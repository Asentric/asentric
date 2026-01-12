package sink

import (
	"encoding/json"
	"testing"

	"github.com/asentric/asentric/internal/testutils"
	"github.com/asentric/asentric/pkg/asentric"
)

func TestFormatter_Format(t *testing.T) {
	formatter := DefaultFormatter()
	ctx := testutils.NewMockContext()

	alert := asentric.NewAlert("test-rule", "Test Alert", asentric.SeverityHigh).
		WithDescription("This is a test").
		WithMetadata("key", "value")

	payload := formatter.Format(ctx, alert)

	if payload.Rule != "test-rule" {
		t.Errorf("Expected rule 'test-rule', got '%s'", payload.Rule)
	}

	if payload.Severity != "HIGH" {
		t.Errorf("Expected severity 'HIGH', got '%s'", payload.Severity)
	}

	if payload.Title != "Test Alert" {
		t.Errorf("Expected title 'Test Alert', got '%s'", payload.Title)
	}

	if payload.Description != "This is a test" {
		t.Errorf("Expected description 'This is a test', got '%s'", payload.Description)
	}

	if payload.Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", payload.Metadata)
	}

	if payload.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
}

func TestFormatter_Format_WithContext(t *testing.T) {
	formatter := NewFormatter(FormatterConfig{IncludeContext: true})
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityInfo)

	payload := formatter.Format(ctx, alert)

	if payload.Context == nil {
		t.Fatal("Expected context to be included")
	}

	if payload.Context.ChainID != 1 {
		t.Errorf("Expected ChainID 1, got %d", payload.Context.ChainID)
	}
}

func TestFormatter_Format_WithoutContext(t *testing.T) {
	formatter := NewFormatter(FormatterConfig{IncludeContext: false})
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityInfo)

	payload := formatter.Format(ctx, alert)

	if payload.Context != nil {
		t.Error("Expected context to be nil when IncludeContext=false")
	}
}

func TestFormatter_Format_WithRef(t *testing.T) {
	formatter := DefaultFormatter()
	ctx := testutils.NewMockContext()

	alert := asentric.NewAlert("test-rule", "Test", asentric.SeverityMedium).
		WithRef(asentric.NewExecutionRef("0xabc123", 12345, 3))

	payload := formatter.Format(ctx, alert)

	if payload.Ref == nil {
		t.Fatal("Expected Ref to be set")
	}

	if payload.Ref.TxHash != "0xabc123" {
		t.Errorf("Expected TxHash '0xabc123', got '%s'", payload.Ref.TxHash)
	}

	if payload.Ref.BlockNumber != 12345 {
		t.Errorf("Expected BlockNumber 12345, got %d", payload.Ref.BlockNumber)
	}

	if payload.Ref.LogIndex != 3 {
		t.Errorf("Expected LogIndex 3, got %d", payload.Ref.LogIndex)
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	formatter := DefaultFormatter()
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test-rule", "Test Alert", asentric.SeverityHigh)

	jsonData, err := formatter.FormatJSON(ctx, alert)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if parsed["rule"] != "test-rule" {
		t.Errorf("Expected rule 'test-rule', got '%v'", parsed["rule"])
	}
}

func TestFormatter_FormatJSONPretty(t *testing.T) {
	formatter := DefaultFormatter()
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityLow)

	jsonData, err := formatter.FormatJSONPretty(ctx, alert)
	if err != nil {
		t.Fatalf("FormatJSONPretty error: %v", err)
	}

	// Should contain indentation
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

func TestDefaultFormatter(t *testing.T) {
	formatter := DefaultFormatter()
	if formatter == nil {
		t.Fatal("DefaultFormatter should return non-nil")
	}

	// Default should include context
	if !formatter.includeContext {
		t.Error("Default formatter should include context")
	}
}
