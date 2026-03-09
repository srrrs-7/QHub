import type { RequestExecutor } from "../client.js";
import type {
  Prompt,
  CreatePromptRequest,
  UpdatePromptRequest,
} from "../types.js";

/** Resource for managing prompts within a project. */
export class PromptsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all prompts in a project. */
  async list(projectId: string): Promise<Prompt[]> {
    return this.client.request<Prompt[]>(
      "GET",
      `/api/v1/projects/${encodeURIComponent(projectId)}/prompts`,
    );
  }

  /** Create a new prompt in a project. */
  async create(
    projectId: string,
    data: CreatePromptRequest,
  ): Promise<Prompt> {
    return this.client.request<Prompt>(
      "POST",
      `/api/v1/projects/${encodeURIComponent(projectId)}/prompts`,
      data,
    );
  }

  /** Get a prompt by slug within a project. */
  async get(projectId: string, promptSlug: string): Promise<Prompt> {
    return this.client.request<Prompt>(
      "GET",
      `/api/v1/projects/${encodeURIComponent(projectId)}/prompts/${encodeURIComponent(promptSlug)}`,
    );
  }

  /** Update a prompt by slug within a project. */
  async update(
    projectId: string,
    promptSlug: string,
    data: UpdatePromptRequest,
  ): Promise<Prompt> {
    return this.client.request<Prompt>(
      "PUT",
      `/api/v1/projects/${encodeURIComponent(projectId)}/prompts/${encodeURIComponent(promptSlug)}`,
      data,
    );
  }
}
