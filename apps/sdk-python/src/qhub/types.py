"""Pydantic models for QHub API request and response types."""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Organizations
# ---------------------------------------------------------------------------


class Organization(BaseModel):
    """An organization."""

    id: str
    name: str
    slug: str
    plan: str


class CreateOrganizationRequest(BaseModel):
    """Request body for creating an organization."""

    name: str
    slug: str


class UpdateOrganizationRequest(BaseModel):
    """Request body for updating an organization."""

    name: str | None = None
    slug: str | None = None
    plan: str | None = None


# ---------------------------------------------------------------------------
# Projects
# ---------------------------------------------------------------------------


class Project(BaseModel):
    """A project within an organization."""

    id: str
    organization_id: str
    name: str
    slug: str
    description: str


class CreateProjectRequest(BaseModel):
    """Request body for creating a project."""

    organization_id: str
    name: str
    slug: str
    description: str = ""


class UpdateProjectRequest(BaseModel):
    """Request body for updating a project."""

    name: str | None = None
    slug: str | None = None
    description: str | None = None


# ---------------------------------------------------------------------------
# Prompts
# ---------------------------------------------------------------------------


class Prompt(BaseModel):
    """A prompt definition."""

    id: str
    project_id: str
    name: str
    slug: str
    prompt_type: str
    description: str
    latest_version: int
    production_version: int | None = None


class CreatePromptRequest(BaseModel):
    """Request body for creating a prompt."""

    name: str
    slug: str
    prompt_type: str
    description: str = ""


class UpdatePromptRequest(BaseModel):
    """Request body for updating a prompt."""

    name: str | None = None
    slug: str | None = None
    description: str | None = None


# ---------------------------------------------------------------------------
# Prompt Versions
# ---------------------------------------------------------------------------


class PromptVersion(BaseModel):
    """A specific version of a prompt."""

    id: str
    prompt_id: str
    version_number: int
    status: str
    content: Any = None
    variables: Any = None
    change_description: str = ""
    author_id: str = ""


class CreateVersionRequest(BaseModel):
    """Request body for creating a prompt version."""

    content: Any
    variables: Any = None
    change_description: str = ""
    author_id: str


class UpdateVersionStatusRequest(BaseModel):
    """Request body for updating a version's status."""

    status: str


# ---------------------------------------------------------------------------
# Execution Logs
# ---------------------------------------------------------------------------


class ExecutionLog(BaseModel):
    """A prompt execution log entry."""

    id: str = ""
    org_id: str
    prompt_id: str
    version_number: int
    request_body: Any
    response_body: Any = None
    model: str
    provider: str
    input_tokens: int = 0
    output_tokens: int = 0
    total_tokens: int = 0
    latency_ms: int = 0
    estimated_cost: str
    status: str
    error_message: str = ""
    environment: str
    metadata: Any = None
    executed_at: str
    created_at: str = ""


class CreateLogRequest(BaseModel):
    """Request body for creating an execution log."""

    org_id: str
    prompt_id: str
    version_number: int
    request_body: Any
    response_body: Any = None
    model: str
    provider: str
    input_tokens: int = 0
    output_tokens: int = 0
    total_tokens: int = 0
    latency_ms: int = 0
    estimated_cost: str
    status: str
    error_message: str = ""
    environment: str
    metadata: Any = None
    executed_at: str


class LogListResponse(BaseModel):
    """Paginated list of execution logs."""

    data: list[ExecutionLog]
    total: int


# ---------------------------------------------------------------------------
# Evaluations
# ---------------------------------------------------------------------------


class Evaluation(BaseModel):
    """An evaluation of a prompt execution."""

    id: str = ""
    execution_log_id: str
    overall_score: str | None = None
    accuracy_score: str | None = None
    relevance_score: str | None = None
    fluency_score: str | None = None
    safety_score: str | None = None
    feedback: str = ""
    evaluator_type: str
    evaluator_id: str = ""
    metadata: Any = None
    created_at: str = ""


class CreateEvaluationRequest(BaseModel):
    """Request body for creating an evaluation."""

    execution_log_id: str
    overall_score: str | None = None
    accuracy_score: str | None = None
    relevance_score: str | None = None
    fluency_score: str | None = None
    safety_score: str | None = None
    feedback: str = ""
    evaluator_type: str
    evaluator_id: str = ""
    metadata: Any = None


# ---------------------------------------------------------------------------
# Consulting
# ---------------------------------------------------------------------------


class ConsultingSession(BaseModel):
    """A consulting session."""

    id: str
    org_id: str
    title: str
    industry_config_id: str | None = None
    status: str
    created_at: str
    updated_at: str


class CreateSessionRequest(BaseModel):
    """Request body for creating a consulting session."""

    org_id: str
    title: str
    industry_config_id: str = ""


class ConsultingMessage(BaseModel):
    """A message within a consulting session."""

    id: str
    session_id: str
    role: str
    content: str
    citations: Any = None
    actions_taken: Any = None
    created_at: str


class CreateMessageRequest(BaseModel):
    """Request body for creating a consulting message."""

    role: str
    content: str
    citations: Any = None
    actions_taken: Any = None


# ---------------------------------------------------------------------------
# Tags
# ---------------------------------------------------------------------------


class Tag(BaseModel):
    """A tag for organizing prompts."""

    id: str
    org_id: str
    name: str
    color: str
    created_at: str


class CreateTagRequest(BaseModel):
    """Request body for creating a tag."""

    org_id: str
    name: str
    color: str


class AddPromptTagRequest(BaseModel):
    """Request body for adding a tag to a prompt."""

    tag_id: str


# ---------------------------------------------------------------------------
# Industries
# ---------------------------------------------------------------------------


class IndustryConfig(BaseModel):
    """An industry configuration."""

    id: str
    slug: str
    name: str
    description: str
    knowledge_base: Any = None
    compliance_rules: Any = None
    created_at: str
    updated_at: str


class CreateIndustryConfigRequest(BaseModel):
    """Request body for creating an industry config."""

    slug: str
    name: str
    description: str = ""
    knowledge_base: Any = None
    compliance_rules: Any = None


class UpdateIndustryConfigRequest(BaseModel):
    """Request body for updating an industry config."""

    name: str | None = None
    description: str | None = None
    knowledge_base: Any = None
    compliance_rules: Any = None


class Benchmark(BaseModel):
    """Industry benchmark data."""

    id: str
    industry_config_id: str
    period: str
    avg_quality_score: str
    avg_latency_ms: int
    avg_cost_per_request: str
    total_executions: int
    p50_quality: str
    p90_quality: str
    opt_in_count: int
    created_at: str


class ComplianceCheckRequest(BaseModel):
    """Request body for checking compliance."""

    content: str


class ComplianceIssue(BaseModel):
    """A single compliance violation."""

    rule: str
    message: str


class ComplianceCheckResponse(BaseModel):
    """Result of a compliance check."""

    compliant: bool
    violations: list[ComplianceIssue]


# ---------------------------------------------------------------------------
# Search
# ---------------------------------------------------------------------------


class SemanticSearchRequest(BaseModel):
    """Request body for semantic search."""

    query: str
    org_id: str
    limit: int = 10
    min_score: float = 0.0


class SearchResult(BaseModel):
    """A single semantic search result."""

    id: str
    prompt_id: str
    prompt_name: str
    prompt_slug: str
    version_number: int
    status: str
    content: Any = None
    change_description: str = ""
    similarity: float
    created_at: str


class SearchResponse(BaseModel):
    """Semantic search response."""

    query: str
    results: list[SearchResult]
    total: int


# ---------------------------------------------------------------------------
# Analytics
# ---------------------------------------------------------------------------


class PromptAnalytics(BaseModel):
    """Analytics for a prompt across all versions."""

    version_number: int
    total_executions: int
    avg_tokens: int
    avg_latency_ms: int
    total_cost: str
    success_count: int
    error_count: int


class VersionAnalytics(BaseModel):
    """Analytics for a specific prompt version."""

    prompt_id: str
    version_number: int
    total_executions: int
    avg_tokens: int
    avg_latency_ms: int
    total_cost: str
    avg_cost: str
    success_count: int
    error_count: int


class ProjectAnalytics(BaseModel):
    """Analytics for a project."""

    prompt_id: str
    prompt_name: str
    total_executions: int
    avg_tokens: int
    avg_latency_ms: int
    total_cost: str


class DailyTrend(BaseModel):
    """Daily trend data point."""

    day: str
    total_executions: int
    avg_tokens: int
    avg_latency_ms: int
    total_cost: str


# ---------------------------------------------------------------------------
# Diff / Lint
# ---------------------------------------------------------------------------


class SemanticDiff(BaseModel):
    """Result of a semantic diff between two prompt versions."""

    model_config = {"extra": "allow"}


class LintResult(BaseModel):
    """Result of linting a prompt version."""

    model_config = {"extra": "allow"}


class TextDiff(BaseModel):
    """Result of a text diff for a prompt version."""

    model_config = {"extra": "allow"}
