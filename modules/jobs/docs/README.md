# Jobs Module

Background job queue for modkit projects.

## Overview

The jobs module provides a Redis-backed job queue for async processing. Uses [Asynq](https://github.com/hibiken/asynq) for Go runtime and [BullMQ](https://bullmq.io) for Bun runtime. Both share the same Redis instance as the cache module — no additional infrastructure needed.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `asynq` | Asynq | MVP | Go only |
| `bullmq` | BullMQ | MVP | Bun only |

## Setup

No additional setup — jobs uses the same Redis instance configured for the `cache` module.

Make sure `REDIS_URL` is set (from the cache module config) and set job-specific vars:
```
JOBS_PROVIDER=asynq      # or bullmq for Bun
JOBS_CONCURRENCY=10
JOBS_MAX_RETRIES=3
```

## BullMQ Dashboard (Bun only)

BullMQ has an optional web dashboard for monitoring jobs:
```bash
bun add @bull-board/api @bull-board/hono
```

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **cache** (required) — Redis instance used as queue backend
- **observability** (optional) — traces job execution
- **error-tracking** (optional) — reports job failures
