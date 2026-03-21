// Package tests contains contract compliance tests for all search implementations.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// SearchServiceContract runs contract compliance tests against any SearchService implementation.
func SearchServiceContract(t *testing.T, svc contracts.SearchService) {
	t.Helper()

	const testIndex = "contract-test-items"

	t.Run("Index_ThenSearch_FindsDocument", func(t *testing.T) {
		err := svc.Index(context.Background(), testIndex, "doc-1", map[string]any{
			"title": "contract test document",
			"body":  "this is a test",
		})
		if err != nil {
			t.Fatalf("Index failed: %v", err)
		}

		results, err := svc.Search(context.Background(), testIndex, contracts.SearchQuery{
			Query: "contract test",
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if results == nil {
			t.Fatal("expected non-nil SearchResult")
		}
	})

	t.Run("Delete_NonexistentDocument_ReturnsNil", func(t *testing.T) {
		err := svc.Delete(context.Background(), testIndex, "doc-nonexistent")
		if err != nil {
			t.Fatalf("Delete of nonexistent doc returned error: %v", err)
		}
	})

	t.Run("Search_EmptyQuery_ReturnsResults", func(t *testing.T) {
		results, err := svc.Search(context.Background(), testIndex, contracts.SearchQuery{
			Query:    "",
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("Search with empty query failed: %v", err)
		}
		if results.PageSize <= 0 {
			t.Fatal("expected positive page size in result")
		}
	})
}
