// Package rules provides reusable built-in detection rules for the Asentric SDK.
//
// Available Rules:
//
//   - LargeTransferRule: Detects large ERC20 and native token transfers
//   - WhaleAlertRule: Detects very large native token transfers
//   - ERC20TransferRule: Monitors specific ERC20 token transfers (including mint/burn)
//
// Usage:
//
//	engine := asentric.NewEngine()
//
//	// Register built-in rules
//	engine.RegisterRule(rules.NewLargeTransferRule())
//	engine.RegisterRule(rules.NewWhaleAlertRule())
//
//	// Register ERC20-specific rule
//	engine.RegisterRule(rules.NewERC20TransferRule(rules.ERC20TransferConfig{
//	    ContractAddress: "0x...",
//	    TokenSymbol:     "USDC",
//	    Decimals:        6,
//	}))
package rules
