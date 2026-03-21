import type { ICICDService, CICDConfig, CICDValidationResult } from '../../../contracts/ts/cicd'

/**
 * GitHubActionsService implements ICICDService for GitHub Actions (Bun runtime).
 */
export class GitHubActionsService implements ICICDService {
  async generateWorkflows(cfg: CICDConfig): Promise<Map<string, string>> {
    // TODO: generate GitHub Actions workflow YAML files for Bun runtime:
    //   .github/workflows/ci.yaml             — bun install + build + test + eslint
    //   .github/workflows/deploy-staging.yaml — docker build + push + deploy on main
    //   .github/workflows/deploy-production.yaml — docker build + push + release on v* tag
    throw new Error('not implemented')
  }

  async validateWorkflows(projectRoot: string): Promise<CICDValidationResult> {
    // TODO: check that .github/workflows/{ci,deploy-staging,deploy-production}.yaml exist and are valid YAML
    throw new Error('not implemented')
  }
}
