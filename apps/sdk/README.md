# QHub Go SDK

Official Go SDK for the QHub prompt/answer version management API.

## Features

- 🚀 **Go 1.25+** support
- 📦 **Zero external dependencies** (standard library only)
- 🔄 **Context-aware** API with timeout support
- ✅ **Type-safe** with Go structs
- 🧪 **Production-ready**

## Installation

```bash
go get sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "sdk"
)

func main() {
    // Initialize client
    client := sdk.NewClient(
        "your-bearer-token",
        sdk.WithBaseURL("http://localhost:8080"),
    )

    ctx := context.Background()

    // Create an organization
    org, err := client.CreateOrganization(ctx, sdk.CreateOrganizationRequest{
        Name: "My Organization",
        Slug: "my-org",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a project
    project, err := client.CreateProject(ctx, sdk.CreateProjectRequest{
        OrganizationID: org.ID,
        Name:          "AI Assistant",
        Slug:          "ai-assistant",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a prompt
    prompt, err := client.CreatePrompt(ctx, sdk.CreatePromptRequest{
        ProjectID: project.ID,
        Name:      "System Prompt",
        Slug:      "system-prompt",
        Content:   map[string]interface{}{"text": "You are a helpful AI assistant."},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a version
    version, err := client.CreateVersion(ctx, sdk.CreateVersionRequest{
        PromptID:      prompt.ID,
        Content:       map[string]interface{}{"text": "You are a helpful AI assistant with expertise in Go."},
        ChangeSummary: "Added Go expertise",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Log execution
    _, err = client.CreateLog(ctx, sdk.CreateLogRequest{
        VersionID:  version.ID,
        InputText:  "How do I use goroutines?",
        OutputText: "Goroutines are lightweight threads...",
        LatencyMs:  250,
        TokenCount: 150,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Successfully created organization, project, prompt, version, and log!")
}
```

## Authentication

QHub API supports two authentication methods:

### Bearer Token (recommended)

```go
client := sdk.NewClient("your-bearer-token")
```

### API Key (for log ingestion)

API keys are managed per organization and used for server-to-server log ingestion:

```go
// Create API key (using bearer token auth)
adminClient := sdk.NewClient("admin-bearer-token")
apiKey, err := adminClient.CreateAPIKey(ctx, sdk.CreateAPIKeyRequest{
    OrganizationID: orgID,
    Name:          "Production Logger",
})

// Use API key for log ingestion
logClient := sdk.NewClient(apiKey.Key)
_, err = logClient.CreateLog(ctx, sdk.CreateLogRequest{...})
```

## Available Methods

### Organizations

```go
// Create organization
org, err := client.CreateOrganization(ctx, sdk.CreateOrganizationRequest{...})

// Get organization
org, err := client.GetOrganization(ctx, "org-slug")

// Update organization
org, err := client.UpdateOrganization(ctx, "org-slug", sdk.UpdateOrganizationRequest{...})
```

### Projects

```go
// List projects
projects, err := client.ListProjects(ctx, orgID)

// Create project
project, err := client.CreateProject(ctx, sdk.CreateProjectRequest{...})

// Get project
project, err := client.GetProject(ctx, orgID, "project-slug")

// Update project
project, err := client.UpdateProject(ctx, orgID, "project-slug", sdk.UpdateProjectRequest{...})

// Delete project
err := client.DeleteProject(ctx, orgID, "project-slug")
```

### Prompts

```go
// List prompts
prompts, err := client.ListPrompts(ctx, projectID)

// Create prompt
prompt, err := client.CreatePrompt(ctx, sdk.CreatePromptRequest{...})

// Get prompt
prompt, err := client.GetPrompt(ctx, projectID, "prompt-slug")

// Update prompt
prompt, err := client.UpdatePrompt(ctx, projectID, "prompt-slug", sdk.UpdatePromptRequest{...})
```

### Versions

```go
// List versions
versions, err := client.ListVersions(ctx, promptID)

// Create version
version, err := client.CreateVersion(ctx, sdk.CreateVersionRequest{...})

// Get version
version, err := client.GetVersion(ctx, promptID, versionNumber)

// Update version status
version, err := client.UpdateVersionStatus(ctx, promptID, versionNumber, "active")
```

### Execution Logs

```go
// List logs
logs, err := client.ListLogs(ctx, sdk.ListLogsRequest{
    VersionID: versionID,
    Limit:     100,
})

// Create log
log, err := client.CreateLog(ctx, sdk.CreateLogRequest{...})

// Batch create logs
batchResp, err := client.CreateLogBatch(ctx, []sdk.CreateLogRequest{...})

// Get log
log, err := client.GetLog(ctx, logID)
```

### Evaluations

```go
// List evaluations for log
evals, err := client.ListEvaluationsForLog(ctx, logID)

// Create evaluation
eval, err := client.CreateEvaluation(ctx, sdk.CreateEvaluationRequest{...})

// Get evaluation
eval, err := client.GetEvaluation(ctx, evalID)

// Update evaluation
eval, err := client.UpdateEvaluation(ctx, evalID, sdk.UpdateEvaluationRequest{...})
```

## Advanced Usage

### Version Comparison

```go
// Get semantic diff
diff, err := client.GetVersionDiff(ctx, promptID, 1, 2)
fmt.Printf("Length change: %d\n", diff.LengthChange)
fmt.Printf("Text diff:\n%s\n", diff.TextDiff)

// Compare metrics (Welch's t-test)
comparison, err := client.CompareVersions(ctx, promptID, 1, 2, "latency_ms")
fmt.Printf("p-value: %.4f\n", comparison.PValue)
fmt.Printf("Significant: %v\n", comparison.Significant)
```

### Lint Prompt

```go
lintResult, err := client.LintVersion(ctx, promptID, 1)
fmt.Printf("Score: %d/100\n", lintResult.Score)
for _, issue := range lintResult.Issues {
    fmt.Printf("- %s: %s\n", issue.Severity, issue.Message)
}
```

### Semantic Search

```go
results, err := client.SemanticSearch(ctx, sdk.SemanticSearchRequest{
    Query: "prompt for customer support chatbot",
    Limit: 5,
})
for _, result := range results {
    fmt.Printf("%s v%d (score: %.2f)\n", result.PromptName, result.Version, result.Score)
}
```

### Analytics

```go
// Project analytics
stats, err := client.GetProjectAnalytics(ctx, projectID)
fmt.Printf("Total prompts: %d\n", stats.TotalPrompts)
fmt.Printf("Total logs: %d\n", stats.TotalLogs)

// Version analytics
versionStats, err := client.GetVersionAnalytics(ctx, promptID, 1)
fmt.Printf("Avg latency: %.2fms\n", versionStats.AvgLatencyMs)
fmt.Printf("Total executions: %d\n", versionStats.TotalExecutions)
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client := sdk.NewClient(
    "your-bearer-token",
    sdk.WithHTTPClient(httpClient),
    sdk.WithBaseURL("https://api.qhub.example.com"),
)
```

## Error Handling

The SDK returns standard Go errors. Check for specific error types:

```go
org, err := client.GetOrganization(ctx, "non-existent")
if err != nil {
    switch {
    case sdk.IsNotFoundError(err):
        fmt.Println("Organization not found")
    case sdk.IsUnauthorizedError(err):
        fmt.Println("Invalid token")
    case sdk.IsValidationError(err):
        fmt.Println("Invalid input")
    default:
        fmt.Printf("API error: %v\n", err)
    }
}
```

## Context and Timeouts

All methods accept a `context.Context` for cancellation and timeout control:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

org, err := client.GetOrganization(ctx, "my-org")

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel after 2 seconds
}()

logs, err := client.ListLogs(ctx, sdk.ListLogsRequest{Limit: 1000})
```

## Development

### Project Structure

```
apps/sdk/
├── client.go       # Main client implementation
├── types.go        # Request/response types
├── errors.go       # Error types and helpers
├── go.mod          # Module definition
└── README.md       # This file
```

### Testing

The SDK is designed to be testable with standard Go testing tools:

```go
func TestMyApp(t *testing.T) {
    // Use a test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock responses
    }))
    defer server.Close()

    client := sdk.NewClient(
        "test-token",
        sdk.WithBaseURL(server.URL),
    )

    // Test your code
}
```

## Requirements

- Go 1.25 or higher
- No external dependencies (standard library only)

## License

MIT

## Related

- [Python SDK](../sdk-python/) - Official Python SDK
- [TypeScript SDK](../sdk-typescript/) - Official TypeScript SDK
- [API Documentation](../api/) - QHub API server
- [CLI](../cli/) - QHub command-line interface
