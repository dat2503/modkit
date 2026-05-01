// Package vercel implements the CICDService interface for Vercel deployments (Go runtime).
package vercel

import (
	"context"
	"log"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Service implements contracts.CICDService for Vercel.
type Service struct{}

// New creates a new Vercel CICD service.
func New() *Service {
	return &Service{}
}

// GenerateWorkflows generates Vercel deployment config and GitHub Actions CI workflow.
func (s *Service) GenerateWorkflows(ctx context.Context, cfg contracts.CICDConfig) (map[string][]byte, error) {
	// TODO: implement using Go text/template to render config and workflow files.
	// Files to generate:
	//   vercel.json                               — build output config, route rewrites for Go serverless
	//   .github/workflows/ci.yaml                — go build + go test + golangci-lint on every PR
	//   .github/workflows/deploy-production.yaml — vercel --prod on push to main
	//
	// vercel.json shape:
	//   { "builds": [{"src": "apps/api/main.go", "use": "@vercel/go"}],
	//     "routes": [{"src": "/api/(.*)", "dest": "apps/api/main.go"}] }
	//
	// deploy-production.yaml steps:
	//   1. actions/checkout
	//   2. actions/setup-go
	//   3. go build ./...
	//   4. vercel pull --environment=production --token=$VERCEL_TOKEN
	//   5. vercel build --prod --token=$VERCEL_TOKEN
	//   6. vercel deploy --prebuilt --prod --token=$VERCEL_TOKEN
	//
	// Required GitHub secrets: VERCEL_TOKEN, VERCEL_ORG_ID, VERCEL_PROJECT_ID
	log.Printf("[vercel] stub: GenerateWorkflows() not implemented")
	return nil, nil
}

// ValidateWorkflows checks that vercel.json and required workflow files exist.
func (s *Service) ValidateWorkflows(ctx context.Context, projectRoot string) (*contracts.CICDValidationResult, error) {
	// TODO: check that the following files exist:
	//   vercel.json
	//   .github/workflows/ci.yaml
	//   .github/workflows/deploy-production.yaml
	log.Printf("[vercel] stub: ValidateWorkflows() not implemented")
	return &contracts.CICDValidationResult{Valid: true}, nil
}
