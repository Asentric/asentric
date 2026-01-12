package abi

import (
	"strings"
	"testing"

	"github.com/asentric/asentric/pkg/domain"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ERC20 Transfer event ABI for testing
const erc20ABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"}]`

func TestRegistry_NewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 ABIs, got %d", registry.Count())
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	// Parse ABI
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		t.Fatalf("Failed to parse ABI: %v", err)
	}

	// Register
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	registry.Register(addr, parsed)

	// Verify count
	if registry.Count() != 1 {
		t.Errorf("Expected 1 ABI, got %d", registry.Count())
	}

	// Get
	result := registry.Get(addr)
	if result == nil {
		t.Fatal("Expected ABI, got nil")
	}

	// Check event exists
	if _, ok := result.Events["Transfer"]; !ok {
		t.Error("Expected Transfer event in ABI")
	}

	if _, ok := result.Events["Approval"]; !ok {
		t.Error("Expected Approval event in ABI")
	}
}

func TestRegistry_RegisterHex(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	err := registry.RegisterHex("0x1234567890123456789012345678901234567890", parsed)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 ABI, got %d", registry.Count())
	}
}

func TestRegistry_RegisterHex_InvalidAddress(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	err := registry.RegisterHex("invalid", parsed)
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	otherAddr := common.HexToAddress("0x9999999999999999999999999999999999999999")

	registry.Register(addr, parsed)

	if !registry.Has(addr) {
		t.Error("Expected Has to return true for registered address")
	}

	if registry.Has(otherAddr) {
		t.Error("Expected Has to return false for unregistered address")
	}
}

func TestRegistry_GetEventByTopic(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	registry.Register(addr, parsed)

	// Transfer event topic: keccak256("Transfer(address,address,uint256)")
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	event := registry.GetEventByTopic(transferTopic)
	if event == nil {
		t.Fatal("Expected event, got nil")
	}

	if event.Name != "Transfer" {
		t.Errorf("Expected Transfer, got %s", event.Name)
	}
}

func TestRegistry_GetEventByTopic_NotFound(t *testing.T) {
	registry := NewRegistry()

	unknownTopic := common.HexToHash("0x1234567890abcdef")
	event := registry.GetEventByTopic(unknownTopic)

	if event != nil {
		t.Error("Expected nil for unknown topic")
	}
}

func TestRegistry_Addresses(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	registry.Register(addr1, parsed)
	registry.Register(addr2, parsed)

	addresses := registry.Addresses()
	if len(addresses) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(addresses))
	}
}

func TestRegistry_DomainInterface(t *testing.T) {
	registry := NewRegistry()
	parsed, _ := abi.JSON(strings.NewReader(erc20ABI))

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	registry.Register(addr, parsed)

	// Test domain.ABIRegistry interface - GetEvent
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	event, found := registry.GetEvent("0x1234567890123456789012345678901234567890", "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	if !found {
		t.Error("GetEvent should find Transfer event")
	}
	if event.Name != "Transfer" {
		t.Errorf("Expected event name 'Transfer', got '%s'", event.Name)
	}

	// Test for unknown event
	_, found = registry.GetEvent("0x1234567890123456789012345678901234567890", "0xunknown")
	if found {
		t.Error("GetEvent should not find unknown event")
	}

	// Test for unknown contract
	_, found = registry.GetEvent("0x9999999999999999999999999999999999999999", domain.Hash(transferTopic.Hex()))
	if found {
		t.Error("GetEvent should not find event for unknown contract")
	}
}
