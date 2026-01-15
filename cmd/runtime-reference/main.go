// Package main provides a standalone runtime binary for Asentric.
// This is a reference implementation that can be used directly
// without writing Go code - just provide a configuration file.
//
// Usage:
//
//	./runtime-reference --config config/asentric.yaml
package main

import (
	"context"
	"flag"
	"fmt"
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

	"github.com/ethereum/go-ethereum/common"
)

var (
	// Version information - set via ldflags at build time
	Version   = "0.2.2"
	GitCommit = "dev"
	BuildDate = "unknown"
)

// MockUSDC contract on Base Sepolia (for demo)
const MockUSDCAddress = "0xc309D45d4119487b30205784efF9abACF20872c0"

func main() {
	configPath := flag.String("config", "config/asentric.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if err := run(*configPath); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func printVersion() {
	fmt.Println("Asentric Runtime Reference")
	fmt.Println("==========================")
	fmt.Printf("  Version:    %s\n", Version)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
	fmt.Printf("  Built:      %s\n", BuildDate)
}

func run(configPath string) error {
	log.Println("========================================")
	log.Printf("Asentric Runtime Reference v%s", Version)
	log.Println("========================================")

	// Load configuration
	log.Printf("Loading configuration from: %s", configPath)
	cfg, err := asentric.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create engine and register rules
	engine := asentric.NewEngine()
	if err := registerRules(engine); err != nil {
		return fmt.Errorf("failed to register rules: %w", err)
	}
	log.Println("Registered built-in rules")

	// Create runtime
	runtime, err := asentric.NewRuntime(cfg, engine)
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	// Build components
	eventSource, abiRegistry, err := buildEventSource(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("failed to create event source: %w", err)
	}
	log.Println("Event source ready")

	alertSink, err := buildAlertSink(cfg)
	if err != nil {
		return fmt.Errorf("failed to create alert sink: %w", err)
	}
	log.Println("Alert sink ready")

	// Create dispatcher with context builder
	contextBuilder := dispatcher.NewContextBuilderAdapter(
		dispatcher.NewDefaultContextBuilder(dispatcher.ContextBuilderConfig{
			ChainID:     domain.ChainID(cfg.Chain.ID),
			ABIRegistry: abiRegistry,
		}),
	)

	disp, err := dispatcher.NewEngineDispatcher(dispatcher.EngineDispatcherConfig{
		Engine:         engine,
		AlertSink:      alertSink,
		ContextBuilder: contextBuilder,
		ABIRegistry:    abiRegistry,
	})
	if err != nil {
		return fmt.Errorf("failed to create dispatcher: %w", err)
	}

	// Wire up runtime
	runtime.
		WithEventSource(eventSource).
		WithAlertSink(alertSink).
		WithDispatcher(disp)

	// Setup graceful shutdown
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	}()

	// Start runtime
	log.Println("----------------------------------------")
	log.Printf("Chain:  %s (ID: %d)", cfg.Chain.Name, cfg.Chain.ID)
	log.Printf("Source: %s", cfg.Source.Type)
	log.Printf("Sink:   %s", cfg.Sink.Type)
	log.Println("----------------------------------------")
	log.Println("Listening for events... (Press Ctrl+C to stop)")

	if err := runtime.Start(runCtx); err != nil && err != context.Canceled {
		return fmt.Errorf("runtime error: %w", err)
	}

	log.Println("Runtime stopped gracefully.")
	return nil
}

// registerRules registers the default detection rules.
func registerRules(engine *asentric.Engine) error {
	// Default ERC20 rule - listens to all ERC20 Transfer/Mint/Burn events with value > 100
	// This is the main default rule for demo
	if err := engine.RegisterRule(rules.NewERC20DefaultRule()); err != nil {
		return err
	}

	// Large transfer detection
	if err := engine.RegisterRule(rules.NewLargeTransferRule()); err != nil {
		return err
	}

	// Whale alert
	if err := engine.RegisterRule(rules.NewWhaleAlertRule()); err != nil {
		return err
	}

	// MockUSDC monitoring (Base Sepolia demo)
	if err := engine.RegisterRule(rules.NewERC20TransferRule(rules.ERC20TransferConfig{
		ContractAddress: MockUSDCAddress,
		TokenSymbol:     "USDC",
		Decimals:        6,
	})); err != nil {
		return err
	}

	return nil
}

// buildEventSource creates the appropriate EventSource based on config.
func buildEventSource(ctx context.Context, cfg *asentric.RuntimeConfig) (asentric.EventSource, domain.ABIRegistry, error) {
	abiRegistry := abiPkg.NewRegistry()

	// Register standard ERC20 ABI for all contracts
	// For demo, we'll register it dynamically when we encounter Transfer events
	// For now, we register it for known contracts
	// Note: In production, you might want to register ABIs from config or discover them dynamically
	loader := abiPkg.NewLoader(abiRegistry)
	err := loader.LoadAndRegisterString(
		common.HexToAddress(MockUSDCAddress),
		abiPkg.StandardERC20ABI,
	)
	if err == nil {
		log.Println("Registered standard ERC20 ABI for default detection")
	}

	switch cfg.Source.Type {
	case "websocket":
		client, err := chain.NewClient(ctx, chain.ClientConfig{
			WSURL:   cfg.Chain.RPCWS,
			ChainID: cfg.Chain.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("chain client: %w", err)
		}

		eventSource, err := source.NewWebSocketSource(source.WebSocketSourceConfig{
			Client:      client,
			BufferSize:  100,
			UseNewHeads: true, // Use newHeads for Alchemy compatibility
		})
		if err != nil {
			return nil, nil, fmt.Errorf("websocket source: %w", err)
		}

		return eventSource, abiRegistry, nil

	case "memory":
		return nil, nil, fmt.Errorf("memory source not supported (use for testing only)")

	default:
		return nil, nil, fmt.Errorf("unknown source type: %s", cfg.Source.Type)
	}
}

// buildAlertSink creates the appropriate AlertSink based on config.
func buildAlertSink(cfg *asentric.RuntimeConfig) (asentric.AlertSink, error) {
	switch cfg.Sink.Type {
	case "console":
		return sink.NewConsoleSink(), nil

	case "webhook":
		if cfg.Sink.URL == "" {
			return nil, fmt.Errorf("webhook URL is required when sink.type=webhook")
		}
		return sink.NewWebhookSink(sink.WebhookSinkConfig{URL: cfg.Sink.URL})

	default:
		return nil, fmt.Errorf("unknown sink type: %s", cfg.Sink.Type)
	}
}
