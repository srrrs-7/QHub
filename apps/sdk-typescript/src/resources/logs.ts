import type { RequestExecutor } from "../client.js";
import type {
  ExecutionLog,
  CreateLogRequest,
  ListLogsResponse,
} from "../types.js";

/** Resource for managing execution logs. */
export class LogsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List execution logs. */
  async list(): Promise<ListLogsResponse> {
    return this.client.request<ListLogsResponse>("GET", "/api/v1/logs");
  }

  /** Create a single execution log. */
  async create(data: CreateLogRequest): Promise<ExecutionLog> {
    return this.client.request<ExecutionLog>("POST", "/api/v1/logs", data);
  }

  /** Create multiple execution logs in a single request. */
  async createBatch(data: CreateLogRequest[]): Promise<ExecutionLog[]> {
    return this.client.request<ExecutionLog[]>("POST", "/api/v1/logs/batch", {
      logs: data,
    });
  }

  /** Get a single execution log by ID. */
  async get(id: string): Promise<ExecutionLog> {
    return this.client.request<ExecutionLog>(
      "GET",
      `/api/v1/logs/${encodeURIComponent(id)}`,
    );
  }
}
