# Search Module — Agent Instructions

## When to use

Include this module when:
- The brief explicitly mentions search functionality with relevance ranking
- App needs to filter >100k records across multiple fields efficiently
- App needs autocomplete / typeahead search
- Faceted search is required (filter by category, price range, tags)

Do NOT include when:
- Simple text matching on a few hundred rows — Postgres `ILIKE` or `tsvector` is sufficient
- Search is just filtering by exact values — use database queries
- The brief doesn't mention search — don't add it speculatively

## How to wire

### Go

1. Import `SearchService` from `contracts/go/search.go`
2. Initialize in bootstrap:
   ```go
   searchSvc, err := elasticsearch.New(elasticsearch.Config{
       URL:         cfg.Search.ElasticsearchURL,
       APIKey:      cfg.Search.APIKey,
       IndexPrefix: cfg.Search.IndexPrefix,
   })
   ```
3. Create indexes during app startup or migration:
   ```go
   // Typically called once during setup, not on every startup
   ```
4. Index documents after database writes
5. Inject into search handler endpoints

### Bun (TypeScript)

1. Import `ISearchService` from `contracts/ts/search.ts`
2. Initialize in bootstrap:
   ```typescript
   const search = new ElasticsearchService({
     url: config.search.elasticsearchUrl,
     apiKey: config.search.apiKey,
     indexPrefix: config.search.indexPrefix,
   })
   ```

## Index naming convention

Use lowercase, hyphen-separated index names prefixed with the app name:
```
myapp_invoices
myapp_users
myapp_products
```

Set `ELASTICSEARCH_INDEX_PREFIX=myapp_` to apply this automatically.

## Keep search in sync with the database

Always write to the database first, then index to search — never the reverse:

```go
// In your handler or job:
invoice, err := repo.Create(ctx, req)  // database first
if err != nil { ... }

// Index asynchronously (via jobs module) to avoid blocking the response
jobs.Enqueue(ctx, "search:index_invoice", IndexInvoicePayload{InvoiceID: invoice.ID})
```

## Search handler pattern (Go)

```go
func (h *InvoiceHandler) Search(w http.ResponseWriter, r *http.Request) {
    ctx, span := h.obs.StartSpan(r.Context(), "invoice.search")
    defer span.End()

    results, err := h.search.Search(ctx, "invoices", contracts.SearchQuery{
        Query:    r.URL.Query().Get("q"),
        Filters:  map[string]any{"freelancer_id": userID},
        Page:     pageFromQuery(r),
        PageSize: 20,
    })
    if err != nil {
        // fall back to database query on search failure
        span.RecordError(err)
    }
    writeJSON(w, http.StatusOK, results)
}
```

## Required env vars

```
SEARCH_PROVIDER=elasticsearch
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_API_KEY=...            # sensitive, use for cloud
ELASTICSEARCH_INDEX_PREFIX=myapp_
```

## Local development

Run Elasticsearch via Docker:
```bash
docker run -p 9200:9200 -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" elasticsearch:8.12.0
```

## Integration spec

After wiring, verify with:

1. Start Elasticsearch: `make infra-up` (or `docker run -p 9200:9200 -e "discovery.type=single-node" -e "xpack.security.enabled=false" elasticsearch:8.13.0`)
2. Verify Elasticsearch is healthy: `curl http://localhost:9200/_cluster/health` should not show `red`
3. Index a test document:
   ```go
   search.Index(ctx, "test_items", "1", map[string]any{"title": "Hello World", "body": "integration test"})
   ```
4. Search for it:
   ```go
   results, err := search.Search(ctx, "test_items", contracts.SearchQuery{Query: "Hello"})
   // results should contain the document with ID "1"
   ```
5. Clean up: delete the test index via `curl -X DELETE http://localhost:9200/test_items`

## Do NOT

- Index sensitive data (passwords, raw card numbers) — only index what users can search
- Use search as the source of truth — database is authoritative, search is a projection
- Block HTTP responses on search indexing — always index asynchronously
- Ignore search failures — implement database fallback for critical search paths
