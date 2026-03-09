// ─── Client Options ───

/** Options for configuring the QHubClient. */
export interface ClientOptions {
  /** Base URL of the QHub API (e.g. "http://localhost:8080"). */
  baseUrl: string;
  /** Bearer token for authentication. */
  bearerToken: string;
  /** Optional custom fetch implementation (defaults to global fetch). */
  fetch?: typeof globalThis.fetch;
}

// ─── Organizations ───

export interface Organization {
  id: string;
  name: string;
  slug: string;
  plan: string;
}

export interface CreateOrganizationRequest {
  name: string;
  slug: string;
}

export interface UpdateOrganizationRequest {
  name?: string;
  slug?: string;
  plan?: string;
}

// ─── Projects ───

export interface Project {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  description: string;
}

export interface CreateProjectRequest {
  organization_id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  slug?: string;
  description?: string;
}

// ─── Prompts ───

export interface Prompt {
  id: string;
  project_id: string;
  name: string;
  slug: string;
  prompt_type: string;
  description: string;
  latest_version: number;
  production_version: number | null;
}

export interface CreatePromptRequest {
  name: string;
  slug: string;
  prompt_type: "system" | "user" | "combined";
  description?: string;
}

export interface UpdatePromptRequest {
  name?: string;
  slug?: string;
  description?: string;
}

// ─── Prompt Versions ───

export interface PromptVersion {
  id: string;
  prompt_id: string;
  version_number: number;
  status: string;
  content: unknown;
  variables: unknown;
  change_description: string;
  author_id: string;
}

export interface CreateVersionRequest {
  content: unknown;
  variables?: unknown;
  change_description?: string;
  author_id: string;
}

export interface UpdateVersionStatusRequest {
  status: "draft" | "review" | "production" | "archived";
}

// ─── Execution Logs ───

export interface ExecutionLog {
  id: string;
  org_id: string;
  prompt_id: string;
  version_number: number;
  request_body: unknown;
  response_body: unknown;
  model: string;
  provider: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  latency_ms: number;
  estimated_cost: string;
  status: string;
  error_message: string;
  environment: string;
  metadata: unknown;
  executed_at: string;
  created_at: string;
}

export interface CreateLogRequest {
  org_id: string;
  prompt_id: string;
  version_number: number;
  request_body: unknown;
  response_body?: unknown;
  model: string;
  provider: string;
  input_tokens?: number;
  output_tokens?: number;
  total_tokens?: number;
  latency_ms?: number;
  estimated_cost: string;
  status: "success" | "error";
  error_message?: string;
  environment: "development" | "staging" | "production";
  metadata?: unknown;
  executed_at: string;
}

export interface ListLogsResponse {
  data: ExecutionLog[];
  total: number;
}

// ─── Evaluations ───

export interface Evaluation {
  id: string;
  execution_log_id: string;
  overall_score: string | null;
  accuracy_score: string | null;
  relevance_score: string | null;
  fluency_score: string | null;
  safety_score: string | null;
  feedback: string;
  evaluator_type: string;
  evaluator_id: string;
  metadata: unknown;
  created_at: string;
}

export interface CreateEvaluationRequest {
  execution_log_id: string;
  overall_score?: string | null;
  accuracy_score?: string | null;
  relevance_score?: string | null;
  fluency_score?: string | null;
  safety_score?: string | null;
  feedback?: string;
  evaluator_type: "human" | "auto";
  evaluator_id?: string;
  metadata?: unknown;
}

// ─── Consulting ───

export interface ConsultingSession {
  id: string;
  org_id: string;
  title: string;
  industry_config_id: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSessionRequest {
  org_id: string;
  title: string;
  industry_config_id?: string;
}

export interface ConsultingMessage {
  id: string;
  session_id: string;
  role: string;
  content: string;
  citations: unknown;
  actions_taken: unknown;
  created_at: string;
}

export interface CreateMessageRequest {
  role: "user" | "assistant" | "system";
  content: string;
  citations?: unknown;
  actions_taken?: unknown;
}

// ─── Tags ───

export interface Tag {
  id: string;
  org_id: string;
  name: string;
  color: string;
  created_at: string;
}

export interface CreateTagRequest {
  org_id: string;
  name: string;
  color: string;
}

export interface AddPromptTagRequest {
  tag_id: string;
}

// ─── Industries ───

export interface IndustryConfig {
  id: string;
  slug: string;
  name: string;
  description: string;
  knowledge_base: unknown;
  compliance_rules: unknown;
  created_at: string;
  updated_at: string;
}

export interface CreateIndustryConfigRequest {
  slug: string;
  name: string;
  description?: string;
  knowledge_base?: unknown;
  compliance_rules?: unknown;
}

export interface UpdateIndustryConfigRequest {
  name?: string;
  description?: string;
  knowledge_base?: unknown;
  compliance_rules?: unknown;
}

export interface Benchmark {
  id: string;
  industry_config_id: string;
  period: string;
  avg_quality_score: string;
  avg_latency_ms: number;
  avg_cost_per_request: string;
  total_executions: number;
  p50_quality: string;
  p90_quality: string;
  opt_in_count: number;
  created_at: string;
}

export interface ComplianceCheckRequest {
  content: string;
}

export interface ComplianceCheckResponse {
  compliant: boolean;
  violations: ComplianceIssue[];
}

export interface ComplianceIssue {
  rule: string;
  message: string;
}

// ─── Intelligence (Diff / Lint) ───

export interface SemanticDiff {
  summary: string;
  changes: DiffChange[];
  tone_shift?: string;
  specificity_change: number;
}

export interface DiffChange {
  category: string;
  description: string;
  impact: string;
}

export interface LintResult {
  score: number;
  issues: LintIssue[];
  passed: string[];
}

export interface LintIssue {
  rule: string;
  severity: string;
  message: string;
  suggestion?: string;
}

export interface TextDiffResult {
  from_version: number;
  to_version: number;
  hunks: TextDiffHunk[];
  stats: TextDiffStats;
}

export interface TextDiffHunk {
  lines: TextDiffLine[];
}

export interface TextDiffLine {
  type: string;
  content: string;
  old_line?: number;
  new_line?: number;
}

export interface TextDiffStats {
  added: number;
  removed: number;
  equal: number;
}

// ─── Analytics ───

export interface PromptAnalytics {
  version_number: number;
  total_executions: number;
  avg_tokens: number;
  avg_latency_ms: number;
  total_cost: string;
  success_count: number;
  error_count: number;
}

export interface VersionAnalytics {
  prompt_id: string;
  version_number: number;
  total_executions: number;
  avg_tokens: number;
  avg_latency_ms: number;
  total_cost: string;
  avg_cost: string;
  success_count: number;
  error_count: number;
}

export interface ProjectAnalytics {
  prompt_id: string;
  prompt_name: string;
  total_executions: number;
  avg_tokens: number;
  avg_latency_ms: number;
  total_cost: string;
}

export interface DailyTrend {
  day: string;
  total_executions: number;
  avg_tokens: number;
  avg_latency_ms: number;
  total_cost: string;
}

// ─── Search ───

export interface SemanticSearchRequest {
  query: string;
  org_id: string;
  limit?: number;
  min_score?: number;
}

export interface SemanticSearchResponse {
  query: string;
  results: SearchResult[];
  total: number;
}

export interface SearchResult {
  id: string;
  prompt_id: string;
  prompt_name: string;
  prompt_slug: string;
  version_number: number;
  status: string;
  content: unknown;
  change_description: string;
  similarity: number;
  created_at: string;
}

export interface EmbeddingStatusResponse {
  embedding_service: string;
}

// ─── API Keys ───

export interface ApiKey {
  id: string;
  organization_id: string;
  name: string;
  key_prefix: string;
  last_used_at: string | null;
  expires_at: string | null;
  revoked_at: string | null;
  created_at: string;
}

export interface ApiKeyCreated {
  id: string;
  organization_id: string;
  name: string;
  key: string;
  key_prefix: string;
  expires_at: string | null;
  created_at: string;
}

export interface CreateApiKeyRequest {
  organization_id: string;
  name: string;
}

// ─── Members ───

export interface Member {
  organization_id: string;
  user_id: string;
  role: string;
  joined_at: string;
}

export interface AddMemberRequest {
  user_id: string;
  role: "owner" | "admin" | "member" | "viewer";
}

export interface UpdateMemberRequest {
  role: "owner" | "admin" | "member" | "viewer";
}
