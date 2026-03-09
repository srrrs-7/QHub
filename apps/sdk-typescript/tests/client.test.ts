import { describe, it, expect, vi, beforeEach } from "vitest";
import { QHubClient } from "../src/client.js";
import {
  QHubError,
  ValidationError,
  AuthenticationError,
  NotFoundError,
  RateLimitError,
  InternalServerError,
} from "../src/errors.js";

/**
 * Creates a mock fetch function that returns the given response.
 */
function mockFetch(
  status: number,
  body: unknown,
  headers?: Record<string, string>,
): typeof globalThis.fetch {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(typeof body === "string" ? body : JSON.stringify(body)),
    headers: new Headers(headers),
  } as Response);
}

function createClient(fetchFn: typeof globalThis.fetch): QHubClient {
  return new QHubClient({
    baseUrl: "http://localhost:8080",
    bearerToken: "test-token",
    fetch: fetchFn,
  });
}

describe("QHubClient", () => {
  describe("constructor", () => {
    it("strips trailing slash from base URL", () => {
      const fetch = mockFetch(200, { id: "1" });
      const client = new QHubClient({
        baseUrl: "http://example.com/",
        bearerToken: "tok",
        fetch,
      });
      // Verify by making a request and checking the URL
      client.organizations.get("test");
      expect(fetch).toHaveBeenCalledWith(
        "http://example.com/api/v1/organizations/test",
        expect.any(Object),
      );
    });

    it("initializes all resource properties", () => {
      const client = createClient(mockFetch(200, {}));
      expect(client.organizations).toBeDefined();
      expect(client.projects).toBeDefined();
      expect(client.prompts).toBeDefined();
      expect(client.versions).toBeDefined();
      expect(client.logs).toBeDefined();
      expect(client.evaluations).toBeDefined();
      expect(client.consulting).toBeDefined();
      expect(client.tags).toBeDefined();
      expect(client.industries).toBeDefined();
      expect(client.analytics).toBeDefined();
      expect(client.apiKeys).toBeDefined();
      expect(client.members).toBeDefined();
      expect(client.search).toBeDefined();
    });
  });

  describe("request", () => {
    it("sends Authorization header with bearer token", async () => {
      const fetch = mockFetch(200, { id: "org-1" });
      const client = createClient(fetch);

      await client.organizations.get("my-org");

      expect(fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
            "Content-Type": "application/json",
          }),
        }),
      );
    });

    it("sends JSON body for POST requests", async () => {
      const fetch = mockFetch(201, { id: "org-1", name: "Test", slug: "test", plan: "free" });
      const client = createClient(fetch);

      await client.organizations.create({ name: "Test", slug: "test" });

      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/organizations",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ name: "Test", slug: "test" }),
        }),
      );
    });

    it("does not include body for GET requests", async () => {
      const fetch = mockFetch(200, []);
      const client = createClient(fetch);

      await client.tags.list();

      const callArgs = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      expect(callArgs[1].body).toBeUndefined();
    });

    it("returns parsed JSON on success", async () => {
      const expected = { id: "org-1", name: "My Org", slug: "my-org", plan: "free" };
      const fetch = mockFetch(200, expected);
      const client = createClient(fetch);

      const result = await client.organizations.get("my-org");
      expect(result).toEqual(expected);
    });

    it("returns undefined for empty response body", async () => {
      const fetch = mockFetch(204, "");
      const client = createClient(fetch);

      const result = await client.tags.delete("tag-1");
      expect(result).toBeUndefined();
    });
  });

  describe("error handling", () => {
    it("throws ValidationError for 400", async () => {
      const fetch = mockFetch(400, "invalid input");
      const client = createClient(fetch);

      await expect(client.organizations.get("bad")).rejects.toThrow(ValidationError);
    });

    it("throws AuthenticationError for 401", async () => {
      const fetch = mockFetch(401, "unauthorized");
      const client = createClient(fetch);

      await expect(client.organizations.get("x")).rejects.toThrow(AuthenticationError);
    });

    it("throws NotFoundError for 404", async () => {
      const fetch = mockFetch(404, "not found");
      const client = createClient(fetch);

      await expect(client.organizations.get("missing")).rejects.toThrow(NotFoundError);
    });

    it("throws RateLimitError for 429", async () => {
      const fetch = mockFetch(429, "rate limited");
      const client = createClient(fetch);

      await expect(client.organizations.get("x")).rejects.toThrow(RateLimitError);
    });

    it("throws InternalServerError for 500", async () => {
      const fetch = mockFetch(500, "internal error");
      const client = createClient(fetch);

      await expect(client.organizations.get("x")).rejects.toThrow(InternalServerError);
    });

    it("throws QHubError for unknown status codes", async () => {
      const fetch = mockFetch(418, "i am a teapot");
      const client = createClient(fetch);

      await expect(client.organizations.get("x")).rejects.toThrow(QHubError);
    });

    it("includes status code and body in error", async () => {
      const fetch = mockFetch(404, "resource not found");
      const client = createClient(fetch);

      try {
        await client.organizations.get("missing");
        expect.fail("should have thrown");
      } catch (err) {
        expect(err).toBeInstanceOf(NotFoundError);
        const qErr = err as NotFoundError;
        expect(qErr.statusCode).toBe(404);
        expect(qErr.body).toBe("resource not found");
        expect(qErr.message).toContain("404");
      }
    });
  });
});

describe("OrganizationsResource", () => {
  it("creates an organization", async () => {
    const org = { id: "1", name: "Org", slug: "org", plan: "free" };
    const fetch = mockFetch(201, org);
    const client = createClient(fetch);

    const result = await client.organizations.create({ name: "Org", slug: "org" });
    expect(result).toEqual(org);
  });

  it("gets an organization by slug", async () => {
    const org = { id: "1", name: "Org", slug: "org", plan: "pro" };
    const fetch = mockFetch(200, org);
    const client = createClient(fetch);

    const result = await client.organizations.get("org");
    expect(result).toEqual(org);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/org",
      expect.any(Object),
    );
  });

  it("updates an organization", async () => {
    const org = { id: "1", name: "Updated", slug: "org", plan: "enterprise" };
    const fetch = mockFetch(200, org);
    const client = createClient(fetch);

    const result = await client.organizations.update("org", { plan: "enterprise" });
    expect(result).toEqual(org);
  });
});

describe("ProjectsResource", () => {
  it("lists projects for an organization", async () => {
    const projects = [{ id: "p1", organization_id: "o1", name: "Proj", slug: "proj", description: "" }];
    const fetch = mockFetch(200, projects);
    const client = createClient(fetch);

    const result = await client.projects.list("o1");
    expect(result).toEqual(projects);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/o1/projects",
      expect.any(Object),
    );
  });

  it("creates a project", async () => {
    const proj = { id: "p1", organization_id: "o1", name: "New", slug: "new", description: "desc" };
    const fetch = mockFetch(201, proj);
    const client = createClient(fetch);

    const result = await client.projects.create("o1", {
      organization_id: "o1",
      name: "New",
      slug: "new",
      description: "desc",
    });
    expect(result).toEqual(proj);
  });

  it("deletes a project", async () => {
    const fetch = mockFetch(204, "");
    const client = createClient(fetch);

    await client.projects.delete("o1", "proj");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/o1/projects/proj",
      expect.objectContaining({ method: "DELETE" }),
    );
  });
});

describe("PromptsResource", () => {
  it("lists prompts for a project", async () => {
    const prompts = [
      { id: "pr1", project_id: "p1", name: "P", slug: "p", prompt_type: "system", description: "", latest_version: 1, production_version: null },
    ];
    const fetch = mockFetch(200, prompts);
    const client = createClient(fetch);

    const result = await client.prompts.list("p1");
    expect(result).toEqual(prompts);
  });

  it("creates a prompt", async () => {
    const prompt = { id: "pr1", project_id: "p1", name: "New", slug: "new", prompt_type: "user", description: "", latest_version: 1, production_version: null };
    const fetch = mockFetch(201, prompt);
    const client = createClient(fetch);

    const result = await client.prompts.create("p1", {
      name: "New",
      slug: "new",
      prompt_type: "user",
    });
    expect(result).toEqual(prompt);
  });
});

describe("VersionsResource", () => {
  it("lists versions", async () => {
    const versions = [{ id: "v1", prompt_id: "pr1", version_number: 1, status: "draft", content: {}, variables: {}, change_description: "", author_id: "u1" }];
    const fetch = mockFetch(200, versions);
    const client = createClient(fetch);

    const result = await client.versions.list("pr1");
    expect(result).toEqual(versions);
  });

  it("gets a specific version", async () => {
    const version = { id: "v1", prompt_id: "pr1", version_number: 2, status: "production", content: { text: "hi" }, variables: [], change_description: "update", author_id: "u1" };
    const fetch = mockFetch(200, version);
    const client = createClient(fetch);

    const result = await client.versions.get("pr1", 2);
    expect(result).toEqual(version);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/versions/2",
      expect.any(Object),
    );
  });

  it("updates version status", async () => {
    const version = { id: "v1", prompt_id: "pr1", version_number: 1, status: "production", content: {}, variables: {}, change_description: "", author_id: "u1" };
    const fetch = mockFetch(200, version);
    const client = createClient(fetch);

    const result = await client.versions.updateStatus("pr1", 1, { status: "production" });
    expect(result.status).toBe("production");
  });

  it("gets lint result", async () => {
    const lint = { score: 85, issues: [], passed: ["excessive-length"] };
    const fetch = mockFetch(200, lint);
    const client = createClient(fetch);

    const result = await client.versions.lint("pr1", 1);
    expect(result.score).toBe(85);
  });

  it("gets text diff with default from version", async () => {
    const diff = { from_version: 1, to_version: 2, hunks: [], stats: { added: 1, removed: 0, equal: 5 } };
    const fetch = mockFetch(200, diff);
    const client = createClient(fetch);

    await client.versions.textDiff("pr1", 2);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/versions/2/text-diff",
      expect.any(Object),
    );
  });

  it("gets text diff with explicit from version", async () => {
    const diff = { from_version: 1, to_version: 3, hunks: [], stats: { added: 2, removed: 1, equal: 3 } };
    const fetch = mockFetch(200, diff);
    const client = createClient(fetch);

    await client.versions.textDiff("pr1", 3, 1);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/versions/3/text-diff?from=1",
      expect.any(Object),
    );
  });

  it("gets semantic diff between two versions", async () => {
    const diff = { summary: "Added context", changes: [], tone_shift: "neutral", specificity_change: 0.1 };
    const fetch = mockFetch(200, diff);
    const client = createClient(fetch);

    const result = await client.versions.semanticDiff("pr1", 1, 2);
    expect(result.summary).toBe("Added context");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/semantic-diff/1/2",
      expect.any(Object),
    );
  });
});

describe("LogsResource", () => {
  it("creates a log", async () => {
    const log = { id: "l1", org_id: "o1", prompt_id: "p1", version_number: 1, request_body: {}, response_body: {}, model: "gpt-4", provider: "openai", input_tokens: 10, output_tokens: 20, total_tokens: 30, latency_ms: 100, estimated_cost: "0.01", status: "success", error_message: "", environment: "production", metadata: null, executed_at: "2025-01-01T00:00:00Z", created_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, log);
    const client = createClient(fetch);

    const result = await client.logs.create({
      org_id: "o1",
      prompt_id: "p1",
      version_number: 1,
      request_body: {},
      model: "gpt-4",
      provider: "openai",
      estimated_cost: "0.01",
      status: "success",
      environment: "production",
      executed_at: "2025-01-01T00:00:00Z",
    });
    expect(result.id).toBe("l1");
  });

  it("creates a batch of logs", async () => {
    const logs = [{ id: "l1" }, { id: "l2" }];
    const fetch = mockFetch(201, logs);
    const client = createClient(fetch);

    await client.logs.createBatch([
      { org_id: "o1", prompt_id: "p1", version_number: 1, request_body: {}, model: "m", provider: "p", estimated_cost: "0", status: "success", environment: "development", executed_at: "2025-01-01T00:00:00Z" },
    ]);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/logs/batch",
      expect.objectContaining({
        method: "POST",
        body: expect.stringContaining('"logs"'),
      }),
    );
  });
});

describe("EvaluationsResource", () => {
  it("creates an evaluation", async () => {
    const evaluation = { id: "e1", execution_log_id: "l1", overall_score: "0.9", accuracy_score: null, relevance_score: null, fluency_score: null, safety_score: null, feedback: "good", evaluator_type: "human", evaluator_id: "u1", metadata: null, created_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, evaluation);
    const client = createClient(fetch);

    const result = await client.evaluations.create({
      execution_log_id: "l1",
      overall_score: "0.9",
      evaluator_type: "human",
      feedback: "good",
    });
    expect(result.id).toBe("e1");
  });

  it("lists evaluations by log ID", async () => {
    const evals = [{ id: "e1" }];
    const fetch = mockFetch(200, evals);
    const client = createClient(fetch);

    await client.evaluations.listByLog("l1");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/logs/l1/evaluations",
      expect.any(Object),
    );
  });
});

describe("ConsultingResource", () => {
  it("creates a session", async () => {
    const session = { id: "s1", org_id: "o1", title: "Help", industry_config_id: null, status: "active", created_at: "2025-01-01T00:00:00Z", updated_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, session);
    const client = createClient(fetch);

    const result = await client.consulting.createSession({ org_id: "o1", title: "Help" });
    expect(result.title).toBe("Help");
  });

  it("posts a message to a session", async () => {
    const msg = { id: "m1", session_id: "s1", role: "user", content: "Hello", citations: null, actions_taken: null, created_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, msg);
    const client = createClient(fetch);

    const result = await client.consulting.createMessage("s1", { role: "user", content: "Hello" });
    expect(result.content).toBe("Hello");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/consulting/sessions/s1/messages",
      expect.objectContaining({ method: "POST" }),
    );
  });
});

describe("TagsResource", () => {
  it("adds a tag to a prompt", async () => {
    const fetch = mockFetch(201, "");
    const client = createClient(fetch);

    await client.tags.addToPrompt("pr1", { tag_id: "t1" });
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/tags",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("removes a tag from a prompt", async () => {
    const fetch = mockFetch(204, "");
    const client = createClient(fetch);

    await client.tags.removeFromPrompt("pr1", "t1");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/tags/t1",
      expect.objectContaining({ method: "DELETE" }),
    );
  });
});

describe("IndustriesResource", () => {
  it("runs a compliance check", async () => {
    const result = { compliant: false, violations: [{ rule: "pii", message: "Contains PII" }] };
    const fetch = mockFetch(200, result);
    const client = createClient(fetch);

    const resp = await client.industries.complianceCheck("healthcare", { content: "patient John" });
    expect(resp.compliant).toBe(false);
    expect(resp.violations).toHaveLength(1);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/industries/healthcare/compliance-check",
      expect.objectContaining({ method: "POST" }),
    );
  });
});

describe("AnalyticsResource", () => {
  it("gets prompt analytics", async () => {
    const data = [{ version_number: 1, total_executions: 100, avg_tokens: 50, avg_latency_ms: 200, total_cost: "1.23", success_count: 95, error_count: 5 }];
    const fetch = mockFetch(200, data);
    const client = createClient(fetch);

    const result = await client.analytics.promptAnalytics("pr1");
    expect(result).toEqual(data);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/analytics",
      expect.any(Object),
    );
  });

  it("gets version analytics", async () => {
    const data = { prompt_id: "pr1", version_number: 1, total_executions: 50, avg_tokens: 30, avg_latency_ms: 150, total_cost: "0.50", avg_cost: "0.01", success_count: 48, error_count: 2 };
    const fetch = mockFetch(200, data);
    const client = createClient(fetch);

    const result = await client.analytics.versionAnalytics("pr1", 1);
    expect(result).toEqual(data);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/versions/1/analytics",
      expect.any(Object),
    );
  });

  it("gets daily trend", async () => {
    const data = [{ day: "2025-01-01", total_executions: 10, avg_tokens: 40, avg_latency_ms: 100, total_cost: "0.10" }];
    const fetch = mockFetch(200, data);
    const client = createClient(fetch);

    const result = await client.analytics.dailyTrend("pr1");
    expect(result).toEqual(data);
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/prompts/pr1/trend",
      expect.any(Object),
    );
  });
});

describe("SearchResource", () => {
  it("performs semantic search", async () => {
    const resp = { query: "test", results: [], total: 0 };
    const fetch = mockFetch(200, resp);
    const client = createClient(fetch);

    const result = await client.search.semantic({ query: "test", org_id: "o1", limit: 5 });
    expect(result.query).toBe("test");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/search/semantic",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("checks embedding status", async () => {
    const fetch = mockFetch(200, { embedding_service: "healthy" });
    const client = createClient(fetch);

    const result = await client.search.embeddingStatus();
    expect(result.embedding_service).toBe("healthy");
  });
});

describe("ApiKeysResource", () => {
  it("creates an API key", async () => {
    const key = { id: "k1", organization_id: "o1", name: "My Key", key: "qhub_abc123", key_prefix: "qhub_abc", expires_at: null, created_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, key);
    const client = createClient(fetch);

    const result = await client.apiKeys.create("o1", { organization_id: "o1", name: "My Key" });
    expect(result.key).toBe("qhub_abc123");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/o1/api-keys",
      expect.objectContaining({ method: "POST" }),
    );
  });
});

describe("MembersResource", () => {
  it("adds a member", async () => {
    const member = { organization_id: "o1", user_id: "u1", role: "member", joined_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(201, member);
    const client = createClient(fetch);

    const result = await client.members.add("o1", { user_id: "u1", role: "member" });
    expect(result.role).toBe("member");
  });

  it("updates a member role", async () => {
    const member = { organization_id: "o1", user_id: "u1", role: "admin", joined_at: "2025-01-01T00:00:00Z" };
    const fetch = mockFetch(200, member);
    const client = createClient(fetch);

    const result = await client.members.update("o1", "u1", { role: "admin" });
    expect(result.role).toBe("admin");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/o1/members/u1",
      expect.objectContaining({ method: "PUT" }),
    );
  });

  it("removes a member", async () => {
    const fetch = mockFetch(204, "");
    const client = createClient(fetch);

    await client.members.remove("o1", "u1");
    expect(fetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/organizations/o1/members/u1",
      expect.objectContaining({ method: "DELETE" }),
    );
  });
});
