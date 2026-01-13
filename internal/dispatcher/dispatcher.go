package dispatcher

import (
	"context"

	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// ContextBuilder defines how an Event is transformed into an Engine Context.
//
// This abstraction decouples event ingestion from rule evaluation logic,
// allowing different implementations to:
//   - Decode raw event payloads
//   - Enrich events with on-chain / off-chain metadata
//   - Prepare a normalized Context for Engine evaluation
//
// Implementations must return a fully-initialized Context suitable for
// Engine.Evaluate.
type ContextBuilder interface {
	// Build converts the given Event into an asentric.Context.
	//
	// The returned Context is passed directly to Engine.Evaluate and must
	// contain all data required by registered rules.
	// Returns an error if the context cannot be built from the event.
	Build(event asentric.Event) (asentric.Context, error)
}

// EngineDispatcher routes incoming Events to the rule Engine.
//
// EngineDispatcher acts as the orchestration layer between an EventSource
// and the Engine by performing the following steps:
//  1. Accepting an incoming Event
//  2. Building an evaluation Context via ContextBuilder
//  3. Evaluating rules using Engine.Evaluate
//  4. Emitting generated Alerts through AlertSink
//
// The dispatcher itself contains no business logic and is designed to be
// lightweight, deterministic, and easy to test.
type EngineDispatcher struct {
	engine         *asentric.Engine
	alertSink      asentric.AlertSink
	contextBuilder ContextBuilder
	abiRegistry    domain.ABIRegistry
}

// EngineDispatcherConfig defines all dependencies required to construct
// an EngineDispatcher.
//
// All fields are required and validated during initialization to ensure
// the dispatcher is always created in a valid state.
type EngineDispatcherConfig struct {
	// Engine evaluates rules against a Context.
	Engine *asentric.Engine

	// AlertSink receives and delivers alerts produced by rule evaluation.
	AlertSink asentric.AlertSink

	// ContextBuilder converts Events into evaluation-ready Contexts.
	ContextBuilder ContextBuilder

	// ABIRegistry provides ABI definitions used when decoding event data.
	ABIRegistry domain.ABIRegistry
}

// NewEngineDispatcher creates and initializes a new EngineDispatcher.
//
// All required dependencies are validated before construction.
// If any dependency is missing, a well-defined dispatcher error is returned.
func NewEngineDispatcher(config EngineDispatcherConfig) (*EngineDispatcher, error) {
	if err := validateNewEngineDispatcher(config); err != nil {
		return nil, err
	}

	return &EngineDispatcher{
		engine:         config.Engine,
		alertSink:      config.AlertSink,
		contextBuilder: config.ContextBuilder,
		abiRegistry:    config.ABIRegistry,
	}, nil
}

// validateNewEngineDispatcher validates required dependencies for
// EngineDispatcher construction.
//
// This function ensures the dispatcher fails fast during initialization
// rather than at runtime.
func validateNewEngineDispatcher(config EngineDispatcherConfig) error {
	if config.Engine == nil {
		return asentric.ErrDispatcherEngine
	}

	if config.AlertSink == nil {
		return asentric.ErrDispatcherAlertSink
	}

	if config.ContextBuilder == nil {
		return asentric.ErrDispatcherContextBuilder
	}

	if config.ABIRegistry == nil {
		return asentric.ErrDispatcherABIRegistry
	}

	return nil
}

// Dispatch processes a single Event through the Engine lifecycle.
//
// Processing flow:
//  1. Build a Context from the incoming Event
//  2. Evaluate rules using Engine.Evaluate
//  3. Emit all generated Alerts via AlertSink
//
// Dispatch uses fail-fast semantics:
//   - If rule evaluation fails, dispatch stops immediately
//   - If emitting any alert fails, dispatch stops and returns the error
//
// A nil error indicates the event was fully processed without failures.
func (e *EngineDispatcher) Dispatch(ctx context.Context, event asentric.Event) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Step 1: Build evaluation context from event
	evalCtx, err := e.contextBuilder.Build(event)
	if err != nil {
		return err
	}

	// Step 2: Evaluate rules against the context
	alerts, err := e.engine.Evaluate(evalCtx)
	if err != nil {
		return err
	}

	// Step 3: Emit generated alerts
	for _, alert := range alerts {
		// Check for cancellation before each emit
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.alertSink.Emit(ctx, evalCtx, alert); err != nil {
			return err
		}
	}

	return nil
}
