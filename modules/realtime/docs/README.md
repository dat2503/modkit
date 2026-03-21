# Realtime Module

Real-time bidirectional communication for modkit projects using native WebSockets.

## Overview

The realtime module enables live updates from server to connected clients. Built on native WebSockets — no external service required. Uses Redis (via the cache module) to track connections and fan out events across multiple server instances.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `websocket` | WebSocket (native) | v2 | Go, Bun |

## Use cases

- Live dashboard updates (new invoice paid, status changed)
- Notification toasts without page refresh
- Live chat / messaging
- Presence indicators (who's online)

## Multi-instance scaling

Connection state is stored in Redis (via the cache module). Events published on any server instance are broadcast to all connected clients across all instances via Redis pub/sub.

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **cache** (required) — connection registry and pub/sub fan-out
- **auth** (required) — authenticate WebSocket connections
- **observability** (optional) — traces connection lifecycle and event publishing
