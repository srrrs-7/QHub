import { describe, it, expect } from "vitest";
import type {
  Organization,
  Project,
  Prompt,
  PromptVersion,
  ExecutionLog,
  Evaluation,
  ConsultingSession,
  ConsultingMessage,
  Tag,
  IndustryConfig,
  Benchmark,
  SemanticDiff,
  LintResult,
  TextDiffResult,
  PromptAnalytics,
  VersionAnalytics,
  ProjectAnalytics,
  DailyTrend,
  SearchResult,
  ApiKey,
  ApiKeyCreated,
  Member,
} from "../src/types.js";

/**
 * These tests verify that the TypeScript interfaces are correctly shaped
 * by assigning concrete values and confirming they compile and hold expected data.
 */

describe("Type definitions", () => {
  it("Organization has correct shape", () => {
    const org: Organization = { id: "1", name: "Org", slug: "org", plan: "free" };
    expect(org.id).toBe("1");
    expect(org.plan).toBe("free");
  });

  it("Project has correct shape", () => {
    const proj: Project = { id: "1", organization_id: "o1", name: "P", slug: "p", description: "desc" };
    expect(proj.organization_id).toBe("o1");
  });

  it("Prompt supports nullable production_version", () => {
    const p1: Prompt = { id: "1", project_id: "p1", name: "P", slug: "p", prompt_type: "system", description: "", latest_version: 1, production_version: null };
    const p2: Prompt = { id: "2", project_id: "p1", name: "Q", slug: "q", prompt_type: "user", description: "", latest_version: 3, production_version: 2 };
    expect(p1.production_version).toBeNull();
    expect(p2.production_version).toBe(2);
  });

  it("PromptVersion has content and variables as unknown", () => {
    const v: PromptVersion = { id: "1", prompt_id: "p1", version_number: 1, status: "draft", content: { text: "hello" }, variables: ["name"], change_description: "", author_id: "u1" };
    expect(v.content).toEqual({ text: "hello" });
    expect(v.variables).toEqual(["name"]);
  });

  it("ExecutionLog has all token fields", () => {
    const log: ExecutionLog = {
      id: "1", org_id: "o1", prompt_id: "p1", version_number: 1,
      request_body: {}, response_body: {}, model: "gpt-4", provider: "openai",
      input_tokens: 10, output_tokens: 20, total_tokens: 30,
      latency_ms: 100, estimated_cost: "0.01", status: "success",
      error_message: "", environment: "production",
      metadata: null, executed_at: "2025-01-01T00:00:00Z", created_at: "2025-01-01T00:00:00Z",
    };
    expect(log.total_tokens).toBe(30);
  });

  it("Evaluation supports nullable score fields", () => {
    const e: Evaluation = {
      id: "1", execution_log_id: "l1",
      overall_score: "0.9", accuracy_score: null, relevance_score: null,
      fluency_score: null, safety_score: null,
      feedback: "good", evaluator_type: "human", evaluator_id: "u1",
      metadata: null, created_at: "2025-01-01T00:00:00Z",
    };
    expect(e.overall_score).toBe("0.9");
    expect(e.accuracy_score).toBeNull();
  });

  it("ConsultingSession supports nullable industry_config_id", () => {
    const s: ConsultingSession = { id: "1", org_id: "o1", title: "Help", industry_config_id: null, status: "active", created_at: "", updated_at: "" };
    expect(s.industry_config_id).toBeNull();
  });

  it("Tag has expected fields", () => {
    const t: Tag = { id: "1", org_id: "o1", name: "important", color: "#ff0000", created_at: "" };
    expect(t.color).toBe("#ff0000");
  });

  it("SemanticDiff has optional tone_shift", () => {
    const d1: SemanticDiff = { summary: "s", changes: [], specificity_change: 0 };
    const d2: SemanticDiff = { summary: "s", changes: [], tone_shift: "casual to formal", specificity_change: 0.5 };
    expect(d1.tone_shift).toBeUndefined();
    expect(d2.tone_shift).toBe("casual to formal");
  });

  it("LintResult has score, issues, and passed", () => {
    const r: LintResult = { score: 85, issues: [{ rule: "vague-instructions", severity: "warning", message: "Too vague" }], passed: ["excessive-length"] };
    expect(r.score).toBe(85);
    expect(r.issues).toHaveLength(1);
  });

  it("TextDiffResult has hunks and stats", () => {
    const d: TextDiffResult = {
      from_version: 1, to_version: 2,
      hunks: [{ lines: [{ type: "added", content: "new line", new_line: 1 }] }],
      stats: { added: 1, removed: 0, equal: 5 },
    };
    expect(d.stats.added).toBe(1);
  });

  it("Analytics types have numeric fields", () => {
    const pa: PromptAnalytics = { version_number: 1, total_executions: 100, avg_tokens: 50, avg_latency_ms: 200, total_cost: "1.23", success_count: 95, error_count: 5 };
    const va: VersionAnalytics = { prompt_id: "p1", version_number: 1, total_executions: 50, avg_tokens: 30, avg_latency_ms: 150, total_cost: "0.50", avg_cost: "0.01", success_count: 48, error_count: 2 };
    const pra: ProjectAnalytics = { prompt_id: "p1", prompt_name: "P", total_executions: 200, avg_tokens: 40, avg_latency_ms: 180, total_cost: "2.00" };
    const dt: DailyTrend = { day: "2025-01-01", total_executions: 10, avg_tokens: 40, avg_latency_ms: 100, total_cost: "0.10" };
    expect(pa.total_executions).toBe(100);
    expect(va.avg_cost).toBe("0.01");
    expect(pra.prompt_name).toBe("P");
    expect(dt.day).toBe("2025-01-01");
  });

  it("SearchResult has similarity score", () => {
    const r: SearchResult = { id: "1", prompt_id: "p1", prompt_name: "P", prompt_slug: "p", version_number: 1, status: "production", content: {}, change_description: "", similarity: 0.95, created_at: "" };
    expect(r.similarity).toBe(0.95);
  });

  it("ApiKey has nullable date fields", () => {
    const k: ApiKey = { id: "1", organization_id: "o1", name: "Key", key_prefix: "qhub_", last_used_at: null, expires_at: null, revoked_at: null, created_at: "" };
    expect(k.last_used_at).toBeNull();
  });

  it("ApiKeyCreated includes the raw key", () => {
    const k: ApiKeyCreated = { id: "1", organization_id: "o1", name: "Key", key: "qhub_abc123xyz", key_prefix: "qhub_abc", expires_at: null, created_at: "" };
    expect(k.key).toContain("qhub_");
  });

  it("Member has role and joined_at", () => {
    const m: Member = { organization_id: "o1", user_id: "u1", role: "admin", joined_at: "2025-01-01T00:00:00Z" };
    expect(m.role).toBe("admin");
  });
});
