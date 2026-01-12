package testutils

import (
	"math/big"

	"github.com/asentric/asentric/pkg/asentric"
)

// -------------------------------------------
// FlexibleAlertRule - Returns alert with configurable severity
// -------------------------------------------

// FlexibleAlertRule always returns an alert with the configured severity.
type FlexibleAlertRule struct {
	ruleID   string
	severity asentric.Severity
}

// NewFlexibleAlertRule creates a rule that always alerts with the given severity.
// This is the preferred way to create test rules when you need to specify severity.
func NewFlexibleAlertRule(ruleID string, severity asentric.Severity) *FlexibleAlertRule {
	return &FlexibleAlertRule{
		ruleID:   ruleID,
		severity: severity,
	}
}

func (r *FlexibleAlertRule) Name() string {
	return r.ruleID
}

func (r *FlexibleAlertRule) Severity() asentric.Severity {
	return r.severity
}

func (r *FlexibleAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	return asentric.NewAlert(r.ruleID, "Test Alert from "+r.ruleID, r.severity), nil
}

// -------------------------------------------
// LargeTransferRule - Alerts on large value transfers
// -------------------------------------------

// LargeTransferRule triggers alert when transaction value exceeds threshold.
type LargeTransferRule struct {
	ruleID    string
	severity  asentric.Severity
	threshold *big.Int
}

// NewLargeTransferRule creates a rule that alerts on transfers > 1 ETH.
func NewLargeTransferRule(ruleID string, severity asentric.Severity) *LargeTransferRule {
	// Default threshold: 1 ETH
	threshold := new(big.Int)
	threshold.SetString("1000000000000000000", 10) // 1e18 wei

	return &LargeTransferRule{
		ruleID:    ruleID,
		severity:  severity,
		threshold: threshold,
	}
}

// NewLargeTransferRuleWithThreshold creates a rule with custom threshold.
func NewLargeTransferRuleWithThreshold(ruleID string, severity asentric.Severity, threshold *big.Int) *LargeTransferRule {
	return &LargeTransferRule{
		ruleID:    ruleID,
		severity:  severity,
		threshold: threshold,
	}
}

func (r *LargeTransferRule) Name() string {
	return r.ruleID
}

func (r *LargeTransferRule) Severity() asentric.Severity {
	return r.severity
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	tx := ctx.Tx()
	value := tx.Value()

	// Check if value exceeds threshold
	if value != nil && value.Cmp(r.threshold) > 0 {
		return asentric.NewAlert(r.ruleID, "Large Transfer Detected", r.severity).
			WithDescription("Transaction value exceeds threshold").
			WithMetadata("value", value.String()).
			WithMetadata("threshold", r.threshold.String()), nil
	}

	return nil, nil
}
