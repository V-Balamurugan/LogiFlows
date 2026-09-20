package logger_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/logiflows/logiflows/backend/internal/logger"
)

func TestLogger_JSON_WithRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.InitWithWriter("production", "info", &buf)

	ctx := logger.ContextWithRequestID(context.Background(), "req-uuid-12345")
	log.InfoContext(ctx, "test event occurred", "user_id", 42)

	output := buf.String()
	if !strings.Contains(output, `"request_id":"req-uuid-12345"`) {
		t.Errorf("expected log output to contain request_id, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"test event occurred"`) {
		t.Errorf("expected log output to contain msg, got: %s", output)
	}
	if !strings.Contains(output, `"user_id":42`) {
		t.Errorf("expected log output to contain user_id, got: %s", output)
	}
}

func TestLogger_Text_WithoutRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.InitWithWriter("development", "debug", &buf)

	log.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected log output to contain 'debug message', got: %s", output)
	}
	if strings.Contains(output, "request_id") {
		t.Errorf("expected log output NOT to contain request_id when absent, got: %s", output)
	}
}
