# API Design

## Base URL

```
https://api.promptlab.dev/api/v1
```

## Authentication

### API Key 認証（SDK / 外部連携用）

```
X-API-Key: pl_live_xxxxxxxxxxxxxxxxxxxx
```

API Key は Organization 単位で発行。プレフィクス: `pl_live_` (本番) / `pl_test_` (テスト)。

### Bearer Token 認証（Web UI 用）

```
Authorization: Bearer <jwt_token>
```

---

## Endpoints

### Organizations

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /organizations | 組織作成 | Bearer |
| GET | /organizations/{org_slug} | 組織取得 | Bearer |
| PUT | /organizations/{org_slug} | 組織更新 | Bearer (owner/admin) |

### Members

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /organizations/{org_slug}/members | メンバー一覧 | Bearer |
| POST | /organizations/{org_slug}/members | メンバー招待 | Bearer (owner/admin) |
| PUT | /organizations/{org_slug}/members/{user_id} | ロール変更 | Bearer (owner) |
| DELETE | /organizations/{org_slug}/members/{user_id} | メンバー削除 | Bearer (owner/admin) |

### API Keys

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /organizations/{org_slug}/api-keys | API Key 作成 | Bearer (owner/admin) |
| GET | /organizations/{org_slug}/api-keys | API Key 一覧 | Bearer (owner/admin) |
| DELETE | /organizations/{org_slug}/api-keys/{key_id} | API Key 無効化 | Bearer (owner/admin) |

### Projects

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /organizations/{org_slug}/projects | プロジェクト作成 | Bearer (admin+) |
| GET | /organizations/{org_slug}/projects | プロジェクト一覧 | Bearer |
| GET | /organizations/{org_slug}/projects/{project_slug} | プロジェクト取得 | Bearer |
| PUT | /organizations/{org_slug}/projects/{project_slug} | プロジェクト更新 | Bearer (admin+) |
| DELETE | /organizations/{org_slug}/projects/{project_slug} | プロジェクト削除 | Bearer (owner) |

### Prompts

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /projects/{project_id}/prompts | プロンプト作成 | Bearer (member+) |
| GET | /projects/{project_id}/prompts | プロンプト一覧 | Bearer |
| GET | /projects/{project_id}/prompts/{prompt_slug} | プロンプト取得 | Bearer |
| PUT | /projects/{project_id}/prompts/{prompt_slug} | プロンプト更新 | Bearer (member+) |

### Prompt Versions

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /prompts/{prompt_id}/versions | 新バージョン作成 (draft) | Bearer (member+) |
| GET | /prompts/{prompt_id}/versions | バージョン一覧 | Bearer |
| GET | /prompts/{prompt_id}/versions/{version} | バージョン取得 | Bearer |
| GET | /prompts/{prompt_id}/versions/latest | 最新バージョン | Bearer / API Key |
| GET | /prompts/{prompt_id}/versions/production | Production バージョン | Bearer / API Key |
| PUT | /prompts/{prompt_id}/versions/{version}/status | ステータス変更 | Bearer (member+) |
| GET | /prompts/{prompt_id}/diff/{v1}...{v2} | テキスト diff | Bearer |
| GET | /prompts/{prompt_id}/semantic-diff/{v1}...{v2} | Semantic diff | Bearer |
| GET | /prompts/{prompt_id}/versions/{version}/lint | Lint 結果 | Bearer |

### Execution Logs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /logs | ログ送信 | API Key |
| POST | /logs/batch | ログ一括送信 | API Key |
| GET | /prompts/{prompt_id}/logs | プロンプト別ログ一覧 | Bearer |
| GET | /logs/{log_id} | ログ詳細 | Bearer |

### Evaluations

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /logs/{log_id}/evaluations | 評価作成 | Bearer / API Key |
| GET | /logs/{log_id}/evaluations | 評価一覧 | Bearer |
| PUT | /evaluations/{evaluation_id} | 評価更新 | Bearer |

### Analytics

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /prompts/{prompt_id}/analytics | プロンプト分析 | Bearer |
| GET | /prompts/{prompt_id}/versions/{v}/analytics | バージョン分析 | Bearer |
| GET | /projects/{project_id}/analytics | プロジェクト分析 | Bearer |

### Tags

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /organizations/{org_slug}/tags | タグ作成 | Bearer |
| GET | /organizations/{org_slug}/tags | タグ一覧 | Bearer |
| POST | /prompts/{prompt_id}/tags | タグ付け | Bearer |
| DELETE | /prompts/{prompt_id}/tags/{tag_id} | タグ解除 | Bearer |

### Consulting Chat

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /consulting/sessions | セッション作成 | Bearer |
| GET | /consulting/sessions | セッション一覧 | Bearer |
| GET | /consulting/sessions/{session_id} | セッション詳細 (全メッセージ) | Bearer |
| POST | /consulting/sessions/{session_id}/messages | メッセージ送信 (SSE レスポンス) | Bearer |
| PUT | /consulting/sessions/{session_id} | セッション更新 (close 等) | Bearer |

### Industry Config

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /organizations/{org_slug}/industries | 業界設定一覧 | Bearer |
| PUT | /organizations/{org_slug}/industries/{industry} | 業界設定更新 | Bearer (admin+) |
| GET | /organizations/{org_slug}/industries/{industry}/benchmarks | ベンチマーク取得 | Bearer |
| POST | /organizations/{org_slug}/industries/{industry}/compliance-check | コンプライアンスチェック | Bearer |

---

## Key Request / Response Examples

### POST /logs (Execution Log Ingestion)

**Request**:
```json
{
  "prompt_id": "550e8400-e29b-41d4-a716-446655440000",
  "version_number": 3,
  "request": {
    "resolved_prompt": "You are a helpful coding assistant...",
    "variables": {
      "language": "Go",
      "question": "How to handle concurrent access?"
    }
  },
  "response": {
    "content": "To handle concurrent access in Go...",
    "finish_reason": "end_turn"
  },
  "provider": "anthropic",
  "model": "claude-sonnet-4-6",
  "token_usage": {
    "input_tokens": 150,
    "output_tokens": 320,
    "total_tokens": 470
  },
  "latency_ms": 1250,
  "executed_at": "2026-03-08T10:30:00Z",
  "metadata": {
    "environment": "production",
    "user_session_id": "sess_abc123"
  }
}
```

**Response** (201 Created):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "prompt_version_id": "770e8400-e29b-41d4-a716-446655440002",
  "created_at": "2026-03-08T10:30:05Z"
}
```

### PUT /prompts/{prompt_id}/versions/{version}/status

**Request**:
```json
{
  "status": "production"
}
```

**Response** (200 OK):
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "prompt_id": "550e8400-e29b-41d4-a716-446655440000",
  "version_number": 4,
  "status": "production",
  "published_at": "2026-03-08T11:00:00Z"
}
```

### POST /consulting/sessions/{session_id}/messages

Server-Sent Events (SSE) でストリーミングレスポンスを返す。

**Request**:
```json
{
  "content": "customer-support-bot の品質が v4 から下がっています。原因を分析してください。"
}
```

**Response** (200 OK, SSE stream):
```
event: message_start
data: {"message_id": "msg_001"}

event: content_delta
data: {"delta": "v4 での変更点を分析しました。\n\n"}

event: content_delta
data: {"delta": "📊 品質変化:\n  - 正確性: 85.2 → 79.1 (-6.1)\n"}

event: citation
data: {"type": "execution_log", "log_id": "uuid", "score": 79.1}

event: content_delta
data: {"delta": "\n💡 改善提案:\n安全性制約を維持しつつ..."}

event: action_available
data: {"type": "create_version", "prompt_id": "uuid", "description": "コンサル提案に基づく改善"}

event: message_end
data: {"message_id": "msg_001", "citations_count": 3}
```

### POST /organizations/{org_slug}/industries/{industry}/compliance-check

**Request**:
```json
{
  "prompt_id": "550e8400-e29b-41d4-a716-446655440000",
  "version_number": 6
}
```

**Response** (200 OK):
```json
{
  "industry": "healthcare",
  "compliance_rate": 75.0,
  "minimum_required": 90.0,
  "passed": false,
  "checks": [
    {"rule": "no_phi_reference", "status": "pass", "message": "PHI の直接参照なし"},
    {"rule": "data_retention_constraint", "status": "pass", "message": "データ保持制約あり"},
    {"rule": "consent_verification", "status": "warn", "message": "患者同意確認フローが不十分"},
    {"rule": "medical_disclaimer", "status": "fail", "message": "免責事項が欠落"}
  ],
  "suggestions": [
    "\"Always verify patient consent before discussing treatment details\" を追加",
    "\"This is not a substitute for professional medical advice\" を追加"
  ]
}
```

### GET /prompts/{prompt_id}/analytics

**Query Parameters**: `from`, `to`, `group_by` (version | day | week | month)

**Response** (200 OK):
```json
{
  "prompt_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": {"from": "2026-02-01T00:00:00Z", "to": "2026-03-08T23:59:59Z"},
  "summary": {
    "total_executions": 1250,
    "total_evaluations": 430,
    "avg_score": 82.5,
    "avg_latency_ms": 980,
    "total_tokens": 587500,
    "estimated_cost_usd": 12.35
  },
  "by_version": [
    {"version_number": 3, "status": "archived", "executions": 800, "avg_score": 78.2},
    {"version_number": 4, "status": "production", "executions": 450, "avg_score": 89.3}
  ],
  "trend": [
    {"date": "2026-02-01", "avg_score": 76.0, "executions": 45},
    {"date": "2026-02-08", "avg_score": 78.5, "executions": 52}
  ]
}
```

---

## Pagination

カーソルベース。

```json
{
  "data": [],
  "pagination": {
    "next_cursor": "eyJpZCI6Ijk5OWUuLi4ifQ",
    "has_more": true
  }
}
```

Query: `?limit=20&cursor=eyJ...`

## Error Response

```json
{
  "error": {
    "type": "ValidationError",
    "message": "title must be between 2 and 200 characters",
    "details": [{"field": "title", "message": "must be between 2 and 200 characters"}]
  }
}
```

## Rate Limiting

| Plan | Log Ingestion | Other Endpoints | Consulting Chat |
|------|--------------|-----------------|-----------------|
| Free | 1,000 req/hour | 100 req/min | 10 sessions/day |
| Pro | 10,000 req/hour | 1,000 req/min | 50 sessions/day |
| Team | 50,000 req/hour | 5,000 req/min | unlimited |
| Enterprise | Custom | Custom | unlimited |
