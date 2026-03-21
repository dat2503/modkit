# Observability Module

Distributed tracing, structured logging, and metrics for modkit projects using OpenTelemetry.

## Overview

The observability module wraps [OpenTelemetry](https://opentelemetry.io/) to provide vendor-neutral telemetry. Export to any OTLP-compatible backend:
- [Jaeger](https://www.jaegertracing.io/) (self-hosted, free)
- [Honeycomb](https://honeycomb.io/) (cloud, generous free tier)
- [Datadog](https://www.datadoghq.com/)
- [Grafana Cloud](https://grafana.com/products/cloud/)

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `otel` | OpenTelemetry | MVP | Go, Bun |

## Setup

For local development, run Jaeger via Docker:
```bash
docker run -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one:latest
```

Then open the Jaeger UI at http://localhost:16686.

Set env vars:
```
OTEL_SERVICE_NAME=myapp-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Configuration

See `config.schema.json` for all environment variables.
