# Asentric SDK – Public API Specification

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH** untuk hackathon  
> **📖 Lihat alur developer:** [developer-overview.md](developer-overview.md) - Alur end-to-end penggunaan Asentric SDK  
> **🏗️ Lihat architecture:** [architecture.md](architecture.md) - Core architecture

**Status:** ✅ **AUTHORITATIVE – CONTRACT LOCKED**  
**Audience:** SDK Users, Rule Authors, Runtime Implementers  
**Breaking Change Policy:** **FORBIDDEN** after this document

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

---

## Table of Contents

1. [Purpose & Stability Guarantee](#1-purpose--stability-guarantee)
2. [Package Boundary Rules](#2-package-boundary-rules)
3. [Core Types Overview](#3-core-types-overview)
4. [Engine Contract](#4-engine-contract)
5. [Rule Contract](#5-rule-contract)
6. [Context Contract](#6-context-contract)
7. [Event Contract](#7-event-contract)
8. [Alert Contract](#8-alert-contract)
9. [EventSource Contract](#9-eventsource-contract)
10. [AlertSink Contract](#10-alertsink-contract)
11. [Dispatcher Contract](#11-dispatcher-contract)
12. [Error Semantics](#12-error-semantics)
13. [Extension Points](#13-extension-points-allowed)
14. [Explicit Non-Extension Points](#14-explicit-non-extension-points-forbidden)
15. [Versioning & Compatibility Rules](#15-versioning--compatibility-rules)

---

## 1. Purpose & Stability Guarantee

Dokumen ini mendefinisikan **SELURUH** kontrak publik Asentric SDK.

Kontrak ini:

* Digunakan oleh **rule authors**
* Digunakan oleh **runtime implementers**
* Digunakan oleh **SDK integrators**

### Stability Guarantee

* ✅ Semua API di `pkg/asentric` **STABLE**
* ✅ Tidak ada breaking change tanpa major version bump
* ✅ Semua perubahan harus backward-compatible

### Package Stability

| Package | Stability |
|---------|-----------|
| `pkg/asentric/*` | ✅ **STABLE** (v1 contract) |
| `cmd/asentric/*` | ⚠️ Best-effort (DX tooling) |
| `internal/*` | ❌ Private, no guarantees |

---

## 2. Package Boundary Rules

### Allowed Imports (USER CODE)

```go
import "github.com/asentric/asentric/pkg/asentric"
```

**User code HANYA BOLEH mengimpor dari `pkg/asentric`.**

### Forbidden Imports

```go
// ❌ DILARANG
import "github.com/asentric/asentric/internal/..."
```

### Rule Mutlak

* ❌ **User TIDAK BOLEH mengimpor `internal/*`**
* ❌ **`pkg/asentric` TIDAK BOLEH mengimpor `internal/*`**
* ✅ Dependency hanya satu arah: `internal/*` → `pkg/asentric`

---

## 3. Core Types Overview

| Type | Responsibility |
|------|---------------|
| **Engine** | Orchestrates rule evaluation |
| **Rule** | Pure detection logic |
| **Context** | Immutable snapshot of event |
| **Event** | Normalized chain event |
| **Alert** | Rule output |
| **EventSource** | Event ingestion interface |
| **AlertSink** | Alert delivery interface |
| **Dispatcher** | Runtime orchestration |

---

## 4. Engine Contract

### Type Definition

```go
type Engine struct {
    // internal state (rule registry, config)
}

func NewEngine() *Engine
func (e *Engine) RegisterRule(rule Rule) error
func (e *Engine) Evaluate(ctx Context) ([]*Alert, error)
```

**Important:** Engine is a **CONCRETE TYPE**, not an interface.

### Extension Point Status

**Engine is NOT an extension point.**

* ❌ Engine tidak boleh di-extend atau di-subclass
* ❌ Engine tidak boleh di-replace dengan custom implementation
* ✅ Engine adalah concrete type dengan invariant yang kuat

**Reasoning:**

* Engine punya invariant kuat (determinism, ordering, error semantics)
* Engine bukan extension point
* Interface hanya akan membuka peluang "fake engine" yang melanggar kontrak

### Semantics

* ✅ **No execution state** — Engine only maintains configuration state (rules, ordering), not runtime mutable state
* ✅ **Deterministic** — Same input → same output
* ✅ **Single-threaded** — Engine tidak concurrency-safe
* ✅ **No side effects** — Engine tidak melakukan I/O

**Note:** Engine maintains internal state (rule registry, config), but does not maintain per-event or cross-event execution state. Each `Evaluate()` call is independent.

### Rules

* ✅ `ctx` **MUST NOT** be nil
* ✅ `ctx.Event` **MUST NOT** be nil (jika Context memiliki Event)
* ✅ Engine **MUST NOT** mutate Context
* ✅ Engine **MUST NOT** perform I/O
* ✅ Engine **MUST NOT** manage concurrency
* ✅ Rules dieksekusi secara **sequential**
* ✅ Rule execution order adalah **deterministic** (registration order)

### Error Behavior

| Condition | Error |
|-----------|-------|
| `ctx == nil` | `ErrInvalidContext` |
| `ctx.Event == nil` | `ErrInvalidEvent` |
| Rule execution failure | Propagated error |
| Rule panic | Engine MUST recover and return `ErrRulePanic` |

### Behavioral Guarantees

* Rules are executed sequentially
* Each rule is executed at most once per `Evaluate` call
* Rules cannot affect each other's execution
* Rule execution order is deterministic (registration order)
* Engine maintains internal state (rule registry, config)
* Engine is **NOT** safe for concurrent use

---

## 5. Rule Contract

### Interface

```go
type Rule interface {
    Name() string
    Evaluate(ctx Context) (*Alert, error)
}
```

### Semantics

* ✅ **Rule adalah pure function**
* ✅ **Deterministic** — Same Context → same output
* ✅ **Side-effect free** — Tidak ada I/O atau mutasi
* ✅ **No state** — Rule tidak menyimpan state (tidak ada per-event atau cross-event state)

### Rules

**Rule **MUST**:**

* ✅ Not perform I/O (network, file, database)
* ✅ Not mutate Context
* ✅ Not depend on global state
* ✅ Return at most one alert per evaluation
* ✅ Return `nil, nil` if no detection

**Rule **MAY**:**

* ✅ Perform computation
* ✅ Decode ABI data via context
* ✅ Return execution errors

### Error Handling Rules

| Condition | Expected Return |
|-----------|----------------|
| Detection not matched | `nil, nil` |
| Detection matched | `alert, nil` |
| Invalid input / decode failure | `nil, error` |

**Errors represent execution failure, not detection outcome.**

### Output Rules

* ✅ **1 Rule → maksimal 1 Alert** per evaluasi
* ✅ **Rule TIDAK BOLEH emit alert langsung** — Alert hanya dikembalikan dari `Evaluate()`
* ✅ **Rule hanya mengembalikan Alert atau error** — Tidak ada side channel

---

## 6. Context Contract

### Type Definition

```go
type Context interface {
    ChainID() domain.ChainID  // Returns chain ID (uint64 alias)
    Tx() domain.Transaction   // Transaction data with Value() *big.Int
    Block() domain.Block      // Block metadata
    Logs() []domain.Log       // Decoded logs
    ABI() domain.ABIRegistry  // ABI access for decoding
}
```

> **Note:** Domain types are defined in `pkg/domain/`. See [SPEC.md](SPEC.md) for complete type definitions.

### Semantics

* ✅ **Immutable** — Context tidak boleh dimutasi
* ✅ **Snapshot** — Context adalah snapshot dari event
* ✅ **Single source of truth** — Semua data execution ada di Context
* ✅ **Deterministic** — Same event → same Context

### Rules

* ✅ **Context HANYA dibuat dari Event** — Context tidak boleh dibuat dari sumber lain
* ✅ **Context TIDAK BOLEH dimutasi** — Context adalah read-only selama evaluasi
* ✅ **Context TIDAK BOLEH menyimpan state lain** — Context hanya berisi data dari event

### Context Guarantees

* No global access
* No hidden state
* Same input → same output
* Safe for replay and testing

### Context Immutability Enforcement

**Context implementers MUST guarantee deep immutability, not only interface-level immutability.**

**Rules:**

* ✅ **Returned objects (Transaction, Block, Log) MUST be immutable** — Objects yang dikembalikan dari Context tidak boleh dimutasi
* ✅ **SDK MAY wrap underlying data in read-only views** — SDK boleh wrap data dengan read-only views untuk enforce immutability
* ✅ **Mutation attempts result in undefined behavior** — Attempts untuk memutasi Context atau objects yang dikembalikan akan menghasilkan undefined behavior. Undefined behavior may include incorrect detections, inconsistent results, or engine errors.

**Implementation Requirements:**

* Context implementers harus memastikan bahwa semua data yang dikembalikan adalah immutable
* Deep immutability berarti tidak hanya interface-level, tapi juga semua nested objects
* Runtime implementers harus memastikan bahwa data yang diberikan ke Context sudah immutable

> **Note:** The SDK assumes Context is valid and complete. Context validation and enrichment (e.g., fetching missing data, decoding events) is the runtime's responsibility. The SDK does not validate or enrich Context.

---

## 7. Event Contract

### Type Definition

```go
type Event struct {
    ChainID     uint64
    BlockNumber uint64
    TxHash      string
    Payload     any
}
```

**Important:** `ChainID` is `uint64`, not `string`.

**Reasoning:**

* ChainID adalah numeric EVM canonical
* String membuka ambiguity ("1", "01", "eth-mainnet")
* `uint64` memastikan type safety dan consistency

> **Note:** Exact structure may vary. This is a conceptual representation. The key is that Event is a normalized representation of chain data.

### Semantics

* ✅ **Normalized representation** — Event adalah format yang sudah di-normalize
* ✅ **Infrastructure-agnostic** — Event tidak mengandung metadata infrastructure
* ✅ **Immutable** — Event tidak boleh dimutasi

### Rules

* ✅ **Payload READ-ONLY** — Payload tidak boleh dimutasi
* ✅ **Payload MUST be safe for concurrent read access** — Payload harus aman untuk concurrent read access (runtime boleh parallelize di Dispatcher)
* ✅ **Event TIDAK BOLEH contain infra metadata** — Event tidak mengandung Redis, RPC, atau metadata infrastructure lainnya
* ✅ **Event TIDAK BOLEH lazy-load data** — Event harus complete saat dibuat

---

## 8. Alert Contract

### Type Definition

```go
type Alert struct {
    Rule        string
    Severity    Severity
    Title       string
    Description string
    Ref         *ExecutionRef  // optional, populated by engine
    Metadata    map[string]any
}

type ExecutionRef struct {
    TxHash      string
    BlockNumber uint64
}
```

### Semantics

* ✅ **Serializable** — Alert dapat di-serialize ke JSON
* ✅ **Immutable** — Alert tidak boleh dimutasi setelah dibuat
* ✅ **Output-only** — Alert adalah output dari rule, bukan input

### Rules

* ✅ **Alert TIDAK BOLEH contain delivery logic** — Alert tidak tahu bagaimana akan dikirim
* ✅ **Alert TIDAK BOLEH mutate Event** — Alert tidak memutasi Event yang menjadi sumbernya
* ✅ **Alert TIDAK BOLEH trigger side effects** — Alert adalah pure data structure
* ✅ **`ExecutionRef` is informational only** — Tidak termasuk chain identity, network, atau RPC endpoints
* ✅ **Metadata MUST NOT be mutated after Alert creation** — Metadata map tidak boleh dimutasi setelah Alert dibuat. Runtime and sinks MUST treat Metadata as read-only. SDK MAY defensively copy Metadata when necessary.

### Alert Design Rules

* Alerts are pure semantic outputs
* Metadata must be JSON-serializable
* Alerts do not imply delivery
* Alerts do not imply persistence
* `ExecutionRef` is populated by the engine, not by rules

### What Alerts Do Not Include

Alerts do not include:

* Chain IDs (chain context is runtime responsibility)
* Network information
* RPC endpoints
* Delivery metadata

---

## 9. EventSource Contract

### Interface

```go
type EventSource interface {
    Start(ctx context.Context) (<-chan Event, error)
}
```

### Semantics

* ✅ **Owned by runtime** — EventSource adalah runtime responsibility
* ✅ **Produces normalized Events** — EventSource menghasilkan Event yang sudah di-normalize
* ✅ **Manages its own goroutines** — EventSource mengelola concurrency sendiri

### Rules

* ✅ **Start MUST be idempotent** — Memanggil Start beberapa kali tidak menyebabkan error
* ✅ **Channel close signals completion** — Close channel menandakan EventSource selesai
* ✅ **EventSource TIDAK BOLEH panic** — Error harus dikembalikan, bukan panic

### Responsibilities

* Fetch chain data (RPC, websocket, etc.)
* Normalize chain data menjadi Event
* Manage subscription lifecycle
* Handle reconnection and retry (internal)

---

## 10. AlertSink Contract

### Interface

```go
type AlertSink interface {
    Emit(ctx context.Context, alert *Alert) error
}
```

### Semantics

* ✅ **External delivery mechanism** — AlertSink adalah interface untuk delivery
* ✅ **Infrastructure-owned** — Implementasi AlertSink adalah runtime responsibility

### Rules

* ✅ **Sink MAY perform I/O** — AlertSink boleh melakukan network calls, database writes, dll
* ✅ **Sink MUST NOT mutate Alert** — AlertSink tidak boleh memutasi Alert
* ✅ **Sink MUST handle retries externally** — Retry logic adalah AlertSink responsibility

### Responsibilities

* Deliver alerts ke external systems (webhook, database, message queue, etc.)
* Handle delivery failures dan retries
* Manage delivery state (jika diperlukan)

---

## 11. Dispatcher Contract

### Status: Internal Component (NOT Public Extension Point)

**Dispatcher is NOT part of the stable SDK surface.**

**Hard Rule:**

* ❌ **Dispatcher interface MAY exist in `pkg/asentric` for documentation only**, but is **NOT a supported public extension point**
* ❌ **Dispatcher MAY change without notice** — Tidak ada stability guarantee
* ❌ **User code MUST NOT depend on Dispatcher interface** — Dispatcher adalah internal implementation detail

### Interface (Documentation Only)

```go
type Dispatcher interface {
    Dispatch(ctx context.Context, event Event) error
}
```

> **Warning:** This interface is shown for documentation purposes only. It is NOT a stable public API and MUST NOT be used by user code.

### Semantics

* ✅ **Runtime orchestration layer** — Dispatcher adalah adapter antara EventSource dan Engine
* ✅ **Bridges Event → Context → Engine → AlertSink** — Dispatcher mengorkestrasi alur data
* ✅ **May handle parallelism** — Dispatcher boleh mengelola concurrency

### Rules

* ✅ **Dispatcher MAY handle parallelism** — Dispatcher boleh menggunakan worker pools, goroutines, dll
* ✅ **Dispatcher MUST NOT mutate Event** — Dispatcher tidak boleh memutasi Event
* ✅ **Dispatcher MUST respect Engine determinism** — Dispatcher harus memastikan Engine tetap deterministic

### Responsibilities

* Receive events dari EventSource
* Convert Event menjadi Context
* Invoke Engine.Evaluate() dengan Context
* Collect alerts dari Engine
* Send alerts ke AlertSink

**Catatan:** Dispatcher adalah komponen internal (bukan public API) dan dapat berubah bebas tanpa breaking change. User code tidak boleh mengimplementasikan atau menggunakan Dispatcher interface.

---

## 12. Error Semantics

### Defined Errors

```go
var (
    ErrInvalidContext  = errors.New("invalid context")
    ErrInvalidEvent     = errors.New("invalid event")
    ErrNoDispatcher     = errors.New("no dispatcher configured")
    ErrAlreadyRunning   = errors.New("runtime already running")
    ErrRulePanic        = errors.New("rule panic")
)
```

**Note:** `ErrRuleNotFound` is reserved for future SDK APIs and may not be returned in v1 execution flow. It is not part of the current public error contract.

### Error Rules

* ✅ **Errors are typed & stable** — Error types tidak berubah tanpa major version bump
* ✅ **No panic in public API** — Public API tidak boleh panic, harus return error
* ✅ **Errors are propagated upward** — Error dari rule di-propagate ke caller
* ✅ **Panic recovery is Engine responsibility** — Engine MUST recover dari rule panic dan return error

### Error Handling by Component

| Component | Error Behavior |
|-----------|----------------|
| **Engine** | Returns error, stops processing current Context. **MUST recover from rule panic and return `ErrRulePanic`** |
| **Rule** | Returns `(nil, error)` for execution failure. **MUST NOT panic** (but if it does, Engine will recover) |
| **EventSource** | Returns error from `Start()`, closes channel on fatal error |
| **AlertSink** | Returns error, caller decides retry strategy |
| **Dispatcher** | Returns error, runtime decides recovery strategy |

### Panic Handling

**Rule panic behavior:**

* ✅ **Engine MUST recover from rule panic** — Engine tidak boleh crash karena rule panic
* ✅ **Engine MUST return `ErrRulePanic`** — Panic di-recover dan dikonversi menjadi error
* ✅ **SDK public API tidak boleh crash runtime** — Panic recovery adalah Engine responsibility, bukan runtime

**Reasoning:**

* SDK public API tidak boleh crash runtime
* Panic recovery adalah Engine responsibility, bukan runtime
* Rule panic harus di-handle gracefully dengan error, bukan propagate

---

## 13. Extension Points (ALLOWED)

| Extension | How |
|-----------|-----|
| **Custom rules** | Implement `Rule` interface |
| **Custom chain source** | Implement `EventSource` interface |
| **Custom alert delivery** | Implement `AlertSink` interface |
| **Runtime orchestration** | Runtime responsibility (Dispatcher is NOT an extension point) |
| **Parallelism** | Runtime layer only (via runtime orchestration) |

### Example: Custom Rule

```go
type MyRule struct{}

func (r *MyRule) Name() string {
    return "my_custom_rule"
}

func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    // Pure detection logic
    if condition {
        return &Alert{
            Rule:        r.Name(),
            Severity:    High,
            Title:       "Detection Title",
            Description: "Detection Description",
            Metadata:    map[string]any{},
        }, nil
    }
    return nil, nil
}
```

---

## 14. Explicit Non-Extension Points (FORBIDDEN)

Anti-pattern berikut **DILARANG**:

* ❌ **Modify Engine behavior** — Engine interface tidak boleh di-extend atau di-modify
* ❌ **Mutate Context** — Context adalah immutable, tidak boleh dimutasi
* ❌ **Inject infra into public API** — Public API tidak boleh mengandung Redis, RPC, atau infrastructure lainnya
* ❌ **Emit alert outside rule** — Alert hanya boleh dihasilkan oleh Rule.Evaluate()
* ❌ **Stateful rule execution** — Rule tidak boleh menyimpan state antar evaluasi
* ❌ **Rule-to-rule communication** — Rules tidak boleh berkomunikasi satu sama lain

### Examples of Forbidden Patterns

```go
// ❌ DILARANG: Mutate Context
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    ctx.SetData(newData)  // DILARANG
}

// ❌ DILARANG: Network call in Rule
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    resp, err := http.Get("https://api.example.com")  // DILARANG
}

// ❌ DILARANG: Stateful Rule
var globalState = make(map[string]bool)  // DILARANG
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    globalState["key"] = true  // DILARANG
}
```

---

## 15. Versioning & Compatibility Rules

### Semantic Versioning

* ✅ **`pkg/asentric` follows semantic versioning** — Major.Minor.Patch
* ✅ **Breaking change → major bump** — Breaking changes memerlukan major version bump
* ✅ **Internal packages → no stability guarantee** — `internal/*` dapat berubah tanpa notice

### Compatibility Guarantees

* ✅ **Backward compatibility** — Minor dan patch versions tidak breaking
* ✅ **Deprecation policy** — Deprecated APIs akan di-mark dan dihapus di major version berikutnya
* ✅ **Documentation is authoritative** — Dokumentasi ini adalah source of truth

### Version History

| Version | Status | Notes |
|---------|--------|-------|
| v1.0.0 | ✅ **LOCKED** | Initial stable contract |

---

## FINAL LOCK STATEMENT

**`sdk-api.md` is now LOCKED.**

From this point forward:

* ✅ **Any breaking change requires Architecture RFC** — Breaking changes harus melalui Architecture Review
* ✅ **`pkg/asentric` is considered frozen** — Public API tidak boleh berubah tanpa major version bump
* ✅ **Internal implementation must conform to this contract** — Implementasi internal harus mengikuti kontrak ini

---

## Related Documentation

* **[SPEC.md](SPEC.md)** - MVP specification (authoritative)
* **[architecture.md](architecture.md)** - Core architecture (authoritative)
* **[project-structure.md](project-structure.md)** - Final project structure
* **[developer-overview.md](developer-overview.md)** - Developer end-to-end guide

---

## Summary

Asentric SDK Public API adalah:

* ✅ **Stable contract** — Kontrak yang stabil dan terkunci
* ✅ **Pure execution engine** — Engine yang pure dan deterministic
* ✅ **Infrastructure-agnostic** — Tidak tergantung pada infrastructure
* ✅ **Extensible** — Dapat di-extend melalui interfaces
* ✅ **Well-defined boundaries** — Batasan yang jelas antara public dan internal

**Dokumen ini bersifat authoritative.**

Jika terjadi konflik antara implementasi dan dokumen ini, maka implementasi dianggap salah.

---

**Last Updated:** 2024  
**Version:** 1.0 (FINAL – LOCKED)
