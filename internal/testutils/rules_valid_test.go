package testutils

import (
	"encoding/json"
	"testing"

	"github.com/asentric/asentric/pkg/asentric"
)

func TestAlwaysAlertRule(t *testing.T) {
	ctx := NewMockContext()
	rule := NewAlwaysAlertRule("test-always-alert")

	// Test Name()
	if rule.Name() != "test-always-alert" {
		t.Errorf("Expected rule name 'test-always-alert', got '%s'", rule.Name())
	}

	// Test Evaluate()
	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	// Verify alert fields
	if alert.Rule != "test-always-alert" {
		t.Errorf("Expected rule 'test-always-alert', got '%s'", alert.Rule)
	}
	if alert.Severity != asentric.SeverityInfo {
		t.Errorf("Expected severity INFO, got %s", alert.Severity)
	}
	if alert.Title != "Test Alert" {
		t.Errorf("Expected title 'Test Alert', got '%s'", alert.Title)
	}

	// Test JSON serialization
	jsonData, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("AlwaysAlertRule JSON output:\n%s", string(jsonData))
}

func TestAlwaysAlertRuleCritical(t *testing.T) {
	ctx := NewMockContext()
	rule := NewAlwaysAlertRuleCritical("test-critical")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	if alert.Severity != asentric.SeverityCritical {
		t.Errorf("Expected severity CRITICAL, got %s", alert.Severity)
	}
	if alert.Title != "Critical Alert" {
		t.Errorf("Expected title 'Critical Alert', got '%s'", alert.Title)
	}

	jsonData, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("AlwaysAlertRuleCritical JSON output:\n%s", string(jsonData))
}

func TestAlwaysAlertRuleHigh(t *testing.T) {
	ctx := NewMockContext()
	rule := NewAlwaysAlertRuleHigh("test-high")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	if alert.Severity != asentric.SeverityHigh {
		t.Errorf("Expected severity HIGH, got %s", alert.Severity)
	}

	jsonData, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("AlwaysAlertRuleHigh JSON output:\n%s", string(jsonData))
}

func TestAlwaysAlertRuleMedium(t *testing.T) {
	ctx := NewMockContext()
	rule := NewAlwaysAlertRuleMedium("test-medium")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	if alert.Severity != asentric.SeverityMedium {
		t.Errorf("Expected severity MEDIUM, got %s", alert.Severity)
	}

	jsonData, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("AlwaysAlertRuleMedium JSON output:\n%s", string(jsonData))
}

func TestAlwaysAlertRuleLow(t *testing.T) {
	ctx := NewMockContext()
	rule := NewAlwaysAlertRuleLow("test-low")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	if alert.Severity != asentric.SeverityLow {
		t.Errorf("Expected severity LOW, got %s", alert.Severity)
	}

	jsonData, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("AlwaysAlertRuleLow JSON output:\n%s", string(jsonData))
}

func TestNeverAlertRule(t *testing.T) {
	ctx := NewMockContext()
	rule := NewNeverAlertRule("test-never")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert != nil {
		t.Errorf("Expected nil alert, got: %+v", alert)
	}

	t.Log("NeverAlertRule correctly returns nil (no alert)")
}

func TestConditionalAlertRule_NoAlert(t *testing.T) {
	// ChainID = 1 (default), should NOT trigger alert
	ctx := NewMockContext()
	rule := NewConditionalAlertRule("test-conditional")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert != nil {
		t.Errorf("Expected nil alert for chainID 1, got: %+v", alert)
	}

	t.Log("ConditionalAlertRule correctly returns nil for chainID <= 1000")
}

func TestConditionalAlertRule_WithAlert(t *testing.T) {
	// ChainID = 56 (BSC), should NOT trigger
	ctx := NewMockContext().WithChainID(56)
	rule := NewConditionalAlertRule("test-conditional")

	alert, err := rule.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if alert != nil {
		t.Errorf("Expected nil alert for chainID 56, got: %+v", alert)
	}

	// ChainID = 5000 (> 1000), should trigger
	ctx2 := NewMockContext().WithChainID(5000)
	alert2, err2 := rule.Evaluate(ctx2)
	if err2 != nil {
		t.Fatalf("Expected no error, got: %v", err2)
	}
	if alert2 == nil {
		t.Fatal("Expected alert for chainID 5000, got nil")
	}

	if alert2.Severity != asentric.SeverityMedium {
		t.Errorf("Expected severity MEDIUM, got %s", alert2.Severity)
	}
	if alert2.Metadata["chainID"] != uint64(5000) {
		t.Errorf("Expected chainID 5000 in metadata, got %v", alert2.Metadata["chainID"])
	}

	jsonData, err := json.MarshalIndent(alert2, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal alert to JSON: %v", err)
	}
	t.Logf("ConditionalAlertRule JSON output:\n%s", string(jsonData))
}

// TestAllRulesImplementInterface verifies all rules implement asentric.Rule
func TestAllRulesImplementInterface(t *testing.T) {
	ctx := NewMockContext()

	rules := []asentric.Rule{
		NewAlwaysAlertRule("test1"),
		NewAlwaysAlertRuleCritical("test2"),
		NewAlwaysAlertRuleHigh("test3"),
		NewAlwaysAlertRuleMedium("test4"),
		NewAlwaysAlertRuleLow("test5"),
		NewNeverAlertRule("test6"),
		NewConditionalAlertRule("test7"),
	}

	for i, rule := range rules {
		if rule.Name() == "" {
			t.Errorf("Rule %d has empty name", i)
		}

		// All should be callable without panic
		_, err := rule.Evaluate(ctx)
		if err != nil {
			t.Errorf("Rule %d (%s) returned unexpected error: %v", i, rule.Name(), err)
		}
	}

	t.Logf("Successfully verified %d rules implement asentric.Rule interface", len(rules))
}

// TestAlertJSONSerialization demonstrates JSON serialization of all alert types
func TestAlertJSONSerialization(t *testing.T) {
	ctx := NewMockContext().WithChainID(5000)

	rules := map[string]asentric.Rule{
		"INFO":        NewAlwaysAlertRule("alert-info"),
		"CRITICAL":    NewAlwaysAlertRuleCritical("alert-critical"),
		"HIGH":        NewAlwaysAlertRuleHigh("alert-high"),
		"MEDIUM":      NewAlwaysAlertRuleMedium("alert-medium"),
		"LOW":         NewAlwaysAlertRuleLow("alert-low"),
		"CONDITIONAL": NewConditionalAlertRule("alert-conditional"),
	}

	t.Log("\n=== JSON Serialization Examples ===")
	for severity, rule := range rules {
		alert, err := rule.Evaluate(ctx)
		if err != nil {
			t.Errorf("Rule %s failed: %v", severity, err)
			continue
		}
		if alert == nil {
			t.Logf("\n[%s] No alert generated", severity)
			continue
		}

		jsonData, err := json.MarshalIndent(alert, "", "  ")
		if err != nil {
			t.Errorf("Failed to marshal %s alert: %v", severity, err)
			continue
		}

		t.Logf("\n[%s] Alert JSON:\n%s", severity, string(jsonData))
	}
}
