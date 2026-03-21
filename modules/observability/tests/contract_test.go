// Package tests contains contract compliance tests for all observability implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// ObservabilityServiceContract runs contract compliance tests against any ObservabilityService implementation.
func ObservabilityServiceContract(t *testing.T, svc contracts.ObservabilityService) {
	t.Helper()

	t.Run("StartSpan_ReturnsNonNilSpanAndContext", func(t *testing.T) {
		ctx, span := svc.StartSpan(context.Background(), "test.operation")
		if span == nil {
			t.Fatal("expected non-nil Span")
		}
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		span.End()
	})

	t.Run("StartSpan_SetAttribute_DoesNotPanic", func(t *testing.T) {
		_, span := svc.StartSpan(context.Background(), "test.attrs")
		defer span.End()
		span.SetAttribute("key", "value")
		span.SetAttribute("number", 42)
	})

	t.Run("StartSpan_RecordError_DoesNotPanic", func(t *testing.T) {
		_, span := svc.StartSpan(context.Background(), "test.error")
		defer span.End()
		span.RecordError(context.DeadlineExceeded)
	})

	t.Run("Log_AllLevels_DoesNotPanic", func(t *testing.T) {
		ctx := context.Background()
		svc.Log(ctx, contracts.LogLevelDebug, "debug message", nil)
		svc.Log(ctx, contracts.LogLevelInfo, "info message", map[string]any{"key": "val"})
		svc.Log(ctx, contracts.LogLevelWarn, "warn message", nil)
		svc.Log(ctx, contracts.LogLevelError, "error message", nil)
	})

	t.Run("Shutdown_DoesNotError", func(t *testing.T) {
		if err := svc.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown returned error: %v", err)
		}
	})
}
