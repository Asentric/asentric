package dispatcher

import (
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
)

// ContextBuilder creates Context from Event.
// This allows customization of how events are converted to contexts.
type ContextBuilder interface {
	Build(event asentric.Event) asentric.Context
}

// EngineDispatcher dispatches events to the engine for rule evaluation.
// It bridges EventSource and Engine, converting Events to Contexts.
type EngineDispatcher struct {
	Engine         *asentric.Engine
	Sink           asentric.AlertSink
	ContextBuilder ContextBuilder
	ABIRegistry    domain.ABIRegistry
}
