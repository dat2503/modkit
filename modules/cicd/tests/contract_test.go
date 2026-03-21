// Package tests contains contract compliance tests for all cicd implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// CICDServiceContract runs contract compliance tests against any CICDService implementation.
func CICDServiceContract(t *testing.T, svc contracts.CICDService) {
	t.Helper()

	t.Run("GenerateWorkflows_GoRuntime_ReturnsThreeFiles", func(t *testing.T) {
		files, err := svc.GenerateWorkflows(context.Background(), contracts.CICDConfig{
			ProjectName: "test-project",
			Runtime:     "go",
			GoVersion:   "1.22",
		})
		if err != nil {
			t.Fatalf("GenerateWorkflows failed: %v", err)
		}
		required := []string{
			".github/workflows/ci.yaml",
			".github/workflows/deploy-staging.yaml",
			".github/workflows/deploy-production.yaml",
		}
		for _, name := range required {
			if _, ok := files[name]; !ok {
				t.Errorf("expected workflow file %q to be generated", name)
			}
		}
	})

	t.Run("GenerateWorkflows_BunRuntime_ReturnsThreeFiles", func(t *testing.T) {
		files, err := svc.GenerateWorkflows(context.Background(), contracts.CICDConfig{
			ProjectName: "test-project",
			Runtime:     "bun",
			BunVersion:  "1.1",
		})
		if err != nil {
			t.Fatalf("GenerateWorkflows failed: %v", err)
		}
		if len(files) < 3 {
			t.Fatalf("expected at least 3 workflow files, got %d", len(files))
		}
	})
}
