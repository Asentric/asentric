# Asentric SDK

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**A Go SDK for building real-time, on-chain security detection logic in a modular, deterministic, and developer-friendly way.**

Asentric SDK provides a pure execution engine, rule system, and explicit runtime context that enable developers to write smart contract security rules without coupling to infrastructure concerns such as message queues, databases, APIs, or deployment systems.

---

## Overview

Asentric SDK is designed to be the **shared security brain** of the Asentric ecosystem. It provides a pure, infrastructure-agnostic layer for defining security detection logic that can be embedded into multiple runtime environments.

The SDK separates security logic from infrastructure concerns, allowing developers to focus on writing effective detection rules while infrastructure systems handle deployment, ingestion, and alert delivery.

---

## Core Capabilities

Asentric SDK empowers developers to:

- **Define Pure, Deterministic Security Rules** — Write detection logic as pure functions with no side effects
- **Process On-Chain Data** — Analyze transactions, logs, and decoded smart contract events
- **Generate Structured Alerts** — Produce standardized security alerts with severity levels and metadata
- **Enable Deterministic Replay** — Debug and test rules offline with guaranteed reproducibility
- **Build Infrastructure-Agnostic Logic** — Write rules once, deploy anywhere without infrastructure dependencies
- **Maintain Developer Focus** — Concentrate on security logic without worrying about deployment or infrastructure

---

## What Asentric SDK Is — and Is Not

### ✅ What This SDK Is

Asentric SDK is:

- A **security detection engine** for blockchain transactions and events
- A **rule execution framework** with a clean, testable API
- A **developer-focused SDK** prioritizing simplicity and clarity
- A **pure domain logic layer** with no infrastructure coupling
- A **shared core** used by multiple runtime systems in the Asentric ecosystem

### ❌ What This SDK Is Not

By design, Asentric SDK does **not** provide:

- Message queues (Redis, Kafka, RabbitMQ, etc.)
- HTTP APIs or web dashboards
- Database connections or persistence layers
- Container orchestration or deployment tools
- Notification systems (Slack, webhooks, email, etc.)
- RPC connections or blockchain node management

These responsibilities are handled by dedicated components in the Asentric ecosystem:

| Repository | Responsibility |
|------------|----------------|
| `asentric-bot` | Runtime watcher, chain data ingestion, alert delivery |
| `asentric-backend` | API server, alert aggregation, persistence layer |
| `asentric-frontend` | Web-based dashboard and visualization |

**The SDK remains a pure security detection engine.**

---

## Design Principles

### 1. Pure Domain Logic First

- Rules are implemented as **pure functions**
- **No side effects** — rules don't modify external state
- **No I/O operations** inside rule evaluation
- Easy to test, reason about, and debug

### 2. Infrastructure-Agnostic by Default

- SDK has **zero knowledge** of Redis, HTTP servers, or databases
- Alerts are **produced**, not delivered — delivery is an infrastructure concern
- All infrastructure integration lives **outside** the SDK

### 3. Explicit Context

- All execution data flows through a **Context** object
- **No global state** — everything is explicit and traceable
- **Deterministic execution** — same input always produces same output

### 4. Developer Experience (DX)

- **Minimal boilerplate** — get started quickly with sensible defaults
- **Idiomatic Go** — follows Go best practices and conventions
- **Library-first, CLI-assisted workflow** — scaffold, test, and replay from the command line
- **Easy local testing** — test rules without external dependencies

> **Important:** The CLI does not run production watchers or connect to blockchains. It is strictly a developer tool for scaffolding, replay, and rule validation.

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Asentric SDK                     │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Engine  │  │  Rules   │  │ Context  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│         │             │             │           │
│         └─────────────┴─────────────┘           │
│                     │                           │
└─────────────────────┼───────────────────────────┘
                      │
         ┌────────────┼────────────┐
         │            │            │
    ┌────▼───┐   ┌───▼────┐   ┌──▼──────┐
    │  Bot   │   │Backend │   │Frontend │
    │(Runtime)│  │  (API) │   │  (UI)   │
    └────────┘   └────────┘   └─────────┘
```

The SDK provides the **core detection logic**, while external systems handle:

- **Bot**: Chain monitoring, transaction ingestion, alert routing
- **Backend**: REST API, data aggregation, persistence
- **Frontend**: User interface, dashboards, visualization

---

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Redis 7+ (required for setup - like Ponder.sh requires Postgres)
- Basic understanding of blockchain transactions and smart contracts

### Installation

Install the CLI tool:

```bash
go install github.com/asentric/asentric@latest
```

### Setup Redis (Required)

Like Ponder.sh requires Postgres for setup, Asentric requires Redis for message queue and state management:

```bash
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

### Scaffold a New Project

Create a new detection project with the CLI:

```bash
asentric init my-protocol-monitor
cd my-protocol-monitor
go mod tidy
```

This generates a complete project structure:

```
my-protocol-monitor/
├── config/
│   ├── asentric.yaml      # Engine configuration
│   ├── registry.yaml      # Target monitoring list (1 chain per project)
│   └── runtime.yaml        # Runtime configuration (Redis, RPC, database)
├── rules/                   # Your security rules
├── abi/                     # Smart contract ABIs
└── cmd/
    └── watcher/
        └── main.go          # Runtime entry point (you build this)
```

### Initialize the Runtime

In your `cmd/watcher/main.go`:

```go
package main

import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/go-redis/redis/v8"
)

func main() {
    // 1. Setup Redis (required - like Ponder.sh needs Postgres)
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 2. Setup engine
    engine := asentric.NewEngine()
    engine.RegisterRule(&LargeSwapRule{})
    engine.RegisterRule(&SuspiciousTransferRule{})
    
    // 3. Setup Database (optional - for saving events/logs)
    // var db *gorm.DB
    // if config.Database.Enabled { ... }
    
    // 4. Start monitoring (you implement this)
    watcher := NewWatcher(engine, redisClient)
    watcher.Start()
}
```

---

## Writing Security Rules

Security rules implement a simple interface:

```go
type Rule interface {
    Name() string
    Evaluate(ctx Context) (*Alert, error)
}
```

### Example: Large Swap Detection

```go
package rules

import (
    "github.com/asentric/asentric-sdk/pkg/asentric"
)

type LargeSwapRule struct{}

func (r *LargeSwapRule) Name() string {
    return "large_swap_detection"
}

func (r *LargeSwapRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // Access transaction data from context
    tx := ctx.Tx()
    
    // Define detection logic
    threshold := big.NewInt(1000)
    if tx.Value().Cmp(threshold) > 0 {
        return &asentric.Alert{
            Severity:    asentric.High,
            Title:       "Large Swap Detected",
            Description: "Transaction value exceeds threshold",
            Metadata: map[string]interface{}{
                "value":     tx.Value().String(),
                "threshold": threshold.String(),
            },
            // Note: Ref (tx_hash, block_number) is populated by the engine, not by rules
        }, nil
    }
    
    // No alert — transaction is normal
    return nil, nil
}
```

### Key Principles for Rules

1. **Pure Functions** — No side effects, no external I/O
2. **Context-Based Data** — All transaction data comes from `Context`
3. **Explicit Returns** — Return `nil` when no alert is needed
4. **Structured Alerts** — Use the `Alert` struct for consistency
5. **Error Handling** — Return errors for processing failures, not detection misses

### Alert Structure

Alerts may include a minimal execution reference (transaction hash and block number) for debugging and traceability. This reference is:

* **Populated by the engine**, not by rules
* **Informational only** — does not imply routing, persistence, or delivery responsibility
* **Optional** — rules do not need to (and cannot) set it

The `ExecutionRef` contains only `tx_hash` and `block_number`. It does not include chain identity, network information, or RPC endpoints. Chain identity remains the responsibility of runtime systems.

---

## Testing Rules

Rules are easy to test because they're pure functions with no external dependencies:

```go
package rules

import (
    "math/big"
    "testing"
    
    "github.com/asentric/asentric-sdk/pkg/asentric"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLargeSwapRule_TriggersOnLargeValue(t *testing.T) {
    // Arrange
    ctx := mockContextWithValue(big.NewInt(2000))
    rule := &LargeSwapRule{}
    
    // Act
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    require.NoError(t, err)
    require.NotNil(t, alert)
    assert.Equal(t, asentric.High, alert.Severity)
}

func TestLargeSwapRule_NoAlertOnSmallValue(t *testing.T) {
    // Arrange
    ctx := mockContextWithValue(big.NewInt(500))
    rule := &LargeSwapRule{}
    
    // Act
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    require.NoError(t, err)
    assert.Nil(t, alert, "No alert should be generated for small values")
}

// Helper function to create test contexts
func mockContextWithValue(value *big.Int) asentric.Context {
    return asentric.NewMockContext(
        asentric.WithTxValue(value),
    )
}
```

**No mocks for Redis, HTTP, or databases are required** — just test your logic.

---

## Replay Mode

Asentric SDK supports **offline, deterministic replay** for debugging and testing.

### Running Replay

```bash
asentric replay --fixture fixtures/suspicious_tx.json
```

### Replay Guarantees

- **No External Dependencies** — Runs completely offline
- **Deterministic** — Same input always produces same output
- **Safe Iteration** — Test rule changes without affecting production
- **Historical Analysis** — Replay past transactions to validate detection logic

### Creating Replay Fixtures

Replay fixtures are JSON files containing transaction data:

```json
{
  "chain_id": 1,
  "block_number": 12345678,
  "tx_hash": "0xabc...",
  "from": "0x123...",
  "to": "0x456...",
  "value": "1000000000000000000",
  "data": "0x...",
  "logs": [...]
}
```

> **Important:** The SDK will never fetch historical data by itself. Fetching live transaction data from RPC nodes is the responsibility of runtime systems (e.g., `asentric-bot`), not the SDK.

---

## Infrastructure Requirements

### Required: Redis (Setup Awal)

**Like Ponder.sh requires Postgres, Asentric requires Redis for:**
- ✅ Message queue (watcher → processor)
- ✅ State management (processed blocks)
- ✅ Worker coordination (multi-worker)
- ✅ Alert queue (processor → alert handler)

**Setup:**
```bash
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

### Optional: Database (Save Events/Logs)

**You can choose a database for:**
- ⚠️ Saving events/transactions/logs
- ⚠️ Historical data storage
- ⚠️ Analytics & reporting

**Options:**
- PostgreSQL (relational)
- MongoDB (document)
- InfluxDB (time-series)
- ClickHouse (analytics)

**See full documentation:** `docs/infrastructure-requirements.md`

---

## Ecosystem Integration

The Asentric SDK is designed to be embedded into multiple runtime environments:

```
┌─────────────────────────────────────────────────┐
│              Asentric Framework                  │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Engine  │  │  Rules   │  │ Context  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│         │             │             │           │
│         └─────────────┴─────────────┘           │
│                     │                           │
└─────────────────────┼───────────────────────────┘
                      │
         ┌────────────┼────────────┐
         │            │            │
    ┌────▼───┐   ┌───▼────┐   ┌──▼──────┐
    │Runtime │   │Backend │   │Frontend │
    │(Self-  │   │  (API) │   │  (UI)   │
    │hosted) │   │        │   │         │
    └────────┘   └────────┘   └─────────┘
```

### Component Roles

| Component | Role | Uses SDK For |
|-----------|------|--------------|
| **SDK** | Security logic & detection | N/A (core library) |
| **Runtime** | Self-hosted watcher (developer builds) | Rule execution, alert generation |
| **Backend** | API, aggregation, persistence | Alert processing, historical analysis |
| **Frontend** | Visualization & dashboard | N/A (consumes Backend API) |
| **CLI** | Development & testing tools (not a runtime) | Replay, rule validation |
| **Reference Runtime** | Example implementation | Reference for developers (not required) |

**The SDK remains the single source of truth for security detection logic.**

**Note:** Developers build their own runtime (self-hosted). The reference runtime (`cmd/runtime-reference/`) is provided as an example, not a requirement.

---

## Repository Structure

```
asentric/ (or asentric-sdk/)
├── pkg/
│   └── asentric/              # PUBLIC SDK API (STABLE)
│       ├── engine.go          # Detection engine
│       ├── rule.go            # Rule interface
│       ├── context.go         # Execution context
│       ├── alert.go           # Alert model
│       └── ...
│
├── internal/                  # PRIVATE SDK IMPLEMENTATION
│   ├── runtime/               # Engine runtime loop
│   ├── rule/                  # Rule registry & executor
│   ├── chain/                 # Chain data models & helpers
│   ├── abi/                   # ABI loading & decoding (internal only)
│   ├── alert/                 # Alert formatting & envelope
│   └── observability/         # Internal execution metrics & diagnostics
│
├── cmd/
│   ├── asentric/              # CLI tools (init, replay, version)
│   │   ├── init.go
│   │   ├── replay.go
│   │   └── version.go
│   │
│   └── runtime-reference/    # Reference Runtime (Example)
│       └── main.go            # Example runtime implementation
│
├── templates/
│   └── project/               # Project templates for `asentric init`
│       ├── config/
│       │   ├── asentric.yaml
│       │   ├── registry.yaml
│       │   └── runtime.yaml
│       ├── rules/
│       └── cmd/watcher/
│
├── examples/
│   ├── simple-watcher/        # Minimal runtime example
│   ├── custom-rules/          # Example custom rules
│   ├── ml-integration/        # Example ML rule
│   └── multi-chain/          # Example multi-chain setup
│
├── docs/
│   ├── architecture.md                    # Architecture deep dive
│   ├── sdk-api.md                        # Complete API reference
│   ├── final-architecture-recommendation.md  # Final architecture decision
│   ├── developer-experience.md           # Developer experience guide
│   ├── infrastructure-requirements.md   # Infrastructure setup guide
│   └── migration-roadmap.md              # Migration strategy
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### Key Directories

- **`pkg/asentric/`** — Public, stable SDK API used by developers
- **`internal/`** — Private implementation details, subject to change
- **`cmd/asentric/`** — CLI tools for scaffolding and testing (not a runtime)
- **`cmd/runtime-reference/`** — Reference runtime example (not required, just a reference)
- **`templates/`** — Project templates for quick start
- **`examples/`** — Working examples demonstrating SDK usage

**Observability Note:** Observability in the SDK is limited to internal execution metrics and diagnostics (e.g., rule execution timing, performance counters). Exporting metrics or logs to external systems (Prometheus, OpenTelemetry, etc.) is the responsibility of the runtime.

---

## Comparison with Ponder.sh

| Aspect | Ponder.sh | Asentric |
|--------|-----------|----------|
| **Repository** | 1 repo (monorepo) | ✅ 1 repo (monorepo) |
| **Framework** | ✅ Core framework | ✅ Core framework |
| **CLI** | ✅ CLI tools | ✅ CLI tools |
| **Examples** | ✅ Examples | ✅ Examples |
| **Runtime** | Managed (Ponder) | Self-hosted (developer) |
| **Required Infrastructure** | Postgres (setup awal) | Redis (setup awal) |
| **Optional Infrastructure** | Database untuk data | Database untuk events/logs |
| **Deployment** | Push to Ponder | Self-hosted |
| **Open Source** | ✅ | ✅ |

**Similarities:**
- ✅ 1 repo (monorepo)
- ✅ Framework + CLI + Examples
- ✅ Developer experience focus
- ✅ Required infrastructure untuk setup awal
- ✅ Optional database untuk data storage

**Differences:**
- ⚠️ Ponder.sh: Managed infrastructure
- ⚠️ Asentric: Self-hosted (developer choice)
- ⚠️ Ponder.sh: Postgres required (state management)
- ⚠️ Asentric: Redis required (message queue & state management)

---

## Roadmap

We're continuously improving Asentric SDK. Upcoming features include:

- [ ] **Rule Grouping & Tagging** — Organize rules by protocol, risk level, or category
- [ ] **Multi-Chain Support** — Built-in support for EVM-compatible chains
- [ ] **Enhanced ABI Decoding** — Automatic event decoding and type safety
- [ ] **Advanced Replay Features** — Time-travel debugging and batch replay
- [ ] **Performance Profiling** — Built-in rule performance analysis

### Long-Term Research Ideas

The following features are under research and may be explored in future versions:

- [ ] **ZK-Friendly Rule Outputs** — Generate zero-knowledge-compatible alert proofs
- [ ] **WASM-Based Rule Sandbox** — Run untrusted rules in isolated WebAssembly environment

---

## Contributing

Contributions are welcome! We especially appreciate help with:

- **Rule Examples** — Share your detection logic with the community
- **ABI Decoding** — Improve support for complex event types
- **Developer Tooling** — Enhance CLI features and DX improvements
- **Documentation** — Help make the SDK more accessible
- **Testing** — Add test coverage and edge case scenarios

### Getting Started

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-rule`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing detection rule'`)
6. Push to your branch (`git push origin feature/amazing-rule`)
7. Open a Pull Request

### Code Standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write tests for new features
- Update documentation as needed
- Keep rules pure and deterministic

---

## License

MIT License © Asentric

See [LICENSE](LICENSE) for full details.

---

## Resources

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issue Tracker**: [GitHub Issues](https://github.com/asentric/asentric-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/asentric/asentric-sdk/discussions)

---

**Built with ❤️ by the Asentric Team**