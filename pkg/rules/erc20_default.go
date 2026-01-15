// Package rules provides reusable built-in detection rules for the Asentric SDK.
package rules

import (
	"fmt"
	logPkg "log"
	"math/big"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/utils"
)

// ERC20DefaultRule detects all ERC20 Transfer/Mint/Burn events with value > 100.
// This is a default rule that listens to all ERC20 contracts.
type ERC20DefaultRule struct {
	// Minimum value threshold in token units (not wei)
	// Default: 100
	minValue *big.Int
}

// ERC20DefaultConfig configures the default ERC20 detection rule.
type ERC20DefaultConfig struct {
	// MinValue is the minimum token value to trigger alert (in token units, not wei)
	// Default: 100
	MinValue *big.Int
}

// NewERC20DefaultRule creates a new default ERC20 detection rule.
// This rule listens to all ERC20 Transfer events with value > 100.
func NewERC20DefaultRule() *ERC20DefaultRule {
	// Default threshold: 100 tokens (will be converted to wei based on decimals)
	minValue := big.NewInt(100)
	return &ERC20DefaultRule{
		minValue: minValue,
	}
}

// NewERC20DefaultRuleWithConfig creates a rule with custom configuration.
func NewERC20DefaultRuleWithConfig(cfg ERC20DefaultConfig) *ERC20DefaultRule {
	minValue := cfg.MinValue
	if minValue == nil {
		minValue = big.NewInt(100)
	}
	return &ERC20DefaultRule{
		minValue: minValue,
	}
}

func (r *ERC20DefaultRule) Name() string {
	return "erc20-default"
}

func (r *ERC20DefaultRule) Severity() asentric.Severity {
	return asentric.SeverityMedium
}

func (r *ERC20DefaultRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	for _, logEntry := range ctx.Logs() {
		// Check if this is a Transfer event
		if logEntry.Event.Name != "Transfer" {
			continue
		}

		// Extract Transfer event fields
		from := utils.GetFieldString(logEntry.Event.Fields, "from")
		to := utils.GetFieldString(logEntry.Event.Fields, "to")
		value := utils.GetFieldBigInt(logEntry.Event.Fields, "value")

		if value == nil || value.Sign() == 0 {
			continue
		}

		// Check if value is > 100 (in token units)
		// We assume 18 decimals by default for demo purposes
		// In production, you would query the contract for decimals
		decimals := 18
		thresholdWei := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		thresholdWei.Mul(thresholdWei, r.minValue)

		if value.Cmp(thresholdWei) <= 0 {
			continue
		}

		// Determine event type
		isMint := utils.IsZeroAddress(from)
		isBurn := utils.IsZeroAddress(to)

		// Format value for display
		valueStr := utils.FormatTokenAmount(value, decimals)

		// Print to console for demo
		eventType := "Transfer"
		if isMint {
			eventType = "MINT"
		} else if isBurn {
			eventType = "BURN"
		}

		logPkg.Printf("[ERC20 Default] %s detected:", eventType)
		logPkg.Printf("  Contract: %s", logEntry.Address.Hex())
		logPkg.Printf("  From: %s", utils.TruncateAddress(from))
		logPkg.Printf("  To: %s", utils.TruncateAddress(to))
		logPkg.Printf("  Value: %s", valueStr)
		logPkg.Printf("  Block: %d", logEntry.BlockNumber)
		logPkg.Printf("  TxHash: %s", logEntry.TxHash.Hex())
		logPkg.Println("")

		// Create alert
		var title, desc string
		severity := r.Severity()

		switch {
		case isMint:
			title = fmt.Sprintf("🎉 ERC20 MINT: %s", valueStr)
			desc = fmt.Sprintf("New tokens minted to %s", utils.TruncateAddress(to))
			severity = asentric.SeverityHigh
		case isBurn:
			title = fmt.Sprintf("🔥 ERC20 BURN: %s", valueStr)
			desc = fmt.Sprintf("Tokens burned from %s", utils.TruncateAddress(from))
			severity = asentric.SeverityHigh
		default:
			title = fmt.Sprintf("ERC20 Transfer: %s", valueStr)
			desc = fmt.Sprintf("Transfer from %s to %s", utils.TruncateAddress(from), utils.TruncateAddress(to))
		}

		return asentric.NewAlert(r.Name(), title, severity).
			WithDescription(desc).
			WithMetadata("from", from).
			WithMetadata("to", to).
			WithMetadata("value", valueStr).
			WithMetadata("value_wei", value.String()).
			WithMetadata("isMint", isMint).
			WithMetadata("isBurn", isBurn).
			WithMetadata("contract", logEntry.Address.Hex()).
			WithRef(asentric.NewExecutionRef(
				logEntry.TxHash.Hex(),
				logEntry.BlockNumber,
				int(logEntry.LogIndex),
			)), nil
	}

	return nil, nil
}
