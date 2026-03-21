package contracts

import "context"

// SearchService provides full-text search and filtering over large datasets.
// Use when Postgres ILIKE queries are insufficient — typically >100k records
// or complex multi-field relevance ranking is needed.
type SearchService interface {
	// Index adds or updates a document in the search index.
	// id is the document's unique identifier (typically the database row UUID).
	// doc must be JSON-serializable — all searchable fields should be included.
	Index(ctx context.Context, indexName string, id string, doc any) error

	// IndexBatch indexes multiple documents in a single operation.
	// More efficient than calling Index in a loop for bulk operations.
	IndexBatch(ctx context.Context, indexName string, docs []IndexDocument) error

	// Search executes a search query and returns matching documents.
	Search(ctx context.Context, indexName string, query SearchQuery) (*SearchResult, error)

	// Delete removes a document from the search index by ID.
	// Returns nil if the document did not exist.
	Delete(ctx context.Context, indexName string, id string) error

	// DeleteIndex removes an entire search index and all its documents.
	// Use with care — this is irreversible.
	DeleteIndex(ctx context.Context, indexName string) error
}

// IndexDocument is a single document to be indexed in a batch operation.
type IndexDocument struct {
	// ID is the document's unique identifier.
	ID string

	// Doc is the document data — must be JSON-serializable.
	Doc any
}

// SearchQuery defines a search request.
type SearchQuery struct {
	// Query is the full-text search string.
	Query string

	// Fields restricts the search to specific fields. If empty, all fields are searched.
	Fields []string

	// Filters are exact-match constraints applied before full-text search.
	Filters map[string]any

	// Page is the 1-based page number. Defaults to 1.
	Page int

	// PageSize is the number of results per page. Defaults to 20, max 100.
	PageSize int

	// SortBy is the field to sort results by. Defaults to relevance score.
	SortBy string

	// SortDesc sorts in descending order if true.
	SortDesc bool
}

// SearchResult holds the results of a search query.
type SearchResult struct {
	// Hits is the list of matching documents for this page.
	Hits []SearchHit

	// Total is the total number of matching documents across all pages.
	Total int

	// Page is the current page number.
	Page int

	// PageSize is the page size used.
	PageSize int
}

// SearchHit is a single search result.
type SearchHit struct {
	// ID is the document's unique identifier (matches what was passed to Index).
	ID string

	// Score is the relevance score (higher = more relevant).
	Score float64

	// Source is the raw document data as originally indexed.
	Source map[string]any
}
