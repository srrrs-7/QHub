import type { RequestExecutor } from "../client.js";
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
} from "../types.js";

/** Resource for managing projects within an organization. */
export class ProjectsResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all projects in an organization. */
  async list(orgId: string): Promise<Project[]> {
    return this.client.request<Project[]>(
      "GET",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/projects`,
    );
  }

  /** Create a new project in an organization. */
  async create(orgId: string, data: CreateProjectRequest): Promise<Project> {
    return this.client.request<Project>(
      "POST",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/projects`,
      data,
    );
  }

  /** Get a project by slug within an organization. */
  async get(orgId: string, projectSlug: string): Promise<Project> {
    return this.client.request<Project>(
      "GET",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/projects/${encodeURIComponent(projectSlug)}`,
    );
  }

  /** Update a project by slug within an organization. */
  async update(
    orgId: string,
    projectSlug: string,
    data: UpdateProjectRequest,
  ): Promise<Project> {
    return this.client.request<Project>(
      "PUT",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/projects/${encodeURIComponent(projectSlug)}`,
      data,
    );
  }

  /** Delete a project by slug within an organization. */
  async delete(orgId: string, projectSlug: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/projects/${encodeURIComponent(projectSlug)}`,
    );
  }
}
