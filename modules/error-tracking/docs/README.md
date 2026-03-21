# Error Tracking Module

Error and exception tracking for modkit projects using Sentry.

## Overview

The error tracking module wraps [Sentry](https://sentry.io) to capture unhandled errors, panics, and performance issues. Free tier includes 5,000 errors/month which is sufficient for MVP.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `sentry` | Sentry | MVP | Go, Bun |

## Setup

1. Create a Sentry account and project at [sentry.io](https://sentry.io)
2. Copy the DSN from Project Settings → Client Keys
3. Set env vars:
   ```
   SENTRY_DSN=https://...@sentry.io/...
   SENTRY_ENVIRONMENT=production
   ```

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **observability** (required) — Sentry integrates with OpenTelemetry for trace context propagation
