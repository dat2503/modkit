// Package otel implements ObservabilityService using log/slog for structured
// logging and no-op spans for tracing. Extend with the OpenTelemetry SDK for
// production OTLP export.
package otel

import (
	"context"
	"log/slog"
	"os"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the observability provider.
type Config struct {
	// ServiceName is the service name in logs and traces.
	ServiceName string

	// Endpoint is the OTLP exporter endpoint. Empty disables remote export.
	Endpoint string

	// LogLevel is the minimum log level (debug, info, warn, error). Default: info.
	LogLevel string

	// LogFormat is the log format (json or text). Default: json.
	LogFormat string
}

// Service implements contracts.ObservabilityService.
type Service struct {
	cfg    Config
	logger *slog.Logger
}

// New creates a new observability service backed by log/slog.
func New(cfg Config) (*Service, error) {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	if cfg.ServiceName != "" {
		logger = logger.With("service", cfg.ServiceName)
	}

	return &Service{cfg: cfg, logger: logger}, nil
}

// StartSpan starts a new trace span. Returns a no-op span.
func (s *Service) StartSpan(ctx context.Context, operationName string) (context.Context, contracts.Span) {
	return ctx, &nopSpan{}
}

// Log emits a structured log entry via slog.
func (s *Service) Log(ctx context.Context, level contracts.LogLevel, msg string, fields map[string]any) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	switch level {
	case contracts.LogLevelDebug:
		s.logger.DebugContext(ctx, msg, args...)
	case contracts.LogLevelWarn:
		s.logger.WarnContext(ctx, msg, args...)
	case contracts.LogLevelError:
		s.logger.ErrorContext(ctx, msg, args...)
	default:
		s.logger.InfoContext(ctx, msg, args...)
	}
}

// RecordMetric is a no-op in this implementation.
func (s *Service) RecordMetric(_ context.Context, _ string, _ float64, _ map[string]string) {}

// Shutdown is a no-op for this slog-based implementation.
func (s *Service) Shutdown(_ context.Context) error { return nil }

// nopSpan is a no-op implementation of contracts.Span.
type nopSpan struct{}

func (n *nopSpan) End()                           {}
func (n *nopSpan) SetAttribute(_ string, _ any)   {}
func (n *nopSpan) RecordError(_ error)             {}
