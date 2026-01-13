// Package adapter provides type conversion between chain types and domain types.
// This is the bridge between raw blockchain data and our SDK's domain model.
package adapter

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// ToEvent converts a RawLog to an asentric.Event.
// This is the main conversion used by WebSocketSource.
func ToEvent(rawLog chain.RawLog, chainID uint64) asentric.Event {
	return asentric.Event{
		ChainID:     chainID,
		BlockNumber: rawLog.BlockNumber,
		TxHash:      rawLog.TxHash.Hex(),
		Payload:     toLogPayload(rawLog),
	}
}

// toLogPayload creates the payload map from a RawLog.
// The payload structure matches what ContextBuilder expects.
// IMPORTANT: logs must be []interface{} for type assertion in ContextBuilder to work.
func toLogPayload(rawLog chain.RawLog) map[string]interface{} {
	topics := make([]string, len(rawLog.Topics))
	for i, t := range rawLog.Topics {
		topics[i] = t.Hex()
	}

	// Use []interface{} instead of []map[string]interface{}
	// Go doesn't support covariance, so []map[string]interface{} cannot be
	// type-asserted to []interface{} in ContextBuilder.extractLogs()
	logEntry := map[string]interface{}{
		"address":          rawLog.Address.Hex(),
		"topics":           topics,
		"data":             common.Bytes2Hex(rawLog.Data),
		"blockNumber":      rawLog.BlockNumber,
		"transactionHash":  rawLog.TxHash.Hex(),
		"transactionIndex": rawLog.TxIndex,
		"blockHash":        rawLog.BlockHash.Hex(),
		"logIndex":         rawLog.LogIndex,
		"removed":          rawLog.Removed,
	}

	return map[string]interface{}{
		"logs": []interface{}{logEntry},
	}
}

// ToTransaction converts a RawTransaction to domain.Transaction.
func ToTransaction(raw *chain.RawTransaction) domain.Transaction {
	var toAddr domain.Address
	if raw.To != nil {
		toAddr = domain.Address(raw.To.Hex())
	}

	var gasPrice string
	if raw.GasPrice != nil {
		gasPrice = raw.GasPrice.String()
	}

	var valueWei string
	if raw.Value != nil {
		valueWei = raw.Value.String()
	} else {
		valueWei = "0"
	}

	return domain.Transaction{
		Hash:        domain.Hash(raw.Hash.Hex()),
		Index:       uint64(raw.TxIndex),
		From:        domain.Address(raw.From.Hex()),
		To:          toAddr,
		Nonce:       raw.Nonce,
		GasLimit:    raw.GasLimit,
		RawValue:    domain.NativeValue{Wei: valueWei},
		GasPrice:    gasPrice,
		BlockNumber: raw.BlockNumber,
		BlockHash:   domain.Hash(raw.BlockHash.Hex()),
	}
}

// ToBlock converts a RawBlock to domain.Block.
func ToBlock(raw *chain.RawBlock) domain.Block {
	var baseFee string
	if raw.BaseFee != nil {
		baseFee = raw.BaseFee.String()
	}

	var number uint64
	if raw.Number != nil {
		number = raw.Number.Uint64()
	}

	return domain.Block{
		Number:    number,
		Hash:      domain.Hash(raw.Hash.Hex()),
		Parent:    domain.Hash(raw.ParentHash.Hex()),
		Timestamp: raw.Timestamp,
		Miner:     domain.Address(raw.Miner.Hex()),
		GasLimit:  raw.GasLimit,
		GasUsed:   raw.GasUsed,
		BaseFee:   baseFee,
	}
}

// ToLog converts a RawLog to domain.Log (without event decoding).
// For decoded events, use the abi.Decoder.
func ToLog(raw chain.RawLog) domain.Log {
	return domain.Log{
		Address:     domain.Address(raw.Address.Hex()),
		LogIndex:    uint64(raw.LogIndex),
		TxHash:      domain.Hash(raw.TxHash.Hex()),
		TxIndex:     uint64(raw.TxIndex),
		BlockNumber: raw.BlockNumber,
		BlockHash:   domain.Hash(raw.BlockHash.Hex()),
	}
}

// HexToAddress converts a hex string to common.Address.
// Useful for building filters.
func HexToAddress(hex string) common.Address {
	return common.HexToAddress(hex)
}

// HexToHash converts a hex string to common.Hash.
// Useful for topic filtering.
func HexToHash(hex string) common.Hash {
	return common.HexToHash(hex)
}
