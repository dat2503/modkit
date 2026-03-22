# Auth Module — Agent Instructions

## When to use

Include this module when:
- The app has user accounts (sign-up, sign-in, profile)
- Any route requires authentication
- The app has role-based access control (admin vs regular users)
- The app uses organizations or teams

Do NOT use when:
- The app has no user accounts (e.g. a public read-only tool)
- All access is via API keys only (use a custom middleware instead)

## Dependencies

**Required:** `cache` — auth uses cache for session storage and token blacklisting.
Always initialize cache before auth.

## How to wire

### Go

1. Import the `AuthService` interface from `contracts/go/auth.go`
2. Initialize in bootstrap **after** cache:
   ```go
   authSvc := clerk.New(clerk.Config{
       SecretKey:      cfg.Auth.SecretKey,
       PublishableKey: cfg.Auth.PublishableKey,
   }, cacheSvc)
   ```
3. Register the auth middleware on protected routes:
   ```go
   protected := router.Group("/api/v1", authMiddleware(authSvc))
   ```
4. In handlers, extract the user from context:
   ```go
   user, err := authSvc.ValidateToken(ctx, tokenFromHeader(r))
   ```

### Bun (TypeScript)

1. Import `IAuthService` from `contracts/ts/auth.ts`
2. Initialize in bootstrap after cache:
   ```typescript
   const auth = new ClerkAuthService({
     secretKey: config.auth.secretKey,
     publishableKey: config.auth.publishableKey,
   }, cache)
   ```
3. Register middleware on protected routes:
   ```typescript
   app.use('/api/v1/*', authMiddleware(auth))
   ```
4. In handlers, read the user from context:
   ```typescript
   const user = ctx.get('user') as AuthUser
   ```

## Middleware pattern (Go)

```go
func authMiddleware(auth contracts.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            user, err := auth.ValidateToken(r.Context(), token)
            if err != nil {
                writeError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
            ctx := context.WithValue(r.Context(), userContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Clerk webhook handling

Clerk emits webhook events for user lifecycle (created, updated, deleted).
Always verify the webhook signature before processing:

```go
event, err := authSvc.ConstructWebhookEvent(body, r.Header.Get("Svix-Signature"))
```

Register a webhook endpoint in your Clerk dashboard:
- URL: `POST /api/v1/webhooks/clerk`
- Events: `user.created`, `user.updated`, `user.deleted`

## Frontend (Next.js)

Use Clerk's Next.js SDK for the frontend — it handles sign-in/sign-up UI:
```typescript
// apps/web — use @clerk/nextjs, not this module directly
import { ClerkProvider, SignIn, SignUp } from '@clerk/nextjs'
```

The backend auth module validates tokens issued by Clerk's frontend SDK.

### Protected routes

The generated project includes a protected layout at `src/app/(protected)/layout.tsx`. Any page placed under `src/app/(protected)/` is automatically guarded — unauthenticated users are redirected to `/sign-in`.

```
src/app/
  (protected)/
    layout.tsx          ← auth guard (generated)
    dashboard/page.tsx  ← protected
    settings/page.tsx   ← protected
  page.tsx              ← public landing page
```

**Do NOT** add per-page auth checks in protected routes — the layout handles it. Only add auth checks in pages outside `(protected)/` if they need conditional auth behavior.

## Cache key usage

The auth module uses the cache module for session storage and token management. Here are the exact keys it writes:

| Key pattern | TTL | Purpose |
|---|---|---|
| `sessions:{sessionID}` | 24h | Clerk session data cached to avoid repeated API calls |
| `auth:user:{userID}` | 15m | Cached user profile to reduce Clerk lookups |
| `auth:blacklist:{tokenJTI}` | matches token expiry | Revoked tokens — checked on every `ValidateToken` call |

**If cache is unavailable:**
- `ValidateToken` falls back to direct Clerk API call (slower but functional)
- Session cache misses result in a Clerk API round-trip per request
- Token blacklist misses mean revoked tokens may remain valid until natural expiry

**Namespacing:** All auth keys use the `sessions:` or `auth:` prefix. Do not use these prefixes for application-level cache keys.

## Required env vars

```
AUTH_PROVIDER=clerk
CLERK_SECRET_KEY=sk_test_...         # sensitive
CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_WEBHOOK_SECRET=whsec_...       # sensitive, only if using webhooks
```

## Do NOT

- Store passwords — Clerk handles all credential storage
- Bypass token validation in handlers — always use the middleware
- Read `CLERK_SECRET_KEY` directly in handler code — inject via constructor
- Share the secret key with the frontend — publishable key only
