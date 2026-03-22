# Orchestration Playbook v1.0

> **This document is authoritative agent context.** Load it before starting any new project. Follow all phases in order. Do not skip phases. Do not proceed past a 🔒 checkpoint without human confirmation.

---

## Overview

This playbook defines the **6-phase workflow** for building a web application using the module registry system. An agent (Claude Code, or any orchestration system) follows these phases sequentially. At each 🔒 checkpoint, present output to the human and wait for explicit approval before proceeding.

**The goal:** Go from a plain-English project idea to a deployed, working application — with human oversight at key decision points and agents handling the mechanical work.

**Phases:**

| # | Phase | Agent does | Human does |
|---|-------|-----------|-----------|
| 0 | Intake | Parse idea → structured brief | Review, answer questions |
| 1 | Module Selection | Select modules from registry | Approve/adjust selection |
| 2 | Architecture Plan | Generate schema, routes, wiring | Approve plan before any code |
| 3 | Scaffold & Wire | Build the application | Monitor progress |
| 4 | Validate | Verify everything works | Review report, test locally |
| 5 | Deploy | Ship to staging → production | Approve each deployment |

---

## Approval Signals

At every 🔒 checkpoint, the agent must present its output and wait. The following signals indicate the human's intent:

**Approval (proceed to next phase):**
- "looks good", "approved", "lgtm", "proceed", "go ahead", "ship it", "yes", "ok"
- Any message that acknowledges the output without requesting changes

**Rejection (revise current phase):**
- "no", "change X", "wrong", "redo", "fix X", specific feedback on what to change
- Any message requesting modifications

**Ambiguous (ask for clarification):**
- "hmm", "maybe", "not sure", "I think so"
- If the signal is unclear, ask: "Should I proceed to the next phase, or would you like changes?"

**Rule:** When in doubt, do NOT proceed. It is always better to ask than to assume approval.

---

## Before You Start

Load the following context before any phase:

```bash
# 1. Load this playbook (you're reading it)

# 2. Load the composition rulebook
cat ~/.modkit/cache/orchestration/composition-rulebook.md

# 3. Load the module registry
modkit list --output json
```

If `modkit` is not installed or the registry cache is missing:
```bash
go install github.com/yourorg/module-registry/modkit@latest
git clone https://github.com/yourorg/module-registry ~/.modkit/cache
```

---

## Phase 0 — Intake [MVP]

**Goal:** Transform the user's plain-English idea into a structured, unambiguous project brief.

### Agent Actions

1. Read the user's project description carefully
2. Extract and structure:
   - **Project name** — short, lowercase, suitable for a directory name
   - **Domain** — business domain (e.g., "fintech / invoicing")
   - **Entities** — data objects with fields
   - **User roles** — who uses the system and what they can do
   - **Key flows** — the 3–5 core user journeys (not technical flows — user actions)
   - **Assumptions** — things you inferred but weren't explicitly stated
3. List all **ambiguities** — questions that, if answered differently, would change the architecture
4. Present the brief and ambiguity list to the human

### Output Format

```yaml
# project-brief.yaml — generated during Phase 0
project_name: "invoicely"
domain: "fintech / invoicing"
runtime: go  # or bun — ask the human if not specified

entities:
  - name: Invoice
    fields: [id, freelancer_id, client_email, items, total, status, due_date, pdf_url]
  - name: InvoiceItem
    fields: [description, quantity, rate, amount]
  - name: Client
    fields: [email, name, company]

user_roles:
  - freelancer: [create_invoice, send_invoice, track_payments, download_pdf]
  - client: [view_invoice, pay_invoice]

key_flows:
  - "Freelancer creates invoice → sends to client via email"
  - "Client receives email → clicks link → views invoice → pays via Stripe"
  - "Freelancer sees payment status update"
  - "Freelancer downloads invoice as PDF"

assumptions:
  - "Clients do not need accounts — they access via unique invoice links"
  - "PDF generation happens server-side"

ambiguities:
  - "Does the freelancer need recurring invoices?"
  - "Multi-currency support?"
  - "Tax calculation?"
  - "Dashboard with analytics?"
```

### 🔒 Human Checkpoint — Brief Review

Present the brief and ambiguity list. Wait for:
- Answers to ambiguity questions
- Corrections to entities, roles, or flows
- Explicit "looks good, proceed" or equivalent

**Do not proceed to Phase 1 until the brief is approved.**

Update `project-brief.yaml` with human feedback before moving on.

### Escalation

If the idea is too vague to produce a brief (no clear entities, no user roles, no flows), ask the human for 3–5 more sentences of context before attempting the brief. Never guess at business logic.

---

## Phase 1 — Module Selection [MVP]

**Goal:** Select the right modules from the registry for this specific project.

### Agent Actions

1. Read the module registry:
   ```bash
   modkit list --output json
   ```
2. Read agent docs for modules that might be relevant:
   ```bash
   modkit info auth --agent
   modkit info payments --agent
   modkit info email --agent
   modkit info storage --agent
   modkit info jobs --agent
   modkit info realtime --agent
   modkit info search --agent
   modkit info feature-flags --agent
   ```
3. Based on the approved project brief, decide which modules to include and which to skip
4. For each selected module, choose which implementation to use (use the default unless the brief specifies otherwise)
5. Generate the module selection manifest

### Output Format

```yaml
# module-selection.yaml — generated during Phase 1
runtime: go

selected_modules:
  - name: auth
    impl: clerk
    reason: "user accounts needed for freelancers"
  - name: payments
    impl: stripe
    reason: "client pays invoices"
  - name: email
    impl: resend
    reason: "send invoice links to clients"
  - name: storage
    impl: s3
    reason: "store generated PDFs"
  - name: cache
    impl: redis
    reason: "session management + dashboard caching"
  - name: observability
    impl: otel
    reason: "always included"
  - name: error-tracking
    impl: sentry
    reason: "always included"
  - name: jobs
    impl: asynq
    reason: "async PDF generation + email sending"

skipped_modules:
  - name: realtime
    reason: "not needed for MVP — freelancer will refresh manually"
  - name: search
    reason: "not needed — freelancers have few invoices"
  - name: feature-flags
    reason: "not needed for MVP"
```

### Selection Rules

| Module | Include when |
|--------|-------------|
| `observability` | **Always** — non-negotiable |
| `error-tracking` | **Always** — non-negotiable |
| `auth` | Any app with user accounts |
| `payments` | Any app with paid transactions |
| `email` | Transactional emails needed |
| `storage` | File uploads or generated files |
| `cache` | Sessions, rate limiting, repeated reads |
| `jobs` | Any operation > 500ms or needing retry |
| `realtime` | Brief explicitly mentions live updates |
| `search` | Search/filter across many records |
| `feature-flags` | Phased rollout or A/B testing mentioned |
| `cicd` | Always — generated by `modkit init` |

### 🔒 Human Checkpoint — Module Review

Present the selection with rationale. Wait for human to:
- Approve the selection
- Add or remove modules
- Change implementation choices

---

## Phase 2 — Architecture Plan [MVP]

**Goal:** Generate the complete architecture — database, API, module wiring, frontend pages — before writing any code.

### Agent Actions

1. Re-read the composition rulebook (§9 for DB, §4 for API, §3 for init order, §10 for frontend)
2. Generate the architecture plan:

**Database schema** — all tables with columns, types, indexes, FK relationships
**API routes** — method, path, auth requirement, description
**Module wiring order** — initialization sequence following rulebook §3
**Frontend pages** — route, auth requirement, component description
**Communication map** — data flow diagram in text form

### Output Format

```yaml
# architecture-plan.yaml — generated during Phase 2

database:
  - table: invoices
    columns:
      - id: uuid DEFAULT gen_random_uuid() PRIMARY KEY
      - freelancer_id: uuid NOT NULL REFERENCES users(id)
      - client_email: text NOT NULL
      - status: text NOT NULL DEFAULT 'draft'  # draft|sent|paid|overdue
      - total: numeric(10,2) NOT NULL DEFAULT 0
      - due_date: date
      - pdf_url: text
      - public_token: text UNIQUE
      - created_at: timestamptz NOT NULL DEFAULT NOW()
      - updated_at: timestamptz NOT NULL DEFAULT NOW()
      - deleted_at: timestamptz
    indexes:
      - idx_invoices_freelancer_id
      - idx_invoices_status
      - idx_invoices_public_token
  - table: invoice_items
    columns:
      - id: uuid DEFAULT gen_random_uuid() PRIMARY KEY
      - invoice_id: uuid NOT NULL REFERENCES invoices(id)
      - description: text NOT NULL
      - quantity: int NOT NULL DEFAULT 1
      - rate: numeric(10,2) NOT NULL
      - amount: numeric(10,2) NOT NULL
      - created_at: timestamptz NOT NULL DEFAULT NOW()

api_routes:
  - method: GET
    path: /api/v1/invoices
    auth: freelancer
    description: List invoices for the authenticated freelancer
  - method: POST
    path: /api/v1/invoices
    auth: freelancer
    description: Create a new invoice draft
  - method: GET
    path: /api/v1/invoices/:id
    auth: freelancer
    description: Get invoice detail
  - method: PATCH
    path: /api/v1/invoices/:id
    auth: freelancer
    description: Update invoice fields
  - method: POST
    path: /api/v1/invoices/:id/send
    auth: freelancer
    description: Send invoice to client via email
  - method: GET
    path: /api/v1/public/:token
    auth: none
    description: View invoice by public token (client access)
  - method: POST
    path: /api/v1/public/:token/pay
    auth: none
    description: Initiate payment for an invoice (creates Stripe checkout)
  - method: POST
    path: /api/v1/webhooks/stripe
    auth: none
    description: Handle Stripe payment webhook
  - method: GET
    path: /api/v1/dashboard
    auth: freelancer
    description: Get dashboard stats (total earned, pending, overdue)

module_wiring_order:
  1: observability
  2: error-tracking
  3: cache
  4: auth
  5: storage
  6: email
  7: payments
  8: jobs

frontend_pages:
  - path: /
    auth: none
    description: Landing page
  - path: /sign-in
    auth: none
    description: Sign-in page (Clerk component)
  - path: /dashboard
    auth: freelancer
    description: Invoice summary with stats
  - path: /invoices
    auth: freelancer
    description: Invoice list with filters
  - path: /invoices/new
    auth: freelancer
    description: Create invoice form
  - path: /invoices/:id
    auth: freelancer
    description: Invoice detail with send/download actions
  - path: /pay/:token
    auth: none
    description: Public payment page for clients
```

### 🔒 Human Checkpoint — Architecture Approval

**This is the most critical checkpoint.** Present the full architecture plan.

Wait for human to:
- Approve the database schema
- Adjust routes or pages
- Explicit **"approved, proceed to build"**

**Do not write a single line of application code until this is approved.** Changes at this stage are cheap. Changes after code is written are expensive.

---

## Phase 3 — Scaffold & Wire [MVP]

**Goal:** Build the application using modkit and write application code.

Phase 3 is split into sub-phases, each ending with a verifiable milestone. Complete each sub-phase fully before moving to the next.

---

### Phase 3a — Scaffold & Infrastructure

**Step 3.1 — Scaffold the project**
```bash
modkit init \
  --name {project_name} \
  --runtime {runtime} \
  --go-module github.com/{org}/{project_name} \
  --modules {comma-separated module:impl pairs} \
  --no-prompt

cd {project_name}
```

**Step 3.2 — Setup infrastructure and verify**
```bash
cp .env.example .env   # fill in real keys
make setup             # starts Postgres + Redis, installs deps
make dev-api           # start the API
```

**Milestone 3a:** `curl http://localhost:8080/health` returns `{"status":"ok"}`. If it doesn't, stop and fix before proceeding.

---

### Phase 3b — Database & Migrations

**Step 3.3 — Write database migrations**

Create files in `infra/migrations/` named `{timestamp}_{description}.sql`.
Use the schema from the approved architecture plan exactly.

```sql
-- infra/migrations/20240101120000_create_invoices.sql
CREATE TABLE invoices (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  freelancer_id UUID NOT NULL REFERENCES users(id),
  -- ...
);
```

**Step 3.4 — Run migrations**
```bash
make migrate
```

**Milestone 3b:** Migrations apply without errors. Verify tables exist with a quick SQL check or `make migrate` exits 0.

---

### Phase 3c — API Handlers & Backend Logic

**Step 3.5 — Write the module bootstrap**

`apps/api/bootstrap.go` (Go) or `apps/api/bootstrap.ts` (Bun) initializes all modules in the wiring order from the architecture plan.

**Step 3.6 — Generate the OpenAPI spec**

Create `apps/api/openapi.yaml` with all routes from the architecture plan. Use OpenAPI 3.1 format.

**Step 3.7 — Write API handlers**

For each route in the architecture plan, write the handler. Follow composition rulebook §4 (REST/OpenAPI) and §12 (error handling).

**Step 3.8 — Generate typed API client** [MVP]

```bash
# Go
oapi-codegen -generate server openapi.yaml > apps/api/gen/server.go

# TypeScript
bunx openapi-typescript apps/api/openapi.yaml -o apps/web/lib/api.types.ts
```

**Step 3.9 — Write background jobs** (if `jobs` module selected) [MVP when jobs included]

Create job definitions in `apps/api/jobs/`. Each job must be idempotent (safe to retry).

**Milestone 3c:** API builds without errors, smoke test auth works:
```bash
# Verify build
cd apps/api && go build ./... (or bun run build)

# Verify health + ready
curl http://localhost:8080/health
curl http://localhost:8080/ready

# Verify auth rejects unauthenticated requests
curl -s http://localhost:8080/api/v1/{any_protected_route} | grep -q "unauthorized"
```

---

### Phase 3d — Frontend & Tests

**Step 3.10 — Write Next.js pages**

For each page in the architecture plan. Use server components for initial data, client components for interactivity. Place protected pages under `src/app/(protected)/`.

**Step 3.11 — Write tests**

- Unit tests for all handlers (mock modules at the interface level)
- Integration tests for the 2–3 critical flows

**Milestone 3d:** Frontend loads at `http://localhost:3000`, sign-in page renders, protected routes redirect to sign-in when not authenticated.

---

### Validate as you go

After each handler is written:
```bash
modkit validate --output json
```

Fix any validation failures before continuing. Never accumulate validation failures.

### Escalation

If you encounter a decision not covered by the architecture plan:
1. **Stop writing code**
2. List all open questions as a numbered list
3. Present them to the human and wait for answers
4. Never guess at business logic or data relationships

---

## Phase 4 — Validate [MVP]

**Goal:** Verify the application builds, tests pass, and critical flows work.

### Agent Actions

```bash
# 1. Check module wiring
modkit validate --output json

# 2. Check overall health
modkit doctor --output json

# 3. Build
make build

# 4. Run tests
make test

# 5. Start local services
docker-compose up -d
sleep 5  # wait for services

# 6. Smoke test
curl -sf http://localhost:8080/health || echo "FAIL: health check"
```

### Output Format

```yaml
# validation-report.yaml — generated during Phase 4
validation:
  modkit_validate: pass
  modkit_doctor: pass
  build: pass
  test_results:
    total: 47
    passed: 47
    failed: 0
    coverage: 74%
  smoke_tests:
    - test: "GET /health"
      status: 200
      result: pass
    - test: "POST /api/v1/invoices (auth required)"
      status: 401
      result: pass
issues: []
```

### 🔒 Human Checkpoint — Validation Review

Present the validation report. Ask the human to:
1. Pull the repo locally
2. Run `make dev`
3. Test the critical flows manually

Wait for: "looks good, ship to staging"

---

## Phase 5 — Deploy [MVP]

**Goal:** Deploy to staging, then production.

### Agent Actions

**Step 5.1 — Configure CI/CD** [MVP]

Ensure `.github/workflows/` contains:
- `ci.yaml` — build + test on every PR
- `deploy-staging.yaml` — auto-deploy on merge to `main`
- `deploy-production.yaml` — deploy on version tag `v*`

**Step 5.2 — Deploy to staging**

```bash
git push origin main
# Wait for CI pipeline to complete
# Check staging URL
```

### 🔒 Human Checkpoint — Staging Approval

Human reviews staging environment:
- Verify all key flows work end to end
- Check error tracking dashboard for any errors
- Confirm "approved for production"

### Step 5.3 — Deploy to production

Only after explicit staging approval:
```bash
git tag v1.0.0
git push origin v1.0.0
```

### 🔒 Human Checkpoint — Production Confirmation

Confirm the production deploy succeeded. Monitor error tracking for 15 minutes post-deploy.

---

## Future Phases (Deferred)

These are not part of the MVP workflow. Plan and implement separately:

### v2 — Advanced Agent Swarm [LATER]
Split Phase 3 across multiple specialist agents (Backend Agent, Frontend Agent, Database Agent, Testing Agent) working in parallel. Requires stable contract layer and solid Phase 0–2 outputs to coordinate between agents.

### v2 — Feature Flag Gating [LATER]
Integrate `feature-flags` module into Phase 3 scaffold to gate new features behind flags. Add flag evaluation step to Phase 1 module selection.

### v2 — Automated Staging Tests [LATER]
Add automated smoke test suite that runs against staging before the Phase 5 human checkpoint. Replace manual testing with a test report.

### v2 — Multi-region Deploy [LATER]
Extend Phase 5 to support multi-region infrastructure. Requires IaC (Terraform/Pulumi) module.

---

## Error Recovery

If any phase fails:

| Failure | Recovery |
|---------|---------|
| Build failure | Fix compilation errors before proceeding |
| Test failure | Diagnose root cause — do not skip tests |
| Validation failure | Run `modkit doctor`, fix wiring issues |
| Human rejects checkpoint | Apply feedback, re-run the failed phase |
| Agent gets stuck | Stop, list open questions, present to human |

**Never proceed to the next phase while the current phase has unresolved failures.**

### Common failure runbooks

**Database migration fails:**
1. Read the error message — common causes: syntax error, missing referenced table, duplicate column
2. Check migration ordering — migrations run alphabetically by filename timestamp
3. If the schema is wrong: fix the migration SQL, run `make migrate-down` then `make migrate`
4. If the database is in a bad state: drop and recreate via `docker compose down -v && make infra-up && make migrate`
5. Never edit a migration that has been applied to a shared database — create a new migration instead

**Auth keys are invalid (Clerk 401 on all requests):**
1. Verify `CLERK_SECRET_KEY` and `CLERK_PUBLISHABLE_KEY` are from the **same** Clerk application
2. Check the key prefix: `sk_test_` for development, `sk_live_` for production — they are not interchangeable
3. If using webhooks, verify `CLERK_WEBHOOK_SECRET` matches the webhook endpoint configured in the Clerk dashboard
4. Test the key directly: `curl -H "Authorization: Bearer sk_test_..." https://api.clerk.com/v1/users?limit=1`

**API returns wrong status codes or response shapes:**
1. Compare the handler response against the OpenAPI spec from Phase 2
2. Check that error responses use `writeError()` (Go) or `c.json({error:...}, status)` (Bun) — never `http.Error()`
3. Verify the response envelope: success = `{"data": ...}`, error = `{"error": {"message": "..."}}`
4. If a handler returns 200 for errors or 500 for validation failures, fix the status code mapping

**Deploy fails halfway (Phase 5):**
1. Check the deployment logs for the specific failure point
2. If the container fails to start: check `docker logs` — usually a missing env var or failed database connection
3. If migrations fail on staging: do NOT retry blindly — check if a partial migration was applied
4. If the health check fails after deploy: verify `ALLOWED_ORIGINS`, `DATABASE_URL`, `REDIS_URL` are set for the target environment
5. If rollback is needed: redeploy the previous known-good image/commit — do not attempt manual patches in production

**Cache connection refused:**
1. Verify Redis is running: `docker compose ps` should show the redis container as healthy
2. Check `REDIS_URL` — default is `redis://localhost:6379` for local development
3. If Redis is up but connection fails: check firewall rules or Docker network configuration
4. The app should degrade gracefully (auth falls back to direct API calls) — but performance will suffer
