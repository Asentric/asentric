package sink

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asentric/asentric/internal/testutils"
	"github.com/asentric/asentric/pkg/asentric"
)

func TestWebhookSink_Emit(t *testing.T) {
	// Create test server
	var receivedPayload AlertPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create sink
	sink, err := NewWebhookSink(WebhookSinkConfig{
		URL: server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create sink: %v", err)
	}

	// Create alert
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test-rule", "Test Alert", asentric.SeverityHigh).
		WithDescription("Test description")

	// Emit
	if err := sink.Emit(context.Background(), ctx, alert); err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Verify received payload
	if receivedPayload.Rule != "test-rule" {
		t.Errorf("Expected rule 'test-rule', got '%s'", receivedPayload.Rule)
	}

	if receivedPayload.Severity != "HIGH" {
		t.Errorf("Expected severity 'HIGH', got '%s'", receivedPayload.Severity)
	}

	if receivedPayload.Title != "Test Alert" {
		t.Errorf("Expected title 'Test Alert', got '%s'", receivedPayload.Title)
	}
}

func TestWebhookSink_EmitWithHeaders(t *testing.T) {
	// Create test server
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create sink with custom headers
	sink, err := NewWebhookSink(WebhookSinkConfig{
		URL: server.URL,
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create sink: %v", err)
	}

	// Emit
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityLow)

	if err := sink.Emit(context.Background(), ctx, alert); err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Verify header
	if receivedAuth != "Bearer test-token" {
		t.Errorf("Expected 'Bearer test-token', got '%s'", receivedAuth)
	}
}

func TestWebhookSink_EmitError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	sink, _ := NewWebhookSink(WebhookSinkConfig{URL: server.URL})
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityLow)

	err := sink.Emit(context.Background(), ctx, alert)
	if err == nil {
		t.Error("Expected error for 500 response")
	}
}

func TestWebhookSink_RequiresURL(t *testing.T) {
	_, err := NewWebhookSink(WebhookSinkConfig{})
	if err == nil {
		t.Error("Expected error for missing URL")
	}
}

func TestConsoleSink_Emit(t *testing.T) {
	sink := NewConsoleSink()
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityInfo)

	// Should not error (output goes to stdout)
	err := sink.Emit(context.Background(), ctx, alert)
	if err != nil {
		t.Errorf("ConsoleSink.Emit failed: %v", err)
	}
}

func TestMultiSink_Emit(t *testing.T) {
	var called1, called2 bool

	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called1 = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called2 = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	// Create multi sink
	sink1, _ := NewWebhookSink(WebhookSinkConfig{URL: server1.URL})
	sink2, _ := NewWebhookSink(WebhookSinkConfig{URL: server2.URL})
	multiSink := NewMultiSink(sink1, sink2)

	// Emit
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityLow)

	if err := multiSink.Emit(context.Background(), ctx, alert); err != nil {
		t.Fatalf("MultiSink.Emit failed: %v", err)
	}

	if !called1 || !called2 {
		t.Errorf("Expected both sinks to be called: sink1=%v, sink2=%v", called1, called2)
	}
}

func TestMultiSink_ContinuesOnError(t *testing.T) {
	var called2 bool

	// First server returns error
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server1.Close()

	// Second server works
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called2 = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	sink1, _ := NewWebhookSink(WebhookSinkConfig{URL: server1.URL})
	sink2, _ := NewWebhookSink(WebhookSinkConfig{URL: server2.URL})
	multiSink := NewMultiSink(sink1, sink2)

	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityLow)

	// Should return error from first sink but still call second
	err := multiSink.Emit(context.Background(), ctx, alert)
	if err == nil {
		t.Error("Expected error from failed sink")
	}

	if !called2 {
		t.Error("Second sink should have been called even though first failed")
	}
}

func TestNullSink_Emit(t *testing.T) {
	sink := NewNullSink()
	ctx := testutils.NewMockContext()
	alert := asentric.NewAlert("test", "Test", asentric.SeverityHigh)

	err := sink.Emit(context.Background(), ctx, alert)
	if err != nil {
		t.Errorf("NullSink.Emit should never error: %v", err)
	}
}
