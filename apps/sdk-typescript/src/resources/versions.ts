import type { RequestExecutor } from "../client.js";
import type {
  PromptVersion,
  CreateVersionRequest,
  UpdateVersionStatusRequest,
  SemanticDiff,
  LintResult,
  TextDiffResult,
} from "../types.js";

/** Resource for managing prompt versions. */
export class VersionsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all versions of a prompt. */
  async list(promptId: string): Promise<PromptVersion[]> {
    return this.client.request<PromptVersion[]>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions`,
    );
  }

  /** Create a new version for a prompt. */
  async create(
    promptId: string,
    data: CreateVersionRequest,
  ): Promise<PromptVersion> {
    return this.client.request<PromptVersion>(
      "POST",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions`,
      data,
    );
  }

  /** Get a specific version of a prompt. */
  async get(promptId: string, version: number): Promise<PromptVersion> {
    return this.client.request<PromptVersion>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions/${version}`,
    );
  }

  /** Update the status of a prompt version. */
  async updateStatus(
    promptId: string,
    version: number,
    data: UpdateVersionStatusRequest,
  ): Promise<PromptVersion> {
    return this.client.request<PromptVersion>(
      "PUT",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions/${version}/status`,
      data,
    );
  }

  /** Get the lint result for a prompt version. */
  async lint(promptId: string, version: number): Promise<LintResult> {
    return this.client.request<LintResult>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions/${version}/lint`,
    );
  }

  /**
   * Get a line-by-line text diff for a prompt version compared to a previous version.
   * If fromVersion is omitted, it defaults to (version - 1) on the server.
   */
  async textDiff(
    promptId: string,
    version: number,
    fromVersion?: number,
  ): Promise<TextDiffResult> {
    const query =
      fromVersion !== undefined ? `?from=${fromVersion}` : "";
    return this.client.request<TextDiffResult>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions/${version}/text-diff${query}`,
    );
  }

  /** Get a semantic diff between two prompt versions. */
  async semanticDiff(
    promptId: string,
    v1: number,
    v2: number,
  ): Promise<SemanticDiff> {
    return this.client.request<SemanticDiff>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/semantic-diff/${v1}/${v2}`,
    );
  }
}
