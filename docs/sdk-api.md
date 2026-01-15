# Asentric SDK - API Reference

**Status:** Stable (v1 Contract)  
**Audience:** SDK Users, Rule Authors, Runtime Implementers

---

## Table of Contents

1. [Overview](#overview)
2. [Package Structure](#package-structure)
3. [Core Types](#core-types)
4. [Engine](#engine)
5. [Rule Interface](#rule-interface)
6. [Context Interface](#context-interface)
7. [Alert](#alert)
8. [Severity Levels](#severity-levels)
9. [EventSource Interface](#eventsource-interface)
10. [AlertSink Interface](#alertsink-interface)
11. [Utility Functions](#utility-functions)
12. [Error Types](#error-types)
13. [Extension Points](#extension-points)

---

## Overview

The Asentric SDK provides a pure, deterministic execution engine for blockchain security detection. This document defines the complete public API.

**Stability Guarantee:**

| Package | Stability |
|---------|-----------|
| `pkg/asentric/*` | Stable |
| `pkg/utils/*` | Stable |
| `pkg/runtime/*` | Stable |
| `internal/*` | Private, no guarantees |

---

## Package Structure

```go
import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/utils"
    "github.com/asentric/asentric/pkg/runtime"
)
```

---

## Core Types

| Type | Description |
|------|-------------|
| `Engine` | Rule execution engine |
| `Rule` | Detection logic interface |
| `Context` | Execution context with transaction data |
| `Alert` | Rule output |
| `Severity` | Alert severity level |
| `EventSource` | Event ingestion interface |
| `AlertSink` | Alert delivery interface |

---

## Engine

The engine orchestrates rule evaluation.

### Constructor

```go
func NewEngine() *Engine
```

### Methods

```go
// RegisterRule adds a rule to the engine
func (e *Engine) RegisterRule(rule Rule) error

// Evaluate runs all registered rules against the context
func (e *Engine) Evaluate(ctx Context) ([]*Alert, error)
```

### Behavior

- Rules execute sequentially in registration order
- Same input produces same output (deterministic)
- No side effects or I/O operations
- Recovers from rule panics

---

## Rule Interface

Rules implement detection logic as pure functions.

```go
type Rule interface {
    // Name returns unique rule identifier
    Name() string
    
    // Severity returns the alert severity level
    Severity() Severity
    
    // Evaluate runs detection logic
    // Returns (alert, nil) if detection triggered
    // Returns (nil, nil) if no detection
    // Returns (nil, error) if execution failed
    Evaluate(ctx Context) (*Alert, error)
}
```

### Example Implementation

```go
type TransferRule struct {
    Threshold *big.Int
}

func (r *TransferRule) Name() string {
    return "large-transfer"
}

func (r *TransferRule) Severity() asentric.Severity {
    return asentric.SeverityHigh
}

func (r *TransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        if log.Event.Name == "Transfer" {
            value := utils.GetFieldBigInt(log.Event.Fields, "value")
            if value != nil && value.Cmp(r.Threshold) > 0 {
                return asentric.NewAlert(
                    r.Name(),
                    "Large Transfer Detected",
                    r.Severity(),
                ).WithMetadata("value", value.String()), nil
            }
        }
    }
    return nil, nil
}
```

### Rules for Rule Authors

**MUST:**
- Be deterministic (same input produces same output)
- Be side-effect free (no I/O, no network calls)
- Not mutate context
- Return `nil, nil` when no detection

**MUST NOT:**
- Perform I/O operations
- Access global state
- Depend on external services
- Store state between evaluations

---

## Context Interface

Context provides read-only access to transaction data.

```go
type Context interface {
    // ChainID returns the chain identifier
    ChainID() uint64
    
    // Tx returns transaction data
    Tx() Transaction
    
    // Block returns block metadata
    Block() Block
    
    // Logs returns decoded event logs
    Logs() []Log
}
```

### Transaction

```go
type Transaction struct {
    Hash     string
    From     Address
    To       Address
    Value    *big.Int
    Input    []byte
    Nonce    uint64
    GasPrice *big.Int
    GasLimit uint64
}
```

### Block

```go
type Block struct {
    Number    uint64
    Hash      string
    Timestamp uint64
    BaseFee   *big.Int
}
```

### Log

```go
type Log struct {
    Address     Address
    Topics      []string
    Data        []byte
    BlockNumber uint64
    TxHash      string
    TxIndex     uint
    LogIndex    uint
    Event       *DecodedEvent
}

type DecodedEvent struct {
    Name   string
    Fields map[string]interface{}
}
```

---

## Alert

Alert represents a detection result.

### Constructor

```go
func NewAlert(rule, title string, severity Severity) *Alert
```

### Methods

```go
// WithDescription adds a description
func (a *Alert) WithDescription(desc string) *Alert

// WithMetadata adds key-value metadata
func (a *Alert) WithMetadata(key string, value interface{}) *Alert
```

### Structure

```go
type Alert struct {
    Rule        string
    Title       string
    Description string
    Severity    Severity
    Timestamp   time.Time
    Metadata    map[string]interface{}
    Context     *AlertContext
}

type AlertContext struct {
    ChainID     uint64
    BlockNumber uint64
    TxHash      string
}
```

---

## Severity Levels

```go
const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)
```

---

## EventSource Interface

EventSource provides events to the runtime.

```go
type EventSource interface {
    // Start begins event streaming
    // Returns a channel that emits events
    // Channel closes when source stops
    Start(ctx context.Context) (<-chan Event, error)
}
```

### Built-in Sources

| Source | Description |
|--------|-------------|
| `WebSocketSource` | WebSocket RPC subscription |
| `MemorySource` | In-memory source for testing |

---

## AlertSink Interface

AlertSink handles alert delivery.

```go
type AlertSink interface {
    // Emit sends an alert
    Emit(ctx context.Context, alert *Alert) error
}
```

### Built-in Sinks

| Sink | Description |
|------|-------------|
| `ConsoleSink` | Prints to stdout |
| `WebhookSink` | HTTP POST to URL |
| `MultiSink` | Broadcasts to multiple sinks |

---

## Utility Functions

```go
import "github.com/asentric/asentric/pkg/utils"
```

### Field Extraction

```go
// GetFieldString extracts string from event fields
func GetFieldString(fields map[string]interface{}, key string) string

// GetFieldBigInt extracts *big.Int from event fields
func GetFieldBigInt(fields map[string]interface{}, key string) *big.Int

// GetFieldAddress extracts address from event fields
func GetFieldAddress(fields map[string]interface{}, key string) string
```

### Address Utilities

```go
// IsZeroAddress checks if address is zero (mint/burn indicator)
func IsZeroAddress(addr string) bool

// TruncateAddress shortens address for display
// "0x1234567890abcdef" -> "0x1234...cdef"
func TruncateAddress(addr string) string
```

### Formatting

```go
// FormatTokenAmount converts wei to token amount
// FormatTokenAmount("1000000", 6) -> "1"
func FormatTokenAmount(wei string, decimals int) string
```

---

## Error Types

```go
var (
    ErrInvalidContext = errors.New("invalid context")
    ErrInvalidEvent   = errors.New("invalid event")
    ErrRulePanic      = errors.New("rule panic")
    ErrAlreadyRunning = errors.New("runtime already running")
)
```

---

## Extension Points

### Allowed Extensions

| Extension | How |
|-----------|-----|
| Custom rules | Implement `Rule` interface |
| Custom sources | Implement `EventSource` interface |
| Custom sinks | Implement `AlertSink` interface |

### Forbidden Patterns

| Anti-pattern | Reason |
|--------------|--------|
| Mutate Context | Context is immutable |
| I/O in rules | Rules must be pure |
| Global state in rules | Rules must be stateless |
| Rule-to-rule communication | Rules are independent |

---

## Related Documentation

- [Quick Start](QUICK-START.md)
- [Developer Overview](developer-overview.md)
- [Architecture](architecture.md)
- [Project Structure](project-structure.md)
