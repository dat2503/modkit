# Observability Module — Agent Instructions

## When to use

**Always include.** This module is `always_include: true` — it is non-negotiable.
Initialize it first, before every other module.

## How to wire

### Go

1. Import `ObservabilityService` from `contracts/go/observability.go`
2. Initialize **first** in bootstrap:
   ```go
   obsSvc, err := otel.New(otel.Config{
       ServiceName: cfg.Obs.ServiceName,
       Endpoint:    cfg.Obs.OTLPEndpoint,
       Headers:     cfg.Obs.OTLPHeaders,
       LogLevel:    cfg.Obs.LogLevel,
       LogFormat:   cfg.Obs.LogFormat,
   })
   defer obsSvc.Shutdown(ctx)
   ```
3. Wrap your HTTP router with the tracing middleware:
   ```go
   router.Use(tracingMiddleware(obsSvc))
   ```
4. Inject into all other modules that accept it as an optional dependency

### Bun (TypeScript)

1. Import `IObservabilityService` from `contracts/ts/observability.ts`
2. Initialize first in bootstrap:
   ```typescript
   const obs = new OtelObservabilityService({
     serviceName: config.obs.serviceName,
     endpoint: config.obs.otlpEndpoint,
     headers: config.obs.otlpHeaders,
   })
   ```
3. Register middleware:
   ```typescript
   app.use('*', tracingMiddleware(obs))
   ```

## Current implementation status

**Structured logging** — fully implemented via `slog`. All `Log()` calls produce real structured log output.

**Distributed tracing** — the `StartSpan()` and `span.End()` methods exist but return **no-op spans**. Trace data is not exported to any backend. This means:
- Your code can call `StartSpan()` everywhere — it won't break, but won't produce traces yet
- Once a real OTel exporter is configured (Jaeger, Honeycomb, etc.), tracing will activate without code changes
- The middleware records `http.method`, `http.url`, `http.status_code` attributes — these will populate once the exporter is wired

Do NOT skip writing span code because tracing is no-op — the spans will become real when the exporter is connected.

## Span patterns

### Handler span (Go)

```go
func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx, span := h.obs.StartSpan(r.Context(), "invoice.create")
    defer span.End()

    // All downstream calls use ctx — spans are propagated automatically
    invoice, err := h.repo.Create(ctx, req)
    if err != nil {
        span.RecordError(err)
        writeError(w, http.StatusInternalServerError, "failed to create invoice")
        return
    }
    span.SetAttribute("invoice.id", invoice.ID)
    writeJSON(w, http.StatusCreated, invoice)
}
```

### Structured logging

```go
obs.Log(ctx, contracts.LogLevelInfo, "invoice created", map[string]any{
    "invoice_id":    invoice.ID,
    "freelancer_id": invoice.FreelancerID,
    "total":         invoice.Total,
})
```

## Required env vars

```
OTEL_SERVICE_NAME=myapp-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317   # empty to disable export
OTEL_EXPORTER_OTLP_HEADERS=                         # for cloud providers
LOG_LEVEL=info
LOG_FORMAT=json
```

## Integration spec

After wiring, verify with:

1. Start the app — structured JSON logs should appear on stdout with `service_name`, `level`, and `timestamp` fields
2. Hit any API endpoint (e.g. `GET /api/v1/health`) — a log line with `http.method` and `http.url` should appear
3. Call `obs.Log(ctx, LogLevelInfo, "test", nil)` from any handler — verify the log line appears with the correct level
4. Shut down the app gracefully (Ctrl+C) — verify `Shutdown()` completes without errors in the logs

## Do NOT

- Skip span creation on handler functions — every handler should have a span
- Log sensitive data (tokens, passwords, card numbers) — ever
- Use `fmt.Println` or `log.Println` — always use the structured logger
- Forget to call `Shutdown()` on graceful exit — flushes pending spans
