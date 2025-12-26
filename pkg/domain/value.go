package domain

// NativeValue represents native blockchain currency (ETH, MNT, etc.)
// Uses decimal string for immutability and safety
type NativeValue struct {
	Wei string // decimal string, immutable
}

func (v NativeValue) IsZero() bool {
	return v.Wei == "0"
}

// TokenAmount represents token value with metadata
type TokenAmount struct {
	Token  Token
	Amount string // decimal string
}

