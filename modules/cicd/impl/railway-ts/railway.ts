import type { ICICDService, CICDConfig, CICDValidationResult } from '../../../contracts/ts/cicd'

/**
 * RailwayService implements ICICDService for Railway deployments (Bun runtime).
 */
export class RailwayService implements ICICDService {
  async generateWorkflows(cfg: CICDConfig): Promise<Map<string, string>> {
    // TODO: generate files:
    //   railway.toml                              — service config (healthcheck, restart policy, build)
    //   .github/workflows/ci.yaml                — bun install + build + test + eslint on every PR
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
    //   2. oven-sh/setup-bun
    //   3. bun install && bun run build
    //   4. Install Railway CLI: npm install -g @railway/cli
    //   5. railway up --service={cfg.projectName} --detach
    //
    // Required GitHub secrets: RAILWAY_TOKEN
    // Note: Postgres and Redis are linked as Railway services in the dashboard, not provisioned here.
    console.warn('[railway] stub: generateWorkflows() not implemented')
    return new Map()
  }

  async validateWorkflows(projectRoot: string): Promise<CICDValidationResult> {
    // TODO: check that the following files exist:
    //   railway.toml
    //   .github/workflows/ci.yaml
    //   .github/workflows/deploy-production.yaml
    console.warn('[railway] stub: validateWorkflows() not implemented')
    return { valid: true, missingWorkflows: [], errors: [] }
  }
}
