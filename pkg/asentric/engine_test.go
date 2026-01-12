package asentric_test

import (
	"errors"
	"testing"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// ========================================
// Mock Context
// ========================================

type mockContext struct {
	chainID domain.ChainID
	tx      domain.Transaction
	block   domain.Block
	logs    []domain.Log
	abi     domain.ABIRegistry
}

func (m *mockContext) ChainID() domain.ChainID { return m.chainID }
func (m *mockContext) Tx() domain.Transaction  { return m.tx }
func (m *mockContext) Block() domain.Block     { return m.block }
func (m *mockContext) Logs() []domain.Log      { return m.logs }
func (m *mockContext) ABI() domain.ABIRegistry { return m.abi }

func newMockContext() *mockContext {
	return &mockContext{
		chainID: 1,
		tx:      domain.Transaction{},
		block:   domain.Block{},
		logs:    []domain.Log{},
		abi:     nil,
	}
}

// ========================================
// Mock Rules - Format Valid
// ========================================

// alwaysAlertRule selalu return alert dengan nil error (format valid)
type alwaysAlertRule struct {
	name string
}

func (r *alwaysAlertRule) Name() string { return r.name }
func (r *alwaysAlertRule) Severity() asentric.Severity { return asentric.SeverityInfo }
func (r *alwaysAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:        r.name,
		Severity:    asentric.SeverityInfo,
		Title:       "Test Alert",
		Description: "This is a test alert from alwaysAlertRule",
		Metadata:    map[string]any{"test": true},
	}, nil
}

// alwaysAlertRuleCritical return alert dengan severity CRITICAL (format valid)
type alwaysAlertRuleCritical struct {
	name string
}

func (r *alwaysAlertRuleCritical) Name() string { return r.name }
func (r *alwaysAlertRuleCritical) Severity() asentric.Severity { return asentric.SeverityCritical }
func (r *alwaysAlertRuleCritical) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:        r.name,
		Severity:    asentric.SeverityCritical,
		Title:       "Critical Alert",
		Description: "This is a critical alert",
		Metadata:    map[string]any{"severity": "critical"},
	}, nil
}

// alwaysAlertRuleHigh return alert dengan severity HIGH (format valid)
type alwaysAlertRuleHigh struct {
	name string
}

func (r *alwaysAlertRuleHigh) Name() string { return r.name }
func (r *alwaysAlertRuleHigh) Severity() asentric.Severity { return asentric.SeverityHigh }
func (r *alwaysAlertRuleHigh) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:        r.name,
		Severity:    asentric.SeverityHigh,
		Title:       "High Alert",
		Description: "This is a high severity alert",
		Metadata:    map[string]any{"severity": "high"},
	}, nil
}

// alwaysAlertRuleMedium return alert dengan severity MEDIUM (format valid)
type alwaysAlertRuleMedium struct {
	name string
}

func (r *alwaysAlertRuleMedium) Name() string { return r.name }
func (r *alwaysAlertRuleMedium) Severity() asentric.Severity { return asentric.SeverityMedium }
func (r *alwaysAlertRuleMedium) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:        r.name,
		Severity:    asentric.SeverityMedium,
		Title:       "Medium Alert",
		Description: "This is a medium severity alert",
		Metadata:    map[string]any{"severity": "medium"},
	}, nil
}

// alwaysAlertRuleLow return alert dengan severity LOW (format valid)
type alwaysAlertRuleLow struct {
	name string
}

func (r *alwaysAlertRuleLow) Name() string { return r.name }
func (r *alwaysAlertRuleLow) Severity() asentric.Severity { return asentric.SeverityLow }
func (r *alwaysAlertRuleLow) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:        r.name,
		Severity:    asentric.SeverityLow,
		Title:       "Low Alert",
		Description: "This is a low severity alert",
		Metadata:    map[string]any{"severity": "low"},
	}, nil
}

// neverAlertRule tidak pernah return alert, hanya nil (format valid)
type neverAlertRule struct {
	name string
}

func (r *neverAlertRule) Name() string { return r.name }
func (r *neverAlertRule) Severity() asentric.Severity { return asentric.SeverityInfo }
func (r *neverAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return nil, nil
}

// conditionalAlertRule return alert berdasarkan kondisi context (format valid)
type conditionalAlertRule struct {
	name string
}

func (r *conditionalAlertRule) Name() string { return r.name }
func (r *conditionalAlertRule) Severity() asentric.Severity { return asentric.SeverityMedium }
func (r *conditionalAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	if ctx == nil {
		return nil, nil
	}
	// Contoh: return alert jika chainID > 1000
	if ctx.ChainID() > 1000 {
		return &asentric.Alert{
			Rule:        r.name,
			Severity:    asentric.SeverityInfo,
			Title:       "Conditional Alert",
			Description: "Chain ID is greater than 1000",
			Metadata:    map[string]any{"chainID": ctx.ChainID()},
		}, nil
	}
	return nil, nil
}

// ========================================
// Mock Rules - Format Tidak Valid
// ========================================

// panicRule panic saat evaluation (format tidak valid)
type panicRule struct {
	name string
}

func (r *panicRule) Name() string { return r.name }
func (r *panicRule) Severity() asentric.Severity { return asentric.SeverityCritical }
func (r *panicRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	panic("intentional panic for testing")
}

// errorRule return error (format tidak valid)
type errorRule struct {
	name string
	err  error
}

func (r *errorRule) Name() string { return r.name }
func (r *errorRule) Severity() asentric.Severity { return asentric.SeverityHigh }
func (r *errorRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	if r.err != nil {
		return nil, r.err
	}
	return nil, errors.New("rule evaluation error")
}

// errorRuleWithAlert return alert dan error bersamaan (format tidak valid - edge case)
type errorRuleWithAlert struct {
	name string
}

func (r *errorRuleWithAlert) Name() string { return r.name }
func (r *errorRuleWithAlert) Severity() asentric.Severity { return asentric.SeverityHigh }
func (r *errorRuleWithAlert) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return &asentric.Alert{
		Rule:     r.name,
		Severity: asentric.SeverityInfo,
		Title:    "Alert with Error",
	}, errors.New("error occurred but alert was generated")
}

// ========================================
// Tests
// ========================================

func TestNewEngine(t *testing.T) {
	engine := asentric.NewEngine()
	if engine == nil {
		t.Error("expected engine to be not nil")
	}
}

func TestEngine_RegisterRule(t *testing.T) {
	t.Run("register valid rule", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &alwaysAlertRule{name: "test_rule"}
		err := engine.RegisterRule(rule)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("register nil rule returns error", func(t *testing.T) {
		engine := asentric.NewEngine()
		err := engine.RegisterRule(nil)
		if err == nil {
			t.Error("expected error for nil rule")
		}
		if err != asentric.ErrInvalidRule {
			t.Errorf("expected ErrInvalidRule, got %v", err)
		}
	})

	t.Run("register multiple rules", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "rule1"}
		rule2 := &alwaysAlertRule{name: "rule2"}
		rule3 := &neverAlertRule{name: "rule3"}

		err1 := engine.RegisterRule(rule1)
		err2 := engine.RegisterRule(rule2)
		err3 := engine.RegisterRule(rule3)

		if err1 != nil {
			t.Errorf("expected no error for rule1, got %v", err1)
		}
		if err2 != nil {
			t.Errorf("expected no error for rule2, got %v", err2)
		}
		if err3 != nil {
			t.Errorf("expected no error for rule3, got %v", err3)
		}
	})
}

func TestEngine_Evaluate_FormatValid(t *testing.T) {
	t.Run("evaluate with nil context returns error", func(t *testing.T) {
		engine := asentric.NewEngine()
		alerts, err := engine.Evaluate(nil)
		if err == nil {
			t.Error("expected error for nil context")
		}
		if err != asentric.ErrInvalidContext {
			t.Errorf("expected ErrInvalidContext, got %v", err)
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
	})

	t.Run("evaluate with empty rules returns empty alerts", func(t *testing.T) {
		engine := asentric.NewEngine()
		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 0 {
			t.Errorf("expected empty alerts, got %d alerts", len(alerts))
		}
	})

	t.Run("evaluate with alwaysAlertRule returns alert", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &alwaysAlertRule{name: "test_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Rule != "test_rule" {
			t.Errorf("expected rule name 'test_rule', got %q", alerts[0].Rule)
		}
		if alerts[0].Severity != asentric.SeverityInfo {
			t.Errorf("expected severity Info, got %q", alerts[0].Severity)
		}
		if alerts[0].Title != "Test Alert" {
			t.Errorf("expected title 'Test Alert', got %q", alerts[0].Title)
		}
	})

	t.Run("evaluate with neverAlertRule returns no alert", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &neverAlertRule{name: "no_alert_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 0 {
			t.Errorf("expected no alerts, got %d alerts", len(alerts))
		}
	})

	t.Run("evaluate with multiple rules returns multiple alerts", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "rule1"}
		rule2 := &alwaysAlertRule{name: "rule2"}
		rule3 := &alwaysAlertRule{name: "rule3"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 3 {
			t.Fatalf("expected 3 alerts, got %d", len(alerts))
		}
		if alerts[0].Rule != "rule1" {
			t.Errorf("expected first alert rule 'rule1', got %q", alerts[0].Rule)
		}
		if alerts[1].Rule != "rule2" {
			t.Errorf("expected second alert rule 'rule2', got %q", alerts[1].Rule)
		}
		if alerts[2].Rule != "rule3" {
			t.Errorf("expected third alert rule 'rule3', got %q", alerts[2].Rule)
		}
	})

	t.Run("evaluate with mixed rules (alert and no alert)", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "alert_rule"}
		rule2 := &neverAlertRule{name: "no_alert_rule"}
		rule3 := &alwaysAlertRule{name: "alert_rule2"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 2 {
			t.Fatalf("expected 2 alerts, got %d", len(alerts))
		}
		if alerts[0].Rule != "alert_rule" {
			t.Errorf("expected first alert rule 'alert_rule', got %q", alerts[0].Rule)
		}
		if alerts[1].Rule != "alert_rule2" {
			t.Errorf("expected second alert rule 'alert_rule2', got %q", alerts[1].Rule)
		}
	})

	t.Run("evaluate with different severity levels", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRuleCritical{name: "critical_rule"}
		rule2 := &alwaysAlertRuleHigh{name: "high_rule"}
		rule3 := &alwaysAlertRuleMedium{name: "medium_rule"}
		rule4 := &alwaysAlertRuleLow{name: "low_rule"}
		rule5 := &alwaysAlertRule{name: "info_rule"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)
		_ = engine.RegisterRule(rule4)
		_ = engine.RegisterRule(rule5)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 5 {
			t.Fatalf("expected 5 alerts, got %d", len(alerts))
		}
		if alerts[0].Severity != asentric.SeverityCritical {
			t.Errorf("expected first alert severity Critical, got %q", alerts[0].Severity)
		}
		if alerts[1].Severity != asentric.SeverityHigh {
			t.Errorf("expected second alert severity High, got %q", alerts[1].Severity)
		}
		if alerts[2].Severity != asentric.SeverityMedium {
			t.Errorf("expected third alert severity Medium, got %q", alerts[2].Severity)
		}
		if alerts[3].Severity != asentric.SeverityLow {
			t.Errorf("expected fourth alert severity Low, got %q", alerts[3].Severity)
		}
		if alerts[4].Severity != asentric.SeverityInfo {
			t.Errorf("expected fifth alert severity Info, got %q", alerts[4].Severity)
		}
	})

	t.Run("evaluate with conditional rule", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &conditionalAlertRule{name: "conditional_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		ctx.chainID = 2000 // > 1000, should trigger alert
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Rule != "conditional_rule" {
			t.Errorf("expected rule name 'conditional_rule', got %q", alerts[0].Rule)
		}
	})

	t.Run("evaluate preserves rule execution order", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "first"}
		rule2 := &alwaysAlertRule{name: "second"}
		rule3 := &alwaysAlertRule{name: "third"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(alerts) != 3 {
			t.Fatalf("expected 3 alerts, got %d", len(alerts))
		}
		// Verify order matches registration order
		if alerts[0].Rule != "first" {
			t.Errorf("expected first alert rule 'first', got %q", alerts[0].Rule)
		}
		if alerts[1].Rule != "second" {
			t.Errorf("expected second alert rule 'second', got %q", alerts[1].Rule)
		}
		if alerts[2].Rule != "third" {
			t.Errorf("expected third alert rule 'third', got %q", alerts[2].Rule)
		}
	})
}

func TestEngine_Evaluate_FormatTidakValid(t *testing.T) {
	t.Run("evaluate with panic rule recovers and returns error", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &panicRule{name: "panic_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err == nil {
			t.Error("expected error for panic rule")
		}
		if err != asentric.ErrRulePanic {
			t.Errorf("expected ErrRulePanic, got %v", err)
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
	})

	t.Run("evaluate with error rule returns error", func(t *testing.T) {
		engine := asentric.NewEngine()
		customErr := errors.New("custom rule error")
		rule := &errorRule{name: "error_rule", err: customErr}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err == nil {
			t.Error("expected error for error rule")
		}
		if err != customErr {
			t.Errorf("expected custom error, got %v", err)
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
	})

	t.Run("evaluate with error rule (default error)", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &errorRule{name: "error_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err == nil {
			t.Error("expected error for error rule")
		}
		if err.Error() != "rule evaluation error" {
			t.Errorf("expected 'rule evaluation error', got %q", err.Error())
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
	})

	t.Run("evaluate stops on first error", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "rule1"}
		rule2 := &errorRule{name: "error_rule"}
		rule3 := &alwaysAlertRule{name: "rule3"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err == nil {
			t.Error("expected error")
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
		// rule3 should not be executed because rule2 returned error
	})

	t.Run("evaluate stops on first panic", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule1 := &alwaysAlertRule{name: "rule1"}
		rule2 := &panicRule{name: "panic_rule"}
		rule3 := &alwaysAlertRule{name: "rule3"}
		_ = engine.RegisterRule(rule1)
		_ = engine.RegisterRule(rule2)
		_ = engine.RegisterRule(rule3)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		if err == nil {
			t.Error("expected error for panic rule")
		}
		if err != asentric.ErrRulePanic {
			t.Errorf("expected ErrRulePanic, got %v", err)
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
		// rule3 should not be executed because rule2 panicked
	})

	t.Run("evaluate with error rule that also returns alert", func(t *testing.T) {
		engine := asentric.NewEngine()
		rule := &errorRuleWithAlert{name: "error_with_alert_rule"}
		_ = engine.RegisterRule(rule)

		ctx := newMockContext()
		alerts, err := engine.Evaluate(ctx)

		// Engine should return error, alert is ignored when error occurs
		if err == nil {
			t.Error("expected error")
		}
		if err.Error() != "error occurred but alert was generated" {
			t.Errorf("expected 'error occurred but alert was generated', got %q", err.Error())
		}
		if alerts != nil {
			t.Errorf("expected nil alerts, got %v", alerts)
		}
	})
}
