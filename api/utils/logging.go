package utils

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*NoOpHandler)(nil)

type NoOpHandler struct{}

// Enabled implements slog.Handler.
func (n *NoOpHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (n *NoOpHandler) Handle(context.Context, slog.Record) error {
	return nil
}

// WithAttrs implements slog.Handler.
func (n *NoOpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

// WithGroup implements slog.Handler.
func (n *NoOpHandler) WithGroup(name string) slog.Handler {
	return n
}

func NewNoOpLogger() *slog.Logger {
	return slog.New(&NoOpHandler{})
}

func LoggerWithOrNoOp(logger *slog.Logger, with ...any) *slog.Logger {
	if logger == nil {
		return NewNoOpLogger()
	}

	return logger.With(with...)
}
