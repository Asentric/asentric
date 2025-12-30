package testutils

import (
	"strings"
	"testing"

	"github.com/asentric/asentric/pkg/asentric"
)

func TestPanicRule(t *testing.T) {
	ctx := NewMockContext()
	rule := NewPanicRule("test-panic")

	// Test Name()
	if rule.Name() != "test-panic" {
		t.Errorf("Expected rule name 'test-panic', got '%s'", rule.Name())
	}

	// Test that Evaluate panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but no panic occurred")
		} else {
			panicMsg := r.(string)
			if !strings.Contains(panicMsg, "intentional panic") {
				t.Errorf("Expected panic message containing 'intentional panic', got '%s'", panicMsg)
			}
			t.Logf("PanicRule correctly panicked with message: %s", panicMsg)
		}
	}()

	// This should panic
	rule.Evaluate(ctx)
}

func TestPanicRuleWithMessage(t *testing.T) {
	ctx := NewMockContext()
	customMsg := "custom panic message for testing"
	rule := NewPanicRuleWithMessage("test-panic-custom", customMsg)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with custom message")
		} else {
			panicMsg := r.(string)
			if panicMsg != customMsg {
				t.Errorf("Expected panic message '%s', got '%s'", customMsg, panicMsg)
			}
			t.Logf("Custom panic message: %s", panicMsg)
		}
	}()

	rule.Evaluate(ctx)
}

func TestErrorRule(t *testing.T) {
	ctx := NewMockContext()
	rule := NewErrorRule("test-error")

	// Test Name()
	if rule.Name() != "test-error" {
		t.Errorf("Expected rule name 'test-error', got '%s'", rule.Name())
	}

	// Test Evaluate returns error
	alert, err := rule.Evaluate(ctx)

	if alert != nil {
		t.Errorf("Expected nil alert, got: %+v", alert)
	}

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rule evaluation error") {
		t.Errorf("Expected error containing 'rule evaluation error', got '%s'", err.Error())
	}

	t.Logf("ErrorRule correctly returned error: %v", err)
}

func TestErrorRuleWithMessage(t *testing.T) {
	ctx := NewMockContext()
	customErr := "custom error message"
	rule := NewErrorRuleWithMessage("test-error-custom", customErr)

	alert, err := rule.Evaluate(ctx)

	if alert != nil {
		t.Errorf("Expected nil alert, got: %+v", alert)
	}

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != customErr {
		t.Errorf("Expected error '%s', got '%s'", customErr, err.Error())
	}

	t.Logf("Custom error message: %v", err)
}

func TestErrorRuleWithAlert(t *testing.T) {
	ctx := NewMockContext()
	rule := NewErrorRuleWithAlert("test-error-with-alert")

	// Test Name()
	if rule.Name() != "test-error-with-alert" {
		t.Errorf("Expected rule name 'test-error-with-alert', got '%s'", rule.Name())
	}

	// Test Evaluate returns BOTH alert and error
	alert, err := rule.Evaluate(ctx)

	// Should have alert
	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}

	// Should have error
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify alert fields
	if alert.Rule != "test-error-with-alert" {
		t.Errorf("Expected rule 'test-error-with-alert', got '%s'", alert.Rule)
	}
	if alert.Severity != asentric.SeverityHigh {
		t.Errorf("Expected severity HIGH, got %s", alert.Severity)
	}
	if alert.Title != "Alert with Error" {
		t.Errorf("Expected title 'Alert with Error', got '%s'", alert.Title)
	}

	// Verify error
	if !strings.Contains(err.Error(), "rule error with alert") {
		t.Errorf("Expected error containing 'rule error with alert', got '%s'", err.Error())
	}

	t.Logf("ErrorRuleWithAlert returned both alert (%s) and error (%v)", alert.Title, err)
}

func TestErrorRuleWithAlertAndMessage(t *testing.T) {
	ctx := NewMockContext()
	customErr := "specific error case"
	rule := NewErrorRuleWithAlertAndMessage("test-edge-case", customErr)

	alert, err := rule.Evaluate(ctx)

	if alert == nil {
		t.Fatal("Expected alert, got nil")
	}
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != customErr {
		t.Errorf("Expected error '%s', got '%s'", customErr, err.Error())
	}

	t.Logf("Edge case: Alert=%s, Error=%v", alert.Title, err)
}

// TestInvalidRulesImplementInterface verifies all invalid rules implement asentric.Rule
func TestInvalidRulesImplementInterface(t *testing.T) {
	rules := []asentric.Rule{
		NewErrorRule("error-1"),
		NewErrorRuleWithAlert("error-2"),
		// PanicRule is also a valid interface implementation, but can't call Evaluate safely
	}

	for i, rule := range rules {
		if rule.Name() == "" {
			t.Errorf("Rule %d has empty name", i)
		}
	}

	t.Logf("Successfully verified %d invalid rules implement asentric.Rule interface", len(rules))
}

// TestPanicRecoveryPattern demonstrates how to safely call rules that might panic
func TestPanicRecoveryPattern(t *testing.T) {
	ctx := NewMockContext()

	rules := []asentric.Rule{
		NewAlwaysAlertRule("safe-rule"),
		NewPanicRule("panic-rule"),
		NewErrorRule("error-rule"),
	}

	var alerts []*asentric.Alert
	var errors []error

	for _, rule := range rules {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Recovered from panic
					err := asentric.Alert{
						Rule:        rule.Name(),
						Title:       "Panic Recovered",
						Description: "Rule panicked during evaluation",
						Severity:    asentric.SeverityCritical,
						Metadata: map[string]any{
							"panic": r,
						},
					}
					t.Logf("Recovered from panic in rule %s: %v", rule.Name(), r)
					alerts = append(alerts, &err)
				}
			}()

			alert, err := rule.Evaluate(ctx)
			if err != nil {
				errors = append(errors, err)
				t.Logf("Rule %s returned error: %v", rule.Name(), err)
			}
			if alert != nil {
				alerts = append(alerts, alert)
			}
		}()
	}

	t.Logf("Processed %d rules: %d alerts, %d errors", len(rules), len(alerts), len(errors))

	// Should have 2 alerts (1 from safe-rule, 1 from panic recovery)
	// Should have 1 error (from error-rule)
	if len(alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(alerts))
	}
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
}
