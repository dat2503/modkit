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

## Default: Better Auth

Better Auth (https://www.better-auth.com) is an open-source, self-hosted auth library.
It runs inside your Bun backend process (or as a sidecar for Go). No external auth service required.

### How to wire (Bun)

1. Install Better Auth: `bun add better-auth`
2. Create `apps/api/src/auth.ts` — configure the Better Auth instance:
   ```typescript
   import { betterAuth } from 'better-auth'
   import { Pool } from 'pg'

   export const auth = betterAuth({
     database: new Pool({ connectionString: process.env.DATABASE_URL }),
     secret: process.env.BETTER_AUTH_SECRET,
     emailAndPassword: { enabled: true },
     socialProviders: {
       google: {
         clientId: process.env.GOOGLE_CLIENT_ID!,
         clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
       },
     },
   })
   ```
3. Mount the Better Auth handler on your Hono/Express app:
   ```typescript
   app.all('/api/auth/*', (c) => auth.handler(c.req.raw))
   ```
4. Initialize the `BetterAuthService` in bootstrap after cache:
   ```typescript
   import { BetterAuthService } from './modules/auth/better_auth'

   const authSvc = new BetterAuthService({
     baseUrl: config.auth.baseUrl,
     secret: config.auth.secret,
   }, cache)
   ```
5. Register middleware on protected routes:
   ```typescript
   app.use('/api/v1/*', authMiddleware(authSvc))
   ```

### How to wire (Go)

For Go, the Better Auth server runs as part of the Bun frontend/auth service.
The Go backend calls Better Auth's REST API endpoints.

1. Initialize the service in bootstrap after cache:
   ```go
   authSvc, err := betterauth.New(betterauth.Config{
       BaseURL: cfg.Auth.BaseURL,    // e.g. "http://auth-service:3000"
       Secret:  cfg.Auth.Secret,
   }, cacheSvc)
   ```
2. Register the auth middleware on protected routes:
   ```go
   protected := router.Group("/api/v1", authMiddleware(authSvc))
   ```

## Auth middleware pattern

### Go
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

### TypeScript (Hono)
```typescript
async function authMiddleware(auth: IAuthService, c: Context, next: Next) {
  const token = c.req.header('Authorization')?.replace('Bearer ', '') ?? ''
  try {
    const user = await auth.validateToken(token)
    c.set('user', user)
    await next()
  } catch {
    return c.json({ error: 'unauthorized' }, 401)
  }
}
```

## Frontend integration

### Vite + React (Better Auth)
```typescript
// apps/web/src/lib/auth.tsx
import { createAuthClient } from 'better-auth/react'

export const authClient = createAuthClient({
  baseURL: import.meta.env.VITE_API_URL,
})

// In your component:
const { data: session } = authClient.useSession()
```

Use `authClient.signIn.email()`, `authClient.signUp.email()`, `authClient.signOut()` for auth actions.

Protected route guard using `authClient.useSession()`:
```tsx
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { data: session, isPending } = authClient.useSession()
  if (isPending) return null
  if (!session) return <Navigate to="/sign-in" replace />
  return <>{children}</>
}
```

### Next.js (Better Auth)
```typescript
// Use better-auth/next-js for server-side session checks
import { auth } from '@/auth'
import { headers } from 'next/headers'

const session = await auth.api.getSession({ headers: await headers() })
if (!session) redirect('/sign-in')
```

## Alternative: Clerk

If you need managed auth with prebuilt UI components, organizations, and enterprise SSO, use Clerk instead:
```
--modules auth:clerk
```

Clerk wiring (Bun):
```typescript
const auth = new ClerkAuthService({
  secretKey: config.auth.secretKey,
  publishableKey: config.auth.publishableKey,
}, cache)
```

Clerk wiring (Go):
```go
authSvc, err := clerk.New(clerk.Config{
    SecretKey:      cfg.Auth.SecretKey,
    PublishableKey: cfg.Auth.PublishableKey,
}, cacheSvc)
```

## Cache key usage

| Key pattern | TTL | Purpose |
|---|---|---|
| `sessions:{sessionID}` | 24h | Session data cached to avoid repeated API calls |
| `auth:user:{userID}` | 15m | Cached user profile to reduce auth provider lookups |
| `auth:blacklist:{tokenJTI}` | matches token expiry | Revoked tokens — checked on every `ValidateToken` call |

**Namespacing:** All auth keys use the `sessions:` or `auth:` prefix. Do not use these prefixes for application-level cache keys.

## Required env vars

**Better Auth (default):**
```
AUTH_PROVIDER=better-auth
BETTER_AUTH_URL=http://localhost:3000     # URL where Better Auth runs
BETTER_AUTH_SECRET=...                   # generate: openssl rand -hex 32
```

**Clerk (alternative):**
```
AUTH_PROVIDER=clerk
CLERK_SECRET_KEY=sk_test_...             # sensitive
CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_WEBHOOK_SECRET=whsec_...           # sensitive, only if using webhooks
```

## Do NOT

- Store passwords — Better Auth handles all credential storage
- Bypass token validation in handlers — always use the middleware
- Read `BETTER_AUTH_SECRET` directly in handler code — inject via constructor
- Share the secret with the frontend — use the Better Auth React client only
- Call Better Auth admin endpoints from the frontend — server-side only
