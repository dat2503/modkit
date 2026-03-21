// Package tests contains contract compliance tests for all error-tracking implementations.
package tests

import (
	"context"
	"errors"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// ErrorTrackingServiceContract runs contract compliance tests against any ErrorTrackingService implementation.
func ErrorTrackingServiceContract(t *testing.T, svc contracts.ErrorTrackingService) {
	t.Helper()

	t.Run("CaptureError_DoesNotReturnError", func(t *testing.T) {
		err := svc.CaptureError(context.Background(), errors.New("test error"), contracts.CaptureOptions{
			Tags: map[string]string{"test": "true"},
		})
		if err != nil {
			t.Fatalf("CaptureError returned error: %v", err)
		}
	})

	t.Run("SetUser_ReturnsContextWithUser", func(t *testing.T) {
		ctx := svc.SetUser(context.Background(), contracts.ErrorUser{
			ID:    "user_test_123",
			Email: "test@example.com",
		})
		if ctx == nil {
			t.Fatal("expected non-nil context from SetUser")
		}
	})

	t.Run("Flush_DoesNotError", func(t *testing.T) {
		if err := svc.Flush(context.Background()); err != nil {
			t.Fatalf("Flush returned error: %v", err)
		}
	})
}
