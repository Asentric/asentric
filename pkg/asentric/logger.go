package asentric

import (
	"fmt"
	"log"
	"os"
	"time"
)

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

type DefaultLogger struct {
	debug bool
}

func NewDefaultLogger(debug bool) *DefaultLogger {
	return &DefaultLogger{debug: debug}
}

// Info logs an informational message.
func (l *DefaultLogger) Info(msg string, args ...any) {
	l.log("INFO", msg, args...)
}
func (l *DefaultLogger) Error(msg string, args ...any) {
	l.log("ERROR", msg, args...)
}

func (l *DefaultLogger) Debug(msg string, args ...any) {
	l.log("DEBUG", msg, args...)
}

func (l *DefaultLogger) log(level, msg string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf(msg, args...)
	log.Printf("[%s] %s: %s", timestamp, level, formatted)
}

// NopLogger is a no-op logger that discards all messages.
type NopLogger struct{}

// NewNopLogger creates a new no-op logger.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (l *NopLogger) Info(msg string, args ...any)  {}
func (l *NopLogger) Error(msg string, args ...any) {}
func (l *NopLogger) Debug(msg string, args ...any) {}

// stdLogger is the default logger used if none is provided
var stdLogger Logger = NewDefaultLogger(false)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	stdLogger = NewDefaultLogger(false)
}
