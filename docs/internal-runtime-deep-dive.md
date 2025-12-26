# Internal Runtime – Event Loop Canonical (Deep Dive)

> **🔒 Lihat MVP Spec:** [SPEC.md](SPEC.md) - **SINGLE SOURCE OF TRUTH**  
> **📖 Lihat architecture:** [architecture.md](architecture.md) - Core architecture  
> **📋 Lihat runtime reference:** [runtime-reference-checklist.md](runtime-reference-checklist.md) - Implementation checklist

**Status:** ✅ **AUTHORITATIVE**  
**Audience:** Core Engineers, Runtime Implementers, Code Reviewers

Dokumen ini menjelaskan **secara detail** bagaimana `internal/runtime` bekerja sebagai **event loop canonical** yang menghubungkan `EventSource` → `Dispatcher` → `Engine` → `AlertSink`.

**Jika terjadi konflik dengan [SPEC.md](SPEC.md), SPEC.md yang benar.**

---

## Table of Contents

1. [Overview & Purpose](#1-overview--purpose)
2. [Runtime Struct – Setiap Field Dijelaskan](#2-runtime-struct--setiap-field-dijelaskan)
3. [Start() Flow – Step by Step](#3-start-flow--step-by-step)
4. [loop() – Event Loop Canonical](#4-loop--event-loop-canonical)
5. [Stop() – Graceful Shutdown](#5-stop--graceful-shutdown)
6. [Mapping ke Canonical Flow SPEC](#6-mapping-ke-canonical-flow-spec)
7. [Contoh Konkret dengan Use Case](#7-contoh-konkret-dengan-use-case)
8. [Invariants & Guarantees](#8-invariants--guarantees)
9. [Error Semantics](#9-error-semantics)
10. [Concurrency Model](#10-concurrency-model)

---

## 1. Overview & Purpose

### 1.1 Apa itu `internal/runtime`?

`internal/runtime` adalah **komponen internal** yang bertanggung jawab untuk:

- ✅ Mengorkestrasi lifecycle `EventSource` dan `Dispatcher`
- ✅ Menjalankan **single-threaded event loop** yang membaca events dari channel
- ✅ Memastikan **canonical flow** dari SPEC dijalankan: Event → Context → Engine → Alert
- ✅ Mengelola **graceful shutdown** via context cancellation

### 1.2 Posisi di Arsitektur

```
┌─────────────────────────────────────────────────────────┐
│              internal/runtime (INI YANG KITA BEDAH)      │
│                                                          │
│  ┌──────────────┐         ┌──────────────┐            │
│  │ EventSource  │ ──────▶ │  Dispatcher  │            │
│  │ (interface) │         │ (interface)  │            │
│  └──────────────┘         └──────────────┘            │
│         │                        │                      │
│         │ <-chan Event           │ Dispatch()          │
│         │                        │                      │
│         └────────────────────────┘                      │
│                    │                                    │
│                    ▼                                    │
│         [Single-threaded loop]                          │
└─────────────────────────────────────────────────────────┘
```

**Penting:** `internal/runtime` adalah **komponen internal**, bukan public API. User tidak langsung menggunakan ini, melainkan melalui `pkg/asentric.Runtime` facade.

---

## 2. Runtime Struct – Setiap Field Dijelaskan

### 2.1 Definisi Struct

```go
type Runtime struct {
    Source     asentric.EventSource
    Dispatcher asentric.Dispatcher

    ctx    context.Context
    cancel context.CancelFunc

    events <-chan asentric.Event

    mu      sync.Mutex
    running bool
}
```

### 2.2 Penjelasan Setiap Field

#### `Source asentric.EventSource`

**Tipe:** Interface dari `pkg/asentric`

**Fungsi:**
- Menyediakan **stream events** dari blockchain (via WebSocket)
- Method `Start(ctx)` mengembalikan `<-chan asentric.Event`
- Implementasi konkret ada di `internal/source/` (misalnya `websocket.go`)

**Contoh implementasi (konseptual):**
```go
// internal/source/websocket.go (contoh)
type WebSocketSource struct {
    rpcURL string
}

func (s *WebSocketSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    ch := make(chan asentric.Event)
    go func() {
        // Connect ke WebSocket
        // Subscribe eth_subscribe("logs", {...})
        // Parse logs → asentric.Event
        // Kirim ke channel ch
    }()
    return ch, nil
}
```

**Di Runtime:**
- `Source` di-inject saat `NewRuntime(source, dispatcher)`
- `Source.Start(ctx)` dipanggil di `Start()` untuk mendapatkan channel

---

#### `Dispatcher asentric.Dispatcher`

**Tipe:** Interface dari `pkg/asentric`

**Fungsi:**
- Menghubungkan **Event → Context → Engine → AlertSink**
- Method `Dispatch(ctx, event)` memproses satu event
- Implementasi konkret ada di `internal/dispatcher/` (misalnya `dispatcher.go`)

**Contoh implementasi (konseptual):**
```go
// internal/dispatcher/dispatcher.go
type EngineDispatcher struct {
    Engine         *asentric.Engine
    Sink           asentric.AlertSink
    ContextBuilder ContextBuilder
    ABIRegistry    domain.ABIRegistry
}

func (d *EngineDispatcher) Dispatch(ctx context.Context, event asentric.Event) error {
    // 1. Build Context dari Event
    execCtx := d.ContextBuilder.Build(event)
    
    // 2. Evaluate rules
    alerts, err := d.Engine.Evaluate(execCtx)
    if err != nil {
        return err
    }
    
    // 3. Emit alerts
    for _, alert := range alerts {
        d.Sink.Emit(ctx, alert)
    }
    
    return nil
}
```

**Di Runtime:**
- `Dispatcher` di-inject saat `NewRuntime(source, dispatcher)`
- `Dispatcher.Dispatch()` dipanggil di `loop()` untuk setiap event

---

#### `ctx context.Context` dan `cancel context.CancelFunc`

**Tipe:** Standard Go context

**Fungsi:**
- **Lifecycle control** – mengontrol kapan runtime harus berhenti
- **Cancellation propagation** – ketika context di-cancel, semua operasi yang menggunakan context ini akan berhenti
- **Graceful shutdown** – memungkinkan runtime menyelesaikan event yang sedang diproses sebelum exit

**Di Runtime:**
- Dibuat di `Start()` dengan `context.WithCancel(parentCtx)`
- Di-cancel di `Stop()` untuk menghentikan runtime
- Di-check di `loop()` dengan `case <-r.ctx.Done()`

**Contoh penggunaan:**
```go
// Di Start()
ctx, cancel := context.WithCancel(parentCtx)
r.ctx = ctx
r.cancel = cancel

// Di Stop()
r.cancel()  // Ini akan trigger r.ctx.Done()

// Di loop()
case <-r.ctx.Done():
    return nil  // Exit gracefully
```

---

#### `events <-chan asentric.Event`

**Tipe:** Receive-only channel dari `asentric.Event`

**Fungsi:**
- **Stream events** dari `EventSource` ke runtime loop
- **Unidirectional** – runtime hanya membaca, tidak menulis
- **Buffered atau unbuffered** – tergantung implementasi `EventSource`

**Di Runtime:**
- Diisi di `Start()` dari `r.Source.Start(ctx)`
- Dibaca di `loop()` dengan `case event, ok := <-r.events`

**Penting:**
- Channel ini **tidak pernah di-close oleh Runtime**
- `EventSource` yang bertanggung jawab menutup channel (biasanya saat context di-cancel atau error)

---

#### `mu sync.Mutex` dan `running bool`

**Tipe:** Mutex untuk thread-safety, bool untuk state tracking

**Fungsi:**
- **Thread-safety** – melindungi akses ke `running` dari multiple goroutines
- **State tracking** – menandai apakah runtime sedang berjalan atau tidak
- **Idempotency** – memastikan `Start()` tidak bisa dipanggil dua kali

**Di Runtime:**
- `mu.Lock()` digunakan di `Start()` dan `Stop()` untuk protect `running`
- `running = true` saat runtime mulai
- `running = false` saat runtime selesai atau error

**Contoh:**
```go
// Di Start()
r.mu.Lock()
if r.running {
    r.mu.Unlock()
    return asentric.ErrAlreadyRunning  // Cegah double start
}
r.running = true
r.mu.Unlock()

// Di loop() defer
defer func() {
    r.mu.Lock()
    r.running = false  // Setelah loop selesai
    r.mu.Unlock()
}()
```

---

## 3. Start() Flow – Step by Step

### 3.1 Signature

```go
func (r *Runtime) Start(parentCtx context.Context) error
```

**Input:**
- `parentCtx` – context dari caller (biasanya `context.Background()` atau context dengan timeout)

**Output:**
- `error` – error jika startup gagal, atau `nil` jika loop selesai dengan sukses

**Behavior:**
- **Blocks** – method ini akan block sampai runtime selesai (loop exit atau error)
- **Idempotent** – memanggil `Start()` kedua kali akan return `ErrAlreadyRunning`

---

### 3.2 Step-by-Step Breakdown

#### Step 1: Cegah Double Start

```go
r.mu.Lock()
if r.running {
    r.mu.Unlock()
    return asentric.ErrAlreadyRunning
}
```

**Apa yang terjadi:**
- Lock mutex untuk protect state
- Cek apakah runtime sudah running
- Jika sudah running → return error, unlock mutex
- Jika belum running → lanjut ke step berikutnya

**Kenapa penting:**
- Mencegah race condition jika `Start()` dipanggil dari multiple goroutines
- Memastikan hanya ada satu instance runtime yang berjalan

---

#### Step 2: Buat Context dengan Cancel

```go
ctx, cancel := context.WithCancel(parentCtx)
r.ctx = ctx
r.cancel = cancel
r.running = true
r.mu.Unlock()
```

**Apa yang terjadi:**
- Buat context baru dengan cancel function dari `parentCtx`
- Simpan context dan cancel function ke struct
- Set `running = true` untuk menandai runtime aktif
- Unlock mutex

**Kenapa penting:**
- Context ini digunakan untuk:
  - Mengontrol lifecycle runtime (kapan harus stop)
  - Di-pass ke `EventSource.Start()` dan `Dispatcher.Dispatch()`
  - Di-check di `loop()` untuk graceful shutdown

---

#### Step 3: Start EventSource

```go
events, err := r.Source.Start(ctx)
if err != nil {
    cancel()
    r.mu.Lock()
    r.running = false
    r.mu.Unlock()
    return err
}
```

**Apa yang terjadi:**
- Panggil `EventSource.Start(ctx)` dengan context yang baru dibuat
- Jika error (misalnya WebSocket connection gagal):
  - Cancel context
  - Set `running = false`
  - Return error
- Jika sukses:
  - Dapat channel `events` yang berisi stream events
  - Lanjut ke step berikutnya

**Kenapa penting:**
- Ini adalah **titik pertama** di canonical flow SPEC:
  ```
  [RPC WebSocket] → EventSource.Start() → <-chan Event
  ```
- Error handling di sini sesuai SPEC: **WebSocket error → exit** (lihat SPEC §12.2)

---

#### Step 4: Simpan Channel dan Start Loop

```go
r.events = events
return r.loop()
```

**Apa yang terjadi:**
- Simpan channel `events` ke struct
- Panggil `r.loop()` yang akan **block** sampai selesai
- Return error dari `loop()` (atau `nil` jika sukses)

**Kenapa penting:**
- `loop()` adalah **core event loop** yang membaca events dan memprosesnya
- Method `Start()` akan block di sini sampai runtime selesai

---

### 3.3 Diagram Flow Start()

```
Start(parentCtx)
    │
    ├─▶ Lock mutex
    │   ├─▶ Cek running?
    │   │   ├─▶ Ya → ErrAlreadyRunning
    │   │   └─▶ Tidak → Lanjut
    │   │
    │   ├─▶ Buat context.WithCancel(parentCtx)
    │   ├─▶ Set running = true
    │   └─▶ Unlock mutex
    │
    ├─▶ Source.Start(ctx)
    │   ├─▶ Error? → Cancel context, set running=false, return error
    │   └─▶ Sukses? → Dapat <-chan Event
    │
    ├─▶ Simpan events ke r.events
    └─▶ return loop()  ← BLOCK DI SINI
```

---

## 4. loop() – Event Loop Canonical

### 4.1 Signature

```go
func (r *Runtime) loop() error
```

**Input:** Tidak ada (menggunakan field dari struct)

**Output:**
- `error` – error jika terjadi masalah, atau `nil` jika exit dengan sukses

**Behavior:**
- **Blocks** – akan block sampai channel ditutup atau context di-cancel
- **Single-threaded** – hanya satu goroutine yang menjalankan loop ini
- **Sequential** – events diproses satu per satu secara berurutan

---

### 4.2 Step-by-Step Breakdown

#### Step 1: Setup Defer untuk Cleanup

```go
defer func() {
    r.mu.Lock()
    r.running = false
    r.mu.Unlock()
}()
```

**Apa yang terjadi:**
- Defer function yang akan dieksekusi saat `loop()` selesai (exit atau panic)
- Set `running = false` untuk menandai runtime sudah tidak aktif lagi
- Lock mutex untuk thread-safety

**Kenapa penting:**
- Memastikan state `running` selalu konsisten, bahkan jika loop exit karena panic
- Memungkinkan `Start()` dipanggil lagi setelah loop selesai

---

#### Step 2: Infinite Loop dengan Select

```go
for {
    select {
    case <-r.ctx.Done():
        return nil
    case event, ok := <-r.events:
        if !ok {
            return nil
        }
        if r.Dispatcher == nil {
            return asentric.ErrNoDispatcher
        }
        if err := r.Dispatcher.Dispatch(r.ctx, event); err != nil {
            return err
        }
    }
}
```

**Apa yang terjadi:**

1. **Infinite loop** – akan terus berjalan sampai ada kondisi yang membuatnya exit
2. **Select statement** – menunggu salah satu dari dua channel:
   - `r.ctx.Done()` – context di-cancel (graceful shutdown)
   - `r.events` – event baru masuk dari EventSource

---

#### Case 1: Context Cancelled (`<-r.ctx.Done()`)

```go
case <-r.ctx.Done():
    return nil
```

**Apa yang terjadi:**
- Context di-cancel (misalnya karena `Stop()` dipanggil atau SIGINT/SIGTERM)
- Exit loop dengan return `nil` (sukses, bukan error)

**Kenapa penting:**
- Ini adalah **graceful shutdown** – runtime berhenti dengan terkontrol
- Tidak ada event yang "terpotong" di tengah proses (karena kita cek context sebelum memproses event)

---

#### Case 2: Event Masuk (`event, ok := <-r.events`)

```go
case event, ok := <-r.events:
    if !ok {
        return nil
    }
    if r.Dispatcher == nil {
        return asentric.ErrNoDispatcher
    }
    if err := r.Dispatcher.Dispatch(r.ctx, event); err != nil {
        return err
    }
```

**Apa yang terjadi:**

1. **Receive event dari channel**
   - `event` – event yang diterima
   - `ok` – `false` jika channel sudah ditutup

2. **Cek channel closed**
   - Jika `!ok` → channel ditutup oleh EventSource (misalnya karena WebSocket disconnect)
   - Return `nil` (exit dengan sukses, bukan error)

3. **Cek Dispatcher**
   - Jika `Dispatcher == nil` → return `ErrNoDispatcher`
   - Ini adalah **safety check** untuk memastikan dispatcher sudah di-set

4. **Dispatch event**
   - Panggil `Dispatcher.Dispatch(r.ctx, event)`
   - Jika error → return error (ini akan menghentikan runtime)

**Kenapa penting:**
- Ini adalah **core canonical flow** dari SPEC:
  ```
  Event → Dispatcher.Dispatch() → Context → Engine → Alert
  ```
- Error handling di sini sesuai SPEC: **Dispatcher error → exit** (lihat SPEC §12.2)

---

### 4.3 Diagram Flow loop()

```
loop()
    │
    ├─▶ defer: Set running = false
    │
    └─▶ for {
            select {
                │
                ├─▶ case <-ctx.Done():
                │       └─▶ return nil  (graceful shutdown)
                │
                └─▶ case event, ok := <-events:
                        │
                        ├─▶ if !ok:
                        │       └─▶ return nil  (channel closed)
                        │
                        ├─▶ if Dispatcher == nil:
                        │       └─▶ return ErrNoDispatcher
                        │
                        └─▶ Dispatcher.Dispatch(ctx, event)
                                │
                                ├─▶ Error? → return error
                                └─▶ Sukses? → Lanjut loop
        }
```

---

### 4.4 Single-Threaded Guarantee

**Penting:** `loop()` adalah **single-threaded**:

- ✅ Hanya **satu goroutine** yang menjalankan loop ini
- ✅ Events diproses **secara sequential** (satu per satu)
- ✅ Tidak ada **concurrent processing** di dalam loop

**Kenapa single-threaded?**

1. **Determinism** – memastikan events diproses dalam urutan yang sama
2. **Engine constraint** – Engine tidak concurrency-safe (lihat SPEC §4.2)
3. **Simplicity** – lebih mudah di-debug dan di-reason

**Jika butuh parallelism:**
- Parallelism dilakukan di **level di atas** runtime (misalnya di `cmd/runtime-reference/pipeline/`)
- Bisa membuat **multiple Runtime instances** dengan worker pool
- Tapi **satu Runtime instance** tetap single-threaded

---

## 5. Stop() – Graceful Shutdown

### 5.1 Signature

```go
func (r *Runtime) Stop() error
```

**Input:** Tidak ada

**Output:**
- `error` – selalu `nil` (idempotent, tidak pernah error)

**Behavior:**
- **Idempotent** – bisa dipanggil berkali-kali tanpa efek samping
- **Non-blocking** – tidak menunggu loop selesai, hanya trigger cancellation
- **Does not close channels** – channel ditutup oleh EventSource, bukan Runtime

---

### 5.2 Step-by-Step Breakdown

#### Step 1: Lock dan Cek State

```go
r.mu.Lock()
defer r.mu.Unlock()

if !r.running {
    return nil
}
```

**Apa yang terjadi:**
- Lock mutex untuk protect state
- Cek apakah runtime sedang running
- Jika tidak running → return `nil` (idempotent, tidak ada yang perlu dilakukan)
- Jika running → lanjut ke step berikutnya

---

#### Step 2: Cancel Context

```go
if r.cancel != nil {
    r.cancel()
}
r.running = false
return nil
```

**Apa yang terjadi:**
- Cancel context jika cancel function ada
- Set `running = false`
- Return `nil`

**Kenapa penting:**
- `cancel()` akan trigger `r.ctx.Done()` di `loop()`
- `loop()` akan detect ini dan exit gracefully
- Event yang sedang diproses akan selesai, tapi event baru tidak akan diproses

---

### 5.3 Diagram Flow Stop()

```
Stop()
    │
    ├─▶ Lock mutex
    │   ├─▶ Cek running?
    │   │   ├─▶ Tidak → return nil (idempotent)
    │   │   └─▶ Ya → Lanjut
    │   │
    │   ├─▶ cancel()  ← Trigger ctx.Done()
    │   ├─▶ Set running = false
    │   └─▶ Unlock mutex
    │
    └─▶ return nil
```

**Catatan:** `Stop()` **tidak menunggu** loop selesai. Jika Anda perlu menunggu, gunakan pattern seperti ini:

```go
runtime.Stop()
// Tunggu sampai loop selesai (misalnya dengan WaitGroup atau channel)
```

---

## 6. Mapping ke Canonical Flow SPEC

### 6.1 Canonical Flow dari SPEC §3

```
[RPC WebSocket]
    │
    │ eth_subscribe("logs", {addresses})
    ▼
[EventSource]  ← Start() di runtime
    │
    │ Event
    ▼
[Redis Queue]  ← (Opsional, bisa langsung ke Runtime)
    │
    │ Event
    ▼
[Runtime]  ← loop() di internal/runtime
    │
    │ Dispatch(event)
    ▼
[Dispatcher]  ← Dispatch() di internal/dispatcher
    │
    │ Context → Engine.Evaluate()
    ▼
[Engine]  ← Evaluate() di pkg/asentric
    │
    │ []*Alert
    ▼
[Redis Queue]  ← (Opsional, bisa langsung ke Sink)
    │
    │ Alert
    ▼
[AlertSink]  ← Emit() di internal/sink
    │
    │ POST JSON
    ▼
[Webhook]
```

---

### 6.2 Bagaimana `internal/runtime` Memetakan ke Flow Ini

| Step di SPEC | Komponen di Runtime | Kode |
|--------------|---------------------|------|
| **EventSource** | `r.Source.Start(ctx)` | `Start()` line 57 |
| **Event masuk** | `<-r.events` | `loop()` line 86 |
| **Runtime** | `r.loop()` | `loop()` seluruh function |
| **Dispatcher** | `r.Dispatcher.Dispatch()` | `loop()` line 95 |
| **Context → Engine** | Di dalam `Dispatcher.Dispatch()` | `internal/dispatcher` |
| **Alert → Sink** | Di dalam `Dispatcher.Dispatch()` | `internal/dispatcher` |

**Penting:** `internal/runtime` hanya mengorkestrasi **EventSource → Dispatcher**. Detail Context building, Engine evaluation, dan Alert delivery ada di `internal/dispatcher`.

---

## 7. Contoh Konkret dengan Use Case

### 7.1 Setup Runtime untuk Large Transfer Detection

```go
// 1. Buat EventSource (contoh: WebSocket)
source := &websocket.WebSocketSource{
    RPCURL: "wss://rpc.mantle.xyz/ws",
    Addresses: []string{"0xVaultAddress..."},
}

// 2. Buat Dispatcher dengan Engine
engine := asentric.NewEngine()
engine.RegisterRule(&rules.LargeTransferRule{
    Threshold: big.NewInt(1e18), // 1 ETH
})

dispatcher := &dispatcher.EngineDispatcher{
    Engine: engine,
    Sink: &webhook.WebhookSink{
        URL: "https://your-webhook.com/alerts",
    },
    ContextBuilder: &context.EventContextBuilder{},
    ABIRegistry: abiRegistry,
}

// 3. Buat Runtime
runtime := runtime.NewRuntime(source, dispatcher)

// 4. Start runtime (block sampai selesai)
ctx := context.Background()
if err := runtime.Start(ctx); err != nil {
    log.Fatal(err)
}
```

---

### 7.2 Flow Saat Event Masuk (Large Transfer)

**Timeline:**

1. **T0: Event masuk ke channel**
   ```
   EventSource (WebSocket) menerima log dari chain
   → Parse log → asentric.Event
   → Kirim ke channel: events <- event
   ```

2. **T1: Runtime loop menerima event**
   ```go
   // Di loop()
   case event, ok := <-r.events:
       // event = {
       //   ChainID: 5000,
       //   BlockNumber: 12345678,
       //   TxHash: "0xabc...",
       //   Payload: <raw log data>
       // }
   ```

3. **T2: Dispatcher memproses event**
   ```go
   // Di Dispatcher.Dispatch()
   // 1. Build Context dari Event
   execCtx := d.ContextBuilder.Build(event)
   // execCtx.Tx() = domain.Transaction dengan Value() = 2 ETH
   
   // 2. Evaluate rules
   alerts, err := d.Engine.Evaluate(execCtx)
   // LargeTransferRule.Evaluate() terdeteksi!
   // alerts = [Alert{Severity: High, Title: "Large Transfer", ...}]
   
   // 3. Emit alerts
   d.Sink.Emit(ctx, alerts[0])
   // HTTP POST ke webhook dengan JSON alert
   ```

4. **T3: Loop lanjut ke event berikutnya**
   ```go
   // Kembali ke select, tunggu event berikutnya
   ```

---

### 7.3 Flow Saat Proxy Upgrade Terdeteksi

**Skenario:** Event `Upgraded(address,address)` dari proxy contract

1. **Event masuk**
   ```
   EventSource menerima log dengan topic = Upgraded
   → Parse → asentric.Event dengan Payload = log data
   ```

2. **Context dibangun**
   ```go
   // Di ContextBuilder.Build()
   // Decode log menggunakan ABI
   // execCtx.Logs() = [Log{
   //   Event: Event{
   //     Name: "Upgraded",
   //     Fields: {
   //       "oldImplementation": "0x...",
   //       "newImplementation": "0x...",
   //     }
   //   }
   // }]
   ```

3. **Rule evaluate**
   ```go
   // ProxyUpgradeRule.Evaluate(execCtx)
   for _, log := range execCtx.Logs() {
       if log.Event.Name == "Upgraded" {
           return Alert{
               Severity: Critical,
               Title: "Proxy Upgrade Detected",
               Metadata: {
                   "old": log.Event.Fields["oldImplementation"],
                   "new": log.Event.Fields["newImplementation"],
               }
           }, nil
       }
   }
   ```

4. **Alert dikirim**
   ```go
   // Sink.Emit() → POST ke webhook
   ```

---

## 8. Invariants & Guarantees

### 8.1 Runtime Invariants (WAJIB DIPATUHI)

| Invariant | Penjelasan | Dijamin oleh |
|-----------|------------|--------------|
| **Single-threaded loop** | Hanya satu goroutine yang menjalankan `loop()` | Desain `Start()` hanya memanggil `loop()` sekali |
| **Sequential processing** | Events diproses satu per satu | `select` di `loop()` hanya memproses satu event per iterasi |
| **Idempotent Start** | `Start()` tidak bisa dipanggil dua kali | Mutex + `running` check |
| **Idempotent Stop** | `Stop()` bisa dipanggil berkali-kali | Mutex + `running` check |
| **Context-driven lifecycle** | Runtime berhenti saat context di-cancel | `case <-r.ctx.Done()` di `loop()` |
| **No channel closing** | Runtime tidak pernah menutup channel | Channel ditutup oleh EventSource |

---

### 8.2 Behavioral Guarantees

| Guarantee | Penjelasan |
|-----------|------------|
| **Deterministic order** | Events diproses dalam urutan yang sama seperti diterima | Single-threaded loop |
| **Graceful shutdown** | Runtime menyelesaikan event yang sedang diproses sebelum exit | Context cancellation di-check sebelum memproses event baru |
| **Error propagation** | Error dari Dispatcher akan menghentikan runtime | `if err := r.Dispatcher.Dispatch(...); err != nil { return err }` |
| **State consistency** | `running` selalu konsisten, bahkan saat panic | Defer di `loop()` |

---

## 9. Error Semantics

### 9.1 Error dari EventSource.Start()

**Kapan terjadi:**
- WebSocket connection gagal
- RPC endpoint tidak tersedia
- Invalid configuration

**Behavior:**
```go
events, err := r.Source.Start(ctx)
if err != nil {
    cancel()
    r.mu.Lock()
    r.running = false
    r.mu.Unlock()
    return err  // ← Runtime exit dengan error
}
```

**Sesuai SPEC:** ✅ **WebSocket error → exit** (SPEC §12.2)

---

### 9.2 Error dari Dispatcher.Dispatch()

**Kapan terjadi:**
- Context building gagal
- Engine evaluation error (rule panic)
- AlertSink error (webhook POST gagal)

**Behavior:**
```go
if err := r.Dispatcher.Dispatch(r.ctx, event); err != nil {
    return err  // ← Runtime exit dengan error
}
```

**Sesuai SPEC:**
- ✅ **Rule error** → propagate error (SPEC §12.2)
- ⚠️ **Webhook error** → seharusnya log & continue (SPEC §12.2), tapi di runtime ini akan exit
  - **Catatan:** Error handling webhook biasanya dilakukan di `AlertSink.Emit()` sendiri, bukan di Dispatcher

---

### 9.3 Channel Closed (`!ok`)

**Kapan terjadi:**
- EventSource menutup channel (misalnya karena WebSocket disconnect)
- Context di-cancel di EventSource

**Behavior:**
```go
case event, ok := <-r.events:
    if !ok {
        return nil  // ← Exit dengan sukses (bukan error)
    }
```

**Sesuai SPEC:** ✅ Channel closed biasanya berarti EventSource sengaja berhenti, bukan error

---

### 9.4 Context Cancelled

**Kapan terjadi:**
- `Stop()` dipanggil
- Parent context di-cancel
- SIGINT/SIGTERM diterima (jika di-handle di parent context)

**Behavior:**
```go
case <-r.ctx.Done():
    return nil  // ← Graceful shutdown, exit dengan sukses
```

**Sesuai SPEC:** ✅ **Graceful shutdown** (SPEC §12.3)

---

## 10. Concurrency Model

### 10.1 Single-Threaded Runtime

**Runtime itu sendiri adalah single-threaded:**

- ✅ Hanya **satu goroutine** yang menjalankan `loop()`
- ✅ Events diproses **secara sequential**
- ✅ Tidak ada **race condition** di dalam Runtime

**Tapi komponen di bawahnya bisa concurrent:**

- ✅ `EventSource.Start()` bisa spawn goroutine untuk WebSocket
- ✅ `Dispatcher.Dispatch()` bisa spawn goroutine untuk async processing (tapi harus hati-hati dengan Engine)
- ✅ `AlertSink.Emit()` bisa spawn goroutine untuk HTTP POST (tapi harus handle error dengan benar)

---

### 10.2 Thread-Safety

**Runtime struct fields:**

| Field | Thread-Safe? | Protection |
|-------|--------------|------------|
| `Source` | ✅ Read-only setelah `NewRuntime()` | Tidak perlu protection |
| `Dispatcher` | ✅ Read-only setelah `NewRuntime()` | Tidak perlu protection |
| `ctx` | ✅ Read-only setelah `Start()` | Tidak perlu protection |
| `cancel` | ⚠️ Bisa di-call dari multiple goroutines | Tidak ada protection (tapi `cancel()` idempotent) |
| `events` | ✅ Channel adalah thread-safe | Go channel semantics |
| `running` | ✅ Protected by mutex | `mu sync.Mutex` |

**Penting:** `running` **harus** di-protect dengan mutex karena bisa di-akses dari:
- `Start()` (set `running = true`)
- `Stop()` (set `running = false`)
- `loop()` defer (set `running = false`)

---

### 10.3 Multiple Runtime Instances

**Bisa membuat multiple Runtime instances untuk parallelism:**

```go
// Worker pool pattern
for i := 0; i < numWorkers; i++ {
    runtime := runtime.NewRuntime(source, dispatcher)
    go runtime.Start(ctx)  // Setiap worker punya Runtime sendiri
}
```

**Tapi perhatikan:**

- ✅ Setiap Runtime instance punya **Engine sendiri** (Engine tidak concurrency-safe)
- ✅ Setiap Runtime instance punya **Dispatcher sendiri**
- ⚠️ Harus hati-hati dengan **shared state** (misalnya Redis queue, state store)

---

## Summary

`internal/runtime` adalah **core event loop** yang:

1. ✅ Mengorkestrasi **EventSource → Dispatcher** sesuai canonical flow SPEC
2. ✅ Menjamin **single-threaded, sequential processing**
3. ✅ Mengelola **graceful shutdown** via context cancellation
4. ✅ Memastikan **deterministic order** dan **state consistency**

**Sebagai Lead, Anda perlu memastikan:**

- ✅ Tim memahami bahwa Runtime adalah **single-threaded**
- ✅ Tim tidak mencoba membuat Runtime concurrent (gunakan worker pool di level di atas)
- ✅ Tim memahami **error semantics** dan kapan runtime exit
- ✅ Tim memahami **lifecycle** (Start → loop → Stop)

---

## Related Documentation

- **[SPEC.md](SPEC.md)** - MVP specification (canonical flow di §3)
- **[architecture.md](architecture.md)** - Core architecture (lifecycle di §7)
- **[sdk-api.md](sdk-api.md)** - Public API contracts (EventSource, Dispatcher interfaces)
- **[runtime-reference-checklist.md](runtime-reference-checklist.md)** - Implementation checklist

---

**Last Updated:** 2024  
**Version:** 1.0




