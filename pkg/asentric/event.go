package asentric

// Event represents a normalized blockchain event.
// Events are:
// - Normalized representation of chain data
// - Infrastructure-agnostic (no Redis, RPC metadata)
// - Immutable
//
// Events do NOT:
// - Contain infrastructure metadata
// - Lazy-load data (must be complete when created)
type Event struct {
	// ChainID is the numeric chain identifier (EVM canonical)
	// Using uint64 ensures type safety and avoids ambiguity
	ChainID uint64

	// BlockNumber is the block height
	BlockNumber uint64

	// TxHash is the transaction hash (hex with 0x prefix)
	TxHash string

	// Payload contains the event-specific data
	// MUST be read-only
	// MUST be safe for concurrent read access
	Payload any
}
