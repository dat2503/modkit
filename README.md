# Module Registry

A shared library of pre-built, swappable components for SaaS web applications. Agents and developers use this registry to scaffold and assemble production-ready projects via the `modkit` CLI.

---

## What's in this repo

```
module-registry/
├── orchestration/        ← Agent playbook, composition rulebook, registry index
├── contracts/            ← Go and TypeScript module interfaces
├── modules/              ← Module implementations (manifests, configs, code stubs)
├── templates/            ← Project scaffold templates (Go + Bun runtimes)
├── modkit/               ← CLI tool source code (Go + Cobra)
└── docs/                 ← Contributor guides and CLI specification
```

---

## Quick Start (Using the CLI)

**Install modkit:**
```bash
go install github.com/dat2503/modkit/modkit@latest
```

**Scaffold a new project:**
```bash
modkit init \
  --name myapp \
  --runtime go \
  --go-module github.com/myorg/myapp \
  --modules auth:clerk,payments:stripe,email:resend \
  --no-prompt
```

**Add a module to an existing project:**
```bash
cd myapp
modkit pull realtime --no-prompt
```

**Validate module wiring:**
```bash
modkit validate --output json
```

---

## Available Modules

| Module | Category | Phase | Default Impl |
|--------|----------|-------|--------------|
| `observability` | observability | mvp | otel |
| `error-tracking` | error-tracking | mvp | sentry |
| `auth` | auth | mvp | clerk |
| `payments` | payments | mvp | stripe |
| `email` | notification | mvp | resend |
| `storage` | storage | mvp | s3 |
| `cache` | cache | mvp | redis |
| `jobs` | jobs | mvp | asynq (Go) / bullmq (Bun) |
| `realtime` | realtime | v2 | websocket |
| `search` | search | v2 | elasticsearch |
| `feature-flags` | feature-flags | v2 | flagsmith |
| `cicd` | cicd | mvp | github-actions |

---

## Supported Runtimes

| Runtime | Language | Build | Test |
|---------|----------|-------|------|
| `go` | Go 1.22+ | `go build ./...` | `go test ./...` |
| `bun` | TypeScript (Bun 1.1+) | `bun build` | `bun test` |

All projects share a Next.js frontend (`apps/web/`). The `--runtime` flag selects the backend only.

---

## Agent Workflow

Agents follow the 6-phase playbook in `orchestration/playbook.md`:

1. **Phase 0 — Intake**: Parse project idea → structured brief
2. **Phase 1 — Module Selection**: Select modules from registry
3. **Phase 2 — Architecture Plan**: DB schema, routes, wiring plan
4. **Phase 3 — Scaffold & Wire**: `modkit init` + write handlers, migrations, tests
5. **Phase 4 — Validate**: `modkit validate` + build + tests
6. **Phase 5 — Deploy**: CI/CD → staging → production

---

## Contributing

### Adding a New Module

See [`docs/module-registry-spec.md`](docs/module-registry-spec.md) for the full contributor guide.

Quick checklist:
1. Define interfaces in `contracts/go/` and `contracts/ts/`
2. Create `modules/{name}/` with `module.yaml`, `config.schema.json`, `docs/AGENT.md`
3. Add implementation stubs in `impl/{provider}-go/` and `impl/{provider}-ts/`
4. Write contract tests in `tests/`
5. Add to `orchestration/registry.yaml`

### CLI Development

See [`docs/modkit-cli-spec.md`](docs/modkit-cli-spec.md) for the full CLI specification.

```bash
cd modkit
go build ./...
./modkit --help
```
