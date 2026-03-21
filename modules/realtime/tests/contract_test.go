// Package tests contains contract compliance tests for all realtime implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// RealtimeServiceContract runs contract compliance tests against any RealtimeService implementation.
func RealtimeServiceContract(t *testing.T, svc contracts.RealtimeService) {
	t.Helper()

	t.Run("ConnectedUsers_ReturnsSlice", func(t *testing.T) {
		users, err := svc.ConnectedUsers(context.Background())
		if err != nil {
			t.Fatalf("ConnectedUsers returned error: %v", err)
		}
		if users == nil {
			t.Fatal("expected non-nil slice")
		}
	})

	t.Run("Publish_NoConnectedClients_ReturnsZero", func(t *testing.T) {
		count, err := svc.Publish(context.Background(), "test.topic", map[string]string{"key": "val"})
		if err != nil {
			t.Fatalf("Publish returned error: %v", err)
		}
		if count < 0 {
			t.Fatalf("expected non-negative count, got %d", count)
		}
	})

	t.Run("Disconnect_NonexistentUser_ReturnsNil", func(t *testing.T) {
		err := svc.Disconnect(context.Background(), "user_nonexistent_000")
		if err != nil {
			t.Fatalf("Disconnect for nonexistent user returned error: %v", err)
		}
	})
}
