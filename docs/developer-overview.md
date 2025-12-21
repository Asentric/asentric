# Asentric SDK – Developer Overview

Dokumen ini menjelaskan **alur end-to-end penggunaan Asentric SDK** dari sudut pandang developer. Tujuannya adalah memberikan gambaran besar (big picture) tentang bagaimana Asentric digunakan, tanpa masuk ke detail teknis implementasi.

Asentric dirancang dengan filosofi **developer experience seperti Ponder.sh**: sederhana untuk memulai, fleksibel untuk use case kompleks, dan bersih secara arsitektur.

---

## Tujuan Asentric

Asentric adalah SDK untuk **real-time smart contract security monitoring** yang memungkinkan developer:

* Mendefinisikan **apa yang dimonitor** melalui konfigurasi
* Menulis **logic deteksi sendiri** melalui custom rules
* Menjalankan engine secara lokal atau di runtime mana pun
* Menghasilkan alert yang bersifat semantik dan deterministik

Asentric **bukan SaaS**, dan **bukan rule-engine berbasis YAML**. Asentric adalah **SDK + runtime pattern**.

---

## Gambaran Besar Alur Developer

Secara ringkas, alur penggunaan Asentric adalah:

1. **Setup Redis** (required - seperti Ponder.sh butuh Postgres)
2. **Install & inisialisasi project**
3. **Konfigurasi engine & target monitoring** (YAML)
4. **Menulis custom rules** (Go)
5. **Menjalankan runtime watcher**
6. **Menghasilkan alert**

Alur ini sengaja dibuat linear dan mudah dipahami.

---

## 1. Setup Infrastructure & Instalasi

### Setup Redis Server (Required)

Seperti Ponder.sh memerlukan Postgres untuk setup awal, Asentric memerlukan Redis server untuk message queue dan state management:

```bash
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

**Catatan:** Framework yang handle Redis client connection. Developer hanya perlu:
- Setup Redis server (docker)
- Konfigurasi di `runtime.yaml`
- Framework otomatis connect ke Redis berdasarkan config

### Install CLI

```bash
go install github.com/asentric/asentric@latest
```

### Inisialisasi Project

Developer memulai dengan membuat project baru:

```bash
asentric init my-protocol-monitor
cd my-protocol-monitor
```

Perintah init akan menghasilkan **struktur project standar** yang siap digunakan.

Tujuan tahap ini adalah **zero-friction onboarding**.

---

## 2. Struktur Project

Struktur project hasil inisialisasi:

```
my-protocol-monitor/
├── config/
│   ├── asentric.yaml      # Engine configuration
│   ├── registry.yaml      # Target monitoring list (1 chain per project)
│   └── runtime.yaml        # Runtime configuration (Redis, RPC, database)
├── rules/
│   ├── large_swap.go      # Custom rule
│   └── upgrade.go         # Custom rule
├── abi/
│   └── contracts.json     # Contract ABIs
├── cmd/
│   └── watcher/
│       └── main.go        # Runtime entry point (you build this)
└── fixtures/
    └── example_tx.json     # Example fixture for replay
```

Struktur ini mencerminkan pemisahan concern yang jelas:

* **config/** → deklaratif (apa & bagaimana dijalankan)
* **rules/** → imperatif (logika deteksi)
* **cmd/** → runtime & orchestration

---

## 3. Konfigurasi (YAML)

### asentric.yaml

Digunakan untuk mengatur **bagaimana engine berjalan**, seperti:

* Rules yang diaktifkan
* Opsi per rule
* Batas eksekusi (timeout, max rules)

File ini **tidak mengandung logic deteksi** dan **tidak mengandung RPC endpoints atau chain identity**.

```yaml
# config/asentric.yaml
engine:
  enabled_rules:
    - large_swap_detection
    - suspicious_transfer
  
  rule_options:
    large_swap_detection:
      threshold: "1000000000000000000"  # 1 ETH in wei
  
  execution:
    max_rules_per_context: 100
    timeout_ms: 5000
```

---

### registry.yaml

Digunakan untuk mendefinisikan **apa yang dimonitor**, seperti:

* Smart contract addresses (1 chain per project)
* EOA addresses
* Label dan status enable/disable

Registry bersifat **declarative** dan hanya menjadi input bagi runtime untuk menentukan target monitoring.

```yaml
# config/registry.yaml
targets:
  contracts:
    - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
      name: "Uniswap V3 Router"
      abi_path: "abi/uniswap_v3_router.json"
      enabled: true
```

---

### runtime.yaml

Digunakan untuk mengatur **infrastructure runtime**, seperti:

* Redis configuration (required)
* RPC endpoint (developer choice)
* Database configuration (optional - untuk save events/logs)

```yaml
# config/runtime.yaml
runtime:
  redis:
    addr: "localhost:6379"
  
  rpc:
    endpoint: "https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY"
  
  database:
    enabled: false  # Optional - untuk save events/logs
    type: "postgres"
```

---

## 4. Custom Rules (Go Code)

Semua logic deteksi ditulis sebagai **custom rules dalam Go**:

* Setiap rule mengimplementasikan interface SDK
* Rules bersifat **pure logic**
* Tidak melakukan I/O, network call, atau side-effect

### Contoh Rule

```go
// rules/large_swap.go
package rules

import (
    "math/big"
    "github.com/asentric/asentric/pkg/asentric"
)

type LargeSwapRule struct{}

func (r *LargeSwapRule) Name() string {
    return "large_swap_detection"
}

func (r *LargeSwapRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    threshold := big.NewInt(1000000000000000000) // 1 ETH
    
    if tx.Value().Cmp(threshold) > 0 {
        return &asentric.Alert{
            Severity:    asentric.High,
            Title:       "Large Swap Detected",
            Description: "Transaction value exceeds threshold",
            Metadata: map[string]interface{}{
                "value":     tx.Value().String(),
                "threshold": threshold.String(),
            },
            // Note: Ref (tx_hash, block_number) is populated by engine
        }, nil
    }
    
    return nil, nil
}
```

Contoh rule lainnya:

* Upgrade / proxy change detection
* Suspicious behavior detection
* ML-based anomaly detection (custom ML integration)

Pendekatan ini memberikan fleksibilitas maksimal dan mencegah keterbatasan rule berbasis config.

---

## 5. Rule Registration

Rules diregistrasikan ke engine sebelum runtime dijalankan:

```go
// cmd/watcher/main.go
engine := asentric.NewEngine()
engine.RegisterRule(&LargeSwapRule{})
engine.RegisterRule(&UpgradeRule{})
```

Secara konsep:

> Engine mengetahui kumpulan rules yang harus dijalankan.

---

## 6. Runtime (Watcher)

Runtime adalah **entry point aplikasi**, biasanya berada di:

```
cmd/watcher/main.go
```

### Tanggung Jawab Runtime

* Load runtime config (Redis config dari runtime.yaml - framework handle koneksi)
* Connect ke RPC (developer choice)
* Setup Database (optional - untuk save events/logs)
* Parse config & registry
* Mengambil event / transaksi dari blockchain
* Mengatur concurrency & lifecycle
* Menjalankan engine
* Mengirim alert ke channel yang dikonfigurasi

**Catatan:** Framework yang handle Redis client connection. Developer hanya perlu konfigurasi di `runtime.yaml`.

Engine **tidak tahu** bagaimana data diambil atau alert dikirim.

### Contoh Runtime Implementation

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

**Catatan:** Framework yang handle Redis client. Developer hanya perlu:
- Setup Redis server (docker)
- Konfigurasi di `runtime.yaml`
- Framework otomatis connect ke Redis berdasarkan config

---

## 7. Menjalankan Aplikasi

Developer menjalankan watcher menggunakan:

```bash
go run cmd/watcher/main.go
```

Setelah dijalankan:

* Runtime mulai memonitor blockchain (1 chain per project)
* Engine mengeksekusi rules terhadap transaksi
* Alert dihasilkan ketika rule terpenuhi
* Alert dikirim ke channel yang dikonfigurasi (runtime responsibility)

---

## 8. Testing & Replay

Developer dapat test rules secara offline tanpa infrastructure:

```bash
asentric replay --fixture fixtures/example_tx.json
```

Replay mode:

* **No External Dependencies** — Runs completely offline
* **Deterministic** — Same input always produces same output
* **Safe Iteration** — Test rule changes without affecting production

---

## Filosofi Desain

Asentric dibangun dengan prinsip berikut:

* **YAML untuk konfigurasi, bukan logic**
* **Rules adalah code, bukan config**
* **Engine deterministic & stateless**
* **Runtime bertanggung jawab atas side-effect**
* **Developer bebas menentukan kompleksitas rules**
* **Redis required** (seperti Ponder.sh butuh Postgres)
* **Database optional** (untuk save events/logs)
* **1 project = 1 chain** (chain agnostic, tapi fokus 1 chain)

Pendekatan ini membuat Asentric:

* Mudah dipelajari
* Mudah dites
* Mudah di-debug
* Tidak cepat mentok untuk use case kompleks

---

## Perbandingan dengan Ponder.sh

| Aspek | Ponder.sh | Asentric |
|-------|-----------|----------|
| **Setup** | Postgres required | Redis required |
| **Configuration** | YAML | YAML (config + registry + runtime) |
| **Logic** | TypeScript | Go (custom rules) |
| **Testing** | Local replay | Local replay |
| **Deployment** | Push to Ponder | Self-hosted |
| **Infrastructure** | Managed | Developer choice |

**Kesamaan:**
- ✅ Setup minimal dengan infrastructure required
- ✅ YAML untuk konfigurasi
- ✅ Custom logic (TypeScript vs Go)
- ✅ Local testing/replay
- ✅ Developer experience focus

**Perbedaan:**
- ⚠️ Ponder.sh: Managed infrastructure
- ⚠️ Asentric: Self-hosted (developer choice)

---

## Penutup

Dengan flow ini, developer dapat:

* Setup monitoring hanya dengan konfigurasi minimal
* Menulis logic deteksi sesuai kebutuhan protocol
* Menjalankan engine secara lokal atau terdistribusi
* Test rules secara offline tanpa infrastructure

Dokumen ini menjadi **acuan utama** untuk memahami bagaimana Asentric digunakan dari awal sampai berjalan.

---

## Next Steps

Setelah memahami alur developer, lihat dokumentasi detail:

* **[architecture.md](architecture.md)** - Deep dive arsitektur
* **[sdk-api.md](sdk-api.md)** - Complete API reference
* **[project-structure.md](project-structure.md)** - Final project structure

