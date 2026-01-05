package testutils

import (
	"context"
	"sync"

	"github.com/asentric/asentric/pkg/asentric"
)

// MockAlertSink captures alerts for testing verification.
type MockAlertSink struct {
	mu      sync.Mutex
	alerts  []*CapturedAlert
	emitErr error
}

// CapturedAlert holds alert data captured by MockAlertSink.
type CapturedAlert struct {
	Ctx   context.Context
	Alert *asentric.Alert
}

// NewMockAlertSink creates a new MockAlertSink.
func NewMockAlertSink() *MockAlertSink {
	return &MockAlertSink{
		alerts: make([]*CapturedAlert, 0),
	}
}

// WithError configures the sink to return the given error on Emit.
func (m *MockAlertSink) WithError(err error) *MockAlertSink {
	m.emitErr = err
	return m
}

// Emit implements asentric.AlertSink.
// Captures alert for later verification.
func (m *MockAlertSink) Emit(ctx context.Context, evalCtx asentric.Context, alert *asentric.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.emitErr != nil {
		return m.emitErr
	}

	m.alerts = append(m.alerts, &CapturedAlert{
		Ctx:   ctx,
		Alert: alert,
	})
	return nil
}

// Alerts returns all captured alerts.
func (m *MockAlertSink) Alerts() []*CapturedAlert {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.alerts
}

// AlertCount returns the number of captured alerts.
func (m *MockAlertSink) AlertCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.alerts)
}

// Reset clears all captured alerts.
func (m *MockAlertSink) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = make([]*CapturedAlert, 0)
}

// Ensure MockAlertSink implements asentric.AlertSink
var _ asentric.AlertSink = (*MockAlertSink)(nil)
