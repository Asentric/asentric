package abi

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/pkg/domain"
)

// Decoder decodes raw logs into typed event data using registered ABIs.
type Decoder struct {
	registry *Registry
}

// DecodedLog represents a log with its decoded event information.
type DecodedLog struct {
	// Log is the original raw log
	Log chain.RawLog

	// Event is the decoded event (nil if not decoded)
	Event *DecodedEvent

	// Error is set if decoding failed
	Error error
}

// DecodedEvent represents a decoded smart contract event.
type DecodedEvent struct {
	// Name is the event name (e.g., "Transfer", "Approval")
	Name string

	// Fields contains the decoded event parameters
	Fields map[string]interface{}
}

// NewDecoder creates a new decoder with the given registry.
func NewDecoder(registry *Registry) *Decoder {
	return &Decoder{
		registry: registry,
	}
}

// Decode decodes a raw log into a DecodedLog.
// If the contract's ABI is not registered or the event is unknown, returns log without event.
func (d *Decoder) Decode(rawLog chain.RawLog) DecodedLog {
	result := DecodedLog{Log: rawLog}

	// Get ABI for the contract
	contractABI := d.registry.Get(rawLog.Address)
	if contractABI == nil {
		// No ABI registered - return without decoded event (not an error)
		return result
	}

	// Need at least one topic (the event signature)
	if len(rawLog.Topics) == 0 {
		return result
	}

	// Find event by topic[0] (event signature hash)
	event, err := contractABI.EventByID(rawLog.Topics[0])
	if err != nil {
		// Unknown event - not in ABI
		return result
	}

	// Decode the event
	decoded, err := d.decodeEvent(event, rawLog)
	if err != nil {
		result.Error = err
		return result
	}

	result.Event = decoded
	return result
}

// decodeEvent decodes event parameters from log data.
func (d *Decoder) decodeEvent(event *abi.Event, rawLog chain.RawLog) (*DecodedEvent, error) {
	fields := make(map[string]interface{})

	// Separate indexed and non-indexed arguments
	var indexedArgs, nonIndexedArgs []abi.Argument
	for _, input := range event.Inputs {
		if input.Indexed {
			indexedArgs = append(indexedArgs, input)
		} else {
			nonIndexedArgs = append(nonIndexedArgs, input)
		}
	}

	// Decode indexed arguments from topics[1:]
	topicIndex := 1
	for _, arg := range indexedArgs {
		if topicIndex >= len(rawLog.Topics) {
			break
		}

		topic := rawLog.Topics[topicIndex]
		value, err := d.decodeIndexedArg(arg, topic)
		if err != nil {
			return nil, fmt.Errorf("decode indexed arg %s: %w", arg.Name, err)
		}

		fields[arg.Name] = d.toSerializable(value)
		topicIndex++
	}

	// Decode non-indexed arguments from data
	if len(rawLog.Data) > 0 && len(nonIndexedArgs) > 0 {
		values, err := event.Inputs.UnpackValues(rawLog.Data)
		if err != nil {
			return nil, fmt.Errorf("decode data: %w", err)
		}

		// Map values to non-indexed args
		valueIndex := 0
		for _, arg := range event.Inputs {
			if !arg.Indexed {
				if valueIndex < len(values) {
					fields[arg.Name] = d.toSerializable(values[valueIndex])
					valueIndex++
				}
			}
		}
	}

	return &DecodedEvent{
		Name:   event.Name,
		Fields: fields,
	}, nil
}

// decodeIndexedArg decodes an indexed argument from its topic hash.
func (d *Decoder) decodeIndexedArg(arg abi.Argument, topic common.Hash) (interface{}, error) {
	switch arg.Type.T {
	case abi.AddressTy:
		return common.BytesToAddress(topic.Bytes()), nil

	case abi.UintTy, abi.IntTy:
		return new(big.Int).SetBytes(topic.Bytes()), nil

	case abi.BoolTy:
		return topic.Bytes()[31] != 0, nil

	case abi.BytesTy, abi.StringTy:
		// For indexed bytes/string, the topic is the keccak256 hash
		// The original value isn't available
		return topic.Hex(), nil

	case abi.FixedBytesTy:
		return topic.Bytes()[:arg.Type.Size], nil

	default:
		// Return hex representation for unknown types
		return topic.Hex(), nil
	}
}

// toSerializable converts ABI values to JSON-serializable types.
func (d *Decoder) toSerializable(v interface{}) interface{} {
	switch val := v.(type) {
	case common.Address:
		return val.Hex()

	case common.Hash:
		return val.Hex()

	case *big.Int:
		if val == nil {
			return "0"
		}
		return val.String()

	case []byte:
		return common.Bytes2Hex(val)

	case [32]byte:
		return common.Bytes2Hex(val[:])

	case [20]byte:
		return common.BytesToAddress(val[:]).Hex()

	default:
		return val
	}
}

// ToDomainEvent converts a DecodedEvent to domain.Event.
func (d *Decoder) ToDomainEvent(decoded *DecodedEvent) domain.Event {
	if decoded == nil {
		return domain.Event{}
	}
	return domain.Event{
		Name:   decoded.Name,
		Fields: decoded.Fields,
	}
}

// DecodeLog decodes a raw log and returns a domain.Log with event data.
func (d *Decoder) DecodeLog(rawLog chain.RawLog) domain.Log {
	decoded := d.Decode(rawLog)

	log := domain.Log{
		Address:     domain.Address(rawLog.Address.Hex()),
		LogIndex:    uint64(rawLog.LogIndex),
		TxHash:      domain.Hash(rawLog.TxHash.Hex()),
		TxIndex:     uint64(rawLog.TxIndex),
		BlockNumber: rawLog.BlockNumber,
		BlockHash:   domain.Hash(rawLog.BlockHash.Hex()),
	}

	if decoded.Event != nil {
		log.Event = d.ToDomainEvent(decoded.Event)
	}

	return log
}

// DecodeLogs decodes multiple raw logs.
func (d *Decoder) DecodeLogs(rawLogs []chain.RawLog) []domain.Log {
	logs := make([]domain.Log, len(rawLogs))
	for i, rawLog := range rawLogs {
		logs[i] = d.DecodeLog(rawLog)
	}
	return logs
}
