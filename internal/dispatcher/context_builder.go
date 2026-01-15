package dispatcher

import (
	"github.com/ethereum/go-ethereum/common"

	abiPkg "github.com/asentric/asentric/internal/abi"
	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/internal/context"
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// DefaultContextBuilder is the standard implementation of ContextBuilder.
type DefaultContextBuilder struct {
	chainID     domain.ChainID   
	abiRegistry domain.ABIRegistry // optional
	decoder     *abiPkg.Decoder    // decoder for ABI decoding
}

// ContextBuilderConfig holds configuration for DefaultContextBuilder.
type ContextBuilderConfig struct {
	ChainID     domain.ChainID   
	ABIRegistry domain.ABIRegistry // optional
}

// NewDefaultContextBuilder creates a new DefaultContextBuilder.
func NewDefaultContextBuilder(cfg ContextBuilderConfig) *DefaultContextBuilder {
	builder := &DefaultContextBuilder{
		chainID:     cfg.ChainID,
		abiRegistry: cfg.ABIRegistry,
	}
	
	// Create decoder if we have an ABI registry
	if cfg.ABIRegistry != nil {
		// Try to cast to *abiPkg.Registry to get decoder
		if reg, ok := cfg.ABIRegistry.(*abiPkg.Registry); ok {
			builder.decoder = abiPkg.NewDecoder(reg)
		}
	}
	
	return builder
}

// Build converts an Event into a Context for Engine evaluation.
func (b *DefaultContextBuilder) Build(event asentric.Event) (asentric.Context, error) {
	// Start building context
	ctx := context.NewEventContext(event)
    
    // Set chain ID
    ctx = ctx.WithChainID(b.chainID)
    
	// Extract and set transaction if available
	if tx := b.extractTransaction(event); tx != (domain.Transaction{}) {
		ctx = ctx.WithTransaction(tx)
	}
    
	// Extract and set block if available
	if block := b.extractBlock(event); block != (domain.Block{}) {
		ctx = ctx.WithBlock(block)
	}
    
    // Extract and set logs if available
    if logs := b.extractLogs(event); len(logs) > 0 {
        ctx = ctx.WithLogs(logs)
    }
    
    // Set ABI registry if available
    if b.abiRegistry != nil {
        ctx = ctx.WithABI(b.abiRegistry)
    }
    
    return ctx, nil
}

// Extracts transaction data from the Event payload.
func (b *DefaultContextBuilder) extractTransaction(event asentric.Event) domain.Transaction {
	// Handle nil or empty payload
	if event.Payload == nil {
		return domain.Transaction{}
	}
	
	// Type assertion: try to get payload as map
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return domain.Transaction{}
	}
	
	// Try to extract transaction data from payload
	txData, ok := payload["transaction"].(map[string]interface{})
	if !ok {
		return domain.Transaction{}
	}
	
	// Build Transaction from extracted data
	tx := domain.Transaction{
		Hash:        domain.Hash(getStringValue(txData, "hash")),
		Index:       getUint64Value(txData, "index"),
		From:        domain.Address(getStringValue(txData, "from")),
		To:          domain.Address(getStringValue(txData, "to")),
		Nonce:       getUint64Value(txData, "nonce"),
		GasLimit:    getUint64Value(txData, "gasLimit"),
		GasUsed:     getUint64Value(txData, "gasUsed"),
		Status:      getBoolValue(txData, "status"),
		GasPrice:    getStringValue(txData, "gasPrice"),
		BlockNumber: event.BlockNumber,
		BlockHash:   domain.Hash(getStringValue(txData, "blockHash")),
		Timestamp:   getUint64Value(txData, "timestamp"),
	}
	
	// Set RawValue (native value/ETH amount)
	if valueStr := getStringValue(txData, "value"); valueStr != "" {
		tx.RawValue = domain.NativeValue{Wei: valueStr}
	}
	
	return tx
}

// Extracts block data from the Event payload.
func (b *DefaultContextBuilder) extractBlock(event asentric.Event) domain.Block {
	// Handle nil or empty payload
	if event.Payload == nil {
		return domain.Block{}
	}
	
	// Type assertion: try to get payload as map
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return domain.Block{}
	}
	
	// Try to extract block data from payload
	blockData, ok := payload["block"].(map[string]interface{})
	if !ok {
		return domain.Block{}
	}
	
	// Build Block from extracted data
	block := domain.Block{
		Number:    event.BlockNumber, // Use event's block number as source of truth
		Hash:      domain.Hash(getStringValue(blockData, "hash")),
		Parent:    domain.Hash(getStringValue(blockData, "parentHash")),
		Timestamp: getUint64Value(blockData, "timestamp"),
		Miner:     domain.Address(getStringValue(blockData, "miner")),
		GasLimit:  getUint64Value(blockData, "gasLimit"),
		GasUsed:   getUint64Value(blockData, "gasUsed"),
		BaseFee:   getStringValue(blockData, "baseFeePerGas"),
		TxCount:   getIntValue(blockData, "transactionCount"),
	}
	
	return block
}

// Extracts log data from the Event payload.
// If ABI registry is available, attempts to decode events using ABI decoder.
func (b *DefaultContextBuilder) extractLogs(event asentric.Event) []domain.Log {
	// Handle nil or empty payload
	if event.Payload == nil {
		return []domain.Log{}
	}
	
	// Type assertion: try to get payload as map
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return []domain.Log{}
	}
	
	// Try to extract logs array from payload
	logsData, ok := payload["logs"].([]interface{})
	if !ok {
		return []domain.Log{}
	}
	
	// Build logs slice
	logs := make([]domain.Log, 0, len(logsData))
	
	for _, logItem := range logsData {
		logMap, ok := logItem.(map[string]interface{})
		if !ok {
			continue // Skip invalid log entries
		}
		
		// Extract event data (decoded event info if already present)
		eventData := domain.Event{}
		if evtMap, ok := logMap["event"].(map[string]interface{}); ok {
			eventData.Name = getStringValue(evtMap, "name")
			if fields, ok := evtMap["fields"].(map[string]interface{}); ok {
				eventData.Fields = fields
			}
		}
		
		// If event is not decoded and we have a decoder, try to decode it
		if eventData.Name == "" && b.decoder != nil {
			// Convert logMap to chain.RawLog for decoding
			if rawLog := b.logMapToRawLog(logMap, event); rawLog != nil {
				decoded := b.decoder.Decode(*rawLog)
				if decoded.Event != nil {
					eventData = b.decoder.ToDomainEvent(decoded.Event)
				}
			}
		}
		
		// Build Log entry
		log := domain.Log{
			Address:     domain.Address(getStringValue(logMap, "address")),
			LogIndex:    getUint64Value(logMap, "logIndex"),
			TxHash:      domain.Hash(event.TxHash),
			TxIndex:     getUint64Value(logMap, "transactionIndex"),
			Event:       eventData,
			BlockNumber: event.BlockNumber,
			BlockHash:   domain.Hash(getStringValue(logMap, "blockHash")),
		}
		
		logs = append(logs, log)
	}
	
	return logs
}

// logMapToRawLog converts a log map to chain.RawLog for ABI decoding.
func (b *DefaultContextBuilder) logMapToRawLog(logMap map[string]interface{}, event asentric.Event) *chain.RawLog {
	addressStr := getStringValue(logMap, "address")
	if addressStr == "" {
		return nil
	}
	
	address := common.HexToAddress(addressStr)
	
	// Extract topics
	topicsInterface, ok := logMap["topics"].([]interface{})
	if !ok {
		return nil
	}
	
	topics := make([]common.Hash, 0, len(topicsInterface))
	for _, t := range topicsInterface {
		if topicStr, ok := t.(string); ok {
			topics = append(topics, common.HexToHash(topicStr))
		}
	}
	
	// Extract data
	dataStr := getStringValue(logMap, "data")
	var data []byte
	if dataStr != "" {
		data = common.FromHex(dataStr)
	}
	
	// Extract block hash
	blockHashStr := getStringValue(logMap, "blockHash")
	blockHash := common.Hash{}
	if blockHashStr != "" {
		blockHash = common.HexToHash(blockHashStr)
	}
	
	return &chain.RawLog{
		Address:     address,
		Topics:      topics,
		Data:        data,
		BlockNumber: event.BlockNumber,
		BlockHash:   blockHash,
		TxHash:      common.HexToHash(event.TxHash),
		TxIndex:     int(getUint64Value(logMap, "transactionIndex")),
		LogIndex:    int(getUint64Value(logMap, "logIndex")),
		Removed:     getBoolValue(logMap, "removed"),
	}
}

// Helper functions for safe type extraction from map[string]interface{}
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getUint64Value safely extracts a uint64 value from a map.
func getUint64Value(m map[string]interface{}, key string) uint64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case uint64:
			return v
		case int64:
			if v >= 0 {
				return uint64(v)
			}
		case int:
			if v >= 0 {
				return uint64(v)
			}
		case float64:
			if v >= 0 {
				return uint64(v)
			}
		}
	}
	return 0
}

// getIntValue safely extracts an int value from a map.
func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case uint64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// getBoolValue safely extracts a bool value from a map.
func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
