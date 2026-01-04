package asentric

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds basic configuration for the Engine.
// Used for simple FailFast behavior.
type Config struct {
	FailFast bool
}

// DefaultConfig returns a default configuration for the Engine.
func DefaultConfig() *Config {
	return &Config{
		FailFast: false,
	}
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

// ChainConfig defines the blockchain configuration.
type ChainConfig struct {
	ID     int64  `yaml:"id"`     // Chain ID (1 for Ethereum, 5000 for Mantle)
	Name   string `yaml:"name"`   // Human readable name
	RPCURL string `yaml:"rpcUrl"` // RPC endpoint
}

// ABIConfig defines the ABI registry configuration.
type ABIConfig struct {
	RegistryPath string `yaml:"registryPath"` // Path to registry.yaml
}

// RuntimeConfig holds all configuration for the Asentric runtime.
type RuntimeConfig struct {
	Version string `yaml:"version"` // Config version (e.g., "1.0")

	Chain  ChainConfig  `yaml:"chain"`
	Source SourceConfig `yaml:"source"`
	Sink   SinkConfig   `yaml:"sink"`
	Queue  QueueConfig  `yaml:"queue"`
	ABI    ABIConfig    `yaml:"abi"`

	// Runtime fields (set programmatically, not from YAML)
	Engine *Engine `yaml:"-"`
	Logger Logger  `yaml:"-"`
}

// LoadConfig loads configuration from a YAML file.
// Returns error if file not found or invalid YAML.
func LoadConfig(path string) (*RuntimeConfig, error) {
	// Check file exists
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("asentric: failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("asentric: failed to parse config file: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid.
func (c *RuntimeConfig) Validate() error {
	// Validate source
	if c.Source.Type == "" {
		return errors.New("asentric: source.type is required")
	}
	if c.Source.Type == "websocket" && c.Source.URL == "" {
		return errors.New("asentric: source.url is required for websocket type")
	}

	// Validate sink
	if c.Sink.Type == "" {
		return errors.New("asentric: sink.type is required")
	}
	if c.Sink.Type == "webhook" && c.Sink.URL == "" {
		return errors.New("asentric: sink.url is required for webhook type")
	}

	// Validate chain
	if c.Chain.ID == 0 {
		return errors.New("asentric: chain.id is required")
	}

	return nil
}

// LoadConfigOrDefault loads config from path, or returns default config if path is empty.
func LoadConfigOrDefault(path string) (*RuntimeConfig, error) {
	if path == "" {
		return DefaultRuntimeConfig(), nil
	}
	return LoadConfig(path)
}

// DefaultRuntimeConfig returns a default configuration for development.
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		Version: "1.0",
		Chain: ChainConfig{
			ID:   1,
			Name: "Mantle",
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
