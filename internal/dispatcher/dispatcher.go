package dispatcher

import (
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// ContextBuilder is responsible for converting an Event into an Engine Context.
//
// This abstraction allows different implementations to customize how raw
// events (e.g. blockchain logs, transactions, or external signals) are
// transformed into a structured Context that can be evaluated by the Engine.
type ContextBuilder interface {
	// Build creates an asentric.Context from the given Event.
	//
	// The returned Context will later be passed to Engine.Evaluate.
	Build(event asentric.Event) asentric.Context
}

// EngineDispatcher dispatches incoming Events to the rule Engine.
//
// It acts as a bridge between an EventSource and the Engine by:
//  1. Receiving an Event
//  2. Building a Context using ContextBuilder
//  3. Evaluating rules via the Engine
//  4. Emitting generated Alerts through AlertSink
type EngineDispatcher struct {
	engine         *asentric.Engine
	alertSink      asentric.AlertSink
	contextBuilder ContextBuilder
	abiRegistry    domain.ABIRegistry
}

// EngineDispatcherConfig contains all dependencies required to construct
// an EngineDispatcher.
//
// All fields are mandatory and will be validated during initialization.
type EngineDispatcherConfig struct {
	// Engine is responsible for evaluating rules against a Context.
	Engine *asentric.Engine

	// AlertSink is used to emit alerts produced by the Engine.
	AlertSink asentric.AlertSink

	// ContextBuilder converts Events into evaluation-ready Contexts.
	ContextBuilder ContextBuilder

	// ABIRegistry provides ABI definitions used during event decoding.
	ABIRegistry domain.ABIRegistry
}

// NewEngineDispatcher creates a new EngineDispatcher instance.
//
// It validates that all required dependencies are provided.
// Returns a descriptive error if any dependency is missing.
func NewEngineDispatcher(config EngineDispatcherConfig) (*EngineDispatcher, error) {
	if config.Engine == nil {
		return nil, asentric.ErrDispatcherEngine
	}

	if config.AlertSink == nil {
		return nil, asentric.ErrDispatcherAlertSink
	}

	if config.ContextBuilder == nil {
		return nil, asentric.ErrDispatcherContextBuilder
	}

	if config.ABIRegistry == nil {
		return nil, asentric.ErrDispatcherABIRegistry
	}

	return &EngineDispatcher{
		engine:         config.Engine,
		alertSink:      config.AlertSink,
		contextBuilder: config.ContextBuilder,
		abiRegistry:    config.ABIRegistry,
	}, nil
}

// Dispatch processes a single Event through the Engine.
//
// Execution flow:
//  1. Build Context from Event using ContextBuilder
//  2. Evaluate rules using Engine.Evaluate
//  3. Emit each generated Alert via AlertSink
//
// Dispatch returns an error if:
//   - Rule evaluation fails
//   - Emitting any alert fails
func (e *EngineDispatcher) Dispatch(event asentric.Event) error {
	// Step 1: Build evaluation context from event
	ctx := e.contextBuilder.Build(event)

	// Step 2: Evaluate rules against the context
	alerts, err := e.engine.Evaluate(ctx)
	if err != nil {
		return err
	}

	// Step 3: Emit generated alerts
	for _, alert := range alerts {
		if err := e.alertSink.Emit(ctx, alert); err != nil {
			return err
		}
	}

	return nil
}
