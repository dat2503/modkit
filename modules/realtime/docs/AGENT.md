# Realtime Module — Agent Instructions

## When to use

Include this module ONLY when the project brief explicitly mentions:
- Live/real-time updates (dashboards, feeds, status pages)
- No page refresh required for new data
- Live chat or messaging
- Collaborative features (shared editing, presence indicators)

Do NOT include by default. Most SaaS apps don't need realtime — polling or page refresh is sufficient for MVP.

## Dependencies

**Required:** `cache` (connection registry, pub/sub backend) and `auth` (authenticate WebSocket connections).
Always initialize cache → auth → realtime in that order.

## How to wire

### Go

1. Import `RealtimeService` from `contracts/go/realtime.go`
2. Initialize after cache and auth:
   ```go
   realtimeSvc := websocket.New(websocket.Config{
       MaxConnectionsPerUser: cfg.Realtime.MaxConnectionsPerUser,
       PingInterval:          cfg.Realtime.PingInterval,
   }, cacheSvc)
   ```
3. Register the WebSocket upgrade endpoint:
   ```go
   router.Get("/ws", wsUpgradeHandler(realtimeSvc, authSvc))
   ```
4. Publish events from any handler:
   ```go
   realtimeSvc.PublishToUser(ctx, userID, contracts.RealtimeEvent{
       Type:    "invoice.paid",
       Payload: map[string]any{"invoiceId": invoice.ID, "amount": invoice.Total},
   })
   ```

### Bun (TypeScript)

1. Import `IRealtimeService` from `contracts/ts/realtime.ts`
2. Initialize after cache and auth:
   ```typescript
   const realtime = new WebSocketRealtimeService({
     maxConnectionsPerUser: config.realtime.maxConnectionsPerUser,
   }, cache)
   ```
3. Register WebSocket upgrade handler:
   ```typescript
   app.get('/ws', wsUpgradeHandler(realtime, auth))
   ```

## WebSocket upgrade pattern (Go)

```go
func wsUpgradeHandler(rt contracts.RealtimeService, auth contracts.AuthService) http.HandlerFunc {
    upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
    return func(w http.ResponseWriter, r *http.Request) {
        // Authenticate before upgrading
        user, err := auth.ValidateToken(r.Context(), tokenFromQuery(r))
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            return
        }
        rt.HandleConnection(r.Context(), conn, user.ID)
    }
}
```

## Frontend (Next.js)

```typescript
// apps/web — connect to WebSocket from client component
const ws = new WebSocket(`wss://api.yourapp.com/ws?token=${sessionToken}`)
ws.onmessage = (e) => {
    const event = JSON.parse(e.data)
    if (event.type === 'invoice.paid') {
        // update UI
    }
}
```

## Connection lifecycle

1. Client connects with auth token as query param or in first message
2. Backend validates token, registers connection in cache
3. Server publishes events → client receives in real-time
4. On disconnect, connection removed from registry
5. Ping/pong keeps connection alive through proxies

## Required env vars

```
REALTIME_PROVIDER=websocket
REALTIME_MAX_CONNECTIONS_PER_USER=5
REALTIME_PING_INTERVAL_SECONDS=30
```

## Do NOT

- Allow unauthenticated WebSocket connections — always validate token first
- Publish sensitive data (card numbers, passwords) over WebSocket
- Use realtime as the primary data source — always persist to database first, then publish
- Skip connection cleanup — always handle disconnect events
