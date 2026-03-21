# CI/CD Module

GitHub Actions workflow generation for modkit projects.

## Overview

The CI/CD module generates GitHub Actions workflows for every project scaffolded with `modkit init`. It produces three workflows:

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yaml` | PR / push | Build, test, lint |
| `deploy-staging.yaml` | Push to `main` | Auto-deploy to staging |
| `deploy-production.yaml` | Tag `v*` | Deploy to production |

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `github-actions` | GitHub Actions | MVP | Go, Bun |

## Generated files

```
.github/
└── workflows/
    ├── ci.yaml                    ← build + test + lint
    ├── deploy-staging.yaml        ← auto-deploy on main
    └── deploy-production.yaml     ← deploy on v* tag
```

## Required secrets

Set in GitHub → Settings → Secrets and variables → Actions:
- `DOCKER_REGISTRY_TOKEN`
- `STAGING_DEPLOY_KEY`
- `PRODUCTION_DEPLOY_KEY`

## Configuration

See `config.schema.json` for environment variables.
