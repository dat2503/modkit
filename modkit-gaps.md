# modkit — Gaps & Improvement Notes

> Observations from actually trying to build a Go app (notepad) with the current registry.
> Written after working through the full 6-phase playbook end-to-end.

---

## Fixed (2026-03-22)

The following issues were resolved in a single session. `modkit init --name notepad --runtime go --go-module github.com/you/notepad --modules auth,cache,cicd --no-prompt --registry <path>` now scaffolds a project that compiles with `go build ./...`.

| # | Issue | Fix |
|---|-------|-----|
| 1 | `modkit init` was a no-op stub | Implemented: registry YAML loader, template renderer, module impl copier (`modkit/cmd/registry.go`, `modkit/cmd/scaffold.go`, `modkit/cmd/init.go`) |
| 2 | All 4 Go module implementations panicked | Implemented: `otel-go` (slog), `sentry-go`, `redis-go`, `clerk-go` |
| 4 | Bootstrap imported `{{.GoModule}}/modules/...` packages that didn't exist | `modkit init` now copies `modules/{name}/impl/{impl}-go/` into `apps/api/modules/{name}/` |
| 5 | Migration path inconsistency | Not changed — `apps/api/migrations/` is the canonical path per composition-rulebook |
| 6 | Go version 1.22 → needs 1.24 for clerk-sdk-go | `go.mod.tmpl` and `--go-version` default updated to 1.24 |
| 9 | `cicd` listed as `phase: v2` but treated as MVP | Was already `phase: mvp` in registry.yaml — description clarified |
| 10 | No CORS middleware | Added `corsMiddleware()` to `middleware.go.tmpl` |
| 11 | No `make deps` target | Added to `Makefile.tmpl` |
| 12 | Sentry init fatal in development | `New()` now logs a warning and continues if `Environment != "production"` |
| — | `writeJSON` was a TODO | Implemented with `encoding/json` in `middleware.go.tmpl` |
| — | Bootstrap config type mismatch | `bootstrap.go.tmpl` now uses explicit `Pkg.Config{...}` struct conversions |
| — | No `replace` directive → contracts unresolvable | `go.mod.tmpl` adds `replace github.com/dat2503/modkit => {{.RegistryPath}}` |
| — | Missing `--registry` flag on CLI | Added as a global flag (default: `~/.modkit/cache`) |

**Still open:** #3 (`go install` broken — repo not public), #7 (`lib/api.ts` not scaffolded), #8 (`modkit list` misleading output), and all v2 commands (`validate`, `doctor`, `pull`).

---

---

## 1. Critical: `modkit init` Does Nothing

The single most important command — `modkit init` — is a no-op stub:

```
modkit init: would scaffold "notepad" (runtime: go)
  modules: [auth cache cicd]
(not yet implemented)
```

This means **zero of the template files get rendered**. The whole value proposition of modkit is that it scaffolds a working project from templates, and that part doesn't exist yet. Everything downstream (migrate, validate, doctor) also prints "not yet implemented".

**What needs to happen:**
- Implement Go `text/template` rendering for all `.tmpl` files
- Walk `templates/project-{runtime}/`, render each file with the `TemplateData` struct, write to the output directory
- Strip the `.tmpl` extension from output filenames
- Create the directory structure automatically

---

## 2. Critical: Module Implementations Are Empty Stubs

Every implementation file (`clerk-go`, `redis-go`, `otel-go`, `sentry-go`) is a skeleton that panics at runtime:

```go
func (s *Service) ValidateToken(ctx context.Context, token string) (*contracts.AuthUser, error) {
    panic("not implemented")
}
```

The contracts are well-designed, but there's nothing behind them. An agent scaffolding a project with these modules gets code that compiles but crashes the moment it handles any real request.

**What needs to happen:**
- Implement `clerk-go`: wire `github.com/clerk/clerk-sdk-go/v2/jwt` for token verification
- Implement `redis-go`: wire `github.com/redis/go-redis/v9` for Get/Set/Delete/etc.
- Implement `otel-go`: wire `go.opentelemetry.io/otel` for tracing + `log/slog` for structured logging
- Implement `sentry-go`: wire `github.com/getsentry/sentry-go` for error capture/flush

---

## 3. Critical: Cannot `go install` the CLI

The agent instructions say:

```bash
go install github.com/dat2503/modkit/modkit@latest
```

This fails — the module is either private or the module path doesn't match the subdirectory layout. The CLI had to be built manually from the local source with `go build`.

**What needs to happen:**
- Either make the repo public, or document the local build path clearly
- Or publish a binary release (GitHub Releases) so agents don't need Go installed to get the CLI

---

## 4. Bootstrap Template Imports Non-Existent Local Packages

The rendered `bootstrap.go` would import:

```go
obspkg "github.com/you/notepad/modules/observability"
authpkg "github.com/you/notepad/modules/auth"
```

But `modkit init` never copies any files into `apps/api/modules/`. So even if init rendered the templates, the generated project would fail `go build` immediately due to missing packages.

**What needs to happen:**
- `modkit init` must copy the relevant implementation files from `modules/{name}/impl/{runtime}/` into the generated project's `apps/api/modules/{name}/` directory
- Or change the import strategy: have the project's `go.mod` use a `replace` directive pointing to the local registry clone, and the bootstrap imports directly from `github.com/dat2503/modkit/modules/...`
- The second approach is simpler since there's already a single `go.mod` in the saas repo

---

## 5. Migration Path Inconsistency

Two different locations are referenced across the docs:

| Source | Path |
|--------|------|
| `composition-rulebook.md §9` | `infra/migrations/` |
| `note-app.md` (spec) | `apps/api/migrations/` |
| `Makefile.tmpl` | `apps/api/cmd/migrate` |
| `agent-instructions.md` | `apps/api/migrations/` |

Pick one and be consistent everywhere. `apps/api/migrations/` makes more sense for a Go project since the migration runner lives in the same module.

---

## 6. Go Version Mismatch

The template hardcodes `go 1.22` in `go.mod.tmpl`, but `github.com/clerk/clerk-sdk-go/v2` requires Go 1.24+. Running `go mod tidy` automatically switches the toolchain to `go1.25.8`, which is surprising and can break CI pipelines that pin a Go version.

**Fix:** Update the template to `go 1.24` or test all default module combinations against the Go version specified in the template.

---

## 7. `lib/api.ts` Not in the Web Template

The agent instructions and CLAUDE.md both reference:

```typescript
import { apiClient } from '@/lib/api-client'
const { data } = await apiClient.GET('/api/v1/notes')
```

But the web template only generates `layout.tsx` and `page.tsx`. There's no `src/lib/api.ts` (or `api-client.ts`) scaffolded. Agents and developers are left to build the frontend API client from scratch every time.

**Fix:** Add `apps/web/src/lib/api.ts.tmpl` — a typed fetch wrapper that reads `NEXT_PUBLIC_API_URL` and wraps GET/POST/PUT/DELETE with the `{ "data": ... }` envelope format.

---

## 8. `modkit list` Output is Misleading

```
Available modules (not yet implemented):
  auth   mvp  Clerk (go, bun)
  cache  mvp  Redis (go, bun)
```

Every module says "not yet implemented" regardless of actual status. This is noise for an agent trying to decide which modules to include — it reads as "nothing works".

**Fix:** Separate "CLI command not implemented" from "module implementation not implemented". Show actual implementation status per module (stub / partial / stable).

---

## 9. `cicd` Module Phase Inconsistency

In `registry.yaml`, `cicd` is listed as `phase: v2`:

```yaml
- name: cicd
  phase: v2   # ← listed as v2
```

But in `agent-instructions.md` and `note-app.md`, it's treated as an always-include MVP module:

> `cicd` | Almost always — generates GitHub Actions workflows

**Fix:** Change `cicd` to `phase: mvp` in `registry.yaml`.

---

## 10. No CORS Middleware in Template

The composition rulebook (§8) defines the middleware chain as:

```
CORS → Request ID → Auth → Rate limit → Handler
```

But `middleware.go.tmpl` has no CORS middleware. Any browser calling the API from a different origin (which is always the case with `localhost:3000` → `localhost:8080`) will get blocked.

**Fix:** Add a CORS middleware to `middleware.go.tmpl` that reads allowed origins from config (defaulting to `*` in development, configurable for production).

---

## 11. No `make deps` Target Despite Being Referenced

`go.mod.tmpl` says:

```
// Run: make deps
```

But `Makefile.tmpl` has no `deps` target. The `setup` target runs `go mod download` but is named differently.

**Fix:** Either add a `deps` target or remove the comment from `go.mod.tmpl`.

---

## 12. Sentry Init Should Not Be Fatal in Development

The current pattern (and the template) treats Sentry init failure as a hard error. But in local development where `SENTRY_DSN=https://dummy@sentry.io/0`, Sentry init reliably fails and blocks startup.

The sentry-go SDK itself returns an error for dummy DSNs.

**Fix:** Sentry init failures should log a warning and continue — especially in `APP_ENV=development`. Only treat it as fatal in production. The template should reflect this.

---

## Summary: Priority Order

| Priority | Issue |
|----------|-------|
| 🔴 P0 | `modkit init` not implemented — nothing works without this |
| 🔴 P0 | Module implementations are stubs that panic |
| 🔴 P0 | `go install` path broken — CLI can't be installed |
| 🟠 P1 | Bootstrap imports non-existent local packages |
| 🟠 P1 | Go version mismatch (1.22 → needs 1.24) |
| 🟡 P2 | `lib/api.ts` not scaffolded in web template |
| 🟡 P2 | Migration path inconsistency across docs |
| 🟡 P2 | No CORS middleware in template |
| 🟢 P3 | `cicd` phase wrong in registry.yaml |
| 🟢 P3 | `modkit list` shows "not yet implemented" for everything |
| 🟢 P3 | Sentry init shouldn't be fatal in dev |
| 🟢 P3 | Missing `make deps` target |
