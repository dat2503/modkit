# modkit CLI Specification v2.0

> The `modkit` CLI is the bridge between the module registry and project directories. Agents use it to scaffold, manage, and validate projects.

---

## Overview

`modkit` is a Go CLI tool that:
- Scaffolds new projects from registry templates
- Pulls modules into existing projects
- Validates module wiring
- Runs diagnostics
- Lists available modules and runtimes

**Install:**
```bash
go install github.com/yourorg/module-registry/modkit@latest
```

**Registry cache:** `modkit` maintains a local cache of the registry at `~/.modkit/cache/`. It clones the registry repo there on first use and updates it on subsequent commands.

---

## Global Flags

These flags are available on every command:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output`, `-o` | string | `table` | Output format: `table` or `json` |
| `--no-prompt` | bool | `false` | Disable interactive prompts (for agent use) |
| `--registry` | string | `~/.modkit/cache` | Path to registry cache |

When `--output json` is set, all output is valid JSON. Agents should always use `--output json` for programmatic parsing.

When `--no-prompt` is set, commands fail rather than asking interactive questions. This is required for agent use.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Validation failure |
| `3` | Missing dependency (module not found in registry) |
| `4` | Config error (invalid `.modkit.yaml` or missing required fields) |
| `5` | Network error (registry unreachable) |
| `6` | Runtime mismatch (module doesn't support the project's runtime) |

---

## Commands

### `modkit init` — Scaffold a new project [MVP]

Creates a new project directory from a registry template with the selected modules.

```bash
modkit init [flags]
```

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name`, `-n` | ✅ | Project name (used for directory name) |
| `--runtime`, `-r` | ❌ | Backend runtime: `go` or `bun` (default: `bun`) |
| `--frontend`, `-f` | ❌ | Frontend framework: `vite` or `next` (default: `vite`) |
| `--modules`, `-m` | ❌ | Comma-separated `module:impl` pairs. If omitted and `--no-prompt` is not set, interactive selector is shown |
| `--go-module` | ❌ | Go module path (e.g., `github.com/org/project`). Go runtime only. |
| `--no-prompt` | ❌ | Non-interactive mode — uses defaults for unspecified flags |

**Examples:**

```bash
# Interactive — uses defaults (bun + vite)
modkit init --name myapp

# Non-interactive with defaults (bun + vite + better-auth)
modkit init \
  --name invoicely \
  --modules auth,payments:stripe,email:resend \
  --no-prompt

# Go backend + Vite frontend
modkit init \
  --name invoicely \
  --runtime go \
  --frontend vite \
  --go-module github.com/myorg/invoicely \
  --modules auth,payments:stripe,email:resend \
  --no-prompt

# Go backend + Next.js frontend with Clerk auth (non-default auth impl)
modkit init \
  --name invoicely \
  --runtime go \
  --frontend next \
  --go-module github.com/myorg/invoicely \
  --modules auth:clerk,payments:stripe \
  --no-prompt
```

**What `modkit init` creates:**

```
./{name}/
├── apps/
│   ├── web/              ← Vite + React frontend (default) or Next.js
│   └── api/              ← Go or Bun backend
├── contracts/            ← copied from registry for selected modules
├── modules/              ← selected module implementations
├── infra/
│   ├── migrations/       ← empty, ready for migration files
│   ├── docker-compose.yaml
│   └── ci/
│       └── .github/workflows/
├── .modkit.yaml          ← tracks project modules and runtime
├── .env.example          ← generated from module config schemas
├── Makefile
├── CLAUDE.md             ← imports registry orchestration docs
└── README.md
```

**`.modkit.yaml` format:**
```yaml
name: invoicely
runtime: go
go_module: github.com/myorg/invoicely
registry: https://github.com/yourorg/module-registry
registry_sha: abc123def456  # pinned to registry commit

modules:
  - name: auth
    impl: better-auth
    version: 1.0.0
  - name: payments
    impl: stripe
    version: 1.0.0
  # ...
```

---

### `modkit list` — List available modules [MVP]

```bash
modkit list [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--runtime`, `-r` | Filter by runtime: `go` or `bun` |
| `--category`, `-c` | Filter by category: `auth`, `payments`, `email`, etc. |
| `--phase` | Filter by phase: `mvp`, `v2`, `later` |

**Examples:**
```bash
modkit list
modkit list --runtime go --output json
modkit list --phase mvp
```

**Output (table):**
```
NAME             CATEGORY       DEFAULT IMPL    RUNTIMES      PHASE   STATUS
auth             auth           better-auth     go, bun       mvp     stable
payments         payments       stripe          go, bun       mvp     stable
email            notification   resend          go, bun       mvp     stable
realtime         realtime       websocket       go, bun       v2      stable
```

**Output (json):**
```json
{
  "modules": [
    {
      "name": "auth",
      "category": "auth",
      "phase": "mvp",
      "default_impl": "better-auth",
      "runtimes": ["go", "bun"],
      "status": "stable",
      "always_include": false
    }
  ]
}
```

---

### `modkit info <module>` — Show module details [MVP]

```bash
modkit info <module> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--agent` | Print AGENT.md contents (for injecting into agent context) |
| `--impl` | Show details for a specific implementation |

**Examples:**
```bash
modkit info auth
modkit info email --agent
modkit info payments --impl stripe --output json
```

**Use in agent workflow (Phase 1):**
```bash
# Agent reads AGENT.md for each candidate module
modkit info auth --agent
modkit info payments --agent
modkit info email --agent
```

---

### `modkit pull <module>` — Add a module to current project [MVP]

Pulls a module from the registry into the current project directory. The project must have a `.modkit.yaml` file (created by `modkit init`). The runtime is auto-detected from `.modkit.yaml`.

```bash
modkit pull <module> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--impl`, `-i` | Implementation to use (default: module's default for the project's runtime) |

**Examples:**
```bash
# From inside a project directory
modkit pull realtime
modkit pull email --impl sendgrid
modkit pull search --no-prompt
```

**What `modkit pull` does:**
1. Reads `.modkit.yaml` to determine runtime
2. Finds the module in the registry
3. Resolves implementation (uses default if `--impl` not specified)
4. Checks runtime compatibility — exits `6` if the implementation doesn't support the project's runtime
5. Copies module files from registry cache into `modules/`
6. Updates `contracts/` with the module's interface (if not already present)
7. Updates `.modkit.yaml` with the new module entry
8. Prints wiring instructions from `docs/AGENT.md`

**Error: runtime mismatch (exit 6):**
```
Error: module 'jobs' implementation 'asynq' does not support runtime 'bun'.
Available implementations for bun: bullmq
Run: modkit pull jobs --impl bullmq
```

---

### `modkit validate` — Validate module wiring [MVP]

Checks that all modules in the current project are correctly wired and configured.

```bash
modkit validate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--strict` | Treat warnings as errors |

**What it checks:**
- All modules in `.modkit.yaml` exist in the registry
- Module implementation files exist in `modules/`
- Initialization order follows composition rulebook §3
- All required env vars from `config.schema.json` are listed in `.env.example`
- Runtime-specific checks:
  - Go: `go vet ./...` passes
  - Bun: `tsc --noEmit` passes

**Output (json):**
```json
{
  "result": "pass",
  "checks": [
    { "name": "modules_exist", "status": "pass" },
    { "name": "init_order", "status": "pass" },
    { "name": "env_vars", "status": "pass" },
    { "name": "go_vet", "status": "pass" }
  ],
  "warnings": [],
  "errors": []
}
```

**Exit codes:** `0` = all pass, `2` = validation failures

---

### `modkit upgrade` — Upgrade modules [v2]

Upgrades one or all modules to the latest version from the registry.

```bash
modkit upgrade [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--module`, `-m` | Module to upgrade (omit to see all with available upgrades) |
| `--all` | Upgrade all modules with available upgrades |
| `--dry-run` | Show what would change without applying |

**Examples:**
```bash
modkit upgrade --dry-run
modkit upgrade --module email
modkit upgrade --all
```

---

### `modkit doctor` — Diagnostics [MVP]

Runs diagnostic checks on the modkit installation and the current project.

```bash
modkit doctor [flags]
```

**Checks performed:**

| Check | What it verifies |
|-------|-----------------|
| Registry cache | `~/.modkit/cache/` exists and registry.yaml is readable |
| Git | `git` is installed (needed for registry updates) |
| Go runtime | `go version` ≥ 1.22 (if a Go project) |
| Bun runtime | `bun --version` ≥ 1.1 (if a Bun project) |
| Docker | `docker` and `docker-compose` are installed |
| Project config | `.modkit.yaml` exists and is valid (if in a project) |
| Module files | All modules in `.modkit.yaml` have their files present |

**Output (table):**
```
CHECK                    STATUS   DETAILS
registry_cache           ✅ ok    ~/.modkit/cache (last updated: 2 hours ago)
git                      ✅ ok    git 2.43.0
go_runtime               ✅ ok    go1.22.1
docker                   ✅ ok    Docker 25.0.3
project_config           ✅ ok    .modkit.yaml (runtime: go, 8 modules)
module_files             ✅ ok    all 8 modules present
```

---

### `modkit runtimes` — List available runtimes [MVP]

```bash
modkit runtimes [flags]
```

**Output (table):**
```
NAME   LABEL              BUILD CMD       TEST CMD        STATUS
go     Go                 go build ./...  go test ./...   stable
bun    Bun (TypeScript)   bun build       bun test        stable
```

---

## Agent Workflow Examples

### Full project scaffold (Phase 3, Step 3.1)
```bash
modkit init \
  --name invoicely \
  --runtime go \
  --go-module github.com/myorg/invoicely \
  --modules auth,payments:stripe,email:resend,storage:s3,cache:redis,jobs:asynq,observability:otel,error-tracking:sentry \
  --no-prompt --output json
```

### Read module context before Phase 1
```bash
modkit list --output json
modkit info auth --agent
modkit info payments --agent
modkit info email --agent
```

### Add a module post-scaffold
```bash
cd invoicely
modkit pull realtime --no-prompt --output json
```

### Validate after wiring (Phase 3, ongoing)
```bash
modkit validate --output json
# Check exit code: 0 = all good, 2 = failures
```

### Pre-deploy check (Phase 4)
```bash
modkit validate --strict --output json
modkit doctor --output json
```

---

## `.modkit.yaml` Schema

```yaml
# Generated by modkit init, updated by modkit pull/upgrade
name: string           # project name
runtime: go | bun      # backend runtime
go_module: string      # Go module path (go runtime only)
registry: string       # registry repo URL
registry_sha: string   # pinned registry commit SHA

modules:
  - name: string       # module name
    impl: string       # implementation name
    version: string    # semver from module.yaml
```

Do not edit `.modkit.yaml` manually. Use `modkit pull` and `modkit upgrade`.
