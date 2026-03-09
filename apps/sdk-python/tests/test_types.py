"""Tests for Pydantic model serialization and deserialization."""

from __future__ import annotations

import pytest

from qhub.types import (
    ComplianceCheckResponse,
    ComplianceIssue,
    ConsultingMessage,
    ConsultingSession,
    Evaluation,
    ExecutionLog,
    IndustryConfig,
    LogListResponse,
    Organization,
    Project,
    Prompt,
    PromptVersion,
    SearchResponse,
    SearchResult,
    Tag,
)


class TestOrganization:
    def test_roundtrip(self) -> None:
        data = {"id": "abc", "name": "Acme", "slug": "acme", "plan": "pro"}
        org = Organization.model_validate(data)
        assert org.id == "abc"
        assert org.name == "Acme"
        assert org.slug == "acme"
        assert org.plan == "pro"
        assert org.model_dump() == data


class TestProject:
    def test_roundtrip(self) -> None:
        data = {
            "id": "p1",
            "organization_id": "o1",
            "name": "My Project",
            "slug": "my-project",
            "description": "A test project",
        }
        proj = Project.model_validate(data)
        assert proj.slug == "my-project"
        assert proj.model_dump() == data


class TestPrompt:
    def test_with_production_version(self) -> None:
        data = {
            "id": "pr1",
            "project_id": "p1",
            "name": "Greeting",
            "slug": "greeting",
            "prompt_type": "system",
            "description": "",
            "latest_version": 3,
            "production_version": 2,
        }
        prompt = Prompt.model_validate(data)
        assert prompt.production_version == 2

    def test_without_production_version(self) -> None:
        data = {
            "id": "pr1",
            "project_id": "p1",
            "name": "Greeting",
            "slug": "greeting",
            "prompt_type": "system",
            "description": "",
            "latest_version": 1,
            "production_version": None,
        }
        prompt = Prompt.model_validate(data)
        assert prompt.production_version is None


class TestPromptVersion:
    def test_with_json_content(self) -> None:
        data = {
            "id": "v1",
            "prompt_id": "pr1",
            "version_number": 1,
            "status": "draft",
            "content": {"text": "Hello, {{name}}!"},
            "variables": [{"name": "name", "type": "string"}],
            "change_description": "Initial version",
            "author_id": "u1",
        }
        ver = PromptVersion.model_validate(data)
        assert ver.content == {"text": "Hello, {{name}}!"}
        assert ver.variables == [{"name": "name", "type": "string"}]


class TestExecutionLog:
    def test_full_roundtrip(self) -> None:
        data = {
            "id": "log1",
            "org_id": "o1",
            "prompt_id": "pr1",
            "version_number": 1,
            "request_body": {"prompt": "hello"},
            "response_body": {"text": "world"},
            "model": "gpt-4",
            "provider": "openai",
            "input_tokens": 10,
            "output_tokens": 20,
            "total_tokens": 30,
            "latency_ms": 150,
            "estimated_cost": "0.001",
            "status": "success",
            "error_message": "",
            "environment": "production",
            "metadata": None,
            "executed_at": "2025-01-01T00:00:00Z",
            "created_at": "2025-01-01T00:00:01Z",
        }
        log = ExecutionLog.model_validate(data)
        assert log.total_tokens == 30
        assert log.model == "gpt-4"


class TestLogListResponse:
    def test_pagination(self) -> None:
        data = {
            "data": [
                {
                    "id": "log1",
                    "org_id": "o1",
                    "prompt_id": "pr1",
                    "version_number": 1,
                    "request_body": {},
                    "response_body": {},
                    "model": "gpt-4",
                    "provider": "openai",
                    "input_tokens": 0,
                    "output_tokens": 0,
                    "total_tokens": 0,
                    "latency_ms": 0,
                    "estimated_cost": "0",
                    "status": "success",
                    "error_message": "",
                    "environment": "production",
                    "metadata": None,
                    "executed_at": "2025-01-01T00:00:00Z",
                    "created_at": "2025-01-01T00:00:01Z",
                }
            ],
            "total": 42,
        }
        resp = LogListResponse.model_validate(data)
        assert resp.total == 42
        assert len(resp.data) == 1


class TestEvaluation:
    def test_nullable_scores(self) -> None:
        data = {
            "id": "ev1",
            "execution_log_id": "log1",
            "overall_score": "0.95",
            "accuracy_score": None,
            "relevance_score": None,
            "fluency_score": None,
            "safety_score": None,
            "feedback": "Good",
            "evaluator_type": "human",
            "evaluator_id": "u1",
            "metadata": None,
            "created_at": "2025-01-01T00:00:00Z",
        }
        ev = Evaluation.model_validate(data)
        assert ev.overall_score == "0.95"
        assert ev.accuracy_score is None


class TestConsultingSession:
    def test_with_industry_config(self) -> None:
        data = {
            "id": "s1",
            "org_id": "o1",
            "title": "Help me",
            "industry_config_id": "ic1",
            "status": "active",
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
        }
        session = ConsultingSession.model_validate(data)
        assert session.industry_config_id == "ic1"


class TestConsultingMessage:
    def test_basic(self) -> None:
        data = {
            "id": "m1",
            "session_id": "s1",
            "role": "user",
            "content": "Hello",
            "citations": None,
            "actions_taken": None,
            "created_at": "2025-01-01T00:00:00Z",
        }
        msg = ConsultingMessage.model_validate(data)
        assert msg.role == "user"


class TestTag:
    def test_basic(self) -> None:
        data = {
            "id": "t1",
            "org_id": "o1",
            "name": "important",
            "color": "#ff0000",
            "created_at": "2025-01-01T00:00:00Z",
        }
        tag = Tag.model_validate(data)
        assert tag.color == "#ff0000"


class TestIndustryConfig:
    def test_basic(self) -> None:
        data = {
            "id": "ic1",
            "slug": "healthcare",
            "name": "Healthcare",
            "description": "HIPAA-compliant",
            "knowledge_base": {"topics": ["hipaa"]},
            "compliance_rules": [{"rule": "no-phi"}],
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
        }
        ic = IndustryConfig.model_validate(data)
        assert ic.slug == "healthcare"


class TestComplianceCheckResponse:
    def test_compliant(self) -> None:
        data = {"compliant": True, "violations": []}
        resp = ComplianceCheckResponse.model_validate(data)
        assert resp.compliant is True
        assert resp.violations == []

    def test_non_compliant(self) -> None:
        data = {
            "compliant": False,
            "violations": [
                {"rule": "no-phi", "message": "Contains PHI references"}
            ],
        }
        resp = ComplianceCheckResponse.model_validate(data)
        assert not resp.compliant
        assert len(resp.violations) == 1
        assert isinstance(resp.violations[0], ComplianceIssue)


class TestSearchResponse:
    def test_with_results(self) -> None:
        data = {
            "query": "hello",
            "results": [
                {
                    "id": "v1",
                    "prompt_id": "pr1",
                    "prompt_name": "Greeting",
                    "prompt_slug": "greeting",
                    "version_number": 1,
                    "status": "production",
                    "content": {"text": "Hello!"},
                    "change_description": "",
                    "similarity": 0.95,
                    "created_at": "2025-01-01T00:00:00Z",
                }
            ],
            "total": 1,
        }
        resp = SearchResponse.model_validate(data)
        assert resp.total == 1
        assert isinstance(resp.results[0], SearchResult)
        assert resp.results[0].similarity == 0.95
