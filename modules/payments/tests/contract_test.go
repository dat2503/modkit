// Package tests contains contract compliance tests for all payments implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// PaymentsServiceContract runs contract compliance tests against any PaymentsService implementation.
func PaymentsServiceContract(t *testing.T, svc contracts.PaymentsService) {
	t.Helper()

	t.Run("ConstructWebhookEvent_InvalidSignature_ReturnsError", func(t *testing.T) {
		_, err := svc.ConstructWebhookEvent([]byte(`{"type":"test"}`), "invalid-sig")
		if err == nil {
			t.Fatal("expected error for invalid webhook signature, got nil")
		}
	})

	t.Run("GetCheckoutSession_NonexistentID_ReturnsError", func(t *testing.T) {
		_, err := svc.GetCheckoutSession(context.Background(), "cs_nonexistent_000")
		if err == nil {
			t.Fatal("expected error for nonexistent session, got nil")
		}
	})

	t.Run("GetCustomer_NonexistentID_ReturnsError", func(t *testing.T) {
		_, err := svc.GetCustomer(context.Background(), "cus_nonexistent_000")
		if err == nil {
			t.Fatal("expected error for nonexistent customer, got nil")
		}
	})
}
