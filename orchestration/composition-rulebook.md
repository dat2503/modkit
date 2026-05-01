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

---

## §14 Agent Verification Loop (addition to Testing Strategy) [MVP]

**Per-handler (agent-only, no human):**
```bash
go build ./...              # must pass before writing the next handler
modkit validate --output json  # fix wiring violations immediately
```

**Per-milestone (full self-review):**
Run the full 5-step Agent Self-Review Protocol (see Playbook). All steps must pass before declaring a milestone complete. Self-improvement loop: up to 3 same-error / 6 different-error iterations, then escalate.

**Self-audit checklist (before any human checkpoint):**
- [ ] All routes in `openapi.yaml` have a corresponding handler
- [ ] All handlers use `writeError()` / `c.json({error:...})` for error responses
- [ ] Bootstrap init order matches `module_wiring_order` from architecture plan
- [ ] No hardcoded env vars — all config via injected struct
- [ ] Tests written for all new handlers (unit) and critical flows (integration)
- [ ] `modkit validate` shows 0 errors
- [ ] Every pattern listed in `patterns_applied` is implemented
- [ ] No N+1 queries in list handlers (§20.20)
- [ ] Every external call has a timeout (§20.18)
- [ ] Every inbound webhook verifies signature (§20.17)
- [ ] Every retryable POST supports idempotency keys (§20.1)
- [ ] Migrations are backward-compatible within a release (§20.10)
- [ ] If `ui-spec.yaml` exists: frontend components match design spec (colors, layout, names)
- [ ] No §22.6 "don't reinvent" smells in custom code

**Rule:** presenting a milestone to the human with known failures is a protocol violation. Fix first, then present.

---

## §17 Guardrails & Loop Control [MVP]

See Playbook §Agent Guardrails & Loop Control for the full rules.

**Hard limits summary:**
- Max 3 same-error retries per milestone → escalate
- Max 6 different-error retries per milestone → escalate
- Max 50 file writes per Phase 3 sub-phase → stop, present progress
- Max 5 consecutive tool failures → stop
- Phase 3 wall time 90 min → stop, summarize
- Token budgets: warn at 80%, hard stop at 100% (see Playbook for per-phase budgets)

**Abort conditions (no retry):** build exit outside [0,1], partial migration failure, DROP/DELETE not in approved plan, `modkit validate` parse error.

**Escalation:** stop → emit ≤200-token failure report → wait for human → never self-resume.

**Agents must respect hard limits even if the user says "keep trying."**

---

## §18 Token Efficiency [MVP]

Minimize token usage at every step. Every rule below saves tokens — follow them.

**Loading rules:**
- Load `playbook.md`, `composition-rulebook.md`, `registry.yaml` ONCE per session — never re-read unless the file has been written to
- Load module `AGENT.md` only for selected modules (after Phase 1 approval), not the full catalog
- Load `ui-spec.yaml` once at Phase 3d start — do not re-read original design files

**Output rules:**
- Default to YAML/JSON for all structured artifacts — 5–10× shorter than prose
- Diff-style edits, not full file rewrites for small changes
- Status reports use fixed schemas (see §Self-Review Telemetry in Playbook) — no recap paragraphs
- Architecture amendments are YAML diffs, not full re-plans

**Verification ordering (always cheap → expensive):**
- Exit codes > HTTP status codes > parsed JSON > log scanning > screenshots > visual diffs
- Stop at first failure — do not gather full state when one signal is enough

**Reference by ID, never paraphrase:**
- Write `see §20.1` not "the idempotency key pattern where the client supplies a header and the server caches the response for 24 hours..."
- Write `see Playbook §Self-Review Protocol` not a re-explanation of the 5 steps

**Rule:** when two approaches are equivalent, pick the one that emits fewer tokens.

---

## §19 Engineering Principles [MVP]

These principles are mandatory defaults. Deliberate deviation requires an explicit note in `deviations:` (see §22.4).

### §19.1 SOLID
- **Single Responsibility**: one handler does one thing; services own one aggregate
- **Open/Closed**: extend via interface implementations (enforced by §1 Interface-First)
- **Liskov Substitution**: any module impl must pass the contract test (§14)
- **Interface Segregation**: small contracts; do not bundle unrelated methods
- **Dependency Inversion**: depend on contracts, never on concrete implementations (§1)

### §19.2 DRY (Don't Repeat Yourself)
- Repeated logic → extract to a service or helper after the 3rd occurrence
- Repeated wiring → put in bootstrap, not in handlers

### §19.3 YAGNI (You Aren't Gonna Need It)
- Do not add fields, routes, or modules not in the approved architecture plan
- Do not add abstractions until a 2nd concrete use exists
- Speculative configurability is a smell

### §19.4 Separation of Concerns
- Layer boundary: **handler → service → repository → DB**
- Handlers never call DB directly; repositories never know about HTTP
- Modules never reach across into each other's internals — only via contracts

### §19.5 Twelve-Factor App
- Config via env vars, never hardcoded (enforced by §13)
- Stateless processes (state in DB, cache, or object store)
- Logs to stdout — never to files (infra collects them)
- Disposability: graceful shutdown on SIGTERM
- Dev/prod parity: same Postgres/Redis versions in compose and production

### §19.6 Coding Principles
- **Fail fast**: validate at system boundaries; reject malformed input with 400
- **Errors as values**: Go returns `error`; TypeScript returns typed Result — never panic in handlers
- **Pure where possible**: services accept inputs, return outputs; side effects at the edges
- **Composition over inheritance**: both Go and TypeScript favor composition
- **Small functions**: if a function exceeds ~50 lines or 3 levels of nesting, split it

**Application rule:** the self-audit checklist checks these at every milestone.

---

## §20 Data-Intensive Application Patterns [MVP]

Reference: Kleppmann, *Designing Data-Intensive Applications*. Each pattern has a **trigger** — apply it only when the trigger is met. Cite the §20.X ID in the architecture plan's `patterns_applied` block. Select patterns using the §21 tier filter.

### Reliability

#### §20.1 Idempotency Keys (DDIA Ch.8, Ch.11)
- **Trigger**: any POST that mutates state and could be retried (payments, webhooks, job enqueue)
- **Pattern**: client supplies `Idempotency-Key` header → server stores (key, response) for 24h → repeat request returns cached response
- **Implementation**: `idempotency_keys` table with PK on (key, route); unique constraint

#### §20.2 Outbox Pattern (DDIA Ch.11)
- **Trigger**: must reliably notify external systems after a DB write (webhook, email, downstream)
- **Pattern**: write event row to `outbox` table in the same DB transaction → background worker reads outbox → emits event → marks delivered
- **Why**: avoids dual-write (DB committed but notification lost)

#### §20.3 At-Least-Once Job Delivery (DDIA Ch.11)
- **Trigger**: any background job that mutates state
- **Pattern**: jobs must be idempotent; rely on queue's at-least-once guarantee; never assume exactly-once
- **Implementation**: job checks "already done?" before acting (often via §20.1 key)

#### §20.4 Retry with Exponential Backoff + Jitter (DDIA Ch.8)
- **Trigger**: any external API call (Stripe, Resend, S3, third-party)
- **Pattern**: retry on 5xx/timeout with backoff (1s, 2s, 4s, 8s) + jitter; cap at 5 attempts
- **Never**: retry on 4xx (except 429 with `Retry-After`)

#### §20.5 Circuit Breaker (DDIA Ch.8) — Tier 2
- **Trigger**: external dependency that can fail and impact latency (defer until ≥2 outages in 30 days)
- **Pattern**: track failure rate; open circuit after threshold; half-open after cooldown

### Scalability

#### §20.6 Cache-Aside (DDIA Ch.5) — see §7
- Full pattern documented in §7

#### §20.7 Read Replicas Awareness (DDIA Ch.5) — Tier 2
- **Trigger**: read-heavy workload >10:1 read/write (defer until DB CPU >70% sustained)
- **Pattern**: use `db.Reader()` for SELECT, `db.Writer()` for writes; understand replica lag
- **MVP**: single DB but code as if reader/writer split — switch is then a config change

#### §20.8 Cursor-Based Pagination (DDIA Ch.6)
- **Trigger**: list endpoint expected to return >1000 rows
- **Pattern**: `(created_at, id)` cursor pagination — never `OFFSET` (degrades at scale)
- **API shape**: `?cursor=<base64>&limit=20` → `{data, next_cursor}`

#### §20.9 Backpressure on Job Queue (DDIA Ch.11) — Tier 2
- **Trigger**: job producer can outpace consumer (defer until queue depth >5000 sustained)
- **Pattern**: bounded queue size; return 429 on full; consumer pulls at controlled rate

### Maintainability

#### §20.10 Schema Evolution: Backward-Compatible Migrations (DDIA Ch.4)
- **Trigger**: any migration on a live table
- **Pattern**: additive-only within a release (new nullable column, new table); destructive changes split across two releases
- **Never**: rename a column in a single migration on a live system

#### §20.11 Versioned APIs
- **Trigger**: public API with external consumers
- **Pattern**: prefix routes with `/api/v1/`; never break v1 within its lifetime; add v2 for breaking changes; sunset v1 with deprecation header

#### §20.12 Structured Logs with Trace IDs (DDIA Ch.10) — extends §11
- All logs include: `trace_id`, `span_id`, `user_id` (if authenticated), `request_id`

### Data Integrity

#### §20.13 Transactions with Appropriate Isolation (DDIA Ch.7)
- **Trigger**: any multi-row write that must be atomic, or any write where lost updates cause harm
- **Default**: `READ COMMITTED` (Postgres default)
- **Use `REPEATABLE READ`**: read-heavy reports
- **Use `SERIALIZABLE` or `SELECT FOR UPDATE`**: payment status, inventory decrements, balance updates — any race causes financial harm

#### §20.14 Optimistic Concurrency / Version Columns (DDIA Ch.7) — Tier 2
- **Trigger**: collaborative edits on the same record (defer until multiple editors exist)
- **Pattern**: `version` column; `UPDATE WHERE id = X AND version = Y`; 0-rows-affected → 409 Conflict

#### §20.15 Event Ordering (DDIA Ch.5, Ch.11) — Tier 2
- **Trigger**: webhook events that arrive out of order (defer until provider sends out-of-order)
- **Pattern**: store event with monotonic timestamp; reconcile from sorted log, not arrival order

#### §20.16 Soft Deletes (DDIA Ch.5) — see §9
- Full pattern documented in §9

### Distributed Systems Hygiene

#### §20.17 Webhook Signature Verification (DDIA Ch.8)
- **Trigger**: every inbound webhook (always — Tier 0)
- **Pattern**: verify HMAC signature with provider secret BEFORE any processing; reject 401 on mismatch

#### §20.18 Timeouts on Every External Call (DDIA Ch.8)
- **Trigger**: every HTTP call, DB query, Redis op (always — Tier 0)
- **Pattern**: explicit timeout (5s HTTP, 1s DB single query, 100ms cache); never `context.Background()` in handlers

#### §20.19 Rate Limiting (DDIA Ch.8)
- **Tier 0**: auth, payment, signup routes
- **Tier 2**: all other public routes (defer until abuse is observed)
- **Pattern**: token bucket per IP/user via Redis; return 429 with `Retry-After`

### Performance

#### §20.20 N+1 Query Prevention
- **Trigger**: any list endpoint that returns related entities (always — Tier 0)
- **Pattern**: batch-load via JOIN or `WHERE id IN (...)`; never loop rows issuing one query each
- **Self-audit**: inspect every list handler for N+1 patterns

#### §20.21 Index by Query Pattern (DDIA Ch.3)
- **Trigger**: every WHERE clause and ORDER BY in a hot path
- **Pattern**: composite indexes matching query columns + order; run `EXPLAIN` before merging
- **Note**: only create indexes on confirmed hot paths — premature indexing slows writes

**Selection rule**: in Phase 2 Pattern Selection, emit `patterns_applied` citing only patterns whose triggers match this project. Use the §21 tier filter. Patterns not listed are NOT implemented.

---

## §21 Lean MVP & Anti-Over-Engineering [MVP]

The default posture is **as simple as possible until proven otherwise.** §19 and §20 are a menu, not a requirement.

### §21.1 Pattern Tiers

| Tier | Meaning | When to apply |
|------|---------|--------------|
| **Tier 0 — Safety** | Non-negotiable; skipping risks data loss, double-charges, or security holes | Always, every project |
| **Tier 1 — Direct trigger** | Apply only when the project brief contains a concrete trigger | Brief explicitly creates the trigger condition |
| **Tier 2 — Deferred** | Defer until a measured post-launch signal fires | Never at MVP — promote via architecture amendment |

**Tier-0 patterns (always apply):**
§20.1 (payment/webhook routes only), §20.10, §20.13 (financial state), §20.17, §20.18, §20.20

**Tier-1 patterns (apply on direct trigger):**
§20.2, §20.3, §20.4, §20.8, §20.11, §20.21

**Tier-2 patterns (defer until measured):**
§20.5, §20.7, §20.9, §20.14, §20.15, §20.19 (most routes)

### §21.2 The Reverse-YAGNI Test (mandatory at Phase 2)

For every Tier-1 candidate in `patterns_applied`, answer:

> **"If we removed this pattern today, what specifically breaks today (not in a hypothetical future)?"**

If the honest answer is "nothing today, but maybe later" → remove it; move to `patterns_deferred` with the trigger that would later promote it.

### §21.3 Lean MVP Default Profile

Unless the brief explicitly says otherwise:

| Concern | MVP default | Upgrade trigger |
|---------|-------------|----------------|
| Architecture | Monolith (one API, one DB, one frontend) | Revisit at >50k req/min sustained |
| Database | Single Postgres, single region | CPU >70% sustained or read latency >200ms p95 |
| Cache | Redis for auth + dashboard reads only | Add keys when measured cache-miss latency hurts |
| Jobs | Async only when sync operation >500ms | Defer queues until actually needed |
| Search | Postgres `ILIKE` + trigram | Promote to full-text when quality complaints arise |
| Auth | One IdP (default impl) | Defer SAML/multi-IdP until enterprise customer asks |
| File storage | S3 equivalent, no CDN | Add CDN when bandwidth bill or latency justifies |
| Real-time | Polling at 5–10s | Promote to WebSocket when brief explicitly needs <1s |

### §21.4 Pattern Budget for MVP

Soft cap: **≤8 patterns from §20** (excluding Tier-0) per MVP. Exceeding the cap requires an explicit justification per pattern in `patterns_applied`.

### §21.5 Forbidden at MVP (over-engineering smells)

The agent MUST NOT introduce these without explicit user request and concrete reasoning:

- Microservices, message buses (Kafka/NATS/Pulsar)
- Event sourcing, CQRS
- Service mesh, sidecar proxies
- Multi-region active-active
- gRPC between internal services
- More than 2 cache layers
- More than 1 database engine
- Premature sharding or partitioning

If the agent is tempted to introduce one → STOP and ask the user.

### §21.6 Evolution Protocol (adding patterns post-launch)

1. **Observe**: a measured signal crosses the threshold in `patterns_deferred`
2. **Document**: write `architecture_amendment.observation` (one line)
3. **Propose**: the amendment citing the §20 pattern to promote
4. **Approve**: human reviews — a real trigger justifies real work
5. **Implement**: move from `patterns_deferred` to `patterns_applied`

**Rule**: patterns are pulled by evidence, never pushed by speculation.

### §21.7 The "Boring Stack" Principle

Prefer boring, proven choices over novel ones at MVP. Postgres over DynamoDB. Monolith over microservices. cron over Airflow. JSON over HTTPS over gRPC streams. The agent's job at MVP is to make the product work — not to build interesting infrastructure.

---

## §22 Reuse Over Reinvent + Project Fit [MVP]

### §22.1 Reuse Hierarchy (mandatory check before writing custom code)

Check in this order. Stop at the first fit:

1. **Existing modkit module** — does an installed module already provide this capability?
2. **MCP tools available in the session** — Playwright for tests, Canva/Excalidraw for design, GitHub for PRs
3. **Runtime standard library** — Go stdlib (`net/http`, `crypto/hmac`, `database/sql`); Bun built-ins
4. **Well-maintained third-party library** — see §22.2 rubric
5. **Custom implementation** — only if 1–4 genuinely don't fit

**Before writing custom code, emit a `reuse_check` block:**
```yaml
reuse_check:
  capability: "idempotency keys for /pay/:token"
  checked:
    - source: "modkit cache module"
      fit: "partial — Redis client only, no keyed-response wrapper"
    - source: "Go stdlib"
      fit: "none"
    - source: "library: github.com/example/go-idempotency"
      fit: "yes — MIT, maintained, semver stable"
  decision: "use go-idempotency library"
  custom_code_required: "thin middleware wrapper, ~30 lines"
```

Absence of this block for custom code is a protocol violation.

### §22.2 Library Selection Rubric

A library is acceptable when ALL apply:
- **Maintained**: commit in last 6 months, OR explicitly marked stable by maintainer
- **License**: MIT, Apache-2, BSD-3, ISC — flag GPL/AGPL for human review
- **Stable API**: semver ≥1.0.0; v0.x.x requires explicit human approval
- **Bounded transitive deps**: <30 for backend, <50 for frontend
- **Bundle budget** (frontend): per-entry chunk ≤200KB gzipped (or ui-spec budget if specified)

**Red flags**: sole maintainer, no tests, last commit >2 years, single-vendor lock-in.

### §22.3 "Best Practice" is Contextual — Project Fit Rule

Apply patterns and libraries that fit THIS project's reality:

| Project signal | Implication |
|---------------|-------------|
| <100 users, prototype, internal tool | Simpler beats "correct" — skip patterns sized for scale you don't have |
| Financial/regulated/multi-tenant SaaS | Apply more rigor; §20 Tier-0 defaults hold |
| Solo dev, weekend project | Familiar tools over "best" ones |
| Large team, long-lived product | Consistency and explicit contracts matter more |
| Mobile-first or low-bandwidth | Bundle size and payload shape outweigh elegance |

The agent must capture `project_signals` in Phase 2 before applying §21 tiers.

### §22.4 Deviation Documentation

When deliberately departing from §19/§20 defaults due to project fit:

```yaml
deviations:
  - rule: §20.2
    decision: "skip outbox pattern"
    reason: "single-user admin tool; email failure rate 1/10000 acceptable; retry from UI is fine"
  - rule: §20.8
    decision: "use OFFSET pagination"
    reason: "max 50 invoices per freelancer; OFFSET performance is irrelevant at this scale"
```

Future agents and humans can tell "deliberate" from "bug." That distinction matters.

### §22.5 Tool Discovery Step (output of Phase 1, input to Phase 2)

After module selection, emit `toolchain.yaml` — see Playbook Phase 1 for schema. This document drives implementation reuse decisions in Phase 3. Never write a feature without first checking if a listed tool already provides it.

### §22.6 The "Don't Reinvent" Smell List

If the agent finds itself implementing one of these from scratch → STOP and find a library:

- HTTP retry / backoff / circuit breaker
- HMAC signature verification
- JWT parsing and validation
- UUID generation
- Date/time arithmetic (`time` / `date-fns`)
- Rate limiting (`golang.org/x/time/rate` or Redis-based)
- CSV / Excel parsing
- File MIME type detection
- HTML sanitization
- Markdown rendering
- Image resizing
- Email template rendering
- Crypto primitives — ALWAYS use stdlib or an audited library; never roll your own

**Self-audit**: scan own code at milestone end for these smells. If found, replace with a library before declaring the milestone complete.

---

## §23 Day-2 Operations [MVP]

These rules govern how an agent manages a deployed application post-launch. They complement Phase 6 in the playbook — the playbook describes *when*, this section describes *how and to what standard*.

### §23.1 SLI/SLO Definitions

Every deployed application must have `slo.yaml` committed to the repo (generated in Phase 6a). Default targets — adjust by project signals (§22.3):

| Signal | SLO target |
|--------|-----------|
| latency p95 | < 500ms (< 200ms for financial/B2B) |
| error rate | < 1% (< 0.1% for payment flows) |
| availability | > 99.5% (> 99.9% for paid B2B SaaS) |

SLOs must be tied to user-facing behavior, not internal metrics. "Database query time" is not an SLO; "checkout endpoint latency p95" is.

### §23.2 Alert Design Rules

- Page **only** on user-impacting breaches sustained >5 minutes
- Warn on budget burn (>50% of error budget consumed in 1h)
- Do NOT alert on every individual error — Sentry handles error visibility; alerts fire on *rates*, not counts
- Every alert must have a runbook reference in its description field
- Alerts without runbooks are protocol violations

### §23.3 Runbook Template

One runbook per known failure mode. Stored in `runbooks/{module}/{symptom}.yaml`:

```yaml
symptom: "API returns 502 on /health"
probable_causes:
  - "Container failed to start (missing env var)"
  - "Database connection refused"
  - "Redis connection refused"
diagnosis:
  - "docker logs <container_name>"
  - "curl -sf http://localhost:8080/health"
  - "docker compose ps"
mitigation:
  - "Check .env for missing DATABASE_URL or REDIS_URL"
  - "Restart: docker compose restart api"
  - "If DB is down: docker compose up -d postgres"
escalate_if: "Health check still failing after 10 minutes"
owner: "on-call"
```

### §23.4 Postmortem Template

After any user-impacting incident (sustained SLO breach, data loss, security event):

```yaml
incident_id: "INC-001"
date: "2026-05-01"
severity: P1 | P2 | P3
duration_minutes: 0
summary: ""
timeline:
  - time: "14:23"
    event: "alert fired"
  - time: "14:31"
    event: "root cause identified"
  - time: "14:45"
    event: "mitigation deployed"
root_cause: ""
contributing_factors: []
what_went_well: []
action_items:
  - description: ""
    owner: ""
    due_date: ""
# Link to §21.6 if a pattern promotion is warranted:
patterns_to_promote: []
```

Postmortems are blameless. Focus on systemic causes, not individual mistakes. Every incident that exceeds P2 severity must produce a postmortem within 48 hours.

### §23.5 Lean Oncall

- Solo dev: pager → mobile push notification via Sentry alerts
- Small team: rotation schedule outside modkit scope; use PagerDuty or equivalent
- Never expect the on-call person to investigate without a runbook — if a runbook doesn't exist, write it as part of the incident

---

## §24 Release Management [MVP]

### §24.1 Conventional Commits (mandatory)

All commits follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add invoice PDF download
fix: correct Stripe webhook retry logic
feat!: remove legacy /api/v0 routes   ← breaking change
docs: update README deploy instructions
chore: upgrade go to 1.23
```

This is the input to automated changelog generation. Non-conventional commits block the `/release` skill.

### §24.2 Semantic Versioning Rules

| Commit type | Version bump |
|-------------|-------------|
| `fix:` only | PATCH (1.0.X) |
| `feat:` present | MINOR (1.X.0) |
| `feat!:` or `BREAKING CHANGE:` footer | MAJOR (X.0.0) |

Tools to use (§22 — don't reinvent):
- Go projects: `git-cliff` for CHANGELOG generation
- Bun/TS projects: `changesets` for monorepo versioning + CHANGELOG

Never hand-write CHANGELOG entries — they must be generated from Conventional Commits.

### §24.3 Rollback Protocol

When a production deploy must be reverted:

```bash
# 1. Identify the last known-good tag
git log --tags --oneline

# 2. Redeploy the known-good image/commit
# github-actions: re-run the deploy-production workflow for the previous tag
# vercel: vercel rollback <deployment-url>
# railway: railway rollback

# 3. Verify health endpoint
curl -sf https://api.yourapp.com/health

# 4. Open a postmortem (§23.4)
```

Never patch production directly. Rollback → fix → redeploy is the only safe path.

### §24.4 Canary / Feature-Flag Rollout

When a feature needs gradual rollout (not all-or-nothing):

1. Gate the feature behind a `feature-flags` module flag (see module AGENT.md)
2. Deploy with flag at 0% for all users
3. Promote: 5% → 20% → 50% → 100% with a verification step at each increment
4. Verification: check SLO metrics (§23.1) hold for each cohort before expanding
5. Rollback: set flag to 0% — no redeploy needed

This replaces blue/green deployment for most MVP rollout needs at lower infrastructure cost.

### §24.5 Release Checklist

Before tagging a release:
- [ ] All Phase 4 critical flows passing on staging
- [ ] No open P1/P2 incidents
- [ ] CHANGELOG generated and reviewed
- [ ] Semver bump matches commit types (§24.2)
- [ ] `slo.yaml` targets still achievable with new code (check staging metrics)
- [ ] Secrets not about to expire (§25.3)

---

## §25 Security Lifecycle [MVP]

### §25.1 CI Security Scans (mandatory additions to generated workflows)

Add these jobs alongside build + test in CI. They must pass before merge to main.

**Go projects:**
```yaml
- name: Dependency vulnerabilities
  run: govulncheck ./...

- name: Static analysis (security)
  run: |
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec -fmt=json -out=sec-report.json ./...

- name: Secret scanning
  uses: gitleaks/gitleaks-action@v2
```

**Bun/TS projects:**
```yaml
- name: Dependency vulnerabilities
  run: npm audit --audit-level=high

- name: Static analysis (security)
  run: npx semgrep --config=auto --error

- name: Secret scanning
  uses: gitleaks/gitleaks-action@v2
```

On any HIGH/CRITICAL finding → fail the build. MEDIUM findings → fail unless explicitly suppressed with a comment explaining why. LOW → log only.

### §25.2 Dependency Scanning Cadence

- On every PR (CI job above)
- Weekly cron in Phase 6c (§Phase 6c — Maintain)
- After any security disclosure affecting the runtime or a direct dependency

### §25.3 Secret Rotation Cadence

| Secret type | Rotation frequency | Trigger for immediate rotation |
|-------------|-------------------|-------------------------------|
| Clerk API keys | Quarterly | Team member departure; suspected exposure |
| Stripe API keys | Quarterly | Same |
| Stripe webhook secrets | Yearly | Suspected exposure |
| Database credentials | Per-incident | Any team member departure |
| Redis credentials | Per-incident | Any team member departure |
| GitHub deploy keys | Yearly | Repository transfer or team change |
| JWT signing keys | Yearly | Suspected exposure |

Rotation procedure:
1. Generate new secret in provider dashboard
2. Update in CI/CD secrets store (GitHub Actions Secrets)
3. Deploy — verify health endpoint immediately after
4. Revoke old secret only after verified deployment
5. Log rotation date + reason in `security-log.yaml`

### §25.4 Threat Model Template (one per public surface, generated at Phase 2)

Keep this short. One YAML block per API surface, updated when routes change:

```yaml
# threat-model.yaml — append one entry per public surface
surfaces:
  - route: "POST /api/v1/pay/:token"
    threats:
      - id: S1
        category: STRIDE-Spoofing
        description: "Attacker replays payment request with guessed token"
        mitigation: "§20.1 idempotency keys + token is random UUID (128-bit entropy)"
      - id: T1
        category: STRIDE-Tampering
        description: "Attacker modifies Stripe webhook payload"
        mitigation: "§20.17 HMAC verification on all webhooks"
    residual_risk: low
    last_reviewed: "2026-05-01"
```

Review threat model at every architecture amendment. A route added without a threat model entry is a protocol violation at the §14 self-audit.

### §25.5 Quarterly Security Review Checklist

- [ ] All §25.1 CI scans passing on main
- [ ] Secret rotation current (§25.3)
- [ ] Threat model updated for any new routes since last review
- [ ] Dependency CVE scan clean or findings triaged
- [ ] Auth module: verify Clerk webhook secret still valid and not expired
- [ ] Payments module: verify Stripe restricted-key permissions are minimal (read-only where possible)
- [ ] Logs: confirm no PII is appearing in log output (spot check Sentry events)
- [ ] Access: confirm only current team members have admin access to Clerk, Stripe, Sentry, CI/CD

---

## §26 Data Governance [MVP]

### §26.1 Data Retention Table

Every entity in the database must have a retention policy. Define this in `data-governance.yaml` at Phase 2 and commit it with the migrations:

```yaml
# data-governance.yaml
retention:
  users:
    hot_days: -1          # keep indefinitely while account active
    soft_delete_on: account_closure
    hard_delete_after_days: 730  # 2 years after closure (GDPR)
  invoices:
    hot_days: -1          # financial records — keep indefinitely
    archive_after_days: 1825  # 5 years (tax/legal requirement)
    hard_delete: never
  sessions:
    hot_days: 30
    hard_delete_after_days: 30
  audit_logs:
    hot_days: -1
    archive_after_days: 365
    hard_delete: never
```

Retention policies must be reviewed at the §25.5 quarterly security review.

### §26.2 GDPR / Privacy Endpoints

Required if `project_signals.regulatory` includes EU users or any data privacy regulation:

| Endpoint | Method | Behavior |
|----------|--------|---------|
| `/api/me/export` | GET | Return all user data as JSON — must cover every table with `user_id` |
| `/api/me/delete` | DELETE | Soft-delete account → schedule hard-delete per §26.1 retention table |

These must be in the architecture plan if the project signal triggers them. They are NOT optional when EU users are in scope. Test them in Phase 4 critical flows.

### §26.3 PII Inventory

At Phase 2, the agent emits a PII inventory block alongside the schema:

```yaml
# pii-inventory — append to architecture-plan.yaml
pii:
  - entity: users
    fields:
      - name: email
        category: contact
        used_for: [auth, transactional_email]
      - name: full_name
        category: identity
        used_for: [invoice_display]
  - entity: invoices
    fields:
      - name: client_email
        category: contact
        used_for: [payment_notification]
```

Rules:
- Never log PII fields (already enforced by §11, reinforced here)
- Never include PII in error messages
- PII fields must not appear in URL parameters (use IDs or opaque tokens)

### §26.4 Backup and Restore

**Backup configuration** (generated by cicd module for production deployments):
- Postgres: automated daily snapshot (provided by Railway/managed Postgres/RDS — not custom)
- Verify: `pg_dump` produces a non-empty file
- Retention: keep 30 daily snapshots; 12 monthly snapshots

**Restore drill (monthly, Phase 6c):**
```bash
# 1. Download latest snapshot
# 2. Restore to staging database
pg_restore -h staging-db -U postgres -d myapp_staging latest.dump
# 3. Verify row counts approximate production
# 4. Run smoke tests against restored staging
# 5. Log: date, backup age at restore, row count match, smoke test result
```

Failing a restore drill means the backup is untrustworthy. Treat it as a P1 incident.
