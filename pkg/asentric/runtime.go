package asentric

import (
	"context"
	"sync"
)

// Runtime is a thin orchestration facade.
//
// Responsibilities:
// - Start/Stop lifecycle
// - Consume events from EventSource
// - Forward events to Dispatcher
//
// Non-responsibilities:
// - NO buffering
// - NO concurrency management
// - NO context building
// - NO rule evaluation
// - NO alert delivery
//
// Any backpressure, queueing, or parallelism
// MUST be handled by EventSource or Dispatcher.

// Runtime NEVER builds implementations itself.
// All components MUST be injected directly via With* methods.
//
// Usage:
//
//	config, _ := asentric.LoadConfig("config/")
//	engine := asentric.NewEngine()
//	engine.RegisterRule(&MyRule{})
//
//	// Build components in runtime layer (using factories)
//	source := buildEventSource(config)  // Factory helper in runtime layer
//	sink := buildAlertSink(config)      // Factory helper in runtime layer
//
//	// Direct injection to Runtime
//	runtime, err := asentric.NewRuntime(config, engine)
//	if err != nil {
//		log.Fatal(err)
//	}
//	runtime.
//		WithEventSource(source).
//		WithAlertSink(sink)
//
//	runtime.Start(context.Background())

type Runtime struct {
	config   *RuntimeConfig
	registry *RegistryConfig
	// Runtime does NOT call engine directly.
	// Engine is owned by Dispatcher and injected into Runtime.
	engine *Engine
	logger Logger

	// Direct Injected Components
	source EventSource
	// AlertSink is injected for lifecycle validation.
	// It is NOT used directly by Runtime.
	// AlertSink owns alert delivery.

	sink       AlertSink
	dispatcher Dispatcher

	// Lifecycle
	shutdown chan struct{}
	done     chan struct{}
	once     sync.Once

	running bool
}

// NewRuntime creates a new runtime with the given config and engine.
// Returns error if config or engine is invalid.
//
// Validations performed:
//   - config must not be nil
//   - engine must not be nil
//   - config.Chain.RPCWS must not be empty
//   - config.Chain.Name must not be empty
//   - config.Redis.Addr must not be empty (if Redis is required)
//   - config.Webhook.URL must not be empty (if Webhook is required)
//
// Note: Source, Sink, and Dispatcher are built later in Start() method.
// This constructor only validates configuration, not runtime components.
func NewRuntime(cfg *RuntimeConfig, engine *Engine) (*Runtime, error) {

	// Validate Config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Validate Engine
	if engine == nil {
		return nil, ErrEngineRequired
	}

	rt := &Runtime{
		config:   cfg,
		engine:   engine,
		logger:   NewDefaultLogger(cfg.Debug),
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
	return rt, nil
}

func (r *Runtime) WithLogger(logger Logger) *Runtime {
	r.logger = logger
	return r
}

// WithEventSource directly injects EventSource.
// This is the ONLY way to provide EventSource to Runtime.
//
// Runtime NEVER builds EventSource itself.
// Use factory helpers in runtime layer to build, then inject here.
//
// Example:
//
//	source := buildEventSource(config)  // Factory helper (in runtime layer)
//	runtime.WithEventSource(source)     // Direct injection
func (r *Runtime) WithEventSource(source EventSource) *Runtime {
	r.source = source
	return r
}

// WithAlertSink directly injects AlertSink.
// This is the ONLY way to provide AlertSink to Runtime.
func (r *Runtime) WithAlertSink(sink AlertSink) *Runtime {
	r.sink = sink
	return r
}

// WithDispatcher directly injects Dispatcher.
// This is the ONLY way to provide Dispatcher to Runtime.
func (r *Runtime) WithDispatcher(dispatcher Dispatcher) *Runtime {
	r.dispatcher = dispatcher
	return r
}

func (r *Runtime) Start(ctx context.Context) error {
	if err := r.validateDependencies(); err != nil {
		return err
	}

	if r.running {
		return ErrAlreadyRunning
	}

	r.logger.Info("Starting runtime...")

	r.running = true
	defer func() {
		r.running = false
		close(r.done)
		err := r.source.Stop()
		if err != nil {
			r.logger.Error("Error stopping event source", "error", err)
		}
	}()

	events, err := r.source.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Context cancelled, stopping runtime")
			return ctx.Err()

		case <-r.shutdown:
			r.logger.Info("Shutdown signal received")
			return nil

		case event, ok := <-events:
			if !ok {
				r.logger.Info("Event source closed")
				return nil
			}

			if err := r.dispatcher.Dispatch(ctx, event); err != nil {
				r.logger.Error("Error dispatching event", "error", err)
			}
		}
	}
}

// validateDependencies validates the dependencies of the runtime.
func (r *Runtime) validateDependencies() error {
	if r.source == nil {
		return ErrSourceRequired
	}
	if r.sink == nil {
		return ErrSinkRequired
	}
	if r.dispatcher == nil {
		return ErrDispatcherRequired
	}
	return nil
}

// Stop gracefully stops the runtime.
func (r *Runtime) Stop(ctx context.Context) error {
	r.logger.Info("Stopping runtime...")
	r.once.Do(func() {
		close(r.shutdown)
	})

	// Wait for done or context timeout
	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns true if the runtime is currently running.
func (r *Runtime) IsRunning() bool {
	return r.running
}
