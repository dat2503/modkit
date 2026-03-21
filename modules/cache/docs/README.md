# Cache Module

Key-value caching for modkit projects using Redis.

## Overview

The cache module provides a Redis-backed key-value store for sessions, rate limiting, distributed locks, and hot data. It is a required dependency of the `auth` and `jobs` modules.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `redis` | Redis | MVP | Go, Bun |

## Setup

1. Run Redis locally via Docker: `docker run -p 6379:6379 redis:7-alpine`
2. Or use Redis Cloud (free tier: 30MB, sufficient for MVP)
3. Set env vars:
   ```
   REDIS_URL=redis://localhost:6379
   ```

## Configuration

See `config.schema.json` for all environment variables.

## Local development

Redis is included in the project's `docker-compose.yaml`:
```bash
docker-compose up -d redis
```
