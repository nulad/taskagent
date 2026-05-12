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
