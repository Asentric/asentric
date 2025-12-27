package asentric

import (
	"fmt"
	"log"
	"os"
)

// Logger is the interface for logging within the Asentric runtime.
// Implementations can redirect logs to any destination.
type Logger interface {

	// Error logs an error message
	Error(msg string, args ...any)

	// Warn logs a warning message
	Warn(msg string, args ...any)

	// Info logs an informational message
	Info(msg string, args ...any)

	// Debug logs a debug message (may be no-op in production)
	Debug(msg string, args ...any)
}

type StdLogger struct {
	logger *log.Logger
	debug  bool
}

func NewStdLogger(logger *log.Logger, debug bool) *StdLogger {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}
	return &StdLogger{logger: logger, debug: debug}
}

func (s *StdLogger) logInternal(level string, msg string, args ...any) {
	// Format message
	formattedMsg := fmt.Sprintf(level+msg, args...)

	// Output message
	_ = s.logger.Output(3, formattedMsg)
}

func (s *StdLogger) Error(msg string, args ...any) {
	s.logInternal("[ERROR] ", msg, args...)
}

// Warn logs a warning message.
func (s *StdLogger) Warn(msg string, args ...any) {
	s.logInternal("[WARN] ", msg, args...)
}

// Info logs an informational message.
func (s *StdLogger) Info(msg string, args ...any) {
	s.logInternal("[INFO] ", msg, args...)
}

func (s *StdLogger) Debug(msg string, args ...any) {
	// If debug is disabled, do nothing
	if !s.debug {
		return
	}
	s.logInternal("[DEBUG] ", msg, args...)
}

// NopLogger is a no-op logger that discards all messages.
type NopLogger struct{}

// NewNopLogger creates a new no-op logger.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (n *NopLogger) Info(msg string, args ...any)  {}
func (n *NopLogger) Error(msg string, args ...any) {}
func (n *NopLogger) Debug(msg string, args ...any) {}
func (n *NopLogger) Warn(msg string, args ...any)  {}
