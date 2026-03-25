# Error Tracking Module — Agent Instructions

## When to use

**Always include.** This module is `always_include: true` — it is non-negotiable.
Initialize it **second**, immediately after observability.

## How to wire

### Go

1. Import `ErrorTrackingService` from `contracts/go/errors.go`
2. Initialize in bootstrap **immediately after observability**:
   ```go
   errSvc, err := sentry.New(sentry.Config{
       DSN:               cfg.ErrorTracking.SentryDSN,
       Environment:       cfg.ErrorTracking.Environment,
       TracesSampleRate:  cfg.ErrorTracking.TracesSampleRate,
   })
   defer errSvc.Flush(ctx)
   ```
3. Register a recovery middleware to capture panics:
   ```go
   router.Use(recoveryMiddleware(errSvc))
   ```
4. In handlers, capture unexpected errors before returning 5xx:
   ```go
   if err != nil {
       errSvc.CaptureError(ctx, err, contracts.CaptureOptions{
           Tags: map[string]string{"operation": "invoice.create"},
       })
       writeError(w, http.StatusInternalServerError, "internal error")
       return
   }
   ```

### Bun (TypeScript)

1. Import `IErrorTrackingService` from `contracts/ts/errors.ts`
2. Initialize second:
   ```typescript
   const errTracking = new SentryErrorTrackingService({
     dsn: config.errorTracking.sentryDsn,
     environment: config.errorTracking.environment,
     tracesSampleRate: config.errorTracking.tracesSampleRate,
   })
   ```
3. Add global error handler:
   ```typescript
   app.onError((err, ctx) => {
     errTracking.captureError(err)
     return ctx.json({ error: 'internal error' }, 500)
   })
   ```

## User context

Set user context after authentication so errors are linked to the right user:

```go
func authMiddleware(auth contracts.AuthService, err contracts.ErrorTrackingService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, authErr := auth.ValidateToken(r.Context(), token)
            if authErr == nil {
                ctx = err.SetUser(r.Context(), contracts.ErrorUser{
                    ID:    user.ID,
                    Email: user.Email,
                })
            }
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## What to capture

- Capture: unexpected errors, panics, 5xx responses
- Do NOT capture: validation errors (400), auth errors (401/403), not found (404) — these are expected

## Required env vars

```
ERROR_TRACKING_PROVIDER=sentry
SENTRY_DSN=https://...@sentry.io/...   # sensitive
SENTRY_ENVIRONMENT=production
SENTRY_TRACES_SAMPLE_RATE=0.1
```

## Integration spec

After wiring, verify with:

1. Set `SENTRY_DSN` to your project's DSN and `SENTRY_ENVIRONMENT=development`
2. Add a temporary test route that panics: `app.get('/api/v1/test-error', () => { throw new Error('test sentry') })`
3. Hit `GET /api/v1/test-error` — the recovery middleware should return 500 and capture the error
4. Open the Sentry dashboard — the `test sentry` error should appear within 30 seconds
5. Remove the test route after verifying

## Do NOT

- Capture expected user errors (bad input, not found) — alerts will fire for noise
- Log sensitive data in error context (passwords, tokens, card numbers)
- Forget to call `Flush()` on graceful shutdown — events may be lost
