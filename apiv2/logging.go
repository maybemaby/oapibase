package api

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
)

type LoggingFormat string

const (
	JSONFormat LoggingFormat = "json"
	TEXTFormat LoggingFormat = "text"
)

type WithLoggerService interface {
	WithLogger(logger *slog.Logger) WithLoggerService
}

func BootstrapLogger(level slog.Level, format LoggingFormat, colorize bool) *slog.Logger {
	handlers := []slog.Handler{}

	if format == "json" {
		handlers = append(handlers, slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	} else if format == "text" {
		if colorize {
			handlers = append(handlers, tint.NewHandler(os.Stderr, &tint.Options{
				Level:      level,
				TimeFormat: time.Kitchen,
			}))
		} else {
			handlers = append(handlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			}))
		}
	}

	handler := slogmulti.Fanout(handlers...)

	logger := slog.New(handler)

	return logger
}

var _ slog.Handler = (*NoOpHandler)(nil)

type NoOpHandler struct{}

// Enabled implements slog.Handler.
func (n *NoOpHandler) Enabled(context.Context, slog.Level) bool {
	return false
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
