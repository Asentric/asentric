# Architecture – Asentric SDK

## 1. Purpose & Scope

Asentric SDK adalah **pure security detection engine** untuk sistem blockchain.

Tanggung jawab tunggalnya adalah:

* Mengeksekusi security rules yang deterministik
* Memproses data on-chain yang terstruktur
* Menghasilkan alert yang terstruktur

SDK secara sengaja **tidak** melakukan:

* Mengambil data blockchain
* Menjalankan watcher yang berjalan lama
* Terhubung ke Redis, Kafka, database, atau HTTP API
* Mengirimkan alert
* Mengelola konfigurasi, deployment, atau infrastruktur

Concern ini didelegasikan ke sistem eksternal (misalnya `asentric-bot`, `asentric-backend`).

SDK dirancang untuk **di-embed**, bukan di-deploy.

---

## 2. Architectural Philosophy

### 2.1 Pure Domain Core

Pada intinya, Asentric SDK mengikuti filosofi **pure domain logic**:

* Rules adalah pure functions
* Tidak ada side effects selama evaluasi rule
* Tidak ada global state
* Tidak ada I/O di dalam engine

Ini menjamin:

* Eksekusi yang deterministik
* Mudah di-test
* Replay yang aman
* Perilaku yang dapat diprediksi

> **Note:** Secara konseptual, SDK berperilaku seperti pure function: dengan Context dan rule set yang sama, selalu menghasilkan alert yang sama. Sifat matematis ini memungkinkan replay dan testing yang deterministik.

---

### 2.2 Infrastructure Inversion

SDK tidak bergantung pada infrastruktur.

Sebaliknya:

* Infrastruktur bergantung pada SDK
* Sistem runtime menyediakan data **ke dalam** SDK
* SDK menghasilkan hasil **ke luar**

Ini adalah dependensi satu arah yang ketat.

```
Infrastructure → SDK → Alerts
```

Inversi ini mencegah concern infrastruktur merembes ke logika keamanan.

---

### 2.3 Explicit Context Boundary

Semua data eksekusi mengalir melalui satu objek: **Context**.

Context adalah:

* Eksplisit
* Immutable selama evaluasi
* Sepenuhnya dikendalikan oleh runtime

Tidak ada:

* Hidden state
* Global variables
* Implicit dependencies

Ini membuat eksekusi dapat ditelusuri, dapat di-debug, dan dapat di-replay.

---

## 3. High-Level Architecture

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

* **Bot**: Chain monitoring, transaction ingestion, alert routing
* **Backend**: REST API, data aggregation, persistence
* **Frontend**: User interface, dashboards, visualization

SDK adalah **closed execution box** yang memproses Context dan menghasilkan Alerts.

---

## 4. Core Components

### 4.1 Engine

Engine bertanggung jawab untuk:

* Mengelola registrasi rule
* Iterasi atas rules
* Mengeksekusi rules terhadap Context yang diberikan
* Mengumpulkan alert yang dihasilkan

Engine:

* Tidak memiliki pengetahuan tentang chain, RPC, atau infrastruktur
* Tidak mengelola concurrency atau scheduling
* Mengeksekusi rules secara sequential dan deterministik

Secara konseptual:

```
for each rule:
  alert = rule.Evaluate(ctx)
  collect(alert)
```

---

### 4.2 Rule System

Rules merepresentasikan **security knowledge**.

Sebuah rule:

* Mengenkapsulasi satu ide deteksi
* Stateless
* Deterministik

Rules:

* Tidak pernah melakukan I/O
* Tidak pernah memutasi state eksternal
* Tidak pernah berkomunikasi dengan rules lain

Rules adalah **isolated units of logic**.

Isolasi ini menjamin:

* Keamanan
* Parallel reasoning
* Testing yang sederhana

---

### 4.3 Context

Context adalah **single source of truth** selama eksekusi.

Berisi:

* Data transaksi
* Metadata block
* Log / event yang sudah di-decode
* Informasi spesifik chain

Context:

* Dibangun oleh runtime
* Diteruskan ke SDK
* Read-only selama evaluasi

> **Note:** SDK mengasumsikan Context valid dan lengkap. Validasi dan enrichment Context (misalnya, mengambil data yang hilang, decoding event) adalah tanggung jawab runtime. SDK tidak memvalidasi atau memperkaya Context.

---

### 4.4 Alerts

Alert adalah **satu-satunya output** dari SDK.

Sebuah alert adalah:

* Terstruktur
* Dapat di-serialize
* Infrastructure-agnostic

Alert tidak:

* Tahu ke mana akan dikirim
* Tahu bagaimana akan disimpan
* Berisi logika delivery

SDK menghasilkan alert — SDK tidak mengirimkannya.

#### Execution Reference

Alert dapat menyertakan `ExecutionRef` opsional yang berisi:

* Transaction hash (`tx_hash`)
* Block number (`block_number`)

**Distingsi penting:**

* **Alert ≠ delivery envelope** — Alert adalah sinyal semantik, bukan pesan infrastruktur
* **ExecutionRef ≠ infrastructure metadata** — Hanya berisi traceability eksekusi, bukan chain identity, network, atau RPC endpoints
* **Chain identity tetap eksternal** — Sistem runtime bertanggung jawab untuk chain context (chain ID, nama network, dll.)

`ExecutionRef` adalah:

* Diisi oleh engine (bukan oleh rules)
* Tidak dapat diakses atau dimodifikasi oleh rules
* Opsional dan hanya informasional
* Dapat di-override atau diperkaya oleh sistem runtime

Desain ini mempertahankan:

* **Rule purity** — Rules tetap tidak menyadari execution context
* **Alert semantics** — Alert tetap fokus pada sinyal keamanan
* **Runtime ownership** — Runtime mengontrol semua infrastruktur dan chain context

---

## 5. Package Structure & Boundaries

### 5.1 Public API (`pkg/asentric`)

Ini adalah **satu-satunya surface integrasi yang didukung**.

Berisi:

* Engine
* Rule interface
* Context interface
* Alert model

Jaminan stabilitas:

* Backward compatibility
* Perubahan dikelola dengan semver

Jika berada di `pkg/asentric`, aman untuk dijadikan dependensi.

---

### 5.2 Internal Implementation (`internal/`)

Semua yang ada di bawah `internal/` adalah:

* Private
* Non-stable
* Bebas untuk diubah

Berisi:

* Internals eksekusi rule
* Runtime helpers
* Logika decoding ABI
* Observability internal

Sistem eksternal tidak boleh mengimpor dari `internal/`.

---

### 5.3 CLI (`cmd/asentric`)

CLI adalah **developer tool**, bukan runtime.

Digunakan untuk:

* Project scaffolding
* Testing rule
* Offline replay

CLI:

* Tidak terhubung ke RPC nodes
* Tidak menjalankan production watchers
* Tidak mengelola proses yang berjalan lama

CLI ada semata-mata untuk meningkatkan developer experience.

---

## 6. Runtime Responsibility Matrix

| Responsibility   | SDK          | Runtime |
| ---------------- | ------------ | ------- |
| Fetch chain data | ❌            | ✅       |
| Decode ABI       | ⚠️ (internal helpers only) | ✅ (full decoding) |
| Rule execution   | ✅            | ❌       |
| Alert creation   | ✅            | ❌       |
| Alert delivery   | ❌            | ✅       |
| Persistence      | ❌            | ✅       |
| Scheduling       | ❌            | ✅       |
| Scaling          | ❌            | ✅       |

> **Note:** SDK menyediakan ABI decoding helpers di `internal/abi/` untuk penggunaan internal, tetapi decoding ABI penuh dan parsing event adalah tanggung jawab runtime. SDK menggunakan helpers ini secara internal tetapi tidak mengeksposnya sebagai bagian dari public API.

SDK tidak pernah melintasi batasnya.

---

## 7. Determinism & Replay

Determinisme adalah invariant inti.

Dengan:

* Context yang sama
* Rule set yang sama

SDK menjamin:

* Output yang identik
* Tidak ada hidden state
* Tidak ada perilaku berbasis waktu

Replay bekerja dengan:

* Merekonstruksi Context dari fixtures
* Menjalankan engine secara offline

SDK **tidak pernah mengambil data historis**.

---

## 8. Observability Scope

Observability di dalam SDK adalah **strictly internal**.

Termasuk:

* Rule execution timing
* Rule evaluation counts
* Engine diagnostics

Tidak termasuk:

* Metrics exporters
* Tracing backends
* Logging pipelines

Mengekspor data observability adalah tanggung jawab runtime.

---

## 9. Ecosystem Integration

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

SDK tetap menjadi **single source of truth** untuk security detection logic.

---

## 10. Architectural Non-Goals

Berikut ini **tidak akan pernah** ditambahkan ke SDK:

* RPC clients
* Redis / Kafka clients
* HTTP servers
* Databases
* Alert delivery
* Deployment tooling

Jika sebuah fitur memerlukan infrastruktur, fitur tersebut tidak termasuk di sini.

---

## 11. Summary

Asentric SDK adalah:

* Pure execution engine
* Infrastructure-agnostic
* Deterministic dan dapat di-test
* Dirancang untuk embedding

SDK ada untuk membuat **security logic menjadi sederhana, aman, dan dapat digunakan kembali**.

Segala sesuatu yang lain berada di tempat lain.

