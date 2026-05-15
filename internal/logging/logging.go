package logging

import (
	"context"
	"log/slog"
)

type requestID struct{}

var ctxRequestID = requestID{}

func WithRequestID(ctx context.Context, id string) context.Context {
	ctx = context.WithValue(ctx, ctxRequestID, id)
	return ctx
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxRequestID).(string)
	return id, ok
}

// LogWithError adds request_id to log attributes if present in context
func LogWithError(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	if reqID, ok := RequestIDFromContext(ctx); ok {
		attrs = append(attrs, slog.String("request_id", reqID))
	}
	attrs = append(attrs, slog.Any("error", err))
	logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// ParseLevel parses log level string
func ParseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
