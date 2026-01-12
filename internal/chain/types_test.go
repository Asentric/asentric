package chain

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestFromEthLog(t *testing.T) {
	// Create a sample go-ethereum log
	ethLog := &types.Log{
		Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Topics:      []common.Hash{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		Data:        []byte{1, 2, 3, 4},
		BlockNumber: 12345,
		TxHash:      common.HexToHash("0xabc123"),
		TxIndex:     2,
		BlockHash:   common.HexToHash("0xdef456"),
		Index:       5,
		Removed:     false,
	}

	// Convert
	rawLog := FromEthLog(ethLog)

	// Verify
	if rawLog.Address != ethLog.Address {
		t.Errorf("Address mismatch: expected %s, got %s", ethLog.Address.Hex(), rawLog.Address.Hex())
	}

	if len(rawLog.Topics) != len(ethLog.Topics) {
		t.Errorf("Topics length mismatch: expected %d, got %d", len(ethLog.Topics), len(rawLog.Topics))
	}

	if rawLog.Topics[0] != ethLog.Topics[0] {
		t.Errorf("Topic[0] mismatch")
	}

	if rawLog.BlockNumber != ethLog.BlockNumber {
		t.Errorf("BlockNumber mismatch: expected %d, got %d", ethLog.BlockNumber, rawLog.BlockNumber)
	}

	if rawLog.TxHash != ethLog.TxHash {
		t.Errorf("TxHash mismatch")
	}

	if rawLog.LogIndex != ethLog.Index {
		t.Errorf("LogIndex mismatch: expected %d, got %d", ethLog.Index, rawLog.LogIndex)
	}

	if rawLog.Removed != ethLog.Removed {
		t.Errorf("Removed mismatch")
	}
}

func TestSubscriptionFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		filter   SubscriptionFilter
		expected bool
	}{
		{
			name:     "empty filter",
			filter:   SubscriptionFilter{},
			expected: true,
		},
		{
			name: "with addresses",
			filter: SubscriptionFilter{
				Addresses: []common.Address{common.HexToAddress("0x1234")},
			},
			expected: false,
		},
		{
			name: "with topics",
			filter: SubscriptionFilter{
				Topics: [][]common.Hash{{common.HexToHash("0xabc")}},
			},
			expected: false,
		},
		{
			name: "with both",
			filter: SubscriptionFilter{
				Addresses: []common.Address{common.HexToAddress("0x1234")},
				Topics:    [][]common.Hash{{common.HexToHash("0xabc")}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestRawLog_Fields(t *testing.T) {
	// Test that all fields can be set and read correctly
	rawLog := RawLog{
		Address:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Topics:      []common.Hash{common.HexToHash("0xabc"), common.HexToHash("0xdef")},
		Data:        []byte{0x01, 0x02, 0x03},
		BlockNumber: 100,
		TxHash:      common.HexToHash("0x111"),
		TxIndex:     1,
		BlockHash:   common.HexToHash("0x222"),
		LogIndex:    2,
		Removed:     true,
	}

	if rawLog.BlockNumber != 100 {
		t.Errorf("BlockNumber mismatch")
	}

	if len(rawLog.Topics) != 2 {
		t.Errorf("Topics length mismatch")
	}

	if !rawLog.Removed {
		t.Errorf("Removed should be true")
	}
}
