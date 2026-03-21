# modkit CLI

The `modkit` CLI scaffolds new SaaS projects and manages pluggable modules from the module registry.

## Installation

```bash
# From source
cd modkit && go install .

# Or build binary
cd modkit && go build -o bin/modkit .
```

## Usage

```
modkit [command] [flags]

Global flags:
  -o, --output string   Output format: "table" or "json" (default "table")
      --no-prompt       Disable interactive prompts; fail on missing required input
```

## Commands

### `modkit init`

Scaffold a new project from the registry templates.

```bash
# Interactive
modkit init

# Non-interactive (Go runtime)
modkit init \
  --name my-app \
  --runtime go \
  --go-module github.com/user/my-app \
  --modules auth,payments,email,cache \
  --no-prompt

# Non-interactive (Bun runtime)
modkit init \
  --name my-app \
  --runtime bun \
  --modules auth,payments \
  --no-prompt
```

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Project name | (prompted) |
| `--runtime` | `go` or `bun` | (prompted) |
| `--modules` | Comma-separated module list | (prompted) |
| `--go-module` | Go module path | (prompted if go) |
| `--go-version` | Go version | `1.22` |
| `--bun-version` | Bun version | `1.1` |

### `modkit list`

List all available modules.

```bash
modkit list
modkit list --runtime go
modkit list --output json
```

### `modkit info <module>`

Show details for a module.

```bash
modkit info auth
modkit info payments --agent   # print AGENT.md
modkit info cache --output json
```

### `modkit pull <module>`

Add a module to the current project.

```bash
modkit pull auth
modkit pull payments --impl stripe
```

### `modkit validate`

Validate module wiring in the current project.

```bash
modkit validate
modkit validate --strict
modkit validate --output json
```

### `modkit upgrade`

Upgrade module implementations.

```bash
modkit upgrade --all
modkit upgrade --module auth
```

### `modkit doctor`

Check local environment.

```bash
modkit doctor
```

### `modkit runtimes`

List supported runtimes.

```bash
modkit runtimes
modkit runtimes --output json
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation failure |
| 3 | Missing dependency |
| 4 | Config error |
| 5 | Network error |
| 6 | Runtime mismatch |
