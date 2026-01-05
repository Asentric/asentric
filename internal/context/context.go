// Package context provides concrete Context implementations.
// Context is the bridge between Event and Engine evaluation.
package context

import (
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// EventContext is a concrete implementation of asentric.Context.
// It wraps an Event and provides access to transaction, block, and log data.
type EventContext struct {
	event       asentric.Event
	chainID     domain.ChainID
	tx          domain.Transaction
	block       domain.Block
	logs        []domain.Log
	abiRegistry domain.ABIRegistry
}

// NewEventContext creates a new EventContext from an Event.
// The context builder extracts and normalizes data from the event payload.
func NewEventContext(event asentric.Event) *EventContext {
	return &EventContext{
		event:   event,
		chainID: domain.ChainID(event.ChainID),
	}
}

// WithChainID sets the chain ID.
func (c *EventContext) WithChainID(chainID domain.ChainID) *EventContext {
	c.chainID = chainID
	return c
}

// WithTransaction sets the transaction data.
func (c *EventContext) WithTransaction(tx domain.Transaction) *EventContext {
	c.tx = tx
	return c
}

// WithBlock sets the block data.
func (c *EventContext) WithBlock(block domain.Block) *EventContext {
	c.block = block
	return c
}

// WithLogs sets the log data.
func (c *EventContext) WithLogs(logs []domain.Log) *EventContext {
	c.logs = logs
	return c
}

// WithABI sets the ABI registry.
func (c *EventContext) WithABI(abi domain.ABIRegistry) *EventContext {
	c.abiRegistry = abi
	return c
}

// ChainID implements asentric.Context.
func (c *EventContext) ChainID() domain.ChainID {
	return c.chainID
}

// Tx implements asentric.Context.
func (c *EventContext) Tx() domain.Transaction {
	return c.tx
}

// Block implements asentric.Context.
func (c *EventContext) Block() domain.Block {
	return c.block
}

// Logs implements asentric.Context.
func (c *EventContext) Logs() []domain.Log {
	return c.logs
}

// ABI implements asentric.Context.
func (c *EventContext) ABI() domain.ABIRegistry {
	return c.abiRegistry
}

// Event returns the original event (for reference, not part of Context interface).
func (c *EventContext) Event() asentric.Event {
	return c.event
}

// Ensure EventContext implements asentric.Context
var _ asentric.Context = (*EventContext)(nil)
