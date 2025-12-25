# Runtime Reference – Deep Dive Implementation Guide

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH**  
> **📖 Lihat architecture:** [architecture.md](architecture.md) - Core architecture  
> **📋 Lihat internal runtime:** [internal-runtime-deep-dive.md](internal-runtime-deep-dive.md) - Event loop details

**Status:** ✅ **AUTHORITATIVE**  
**Audience:** Core Engineers, Runtime Implementers, Code Reviewers, Lead

Dokumen ini menjelaskan **secara detail** bagaimana mengimplementasikan `cmd/runtime-reference/` sebagai **reference runtime** yang menggabungkan SDK publik dengan infrastruktur internal, mengikuti canonical flow dari SPEC.

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

---

## Table of Contents

1. [Overview & Purpose](#1-overview--purpose)
2. [Struktur Direktori & Peran Setiap Komponen](#2-struktur-direktori--peran-setiap-komponen)
3. [main.go – Entry Point](#3-maingo--entry-point)
4. [runtime.go – Glue Code](#4-runtimego--glue-code)
5. [config/ – Configuration Bridge](#5-config--configuration-bridge)
6. [ingest/ – Event Source Implementation](#6-ingest--event-source-implementation)
7. [pipeline/ – Event Processing Pipeline](#7-pipeline--event-processing-pipeline)
8. [state/ – Runtime State Management](#8-state--runtime-state-management)
9. [alert/ – Alert Delivery](#9-alert--alert-delivery)
10. [Integrasi Global & Boundary Rules](#10-integrasi-global--boundary-rules)
11. [Contoh Implementasi Lengkap](#11-contoh-implementasi-lengkap)

---

## 1. Overview & Purpose

### 1.1 Apa itu Runtime Reference?

`cmd/runtime-reference/` adalah **contoh runtime produksi** yang:

- ✅ Menggunakan **public SDK** (`pkg/asentric`, `pkg/domain`) seperti user eksternal
- ✅ Menggunakan **internal implementation** (`internal/*`) untuk infrastruktur
- ✅ Mengimplementasikan **canonical flow** dari SPEC end-to-end
- ✅ Menjadi **golden reference** untuk tim dan user eksternal

**Penting:** Runtime reference adalah **contoh**, bukan requirement. User bisa membuat runtime sendiri dengan pola yang berbeda, selama mengikuti kontrak public SDK.

---

### 1.2 Posisi di Arsitektur

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/runtime-reference/                     │
│                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │  main.go │→ │runtime.go│→ │ internal │                 │
│  └──────────┘  └──────────┘  │ /runtime │                 │
│       │             │         └──────────┘                 │
│       │             │              │                         │
│       ▼             ▼              ▼                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │ config/  │  │ ingest/  │  │ pipeline/│                 │
│  └──────────┘  └──────────┘  └──────────┘                 │
│                                                               │
│  ┌──────────┐  ┌──────────┐                                │
│  │  state/  │  │  alert/  │                                │
│  └──────────┘  └──────────┘                                │
└─────────────────────────────────────────────────────────────┘
         │              │              │
         ▼              ▼              ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ pkg/asentric │ │ pkg/domain   │ │ internal/*   │
│ (Public SDK) │ │ (Domain)     │ │ (Infra)      │
└──────────────┘ └──────────────┘ └──────────────┘
```

---

### 1.3 Target Struktur (dari project-structure.md)

```
cmd/runtime-reference/
├── main.go                 # Entry point runtime
│
├── runtime.go              # Glue code - menyatukan semua komponen
│
├── config/
│   ├── loader.go           # Load yaml config (di atas asentric.LoadConfig)
│   └── schema.go           # Schema tambahan jika perlu
│
├── ingest/
│   ├── evm_logs.go         # Subscribe logs via WebSocket
│   └── blocks.go           # (optional) block stream
│
├── pipeline/
│   ├── dispatcher.go       # Fan-out events (jika perlu)
│   └── worker.go           # Engine workers (jika perlu worker pool)
│
├── state/                  # RUNTIME STATE (Redis here)
│   ├── store.go            # Interface untuk state store
│   ├── redis.go            # Redis implementation
│   └── memory.go           # In-memory implementation (dev/test)
│
└── alert/
    ├── webhook.go          # Webhook AlertSink implementation
    ├── telegram.go         # (optional) Telegram sink
    └── dispatcher.go        # (optional) Fan-out ke multiple sinks
```

---

## 2. Struktur Direktori & Peran Setiap Komponen

### 2.1 Hierarki Tanggung Jawab

| Komponen | Tanggung Jawab | Menggunakan |
|----------|----------------|-------------|
| **main.go** | Entry point, load config, register rules, start runtime | `pkg/asentric`, `config/`, `runtime.go` |
| **runtime.go** | Glue code - build semua komponen dari config | `ingest/`, `pipeline/`, `state/`, `alert/`, `internal/*` |
| **config/** | Bridge YAML → runtime config | `pkg/asentric` (LoadConfig) |
| **ingest/** | EventSource implementation | `pkg/asentric.EventSource`, `internal/source` |
| **pipeline/** | Event processing orchestration | `internal/runtime`, `internal/dispatcher` |
| **state/** | Runtime state (processed blocks, offsets) | Redis client, `pkg/asentric` (tidak ada) |
| **alert/** | AlertSink implementation | `pkg/asentric.AlertSink` |

---

### 2.2 Dependency Graph

```
main.go
    │
    ├─▶ asentric.LoadConfig()  (pkg/asentric)
    ├─▶ asentric.NewEngine()   (pkg/asentric)
    └─▶ runtime.go
            │
            ├─▶ config/loader.go
            ├─▶ ingest/evm_logs.go
            │       └─▶ internal/source/websocket.go
            ├─▶ pipeline/dispatcher.go
            │       └─▶ internal/dispatcher/dispatcher.go
            ├─▶ state/redis.go
            │       └─▶ Redis client
            ├─▶ alert/webhook.go
            │       └─▶ HTTP client
            └─▶ internal/runtime/runtime.go
                    ├─▶ EventSource (dari ingest/)
                    └─▶ Dispatcher (dari pipeline/)
```

**Penting:** Semua dependency ke `internal/*` hanya di `cmd/runtime-reference/`, tidak pernah di `pkg/*`.

---

## 3. main.go – Entry Point

### 3.1 Peran

`main.go` adalah **entry point** runtime reference yang:

- ✅ Load konfigurasi dari YAML
- ✅ Membuat Engine dan register rules
- ✅ Membangun runtime dari config + engine
- ✅ Menjalankan runtime sampai SIGINT/SIGTERM

---

### 3.2 Struktur Kode (Konseptual)

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/cmd/runtime-reference/rules"
    "github.com/asentric/asentric/cmd/runtime-reference/runtime"
)

func main() {
    // 1. Load configuration
    config, registry, err := loadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 2. Create engine and register rules
    engine := asentric.NewEngine()
    registerRules(engine)
    
    // 3. Build runtime from config + engine
    rt, err := runtime.NewRuntimeReference(config, registry, engine)
    if err != nil {
        log.Fatalf("Failed to build runtime: %v", err)
    }
    
    // 4. Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    
    // 5. Start runtime in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- rt.Start(ctx)
    }()
    
    // 6. Wait for signal or error
    select {
    case sig := <-sigCh:
        log.Printf("Received signal: %v, shutting down...", sig)
        cancel()
        if err := rt.Stop(); err != nil {
            log.Printf("Error stopping runtime: %v", err)
        }
        // Wait for runtime to finish
        if err := <-errCh; err != nil {
            log.Printf("Runtime error: %v", err)
            os.Exit(1)
        }
    case err := <-errCh:
        if err != nil {
            log.Fatalf("Runtime error: %v", err)
        }
    }
}

func loadConfig() (*asentric.RuntimeConfig, *asentric.RegistryConfig, error) {
    // Panggil asentric.LoadConfig() dari SDK
    config, err := asentric.LoadConfig("config/")
    if err != nil {
        return nil, nil, err
    }
    
    // Load registry.yaml (jika ada helper di config/)
    registry, err := config.LoadRegistry("config/registry.yaml")
    if err != nil {
        return nil, nil, err
    }
    
    return config, registry, nil
}

func registerRules(engine *asentric.Engine) {
    // Register rules untuk MVP use cases
    engine.RegisterRule(&rules.LargeTransferRule{
        Threshold: big.NewInt(1e18), // 1 ETH
    })
    
    engine.RegisterRule(&rules.ProxyUpgradeRule{})
    engine.RegisterRule(&rules.OwnershipChangeRule{})
    engine.RegisterRule(&rules.VaultWithdrawalRule{
        Threshold: big.NewInt(1e20), // 100 ETH
    })
}
```

---

### 3.3 Checklist Implementasi

- [ ] **Load config:**
  - Panggil `asentric.LoadConfig("config/")` untuk mendapatkan `RuntimeConfig`
  - Load `registry.yaml` untuk mendapatkan `RegistryConfig` (atau gunakan helper dari `config/`)
  - Handle error dengan log + exit non-zero

- [ ] **Create engine:**
  - `engine := asentric.NewEngine()`
  - Register rules untuk MVP use cases (large transfer, vault, upgrade, ownership)

- [ ] **Build runtime:**
  - Panggil fungsi dari `runtime.go` (misalnya `runtime.NewRuntimeReference()`)
  - Pass `config`, `registry`, dan `engine`
  - Handle error dengan log + exit non-zero

- [ ] **Graceful shutdown:**
  - Setup signal handler untuk SIGINT/SIGTERM
  - Start runtime di goroutine terpisah
  - Wait for signal atau error
  - Panggil `rt.Stop()` saat signal diterima
  - Exit dengan code yang sesuai (0 untuk graceful, 1 untuk error)

---

### 3.4 Error Handling

| Error Scenario | Behavior | Sesuai SPEC |
|----------------|----------|-------------|
| Config load error | Log + exit(1) | ✅ |
| Runtime build error | Log + exit(1) | ✅ |
| Runtime start error | Log + exit(1) | ✅ |
| SIGINT/SIGTERM | Call Stop(), wait, exit(0) | ✅ SPEC §12.3 |
| Runtime error during execution | Log + exit(1) | ✅ |

---

## 4. runtime.go – Glue Code

### 4.1 Peran

`runtime.go` adalah **glue code** yang:

- ✅ Menerima config + engine dari `main.go`
- ✅ Membangun semua komponen internal (EventSource, Dispatcher, AlertSink, dll.)
- ✅ Menyatukan mereka menjadi `internal/runtime.Runtime`
- ✅ Menyediakan facade untuk start/stop runtime

---

### 4.2 Struktur Kode (Konseptual)

```go
package runtime

import (
    "context"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/internal/runtime"
    "github.com/asentric/asentric/internal/dispatcher"
    "github.com/asentric/asentric/cmd/runtime-reference/config"
    "github.com/asentric/asentric/cmd/runtime-reference/ingest"
    "github.com/asentric/asentric/cmd/runtime-reference/alert"
    "github.com/asentric/asentric/cmd/runtime-reference/state"
)

// RuntimeReference adalah facade untuk reference runtime
type RuntimeReference struct {
    config    *asentric.RuntimeConfig
    registry  *asentric.RegistryConfig
    engine    *asentric.Engine
    
    // Internal components
    source     asentric.EventSource
    dispatcher asentric.Dispatcher
    sink       asentric.AlertSink
    
    // Runtime instance
    runtime    *runtime.Runtime
    
    // State store (optional, untuk resume/restart)
    stateStore state.Store
}

// NewRuntimeReference builds semua komponen dari config + engine
func NewRuntimeReference(
    config *asentric.RuntimeConfig,
    registry *asentric.RegistryConfig,
    engine *asentric.Engine,
) (*RuntimeReference, error) {
    rt := &RuntimeReference{
        config:   config,
        registry: registry,
        engine:   engine,
    }
    
    // 1. Build state store (Redis atau memory)
    if err := rt.buildStateStore(); err != nil {
        return nil, err
    }
    
    // 2. Build EventSource (WebSocket)
    if err := rt.buildEventSource(); err != nil {
        return nil, err
    }
    
    // 3. Build ABIRegistry dari registry.yaml
    abiRegistry, err := rt.buildABIRegistry()
    if err != nil {
        return nil, err
    }
    
    // 4. Build ContextBuilder
    contextBuilder := rt.buildContextBuilder(abiRegistry)
    
    // 5. Build AlertSink (webhook)
    if err := rt.buildAlertSink(); err != nil {
        return nil, err
    }
    
    // 6. Build Dispatcher
    if err := rt.buildDispatcher(contextBuilder, abiRegistry); err != nil {
        return nil, err
    }
    
    // 7. Build internal/runtime.Runtime
    rt.runtime = runtime.NewRuntime(rt.source, rt.dispatcher)
    
    return rt, nil
}

func (rt *RuntimeReference) buildStateStore() error {
    // Pilih implementasi berdasarkan config atau env
    if rt.config.Redis.Addr != "" {
        store, err := state.NewRedisStore(rt.config.Redis)
        if err != nil {
            return err
        }
        rt.stateStore = store
    } else {
        // Fallback ke memory store untuk dev
        rt.stateStore = state.NewMemoryStore()
    }
    return nil
}

func (rt *RuntimeReference) buildEventSource() error {
    // Build WebSocket EventSource
    // Menggunakan internal/source atau implementasi di ingest/
    addresses := make([]string, 0, len(rt.registry.Targets))
    for _, target := range rt.registry.Targets {
        addresses = append(addresses, target.Address)
    }
    
    source, err := ingest.NewWebSocketSource(
        rt.config.Chain.RPCWS,
        addresses,
    )
    if err != nil {
        return err
    }
    
    rt.source = source
    return nil
}

func (rt *RuntimeReference) buildABIRegistry() (domain.ABIRegistry, error) {
    // Load ABI files dari registry.yaml
    // Menggunakan internal/abi
    registry := abi.NewRegistry()
    
    for _, target := range rt.registry.Targets {
        if err := registry.LoadABI(target.Address, target.ABIPath); err != nil {
            return nil, err
        }
    }
    
    return registry, nil
}

func (rt *RuntimeReference) buildContextBuilder(abiRegistry domain.ABIRegistry) *context.EventContextBuilder {
    // Build ContextBuilder menggunakan internal/context + internal/adapter
    return context.NewEventContextBuilder(abiRegistry)
}

func (rt *RuntimeReference) buildAlertSink() error {
    // Build webhook AlertSink
    sink, err := alert.NewWebhookSink(rt.config.Webhook)
    if err != nil {
        return err
    }
    
    rt.sink = sink
    return nil
}

func (rt *RuntimeReference) buildDispatcher(
    contextBuilder *context.EventContextBuilder,
    abiRegistry domain.ABIRegistry,
) error {
    // Build EngineDispatcher menggunakan internal/dispatcher
    dispatcher := dispatcher.NewEngineDispatcher(
        rt.engine,
        rt.sink,
        contextBuilder,
        abiRegistry,
    )
    
    rt.dispatcher = dispatcher
    return nil
}

// Start starts the runtime (blocks until stopped)
func (rt *RuntimeReference) Start(ctx context.Context) error {
    return rt.runtime.Start(ctx)
}

// Stop gracefully stops the runtime
func (rt *RuntimeReference) Stop() error {
    return rt.runtime.Stop()
}
```

---

### 4.3 Checklist Implementasi

- [ ] **Menerima input:**
  - `RuntimeConfig` + `RegistryConfig` dari `asentric.LoadConfig()`
  - Pointer ke `*asentric.Engine` dari user

- [ ] **Build state store:**
  - Pilih implementasi (Redis atau memory) berdasarkan config
  - Handle error connection

- [ ] **Build EventSource:**
  - Extract addresses dari `registry.Targets`
  - Build WebSocket source dengan `chain.rpc_ws`
  - Handle error connection

- [ ] **Build ABIRegistry:**
  - Load ABI files dari `abi_path` di setiap target
  - Menggunakan `internal/abi` untuk parsing
  - Handle error file tidak ditemukan atau invalid ABI

- [ ] **Build ContextBuilder:**
  - Menggunakan `internal/context` + `internal/adapter`
  - Pass `ABIRegistry` untuk decoding

- [ ] **Build AlertSink:**
  - Build webhook sink dengan `webhook.url`
  - Handle error invalid URL

- [ ] **Build Dispatcher:**
  - Menggunakan `internal/dispatcher.EngineDispatcher`
  - Pass `Engine`, `AlertSink`, `ContextBuilder`, `ABIRegistry`

- [ ] **Build internal/runtime.Runtime:**
  - `runtime.NewRuntime(source, dispatcher)`
  - Return facade untuk start/stop

---

### 4.4 Error Handling

| Error Scenario | Behavior | Return |
|----------------|----------|--------|
| Redis connection error | Return error | ❌ |
| WebSocket connection test error | Return error | ❌ |
| ABI file not found | Return error | ❌ |
| Invalid webhook URL | Return error | ❌ |
| Dispatcher build error | Return error | ❌ |

**Semua error di build time harus di-handle dan return error, tidak boleh panic.**

---

## 5. config/ – Configuration Bridge

### 5.1 Peran

`config/` adalah **bridge** antara:

- ✅ YAML files (`asentric.yaml`, `registry.yaml`) → `RuntimeConfig`/`RegistryConfig`
- ✅ Runtime-specific config needs (jika ada) → struct lokal

**Penting:** `config/` **tidak menduplikasi** parsing yang sudah dilakukan `asentric.LoadConfig()`. Ini hanya helper untuk akses yang lebih mudah.

---

### 5.2 Struktur Kode (Konseptual)

```go
package config

import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
)

// Loader adalah helper untuk load dan akses config
type Loader struct {
    RuntimeConfig *asentric.RuntimeConfig
    RegistryConfig *asentric.RegistryConfig
}

// NewLoader loads config dari directory
func NewLoader(configDir string) (*Loader, error) {
    // Panggil asentric.LoadConfig() - jangan duplikasi parsing
    runtimeConfig, err := asentric.LoadConfig(configDir)
    if err != nil {
        return nil, err
    }
    
    // Load registry.yaml (jika ada helper di SDK, gunakan itu)
    // Jika tidak, parse sendiri tapi tetap konsisten dengan format SPEC
    registryConfig, err := loadRegistryConfig(configDir + "/registry.yaml")
    if err != nil {
        return nil, err
    }
    
    return &Loader{
        RuntimeConfig:  runtimeConfig,
        RegistryConfig: registryConfig,
    }, nil
}

// GetTargetAddresses returns semua addresses dari registry
func (l *Loader) GetTargetAddresses() []domain.Address {
    addresses := make([]domain.Address, 0, len(l.RegistryConfig.Targets))
    for _, target := range l.RegistryConfig.Targets {
        addresses = append(addresses, domain.Address(target.Address))
    }
    return addresses
}

// GetABIPaths returns map address -> abi_path
func (l *Loader) GetABIPaths() map[domain.Address]string {
    paths := make(map[domain.Address]string)
    for _, target := range l.RegistryConfig.Targets {
        paths[domain.Address(target.Address)] = target.ABIPath
    }
    return paths
}

// GetRedisConfig returns Redis config untuk state store
func (l *Loader) GetRedisConfig() *asentric.RedisConfig {
    return &l.RuntimeConfig.Redis
}

// GetWebhookConfig returns Webhook config untuk alert sink
func (l *Loader) GetWebhookConfig() *asentric.WebhookConfig {
    return &l.RuntimeConfig.Webhook
}

// GetChainConfig returns Chain config untuk EventSource
func (l *Loader) GetChainConfig() *asentric.ChainConfig {
    return &l.RuntimeConfig.Chain
}
```

---

### 5.3 Checklist Implementasi

- [ ] **Tidak duplikasi parsing:**
  - Gunakan `asentric.LoadConfig()` untuk `asentric.yaml`
  - Jangan re-implement YAML parsing yang sudah ada di SDK

- [ ] **Helper functions:**
  - `GetTargetAddresses()` - extract addresses dari registry
  - `GetABIPaths()` - map address ke abi_path
  - `GetRedisConfig()`, `GetWebhookConfig()`, `GetChainConfig()` - akses mudah ke config sections

- [ ] **Konsistensi dengan SPEC:**
  - Semua field dan default harus match dengan `SPEC.md §4.4`
  - Validasi field required (jika perlu)

---

## 6. ingest/ – Event Source Implementation

### 6.1 Peran

`ingest/` mengimplementasikan **EventSource** yang:

- ✅ Connect ke WebSocket RPC endpoint
- ✅ Subscribe `eth_subscribe("logs", {addresses})`
- ✅ Parse logs → `asentric.Event`
- ✅ Stream events via channel

---

### 6.2 Struktur Kode (Konseptual)

```go
package ingest

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
)

// WebSocketSource implements asentric.EventSource
type WebSocketSource struct {
    rpcURL    string
    addresses []domain.Address
    
    // WebSocket connection (gunakan library seperti gorilla/websocket)
    conn *websocket.Conn
}

// NewWebSocketSource creates new WebSocket EventSource
func NewWebSocketSource(rpcURL string, addresses []domain.Address) (*WebSocketSource, error) {
    return &WebSocketSource{
        rpcURL:    rpcURL,
        addresses: addresses,
    }, nil
}

// Start implements asentric.EventSource
func (s *WebSocketSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    // 1. Connect ke WebSocket
    conn, err := websocket.Dial(s.rpcURL, "", "http://localhost/")
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }
    s.conn = conn
    
    // 2. Subscribe logs
    addressesHex := make([]string, 0, len(s.addresses))
    for _, addr := range s.addresses {
        addressesHex = append(addressesHex, addr.Hex())
    }
    
    subscribeReq := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      1,
        "method":  "eth_subscribe",
        "params": []interface{}{
            "logs",
            map[string]interface{}{
                "address": addressesHex,
            },
        },
    }
    
    if err := conn.WriteJSON(subscribeReq); err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to subscribe: %w", err)
    }
    
    // 3. Read subscription ID
    var subscribeResp map[string]interface{}
    if err := conn.ReadJSON(&subscribeResp); err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to read subscription: %w", err)
    }
    
    // 4. Create channel dan start goroutine untuk read logs
    ch := make(chan asentric.Event, 100) // Buffered channel
    
    go s.readLogs(ctx, ch)
    
    return ch, nil
}

func (s *WebSocketSource) readLogs(ctx context.Context, ch chan<- asentric.Event) {
    defer close(ch)
    defer s.conn.Close()
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Read log notification dari WebSocket
            var notification map[string]interface{}
            if err := s.conn.ReadJSON(&notification); err != nil {
                // Error atau connection closed
                // Sesuai SPEC: log error dan exit (tidak auto-reconnect)
                return
            }
            
            // Parse log notification
            event, err := s.parseLogNotification(notification)
            if err != nil {
                // Log error tapi continue (skip invalid log)
                continue
            }
            
            // Send event ke channel
            select {
            case ch <- event:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (s *WebSocketSource) parseLogNotification(notification map[string]interface{}) (asentric.Event, error) {
    // Extract params dari notification
    params, ok := notification["params"].(map[string]interface{})
    if !ok {
        return asentric.Event{}, fmt.Errorf("invalid notification format")
    }
    
    result, ok := params["result"].(map[string]interface{})
    if !ok {
        return asentric.Event{}, fmt.Errorf("invalid result format")
    }
    
    // Parse log fields
    blockNumberHex := result["blockNumber"].(string)
    txHash := result["transactionHash"].(string)
    
    blockNumber, err := hexutil.DecodeUint64(blockNumberHex)
    if err != nil {
        return asentric.Event{}, err
    }
    
    // Build asentric.Event
    event := asentric.Event{
        ChainID:     s.chainID, // Dari config atau auto-detect
        BlockNumber: blockNumber,
        TxHash:      txHash,
        Payload:     result, // Raw log data untuk decoding nanti
    }
    
    return event, nil
}
```

---

### 6.3 Checklist Implementasi

- [ ] **Implement EventSource interface:**
  - `Start(ctx) (<-chan Event, error)`
  - Return channel yang akan di-close saat context cancelled atau error

- [ ] **WebSocket connection:**
  - Connect ke `chain.rpc_ws` dari config
  - Handle connection error → return error (sesuai SPEC: exit, tidak retry)

- [ ] **Subscribe logs:**
  - `eth_subscribe("logs", {addresses})`
  - Addresses dari `registry.yaml` targets
  - Handle subscription error → return error

- [ ] **Parse logs:**
  - Parse WebSocket notification → `asentric.Event`
  - Extract `ChainID`, `BlockNumber`, `TxHash`, `Payload`
  - Handle parse error → skip log (log error, continue)

- [ ] **Stream events:**
  - Send events ke channel
  - Close channel saat context cancelled atau connection error
  - **Tidak auto-reconnect** (sesuai SPEC §12.2)

---

### 6.4 Error Semantics

| Error Scenario | Behavior | Sesuai SPEC |
|----------------|----------|-------------|
| WebSocket connection error | Return error dari `Start()` | ✅ Exit |
| Subscription error | Return error dari `Start()` | ✅ Exit |
| Connection lost during execution | Close channel, goroutine exit | ✅ Exit (tidak retry) |
| Parse error (invalid log) | Log error, skip log, continue | ✅ |

---

## 7. pipeline/ – Event Processing Pipeline

### 7.1 Peran

`pipeline/` mengatur **aliran event processing**:

- ✅ **MVP:** Simple pipeline → langsung ke `internal/runtime.Runtime`
- ✅ **Advanced:** Worker pool untuk parallelism (jika perlu)

**Penting:** Engine tetap single-threaded. Parallelism dilakukan di level worker, bukan di Engine.

---

### 7.2 Struktur Kode (Konseptual - MVP Simple)

```go
package pipeline

import (
    "github.com/asentric/asentric/internal/runtime"
    "github.com/asentric/asentric/pkg/asentric"
)

// SimplePipeline adalah pipeline sederhana untuk MVP
// Langsung menggunakan internal/runtime.Runtime tanpa worker pool
type SimplePipeline struct {
    runtime *runtime.Runtime
}

// NewSimplePipeline creates simple pipeline
func NewSimplePipeline(
    source asentric.EventSource,
    dispatcher asentric.Dispatcher,
) *SimplePipeline {
    return &SimplePipeline{
        runtime: runtime.NewRuntime(source, dispatcher),
    }
}

// Start starts the pipeline (delegates to internal runtime)
func (p *SimplePipeline) Start(ctx context.Context) error {
    return p.runtime.Start(ctx)
}

// Stop stops the pipeline
func (p *SimplePipeline) Stop() error {
    return p.runtime.Stop()
}
```

---

### 7.3 Struktur Kode (Konseptual - Worker Pool - Optional)

```go
package pipeline

import (
    "context"
    "sync"
    
    "github.com/asentric/asentric/internal/runtime"
    "github.com/asentric/asentric/pkg/asentric"
)

// WorkerPoolPipeline adalah pipeline dengan worker pool
// Setiap worker punya Engine sendiri (Engine tidak concurrency-safe)
type WorkerPoolPipeline struct {
    source     asentric.EventSource
    numWorkers int
    
    workers []*Worker
    wg      sync.WaitGroup
}

type Worker struct {
    id        int
    engine    *asentric.Engine  // Setiap worker punya engine sendiri
    dispatcher asentric.Dispatcher
    runtime   *runtime.Runtime
}

// NewWorkerPoolPipeline creates pipeline dengan worker pool
func NewWorkerPoolPipeline(
    source asentric.EventSource,
    numWorkers int,
    engineFactory func() *asentric.Engine, // Factory untuk create engine per worker
    dispatcherFactory func(*asentric.Engine) asentric.Dispatcher,
) *WorkerPoolPipeline {
    workers := make([]*Worker, numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        engine := engineFactory()
        dispatcher := dispatcherFactory(engine)
        
        workers[i] = &Worker{
            id:        i,
            engine:    engine,
            dispatcher: dispatcher,
            runtime:   runtime.NewRuntime(source, dispatcher), // Setiap worker punya runtime sendiri
        }
    }
    
    return &WorkerPoolPipeline{
        source:     source,
        numWorkers: numWorkers,
        workers:    workers,
    }
}

// Start starts all workers
func (p *WorkerPoolPipeline) Start(ctx context.Context) error {
    // Round-robin: distribute events ke workers
    // Atau gunakan channel fan-out pattern
    // ...
    return nil
}
```

---

### 7.4 Checklist Implementasi

- [ ] **Tentukan strategi:**
  - **MVP:** Simple pipeline (langsung ke `internal/runtime.Runtime`)
  - **Advanced (optional):** Worker pool jika butuh throughput tinggi

- [ ] **Jika simple pipeline:**
  - Wrap `internal/runtime.Runtime`
  - Delegate `Start()` dan `Stop()` ke internal runtime

- [ ] **Jika worker pool:**
  - Setiap worker punya Engine sendiri (Engine tidak concurrency-safe)
  - Setiap worker punya Dispatcher sendiri
  - Distribute events ke workers (round-robin atau fan-out)
  - **Tidak** share Engine antar workers

- [ ] **Jamin determinism:**
  - Same Event + same rule set → same Alerts (bahkan dengan worker pool)

---

## 8. state/ – Runtime State Management

### 8.1 Peran

`state/` mengelola **runtime state ephemeral**:

- ✅ Processed block number
- ✅ Log index terakhir
- ✅ Offset untuk resume/restart

**Penting:** State ini untuk **runtime orchestration**, bukan untuk rule logic. Engine tetap stateless.

---

### 8.2 Struktur Kode (Konseptual)

```go
package state

import (
    "github.com/asentric/asentric/pkg/asentric"
)

// Store adalah interface untuk state store
type Store interface {
    // GetLastProcessedBlock returns last processed block number
    GetLastProcessedBlock(chainID uint64) (uint64, error)
    
    // SetLastProcessedBlock sets last processed block number
    SetLastProcessedBlock(chainID uint64, blockNumber uint64) error
    
    // GetLastLogIndex returns last processed log index untuk block tertentu
    GetLastLogIndex(chainID uint64, blockNumber uint64) (uint64, error)
    
    // SetLastLogIndex sets last processed log index
    SetLastLogIndex(chainID uint64, blockNumber uint64, logIndex uint64) error
}

// RedisStore implements Store menggunakan Redis
type RedisStore struct {
    client *redis.Client
}

func NewRedisStore(config *asentric.RedisConfig) (*RedisStore, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     config.Addr,
        Password: config.Password,
        DB:       config.DB,
    })
    
    // Test connection
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }
    
    return &RedisStore{client: client}, nil
}

func (s *RedisStore) GetLastProcessedBlock(chainID uint64) (uint64, error) {
    key := fmt.Sprintf("asentric:chain:%d:last_block", chainID)
    val, err := s.client.Get(context.Background(), key).Uint64()
    if err == redis.Nil {
        return 0, nil // Belum ada state
    }
    return val, err
}

func (s *RedisStore) SetLastProcessedBlock(chainID uint64, blockNumber uint64) error {
    key := fmt.Sprintf("asentric:chain:%d:last_block", chainID)
    return s.client.Set(context.Background(), key, blockNumber, 0).Err()
}

// MemoryStore implements Store menggunakan in-memory map (untuk dev/test)
type MemoryStore struct {
    lastBlocks map[uint64]uint64
    lastLogs   map[string]uint64
    mu         sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        lastBlocks: make(map[uint64]uint64),
        lastLogs:   make(map[string]uint64),
    }
}

func (s *MemoryStore) GetLastProcessedBlock(chainID uint64) (uint64, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.lastBlocks[chainID], nil
}

func (s *MemoryStore) SetLastProcessedBlock(chainID uint64, blockNumber uint64) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.lastBlocks[chainID] = blockNumber
    return nil
}
```

---

### 8.3 Checklist Implementasi

- [ ] **Define Store interface:**
  - Methods untuk get/set last processed block
  - Methods untuk get/set last log index (jika perlu)

- [ ] **Implement RedisStore:**
  - Connect ke Redis menggunakan config
  - Key naming convention (misalnya `asentric:chain:{chainID}:last_block`)
  - Handle connection error

- [ ] **Implement MemoryStore:**
  - In-memory map untuk dev/test
  - Thread-safe dengan mutex

- [ ] **Jamin Engine stateless:**
  - State hanya digunakan oleh runtime untuk resume
  - Engine tidak pernah akses state store

---

## 9. alert/ – Alert Delivery

### 9.1 Peran

`alert/` mengimplementasikan **AlertSink** untuk:

- ✅ Webhook delivery (required untuk MVP)
- ✅ (Optional) Multiple sinks (Telegram, stdout, dll.)

---

### 9.2 Struktur Kode (Konseptual)

```go
package alert

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
)

// WebhookSink implements asentric.AlertSink
type WebhookSink struct {
    url     string
    timeout time.Duration
    client  *http.Client
}

// NewWebhookSink creates new webhook AlertSink
func NewWebhookSink(config *asentric.WebhookConfig) (*WebhookSink, error) {
    timeout := 10 * time.Second
    if config.Timeout > 0 {
        timeout = config.Timeout
    }
    
    return &WebhookSink{
        url:     config.URL,
        timeout: timeout,
        client: &http.Client{
            Timeout: timeout,
        },
    }, nil
}

// Emit implements asentric.AlertSink
func (s *WebhookSink) Emit(ctx context.Context, alert *asentric.Alert) error {
    // 1. Build JSON payload sesuai SPEC §7.1
    payload := s.buildPayload(alert)
    
    // 2. Serialize ke JSON
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal alert: %w", err)
    }
    
    // 3. HTTP POST ke webhook
    req, err := http.NewRequestWithContext(ctx, "POST", s.url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := s.client.Do(req)
    if err != nil {
        // Sesuai SPEC: log error, continue (tidak return error)
        // Tapi untuk sekarang, kita return error dan letakkan handling di Dispatcher
        return fmt.Errorf("failed to send webhook: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("webhook returned status %d", resp.StatusCode)
    }
    
    return nil
}

// buildPayload builds JSON payload sesuai SPEC §7.1
func (s *WebhookSink) buildPayload(alert *asentric.Alert) map[string]interface{} {
    payload := map[string]interface{}{
        "severity":    string(alert.Severity),
        "rule":        alert.Rule,
        "title":       alert.Title,
        "description": alert.Description,
    }
    
    // Network info (dari context atau config)
    if alert.Ref != nil {
        payload["network"] = map[string]interface{}{
            "name":     "Mantle", // Dari config
            "chain_id": 5000,     // Dari context atau config
        }
        
        payload["context"] = map[string]interface{}{
            "block_number": alert.Ref.BlockNumber,
            "tx_hash":      alert.Ref.TxHash,
            "timestamp":    time.Now().UTC().Format(time.RFC3339),
        }
    }
    
    // Details dari Metadata
    payload["details"] = alert.Metadata
    
    return payload
}
```

---

### 9.3 Checklist Implementasi

- [ ] **Implement AlertSink interface:**
  - `Emit(ctx, *Alert) error`
  - Return error hanya untuk fatal errors (bukan untuk retry-able errors)

- [ ] **Build JSON payload:**
  - Sesuai format di `SPEC.md §7.1`:
    - `severity`, `rule`, `title`, `description`
    - `network` (name, chain_id)
    - `context` (block_number, tx_hash, timestamp)
    - `details` (dari `Alert.Metadata`)

- [ ] **HTTP POST:**
  - POST ke `webhook.url` dari config
  - Set `Content-Type: application/json`
  - Handle timeout dari config

- [ ] **Error handling:**
  - **Sesuai SPEC §12.2:** Webhook error → log error, **continue** (tidak stop runtime)
  - Implementasi: return error dari `Emit()`, tapi Dispatcher harus handle dengan log & continue

---

### 9.4 Error Semantics

| Error Scenario | Behavior | Sesuai SPEC |
|----------------|----------|-------------|
| Invalid webhook URL | Return error dari constructor | ✅ |
| HTTP request error | Return error dari `Emit()` | ⚠️ Dispatcher harus log & continue |
| HTTP status != 2xx | Return error dari `Emit()` | ⚠️ Dispatcher harus log & continue |
| Timeout | Return error dari `Emit()` | ⚠️ Dispatcher harus log & continue |

**Catatan:** Error dari `Emit()` seharusnya **tidak** menghentikan runtime. Dispatcher harus catch error dan log, tapi continue processing.

---

## 10. Integrasi Global & Boundary Rules

### 10.1 Boundary & Imports

**Allowed:**
```go
// ✅ BOLEH
import "github.com/asentric/asentric/pkg/asentric"
import "github.com/asentric/asentric/pkg/domain"
import "github.com/asentric/asentric/internal/runtime"
import "github.com/asentric/asentric/internal/dispatcher"
// ... semua internal/*
```

**Forbidden:**
```go
// ❌ TIDAK BOLEH - tidak ada yang perlu di-forbid di runtime-reference
// karena ini adalah binary, bukan library
```

**Penting:** Runtime reference adalah **binary**, bukan library. Jadi tidak ada constraint "tidak boleh expose internal". Tapi tetap **best practice** untuk tidak expose internal types di public API jika nanti ada bagian yang di-reuse.

---

### 10.2 Flow Canonical Verification

Verifikasi bahwa flow mengikuti `SPEC.md §3`:

```
[RPC WebSocket]
    │
    │ eth_subscribe("logs", {addresses})
    ▼
[ingest/WebSocketSource]  ← EventSource.Start()
    │
    │ <-chan Event
    ▼
[internal/runtime.Runtime]  ← loop() membaca dari channel
    │
    │ Dispatch(event)
    ▼
[internal/dispatcher.EngineDispatcher]  ← Dispatcher.Dispatch()
    │
    │ Build Context → Engine.Evaluate()
    ▼
[asentric.Engine]  ← Evaluate(ctx)
    │
    │ []*Alert
    ▼
[alert/WebhookSink]  ← AlertSink.Emit()
    │
    │ HTTP POST
    ▼
[Webhook Receiver]
```

**Checklist:**
- [ ] Event hanya dari WebSocket subscription (tidak dari sumber lain)
- [ ] Event masuk channel dari EventSource
- [ ] Runtime loop membaca event dari channel (single-threaded)
- [ ] Dispatcher membangun Context dari Event
- [ ] Engine.Evaluate() dipanggil dengan Context
- [ ] Alert dikirim ke AlertSink
- [ ] AlertSink mengirim HTTP POST ke webhook

---

### 10.3 Error Semantics & Shutdown

**Konsisten dengan `SPEC.md §12`:**

| Error Type | Behavior | Implementation |
|------------|----------|----------------|
| **Rule error/panic** | Engine recover, return `ErrRulePanic` | Di `internal/runtime` dan `internal/dispatcher` |
| **WebSocket error** | Log error, **exit** | Di `ingest/WebSocketSource` |
| **Redis error** | Log error, **exit** | Di `state/RedisStore` atau `runtime.go` |
| **Webhook error** | Log error, **continue** | Di `alert/WebhookSink` + `internal/dispatcher` |
| **Graceful shutdown** | Handle SIGINT/SIGTERM | Di `main.go` |

---

## 11. Contoh Implementasi Lengkap

### 11.1 Flow End-to-End untuk Large Transfer Detection

**Timeline:**

1. **T0: main.go start**
   ```go
   config, _ := asentric.LoadConfig("config/")
   engine := asentric.NewEngine()
   engine.RegisterRule(&rules.LargeTransferRule{Threshold: big.NewInt(1e18)})
   runtime := runtime.NewRuntimeReference(config, registry, engine)
   runtime.Start(ctx)
   ```

2. **T1: runtime.go build components**
   ```go
   // Build EventSource
   source := ingest.NewWebSocketSource("wss://...", addresses)
   
   // Build Dispatcher
   dispatcher := dispatcher.NewEngineDispatcher(engine, sink, contextBuilder, abiRegistry)
   
   // Build internal/runtime.Runtime
   rt := runtime.NewRuntime(source, dispatcher)
   ```

3. **T2: EventSource.Start()**
   ```go
   // Connect WebSocket
   // Subscribe eth_subscribe("logs", {addresses})
   // Start goroutine untuk read logs
   // Return <-chan Event
   ```

4. **T3: internal/runtime.Runtime.Start()**
   ```go
   // Panggil source.Start(ctx) → dapat channel
   // Start loop() yang membaca dari channel
   ```

5. **T4: Event masuk**
   ```go
   // WebSocket menerima log
   // Parse → asentric.Event
   // Send ke channel: ch <- event
   ```

6. **T5: Runtime loop menerima event**
   ```go
   // Di loop(): case event, ok := <-r.events
   // Panggil dispatcher.Dispatch(ctx, event)
   ```

7. **T6: Dispatcher memproses**
   ```go
   // Build Context dari Event
   execCtx := contextBuilder.Build(event)
   
   // Evaluate rules
   alerts, _ := engine.Evaluate(execCtx)
   
   // Emit alerts
   for _, alert := range alerts {
       sink.Emit(ctx, alert)
   }
   ```

8. **T7: Alert dikirim**
   ```go
   // WebhookSink.Emit()
   // Build JSON payload
   // HTTP POST ke webhook.url
   ```

---

### 11.2 Contoh Rule untuk MVP Use Cases

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

func (r *LargeTransferRule) Name() string {
    return "large_transfer_detection"
}

func (r *LargeTransferRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    tx := ctx.Tx()
    
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

```go
// rules/proxy_upgrade.go
package rules

import (
    "github.com/asentric/asentric/pkg/asentric"
)

type ProxyUpgradeRule struct{}

func (r *ProxyUpgradeRule) Name() string {
    return "proxy_upgrade_detection"
}

func (r *ProxyUpgradeRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        if log.Event.Name == "Upgraded" {
            return asentric.NewAlert(
                r.Name(),
                asentric.SeverityCritical,
                "Proxy Upgrade Detected",
                "Proxy implementation has been upgraded",
            ).WithMetadata("old_implementation", log.Event.Fields["oldImplementation"]).
              WithMetadata("new_implementation", log.Event.Fields["newImplementation"]).
              WithMetadata("proxy", log.Address.String()), nil
        }
    }
    
    return nil, nil
}
```

---

## Summary

Runtime reference adalah **contoh implementasi lengkap** yang:

1. ✅ Menggunakan public SDK seperti user eksternal
2. ✅ Menggunakan internal implementation untuk infrastruktur
3. ✅ Mengikuti canonical flow dari SPEC end-to-end
4. ✅ Menjadi golden reference untuk tim dan user

**Sebagai Lead, pastikan:**

- ✅ Setiap komponen mengikuti checklist di dokumen ini
- ✅ Flow canonical sesuai SPEC §3
- ✅ Error semantics sesuai SPEC §12
- ✅ Boundary rules tidak dilanggar

---

## Related Documentation

- **[SPEC.md](SPEC.md)** - MVP specification (canonical flow di §3, config di §4, error di §12)
- **[architecture.md](architecture.md)** - Core architecture (boundaries di §9)
- **[project-structure.md](project-structure.md)** - Final structure (runtime-reference di §4.2)
- **[internal-runtime-deep-dive.md](internal-runtime-deep-dive.md)** - Event loop details
- **[sdk-api.md](sdk-api.md)** - Public API contracts

---

**Last Updated:** 2024  
**Version:** 1.0




