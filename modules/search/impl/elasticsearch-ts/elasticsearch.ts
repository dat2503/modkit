import type { ISearchService, IndexDocument, SearchQuery, SearchResult } from '../../../contracts/ts/search'

export interface ElasticsearchConfig {
  url: string
  apiKey?: string
  indexPrefix?: string
}

/**
 * ElasticsearchService implements ISearchService using Elasticsearch.
 */
export class ElasticsearchService implements ISearchService {
  constructor(private readonly config: ElasticsearchConfig) {}

  async index(indexName: string, id: string, doc: Record<string, unknown>): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.index()
    console.warn('[elasticsearch] stub: index() not implemented')
  }

  async indexBatch(indexName: string, docs: IndexDocument[]): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.bulk()
    console.warn('[elasticsearch] stub: indexBatch() not implemented')
  }

  async search(indexName: string, query: SearchQuery): Promise<SearchResult> {
    // TODO: implement using @elastic/elasticsearch client.search() with query DSL
    console.warn('[elasticsearch] stub: search() not implemented')
    return { hits: [], total: 0, page: query.page ?? 1, pageSize: query.pageSize ?? 20 }
  }

  async delete(indexName: string, id: string): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.delete()
    console.warn('[elasticsearch] stub: delete() not implemented')
  }

  async deleteIndex(indexName: string): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.indices.delete()
    console.warn('[elasticsearch] stub: deleteIndex() not implemented')
  }
}
