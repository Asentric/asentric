package abi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Loader loads ABI files and registers them in a registry.
type Loader struct {
	registry *Registry
}

// RegistryEntry represents an entry in a registry configuration file.
type RegistryEntry struct {
	Address string `yaml:"address" json:"address"`
	Name    string `yaml:"name" json:"name"`
	ABIPath string `yaml:"abi_path" json:"abi_path"`
}

// RegistryConfig represents a registry configuration with target contracts.
type RegistryConfig struct {
	Targets []RegistryEntry `yaml:"targets" json:"targets"`
}

// NewLoader creates a new ABI loader for the given registry.
func NewLoader(registry *Registry) *Loader {
	return &Loader{
		registry: registry,
	}
}

// LoadFromFile loads an ABI from a JSON file.
func (l *Loader) LoadFromFile(path string) (abi.ABI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("abi: read file failed: %w", err)
	}

	return l.LoadFromBytes(data)
}

// LoadFromBytes loads an ABI from JSON bytes.
// Supports both direct ABI arrays and Hardhat/Foundry artifact format.
func (l *Loader) LoadFromBytes(data []byte) (abi.ABI, error) {
	// Try to parse as direct ABI array first
	parsed, err := abi.JSON(strings.NewReader(string(data)))
	if err == nil {
		return parsed, nil
	}

	// Try to parse as Hardhat/Foundry artifact (has "abi" field)
	var artifact struct {
		ABI json.RawMessage `json:"abi"`
	}
	if jsonErr := json.Unmarshal(data, &artifact); jsonErr == nil && len(artifact.ABI) > 0 {
		parsed, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
		if err == nil {
			return parsed, nil
		}
	}

	return abi.ABI{}, fmt.Errorf("abi: failed to parse ABI: %w", err)
}

// LoadFromString loads an ABI from a JSON string.
func (l *Loader) LoadFromString(abiJSON string) (abi.ABI, error) {
	return l.LoadFromBytes([]byte(abiJSON))
}

// LoadAndRegister loads an ABI file and registers it for an address.
func (l *Loader) LoadAndRegister(address common.Address, abiPath string) error {
	parsed, err := l.LoadFromFile(abiPath)
	if err != nil {
		return err
	}

	l.registry.Register(address, parsed)
	return nil
}

// LoadAndRegisterHex loads an ABI file and registers it for an address (hex string).
func (l *Loader) LoadAndRegisterHex(addressHex, abiPath string) error {
	if !common.IsHexAddress(addressHex) {
		return fmt.Errorf("abi: invalid address: %s", addressHex)
	}
	return l.LoadAndRegister(common.HexToAddress(addressHex), abiPath)
}

// LoadAndRegisterString loads an ABI from string and registers it.
func (l *Loader) LoadAndRegisterString(address common.Address, abiJSON string) error {
	parsed, err := l.LoadFromString(abiJSON)
	if err != nil {
		return err
	}

	l.registry.Register(address, parsed)
	return nil
}

// LoadRegistry loads all ABIs from a registry configuration.
func (l *Loader) LoadRegistry(config RegistryConfig, basePath string) error {
	for _, entry := range config.Targets {
		// Resolve relative path
		abiPath := entry.ABIPath
		if !filepath.IsAbs(abiPath) {
			abiPath = filepath.Join(basePath, abiPath)
		}

		if err := l.LoadAndRegisterHex(entry.Address, abiPath); err != nil {
			return fmt.Errorf("abi: failed to load %s (%s): %w", entry.Name, entry.Address, err)
		}
	}
	return nil
}

// LoadDirectory loads all .json files from a directory where filename is the address.
// Files should be named like "0x1234...5678.json"
func (l *Loader) LoadDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("abi: read dir failed: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		// Extract address from filename
		basename := strings.TrimSuffix(name, ".json")
		if !common.IsHexAddress(basename) {
			// Skip files that aren't named as addresses
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		if err := l.LoadAndRegisterHex(basename, fullPath); err != nil {
			return fmt.Errorf("abi: failed to load %s: %w", name, err)
		}
	}

	return nil
}

// Registry returns the underlying registry.
func (l *Loader) Registry() *Registry {
	return l.registry
}
