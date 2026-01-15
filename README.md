# Asentric SDK

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**A Go SDK for building real-time blockchain security monitoring with custom detection rules.**

Asentric SDK provides a pure execution engine, rule system, and explicit runtime context that enable developers to write smart contract security rules without coupling to infrastructure concerns.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Writing Rules](#writing-rules)
- [API Reference](#api-reference)
- [Infrastructure](#infrastructure)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Asentric SDK is a framework for real-time smart contract security monitoring that enables developers to:

- Define monitoring targets via YAML configuration
- Write custom detection logic in Go
- Run self-hosted watchers
- Receive alerts via webhook or console in real-time

Asentric is not a SaaS platform or YAML-based rule engine. It is an SDK with a runtime pattern similar to [Ponder.sh](https://ponder.sh), designed for developers who want full control over their security monitoring infrastructure.

---

## Features

| Feature | Description |
|---------|-------------|
| **Pure Detection Rules** | Write detection logic as pure functions with no side effects |
| **Real-time Monitoring** | WebSocket-based event streaming from any EVM chain |
| **Deterministic Execution** | Same input always produces same output, enabling replay and testing |
| **Zero Dependencies** | No external infrastructure required for basic usage |
| **Flexible Alerting** | Console output for development, webhook for production |
| **EVM Compatible** | Works with any EVM-compatible blockchain |

---

## Quick Start

### Prerequisites

- Go 1.22 or higher

### Installation

```bash
go install github.com/asentric/asentric@latest
```

### Create a New Project

```bash
asentric init my-watcher
cd my-watcher
go mod tidy
```

This generates a ready-to-run project structure:

```
my-watcher/
├── config/
│   ├── asentric.yaml      # Runtime configuration
│   └── registry.yaml      # Target contracts
├── rules/
│   └── example_rule.go    # Detection rule
├── abi/                   # Contract ABI files
├── cmd/
│   └── watcher/
│       └── main.go        # Entry point
├── go.mod
└── README.md
```

### Run

```bash
go run cmd/watcher/main.go
```

Expected output:

```
===========================================
  my-watcher - Asentric Watcher
===========================================
[OK] Rules registered
Connecting to Mantle Sepolia...
[OK] Runtime ready
-------------------------------------------
Chain:  Mantle Sepolia (ID: 5003)
Source: websocket
Sink:   console
-------------------------------------------
Listening for events... (Press Ctrl+C to stop)
```

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
    │Runtime │   │Backend │   │Frontend │
    │(Watcher)│  │  (API) │   │  (UI)   │
    └────────┘   └────────┘   └─────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| **SDK** | Core detection engine, rule execution, context management |
| **Runtime** | Chain connection, event ingestion, alert delivery |
| **Backend** | REST API, data persistence, aggregation (optional) |
| **Frontend** | Dashboard, visualization (optional) |

---

## Configuration

### Runtime Configuration

**config/asentric.yaml**

```yaml
version: "1.0"

chain:
  id: 5003
  name: "Mantle Sepolia"
  rpcUrl: "https://rpc.sepolia.mantle.xyz"
  rpcWs: "wss://mantle-sepolia.drpc.org"

source:
  type: "websocket"
  url: "wss://mantle-sepolia.drpc.org"

sink:
  type: "console"      # Options: console, webhook
  url: ""              # Required if type is webhook

debug: true
```

### Target Registry

**config/registry.yaml**

```yaml
targets:
  - address: "0xYourContractAddress"
    name: "Token Contract"
    abi_path: "abi/erc20.json"
```

### Supported Networks

| Network | Chain ID | WebSocket RPC |
|---------|----------|---------------|
| Mantle Sepolia (default) | 5003 | `wss://mantle-sepolia.drpc.org` |
| Base Sepolia | 84532 | `wss://base-sepolia.drpc.org` |
| Ethereum Sepolia | 11155111 | `wss://eth-sepolia.drpc.org` |

---

## Writing Rules

Rules implement the `asentric.Rule` interface:

```go
package rules

import (
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/utils"
)

type LargeTransferRule struct {
    Threshold *big.Int
}

func NewLargeTransferRule() *LargeTransferRule {
    threshold := new(big.Int)
    threshold.SetString("1000000000000000000", 10) // 1 token
    return &LargeTransferRule{Threshold: threshold}
}

func (r *LargeTransferRule) Name() string {
    return "large-transfer"
}

func (r *LargeTransferRule) Severity() asentric.Severity {
    return asentric.SeverityHigh
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        if log.Event.Name == "Transfer" {
            value := utils.GetFieldBigInt(log.Event.Fields, "value")
            from := utils.GetFieldString(log.Event.Fields, "from")
            
            if value != nil && value.Cmp(r.Threshold) > 0 {
                isMint := utils.IsZeroAddress(from)
                title := "Large Transfer Detected"
                if isMint {
                    title = "Token Mint Detected"
                }
                
                return asentric.NewAlert(r.Name(), title, r.Severity()).
                    WithMetadata("value", value.String()).
                    WithMetadata("isMint", isMint), nil
            }
        }
    }
    return nil, nil
}
```

### Registering Rules

```go
// cmd/watcher/main.go
engine := asentric.NewEngine()
engine.RegisterRule(rules.NewLargeTransferRule())
```

---

## API Reference

### Core Types

| Type | Description |
|------|-------------|
| `asentric.Engine` | Rule execution engine |
| `asentric.Rule` | Interface for detection rules |
| `asentric.Context` | Execution context with transaction and log data |
| `asentric.Alert` | Structured alert output |
| `asentric.Severity` | Alert severity level (Low, Medium, High, Critical) |

### Utility Functions

```go
import "github.com/asentric/asentric/pkg/utils"

// Extract fields from event data
utils.GetFieldString(fields, "from")
utils.GetFieldBigInt(fields, "value")

// Address utilities
utils.IsZeroAddress(addr)      // Check if mint/burn
utils.TruncateAddress(addr)    // "0x1234...5678"

// Token formatting
utils.FormatTokenAmount(wei, decimals)
```

---

## Infrastructure

### Minimal Setup (Default)

The SDK works with zero external dependencies:

- WebSocket RPC endpoint (free from dRPC, Infura, Alchemy)
- Console sink for development
- In-memory queue

### Production Setup

For production deployments:

| Component | Purpose |
|-----------|---------|
| Webhook Backend | Receive and store alerts |
| PostgreSQL | Historical data and analytics |
| Redis | Message queue for multi-worker setups |

---

## Documentation

| Document | Description |
|----------|-------------|
| [SPEC.md](docs/SPEC.md) | MVP Specification |
| [architecture.md](docs/architecture.md) | System architecture |
| [developer-overview.md](docs/developer-overview.md) | Developer guide |
| [sdk-api.md](docs/sdk-api.md) | API reference |

---

## Project Structure

```
asentric-sdk/
├── pkg/asentric/          # Public SDK API
│   ├── engine.go          # Rule execution engine
│   ├── rule.go            # Rule interface
│   ├── context.go         # Execution context
│   ├── alert.go           # Alert model
│   └── config.go          # Configuration
├── pkg/runtime/           # Runtime builder
├── pkg/utils/             # Utility functions
├── internal/              # Private implementation
├── cmd/asentric/          # CLI tools
├── templates/             # Project templates
└── examples/              # Usage examples
```

---

## Testing

Rules are pure functions, making them easy to test:

```go
func TestLargeTransferRule(t *testing.T) {
    ctx := mockContextWithTransfer(big.NewInt(2e18))
    rule := NewLargeTransferRule()
    
    alert, err := rule.Evaluate(ctx)
    
    require.NoError(t, err)
    require.NotNil(t, alert)
    assert.Equal(t, asentric.SeverityHigh, alert.Severity)
}
```

Run tests:

```bash
go test ./...
```

---

## Contributing

Contributions are welcome. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Submit a pull request

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Links

- [GitHub Repository](https://github.com/asentric/asentric)
- [Documentation](docs/)
- [Examples](examples/)
- [Issue Tracker](https://github.com/asentric/asentric/issues)