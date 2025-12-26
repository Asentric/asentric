package domain

// Log represents a blockchain log entry
// ABI-decoded and ready for rule evaluation
// No topic-level exposure to rule authors
type Log struct {
	// Location
	Address  Address
	LogIndex uint64
	TxHash   Hash
	TxIndex  uint64

	// Decoded event data
	Event Event

	// Block context
	BlockNumber uint64
	BlockHash   Hash
}
