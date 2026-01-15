# Project Structure

This document defines the final folder and file structure for the Asentric SDK.

---

## Repository Layout

```
asentric-sdk/
├── pkg/                       # Public SDK API
│   ├── asentric/              # Core SDK
│   ├── domain/                # Domain types
│   ├── runtime/               # Runtime builder
│   └── utils/                 # Utility functions
├── internal/                  # Private implementation
├── cmd/                       # CLI tools
├── templates/                 # Project scaffolding
├── examples/                  # Usage examples
├── docs/                      # Documentation
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## Public SDK - `pkg/`

The public API that external developers depend on.

### Core SDK (`pkg/asentric/`)

```
pkg/asentric/
├── engine.go        # Engine struct and execution
├── rule.go          # Rule interface
├── context.go       # Context interface
├── alert.go         # Alert and Severity models
├── event.go         # Event model
├── config.go        # Configuration loading
├── errors.go        # Error types
├── event_source.go  # EventSource interface
├── alert_sink.go    # AlertSink interface
└── dispatcher.go    # Dispatcher interface
```

### Domain Types (`pkg/domain/`)

```
pkg/domain/
├── address.go       # Address type
├── hash.go          # Hash type
├── chain.go         # ChainID, Chain
├── transaction.go   # Transaction struct
├── block.go         # Block struct
├── log.go           # Log struct
├── event.go         # Decoded event
├── value.go         # NativeValue, TokenAmount
└── token.go         # Token metadata
```

### Runtime (`pkg/runtime/`)

```
pkg/runtime/
├── builder.go       # Runtime builder pattern
└── runtime.go       # Runtime lifecycle
```

### Utilities (`pkg/utils/`)

```
pkg/utils/
├── address.go       # Address utilities
├── format.go        # Formatting helpers
└── field.go         # Event field extraction
```

**Design Principles:**

- No external infrastructure (Redis, databases)
- No network I/O
- No goroutines in public API
- Pure execution only
- String-based types for ergonomics

---

## Internal Implementation - `internal/`

Private implementation details that may change without notice.

```
internal/
├── chain/                  # Chain interaction
│   ├── types.go            # Raw types
│   └── client.go           # Chain client
├── adapter/                # Type conversion
│   └── converter.go        # Chain to domain conversion
├── source/                 # Event sources
│   ├── websocket.go        # WebSocket source
│   └── memory.go           # In-memory source
├── sink/                   # Alert sinks
│   ├── console.go          # Console output
│   ├── webhook.go          # Webhook delivery
│   └── multi.go            # Multiple sinks
├── context/                # Context implementations
│   └── evm_context.go      # EVM context
├── abi/                    # ABI handling
│   ├── loader.go           # ABI file loading
│   └── decoder.go          # Event decoding
└── dispatcher/             # Event dispatching
    └── dispatcher.go       # Dispatcher implementation
```

---

## CLI Tools - `cmd/`

```
cmd/asentric/
├── main.go
└── cmd/
    ├── root.go          # Root command
    ├── init.go          # Project scaffolding
    ├── version.go       # Version info
    └── templates/       # Embedded templates
```

**Purpose:** Developer tools for project initialization and testing. Not a runtime.

---

## Templates - `templates/`

Used by `asentric init` to scaffold new projects.

```
templates/project/
├── config/
│   ├── asentric.yaml      # Runtime configuration
│   └── registry.yaml      # Target contracts
├── rules/
│   └── example_rule.go    # Example detection rule
├── abi/
│   └── .gitkeep
├── cmd/
│   └── watcher/
│       └── main.go        # Entry point
├── go.mod.tmpl
└── README.md.tmpl
```

---

## Examples - `examples/`

Reference implementations for common use cases.

```
examples/
├── simple-watcher/        # Basic watcher example
├── webhook-integration/   # Webhook backend integration
└── custom-rules/          # Advanced rule examples
```

---

## Documentation - `docs/`

```
docs/
├── QUICK-START.md         # 5-minute quick start
├── developer-overview.md  # Complete developer guide
├── architecture.md        # System architecture
├── sdk-api.md             # API reference
├── project-structure.md   # This file
└── SPEC.md                # MVP specification
```

---

## Layer Boundaries

| Layer | Infrastructure | State | Deterministic |
|-------|----------------|-------|---------------|
| Engine | None | None | Yes |
| Rules | None | None | Yes |
| Runtime | WebSocket, HTTP | Ephemeral | No |

**Principles:**

- **Engine and Rules:** Pure, deterministic, no infrastructure dependencies
- **Runtime:** Infrastructure-aware, handles connections and delivery
- **Public API:** Stable, backward compatible

---

## Related Documentation

- [Quick Start](QUICK-START.md)
- [Developer Overview](developer-overview.md)
- [Architecture](architecture.md)
- [SDK API Reference](sdk-api.md)
