// Package tests contains contract compliance tests for all feature-flags implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// FeatureFlagsServiceContract runs contract compliance tests against any FeatureFlagsService implementation.
func FeatureFlagsServiceContract(t *testing.T, svc contracts.FeatureFlagsService) {
	t.Helper()

	t.Run("IsEnabled_UnknownFlag_ReturnsFalseNoError", func(t *testing.T) {
		// Unknown flags should default to false (safe default), not error
		enabled, err := svc.IsEnabled(context.Background(), "nonexistent_flag_xyz", contracts.FlagContext{})
		if err != nil {
			t.Fatalf("IsEnabled returned error for unknown flag: %v", err)
		}
		if enabled {
			t.Fatal("expected unknown flag to default to false")
		}
	})

	t.Run("GetAllFlags_ReturnsMap", func(t *testing.T) {
		flags, err := svc.GetAllFlags(context.Background(), contracts.FlagContext{})
		if err != nil {
			t.Fatalf("GetAllFlags returned error: %v", err)
		}
		if flags == nil {
			t.Fatal("expected non-nil map")
		}
	})

	t.Run("GetVariant_UnknownFlag_ReturnsDefaultNoError", func(t *testing.T) {
		_, err := svc.GetVariant(context.Background(), "nonexistent_variant_xyz", contracts.FlagContext{})
		if err != nil {
			t.Fatalf("GetVariant returned error for unknown flag: %v", err)
		}
	})
}
