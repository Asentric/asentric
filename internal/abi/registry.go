// Package abi provides ABI loading, storage, and event decoding for smart contracts.
// It enables the SDK to decode raw log data into meaningful event structures.
package abi

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/asentric/asentric/pkg/domain"
)

// Registry stores and retrieves ABIs by contract address.
// It implements domain.ABIRegistry for use in Context.
type Registry struct {
	mu   sync.RWMutex
	abis map[common.Address]abi.ABI
}

// NewRegistry creates a new empty ABI registry.
func NewRegistry() *Registry {
	return &Registry{
		abis: make(map[common.Address]abi.ABI),
	}
}

// Register adds an ABI for a contract address.
func (r *Registry) Register(address common.Address, contractABI abi.ABI) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.abis[address] = contractABI
}

// RegisterHex adds an ABI for a contract address specified as hex string.
func (r *Registry) RegisterHex(addressHex string, contractABI abi.ABI) error {
	if !common.IsHexAddress(addressHex) {
		return fmt.Errorf("abi: invalid address: %s", addressHex)
	}
	r.Register(common.HexToAddress(addressHex), contractABI)
	return nil
}

// Get retrieves the ABI for a contract address.
// Returns nil if not found.
func (r *Registry) Get(address common.Address) *abi.ABI {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.abis[address]; ok {
		return &a
	}
	return nil
}

// GetHex retrieves the ABI for a contract address specified as hex string.
func (r *Registry) GetHex(addressHex string) *abi.ABI {
	if !common.IsHexAddress(addressHex) {
		return nil
	}
	return r.Get(common.HexToAddress(addressHex))
}

// Has checks if an ABI is registered for an address.
func (r *Registry) Has(address common.Address) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.abis[address]
	return ok
}

// HasHex checks if an ABI is registered for an address (hex string).
func (r *Registry) HasHex(addressHex string) bool {
	if !common.IsHexAddress(addressHex) {
		return false
	}
	return r.Has(common.HexToAddress(addressHex))
}

// Addresses returns all registered contract addresses.
func (r *Registry) Addresses() []common.Address {
	r.mu.RLock()
	defer r.mu.RUnlock()
	addresses := make([]common.Address, 0, len(r.abis))
	for addr := range r.abis {
		addresses = append(addresses, addr)
	}
	return addresses
}

// Count returns the number of registered ABIs.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.abis)
}

// GetEventByTopic finds an event in any registered ABI by its topic hash.
// This is useful when you have a log but don't know which contract emitted it.
func (r *Registry) GetEventByTopic(topic common.Hash) *abi.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, contractABI := range r.abis {
		for _, event := range contractABI.Events {
			if event.ID == topic {
				return &event
			}
		}
	}
	return nil
}

// GetEventForAddress finds an event by topic for a specific contract address.
func (r *Registry) GetEventForAddress(address common.Address, topic common.Hash) *abi.Event {
	contractABI := r.Get(address)
	if contractABI == nil {
		return nil
	}

	for _, event := range contractABI.Events {
		if event.ID == topic {
			return &event
		}
	}
	return nil
}

// --- domain.ABIRegistry interface implementation ---

// GetMethod returns the method metadata for a contract address and selector.
// Implements domain.ABIRegistry.
func (r *Registry) GetMethod(address domain.Address, selector string) (domain.Method, bool) {
	contractABI := r.GetHex(string(address))
	if contractABI == nil {
		return domain.Method{}, false
	}

	// Try to find method by selector (first 4 bytes of keccak256)
	for _, method := range contractABI.Methods {
		if common.Bytes2Hex(method.ID) == selector || method.Sig == selector {
			args := make([]domain.ABIArg, len(method.Inputs))
			for i, input := range method.Inputs {
				args[i] = domain.ABIArg{
					Name: input.Name,
					Type: input.Type.String(),
				}
			}
			return domain.Method{
				Name: method.Name,
				Args: args,
			}, true
		}
	}

	return domain.Method{}, false
}

// GetEvent returns the event metadata for a contract address and topic.
// Implements domain.ABIRegistry.
func (r *Registry) GetEvent(address domain.Address, topic domain.Hash) (domain.Event, bool) {
	contractABI := r.GetHex(string(address))
	if contractABI == nil {
		return domain.Event{}, false
	}

	topicHash := common.HexToHash(string(topic))

	for _, event := range contractABI.Events {
		if event.ID == topicHash {
			return domain.Event{
				Name:   event.Name,
				Fields: nil, // Empty until decoded
			}, true
		}
	}

	return domain.Event{}, false
}

// Ensure Registry implements domain.ABIRegistry
var _ domain.ABIRegistry = (*Registry)(nil)
