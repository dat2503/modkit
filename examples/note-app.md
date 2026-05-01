# Build: Simple Note App

> **Instructions for a Claude Code agent.** Open this file in a new project folder and follow the steps below. The modkit registry lives at `D:\desperatelylazy\saas` (or wherever you cloned it).

---

## Step 0 — Load the Kit

Before writing any code, load the agent instructions:

```
Read: D:\desperatelylazy\saas\docs\agent-instructions.md
Read: D:\desperatelylazy\saas\orchestration\playbook.md
Read: D:\desperatelylazy\saas\orchestration\composition-rulebook.md
```

Install the CLI if not already installed:
```bash
go install github.com/dat2503/modkit/modkit@latest
```

---

## Step 1 — Project Brief

```yaml
project_name: "notepad"
domain: "productivity / personal notes"
runtime: go   # change to "bun" if you prefer TypeScript backend

entities:
  - name: Note
    fields: [id (uuid), user_id, title, content (text), created_at, updated_at, deleted_at]

user_roles:
  - user:
      - create a note
      - list their notes
      - view a single note
      - edit a note (title + content)
      - delete a note

key_flows:
  - "User signs up → lands on dashboard → sees empty notes list"
  - "User creates note → fills title + content → saves → redirected to note view"
  - "User edits note → changes content → saves"
  - "User deletes note → note disappears from list"
  - "User signs out → signs back in → notes are still there"

assumptions:
  - Notes are private — each user only sees their own notes
  - No sharing, collaboration, or public links
  - No folders or tags (keep it simple)
  - Soft delete — deleted notes are not hard-removed from DB
  - No rich text — plain text content only
  - No pagination needed for MVP (users unlikely to have >50 notes in testing)
```

---

## Step 2 — Module Selection

Selected modules and rationale:

| Module | Include | Rationale |
|--------|---------|-----------|
| `observability` | ✅ always | Required |
| `error-tracking` | ✅ always | Required |
| `auth` | ✅ | Users need accounts; all routes are private |
| `cache` | ✅ | Required by auth for session storage |
| `email` | ❌ | No transactional emails needed for MVP |
| `payments` | ❌ | Free app |
| `storage` | ❌ | No file uploads |
| `jobs` | ❌ | No async operations |
| `realtime` | ❌ | No live updates needed |
| `search` | ❌ | <100 notes per user; SQL is fine |
| `feature-flags` | ❌ | No staged rollouts |
| `cicd` | ✅ | Generate GitHub Actions workflows |

**Final module list:** `auth,cache,cicd`

---

## Step 3 — Scaffold the Project

Run from the parent directory of where you want the project:

```bash
# Go runtime
modkit init \
  --name notepad \
  --runtime go \
  --go-module github.com/you/notepad \
  --modules auth,cache,cicd \
  --no-prompt

# Bun runtime
modkit init \
  --name notepad \
  --runtime bun \
  --modules auth,cache,cicd \
  --no-prompt
```

Then enter the project directory:
```bash
cd notepad
```

---

## Step 4 — Architecture

### Database schema

```sql
-- migrations/001_create_notes.sql
CREATE TABLE notes (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     TEXT        NOT NULL,           -- Clerk user ID (e.g. user_abc123)
  title       TEXT        NOT NULL DEFAULT '',
  content     TEXT        NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ                     -- soft delete
);

CREATE INDEX notes_user_id_idx ON notes (user_id) WHERE deleted_at IS NULL;
```

### API routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | none | Health check |
| GET | `/api/notes` | required | List user's notes (excluding deleted) |
| POST | `/api/notes` | required | Create a note |
| GET | `/api/notes/:id` | required | Get a single note |
| PUT | `/api/notes/:id` | required | Update a note |
| DELETE | `/api/notes/:id` | required | Soft-delete a note |

All responses: `{"data": ...}` on success, `{"error": "..."}` on failure.

### Frontend pages

| Route | Description |
|-------|-------------|
| `/` | Landing page with sign-in button (redirect to `/notes` if signed in) |
| `/notes` | Notes list (protected) |
| `/notes/new` | Create note form (protected) |
| `/notes/[id]` | View + edit note (protected) |

### Module wiring (bootstrap order)

```
1. observability  (OtelObservabilityService)
2. error-tracking (SentryErrorTrackingService)
3. cache          (RedisCacheService)
4. auth           (ClerkAuthService)
```

---

## Step 5 — Implementation Checklist

Work through these in order. Run `modkit validate` after each group.

### 5a — Infrastructure
- [ ] `cp .env.example .env` and fill in keys (Clerk + Sentry — Redis/Postgres come from Docker)
- [ ] `make setup` (starts Postgres + Redis, installs deps)
- [ ] Verify: `curl http://localhost:8080/health` returns `{"status":"ok"}`

### 5b — Database
- [ ] Create `apps/api/migrations/001_create_notes.sql` with the schema above
- [ ] Create `apps/api/cmd/migrate/main.go` — reads `DATABASE_URL`, applies SQL files in order
- [ ] Run `make migrate`
- [ ] Add `DATABASE_URL` handling to `config/config.go` if not already there

### 5c — Note service (Go)
Create `apps/api/service/notes.go`:
- [ ] `NoteService` struct — takes `db *sql.DB` and `auth contracts.AuthService`
- [ ] `List(ctx, userID string) ([]Note, error)` — SELECT WHERE user_id = $1 AND deleted_at IS NULL ORDER BY updated_at DESC
- [ ] `Create(ctx, userID, title, content string) (*Note, error)` — INSERT
- [ ] `Get(ctx, id, userID string) (*Note, error)` — SELECT with ownership check
- [ ] `Update(ctx, id, userID, title, content string) (*Note, error)` — UPDATE with ownership check
- [ ] `Delete(ctx, id, userID string) error` — UPDATE SET deleted_at = now()

### 5d — API handlers (Go)
Add to `apps/api/api/router.go`:
- [ ] Wire `NoteService` into `Deps` struct
- [ ] `GET /api/notes` → `authRequired` → list notes for `UserFromContext`
- [ ] `POST /api/notes` → `authRequired` → decode body → create note
- [ ] `GET /api/notes/{id}` → `authRequired` → get note (404 if not found or wrong user)
- [ ] `PUT /api/notes/{id}` → `authRequired` → decode body → update note
- [ ] `DELETE /api/notes/{id}` → `authRequired` → soft-delete note

### 5e — Frontend pages (Next.js)
- [ ] `src/app/notes/page.tsx` — fetch `GET /api/notes`, list with title + date, link to each note, "New Note" button
- [ ] `src/app/notes/new/page.tsx` — form (title + textarea), POST to `/api/notes`, redirect to `/notes/[id]` on success
- [ ] `src/app/notes/[id]/page.tsx` — fetch note, show content, edit in-place, save button (PUT), delete button (DELETE → redirect to `/notes`)
- [ ] Protect all `/notes/*` routes with Clerk's `<SignedIn>` / redirect
- [ ] Update `src/app/page.tsx` — landing with sign-in button, redirect to `/notes` if already signed in

---

## Step 6 — Environment Variables Needed

Fill these in `.env`:

```bash
# Clerk — get from https://clerk.com (free)
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/notes
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/notes

# Sentry — get from https://sentry.io (free) or set to a dummy value to skip
SENTRY_DSN=https://dummy@sentry.io/0
SENTRY_ENVIRONMENT=development

# These come from Docker — leave as-is
DATABASE_URL=postgres://notepad:secret@localhost:5432/notepad_dev?sslmode=disable
REDIS_URL=redis://localhost:6379

# App
PORT=8080
APP_ENV=development
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Step 7 — Run and Validate

```bash
make dev         # API on :8080, web on :3000

# In another terminal:
modkit validate  # check module wiring
modkit doctor    # check environment

# Smoke tests:
curl http://localhost:8080/health
# Open http://localhost:3000 — sign up, create a note, reload, note is still there
```

---

## What to Verify (Testing Checklist)

- [ ] Sign up creates a new user in Clerk dashboard
- [ ] Sign in works after signing out
- [ ] Creating a note shows it in the list
- [ ] Editing a note updates the content
- [ ] Deleting a note removes it from the list
- [ ] Notes from one user are NOT visible when signed in as another user
- [ ] `GET /api/notes` without a Bearer token returns `{"error":"unauthorized"}`
- [ ] `GET /api/notes/{id}` for another user's note returns 404

---

## Troubleshooting

**`config: missing required environment variables: CLERK_SECRET_KEY`**
→ Your `.env` file is missing or not loaded. Check that `godotenv` (or similar) is loading it.

**`dial tcp 127.0.0.1:5432: connect: connection refused`**
→ Postgres isn't running. Run `make infra-up`.

**Clerk 401 on all requests**
→ Make sure you're passing the session token as `Authorization: Bearer <token>` and that `CLERK_SECRET_KEY` matches the frontend `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` (same Clerk app).

**Next.js can't reach the API**
→ Check `NEXT_PUBLIC_API_URL=http://localhost:8080` is set and the API is running.
