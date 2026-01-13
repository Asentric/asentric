// Package utils provides common utility functions for the Asentric SDK.
package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// FormatTokenAmount converts wei to human-readable token amount.
// decimals specifies the token's decimal places (e.g., 6 for USDC, 18 for ETH).
func FormatTokenAmount(wei *big.Int, decimals int) string {
	if wei == nil || wei.Sign() == 0 {
		return "0"
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	wholePart := new(big.Int).Div(wei, divisor)
	remainder := new(big.Int).Mod(wei, divisor)

	if remainder.Sign() == 0 {
		return wholePart.String()
	}

	remainderStr := remainder.String()
	padding := decimals - len(remainderStr)
	for i := 0; i < padding; i++ {
		remainderStr = "0" + remainderStr
	}

	remainderStr = strings.TrimRight(remainderStr, "0")
	if remainderStr == "" {
		return wholePart.String()
	}

	return fmt.Sprintf("%s.%s", wholePart.String(), remainderStr)
}

// FormatETH formats wei as ETH string with "ETH" suffix.
func FormatETH(wei *big.Int) string {
	return FormatTokenAmount(wei, 18) + " ETH"
}

// ParseHexValue parses a hex string (with or without 0x prefix) to big.Int.
func ParseHexValue(data string) *big.Int {
	if data == "" || data == "0x" {
		return big.NewInt(0)
	}

	if strings.HasPrefix(data, "0x") {
		data = data[2:]
	}

	value := new(big.Int)
	value.SetString(data, 16)
	return value
}
