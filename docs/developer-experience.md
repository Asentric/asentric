# Developer Experience - Asentric SDK

Dokumen ini menjelaskan developer experience Asentric SDK yang terinspirasi dari Ponder.sh: **setup minimal, fokus pada logika, tanpa khawatir infrastruktur**.

---

## Filosofi Developer Experience

### Prinsip Utama

1. **YAML untuk Setup** — Konfigurasi dan target list, bukan logika
2. **Go untuk Rules** — Logika deteksi sebagai pure functions
3. **Engine Pure & Deterministic** — Kotak hitam yang dapat di-test dan di-replay
4. **Runtime Menangani Infrastruktur** — Developer tidak perlu memikirkan deployment

### Workflow: Write → Test → Deploy

```
Write Rules (Go) → Test Locally (Replay) → Deploy (Runtime)
```

---

## Quick Start: Developer Journey

### Step 1: Setup Project

```bash
# Scaffold project baru
asentric init my-protocol-monitor
cd my-protocol-monitor

# Struktur project
my-protocol-monitor/
├── config/
│   ├── asentric.yaml      # Engine config
│   └── registry.yaml      # Target monitoring list
├── rules/
│   └── large_swap.go      # Custom rules
├── abi/
│   └── uniswap_v3.json    # Contract ABIs
└── cmd/
    └── watcher/
        └── main.go        # Runtime entry point
```

### Step 2: Konfigurasi YAML

#### `config/asentric.yaml` - Engine Configuration

```yaml
# Engine behavior configuration
engine:
  # Rules yang diaktifkan
  enabled_rules:
    - large_swap_detection
    - suspicious_transfer
    - flash_loan_attack
  
  # Opsi per rule
  rule_options:
    large_swap_detection:
      threshold: "1000000000000000000"  # 1 ETH in wei
      severity: "high"
    
    suspicious_transfer:
      min_value: "10000000000000000000"  # 10 ETH
      check_blacklist: true
  
  # Batas eksekusi
  execution:
    max_rules_per_context: 100
    timeout_ms: 5000

# Catatan penting:
# - RPC endpoints TIDAK ada di sini (di runtime)
# - Chain ID TIDAK ada di sini (di runtime)
# - Network info TIDAK ada di sini (di runtime)
# Ini adalah config engine behavior, bukan infrastruktur
```

#### `config/registry.yaml` - Target Monitoring List

```yaml
# Daftar target yang akan dimonitor
targets:
  # Smart contract addresses
  contracts:
    - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"  # Uniswap V3 Router
      name: "Uniswap V3 Router"
      abi_path: "abi/uniswap_v3_router.json"
      enabled: true
    
    - address: "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"  # Uniswap V2 Router
      name: "Uniswap V2 Router"
      abi_path: "abi/uniswap_v2_router.json"
      enabled: true
    
    - address: "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9"  # Aave Lending Pool
      name: "Aave Lending Pool V2"
      abi_path: "abi/aave_pool.json"
      enabled: true
  
  # EOA addresses (opsional)
  eoa:
    - address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
      label: "Suspicious Wallet"
      enabled: true
  
  # Threshold global (opsional, bisa di-override per rule)
  global_thresholds:
    large_transfer: "1000000000000000000"  # 1 ETH
    suspicious_amount: "10000000000000000000"  # 10 ETH
```

**Penting:**
- Registry adalah **metadata untuk runtime**, bukan untuk SDK langsung
- Runtime menggunakan registry untuk:
  - Menentukan address mana yang perlu di-fetch dari blockchain
  - Filter transaksi sebelum memanggil engine
  - Membuat Context dengan data yang relevan

### Step 3: Tulis Custom Rules

```go
// rules/large_swap.go
package rules

import (
    "math/big"
    "github.com/asentric/asentric-sdk/pkg/asentric"
)

// LargeSwapRule mendeteksi swap dengan nilai besar
type LargeSwapRule struct {
    threshold *big.Int
}

// NewLargeSwapRule membuat rule baru dengan threshold
func NewLargeSwapRule(threshold *big.Int) *LargeSwapRule {
    return &LargeSwapRule{threshold: threshold}
}

func (r *LargeSwapRule) Name() string {
    return "large_swap_detection"
}

func (r *LargeSwapRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    // Logika deteksi murni - tidak ada I/O, tidak ada side effects
    if tx.Value().Cmp(r.threshold) > 0 {
        return &asentric.Alert{
            Severity:    asentric.High,
            Title:       "Large Swap Detected",
            Description: "Transaction value exceeds configured threshold",
            Metadata: map[string]interface{}{
                "value":     tx.Value().String(),
                "threshold": r.threshold.String(),
                "from":      tx.From().String(),
                "to":        tx.To().String(),
            },
        }, nil
    }
    
    // Tidak ada alert jika tidak terdeteksi
    return nil, nil
}
```

**Karakteristik Rules:**
- ✅ Pure functions — tidak ada I/O, tidak ada side effects
- ✅ Deterministik — input yang sama selalu menghasilkan output yang sama
- ✅ Stateless — tidak bergantung pada state eksternal
- ✅ Context-based — semua data dari Context

### Step 4: Test Locally dengan Replay

```bash
# Buat fixture dari transaksi yang ingin di-test
cat > fixtures/large_swap_tx.json <<EOF
{
  "chain_id": 1,
  "block_number": 18500000,
  "tx_hash": "0xabc123...",
  "from": "0x1234...",
  "to": "0x5678...",
  "value": "2000000000000000000",
  "data": "0x...",
  "logs": []
}
EOF

# Replay offline untuk test rules
asentric replay --fixture fixtures/large_swap_tx.json

# Output:
# ✓ Rule: large_swap_detection
#   Alert: High - Large Swap Detected
#   Metadata: {value: 2000000000000000000, threshold: 1000000000000000000}
```

**Keuntungan Replay:**
- ✅ Test rules tanpa koneksi blockchain
- ✅ Deterministik — hasil selalu sama
- ✅ Cepat — tidak perlu menunggu block baru
- ✅ Safe — tidak mempengaruhi production

### Step 5: Deploy ke Runtime

Runtime (bot/worker) membaca config dan registry, kemudian menjalankan engine:

```go
// cmd/watcher/main.go
package main

import (
    "github.com/asentric/asentric-sdk/pkg/asentric"
    "gopkg.in/yaml.v3"
)

func main() {
    // 1. Load config (runtime responsibility)
    config := loadConfig("config/asentric.yaml")
    registry := loadRegistry("config/registry.yaml")
    
    // 2. Setup RPC connection (runtime responsibility)
    rpcClient := connectRPC(os.Getenv("RPC_ENDPOINT"))
    
    // 3. Create engine dengan config
    engine := setupEngine(config)
    
    // 4. Start monitoring (runtime responsibility)
    watcher := NewWatcher(engine, rpcClient, registry)
    watcher.Start()
}

func setupEngine(config *Config) *asentric.Engine {
    engine := asentric.NewEngine()
    
    // Register rules dengan options dari config
    for _, ruleName := range config.Engine.EnabledRules {
        options := config.Engine.RuleOptions[ruleName]
        rule := createRule(ruleName, options)
        engine.RegisterRule(rule)
    }
    
    return engine
}
```

---

## Arsitektur: Developer vs Runtime

```
┌─────────────────────────────────────────────────────────────┐
│                    Developer Workspace                      │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ config.yaml  │  │registry.yaml │  │  rules/*.go  │    │
│  │ (engine cfg) │  │ (target list)│  │ (Go code)    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│         │                  │                  │            │
│         └──────────────────┴──────────────────┘            │
│                            │                               │
│                            ▼                               │
│                   ┌─────────────────┐                     │
│                   │  CLI / Runtime  │                     │
│                   │  (Parse & Setup)│                     │
│                   └─────────────────┘                     │
└────────────────────────────┼───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Runtime Environment                     │
│                                                             │
│  ┌──────────────┐         ┌──────────────┐                │
│  │  RPC Client  │────────▶│   Context    │                │
│  │  (Fetch Data)│         │  (Immutable) │                │
│  └──────────────┘         └──────────────┘                │
│         │                         │                        │
│         │                         ▼                        │
│         │              ┌─────────────────┐                │
│         │              │  Asentric SDK   │                │
│         │              │     Engine      │                │
│         │              │  (Pure & Det.)  │                │
│         │              └─────────────────┘                │
│         │                         │                        │
│         │                         ▼                        │
│         │              ┌─────────────────┐                │
│         └─────────────▶│     Alerts      │                │
│                        │   (Semantic)    │                │
│                        └─────────────────┘                │
│                                 │                          │
│                                 ▼                          │
│              ┌──────────────────────────────────┐         │
│              │  Runtime Alert Handling          │         │
│              │  - Enrich with chain context     │         │
│              │  - Deliver (Slack, Email, etc.)  │         │
│              │  - Persist (Database)            │         │
│              └──────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## Boundary yang Jelas

### Developer Hanya Peduli:

✅ **YAML Config** — Engine behavior dan target list  
✅ **Go Rules** — Logika deteksi murni  
✅ **Local Testing** — Replay dengan fixtures  
✅ **Deploy** — Push ke runtime (bot/worker)

### Developer Tidak Perlu Peduli:

❌ RPC connection — ditangani runtime  
❌ Data fetching — ditangani runtime  
❌ Alert delivery — ditangani runtime  
❌ Persistence — ditangani runtime  
❌ Deployment infrastructure — ditangani runtime

### Engine (SDK) Adalah:

✅ **Pure execution box** — Context → Alerts  
✅ **Deterministic** — input sama = output sama  
✅ **Testable** — bisa di-test tanpa infrastruktur  
✅ **Replayable** — bisa di-replay offline

### Runtime Adalah:

✅ **Infrastructure layer** — RPC, database, delivery  
✅ **Orchestration** — fetch data, create context, call engine  
✅ **Enrichment** — tambahkan chain context ke alerts  
✅ **Delivery** — kirim dan simpan alerts

---

## Workflow: Write → Test → Deploy

### 1. Write (Development)

```bash
# Developer menulis rules
vim rules/large_swap.go

# Developer update config jika perlu
vim config/asentric.yaml
vim config/registry.yaml
```

### 2. Test (Local)

```bash
# Test dengan replay offline
asentric replay --fixture fixtures/test_tx.json

# Atau test dengan unit test
go test ./rules/...

# Pastikan rules deterministic
asentric replay --fixture fixtures/test_tx.json
# Run lagi, hasil harus sama
```

### 3. Deploy (Production)

```bash
# Commit dan push
git add .
git commit -m "Add large swap detection rule"
git push

# Runtime (bot) otomatis:
# 1. Pull latest code
# 2. Parse config & registry
# 3. Setup engine dengan rules baru
# 4. Start monitoring
```

---

## Contoh Lengkap: End-to-End

### Setup Project

```bash
asentric init uniswap-monitor
cd uniswap-monitor
```

### Config Files

**config/asentric.yaml:**
```yaml
engine:
  enabled_rules:
    - large_swap_detection
    - suspicious_transfer
  
  rule_options:
    large_swap_detection:
      threshold: "1000000000000000000"  # 1 ETH
```

**config/registry.yaml:**
```yaml
targets:
  contracts:
    - address: "0xE592427A0AEce92De3Edee1F18E0157C05861564"
      name: "Uniswap V3 Router"
      abi_path: "abi/uniswap_v3.json"
```

### Custom Rule

```go
// rules/large_swap.go
package rules

import (
    "math/big"
    "github.com/asentric/asentric-sdk/pkg/asentric"
)

type LargeSwapRule struct {
    threshold *big.Int
}

func NewLargeSwapRule(threshold *big.Int) *LargeSwapRule {
    return &LargeSwapRule{threshold: threshold}
}

func (r *LargeSwapRule) Name() string {
    return "large_swap_detection"
}

func (r *LargeSwapRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    if tx.Value().Cmp(r.threshold) > 0 {
        return &asentric.Alert{
            Severity:    asentric.High,
            Title:       "Large Swap Detected",
            Description: "Transaction value exceeds threshold",
            Metadata: map[string]interface{}{
                "value": tx.Value().String(),
            },
        }, nil
    }
    return nil, nil
}
```

### Test Locally

```bash
# Buat fixture
echo '{
  "tx_hash": "0x123...",
  "value": "2000000000000000000",
  ...
}' > fixtures/test.json

# Replay
asentric replay --fixture fixtures/test.json
# ✓ Alert: High - Large Swap Detected
```

### Runtime (Bot) Setup

```go
// cmd/watcher/main.go
func main() {
    // Load config
    config := loadConfig("config/asentric.yaml")
    registry := loadRegistry("config/registry.yaml")
    
    // Setup engine
    engine := asentric.NewEngine()
    threshold := parseBigInt(config.Engine.RuleOptions["large_swap_detection"]["threshold"])
    engine.RegisterRule(rules.NewLargeSwapRule(threshold))
    
    // Connect RPC (runtime responsibility)
    rpc := connectRPC(os.Getenv("RPC_ENDPOINT"))
    
    // Start monitoring
    watcher := NewWatcher(engine, rpc, registry)
    watcher.Start()
}
```

---

## Perbandingan dengan Ponder.sh

| Aspek | Ponder.sh | Asentric SDK |
|-------|-----------|--------------|
| **Config** | YAML (schema, chains) | YAML (engine, registry) |
| **Logic** | TypeScript functions | Go functions (rules) |
| **Testing** | Local replay | Offline replay |
| **Deploy** | Push to Ponder | Push to runtime (bot) |
| **Infrastructure** | Managed by Ponder | Managed by runtime |
| **Deterministic** | ✅ | ✅ |
| **Pure Logic** | ✅ | ✅ |

**Kesamaan:**
- Developer hanya fokus pada logika
- Setup minimal dengan YAML
- Testing lokal tanpa infrastruktur
- Deploy sederhana

**Perbedaan:**
- Ponder.sh: Managed infrastructure
- Asentric SDK: Runtime (bot) mengelola infrastruktur sendiri

---

## Kesimpulan

Developer experience Asentric SDK dirancang untuk:

1. **Setup Minimal** — Hanya YAML + Go rules
2. **Fokus pada Logika** — Tidak perlu khawatir infrastruktur
3. **Testing Mudah** — Replay offline, deterministic
4. **Deploy Sederhana** — Push ke runtime, runtime yang handle sisanya

Dengan boundary yang jelas:
- **YAML** = setup & target list
- **Rules** = logika murni (Go)
- **Engine** = kotak hitam deterministic
- **Runtime** = infrastruktur & orchestration

Developer bisa langsung mulai monitoring dengan konfigurasi minimal dan custom rules, tanpa harus memikirkan RPC, database, atau deployment infrastructure.

