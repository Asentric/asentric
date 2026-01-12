// Package chain provides low-level blockchain connectivity via op-geth.
// This package handles RPC connections and raw chain data types.
package chain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// RawLog represents a log entry from the blockchain.
// This is our internal representation, decoupled from go-ethereum types.
type RawLog struct {
	// Address is the contract that emitted the event
	Address common.Address

	// Topics contains the indexed event parameters
	// Topics[0] is typically the event signature hash
	Topics []common.Hash

	// Data contains the non-indexed event parameters
	Data []byte

	// BlockNumber is the block height where this log was emitted
	BlockNumber uint64

	// TxHash is the transaction hash that generated this log
	TxHash common.Hash

	// TxIndex is the transaction's position in the block
	TxIndex uint

	// BlockHash is the hash of the block containing this log
	BlockHash common.Hash

	// LogIndex is this log's position in the block
	LogIndex uint

	// Removed is true if this log was reverted due to chain reorganization
	Removed bool
}

// RawBlock represents block header information.
type RawBlock struct {
	// Number is the block height
	Number *big.Int

	// Hash is the block's hash
	Hash common.Hash

	// ParentHash is the parent block's hash
	ParentHash common.Hash

	// Timestamp is the block's Unix timestamp
	Timestamp uint64

	// GasLimit is the maximum gas allowed in this block
	GasLimit uint64

	// GasUsed is the total gas used by all transactions
	GasUsed uint64

	// BaseFee is the base fee per gas (EIP-1559)
	BaseFee *big.Int

	// Miner is the address of the block producer
	Miner common.Address
}

// RawTransaction represents transaction data.
type RawTransaction struct {
	// Hash is the transaction hash
	Hash common.Hash

	// Nonce is the sender's transaction count
	Nonce uint64

	// GasPrice is the gas price in wei (legacy or effective)
	GasPrice *big.Int

	// GasLimit is the maximum gas for this transaction
	GasLimit uint64

	// To is the recipient address (nil for contract creation)
	To *common.Address

	// Value is the amount of native token transferred
	Value *big.Int

	// Data is the transaction input data
	Data []byte

	// From is the sender address
	From common.Address

	// BlockNumber is the block containing this transaction
	BlockNumber uint64

	// BlockHash is the hash of the containing block
	BlockHash common.Hash

	// TxIndex is the transaction's position in the block
	TxIndex uint
}

// RawReceipt represents transaction receipt data.
type RawReceipt struct {
	// Status is 1 for success, 0 for failure
	Status uint64

	// GasUsed is the actual gas consumed
	GasUsed uint64

	// Logs are the event logs emitted by this transaction
	Logs []*types.Log

	// BlockNumber is the block containing this transaction
	BlockNumber uint64

	// BlockHash is the hash of the containing block
	BlockHash common.Hash

	// TxHash is the transaction hash
	TxHash common.Hash
}

// SubscriptionFilter defines the filter for log subscriptions.
type SubscriptionFilter struct {
	// Addresses filters logs by emitting contract address
	// Empty means all addresses
	Addresses []common.Address

	// Topics filters by indexed event parameters
	// Topics[0] is typically the event signature
	// Each topic position can have multiple options (OR)
	// Empty means all topics
	Topics [][]common.Hash
}

// FromEthLog converts go-ethereum Log to our RawLog.
func FromEthLog(log *types.Log) RawLog {
	return RawLog{
		Address:     log.Address,
		Topics:      log.Topics,
		Data:        log.Data,
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash,
		TxIndex:     log.TxIndex,
		BlockHash:   log.BlockHash,
		LogIndex:    log.Index,
		Removed:     log.Removed,
	}
}

// IsEmpty returns true if the filter has no constraints.
func (f SubscriptionFilter) IsEmpty() bool {
	return len(f.Addresses) == 0 && len(f.Topics) == 0
}
