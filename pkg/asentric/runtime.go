package asentric

import (
	"context"
)

// Runtime is the public facade for running Asentric.
// It orchestrates EventSource, Engine, and AlertSink.
//
// Usage:
//
//	config, _ := asentric.LoadConfig("config/")
//	engine := asentric.NewEngine()
//	engine.RegisterRule(&MyRule{})
//	runtime := asentric.NewRuntime(config, engine)
//	runtime.Start(context.Background())
type Runtime struct {
	config   *RuntimeConfig
	registry *RegistryConfig
	engine   *Engine
	logger   Logger

	// Internal components (injected by internal packages)
	source     EventSource
	sink       AlertSink
	dispatcher Dispatcher

	// Lifecycle
	cancel  context.CancelFunc
	running bool
}
