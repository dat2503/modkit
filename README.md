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

## Documentation

| Doc | Audience | What it covers |
|-----|----------|---------------|
| [`docs/getting-started.md`](docs/getting-started.md) | **Developers** | Install modkit, scaffold a project, run locally, add handlers and pages |
| [`docs/agent-instructions.md`](docs/agent-instructions.md) | **AI agents** | System overview, module selection, wiring patterns, CLI reference, common mistakes |
| [`orchestration/playbook.md`](orchestration/playbook.md) | Agents | 6-phase workflow for building SaaS apps (authoritative) |
| [`orchestration/composition-rulebook.md`](orchestration/composition-rulebook.md) | Agents | All wiring rules with Go + TypeScript examples |
| [`docs/module-registry-spec.md`](docs/module-registry-spec.md) | Contributors | How to add a new module to the registry |
| [`docs/modkit-cli-spec.md`](docs/modkit-cli-spec.md) | Contributors | Full modkit CLI specification |
| [`docs/workflow-improvements.md`](docs/workflow-improvements.md) | All | All workflow additions explained — self-review, guardrails, Phase 6, learning loop, compliance postures |
| [`learnings/catalog.yaml`](learnings/catalog.yaml) | Agents | Cross-project lessons catalog — loaded at `/new-app` start, populated via `/learn` |

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

| Module | Category | Phase | Default Impl | Selection |
|--------|----------|-------|--------------|-----------|
| `observability` | observability | mvp | otel | Optional — include for production, skip for prototypes |
| `error-tracking` | error-tracking | mvp | sentry | Optional — include for production, skip for prototypes |
| `auth` | auth | mvp | clerk | Include when app has user accounts |
| `payments` | payments | mvp | stripe | Include when app processes money |
| `email` | notification | mvp | resend | Include when app sends transactional email |
| `storage` | storage | mvp | s3 | Include when app stores user files |
| `cache` | cache | mvp | redis | Required by auth and jobs |
| `jobs` | jobs | mvp | asynq (Go) / bullmq (Bun) | Include for background processing |
| `realtime` | realtime | v2 | websocket | Include only when <1s live updates are required |
| `search` | search | v2 | elasticsearch | Include when full-text search or >1k rows needed |
| `feature-flags` | feature-flags | v2 | flagsmith | Include for phased rollouts / A/B testing |
| `cicd` | cicd | mvp | github-actions | Always generated; also supports vercel, railway |

Module selection is user-driven — when using `/new-app`, the agent presents the full menu and asks for explicit yes/no per module. You can add or remove modules from an existing project with `/configure`.

---

## Supported Runtimes

| Runtime | Language | Build | Test |
|---------|----------|-------|------|
| `go` | Go 1.22+ | `go build ./...` | `go test ./...` |
| `bun` | TypeScript (Bun 1.1+) | `bun build` | `bun test` |

All projects share a Next.js frontend (`apps/web/`). The `--runtime` flag selects the backend only.

---

## Agent Workflow

Agents follow the 7-phase playbook in `orchestration/playbook.md`. Entry point: `/new-app` Claude Code skill.

| Phase | Name | Human gate |
|-------|------|-----------|
| 0 | Intake — structured brief from user idea | ✅ required |
| 1 | Module selection + compliance posture | ✅ required |
| 2 | Architecture plan (schema, routes, patterns, §21/§22) | ✅ required |
| 2.5 | Design analysis — `ui-spec.yaml` from Canva/Figma/image | ✅ if design provided |
| 3 | Scaffold + implement (milestones 3a–3d with self-review) | optional after 3c |
| 4 | Validate — E2E with docker-compose + Playwright | ✅ required |
| 5 | Deploy — staging gate → production | ✅ required |
| 6 | Operate — SLOs, evolution loop, weekly maintenance | agent-driven; human on findings |

**Agent skills** (Claude Code slash commands):

| Skill | Purpose |
|-------|---------|
| `/new-app` | Start a new project — full Phase 0–6 workflow |
| `/operate` | Run Phase 6 (6a instrument, 6b iterate, 6c maintain) |
| `/postmortem` | Guided blameless postmortem → architecture amendment |
| `/release` | Semver bump + CHANGELOG + tag + deploy |
| `/learn` | Extract lessons from postmortems → catalog proposals |
| `/configure` | Add or remove modules from an existing project |
| `/update-modkit` | Sync local registry cache via git pull |

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
