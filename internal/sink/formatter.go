// Package sink provides AlertSink implementations for delivering alerts.
package sink

import (
	"encoding/json"
	"time"

	"github.com/asentric/asentric/pkg/asentric"
)

// AlertPayload is the JSON structure for alert delivery.
type AlertPayload struct {
	Rule        string                 `json:"rule"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Timestamp   string                 `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Ref         *RefPayload            `json:"ref,omitempty"`
	Context     *ContextPayload        `json:"context,omitempty"`
}

// RefPayload is the execution reference in JSON.
type RefPayload struct {
	TxHash      string `json:"txHash"`
	BlockNumber uint64 `json:"blockNumber"`
	LogIndex    int    `json:"logIndex"`
}

// ContextPayload contains context information for the alert.
type ContextPayload struct {
	ChainID     uint64 `json:"chainId,omitempty"`
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	TxHash      string `json:"txHash,omitempty"`
}

// Formatter formats alerts for delivery.
type Formatter struct {
	includeContext bool
}

// FormatterConfig holds formatter configuration.
type FormatterConfig struct {
	// IncludeContext adds context data to the payload
	IncludeContext bool
}

// NewFormatter creates a new formatter.
func NewFormatter(cfg FormatterConfig) *Formatter {
	return &Formatter{
		includeContext: cfg.IncludeContext,
	}
}

// DefaultFormatter returns a formatter with default settings.
func DefaultFormatter() *Formatter {
	return NewFormatter(FormatterConfig{
		IncludeContext: true,
	})
}

// Format converts an alert to AlertPayload.
func (f *Formatter) Format(ctx asentric.Context, alert *asentric.Alert) AlertPayload {
	payload := AlertPayload{
		Rule:        alert.Rule,
		Severity:    string(alert.Severity),
		Title:       alert.Title,
		Description: alert.Description,
		Timestamp:   alert.Timestamp.Format(time.RFC3339),
		Metadata:    alert.Metadata,
	}

	// Add execution reference if available
	if alert.Ref != nil {
		payload.Ref = &RefPayload{
			TxHash:      alert.Ref.TxHash,
			BlockNumber: alert.Ref.BlockNumber,
			LogIndex:    alert.Ref.LogIndex,
		}
	}

	// Add context if configured and available
	if f.includeContext && ctx != nil {
		tx := ctx.Tx()
		payload.Context = &ContextPayload{
			ChainID:     uint64(ctx.ChainID()),
			BlockNumber: tx.BlockNumber,
			TxHash:      string(tx.Hash),
		}
	}

	return payload
}

// FormatJSON converts an alert to JSON bytes.
func (f *Formatter) FormatJSON(ctx asentric.Context, alert *asentric.Alert) ([]byte, error) {
	payload := f.Format(ctx, alert)
	return json.Marshal(payload)
}

// FormatJSONPretty converts an alert to pretty-printed JSON bytes.
func (f *Formatter) FormatJSONPretty(ctx asentric.Context, alert *asentric.Alert) ([]byte, error) {
	payload := f.Format(ctx, alert)
	return json.MarshalIndent(payload, "", "  ")
}
