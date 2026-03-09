"""Tests for resource classes with mocked HTTP responses."""

from __future__ import annotations

import httpx
import pytest
import respx

from qhub.client import QHubClient
from qhub.types import CreateEvaluationRequest, CreateLogRequest

BASE_URL = "http://localhost:8080"
TOKEN = "test-token"


# ---------------------------------------------------------------------------
# Organizations
# ---------------------------------------------------------------------------


class TestOrganizations:
    @respx.mock(base_url=BASE_URL)
    def test_get(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/organizations/acme").mock(
            return_value=httpx.Response(
                200, json={"id": "o1", "name": "Acme", "slug": "acme", "plan": "pro"}
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        org = c.organizations.get("acme")
        assert org.slug == "acme"

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/organizations").mock(
            return_value=httpx.Response(
                201,
                json={"id": "o1", "name": "Acme", "slug": "acme", "plan": "free"},
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        org = c.organizations.create(name="Acme", slug="acme")
        assert org.plan == "free"

    @respx.mock(base_url=BASE_URL)
    def test_update(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.put("/api/v1/organizations/acme").mock(
            return_value=httpx.Response(
                200,
                json={"id": "o1", "name": "Acme Inc", "slug": "acme", "plan": "pro"},
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        org = c.organizations.update("acme", name="Acme Inc")
        assert org.name == "Acme Inc"


# ---------------------------------------------------------------------------
# Projects
# ---------------------------------------------------------------------------


class TestProjects:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/organizations/o1/projects").mock(
            return_value=httpx.Response(
                200,
                json=[
                    {
                        "id": "p1",
                        "organization_id": "o1",
                        "name": "Proj",
                        "slug": "proj",
                        "description": "",
                    }
                ],
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        projects = c.projects.list("o1")
        assert len(projects) == 1
        assert projects[0].slug == "proj"

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/organizations/o1/projects").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "p1",
                    "organization_id": "o1",
                    "name": "Proj",
                    "slug": "proj",
                    "description": "A project",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        proj = c.projects.create("o1", name="Proj", slug="proj", description="A project")
        assert proj.description == "A project"

    @respx.mock(base_url=BASE_URL)
    def test_delete(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.delete("/api/v1/organizations/o1/projects/proj").mock(
            return_value=httpx.Response(204)
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        c.projects.delete("o1", "proj")


# ---------------------------------------------------------------------------
# Prompts
# ---------------------------------------------------------------------------


class TestPrompts:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/projects/p1/prompts").mock(
            return_value=httpx.Response(
                200,
                json=[
                    {
                        "id": "pr1",
                        "project_id": "p1",
                        "name": "Greeting",
                        "slug": "greeting",
                        "prompt_type": "system",
                        "description": "",
                        "latest_version": 1,
                        "production_version": None,
                    }
                ],
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        prompts = c.prompts.list("p1")
        assert len(prompts) == 1

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/projects/p1/prompts").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "pr1",
                    "project_id": "p1",
                    "name": "Greeting",
                    "slug": "greeting",
                    "prompt_type": "system",
                    "description": "",
                    "latest_version": 0,
                    "production_version": None,
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        prompt = c.prompts.create("p1", name="Greeting", slug="greeting", prompt_type="system")
        assert prompt.prompt_type == "system"


# ---------------------------------------------------------------------------
# Versions
# ---------------------------------------------------------------------------


class TestVersions:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/versions").mock(
            return_value=httpx.Response(
                200,
                json=[
                    {
                        "id": "v1",
                        "prompt_id": "pr1",
                        "version_number": 1,
                        "status": "draft",
                        "content": {},
                        "variables": None,
                        "change_description": "",
                        "author_id": "u1",
                    }
                ],
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        versions = c.versions.list("pr1")
        assert len(versions) == 1

    @respx.mock(base_url=BASE_URL)
    def test_get(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/versions/1").mock(
            return_value=httpx.Response(
                200,
                json={
                    "id": "v1",
                    "prompt_id": "pr1",
                    "version_number": 1,
                    "status": "draft",
                    "content": {"text": "Hello!"},
                    "variables": None,
                    "change_description": "",
                    "author_id": "u1",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        ver = c.versions.get("pr1", 1)
        assert ver.content == {"text": "Hello!"}

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/prompts/pr1/versions").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "v2",
                    "prompt_id": "pr1",
                    "version_number": 2,
                    "status": "draft",
                    "content": {"text": "Hi!"},
                    "variables": None,
                    "change_description": "Shortened greeting",
                    "author_id": "u1",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        ver = c.versions.create(
            "pr1",
            content={"text": "Hi!"},
            author_id="u1",
            change_description="Shortened greeting",
        )
        assert ver.version_number == 2

    @respx.mock(base_url=BASE_URL)
    def test_update_status(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.put("/api/v1/prompts/pr1/versions/1/status").mock(
            return_value=httpx.Response(
                200,
                json={
                    "id": "v1",
                    "prompt_id": "pr1",
                    "version_number": 1,
                    "status": "production",
                    "content": {},
                    "variables": None,
                    "change_description": "",
                    "author_id": "u1",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        ver = c.versions.update_status("pr1", 1, status="production")
        assert ver.status == "production"

    @respx.mock(base_url=BASE_URL)
    def test_lint(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/versions/1/lint").mock(
            return_value=httpx.Response(200, json={"score": 85, "issues": []})
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.versions.lint("pr1", 1)
        assert result.score == 85  # type: ignore[attr-defined]

    @respx.mock(base_url=BASE_URL)
    def test_semantic_diff(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/semantic-diff/1/2").mock(
            return_value=httpx.Response(
                200, json={"changes": [{"type": "tone", "detail": "more formal"}]}
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        diff = c.versions.semantic_diff("pr1", 1, 2)
        assert diff.changes == [{"type": "tone", "detail": "more formal"}]  # type: ignore[attr-defined]


# ---------------------------------------------------------------------------
# Logs
# ---------------------------------------------------------------------------


class TestLogs:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/logs").mock(
            return_value=httpx.Response(200, json={"data": [], "total": 0})
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.logs.list()
        assert result.total == 0

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/logs").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "log1",
                    "org_id": "o1",
                    "prompt_id": "pr1",
                    "version_number": 1,
                    "request_body": {},
                    "response_body": {},
                    "model": "gpt-4",
                    "provider": "openai",
                    "input_tokens": 10,
                    "output_tokens": 20,
                    "total_tokens": 30,
                    "latency_ms": 100,
                    "estimated_cost": "0.001",
                    "status": "success",
                    "error_message": "",
                    "environment": "production",
                    "metadata": None,
                    "executed_at": "2025-01-01T00:00:00Z",
                    "created_at": "2025-01-01T00:00:00Z",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        req = CreateLogRequest(
            org_id="o1",
            prompt_id="pr1",
            version_number=1,
            request_body={},
            model="gpt-4",
            provider="openai",
            estimated_cost="0.001",
            status="success",
            environment="production",
            executed_at="2025-01-01T00:00:00Z",
        )
        log = c.logs.create(req)
        assert log.id == "log1"

    @respx.mock(base_url=BASE_URL)
    def test_create_batch(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/logs/batch").mock(
            return_value=httpx.Response(201, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.logs.create_batch([])
        assert result == []


# ---------------------------------------------------------------------------
# Evaluations
# ---------------------------------------------------------------------------


class TestEvaluations:
    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/evaluations").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "ev1",
                    "execution_log_id": "log1",
                    "overall_score": "0.9",
                    "accuracy_score": None,
                    "relevance_score": None,
                    "fluency_score": None,
                    "safety_score": None,
                    "feedback": "Great",
                    "evaluator_type": "human",
                    "evaluator_id": "u1",
                    "metadata": None,
                    "created_at": "2025-01-01T00:00:00Z",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        req = CreateEvaluationRequest(
            execution_log_id="log1",
            overall_score="0.9",
            feedback="Great",
            evaluator_type="human",
            evaluator_id="u1",
        )
        ev = c.evaluations.create(req)
        assert ev.overall_score == "0.9"

    @respx.mock(base_url=BASE_URL)
    def test_list_by_log(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/logs/log1/evaluations").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.evaluations.list_by_log("log1")
        assert result == []


# ---------------------------------------------------------------------------
# Consulting
# ---------------------------------------------------------------------------


class TestConsulting:
    @respx.mock(base_url=BASE_URL)
    def test_list_sessions(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/consulting/sessions").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        sessions = c.consulting.list_sessions()
        assert sessions == []

    @respx.mock(base_url=BASE_URL)
    def test_create_session(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/consulting/sessions").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "s1",
                    "org_id": "o1",
                    "title": "Help",
                    "industry_config_id": None,
                    "status": "active",
                    "created_at": "2025-01-01T00:00:00Z",
                    "updated_at": "2025-01-01T00:00:00Z",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        session = c.consulting.create_session(org_id="o1", title="Help")
        assert session.title == "Help"

    @respx.mock(base_url=BASE_URL)
    def test_create_message(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/consulting/sessions/s1/messages").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "m1",
                    "session_id": "s1",
                    "role": "user",
                    "content": "Hello",
                    "citations": None,
                    "actions_taken": None,
                    "created_at": "2025-01-01T00:00:00Z",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        msg = c.consulting.create_message("s1", role="user", content="Hello")
        assert msg.content == "Hello"


# ---------------------------------------------------------------------------
# Tags
# ---------------------------------------------------------------------------


class TestTags:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        tags = c.tags.list()
        assert tags == []

    @respx.mock(base_url=BASE_URL)
    def test_create(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/tags").mock(
            return_value=httpx.Response(
                201,
                json={
                    "id": "t1",
                    "org_id": "o1",
                    "name": "important",
                    "color": "#ff0000",
                    "created_at": "2025-01-01T00:00:00Z",
                },
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        tag = c.tags.create(org_id="o1", name="important", color="#ff0000")
        assert tag.name == "important"

    @respx.mock(base_url=BASE_URL)
    def test_add_to_prompt(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/prompts/pr1/tags").mock(
            return_value=httpx.Response(201, json={})
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        c.tags.add_to_prompt("pr1", tag_id="t1")

    @respx.mock(base_url=BASE_URL)
    def test_remove_from_prompt(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.delete("/api/v1/prompts/pr1/tags/t1").mock(
            return_value=httpx.Response(204)
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        c.tags.remove_from_prompt("pr1", tag_id="t1")


# ---------------------------------------------------------------------------
# Industries
# ---------------------------------------------------------------------------


class TestIndustries:
    @respx.mock(base_url=BASE_URL)
    def test_list(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/industries").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        industries = c.industries.list()
        assert industries == []

    @respx.mock(base_url=BASE_URL)
    def test_compliance_check(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/industries/healthcare/compliance-check").mock(
            return_value=httpx.Response(
                200, json={"compliant": True, "violations": []}
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.industries.compliance_check("healthcare", content="Test prompt")
        assert result.compliant is True

    @respx.mock(base_url=BASE_URL)
    def test_list_benchmarks(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/industries/healthcare/benchmarks").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        benchmarks = c.industries.list_benchmarks("healthcare")
        assert benchmarks == []


# ---------------------------------------------------------------------------
# Search
# ---------------------------------------------------------------------------


class TestSearch:
    @respx.mock(base_url=BASE_URL)
    def test_semantic(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.post("/api/v1/search/semantic").mock(
            return_value=httpx.Response(
                200, json={"query": "hello", "results": [], "total": 0}
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.search.semantic(query="hello", org_id="o1")
        assert result.total == 0

    @respx.mock(base_url=BASE_URL)
    def test_embedding_status(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/search/embedding-status").mock(
            return_value=httpx.Response(
                200, json={"embedding_service": "healthy"}
            )
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        status = c.search.embedding_status()
        assert status["embedding_service"] == "healthy"


# ---------------------------------------------------------------------------
# Analytics
# ---------------------------------------------------------------------------


class TestAnalytics:
    @respx.mock(base_url=BASE_URL)
    def test_prompt_analytics(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/analytics").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.analytics.prompt("pr1")
        assert result == []

    @respx.mock(base_url=BASE_URL)
    def test_daily_trend(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/prompts/pr1/trend").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.analytics.daily_trend("pr1")
        assert result == []
