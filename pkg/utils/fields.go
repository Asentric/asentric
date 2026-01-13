// Package utils provides common utility functions for the Asentric SDK.
package utils

import (
	"fmt"
	"math/big"
)

// GetFieldString safely extracts a string value from a map.
func GetFieldString(fields map[string]any, key string) string {
	if fields == nil {
		return ""
	}
	if v, ok := fields[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// GetFieldBigInt safely extracts a *big.Int value from a map.
func GetFieldBigInt(fields map[string]any, key string) *big.Int {
	if fields == nil {
		return big.NewInt(0)
	}
	if v, ok := fields[key]; ok {
		switch val := v.(type) {
		case *big.Int:
			return val
		case string:
			n := new(big.Int)
			n.SetString(val, 10)
			return n
		case int64:
			return big.NewInt(val)
		case float64:
			return big.NewInt(int64(val))
		}
	}
	return big.NewInt(0)
}

// GetFieldUint64 safely extracts a uint64 value from a map.
func GetFieldUint64(fields map[string]any, key string) uint64 {
	if fields == nil {
		return 0
	}
	if v, ok := fields[key]; ok {
		switch val := v.(type) {
		case uint64:
			return val
		case int64:
			if val >= 0 {
				return uint64(val)
			}
		case int:
			if val >= 0 {
				return uint64(val)
			}
		case float64:
			if val >= 0 {
				return uint64(val)
			}
		}
	}
	return 0
}

// GetFieldBool safely extracts a bool value from a map.
func GetFieldBool(fields map[string]any, key string) bool {
	if fields == nil {
		return false
	}
	if v, ok := fields[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
