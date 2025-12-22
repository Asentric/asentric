# Asentric SDK – MVP Specification (LOCKED)

> **Status:** 🔒 **LOCKED – FINAL FOR HACKATHON**  
> **Version:** 1.0  
> **Last Updated:** December 2024

**Dokumen ini adalah SINGLE SOURCE OF TRUTH untuk implementasi Asentric SDK.**

Jika terjadi konflik dengan dokumen lain, dokumen ini yang benar.

---

## Table of Contents

1. [Scope & Goals](#1-scope--goals)
2. [Infrastructure Requirements](#2-infrastructure-requirements)
3. [Data Flow](#3-data-flow)
4. [Configuration](#4-configuration)
5. [Public API Contract](#5-public-api-contract)
6. [Domain Types](#6-domain-types)
7. [Alert Format](#7-alert-format)
8. [Developer Experience](#8-developer-experience)
9. [Project Structure](#9-project-structure)
10. [Internal Architecture](#10-internal-architecture)
11. [CLI Commands](#11-cli-commands)
12. [Error Handling](#12-error-handling)
13. [Out of Scope](#13-out-of-scope)

---

## 1. Scope & Goals

### 1.1 Target

| Aspek | Keputusan |
|-------|-----------|
| **Scope** | MVP Hackathon |
| **Target User** | Developer self-hosted |
| **Chain Support** | EVM only |
| **Chain per Project** | 1 chain per project (fixed) |
| **Filosofi** | Simple DX seperti Ponder.sh |

### 1.2 Core Value Proposition

Asentric adalah SDK untuk **real-time smart contract security monitoring** yang memungkinkan developer:

- ✅ Mendefinisikan **apa yang dimonitor** melalui konfigurasi YAML
- ✅ Menulis **logic deteksi sendiri** melalui custom rules (Go)
- ✅ Menjalankan **self-hosted runtime** 
- ✅ Menerima **alert via webhook** secara real-time

### 1.3 Design Principles

- **YAML untuk konfigurasi, bukan logic**
- **Rules adalah code, bukan config**
- **Engine deterministic & stateless**
- **Runtime handle semua infrastructure**
- **Simple > Complex**

---

## 2. Infrastructure Requirements

### 2.1 Required

| Component | Purpose | Status |
|-----------|---------|--------|
| **Redis** | Message queue (internal framework) | ✅ Required |
| **RPC WebSocket** | Chain data subscription | ✅ Required |
| **Webhook URL** | Alert delivery | ✅ Required |

### 2.2 Not Required

| Component | Status |
|-----------|--------|
| Database | ❌ Not in scope |
| Dashboard | ❌ Not in scope |
| Multiple alert sinks | ❌ Not in scope |

### 2.3 Setup

```bash
# Redis
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

---

## 3. Data Flow

### 3.1 Canonical Flow (WAJIB)

```
┌─────────────────────────────────────────────────────────────────┐
│                         DATA FLOW                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [RPC WebSocket]                                                │
│        │                                                         │
│        │ eth_subscribe("logs", {addresses})                     │
│        ▼                                                         │
│  ┌─────────────────┐                                            │
│  │  EventSource    │  Subscribe logs from chain                 │
│  │  (WebSocket)    │                                            │
│  └────────┬────────┘                                            │
│           │ Event                                                │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │  Redis Queue    │  Event queue (internal)                    │
│  │  (events)       │                                            │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │    Runtime      │  Pop event, build Context                  │
│  │                 │  (framework internal)                      │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │    Engine       │  Evaluate all registered rules             │
│  │  (Evaluate)     │  Pure, deterministic, no I/O               │
│  └────────┬────────┘                                            │
│           │ []*Alert                                             │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │  Redis Queue    │  Alert queue (internal)                    │
│  │  (alerts)       │                                            │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │   AlertSink     │  POST JSON to webhook                      │
│  │   (Webhook)     │                                            │
│  └─────────────────┘                                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Flow Rules (WAJIB DIPATUHI)

- ✅ Event hanya dari WebSocket subscription
- ✅ Event masuk Redis queue sebelum diproses
- ✅ Context dibuat dari Event (immutable)
- ✅ Engine.Evaluate() sequential, deterministic
- ✅ Alert masuk Redis queue sebelum dikirim
- ✅ Alert dikirim via Webhook POST

---

## 4. Configuration

### 4.1 File Structure

```
config/
├── asentric.yaml    # Runtime & engine config
└── registry.yaml    # Target monitoring list
```

### 4.2 asentric.yaml

```yaml
# config/asentric.yaml

# Chain configuration
chain:
  rpc_ws: "wss://rpc.mantle.xyz/ws"
  name: "Mantle"           # Network name for alerts
  chain_id: 5000           # Optional, auto-detect if not provided

# Redis configuration (required)
redis:
  addr: "localhost:6379"
  password: ""             # Optional
  db: 0                    # Optional, default 0

# Webhook configuration (required)
webhook:
  url: "https://your-webhook.com/alerts"
  timeout: 10s             # Optional, default 10s

# Engine configuration (optional)
engine:
  fail_fast: false         # Stop on first rule error
```

### 4.3 registry.yaml

```yaml
# config/registry.yaml

targets:
  - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
    name: "Uniswap V3 Router"        # Required
    abi_path: "abi/uniswap_v3.json"  # Required
    
  - address: "0x..."
    name: "My Protocol Vault"
    abi_path: "abi/vault.json"
```

### 4.4 Field Reference

**asentric.yaml:**

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `chain.rpc_ws` | ✅ Yes | - | WebSocket RPC endpoint |
| `chain.name` | ✅ Yes | - | Network name for alerts |
| `chain.chain_id` | ❌ No | Auto-detect | Chain ID |
| `redis.addr` | ✅ Yes | - | Redis address |
| `redis.password` | ❌ No | "" | Redis password |
| `redis.db` | ❌ No | 0 | Redis database |
| `webhook.url` | ✅ Yes | - | Webhook URL |
| `webhook.timeout` | ❌ No | 10s | Request timeout |
| `engine.fail_fast` | ❌ No | false | Stop on error |

**registry.yaml:**

| Field | Required | Description |
|-------|----------|-------------|
| `address` | ✅ Yes | Contract address (0x...) |
| `name` | ✅ Yes | Contract name for alerts |
| `abi_path` | ✅ Yes | Path to ABI JSON file |

---

## 5. Public API Contract

### 5.1 Package: `pkg/asentric`

```go
import "github.com/asentric/asentric/pkg/asentric"
```

### 5.2 Engine

```go
// Engine is a CONCRETE TYPE, not an interface
type Engine struct {
    // internal state
}

// NewEngine creates a new Engine instance
func NewEngine() *Engine

// RegisterRule registers a rule into the engine
// Rules are executed in registration order (deterministic)
func (e *Engine) RegisterRule(rule Rule) error

// Evaluate evaluates all registered rules against the Context
// Returns zero or more Alerts
// Engine is stateless - each Evaluate() call is independent
func (e *Engine) Evaluate(ctx Context) ([]*Alert, error)
```

**Engine Invariants:**
- ✅ Stateless (no per-event state)
- ✅ Deterministic (same input → same output)
- ✅ Single-threaded (not concurrency-safe)
- ✅ No I/O (pure computation)
- ✅ Panic recovery (returns ErrRulePanic)

### 5.3 Rule Interface

```go
type Rule interface {
    // Name returns unique identifier for this rule
    Name() string
    
    // Evaluate processes context and returns alert if triggered
    // Returns (nil, nil) if no detection
    // Returns (alert, nil) if detection triggered
    // Returns (nil, error) if execution error
    Evaluate(ctx Context) (*Alert, error)
}
```

**Rule Invariants:**
- ✅ Pure function (no side effects)
- ✅ Deterministic
- ✅ No I/O (no network, file, database)
- ✅ No global state
- ✅ Max 1 alert per evaluation

### 5.4 Context Interface

```go
type Context interface {
    ChainID() domain.ChainID    // Returns chain ID (uint64 alias)
    Tx() domain.Transaction     // Transaction data
    Block() domain.Block        // Block metadata
    Logs() []domain.Log         // Decoded logs
    ABI() domain.ABIRegistry    // ABI access
}
```

**Context Invariants:**
- ✅ Immutable (read-only)
- ✅ Created from Event only
- ✅ Complete snapshot (no lazy loading)

### 5.5 Alert

```go
type Severity string

const (
    SeverityCritical Severity = "CRITICAL"
    SeverityHigh     Severity = "HIGH"
    SeverityMedium   Severity = "MEDIUM"
    SeverityLow      Severity = "LOW"
    SeverityInfo     Severity = "INFO"
)

type Alert struct {
    Rule        string            // Rule name
    Severity    Severity          // Alert severity
    Title       string            // Short title
    Description string            // Detailed description
    Ref         *ExecutionRef     // Execution context (populated by engine)
    Metadata    map[string]any    // Flexible metadata
}

type ExecutionRef struct {
    TxHash      string
    BlockNumber uint64
}

// Helper methods
func NewAlert(rule string, severity Severity, title, description string) *Alert
func (a *Alert) WithMetadata(key string, value any) *Alert
```

### 5.6 Event

```go
type Event struct {
    ChainID     uint64    // Chain identifier
    BlockNumber uint64    // Block height
    TxHash      string    // Transaction hash
    Payload     any       // Event-specific data (read-only)
}
```

### 5.7 Interfaces for Extension

```go
// EventSource provides chain data ingestion
type EventSource interface {
    Start(ctx context.Context) (<-chan Event, error)
}

// AlertSink delivers alerts to external systems
type AlertSink interface {
    Emit(ctx context.Context, alert *Alert) error
}

// Dispatcher bridges EventSource and Engine (internal)
type Dispatcher interface {
    Dispatch(ctx context.Context, event Event) error
}
```

### 5.8 Errors

```go
var (
    ErrInvalidContext = errors.New("invalid context")
    ErrInvalidRule    = errors.New("invalid rule")
    ErrInvalidEvent   = errors.New("invalid event")
    ErrRulePanic      = errors.New("rule panic")
    ErrAlreadyRunning = errors.New("runtime already running")
    ErrNoDispatcher   = errors.New("dispatcher is not set")
)
```

### 5.9 Runtime (Public Facade)

```go
// RuntimeConfig holds all configuration
type RuntimeConfig struct {
    Chain   ChainConfig
    Redis   RedisConfig
    Webhook WebhookConfig
    Engine  EngineConfig
}

// LoadConfig loads configuration from directory
func LoadConfig(configDir string) (*RuntimeConfig, error)

// Runtime orchestrates the entire system
type Runtime struct {
    // internal
}

// NewRuntime creates a new runtime with config and engine
func NewRuntime(config *RuntimeConfig, engine *Engine) *Runtime

// WithLogger sets custom logger (optional)
func (r *Runtime) WithLogger(logger Logger) *Runtime

// Start begins the runtime (blocks until stopped)
func (r *Runtime) Start(ctx context.Context) error

// Stop gracefully stops the runtime
func (r *Runtime) Stop() error
```

---

## 6. Domain Types

### 6.1 Package: `pkg/domain`

```go
import "github.com/asentric/asentric/pkg/domain"
```

### 6.2 Primitive Types

```go
// Address represents Ethereum address (hex string with 0x)
type Address string

func (a Address) String() string
func (a Address) Hex() string
func (a Address) IsZero() bool

// Hash represents 32-byte hash (hex string with 0x)
type Hash string

func (h Hash) String() string
func (h Hash) Hex() string
func (h Hash) IsZero() bool

// ChainID represents chain identifier
type ChainID uint64
```

### 6.3 Transaction

```go
type TxActionType string

const (
    TxActionCall   TxActionType = "CALL"
    TxActionCreate TxActionType = "CREATE"
)

type Transaction struct {
    // Identity
    Hash  Hash
    Index uint64
    
    // Parties
    From Address
    To   Address    // Empty for contract creation
    
    // Execution
    Nonce    uint64
    GasLimit uint64
    GasUsed  uint64
    Status   bool    // true = success
    
    // Value
    value    *big.Int  // internal
    
    // Gas pricing
    GasPrice     string
    MaxFeePerGas string
    MaxPriority  string
    
    // Type info
    Type   TxType
    Action TxActionType
    
    // Decoded call (nil if not decoded)
    Call *ContractCall
    
    // Block context
    BlockNumber uint64
    BlockHash   Hash
    Timestamp   uint64
}

// Value returns transaction value as *big.Int
// This is the preferred method for rule authors
func (t Transaction) Value() *big.Int

type ContractCall struct {
    Contract Address
    Method   string
    Args     map[string]any
}
```

### 6.4 Block

```go
type Block struct {
    Number    uint64
    Hash      Hash
    Parent    Hash
    Timestamp uint64
    Miner     Address
    GasLimit  uint64
    GasUsed   uint64
    BaseFee   string    // Decimal string
    TxCount   int
}
```

### 6.5 Log

```go
type Log struct {
    Address     Address
    LogIndex    uint64
    TxHash      Hash
    TxIndex     uint64
    BlockNumber uint64
    BlockHash   Hash
    Event       Event     // Decoded event
}

type Event struct {
    Name   string
    Fields map[string]any
}
```

### 6.6 ABI Registry

```go
type ABIRegistry interface {
    GetMethod(address Address, selector string) (Method, bool)
    GetEvent(address Address, topic Hash) (Event, bool)
}

type Method struct {
    Name string
    Args []ABIArg
}

type ABIArg struct {
    Name string
    Type string
}
```

---

## 7. Alert Format

### 7.1 Webhook JSON Structure

```json
{
  "severity": "HIGH",
  "rule": "large_transfer_detection",
  "title": "Large Transfer Detected",
  "description": "Transfer of 1,000,000 USDC detected from vault",
  
  "network": {
    "name": "Mantle",
    "chain_id": 5000
  },
  
  "context": {
    "block_number": 12345678,
    "tx_hash": "0x...",
    "timestamp": "2024-12-23T10:30:00Z"
  },
  
  "details": {
    // Rule-specific fields (flexible)
  }
}
```

### 7.2 Alert Types & Details Examples

**Large Value Transfer:**
```json
"details": {
  "type": "LARGE_TRANSFER",
  "from": "0x...",
  "to": "0x...",
  "token": {
    "address": "0x...",
    "symbol": "USDC",
    "decimals": 6
  },
  "amount": "1000000000000",
  "amount_formatted": "1,000,000 USDC",
  "threshold": "500000000000",
  "threshold_formatted": "500,000 USDC"
}
```

**Contract Upgrade:**
```json
"details": {
  "type": "PROXY_UPGRADE",
  "proxy": "0x...",
  "proxy_name": "TransparentUpgradeableProxy",
  "old_implementation": "0x...",
  "new_implementation": "0x..."
}
```

**Ownership Change:**
```json
"details": {
  "type": "OWNERSHIP_CHANGE",
  "contract": "0x...",
  "contract_name": "Vault",
  "old_owner": "0x...",
  "new_owner": "0x..."
}
```

### 7.3 Details Field Rules

- ✅ `details` is flexible - rule author defines fields
- ✅ All formatted values (e.g., `amount_formatted`) provided by rule author
- ✅ Token info (symbol, decimals) provided by rule author
- ✅ Must be JSON-serializable

---

## 8. Developer Experience

### 8.1 Target Usage

```go
// cmd/watcher/main.go
package main

import (
    "context"
    "log"
    
    "github.com/asentric/asentric/pkg/asentric"
    "my-project/rules"
)

func main() {
    // 1. Load configuration
    config, err := asentric.LoadConfig("config/")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Create engine and register rules
    engine := asentric.NewEngine()
    engine.RegisterRule(&rules.LargeTransferRule{})
    engine.RegisterRule(&rules.UpgradeDetectionRule{})
    
    // 3. Create and start runtime
    runtime := asentric.NewRuntime(config, engine)
    
    // 4. Run (blocks until SIGINT/SIGTERM)
    if err := runtime.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 8.2 Example Rule

```go
// rules/large_transfer.go
package rules

import (
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
)

type LargeTransferRule struct {
    Threshold *big.Int
}

func NewLargeTransferRule(threshold *big.Int) *LargeTransferRule {
    return &LargeTransferRule{Threshold: threshold}
}

func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    // Check if value exceeds threshold
    if tx.Value().Cmp(r.Threshold) > 0 {
        return asentric.NewAlert(
            r.Name(),
            asentric.SeverityHigh,
            "Large Transfer Detected",
            "Transaction value exceeds threshold",
        ).WithMetadata("value", tx.Value().String()).
          WithMetadata("threshold", r.Threshold.String()).
          WithMetadata("from", tx.From.String()).
          WithMetadata("to", tx.To.String()), nil
    }
    
    return nil, nil
}
```

---

## 9. Project Structure

### 9.1 SDK Structure

```
asentric-sdk/
├── pkg/
│   ├── asentric/           # Public API (STABLE)
│   │   ├── engine.go
│   │   ├── rule.go
│   │   ├── context.go
│   │   ├── alert.go
│   │   ├── event.go
│   │   ├── config.go       # RuntimeConfig, LoadConfig
│   │   ├── runtime.go      # Runtime facade
│   │   ├── errors.go
│   │   ├── event_source.go
│   │   ├── alert_sink.go
│   │   └── dispatcher.go
│   │
│   └── domain/             # Domain types (STABLE)
│       ├── address.go
│       ├── hash.go
│       ├── chain.go
│       ├── transaction.go
│       ├── block.go
│       ├── log.go
│       ├── event.go
│       ├── abi.go
│       └── value.go
│
├── internal/               # Private implementation
│   ├── source/             # EventSource implementations
│   │   └── websocket.go
│   │
│   ├── sink/               # AlertSink implementations
│   │   └── webhook.go
│   │
│   ├── queue/              # Redis queue
│   │   └── redis.go
│   │
│   ├── config/             # Config loading
│   │   └── loader.go
│   │
│   ├── abi/                # ABI loading & decoding
│   │   ├── loader.go
│   │   └── decoder.go
│   │
│   ├── chain/              # Raw chain types (geth-compatible)
│   │   ├── types.go
│   │   └── client.go
│   │
│   ├── adapter/            # Type conversion
│   │   ├── converter.go
│   │   └── geth.go
│   │
│   ├── context/            # Context implementations
│   │   └── context.go
│   │
│   ├── dispatcher/         # Dispatcher implementation
│   │   └── dispatcher.go
│   │
│   └── runtime/            # Runtime implementation
│       ├── runtime.go
│       └── shutdown.go
│
├── cmd/
│   └── asentric/           # CLI
│       ├── main.go
│       ├── init.go
│       └── version.go
│
├── templates/              # Project templates
│   └── project/
│
├── examples/               # Example projects
│
└── docs/                   # Documentation
    ├── SPEC.md             # This file (LOCKED)
    └── ...
```

### 9.2 User Project Structure (Generated by `asentric init`)

```
my-project/
├── config/
│   ├── asentric.yaml       # Runtime config
│   └── registry.yaml       # Target list
│
├── rules/
│   └── example_rule.go     # Example rule
│
├── abi/
│   └── .gitkeep
│
├── cmd/
│   └── watcher/
│       └── main.go         # Entry point
│
├── go.mod
└── README.md
```

---

## 10. Internal Architecture

### 10.1 Hybrid Architecture

| Layer | geth types? | Description |
|-------|-------------|-------------|
| `pkg/asentric/` | ❌ No | Public API, stable |
| `pkg/domain/` | ❌ No | Domain types, string-based |
| `internal/chain/` | ✅ Yes | Raw types, geth-compatible |
| `internal/adapter/` | ✅ Yes | Conversion layer |

### 10.2 Dependency Rules

```
pkg/asentric  ←──  internal/*  (allowed)
pkg/domain    ←──  internal/*  (allowed)

pkg/asentric  ──→  internal/*  (FORBIDDEN)
pkg/domain    ──→  internal/*  (FORBIDDEN)
```

### 10.3 Internal Components

| Component | Location | Responsibility |
|-----------|----------|----------------|
| WebSocket Source | `internal/source/websocket.go` | Subscribe to chain logs |
| Webhook Sink | `internal/sink/webhook.go` | POST alerts to webhook |
| Redis Queue | `internal/queue/redis.go` | Event & alert queuing |
| Config Loader | `internal/config/loader.go` | YAML parsing |
| ABI Loader | `internal/abi/loader.go` | Load ABI from files |
| ABI Decoder | `internal/abi/decoder.go` | Decode call/event data |
| Converter | `internal/adapter/converter.go` | chain types → domain types |
| Context | `internal/context/context.go` | Concrete Context impl |
| Dispatcher | `internal/dispatcher/dispatcher.go` | Event → Context → Engine |
| Runtime | `internal/runtime/runtime.go` | Orchestration & lifecycle |

---

## 11. CLI Commands

### 11.1 Available Commands

| Command | Priority | Description |
|---------|----------|-------------|
| `asentric init <name>` | P0 | Initialize new project |
| `asentric version` | P0 | Show version |
| `asentric replay` | P2 (Nice to have) | Replay from fixture |

### 11.2 `asentric init`

```bash
asentric init my-protocol-monitor
cd my-protocol-monitor
go mod tidy
```

Generates:
- `config/asentric.yaml` (template)
- `config/registry.yaml` (template)
- `rules/example_rule.go`
- `cmd/watcher/main.go`
- `go.mod`
- `README.md`

---

## 12. Error Handling

### 12.1 Logging

```go
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

// Default: stdout logger
// Custom logger optional via runtime.WithLogger()
```

### 12.2 Error Strategy

| Scenario | Behavior |
|----------|----------|
| Rule error | Log error, continue to next rule |
| Rule panic | Recover, return ErrRulePanic, continue |
| WebSocket disconnect | Log error, exit |
| Redis error | Log error, exit |
| Webhook error | Log error, continue |

### 12.3 Graceful Shutdown

- Framework handles `SIGINT` and `SIGTERM`
- Completes current processing before exit
- Closes connections gracefully

---

## 13. Out of Scope

The following are **explicitly NOT in scope** for this MVP:

| Feature | Status | Notes |
|---------|--------|-------|
| Database integration | ❌ | Not needed for MVP |
| Multiple alert sinks | ❌ | Webhook only |
| Multi-chain per project | ❌ | 1 chain per project |
| Historical replay from chain | ❌ | Nice to have |
| Dashboard/UI | ❌ | Separate project |
| `asentric replay` CLI | ⚠️ | Nice to have |
| Auto-reconnection | ❌ | Log error and exit |
| Custom event sources | ❌ | WebSocket only |
| Alert batching | ❌ | One at a time |
| Rate limiting | ❌ | Not needed |

---

## Appendix A: Quick Reference

### A.1 Import Paths

```go
import "github.com/asentric/asentric/pkg/asentric"
import "github.com/asentric/asentric/pkg/domain"
```

### A.2 Minimal Working Example

```go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
)

type SimpleRule struct{}

func (r *SimpleRule) Name() string { return "simple_rule" }

func (r *SimpleRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    if ctx.Tx().Value().Cmp(big.NewInt(1e18)) > 0 {
        return asentric.NewAlert(
            r.Name(),
            asentric.SeverityHigh,
            "Large Value",
            "Transaction > 1 ETH",
        ), nil
    }
    return nil, nil
}

func main() {
    config, _ := asentric.LoadConfig("config/")
    engine := asentric.NewEngine()
    engine.RegisterRule(&SimpleRule{})
    
    runtime := asentric.NewRuntime(config, engine)
    runtime.Start(context.Background())
}
```

---

## Related Documentation

| Document | Purpose |
|----------|---------|
| **[IMPL-GUIDE.md](IMPL-GUIDE.md)** | Step-by-step implementation guide untuk tim |
| **[architecture.md](architecture.md)** | Core principles & boundaries |
| **[sdk-api.md](sdk-api.md)** | API contracts |
| **[developer-overview.md](developer-overview.md)** | End-to-end developer guide |
| **[project-structure.md](project-structure.md)** | Folder structure |

---

## Document Control

| Field | Value |
|-------|-------|
| **Status** | 🔒 LOCKED |
| **Version** | 1.0 |
| **Created** | December 2024 |
| **Author** | Asentric Team |
| **Reviewers** | - |

**This document is the SINGLE SOURCE OF TRUTH.**

Any conflict with other documentation → this document is correct.

---

**END OF SPECIFICATION**

