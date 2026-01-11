// Package source provides EventSource implementations for the Asentric SDK.
// These implementations subscribe to blockchain events and emit them for processing.
package source

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum"

	"github.com/asentric/asentric/internal/adapter"
	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/pkg/asentric"
)

// WebSocketSource implements asentric.EventSource using WebSocket subscription.
// It connects to an EVM node and subscribes to log events.
type WebSocketSource struct {
	client  *chain.Client
	filter  chain.SubscriptionFilter
	chainID uint64

	// Lifecycle management
	sub     ethereum.Subscription
	events  chan asentric.Event
	rawLogs <-chan chain.RawLog

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

// WebSocketSourceConfig holds configuration for WebSocketSource.
type WebSocketSourceConfig struct {
	// Client is the chain client (required)
	Client *chain.Client

	// Filter defines which logs to subscribe to (optional)
	// If nil, subscribes to all logs
	Filter *chain.SubscriptionFilter

	// BufferSize is the event channel buffer size (default: 100)
	BufferSize int
}

// NewWebSocketSource creates a new WebSocketSource.
func NewWebSocketSource(cfg WebSocketSourceConfig) (*WebSocketSource, error) {
	if cfg.Client == nil {
		return nil, fmt.Errorf("source: client is required")
	}

	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100
	}

	filter := chain.SubscriptionFilter{}
	if cfg.Filter != nil {
		filter = *cfg.Filter
	}

	return &WebSocketSource{
		client:  cfg.Client,
		filter:  filter,
		chainID: cfg.Client.ChainIDUint64(),
		events:  make(chan asentric.Event, bufferSize),
	}, nil
}

// Start begins event subscription and returns event channel.
// Implements asentric.EventSource interface.
func (s *WebSocketSource) Start(parentCtx context.Context) (<-chan asentric.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil, fmt.Errorf("source: already running")
	}

	// Create cancellable context for this subscription
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	// Subscribe to logs via chain client
	rawLogs, sub, err := s.client.SubscribeLogs(ctx, s.filter)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("source: subscribe failed: %w", err)
	}

	s.sub = sub
	s.rawLogs = rawLogs
	s.running = true

	// Start goroutine to process incoming logs
	go s.processLogs(ctx)

	return s.events, nil
}

// processLogs converts raw logs to events and sends them to the output channel.
func (s *WebSocketSource) processLogs(ctx context.Context) {
	defer func() {
		s.mu.Lock()
		s.running = false
		close(s.events)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case err := <-s.sub.Err():
			if err != nil {
				// Subscription error - log and exit
				// In production: implement reconnection logic
				fmt.Printf("source: subscription error: %v\n", err)
			}
			return

		case rawLog, ok := <-s.rawLogs:
			if !ok {
				// Channel closed
				return
			}

			// Skip removed logs (from chain reorganization)
			if rawLog.Removed {
				continue
			}

			// Convert RawLog to Event
			event := adapter.ToEvent(rawLog, s.chainID)

			// Send to output channel
			select {
			case s.events <- event:
				// Successfully sent
			case <-ctx.Done():
				return
			default:
				// Channel full - drop event with warning
				// In production: implement backpressure handling
				fmt.Printf("source: event channel full, dropping event at block %d\n", event.BlockNumber)
			}
		}
	}
}

// Stop gracefully stops the event source.
// Implements asentric.EventSource interface.
func (s *WebSocketSource) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Unsubscribe from node
	if s.sub != nil {
		s.sub.Unsubscribe()
	}

	// Cancel context to stop processing goroutine
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

// IsRunning returns whether the source is currently running.
func (s *WebSocketSource) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Ensure WebSocketSource implements asentric.EventSource
var _ asentric.EventSource = (*WebSocketSource)(nil)
