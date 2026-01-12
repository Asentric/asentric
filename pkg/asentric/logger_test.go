package asentric_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/asentric/asentric/pkg/asentric"
)

func TestNewStdLogger_DefaultLogger(t *testing.T) {
	logger := asentric.NewStdLogger(nil, false)

	// should not panic
	logger.Info("hello")
}

func TestStdLogger_Error_WithFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := log.New(&buf, "", 0)
	logger := asentric.NewStdLogger(l, false)

	logger.Error("failed at step %d", 3)

	if !strings.Contains(buf.String(), "[ERROR] failed at step 3") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestStdLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	l := log.New(&buf, "", 0)
	logger := asentric.NewStdLogger(l, false)

	logger.Warn("disk almost full")

	if !strings.Contains(buf.String(), "[WARN] disk almost full") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestStdLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	l := log.New(&buf, "", 0)
	logger := asentric.NewStdLogger(l, false)

	logger.Info("hello %s", "world")

	expected := "[INFO] hello world"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestStdLogger_Debug(t *testing.T) {
	t.Run("should log when debug is true", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.New(&buf, "", 0)
		logger := asentric.NewStdLogger(l, true)

		logger.Debug("test debug")
		if !strings.Contains(buf.String(), "[DEBUG]") {
			t.Error("expected debug message to be logged")
		}
	})

	t.Run("should NOT log when debug is false", func(t *testing.T) {
		var buf bytes.Buffer
		l := log.New(&buf, "", 0)
		logger := asentric.NewStdLogger(l, false)

		logger.Debug("test debug")
		if buf.Len() > 0 {
			t.Error("expected no log output when debug is disabled")
		}
	})
}

func TestNopLogger(t *testing.T) {
	logger := asentric.NewNopLogger()
	logger.Info("test")
	logger.Error("test")
	logger.Debug("test")
	logger.Warn("test")

}
