package adapter

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/asentric/asentric/internal/chain"
)

func TestToEvent(t *testing.T) {
	rawLog := chain.RawLog{
		Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Topics:      []common.Hash{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		Data:        []byte{1, 2, 3},
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabc123def456"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0xdef456"),
		LogIndex:    0,
	}

	event := ToEvent(rawLog, 5000)

	if event.ChainID != 5000 {
		t.Errorf("Expected ChainID 5000, got %d", event.ChainID)
	}

	if event.BlockNumber != 12345 {
		t.Errorf("Expected BlockNumber 12345, got %d", event.BlockNumber)
	}

	if event.TxHash != rawLog.TxHash.Hex() {
		t.Errorf("Expected TxHash %s, got %s", rawLog.TxHash.Hex(), event.TxHash)
	}

	if event.Payload == nil {
		t.Error("Expected Payload to be non-nil")
	}

	// Verify payload structure
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		t.Fatal("Payload should be a map")
	}

	logs, ok := payload["logs"].([]map[string]interface{})
	if !ok || len(logs) == 0 {
		t.Fatal("Payload should have logs array")
	}

	if logs[0]["address"] != rawLog.Address.Hex() {
		t.Errorf("Log address mismatch")
	}
}

func TestToTransaction(t *testing.T) {
	to := common.HexToAddress("0x9876543210987654321098765432109876543210")
	raw := &chain.RawTransaction{
		Hash:        common.HexToHash("0xabc123"),
		Nonce:       5,
		GasPrice:    big.NewInt(20000000000),
		GasLimit:    21000,
		To:          &to,
		Value:       big.NewInt(1000000000000000000),
		From:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		BlockNumber: 12345,
		BlockHash:   common.HexToHash("0xblock123"),
		TxIndex:     3,
	}

	tx := ToTransaction(raw)

	if tx.Nonce != 5 {
		t.Errorf("Expected Nonce 5, got %d", tx.Nonce)
	}

	if tx.GasLimit != 21000 {
		t.Errorf("Expected GasLimit 21000, got %d", tx.GasLimit)
	}

	if tx.RawValue.Wei != "1000000000000000000" {
		t.Errorf("Expected Wei 1e18, got %s", tx.RawValue.Wei)
	}

	if tx.BlockNumber != 12345 {
		t.Errorf("Expected BlockNumber 12345, got %d", tx.BlockNumber)
	}

	if tx.Index != 3 {
		t.Errorf("Expected Index 3, got %d", tx.Index)
	}
}

func TestToTransaction_NilTo(t *testing.T) {
	raw := &chain.RawTransaction{
		Hash:     common.HexToHash("0xabc123"),
		To:       nil, // Contract creation
		Value:    big.NewInt(0),
		GasPrice: big.NewInt(0),
	}

	tx := ToTransaction(raw)

	if tx.To != "" {
		t.Errorf("Expected empty To for contract creation, got %s", tx.To)
	}
}

func TestToBlock(t *testing.T) {
	raw := &chain.RawBlock{
		Number:     big.NewInt(12345),
		Hash:       common.HexToHash("0xblock123"),
		ParentHash: common.HexToHash("0xparent456"),
		Timestamp:  1704067200,
		GasLimit:   30000000,
		GasUsed:    15000000,
		BaseFee:    big.NewInt(1000000000),
		Miner:      common.HexToAddress("0xminer"),
	}

	block := ToBlock(raw)

	if block.Number != 12345 {
		t.Errorf("Expected Number 12345, got %d", block.Number)
	}

	if block.Timestamp != 1704067200 {
		t.Errorf("Expected Timestamp 1704067200, got %d", block.Timestamp)
	}

	if block.GasLimit != 30000000 {
		t.Errorf("Expected GasLimit 30000000, got %d", block.GasLimit)
	}

	if block.BaseFee != "1000000000" {
		t.Errorf("Expected BaseFee 1000000000, got %s", block.BaseFee)
	}
}

func TestToLog(t *testing.T) {
	raw := chain.RawLog{
		Address:     common.HexToAddress("0xcontract"),
		LogIndex:    5,
		TxHash:      common.HexToHash("0xtx123"),
		TxIndex:     2,
		BlockNumber: 100,
		BlockHash:   common.HexToHash("0xblock"),
	}

	log := ToLog(raw)

	if log.LogIndex != 5 {
		t.Errorf("Expected LogIndex 5, got %d", log.LogIndex)
	}

	if log.TxIndex != 2 {
		t.Errorf("Expected TxIndex 2, got %d", log.TxIndex)
	}

	if log.BlockNumber != 100 {
		t.Errorf("Expected BlockNumber 100, got %d", log.BlockNumber)
	}
}

func TestHexToAddress(t *testing.T) {
	addr := HexToAddress("0x1234567890123456789012345678901234567890")
	expected := common.HexToAddress("0x1234567890123456789012345678901234567890")

	if addr != expected {
		t.Errorf("Address mismatch")
	}
}

func TestHexToHash(t *testing.T) {
	hash := HexToHash("0xabc123")
	expected := common.HexToHash("0xabc123")

	if hash != expected {
		t.Errorf("Hash mismatch")
	}
}
