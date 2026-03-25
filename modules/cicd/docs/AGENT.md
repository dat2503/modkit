# CI/CD Module — Agent Instructions

## When to use

This module is automatically included by `modkit init` for every project. You do not select it — it is always generated.

The module generates three GitHub Actions workflows into `.github/workflows/`:
1. `ci.yaml` — builds, tests, and lints on every PR
2. `deploy-staging.yaml` — deploys to staging on merge to `main`
3. `deploy-production.yaml` — deploys to production on version tag `v*`

## What gets generated

### `ci.yaml`

Runs on: `push` to any branch, `pull_request` to `main`

Steps (Go runtime):
1. `actions/checkout`
2. `actions/setup-go` with version from `registry.yaml`
3. `go build ./...`
4. `go test ./...`
5. `golangci-lint run`

Steps (Bun runtime):
1. `actions/checkout`
2. `oven-sh/setup-bun` with version from `registry.yaml`
3. `bun install`
4. `bun build`
5. `bun test`
6. `bunx eslint . --ext .ts`

### `deploy-staging.yaml`

Runs on: `push` to `main`

Steps:
1. Build Docker image and push to `DOCKER_REGISTRY`
2. Deploy to staging environment (SSH deploy or container registry update)
3. Run smoke tests against staging URL
4. Notify on failure

### `deploy-production.yaml`

Runs on: `push` tag matching `v*`

Steps:
1. Build and push production Docker image (tagged with git tag)
2. Deploy to production environment
3. Create GitHub Release with changelog
4. Notify on success/failure

## Required GitHub secrets

These must be set in GitHub → Settings → Secrets and variables → Actions:

```
DOCKER_REGISTRY_TOKEN    # container registry authentication
STAGING_DEPLOY_KEY       # SSH key or deploy token for staging
PRODUCTION_DEPLOY_KEY    # SSH key or deploy token for production
STAGING_URL              # base URL for smoke tests
```

## Deploy to staging workflow (Phase 5 of playbook)

```bash
# Triggers automatically on push to main
git push origin main
# Watch: GitHub Actions → deploy-staging.yaml
```

## Deploy to production (Phase 5 of playbook)

```bash
# Only after staging is approved:
git tag v1.0.0
git push origin v1.0.0
# Watch: GitHub Actions → deploy-production.yaml
```

## Integration spec

After wiring, verify with:

1. Push the scaffolded project to a GitHub repository
2. Create a branch and open a pull request — the `ci.yaml` workflow should trigger automatically
3. Verify the CI workflow runs: build, test, and lint steps should all pass on the scaffold's default code
4. Merge the PR to `main` — the `deploy-staging.yaml` workflow should trigger (it will fail without secrets configured, but it should trigger)
5. Confirm the workflow files exist at `.github/workflows/ci.yaml`, `deploy-staging.yaml`, and `deploy-production.yaml`

## Do NOT

- Deploy to production without a staging approval
- Store deployment credentials in `.env` — use GitHub Actions secrets
- Skip the CI workflow on PRs — never merge with failing tests
- Manually edit generated workflow files — regenerate via `modkit` if changes needed
