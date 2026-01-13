// Package rules provides reusable built-in detection rules for the Asentric SDK.
package rules

import (
	"math/big"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/utils"
)

// WhaleAlertRule detects very large native token transfers.
type WhaleAlertRule struct {
	Threshold *big.Int
}

// NewWhaleAlertRule creates a rule for whale alerts (default: 10 tokens).
func NewWhaleAlertRule() *WhaleAlertRule {
	threshold := new(big.Int)
	threshold.SetString("10000000000000000000", 10) // 10 * 10^18
	return &WhaleAlertRule{Threshold: threshold}
}

// NewWhaleAlertRuleWithThreshold creates a rule with custom threshold.
func NewWhaleAlertRuleWithThreshold(threshold *big.Int) *WhaleAlertRule {
	return &WhaleAlertRule{Threshold: threshold}
}

func (r *WhaleAlertRule) Name() string {
	return "whale-alert"
}

func (r *WhaleAlertRule) Severity() asentric.Severity {
	return asentric.SeverityCritical
}

func (r *WhaleAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	tx := ctx.Tx()

	if tx.RawValue.Wei == "" {
		return nil, nil
	}

	value := new(big.Int)
	value.SetString(tx.RawValue.Wei, 10)

	if value.Cmp(r.Threshold) > 0 {
		return asentric.NewAlert(
			r.Name(),
			"🐋 Whale Alert!",
			r.Severity(),
		).
			WithDescription("Large native token transfer detected").
			WithMetadata("value", utils.FormatETH(value)).
			WithMetadata("from", tx.From.Hex()).
			WithMetadata("to", tx.To.Hex()), nil
	}

	return nil, nil
}
