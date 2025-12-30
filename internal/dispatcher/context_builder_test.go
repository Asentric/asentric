package dispatcher

import (
	"testing"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

func TestNewDefaultContextBuilder(t *testing.T) {
	cfg := ContextBuilderConfig{
		ChainID: 1,
	}

	builder := NewDefaultContextBuilder(cfg)

	if builder == nil {
		t.Fatal("Expected builder, got nil")
	}
	if builder.chainID != 1 {
		t.Errorf("Expected chainID 1, got %d", builder.chainID)
	}
}

func TestBuild_NilEvent(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	// asentric.Event is a struct, so we pass zero value to test nil payload
	event := asentric.Event{}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if ctx == nil {
		t.Fatal("Expected context, got nil")
	}
}

func TestBuild_BasicEvent(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0x123",
		Payload:     make(map[string]interface{}),
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if ctx == nil {
		t.Fatal("Expected context, got nil")
	}
	if ctx.ChainID() != 1 {
		t.Errorf("Expected chainID 1, got %d", ctx.ChainID())
	}
}

func TestBuild_WithTransaction(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0xabc",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash":  "0xabc",
				"from":  "0x111",
				"to":    "0x222",
				"nonce": float64(5),
			},
		},
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify transaction was set
	ctxTx := ctx.Tx()
	if ctxTx.Hash == "" {
		t.Fatal("Expected transaction in context")
	}
	if ctxTx.From != domain.Address("0x111") {
		t.Errorf("Expected from 0x111, got %s", ctxTx.From)
	}
}

func TestBuild_WithBlock(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 12345,
		TxHash:      "0xabc",
		Payload: map[string]interface{}{
			"block": map[string]interface{}{
				"hash":      "0xblock",
				"number":    float64(12345),
				"timestamp": float64(1234567890),
			},
		},
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	ctxBlock := ctx.Block()
	if ctxBlock.Number == 0 {
		t.Fatal("Expected block in context")
	}
	if ctxBlock.Number != 12345 {
		t.Errorf("Expected block number 12345, got %d", ctxBlock.Number)
	}
}

func TestBuild_WithLogs(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0xabc",
		Payload: map[string]interface{}{
			"logs": []interface{}{
				map[string]interface{}{
					"address":     "0x123",
					"blockNumber": float64(100),
					"logIndex":    float64(0),
				},
				map[string]interface{}{
					"address":     "0x456",
					"blockNumber": float64(100),
					"logIndex":    float64(1),
				},
			},
		},
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	ctxLogs := ctx.Logs()
	if len(ctxLogs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(ctxLogs))
	}
}

func TestBuild_TransactionFromMap(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0xabc123",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash":  "0xabc123",
				"from":  "0x111",
				"to":    "0x222",
				"nonce": float64(5),
			},
		},
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	ctxTx := ctx.Tx()
	if ctxTx.Hash == "" {
		t.Fatal("Expected transaction in context")
	}
	if ctxTx.Nonce != 5 {
		t.Errorf("Expected nonce 5, got %d", ctxTx.Nonce)
	}
}

func TestBuild_EmptyData(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 5000})

	event := asentric.Event{
		ChainID:     5000,
		BlockNumber: 100,
		TxHash:      "0x123",
		Payload:     nil,
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should still have chainID
	if ctx.ChainID() != 5000 {
		t.Errorf("Expected chainID 5000, got %d", ctx.ChainID())
	}

	// Optional fields should be empty/zero value
	ctxTx := ctx.Tx()
	if ctxTx.Hash != "" {
		t.Error("Expected empty transaction for empty data")
	}
}

func TestBuild_CompleteEvent(t *testing.T) {
	builder := NewDefaultContextBuilder(ContextBuilderConfig{ChainID: 1})

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 999,
		TxHash:      "0xtx",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash": "0xtx",
				"from": "0xsender",
				"to":   "0xToken",
			},
			"block": map[string]interface{}{
				"number": float64(999),
			},
			"logs": []interface{}{
				map[string]interface{}{
					"address": "0xToken",
				},
			},
		},
	}

	ctx, err := builder.Build(event)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify all data is present
	if ctx.Tx().Hash == "" {
		t.Error("Expected transaction in context")
	}
	if ctx.Block().Number == 0 {
		t.Error("Expected block in context")
	}
	if len(ctx.Logs()) != 1 {
		t.Error("Expected 1 log in context")
	}
}
