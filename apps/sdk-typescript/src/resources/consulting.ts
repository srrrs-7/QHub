import type { RequestExecutor } from "../client.js";
import type {
  ConsultingSession,
  CreateSessionRequest,
  ConsultingMessage,
  CreateMessageRequest,
} from "../types.js";

/** Resource for managing consulting sessions and messages. */
export class ConsultingResource {
  constructor(private readonly client: RequestExecutor) {}

  /** List all consulting sessions. */
  async listSessions(): Promise<ConsultingSession[]> {
    return this.client.request<ConsultingSession[]>(
      "GET",
      "/api/v1/consulting/sessions",
    );
  }

  /** Create a new consulting session. */
  async createSession(
    data: CreateSessionRequest,
  ): Promise<ConsultingSession> {
    return this.client.request<ConsultingSession>(
      "POST",
      "/api/v1/consulting/sessions",
      data,
    );
  }

  /** Get a consulting session by ID. */
  async getSession(sessionId: string): Promise<ConsultingSession> {
    return this.client.request<ConsultingSession>(
      "GET",
      `/api/v1/consulting/sessions/${encodeURIComponent(sessionId)}`,
    );
  }

  /** List messages in a consulting session. */
  async listMessages(sessionId: string): Promise<ConsultingMessage[]> {
    return this.client.request<ConsultingMessage[]>(
      "GET",
      `/api/v1/consulting/sessions/${encodeURIComponent(sessionId)}/messages`,
    );
  }

  /** Post a message to a consulting session. */
  async createMessage(
    sessionId: string,
    data: CreateMessageRequest,
  ): Promise<ConsultingMessage> {
    return this.client.request<ConsultingMessage>(
      "POST",
      `/api/v1/consulting/sessions/${encodeURIComponent(sessionId)}/messages`,
      data,
    );
  }
}
