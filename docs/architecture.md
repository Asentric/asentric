# Architecture – Asentric SDK

## 1. Purpose & Scope

Asentric SDK is a **pure security detection engine** for blockchain systems.

Its sole responsibility is to:

* Execute deterministic security rules
* Process structured on-chain data
* Produce structured alerts

The SDK intentionally **does not**:

* Fetch blockchain data
* Run long-lived watchers
* Connect to Redis, Kafka, databases, or HTTP APIs
* Deliver alerts
* Manage configuration, deployment, or infrastructure

These concerns are delegated to external systems (e.g. `asentric-bot`, `asentric-backend`).

The SDK is designed to be **embedded**, not deployed.

---

## 2. Architectural Philosophy

### 2.1 Pure Domain Core

At its core, Asentric SDK follows a **pure domain logic** philosophy:

* Rules are pure functions
* No side effects during rule evaluation
* No global state
* No I/O inside the engine

This guarantees:

* Deterministic execution
* Easy testing
* Safe replay
* Predictable behavior

> **Note:** Conceptually, the SDK behaves like a pure function: given the same Context and rule set, it always produces the same alerts. This mathematical property enables deterministic replay and testing.

---

### 2.2 Infrastructure Inversion

The SDK does not depend on infrastructure.

Instead:

* Infrastructure depends on the SDK
* Runtime systems provide data **into** the SDK
* The SDK produces results **outward**

This is a strict, one-way dependency.

```
Infrastructure → SDK → Alerts
```

This inversion prevents infrastructure concerns from leaking into security logic.

---

### 2.3 Explicit Context Boundary

All execution data flows through a single object: **Context**.

Context is:

* Explicit
* Immutable during evaluation
* Fully controlled by the runtime

There is:

* No hidden state
* No global variables
* No implicit dependencies

This makes execution traceable, debuggable, and replayable.

---

## 3. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Asentric Framework Repository                    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Framework Core (pkg/asentric)                │   │
│  │         - Engine                                     │   │
│  │         - Rules                                      │   │
│  │         - Context                                    │   │
│  │         - Alerts                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│                          │ Used by                           │
│                          ▼                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         CLI Tools (cmd/asentric)                      │   │
│  │         - init: Generate project                     │   │
│  │         - replay: Test offline                       │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Reference Runtime (cmd/runtime-reference)    │   │
│  │         - Example implementation                     │   │
│  │         - Not required, hanya contoh                 │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Examples (examples/)                         │   │
│  │         - simple-watcher                             │   │
│  │         - custom-rules                               │   │
│  │         - ml-integration                             │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ Used by
                          ▼
┌─────────────────────────────────────────────────────────────┐
│         Developer Project (Self-Hosted)                     │
│                                                              │
│  my-protocol-monitor/                                       │
│  ├── config/                                                │
│  │   ├── asentric.yaml      # Engine config                │
│  │   ├── registry.yaml      # Target list (1 chain)         │
│  │   └── runtime.yaml        # Runtime config (Redis, DB)   │
│  ├── rules/                                                 │
│  │   ├── custom_rule.go     # Developer rules             │
│  │   └── ml_rule.go         # Custom ML rule               │
│  ├── abi/                                                   │
│  └── cmd/watcher/                                           │
│      └── main.go            # Runtime (developer buat)     │
│                                                              │
│  Runtime Responsibilities:                                  │
│  - Setup Redis (required - message queue & state)           │
│  - Connect to RPC (developer choice)                        │
│  - Setup Database (optional - untuk save events/logs)      │
│  - Parse config & registry                                  │
│  - Setup engine                                             │
│  - Register rules                                           │
│  - Monitoring loop                                          │
│  - Alert delivery (developer choice)                         │
└─────────────────────────────────────────────────────────────┘
```

The SDK provides the **core detection logic**, while external systems handle:

* **Runtime** (Self-hosted): Chain monitoring, transaction ingestion, alert routing (developer builds)
* **Backend**: REST API, data aggregation, persistence
* **Frontend**: User interface, dashboards, visualization

The SDK is a **closed execution box** that processes Context and produces Alerts.

**Note:** Developers build their own runtime (self-hosted). The reference runtime (`cmd/runtime-reference/`) is provided as an example, not a requirement.

---

## 4. Core Components

### 4.1 Engine

The Engine is responsible for:

* Managing rule registration
* Iterating over rules
* Executing rules against a given Context
* Collecting produced alerts

The Engine:

* Has no knowledge of chains, RPCs, or infrastructure
* Does not manage concurrency or scheduling
* Executes rules sequentially and deterministically

Conceptually:

```
for each rule:
  alert = rule.Evaluate(ctx)
  collect(alert)
```

---

### 4.2 Rule System

Rules represent **security knowledge**.

A rule:

* Encapsulates one detection idea
* Is stateless
* Is deterministic

Rules:

* Never perform I/O
* Never mutate external state
* Never communicate with other rules

Rules are **isolated units of logic**.

This isolation guarantees:

* Safety
* Parallel reasoning
* Simple testing

---

### 4.3 Context

Context is the **single source of truth** during execution.

It contains:

* Transaction data
* Block metadata
* Decoded logs / events
* Chain-specific information

Context:

* Is constructed by the runtime
* Is passed into the SDK
* Is read-only during evaluation

> **Note:** The SDK assumes Context is valid and complete. Context validation and enrichment (e.g., fetching missing data, decoding events) is the runtime's responsibility. The SDK does not validate or enrich Context.

---

### 4.4 Alerts

Alerts are the **only output** of the SDK.

An alert is:

* Structured
* Serializable
* Infrastructure-agnostic

Alerts do not:

* Know where they will be sent
* Know how they will be stored
* Contain delivery logic

The SDK produces alerts — it does not deliver them.

#### Execution Reference

Alerts may include an optional `ExecutionRef` containing:

* Transaction hash (`tx_hash`)
* Block number (`block_number`)

**Important distinctions:**

* **Alert ≠ delivery envelope** — Alerts are semantic signals, not infrastructure messages
* **ExecutionRef ≠ infrastructure metadata** — It contains only execution traceability, not chain identity, network, or RPC endpoints
* **Chain identity remains external** — Runtime systems are responsible for chain context (chain ID, network name, etc.)

The `ExecutionRef` is:

* Populated by the engine (not by rules)
* Not accessible or modifiable by rules
* Optional and informational only
* Can be overridden or enriched by runtime systems

This design maintains:

* **Rule purity** — Rules remain unaware of execution context
* **Alert semantics** — Alerts stay focused on security signals
* **Runtime ownership** — Runtime controls all infrastructure and chain context

---

## 5. Package Structure & Boundaries

### 5.1 Public API (`pkg/asentric`)

This is the **only supported integration surface**.

Contains:

* Engine
* Rule interface
* Context interface
* Alert model

Stability guarantees:

* Backward compatibility
* Semver-managed changes

If it lives in `pkg/asentric`, it is safe to depend on.

---

### 5.2 Internal Implementation (`internal/`)

Everything under `internal/` is:

* Private
* Non-stable
* Free to change

Contains:

* Rule execution internals
* Runtime helpers
* ABI decoding logic
* Internal observability

External systems must never import from `internal/`.

---

### 5.3 CLI (`cmd/asentric`)

The CLI is a **developer tool**, not a runtime.

It is used for:

* Project scaffolding
* Rule testing
* Offline replay

The CLI:

* Does not connect to RPC nodes
* Does not run production watchers
* Does not manage long-lived processes

It exists purely to improve developer experience.

---

## 6. Runtime Responsibility Matrix

| Responsibility   | SDK          | Runtime |
| ---------------- | ------------ | ------- |
| Fetch chain data | ❌            | ✅       |
| Decode ABI       | ⚠️ (internal helpers only) | ✅ (full decoding) |
| Rule execution   | ✅            | ❌       |
| Alert creation   | ✅            | ❌       |
| Alert delivery   | ❌            | ✅       |
| Persistence      | ❌            | ✅       |
| Scheduling       | ❌            | ✅       |
| Scaling          | ❌            | ✅       |

> **Note:** The SDK provides ABI decoding helpers in `internal/abi/` for internal use, but full ABI decoding and event parsing is the runtime's responsibility. The SDK uses these helpers internally but does not expose them as part of the public API.

The SDK never crosses its boundary.

---

## 7. Determinism & Replay

Determinism is a core invariant.

Given:

* The same Context
* The same rule set

The SDK guarantees:

* Identical outputs
* No hidden state
* No time-based behavior

Replay works by:

* Reconstructing Context from fixtures
* Running the engine offline

The SDK **never fetches historical data**.

---

## 8. Observability Scope

Observability inside the SDK is **strictly internal**.

Includes:

* Rule execution timing
* Rule evaluation counts
* Engine diagnostics

Excludes:

* Metrics exporters
* Tracing backends
* Logging pipelines

Exporting observability data is the runtime's responsibility.

---

## 9. Ecosystem Integration

The Asentric SDK is designed to be embedded into multiple runtime environments:

```
┌─────────────────────────────────────────────────────────────┐
│              Asentric Framework Repository                    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Framework Core (pkg/asentric)                │   │
│  │         - Engine                                     │   │
│  │         - Rules                                      │   │
│  │         - Context                                    │   │
│  │         - Alerts                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│                          │ Used by                           │
│                          ▼                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         CLI Tools (cmd/asentric)                      │   │
│  │         - init: Generate project                     │   │
│  │         - replay: Test offline                       │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Reference Runtime (cmd/runtime-reference)    │   │
│  │         - Example implementation                     │   │
│  │         - Not required, hanya contoh                 │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Examples (examples/)                         │   │
│  │         - simple-watcher                             │   │
│  │         - custom-rules                               │   │
│  │         - ml-integration                             │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ Used by
                          ▼
┌─────────────────────────────────────────────────────────────┐
│         Developer Project (Self-Hosted)                     │
│                                                              │
│  Runtime Responsibilities:                                  │
│  - Setup Redis (required - message queue & state)           │
│  - Connect to RPC (developer choice)                        │
│  - Setup Database (optional - untuk save events/logs)      │
│  - Parse config & registry                                  │
│  - Setup engine                                             │
│  - Register rules                                           │
│  - Monitoring loop                                          │
│  - Alert delivery (developer choice)                         │
└─────────────────────────────────────────────────────────────┘
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

The SDK remains the **single source of truth** for security detection logic.

**Note:** Developers build their own runtime (self-hosted). The reference runtime (`cmd/runtime-reference/`) is provided as an example, not a requirement.

---

## 10. Infrastructure Requirements

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

## 11. Architectural Non-Goals

The following will **never** be added to the SDK:

* RPC clients
* HTTP servers
* Alert delivery
* Deployment tooling

**Note:** While Redis is required for runtime setup (like Ponder.sh requires Postgres), the SDK itself does not include Redis clients. Redis is used by runtime systems (developer-built) for message queue and state management.

If a feature requires infrastructure, it does not belong here.

---

## 12. Summary

Asentric SDK is:

* A pure execution engine
* Infrastructure-agnostic
* Deterministic and testable
* Designed for embedding
* **1 Repository (Monorepo)** - Framework + CLI + Examples + Reference Runtime
* **Self-hosted** - Developer builds runtime
* **Redis required** - For setup (like Ponder.sh needs Postgres)
* **Database optional** - For saving events/logs

It exists to make **security logic simple, safe, and reusable**.

Everything else belongs elsewhere.
