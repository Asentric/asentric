package asentric

import "context"

// AlertSink defines how alerts are delivered to destinations.
// Implementations handle the actual delivery mechanism (webhook, console, etc.)
type AlertSink interface {
	// Emit sends an alert to the configured destination.
	// ctx is used for cancellation and timeouts.
	// evalCtx provides the evaluation context for additional alert metadata.
	Emit(ctx context.Context, evalCtx Context, alert *Alert) error
}

