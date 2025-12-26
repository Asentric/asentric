# Asentric SDK – Onboarding Guide

> **Panduan untuk developer baru memulai kontribusi ke Asentric SDK**
>
> **Waktu estimasi:** 2-3 jam untuk setup + pemahaman dasar
>
> **Prerequisite:** Familiar dengan Go, basic blockchain concepts

---

## Daftar Isi

1. [Welcome & Overview](#1-welcome--overview)
2. [Prerequisite & Setup Environment](#2-prerequisite--setup-environment)
3. [Clone & Verify Repository](#3-clone--verify-repository)
4. [Memahami Project Structure](#4-memahami-project-structure)
5. [Memahami Core Concepts](#5-memahami-core-concepts)
6. [First Task: Tulis Rule Sederhana](#6-first-task-tulis-rule-sederhana)
7. [Development Workflow](#7-development-workflow)
8. [Dokumentasi Wajib Baca](#8-dokumentasi-wajib-baca)
9. [Communication & Escalation](#9-communication--escalation)
10. [Checklist Onboarding](#10-checklist-onboarding)

---

## 1. Welcome & Overview

### 1.1 Apa itu Asentric?

Asentric adalah **SDK untuk real-time smart contract security monitoring**. 

**Analogi sederhana:**
- Bayangkan Anda ingin dinotifikasi setiap kali ada transaksi besar di blockchain
- Asentric memungkinkan Anda menulis "rules" untuk mendeteksi pattern tersebut
- Ketika rule terdeteksi, Asentric mengirim alert ke webhook Anda

### 1.2 Target MVP

| Aspek | Keputusan |
|-------|-----------|
| **Scope** | MVP Hackathon |
| **Target User** | Developer self-hosted |
| **Chain Support** | EVM only |
| **Chain per Project** | 1 chain per project |
| **Infrastructure** | Redis + WebSocket RPC + Webhook |

### 1.3 Tim & Timeline

- **Tim:** 1 Lead + 5 Developer
- **Deadline:** 16 Januari 2026
- **Total Sprint:** 5 sprint
- **Model:** Part-time, learning by doing

---

## 2. Prerequisite & Setup Environment

### 2.1 Required Software

| Software | Version | Check Command |
|----------|---------|---------------|
| **Go** | 1.21+ | `go version` |
| **Git** | 2.x+ | `git --version` |
| **Docker** | 20.x+ | `docker --version` |
| **Redis** (via Docker) | 7.x | `docker run redis:7-alpine` |
| **Code Editor** | VSCode/GoLand | - |

### 2.2 Install Go

```bash
# macOS (via Homebrew)
brew install go

# Linux (via apt)
sudo apt update && sudo apt install golang-go

# Verify
go version
# Expected: go version go1.21.x or higher
```

### 2.3 Install Docker

```bash
# macOS
brew install --cask docker

# Linux (Ubuntu/Debian)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Verify
docker --version
```

### 2.4 Start Redis (akan digunakan nanti)

```bash
# Pull dan run Redis
docker run -d -p 6379:6379 --name asentric-redis redis:7-alpine

# Verify
docker ps | grep redis
```

### 2.5 Go Tools (Recommended)

```bash
# Formatter
go install golang.org/x/tools/cmd/goimports@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Test coverage viewer
go install github.com/mattn/goveralls@latest
```

### 2.6 VSCode Extensions (Recommended)

- **Go** (by Go Team at Google) - wajib
- **GitLens** - git history
- **Error Lens** - inline errors
- **YAML** - untuk config files

---

## 3. Clone & Verify Repository

### 3.1 Clone Repository

```bash
# Clone
git clone https://github.com/asentric/asentric.git asentric-sdk
cd asentric-sdk

# Atau jika sudah ada
cd /path/to/asentric-sdk
git pull origin main
```

### 3.2 Install Dependencies

```bash
# Download dependencies
go mod download

# Verify
go mod verify
```

### 3.3 Run Tests (Verify Setup)

```bash
# Run all tests
go test ./...

# Expected: All tests pass (atau PASS jika belum ada tests)
```

### 3.4 Build Verification

```bash
# Verify code compiles
go build ./...

# Expected: No errors
```

---

## 4. Memahami Project Structure

### 4.1 High-Level Structure

```
asentric-sdk/
│
├── pkg/                    # 📦 PUBLIC SDK - Yang user import
│   ├── asentric/          # Core SDK (Engine, Rule, Context, Alert)
│   └── domain/            # Domain types (Transaction, Block, Address)
│
├── internal/              # 🔒 PRIVATE - Implementation details
│   ├── runtime/           # Event loop
│   ├── dispatcher/        # Event → Engine bridge
│   ├── context/           # Concrete Context implementation
│   ├── source/            # ❌ KOSONG - WebSocket source
│   ├── sink/              # ❌ KOSONG - Webhook sink
│   ├── queue/             # ❌ KOSONG - Redis queue
│   ├── abi/               # ❌ KOSONG - ABI decoder
│   ├── chain/             # ❌ KOSONG - Raw chain types
│   ├── adapter/           # ❌ KOSONG - Type converter
│   └── config/            # ❌ KOSONG - YAML loader
│
├── cmd/                   # 🖥️ CLI & Runtime
│   ├── asentric/          # ❌ KOSONG - CLI tool
│   └── runtime-reference/ # ❌ KOSONG - Reference runtime
│
├── templates/             # 📝 Project scaffolding
├── examples/              # 📚 Usage examples
└── docs/                  # 📖 Documentation
```

### 4.2 Apa yang Sudah Ada vs Belum

| Status | Komponen | Lokasi |
|--------|----------|--------|
| ✅ Ready | Engine, Rule, Context, Alert interfaces | `pkg/asentric/` |
| ✅ Ready | Domain types (Transaction, Block, etc.) | `pkg/domain/` |
| ✅ Ready | Internal runtime loop | `internal/runtime/` |
| ✅ Ready | EventContext implementation | `internal/context/` |
| ⚠️ Partial | Dispatcher (struct only) | `internal/dispatcher/` |
| ❌ Empty | WebSocket source | `internal/source/` |
| ❌ Empty | Webhook sink | `internal/sink/` |
| ❌ Empty | Redis queue | `internal/queue/` |
| ❌ Empty | ABI decoder | `internal/abi/` |
| ❌ Empty | CLI | `cmd/asentric/` |

---

## 5. Memahami Core Concepts

### 5.1 Data Flow (WAJIB PAHAM)

```
┌─────────────────────────────────────────────────────────────────┐
│                         DATA FLOW                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [Blockchain]                                                    │
│       │                                                          │
│       │ WebSocket subscription                                   │
│       ▼                                                          │
│  ┌─────────────┐                                                │
│  │ EventSource │  → Produces Events                              │
│  └──────┬──────┘                                                │
│         │                                                        │
│         ▼                                                        │
│  ┌─────────────┐                                                │
│  │  Dispatcher │  → Converts Event to Context                   │
│  └──────┬──────┘                                                │
│         │                                                        │
│         ▼                                                        │
│  ┌─────────────┐    ┌─────────────┐                            │
│  │   Context   │ →  │   Engine    │  → Evaluates Rules          │
│  │ (immutable) │    │ (stateless) │                             │
│  └─────────────┘    └──────┬──────┘                            │
│                            │                                     │
│                            ▼                                     │
│                     ┌─────────────┐                             │
│                     │   Alerts    │  → Rule outputs              │
│                     └──────┬──────┘                             │
│                            │                                     │
│                            ▼                                     │
│                     ┌─────────────┐                             │
│                     │  AlertSink  │  → Sends to Webhook          │
│                     └─────────────┘                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 Core Components

| Component | Responsibility | Pure? |
|-----------|----------------|-------|
| **Engine** | Jalankan semua rules, return alerts | ✅ Yes |
| **Rule** | Logic deteksi, return alert atau nil | ✅ Yes |
| **Context** | Data snapshot untuk rule baca | ✅ Yes (immutable) |
| **Alert** | Output dari rule | ✅ Yes (data only) |
| **EventSource** | Stream events dari blockchain | ❌ No (I/O) |
| **AlertSink** | Kirim alert ke webhook | ❌ No (I/O) |
| **Dispatcher** | Bridge EventSource → Engine | ❌ No |

### 5.3 Key Principles

1. **Engine & Rules = Pure, Deterministic, No I/O**
2. **Context = Immutable snapshot**
3. **Infrastructure (Redis, WebSocket) = Runtime responsibility**
4. **1 Rule = Maksimal 1 Alert per evaluation**

---

## 6. First Task: Tulis Rule Sederhana

### 6.1 Objective

Tulis rule yang mendeteksi transaksi dengan nilai > 1 ETH.

### 6.2 Step-by-Step

**Step 1: Buat file baru**

```bash
mkdir -p examples/first-rule
touch examples/first-rule/large_transfer.go
```

**Step 2: Tulis rule**

```go
// examples/first-rule/large_transfer.go
package firstrule

import (
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
)

// LargeTransferRule detects transfers > threshold
type LargeTransferRule struct {
    Threshold *big.Int
}

// NewLargeTransferRule creates rule with threshold in wei
func NewLargeTransferRule(thresholdWei *big.Int) *LargeTransferRule {
    return &LargeTransferRule{Threshold: thresholdWei}
}

// Name returns unique rule identifier
func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

// Evaluate checks if transaction value exceeds threshold
func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // 1. Get transaction data from context
    tx := ctx.Tx()
    
    // 2. Compare value with threshold
    if tx.Value().Cmp(r.Threshold) > 0 {
        // 3. Create alert
        return &asentric.Alert{
            Rule:        r.Name(),
            Severity:    asentric.SeverityHigh,
            Title:       "Large Transfer Detected",
            Description: "Transaction value exceeds threshold",
            Metadata: map[string]any{
                "value":     tx.Value().String(),
                "threshold": r.Threshold.String(),
                "from":      string(tx.From),
                "to":        string(tx.To),
            },
        }, nil
    }
    
    // 4. No detection = return nil, nil
    return nil, nil
}
```

**Step 3: Tulis test**

```go
// examples/first-rule/large_transfer_test.go
package firstrule

import (
    "math/big"
    "testing"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
)

// MockContext for testing
type mockContext struct {
    tx domain.Transaction
}

func (m *mockContext) ChainID() domain.ChainID     { return 1 }
func (m *mockContext) Tx() domain.Transaction      { return m.tx }
func (m *mockContext) Block() domain.Block         { return domain.Block{} }
func (m *mockContext) Logs() []domain.Log          { return nil }
func (m *mockContext) ABI() domain.ABIRegistry     { return nil }

func TestLargeTransferRule_Detects(t *testing.T) {
    // Threshold: 1 ETH = 1e18 wei
    threshold := big.NewInt(1e18)
    rule := NewLargeTransferRule(threshold)
    
    // Create mock context with 2 ETH transfer
    ctx := &mockContext{
        tx: domain.Transaction{
            RawValue: domain.NativeValue{Wei: "2000000000000000000"}, // 2 ETH
            From:     "0xSender",
            To:       "0xReceiver",
        },
    }
    
    // Evaluate
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if alert == nil {
        t.Fatal("expected alert, got nil")
    }
    if alert.Rule != "large_transfer_detection" {
        t.Errorf("expected rule name 'large_transfer_detection', got '%s'", alert.Rule)
    }
    if alert.Severity != asentric.SeverityHigh {
        t.Errorf("expected severity HIGH, got %s", alert.Severity)
    }
}

func TestLargeTransferRule_NoDetection(t *testing.T) {
    // Threshold: 1 ETH
    threshold := big.NewInt(1e18)
    rule := NewLargeTransferRule(threshold)
    
    // Create mock context with 0.5 ETH transfer
    ctx := &mockContext{
        tx: domain.Transaction{
            RawValue: domain.NativeValue{Wei: "500000000000000000"}, // 0.5 ETH
        },
    }
    
    // Evaluate
    alert, err := rule.Evaluate(ctx)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if alert != nil {
        t.Errorf("expected no alert, got: %+v", alert)
    }
}
```

**Step 4: Run test**

```bash
cd examples/first-rule
go test -v
```

### 6.3 Verification

- [ ] Code compiles tanpa error
- [ ] Tests pass
- [ ] Anda memahami flow Rule → Context → Alert

---

## 7. Development Workflow

### 7.1 Branch Strategy

```bash
# Create feature branch
git checkout -b feature/DEV-X-component-name

# Example
git checkout -b feature/DEV-3-websocket-source
```

### 7.2 Commit Convention

```bash
# Format: type(scope): message

# Examples:
git commit -m "feat(source): implement websocket event source"
git commit -m "fix(dispatcher): handle nil context"
git commit -m "test(engine): add panic recovery tests"
git commit -m "docs(readme): update installation steps"
```

**Types:**
- `feat` - new feature
- `fix` - bug fix
- `test` - adding tests
- `docs` - documentation
- `refactor` - code refactoring
- `chore` - maintenance

### 7.3 Before Commit

```bash
# Format code
goimports -w .

# Run linter
golangci-lint run

# Run tests
go test ./...
```

### 7.4 Pull Request

1. Push branch
2. Create PR dengan template:
   - **What**: Apa yang diubah
   - **Why**: Kenapa diubah
   - **How**: Bagaimana ditest
3. Request review dari Lead
4. Address feedback
5. Merge setelah approved

---

## 8. Dokumentasi Wajib Baca

### 8.1 Priority 1 (Wajib Sebelum Mulai)

| Dokumen | Waktu | Tujuan |
|---------|-------|--------|
| [SPEC.md](SPEC.md) | 30 min | Single Source of Truth |
| [UNDERSTAND-PKG.md](UNDERSTAND-PKG.md) | 45 min | Memahami pkg/asentric |
| [TESTING-GUIDE.md](TESTING-GUIDE.md) | 30 min | Cara testing |

### 8.2 Priority 2 (Setelah Dapat Assignment)

| Dokumen | Tujuan |
|---------|--------|
| [architecture.md](architecture.md) | Core principles |
| [sdk-api.md](sdk-api.md) | API contracts |
| [project-structure.md](project-structure.md) | Folder structure |

### 8.3 Priority 3 (Deep Dive)

| Dokumen | Tujuan |
|---------|--------|
| [internal-runtime-deep-dive.md](internal-runtime-deep-dive.md) | Runtime loop details |
| [runtime-reference-deep-dive.md](runtime-reference-deep-dive.md) | Reference runtime |
| [IMPL-GUIDE.md](IMPL-GUIDE.md) | Implementation guide |

---

## 9. Communication & Escalation

### 9.1 Channels

| Channel | Untuk |
|---------|-------|
| **Daily Standup** | Progress update, blockers |
| **Slack/Discord** | Quick questions |
| **GitHub Issues** | Bug reports, feature requests |
| **PR Comments** | Code-specific discussions |

### 9.2 Kapan Escalate ke Lead

Escalate **segera** jika:

- ❌ Blocked > 2 jam tanpa progress
- ❌ Tidak yakin dengan approach
- ❌ Menemukan bug di komponen lain
- ❌ Konflik dengan dokumentasi
- ❌ Dependency dengan developer lain

### 9.3 Asking Good Questions

**❌ Bad:**
> "Ini ga jalan, gimana ya?"

**✅ Good:**
> "Saya mencoba implement WebSocket subscription di `internal/source/websocket.go`. Ketika subscribe ke `eth_subscribe("logs", ...)`, saya dapat error `connection refused`. 
> 
> Yang sudah saya coba:
> 1. Verify RPC URL valid
> 2. Test dengan curl
> 
> Error message: `dial tcp: connection refused`
> 
> Apakah ada step yang saya lewatkan?"

---

## 10. Checklist Onboarding

### 10.1 Environment Setup

- [ ] Go 1.21+ installed
- [ ] Docker installed
- [ ] Redis running (via Docker)
- [ ] VSCode/GoLand dengan Go extension
- [ ] Repository cloned
- [ ] `go mod download` success
- [ ] `go build ./...` success
- [ ] `go test ./...` success

### 10.2 Understanding

- [ ] Baca SPEC.md
- [ ] Baca UNDERSTAND-PKG.md
- [ ] Baca TESTING-GUIDE.md
- [ ] Pahami Data Flow (Event → Context → Engine → Alert)
- [ ] Pahami perbedaan `pkg/` vs `internal/`

### 10.3 First Task

- [ ] Tulis LargeTransferRule (contoh di atas)
- [ ] Tulis unit test untuk rule
- [ ] Tests pass

### 10.4 Ready to Start

- [ ] Dapat assignment dari Lead
- [ ] Baca SPRINT-X.md overview
- [ ] Baca SPRINT-X-DEV-X.md detail
- [ ] Buat branch feature

---

## Quick Reference

### Useful Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v -run TestName ./path/to/package

# Format code
goimports -w .

# Lint
golangci-lint run

# Build
go build ./...
```

### Key Files to Know

```bash
# Public SDK
pkg/asentric/engine.go      # Engine implementation
pkg/asentric/rule.go        # Rule interface
pkg/asentric/context.go     # Context interface
pkg/asentric/alert.go       # Alert struct

# Domain Types
pkg/domain/transaction.go   # Transaction + Value()
pkg/domain/block.go         # Block struct
pkg/domain/log.go           # Log + Event

# Internal (yang akan kita implement)
internal/runtime/runtime.go     # Event loop
internal/context/context.go     # EventContext
internal/dispatcher/dispatcher.go
```

---

**Selamat bergabung di tim Asentric! 🚀**

Jika ada pertanyaan, jangan ragu untuk bertanya ke Lead.

---

**Last Updated:** December 2024  
**Version:** 1.0

