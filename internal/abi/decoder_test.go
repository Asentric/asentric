package abi

import (
	"math/big"
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/asentric/asentric/internal/chain"
)

const transferABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

func TestDecoder_DecodeTransfer(t *testing.T) {
	// Setup registry with ERC20 ABI
	registry := NewRegistry()
	parsed, _ := ethabi.JSON(strings.NewReader(transferABI))

	contractAddr := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC
	registry.Register(contractAddr, parsed)

	decoder := NewDecoder(registry)

	// Create a mock Transfer log
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(1000000) // 1 USDC (6 decimals)

	// Pack value for data field (32 bytes)
	valueBytes := common.LeftPadBytes(value.Bytes(), 32)

	rawLog := chain.RawLog{
		Address: contractAddr,
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			common.BytesToHash(from.Bytes()),
			common.BytesToHash(to.Bytes()),
		},
		Data:        valueBytes,
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabc123"),
		LogIndex:    0,
	}

	// Decode
	result := decoder.Decode(rawLog)

	if result.Error != nil {
		t.Fatalf("Decode error: %v", result.Error)
	}

	if result.Event == nil {
		t.Fatal("Expected decoded event")
	}

	if result.Event.Name != "Transfer" {
		t.Errorf("Expected event name 'Transfer', got '%s'", result.Event.Name)
	}

	// Check fields
	if result.Event.Fields["from"] != from.Hex() {
		t.Errorf("Expected from %s, got %v", from.Hex(), result.Event.Fields["from"])
	}

	if result.Event.Fields["to"] != to.Hex() {
		t.Errorf("Expected to %s, got %v", to.Hex(), result.Event.Fields["to"])
	}

	if result.Event.Fields["value"] != value.String() {
		t.Errorf("Expected value %s, got %v", value.String(), result.Event.Fields["value"])
	}
}

func TestDecoder_UnknownContract(t *testing.T) {
	registry := NewRegistry()
	decoder := NewDecoder(registry)

	rawLog := chain.RawLog{
		Address:     common.HexToAddress("0x9999999999999999999999999999999999999999"),
		Topics:      []common.Hash{common.HexToHash("0xabc123")},
		BlockNumber: 12345,
	}

	result := decoder.Decode(rawLog)

	// Should not error, just no decoded event
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	if result.Event != nil {
		t.Error("Expected nil event for unknown contract")
	}
}

func TestDecoder_UnknownEvent(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := ethabi.JSON(strings.NewReader(transferABI))
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	registry.Register(contractAddr, parsed)

	decoder := NewDecoder(registry)

	rawLog := chain.RawLog{
		Address: contractAddr,
		Topics:  []common.Hash{common.HexToHash("0xunknownevent")},
	}

	result := decoder.Decode(rawLog)

	// No error, but no event either
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}

	if result.Event != nil {
		t.Error("Expected nil event for unknown event signature")
	}
}

func TestDecoder_NoTopics(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := ethabi.JSON(strings.NewReader(transferABI))
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	registry.Register(contractAddr, parsed)

	decoder := NewDecoder(registry)

	rawLog := chain.RawLog{
		Address: contractAddr,
		Topics:  []common.Hash{}, // No topics
	}

	result := decoder.Decode(rawLog)

	if result.Event != nil {
		t.Error("Expected nil event for log with no topics")
	}
}

func TestDecoder_DecodeLog(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := ethabi.JSON(strings.NewReader(transferABI))
	contractAddr := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	registry.Register(contractAddr, parsed)

	decoder := NewDecoder(registry)

	rawLog := chain.RawLog{
		Address: contractAddr,
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
		},
		Data:        common.LeftPadBytes(big.NewInt(100).Bytes(), 32),
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xdef456"),
		LogIndex:    5,
		TxIndex:     2,
		BlockHash:   common.HexToHash("0xblock123"),
	}

	domainLog := decoder.DecodeLog(rawLog)

	if domainLog.BlockNumber != 12345 {
		t.Errorf("Expected BlockNumber 12345, got %d", domainLog.BlockNumber)
	}

	if domainLog.LogIndex != 5 {
		t.Errorf("Expected LogIndex 5, got %d", domainLog.LogIndex)
	}

	if domainLog.TxIndex != 2 {
		t.Errorf("Expected TxIndex 2, got %d", domainLog.TxIndex)
	}
}

func TestDecoder_DecodeLogs(t *testing.T) {
	registry := NewRegistry()
	decoder := NewDecoder(registry)

	rawLogs := []chain.RawLog{
		{Address: common.HexToAddress("0x1111"), BlockNumber: 100},
		{Address: common.HexToAddress("0x2222"), BlockNumber: 200},
	}

	domainLogs := decoder.DecodeLogs(rawLogs)

	if len(domainLogs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(domainLogs))
	}

	if domainLogs[0].BlockNumber != 100 {
		t.Errorf("Expected first log BlockNumber 100, got %d", domainLogs[0].BlockNumber)
	}

	if domainLogs[1].BlockNumber != 200 {
		t.Errorf("Expected second log BlockNumber 200, got %d", domainLogs[1].BlockNumber)
	}
}

func TestDecoder_ToDomainEvent(t *testing.T) {
	decoder := NewDecoder(NewRegistry())

	decoded := &DecodedEvent{
		Name:   "Transfer",
		Fields: map[string]interface{}{"from": "0x123", "value": "100"},
	}

	domainEvent := decoder.ToDomainEvent(decoded)

	if domainEvent.Name != "Transfer" {
		t.Errorf("Expected name 'Transfer', got '%s'", domainEvent.Name)
	}

	if domainEvent.Fields["from"] != "0x123" {
		t.Errorf("Expected from '0x123', got '%v'", domainEvent.Fields["from"])
	}
}

func TestDecoder_ToDomainEvent_Nil(t *testing.T) {
	decoder := NewDecoder(NewRegistry())

	domainEvent := decoder.ToDomainEvent(nil)

	if domainEvent.Name != "" {
		t.Error("Expected empty name for nil input")
	}
}
