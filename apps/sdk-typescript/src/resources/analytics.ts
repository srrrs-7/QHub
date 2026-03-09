import type { RequestExecutor } from "../client.js";
import type {
  PromptAnalytics,
  VersionAnalytics,
  ProjectAnalytics,
  DailyTrend,
} from "../types.js";

/** Resource for accessing analytics data. */
export class AnalyticsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** Get analytics for all versions of a prompt. */
  async promptAnalytics(promptId: string): Promise<PromptAnalytics[]> {
    return this.client.request<PromptAnalytics[]>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/analytics`,
    );
  }

  /** Get analytics for a specific prompt version. */
  async versionAnalytics(
    promptId: string,
    version: number,
  ): Promise<VersionAnalytics> {
    return this.client.request<VersionAnalytics>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/versions/${version}/analytics`,
    );
  }

  /** Get analytics for a project. */
  async projectAnalytics(projectId: string): Promise<ProjectAnalytics[]> {
    return this.client.request<ProjectAnalytics[]>(
      "GET",
      `/api/v1/projects/${encodeURIComponent(projectId)}/analytics`,
    );
  }

  /** Get daily trend data for a prompt. */
  async dailyTrend(promptId: string): Promise<DailyTrend[]> {
    return this.client.request<DailyTrend[]>(
      "GET",
      `/api/v1/prompts/${encodeURIComponent(promptId)}/trend`,
    );
  }
}
