/**
 * ICICDService generates and manages CI/CD pipeline configuration.
 * This module produces workflow files (e.g. GitHub Actions YAML) — it does not
 * execute pipelines directly. Generated automatically by `modkit init`.
 */
export interface ICICDService {
  /**
   * Generates CI/CD workflow files for the project.
   * Returns a map of filename → file content for each generated workflow.
   */
  generateWorkflows(cfg: CICDConfig): Promise<Map<string, string>>;

  /**
   * Checks that all required workflows exist and are syntactically valid.
   */
  validateWorkflows(projectRoot: string): Promise<CICDValidationResult>;
}

/** Describes the project for which to generate workflows. */
export interface CICDConfig {
  /** Name of the project. */
  projectName: string;

  /** Backend runtime: "go" or "bun". */
  runtime: 'go' | 'bun';

  /** Go version to use in workflows (Go runtime only). */
  goVersion?: string;

  /** Bun version to use in workflows (Bun runtime only). */
  bunVersion?: string;

  /** Container registry to push images to (e.g. "ghcr.io/org/project"). */
  dockerRegistry?: string;

  /** Deployment environments (e.g. ["staging", "production"]). */
  deployEnvs?: string[];
}

/** Result of validateWorkflows. */
export interface CICDValidationResult {
  /** True if all required workflows are present and valid. */
  valid: boolean;

  /** Required workflows that are absent. */
  missingWorkflows: string[];

  /** Syntax or schema errors found. */
  errors: string[];
}
