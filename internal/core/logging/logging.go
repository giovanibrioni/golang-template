// Package logging configures the process-wide slog logger.
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type requestIDKey struct{}

// WithRequestID stores the request id on the context for the slog handler.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID returns the request id stored by WithRequestID.
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok && id != ""
}

// Setup installs the default slog logger. JSON output is used when jsonLogs is true.
func Setup(level string, jsonLogs bool) {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	var handler slog.Handler
	if jsonLogs {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(contextHandler{Handler: handler}))
}

func parseLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if id, ok := RequestID(ctx); ok {
		record.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, record)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}
