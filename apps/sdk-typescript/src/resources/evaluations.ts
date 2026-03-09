import type { RequestExecutor } from "../client.js";
import type { Evaluation, CreateEvaluationRequest } from "../types.js";

/** Resource for managing evaluations. */
export class EvaluationsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** Create a new evaluation. */
  async create(data: CreateEvaluationRequest): Promise<Evaluation> {
    return this.client.request<Evaluation>(
      "POST",
      "/api/v1/evaluations",
      data,
    );
  }

  /** Get an evaluation by ID. */
  async get(id: string): Promise<Evaluation> {
    return this.client.request<Evaluation>(
      "GET",
      `/api/v1/evaluations/${encodeURIComponent(id)}`,
    );
  }

  /** List evaluations for a specific execution log. */
  async listByLog(logId: string): Promise<Evaluation[]> {
    return this.client.request<Evaluation[]>(
      "GET",
      `/api/v1/logs/${encodeURIComponent(logId)}/evaluations`,
    );
  }
}
