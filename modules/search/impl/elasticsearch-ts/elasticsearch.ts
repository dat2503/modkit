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
    throw new Error('not implemented')
  }

  async indexBatch(indexName: string, docs: IndexDocument[]): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.bulk()
    throw new Error('not implemented')
  }

  async search(indexName: string, query: SearchQuery): Promise<SearchResult> {
    // TODO: implement using @elastic/elasticsearch client.search() with query DSL
    throw new Error('not implemented')
  }

  async delete(indexName: string, id: string): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.delete()
    throw new Error('not implemented')
  }

  async deleteIndex(indexName: string): Promise<void> {
    // TODO: implement using @elastic/elasticsearch client.indices.delete()
    throw new Error('not implemented')
  }
}
