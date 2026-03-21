# Feature Flags Module

Feature flag management for modkit projects using Flagsmith.

## Overview

The feature flags module wraps [Flagsmith](https://flagsmith.com) to provide feature gating, phased rollouts, and A/B testing. Flagsmith is open source and can be self-hosted or used via their cloud service.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `flagsmith` | Flagsmith | v2 | Go, Bun |

## Setup

1. Create a Flagsmith account at [flagsmith.com](https://flagsmith.com) (free tier: unlimited flags, 50k requests/month)
2. Create a project and environment
3. Copy the Server-side SDK Key
4. Set env vars:
   ```
   FLAGSMITH_SERVER_KEY=ser....
   ```

## Self-hosting

Flagsmith is open source. Deploy with Docker:
```bash
docker run -p 8000:8000 flagsmith/flagsmith:latest
```

Then set `FLAGSMITH_API_URL=http://localhost:8000/api/v1/`.

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **cache** (optional but recommended) — caches flag values to reduce API calls across instances
