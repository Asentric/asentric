# Asentric SDK – API Reference

Dokumen ini mendefinisikan kontrak API publik dan stabil dari Asentric SDK.

Ditujukan untuk:

* Pengguna SDK (penulis rule)
* Implementer runtime (asentric-bot)
* Konsumen backend
* Kontributor masa depan

> **Important:** Apa pun di luar `pkg/asentric` bukan bagian dari public API dan dapat berubah tanpa pemberitahuan.

---

## API Stability Guarantees

| Package | Stability |
|---------|-----------|
| `pkg/asentric/*` | STABLE (v1 contract) |
| `cmd/asentric/*` | Best-effort (DX tooling) |
| `internal/*` | Private, no guarantees |

Dokumen ini hanya mencakup public API.

---

## Core Concepts Overview

Pada tingkat tinggi, SDK terdiri dari:

* **Engine** — mengorchestrasi eksekusi rule
* **Rule** — logika deteksi murni
* **Context** — data eksekusi yang immutable
* **Alert** — output deteksi semantik
* **Severity** — enum klasifikasi yang ketat

SDK tidak menangani concern infrastruktur seperti RPC, queue, database, atau delivery.

---

## Engine

### Purpose

Engine adalah koordinator pusat yang:

* Menyimpan rules yang terdaftar (stateful)
* Mengeksekusi rules secara sequential untuk context yang diberikan
* Mengumpulkan alert yang dihasilkan oleh rules

Engine adalah stateful tetapi tidak concurrency-safe.

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

* Rules dieksekusi secara sequential
* Setiap rule dieksekusi paling banyak sekali per panggilan `Process`
* Rules tidak dapat mempengaruhi eksekusi satu sama lain
* Urutan eksekusi rule adalah deterministik (urutan registrasi)
* Engine mempertahankan internal state (rule registry, config)
* Engine **TIDAK** aman untuk penggunaan concurrent

### Error Semantics

| Situation | Behavior |
|-----------|----------|
| Rule mengembalikan `(nil, nil)` | Tidak ada alert |
| Rule mengembalikan `(alert, nil)` | Alert dikumpulkan |
| Rule mengembalikan `(nil, error)` | Engine mengembalikan error |
| Rule panic | Panic di-propagate |

Lapisan infrastruktur memutuskan strategi retry / recovery.

---

## Rule

### Purpose

Rule merepresentasikan satu unit logika deteksi.

Rules adalah:

* Pure
* Deterministik
* Bebas side-effect
* Stateless

### Rule Interface

```go
type Rule interface {
    Name() string
    Evaluate(ctx Context) (*Alert, error)
}
```

### Rule Contract

Rules **HARUS**:

* Tidak melakukan I/O
* Tidak memutasi context
* Tidak bergantung pada global state
* Mengembalikan paling banyak satu alert
* Mengembalikan `nil, nil` jika tidak ada deteksi

Rules **BOLEH**:

* Melakukan komputasi
* Decode data ABI melalui context
* Mengembalikan execution errors

### Error Handling Rules

| Condition | Expected Return |
|-----------|----------------|
| Detection tidak cocok | `nil, nil` |
| Detection cocok | `alert, nil` |
| Invalid input / decode failure | `nil, error` |

Error merepresentasikan kegagalan eksekusi, bukan hasil deteksi.

---

## Context

### Purpose

Context menyediakan semua data eksekusi yang diperlukan oleh rules.

Context adalah:

* Immutable
* Berbasis snapshot
* Eksplisit
* Deterministik

Rules tidak dapat memutasi context.

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

> **Note:** Sub-interface yang tepat (`Transaction`, `Block`, `Log`) didefinisikan dalam tipe SDK.

### Context Guarantees

* Tidak ada akses global
* Tidak ada hidden state
* Input yang sama → output yang sama
* Aman untuk replay dan testing

> **Note:** SDK mengasumsikan Context valid dan lengkap. Validasi dan enrichment Context (misalnya, mengambil data yang hilang, decoding event) adalah tanggung jawab runtime. SDK tidak memvalidasi atau memperkaya Context.

---

## Alert

### Purpose

Alert merepresentasikan sinyal keamanan semantik, bukan delivery envelope.

### Execution Reference

Alert dapat menyertakan execution reference opsional untuk debugging dan traceability:

```go
type ExecutionRef struct {
    TxHash      string
    BlockNumber uint64
}
```

**Important:** `ExecutionRef` diisi oleh engine, bukan oleh rules. Rules tidak dapat mengakses atau memodifikasinya.

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

* Alert adalah output semantik murni
* Metadata harus dapat di-serialize ke JSON
* Alert tidak menyiratkan delivery
* Alert tidak menyiratkan persistence
* `ExecutionRef` hanya informasional — tidak termasuk chain identity, network, atau RPC endpoints

### What Alerts Do Not Include

Alert tidak termasuk:

* Chain IDs
* Informasi network
* RPC endpoints
* Delivery metadata

> **Note:** Menyertakan `tx_hash` dan `block_number` dalam `ExecutionRef` tidak menyiratkan SDK memahami chain identity. Chain context tetap menjadi tanggung jawab sistem runtime.

---

## Severity

### Purpose

Severity menyediakan klasifikasi yang ketat dan terstandar untuk alert.

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

* Enum tertutup (tidak ada nilai kustom)
* Urutan bermakna
* Mapping ke string adalah tanggung jawab runtime / backend

---

## Configuration

### Purpose

Konfigurasi SDK mengontrol perilaku engine, bukan infrastruktur.

Contoh:

* Rules yang diaktifkan
* Opsi rule
* Batas eksekusi

### Config Principles

* Di-parse di luar rules
* Di-inject ke engine
* Immutable selama eksekusi

> **Note:** Skema yang tepat didokumentasikan di [architecture.md](architecture.md).

---

## CLI API (`cmd/asentric`)

### Scope

CLI hanya ada untuk developer tooling.

CLI tidak:

* Menjalankan watchers
* Terhubung ke RPC nodes
* Mengirimkan alert
* Mengelola infrastruktur

> **Important:** CLI tidak menjalankan production watchers atau terhubung ke blockchain. CLI adalah tool developer untuk scaffolding, replay, dan validasi rule.

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

* Instance engine adalah single-threaded
* Engine tidak concurrency-safe
* Paralelisme harus ditangani oleh sistem runtime
* **Recommended:** satu instance engine per worker

---

## Testing Guidelines

Rules mudah di-test:

```go
ctx := asentric.NewMockContext(...)
rule := &MyRule{}

alert, err := rule.Evaluate(ctx)
```

Tidak diperlukan mock infrastruktur.

---

## Non-Goals (Explicit)

Asentric SDK **TIDAK**:

* Mengelola koneksi RPC
* Mengambil data blockchain
* Menyimpan alert
* Mengirimkan notifikasi
* Menyediakan API
* Menangani deployment
* Memahami chain identity (chain ID, nama network)
* Mengelola informasi network atau RPC endpoint

> **Important:** Menyertakan `tx_hash` dan `block_number` dalam `ExecutionRef` tidak menyiratkan SDK memahami chain identity. SDK tetap chain-agnostic. Semua chain context (chain ID, network, RPC endpoints) adalah tanggung jawab sistem runtime.

Ini ditangani oleh repository lain.

---

## Summary

Asentric SDK adalah:

* Pure security detection engine
* Kontrak stabil untuk penulis rule
* Shared brain di berbagai runtime
* Batas ketat antara logika dan infrastruktur

Kontrak API ini adalah fondasi untuk seluruh ekosistem Asentric.

