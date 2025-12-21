# Asentric SDK – API Reference

This document defines the public, stable API contract of the Asentric SDK.

It is intended for:

* SDK users (rule authors)
* Runtime implementers (asentric-bot)
* Backend consumers
* Future contributors

> **Important:** Anything outside `pkg/asentric` is not part of the public API and may change without notice.

---

## API Stability Guarantees

| Package | Stability |
|---------|-----------|
| `pkg/asentric/*` | STABLE (v1 contract) |
| `cmd/asentric/*` | Best-effort (DX tooling) |
| `internal/*` | Private, no guarantees |

This document only covers public API.

---

## Core Concepts Overview

At a high level, the SDK consists of:

* **Engine** — orchestrates rule execution
* **Rule** — pure detection logic
* **Context** — immutable execution data
* **Alert** — semantic detection output
* **Severity** — strict classification enum

The SDK does not handle infrastructure concerns such as RPC, queues, databases, or delivery.

---

## Engine

### Purpose

The Engine is the central coordinator that:

* Holds registered rules (stateful)
* Executes rules sequentially for a given context
* Collects alerts produced by rules

The engine is stateful but not concurrency-safe.

### Engine Lifecycle

```go
engine := asentric.NewEngine()

engine.RegisterRule(ruleA)
engine.RegisterRule(ruleB)

alerts, err := engine.Process(ctx)
```

### Engine API

```go
type Engine interface {
    RegisterRule(rule Rule) error
    Process(ctx Context) ([]*Alert, error)
}
```

### Behavioral Guarantees

* Rules are executed sequentially
* Each rule is executed at most once per `Process` call
* Rules cannot affect each other's execution
* Rule execution order is deterministic (registration order)
* Engine maintains internal state (rule registry, config)
* Engine is **NOT** safe for concurrent use

### Error Semantics

| Situation | Behavior |
|-----------|----------|
| Rule returns `(nil, nil)` | No alert |
| Rule returns `(alert, nil)` | Alert collected |
| Rule returns `(nil, error)` | Engine returns error |
| Rule panics | Panic propagates |

Infrastructure layers decide retry / recovery strategies.

---

## Rule

### Purpose

A Rule represents a single unit of detection logic.

Rules are:

* Pure
* Deterministic
* Side-effect free
* Stateless

### Rule Interface

```go
type Rule interface {
    Name() string
    Evaluate(ctx Context) (*Alert, error)
}
```

### Rule Contract

Rules **MUST**:

* Not perform I/O
* Not mutate context
* Not depend on global state
* Return at most one alert
* Return `nil, nil` if no detection

Rules **MAY**:

* Perform computation
* Decode ABI data via context
* Return execution errors

### Error Handling Rules

| Condition | Expected Return |
|-----------|----------------|
| Detection not matched | `nil, nil` |
| Detection matched | `alert, nil` |
| Invalid input / decode failure | `nil, error` |

Errors represent execution failure, not detection outcome.

---

## Context

### Purpose

Context provides all execution data required by rules.

It is:

* Immutable
* Snapshot-based
* Explicit
* Deterministic

Rules cannot mutate context.

### Context Interface (Conceptual)

```go
type Context interface {
    ChainID() uint64
    Tx() Transaction
    Block() Block
    Logs() []Log
    ABI() ABIRegistry
}
```

> **Note:** Exact sub-interfaces (`Transaction`, `Block`, `Log`) are defined in SDK types.

### Context Guarantees

* No global access
* No hidden state
* Same input → same output
* Safe for replay and testing

> **Note:** The SDK assumes Context is valid and complete. Context validation and enrichment (e.g., fetching missing data, decoding events) is the runtime's responsibility. The SDK does not validate or enrich Context.

---

## Alert

### Purpose

An Alert represents a semantic security signal, not a delivery envelope.

### Execution Reference

Alerts may include an optional execution reference for debugging and traceability:

```go
type ExecutionRef struct {
    TxHash      string
    BlockNumber uint64
}
```

**Important:** The `ExecutionRef` is populated by the engine, not by rules. Rules cannot access or modify it.

### Alert Type

```go
type Alert struct {
    Rule        string
    Severity    Severity
    Title       string
    Description string
    Ref         *ExecutionRef  // optional, populated by engine
    Metadata    map[string]any
}
```

### Alert Design Rules

* Alerts are pure semantic outputs
* Metadata must be JSON-serializable
* Alerts do not imply delivery
* Alerts do not imply persistence
* `ExecutionRef` is informational only — does not include chain identity, network, or RPC endpoints

### What Alerts Do Not Include

Alerts do not include:

* Chain IDs
* Network information
* RPC endpoints
* Delivery metadata

> **Note:** Including `tx_hash` and `block_number` in `ExecutionRef` does not imply the SDK understands chain identity. Chain context remains the responsibility of runtime systems.

---

## Severity

### Purpose

Severity provides strict, normalized classification for alerts.

### Severity Enum

```go
type Severity int

const (
    Info Severity = iota
    Low
    Medium
    High
    Critical
)
```

### Severity Rules

* Enum is closed (no custom values)
* Ordering is meaningful
* Mapping to strings is responsibility of runtime / backend

---

## Configuration

### Purpose

SDK configuration controls engine behavior, not infrastructure.

Examples:

* Enabled rules
* Rule options
* Execution limits

### Config Principles

* Parsed outside rules
* Injected into engine
* Immutable during execution

> **Note:** Exact schema is documented in [architecture.md](architecture.md).

---

## CLI API (`cmd/asentric`)

### Scope

The CLI exists only for developer tooling.

It does not:

* Run watchers
* Connect to RPC nodes
* Deliver alerts
* Manage infrastructure

> **Important:** The CLI does not run production watchers or connect to blockchains. It is strictly a developer tool for scaffolding, replay, and rule validation.

### Supported Commands (v1)

```bash
asentric init <project>
asentric replay --fixture <file>
asentric version
```

### CLI Responsibilities

* Project scaffolding
* Replay testing
* Local development workflows

---

## Concurrency Model

* Engine instances are single-threaded
* Engine is not concurrency-safe
* Parallelism must be handled by runtime systems
* **Recommended:** one engine instance per worker

---

## Testing Guidelines

Rules are easy to test:

```go
ctx := asentric.NewMockContext(...)
rule := &MyRule{}

alert, err := rule.Evaluate(ctx)
```

No infrastructure mocks required.

---

## Non-Goals (Explicit)

Asentric SDK does **NOT**:

* Manage RPC connections
* Fetch blockchain data
* Store alerts
* Deliver notifications
* Provide APIs
* Handle deployment
* Understand chain identity (chain ID, network name)
* Manage network or RPC endpoint information

> **Important:** Including `tx_hash` and `block_number` in `ExecutionRef` does not imply the SDK understands chain identity. The SDK remains chain-agnostic. All chain context (chain ID, network, RPC endpoints) is the responsibility of runtime systems.

Those are handled by other repositories.

---

## Summary

Asentric SDK is:

* A pure security detection engine
* A stable contract for rule authors
* A shared brain across runtimes
* A strict boundary between logic and infrastructure

This API contract is the foundation for the entire Asentric ecosystem.
