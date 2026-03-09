import type { RequestExecutor } from "../client.js";
import type {
  Organization,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
} from "../types.js";

/** Resource for managing organizations. */
export class OrganizationsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** Create a new organization. */
  async create(data: CreateOrganizationRequest): Promise<Organization> {
    return this.client.request<Organization>(
      "POST",
      "/api/v1/organizations",
      data,
    );
  }

  /** Get an organization by slug. */
  async get(orgSlug: string): Promise<Organization> {
    return this.client.request<Organization>(
      "GET",
      `/api/v1/organizations/${encodeURIComponent(orgSlug)}`,
    );
  }

  /** Update an organization by slug. */
  async update(
    orgSlug: string,
    data: UpdateOrganizationRequest,
  ): Promise<Organization> {
    return this.client.request<Organization>(
      "PUT",
      `/api/v1/organizations/${encodeURIComponent(orgSlug)}`,
      data,
    );
  }
}
