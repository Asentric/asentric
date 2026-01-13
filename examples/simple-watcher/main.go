// Package main provides a simple watcher example for the Asentric SDK.
// This example demonstrates monitoring MockUSDC contract on Base Sepolia,
// detecting Transfer events (including mints) in real-time.
//
// Usage:
//
//	go run examples/simple-watcher/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	abiPkg "github.com/asentric/asentric/internal/abi"
	"github.com/asentric/asentric/internal/chain"
	"github.com/asentric/asentric/internal/dispatcher"
	"github.com/asentric/asentric/internal/sink"
	"github.com/asentric/asentric/internal/source"
	"github.com/asentric/asentric/pkg/asentric"
	"github.com/asentric/asentric/pkg/domain"
	"github.com/asentric/asentric/pkg/rules"
)

// MockUSDC contract on Base Sepolia
const MockUSDCAddress = "0xc309D45d4119487b30205784efF9abACF20872c0"

func main() {
	log.Println("===========================================")
	log.Println("  Asentric Simple Watcher - MockUSDC Demo")
	log.Println("===========================================")

	// Load configuration
	cfg, err := asentric.LoadConfig("examples/simple-watcher/config/asentric.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create engine and register rules
	engine := asentric.NewEngine()
	registerRules(engine)

	log.Println("Registered rules:")
	log.Println("  - usdc-transfer: Detects MockUSDC Transfer events (including mint/burn)")

	// Create runtime
	runtime, err := asentric.NewRuntime(cfg, engine)
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}

	// Connect to blockchain
	log.Printf("Connecting to %s...", cfg.Chain.Name)
	chainClient, err := chain.NewClient(context.Background(), chain.ClientConfig{
		WSURL:   cfg.Chain.RPCWS,
		ChainID: cfg.Chain.ID,
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("Connected to blockchain")

	// Create event source
	eventSource, err := source.NewWebSocketSource(source.WebSocketSourceConfig{
		Client:      chainClient,
		BufferSize:  100,
		UseNewHeads: true, // Use newHeads for Alchemy compatibility
	})
	if err != nil {
		log.Fatalf("Failed to create event source: %v", err)
	}

	// Create alert sink (console for demo)
	alertSink := sink.NewConsoleSink()

	// Create ABI registry and context builder
	abiRegistry := abiPkg.NewRegistry()
	contextBuilder := dispatcher.NewContextBuilderAdapter(
		dispatcher.NewDefaultContextBuilder(dispatcher.ContextBuilderConfig{
			ChainID:     domain.ChainID(cfg.Chain.ID),
			ABIRegistry: abiRegistry,
		}),
	)

	// Create dispatcher
	disp, err := dispatcher.NewEngineDispatcher(dispatcher.EngineDispatcherConfig{
		Engine:         engine,
		AlertSink:      alertSink,
		ContextBuilder: contextBuilder,
		ABIRegistry:    abiRegistry,
	})
	if err != nil {
		log.Fatalf("Failed to create dispatcher: %v", err)
	}

	// Wire up runtime
	runtime.
		WithEventSource(eventSource).
		WithAlertSink(alertSink).
		WithDispatcher(disp)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("\nShutting down...")
		cancel()
	}()

	// Start
	log.Println("-------------------------------------------")
	log.Printf("Chain:    %s (ID: %d)", cfg.Chain.Name, cfg.Chain.ID)
	log.Printf("Contract: MockUSDC @ %s", MockUSDCAddress)
	log.Printf("Watching: Transfer events (mint/burn/transfer)")
	log.Println("-------------------------------------------")
	log.Println("Listening for events... (Press Ctrl+C to stop)")
	log.Println("")

	if err := runtime.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Runtime error: %v", err)
	}

	log.Println("Watcher stopped.")
}

// registerRules registers the MockUSDC monitoring rules.
func registerRules(engine *asentric.Engine) {
	// MockUSDC Transfer detection (includes mint/burn detection)
	engine.RegisterRule(rules.NewERC20TransferRule(rules.ERC20TransferConfig{
		ContractAddress: MockUSDCAddress,
		TokenSymbol:     "USDC",
		Decimals:        6,
	}))
}
