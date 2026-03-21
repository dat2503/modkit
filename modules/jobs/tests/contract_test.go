// Package tests contains contract compliance tests for all jobs implementations.
package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// JobsServiceContract runs contract compliance tests against any JobsService implementation.
func JobsServiceContract(t *testing.T, svc contracts.JobsService) {
	t.Helper()

	t.Run("RegisterHandler_ThenEnqueue_HandlerIsCalled", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		called := false

		err := svc.RegisterHandler("test:contract", func(ctx context.Context, payload []byte) error {
			called = true
			wg.Done()
			return nil
		})
		if err != nil {
			t.Fatalf("RegisterHandler failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() { svc.Start(ctx) }()
		defer svc.Stop(context.Background())

		_, err = svc.Enqueue(ctx, "test:contract", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}

		wg.Wait()
		if !called {
			t.Fatal("expected handler to be called")
		}
	})

	t.Run("Enqueue_ReturnsNonEmptyHandle", func(t *testing.T) {
		handle, err := svc.Enqueue(context.Background(), "test:noop", nil)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
		if handle == nil || handle.ID == "" {
			t.Fatal("expected non-empty job handle")
		}
	})
}
