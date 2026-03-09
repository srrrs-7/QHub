# QHub TypeScript SDK

TypeScript SDK for the QHub prompt/answer version management API.

## Features

- 🚀 **Modern TypeScript** with full type safety
- 🌐 **Native fetch** API (Node.js 18+, browsers, edge runtimes)
- 📦 **Zero runtime dependencies**
- 🔄 **Promise-based** async/await API
- 🧪 **Fully tested** with Vitest

## Installation

```bash
# npm
npm install @qhub/sdk

# yarn
yarn add @qhub/sdk

# pnpm
pnpm add @qhub/sdk
```

## Quick Start

```typescript
import { QHubClient } from '@qhub/sdk';

// Initialize client
const client = new QHubClient({
  baseUrl: 'http://localhost:8080', // optional, defaults to localhost:8080
  bearerToken: 'your-bearer-token',
});

// Create an organization
const org = await client.organizations.create({
  name: 'My Organization',
  slug: 'my-org',
});

// Create a project
const project = await client.projects.create({
  organizationId: org.id,
  name: 'AI Assistant',
  slug: 'ai-assistant',
});

// Create a prompt
const prompt = await client.prompts.create({
  projectId: project.id,
  name: 'System Prompt',
  slug: 'system-prompt',
  content: { text: 'You are a helpful AI assistant.' },
});

// Create a version
const version = await client.versions.create({
  promptId: prompt.id,
  content: { text: 'You are a helpful AI assistant with expertise in TypeScript.' },
  changeSummary: 'Added TypeScript expertise',
});

// Log execution
const log = await client.logs.create({
  versionId: version.id,
  inputText: 'How do I use async/await?',
  outputText: 'Async/await is syntactic sugar...',
  latencyMs: 250,
  tokenCount: 150,
});
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
| `client.apiKeys` | API key management |
| `client.members` | Organization member management |
| `client.industries` | Industry configurations |

## Authentication

QHub API supports two authentication methods:

### Bearer Token (recommended)

```typescript
const client = new QHubClient({
  bearerToken: 'your-bearer-token',
});
```

### API Key (for log ingestion)

API keys are managed per organization and used for server-to-server log ingestion:

```typescript
// Create API key via apiKeys resource
const apiKey = await client.apiKeys.create({
  organizationId: orgId,
  name: 'Production Logger',
});

// Use in log ingestion
const logClient = new QHubClient({
  bearerToken: apiKey.key,
});
await logClient.logs.create({...});
```

## Advanced Usage

### Version Comparison

```typescript
// Get semantic diff between versions
const diff = await client.versions.diff({
  promptId,
  v1: 1,
  v2: 2,
});

console.log(`Length change: ${diff.lengthChange}`);
console.log(`Variable changes: ${diff.variableChanges}`);
console.log(`Text diff:\n${diff.textDiff}`);

// Compare version metrics (Welch's t-test)
const comparison = await client.versions.compare({
  promptId,
  v1: 1,
  v2: 2,
  metric: 'latency_ms',
});

console.log(`p-value: ${comparison.pValue}`);
console.log(`Statistically significant: ${comparison.significant}`);
```

### Lint Prompt

```typescript
const lintResult = await client.versions.lint({
  promptId,
  version: 1,
});

console.log(`Score: ${lintResult.score}/100`);
lintResult.issues.forEach((issue) => {
  console.log(`- ${issue.severity}: ${issue.message}`);
});
```

### Semantic Search

```typescript
const results = await client.search.semantic({
  query: 'prompt for customer support chatbot',
  limit: 5,
});

results.forEach((result) => {
  console.log(`${result.promptName} v${result.version} (score: ${result.score})`);
});
```

### Consulting (RAG)

```typescript
// Create session
const session = await client.consulting.createSession({
  projectId,
  title: 'Improve conversion rate',
});

// Send message (non-streaming)
const response = await client.consulting.sendMessage({
  sessionId: session.id,
  message: 'How can I improve my prompt for better responses?',
});

console.log(response.content);
```

### Analytics

```typescript
// Project analytics
const stats = await client.analytics.getProjectStats(projectId);
console.log(`Total prompts: ${stats.totalPrompts}`);
console.log(`Total logs: ${stats.totalLogs}`);

// Prompt version analytics
const versionStats = await client.analytics.getVersionStats({
  promptId,
  version: 1,
});
console.log(`Avg latency: ${versionStats.avgLatencyMs}ms`);
console.log(`Total executions: ${versionStats.totalExecutions}`);
```

### Organization Management

```typescript
// Add member
await client.members.add({
  organizationId: orgId,
  userId: 'user-uuid',
  role: 'member',
});

// Update member role
await client.members.update({
  organizationId: orgId,
  userId: 'user-uuid',
  role: 'admin',
});

// Remove member
await client.members.remove({
  organizationId: orgId,
  userId: 'user-uuid',
});
```

## Error Handling

The SDK throws specific error classes for different scenarios:

```typescript
import {
  QHubAPIError,
  ValidationError,
  NotFoundError,
  UnauthorizedError,
} from '@qhub/sdk';

try {
  const org = await client.organizations.get('non-existent');
} catch (error) {
  if (error instanceof NotFoundError) {
    console.error('Organization not found:', error.message);
  } else if (error instanceof UnauthorizedError) {
    console.error('Invalid token:', error.message);
  } else if (error instanceof ValidationError) {
    console.error('Invalid input:', error.message);
  } else if (error instanceof QHubAPIError) {
    console.error(`API error: ${error.status} - ${error.message}`);
  }
}
```

## Custom Fetch

You can provide a custom fetch implementation:

```typescript
import { QHubClient } from '@qhub/sdk';

const client = new QHubClient({
  bearerToken: 'your-token',
  fetchFn: customFetch, // your custom fetch function
});
```

This is useful for:
- Adding request/response interceptors
- Using fetch polyfills in older environments
- Custom retry logic
- Request debugging

## Development

### Run Tests

```bash
# Install dependencies
npm install

# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Type check
npm run typecheck
```

### Build

```bash
npm run build
```

### Project Structure

```
apps/sdk-typescript/
├── src/
│   ├── client.ts          # Main client
│   ├── types.ts           # TypeScript types
│   ├── errors.ts          # Error classes
│   ├── index.ts           # Public API
│   └── resources/         # Resource implementations
│       ├── organizations.ts
│       ├── projects.ts
│       ├── prompts.ts
│       ├── versions.ts
│       ├── logs.ts
│       ├── evaluations.ts
│       ├── consulting.ts
│       ├── search.ts
│       ├── analytics.ts
│       ├── apikeys.ts
│       ├── members.ts
│       ├── tags.ts
│       └── industries.ts
└── tests/                 # Test suite
```

## Requirements

- Node.js 18.0.0 or higher (for native fetch support)
- TypeScript 5.4+ (for development)

## Browser Support

This SDK uses native fetch, which is supported in:
- All modern browsers (Chrome, Firefox, Safari, Edge)
- Node.js 18+
- Deno
- Cloudflare Workers
- Vercel Edge Runtime

## License

MIT

## Related

- [Go SDK](../sdk/) - Official Go SDK
- [Python SDK](../sdk-python/) - Official Python SDK
- [API Documentation](../api/) - QHub API server
