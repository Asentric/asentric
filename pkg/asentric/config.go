package asentric

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuntimeConfig holds all configuration for the Asentric runtime.
// This is loaded from config/asentric.yaml
//
// Design: Infrastructure-agnostic. Specific implementations (WebSocket, Redis, Webhook)
// are injected at runtime via With* methods on Runtime.
type RuntimeConfig struct {
	Version string `yaml:"version"` // Config version (e.g., "1.0")

	Chain  ChainConfig  `yaml:"chain"`
	Source SourceConfig `yaml:"source"`
	Sink   SinkConfig   `yaml:"sink"`
	Queue  QueueConfig  `yaml:"queue"`
	ABI    ABIConfig    `yaml:"abi"`

	Debug bool `yaml:"debug"` // Enable debug logging

	// Runtime fields (set programmatically, not from YAML)
	Engine *Engine `yaml:"-"`
	Logger Logger  `yaml:"-"`
}

// ChainConfig defines the blockchain configuration.
type ChainConfig struct {
	ID     int64  `yaml:"id"`     // Chain ID (1 for Ethereum, 5000 for Mantle)
	Name   string `yaml:"name"`   // Human readable name
	RPCURL string `yaml:"rpcUrl"` // HTTP RPC endpoint
	RPCWS  string `yaml:"rpcWs"`  // WebSocket RPC endpoint (for subscriptions)
}

// SourceConfig defines the event source configuration.
type SourceConfig struct {
	Type string `yaml:"type"` // "websocket", "memory"
	URL  string `yaml:"url"`  // WebSocket URL or empty for memory
}

// SinkConfig defines the alert sink configuration.
type SinkConfig struct {
	Type string `yaml:"type"` // "webhook", "console"
	URL  string `yaml:"url"`  // Webhook URL or empty for console
}

// QueueConfig defines the queue configuration.
type QueueConfig struct {
	Type string `yaml:"type"` // "redis", "memory"
	URL  string `yaml:"url"`  // Redis URL or empty for memory
}

// ABIConfig defines the ABI registry configuration.
type ABIConfig struct {
	RegistryPath string `yaml:"registryPath"` // Path to registry.yaml
}

// LoadConfig loads configuration from a YAML file.
// Returns error if file not found or invalid YAML.
func LoadConfig(path string) (*RuntimeConfig, error) {
	// Check file exists
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigReadFailed, err)
	}

	// Parse YAML
	var cfg RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigParseFailed, err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadConfigOrDefault loads config from path, or returns default config if path is empty.
func LoadConfigOrDefault(path string) (*RuntimeConfig, error) {
	if path == "" {
		return DefaultRuntimeConfig(), nil
	}
	return LoadConfig(path)
}

// Validate checks if the configuration is valid.
func (c *RuntimeConfig) Validate() error {
	if c == nil {
		return ErrInvalidConfig
	}

	// Validate source
	if c.Source.Type == "" {
		return ErrSourceTypeRequired
	}
	if c.Source.Type == "websocket" && c.Source.URL == "" {
		return ErrSourceURLRequired
	}

	// Validate sink
	if c.Sink.Type == "" {
		return ErrSinkTypeRequired
	}
	if c.Sink.Type == "webhook" && c.Sink.URL == "" {
		return ErrSinkURLRequired
	}

	// Validate chain
	if c.Chain.ID == 0 {
		return ErrChainIDRequired
	}
	if c.Chain.Name == "" {
		return ErrChainNameRequired
	}

	// Validate websocket source requires RPCWS
	if c.Source.Type == "websocket" && c.Chain.RPCWS == "" {
		return ErrChainRPCWSRequired
	}

	return nil
}

// DefaultRuntimeConfig returns a default configuration for development.
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		Version: "1.0",
		Chain: ChainConfig{
			ID:   1,
			Name: "Ethereum",
		},
		Source: SourceConfig{
			Type: "memory",
		},
		Sink: SinkConfig{
			Type: "console",
		},
		Queue: QueueConfig{
			Type: "memory",
		},
	}
}
