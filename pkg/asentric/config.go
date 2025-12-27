package asentric

import (
	"fmt"
	"strings"
	"time"
)

// RuntimeConfig holds all configuration for the Asentric runtime.
// This is loaded from config/asentric.yaml
type RuntimeConfig struct {
	Chain   ChainConfig   `yaml:"chain"`
	Redis   RedisConfig   `yaml:"redis"`
	Webhook WebhookConfig `yaml:"webhook"`
	Engine  EngineConfig  `yaml:"engine"`
	Debug   bool          `yaml:"debug"`
}

// Validate checks if the configuration is valid.
func (c *RuntimeConfig) Validate() error {

	if c == nil {
		return ErrInvalidConfig
	}

	if err := validateChainConfig(c.Chain); err != nil {
		return err
	}
	if err := validateRedisConfig(c.Redis); err != nil {
		return err
	}
	if err := validateWebhookConfig(c.Webhook); err != nil {
		return err
	}
	return nil
}

// ChainConfig holds chain-specific configuration.
type ChainConfig struct {
	// RPCWS is the WebSocket RPC endpoint (required)
	// Example: "wss://rpc.mantle.xyz/ws"
	RPCWS string `yaml:"rpc_ws"`

	// Name is the network name for alerts (required)
	// Example: "Mantle", "Ethereum", "Arbitrum"
	Name string `yaml:"name"`

	// ChainID is the numeric chain identifier (optional, auto-detect if not provided)
	ChainID uint64 `yaml:"chain_id"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	// Addr is the Redis server address (required)
	// Example: "localhost:6379"
	Addr string `yaml:"addr"`

	// Password is the Redis password (optional)
	Password string `yaml:"password"`

	// DB is the Redis database number (optional, default 0)
	DB int `yaml:"db"`

	// PoolSize is the connection pool size (optional, default 10)
	PoolSize int `yaml:"pool_size"`
}

// WebhookConfig holds webhook delivery configuration.
type WebhookConfig struct {
	// URL is the webhook endpoint (required)
	// Example: "https://your-webhook.com/alerts"
	URL string `yaml:"url"`

	// Timeout is the request timeout (optional, default 10s)
	Timeout time.Duration `yaml:"timeout"`

	// RetryCount is the number of retries on failure (optional, default 3)
	RetryCount int `yaml:"retry_count"`

	// Headers are additional HTTP headers to send (optional)
	Headers map[string]string `yaml:"headers"`
}

// EngineConfig holds engine-specific configuration.
type EngineConfig struct {
	// FailFast stops processing on first rule error (default false)
	FailFast bool `yaml:"fail_fast"`
}

// RegistryConfig holds target monitoring configuration.
// This is loaded from config/registry.yaml
type RegistryConfig struct {
	Targets []TargetConfig `yaml:"targets"`
}

// TargetConfig holds configuration for a single monitoring target.
type TargetConfig struct {
	// Address is the contract address (required)
	// Example: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
	Address string `yaml:"address"`

	// Name is the contract name for alerts (required)
	// Example: "Uniswap V3 Router"
	Name string `yaml:"name"`

	// ABIPath is the path to the ABI JSON file (required)
	// Example: "abi/uniswap_v3.json"
	ABIPath string `yaml:"abi_path"`
}

func (r *RuntimeConfig) ApplyDefaults() {
	if r.Redis.PoolSize == 0 {
		r.Redis.PoolSize = 10
	}
	if r.Webhook.Timeout == 0 {
		r.Webhook.Timeout = 10 * time.Second
	}
	if r.Webhook.RetryCount == 0 {
		r.Webhook.RetryCount = 3
	}
}

// validateChainConfig validates chain configuration.
func validateChainConfig(chain ChainConfig) error {
	if chain.RPCWS == "" {
		return fmt.Errorf("%w: chain.rpc_ws is required", ErrInvalidChainConfig)
	}

	if !strings.HasPrefix(chain.RPCWS, "ws://") && !strings.HasPrefix(chain.RPCWS, "wss://") {
		return fmt.Errorf("%w: chain.rpc_ws must be ws:// or wss://", ErrInvalidChainConfig)
	}

	if chain.Name == "" {
		return fmt.Errorf("%w: chain.name is required", ErrInvalidChainConfig)
	}
	return nil
}

// validateRedisConfig validates Redis configuration.
func validateRedisConfig(redis RedisConfig) error {
	if redis.Addr == "" {
		return fmt.Errorf("%w: redis.addr is required", ErrInvalidRedisConfig)
	}
	return nil
}

// validateWebhookConfig validates webhook configuration.
func validateWebhookConfig(webhook WebhookConfig) error {
	if webhook.URL == "" {
		return fmt.Errorf("%w: webhook.url is required", ErrInvalidWebhookConfig)
	}
	return nil
}
