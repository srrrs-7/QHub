import type { RequestExecutor } from "../client.js";
import type {
  Member,
  AddMemberRequest,
  UpdateMemberRequest,
} from "../types.js";

/** Resource for managing organization members. */
export class MembersResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all members of an organization. */
  async list(orgId: string): Promise<Member[]> {
    return this.client.request<Member[]>(
      "GET",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/members`,
    );
  }

  /** Add a member to an organization. */
  async add(orgId: string, data: AddMemberRequest): Promise<Member> {
    return this.client.request<Member>(
      "POST",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/members`,
      data,
    );
  }

  /** Update a member's role in an organization. */
  async update(
    orgId: string,
    userId: string,
    data: UpdateMemberRequest,
  ): Promise<Member> {
    return this.client.request<Member>(
      "PUT",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/members/${encodeURIComponent(userId)}`,
      data,
    );
  }

  /** Remove a member from an organization. */
  async remove(orgId: string, userId: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/api/v1/organizations/${encodeURIComponent(orgId)}/members/${encodeURIComponent(userId)}`,
    );
  }
}
