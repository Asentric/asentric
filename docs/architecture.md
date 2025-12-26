# Asentric Architecture

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH** untuk hackathon  
> **📖 Lihat alur developer:** [developer-overview.md](developer-overview.md) - Alur end-to-end penggunaan Asentric SDK

**Status:** ✅ **Authoritative**  
**Audience:** Core Engineers, Contributors, Reviewers

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

---

## Table of Contents

1. [Purpose & Scope](#1-purpose--scope)
2. [Core Principles](#2-core-principles)
3. [Non-Goals](#3-non-goals)
4. [High-Level System Overview](#4-high-level-system-overview)
5. [Canonical Data Flow](#5-canonical-data-flow)
6. [Core Components & Responsibilities](#6-core-components--responsibilities)
7. [Lifecycle Overview](#7-lifecycle-overview)
8. [State & Determinism Model](#8-state--determinism-model)
9. [Boundaries & Ownership](#9-boundaries--ownership)
10. [Explicit Anti-Patterns](#10-explicit-anti-patterns)

---

## 1. Purpose & Scope

Asentric adalah SDK untuk real-time on-chain security alerting.

Dokumen ini mendefinisikan arsitektur inti dan batasan sistem.

**Dokumen ini bersifat authoritative.**

Jika terjadi konflik antara implementasi dan dokumen ini, maka implementasi dianggap salah.

### Scope

Asentric SDK menyediakan:

* **Pure execution engine** untuk evaluasi security rules
* **Deterministic rule system** yang menghasilkan alerts
* **Infrastructure-agnostic core** yang dapat di-embed di berbagai runtime

Asentric SDK **tidak** menyediakan:

* Chain data fetching
* Infrastructure management (Redis, databases, HTTP servers)
* Alert delivery mechanisms
* Deployment tooling

Tanggung jawab ini didelegasikan ke runtime (developer-built, self-hosted).

---

## 2. Core Principles

Asentric dibangun berdasarkan prinsip berikut:

### 2.1 Stateless Engine

* Engine tidak menyimpan state jangka panjang
* Setiap evaluasi adalah independen
* Tidak ada shared mutable state antar evaluasi

### 2.2 Deterministic Rule Execution

* Rule evaluation harus **pure dan deterministic**
* Same input → same output, selalu
* Tidak bergantung pada waktu, network, atau external state
* Infrastructure tidak boleh mempengaruhi rule semantics

### 2.3 Immutable Event & Context

* Event adalah immutable snapshot dari chain data
* Context adalah immutable snapshot dari event
* Tidak ada mutasi selama evaluasi
* Context hanya dibuat dari event, tidak dari sumber lain

### 2.4 Clear Separation Between Public API and Internal Implementation

* Public API hanya di `pkg/asentric`
* Internal implementation di `internal/*`
* Internal packages boleh depend ke `pkg/asentric`, **tidak pernah sebaliknya**
* Public contracts harus stabil sebelum internal implementation

### 2.5 Infrastructure-Agnostic Core

* Engine tidak tahu tentang Redis, RPC, atau infrastructure lainnya
* Infrastructure adalah runtime responsibility
* Core hanya fokus pada domain logic

### 2.6 Single-Threaded Engine Design

* Engine **tidak concurrency-safe**
* Engine mengeksekusi rules secara sequential
* Parallelism adalah runtime responsibility

**Catatan:** Prinsip-prinsip ini **harus konsisten** dengan roadmap yang sudah dikunci di [implementation-roadmap.md](implementation-roadmap.md).

---

## 3. Non-Goals

Asentric secara eksplisit **TIDAK** bertujuan untuk:

### 3.1 Menjadi SIEM Lengkap

* Tidak menyediakan query engine untuk historical data
* Tidak menyediakan dashboard atau visualization
* Tidak menyediakan alert aggregation atau correlation

### 3.2 Menyimpan Historical State sebagai Source of Truth

* Engine tidak menyimpan state jangka panjang
* Redis adalah transport, bukan state storage
* Tidak ada persistent state di engine

### 3.3 Menyediakan Query Engine

* Tidak ada query API untuk historical data
* Tidak ada indexing atau search capabilities
* Historical analysis adalah runtime/backend responsibility

### 3.4 Menjadi Workflow Engine

* Tidak ada orchestration antar rules
* Tidak ada dependency management antar rules
* Rules adalah isolated units of logic

### 3.5 Menyediakan Infrastructure Management

* Tidak ada built-in Redis management
* Tidak ada built-in RPC connection pooling
* Tidak ada built-in database migrations
* Infrastructure adalah developer responsibility

**Tujuan bagian ini adalah membunuh ekspektasi sejak awal.**

---

## 4. High-Level System Overview

Secara konseptual, Asentric terdiri dari:

* **Engine**: orchestration & lifecycle
* **Chain source**: penyedia event (via EventSource interface)
* **Rule**: pure detection logic
* **Alert sink**: external delivery (via AlertSink interface)

### Diagram ASCII

```
[ Chain Source ] → [ EventSource ] → [ Engine ] → [ Rule ] → [ Alert ] → [ AlertSink ]
```

### Komponen Kunci

```
┌─────────────────────────────────────────────────────────────┐
│                    Asentric SDK Core                          │
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Engine     │───▶│    Rule      │───▶│    Alert     │  │
│  │ (orchestrate)│    │  (detection) │    │   (output)   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                    │                    │          │
│         │                    │                    │          │
│         ▼                    ▼                    ▼          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Context    │    │ EventSource  │    │  AlertSink   │  │
│  │ (immutable)  │    │  (interface) │    │  (interface)  │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. Canonical Data Flow

**INI BAGIAN PALING KRITIS**

Alur data canonical Asentric adalah:

```
Chain Event
  → Internal Event (via EventSource)
    → Context (immutable snapshot)
      → Rule Evaluation
        → Alert (optional)
          → AlertSink
```

### Step-by-Step Breakdown

#### Step 1: Chain Event

* Raw data dari blockchain (block, log, transaction)
* Disediakan oleh runtime melalui EventSource interface

#### Step 2: Internal Event

* Event yang sudah di-normalize oleh EventSource
* Masih belum menjadi Context

#### Step 3: Context (Immutable Snapshot)

* Context dibuat dari Event
* Context adalah immutable snapshot
* Context tidak boleh dimutasi selama evaluasi

#### Step 4: Rule Evaluation

* Engine mengeksekusi rules terhadap Context
* Setiap rule adalah pure function
* Rule hanya membaca Context, tidak memutasi

#### Step 5: Alert (Optional)

* Rule menghasilkan maksimal satu Alert per evaluasi
* Alert adalah output akhir dari rule
* Alert tidak memiliki side-effect

#### Step 6: AlertSink

* Alert dikirim ke AlertSink interface
* AlertSink adalah external delivery mechanism
* AlertSink adalah runtime responsibility

### Rules (WAJIB DIPATUHI)

* ✅ **Tidak boleh lompat step** — Setiap step harus dijalankan secara berurutan
* ✅ **Context hanya dibuat dari event** — Context tidak boleh dibuat dari sumber lain
* ✅ **Rule hanya membaca context** — Rule tidak boleh memutasi Context atau melakukan I/O
* ✅ **Alert hanya dari rule** — Alert hanya bisa dihasilkan oleh rule evaluation

**Jika bagian ini kabur → TIER 0 gagal.**

---

## 6. Core Components & Responsibilities

### 6.1 Engine

**Tanggung Jawab:**

* Mengelola lifecycle (start, stop, shutdown)
* Mengatur subscription ke EventSource
* Menjalankan rule evaluation terhadap Context
* Mengumpulkan alerts dari rules
* Mengirim alerts ke AlertSink

**Tidak Bertanggung Jawab:**

* ❌ Fetch chain data
* ❌ Manage infrastructure (Redis, RPC, database)
* ❌ Deliver alerts (hanya mengirim ke AlertSink)
* ❌ Manage concurrency (engine adalah single-threaded)

**Invariants:**

* Engine adalah stateless
* Engine mengeksekusi rules secara sequential
* Engine tidak tahu tentang infrastructure

### 6.2 Rule

**Tanggung Jawab:**

* Pure function terhadap Context
* Mendeteksi security patterns
* Menghasilkan maksimal satu Alert per evaluasi

**Tidak Bertanggung Jawab:**

* ❌ Melakukan I/O (network, file, database)
* ❌ Memutasi external state
* ❌ Berkomunikasi dengan rules lain
* ❌ Mengetahui tentang infrastructure

**Invariants:**

* Rule adalah pure function
* Rule tidak memiliki side effects
* Rule adalah deterministic
* Rule adalah isolated unit of logic

### 6.3 Context

**Tanggung Jawab:**

* Menyediakan immutable snapshot dari event
* Menyediakan akses ke transaction data, block metadata, decoded logs
* Menjadi single source of truth selama evaluasi

**Tidak Bertanggung Jawab:**

* ❌ Memvalidasi data (runtime responsibility)
* ❌ Meng-enrich data (runtime responsibility)
* ❌ Menyimpan state jangka panjang

**Invariants:**

* Context adalah immutable
* Context hanya dibuat dari event
* Context tidak mengandung behavior
* Context tidak boleh dimutasi selama evaluasi

### 6.4 Alert

**Tanggung Jawab:**

* Menyediakan structured output dari rule
* Menyediakan severity level (Critical, High, Medium, Low, Info)
* Menyediakan metadata untuk debugging

**Tidak Bertanggung Jawab:**

* ❌ Mengetahui bagaimana alert akan dikirim
* ❌ Mengetahui bagaimana alert akan disimpan
* ❌ Mengandung delivery logic

**Invariants:**

* Alert adalah serializable
* Alert adalah infrastructure-agnostic
* Alert tidak memiliki side-effect
* Alert adalah output akhir dari rule

### 6.5 EventSource

**Tanggung Jawab:**

* Menyediakan interface untuk chain data ingestion
* Menyediakan subscription mechanism untuk events
* Menyediakan normalized event format

**Tidak Bertanggung Jawab:**

* ❌ Implementasi chain client (internal responsibility)
* ❌ State management (runtime responsibility)

**Invariants:**

* EventSource adalah interface (public API)
* Chain clients adalah internal implementation
* EventSource tidak tahu tentang infrastructure

### 6.6 AlertSink

**Tanggung Jawab:**

* Menyediakan interface untuk alert delivery
* Menyediakan mechanism untuk mengirim alerts ke external systems

**Tidak Bertanggung Jawab:**

* ❌ Implementasi delivery mechanism (runtime responsibility)
* ❌ Alert storage (runtime responsibility)

**Invariants:**

* AlertSink adalah interface (public API)
* Delivery mechanisms adalah runtime responsibility
* AlertSink tidak tahu tentang infrastructure details

### 6.7 Dispatcher

**Tanggung Jawab:**

* Mengirim event ke engine untuk diproses
* Mengorkestrasi rule evaluation terhadap Context
* Menghubungkan EventSource dengan Engine

**Tidak Bertanggung Jawab:**

* ❌ Implementasi rule logic (Rule responsibility)
* ❌ Chain data fetching (EventSource responsibility)
* ❌ Alert delivery (AlertSink responsibility)

**Invariants:**

* Dispatcher adalah komponen internal (bukan public API)
* Dispatcher dapat berubah bebas tanpa breaking change
* Dispatcher adalah adapter antara EventSource dan Engine

**Catatan:** Dispatcher adalah komponen internal yang bertanggung jawab untuk mengirim event ke engine dan mengorkestrasi rule evaluation. Dispatcher bukan public API dan dapat berubah bebas tanpa mempengaruhi kontrak publik.

---

## 7. Lifecycle Overview

### Engine Lifecycle

```
INIT → STARTING → RUNNING → STOPPING → STOPPED
```

### State Transitions

#### INIT

* Engine dalam state awal
* Belum ada subscription
* Belum ada rule yang terdaftar

#### STARTING

* Engine sedang memulai lifecycle
* Subscription ke EventSource sedang dibuat
* Rules sedang di-register

#### RUNNING

* Engine aktif memproses events
* Rules sedang dievaluasi
* Alerts sedang dihasilkan

#### STOPPING

* Engine sedang menghentikan operasi
* Subscription sedang di-unsubscribe
* Graceful shutdown sedang dilakukan

#### STOPPED

* Engine sudah berhenti
* Tidak ada operasi yang berjalan
* State sudah dibersihkan

### Aturan Lifecycle

* ✅ **Start hanya valid dari INIT** — Engine tidak bisa start dari state lain
* ✅ **Stop idempotent** — Memanggil stop beberapa kali tidak menyebabkan error
* ✅ **Tidak ada restart implisit** — Setelah stop, engine harus dibuat ulang untuk restart
* ✅ **Graceful shutdown** — Engine harus menyelesaikan evaluasi yang sedang berjalan sebelum stop

### Failure Semantics

Jika rule menghasilkan error selama evaluasi:

* **Engine boleh stop atau skip** — Policy ditentukan oleh runtime, bukan engine
* **Engine tidak retry secara otomatis** — Retry adalah runtime responsibility
* **Error tidak mempengaruhi determinism** — Error adalah output yang deterministic dari rule evaluation

**Catatan:** Ini hanya mengunci ekspektasi tentang behavior engine saat terjadi error. Implementasi spesifik (stop vs skip, error handling strategy) adalah runtime responsibility.

---

## 8. State & Determinism Model

### 8.1 Engine Bersifat Stateless

Engine tidak menyimpan state jangka panjang:

* Tidak ada persistent state di engine
* Tidak ada shared mutable state antar evaluasi
* Setiap evaluasi adalah independen

### 8.2 Rule Harus Deterministic

Semua rule harus deterministic:

* **Input sama → output sama** — Same Context dan rule set selalu menghasilkan same alerts
* **Tidak bergantung pada waktu** — Rule tidak boleh menggunakan `time.Now()` atau timestamp
* **Tidak bergantung pada network** — Rule tidak boleh melakukan network calls
* **Tidak bergantung pada external state** — Rule tidak boleh membaca dari database atau external systems

### 8.3 Context Adalah Immutable Snapshot

Context adalah immutable snapshot dari event:

* Context tidak boleh dimutasi selama evaluasi
* Context hanya dibuat dari event, tidak dari sumber lain
* Context tidak mengandung mutable state

### 8.4 Kenapa Ini Penting

**Untuk Scaling:**

* Stateless engine memungkinkan horizontal scaling
* Deterministic rules memungkinkan parallel execution
* Immutable context memungkinkan safe concurrent access

**Untuk Replay:**

* Deterministic execution memungkinkan replay dari fixtures
* Stateless engine memungkinkan replay tanpa side effects
* Immutable context memungkinkan replay dengan confidence

**Untuk Testing:**

* Deterministic rules mudah di-test
* Stateless engine tidak memerlukan complex test setup
* Immutable context memungkinkan isolated testing

---

## 9. Boundaries & Ownership

### 9.1 Public API (`pkg/asentric`)

**Ownership:**

* Public API adalah **satu-satunya** integration surface yang didukung
* Public API adalah stable dan versioned
* Breaking changes memerlukan semver bump

**Contains:**

* Engine interface
* Rule interface
* Context interface
* Alert model
* EventSource interface
* AlertSink interface
* Dispatcher interface
* Error types

**Invariants:**

* ✅ Stable & versioned
* ✅ Backward compatible (semver-managed)
* ✅ Safe to depend on
* ❌ No infrastructure code
* ❌ No internal implementation details

### 9.2 Internal Implementation (`internal/*`)

**Ownership:**

* Internal implementation adalah private
* Internal implementation bebas berubah tanpa breaking change
* Internal implementation tidak boleh di-import oleh user

**Contains:**

* Rule execution internals (`internal/rule/`)
* Runtime helpers (`internal/runtime/`)
* Chain clients (`internal/chain/`)
* ABI decoding logic (`internal/abi/`)
* Alert formatting (`internal/alert/`)

**Invariants:**

* ✅ Private (tidak boleh di-import oleh user)
* ✅ Non-stable (bebas berubah)
* ✅ Free to change tanpa breaking change
* ✅ Boleh depend ke `pkg/asentric`
* ❌ Tidak boleh di-depend oleh `pkg/asentric`

### 9.3 Larangan Eksplisit

**WAJIB DIPATUHI:**

* ❌ **User tidak boleh import dari `internal/*`** — Hanya `pkg/asentric` yang boleh di-import
* ❌ **Internal tidak boleh di-depend oleh public API** — Dependency hanya satu arah: `internal/*` → `pkg/asentric`
* ❌ **Infrastructure tidak boleh bocor ke public API** — Public API tidak boleh mengandung Redis, RPC, atau infrastructure lainnya

---

## 10. Explicit Anti-Patterns

Anti-pattern berikut **dilarang** dan akan menyebabkan refactor besar:

### 10.1 Rule Melakukan Network Call

**❌ DILARANG:**

```go
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    // DILARANG: Network call di dalam rule
    resp, err := http.Get("https://api.example.com/data")
    // ...
}
```

**✅ BENAR:**

```go
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    // BENAR: Rule hanya membaca dari Context
    data := ctx.GetData()
    // ...
}
```

**Alasan:** Rule harus pure dan deterministic. Network call membuat rule non-deterministic dan sulit di-test.

### 10.2 Engine Menyimpan State Jangka Panjang

**❌ DILARANG:**

```go
type Engine struct {
    processedBlocks map[string]bool  // DILARANG: State jangka panjang
    alertHistory    []Alert          // DILARANG: State jangka panjang
}
```

**✅ BENAR:**

```go
type Engine struct {
    rules []Rule  // BENAR: Hanya menyimpan rules
    // State jangka panjang adalah runtime responsibility
}
```

**Alasan:** Engine harus stateless untuk memungkinkan scaling dan replay.

### 10.3 Infrastruktur (Redis, RPC) Bocor ke Public API

**❌ DILARANG:**

```go
// DILARANG: Redis di public API
package asentric

type Engine struct {
    redis *redis.Client  // DILARANG
}
```

**✅ BENAR:**

```go
// BENAR: Redis di internal atau runtime
package internal

type Runtime struct {
    redis *redis.Client  // BENAR: Di internal
}
```

**Alasan:** Public API harus infrastructure-agnostic. Infrastructure adalah runtime responsibility.

### 10.4 Business Logic di CLI

**❌ DILARANG:**

```go
// DILARANG: Business logic di CLI
func main() {
    engine := NewEngine()
    engine.Start()  // DILARANG: CLI tidak boleh run engine
}
```

**✅ BENAR:**

```go
// BENAR: CLI hanya untuk scaffolding dan testing
func main() {
    if cmd == "init" {
        scaffoldProject()  // BENAR: Scaffolding
    } else if cmd == "replay" {
        replayFixture()    // BENAR: Testing
    }
}
```

**Alasan:** CLI adalah developer tool, bukan runtime. Business logic harus di runtime.

### 10.5 Context Dimutasi Selama Evaluasi

**❌ DILARANG:**

```go
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    ctx.SetData(newData)  // DILARANG: Mutasi Context
    // ...
}
```

**✅ BENAR:**

```go
func (r *MyRule) Evaluate(ctx Context) (*Alert, error) {
    data := ctx.GetData()  // BENAR: Hanya membaca Context
    // ...
}
```

**Alasan:** Context harus immutable untuk memastikan determinism dan safe concurrent access.

### 10.6 Rule Berkomunikasi dengan Rules Lain

**❌ DILARANG:**

```go
var globalState = make(map[string]interface{})

func (r *Rule1) Evaluate(ctx Context) (*Alert, error) {
    globalState["key"] = "value"  // DILARANG: Shared state
}

func (r *Rule2) Evaluate(ctx Context) (*Alert, error) {
    value := globalState["key"]  // DILARANG: Membaca shared state
}
```

**✅ BENAR:**

```go
func (r *Rule1) Evaluate(ctx Context) (*Alert, error) {
    // BENAR: Rule isolated, tidak ada shared state
    data := ctx.GetData()
    // ...
}
```

**Alasan:** Rules harus isolated untuk memastikan determinism dan parallel execution.

### 10.7 Alert Mengandung Delivery Logic

**❌ DILARANG:**

```go
type Alert struct {
    Message string
    SendToSlack() error  // DILARANG: Delivery logic di Alert
}
```

**✅ BENAR:**

```go
type Alert struct {
    Message string  // BENAR: Alert hanya data
    // Delivery logic di AlertSink
}
```

**Alasan:** Alert harus infrastructure-agnostic. Delivery adalah AlertSink responsibility.

---

## Related Documentation

* **[SPEC.md](SPEC.md)** - MVP specification (authoritative)
* **[project-structure.md](project-structure.md)** - Final project structure
* **[sdk-api.md](sdk-api.md)** - Complete API reference
* **[developer-overview.md](developer-overview.md)** - Developer end-to-end guide

---

## Summary

Asentric Architecture adalah:

* ✅ **Stateless** — Engine tidak menyimpan state jangka panjang
* ✅ **Deterministic** — Rules selalu menghasilkan output yang sama untuk input yang sama
* ✅ **Immutable** — Context adalah immutable snapshot
* ✅ **Infrastructure-Agnostic** — Core tidak tahu tentang infrastructure
* ✅ **Single-Threaded** — Engine mengeksekusi rules secara sequential
* ✅ **Pure** — Rules adalah pure functions tanpa side effects

**Dokumen ini bersifat authoritative.**

Jika terjadi konflik antara implementasi dan dokumen ini, maka implementasi dianggap salah.

---

**Last Updated:** 2024  
**Version:** 1.0 (FINAL)
