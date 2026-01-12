package testutils

import (
	"errors"

	"github.com/asentric/asentric/pkg/asentric"
)

type PanicRule struct {
	ruleID   string
	panicMsg string
}

func NewPanicRule(ruleID string) *PanicRule {
	return &PanicRule{
		ruleID:   ruleID,
		panicMsg: "intentional panic for testing",
	}
}

func NewPanicRuleWithMessage(ruleID, panicMsg string) *PanicRule {
	return &PanicRule{
		ruleID:   ruleID,
		panicMsg: panicMsg,
	}
}

func (r *PanicRule) Name() string {
	return r.ruleID
}

func (r *PanicRule) Severity() asentric.Severity {
	return asentric.SeverityCritical
}

func (r *PanicRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	// Intentionally panic to test recovery
	panic(r.panicMsg)
}

type ErrorRule struct {
	ruleID   string
	errorMsg string
}

func NewErrorRule(ruleID string) *ErrorRule {
	return &ErrorRule{
		ruleID:   ruleID,
		errorMsg: "rule evaluation error",
	}
}

func NewErrorRuleWithMessage(ruleID, errorMsg string) *ErrorRule {
	return &ErrorRule{
		ruleID:   ruleID,
		errorMsg: errorMsg,
	}
}

func (r *ErrorRule) Name() string {
	return r.ruleID
}

func (r *ErrorRule) Severity() asentric.Severity {
	return asentric.SeverityHigh
}

func (r *ErrorRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	// Return nil alert with error
	return nil, errors.New(r.errorMsg)
}

// -------------------------------------------
// ErrorRuleWithAlert - Returns BOTH alert AND error
// Use case: Test edge case where rule returns alert + error
// This is technically invalid behavior but should be handled gracefully
// -------------------------------------------

type ErrorRuleWithAlert struct {
	ruleID   string
	errorMsg string
}

func NewErrorRuleWithAlert(ruleID string) *ErrorRuleWithAlert {
	return &ErrorRuleWithAlert{
		ruleID:   ruleID,
		errorMsg: "rule error with alert",
	}
}

func NewErrorRuleWithAlertAndMessage(ruleID, errorMsg string) *ErrorRuleWithAlert {
	return &ErrorRuleWithAlert{
		ruleID:   ruleID,
		errorMsg: errorMsg,
	}
}

func (r *ErrorRuleWithAlert) Name() string {
	return r.ruleID
}

func (r *ErrorRuleWithAlert) Severity() asentric.Severity {
	return asentric.SeverityHigh
}

func (r *ErrorRuleWithAlert) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	// Return BOTH alert and error (edge case)
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "Alert with Error",
		Description: "This alert was generated alongside an error",
		Severity:    asentric.SeverityHigh,
		Metadata:    make(map[string]any),
	}

	return alert, errors.New(r.errorMsg)
}
