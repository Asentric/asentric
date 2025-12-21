# Asentric – Final Project Structure (v1)

Dokumen ini mendefinisikan **struktur folder & file final** untuk seluruh ekosistem Asentric.
Struktur ini adalah **versi jadi (authoritative)** dan menjadi acuan implementasi untuk seluruh tim.

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

## 2. Public SDK – `pkg/asentric/`

**Ini adalah kontrak resmi SDK.**
Semua developer eksternal **hanya boleh** bergantung ke folder ini.

```
pkg/asentric/
├── engine.go        # Engine interface & implementation
├── rule.go          # Rule interface
├── context.go       # Context interface & core models
├── alert.go         # Alert & Severity model
├── config.go        # Engine-level config (non-infra)
├── mock_context.go  # Test helpers (pure)
└── version.go
```

**Prinsip penting:**

* ❌ Tidak ada Redis
* ❌ Tidak ada RPC
* ❌ Tidak ada goroutine
* ❌ Tidak ada IO
* ✅ Pure execution only

---

## 3. Internal SDK – `internal/`

Semua di sini **private**, boleh berubah tanpa breaking change.

```
internal/
├── engine/                 # Engine internals
│   ├── executor.go         # Rule execution loop
│   └── registry.go         # Rule registry
│
├── rule/                   # Rule helpers
│   └── metadata.go
│
├── context/                # Concrete context implementations
│   └── evm_context.go
│
├── chain/                  # Chain data models (EVM, etc)
│   ├── transaction.go
│   ├── block.go
│   └── log.go
│
├── abi/                    # ABI decoding helpers (INTERNAL)
│   ├── loader.go
│   └── decoder.go
│
└── alert/                  # Alert envelope helpers
    └── formatter.go
```

**Catatan:**

* ABI helpers **tidak diexpose** ke rule author
* Semua internal helpers tidak accessible dari luar

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
│   ├── asentric.yaml      # Engine & runtime config
│   ├── registry.yaml      # What to monitor
│   └── runtime.yaml       # Runtime config (Redis, RPC, database)
│
├── rules/
│   └── example_rule.go
│
├── abi/
│   └── example.json
│
└── cmd/
    └── watcher/
        └── main.go
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
├── developer-overview.md    # ⭐ Start here - Alur end-to-end developer
├── architecture.md          # Core philosophy & boundaries
├── sdk-api.md              # Public API reference
├── project-structure.md     # This file - Final structure
├── final-architecture-recommendation.md  # Final architecture decision
└── migration-roadmap.md    # Migration strategy
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

* **[developer-overview.md](developer-overview.md)** - Alur end-to-end developer
* **[architecture.md](architecture.md)** - Core philosophy & boundaries
* **[sdk-api.md](sdk-api.md)** - Public API reference
* **[final-architecture-recommendation.md](final-architecture-recommendation.md)** - Final architecture decision

