// Package tests contains contract compliance tests for all cache implementations.
package tests

import (
	"context"
	"testing"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// CacheServiceContract runs contract compliance tests against any CacheService implementation.
func CacheServiceContract(t *testing.T, svc contracts.CacheService) {
	t.Helper()

	t.Run("Set_ThenGet_ReturnsValue", func(t *testing.T) {
		err := svc.Set(context.Background(), "test:key", []byte("hello"), time.Minute)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, err := svc.Get(context.Background(), "test:key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if string(val) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(val))
		}
	})

	t.Run("Get_MissingKey_ReturnsError", func(t *testing.T) {
		_, err := svc.Get(context.Background(), "test:nonexistent")
		if err == nil {
			t.Fatal("expected error for missing key, got nil")
		}
	})

	t.Run("SetNX_WhenKeyAbsent_ReturnsTrue", func(t *testing.T) {
		svc.Delete(context.Background(), "test:nx")
		ok, err := svc.SetNX(context.Background(), "test:nx", []byte("1"), time.Minute)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if !ok {
			t.Fatal("expected SetNX to return true for absent key")
		}
	})

	t.Run("SetNX_WhenKeyPresent_ReturnsFalse", func(t *testing.T) {
		ok, err := svc.SetNX(context.Background(), "test:nx", []byte("2"), time.Minute)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if ok {
			t.Fatal("expected SetNX to return false for existing key")
		}
	})

	t.Run("Increment_ReturnsNewValue", func(t *testing.T) {
		svc.Delete(context.Background(), "test:counter")
		val, err := svc.Increment(context.Background(), "test:counter", 1)
		if err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
		if val != 1 {
			t.Fatalf("expected 1, got %d", val)
		}
		val, _ = svc.Increment(context.Background(), "test:counter", 5)
		if val != 6 {
			t.Fatalf("expected 6, got %d", val)
		}
	})
}
