package asentric

// Logger is the interface for logging within the Asentric runtime.
// Implementations can redirect logs to any destination.
type Logger interface {
	// Info logs an informational message
	Info(msg string, args ...any)

	// Error logs an error message
	Error(msg string, args ...any)

	// Debug logs a debug message (may be no-op in production)
	Debug(msg string, args ...any)
}
