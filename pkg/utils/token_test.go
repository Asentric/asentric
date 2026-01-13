package utils

import (
	"math/big"
	"testing"
)

func TestFormatTokenAmount(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		decimals int
		expected string
	}{
		{"nil", nil, 18, "0"},
		{"zero", big.NewInt(0), 18, "0"},
		{"1 ETH", big.NewInt(1e18), 18, "1"},
		{"0.5 ETH", big.NewInt(5e17), 18, "0.5"},
		{"1.5 ETH", new(big.Int).Add(big.NewInt(1e18), big.NewInt(5e17)), 18, "1.5"},
		{"1000000 USDC", big.NewInt(1000000e6), 6, "1000000"},
		{"1.5 USDC", big.NewInt(1500000), 6, "1.5"},
		{"0.000001 USDC", big.NewInt(1), 6, "0.000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTokenAmount(tt.wei, tt.decimals)
			if got != tt.expected {
				t.Errorf("FormatTokenAmount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseHexValue(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected *big.Int
	}{
		{"empty", "", big.NewInt(0)},
		{"0x", "0x", big.NewInt(0)},
		{"0x1", "0x1", big.NewInt(1)},
		{"0xff", "0xff", big.NewInt(255)},
		{"without prefix", "ff", big.NewInt(255)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHexValue(tt.data)
			if got.Cmp(tt.expected) != 0 {
				t.Errorf("ParseHexValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}
