# CI/CD Module — Agent Instructions

## When to use

This module is automatically included by `modkit init` for every project. You do not select it separately — it is always generated. You select the *implementation* using `--modules cicd:<impl>`.

## Choosing an implementation

| Implementation | Best for | Infra model | Secrets needed |
|---------------|---------|------------|---------------|
| `github-actions` (default) | Self-hosted or Docker-based deploys | Docker + your own servers | `DOCKER_REGISTRY_TOKEN`, `STAGING_DEPLOY_KEY`, `PRODUCTION_DEPLOY_KEY` |
| `vercel` | Bun/Next.js frontends, serverless Go | Managed (Vercel platform) | `VERCEL_TOKEN`, `VERCEL_ORG_ID`, `VERCEL_PROJECT_ID` |
| `railway` | Full-stack with managed Postgres + Redis | Managed (Railway platform) | `RAILWAY_TOKEN` |

If the user does not specify, default to `github-actions`.

---

## github-actions

Generates three GitHub Actions workflows into `.github/workflows/`:

1. `ci.yaml` — builds, tests, and lints on every PR
2. `deploy-staging.yaml` — deploys to staging on merge to `main`
3. `deploy-production.yaml` — deploys to production on version tag `v*`

**CI steps (Go):** checkout → setup-go → `go build ./...` → `go test ./...` → golangci-lint
**CI steps (Bun):** checkout → setup-bun → `bun install` → `bun build` → `bun test` → eslint

**Required GitHub secrets:**
```
DOCKER_REGISTRY_TOKEN    # container registry auth
STAGING_DEPLOY_KEY       # SSH key or deploy token for staging
PRODUCTION_DEPLOY_KEY    # SSH key or deploy token for production
STAGING_URL              # base URL for smoke tests
```

**Deploy to staging:**
```bash
git push origin main    # triggers deploy-staging.yaml automatically
```

**Deploy to production:**
```bash
git tag v1.0.0
git push origin v1.0.0  # triggers deploy-production.yaml
```

---

## vercel

Generates:
- `vercel.json` — build and route config
- `.github/workflows/ci.yaml` — build + test on every PR
- `.github/workflows/deploy-production.yaml` — Vercel CLI deploy on push to `main`

No staging workflow — Vercel automatically creates preview deployments for every PR.

**Required GitHub secrets:**
```
VERCEL_TOKEN        # Vercel personal access token
VERCEL_ORG_ID       # from `vercel whoami` or dashboard
VERCEL_PROJECT_ID   # from .vercel/project.json after first `vercel link`
```

**One-time setup:**
```bash
cd apps/web
npx vercel link     # links project, writes .vercel/project.json
# commit .vercel/project.json — it contains VERCEL_ORG_ID and VERCEL_PROJECT_ID
```

**Deploy:**
```bash
git push origin main    # triggers deploy-production.yaml automatically
```

**Note for Go runtime:** uses `@vercel/go` builder. API routes must be in `apps/api/*.go` and conform to Vercel's Go serverless function signature.

---

## railway

Generates:
- `railway.toml` — service config (healthcheck path, restart policy, build settings)
- `.github/workflows/ci.yaml` — build + test on every PR
- `.github/workflows/deploy-production.yaml` — Railway CLI deploy on push to `main`

**Required GitHub secrets:**
```
RAILWAY_TOKEN    # from Railway dashboard → Account Settings → Tokens
```

**One-time setup (Railway dashboard):**
1. Create a new project in Railway
2. Add a Postgres service (Railway provisions it automatically)
3. Add a Redis service (Railway provisions it automatically)
4. Link your GitHub repo to the Railway service
5. Copy the service environment variables (`DATABASE_URL`, `REDIS_URL`) — Railway injects them automatically at runtime

**Deploy:**
```bash
git push origin main    # triggers deploy-production.yaml automatically
```

**Note:** Railway injects `DATABASE_URL` and `REDIS_URL` at runtime from linked services. Do not set these manually in `.env` for Railway deploys.

---

## Compliance Posture (§28)

The `COMPLIANCE_POSTURE` config field controls which additional CI/CD files are generated beyond the §24/§25 baseline. Posture is auto-suggested from project signals (§28.2) and confirmed at Phase 1.

### Posture: `solo` (team of 1, prototype, <100 users)

No additional files. §24/§25 baseline is sufficient:
- gosec/semgrep + govulncheck/npm audit + gitleaks in CI (already in §25)
- Conventional Commits + CHANGELOG (already in §24)

### Posture: `startup` (small team or public users)

Generated from `modules/cicd/templates/startup/`:

| File | Purpose |
|------|---------|
| `.github/dependabot.yml` | Auto-update PRs for gomod/npm/github-actions weekly |
| `.github/PULL_REQUEST_TEMPLATE.md` | Lightweight PR checklist |
| `codecov.yml` | 60% coverage threshold gate |
| `scripts/setup-branch-protection.sh` | Branch protection via `gh` CLI — run once after repo setup |

CI job additions (added to `ci.yaml`):
```yaml
- name: Container image scan
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.IMAGE_TAG }}
    severity: HIGH,CRITICAL
    exit-code: 1

- name: Upload coverage
  uses: codecov/codecov-action@v4
```

### Posture: `enterprise` (regulated, B2B SaaS, enterprise customers)

Everything in `startup`, plus generated from `modules/cicd/templates/enterprise/`:

| File | Purpose |
|------|---------|
| `.github/CODEOWNERS` | Required reviewers per directory |
| `.github/ISSUE_TEMPLATE/bug_report.md` | Structured bug report |
| `.github/ISSUE_TEMPLATE/feature_request.md` | Structured feature request |
| `codecov.yml` | 80% coverage threshold (overrides startup version) |

CI/deploy job additions:
```yaml
# In ci.yaml:
- name: Generate SBOM
  uses: anchore/sbom-action@v0
  with:
    image: ${{ env.IMAGE_TAG }}
    format: spdx-json
    output-file: sbom.spdx.json

- name: License compliance
  run: |
    go install github.com/google/go-licenses@latest
    go-licenses check ./... --disallowed_types=restricted

# In deploy-production.yaml:
- name: Sign container image
  uses: sigstore/cosign-installer@v3
  run: cosign sign --yes ${{ env.IMAGE_TAG }}
```

### Posture promotion

Posture changes follow §21.6 evolution protocol — the agent documents the trigger, proposes the upgrade, human approves. New template files are added, no existing files removed.

---

## Integration spec

After scaffold, verify with:

1. Push the project to a GitHub repository
2. Open a PR — the `ci.yaml` workflow should trigger and pass on the scaffold's default code
3. Merge to `main` — the deploy workflow should trigger (may fail without secrets configured, but must trigger)
4. Confirm generated files exist:
   - `github-actions`: `.github/workflows/{ci,deploy-staging,deploy-production}.yaml`
   - `vercel`: `vercel.json` + `.github/workflows/{ci,deploy-production}.yaml`
   - `railway`: `railway.toml` + `.github/workflows/{ci,deploy-production}.yaml`

## Do NOT

- Deploy to production without staging approval (`github-actions` impl)
- Store deployment credentials in `.env` — use GitHub Actions secrets for all of them
- Skip CI on PRs — never merge with failing tests
- Manually edit generated workflow files — regenerate via `modkit` if changes are needed
