// Package elasticsearch implements the SearchService interface using Elasticsearch.
package elasticsearch

import (
	"context"
	"fmt"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Elasticsearch search provider.
type Config struct {
	// URL is the Elasticsearch endpoint (e.g. http://localhost:9200).
	URL string

	// APIKey is used for authenticated access to Elastic Cloud.
	APIKey string

	// IndexPrefix is prepended to all index names (e.g. "myapp_").
	IndexPrefix string
}

// Service implements contracts.SearchService using Elasticsearch.
type Service struct {
	cfg Config
	// TODO: add go-elasticsearch client
}

// New creates a new Elasticsearch search service.
func New(cfg Config) (*Service, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("elasticsearch: URL is required")
	}
	return &Service{cfg: cfg}, nil
}

func (s *Service) Index(ctx context.Context, indexName string, id string, doc any) error {
	// TODO: implement using github.com/elastic/go-elasticsearch/v8 esapi.IndexRequest
	panic("not implemented")
}

func (s *Service) IndexBatch(ctx context.Context, indexName string, docs []contracts.IndexDocument) error {
	// TODO: implement using Elasticsearch Bulk API
	panic("not implemented")
}

func (s *Service) Search(ctx context.Context, indexName string, query contracts.SearchQuery) (*contracts.SearchResult, error) {
	// TODO: implement using esapi.SearchRequest with query DSL
	panic("not implemented")
}

func (s *Service) Delete(ctx context.Context, indexName string, id string) error {
	// TODO: implement using esapi.DeleteRequest
	panic("not implemented")
}

func (s *Service) DeleteIndex(ctx context.Context, indexName string) error {
	// TODO: implement using esapi.IndicesDeleteRequest
	panic("not implemented")
}
