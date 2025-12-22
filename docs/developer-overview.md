# Asentric SDK – Developer Overview

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH** untuk hackathon

Dokumen ini menjelaskan **alur end-to-end penggunaan Asentric SDK** dari sudut pandang developer. Tujuannya adalah memberikan gambaran besar (big picture) tentang bagaimana Asentric digunakan, tanpa masuk ke detail teknis implementasi.

Asentric dirancang dengan filosofi **developer experience seperti Ponder.sh**: sederhana untuk memulai, fleksibel untuk use case kompleks, dan bersih secara arsitektur.

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

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
│   ├── asentric.yaml      # Runtime & engine configuration
│   └── registry.yaml      # Target monitoring list (1 chain per project)
├── rules/
│   └── example_rule.go    # Custom rules
├── abi/
│   └── .gitkeep           # Contract ABIs go here
├── cmd/
│   └── watcher/
│       └── main.go        # Runtime entry point
├── go.mod
└── README.md
```

Struktur ini mencerminkan pemisahan concern yang jelas:

* **config/** → deklaratif (apa & bagaimana dijalankan)
* **rules/** → imperatif (logika deteksi)
* **cmd/** → runtime & orchestration

---

## 3. Konfigurasi (YAML)

Asentric menggunakan **2 file konfigurasi**:

### asentric.yaml

Digunakan untuk mengatur **runtime & engine configuration**:

* Chain RPC endpoint (WebSocket)
* Redis configuration
* Webhook URL
* Engine options

```yaml
# config/asentric.yaml

# Chain configuration
chain:
  rpc_ws: "wss://rpc.mantle.xyz/ws"
  name: "Mantle"           # Network name for alerts
  chain_id: 5000           # Optional, auto-detect if not provided

# Redis configuration (required)
redis:
  addr: "localhost:6379"

# Webhook configuration (required)
webhook:
  url: "https://your-webhook.com/alerts"

# Engine configuration (optional)
engine:
  fail_fast: false
```

---

### registry.yaml

Digunakan untuk mendefinisikan **apa yang dimonitor**:

* Smart contract addresses (1 chain per project)
* Contract names
* ABI file paths

```yaml
# config/registry.yaml
targets:
  - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
    name: "Uniswap V3 Router"
    abi_path: "abi/uniswap_v3_router.json"
    
  - address: "0x..."
    name: "My Protocol Vault"
    abi_path: "abi/vault.json"
```

---

## 4. Custom Rules (Go Code)

Semua logic deteksi ditulis sebagai **custom rules dalam Go**:

* Setiap rule mengimplementasikan interface SDK
* Rules bersifat **pure logic**
* Tidak melakukan I/O, network call, atau side-effect

### Contoh Rule

```go
// rules/large_transfer.go
package rules

import (
    "math/big"
    "github.com/asentric/asentric/pkg/asentric"
)

type LargeTransferRule struct {
    Threshold *big.Int
}

func NewLargeTransferRule(threshold *big.Int) *LargeTransferRule {
    return &LargeTransferRule{Threshold: threshold}
}

func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    // tx.Value() returns *big.Int for easy comparison
    if tx.Value().Cmp(r.Threshold) > 0 {
        return asentric.NewAlert(
            r.Name(),
            asentric.SeverityHigh,
            "Large Transfer Detected",
            "Transaction value exceeds threshold",
        ).WithMetadata("value", tx.Value().String()).
          WithMetadata("threshold", r.Threshold.String()).
          WithMetadata("from", tx.From.String()).
          WithMetadata("to", tx.To.String()), nil
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
engine.RegisterRule(rules.NewLargeTransferRule(big.NewInt(1e18)))
engine.RegisterRule(&rules.UpgradeDetectionRule{})
```

Secara konsep:

> Engine mengetahui kumpulan rules yang harus dijalankan.

---

## 6. Runtime

Runtime adalah **entry point aplikasi**, berada di:

```
cmd/watcher/main.go
```

### Developer Experience Target

```go
// cmd/watcher/main.go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
    "my-project/rules"
)

func main() {
    // 1. Load configuration (from config/ directory)
    config, err := asentric.LoadConfig("config/")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Create engine and register rules
    engine := asentric.NewEngine()
    engine.RegisterRule(rules.NewLargeTransferRule(big.NewInt(1e18)))
    engine.RegisterRule(&rules.UpgradeDetectionRule{})
    
    // 3. Create runtime and start
    runtime := asentric.NewRuntime(config, engine)
    
    // 4. Run (blocks until SIGINT/SIGTERM)
    if err := runtime.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Framework Handles

- ✅ WebSocket subscription to chain
- ✅ Redis queue management
- ✅ Context building from events
- ✅ Engine evaluation
- ✅ Webhook alert delivery
- ✅ Graceful shutdown (SIGINT/SIGTERM)

### Developer Provides

- ✅ Configuration files
- ✅ Custom rules
- ✅ ABI files

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

