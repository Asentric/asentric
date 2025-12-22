package runtime

import (
	"context"
	"sync"

	"github.com/asentric/asentric/pkg/asentric"
)

// Runtime orchestrates the lifecycle of EventSource and Dispatcher.
// Runtime is single-threaded and context-driven.
type Runtime struct {
	Source     asentric.EventSource
	Dispatcher asentric.Dispatcher

	ctx    context.Context
	cancel context.CancelFunc

	events <-chan asentric.Event

	mu      sync.Mutex
	running bool
}

func NewRuntime(
	source asentric.EventSource,
	dispatcher asentric.Dispatcher,
) *Runtime {
	return &Runtime{
		Source:     source,
		Dispatcher: dispatcher,
	}
}

// Start runs the runtime and blocks until completion.
// - Idempotent: a second start returns ErrAlreadyRunning
// - Lifecycle is controlled via context
func (r *Runtime) Start(parentCtx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return asentric.ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(parentCtx)
	r.ctx = ctx
	r.cancel = cancel
	r.running = true
	r.mu.Unlock()

	events, err := r.Source.Start(ctx)
	if err != nil {
		cancel()

		r.mu.Lock()
		r.running = false
		r.mu.Unlock()

		return err
	}

	r.events = events
	return r.loop()
}

// loop is the main single-threaded event loop.
func (r *Runtime) loop() error {
	defer func() {
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
	}()

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
}

// Stop gracefully halts the runtime.
// - Idempotent
// - Does not close channels
func (r *Runtime) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	if r.cancel != nil {
		r.cancel()
	}

	r.running = false
	return nil
}
