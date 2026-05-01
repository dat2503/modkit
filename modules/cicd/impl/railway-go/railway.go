// Package railway implements the CICDService interface for Railway deployments (Go runtime).
package railway

import (
	"context"
	"log"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Service implements contracts.CICDService for Railway.
type Service struct{}

// New creates a new Railway CICD service.
func New() *Service {
	return &Service{}
}

// GenerateWorkflows generates Railway deployment config and GitHub Actions CI workflow.
func (s *Service) GenerateWorkflows(ctx context.Context, cfg contracts.CICDConfig) (map[string][]byte, error) {
	// TODO: implement using Go text/template to render config and workflow files.
	// Files to generate:
	//   railway.toml                              — service config (healthcheck, restart policy, build)
	//   .github/workflows/ci.yaml                — go build + go test + golangci-lint on every PR
	//   .github/workflows/deploy-production.yaml — railway up on push to main
	//
	// railway.toml shape:
	//   [build]
	//     builder = "DOCKERFILE"
	//     dockerfilePath = "apps/api/Dockerfile"
	//   [deploy]
	//     healthcheckPath = "/health"
	//     restartPolicyType = "ON_FAILURE"
	//     restartPolicyMaxRetries = 3
	//
	// deploy-production.yaml steps:
	//   1. actions/checkout
	//   2. actions/setup-go
	//   3. go build ./...
	//   4. Install Railway CLI: npm install -g @railway/cli
	//   5. railway up --service={cfg.ProjectName} --detach
	//
	// Required GitHub secrets: RAILWAY_TOKEN
	// Note: Postgres and Redis are linked as Railway services in the dashboard, not provisioned here.
	log.Printf("[railway] stub: GenerateWorkflows() not implemented")
	return nil, nil
}

// ValidateWorkflows checks that railway.toml and required workflow files exist.
func (s *Service) ValidateWorkflows(ctx context.Context, projectRoot string) (*contracts.CICDValidationResult, error) {
	// TODO: check that the following files exist:
	//   railway.toml
	//   .github/workflows/ci.yaml
	//   .github/workflows/deploy-production.yaml
	log.Printf("[railway] stub: ValidateWorkflows() not implemented")
	return &contracts.CICDValidationResult{Valid: true}, nil
}
