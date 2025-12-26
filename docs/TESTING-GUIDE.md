# Asentric SDK – Testing Guide

> **Panduan lengkap untuk testing di Asentric SDK**
>
> **Audience:** Semua developer
>
> **Prerequisite:** Sudah baca [ONBOARDING.md](ONBOARDING.md)

---

## Daftar Isi

1. [Philosophy & Principles](#1-philosophy--principles)
2. [Test Categories](#2-test-categories)
3. [Testing Tools & Setup](#3-testing-tools--setup)
4. [Unit Testing](#4-unit-testing)
5. [Integration Testing](#5-integration-testing)
6. [Mock Patterns](#6-mock-patterns)
7. [Test Coverage](#7-test-coverage)
8. [Testing Checklist per Component](#8-testing-checklist-per-component)
9. [Common Testing Patterns](#9-common-testing-patterns)
10. [Troubleshooting](#10-troubleshooting)

---

## 1. Philosophy & Principles

### 1.1 Testing Goals

| Goal | Mengapa Penting |
|------|-----------------|
| **Correctness** | Memastikan code bekerja sesuai spec |
| **Determinism** | Same input → same output (terutama untuk Engine) |
| **Regression Prevention** | Mencegah bug yang sudah diperbaiki muncul lagi |
| **Documentation** | Tests adalah dokumentasi executable |
| **Confidence** | Berani refactor dengan tests sebagai safety net |

### 1.2 Core Principles

1. **Test behavior, bukan implementation**
   - ✅ Test: "Jika nilai > threshold, return alert"
   - ❌ Test: "Field X harus di-set ke Y"

2. **Isolate tests**
   - Setiap test harus bisa jalan sendiri
   - Tidak depend pada test lain
   - Tidak depend pada external state

3. **Fast tests**
   - Unit tests harus < 100ms per test
   - Integration tests boleh lebih lama, tapi pisahkan

4. **Readable tests**
   - Test name menjelaskan apa yang ditest
   - Use table-driven tests untuk multiple cases
   - Comments untuk explain complex scenarios

### 1.3 Test Naming Convention

```go
// Format: TestComponentName_Scenario_ExpectedBehavior

// Examples:
func TestEngine_NilContext_ReturnsError(t *testing.T)
func TestEngine_RulePanic_RecoversAndReturnsError(t *testing.T)
func TestLargeTransferRule_ValueExceedsThreshold_ReturnsAlert(t *testing.T)
func TestWebSocketSource_ConnectionFailed_ReturnsError(t *testing.T)
```

---

## 2. Test Categories

### 2.1 Unit Tests

| Characteristic | Description |
|----------------|-------------|
| **Scope** | Single function/method |
| **Speed** | Fast (< 100ms) |
| **Dependencies** | No external deps (mock everything) |
| **Location** | Same package as code |
| **File naming** | `*_test.go` |

**Example:**
```go
// pkg/asentric/engine_test.go
func TestEngine_Evaluate_EmptyRules_ReturnsEmptyAlerts(t *testing.T) {
    engine := NewEngine()
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    assert.NoError(t, err)
    assert.Empty(t, alerts)
}
```

### 2.2 Integration Tests

| Characteristic | Description |
|----------------|-------------|
| **Scope** | Multiple components together |
| **Speed** | Slower (seconds) |
| **Dependencies** | May use real deps (Redis, etc.) |
| **Location** | Separate `_integration_test.go` or `/test` folder |
| **Build tag** | `//go:build integration` |

**Example:**
```go
//go:build integration

// internal/runtime/runtime_integration_test.go
func TestRuntime_FullFlow_EventToAlert(t *testing.T) {
    // Setup real components
    source := &mockEventSource{}
    sink := &mockAlertSink{}
    // ... test full flow
}
```

### 2.3 Test Matrix

| Component | Unit Test | Integration Test |
|-----------|-----------|------------------|
| `pkg/asentric/engine.go` | ✅ Required | ❌ Optional |
| `pkg/asentric/rule.go` | ✅ Required | ❌ Not needed |
| `internal/runtime/` | ✅ Required | ✅ Required |
| `internal/dispatcher/` | ✅ Required | ✅ Required |
| `internal/source/` | ✅ Required | ✅ Required (dengan mock WS) |
| `internal/sink/` | ✅ Required | ✅ Required (dengan mock HTTP) |
| `internal/queue/` | ✅ Required | ✅ Required (dengan Docker Redis) |
| `internal/abi/` | ✅ Required | ❌ Optional |
| `internal/config/` | ✅ Required | ❌ Optional |

---

## 3. Testing Tools & Setup

### 3.1 Standard Library

Go standard library sudah cukup powerful untuk testing:

```go
import "testing"

func TestSomething(t *testing.T) {
    // t.Error() - fail but continue
    // t.Fatal() - fail and stop
    // t.Skip() - skip test
    // t.Parallel() - run in parallel
}
```

### 3.2 Recommended: testify

Untuk assertions yang lebih readable:

```bash
go get github.com/stretchr/testify
```

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWithTestify(t *testing.T) {
    // assert - fail but continue
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.Nil(t, obj)
    assert.NotNil(t, obj)
    assert.True(t, condition)
    
    // require - fail and stop immediately
    require.NoError(t, err)  // Stop if error
    require.NotNil(t, obj)   // Stop if nil
}
```

### 3.3 Table-Driven Tests

Pattern yang sangat umum di Go:

```go
func TestEngine_Evaluate(t *testing.T) {
    tests := []struct {
        name        string
        ctx         asentric.Context
        rules       []asentric.Rule
        wantAlerts  int
        wantErr     error
    }{
        {
            name:       "nil context returns error",
            ctx:        nil,
            rules:      nil,
            wantAlerts: 0,
            wantErr:    asentric.ErrInvalidContext,
        },
        {
            name:       "empty rules returns empty alerts",
            ctx:        &mockContext{},
            rules:      nil,
            wantAlerts: 0,
            wantErr:    nil,
        },
        {
            name:       "single rule triggers alert",
            ctx:        &mockContext{},
            rules:      []asentric.Rule{&alwaysAlertRule{}},
            wantAlerts: 1,
            wantErr:    nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := asentric.NewEngine()
            for _, rule := range tt.rules {
                engine.RegisterRule(rule)
            }
            
            alerts, err := engine.Evaluate(tt.ctx)
            
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
            }
            assert.Len(t, alerts, tt.wantAlerts)
        })
    }
}
```

### 3.4 Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./pkg/asentric/...

# Run specific test
go test -v -run TestEngine_Evaluate ./pkg/asentric/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## 4. Unit Testing

### 4.1 Testing Engine

**File:** `pkg/asentric/engine_test.go`

```go
package asentric_test

import (
    "testing"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// ========================================
// Mock Context
// ========================================

type mockContext struct {
    chainID domain.ChainID
    tx      domain.Transaction
    block   domain.Block
    logs    []domain.Log
}

func (m *mockContext) ChainID() domain.ChainID     { return m.chainID }
func (m *mockContext) Tx() domain.Transaction      { return m.tx }
func (m *mockContext) Block() domain.Block         { return m.block }
func (m *mockContext) Logs() []domain.Log          { return m.logs }
func (m *mockContext) ABI() domain.ABIRegistry     { return nil }

// ========================================
// Mock Rules
// ========================================

// alwaysAlertRule always returns an alert
type alwaysAlertRule struct {
    name string
}

func (r *alwaysAlertRule) Name() string { return r.name }
func (r *alwaysAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    return &asentric.Alert{
        Rule:     r.name,
        Severity: asentric.SeverityInfo,
        Title:    "Test Alert",
    }, nil
}

// neverAlertRule never returns an alert
type neverAlertRule struct {
    name string
}

func (r *neverAlertRule) Name() string { return r.name }
func (r *neverAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    return nil, nil
}

// panicRule panics during evaluation
type panicRule struct{}

func (r *panicRule) Name() string { return "panic_rule" }
func (r *panicRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    panic("intentional panic for testing")
}

// ========================================
// Tests
// ========================================

func TestEngine_NewEngine(t *testing.T) {
    engine := asentric.NewEngine()
    assert.NotNil(t, engine)
}

func TestEngine_RegisterRule_NilRule(t *testing.T) {
    engine := asentric.NewEngine()
    err := engine.RegisterRule(nil)
    assert.ErrorIs(t, err, asentric.ErrInvalidRule)
}

func TestEngine_RegisterRule_ValidRule(t *testing.T) {
    engine := asentric.NewEngine()
    rule := &alwaysAlertRule{name: "test"}
    err := engine.RegisterRule(rule)
    assert.NoError(t, err)
}

func TestEngine_Evaluate_NilContext(t *testing.T) {
    engine := asentric.NewEngine()
    alerts, err := engine.Evaluate(nil)
    
    assert.ErrorIs(t, err, asentric.ErrInvalidContext)
    assert.Nil(t, alerts)
}

func TestEngine_Evaluate_EmptyRules(t *testing.T) {
    engine := asentric.NewEngine()
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    assert.NoError(t, err)
    assert.Empty(t, alerts)
}

func TestEngine_Evaluate_SingleAlert(t *testing.T) {
    engine := asentric.NewEngine()
    engine.RegisterRule(&alwaysAlertRule{name: "test_rule"})
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    require.NoError(t, err)
    require.Len(t, alerts, 1)
    assert.Equal(t, "test_rule", alerts[0].Rule)
}

func TestEngine_Evaluate_MultipleRules_Order(t *testing.T) {
    engine := asentric.NewEngine()
    engine.RegisterRule(&alwaysAlertRule{name: "rule_1"})
    engine.RegisterRule(&alwaysAlertRule{name: "rule_2"})
    engine.RegisterRule(&alwaysAlertRule{name: "rule_3"})
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    require.NoError(t, err)
    require.Len(t, alerts, 3)
    
    // Verify order is preserved
    assert.Equal(t, "rule_1", alerts[0].Rule)
    assert.Equal(t, "rule_2", alerts[1].Rule)
    assert.Equal(t, "rule_3", alerts[2].Rule)
}

func TestEngine_Evaluate_MixedRules(t *testing.T) {
    engine := asentric.NewEngine()
    engine.RegisterRule(&alwaysAlertRule{name: "alert_rule"})
    engine.RegisterRule(&neverAlertRule{name: "no_alert"})
    engine.RegisterRule(&alwaysAlertRule{name: "another_alert"})
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    require.NoError(t, err)
    require.Len(t, alerts, 2)
    assert.Equal(t, "alert_rule", alerts[0].Rule)
    assert.Equal(t, "another_alert", alerts[1].Rule)
}

func TestEngine_Evaluate_RulePanic_Recovers(t *testing.T) {
    engine := asentric.NewEngine()
    engine.RegisterRule(&panicRule{})
    ctx := &mockContext{}
    
    alerts, err := engine.Evaluate(ctx)
    
    assert.ErrorIs(t, err, asentric.ErrRulePanic)
    assert.Nil(t, alerts)
}
```

### 4.2 Testing Rules

**File:** `examples/rules/large_transfer_test.go`

```go
package rules_test

import (
    "math/big"
    "testing"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Mock context yang bisa dikonfigurasi
type mockContext struct {
    tx domain.Transaction
}

func (m *mockContext) ChainID() domain.ChainID     { return 1 }
func (m *mockContext) Tx() domain.Transaction      { return m.tx }
func (m *mockContext) Block() domain.Block         { return domain.Block{} }
func (m *mockContext) Logs() []domain.Log          { return nil }
func (m *mockContext) ABI() domain.ABIRegistry     { return nil }

// Helper untuk create transaction dengan value
func txWithValue(weiStr string) domain.Transaction {
    return domain.Transaction{
        RawValue: domain.NativeValue{Wei: weiStr},
        From:     "0xSender",
        To:       "0xReceiver",
    }
}

func TestLargeTransferRule(t *testing.T) {
    oneEth := big.NewInt(1e18)
    
    tests := []struct {
        name           string
        threshold      *big.Int
        txValue        string  // wei as string
        wantAlert      bool
        wantSeverity   asentric.Severity
    }{
        {
            name:         "value exceeds threshold - should alert",
            threshold:    oneEth,
            txValue:      "2000000000000000000", // 2 ETH
            wantAlert:    true,
            wantSeverity: asentric.SeverityHigh,
        },
        {
            name:      "value equals threshold - no alert",
            threshold: oneEth,
            txValue:   "1000000000000000000", // 1 ETH
            wantAlert: false,
        },
        {
            name:      "value below threshold - no alert",
            threshold: oneEth,
            txValue:   "500000000000000000", // 0.5 ETH
            wantAlert: false,
        },
        {
            name:      "zero value - no alert",
            threshold: oneEth,
            txValue:   "0",
            wantAlert: false,
        },
        {
            name:         "large value - should alert",
            threshold:    oneEth,
            txValue:      "100000000000000000000", // 100 ETH
            wantAlert:    true,
            wantSeverity: asentric.SeverityHigh,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rule := NewLargeTransferRule(tt.threshold)
            ctx := &mockContext{tx: txWithValue(tt.txValue)}
            
            alert, err := rule.Evaluate(ctx)
            
            require.NoError(t, err)
            
            if tt.wantAlert {
                require.NotNil(t, alert, "expected alert")
                assert.Equal(t, tt.wantSeverity, alert.Severity)
                assert.Equal(t, "large_transfer_detection", alert.Rule)
            } else {
                assert.Nil(t, alert, "expected no alert")
            }
        })
    }
}

func TestLargeTransferRule_Name(t *testing.T) {
    rule := NewLargeTransferRule(big.NewInt(1))
    assert.Equal(t, "large_transfer_detection", rule.Name())
}
```

---

## 5. Integration Testing

### 5.1 Runtime Integration Test

**File:** `internal/runtime/runtime_integration_test.go`

```go
//go:build integration

package runtime_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/internal/runtime"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// ========================================
// Mock Components
// ========================================

// mockEventSource simulates EventSource
type mockEventSource struct {
    events chan asentric.Event
}

func newMockEventSource() *mockEventSource {
    return &mockEventSource{
        events: make(chan asentric.Event, 10),
    }
}

func (m *mockEventSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    return m.events, nil
}

func (m *mockEventSource) SendEvent(event asentric.Event) {
    m.events <- event
}

func (m *mockEventSource) Close() {
    close(m.events)
}

// mockAlertSink collects alerts
type mockAlertSink struct {
    alerts []*asentric.Alert
}

func (m *mockAlertSink) Emit(ctx context.Context, alert *asentric.Alert) error {
    m.alerts = append(m.alerts, alert)
    return nil
}

// mockDispatcher bridges event to engine
type mockDispatcher struct {
    engine *asentric.Engine
    sink   *mockAlertSink
}

func (d *mockDispatcher) Dispatch(ctx context.Context, event asentric.Event) error {
    // Simplified: create mock context and evaluate
    mockCtx := &mockContext{}
    alerts, err := d.engine.Evaluate(mockCtx)
    if err != nil {
        return err
    }
    for _, alert := range alerts {
        d.sink.Emit(ctx, alert)
    }
    return nil
}

type mockContext struct{}

func (m *mockContext) ChainID() domain.ChainID     { return 1 }
func (m *mockContext) Tx() domain.Transaction      { return domain.Transaction{} }
func (m *mockContext) Block() domain.Block         { return domain.Block{} }
func (m *mockContext) Logs() []domain.Log          { return nil }
func (m *mockContext) ABI() domain.ABIRegistry     { return nil }

// alwaysAlertRule for testing
type alwaysAlertRule struct{}

func (r *alwaysAlertRule) Name() string { return "test_rule" }
func (r *alwaysAlertRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    return &asentric.Alert{
        Rule:     r.Name(),
        Severity: asentric.SeverityInfo,
        Title:    "Test",
    }, nil
}

// ========================================
// Integration Tests
// ========================================

func TestRuntime_Integration_EventToAlert(t *testing.T) {
    // Setup
    source := newMockEventSource()
    sink := &mockAlertSink{}
    
    engine := asentric.NewEngine()
    engine.RegisterRule(&alwaysAlertRule{})
    
    dispatcher := &mockDispatcher{engine: engine, sink: sink}
    
    rt := runtime.NewRuntime(source, dispatcher)
    
    // Start runtime in goroutine
    ctx, cancel := context.WithCancel(context.Background())
    errCh := make(chan error, 1)
    
    go func() {
        errCh <- rt.Start(ctx)
    }()
    
    // Give time for runtime to start
    time.Sleep(50 * time.Millisecond)
    
    // Send event
    source.SendEvent(asentric.Event{
        ChainID:     1,
        BlockNumber: 12345,
        TxHash:      "0xtest",
    })
    
    // Wait for processing
    time.Sleep(100 * time.Millisecond)
    
    // Stop runtime
    cancel()
    
    // Wait for runtime to finish
    err := <-errCh
    require.NoError(t, err)
    
    // Verify alert was generated
    assert.Len(t, sink.alerts, 1)
    assert.Equal(t, "test_rule", sink.alerts[0].Rule)
}

func TestRuntime_Integration_GracefulShutdown(t *testing.T) {
    source := newMockEventSource()
    dispatcher := &mockDispatcher{
        engine: asentric.NewEngine(),
        sink:   &mockAlertSink{},
    }
    
    rt := runtime.NewRuntime(source, dispatcher)
    
    ctx, cancel := context.WithCancel(context.Background())
    errCh := make(chan error, 1)
    
    go func() {
        errCh <- rt.Start(ctx)
    }()
    
    // Give time to start
    time.Sleep(50 * time.Millisecond)
    
    // Cancel context (graceful shutdown)
    cancel()
    
    // Should exit without error
    select {
    case err := <-errCh:
        assert.NoError(t, err)
    case <-time.After(2 * time.Second):
        t.Fatal("timeout waiting for runtime to stop")
    }
}

func TestRuntime_Integration_MultipleEvents(t *testing.T) {
    source := newMockEventSource()
    sink := &mockAlertSink{}
    
    engine := asentric.NewEngine()
    engine.RegisterRule(&alwaysAlertRule{})
    
    dispatcher := &mockDispatcher{engine: engine, sink: sink}
    rt := runtime.NewRuntime(source, dispatcher)
    
    ctx, cancel := context.WithCancel(context.Background())
    errCh := make(chan error, 1)
    
    go func() {
        errCh <- rt.Start(ctx)
    }()
    
    time.Sleep(50 * time.Millisecond)
    
    // Send multiple events
    for i := 0; i < 5; i++ {
        source.SendEvent(asentric.Event{
            ChainID:     1,
            BlockNumber: uint64(i),
            TxHash:      fmt.Sprintf("0x%d", i),
        })
    }
    
    time.Sleep(200 * time.Millisecond)
    cancel()
    
    err := <-errCh
    require.NoError(t, err)
    
    // Should have 5 alerts
    assert.Len(t, sink.alerts, 5)
}
```

### 5.2 Running Integration Tests

```bash
# Run integration tests only
go test -v -tags=integration ./...

# Run with specific timeout
go test -v -tags=integration -timeout=30s ./...
```

---

## 6. Mock Patterns

### 6.1 Mock Context

```go
// Reusable mock context
type mockContext struct {
    chainID domain.ChainID
    tx      domain.Transaction
    block   domain.Block
    logs    []domain.Log
    abi     domain.ABIRegistry
}

func (m *mockContext) ChainID() domain.ChainID     { return m.chainID }
func (m *mockContext) Tx() domain.Transaction      { return m.tx }
func (m *mockContext) Block() domain.Block         { return m.block }
func (m *mockContext) Logs() []domain.Log          { return m.logs }
func (m *mockContext) ABI() domain.ABIRegistry     { return m.abi }

// Builder pattern untuk easy setup
func newMockContext() *mockContext {
    return &mockContext{
        chainID: 1,
    }
}

func (m *mockContext) WithChainID(id domain.ChainID) *mockContext {
    m.chainID = id
    return m
}

func (m *mockContext) WithTransaction(tx domain.Transaction) *mockContext {
    m.tx = tx
    return m
}

func (m *mockContext) WithBlock(block domain.Block) *mockContext {
    m.block = block
    return m
}

func (m *mockContext) WithLogs(logs []domain.Log) *mockContext {
    m.logs = logs
    return m
}

// Usage
ctx := newMockContext().
    WithChainID(5000).
    WithTransaction(domain.Transaction{
        RawValue: domain.NativeValue{Wei: "1000000000000000000"},
    })
```

### 6.2 Mock EventSource

```go
// Controllable mock EventSource
type mockEventSource struct {
    events   chan asentric.Event
    startErr error
}

func newMockEventSource() *mockEventSource {
    return &mockEventSource{
        events: make(chan asentric.Event, 100),
    }
}

func (m *mockEventSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
    if m.startErr != nil {
        return nil, m.startErr
    }
    return m.events, nil
}

func (m *mockEventSource) Send(event asentric.Event) {
    m.events <- event
}

func (m *mockEventSource) Close() {
    close(m.events)
}

func (m *mockEventSource) WithStartError(err error) *mockEventSource {
    m.startErr = err
    return m
}
```

### 6.3 Mock AlertSink

```go
// Mock AlertSink that records all alerts
type mockAlertSink struct {
    alerts  []*asentric.Alert
    emitErr error
    mu      sync.Mutex
}

func newMockAlertSink() *mockAlertSink {
    return &mockAlertSink{
        alerts: make([]*asentric.Alert, 0),
    }
}

func (m *mockAlertSink) Emit(ctx context.Context, alert *asentric.Alert) error {
    if m.emitErr != nil {
        return m.emitErr
    }
    m.mu.Lock()
    defer m.mu.Unlock()
    m.alerts = append(m.alerts, alert)
    return nil
}

func (m *mockAlertSink) Alerts() []*asentric.Alert {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.alerts
}

func (m *mockAlertSink) WithEmitError(err error) *mockAlertSink {
    m.emitErr = err
    return m
}
```

### 6.4 Mock Rule

```go
// Configurable mock rule
type mockRule struct {
    name     string
    alert    *asentric.Alert
    err      error
    doPanic  bool
}

func newMockRule(name string) *mockRule {
    return &mockRule{name: name}
}

func (r *mockRule) Name() string { return r.name }

func (r *mockRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    if r.doPanic {
        panic("intentional panic")
    }
    return r.alert, r.err
}

func (r *mockRule) ReturnsAlert(alert *asentric.Alert) *mockRule {
    r.alert = alert
    return r
}

func (r *mockRule) ReturnsError(err error) *mockRule {
    r.err = err
    return r
}

func (r *mockRule) Panics() *mockRule {
    r.doPanic = true
    return r
}

// Usage
rule := newMockRule("test").ReturnsAlert(&asentric.Alert{
    Rule:     "test",
    Severity: asentric.SeverityHigh,
})
```

---

## 7. Test Coverage

### 7.1 Coverage Goals

| Component | Minimum Coverage |
|-----------|------------------|
| `pkg/asentric/engine.go` | 90% |
| `pkg/asentric/` (lainnya) | 80% |
| `internal/runtime/` | 80% |
| `internal/dispatcher/` | 80% |
| `internal/source/` | 70% |
| `internal/sink/` | 70% |
| `internal/queue/` | 70% |
| `internal/abi/` | 80% |
| `internal/config/` | 80% |

### 7.2 Generating Coverage Report

```bash
# Generate coverage file
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open HTML report
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### 7.3 Coverage per Package

```bash
# Coverage for specific package
go test -cover -coverprofile=pkg.out ./pkg/asentric/
go tool cover -func=pkg.out
```

---

## 8. Testing Checklist per Component

### 8.1 Engine (`pkg/asentric/engine.go`)

- [ ] `NewEngine()` returns valid instance
- [ ] `RegisterRule(nil)` returns `ErrInvalidRule`
- [ ] `RegisterRule(valid)` succeeds
- [ ] `Evaluate(nil)` returns `ErrInvalidContext`
- [ ] `Evaluate()` with empty rules returns empty alerts
- [ ] `Evaluate()` runs rules in registration order
- [ ] `Evaluate()` recovers from rule panic
- [ ] `Evaluate()` returns `ErrRulePanic` on panic

### 8.2 Rules

- [ ] `Name()` returns unique identifier
- [ ] `Evaluate()` returns `nil, nil` when no detection
- [ ] `Evaluate()` returns `alert, nil` when detection
- [ ] `Evaluate()` returns `nil, error` on error
- [ ] Rule is deterministic (same input → same output)
- [ ] Rule does not mutate Context
- [ ] Edge cases tested (zero values, nil, etc.)

### 8.3 Internal Runtime

- [ ] `Start()` idempotent (second call returns `ErrAlreadyRunning`)
- [ ] `Start()` returns error if EventSource fails
- [ ] `Stop()` idempotent
- [ ] `Stop()` triggers graceful shutdown
- [ ] Loop processes events correctly
- [ ] Loop exits on context cancellation
- [ ] Loop exits on channel close
- [ ] Dispatcher errors are handled

### 8.4 Dispatcher

- [ ] `Dispatch()` builds Context from Event
- [ ] `Dispatch()` calls `Engine.Evaluate()`
- [ ] `Dispatch()` sends alerts to Sink
- [ ] Error handling for Context build failure
- [ ] Error handling for Engine evaluation failure
- [ ] Webhook errors are logged but don't stop processing

### 8.5 EventSource (WebSocket)

- [ ] `Start()` connects to WebSocket
- [ ] `Start()` returns error on connection failure
- [ ] Subscription to `eth_subscribe` works
- [ ] Events are parsed correctly
- [ ] Channel is closed on context cancellation
- [ ] Connection errors close channel (no auto-reconnect)

### 8.6 AlertSink (Webhook)

- [ ] `Emit()` sends HTTP POST
- [ ] `Emit()` uses correct JSON format
- [ ] `Emit()` handles timeout
- [ ] `Emit()` returns error on failure
- [ ] Headers are set correctly

### 8.7 Config Loader

- [ ] `Load()` reads YAML file
- [ ] `Load()` validates required fields
- [ ] `Load()` applies defaults
- [ ] `Load()` returns error on invalid YAML
- [ ] `Load()` returns error on missing required fields

### 8.8 ABI Decoder

- [ ] `LoadABI()` reads JSON file
- [ ] `LoadABI()` parses ABI correctly
- [ ] `DecodeMethod()` decodes calldata
- [ ] `DecodeEvent()` decodes log data
- [ ] Error handling for invalid ABI
- [ ] Error handling for unknown selector/topic

---

## 9. Common Testing Patterns

### 9.1 Testing Error Conditions

```go
func TestSomething_ReturnsError(t *testing.T) {
    // Arrange
    input := invalidInput()
    
    // Act
    result, err := DoSomething(input)
    
    // Assert
    assert.ErrorIs(t, err, ErrExpected)
    assert.Nil(t, result)
}
```

### 9.2 Testing with Timeout

```go
func TestWithTimeout(t *testing.T) {
    done := make(chan bool)
    
    go func() {
        // Long-running operation
        DoSomethingSlow()
        done <- true
    }()
    
    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("test timed out")
    }
}
```

### 9.3 Testing Goroutines

```go
func TestGoroutine(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 1)
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := DoAsyncWork(); err != nil {
            errors <- err
        }
    }()
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("async error: %v", err)
    }
}
```

### 9.4 Testing Channels

```go
func TestChannel(t *testing.T) {
    ch := make(chan int, 1)
    
    // Send
    ch <- 42
    
    // Receive with timeout
    select {
    case val := <-ch:
        assert.Equal(t, 42, val)
    case <-time.After(time.Second):
        t.Fatal("expected value from channel")
    }
}
```

---

## 10. Troubleshooting

### 10.1 Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| `undefined: assert.Equal` | Missing testify | `go get github.com/stretchr/testify` |
| Test hangs | Goroutine leak / channel block | Add timeout, check channel close |
| Flaky test | Race condition | Use `-race` flag, add proper sync |
| Coverage low | Not testing edge cases | Add tests for nil, zero, error paths |

### 10.2 Debugging Tests

```bash
# Run with verbose
go test -v ./...

# Run with race detector
go test -race ./...

# Run specific test with more output
go test -v -run TestName ./path/to/package

# Print to stdout (for debugging)
t.Log("debug info")
t.Logf("value: %v", value)
```

### 10.3 Test Isolation

Jika tests saling mempengaruhi:

1. **Jangan gunakan global state**
2. **Setup dan teardown per test**
3. **Gunakan `t.Parallel()` dengan hati-hati**

```go
func TestIsolated(t *testing.T) {
    // Setup
    resource := createResource()
    
    // Teardown
    defer resource.Cleanup()
    
    // Test
    // ...
}
```

---

## Quick Reference

### Test Commands

```bash
# Run all tests
go test ./...

# Run with verbose
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v -run TestName ./path/

# Run with race detector
go test -race ./...

# Run integration tests
go test -v -tags=integration ./...

# Generate coverage HTML
go test -coverprofile=c.out ./... && go tool cover -html=c.out
```

### Assert Cheatsheet (testify)

```go
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Nil(t, obj)
assert.NotNil(t, obj)
assert.True(t, condition)
assert.False(t, condition)
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, targetErr)
assert.Len(t, slice, expectedLen)
assert.Empty(t, slice)
assert.Contains(t, slice, element)

// require stops test immediately on failure
require.NoError(t, err)
require.NotNil(t, obj)
```

---

**Happy Testing! 🧪**

---

**Last Updated:** December 2024  
**Version:** 1.0

