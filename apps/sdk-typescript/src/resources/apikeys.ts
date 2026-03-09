import type { RequestExecutor } from "../client.js";
import type {
  ApiKey,
  ApiKeyCreated,
  CreateApiKeyRequest,
} from "../types.js";

/** Resource for managing organization API keys. */
export class ApiKeysResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all API keys for an organization. */
  async list(orgId: string): Promise<ApiKey[]> {
    return this.client.request<ApiKey[]>(
      "GET",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/api-keys`,
    );
  }

  /** Create a new API key for an organization. Returns the raw key (shown once). */
  async create(
    orgId: string,
    data: CreateApiKeyRequest,
  ): Promise<ApiKeyCreated> {
    return this.client.request<ApiKeyCreated>(
      "POST",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/api-keys`,
      data,
    );
  }

  /** Delete an API key by ID. */
  async delete(orgId: string, id: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/api-keys/${encodeURIComponent(id)}`,
    );
  }
}
