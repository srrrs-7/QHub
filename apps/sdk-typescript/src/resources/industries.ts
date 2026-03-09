import type { RequestExecutor } from "../client.js";
import type {
  IndustryConfig,
  CreateIndustryConfigRequest,
  UpdateIndustryConfigRequest,
  Benchmark,
  ComplianceCheckRequest,
  ComplianceCheckResponse,
} from "../types.js";

/** Resource for managing industry configurations. */
export class IndustriesResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all industry configurations. */
  async list(): Promise<IndustryConfig[]> {
    return this.client.request<IndustryConfig[]>(
      "GET",
      "/api/v1/industries",
    );
  }

  /** Create a new industry configuration. */
  async create(data: CreateIndustryConfigRequest): Promise<IndustryConfig> {
    return this.client.request<IndustryConfig>(
      "POST",
      "/api/v1/industries",
      data,
    );
  }

  /** Get an industry configuration by slug. */
  async get(slug: string): Promise<IndustryConfig> {
    return this.client.request<IndustryConfig>(
      "GET",
      `/api/v1/industries/${encodeURIComponent(slug)}`,
    );
  }

  /** Update an industry configuration by slug. */
  async update(
    slug: string,
    data: UpdateIndustryConfigRequest,
  ): Promise<IndustryConfig> {
    return this.client.request<IndustryConfig>(
      "PUT",
      `/api/v1/industries/${encodeURIComponent(slug)}`,
      data,
    );
  }

  /** List benchmarks for an industry. */
  async listBenchmarks(slug: string): Promise<Benchmark[]> {
    return this.client.request<Benchmark[]>(
      "GET",
      `/api/v1/industries/${encodeURIComponent(slug)}/benchmarks`,
    );
  }

  /** Run a compliance check against an industry's rules. */
  async complianceCheck(
    slug: string,
    data: ComplianceCheckRequest,
  ): Promise<ComplianceCheckResponse> {
    return this.client.request<ComplianceCheckResponse>(
      "POST",
      `/api/v1/industries/${encodeURIComponent(slug)}/compliance-check`,
      data,
    );
  }
}
