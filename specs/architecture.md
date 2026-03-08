# System Architecture

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Clients                              │
├──────────────┬──────────────────────┬───────────────────────┤
│   Web UI     │   SDK (Go/Py/TS)    │   REST API            │
│ (templ+HTMX) │                      │                       │
└──────┬───────┴──────────┬───────────┴──────────┬────────────┘
       │ Bearer Token     │ API Key              │
       │                  │                      │
┌──────▼──────────────────▼──────────────────────▼────────────┐
│                     API Gateway (ALB + WAF)                   │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                     API Server (Go)                          │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Middleware: Auth → Tenant → RBAC → RateLimit → Log  │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌──────────────── Layer 1 ────────────────────────────┐   │
│  │  Routes: Org | Project | Prompt | Version | Log |   │   │
│  │          Evaluation | Analytics | Tag               │   │
│  │                                                     │   │
│  │  Domain: Organization | Project | Prompt |          │   │
│  │          Execution | Evaluation                     │   │
│  │                                                     │   │
│  │  Infra: RDS Repositories (sqlc)                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌──────────────── Layer 2 ────────────────────────────┐   │
│  │  Routes: Consulting Chat | Industry | Compliance    │   │
│  │                                                     │   │
│  │  Domain: ConsultingSession | IndustryConfig |       │   │
│  │          PlatformBenchmark                          │   │
│  │                                                     │   │
│  │  Service: ChatEngine (RAG + LLM)                    │   │
│  │          KnowledgeRetriever | BenchmarkAggregator   │   │
│  │                                                     │   │
│  │  Infra: RDS Repos | LLM Client | Vector Store      │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────┬───────────────┬────────────────┬───────────────┘
             │               │                │
     ┌───────▼──────┐ ┌─────▼─────┐  ┌───────▼──────┐
     │ PostgreSQL   │ │  Redis    │  │  LLM API     │
     │ (Aurora)     │ │ (Cache)   │  │ (Anthropic)  │
     └──────────────┘ └───────────┘  └──────────────┘
```

## Module Structure

```
apps/
├── api/src/
│   ├── cmd/main.go
│   ├── routes/
│   │   ├── routes.go                    # ルーター設定、DI
│   │   ├── middleware/
│   │   │   ├── bearer_auth.go           # JWT 認証
│   │   │   ├── apikey_auth.go           # API Key 認証
│   │   │   ├── tenant.go               # テナントコンテキスト
│   │   │   ├── rbac.go                 # ロールベースアクセス制御
│   │   │   ├── rate_limit.go           # レート制限
│   │   │   └── logger.go
│   │   ├── response/response.go
│   │   ├── tasks/                       # 既存サンプル
│   │   │
│   │   │── # Layer 1
│   │   ├── organizations/               # 組織 CRUD + メンバー管理
│   │   ├── projects/                    # プロジェクト CRUD
│   │   ├── prompts/                     # プロンプト CRUD + バージョン管理
│   │   ├── logs/                        # ログ ingestion
│   │   ├── evaluations/                 # 評価 CRUD
│   │   ├── analytics/                   # 分析 API
│   │   │
│   │   │── # Layer 2
│   │   ├── consulting/                  # コンサルチャット
│   │   │   ├── session.go              # セッション管理
│   │   │   └── message.go             # メッセージ送信 (SSE)
│   │   └── industries/                  # 業界設定 + コンプライアンス
│   │
│   ├── domain/
│   │   ├── apperror/                    # 既存
│   │   ├── task/                        # 既存サンプル
│   │   │
│   │   │── # Layer 1
│   │   ├── organization/
│   │   ├── user/
│   │   ├── project/
│   │   ├── prompt/                      # Prompt + Version + Content
│   │   ├── execution/                   # ExecutionLog + TokenUsage
│   │   ├── evaluation/
│   │   │
│   │   │── # Layer 2
│   │   ├── consulting/                  # Session + Message
│   │   └── industry/                    # IndustryConfig + Benchmark
│   │
│   ├── service/                         # Layer 2 のビジネスロジック
│   │   ├── chat_engine.go              # RAG + LLM オーケストレーション
│   │   ├── knowledge_retriever.go      # 知識ソース統合
│   │   ├── semantic_diff.go            # Semantic Diff 生成
│   │   ├── prompt_linter.go            # Prompt Linting
│   │   ├── compliance_checker.go       # コンプライアンスチェック
│   │   └── benchmark_aggregator.go     # ベンチマーク集計
│   │
│   └── infra/
│       ├── rds/                         # PostgreSQL リポジトリ
│       │   ├── task_repository/         # 既存サンプル
│       │   ├── organization_repository/
│       │   ├── project_repository/
│       │   ├── prompt_repository/
│       │   ├── execution_repository/
│       │   ├── evaluation_repository/
│       │   ├── consulting_repository/
│       │   └── benchmark_repository/
│       ├── llm/                         # LLM API クライアント
│       │   ├── client.go               # Anthropic API クライアント
│       │   └── streaming.go            # SSE ストリーミング
│       └── cache/                       # Redis キャッシュ
│           └── client.go
│
├── web/src/                             # templ + HTMX フロントエンド
│   ├── templates/
│   │   ├── layout.templ
│   │   ├── dashboard.templ              # ダッシュボード
│   │   ├── prompts/                     # プロンプト管理
│   │   │   ├── list.templ
│   │   │   ├── detail.templ
│   │   │   ├── version_diff.templ      # Diff 表示
│   │   │   └── editor.templ            # プロンプトエディタ
│   │   ├── logs/                        # ログ一覧・詳細
│   │   ├── analytics/                   # 分析画面
│   │   ├── consulting/                  # コンサルチャット UI
│   │   │   ├── chat.templ              # チャット画面
│   │   │   └── sessions.templ          # セッション一覧
│   │   └── settings/                    # 組織設定
│   └── ...
│
├── pkgs/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── tasks.sql               # 既存
│   │   │   ├── organizations.sql
│   │   │   ├── projects.sql
│   │   │   ├── prompts.sql
│   │   │   ├── prompt_versions.sql
│   │   │   ├── execution_logs.sql
│   │   │   ├── evaluations.sql
│   │   │   ├── consulting.sql
│   │   │   └── benchmarks.sql
│   │   └── migrations/
│   └── ...
│
└── sdk/                                 # クライアント SDK
    ├── go/
    ├── python/
    └── typescript/
```

## Consulting Chat Engine Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Chat Engine                              │
│                                                            │
│  User Message                                              │
│       │                                                    │
│       ▼                                                    │
│  ┌──────────────────────────────────────────────────┐     │
│  │  Intent Classifier                                │     │
│  │  ・改善提案依頼 → ImprovementFlow                 │     │
│  │  ・ベンチマーク比較 → BenchmarkFlow               │     │
│  │  ・コンプライアンス → ComplianceFlow              │     │
│  │  ・一般質問 → GeneralFlow                         │     │
│  └──────────────┬───────────────────────────────────┘     │
│                 │                                          │
│                 ▼                                          │
│  ┌──────────────────────────────────────────────────┐     │
│  │  Knowledge Retriever                              │     │
│  │                                                   │     │
│  │  ┌────────────┐ ┌────────────┐ ┌──────────────┐  │     │
│  │  │ Org Data   │ │ Industry   │ │ Platform     │  │     │
│  │  │ Retriever  │ │ KB Lookup  │ │ Benchmarks   │  │     │
│  │  │            │ │            │ │              │  │     │
│  │  │ Prompts    │ │ Regulations│ │ Anonymized   │  │     │
│  │  │ Versions   │ │ Best       │ │ cross-org    │  │     │
│  │  │ Logs       │ │ Practices  │ │ statistics   │  │     │
│  │  │ Evaluations│ │ Templates  │ │              │  │     │
│  │  └────────────┘ └────────────┘ └──────────────┘  │     │
│  └──────────────┬───────────────────────────────────┘     │
│                 │ Context                                   │
│                 ▼                                          │
│  ┌──────────────────────────────────────────────────┐     │
│  │  LLM (Claude API)                                 │     │
│  │                                                   │     │
│  │  System Prompt:                                   │     │
│  │  "You are a prompt engineering consultant         │     │
│  │   specializing in {industry}.                     │     │
│  │   Use the provided org data and industry          │     │
│  │   knowledge to give specific, actionable advice." │     │
│  │                                                   │     │
│  │  Context: [retrieved knowledge chunks]            │     │
│  │  Chat History: [previous messages]                │     │
│  │  User Message: [current message]                  │     │
│  └──────────────┬───────────────────────────────────┘     │
│                 │ Stream                                    │
│                 ▼                                          │
│  ┌──────────────────────────────────────────────────┐     │
│  │  Response Processor                               │     │
│  │  ・Citations を抽出・構造化                        │     │
│  │  ・アクション候補を検出                            │     │
│  │  ・SSE ストリームとしてクライアントに返却          │     │
│  └──────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────┘
```

### Knowledge Retriever の詳細

| Source | 取得方法 | 用途 |
|--------|---------|------|
| Org Prompts | SQL: prompt + versions by org_id | 現在のプロンプト構成の把握 |
| Org Logs | SQL: recent execution_logs + evaluations | 品質傾向の分析 |
| Org Analytics | SQL: aggregated scores by version | バージョン比較データ |
| Industry KB | 静的 JSON/Markdown files (初期)、将来的に DB | 業界規制・ベストプラクティス |
| Platform Benchmarks | SQL: platform_benchmarks table | 匿名クロス組織統計 |

初期段階では Vector Store は不要。SQL クエリ + 静的ファイルで十分な精度が出る。
データ量が増えた段階で pgvector or 外部ベクトル DB を検討。

---

## Multi-Tenancy

### Shared Database, Shared Schema

`organization_id` による論理分離。

```go
// middleware/tenant.go
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        orgID := extractOrgID(r) // JWT or API Key → org_id
        ctx := context.WithValue(r.Context(), TenantKey, orgID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

すべての Repository メソッドで `OrganizationID` フィルタを強制。

---

## Authentication

### JWT (Web UI)

```
User → Cognito → JWT → API Server → Extract user_id, org_id from claims
```

### API Key (SDK)

```
SDK → X-API-Key header → API Server → Hash → Lookup api_keys → Resolve org_id
```

---

## RBAC

| Action | Owner | Admin | Member | Viewer |
|--------|-------|-------|--------|--------|
| 組織設定変更 | o | x | x | x |
| メンバー管理 | o | o | x | x |
| API Key 管理 | o | o | x | x |
| プロジェクト作成/削除 | o | o | x | x |
| プロンプト作成/編集 | o | o | o | x |
| バージョン昇格 (→ production) | o | o | o | x |
| ログ閲覧 | o | o | o | o |
| 評価作成 | o | o | o | x |
| 分析閲覧 | o | o | o | o |
| コンサルチャット利用 | o | o | o | o |
| 業界設定変更 | o | o | x | x |

---

## OSS vs Cloud の境界

```
┌─────────────────────────────────┬────────────────────────────────┐
│          OSS Core               │         Cloud (SaaS)           │
│         (Apache 2.0)            │         (Proprietary)          │
├─────────────────────────────────┼────────────────────────────────┤
│ Prompt CRUD + Versioning        │ Semantic Diff (LLM)            │
│ Lifecycle (draft→prod)          │ Prompt Linting (LLM)           │
│ Execution Log Ingestion         │ Consulting Chat                │
│ Basic Evaluation (manual)       │ Industry Knowledge Base        │
│ Basic Analytics (counts, avg)   │ Compliance Check               │
│ Text Diff                       │ Platform Benchmarks            │
│ CLI + SDK (Go/Py/TS)            │ Advanced Analytics             │
│ Basic Web UI                    │ SSO / Audit Log                │
│ API Key / Bearer Auth           │ Managed hosting                │
│ Docker Compose for self-host    │ SLA                            │
└─────────────────────────────────┴────────────────────────────────┘
```

---

## Scalability Phases

### Phase 1 (MVP)
- 単一 API + Aurora Serverless v2
- Redis: セッション、レート制限、Semantic Diff キャッシュ
- 同期書き込み
- Industry KB: 静的 JSON ファイル

### Phase 2 (Growth)
- execution_logs の非同期書き込み (SQS → Worker)
- Read Replica 活用
- execution_logs 月次パーティショニング
- Industry KB を DB に移行

### Phase 3 (Scale)
- ログ ingestion を専用サービスに分離
- Analytics 用のデータウェアハウス
- Vector Store for Knowledge Retrieval
- CDN でフロントエンド配信
