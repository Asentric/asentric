package domain

// Block represents a blockchain block
// Contains enough information for correlation & heuristics
// Not overfitted with unnecessary details
type Block struct {
	Number    uint64
	Hash      Hash
	Parent    Hash
	Timestamp uint64

	// Producer info
	Miner Address

	// Gas info
	GasLimit uint64
	GasUsed  uint64
	BaseFee  string // decimal string, empty if pre-EIP-1559

	// Transaction count
	TxCount int
}
