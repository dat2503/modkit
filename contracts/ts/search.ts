/**
 * ISearchService provides full-text search and filtering over large datasets.
 * Use when Postgres ILIKE queries are insufficient — typically >100k records
 * or complex multi-field relevance ranking is needed.
 */
export interface ISearchService {
  /**
   * Adds or updates a document in the search index.
   * id is the document's unique identifier (typically the database row UUID).
   */
  index(indexName: string, id: string, doc: Record<string, unknown>): Promise<void>;

  /**
   * Indexes multiple documents in a single operation.
   * More efficient than calling index() in a loop for bulk operations.
   */
  indexBatch(indexName: string, docs: IndexDocument[]): Promise<void>;

  /**
   * Executes a search query and returns matching documents.
   */
  search(indexName: string, query: SearchQuery): Promise<SearchResult>;

  /**
   * Removes a document from the search index by ID.
   * Returns without error if the document did not exist.
   */
  delete(indexName: string, id: string): Promise<void>;

  /**
   * Removes an entire search index and all its documents.
   * Use with care — this is irreversible.
   */
  deleteIndex(indexName: string): Promise<void>;
}

/** A single document to be indexed in a batch operation. */
export interface IndexDocument {
  id: string;
  doc: Record<string, unknown>;
}

/** Defines a search request. */
export interface SearchQuery {
  /** Full-text search string. */
  query: string;

  /** Restricts the search to specific fields. If empty, all fields are searched. */
  fields?: string[];

  /** Exact-match constraints applied before full-text search. */
  filters?: Record<string, unknown>;

  /** 1-based page number. Defaults to 1. */
  page?: number;

  /** Results per page. Defaults to 20, max 100. */
  pageSize?: number;

  /** Field to sort results by. Defaults to relevance score. */
  sortBy?: string;

  /** Sorts in descending order if true. */
  sortDesc?: boolean;
}

/** Holds the results of a search query. */
export interface SearchResult {
  /** Matching documents for this page. */
  hits: SearchHit[];

  /** Total number of matching documents across all pages. */
  total: number;

  page: number;
  pageSize: number;
}

/** A single search result. */
export interface SearchHit {
  /** The document's unique identifier. */
  id: string;

  /** Relevance score (higher = more relevant). */
  score: number;

  /** The raw document data as originally indexed. */
  source: Record<string, unknown>;
}
