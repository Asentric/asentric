package domain

import "math/big"

// TxActionType represents the semantic type of transaction action
type TxActionType string

const (
	TxActionCall     TxActionType = "CALL"
	TxActionCreate   TxActionType = "CREATE"
	TxActionDelegate TxActionType = "DELEGATECALL"
)

// TxType represents the transaction type (legacy, EIP-2930, EIP-1559, etc.)
type TxType uint8

const (
	TxTypeLegacy TxType = iota
	TxTypeAccessList
	TxTypeDynamicFee
	TxTypeBlob
)

// ContractCall represents a decoded contract call
// Semantic representation - rule authors don't need to decode calldata
type ContractCall struct {
	Contract Address
	Method   string
	Args     map[string]any // decoded ABI
}

// Transaction represents a blockchain transaction
// Call is nil-safe - nil if not a contract call
// Status is final (no pending state)
// Raw calldata is not exposed to rule authors
type Transaction struct {
	// Identity
	Hash  Hash
	Index uint64 // position in block

	// Parties
	From Address
	To   Address // empty for contract creation

	// Execution
	Nonce    uint64
	GasLimit uint64
	GasUsed  uint64
	Status   bool // true = success, false = reverted

	// Value (internal storage)
	RawValue NativeValue

	// Gas pricing (strings for precision)
	GasPrice     string // legacy or effective gas price
	MaxFeePerGas string // EIP-1559
	MaxPriority  string // EIP-1559

	// Type info
	Type   TxType
	Action TxActionType

	// Decoded call (nil if not a contract call or not decoded)
	Call *ContractCall

	// Block context
	BlockNumber uint64
	BlockHash   Hash
	Timestamp   uint64
}

// Value returns the transaction value as *big.Int.
// This is the preferred method for rule authors to access value.
// Returns 0 if value is empty or invalid.
func (t Transaction) Value() *big.Int {
	if t.RawValue.Wei == "" || t.RawValue.Wei == "0" {
		return big.NewInt(0)
	}
	val, ok := new(big.Int).SetString(t.RawValue.Wei, 10)
	if !ok {
		return big.NewInt(0)
	}
	return val
}
