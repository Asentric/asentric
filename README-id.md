# Asentric SDK

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**SDK Go untuk membangun logika deteksi keamanan on-chain real-time dengan cara yang modular, deterministik, dan ramah developer.**

Asentric SDK menyediakan execution engine murni, sistem rule, dan runtime context eksplisit yang memungkinkan developer menulis aturan keamanan smart contract tanpa terikat pada concern infrastruktur seperti message queue, database, API, atau sistem deployment.

---

## Overview

Asentric SDK dirancang untuk menjadi **shared security brain** dari ekosistem Asentric. SDK ini menyediakan lapisan murni dan infrastructure-agnostic untuk mendefinisikan logika deteksi keamanan yang dapat di-embed ke berbagai lingkungan runtime.

SDK ini memisahkan logika keamanan dari concern infrastruktur, memungkinkan developer fokus menulis aturan deteksi yang efektif sementara sistem infrastruktur menangani deployment, ingestion, dan pengiriman alert.

---

## Core Capabilities

Asentric SDK memberdayakan developer untuk:

- **Define Pure, Deterministic Security Rules** — Menulis logika deteksi sebagai pure functions tanpa side effects
- **Process On-Chain Data** — Menganalisis transaksi, log, dan event smart contract yang sudah di-decode
- **Generate Structured Alerts** — Menghasilkan alert keamanan terstandar dengan tingkat severity dan metadata
- **Enable Deterministic Replay** — Debug dan test rules secara offline dengan jaminan reproducibility
- **Build Infrastructure-Agnostic Logic** — Tulis rules sekali, deploy di mana saja tanpa dependensi infrastruktur
- **Maintain Developer Focus** — Fokus pada logika keamanan tanpa khawatir tentang deployment atau infrastruktur

---

## What Asentric SDK Is — and Is Not

### ✅ What This SDK Is

Asentric SDK adalah:

- **Security detection engine** untuk transaksi dan event blockchain
- **Rule execution framework** dengan API yang bersih dan dapat di-test
- **Developer-focused SDK** yang memprioritaskan kesederhanaan dan kejelasan
- **Pure domain logic layer** tanpa coupling dengan infrastruktur
- **Shared core** yang digunakan oleh berbagai sistem runtime dalam ekosistem Asentric

### ❌ What This SDK Is Not

Secara desain, Asentric SDK **tidak** menyediakan:

- Message queue (Redis, Kafka, RabbitMQ, dll.)
- HTTP API atau web dashboard
- Koneksi database atau persistence layer
- Container orchestration atau deployment tools
- Sistem notifikasi (Slack, webhooks, email, dll.)
- Koneksi RPC atau manajemen blockchain node

Tanggung jawab ini ditangani oleh komponen khusus dalam ekosistem Asentric:

| Repository | Responsibility |
|------------|----------------|
| `asentric-bot` | Runtime watcher, chain data ingestion, alert delivery |
| `asentric-backend` | API server, alert aggregation, persistence layer |
| `asentric-frontend` | Web-based dashboard and visualization |

**SDK tetap menjadi pure security detection engine.**

---

## Design Principles

### 1. Pure Domain Logic First

- Rules diimplementasikan sebagai **pure functions**
- **No side effects** — rules tidak memodifikasi state eksternal
- **No I/O operations** di dalam evaluasi rule
- Mudah di-test, di-reason, dan di-debug

### 2. Infrastructure-Agnostic by Default

- SDK memiliki **zero knowledge** tentang Redis, HTTP server, atau database
- Alert **diproduksi**, bukan dikirim — pengiriman adalah concern infrastruktur
- Semua integrasi infrastruktur berada **di luar** SDK

### 3. Explicit Context

- Semua data eksekusi mengalir melalui objek **Context**
- **No global state** — semuanya eksplisit dan dapat ditelusuri
- **Deterministic execution** — input yang sama selalu menghasilkan output yang sama

### 4. Developer Experience (DX)

- **Minimal boilerplate** — mulai cepat dengan default yang masuk akal
- **Idiomatic Go** — mengikuti best practices dan konvensi Go
- **Library-first, CLI-assisted workflow** — scaffold, test, dan replay dari command line
- **Easy local testing** — test rules tanpa dependensi eksternal

> **Important:** CLI tidak menjalankan production watchers atau terhubung ke blockchain. CLI adalah tool developer untuk scaffolding, replay, dan validasi rule.

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Asentric SDK                     │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Engine  │  │  Rules   │  │ Context  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│         │             │             │           │
│         └─────────────┴─────────────┘           │
│                     │                           │
└─────────────────────┼───────────────────────────┘
                      │
         ┌────────────┼────────────┐
         │            │            │
    ┌────▼───┐   ┌───▼────┐   ┌──▼──────┐
    │  Bot   │   │Backend │   │Frontend │
    │(Runtime)│  │  (API) │   │  (UI)   │
    └────────┘   └────────┘   └─────────┘
```

SDK menyediakan **core detection logic**, sementara sistem eksternal menangani:

- **Bot**: Chain monitoring, transaction ingestion, alert routing
- **Backend**: REST API, data aggregation, persistence
- **Frontend**: User interface, dashboards, visualization

---

## Quick Start

### Prerequisites

- Go 1.21 atau lebih tinggi
- Pemahaman dasar tentang transaksi blockchain dan smart contract

### Installation

Install SDK menggunakan Go modules:

```bash
go get github.com/asentric/asentric-sdk
```

### Scaffold a New Project

Buat project deteksi baru dengan CLI:

```bash
asentric init my-asentric-protocol
cd my-asentric-protocol
go mod tidy
```

Ini akan menghasilkan struktur project lengkap:

```
my-asentric-protocol/
├── cmd/
│   └── watcher/
│       └── main.go          # Entry point
├── rules/                   # Your security rules
├── abi/                     # Smart contract ABIs
├── config/
│   └── asentric.yaml        # Configuration
└── README.md
```

### Initialize the Engine

Di `main.go` Anda:

```go
package main

import (
    "github.com/asentric/asentric-sdk/pkg/asentric"
)

func main() {
    // Create a new detection engine
    engine := asentric.NewEngine()
    
    // Register your security rules
    engine.RegisterRule(&LargeSwapRule{})
    engine.RegisterRule(&SuspiciousTransferRule{})
    
    // Process transactions (context provided by runtime)
    // engine.Process(ctx)
}
```

---

## Writing Security Rules

Security rules mengimplementasikan interface sederhana:

```go
type Rule interface {
    Name() string
    Evaluate(ctx Context) (*Alert, error)
}
```

### Example: Large Swap Detection

```go
package rules

import (
    "github.com/asentric/asentric-sdk/pkg/asentric"
)

type LargeSwapRule struct{}

func (r *LargeSwapRule) Name() string {
    return "large_swap_detection"
}

func (r *LargeSwapRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // Access transaction data from context
    tx := ctx.Tx()
    
    // Define detection logic
    threshold := big.NewInt(1000)
    if tx.Value().Cmp(threshold) > 0 {
        return &asentric.Alert{
            Severity:    asentric.High,
            Title:       "Large Swap Detected",
            Description: "Transaction value exceeds threshold",
            Metadata: map[string]interface{}{
                "value":     tx.Value().String(),
                "threshold": threshold.String(),
            },
            // Note: Ref (tx_hash, block_number) is populated by the engine, not by rules
        }, nil
    }
    
    // No alert — transaction is normal
    return nil, nil
}
```

### Key Principles for Rules

1. **Pure Functions** — No side effects, no external I/O
2. **Context-Based Data** — Semua data transaksi berasal dari `Context`
3. **Explicit Returns** — Return `nil` ketika tidak ada alert yang diperlukan
4. **Structured Alerts** — Gunakan struct `Alert` untuk konsistensi
5. **Error Handling** — Return error untuk kegagalan processing, bukan detection miss

### Alert Structure

Alert dapat menyertakan execution reference minimal (transaction hash dan block number) untuk debugging dan traceability. Reference ini:

* **Diisi oleh engine**, bukan oleh rules
* **Hanya informasional** — tidak menyiratkan tanggung jawab routing, persistence, atau delivery
* **Opsional** — rules tidak perlu (dan tidak bisa) mengesetnya

`ExecutionRef` hanya berisi `tx_hash` dan `block_number`. Tidak termasuk chain identity, informasi network, atau RPC endpoints. Chain identity tetap menjadi tanggung jawab sistem runtime.

---

## Testing Rules

Rules mudah di-test karena mereka adalah pure functions tanpa dependensi eksternal:

```go
package rules

import (
    "math/big"
    "testing"
    
    "github.com/asentric/asentric-sdk/pkg/asentric"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLargeSwapRule_TriggersOnLargeValue(t *testing.T) {
    // Arrange
    ctx := mockContextWithValue(big.NewInt(2000))
    rule := &LargeSwapRule{}
    
    // Act
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    require.NoError(t, err)
    require.NotNil(t, alert)
    assert.Equal(t, asentric.High, alert.Severity)
}

func TestLargeSwapRule_NoAlertOnSmallValue(t *testing.T) {
    // Arrange
    ctx := mockContextWithValue(big.NewInt(500))
    rule := &LargeSwapRule{}
    
    // Act
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    require.NoError(t, err)
    assert.Nil(t, alert, "No alert should be generated for small values")
}

// Helper function to create test contexts
func mockContextWithValue(value *big.Int) asentric.Context {
    return asentric.NewMockContext(
        asentric.WithTxValue(value),
    )
}
```

**Tidak perlu mock untuk Redis, HTTP, atau database** — cukup test logika Anda.

---

## Replay Mode

Asentric SDK mendukung **offline, deterministic replay** untuk debugging dan testing.

### Running Replay

```bash
asentric replay --fixture fixtures/suspicious_tx.json
```

### Replay Guarantees

- **No External Dependencies** — Berjalan sepenuhnya offline
- **Deterministic** — Input yang sama selalu menghasilkan output yang sama
- **Safe Iteration** — Test perubahan rule tanpa mempengaruhi production
- **Historical Analysis** — Replay transaksi masa lalu untuk memvalidasi logika deteksi

### Creating Replay Fixtures

Replay fixtures adalah file JSON yang berisi data transaksi:

```json
{
  "chain_id": 1,
  "block_number": 12345678,
  "tx_hash": "0xabc...",
  "from": "0x123...",
  "to": "0x456...",
  "value": "1000000000000000000",
  "data": "0x...",
  "logs": [...]
}
```

> **Important:** SDK tidak akan pernah mengambil data historis sendiri. Mengambil data transaksi live dari RPC nodes adalah tanggung jawab sistem runtime (misalnya, `asentric-bot`), bukan SDK.

---

## Ecosystem Integration

Asentric SDK dirancang untuk di-embed ke berbagai lingkungan runtime:

```
┌─────────────┐
│  Asentric   │ ◄── Core detection logic
│     SDK     │     (pure, reusable)
└─────────────┘
       │
       ├──► Bot        (ingestion, real-time monitoring)
       ├──► Backend    (API, aggregation, persistence)
       ├──► CLI        (replay, testing, development)
       └──► Lambda     (serverless detection)
```

### Component Roles

| Component | Role | Uses SDK For |
|-----------|------|--------------|
| **SDK** | Security logic & detection | N/A (core library) |
| **Bot** | Chain ingestion & alert delivery | Rule execution, alert generation |
| **Backend** | API, aggregation, persistence | Alert processing, historical analysis |
| **Frontend** | Visualization & dashboard | N/A (consumes Backend API) |
| **CLI** | Development & testing tools (not a runtime) | Replay, rule validation |

**SDK tetap menjadi single source of truth untuk security detection logic.**

---

## Repository Structure

```
asentric-sdk/
├── cmd/
│   └── asentric/              # CLI tools (init, replay)
│       └── main.go
│
├── pkg/
│   └── asentric/              # PUBLIC SDK API (STABLE)
│       ├── engine.go          # Detection engine
│       ├── rule.go            # Rule interface
│       ├── context.go         # Execution context
│       ├── alert.go           # Alert model
│       └── config.go          # SDK configuration
│
├── internal/                  # PRIVATE SDK IMPLEMENTATION
│   ├── runtime/               # Engine runtime loop
│   ├── rule/                  # Rule registry & executor
│   ├── chain/                 # Chain data models & helpers
│   ├── abi/                   # ABI loading & decoding
│   ├── alert/                 # Alert formatting & envelope
│   └── observability/         # Internal execution metrics & diagnostics
│
├── templates/
│   └── project/               # Project templates for `asentric init`
│
├── examples/
│   └── simple-watcher/        # Minimal SDK usage example
│
├── docs/
│   ├── architecture.md        # Architecture deep dive
│   ├── sdk-api.md             # Complete API reference
│   └── cli.md                 # CLI documentation
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### Key Directories

- **`pkg/asentric/`** — Public, stable SDK API yang digunakan oleh developer
- **`internal/`** — Detail implementasi private, dapat berubah
- **`cmd/asentric/`** — CLI tools untuk scaffolding dan testing (bukan runtime)
- **`templates/`** — Project templates untuk quick start
- **`examples/`** — Contoh penggunaan SDK yang berfungsi

**Observability Note:** Observability di SDK terbatas pada metrik eksekusi internal dan diagnostik (misalnya, rule execution timing, performance counters). Mengekspor metrik atau log ke sistem eksternal (Prometheus, OpenTelemetry, dll.) adalah tanggung jawab runtime.

---

## Roadmap

Kami terus meningkatkan Asentric SDK. Fitur yang akan datang meliputi:

- [ ] **Rule Grouping & Tagging** — Mengorganisir rules berdasarkan protocol, tingkat risiko, atau kategori
- [ ] **Multi-Chain Support** — Dukungan built-in untuk chain yang kompatibel dengan EVM
- [ ] **Enhanced ABI Decoding** — Decoding event otomatis dan type safety
- [ ] **Community Rule Registry** — Library bersama dari aturan deteksi open-source
- [ ] **Advanced Replay Features** — Time-travel debugging dan batch replay
- [ ] **Performance Profiling** — Analisis performa rule built-in

### Long-Term Research Ideas

Fitur-fitur berikut sedang diteliti dan mungkin dieksplorasi di versi mendatang:

- [ ] **ZK-Friendly Rule Outputs** — Menghasilkan alert proofs yang kompatibel dengan zero-knowledge
- [ ] **WASM-Based Rule Sandbox** — Menjalankan rules yang tidak terpercaya di lingkungan WebAssembly yang terisolasi

---

## Contributing

Kontribusi sangat diterima! Kami terutama menghargai bantuan dengan:

- **Rule Examples** — Bagikan logika deteksi Anda dengan komunitas
- **ABI Decoding** — Tingkatkan dukungan untuk tipe event yang kompleks
- **Developer Tooling** — Tingkatkan fitur CLI dan peningkatan DX
- **Documentation** — Bantu membuat SDK lebih mudah diakses
- **Testing** — Tambahkan test coverage dan skenario edge case

### Getting Started

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-rule`)
3. Tulis tests untuk perubahan Anda
4. Pastikan semua tests pass (`go test ./...`)
5. Commit perubahan Anda (`git commit -m 'Add amazing detection rule'`)
6. Push ke branch Anda (`git push origin feature/amazing-rule`)
7. Buka Pull Request

### Code Standards

- Ikuti [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Tulis tests untuk fitur baru
- Update dokumentasi sesuai kebutuhan
- Jaga rules tetap pure dan deterministic

---

## License

MIT License © Asentric

Lihat [LICENSE](LICENSE) untuk detail lengkap.

---

## Resources

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issue Tracker**: [GitHub Issues](https://github.com/asentric/asentric-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/asentric/asentric-sdk/discussions)

---

**Built with ❤️ by the Asentric Team**

