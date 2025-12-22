# Memahami pkg/asentric dan pkg/domain

> **Dokumentasi untuk tim memahami code yang ada sebelum implementasi**

---

## Daftar Isi

1. [Gambaran Besar](#1-gambaran-besar)
2. [pkg/asentric - Public API](#2-pkgasentric---public-api)
3. [pkg/domain - Data Structures](#3-pkgdomain---data-structures)
4. [Hubungan Antar Komponen](#4-hubungan-antar-komponen)
5. [Contoh Penggunaan](#5-contoh-penggunaan)

---

## 1. Gambaran Besar

### Flow Data Sederhana

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│   [1]           [2]              [3]              [4]            │
│                                                                  │
│   Event    →   Context     →   Engine.Evaluate() →   Alert      │
│                                                                  │
│   (input)      (data untuk      (jalankan         (output)       │
│                 rule baca)       semua rules)                    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### Apa itu Asentric?

Asentric adalah **engine untuk evaluasi security rules** pada blockchain events.

**Analogi sederhana:**
- **Event** = "Ada transaksi baru di blockchain"
- **Context** = "Detail transaksi: siapa kirim, berapa nilai, ke mana"
- **Rule** = "Jika nilai > 100 ETH, beri tahu saya"
- **Alert** = "Peringatan: Transaksi besar terdeteksi!"

---

## 2. pkg/asentric - Public API

### 2.1 Engine (⭐ INTI)

**File:** `engine.go`

**Apa ini?** Engine adalah "otak" yang menjalankan semua rules.

```go
type Engine struct {
    rules []Rule  // daftar rules yang terdaftar
}
```

**Methods:**

| Method | Fungsi |
|--------|--------|
| `NewEngine()` | Buat engine baru |
| `RegisterRule(rule)` | Tambah rule ke engine |
| `Evaluate(ctx)` | Jalankan semua rules, return alerts |

**Contoh:**

```go
// 1. Buat engine
engine := asentric.NewEngine()

// 2. Daftarkan rule
engine.RegisterRule(&MyRule{})

// 3. Evaluasi (jalankan rules)
alerts, err := engine.Evaluate(ctx)
```

**Sifat Engine:**
- ✅ **Deterministic** - input sama = output sama
- ✅ **Single-threaded** - tidak concurrent
- ✅ **No I/O** - tidak baca file, network, dll
- ✅ **Sequential** - rules dijalankan urut sesuai pendaftaran

---

### 2.2 Rule (⭐ INTI)

**File:** `rule.go`

**Apa ini?** Rule adalah logic deteksi yang kamu tulis sendiri.

```go
type Rule interface {
    Name() string                            // nama unik rule
    Evaluate(ctx Context) (*Alert, error)    // logic deteksi
}
```

**Contoh implementasi:**

```go
type LargeTransferRule struct {
    Threshold *big.Int
}

func (r *LargeTransferRule) Name() string {
    return "large_transfer"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // Baca data dari context
    tx := ctx.Tx()
    
    // Logic deteksi
    if tx.Value().Cmp(r.Threshold) > 0 {
        // Terdeteksi! Return alert
        return &asentric.Alert{
            Rule:        r.Name(),
            Severity:    asentric.SeverityHigh,
            Title:       "Large Transfer Detected",
            Description: "Transaction value exceeds threshold",
        }, nil
    }
    
    // Tidak terdeteksi = return nil, nil
    return nil, nil
}
```

**Aturan Rule:**
- ✅ **Pure function** - tidak ada side effect
- ✅ **Hanya baca Context** - tidak modifikasi
- ✅ **Tidak I/O** - tidak network, file, database
- ✅ **Return nil, nil** jika tidak ada deteksi
- ✅ **Return alert, nil** jika ada deteksi
- ✅ **Return nil, error** jika ada error

---

### 2.3 Context (⭐ INTI)

**File:** `context.go`

**Apa ini?** Context adalah "data yang bisa dibaca oleh Rule".

```go
type Context interface {
    ChainID() domain.ChainID      // ID chain (misal: 1 = Ethereum)
    Tx() domain.Transaction       // data transaksi
    Block() domain.Block          // data block
    Logs() []domain.Log           // event logs
    ABI() domain.ABIRegistry      // registry untuk decode
}
```

**Diagram Context:**

```
                    Context
                       │
       ┌───────────────┼───────────────┐
       │               │               │
       ▼               ▼               ▼
   ChainID()        Tx()           Block()
       │               │               │
       ▼               ▼               ▼
    uint64       Transaction        Block
                       │
                       ▼
                   Value() → *big.Int
```

**Contoh penggunaan dalam Rule:**

```go
func (r *MyRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // Ambil chain ID
    chainID := ctx.ChainID()  // misal: 1 (Ethereum)
    
    // Ambil data transaksi
    tx := ctx.Tx()
    from := tx.From           // siapa yang kirim
    to := tx.To               // ke mana
    value := tx.Value()       // berapa nilainya (*big.Int)
    
    // Ambil data block
    block := ctx.Block()
    blockNumber := block.Number
    
    // Ambil event logs
    logs := ctx.Logs()
    for _, log := range logs {
        eventName := log.Event.Name  // misal: "Transfer"
    }
    
    // ... logic deteksi ...
}
```

---

### 2.4 Alert

**File:** `alert.go`

**Apa ini?** Alert adalah output dari Rule ketika mendeteksi sesuatu.

```go
type Alert struct {
    Rule        string            // nama rule yang generate
    Severity    Severity          // tingkat keparahan
    Title       string            // judul singkat
    Description string            // deskripsi detail
    Ref         *ExecutionRef     // referensi transaksi (diisi engine)
    Metadata    map[string]any    // data tambahan (bebas)
}
```

**Severity levels:**

```go
SeverityCritical = "CRITICAL"   // Sangat serius, perlu tindakan segera
SeverityHigh     = "HIGH"       // Serius
SeverityMedium   = "MEDIUM"     // Perlu perhatian
SeverityLow      = "LOW"        // Informasi penting
SeverityInfo     = "INFO"       // Informasi biasa
```

**Contoh membuat Alert:**

```go
return &asentric.Alert{
    Rule:        r.Name(),
    Severity:    asentric.SeverityHigh,
    Title:       "Large Transfer Detected",
    Description: "Transaction value exceeds 100 ETH",
    Metadata: map[string]any{
        "value":     tx.Value().String(),
        "from":      tx.From.String(),
        "to":        tx.To.String(),
    },
}, nil
```

---

### 2.5 Event

**File:** `event.go`

**Apa ini?** Event adalah data mentah dari blockchain yang masuk ke sistem.

```go
type Event struct {
    ChainID     uint64    // ID chain
    BlockNumber uint64    // nomor block
    TxHash      string    // hash transaksi
    Payload     any       // data tambahan
}
```

**Hubungan Event dan Context:**

```
Event (raw dari blockchain)
    │
    ▼
Context Builder (di internal)
    │
    ▼
Context (siap untuk Rule)
```

**Kamu tidak perlu membuat Event langsung** - EventSource yang akan membuatnya.

---

### 2.6 Interfaces untuk Implementasi

**Ini adalah interfaces yang PERLU kamu implement:**

#### EventSource

**File:** `event_source.go`

```go
type EventSource interface {
    Start(ctx context.Context) (<-chan Event, error)
}
```

**Fungsi:** Menghasilkan Event dari blockchain (via WebSocket).

**Contoh skeleton:**

```go
type MyEventSource struct {
    rpcURL string
}

func (s *MyEventSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    ch := make(chan asentric.Event)
    
    go func() {
        defer close(ch)
        // Connect ke WebSocket
        // Subscribe logs
        // Parse dan kirim ke channel
    }()
    
    return ch, nil
}
```

---

#### AlertSink

**File:** `alert_sink.go`

```go
type AlertSink interface {
    Emit(ctx context.Context, alert *Alert) error
}
```

**Fungsi:** Mengirim Alert ke tujuan (webhook, telegram, dll).

**Contoh skeleton:**

```go
type MyWebhookSink struct {
    url string
}

func (s *MyWebhookSink) Emit(ctx context.Context, alert *asentric.Alert) error {
    // Serialize alert ke JSON
    // HTTP POST ke webhook URL
    return nil
}
```

---

#### Dispatcher

**File:** `dispatcher.go`

```go
type Dispatcher interface {
    Dispatch(ctx context.Context, event Event) error
}
```

**Fungsi:** Menghubungkan Event → Context → Engine → AlertSink.

**Contoh skeleton:**

```go
type MyDispatcher struct {
    engine *asentric.Engine
    sink   asentric.AlertSink
}

func (d *MyDispatcher) Dispatch(ctx context.Context, event asentric.Event) error {
    // 1. Convert Event ke Context
    execCtx := buildContext(event)
    
    // 2. Jalankan engine
    alerts, err := d.engine.Evaluate(execCtx)
    if err != nil {
        return err
    }
    
    // 3. Kirim alerts
    for _, alert := range alerts {
        d.sink.Emit(ctx, alert)
    }
    
    return nil
}
```

---

#### Logger

**File:** `logger.go`

```go
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
    Debug(msg string, args ...any)
}
```

**Fungsi:** Logging untuk debugging.

---

### 2.7 Errors

**File:** `errors.go`

```go
var (
    ErrInvalidContext = errors.New("invalid context")
    ErrInvalidRule    = errors.New("invalid rule")
    ErrDuplicateRule  = errors.New("duplicate rule id")
    ErrInvalidEvent   = errors.New("invalid event")
    ErrInvalidConfig  = errors.New("invalid configuration")
    ErrNoDispatcher   = errors.New("dispatcher is not set")
    ErrAlreadyRunning = errors.New("runtime is already running")
    ErrRulePanic      = errors.New("rule panic")
)
```

---

### 2.8 Config

**File:** `config.go`

**Apa ini?** Struct untuk konfigurasi runtime.

```go
type RuntimeConfig struct {
    Chain   ChainConfig    // konfigurasi chain
    Redis   RedisConfig    // konfigurasi Redis
    Webhook WebhookConfig  // konfigurasi webhook
    Engine  EngineConfig   // konfigurasi engine
}
```

**Contoh YAML:**

```yaml
# config/asentric.yaml
chain:
  rpc_ws: "wss://rpc.mantle.xyz/ws"
  name: "Mantle"
  chain_id: 5000

redis:
  addr: "localhost:6379"

webhook:
  url: "https://your-webhook.com/alerts"
  timeout: 10s
  retry_count: 3
```

---

## 3. pkg/domain - Data Structures

### 3.1 Primitive Types

#### Address

```go
type Address string

// Methods:
a.String()  // "0x1234..."
a.Hex()     // "0x1234..."
a.IsZero()  // true jika kosong atau 0x000...
```

#### Hash

```go
type Hash string

// Methods:
h.String()  // "0xabc123..."
h.Hex()     // "0xabc123..."
h.IsZero()  // true jika kosong atau 0x000...
```

#### ChainID

```go
type ChainID uint64

// Contoh:
// 1 = Ethereum Mainnet
// 5000 = Mantle
// 42161 = Arbitrum
```

---

### 3.2 Transaction (⭐ PALING SERING DIPAKAI)

```go
type Transaction struct {
    // Identity
    Hash  Hash       // hash transaksi
    Index uint64     // posisi di block
    
    // Parties
    From Address     // pengirim
    To   Address     // penerima (kosong jika create contract)
    
    // Execution
    Nonce    uint64  // nonce pengirim
    GasLimit uint64  // gas limit
    GasUsed  uint64  // gas yang dipakai
    Status   bool    // true = sukses, false = reverted
    
    // Value (internal)
    RawValue NativeValue
    
    // Gas pricing
    GasPrice     string  // legacy
    MaxFeePerGas string  // EIP-1559
    MaxPriority  string  // EIP-1559
    
    // Type
    Type   TxType       // Legacy, AccessList, DynamicFee, Blob
    Action TxActionType // CALL, CREATE, DELEGATECALL
    
    // Decoded call (nil jika bukan contract call)
    Call *ContractCall
    
    // Block context
    BlockNumber uint64
    BlockHash   Hash
    Timestamp   uint64
}
```

**Method paling penting:**

```go
// Value() returns *big.Int - GUNAKAN INI untuk compare nilai
value := tx.Value()

// Contoh compare dengan threshold
threshold := big.NewInt(1e18)  // 1 ETH dalam wei
if tx.Value().Cmp(threshold) > 0 {
    // nilai lebih besar dari 1 ETH
}
```

**Contoh penggunaan:**

```go
func (r *MyRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    // Cek siapa pengirim
    if tx.From.String() == "0xHacker..." {
        return &asentric.Alert{...}, nil
    }
    
    // Cek nilai transaksi
    oneEth := big.NewInt(1e18)
    if tx.Value().Cmp(oneEth) > 0 {
        return &asentric.Alert{...}, nil
    }
    
    // Cek apakah transaksi sukses
    if !tx.Status {
        // transaksi reverted
    }
    
    // Cek apakah contract call
    if tx.Call != nil {
        methodName := tx.Call.Method  // misal: "transfer"
        args := tx.Call.Args          // map[string]any
    }
    
    return nil, nil
}
```

---

### 3.3 Block

```go
type Block struct {
    Number    uint64   // nomor block
    Hash      Hash     // hash block
    Parent    Hash     // hash parent block
    Timestamp uint64   // unix timestamp
    Miner     Address  // who mined/proposed
    GasLimit  uint64
    GasUsed   uint64
    BaseFee   string   // EIP-1559 base fee
    TxCount   int      // jumlah transaksi
}
```

**Contoh penggunaan:**

```go
func (r *MyRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    block := ctx.Block()
    
    // Cek block number
    if block.Number < 1000000 {
        // block lama
    }
    
    // Cek gas usage
    gasPercent := float64(block.GasUsed) / float64(block.GasLimit)
    if gasPercent > 0.9 {
        // block hampir penuh
    }
    
    return nil, nil
}
```

---

### 3.4 Log dan Event

```go
type Log struct {
    Address     Address  // kontrak yang emit
    LogIndex    uint64   // posisi log
    TxHash      Hash
    TxIndex     uint64
    Event       Event    // decoded event
    BlockNumber uint64
    BlockHash   Hash
}

type Event struct {
    Name   string         // nama event, misal: "Transfer"
    Fields map[string]any // decoded fields
}
```

**Contoh penggunaan:**

```go
func (r *MyRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        // Cek event name
        if log.Event.Name == "Transfer" {
            from := log.Event.Fields["from"]
            to := log.Event.Fields["to"]
            value := log.Event.Fields["value"]
            
            // ... logic ...
        }
        
        // Cek kontrak mana yang emit
        if log.Address.String() == "0xTokenContract..." {
            // ...
        }
    }
    
    return nil, nil
}
```

---

### 3.5 Value Types

```go
type NativeValue struct {
    Wei string  // decimal string
}

func (v NativeValue) IsZero() bool

type TokenAmount struct {
    Token  Token
    Amount string
}

type Token struct {
    Address  Address
    Symbol   string
    Decimals uint8
}
```

---

### 3.6 ABI Registry

```go
type ABIRegistry interface {
    GetMethod(address Address, selector string) (Method, bool)
    GetEvent(address Address, topic Hash) (Event, bool)
}

type Method struct {
    Name string
    Args []ABIArg
}

type ABIArg struct {
    Name string
    Type string
}
```

**Ini perlu diimplementasi** untuk decode transaction input dan event logs.

---

## 4. Hubungan Antar Komponen

### Diagram Lengkap

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                        FLOW DATA ASENTRIC                                   │
│                                                                             │
│  ┌─────────────┐                                                            │
│  │  Blockchain │                                                            │
│  │  (via WS)   │                                                            │
│  └──────┬──────┘                                                            │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────┐                                                        │
│  │  EventSource    │  ← Interface yang perlu implement                      │
│  │  (WebSocket)    │                                                        │
│  └────────┬────────┘                                                        │
│           │                                                                 │
│           │  Event { ChainID, BlockNumber, TxHash, Payload }                │
│           ▼                                                                 │
│  ┌─────────────────┐                                                        │
│  │   Dispatcher    │  ← Interface yang perlu implement                      │
│  └────────┬────────┘                                                        │
│           │                                                                 │
│           │  1. Build Context dari Event                                    │
│           │  2. Call engine.Evaluate(ctx)                                   │
│           │  3. Kirim alerts ke sink                                        │
│           ▼                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                               │
│  │    Context      │ ──▶ │     Engine      │                               │
│  │  (immutable)    │     │   (evaluator)   │                               │
│  │                 │     │                 │                               │
│  │  • ChainID()    │     │  ┌───────────┐  │                               │
│  │  • Tx()         │     │  │  Rule 1   │  │                               │
│  │  • Block()      │     │  │  Rule 2   │  │                               │
│  │  • Logs()       │     │  │  Rule 3   │  │                               │
│  │  • ABI()        │     │  └───────────┘  │                               │
│  └─────────────────┘     └────────┬────────┘                               │
│                                   │                                         │
│                                   │  []*Alert                               │
│                                   ▼                                         │
│                          ┌─────────────────┐                               │
│                          │   AlertSink     │  ← Interface yang perlu impl  │
│                          │   (Webhook)     │                               │
│                          └─────────────────┘                               │
│                                   │                                         │
│                                   ▼                                         │
│                          ┌─────────────────┐                               │
│                          │  Your Webhook   │                               │
│                          │  (Slack, etc)   │                               │
│                          └─────────────────┘                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Dependency Graph

```
pkg/asentric/
    │
    ├── engine.go      uses→  rule.go, context.go, alert.go, errors.go
    │
    ├── rule.go        uses→  context.go, alert.go
    │
    ├── context.go     uses→  pkg/domain/*
    │
    ├── dispatcher.go  uses→  event.go, context.Context
    │
    ├── event_source.go uses→ event.go, context.Context
    │
    └── alert_sink.go  uses→  alert.go, context.Context


pkg/domain/
    │
    ├── transaction.go uses→  address.go, hash.go, value.go
    │
    ├── block.go       uses→  address.go, hash.go
    │
    ├── log.go         uses→  address.go, hash.go, event.go
    │
    └── abi.go         uses→  address.go, hash.go, event.go
```

---

## 5. Contoh Penggunaan

### Contoh Rule Lengkap

```go
package rules

import (
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
)

// LargeTransferRule deteksi transaksi dengan nilai besar
type LargeTransferRule struct {
    Threshold *big.Int  // threshold dalam wei
}

// Name returns nama unik rule
func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

// Evaluate menjalankan logic deteksi
func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    // 1. Ambil data transaksi
    tx := ctx.Tx()
    
    // 2. Skip jika transaksi gagal
    if !tx.Status {
        return nil, nil
    }
    
    // 3. Compare nilai dengan threshold
    if tx.Value().Cmp(r.Threshold) > 0 {
        // 4. Terdeteksi! Buat alert
        return &asentric.Alert{
            Rule:        r.Name(),
            Severity:    asentric.SeverityHigh,
            Title:       "Large Transfer Detected",
            Description: "Transaction value exceeds configured threshold",
            Metadata: map[string]any{
                "value_wei":  tx.Value().String(),
                "threshold":  r.Threshold.String(),
                "from":       tx.From.String(),
                "to":         tx.To.String(),
                "tx_hash":    tx.Hash.String(),
                "block":      tx.BlockNumber,
            },
        }, nil
    }
    
    // 5. Tidak terdeteksi
    return nil, nil
}
```

### Contoh Main.go (Pseudocode)

```go
package main

import (
    "context"
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
    "myproject/rules"
)

func main() {
    // 1. Buat engine
    engine := asentric.NewEngine()
    
    // 2. Register rules
    engine.RegisterRule(&rules.LargeTransferRule{
        Threshold: big.NewInt(1e18),  // 1 ETH
    })
    
    // 3. Buat components (TIM IMPLEMENT INI)
    eventSource := NewMyEventSource("wss://...")
    alertSink := NewMyWebhookSink("https://...")
    dispatcher := NewMyDispatcher(engine, alertSink)
    
    // 4. Start event source
    events, _ := eventSource.Start(context.Background())
    
    // 5. Event loop
    for event := range events {
        dispatcher.Dispatch(context.Background(), event)
    }
}
```

---

## Checklist Pemahaman

Sebelum mulai implement, pastikan kamu paham:

- [ ] **Engine.Evaluate()** - Menerima Context, return Alerts
- [ ] **Rule interface** - Name() dan Evaluate()
- [ ] **Context interface** - ChainID(), Tx(), Block(), Logs(), ABI()
- [ ] **Transaction.Value()** - Return `*big.Int` untuk compare
- [ ] **Alert struct** - Rule, Severity, Title, Description, Metadata
- [ ] **Event struct** - ChainID, BlockNumber, TxHash, Payload
- [ ] **EventSource interface** - Start() return channel Event
- [ ] **AlertSink interface** - Emit() kirim Alert
- [ ] **Dispatcher interface** - Dispatch() process Event

---

## Tips

1. **Mulai dari Rule** - Tulis rule sederhana dulu, test dengan mock context
2. **Pakai tx.Value()** - Jangan akses RawValue langsung
3. **Return nil, nil** - Jika tidak ada deteksi
4. **Metadata JSON-safe** - Hanya string, number, bool, array, object
5. **Jangan I/O di Rule** - Tidak network, file, database

---

**Selamat belajar!** 🚀

