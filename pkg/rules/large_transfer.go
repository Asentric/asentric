// Package rules provides reusable built-in detection rules for the Asentric SDK.
package rules

import (
	"math/big"

	"github.com/asentric/asentric/pkg/asentric"
)

// LargeTransferRule detects ERC20 and native token transfers above a threshold.
type LargeTransferRule struct {
	Threshold *big.Int
}

// NewLargeTransferRule creates a rule with 1000 token threshold (18 decimals).
func NewLargeTransferRule() *LargeTransferRule {
	threshold := new(big.Int)
	threshold.SetString("1000000000000000000000", 10) // 1000 * 10^18
	return &LargeTransferRule{Threshold: threshold}
}

// NewLargeTransferRuleWithThreshold creates a rule with custom threshold.
func NewLargeTransferRuleWithThreshold(threshold *big.Int) *LargeTransferRule {
	return &LargeTransferRule{Threshold: threshold}
}

func (r *LargeTransferRule) Name() string {
	return "large-transfer"
}

func (r *LargeTransferRule) Severity() asentric.Severity {
	return asentric.SeverityHigh
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	// Check ERC20 Transfer events in logs
	for _, log := range ctx.Logs() {
		if log.Event.Name == "Transfer" {
			valueField, ok := log.Event.Fields["value"]
			if !ok {
				continue
			}

			value, ok := valueField.(*big.Int)
			if !ok {
				continue
			}

			if value.Cmp(r.Threshold) > 0 {
				return asentric.NewAlert(
					r.Name(),
					"Large ERC20 Transfer Detected",
					r.Severity(),
				).
					WithDescription("ERC20 transfer value exceeds threshold").
					WithMetadata("value", value.String()).
					WithMetadata("contract", log.Address.Hex()).
					WithRef(asentric.NewExecutionRef(
						ctx.Tx().Hash.Hex(),
						ctx.Block().Number,
						int(log.LogIndex),
					)), nil
			}
		}
	}

	// Check native token transfer
	tx := ctx.Tx()
	if tx.RawValue.Wei != "" {
		value := new(big.Int)
		value.SetString(tx.RawValue.Wei, 10)

		if value.Cmp(r.Threshold) > 0 {
			return asentric.NewAlert(
				r.Name(),
				"Large Native Transfer Detected",
				r.Severity(),
			).
				WithDescription("Native token transfer exceeds threshold").
				WithMetadata("value_wei", tx.RawValue.Wei).
				WithMetadata("from", tx.From.Hex()).
				WithMetadata("to", tx.To.Hex()), nil
		}
	}

	return nil, nil
}
