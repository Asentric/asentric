package testutils

import (
	"testing"

	"github.com/asentric/asentric/pkg/domain"
)

func TestNewMockContext(t *testing.T) {
	ctx := NewMockContext()

	if ctx == nil {
		t.Fatal("NewMockContext returned nil")
	}

	// Test default chain ID
	if ctx.ChainID() != domain.ChainID(1) {
		t.Errorf("Expected chainID 1, got %d", ctx.ChainID())
	}

	// Test default transaction
	tx := ctx.Tx()
	if tx.Hash != domain.Hash("0xdefault123456789abcdef") {
		t.Errorf("Expected default tx hash, got %s", tx.Hash)
	}

	// Test default block
	block := ctx.Block()
	if block.Number != 1000000 {
		t.Errorf("Expected block number 1000000, got %d", block.Number)
	}

	// Test default logs
	logs := ctx.Logs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	// Test ABI is nil by default
	if ctx.ABI() != nil {
		t.Error("Expected nil ABI")
	}
}

func TestMockContext_WithChainID(t *testing.T) {
	ctx := NewMockContext().WithChainID(domain.ChainID(137)) // Polygon

	if ctx.ChainID() != domain.ChainID(137) {
		t.Errorf("Expected chainID 137, got %d", ctx.ChainID())
	}
}

func TestMockContext_WithTransaction(t *testing.T) {
	customTx := &domain.Transaction{
		Hash:  domain.Hash("0xcustom123"),
		Nonce: 42,
	}

	ctx := NewMockContext().WithTransaction(customTx)
	tx := ctx.Tx()

	if tx.Hash != domain.Hash("0xcustom123") {
		t.Errorf("Expected custom tx hash, got %s", tx.Hash)
	}
	if tx.Nonce != 42 {
		t.Errorf("Expected nonce 42, got %d", tx.Nonce)
	}
}

func TestMockContext_WithBlock(t *testing.T) {
	customBlock := &domain.Block{
		Hash:   domain.Hash("0xcustomblock"),
		Number: 999,
	}

	ctx := NewMockContext().WithBlock(customBlock)
	block := ctx.Block()

	if block.Hash != domain.Hash("0xcustomblock") {
		t.Errorf("Expected custom block hash, got %s", block.Hash)
	}
	if block.Number != 999 {
		t.Errorf("Expected block number 999, got %d", block.Number)
	}
}

func TestMockContext_WithLogs(t *testing.T) {
	customLogs := []domain.Log{
		{
			Address:     domain.Address("0xCustomContract"),
			BlockNumber: 12345,
			LogIndex:    5,
		},
		{
			Address:     domain.Address("0xAnotherContract"),
			BlockNumber: 12346,
			LogIndex:    6,
		},
	}

	ctx := NewMockContext().WithLogs(customLogs)
	logs := ctx.Logs()

	if len(logs) != 2 {
		t.Fatalf("Expected 2 logs, got %d", len(logs))
	}
	if logs[0].Address != domain.Address("0xCustomContract") {
		t.Errorf("Expected first log address 0xCustomContract, got %s", logs[0].Address)
	}
	if logs[1].LogIndex != 6 {
		t.Errorf("Expected second log index 6, got %d", logs[1].LogIndex)
	}
}

func TestMockContext_MethodChaining(t *testing.T) {
	// Test fluent API / method chaining
	ctx := NewMockContext().
		WithChainID(domain.ChainID(56)). // BSC
		WithTransaction(&domain.Transaction{Hash: domain.Hash("0xchain")}).
		WithLogs([]domain.Log{})

	if ctx.ChainID() != domain.ChainID(56) {
		t.Error("Method chaining failed for ChainID")
	}
	if ctx.Tx().Hash != domain.Hash("0xchain") {
		t.Error("Method chaining failed for Transaction")
	}
	if len(ctx.Logs()) != 0 {
		t.Error("Method chaining failed for Logs")
	}
}

func TestMockContext_NilTransaction(t *testing.T) {
	ctx := NewMockContext().WithTransaction(nil)
	tx := ctx.Tx()

	// Should return zero value, not panic
	if tx.Hash != "" {
		t.Error("Expected empty hash for nil transaction")
	}
}

func TestMockContext_NilBlock(t *testing.T) {
	ctx := NewMockContext().WithBlock(nil)
	block := ctx.Block()

	// Should return zero value, not panic
	if block.Number != 0 {
		t.Error("Expected 0 block number for nil block")
	}
}

func TestDefaultTransaction(t *testing.T) {
	tx := defaultTransaction()

	if tx == nil {
		t.Fatal("defaultTransaction returned nil")
	}
	if tx.From != domain.Address("0xSender123") {
		t.Errorf("Expected from 0xSender123, got %s", tx.From)
	}
	if tx.To != domain.Address("0xReceiver456") {
		t.Errorf("Expected to 0xReceiver456, got %s", tx.To)
	}
	if tx.GasLimit != 21000 {
		t.Errorf("Expected gas limit 21000, got %d", tx.GasLimit)
	}
}

func TestDefaultBlock(t *testing.T) {
	block := defaultBlock()

	if block == nil {
		t.Fatal("defaultBlock returned nil")
	}
	if block.Number != 1000000 {
		t.Errorf("Expected block number 1000000, got %d", block.Number)
	}
	if block.Miner != domain.Address("0xMiner789") {
		t.Errorf("Expected miner 0xMiner789, got %s", block.Miner)
	}
}

func TestDefaultLogs(t *testing.T) {
	logs := defaultLogs()

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]
	if log.Address != domain.Address("0xContract123") {
		t.Errorf("Expected address 0xContract123, got %s", log.Address)
	}
	if log.LogIndex != 0 {
		t.Errorf("Expected log index 0, got %d", log.LogIndex)
	}
	if log.TxHash != domain.Hash("0xdefault123456789abcdef") {
		t.Errorf("Expected tx hash 0xdefault123456789abcdef, got %s", log.TxHash)
	}
	if log.TxIndex != 0 {
		t.Errorf("Expected tx index 0, got %d", log.TxIndex)
	}
	if log.Event.Name != "Transfer" {
		t.Errorf("Expected event name Transfer, got %s", log.Event.Name)
	}
	if log.BlockNumber != 1000000 {
		t.Errorf("Expected block number 1000000, got %d", log.BlockNumber)
	}
	if log.BlockHash != domain.Hash("0xblock123") {
		t.Errorf("Expected block hash 0xblock123, got %s", log.BlockHash)
	}
}
