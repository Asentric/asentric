# Rekomendasi Arsitektur Final: 1 Repository (Monorepo)

## Keputusan Final

✅ **1 Repository (Monorepo)** - Framework + CLI + Examples + Reference Runtime

**Alasan:**
- ✅ Developer experience (satu tempat untuk semua)
- ✅ Open source (satu repo untuk contribute)
- ✅ Maintenance (framework + examples selalu sync)
- ✅ Reference runtime sebagai contoh (tidak required)

---

## Struktur Repository Final

> **📖 Lihat struktur detail:** [project-structure.md](project-structure.md) - Struktur folder & file final (authoritative)

```
asentric/
│
├── pkg/asentric/              # PUBLIC SDK (Stable API)
│   ├── engine.go              # Engine interface & implementation
│   ├── rule.go                # Rule interface
│   ├── context.go             # Context interface & core models
│   ├── alert.go               # Alert & Severity model
│   ├── config.go              # Engine-level config (non-infra)
│   ├── mock_context.go        # Test helpers (pure)
│   └── version.go
│
├── internal/                  # PRIVATE SDK implementation
│   ├── engine/                # Engine internals
│   ├── rule/                  # Rule helpers
│   ├── context/               # Concrete context implementations
│   ├── chain/                 # Chain data models (EVM, etc)
│   ├── abi/                   # ABI decoding helpers (INTERNAL)
│   └── alert/                 # Alert envelope helpers
│
├── cmd/
│   ├── asentric/              # CLI Tools
│   │   ├── main.go
│   │   ├── init.go            # asentric init
│   │   ├── replay.go          # offline deterministic replay
│   │   ├── version.go
│   │   └── internal/
│   │       └── templates.go
│   │
│   └── runtime-reference/     # Reference Runtime (Example)
│       ├── main.go             # Entry point runtime
│       ├── config/             # Load yaml config
│       ├── ingest/             # Subscribe logs/blocks
│       ├── pipeline/           # Dispatcher & workers
│       ├── state/              # RUNTIME STATE (Redis here)
│       ├── alert/              # Alert delivery
│       └── runtime.go          # Glue code
│
├── templates/
│   └── project/               # Project templates
│       ├── config/
│       │   ├── asentric.yaml
│       │   ├── registry.yaml
│       │   └── runtime.yaml
│       ├── rules/
│       ├── abi/
│       └── cmd/watcher/
│
├── examples/
│   ├── simple-watcher/        # Minimal runtime example
│   ├── custom-rules/          # Example custom rules
│   ├── ml-integration/        # Example ML rule
│   └── multi-chain/          # Example multi-chain setup
│
├── docs/
│   ├── developer-overview.md
│   ├── architecture.md
│   ├── sdk-api.md
│   ├── project-structure.md   # Final structure (authoritative)
│   ├── final-architecture-recommendation.md
│   └── migration-roadmap.md
│
├── .github/
│   ├── workflows/             # CI/CD
│   └── ISSUE_TEMPLATE/
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## Komponen Repository

### 1. Framework Core (`pkg/asentric/`)

**Public API (STABLE):**
- `Engine` - Core engine interface
- `Rule` - Rule interface
- `Context` - Execution context
- `Alert` - Alert model

**Karakteristik:**
- ✅ Pure, deterministic
- ✅ Infrastructure-agnostic
- ✅ Chain-agnostic (EVM)
- ✅ Stable API contract

---

### 2. CLI Tools (`cmd/asentric/`)

**Commands:**
- `asentric init <project>` - Generate project template
- `asentric replay --fixture <file>` - Test offline
- `asentric version` - Version info

**Karakteristik:**
- ✅ Developer tool only
- ✅ No infrastructure needed
- ✅ Generate templates
- ✅ Test rules offline

---

### 3. Reference Runtime (`cmd/runtime-reference/`)

**Purpose:**
- ✅ Example implementation
- ✅ Reference untuk developer
- ✅ Tidak required (developer bisa buat sendiri)

**Karakteristik:**
- ✅ Minimal implementation
- ✅ Self-hosted example
- ✅ Redis required (message queue)
- ✅ Database optional (save events/logs)
- ✅ Documented

---

### 4. Examples (`examples/`)

**Examples:**
- `simple-watcher/` - Minimal runtime
- `custom-rules/` - Custom rule examples
- `ml-integration/` - ML rule example
- `multi-chain/` - Multi-chain setup (advanced)

**Karakteristik:**
- ✅ Real-world examples
- ✅ Documented
- ✅ Always up-to-date with framework

---

### 5. Templates (`templates/project/`)

**Purpose:**
- ✅ Project templates untuk `asentric init`
- ✅ Config templates (asentric.yaml, registry.yaml, runtime.yaml)
- ✅ Rule templates
- ✅ Runtime templates

**Karakteristik:**
- ✅ Used by CLI
- ✅ Minimal, clean
- ✅ Documented

---

## Developer Journey

### 1. Setup

```bash
# 1. Setup Redis (Required - seperti Ponder.sh butuh Postgres)
docker run -d -p 6379:6379 --name redis redis:7-alpine

# 2. Install CLI
go install github.com/asentric/asentric@latest

# 3. Generate project
asentric init my-protocol-monitor
cd my-protocol-monitor
```

**Output:**
- ✅ Project structure generated
- ✅ Config templates (asentric.yaml, registry.yaml, runtime.yaml)
- ✅ Example rule
- ✅ Runtime template (reference)

---

### 2. Development

```bash
# Write custom rules
vim rules/large_swap.go
vim rules/ml_anomaly.go  # ML integration

# Test offline
asentric replay --fixture fixtures/test_tx.json

# Look at examples (optional)
# - examples/simple-watcher
# - examples/ml-integration
# - cmd/runtime-reference
```

**Konsep:**
- ✅ Developer write rules (pure Go)
- ✅ Test offline (no infrastructure)
- ✅ Examples sebagai reference (optional)

---

### 3. Runtime Implementation

```go
// cmd/watcher/main.go
package main

import (
    "github.com/asentric/asentric/pkg/asentric"
)

func main() {
    // 1. Load runtime config (Redis config ada di runtime.yaml)
    config := loadRuntimeConfig("config/runtime.yaml")
    
    // 2. Setup engine
    engine := asentric.NewEngine()
    engine.RegisterRule(&LargeSwapRule{})
    
    // 3. Setup Database (optional - untuk save events/logs)
    // var db *gorm.DB
    // if config.Database.Enabled { ... }
    
    // 4. Start monitoring (framework handle Redis client)
    watcher := asentric.NewWatcher(engine, config)
    watcher.Start()
}
```

**Konsep:**
- ✅ **Redis required** - Untuk message queue & state management
- ✅ **Framework handle Redis client** - Developer hanya perlu konfigurasi di runtime.yaml
- ✅ Developer buat runtime sendiri (self-hosted)
- ✅ **Database optional** - Untuk save events/logs
- ✅ Framework handle Redis connection, developer tidak perlu setup manual

---

### 4. Deployment

```bash
# Developer deploy runtime sendiri
# - Self-hosted (VPS, cloud, dll)
# - Infrastructure: Redis (required) + Database (optional)

# Runtime akan:
# 1. Load runtime config (framework handle Redis client connection)
# 2. Load config & registry
# 3. Connect to RPC (1 chain per project)
# 4. Setup Database (optional - untuk save events/logs)
# 5. Setup engine
# 6. Register rules
# 7. Start monitoring
```

---

## Infrastructure Requirements

### Required: Redis (Setup Awal)

**Seperti Ponder.sh butuh Postgres, Asentric butuh Redis untuk:**
- ✅ Message queue (watcher → processor)
- ✅ State management (processed blocks)
- ✅ Worker coordination (multi-worker)
- ✅ Alert queue (processor → alert handler)

**Setup:**
```bash
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

---

### Optional: Database (Save Events/Logs)

**Developer bisa pilih database untuk:**
- ⚠️ Save events/transactions/logs
- ⚠️ Historical data storage
- ⚠️ Analytics & reporting

**Options:**
- PostgreSQL (relational)
- MongoDB (document)
- InfluxDB (time-series)
- ClickHouse (analytics)

**Lihat detail:** [developer-overview.md](developer-overview.md#1-setup-infrastructure--instalasi)

---

## Perbandingan dengan Ponder.sh

| Aspek | Ponder.sh | Asentric |
|-------|-----------|----------|
| **Repository** | 1 repo (monorepo) | ✅ 1 repo (monorepo) |
| **Framework** | ✅ Core framework | ✅ Core framework |
| **CLI** | ✅ CLI tools | ✅ CLI tools |
| **Examples** | ✅ Examples | ✅ Examples |
| **Runtime** | Managed (Ponder) | Self-hosted (developer) |
| **Required Infrastructure** | Postgres (setup awal) | Redis (setup awal) |
| **Optional Infrastructure** | Database untuk data | Database untuk events/logs |
| **Deployment** | Push to Ponder | Self-hosted |
| **Open Source** | ✅ | ✅ |

**Kesamaan:**
- ✅ 1 repo (monorepo)
- ✅ Framework + CLI + Examples
- ✅ Developer experience focus
- ✅ Required infrastructure untuk setup awal
- ✅ Optional database untuk data storage

**Perbedaan:**
- ⚠️ Ponder.sh: Managed infrastructure
- ⚠️ Asentric: Self-hosted (developer choice)
- ⚠️ Ponder.sh: Postgres required (state management)
- ⚠️ Asentric: Redis required (message queue & state management)

---

## Key Principles

- ✅ **1 Repository (Monorepo)** - Framework + CLI + Examples + Reference Runtime
- ✅ **Self-hosted** - Developer buat runtime sendiri
- ✅ **Redis required** - Untuk setup awal (seperti Ponder.sh butuh Postgres)
- ✅ **Database optional** - Untuk save events/logs (developer choice)
- ✅ **1 project = 1 chain** - Chain agnostic, tapi fokus 1 chain
- ✅ **ML sebagai custom rule** - Developer implement sendiri
- ✅ **Open source** - Tidak ada SaaS, fokus build open source
- ✅ **Reference runtime** - Contoh, bukan requirement

---

## Roadmap Implementasi

### Phase 1: Framework Core (Weeks 1-8)
- ✅ Core engine
- ✅ Rule system
- ✅ Context & Alert models
- ✅ Internal implementation

### Phase 2: CLI Tools (Weeks 9-13)
- ✅ `asentric init` - Project scaffolding
- ✅ `asentric replay` - Offline testing
- ✅ Templates
- ✅ Documentation

### Phase 3: Reference Runtime (Weeks 14-16)
- ✅ Minimal runtime example (dengan Redis required)
- ✅ Documentation
- ✅ Examples

### Phase 4: Examples (Weeks 17+)
- ✅ Examples (simple-watcher, custom-rules, ml-integration)

---

## Next Steps

1. ✅ Finalize repository structure (1 repo - monorepo)
2. ✅ Implement framework core
3. ✅ Implement CLI tools
4. ✅ Create reference runtime (dengan Redis required)
5. ✅ Create examples
6. ✅ Document infrastructure requirements

---

## Kesimpulan

**Keputusan Final: 1 Repository (Monorepo)**

- ✅ Framework core
- ✅ CLI tools
- ✅ Reference runtime
- ✅ Examples
- ✅ Templates
- ✅ Documentation

**Semua dalam satu repository untuk developer experience yang optimal!**
