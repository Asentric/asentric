// Package runtime provides a high-level builder for creating Asentric runtime instances.
// This package exposes a simplified API that wraps internal components.
package runtime

import (
	"context"
	"fmt"

	abiPkg "github.com/asentric/asentric/internal/abi"
	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/internal/dispatcher"
	"github.com/asentric/asentric/internal/sink"
	"github.com/asentric/asentric/internal/source"
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// Builder provides a fluent API for creating runtime instances.
type Builder struct {
	config      *asentric.RuntimeConfig
	engine      *asentric.Engine
	eventSource asentric.EventSource
	alertSink   asentric.AlertSink
	chainClient *chain.Client
	err         error
}

// NewBuilder creates a new runtime builder.
func NewBuilder(config *asentric.RuntimeConfig, engine *asentric.Engine) *Builder {
	return &Builder{
		config: config,
		engine: engine,
	}
}

// WithWebSocketSource creates a WebSocket event source from config.
func (b *Builder) WithWebSocketSource(ctx context.Context) *Builder {
	if b.err != nil {
		return b
	}

	// Create chain client
	client, err := chain.NewClient(ctx, chain.ClientConfig{
		WSURL:   b.config.Chain.RPCWS,
		ChainID: b.config.Chain.ID,
	})
	if err != nil {
		b.err = fmt.Errorf("failed to create chain client: %w", err)
		return b
	}
	b.chainClient = client

	// Create event source
	eventSource, err := source.NewWebSocketSource(source.WebSocketSourceConfig{
		Client:      client,
		BufferSize:  100,
		UseNewHeads: true,
	})
	if err != nil {
		b.err = fmt.Errorf("failed to create event source: %w", err)
		return b
	}
	b.eventSource = eventSource

	return b
}

// WithConsoleSink creates a console alert sink.
func (b *Builder) WithConsoleSink() *Builder {
	if b.err != nil {
		return b
	}
	b.alertSink = sink.NewConsoleSink()
	return b
}

// WithWebhookSink creates a webhook alert sink.
func (b *Builder) WithWebhookSink(url string) *Builder {
	if b.err != nil {
		return b
	}

	s, err := sink.NewWebhookSink(sink.WebhookSinkConfig{URL: url})
	if err != nil {
		b.err = fmt.Errorf("failed to create webhook sink: %w", err)
		return b
	}
	b.alertSink = s
	return b
}

// WithSinkFromConfig creates alert sink based on config.
func (b *Builder) WithSinkFromConfig() *Builder {
	if b.err != nil {
		return b
	}

	switch b.config.Sink.Type {
	case "console":
		return b.WithConsoleSink()
	case "webhook":
		return b.WithWebhookSink(b.config.Sink.URL)
	default:
		b.err = fmt.Errorf("unknown sink type: %s", b.config.Sink.Type)
	}
	return b
}

// Build creates the final runtime instance.
func (b *Builder) Build() (*asentric.Runtime, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.eventSource == nil {
		return nil, fmt.Errorf("event source not configured")
	}
	if b.alertSink == nil {
		return nil, fmt.Errorf("alert sink not configured")
	}

	// Create runtime
	runtime, err := asentric.NewRuntime(b.config, b.engine)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	// Create context builder and dispatcher
	abiRegistry := abiPkg.NewRegistry()
	contextBuilder := dispatcher.NewContextBuilderAdapter(
		dispatcher.NewDefaultContextBuilder(dispatcher.ContextBuilderConfig{
			ChainID:     domain.ChainID(b.config.Chain.ID),
			ABIRegistry: abiRegistry,
		}),
	)

	disp, err := dispatcher.NewEngineDispatcher(dispatcher.EngineDispatcherConfig{
		Engine:         b.engine,
		AlertSink:      b.alertSink,
		ContextBuilder: contextBuilder,
		ABIRegistry:    abiRegistry,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dispatcher: %w", err)
	}

	// Wire up runtime
	runtime.
		WithEventSource(b.eventSource).
		WithAlertSink(b.alertSink).
		WithDispatcher(disp)

	return runtime, nil
}

// Error returns any error that occurred during building.
func (b *Builder) Error() error {
	return b.err
}
