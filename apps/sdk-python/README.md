# QHub Python SDK

Python SDK for the QHub prompt/answer version management API.

## Features

- 🐍 **Python 3.10+** support
- 🔄 **Synchronous client** with httpx
- ✅ **Type-safe** with Pydantic v2
- 📦 **Zero external dependencies** (httpx + pydantic only)
- 🧪 **Fully tested** with pytest

## Installation

```bash
pip install qhub-sdk
```

For development:

```bash
pip install -e ".[dev]"
```

## Quick Start

```python
from qhub import QHubClient

# Initialize client
client = QHubClient(
    bearer_token="your-bearer-token",
    base_url="http://localhost:8080"  # optional, defaults to localhost:8080
)

# Create an organization
org = client.organizations.create(
    name="My Organization",
    slug="my-org"
)

# Create a project
project = client.projects.create(
    organization_id=org.id,
    name="AI Assistant",
    slug="ai-assistant"
)

# Create a prompt
prompt = client.prompts.create(
    project_id=project.id,
    name="System Prompt",
    slug="system-prompt",
    content={"text": "You are a helpful AI assistant."}
)

# Create a version
version = client.versions.create(
    prompt_id=prompt.id,
    content={"text": "You are a helpful AI assistant with expertise in Python."},
    change_summary="Added Python expertise"
)

# Log execution
log = client.logs.create(
    version_id=version.id,
    input_text="How do I use list comprehensions?",
    output_text="List comprehensions provide...",
    latency_ms=250,
    token_count=150
)
```

## Available Resources

The client provides access to the following resources:

| Resource | Description |
|----------|-------------|
| `client.organizations` | Organization management |
| `client.projects` | Project management |
| `client.prompts` | Prompt management |
| `client.versions` | Version management and diff |
| `client.logs` | Execution log ingestion |
| `client.evaluations` | Evaluation management |
| `client.tags` | Tag management |
| `client.consulting` | AI consulting sessions (RAG) |
| `client.search` | Semantic search |
| `client.analytics` | Analytics and metrics |
| `client.industries` | Industry configurations |

## Authentication

QHub API supports two authentication methods:

### Bearer Token (recommended)

```python
client = QHubClient(bearer_token="your-bearer-token")
```

### API Key (for log ingestion)

API keys are managed per organization and used for server-to-server log ingestion:

```python
# Create API key via organizations resource
api_key = client.organizations.create_api_key(
    organization_id=org_id,
    name="Production Logger"
)

# Use in log ingestion
log_client = QHubClient(bearer_token=api_key.key)
log_client.logs.create(...)
```

## Advanced Usage

### Version Comparison

```python
# Get semantic diff between versions
diff = client.versions.diff(
    prompt_id=prompt_id,
    v1=1,
    v2=2
)

print(f"Length change: {diff.length_change}")
print(f"Variable changes: {diff.variable_changes}")
print(f"Text diff:\n{diff.text_diff}")

# Compare version metrics (Welch's t-test)
comparison = client.versions.compare(
    prompt_id=prompt_id,
    v1=1,
    v2=2,
    metric="latency_ms"
)

print(f"p-value: {comparison.p_value}")
print(f"Statistically significant: {comparison.significant}")
```

### Lint Prompt

```python
lint_result = client.versions.lint(
    prompt_id=prompt_id,
    version=1
)

print(f"Score: {lint_result.score}/100")
for issue in lint_result.issues:
    print(f"- {issue.severity}: {issue.message}")
```

### Semantic Search

```python
results = client.search.semantic(
    query="prompt for customer support chatbot",
    limit=5
)

for result in results:
    print(f"{result.prompt_name} v{result.version} (score: {result.score})")
```

### Consulting (RAG)

```python
# Create session
session = client.consulting.create_session(
    project_id=project_id,
    title="Improve conversion rate"
)

# Send message and stream response
for chunk in client.consulting.stream_response(
    session_id=session.id,
    message="How can I improve my prompt for better responses?"
):
    print(chunk, end="", flush=True)
```

### Analytics

```python
# Project analytics
stats = client.analytics.get_project_stats(project_id)
print(f"Total prompts: {stats.total_prompts}")
print(f"Total logs: {stats.total_logs}")

# Prompt version analytics
version_stats = client.analytics.get_version_stats(
    prompt_id=prompt_id,
    version=1
)
print(f"Avg latency: {version_stats.avg_latency_ms}ms")
print(f"Total executions: {version_stats.total_executions}")
```

## Error Handling

The SDK raises specific exceptions for different error scenarios:

```python
from qhub.errors import (
    QHubAPIError,
    ValidationError,
    NotFoundError,
    UnauthorizedError,
)

try:
    org = client.organizations.get("non-existent")
except NotFoundError as e:
    print(f"Organization not found: {e}")
except UnauthorizedError as e:
    print(f"Invalid token: {e}")
except ValidationError as e:
    print(f"Invalid input: {e}")
except QHubAPIError as e:
    print(f"API error: {e.status_code} - {e.message}")
```

## Development

### Run Tests

```bash
# Install dev dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run with coverage
pytest --cov=qhub --cov-report=term-missing
```

### Project Structure

```
apps/sdk-python/
├── src/qhub/
│   ├── client.py          # Main client
│   ├── types.py           # Pydantic models
│   ├── errors.py          # Exception classes
│   └── resources/         # Resource implementations
│       ├── organizations.py
│       ├── projects.py
│       ├── prompts.py
│       ├── versions.py
│       ├── logs.py
│       ├── evaluations.py
│       ├── consulting.py
│       ├── search.py
│       ├── analytics.py
│       ├── tags.py
│       └── industries.py
└── tests/                 # Test suite
```

## Requirements

- Python 3.10 or higher
- httpx >= 0.27.0
- pydantic >= 2.0

## License

MIT

## Related

- [Go SDK](../sdk/) - Official Go SDK
- [TypeScript SDK](../sdk-typescript/) - Official TypeScript SDK
- [API Documentation](../api/) - QHub API server
