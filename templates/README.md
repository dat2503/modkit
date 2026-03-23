# Project Scaffold Templates

This directory contains Go template files (`.tmpl` extension) that `modkit init` processes when scaffolding a new project.

## Templates

| Directory | Runtime | API Framework |
|-----------|---------|---------------|
| `project-go/` | Go | `net/http` (stdlib) |
| `project-bun/` | Bun | Hono |

Both templates generate a **monorepo** with:
- `apps/api/` — backend API
- `apps/web/` — Next.js 14 frontend (TypeScript + Tailwind)
- `infra/` — Docker Compose for local infrastructure
- `Makefile` — common tasks
- `.env.example` — required environment variables
- `.modkit.yaml` — module registry manifest
- `CLAUDE.md` — project-level Claude Code instructions

## Template Variables

Templates use Go's `text/template` syntax. Available variables:

### Always Available

| Variable | Type | Description |
|----------|------|-------------|
| `{{.Name}}` | `string` | Project name (e.g., `my-app`) |
| `{{.Modules}}` | `[]Module` | List of selected modules |
| `{{.RegistryVersion}}` | `string` | modkit registry version used |

### Go Runtime Only

| Variable | Type | Description |
|----------|------|-------------|
| `{{.GoModule}}` | `string` | Go module path (e.g., `github.com/user/my-app`) |
| `{{.GoVersion}}` | `string` | Go version (e.g., `1.22`) |

### Bun Runtime Only

| Variable | Type | Description |
|----------|------|-------------|
| `{{.BunVersion}}` | `string` | Bun version (e.g., `1.1`) |

### Module Type Fields

Each item in `{{.Modules}}` has:

| Field | Type | Description |
|-------|------|-------------|
| `{{.Name}}` | `string` | Module name (e.g., `auth`) |
| `{{.Impl}}` | `string` | Implementation (e.g., `better-auth`) |
| `{{.ImplDir}}` | `string` | Registry-relative path to implementation directory |

## Helper Functions

| Function | Description |
|----------|-------------|
| `{{.HasModule "name"}}` | Returns `true` if the named module is selected |
| `{{.ImplFor "name"}}` | Returns the implementation name for a module (e.g., `"better-auth"`) |

### Example Usage

```go
{{- if .HasModule "auth"}}
{{- if eq (.ImplFor "auth") "better-auth"}}
// Better Auth is the default
{{- else if eq (.ImplFor "auth") "clerk"}}
// Clerk is an alternative
{{- end}}
{{- end}}
```

```go
{{- range .Modules}}
// Module: {{.Name}} ({{.Impl}})
{{- end}}
```

## Adding a New Template

1. Create a new directory: `templates/project-{runtime}/`
2. Follow the same structure as existing templates
3. All files that need variable substitution must have the `.tmpl` extension
4. Static files (no substitution needed) do not need `.tmpl`
5. Update this README with the new template's variables

## Template Processing

`modkit init` processes templates by:
1. Walking the template directory recursively
2. For each `.tmpl` file: executing Go template → writing output file (without `.tmpl` extension)
3. For non-`.tmpl` files: copying as-is
4. Preserving directory structure
