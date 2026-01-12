package abi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const testERC20ABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

func TestLoader_LoadFromBytes(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	parsed, err := loader.LoadFromBytes([]byte(testERC20ABI))
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if _, ok := parsed.Events["Transfer"]; !ok {
		t.Error("Expected Transfer event")
	}
}

func TestLoader_LoadFromBytes_Artifact(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	// Hardhat/Foundry artifact format
	artifact := `{"abi":[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]}`

	parsed, err := loader.LoadFromBytes([]byte(artifact))
	if err != nil {
		t.Fatalf("Failed to load artifact: %v", err)
	}

	if _, ok := parsed.Events["Transfer"]; !ok {
		t.Error("Expected Transfer event")
	}
}

func TestLoader_LoadFromBytes_Invalid(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	_, err := loader.LoadFromBytes([]byte("not valid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoader_LoadFromString(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	parsed, err := loader.LoadFromString(testERC20ABI)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if _, ok := parsed.Events["Transfer"]; !ok {
		t.Error("Expected Transfer event")
	}
}

func TestLoader_LoadFromFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	abiPath := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(abiPath, []byte(testERC20ABI), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	registry := NewRegistry()
	loader := NewLoader(registry)

	parsed, err := loader.LoadFromFile(abiPath)
	if err != nil {
		t.Fatalf("Failed to load from file: %v", err)
	}

	if _, ok := parsed.Events["Transfer"]; !ok {
		t.Error("Expected Transfer event")
	}
}

func TestLoader_LoadFromFile_NotFound(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	_, err := loader.LoadFromFile("/nonexistent/path.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoader_LoadAndRegister(t *testing.T) {
	tmpDir := t.TempDir()
	abiPath := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(abiPath, []byte(testERC20ABI), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	registry := NewRegistry()
	loader := NewLoader(registry)

	addr := "0x1234567890123456789012345678901234567890"
	if err := loader.LoadAndRegisterHex(addr, abiPath); err != nil {
		t.Fatalf("Failed to load and register: %v", err)
	}

	if !registry.HasHex(addr) {
		t.Error("Expected ABI to be registered")
	}

	if registry.Count() != 1 {
		t.Errorf("Expected 1 ABI, got %d", registry.Count())
	}
}

func TestLoader_LoadAndRegisterHex_InvalidAddress(t *testing.T) {
	tmpDir := t.TempDir()
	abiPath := filepath.Join(tmpDir, "test.json")
	os.WriteFile(abiPath, []byte(testERC20ABI), 0644)

	registry := NewRegistry()
	loader := NewLoader(registry)

	err := loader.LoadAndRegisterHex("invalid", abiPath)
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

func TestLoader_LoadAndRegisterString(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	if err := loader.LoadAndRegisterString(
		addr,
		testERC20ABI,
	); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if !registry.Has(addr) {
		t.Error("Expected ABI to be registered")
	}
}

func TestLoader_LoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files named as addresses
	addr1 := "0x1111111111111111111111111111111111111111"
	addr2 := "0x2222222222222222222222222222222222222222"

	os.WriteFile(filepath.Join(tmpDir, addr1+".json"), []byte(testERC20ABI), 0644)
	os.WriteFile(filepath.Join(tmpDir, addr2+".json"), []byte(testERC20ABI), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not an abi"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notanaddress.json"), []byte(testERC20ABI), 0644)

	registry := NewRegistry()
	loader := NewLoader(registry)

	if err := loader.LoadDirectory(tmpDir); err != nil {
		t.Fatalf("Failed to load directory: %v", err)
	}

	// Should only load the two address-named files
	if registry.Count() != 2 {
		t.Errorf("Expected 2 ABIs, got %d", registry.Count())
	}

	if !registry.HasHex(addr1) {
		t.Error("Expected addr1 to be registered")
	}

	if !registry.HasHex(addr2) {
		t.Error("Expected addr2 to be registered")
	}
}

func TestLoader_LoadDirectory_NotFound(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	err := loader.LoadDirectory("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestLoader_Registry(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)

	if loader.Registry() != registry {
		t.Error("Registry() should return the same registry")
	}
}
