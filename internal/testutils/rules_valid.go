package testutils

import (
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

type AlwaysAlertRule struct {
	ruleID string
}

func NewAlwaysAlertRule(ruleID string) *AlwaysAlertRule {
	return &AlwaysAlertRule{
		ruleID: ruleID,
	}
}

func (r *AlwaysAlertRule) Name() string {
	return r.ruleID
}

func (r *AlwaysAlertRule) Severity() asentric.Severity {
	return asentric.SeverityInfo
}

func (r *AlwaysAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "Test Alert",
		Description: "This is a test alert with INFO severity",
		Severity:    asentric.SeverityInfo,
		Metadata:    make(map[string]any),
	}
	return alert, nil
}

type AlwaysAlertRuleCritical struct {
	ruleID string
}

func NewAlwaysAlertRuleCritical(ruleID string) *AlwaysAlertRuleCritical {
	return &AlwaysAlertRuleCritical{
		ruleID: ruleID,
	}
}

func (r *AlwaysAlertRuleCritical) Name() string {
	return r.ruleID
}

func (r *AlwaysAlertRuleCritical) Severity() asentric.Severity {
	return asentric.SeverityCritical
}

func (r *AlwaysAlertRuleCritical) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "Critical Alert",
		Description: "This is a critical test alert",
		Severity:    asentric.SeverityCritical,
		Metadata:    make(map[string]any),
	}
	return alert, nil
}

type AlwaysAlertRuleHigh struct {
	ruleID string
}

func NewAlwaysAlertRuleHigh(ruleID string) *AlwaysAlertRuleHigh {
	return &AlwaysAlertRuleHigh{
		ruleID: ruleID,
	}
}

func (r *AlwaysAlertRuleHigh) Name() string {
	return r.ruleID
}

func (r *AlwaysAlertRuleHigh) Severity() asentric.Severity {
	return asentric.SeverityHigh
}

func (r *AlwaysAlertRuleHigh) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "High Priority Alert",
		Description: "This is a high priority test alert",
		Severity:    asentric.SeverityHigh,
		Metadata:    make(map[string]any),
	}
	return alert, nil
}

// -------------------------------------------
// AlwaysAlertRuleMedium - Returns alert with MEDIUM severity
// -------------------------------------------

type AlwaysAlertRuleMedium struct {
	ruleID string
}

func NewAlwaysAlertRuleMedium(ruleID string) *AlwaysAlertRuleMedium {
	return &AlwaysAlertRuleMedium{
		ruleID: ruleID,
	}
}

func (r *AlwaysAlertRuleMedium) Name() string {
	return r.ruleID
}

func (r *AlwaysAlertRuleMedium) Severity() asentric.Severity {
	return asentric.SeverityMedium
}

func (r *AlwaysAlertRuleMedium) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "Medium Priority Alert",
		Description: "This is a medium priority test alert",
		Severity:    asentric.SeverityMedium,
		Metadata:    make(map[string]any),
	}
	return alert, nil
}

// -------------------------------------------
// AlwaysAlertRuleLow - Returns alert with LOW severity
// -------------------------------------------

type AlwaysAlertRuleLow struct {
	ruleID string
}

func NewAlwaysAlertRuleLow(ruleID string) *AlwaysAlertRuleLow {
	return &AlwaysAlertRuleLow{
		ruleID: ruleID,
	}
}

func (r *AlwaysAlertRuleLow) Name() string {
	return r.ruleID
}

func (r *AlwaysAlertRuleLow) Severity() asentric.Severity {
	return asentric.SeverityLow
}

func (r *AlwaysAlertRuleLow) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	alert := &asentric.Alert{
		Rule:        r.ruleID,
		Title:       "Low Priority Alert",
		Description: "This is a low priority test alert",
		Severity:    asentric.SeverityLow,
		Metadata:    make(map[string]any),
	}
	return alert, nil
}

// -------------------------------------------
// NeverAlertRule - Returns nil (no alert)
// -------------------------------------------

type NeverAlertRule struct {
	ruleID string
}

func NewNeverAlertRule(ruleID string) *NeverAlertRule {
	return &NeverAlertRule{
		ruleID: ruleID,
	}
}

func (r *NeverAlertRule) Name() string {
	return r.ruleID
}

func (r *NeverAlertRule) Severity() asentric.Severity {
	return asentric.SeverityInfo
}

func (r *NeverAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return nil, nil
}

// -------------------------------------------
// ConditionalAlertRule - Returns alert only if chainID > 1000
// -------------------------------------------

type ConditionalAlertRule struct {
	ruleID string
}

func NewConditionalAlertRule(ruleID string) *ConditionalAlertRule {
	return &ConditionalAlertRule{
		ruleID: ruleID,
	}
}

func (r *ConditionalAlertRule) Name() string {
	return r.ruleID
}

func (r *ConditionalAlertRule) Severity() asentric.Severity {
	return asentric.SeverityMedium
}

func (r *ConditionalAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	chainID := ctx.ChainID()

	// Return alert only if chainID > 1000
	if chainID > domain.ChainID(1000) {
		alert := &asentric.Alert{
			Rule:        r.ruleID,
			Title:       "Conditional Alert",
			Description: "This alert is triggered when chainID > 1000",
			Severity:    asentric.SeverityMedium,
			Metadata: map[string]any{
				"chainID": uint64(chainID),
			},
		}
		return alert, nil
	}

	return nil, nil
}
