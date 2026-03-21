# modkit — Agent Instructions

> Load this document at the start of any session where you will build a SaaS application using the modkit registry. It is your primary reference for the entire build.

---

## 1. What This Kit Is

modkit is a **module registry for SaaS scaffolding**. It is a library — not a framework, not an app. It provides four layers:

| Layer | What it is | Where it lives |
|-------|-----------|----------------|
| **Contracts** | Go + TypeScript interfaces — the "shape" of each capability | `contracts/go/`, `contracts/ts/` |
| **Modules** | Named capabilities (auth, payments, email, etc.) with one or more interchangeable implementations | `modules/{name}/` |
| **Templates** | Starter project scaffolds for each runtime, processed by `modkit init` | `templates/project-{runtime}/` |
| **CLI** | `modkit` — the tool that scaffolds projects, manages modules, and validates wiring | `modkit/` |

Your job as an agent is to take a user's project idea, select the right modules, scaffold a project, and wire everything together following the composition rulebook.

---

## 2. System Map

```
modkit registry repo (this repo)
├── contracts/          ← interfaces for all modules
├── modules/            ← implementations + manifests + agent docs
│   └── auth/
│       ├── module.yaml          ← metadata, dependencies, impls
│       ├── config.schema.json   ← required env vars
│       ├── docs/AGENT.md        ← how to wire this module
│       └── impl/clerk-go/       ← concrete implementation
├── templates/          ← project scaffolds
└── modkit/             ← CLI source code
        │
        │  modkit init --name my-app --runtime go --modules auth,payments
        ▼
generated project (my-app/)
├── apps/
│   ├── api/           ← Go or Bun backend
│   │   ├── main.go / index.ts    ← entry point
│   │   ├── bootstrap/            ← module initialization
│   │   ├── config/               ← env-based config loading
│   │   └── api/                  ← router + handlers
│   └── web/           ← Next.js 14 frontend
├── infra/
│   └── docker-compose.yaml       ← local dev services
├── Makefile            ← common tasks
├── .env.example        ← required env vars (generated per modules)
└── .modkit.yaml        ← project manifest (tracks selected modules)
        │
        │  make dev
        ▼
running application
├── API on :8080
└── Web on :3000
```

---

## 3. Before Every Session

Load context in this order before writing any code:

```bash
# 1. Read the 6-phase workflow
cat orchestration/playbook.md

# 2. Read the wiring rules
cat orchestration/composition-rulebook.md

# 3. See what modules are available
modkit list --output json

# 4. For any module you'll use, read its agent docs
modkit info <module> --agent
# e.g.: modkit info auth --agent
#       modkit info payments --agent
```

If `modkit` is not installed:
```bash
go install github.com/dat2503/modkit/modkit@latest
```

---

## 4. The 6-Phase Workflow (Quick Reference)

Full detail in `orchestration/playbook.md`. Never skip phases. Always stop at 🔒 checkpoints.

| Phase | What you do | Key output | 🔒 Checkpoint |
|-------|------------|------------|--------------|
| **0 — Intake** | Parse the idea into a structured brief | YAML brief with entities, roles, flows, assumptions | Human reviews brief |
| **1 — Module Selection** | Run `modkit list`, read AGENT.md per candidate module, select with rationale | Selected/skipped list with reasons | Human approves selection |
| **2 — Architecture Plan** | Generate DB schema, API routes, module wiring, frontend pages | Architecture document | Human approves before any code |
| **3 — Scaffold & Wire** | `modkit init`, write migrations + handlers + bootstrap + tests | Running project that builds and passes tests | Human monitors |
| **4 — Validate** | `modkit validate`, `modkit doctor`, `make build`, `make test`, smoke tests | Validation report | Human reviews, tests locally |
| **5 — Deploy** | Set up CI/CD, deploy to staging, then production | Live staging URL | Human approves each env |

---

## 5. Module Selection Guide

### Always included (no decision needed)
- **observability** — OpenTelemetry. Always first in init order.
- **error-tracking** — Sentry. Always second in init order.

### Decision rules for optional modules

| Module | Include when | Skip when | Requires |
|--------|-------------|-----------|---------|
| **auth** | App has user accounts, any route needs auth | Public-only API | cache |
| **cache** | auth included, jobs included, sessions, rate limiting | Stateless API with no auth/jobs | — |
| **payments** | App processes money, subscriptions, one-time purchases | Free tool, no monetization | — |
| **email** | Confirmations, notifications, invoices, any transactional email | No user communication needed | — (pair with jobs) |
| **storage** | File uploads, image hosting, generated PDFs/exports | No files or binary data | — |
| **jobs** | Email sending, file generation, any operation >500ms, retries | All operations are fast and sync | cache |
| **realtime** *(v2)* | Live updates, dashboards, chat, notifications without polling | Standard request/response is enough | cache, auth |
| **search** *(v2)* | Full-text search with relevance, >100k records, multi-field filtering | Postgres ILIKE is enough (<100k rows) | — |
| **feature-flags** *(v2)* | Phased rollout, A/B testing, kill switches | No staged feature releases needed | — |
| **cicd** | Almost always — generates GitHub Actions workflows | Non-GitHub VCS or existing CI | — |

### Initialization order (mandatory)

```
1. observability   ← always first
2. error-tracking  ← always second
3. cache           ← before auth and jobs
4. auth            ← after cache
5. payments        ← any order
6. email           ← any order
7. storage         ← any order
8. jobs            ← after cache
9. realtime        ← after auth
10. search          ← any order
11. feature-flags   ← any order
```

Violating this order causes runtime panics or incorrect behavior. `modkit validate` enforces it.

### Runtime choice

| Pick Go if | Pick Bun if |
|-----------|------------|
| Team has Go experience | Team prefers TypeScript end-to-end |
| Performance-critical workloads | Shared types between frontend and backend |
| Strong typing without a transpiler | Faster iteration cycles |

---

## 6. Generated Project Structure

After `modkit init`, the project looks like this (Go runtime example):

```
my-app/
├── apps/
│   ├── api/
│   │   ├── main.go              ← HTTP server, graceful shutdown (30s timeout)
│   │   ├── go.mod               ← Go module (module path from --go-module)
│   │   ├── bootstrap/
│   │   │   └── bootstrap.go     ← ALL module initialization, correct order enforced
│   │   ├── config/
│   │   │   └── config.go        ← config.Load() reads ALL env vars once, fails fast on missing
│   │   └── api/
│   │       ├── router.go        ← http.NewServeMux(), Deps struct, all routes
│   │       └── middleware.go    ← tracing, logging, recovery, authRequired
│   └── web/
│       ├── package.json
│       ├── src/
│       │   ├── app/
│       │   │   ├── layout.tsx   ← ClerkProvider wraps app (if auth selected)
│       │   │   └── page.tsx     ← placeholder home page
│       │   └── lib/
│       │       └── api.ts       ← typed fetch client (api.get, api.post, etc.)
│       └── tailwind.config.ts
├── infra/
│   └── docker-compose.yaml      ← Postgres (always) + Redis (if cache) + ES (if search)
├── Makefile                     ← make dev, make build, make test, make migrate
├── .env.example                 ← one section per selected module
├── .modkit.yaml                 ← project manifest
└── CLAUDE.md                    ← project-level Claude instructions
```

**Key files to know:**

- **`bootstrap/bootstrap.go`** — The only place modules are instantiated. Edit this to wire new modules. Init order is enforced here.
- **`config/config.go`** — Reads all env vars. Add new vars here when pulling new modules. Never read `os.Getenv` anywhere else.
- **`api/router.go`** — Add all new routes here. The `Deps` struct holds injected modules.
- **`api/middleware.go`** — `authRequired` middleware; use `UserFromContext(r.Context())` in handlers.

---

## 7. Essential Wiring Patterns

These are the rules you must follow when writing application code. Full detail in `orchestration/composition-rulebook.md`.

### Interface-first — never import concrete implementations
```go
// ✅ Always depend on the interface
type TaskService struct {
    email contracts.EmailService
    jobs  contracts.JobsService
}

// ❌ Never import the implementation directly
type TaskService struct {
    email *resend.Client
}
```

### Constructor injection — no globals, no init() magic
```go
// ✅ Constructor injection
func NewTaskService(email contracts.EmailService, jobs contracts.JobsService) *TaskService {
    return &TaskService{email: email, jobs: jobs}
}

// ❌ No global singletons
var emailClient = resend.NewClient(os.Getenv("RESEND_API_KEY"))
```

### Config loading — once, at startup, fail fast
```go
// ✅ Load config once in main, inject everywhere
cfg, err := config.Load()
if err != nil {
    log.Fatalf("config: %v", err)  // lists ALL missing vars
}

// ❌ Never read env vars in handlers or services
apiKey := os.Getenv("RESEND_API_KEY")
```

### Response envelope — consistent shape
```go
// ✅ All API responses
// Success:
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]any{"data": result})

// Error:
http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
```

### Auth middleware — token extraction and context propagation
```go
// ✅ Extract in middleware, use in handler
func authRequired(auth contracts.AuthService, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
        user, err := auth.ValidateToken(r.Context(), token)
        if err != nil {
            writeError(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
    })
}

// In handler:
user := UserFromContext(r.Context())
```

### Error handling — structured errors, never panic
```go
// ✅ Return errors, let middleware handle them
if err != nil {
    return nil, fmt.Errorf("tasks: get by id %s: %w", id, err)
}

// ❌ Never panic in handlers
panic("something went wrong")
```

### Database conventions
- All tables: UUID primary keys (`gen_random_uuid()`), `created_at`, `updated_at` TIMESTAMPTZ
- User-generated content: `deleted_at` column for soft deletes
- All queries: parameterized inputs — never string-concat SQL

---

## 8. modkit CLI Quick Reference

```
modkit [command] [flags]

Global flags:
  -o, --output string   "table" or "json" (default "table")
      --no-prompt       Fail on missing input instead of prompting
```

| Command | Purpose | Key flags |
|---------|---------|-----------|
| `modkit init` | Scaffold new project from template | `--name`, `--runtime go\|bun`, `--modules`, `--go-module`, `--no-prompt` |
| `modkit list` | List all available modules | `--runtime go\|bun` |
| `modkit info <module>` | Show module details | `--agent` (prints AGENT.md) |
| `modkit pull <module>` | Add a module to existing project | `--impl <name>` |
| `modkit validate` | Check module wiring + init order | `--strict` |
| `modkit upgrade` | Upgrade module implementations | `--all`, `--module <name>` |
| `modkit doctor` | Check local environment | — |
| `modkit runtimes` | List supported runtimes | — |

**Example — scaffold with common modules:**
```bash
modkit init \
  --name my-app \
  --runtime go \
  --go-module github.com/you/my-app \
  --modules auth,payments,email,cache,jobs,cicd \
  --no-prompt
```

**Example — add a module to existing project:**
```bash
cd my-app
modkit pull realtime
modkit validate
```

---

## 9. Common Mistakes

1. **Wrong init order in bootstrap** — Always: observability → error-tracking → cache → auth → jobs → rest. Run `modkit validate` to check.

2. **Reading env vars in handlers or services** — All env reading belongs in `config/config.go`. Inject the config struct.

3. **Importing concrete implementations in app code** — `apps/api/` must only import from `contracts/`, never from `modules/`.

4. **Forgetting to add env vars to `.env.example`** — Every new required config must be documented there.

5. **Creating a global cache/DB client** — Pass instances through constructors. No package-level vars that hold connections.

6. **Mutating shared state in handlers** — Handlers are called concurrently. Use context, not struct fields, for request-scoped data.

7. **Calling `svc.Shutdown()` before cleanup** — Shutdown order is reverse of init order. The bootstrap template handles this; don't break it.

8. **Skipping `modkit validate` before committing** — Always run before pushing. It catches init-order violations and missing deps.

9. **Hard-coding secrets** — Every credential goes in env vars. Fail fast if missing (the config template does this by default).

10. **Writing SQL with string concatenation** — Always use parameterized queries. No exceptions.

---

## 10. Extending the Generated Project

### Add a new HTTP handler (Go)

1. Add the route in `apps/api/api/router.go`:
```go
mux.HandleFunc("GET /api/tasks", authRequired(deps.Auth, http.HandlerFunc(handleListTasks)))
```

2. Write the handler in a new file `apps/api/api/tasks.go`:
```go
func handleListTasks(w http.ResponseWriter, r *http.Request) {
    user := UserFromContext(r.Context())
    // ... use injected deps via closure or pass deps struct
    writeJSON(w, result)
}
```

3. If the handler needs a service, add it to the `Deps` struct and inject in `bootstrap.go`.

### Add a background job (Go / Asynq)

1. Define the job payload type in `apps/api/jobs/types.go`
2. Enqueue in a handler: `deps.Jobs.Enqueue(ctx, "send_welcome_email", payload)`
3. Register the handler in `bootstrap.go`: `jobs.RegisterHandler("send_welcome_email", handleSendWelcomeEmail)`

### Add a Next.js page

1. Create `apps/web/src/app/tasks/page.tsx`
2. Use the API client: `import { api } from '@/lib/api'`
3. For auth-protected pages, wrap with Clerk's `<SignedIn>` component

### Add a module after initial scaffold

```bash
modkit pull search          # copies impl files
modkit validate             # confirms wiring is correct
# Then: add ELASTICSEARCH_URL to .env, wire in bootstrap.go
```

---

## Reference Documents

| Document | When to read |
|----------|-------------|
| `orchestration/playbook.md` | Start of every session — full 6-phase workflow |
| `orchestration/composition-rulebook.md` | Before writing any application code |
| `orchestration/registry.yaml` | When selecting or looking up modules |
| `modules/{name}/docs/AGENT.md` | Before wiring a specific module |
| `docs/modkit-cli-spec.md` | Full CLI specification |
| `docs/module-registry-spec.md` | When adding a new module to the registry |
