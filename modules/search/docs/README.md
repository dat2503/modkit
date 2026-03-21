# Search Module

Full-text search for modkit projects using Elasticsearch.

## Overview

The search module wraps [Elasticsearch](https://www.elastic.co/) to provide full-text search, faceted filtering, and relevance ranking. Use when Postgres `tsvector` or `ILIKE` is not powerful enough.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `elasticsearch` | Elasticsearch | v2 | Go, Bun |

## Setup

**Local (Docker):**
```bash
docker run -p 9200:9200 -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" elasticsearch:8.12.0
```

**Cloud:** Use [Elastic Cloud](https://cloud.elastic.co) — free trial includes 14 days, then ~$16/month for a small cluster.

Set env vars:
```
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_API_KEY=...
```

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **observability** (optional) — traces search queries
