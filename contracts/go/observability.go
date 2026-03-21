package contracts

import "context"

// ObservabilityService provides distributed tracing, structured logging, and metrics.
// Must be initialized first, before all other modules.
// Wrap all module calls and HTTP handlers with spans for end-to-end tracing.
type ObservabilityService interface {
	// StartSpan starts a new trace span with the given operation name.
	// The returned context carries the span — always pass it to downstream calls.
	// Call span.End() when the operation is complete (typically via defer).
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)

	// Log emits a structured log entry at the given level.
	Log(ctx context.Context, level LogLevel, msg string, fields map[string]any)

	// RecordMetric records a numeric metric value with optional labels.
	RecordMetric(ctx context.Context, name string, value float64, labels map[string]string)

	// Shutdown flushes all pending telemetry and releases resources.
	// Call this during graceful shutdown.
	Shutdown(ctx context.Context) error
}

// Span represents a single unit of work in a distributed trace.
type Span interface {
	// End marks the span as complete.
	End()

	// SetAttribute attaches a key-value attribute to this span.
	SetAttribute(key string, value any)

	// RecordError records an error on this span and marks it as failed.
	RecordError(err error)
}

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)
