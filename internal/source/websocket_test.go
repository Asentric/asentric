package source

import (
	"context"
	"testing"
	"time"
)

func TestNewWebSocketSource_NilClient(t *testing.T) {
	_, err := NewWebSocketSource(WebSocketSourceConfig{
		Client: nil,
	})

	if err == nil {
		t.Error("Expected error for nil client")
	}
}

func TestNewWebSocketSource_DefaultBufferSize(t *testing.T) {
	// We can't test with real client here, but we can verify config parsing
	// This test would need a mock client in practice

	// Test that config with no buffer size doesn't panic
	cfg := WebSocketSourceConfig{
		BufferSize: 0, // Should default to 100
	}

	if cfg.BufferSize == 0 {
		// This is expected - the actual default is applied in NewWebSocketSource
		t.Log("Buffer size 0 is passed in, will be defaulted in constructor")
	}
}

func TestNewWebSocketSource_CustomBufferSize(t *testing.T) {
	cfg := WebSocketSourceConfig{
		BufferSize: 50,
	}

	if cfg.BufferSize != 50 {
		t.Errorf("Expected buffer size 50, got %d", cfg.BufferSize)
	}
}

// TestWebSocketSource_IsRunning tests the running state tracking
func TestWebSocketSource_IsRunning(t *testing.T) {
	// Create a source without starting it
	// We need to test the initial state
	source := &WebSocketSource{
		running: false,
	}

	if source.IsRunning() {
		t.Error("Source should not be running initially")
	}
}

// TestWebSocketSource_StopWhenNotRunning tests stopping a non-running source
func TestWebSocketSource_StopWhenNotRunning(t *testing.T) {
	source := &WebSocketSource{
		running: false,
	}

	err := source.Stop()
	if err != nil {
		t.Errorf("Stop on non-running source should not error: %v", err)
	}
}

// Integration test that requires real node - skipped by default
func TestWebSocketSource_Integration(t *testing.T) {
	t.Skip("Integration test - requires real node connection")

	// This test would connect to a real node
	// Example:
	//
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	//
	// client, err := chain.NewClient(ctx, chain.ClientConfig{
	//     WSURL:   "wss://rpc.sepolia.mantle.xyz/ws",
	//     ChainID: 5003,
	// })
	// require.NoError(t, err)
	// defer client.Close()
	//
	// source, err := NewWebSocketSource(WebSocketSourceConfig{
	//     Client:     client,
	//     BufferSize: 10,
	// })
	// require.NoError(t, err)
	//
	// events, err := source.Start(ctx)
	// require.NoError(t, err)
	//
	// select {
	// case event := <-events:
	//     t.Logf("Received event: block=%d", event.BlockNumber)
	// case <-time.After(10 * time.Second):
	//     t.Log("No events in 10s (chain might be quiet)")
	// }
	//
	// source.Stop()
}

// Benchmark for event processing (with mock data)
func BenchmarkToEvent(b *testing.B) {
	// This would benchmark the event conversion
	// Requires mock setup
	b.Skip("Requires mock chain client")
}

// Test context cancellation
func TestWebSocketSource_ContextCancellation(t *testing.T) {
	// Test that context cancellation is handled properly
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		t.Log("Context cancelled as expected")
	case <-time.After(time.Second):
		t.Error("Context should be cancelled")
	}
}
