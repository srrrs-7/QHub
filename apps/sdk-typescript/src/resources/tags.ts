import type { RequestExecutor } from "../client.js";
import type { Tag, CreateTagRequest, AddPromptTagRequest } from "../types.js";

/** Resource for managing tags and prompt-tag associations. */
export class TagsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all tags. */
  async list(): Promise<Tag[]> {
    return this.client.request<Tag[]>("GET", "/api/v1/tags");
  }

  /** Create a new tag. */
  async create(data: CreateTagRequest): Promise<Tag> {
    return this.client.request<Tag>("POST", "/api/v1/tags", data);
  }

  /** Delete a tag by ID. */
  async delete(id: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/api/v1/tags/${encodeURIComponent(id)}`,
    );
  }

  /** List tags for a prompt. */
  async listByPrompt(promptId: string): Promise<Tag[]> {
    return this.client.request<Tag[]>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/tags`,
    );
  }

  /** Add a tag to a prompt. */
  async addToPrompt(
    promptId: string,
    data: AddPromptTagRequest,
  ): Promise<void> {
    return this.client.request<void>(
      "POST",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/tags`,
      data,
    );
  }

  /** Remove a tag from a prompt. */
  async removeFromPrompt(promptId: string, tagId: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/tags/${encodeURIComponent(tagId)}`,
    );
  }
}
