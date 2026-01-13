// Package internal provides integration tests for the Asentric SDK.
// These tests verify the complete event processing flow from source to sink.
package internal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/asentric/asentric/internal/abi"
	"github.com/asentric/asentric/internal/dispatcher"
	"github.com/asentric/asentric/internal/sink"
	"github.com/asentric/asentric/internal/testutils"
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// createTestDispatcher is a helper to create a dispatcher with common setup.
func createTestDispatcher(engine *asentric.Engine, alertSink asentric.AlertSink, chainID domain.ChainID) (*dispatcher.EngineDispatcher, error) {
	abiRegistry := abi.NewRegistry()

	contextBuilder := dispatcher.NewDefaultContextBuilder(dispatcher.ContextBuilderConfig{
		ChainID:     chainID,
		ABIRegistry: abiRegistry,
	})

	return dispatcher.NewEngineDispatcher(dispatcher.EngineDispatcherConfig{
		Engine:         engine,
		AlertSink:      alertSink,
		ContextBuilder: contextBuilder,
		ABIRegistry:    abiRegistry,
	})
}

// TestIntegration_FullFlow tests the complete event processing flow:
// Event → ContextBuilder → Engine → AlertSink
func TestIntegration_FullFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Create engine with test rule
	engine := asentric.NewEngine()
	rule := testutils.NewLargeTransferRule("large-transfer", asentric.SeverityHigh)
	if err := engine.RegisterRule(rule); err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// 2. Create mock sink to capture alerts
	mockSink := testutils.NewMockAlertSink()

	// 3. Create dispatcher
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(5000))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	// 4. Create test event with large value (should trigger alert)
	event := asentric.Event{
		ChainID:     5000,
		BlockNumber: 12345,
		TxHash:      "0xabc123def456",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash":   "0xabc123def456",
				"from":   "0x1111111111111111111111111111111111111111",
				"to":     "0x2222222222222222222222222222222222222222",
				"value":  "5000000000000000000", // 5 ETH - large transfer
				"status": true,
			},
		},
	}

	// 5. Dispatch event
	if err := disp.Dispatch(ctx, event); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// 6. Verify alert was emitted
	if mockSink.AlertCount() != 1 {
		t.Errorf("Expected 1 alert, got %d", mockSink.AlertCount())
	}

	alerts := mockSink.Alerts()
	if len(alerts) > 0 {
		if alerts[0].Alert.Rule != "large-transfer" {
			t.Errorf("Expected rule 'large-transfer', got '%s'", alerts[0].Alert.Rule)
		}
		if alerts[0].Alert.Severity != asentric.SeverityHigh {
			t.Errorf("Expected severity HIGH, got '%s'", alerts[0].Alert.Severity)
		}
	}
}

// TestIntegration_NoAlert tests that small transfers don't trigger alerts.
func TestIntegration_NoAlert(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewLargeTransferRule("large-transfer", asentric.SeverityHigh))

	mockSink := testutils.NewMockAlertSink()
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	// Small transfer - should NOT trigger alert
	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0x123",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash":   "0x123",
				"value":  "100000000000000000", // 0.1 ETH - small
				"status": true,
			},
		},
	}

	disp.Dispatch(ctx, event)

	if mockSink.AlertCount() != 0 {
		t.Errorf("Expected 0 alerts for small transfer, got %d", mockSink.AlertCount())
	}
}

// TestIntegration_MultipleRules tests evaluation of multiple rules.
func TestIntegration_MultipleRules(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewFlexibleAlertRule("rule-1", asentric.SeverityCritical))
	engine.RegisterRule(testutils.NewFlexibleAlertRule("rule-2", asentric.SeverityHigh))
	engine.RegisterRule(testutils.NewNeverAlertRule("rule-3"))

	mockSink := testutils.NewMockAlertSink()
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0x123",
	}

	if err := disp.Dispatch(ctx, event); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// Should have 2 alerts (rule-1 and rule-2, not rule-3)
	if mockSink.AlertCount() != 2 {
		t.Errorf("Expected 2 alerts, got %d", mockSink.AlertCount())
	}

	// Verify severities
	alerts := mockSink.Alerts()
	severities := make(map[asentric.Severity]bool)
	for _, a := range alerts {
		severities[a.Alert.Severity] = true
	}

	if !severities[asentric.SeverityCritical] {
		t.Error("Expected CRITICAL alert")
	}
	if !severities[asentric.SeverityHigh] {
		t.Error("Expected HIGH alert")
	}
}

// TestIntegration_ContextData tests that context data is properly passed to rules.
func TestIntegration_ContextData(t *testing.T) {
	ctx := context.Background()

	// Create a rule that captures context for verification
	capturedContext := make(chan asentric.Context, 1)
	rule := &contextCaptureRule{
		name:    "capture-rule",
		capture: capturedContext,
	}

	engine := asentric.NewEngine()
	engine.RegisterRule(rule)

	mockSink := testutils.NewMockAlertSink()
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(5000))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	event := asentric.Event{
		ChainID:     5000,
		BlockNumber: 12345,
		TxHash:      "0xdef456",
		Payload: map[string]interface{}{
			"transaction": map[string]interface{}{
				"hash":        "0xdef456",
				"from":        "0xaaaa",
				"to":          "0xbbbb",
				"value":       "1000000000000000000",
				"blockNumber": float64(12345),
			},
		},
	}

	disp.Dispatch(ctx, event)

	// Check captured context
	select {
	case capturedCtx := <-capturedContext:
		if capturedCtx.ChainID() != 5000 {
			t.Errorf("Expected ChainID 5000, got %d", capturedCtx.ChainID())
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for context capture")
	}
}

// TestIntegration_Concurrency tests concurrent event processing.
func TestIntegration_Concurrency(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewFlexibleAlertRule("concurrent-rule", asentric.SeverityLow))

	mockSink := testutils.NewMockAlertSink()
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	// Dispatch 100 events concurrently
	numEvents := 100
	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(idx int) {
			defer wg.Done()
			event := asentric.Event{
				ChainID:     1,
				BlockNumber: uint64(idx),
				TxHash:      "0x" + string(rune('a'+idx%26)),
			}
			disp.Dispatch(ctx, event)
		}(i)
	}

	wg.Wait()

	// Should have 100 alerts
	if mockSink.AlertCount() != numEvents {
		t.Errorf("Expected %d alerts, got %d", numEvents, mockSink.AlertCount())
	}
}

// TestIntegration_WithConsoleSink tests with ConsoleSink for debugging.
func TestIntegration_WithConsoleSink(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewFlexibleAlertRule("console-test", asentric.SeverityInfo))

	// Use console sink (output goes to stdout)
	consoleSink := sink.NewConsoleSink()
	disp, err := createTestDispatcher(engine, consoleSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 12345,
		TxHash:      "0xtest123",
	}

	// Should not error
	if err := disp.Dispatch(ctx, event); err != nil {
		t.Fatalf("Dispatch with ConsoleSink failed: %v", err)
	}
}

// TestIntegration_WithMultiSink tests with multiple sinks.
func TestIntegration_WithMultiSink(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewFlexibleAlertRule("multi-test", asentric.SeverityMedium))

	// Create multiple sinks
	mockSink1 := testutils.NewMockAlertSink()
	mockSink2 := testutils.NewMockAlertSink()
	multiSink := sink.NewMultiSink(mockSink1, mockSink2)

	disp, err := createTestDispatcher(engine, multiSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	event := asentric.Event{
		ChainID:     1,
		BlockNumber: 100,
		TxHash:      "0xmulti",
	}

	disp.Dispatch(ctx, event)

	// Both sinks should receive the alert
	if mockSink1.AlertCount() != 1 {
		t.Errorf("Sink1: expected 1 alert, got %d", mockSink1.AlertCount())
	}
	if mockSink2.AlertCount() != 1 {
		t.Errorf("Sink2: expected 1 alert, got %d", mockSink2.AlertCount())
	}
}

// TestIntegration_RulePanic tests that panicking rules don't crash the system.
func TestIntegration_RulePanic(t *testing.T) {
	ctx := context.Background()

	engine := asentric.NewEngine()
	engine.RegisterRule(testutils.NewPanicRule("panic-rule"))
	engine.RegisterRule(testutils.NewFlexibleAlertRule("after-panic", asentric.SeverityLow))

	mockSink := testutils.NewMockAlertSink()
	disp, err := createTestDispatcher(engine, mockSink, domain.ChainID(1))
	if err != nil {
		t.Fatalf("Failed to create dispatcher: %v", err)
	}

	event := asentric.Event{ChainID: 1, BlockNumber: 1}

	// Should not panic - engine recovers from rule panics
	err = disp.Dispatch(ctx, event)

	// Engine returns error when rule panics
	if err != nil {
		t.Logf("Dispatch returned error (expected for panic): %v", err)
	}
}

// Helper rule that captures context for testing
type contextCaptureRule struct {
	name    string
	capture chan<- asentric.Context
}

func (r *contextCaptureRule) Name() string {
	return r.name
}

func (r *contextCaptureRule) Severity() asentric.Severity {
	return asentric.SeverityInfo
}

func (r *contextCaptureRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	select {
	case r.capture <- ctx:
	default:
	}
	return nil, nil
}
