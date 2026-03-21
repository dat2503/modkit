# Feature Flags Module — Agent Instructions

## When to use

Include this module when:
- The brief mentions phased rollout, gradual rollout, or canary releases
- A/B testing is needed (show different UI/pricing to different user segments)
- Features need to be enabled/disabled without deploying code
- Different behavior for beta users vs. all users

Do NOT include by default. Feature flags add complexity — only add when the brief specifically calls for them.

## How to wire

### Go

1. Import `FeatureFlagsService` from `contracts/go/feature_flags.go`
2. Initialize in bootstrap:
   ```go
   ffSvc := flagsmith.New(flagsmith.Config{
       ServerKey:   cfg.FeatureFlags.ServerKey,
       APIURL:      cfg.FeatureFlags.APIURL,
       CacheTTL:    cfg.FeatureFlags.CacheTTLSeconds,
   }, cacheSvc) // cache is optional but recommended
   ```
3. Inject into handlers that gate features

### Bun (TypeScript)

1. Import `IFeatureFlagsService` from `contracts/ts/feature-flags.ts`
2. Initialize in bootstrap:
   ```typescript
   const flags = new FlagsmithService({
     serverKey: config.featureFlags.serverKey,
     apiUrl: config.featureFlags.apiUrl,
   }, cache)
   ```

## Common patterns

### Gate a feature in a handler

```go
func (h *BillingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
    // Check if new checkout flow is enabled for this user
    enabled, err := h.flags.IsEnabled(ctx, "new_checkout_flow", contracts.FlagContext{
        UserID: user.ID,
        Email:  user.Email,
        Traits: map[string]any{"plan": user.Plan},
    })
    if err != nil || !enabled {
        // Fall back to old flow
        h.legacyCheckout(w, r)
        return
    }
    h.newCheckout(w, r)
}
```

### A/B test variant

```go
variant, _ := flags.GetVariant(ctx, "pricing_page_experiment", contracts.FlagContext{
    UserID: user.ID,
})
switch variant {
case "variant_a":
    renderPricingV1(w)
case "variant_b":
    renderPricingV2(w)
default:
    renderPricingDefault(w)
}
```

### Batch fetch (avoid multiple round-trips)

```go
// At the start of a request — fetch all flags once
allFlags, _ := flags.GetAllFlags(ctx, contracts.FlagContext{UserID: user.ID})

// Later in the same request — use cached values
if allFlags["dark_mode"].Enabled { ... }
if allFlags["new_dashboard"].Enabled { ... }
```

## Caching

Flagsmith SDK has built-in caching. Set `FLAGSMITH_CACHE_TTL_SECONDS=60` for 1-minute
cache — balances freshness with latency. Use the `cache` module as the cache backend
for shared caching across multiple server instances.

## Required env vars

```
FEATURE_FLAGS_PROVIDER=flagsmith
FLAGSMITH_SERVER_KEY=ser....         # sensitive
FLAGSMITH_API_URL=https://edge.api.flagsmith.com/api/v1/
FLAGSMITH_CACHE_TTL_SECONDS=60
```

## Do NOT

- Use feature flags as a substitute for proper config management
- Gate security-critical features behind flags — hardcode those
- Leave stale flags in code indefinitely — clean up after rollouts complete
- Block requests on flag evaluation — always have a safe default
