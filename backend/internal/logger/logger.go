package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// ContextWithRequestID attaches a request ID to the given context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from the context, or returns an empty string.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}

// requestIDHandler is a custom slog.Handler that automatically appends request_id
// from context to log records.
type requestIDHandler struct {
	slog.Handler
}

func (h *requestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID := RequestIDFromContext(ctx); reqID != "" {
		r.AddAttrs(slog.String("request_id", reqID))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithGroup(name)}
}

// Init creates and registers a structured slog.Logger based on environment and level.
func Init(env, levelStr string) *slog.Logger {
	return InitWithWriter(env, levelStr, os.Stdout)
}

// InitWithWriter creates a logger outputting to the specified writer (useful for tests).
func InitWithWriter(env, levelStr string, w io.Writer) *slog.Logger {
	level := parseLevel(levelStr)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var baseHandler slog.Handler
	if strings.ToLower(env) == "production" || strings.ToLower(env) == "staging" {
		baseHandler = slog.NewJSONHandler(w, opts)
	} else {
		// Pretty text format for local development and test
		baseHandler = slog.NewTextHandler(w, opts)
	}

	wrappedHandler := &requestIDHandler{Handler: baseHandler}
	logger := slog.New(wrappedHandler)
	slog.SetDefault(logger)

	return logger
}

func parseLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
