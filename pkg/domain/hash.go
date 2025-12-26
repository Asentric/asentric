package domain

// Hash represents a 32-byte hash (transaction hash, block hash, topic, etc.)
// Uses string for ergonomics - hex-encoded with 0x prefix
type Hash string

func (h Hash) String() string {
	return string(h)
}

func (h Hash) Hex() string {
	return h.String()
}

func (h Hash) IsZero() bool {
	return h == "" || h == "0x0000000000000000000000000000000000000000000000000000000000000000"
}
