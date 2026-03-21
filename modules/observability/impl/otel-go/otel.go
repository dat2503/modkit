// Package otel implements the ObservabilityService interface using OpenTelemetry.
package otel

import (
	"context"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the OpenTelemetry observability provider.
type Config struct {
	// ServiceName is the service name as it appears in traces and metrics.
	ServiceName string

	// Endpoint is the OTLP exporter endpoint. Empty to disable export (stdout only).
	Endpoint string

	// Headers are additional headers for authenticated OTLP endpoints.
	Headers map[string]string

	// LogLevel is the minimum log level to emit (debug, info, warn, error).
	LogLevel string

	// LogFormat is the log output format (json or text).
	LogFormat string
}

// Service implements contracts.ObservabilityService using OpenTelemetry.
type Service struct {
	cfg Config
	// TODO: add otel tracer, meter, and logger providers
}

// New creates a new OpenTelemetry observability service.
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// StartSpan starts a new trace span.
func (s *Service) StartSpan(ctx context.Context, operationName string) (context.Context, contracts.Span) {
	// TODO: implement using go.opentelemetry.io/otel tracer.Start(ctx, operationName)
	panic("not implemented")
}

// Log emits a structured log entry.
func (s *Service) Log(ctx context.Context, level contracts.LogLevel, msg string, fields map[string]any) {
	// TODO: implement using log/slog with OTLP log exporter
	panic("not implemented")
}

// RecordMetric records a numeric metric.
func (s *Service) RecordMetric(ctx context.Context, name string, value float64, labels map[string]string) {
	// TODO: implement using otel meter
	panic("not implemented")
}

// Shutdown flushes pending telemetry and releases resources.
func (s *Service) Shutdown(ctx context.Context) error {
	// TODO: implement graceful shutdown of tracer, meter, and logger providers
	panic("not implemented")
}
