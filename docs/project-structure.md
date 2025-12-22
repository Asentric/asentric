# Asentric – Final Project Structure (v1)

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH** untuk hackathon

Dokumen ini mendefinisikan **struktur folder & file final** untuk seluruh ekosistem Asentric.
Struktur ini adalah **versi jadi (authoritative)** dan menjadi acuan implementasi untuk seluruh tim.

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

Tujuan utama struktur ini:

* Boundary jelas antara **Engine**, **Runtime**, dan **Infrastructure**
* SDK tetap **pure & deterministic**
* Runtime bebas berevolusi dan scalable
* Mudah dipahami oleh contributor baru

---

## 1. High-Level Repository Layout

```
asentric/
│
├── pkg/asentric/              # PUBLIC SDK (Stable API)
├── internal/                  # PRIVATE SDK implementation
├── cmd/                       # CLI & reference runtime
├── templates/                 # Project scaffolding templates
├── examples/                  # Usage examples
├── docs/                      # Architecture & DX docs
├── .github/                   # CI / GitHub config
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 2. Public SDK – `pkg/`

**Ini adalah kontrak resmi SDK.**
Semua developer eksternal **hanya boleh** bergantung ke folder ini.

### 2.1 Core SDK (`pkg/asentric/`)

```
pkg/asentric/
├── engine.go        # Engine struct & implementation
├── rule.go          # Rule interface
├── context.go       # Context interface
├── alert.go         # Alert & Severity model
├── event.go         # Event model
├── config.go        # Engine-level config (non-infra)
├── errors.go        # Error types
├── event_source.go  # EventSource interface
├── alert_sink.go    # AlertSink interface
└── dispatcher.go    # Dispatcher interface
```

### 2.2 Domain Types (`pkg/domain/`)

Lightweight, infrastructure-agnostic types untuk public API.

```
pkg/domain/
├── address.go       # Address (string-based)
├── hash.go          # Hash (string-based)
├── chain.go         # ChainID, Chain
├── transaction.go   # Transaction struct
├── block.go         # Block struct
├── log.go           # Log struct
├── event.go         # Event (decoded)
├── value.go         # NativeValue, TokenAmount
├── token.go         # Token metadata
└── abi.go           # ABIRegistry interface
```

**Prinsip penting:**

* ❌ Tidak ada Redis
* ❌ Tidak ada RPC
* ❌ Tidak ada goroutine
* ❌ Tidak ada IO
* ❌ Tidak ada geth imports
* ✅ Pure execution only
* ✅ String-based types untuk ergonomics

---

## 3. Internal SDK – `internal/`

Semua di sini **private**, boleh berubah tanpa breaking change.
**geth types BOLEH digunakan di sini.**

```
internal/
├── chain/                  # Raw chain types (geth-compatible)
│   ├── types.go            # RawAddress, RawHash, RawTransaction, etc.
│   └── client.go           # Chain client interface
│
├── adapter/                # Conversion layer
│   ├── converter.go        # chain types → domain types
│   └── geth.go             # geth types → chain types
│
├── runtime/                # Runtime lifecycle
│   ├── runtime.go          # Runtime struct & event loop
│   └── shutdown.go         # Graceful shutdown
│
├── dispatcher/             # Event dispatching
│   └── dispatcher.go       # Dispatcher implementation
│
├── context/                # Concrete context implementations
│   └── evm_context.go      # EVM-specific context
│
├── abi/                    # ABI decoding helpers
│   ├── loader.go           # ABI file loading
│   └── decoder.go          # Event/method decoding
│
└── alert/                  # Alert helpers
    └── formatter.go        # Alert formatting
```

**Hybrid Architecture:**

* ✅ **geth types ALLOWED** di `internal/chain/` dan `internal/adapter/`
* ✅ Conversion layer maintains boundary
* ❌ geth types TIDAK BOLEH di `pkg/`

---

## 4. CLI & Reference Runtime – `cmd/`

### 4.1 CLI Tool

```
cmd/asentric/
├── main.go
├── init.go        # asentric init
├── replay.go      # offline deterministic replay
├── version.go
└── internal/
    └── templates.go
```

**CLI adalah developer tool, bukan runtime.**

---

### 4.2 Reference Runtime (Example)

Ini **contoh runtime produksi**, bukan bagian dari SDK core.

```
cmd/runtime-reference/
├── main.go                 # Entry point runtime
│
├── config/
│   ├── loader.go           # Load yaml config
│   └── schema.go
│
├── ingest/
│   ├── evm_logs.go         # Subscribe logs
│   └── blocks.go           # (optional) block stream
│
├── pipeline/
│   ├── dispatcher.go       # Fan-out events
│   └── worker.go           # Engine workers
│
├── state/                  # RUNTIME STATE (Redis here)
│   ├── store.go            # Interface
│   ├── redis.go            # Redis impl
│   └── memory.go           # Dev impl
│
├── alert/
│   ├── webhook.go
│   ├── telegram.go
│   └── dispatcher.go
│
└── runtime.go               # Glue code
```

**Di sinilah Redis hidup.**

---

## 5. Templates – `templates/`

Digunakan oleh `asentric init`.

```
templates/project/
├── config/
│   ├── asentric.yaml      # Runtime & engine config
│   └── registry.yaml      # What to monitor
│
├── rules/
│   └── example_rule.go
│
├── abi/
│   └── .gitkeep
│
├── cmd/
│   └── watcher/
│       └── main.go
│
├── go.mod.tmpl
└── README.md.tmpl
```

---

## 6. Examples – `examples/`

```
examples/
├── simple-watcher/
├── custom-rules/
├── multi-chain/
└── advanced-replay/
```

Digunakan sebagai referensi, **bukan production**.

---

## 7. Documentation – `docs/`

```
docs/
├── SPEC.md                  # ⭐ MVP Specification (SINGLE SOURCE OF TRUTH)
├── developer-overview.md    # Alur end-to-end developer
├── architecture.md          # Core philosophy & boundaries
├── sdk-api.md               # Public API reference
└── project-structure.md     # This file - Final structure
```

---

## 8. Boundary Summary (WAJIB DIPATUHI)

| Layer   | Redis | RPC | State      | Deterministic |
| ------- | ----- | --- | ---------- | ------------- |
| Engine  | ❌     | ❌   | ❌          | ✅             |
| Rule    | ❌     | ❌   | ❌          | ✅             |
| Runtime | ✅     | ✅   | Ephemeral  | ❌             |
| Backend | ✅     | ✅   | Persistent | ❌             |

**Prinsip:**

* **Engine & Rules:** Pure, deterministic, no infrastructure
* **Runtime:** Infrastructure-aware, handles Redis, RPC, state
* **Backend:** Persistent storage, API, alert delivery

---

## 9. Key Principles

### 9.1 Public SDK (`pkg/asentric/`)

* ✅ **Stable API** - Backward compatible, semver-managed
* ✅ **Pure execution** - No Redis, RPC, IO, goroutines
* ✅ **Deterministic** - Same input → same output
* ✅ **Chain-agnostic** - Works with any EVM chain

### 9.2 Internal SDK (`internal/`)

* ✅ **Private** - No external imports allowed
* ✅ **Free to change** - No stability guarantees
* ✅ **Implementation details** - ABI helpers, context implementations

### 9.3 Runtime (`cmd/runtime-reference/`)

* ✅ **Example only** - Not required, just reference
* ✅ **Infrastructure-aware** - Redis, RPC, state management
* ✅ **Scalable** - Worker pools, message queues
* ✅ **Developer choice** - Self-hosted, custom implementation

---

## 10. Final Note

Struktur ini:

* Sudah **final untuk v1**
* Cukup untuk implementasi engine + runtime
* Tidak perlu diubah kecuali ada kebutuhan besar

👉 **Tim boleh mulai implementasi sekarang tanpa ragu.**

---

## Related Documentation

* **[SPEC.md](SPEC.md)** - MVP specification (authoritative)
* **[developer-overview.md](developer-overview.md)** - Alur end-to-end developer
* **[architecture.md](architecture.md)** - Core philosophy & boundaries
* **[sdk-api.md](sdk-api.md)** - Public API reference

