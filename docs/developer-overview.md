# Asentric SDK - Developer Overview

This document provides an end-to-end overview of using Asentric SDK from a developer perspective.

---

## Table of Contents

- [Purpose](#purpose)
- [Developer Flow](#developer-flow)
- [Installation](#installation)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Writing Rules](#writing-rules)
- [Running the Watcher](#running-the-watcher)
- [Testing](#testing)
- [Design Philosophy](#design-philosophy)

---

## Purpose

Asentric SDK is a framework for real-time smart contract security monitoring that enables developers to:

- Define monitoring targets via configuration
- Write custom detection logic in Go
- Run the engine locally or in any runtime environment
- Generate semantic, deterministic alerts

Asentric is not a SaaS platform or YAML-based rule engine. It is an SDK with a runtime pattern.

---

## Developer Flow

1. Install CLI
2. Initialize project
3. Configure targets (YAML)
4. Write custom rules (Go)
5. Run watcher
6. Receive alerts

---

## Installation

### Prerequisites

- Go 1.22 or higher

### Install CLI

```bash
go install github.com/asentric/asentric@latest
```

### Initialize Project

```bash
asentric init my-watcher
cd my-watcher
go mod tidy
```

---

## Project Structure

```
my-watcher/
├── config/
│   ├── asentric.yaml      # Runtime configuration
│   └── registry.yaml      # Target contracts
├── rules/
│   └── example_rule.go    # Detection rules
├── abi/                   # Contract ABI files
├── cmd/
│   └── watcher/
│       └── main.go        # Entry point
├── go.mod
└── README.md
```

**Separation of Concerns:**

| Directory | Purpose |
|-----------|---------|
| `config/` | Declarative configuration |
| `rules/` | Imperative detection logic |
| `cmd/` | Runtime orchestration |

---

## Configuration

### asentric.yaml

Runtime and engine configuration:

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
  url: ""              # Required for webhook

debug: true
```

### registry.yaml

Target contracts to monitor:

```yaml
targets:
  - address: "0xYourContractAddress"
    name: "Token Contract"
    abi_path: "abi/erc20.json"
```

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

func NewLargeTransferRule(threshold *big.Int) *LargeTransferRule {
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

### Rule Characteristics

- **Pure functions** - No side effects
- **No I/O** - No network calls or file access
- **Deterministic** - Same input produces same output
- **Testable** - Easy to unit test

### Registering Rules

```go
// cmd/watcher/main.go
engine := asentric.NewEngine()
engine.RegisterRule(rules.NewLargeTransferRule(big.NewInt(1e18)))
```

---

## Running the Watcher

### Entry Point

```go
// cmd/watcher/main.go
package main

import (
    "context"
    "log"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/runtime"
    "my-watcher/rules"
)

func main() {
    // Load configuration
    cfg, err := asentric.LoadConfig("config/asentric.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create engine and register rules
    engine := asentric.NewEngine()
    engine.RegisterRule(rules.NewLargeTransferRule())
    
    // Build and start runtime
    ctx := context.Background()
    rt, err := runtime.NewBuilder(cfg, engine).
        WithWebSocketSource(ctx).
        WithSinkFromConfig().
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Run (blocks until interrupt)
    if err := rt.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Run Command

```bash
go run cmd/watcher/main.go
```

### Expected Output

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

## Testing

Rules are pure functions, making them easy to test:

```go
func TestLargeTransferRule(t *testing.T) {
    ctx := mockContextWithTransfer(big.NewInt(2e18))
    rule := NewLargeTransferRule(big.NewInt(1e18))
    
    alert, err := rule.Evaluate(ctx)
    
    require.NoError(t, err)
    require.NotNil(t, alert)
    assert.Equal(t, "large-transfer", alert.Rule)
}
```

### Replay Mode

Test rules offline with recorded transactions:

```bash
asentric replay --fixture fixtures/example_tx.json
```

---

## Design Philosophy

| Principle | Description |
|-----------|-------------|
| **YAML for configuration** | Configuration is declarative, not logic |
| **Rules are code** | Detection logic written in Go, not config |
| **Deterministic engine** | Same input always produces same output |
| **Runtime handles I/O** | Side effects managed by runtime, not rules |
| **Zero dependencies** | No external infrastructure required for basic usage |
| **Single chain per project** | Focused monitoring, chain agnostic |

---

## Next Steps

- [Architecture](architecture.md) - System architecture deep dive
- [SDK API Reference](sdk-api.md) - Complete API documentation
- [Testing Guide](TESTING-GUIDE.md) - Testing strategies
