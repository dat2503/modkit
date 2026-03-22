# Composition Rulebook v1.0

> **This document is authoritative agent context.** Read it in full before writing any code. All rules marked `[MVP]` are mandatory. Rules marked `[OPTIONAL]` are recommended but not enforced. Rules marked `[LATER]` are for future phases.

---

## §1 Interface-First Design [MVP]

Every module concern must be expressed as an interface before any implementation is written.

**Go:** Define a Go interface in `contracts/go/{module}.go`. Every struct or package that provides this capability must satisfy the interface. Application code imports and uses only the interface — never the concrete type.

**TypeScript:** Define a TypeScript interface in `contracts/ts/{module}.ts`. Every class that provides this capability must implement the interface.

```go
// ✅ Correct — depend on the interface
type InvoiceService struct {
    email contracts.EmailService
}

// ❌ Wrong — depend on the implementation
type InvoiceService struct {
    email *resend.Client
}
```

```typescript
// ✅ Correct
class InvoiceService {
  constructor(private readonly email: IEmailService) {}
}

// ❌ Wrong
class InvoiceService {
  constructor(private readonly email: ResendClient) {}
}
```

**Why this matters:** When you need to swap Resend for SendGrid, you build a new `SendGridEmailService` that satisfies `IEmailService`. Nothing else changes. The interface is the contract — implementations are interchangeable.

**Rule:** Never import an implementation package (`modules/email/impl/resend-go`) directly in `apps/api/` code. Always import from `contracts/`.

---

## §2 Dependency Injection [MVP]

All modules are initialized once at startup and injected into components that need them. No global singletons, no `init()` magic.

**Go pattern — constructor injection:**
```go
// Correct ✅
func NewInvoiceHandler(
    db *sql.DB,
    email contracts.EmailService,
    payments contracts.PaymentsService,
    jobs contracts.JobsService,
) *InvoiceHandler {
    return &InvoiceHandler{
        db:       db,
        email:    email,
        payments: payments,
        jobs:     jobs,
    }
}
```

**TypeScript pattern — constructor injection:**
```typescript
// Correct ✅
class InvoiceHandler {
  constructor(
    private readonly db: Database,
    private readonly email: IEmailService,
    private readonly payments: IPaymentsService,
    private readonly jobs: IJobsService,
  ) {}
}
```

**Bootstrap file** wires everything together:
```go
// apps/api/bootstrap.go
func Bootstrap(cfg *Config) (*App, error) {
    obs, err := otel.New(cfg.Observability)
    if err != nil { return nil, err }

    errTrack := sentry.New(cfg.Sentry)
    cache := redis.New(cfg.Redis)
    auth := clerk.New(cfg.Clerk)
    // ... etc

    invoiceHandler := handlers.NewInvoiceHandler(db, email, payments, jobs)
    // ...
    return &App{...}, nil
}
```

**Rules:**
- Config is always injected via constructor, never read from `os.Getenv()` inside module code
- Never use package-level variables for module instances
- Never call `init()` to set up module state

---

## §3 Module Initialization Order [MVP]

Modules must be initialized in this exact order. Each depends only on modules initialized before it.

```
1. observability    ← wraps everything, must be first
2. error-tracking   ← needs observability for trace context
3. cache            ← used by auth (sessions) and jobs (queue)
4. auth             ← needs cache for session storage
5. storage          ← independent
6. email            ← independent, pairs with jobs
7. payments         ← independent
8. search           ← independent [v2]
9. feature-flags    ← independent, may cache results [v2]
10. jobs            ← needs cache (queue backend), pairs with email/storage
11. realtime        ← needs cache and auth [v2]
```

**Why this order?**
- `observability` must be first because it wraps all downstream calls with trace context
- `error-tracking` must be second because it needs the trace IDs from observability
- `cache` before `auth` because Clerk/Privy store sessions in Redis
- `jobs` after `email`/`storage` because job handlers often call email and storage

**Go bootstrap order:**
```go
obs   := initObservability(cfg)
errs  := initErrorTracking(cfg, obs)
cache := initCache(cfg)
auth  := initAuth(cfg, cache)
store := initStorage(cfg)
email := initEmail(cfg)
pay   := initPayments(cfg)
jobs  := initJobs(cfg, cache)
// realtime last if included
```

**Rule:** Any deviation from this order requires explicit documentation of why. `modkit validate` will flag initialization order violations.

---

## §4 REST/OpenAPI Communication [MVP]

All communication between the frontend (Vite/React or Next.js) and backend (Go/Bun) uses REST with an OpenAPI 3.1 specification as the single source of truth.

**Workflow (spec-first):**
1. Write `apps/api/openapi.yaml` first — all routes, request/response shapes, auth requirements
2. Generate server stubs from the spec:
   ```bash
   # Go
   oapi-codegen -generate server,types openapi.yaml > gen/server.go

   # TypeScript (Bun)
   bunx openapi-typescript openapi.yaml -o gen/api.types.ts
   ```
3. Generate the typed frontend API client from the same spec:
   ```bash
   bunx openapi-typescript apps/api/openapi.yaml -o apps/web/lib/api.types.ts
   ```

**URL conventions:**
- All routes under `/api/v1/`
- Plural nouns: `/api/v1/invoices` not `/api/v1/invoice`
- Nested resources: `/api/v1/invoices/:id/items`
- Verbs only when necessary for non-CRUD actions: `/api/v1/invoices/:id/send`
- Kebab-case for multi-word resources: `/api/v1/invoice-items`

**Standard response shapes:**

Success (single resource):
```json
{ "data": { "id": "...", "...": "..." } }
```

Success (list):
```json
{
  "data": [...],
  "pagination": { "total": 100, "page": 1, "per_page": 20 }
}
```

Error:
```json
{
  "error": {
    "code": "INVOICE_NOT_FOUND",
    "message": "Invoice not found",
    "details": {}
  }
}
```

**HTTP status codes:**
- `200` — success
- `201` — created
- `400` — bad request (validation error)
- `401` — unauthenticated
- `403` — unauthorized (authenticated but insufficient permissions)
- `404` — not found
- `409` — conflict
- `422` — unprocessable entity
- `500` — internal server error

**Rules:**
- Never return raw objects — always wrap in `{ "data": ... }` or `{ "error": ... }`
- Never return stack traces or internal error details to clients
- All API changes require updating the OpenAPI spec first

---

## §5 Async Jobs Pattern [MVP when jobs included]

Long-running operations, operations needing retry logic, and side effects that don't need to block the HTTP response must use the jobs module.

**When to use jobs (not direct calls):**
- Sending emails (> 100ms, external service, can fail)
- Generating PDFs or files
- Calling external APIs that may be slow
- Database operations on many rows
- Any side effect where the user doesn't need to wait

**Go (Asynq) pattern:**
```go
// Define the job payload type
type GeneratePDFPayload struct {
    InvoiceID string `json:"invoice_id"`
    UserID    string `json:"user_id"`
}

// In the handler — enqueue and return immediately
func (h *InvoiceHandler) SendInvoice(w http.ResponseWriter, r *http.Request) {
    invoiceID := r.PathValue("id")

    // Enqueue the work — do NOT wait for it
    _, err := h.jobs.Enqueue(r.Context(), "invoice:generate_pdf",
        GeneratePDFPayload{InvoiceID: invoiceID},
        contracts.WithMaxRetry(3),
        contracts.WithTimeout(30*time.Second),
    )
    if err != nil {
        writeError(w, err)
        return
    }

    writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

// The job handler — runs asynchronously
func HandleGeneratePDF(ctx context.Context, payload GeneratePDFPayload) error {
    // idempotent: safe to retry if it fails
    // ...
    return nil
}
```

**TypeScript (BullMQ) pattern:**
```typescript
// Enqueue
await jobs.enqueue('invoice:generate_pdf', { invoiceId }, { maxRetry: 3, timeoutMs: 30_000 })
res.status(202).json({ status: 'queued' })

// Worker (in a separate process or worker thread)
worker.process('invoice:generate_pdf', async (job) => {
  // idempotent implementation
})
```

**Job naming convention:** `{entity}:{action}` — e.g., `invoice:generate_pdf`, `user:send_welcome_email`

**Rules:**
- Job handlers must be **idempotent** — safe to run multiple times with the same payload
- Always set `MaxRetry` and `Timeout` explicitly
- Never call job handlers directly from handlers — always enqueue
- Job failure must not affect the HTTP response (return 202 Accepted, not 500)

---

## §6 Realtime Pattern [v2]

Real-time updates use WebSockets via the realtime module. The pattern is pub/sub: backend publishes events, frontend subscribes to channels.

**Channel naming:** `{entity}:{id}:{event}` — e.g., `invoice:abc123:status_changed`

**Go backend — publish:**
```go
err := h.realtime.Publish(ctx, "invoice:"+invoiceID+":status_changed", contracts.RealtimeEvent{
    Type:    "status_changed",
    Payload: map[string]any{"status": "paid", "paid_at": time.Now()},
})
```

**Next.js frontend — subscribe:**
```typescript
// In a client component
const { lastEvent } = useWebSocket(`/ws/invoice/${invoiceId}`)
useEffect(() => {
  if (lastEvent?.type === 'status_changed') {
    setStatus(lastEvent.payload.status)
  }
}, [lastEvent])
```

**WebSocket endpoint:** Mount at `/ws` with auth middleware. Validate token during the upgrade handshake.

**Rules:**
- WebSocket connections must be authenticated — validate the auth token during upgrade
- Never send sensitive data (payment details, PII) over WebSocket without justification
- Channels are per-resource — don't create a single global channel for all updates
- Always handle reconnection on the client side

---

## §7 Caching Strategy [MVP when cache included]

Use Redis cache for: session data, rate limiting counters, hot frequently-read data, job deduplication.

**Cache key naming:** `{service}:{entity}:{id}` — e.g., `invoice:detail:abc123`, `user:session:xyz`

**TTL guidelines:**
| Data type | TTL |
|-----------|-----|
| Sessions | 24h |
| Rate limit windows | 1m–1h |
| Frequently-read, rarely-changed data | 5–15m |
| Job deduplication keys | Match job timeout |
| Feature flag evaluations | 5m [v2] |

**Go pattern — cache-aside:**
```go
func (r *InvoiceRepo) GetByID(ctx context.Context, id string) (*Invoice, error) {
    key := "invoice:detail:" + id
    var invoice Invoice

    // Try cache first
    if err := r.cache.Get(ctx, key, &invoice); err == nil {
        return &invoice, nil
    }

    // Cache miss — query database
    invoice, err := r.db.QueryRow(...)
    if err != nil {
        return nil, err
    }

    // Populate cache for next request
    _ = r.cache.Set(ctx, key, invoice, 10*time.Minute)
    return &invoice, nil
}

// Invalidate on update
func (r *InvoiceRepo) Update(ctx context.Context, id string, ...) error {
    err := r.db.Update(...)
    if err == nil {
        _ = r.cache.Delete(ctx, "invoice:detail:"+id)
    }
    return err
}
```

**Rules:**
- Always invalidate cache entries when underlying data changes
- Never cache financial or auth-related data without a very short TTL (< 1m)
- Never serve stale cache for operations where consistency matters (payments, permissions)
- SetNX (set-if-not-exists) for distributed locks and deduplication

---

## §8 Auth / Authorization Flow [MVP when auth included]

Authentication is handled entirely by the auth module. Never build custom JWT signing or password hashing.

**Middleware chain (applied in this order):**
```
Request
  → CORS middleware
  → Request ID middleware (add trace ID to context)
  → Auth middleware (validate token → inject user into context)
  → Rate limit middleware [OPTIONAL]
  → Handler
```

**CORS middleware (always included):**

CORS is wired in the generated project by default. Configuration:
- `ALLOWED_ORIGINS` env var — comma-separated list of allowed origins (e.g., `http://localhost:3000,https://app.example.com`)
- Defaults to `*` (allow all) if not set
- In production, always set `ALLOWED_ORIGINS` to your actual frontend domain(s)
- Go: `corsMiddleware(allowedOrigins)` in `middleware.go`, wired outermost in `router.go`
- Bun: `cors()` from `hono/cors`, wired first in `router.ts`

**Auth middleware behavior:**
- Public routes: pass through without validation
- Protected routes: extract `Authorization: Bearer {token}` header, validate via `auth.ValidateToken()`, inject `AuthUser` into context
- Return `401 Unauthorized` for missing or expired token
- Return `403 Forbidden` if the user doesn't have the required role (check in handler, not middleware)

**Passing user context (Go):**
```go
type contextKey string
const UserContextKey contextKey = "auth_user"

// In auth middleware
ctx = context.WithValue(r.Context(), UserContextKey, authUser)

// In handler
user, ok := r.Context().Value(UserContextKey).(*contracts.AuthUser)
if !ok {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}
```

**Authorization in handlers:**
```go
// Role check in the handler — not in middleware
func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
    user := getUserFromContext(r.Context())
    invoiceID := r.PathValue("id")

    invoice, err := h.repo.GetByID(r.Context(), invoiceID)
    if err != nil { ... }

    // Ownership check
    if invoice.FreelancerID != user.ID {
        writeError(w, &AppError{Status: 403, Code: "FORBIDDEN"})
        return
    }
    // ...
}
```

**Rules:**
- Never trust client-provided user IDs — always extract from validated JWT in context
- Authentication (who are you?) is in middleware. Authorization (can you do this?) is in handlers.
- Never check roles in middleware for resource-specific permissions
- Every protected route must call `getUserFromContext()` and validate ownership/role

---

## §9 Database Conventions [MVP]

**Migration files:**
- Location: `infra/migrations/`
- Naming: `{timestamp}_{snake_case_description}.sql` — e.g., `20240101120000_create_invoices.sql`
- Each migration is **additive** — never drop or rename columns in a migration that production data depends on
- Use `UP` and `DOWN` migrations when supported by your migration tool

**Table naming:** `snake_case`, plural — `invoices`, `invoice_items`, `users`, `webhook_events`

**Column naming:** `snake_case` — `created_at`, `updated_at`, `freelancer_id`

**Primary keys:** Always `id UUID DEFAULT gen_random_uuid() PRIMARY KEY` — never integer sequences for new tables

**Required columns on every table:**
```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

**Soft deletes for user-generated content:**
```sql
deleted_at TIMESTAMPTZ  -- NULL means not deleted
```
Always filter `WHERE deleted_at IS NULL` in queries.

**Indexes:**
- Add index on every foreign key column
- Add index on columns used in frequent `WHERE` clauses
- Naming: `idx_{table}_{column(s)}` — e.g., `idx_invoices_freelancer_id`

**Repository pattern:**
```go
type InvoiceRepository interface {
    Create(ctx context.Context, invoice *Invoice) error
    GetByID(ctx context.Context, id string) (*Invoice, error)
    ListByFreelancer(ctx context.Context, freelancerID string, opts PaginationOpts) ([]*Invoice, int64, error)
    Update(ctx context.Context, id string, updates InvoiceUpdates) (*Invoice, error)
    SoftDelete(ctx context.Context, id string) error
}
```

**Rules:**
- Never write raw SQL in HTTP handlers — use the repository layer
- All queries must use parameterized inputs — no string concatenation in SQL
- Use `pgx` (Go) or `drizzle`/`prisma` (Bun) for database access
- Run migrations in CI before tests — never test against a fresh empty database

---

## §10a Frontend Structure — Vite + React [MVP, Default]

The default frontend is **Vite + React** with React Router and TanStack Query.

```
apps/web/
├── src/
│   ├── main.tsx              ← entry point (BrowserRouter + QueryClientProvider)
│   ├── App.tsx               ← root routes
│   ├── index.css             ← Tailwind directives
│   ├── pages/
│   │   ├── Home.tsx          ← public landing page
│   │   ├── Dashboard.tsx     ← protected dashboard
│   │   ├── SignIn.tsx        ← auth sign-in (if auth module selected)
│   │   └── ...               ← additional pages
│   ├── components/
│   │   ├── ProtectedRoute.tsx ← auth guard (redirects to /sign-in)
│   │   ├── ui/               ← generic, reusable components
│   │   └── features/         ← feature-specific components
│   ├── lib/
│   │   ├── api.ts            ← typed fetch client (api.get, api.post, etc.)
│   │   └── auth.tsx          ← Better Auth React client (if auth module selected)
│   └── hooks/                ← custom React hooks
│       └── use-invoices.ts
├── index.html                ← Vite entry HTML
├── vite.config.ts            ← Vite config (proxy /api to backend)
└── package.json
```

**Data fetching rules:**
- Use **TanStack Query** (`useQuery`, `useMutation`) for server data
- Use `api.get()` / `api.post()` from `lib/api.ts` — never raw `fetch` in components
- The Vite dev server proxies `/api/*` to the backend — no CORS issues in development

**Auth in Vite (Better Auth):**
```typescript
// lib/auth.tsx
import { createAuthClient } from 'better-auth/react'
export const authClient = createAuthClient({ baseURL: import.meta.env.VITE_API_URL })

// ProtectedRoute — wraps any route that requires authentication
import { Navigate } from 'react-router-dom'
import { authClient } from '../lib/auth'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { data: session, isPending } = authClient.useSession()
  if (isPending) return null
  if (!session) return <Navigate to="/sign-in" replace />
  return <>{children}</>
}

// Dashboard — access current user
const { data: session } = authClient.useSession()
const user = session?.user
```

**Routing pattern:**
```typescript
// App.tsx
<Routes>
  <Route path="/" element={<Home />} />
  <Route path="/sign-in" element={<SignIn />} />
  <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
  <Route path="/invoices" element={<ProtectedRoute><Invoices /></ProtectedRoute>} />
</Routes>
```

**Rules:**
- Never put API keys, secrets, or server-only config in frontend code
- Never call the database directly from frontend — always go through the API
- Use `VITE_` prefix for env vars (not `NEXT_PUBLIC_`)
- All protected routes must be wrapped in `<ProtectedRoute>`

---

## §10b Frontend Structure — Next.js [Alternative]

Use `--frontend next` when the app requires server-side rendering, SEO-critical pages, or Next.js-specific features.

```
apps/web/
├── src/
│   └── app/
│       ├── layout.tsx            ← root layout
│       ├── page.tsx              ← landing page
│       └── (protected)/
│           ├── layout.tsx        ← auth guard (server-side session check)
│           ├── dashboard/
│           │   └── page.tsx
│           └── invoices/
│               └── page.tsx
├── lib/
│   ├── api.ts                    ← typed fetch client
│   └── auth.ts                   ← Better Auth server-side helper
└── package.json
```

**Auth in Next.js (Better Auth):**
```typescript
// lib/auth.ts — server-side auth instance
import { betterAuth } from 'better-auth'
export const auth = betterAuth({ ... })

// (protected)/layout.tsx — server-side session check
import { headers } from 'next/headers'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'

export default async function ProtectedLayout({ children }) {
  const session = await auth.api.getSession({ headers: await headers() })
  if (!session) redirect('/sign-in')
  return <>{children}</>
}
```

**Rules:**
- Use server components for initial data — reduces client JS
- Pages under `(protected)/` are automatically guarded by the layout
- Use `NEXT_PUBLIC_` prefix for env vars exposed to the browser

---

## §11 Observability [MVP]

The observability module (OpenTelemetry) instruments all module calls and HTTP handlers.

**Structured logging (Go):**
```go
// Use slog with context for trace propagation
slog.InfoContext(ctx, "invoice created",
    slog.String("invoice_id", invoice.ID),
    slog.String("freelancer_id", invoice.FreelancerID),
    slog.Float64("total", invoice.Total),
)

// Error logging
slog.ErrorContext(ctx, "failed to send invoice email",
    slog.String("invoice_id", invoice.ID),
    slog.Any("error", err),
)
```

**Tracing (Go):**
```go
func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx, span := h.obs.StartSpan(r.Context(), "invoice.create")
    defer span(nil)  // pass error if one occurs: defer span(&err)

    // ... handler logic
}
```

**TypeScript logging:**
```typescript
obs.log('info', 'invoice created', { invoiceId: invoice.id, freelancerId: invoice.freelancerId })
obs.log('error', 'failed to send email', { invoiceId, error: err.message })
```

**What to log:**
- ✅ Request ID, user ID (when authenticated), entity IDs, operation names
- ✅ Durations for external calls (DB queries, API calls)
- ✅ Job enqueue/complete/fail events
- ❌ Never log passwords, payment card numbers, JWT tokens, or raw secrets
- ❌ Never log PII without a specific compliance justification

**Rules:**
- All HTTP handlers must be wrapped with observability middleware (logs request + response)
- All external service calls (DB, email, payments) must be wrapped in a span
- Log at `INFO` for normal operations, `WARN` for recoverable issues, `ERROR` for failures requiring action

---

## §12 Error Handling [MVP]

**Error types (Go):**
```go
type AppError struct {
    Code    string         `json:"code"`     // machine-readable: "INVOICE_NOT_FOUND"
    Message string         `json:"message"`  // user-facing: "Invoice not found"
    Status  int            `json:"-"`        // HTTP status code
    Err     error          `json:"-"`        // original error (for logs, not sent to client)
    Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Common constructors
func NotFound(entity, id string) *AppError {
    return &AppError{Code: strings.ToUpper(entity)+"_NOT_FOUND", Message: entity+" not found", Status: 404}
}
func Unauthorized() *AppError {
    return &AppError{Code: "UNAUTHORIZED", Message: "Authentication required", Status: 401}
}
func Forbidden() *AppError {
    return &AppError{Code: "FORBIDDEN", Message: "Access denied", Status: 403}
}
```

**Error propagation (Go):**
```go
// Wrap errors with context as they bubble up
if err := r.db.Query(...); err != nil {
    return nil, fmt.Errorf("invoice.repo.getByID: %w", err)
}

// At the HTTP handler boundary — convert to AppError and respond
func writeError(w http.ResponseWriter, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        w.WriteHeader(appErr.Status)
        json.NewEncoder(w).Encode(map[string]any{"error": appErr})
        return
    }
    // Unknown error — log it, return generic 500
    slog.Error("unhandled error", slog.Any("error", err))
    w.WriteHeader(500)
    json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{
        "code": "INTERNAL_ERROR", "message": "An unexpected error occurred",
    }})
}
```

**TypeScript error types:**
```typescript
export class AppError extends Error {
  constructor(
    public readonly code: string,
    public readonly message: string,
    public readonly status: number,
    public readonly details?: Record<string, unknown>,
  ) {
    super(message)
  }
}
```

**Rules:**
- Never expose raw Go/TS errors to API clients — always map to AppError
- Never swallow errors silently — either handle them or propagate them
- Error codes must be SCREAMING_SNAKE_CASE and describe the specific problem
- Report all unexpected (non-AppError) errors to the error-tracking module

---

## §13 Config Management [MVP]

All configuration comes from environment variables. No hardcoded values anywhere in application code.

**Config struct (Go):**
```go
type Config struct {
    // Server
    Port        int    `env:"PORT,required"`
    DatabaseURL string `env:"DATABASE_URL,required"`

    // Auth
    ClerkSecretKey string `env:"CLERK_SECRET_KEY,required"`
    ClerkPublicKey string `env:"CLERK_PUBLIC_KEY,required"`

    // Email
    EmailProvider string `env:"EMAIL_PROVIDER,required"` // resend|sendgrid
    EmailAPIKey   string `env:"EMAIL_API_KEY,required"`
    EmailFrom     string `env:"EMAIL_FROM_DEFAULT,required"`

    // Payments
    StripeSecretKey    string `env:"STRIPE_SECRET_KEY,required"`
    StripeWebhookSecret string `env:"STRIPE_WEBHOOK_SECRET,required"`

    // Cache
    RedisURL string `env:"REDIS_URL,required"`

    // Storage
    StorageProvider  string `env:"STORAGE_PROVIDER,required"` // s3
    StorageBucket    string `env:"STORAGE_BUCKET,required"`
    StorageRegion    string `env:"STORAGE_REGION,required"`
    StorageAccessKey string `env:"STORAGE_ACCESS_KEY,required"`
    StorageSecretKey string `env:"STORAGE_SECRET_KEY,required"`

    // Observability
    OtelEndpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
    OtelServiceName string `env:"OTEL_SERVICE_NAME,required"`

    // Error Tracking
    SentryDSN string `env:"SENTRY_DSN,required"`
}
```

**Validation at startup:**
```go
func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("config: %w", err)  // fail fast with clear message
    }
    return cfg, nil
}
```

**`.env.example` generation:**
`modkit init` generates `.env.example` automatically from the `config.schema.json` of each selected module. This file is committed to git. `.env` is gitignored.

**Rules:**
- Validate all config at startup — never let the app start in a partially-configured state
- Use required field validation — fail with a clear error message, not a nil pointer panic later
- Never commit secrets to git — `.env.example` has placeholder values only
- Config struct must be defined per-app, not shared across modules

---

## §14 Testing Strategy [MVP]

**Test types and when to write them:**

| Type | What it tests | When |
|------|--------------|------|
| Unit test | Handler/service logic with mocked modules | Every handler |
| Integration test | Full stack against real DB | Critical flows |
| Contract test | Module interface compliance | Every module impl |

**Unit tests — mock at the interface boundary:**
```go
// Go — create a mock that satisfies the interface
type mockEmailService struct {
    sendCalled bool
    lastMsg    contracts.EmailMessage
}

func (m *mockEmailService) Send(ctx context.Context, msg contracts.EmailMessage) (*contracts.EmailResult, error) {
    m.sendCalled = true
    m.lastMsg = msg
    return &contracts.EmailResult{ID: "test-id"}, nil
}

// Use in test
func TestSendInvoice(t *testing.T) {
    emailMock := &mockEmailService{}
    handler := NewInvoiceHandler(testDB, emailMock, ...)

    // ... test the handler
    assert.True(t, emailMock.sendCalled)
}
```

**Integration tests — real database:**
```go
// Use testcontainers or a test DB with applied migrations
// Never mock the database in integration tests
func TestInvoiceFlow(t *testing.T) {
    db := testhelpers.NewTestDB(t)  // real Postgres, migrations applied
    // ...
}
```

**Contract tests — verify implementations satisfy the interface:**
```go
// Every module must have a contract test
func RunEmailServiceContractTests(t *testing.T, svc contracts.EmailService) {
    t.Run("Send delivers email", func(t *testing.T) { ... })
    t.Run("GetStatus returns status", func(t *testing.T) { ... })
}

// Run against each implementation
func TestResendEmailServiceContract(t *testing.T) {
    svc := resend.New(testConfig)
    RunEmailServiceContractTests(t, svc)
}
```

**Rules:**
- New features require tests before merging
- Minimum coverage: 70% for new code
- Integration tests must use a real database — do not mock Postgres
- Contract tests must pass for every module implementation

---

## §15 Naming Conventions [MVP]

**Files:**
- Go: `snake_case.go` — `invoice_handler.go`, `invoice_repo.go`, `auth_middleware.go`
- TypeScript: `kebab-case.ts` — `invoice-handler.ts`, `invoice-repo.ts`

**Go:**
- Packages: single-word lowercase — `handlers`, `repos`, `jobs`, `middleware`, `contracts`
- Types: `PascalCase` — `InvoiceHandler`, `CreateInvoiceRequest`
- Functions/methods: `camelCase` for private, `PascalCase` for exported
- Constants: `PascalCase` for exported, `camelCase` for private

**TypeScript:**
- Variables and functions: `camelCase`
- Classes, interfaces, types: `PascalCase`
- Interfaces: `I` prefix — `IEmailService`, `IAuthUser`
- Constants: `SCREAMING_SNAKE_CASE` for env vars and truly constant values

**Database:**
- Tables: `snake_case` plural — `invoices`, `invoice_items`, `webhook_events`
- Columns: `snake_case` — `created_at`, `freelancer_id`, `pdf_url`
- Indexes: `idx_{table}_{column(s)}` — `idx_invoices_freelancer_id`
- Foreign keys: `fk_{table}_{referenced_table}` — `fk_invoice_items_invoices`

**API routes:** `kebab-case` — `/api/v1/invoice-items` not `/api/v1/invoiceItems`

**Environment variables:** `SCREAMING_SNAKE_CASE` — `DATABASE_URL`, `EMAIL_API_KEY`, `STRIPE_SECRET_KEY`

**Job types:** `{entity}:{action}` — `invoice:generate_pdf`, `user:send_welcome_email`

**WebSocket channels (v2):** `{entity}:{id}:{event}` — `invoice:abc123:status_changed`

---

## §16 CI/CD Rules [MVP]

**Required pipelines:**

**On pull request (`.github/workflows/ci.yaml`):**
```yaml
- go build ./... (or bun build)
- go test ./... (or bun test)
- golangci-lint run (or bunx eslint)
- modkit validate
```
All four must pass. No exceptions.

**On merge to `main` (`.github/workflows/deploy-staging.yaml`):**
```yaml
- All CI checks pass
- Build Docker image
- Push to container registry
- Deploy to staging environment
- Run smoke tests
```

**On version tag `v*` (`.github/workflows/deploy-production.yaml`):**
```yaml
- Require manual approval (GitHub environment protection)
- Run full test suite
- Deploy to production
- Monitor error dashboard for 15 minutes
```

**Docker build:**
- Multi-stage builds — separate build and runtime stages
- Runtime image must be minimal (alpine or distroless)
- Never include `.env` or secrets in the Docker image

**Rules:**
- Never deploy directly to production — always staging first
- Never skip CI — if CI is broken, fix it before merging anything
- Production deploys require manual approval gate
- Keep CI fast: target < 3 minutes for the full PR pipeline
- Docker images must be tagged with the git SHA, not just `latest`
