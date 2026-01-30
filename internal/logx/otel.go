package logx

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/bpva/ad-marketplace/internal/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
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

func NewLogger(
	ctx context.Context,
	cfg config.Logger,
	env string,
) (*slog.Logger, func(context.Context) error, error) {
	level := parseLevel(cfg.Level)
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	stdoutLogger := slog.New(stdoutHandler).With("env", env)

	if !cfg.OTLPEnabled {
		return stdoutLogger, func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("app"),
			semconv.DeploymentEnvironment(env),
		),
	)
	if err != nil {
		stdoutLogger.Error("failed to create otel resource", "error", err)
		return nil, func(context.Context) error { return nil }, err
	}

	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		stdoutLogger.Error("failed to create otel exporter", "error", err)
		return nil, func(context.Context) error { return nil }, err
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	)

	otelHandler := otelslog.NewHandler("app", otelslog.WithLoggerProvider(provider))

	multi := &multiHandler{handlers: []slog.Handler{stdoutHandler, otelHandler}}

	return slog.New(multi).With("env", env), provider.Shutdown, nil
}
