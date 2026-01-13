// Package runtime provides a high-level builder API for creating Asentric runtime instances.
//
// This package simplifies the process of wiring up all the components needed for
// a working Asentric watcher. It wraps internal components and provides a clean
// public API.
//
// Basic Usage:
//
//	cfg, _ := asentric.LoadConfig("config/asentric.yaml")
//	engine := asentric.NewEngine()
//	engine.RegisterRule(myRule)
//
//	runtime, err := runtime.NewBuilder(cfg, engine).
//	    WithWebSocketSource(ctx).
//	    WithSinkFromConfig().
//	    Build()
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	runtime.Start(ctx)
//
// Custom Configuration:
//
//	runtime, err := runtime.NewBuilder(cfg, engine).
//	    WithWebSocketSource(ctx).
//	    WithWebhookSink("https://my-webhook.com/alerts").
//	    Build()
package runtime
