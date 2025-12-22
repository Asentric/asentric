# Asentric Implementation Guide

> **Panduan untuk memahami dan mengimplementasi Asentric step-by-step**
>
> Dokumen ini menjelaskan **alur data, hubungan komponen, dan apa yang perlu diimplementasi**.
>
> **Prerequisite:** Baca [UNDERSTAND-PKG.md](UNDERSTAND-PKG.md) terlebih dahulu untuk memahami pkg/asentric dan pkg/domain.

---

## 1. Struktur Saat Ini

```
asentric-sdk/
├── pkg/asentric/          # ✅ PUBLIC API (Ready)
│   ├── engine.go          # Engine struct + Evaluate()
│   ├── rule.go            # Rule interface
│   ├── context.go         # Context interface
│   ├── alert.go           # Alert struct
│   ├── event.go           # Event struct
│   ├── event_source.go    # EventSource interface
│   ├── alert_sink.go      # AlertSink interface
│   ├── dispatcher.go      # Dispatcher interface
│   ├── config.go          # Config structs
│   ├── logger.go          # Logger interface
│   ├── runtime.go         # Runtime struct
│   └── errors.go          # Error types
│
├── pkg/domain/            # ✅ DOMAIN TYPES (Ready)
│   ├── address.go         # Address type
│   ├── hash.go            # Hash type
│   ├── chain.go           # ChainID type
│   ├── transaction.go     # Transaction + Value()
│   ├── block.go           # Block struct
│   ├── log.go             # Log struct
│   ├── event.go           # Event (decoded)
│   ├── value.go           # NativeValue, TokenAmount
│   ├── token.go           # Token metadata
│   └── abi.go             # ABIRegistry interface
│
├── internal/              # 🔧 CONTOH + PERLU IMPLEMENTASI
│   ├── context/           # ✅ Contoh implementation
│   │   └── context.go     # EventContext
│   │
│   ├── dispatcher/        # ✅ Contoh implementation  
│   │   └── dispatcher.go  # EngineDispatcher (partial)
│   │
│   ├── runtime/           # ✅ Contoh implementation
│   │   ├── runtime.go     # Runtime loop
│   │   └── shutdown.go    # Signal handling
│   │
│   ├── source/            # ❌ KOSONG - tim implement
│   ├── sink/              # ❌ KOSONG - tim implement
│   ├── abi/               # ❌ KOSONG - tim implement
│   ├── chain/             # ❌ KOSONG - tim implement
│   ├── adapter/           # ❌ KOSONG - tim implement
│   └── queue/             # ❌ KOSONG - tim implement (optional)
```

---

## 2. Alur Data

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   [1]              [2]              [3]              [4]        │
│                                                                 │
│  Blockchain   →   Context      →   Engine       →   Alert      │
│  (WebSocket)     (snapshot)       .Evaluate()      (output)     │
│                                                                 │
│       │               │                │               │        │
│       ▼               ▼                ▼               ▼        │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐    ┌─────────┐   │
│  │ Event   │    │ Context  │    │  Rules   │    │  Alert  │   │
│  │ Source  │ →  │ Builder  │ →  │  Loop    │ →  │  Sink   │   │
│  └─────────┘    └──────────┘    └──────────┘    └─────────┘   │
│                                                                 │
│   TIM BUAT       TIM BUAT          READY         TIM BUAT      │
└─────────────────────────────────────────────────────────────────┘
```

### Flow Sederhana:

```
WebSocket → Event → Context → Engine.Evaluate() → Alert → Webhook
```

---

## 3. Yang Sudah Ready (Tinggal Pakai)

### 3.1 Engine (`pkg/asentric/engine.go`)

```go
engine := asentric.NewEngine()
engine.RegisterRule(&MyRule{})
alerts, err := engine.Evaluate(ctx)
```

### 3.2 Context Interface (`pkg/asentric/context.go`)

```go
type Context interface {
    ChainID() domain.ChainID
    Tx() domain.Transaction
    Block() domain.Block
    Logs() []domain.Log
    ABI() domain.ABIRegistry
}
```

### 3.3 Contoh Context Implementation (`internal/context/context.go`)

```go
ctx := context.NewEventContext(event).
    WithTransaction(tx).
    WithBlock(block).
    WithLogs(logs)
```

---

## 4. Yang Perlu Tim Implement

### 4.1 EventSource (Priority: P0)

**File:** `internal/source/websocket.go`

**Interface:**
```go
type EventSource interface {
    Start(ctx context.Context) (<-chan Event, error)
}
```

**Yang harus dilakukan:**
1. Connect ke WebSocket RPC
2. Subscribe ke logs: `eth_subscribe("logs", {...})`
3. Parse response ke `asentric.Event`
4. Kirim ke channel

**Skeleton:**
```go
package source

type WebSocketSource struct {
    rpcURL string
}

func NewWebSocketSource(rpcURL string) *WebSocketSource {
    return &WebSocketSource{rpcURL: rpcURL}
}

func (s *WebSocketSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    ch := make(chan asentric.Event)
    
    go func() {
        defer close(ch)
        // 1. Connect ke WebSocket
        // 2. Subscribe logs
        // 3. Loop baca message
        // 4. Parse dan kirim ke channel
    }()
    
    return ch, nil
}
```

---

### 4.2 AlertSink (Priority: P0)

**File:** `internal/sink/webhook.go`

**Interface:**
```go
type AlertSink interface {
    Emit(ctx context.Context, alert *Alert) error
}
```

**Yang harus dilakukan:**
1. Serialize Alert ke JSON
2. HTTP POST ke webhook URL
3. Handle error/retry

**Skeleton:**
```go
package sink

type WebhookSink struct {
    url string
}

func NewWebhookSink(url string) *WebhookSink {
    return &WebhookSink{url: url}
}

func (s *WebhookSink) Emit(ctx context.Context, alert *asentric.Alert) error {
    // 1. json.Marshal(alert)
    // 2. http.Post(s.url, ...)
    // 3. Handle response
    return nil
}
```

---

### 4.3 ABIRegistry (Priority: P1)

**File:** `internal/abi/registry.go`

**Interface:**
```go
type ABIRegistry interface {
    GetMethod(address Address, selector string) (Method, bool)
    GetEvent(address Address, topic Hash) (Event, bool)
}
```

**Yang harus dilakukan:**
1. Load ABI JSON dari file
2. Parse methods dan events
3. Provide lookup by selector/topic

---

### 4.4 Chain Client (Priority: P1)

**File:** `internal/chain/client.go`

**Yang harus dilakukan:**
1. Wrapper untuk go-ethereum ethclient
2. Fetch transaction by hash
3. Fetch block by number

---

## 5. Contoh Rule

```go
package rules

import (
    "math/big"
    "github.com/asentric/asentric/pkg/asentric"
)

type LargeTransferRule struct {
    Threshold *big.Int
}

func (r *LargeTransferRule) Name() string {
    return "large_transfer"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
    if tx.Value().Cmp(r.Threshold) > 0 {
        return &asentric.Alert{
            Rule:        r.Name(),
            Severity:    asentric.SeverityHigh,
            Title:       "Large Transfer Detected",
            Description: "Transaction exceeds threshold",
            Metadata: map[string]any{
                "value": tx.Value().String(),
                "from":  tx.From.String(),
                "to":    tx.To.String(),
            },
        }, nil
    }
    
    return nil, nil
}
```

---

## 6. Contoh Main.go

```go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/asentric/asentric/pkg/asentric"
    internalctx "github.com/asentric/asentric/internal/context"
)

func main() {
    // 1. Buat engine
    engine := asentric.NewEngine()
    
    // 2. Register rules
    engine.RegisterRule(&LargeTransferRule{
        Threshold: big.NewInt(1e18),
    })
    
    // 3. Buat components (TIM IMPLEMENT)
    source := NewWebSocketSource("wss://...")  // implement
    sink := NewWebhookSink("https://...")      // implement
    
    // 4. Start source
    events, _ := source.Start(context.Background())
    
    // 5. Event loop
    for event := range events {
        // Build context
        ctx := internalctx.NewEventContext(event)
        
        // Evaluate
        alerts, err := engine.Evaluate(ctx)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        // Send alerts
        for _, alert := range alerts {
            sink.Emit(context.Background(), alert)
        }
    }
}
```

---

## 7. Urutan Implementasi

### Phase 1: Pahami Code (1-2 hari)

📚 **Baca [UNDERSTAND-PKG.md](UNDERSTAND-PKG.md)**

1. Pahami Engine, Rule, Context, Alert
2. Pahami Transaction.Value()
3. Tulis 1 simple rule

### Phase 2: Implement EventSource (2-3 hari)

1. Buat `internal/source/websocket.go`
2. Test: connect ke testnet, print events

### Phase 3: Implement AlertSink (1 hari)

1. Buat `internal/sink/webhook.go`
2. Test: kirim ke webhook.site

### Phase 4: Integration (1-2 hari)

1. Wire semuanya di main.go
2. Test end-to-end

### Phase 5: ABI Decoding (2-3 hari)

1. Buat `internal/abi/registry.go`
2. Decode event logs

---

## 8. Files yang Perlu Dibuat

| File | Priority | Description |
|------|----------|-------------|
| `internal/source/websocket.go` | P0 | WebSocket subscription |
| `internal/sink/webhook.go` | P0 | Webhook delivery |
| `internal/abi/registry.go` | P1 | ABI loading |
| `internal/chain/client.go` | P1 | Chain client wrapper |
| `internal/queue/redis.go` | P2 | Redis queue (optional) |

---

## 9. Testing

### Unit Test Rule

```go
func TestMyRule(t *testing.T) {
    // Mock context
    ctx := &MockContext{
        tx: domain.Transaction{
            RawValue: domain.NativeValue{Wei: "2000000000000000000"},
        },
    }
    
    rule := &LargeTransferRule{Threshold: big.NewInt(1e18)}
    alert, err := rule.Evaluate(ctx)
    
    assert.NoError(t, err)
    assert.NotNil(t, alert)
}
```

---

## 10. Referensi

- **[UNDERSTAND-PKG.md](UNDERSTAND-PKG.md)** - Dokumentasi pkg/
- **[SPEC.md](SPEC.md)** - MVP specification
- **[architecture.md](architecture.md)** - Core principles

---

## Summary

**Yang sudah ready:**
- ✅ `pkg/asentric/*` - Public API
- ✅ `pkg/domain/*` - Domain types
- ✅ `internal/context/` - Contoh Context implementation
- ✅ `internal/dispatcher/` - Contoh Dispatcher
- ✅ `internal/runtime/` - Contoh Runtime loop

**Yang tim implement:**
- ❌ `internal/source/websocket.go`
- ❌ `internal/sink/webhook.go`
- ❌ `internal/abi/registry.go`
- ❌ `internal/chain/client.go`

**Flow:**
```
WebSocket → Event → Context → Engine.Evaluate() → Alert → Webhook
```

**Build step by step!**
