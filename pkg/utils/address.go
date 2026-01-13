// Package utils provides common utility functions for the Asentric SDK.
package utils

import "strings"

// ZeroAddress is the Ethereum zero address used to detect mints.
const ZeroAddress = "0x0000000000000000000000000000000000000000"

// ExtractAddressFromTopic extracts a 20-byte address from a 32-byte topic.
// Topics are left-padded with zeros: 0x000000000000000000000000<address>
func ExtractAddressFromTopic(topic interface{}) string {
	topicStr, ok := topic.(string)
	if !ok {
		return ""
	}
	// Topic is 32 bytes (64 hex chars + 0x), address is last 20 bytes (40 hex chars)
	if len(topicStr) >= 42 {
		return "0x" + topicStr[len(topicStr)-40:]
	}
	return topicStr
}

// TruncateAddress shortens an address for display (e.g., 0x1234...5678).
func TruncateAddress(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

// IsZeroAddress checks if the address is the zero address (used in mints).
func IsZeroAddress(addr string) bool {
	return addr == ZeroAddress || strings.Contains(addr, "000000000000000000000000000000000000")
}

// NormalizeAddress lowercases an address for comparison.
func NormalizeAddress(addr string) string {
	return strings.ToLower(addr)
}
