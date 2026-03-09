import type { RequestExecutor } from "../client.js";
import type {
  SemanticSearchRequest,
  SemanticSearchResponse,
  EmbeddingStatusResponse,
} from "../types.js";

/** Resource for semantic search operations. */
export class SearchResource {
  constructor(private readonly client: RequestExecutor) {}

  /** Perform a semantic search across prompt versions. */
  async semantic(data: SemanticSearchRequest): Promise<SemanticSearchResponse> {
    return this.client.request<SemanticSearchResponse>(
      "POST",
      "/api/v1/search/semantic",
      data,
    );
  }

  /** Check the status of the embedding service. */
  async embeddingStatus(): Promise<EmbeddingStatusResponse> {
    return this.client.request<EmbeddingStatusResponse>(
      "GET",
      "/api/v1/search/embedding-status",
    );
  }
}
