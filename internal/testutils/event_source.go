package testutils

import (
	"context"

	"github.com/asentric/asentric/pkg/asentric"
)

// MockEventSource is a configurable EventSource for testing.
// It allows injecting events and controlling lifecycle behavior.
type MockEventSource struct {
	events   []asentric.Event
	eventsCh chan asentric.Event
	started  bool
	stopped  bool
}

// NewMockEventSource creates a new MockEventSource.
func NewMockEventSource() *MockEventSource {
	return &MockEventSource{
		events: make([]asentric.Event, 0),
	}
}

// WithEvents configures the source to emit the given events when started.
func (m *MockEventSource) WithEvents(events ...asentric.Event) *MockEventSource {
	m.events = append(m.events, events...)
	return m
}

// Start implements asentric.EventSource.
// Returns a channel that will emit configured events, then close.
func (m *MockEventSource) Start(ctx context.Context) (<-chan asentric.Event, error) {
	m.started = true
	m.eventsCh = make(chan asentric.Event, len(m.events))

	// Send all configured events
	go func() {
		defer close(m.eventsCh)
		for _, event := range m.events {
			select {
			case <-ctx.Done():
				return
			case m.eventsCh <- event:
			}
		}
	}()

	return m.eventsCh, nil
}

// Stop implements asentric.EventSource.
func (m *MockEventSource) Stop() error {
	m.stopped = true
	return nil
}

// IsStarted returns true if Start was called.
func (m *MockEventSource) IsStarted() bool {
	return m.started
}

// IsStopped returns true if Stop was called.
func (m *MockEventSource) IsStopped() bool {
	return m.stopped
}

// Ensure MockEventSource implements asentric.EventSource
var _ asentric.EventSource = (*MockEventSource)(nil)
