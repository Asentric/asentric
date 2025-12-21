# Infrastructure Requirements: Setup Awal vs Opsional

## Konsep: Required vs Optional Infrastructure

Berdasarkan referensi **Ponder.sh** yang memerlukan **Postgres untuk setup awal**, kita perlu membedakan:

1. ✅ **Required Infrastructure** - Diperlukan untuk setup awal (seperti Ponder.sh butuh Postgres)
2. ⚠️ **Optional Infrastructure** - Untuk features tambahan (save events/logs)

---

## Perbandingan dengan Ponder.sh

### Ponder.sh Infrastructure

```
Ponder.sh Setup:
├── Required: Postgres (untuk setup awal)
│   - Store schema
│   - Store indexing state
│   - Store processed blocks
│
└── Optional: Database untuk data aplikasi
    - Store events (opsional)
    - Store custom data (opsional)
```

**Konsep:**
- ✅ **Postgres required** - Framework butuh database untuk state management
- ⚠️ **Database optional** - Developer bisa pilih database untuk data aplikasi

---

## Asentric: Infrastructure Requirements

### Opsi 1: Redis Required untuk Setup Awal (Seperti Ponder.sh)

**Jika Redis diperlukan untuk setup awal:**

```
Asentric Setup:
├── Required: Redis (untuk setup awal)
│   - Message queue (watcher → processor)
│   - State management (processed blocks)
│   - Coordination (multi-worker)
│
└── Optional: Database untuk save events/logs
    - PostgreSQL (opsional)
    - MongoDB (opsional)
    - InfluxDB (opsional)
```

**Karakteristik:**
- ✅ **Redis required** - Framework butuh Redis untuk message queue
- ⚠️ **Database optional** - Developer bisa pilih database untuk save events/logs

---

### Opsi 2: Redis Optional (Current Design)

**Jika Redis opsional (current design):**

```
Asentric Setup:
├── Required: None (untuk setup awal)
│   - Framework pure, tidak butuh infrastructure
│   - Runtime bisa pilih infrastructure sendiri
│
├── Recommended: Redis (untuk production)
│   - Message queue (high-volume)
│   - Worker coordination
│
└── Optional: Database untuk save events/logs
    - PostgreSQL (opsional)
    - MongoDB (opsional)
    - InfluxDB (opsional)
```

**Karakteristik:**
- ✅ **No required infrastructure** - Framework pure
- ⚠️ **Redis recommended** - Untuk production (high-volume)
- ⚠️ **Database optional** - Untuk save events/logs

---

## Rekomendasi: Redis Required untuk Setup Awal

### Alasan: Seperti Ponder.sh

**Ponder.sh memerlukan Postgres karena:**
- ✅ State management (processed blocks)
- ✅ Schema storage
- ✅ Coordination (multi-instance)

**Asentric memerlukan Redis karena:**
- ✅ Message queue (watcher → processor)
- ✅ State management (processed blocks)
- ✅ Worker coordination (multi-worker)

---

## Arsitektur: Required vs Optional

### Diagram Infrastructure

```
┌─────────────────────────────────────────────────────────────┐
│              Required Infrastructure (Setup Awal)            │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Redis (Required)                             │   │
│  │                                                       │   │
│  │  Purpose:                                             │   │
│  │  - Message queue (watcher → processor)                │   │
│  │  - State management (processed blocks)                │   │
│  │  - Worker coordination (multi-worker)                 │   │
│  │  - Alert queue (processor → alert handler)            │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ Optional
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Optional Infrastructure                         │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Database (Optional)                         │   │
│  │                                                       │   │
│  │  Purpose:                                             │   │
│  │  - Save events/transactions/logs                     │   │
│  │  - Historical data storage                            │   │
│  │  - Analytics & reporting                              │   │
│  │                                                       │   │
│  │  Options:                                             │   │
│  │  - PostgreSQL (relational)                           │   │
│  │  - MongoDB (document)                                 │   │
│  │  - InfluxDB (time-series)                             │   │
│  │  - ClickHouse (analytics)                            │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Setup Awal: Redis Required

### Step 1: Install Redis

```bash
# Docker (recommended)
docker run -d -p 6379:6379 --name redis redis:7-alpine

# Or install locally
# macOS
brew install redis
redis-server

# Linux
sudo apt-get install redis-server
sudo systemctl start redis
```

### Step 2: Setup Project

```bash
# Install CLI
go install github.com/asentric/asentric@latest

# Generate project
asentric init my-protocol-monitor
cd my-protocol-monitor
```

### Step 3: Configure Redis

```yaml
# config/runtime.yaml (Runtime Config)
runtime:
  redis:
    addr: "localhost:6379"
    password: ""  # Optional
    db: 0
    
  queues:
    transactions: "asentric:transactions"
    alerts: "asentric:alerts"
    
  state:
    stream: "asentric:state"
    consumer_group: "asentric"
```

**Konsep:**
- ✅ **Redis required** - Untuk setup awal
- ✅ **Configuration** - Di runtime config, bukan framework config

---

## Optional: Database untuk Save Events/Logs

### Step 1: Choose Database (Optional)

```bash
# Option 1: PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  --name postgres postgres:15-alpine

# Option 2: MongoDB
docker run -d -p 27017:27017 \
  --name mongo mongo:7

# Option 3: InfluxDB
docker run -d -p 8086:8086 \
  --name influxdb influxdb:2.7
```

### Step 2: Configure Database (Optional)

```yaml
# config/runtime.yaml (Runtime Config)
runtime:
  # Redis (required)
  redis:
    addr: "localhost:6379"
  
  # Database (optional)
  database:
    enabled: true
    type: "postgres"  # postgres, mongodb, influxdb
    
    postgres:
      host: "localhost"
      port: 5432
      database: "asentric"
      user: "postgres"
      password: "postgres"
    
    # Or MongoDB
    mongodb:
      uri: "mongodb://localhost:27017"
      database: "asentric"
    
    # Or InfluxDB
    influxdb:
      url: "http://localhost:8086"
      token: "..."
      org: "asentric"
      bucket: "events"
```

### Step 3: Save Events (Optional)

```go
// cmd/watcher/main.go (Runtime Implementation)
package main

import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
)

type Watcher struct {
    engine   *asentric.Engine
    redis    *redis.Client
    db       *gorm.DB  // Optional
}

func (w *Watcher) processTransaction(tx *Transaction) {
    // 1. Process dengan engine
    ctx := asentric.NewContext(tx)
    alerts, _ := w.engine.Process(ctx)
    
    // 2. Publish alerts to Redis (required)
    for _, alert := range alerts {
        w.redis.Publish("asentric:alerts", alert)
    }
    
    // 3. Save events to database (optional)
    if w.db != nil {
        w.db.Create(&Event{
            TxHash:      tx.Hash,
            BlockNumber:  tx.BlockNumber,
            Alerts:      alerts,
            Timestamp:    time.Now(),
        })
    }
}
```

**Konsep:**
- ✅ **Redis required** - Untuk message queue
- ⚠️ **Database optional** - Untuk save events/logs

---

## Developer Journey: Setup Awal

### Phase 1: Setup Required Infrastructure

```bash
# 1. Install Redis (required)
docker run -d -p 6379:6379 --name redis redis:7-alpine

# 2. Install CLI
go install github.com/asentric/asentric@latest

# 3. Generate project
asentric init my-protocol-monitor
cd my-protocol-monitor

# 4. Configure Redis
vim config/runtime.yaml
```

**Required:**
- ✅ Redis (untuk message queue)

---

### Phase 2: Setup Optional Infrastructure

```bash
# 1. Choose database (optional)
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  --name postgres postgres:15-alpine

# 2. Configure database (optional)
vim config/runtime.yaml
# Enable database section
```

**Optional:**
- ⚠️ Database (untuk save events/logs)

---

## Configuration: Required vs Optional

### Runtime Configuration

```yaml
# config/runtime.yaml
runtime:
  # Required: Redis
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
  
  # Optional: Database
  database:
    enabled: false  # Set to true jika ingin save events
    
    # PostgreSQL
    postgres:
      host: "localhost"
      port: 5432
      database: "asentric"
      user: "postgres"
      password: "postgres"
    
    # MongoDB
    mongodb:
      uri: "mongodb://localhost:27017"
      database: "asentric"
    
    # InfluxDB
    influxdb:
      url: "http://localhost:8086"
      token: "..."
      org: "asentric"
      bucket: "events"
```

**Konsep:**
- ✅ **Redis required** - Harus dikonfigurasi
- ⚠️ **Database optional** - Bisa di-enable jika perlu

---

## Use Cases

### Use Case 1: Minimal Setup (Redis Only)

```yaml
# config/runtime.yaml
runtime:
  redis:
    addr: "localhost:6379"
  
  database:
    enabled: false  # No database
```

**Karakteristik:**
- ✅ Redis untuk message queue
- ❌ Tidak save events/logs
- ✅ Alerts langsung di-handle (webhook, dll)

---

### Use Case 2: Full Setup (Redis + Database)

```yaml
# config/runtime.yaml
runtime:
  redis:
    addr: "localhost:6379"
  
  database:
    enabled: true
    type: "postgres"
    postgres:
      host: "localhost"
      port: 5432
      database: "asentric"
```

**Karakteristik:**
- ✅ Redis untuk message queue
- ✅ Database untuk save events/logs
- ✅ Historical data storage
- ✅ Analytics & reporting

---

## Comparison: Ponder.sh vs Asentric

| Aspek | Ponder.sh | Asentric |
|-------|-----------|----------|
| **Required** | Postgres | Redis |
| **Purpose** | State management | Message queue |
| **Optional** | Database untuk data | Database untuk events |
| **Setup** | `ponder create` butuh Postgres | `asentric init` butuh Redis |

**Kesamaan:**
- ✅ Required infrastructure untuk setup awal
- ✅ Optional database untuk data storage

**Perbedaan:**
- ⚠️ Ponder.sh: Postgres (state management)
- ⚠️ Asentric: Redis (message queue)

---

## Implementation: Runtime Setup

### Example: Runtime dengan Redis Required

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
    
    // 3. Setup watcher (framework handle Redis client)
    watcher := asentric.NewWatcher(engine, config)
    watcher.Start()
}
```

**Catatan:** Framework yang handle Redis client connection. Developer hanya perlu konfigurasi di `runtime.yaml`.

### Example: Runtime dengan Redis + Database (Optional)

```go
// cmd/watcher/main.go
package main

import (
    "github.com/asentric/asentric/pkg/asentric"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // 1. Load runtime config (Redis config ada di runtime.yaml)
    config := loadRuntimeConfig("config/runtime.yaml")
    
    // 2. Setup Database (optional - untuk save events/logs)
    var db *gorm.DB
    if config.Database.Enabled {
        db, _ = gorm.Open(postgres.Open(config.Database.DSN), &gorm.Config{})
    }
    
    // 3. Setup engine
    engine := asentric.NewEngine()
    engine.RegisterRule(&LargeSwapRule{})
    
    // 4. Setup watcher (framework handle Redis client)
    watcher := asentric.NewWatcher(engine, config, db)
    watcher.Start()
}
```

**Catatan:** Framework yang handle Redis client connection. Developer hanya perlu konfigurasi di `runtime.yaml`.

---

## Migration Path

### From No Infrastructure to Redis Required

**Before (Current Design):**
```bash
# Setup tanpa infrastructure
asentric init my-project
# No Redis needed
```

**After (Redis Required):**
```bash
# Setup dengan Redis
docker run -d -p 6379:6379 redis:7-alpine
asentric init my-project
# Redis required
```

**Migration:**
- ✅ Update documentation
- ✅ Update CLI `init` command (check Redis connection)
- ✅ Update runtime template (require Redis config)

---

## Best Practices

### 1. Redis Configuration

```yaml
# config/runtime.yaml
runtime:
  redis:
    addr: "${REDIS_ADDR:localhost:6379}"
    password: "${REDIS_PASSWORD:}"
    db: 0
    
  # Redis connection pool
  pool:
    max_connections: 10
    min_idle: 5
```

---

### 2. Database Configuration (Optional)

```yaml
# config/runtime.yaml
runtime:
  database:
    enabled: "${DATABASE_ENABLED:false}"
    
    # Only configure if enabled
    postgres:
      host: "${POSTGRES_HOST:localhost}"
      port: "${POSTGRES_PORT:5432}"
      database: "${POSTGRES_DB:asentric}"
      user: "${POSTGRES_USER:postgres}"
      password: "${POSTGRES_PASSWORD:}"
```

---

### 3. Environment Variables

```bash
# .env
# Required
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Optional
DATABASE_ENABLED=true
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=asentric
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
```

---

## Summary

### Required Infrastructure

✅ **Redis (Required untuk Setup Awal)**
- Message queue (watcher → processor)
- State management (processed blocks)
- Worker coordination (multi-worker)
- Alert queue (processor → alert handler)

**Seperti Ponder.sh butuh Postgres untuk setup awal.**

---

### Optional Infrastructure

⚠️ **Database (Optional untuk Save Events/Logs)**
- PostgreSQL (relational)
- MongoDB (document)
- InfluxDB (time-series)
- ClickHouse (analytics)

**Developer bisa pilih database sesuai kebutuhan.**

---

### Developer Journey

1. **Setup Redis Server** (required - framework handle client connection)
2. **Generate project** (`asentric init`)
3. **Configure Redis** (runtime.yaml - framework handle connection)
4. **Setup Database** (optional, jika ingin save events)
5. **Run runtime** (watcher + processor)

---

### Key Points

- ✅ **Redis required** - Seperti Ponder.sh butuh Postgres
- ✅ **Framework handle Redis client** - Developer hanya perlu konfigurasi di runtime.yaml
- ⚠️ **Database optional** - Untuk save events/logs
- ✅ **Configuration** - Di runtime config, bukan framework config
- ✅ **Environment variables** - Untuk override config

**Framework handle Redis client connection. Developer hanya perlu setup Redis server dan konfigurasi di runtime.yaml!**

