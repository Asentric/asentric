# Asentric SDK

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**A Go SDK for building real-time, on-chain security detection logic in a modular, deterministic, and developer-friendly way.**

Asentric SDK provides a pure execution engine, rule system, and explicit runtime context that enable developers to write smart contract security rules without coupling to infrastructure concerns such as message queues, databases, APIs, or deployment systems.

---

## 📚 Documentation

> **🔒 MVP Specification:** [docs/SPEC.md](docs/SPEC.md) - **SINGLE SOURCE OF TRUTH** (locked)  
> **📚 Memahami Code:** [docs/UNDERSTAND-PKG.md](docs/UNDERSTAND-PKG.md) - Dokumentasi pkg/asentric dan pkg/domain  
> **🔧 Implementation Guide:** [docs/IMPL-GUIDE.md](docs/IMPL-GUIDE.md) - Step-by-step build guide  
> **🚀 Quick Start:** [docs/developer-overview.md](docs/developer-overview.md) - Alur end-to-end developer  
> **🏗️ Architecture:** [docs/architecture.md](docs/architecture.md) - Core architecture

---

## Overview

Asentric SDK adalah **framework untuk real-time smart contract security monitoring** yang memungkinkan developer:

* Mendefinisikan **apa yang dimonitor** melalui konfigurasi YAML
* Menulis **logic deteksi sendiri** melalui custom rules (Go)
* Menjalankan **self-hosted runtime**
* Menerima **alert via webhook** secara real-time

Asentric **bukan SaaS**, dan **bukan rule-engine berbasis YAML**. Asentric adalah **SDK + runtime pattern** dengan developer experience seperti Ponder.sh.

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
- Redis 7+ (required)
- WebSocket RPC endpoint (Alchemy, Infura, etc.)

### 1. Setup Redis

```bash
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

### 2. Install CLI

```bash
go install github.com/asentric/asentric@latest
```

### 3. Initialize Project

```bash
asentric init my-protocol-monitor
cd my-protocol-monitor
go mod tidy
```

This generates:

```
my-protocol-monitor/
├── config/
│   ├── asentric.yaml      # Runtime & engine config
│   └── registry.yaml      # Target monitoring list
├── rules/
│   └── example_rule.go    # Custom rules
├── abi/
│   └── .gitkeep
├── cmd/
│   └── watcher/
│       └── main.go
├── go.mod
└── README.md
```

### 4. Configure

**config/asentric.yaml:**
```yaml
chain:
  rpc_ws: "wss://rpc.mantle.xyz/ws"
  name: "Mantle"

redis:
  addr: "localhost:6379"

webhook:
  url: "https://your-webhook.com/alerts"
```

**config/registry.yaml:**
```yaml
targets:
  - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
    name: "Uniswap V3 Router"
    abi_path: "abi/uniswap_v3_router.json"
```

### 5. Write Custom Rules

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

func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    if tx.Value().Cmp(r.Threshold) > 0 {
        return asentric.NewAlert(
            r.Name(),
            asentric.SeverityHigh,
            "Large Transfer Detected",
            "Transaction value exceeds threshold",
        ).WithMetadata("value", tx.Value().String()).
          WithMetadata("threshold", r.Threshold.String()), nil
    }
    
    return nil, nil
}
```

### 6. Run

```go
// cmd/watcher/main.go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
    "my-project/rules"
)

func main() {
    config, _ := asentric.LoadConfig("config/")
    
    engine := asentric.NewEngine()
    engine.RegisterRule(&rules.LargeTransferRule{
        Threshold: big.NewInt(1e18), // 1 ETH
    })
    
    runtime := asentric.NewRuntime(config, engine)
    if err := runtime.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

```bash
go run cmd/watcher/main.go
```

> **📖 Full Guide:** [docs/developer-overview.md](docs/developer-overview.md)  
> **🔒 Specification:** [docs/SPEC.md](docs/SPEC.md)

---

## Filosofi Desain

Asentric dibangun dengan prinsip berikut:

* **YAML untuk konfigurasi, bukan logic** — Config hanya untuk setup engine & target list
* **Rules adalah code, bukan config** — Semua logic deteksi ditulis dalam Go
* **Engine deterministic & stateless** — Same input selalu menghasilkan same output
* **Runtime bertanggung jawab atas side-effect** — RPC, database, alert delivery
* **Developer bebas menentukan kompleksitas rules** — Dari simple threshold hingga ML integration
* **Redis required** (seperti Ponder.sh butuh Postgres) — Untuk message queue & state management
* **Database optional** (untuk save events/logs) — Developer choice
* **1 project = 1 chain** (chain agnostic, tapi fokus 1 chain)

Pendekatan ini membuat Asentric:
* Mudah dipelajari
* Mudah dites
* Mudah di-debug
* Tidak cepat mentok untuk use case kompleks

---

## Testing & Replay

Developer dapat test rules secara offline tanpa infrastructure:

```bash
asentric replay --fixture fixtures/example_tx.json
```

Replay mode:
* **No External Dependencies** — Runs completely offline
* **Deterministic** — Same input always produces same output
* **Safe Iteration** — Test rule changes without affecting production

Rules mudah di-test karena pure functions tanpa external dependencies:

```go
func TestLargeSwapRule_TriggersOnLargeValue(t *testing.T) {
    ctx := mockContextWithValue(big.NewInt(2000000000000000000)) // 2 ETH
    rule := &LargeSwapRule{}
    
    alert, err := rule.Evaluate(ctx)
    
    require.NoError(t, err)
    require.NotNil(t, alert)
    assert.Equal(t, asentric.High, alert.Severity)
}
```

**No mocks for Redis, HTTP, or databases are required** — just test your logic.

---


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

**Lihat detail:** [docs/developer-overview.md](docs/developer-overview.md#1-setup-infrastructure--instalasi)

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

> **📖 Lihat struktur detail:** [docs/project-structure.md](docs/project-structure.md) - Struktur folder & file final (authoritative)

```
asentric/
├── pkg/asentric/              # PUBLIC SDK (Stable API)
│   ├── engine.go              # Engine interface & implementation
│   ├── rule.go                # Rule interface
│   ├── context.go             # Context interface & core models
│   ├── alert.go               # Alert & Severity model
│   ├── config.go              # Engine-level config (non-infra)
│   ├── mock_context.go        # Test helpers (pure)
│   └── version.go
│
├── internal/                  # PRIVATE SDK implementation
│   ├── engine/                # Engine internals
│   ├── rule/                  # Rule helpers
│   ├── context/               # Concrete context implementations
│   ├── chain/                 # Chain data models (EVM, etc)
│   ├── abi/                   # ABI decoding helpers (INTERNAL)
│   └── alert/                 # Alert envelope helpers
│
├── cmd/
│   ├── asentric/              # CLI Tools
│   │   ├── main.go
│   │   ├── init.go            # asentric init
│   │   ├── replay.go          # offline deterministic replay
│   │   ├── version.go
│   │   └── internal/
│   │       └── templates.go
│   │
│   └── runtime-reference/     # Reference Runtime (Example)
│       ├── main.go             # Entry point runtime
│       ├── config/             # Load yaml config
│       ├── ingest/             # Subscribe logs/blocks
│       ├── pipeline/           # Dispatcher & workers
│       ├── state/              # RUNTIME STATE (Redis here)
│       ├── alert/              # Alert delivery
│       └── runtime.go          # Glue code
│
├── templates/
│   └── project/               # Project templates
│       ├── config/
│       │   ├── asentric.yaml
│       │   ├── registry.yaml
│       │   └── runtime.yaml
│       ├── rules/
│       ├── abi/
│       └── cmd/watcher/
│
├── examples/
│   ├── simple-watcher/        # Minimal runtime example
│   ├── custom-rules/          # Example custom rules
│   ├── ml-integration/        # Example ML rule
│   └── multi-chain/          # Example multi-chain setup
│
├── docs/
│   ├── SPEC.md                           # ⭐ MVP Specification (SINGLE SOURCE OF TRUTH)
│   ├── developer-overview.md             # Alur end-to-end developer
│   ├── architecture.md                    # Architecture deep dive
│   ├── sdk-api.md                        # Complete API reference
│   └── project-structure.md              # Final structure (authoritative)
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

- **🔒 MVP Specification**: [docs/SPEC.md](docs/SPEC.md) - **Single source of truth** (locked)
- **📚 Memahami Code**: [docs/UNDERSTAND-PKG.md](docs/UNDERSTAND-PKG.md) - Dokumentasi pkg/asentric dan pkg/domain
- **🔧 Implementation Guide**: [docs/IMPL-GUIDE.md](docs/IMPL-GUIDE.md) - Step-by-step build guide
- **🚀 Quick Start**: [docs/developer-overview.md](docs/developer-overview.md) - End-to-end developer guide
- **🏗️ Architecture**: [docs/architecture.md](docs/architecture.md) - Core architecture
- **📋 API Reference**: [docs/sdk-api.md](docs/sdk-api.md) - Public API specification
- **📁 Project Structure**: [docs/project-structure.md](docs/project-structure.md) - Folder structure
- **Examples**: [examples/](examples/)
- **Issue Tracker**: [GitHub Issues](https://github.com/asentric/asentric-sdk/issues)

---

**Built with ❤️ by the Asentric Team**