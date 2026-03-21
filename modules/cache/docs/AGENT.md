# Cache Module — Agent Instructions

## When to use

Include this module when:
- The `auth` module is selected (required — auth uses cache for session storage)
- The `jobs` module is selected (required — jobs uses cache as the queue backend)
- App needs rate limiting (API endpoints, login attempts)
- App has frequently-read, rarely-written data (dashboard stats, feature config)
- App needs distributed locks or deduplication

## How to wire

### Go

1. Import `CacheService` from `contracts/go/cache.go`
2. Initialize in bootstrap **before auth and jobs**:
   ```go
   cacheSvc := redis.New(redis.Config{
       URL:            cfg.Cache.RedisURL,
       MaxConnections: cfg.Cache.MaxConnections,
   })
   ```
3. Inject into auth, jobs, and any handler that needs caching

### Bun (TypeScript)

1. Import `ICacheService` from `contracts/ts/cache.ts`
2. Initialize before auth and jobs:
   ```typescript
   const cache = new RedisCacheService({
     url: config.cache.redisUrl,
     maxConnections: config.cache.maxConnections,
   })
   ```

## Common patterns

### Cache-aside (read-through)

```go
cacheKey := "dashboard:stats:" + userID
cached, err := cache.Get(ctx, cacheKey)
if err == nil {
    json.Unmarshal(cached, &stats)
    return stats, nil
}
// Cache miss — compute and store
stats, err = computeDashboardStats(ctx, db, userID)
data, _ := json.Marshal(stats)
cache.Set(ctx, cacheKey, data, 5*time.Minute)
return stats, nil
```

### Rate limiting (token bucket via Increment)

```go
key := "rate:login:" + ipAddress
count, _ := cache.Increment(ctx, key, 1)
if count == 1 {
    cache.Expire(ctx, key, time.Minute) // set TTL on first increment
}
if count > 5 {
    return errors.New("rate limit exceeded")
}
```

### Distributed lock (SetNX)

```go
lockKey := "lock:invoice-generate:" + invoiceID
acquired, _ := cache.SetNX(ctx, lockKey, []byte("1"), 30*time.Second)
if !acquired {
    return errors.New("already processing")
}
defer cache.Delete(ctx, lockKey)
// ... do work
```

## Key namespacing

Always namespace your cache keys to avoid collisions:
```
sessions:{sessionID}
rate:{action}:{identifier}
lock:{resource}:{id}
cache:{entity}:{id}
```

## Required env vars

```
CACHE_PROVIDER=redis
REDIS_URL=redis://localhost:6379      # sensitive (contains password in production)
REDIS_MAX_CONNECTIONS=10
```

## Do NOT

- Cache sensitive data without encryption (passwords, full credit card info)
- Use cache as the primary store — it's volatile, data can be evicted
- Use long TTLs (>1 hour) for user-specific data without invalidation logic
- Share cache keyspace between environments (prefix keys with env name in production)
