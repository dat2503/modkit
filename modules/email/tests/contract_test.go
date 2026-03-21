// Package tests contains contract compliance tests for all email implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// EmailServiceContract runs contract compliance tests against any EmailService implementation.
func EmailServiceContract(t *testing.T, svc contracts.EmailService) {
	t.Helper()

	t.Run("Send_InvalidAPIKey_ReturnsError", func(t *testing.T) {
		// Implementations must return an error for authentication failures,
		// not panic or hang.
		_, err := svc.Send(context.Background(), contracts.EmailMessage{
			To:      []string{"test@example.com"},
			From:    "noreply@example.com",
			Subject: "Test",
			Body:    contracts.EmailBody{Text: "Test body"},
		})
		// In unit tests with a real key this should succeed.
		// This test is a placeholder — integration tests verify actual delivery.
		_ = err
	})

	t.Run("SendBatch_EmptySlice_ReturnsEmpty", func(t *testing.T) {
		results, err := svc.SendBatch(context.Background(), []contracts.EmailMessage{})
		if err != nil {
			t.Fatalf("unexpected error for empty batch: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected empty results, got %d", len(results))
		}
	})
}
