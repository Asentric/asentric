// Package rules provides reusable built-in detection rules for the Asentric SDK.
package rules

import (
	"fmt"
	"strings"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/utils"
)

// ERC20 Transfer event topic hash: keccak256("Transfer(address,address,uint256)")
const TransferEventTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// ERC20TransferRule detects Transfer events from a specific ERC20 contract.
type ERC20TransferRule struct {
	contractAddress string
	tokenSymbol     string
	decimals        int
}

// ERC20TransferConfig configures the ERC20 transfer detection rule.
type ERC20TransferConfig struct {
	ContractAddress string
	TokenSymbol     string
	Decimals        int
}

// NewERC20TransferRule creates a new ERC20 transfer detection rule.
func NewERC20TransferRule(cfg ERC20TransferConfig) *ERC20TransferRule {
	decimals := cfg.Decimals
	if decimals == 0 {
		decimals = 18 // default to 18 decimals
	}
	return &ERC20TransferRule{
		contractAddress: utils.NormalizeAddress(cfg.ContractAddress),
		tokenSymbol:     cfg.TokenSymbol,
		decimals:        decimals,
	}
}

func (r *ERC20TransferRule) Name() string {
	return fmt.Sprintf("%s-transfer", strings.ToLower(r.tokenSymbol))
}

func (r *ERC20TransferRule) Severity() asentric.Severity {
	return asentric.SeverityMedium
}

func (r *ERC20TransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
	for _, log := range ctx.Logs() {
		// Check if log is from target contract
		if utils.NormalizeAddress(log.Address.Hex()) != r.contractAddress {
			continue
		}

		// Check if this is a Transfer event
		if log.Event.Name != "Transfer" {
			continue
		}

		from := utils.GetFieldString(log.Event.Fields, "from")
		to := utils.GetFieldString(log.Event.Fields, "to")
		value := utils.GetFieldBigInt(log.Event.Fields, "value")
		valueStr := utils.FormatTokenAmount(value, r.decimals)

		isMint := utils.IsZeroAddress(from)
		isBurn := utils.IsZeroAddress(to)

		var title, desc string
		severity := r.Severity()

		switch {
		case isMint:
			title = fmt.Sprintf("🎉 %s MINT: %s", r.tokenSymbol, valueStr)
			desc = fmt.Sprintf("New %s tokens minted to %s", r.tokenSymbol, utils.TruncateAddress(to))
			severity = asentric.SeverityHigh
		case isBurn:
			title = fmt.Sprintf("🔥 %s BURN: %s", r.tokenSymbol, valueStr)
			desc = fmt.Sprintf("%s tokens burned from %s", r.tokenSymbol, utils.TruncateAddress(from))
			severity = asentric.SeverityHigh
		default:
			title = fmt.Sprintf("%s Transfer: %s", r.tokenSymbol, valueStr)
			desc = fmt.Sprintf("Transfer from %s to %s", utils.TruncateAddress(from), utils.TruncateAddress(to))
		}

		return asentric.NewAlert(r.Name(), title, severity).
			WithDescription(desc).
			WithMetadata("from", from).
			WithMetadata("to", to).
			WithMetadata("value", valueStr).
			WithMetadata("isMint", isMint).
			WithMetadata("isBurn", isBurn).
			WithMetadata("contract", r.contractAddress), nil
	}

	return nil, nil
}

// ContractAddress returns the monitored contract address.
func (r *ERC20TransferRule) ContractAddress() string {
	return r.contractAddress
}
