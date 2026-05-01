import type { ICICDService, CICDConfig, CICDValidationResult } from '../../../contracts/ts/cicd'

/**
 * VercelService implements ICICDService for Vercel deployments (Bun runtime).
 */
export class VercelService implements ICICDService {
  async generateWorkflows(cfg: CICDConfig): Promise<Map<string, string>> {
    // TODO: generate files:
    //   vercel.json                               — build output config for Bun/Next.js
    //   .github/workflows/ci.yaml                — bun install + build + test + eslint on every PR
    //   .github/workflows/deploy-production.yaml — vercel --prod on push to main
    //
    // vercel.json shape (Next.js):
    //   { "framework": "nextjs", "buildCommand": "bun run build", "outputDirectory": ".next" }
    //
    // deploy-production.yaml steps:
    //   1. actions/checkout
    //   2. oven-sh/setup-bun
    //   3. bun install
    //   4. bun run build
    //   5. vercel pull --environment=production --token=$VERCEL_TOKEN
    //   6. vercel build --prod --token=$VERCEL_TOKEN
    //   7. vercel deploy --prebuilt --prod --token=$VERCEL_TOKEN
    //
    // Required GitHub secrets: VERCEL_TOKEN, VERCEL_ORG_ID, VERCEL_PROJECT_ID
    console.warn('[vercel] stub: generateWorkflows() not implemented')
    return new Map()
  }

  async validateWorkflows(projectRoot: string): Promise<CICDValidationResult> {
    // TODO: check that the following files exist:
    //   vercel.json
    //   .github/workflows/ci.yaml
    //   .github/workflows/deploy-production.yaml
    console.warn('[vercel] stub: validateWorkflows() not implemented')
    return { valid: true, missingWorkflows: [], errors: [] }
  }
}
