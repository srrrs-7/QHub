import { createErrorFromStatus } from "./errors.js";
import { OrganizationsResource } from "./resources/organizations.js";
import { ProjectsResource } from "./resources/projects.js";
import { PromptsResource } from "./resources/prompts.js";
import { VersionsResource } from "./resources/versions.js";
import { LogsResource } from "./resources/logs.js";
import { EvaluationsResource } from "./resources/evaluations.js";
import { ConsultingResource } from "./resources/consulting.js";
import { TagsResource } from "./resources/tags.js";
import { IndustriesResource } from "./resources/industries.js";
import { AnalyticsResource } from "./resources/analytics.js";
import { ApiKeysResource } from "./resources/apikeys.js";
import { MembersResource } from "./resources/members.js";
import { SearchResource } from "./resources/search.js";
import type { ClientOptions } from "./types.js";

/**
 * HTTP method type used by the internal request method.
 */
type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";

/**
 * Internal interface for making API requests, shared with resource classes.
 */
export interface RequestExecutor {
  request<T>(method: HttpMethod, path: string, body?: unknown): Promise<T>;
}

/**
 * QHubClient is the main entry point for the QHub TypeScript SDK.
 *
 * Usage:
 * ```ts
 * const client = new QHubClient({
 *   baseUrl: "http://localhost:8080",
 *   bearerToken: "your-token",
 * });
 *
 * const org = await client.organizations.get("my-org");
 * ```
 */
export class QHubClient implements RequestExecutor {
  private readonly baseUrl: string;
  private readonly bearerToken: string;
  private readonly fetchFn: typeof globalThis.fetch;

  readonly organizations: OrganizationsResource;
  readonly projects: ProjectsResource;
  readonly prompts: PromptsResource;
  readonly versions: VersionsResource;
  readonly logs: LogsResource;
  readonly evaluations: EvaluationsResource;
  readonly consulting: ConsultingResource;
  readonly tags: TagsResource;
  readonly industries: IndustriesResource;
  readonly analytics: AnalyticsResource;
  readonly apiKeys: ApiKeysResource;
  readonly members: MembersResource;
  readonly search: SearchResource;

  constructor(options: ClientOptions) {
    // Strip trailing slash from base URL
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.bearerToken = options.bearerToken;
    this.fetchFn = options.fetch ?? globalThis.fetch;

    this.organizations = new OrganizationsResource(this);
    this.projects = new ProjectsResource(this);
    this.prompts = new PromptsResource(this);
    this.versions = new VersionsResource(this);
    this.logs = new LogsResource(this);
    this.evaluations = new EvaluationsResource(this);
    this.consulting = new ConsultingResource(this);
    this.tags = new TagsResource(this);
    this.industries = new IndustriesResource(this);
    this.analytics = new AnalyticsResource(this);
    this.apiKeys = new ApiKeysResource(this);
    this.members = new MembersResource(this);
    this.search = new SearchResource(this);
  }

  /**
   * Performs an HTTP request against the QHub API.
   * Handles JSON serialization/deserialization and error mapping.
   */
  async request<T>(
    method: HttpMethod,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;

    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.bearerToken}`,
      "Content-Type": "application/json",
    };

    const init: RequestInit = { method, headers };
    if (body !== undefined) {
      init.body = JSON.stringify(body);
    }

    const resp = await this.fetchFn(url, init);
    const text = await resp.text();

    if (!resp.ok) {
      throw createErrorFromStatus(resp.status, text);
    }

    if (text.length === 0) {
      return undefined as T;
    }

    return JSON.parse(text) as T;
  }
}
