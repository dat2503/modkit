# Jobs Module — Agent Instructions

## When to use

Include this module when:
- App sends emails (always pair email with jobs for async sending)
- App generates files (PDF invoices, CSV exports, image processing)
- Any HTTP handler does work that takes >500ms
- Any operation needs retry logic on failure (webhooks, third-party API calls)
- App has scheduled or recurring tasks (daily reports, cleanup jobs)

The rule of thumb: if it doesn't need to be done before the HTTP response returns, use a job.

## Dependencies

**Required:** `cache` — jobs uses Redis (via the cache module) as the queue backend.
No separate Redis instance needed — the same Redis serves both.

## How to wire

### Go (Asynq)

1. Import `JobsService` from `contracts/go/jobs.go`
2. Initialize in bootstrap **after cache**:
   ```go
   jobsSvc, err := asynq.New(asynq.Config{
       RedisURL:    cfg.Cache.RedisURL,  // reuse cache Redis
       Concurrency: cfg.Jobs.Concurrency,
       MaxRetries:  cfg.Jobs.MaxRetries,
   })
   ```
3. Register handlers during bootstrap:
   ```go
   jobsSvc.RegisterHandler("email:send_welcome", handlers.SendWelcomeEmail)
   jobsSvc.RegisterHandler("pdf:generate_invoice", handlers.GenerateInvoicePDF)
   ```
4. Start the worker in a goroutine:
   ```go
   go func() {
       if err := jobsSvc.Start(ctx); err != nil {
           log.Error("jobs worker stopped", "error", err)
       }
   }()
   ```
5. Inject into HTTP handlers that enqueue jobs

### Bun (BullMQ)

1. Import `IJobsService` from `contracts/ts/jobs.ts`
2. Initialize after cache:
   ```typescript
   const jobs = new BullMQJobsService({
     redisUrl: config.cache.redisUrl,
     concurrency: config.jobs.concurrency,
   })
   ```
3. Register handlers and start worker similarly

## Job handler pattern (Go)

**Critical: all handlers must be idempotent** — they may be called multiple times for the same job.

```go
func (h *EmailHandlers) SendWelcomeEmail(ctx context.Context, payload []byte) error {
    var p WelcomeEmailPayload
    if err := json.Unmarshal(payload, &p); err != nil {
        return err  // permanent failure — won't retry
    }

    // Check if already sent (idempotency)
    sent, _ := h.cache.Exists(ctx, "email:welcome:sent:"+p.UserID)
    if sent {
        return nil  // already done — success
    }

    user, err := h.userRepo.Get(ctx, p.UserID)
    if err != nil {
        return fmt.Errorf("get user: %w", err)  // transient — will retry
    }

    _, err = h.email.Send(ctx, contracts.EmailMessage{...})
    if err != nil {
        return fmt.Errorf("send email: %w", err)  // transient — will retry
    }

    h.cache.Set(ctx, "email:welcome:sent:"+p.UserID, []byte("1"), 24*time.Hour)
    return nil
}
```

## Enqueueing from an HTTP handler

```go
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    user, err := h.userRepo.Create(ctx, req)
    if err != nil { ... }

    // Don't block on email — enqueue async
    _, _ = h.jobs.Enqueue(ctx, "email:send_welcome", WelcomeEmailPayload{UserID: user.ID})

    writeJSON(w, http.StatusCreated, user)
}
```

## Scheduled jobs

```go
// Run a cleanup job every night at 2am
jobsSvc.EnqueueAt(ctx, "cleanup:expired_invoices", nil, nextRunAt(2, 0))
```

## Required env vars

```
JOBS_PROVIDER=asynq          # or bullmq for Bun
JOBS_CONCURRENCY=10
JOBS_DEFAULT_QUEUE=default
JOBS_MAX_RETRIES=3
# REDIS_URL comes from the cache module — reused automatically
```

## Integration spec

After wiring, verify with:

1. Ensure Redis is running (`make infra-up`)
2. Register a test handler during bootstrap:
   ```go
   jobsSvc.RegisterHandler("test:ping", func(ctx context.Context, payload []byte) error {
       log.Info("job executed", "payload", string(payload))
       return nil
   })
   ```
3. Add a temporary test route that enqueues a job:
   ```go
   handle, err := jobs.Enqueue(ctx, "test:ping", map[string]string{"msg": "hello"})
   // handle.ID should be non-empty
   ```
4. Hit the test route — the job handler should log `job executed` within a few seconds
5. Verify in Redis: `redis-cli KEYS asynq:*` should show queue entries
6. Remove the test handler and route after verifying

## Do NOT

- Write handlers that are NOT idempotent — jobs will be retried on failure
- Enqueue jobs inside database transactions — the job may enqueue before the transaction commits
- Pass large payloads in job messages — store data in the DB, pass only IDs
- Return errors for permanent failures that shouldn't be retried (wrap with a sentinel error type)
