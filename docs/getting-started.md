# Getting Started with modkit

This guide walks you through building your first SaaS project with modkit — from installation to a running local app. We'll build a **task tracker** as the running example.

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl) |
| Bun | 1.1+ | `curl -fsSL https://bun.sh/install \| bash` |
| Docker | 24+ | [docker.com](https://www.docker.com/get-started) |
| Git | any | already installed on most systems |

You only need Go **or** Bun depending on which backend runtime you choose. You need Docker for local infrastructure (Postgres, Redis).

---

## Install modkit

```bash
go install github.com/dat2503/modkit/modkit@latest
```

Verify it works:

```bash
modkit --version
modkit doctor
```

`modkit doctor` checks that your environment has the tools it needs. Fix any failures it reports before continuing.

---

## Choose Your Modules

Before scaffolding, decide which modules your project needs. Run `modkit list` to see everything available:

```bash
modkit list
```

For the task tracker, we'll use:
- **auth** — users need accounts
- **cache** — required by auth (and good for sessions)
- **email** — send task assignment notifications
- **jobs** — send emails asynchronously
- **cicd** — generate GitHub Actions workflows

`observability` and `error-tracking` are always included automatically.

To learn what a module does before choosing it:

```bash
modkit info auth
modkit info payments  # we'll skip this — task tracker is free
```

---

## Scaffold the Project

```bash
modkit init \
  --name task-tracker \
  --runtime go \
  --go-module github.com/you/task-tracker \
  --modules auth,cache,email,jobs,cicd \
  --no-prompt
```

For Bun runtime instead:
```bash
modkit init \
  --name task-tracker \
  --runtime bun \
  --modules auth,cache,email,jobs,cicd \
  --no-prompt
```

After a few seconds, you'll have a `task-tracker/` directory. Let's look inside.

---

## What Was Generated

```
task-tracker/
├── apps/
│   ├── api/                      ← Go backend
│   │   ├── main.go               ← server entry point with graceful shutdown
│   │   ├── go.mod                ← github.com/you/task-tracker
│   │   ├── bootstrap/
│   │   │   └── bootstrap.go      ← wires all modules in the right order
│   │   ├── config/
│   │   │   └── config.go         ← reads ALL env vars once, fails with a clear list if any are missing
│   │   └── api/
│   │       ├── router.go         ← HTTP routes + Deps struct
│   │       └── middleware.go     ← auth, logging, tracing, recovery
│   └── web/                      ← Next.js 14 frontend
│       ├── src/
│       │   ├── app/
│       │   │   ├── layout.tsx    ← ClerkProvider (auth) wraps everything
│       │   │   └── page.tsx      ← placeholder home page
│       │   └── lib/
│       │       └── api.ts        ← typed fetch wrapper: api.get(), api.post(), etc.
│       └── package.json
├── infra/
│   └── docker-compose.yaml       ← Postgres + Redis
├── Makefile                      ← make dev, make test, make migrate, etc.
├── .env.example                  ← one section per module with all required keys
├── .modkit.yaml                  ← records which modules are active
└── CLAUDE.md                     ← project instructions for Claude Code
```

A few things to notice:

**`bootstrap/bootstrap.go`** initializes all modules in the required order:
```
observability → error-tracking → cache → auth → email → jobs
```
This order is enforced — swap it and things will break at runtime.

**`config/config.go`** collects all missing env vars and fails with one clear error message listing everything that's needed — instead of failing one var at a time.

**`api/router.go`** has a `Deps` struct that holds all injected modules. Add new routes here; add new dependencies to `Deps` and wire them in `bootstrap.go`.

**`.env.example`** has one section per selected module:
```
# ── Auth (Clerk) ──────
CLERK_SECRET_KEY=sk_test_changeme
...

# ── Cache (Redis) ─────
REDIS_URL=redis://localhost:6379

# ── Email (Resend) ────
RESEND_API_KEY=re_changeme
...
```

---

## Configure and Run Locally

**Step 1: Set up your environment file**

```bash
cd task-tracker
cp .env.example .env
```

Open `.env` and fill in the required keys:
- `CLERK_SECRET_KEY` and `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` — from [clerk.com](https://clerk.com) (free tier)
- `RESEND_API_KEY` — from [resend.com](https://resend.com) (free tier)
- Leave `DATABASE_URL` and `REDIS_URL` as-is (Docker will provide them)

**Step 2: Start infrastructure and install dependencies**

```bash
make setup
```

This installs Go and Node dependencies, then starts Postgres and Redis via Docker Compose.

**Step 3: Start the app**

```bash
make dev
```

This starts both the API (`localhost:8080`) and the web app (`localhost:3000`) concurrently.

**Verify everything is running:**

```bash
curl http://localhost:8080/health
# → {"status":"ok"}
```

Open `http://localhost:3000` in your browser.

---

## Add a Handler

The generated project has a health check but no business logic yet. Let's add a `GET /api/tasks` endpoint.

**1. Add the route to `apps/api/api/router.go`:**

```go
mux.HandleFunc("GET /api/tasks", authRequired(deps.Auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    user := UserFromContext(r.Context())
    // TODO: query tasks for user.ID from database
    writeJSON(w, []map[string]any{
        {"id": "1", "title": "First task", "owner": user.ID},
    })
})))
```

**2. Test it:**

```bash
# Get an auth token from Clerk, then:
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/tasks
# → {"data":[{"id":"1","title":"First task","owner":"..."}]}
```

All successful responses are wrapped in `{"data": ...}`. All error responses use `{"error": "..."}`.

---

## Add a Frontend Page

**1. Create `apps/web/src/app/tasks/page.tsx`:**

```tsx
'use client'
import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

type Task = { id: string; title: string; owner: string }

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([])

  useEffect(() => {
    api.get<Task[]>('/api/tasks').then(setTasks).catch(console.error)
  }, [])

  return (
    <main className="p-8">
      <h1 className="text-2xl font-bold mb-4">Tasks</h1>
      <ul className="space-y-2">
        {tasks.map(t => (
          <li key={t.id} className="p-3 bg-white rounded shadow">{t.title}</li>
        ))}
      </ul>
    </main>
  )
}
```

**2. Navigate to `http://localhost:3000/tasks`.**

The `api` client automatically uses `NEXT_PUBLIC_API_URL` (set to `http://localhost:8080` by default). For authenticated requests, pass the Clerk session token in headers.

---

## Add a Module After Init

Suppose you decide the task tracker needs payments (Pro tier). Pull it in:

```bash
modkit pull payments
```

This copies the Stripe implementation files into your project and updates `.modkit.yaml`. Then:

1. Add `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` to `.env`
2. Wire Stripe in `bootstrap/bootstrap.go` (modkit prints instructions after pull)
3. Add `payments contracts.PaymentsService` to the `Deps` struct in `router.go`
4. Validate everything: `modkit validate`

---

## Validate Your Wiring

```bash
modkit validate
```

This checks:
- All required module dependencies are present
- Module initialization order is correct
- Required env vars are declared in config

```bash
modkit doctor
```

This checks your local environment: Go version, Bun version, Docker, registry access.

---

## Common Makefile Commands

```bash
make dev          # start API + web concurrently
make build        # build API binary + web bundle
make test         # run Go tests (or bun test)
make lint         # go vet + eslint
make type-check   # tsc --noEmit on the web app
make migrate      # apply pending DB migrations
make infra-up     # start Docker services
make infra-down   # stop Docker services
```

---

## Next Steps

**Add database migrations:**
Create SQL files in `apps/api/migrations/` and run `make migrate`.

**Learn the wiring rules:**
Read `orchestration/composition-rulebook.md` — it covers all the patterns for auth, jobs, caching, error handling, and more, with Go and TypeScript examples.

**Build with AI assistance:**
The project includes a `CLAUDE.md` with instructions for Claude Code. Load it in a new Claude Code session and follow the orchestration playbook (`orchestration/playbook.md`) for a guided build.

**Add more modules:**
```bash
modkit list              # see what's available
modkit info realtime     # learn about a module
modkit pull realtime     # add it
```

**Deploy:**
The `cicd` module generates `.github/workflows/` with CI, staging deploy, and production deploy pipelines. Push to GitHub and configure the required secrets in your repo settings.
