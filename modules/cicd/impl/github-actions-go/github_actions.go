// Package githubactions implements the CICDService interface for GitHub Actions (Go runtime).
package githubactions

import (
	"context"
	"log"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Service implements contracts.CICDService for GitHub Actions.
type Service struct{}

// New creates a new GitHub Actions CICD service.
func New() *Service {
	return &Service{}
}

// GenerateWorkflows generates GitHub Actions workflow files for a Go project.
func (s *Service) GenerateWorkflows(ctx context.Context, cfg contracts.CICDConfig) (map[string][]byte, error) {
	// TODO: implement using Go text/template to render workflow YAML files
	// Files to generate:
	//   .github/workflows/ci.yaml             — build + test + golangci-lint
	//   .github/workflows/deploy-staging.yaml — docker build + push + deploy on main
	//   .github/workflows/deploy-production.yaml — docker build + push + release on v* tag
	log.Printf("[github-actions] stub: GenerateWorkflows() not implemented")
	return nil, nil
}

// ValidateWorkflows checks that all required workflows exist in the project.
func (s *Service) ValidateWorkflows(ctx context.Context, projectRoot string) (*contracts.CICDValidationResult, error) {
	// TODO: check that .github/workflows/{ci,deploy-staging,deploy-production}.yaml exist and are valid YAML
	log.Printf("[github-actions] stub: ValidateWorkflows() not implemented")
	return &contracts.CICDValidationResult{Valid: true}, nil
}
