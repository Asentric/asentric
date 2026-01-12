package sink

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asentric/asentric/pkg/asentric"
)

// WebhookSink implements asentric.AlertSink for HTTP webhooks.
type WebhookSink struct {
	url       string
	client    *http.Client
	formatter *Formatter
	headers   map[string]string
}

// WebhookSinkConfig holds configuration for WebhookSink.
type WebhookSinkConfig struct {
	// URL is the webhook endpoint (required)
	URL string

	// Timeout for HTTP requests (default: 10s)
	Timeout time.Duration

	// Headers to include in requests (optional)
	Headers map[string]string

	// Formatter for alert payloads (optional, uses default if nil)
	Formatter *Formatter
}

// NewWebhookSink creates a new webhook sink.
func NewWebhookSink(cfg WebhookSinkConfig) (*WebhookSink, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("sink: webhook URL is required")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	formatter := cfg.Formatter
	if formatter == nil {
		formatter = DefaultFormatter()
	}

	return &WebhookSink{
		url: cfg.URL,
		client: &http.Client{
			Timeout: timeout,
		},
		formatter: formatter,
		headers:   cfg.Headers,
	}, nil
}

// Emit sends an alert to the webhook.
// Implements asentric.AlertSink interface.
func (s *WebhookSink) Emit(ctx context.Context, evalCtx asentric.Context, alert *asentric.Alert) error {
	// Format alert to JSON
	payload, err := s.formatter.FormatJSON(evalCtx, alert)
	if err != nil {
		return fmt.Errorf("sink: format failed: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("sink: create request failed: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sink: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("sink: webhook returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Ensure WebhookSink implements asentric.AlertSink
var _ asentric.AlertSink = (*WebhookSink)(nil)

// ConsoleSink implements asentric.AlertSink for stdout output.
// Useful for debugging and development.
type ConsoleSink struct {
	formatter *Formatter
}

// NewConsoleSink creates a new console sink.
func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{
		formatter: DefaultFormatter(),
	}
}

// Emit prints the alert to stdout.
// Implements asentric.AlertSink interface.
func (s *ConsoleSink) Emit(ctx context.Context, evalCtx asentric.Context, alert *asentric.Alert) error {
	payload, err := s.formatter.FormatJSONPretty(evalCtx, alert)
	if err != nil {
		return fmt.Errorf("sink: format failed: %w", err)
	}

	fmt.Printf("\n=== ALERT [%s] ===\n%s\n==================\n\n", alert.Severity, string(payload))
	return nil
}

// Ensure ConsoleSink implements asentric.AlertSink
var _ asentric.AlertSink = (*ConsoleSink)(nil)

// MultiSink sends alerts to multiple sinks.
type MultiSink struct {
	sinks []asentric.AlertSink
}

// NewMultiSink creates a sink that sends to multiple destinations.
func NewMultiSink(sinks ...asentric.AlertSink) *MultiSink {
	return &MultiSink{sinks: sinks}
}

// Emit sends the alert to all configured sinks.
// Returns the first error encountered but continues sending to remaining sinks.
func (s *MultiSink) Emit(ctx context.Context, evalCtx asentric.Context, alert *asentric.Alert) error {
	var firstErr error
	for _, sink := range s.sinks {
		if err := sink.Emit(ctx, evalCtx, alert); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Ensure MultiSink implements asentric.AlertSink
var _ asentric.AlertSink = (*MultiSink)(nil)

// NullSink discards all alerts. Useful for testing or disabling alerts.
type NullSink struct{}

// NewNullSink creates a sink that discards all alerts.
func NewNullSink() *NullSink {
	return &NullSink{}
}

// Emit does nothing and returns nil.
func (s *NullSink) Emit(ctx context.Context, evalCtx asentric.Context, alert *asentric.Alert) error {
	return nil
}

// Ensure NullSink implements asentric.AlertSink
var _ asentric.AlertSink = (*NullSink)(nil)
