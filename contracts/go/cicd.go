package contracts

import "context"

// CICDService generates and manages CI/CD pipeline configuration.
// This module produces workflow files (e.g. GitHub Actions YAML) — it does not
// execute pipelines directly. Generated automatically by `modkit init`.
type CICDService interface {
	// GenerateWorkflows generates CI/CD workflow files for the project.
	// Returns a map of filename → file content for each generated workflow.
	GenerateWorkflows(ctx context.Context, cfg CICDConfig) (map[string][]byte, error)

	// ValidateWorkflows checks that all required workflows exist and are syntactically valid.
	ValidateWorkflows(ctx context.Context, projectRoot string) (*CICDValidationResult, error)
}

// CICDConfig describes the project for which to generate workflows.
type CICDConfig struct {
	// ProjectName is the name of the project.
	ProjectName string

	// Runtime is the backend runtime ("go" or "bun").
	Runtime string

	// GoVersion is the Go version to use in workflows (Go runtime only).
	GoVersion string

	// BunVersion is the Bun version to use in workflows (Bun runtime only).
	BunVersion string

	// DockerRegistry is the container registry to push images to (e.g. "ghcr.io/org/project").
	DockerRegistry string

	// DeployEnvs lists the deployment environments (e.g. ["staging", "production"]).
	DeployEnvs []string

	// DeployTarget identifies the deployment platform ("github-actions", "vercel", "railway").
	// Defaults to "github-actions" if empty.
	DeployTarget string

	// VercelOrgID is the Vercel organization ID (vercel impl only).
	VercelOrgID string

	// VercelProjectID is the Vercel project ID (vercel impl only).
	VercelProjectID string

	// RailwayProjectID is the Railway project ID (railway impl only).
	RailwayProjectID string
}

// CICDValidationResult is the result of ValidateWorkflows.
type CICDValidationResult struct {
	// Valid is true if all required workflows are present and valid.
	Valid bool

	// MissingWorkflows lists any required workflows that are absent.
	MissingWorkflows []string

	// Errors lists any syntax or schema errors found.
	Errors []string
}
