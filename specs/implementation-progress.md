# 実装進捗レポート

最終更新: 2026-03-09

---

## 全体サマリー

ロードマップの **Phase 1〜7** を全て実装完了。セマンティック検索（Embedding）統合も完了済み。

| Phase | ステータス | 備考 |
|-------|-----------|------|
| 1: Foundation | ✅ 完了 | Organization, Member, API Key, Middleware |
| 2: Core | ✅ 完了 | Project, Prompt, Version, Lifecycle, Text Diff |
| 3: Ingestion | ✅ 完了 | Execution Log, Evaluation, Go SDK |
| 4: Intelligence | ✅ 完了 | Semantic Diff, Prompt Linting, Text Diff |
| 5: Analytics | ✅ 完了 | Prompt/Version/Project Analytics, Daily Trend |
| 6: Consulting | ✅ 完了 | Session, Message, Industry Config |
| 7: Polish | ✅ 完了 | Compliance Check, Benchmarks, Tags, Web UI |
| + Embedding | ✅ 完了 | TEI統合, セマンティック検索, 非同期埋め込み生成 |

---

## Phase 1: Foundation

### 1.1 DB Migration ✅
- `20260308000001_add_organizations_users_apikeys.sql`
- テーブル: `organizations`, `users`, `organization_members`, `api_keys`
- インデックス: slug (UNIQUE), email (UNIQUE), key_hash (UNIQUE)

### 1.2 Domain Layer ✅
- **Organization**: `OrganizationID`, `OrganizationName`, `OrganizationSlug`, `Plan`, `MemberRole`
- **値オブジェクト**: バリデーション付きコンストラクタ
- **Repository Interface**: `OrganizationRepository`

### 1.3 Infrastructure ✅
- `infra/rds/organization_repository/` — read.go, write.go
- sqlc 生成コードによる型安全なクエリ

### 1.4 Middleware ✅
- **Bearer Auth** (`middleware/bearer.go`) — Bearer トークン認証
- **API Key Auth** (`middleware/apikey.go`) — X-API-Key ヘッダー認証、SHA-256 ハッシュ照合、有効期限チェック
- **Logger** (`middleware/logger.go`) — リクエストログ
- **CORS** (`middleware/cors.go`) — `CORS_ORIGINS` 環境変数によるクロスオリジン設定
- **RBAC** (`middleware/rbac.go`) — ロールベースアクセス制御 (owner > admin > member > viewer)
- **Rate Limit** (`middleware/ratelimit.go`) — Token Bucket (60 req/min, burst 10)
- **Tenant** (`middleware/tenant.go`) — 組織スコープのリクエスト分離

### 1.5 Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/organizations` | POST | `Organization.Post()` |
| `/api/v1/organizations/{org_slug}` | GET | `Organization.Get()` |
| `/api/v1/organizations/{org_slug}` | PUT | `Organization.Put()` |
| `/api/v1/organizations/{org_id}/members` | GET, POST | `Member.List()`, `Member.Post()` |
| `/api/v1/organizations/{org_id}/members/{user_id}` | PUT, DELETE | `Member.Put()`, `Member.Delete()` |
| `/api/v1/organizations/{org_id}/api-keys` | GET, POST | `ApiKey.List()`, `ApiKey.Post()` |
| `/api/v1/organizations/{org_id}/api-keys/{id}` | DELETE | `ApiKey.Delete()` |

---

## Phase 2: Core — Prompt & Version

### 2.1 DB Migration ✅
- `20260308000002_add_projects_prompts_versions.sql`
- テーブル: `projects`, `prompts`, `prompt_versions`
- ステータス遷移: draft → review → production → archived

### 2.2 Domain Layer ✅
- **Project**: `ProjectID`, `ProjectName`, `ProjectSlug`, `ProjectDescription`
- **Prompt**: `PromptID`, `PromptName`, `PromptSlug`, `PromptType`, `PromptDescription`
- **PromptVersion**: イミュータブル、`ChangeDescription` 値オブジェクト
- **バージョン自動インクリメント**: `latestVersion + 1`

### 2.3 Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/organizations/{org_id}/projects` | GET, POST | `Project.List()`, `Project.Post()` |
| `/api/v1/organizations/{org_id}/projects/{slug}` | GET, PUT, DELETE | `Project.Get()`, `Project.Put()`, `Project.Delete()` |
| `/api/v1/projects/{project_id}/prompts` | GET, POST | `Prompt.List()`, `Prompt.Post()` |
| `/api/v1/projects/{project_id}/prompts/{slug}` | GET, PUT | `Prompt.Get()`, `Prompt.Put()` |
| `/api/v1/prompts/{prompt_id}/versions` | GET, POST | `Prompt.ListVersions()`, `Prompt.PostVersion()` |
| `/api/v1/prompts/{prompt_id}/versions/{v}` | GET | `Prompt.GetVersion()` |
| `/api/v1/prompts/{prompt_id}/versions/{v}/status` | PUT | `Prompt.PutVersionStatus()` |
| `/api/v1/prompts/{prompt_id}/versions/{v}/text-diff` | GET | `Prompt.GetTextDiff()` |
| `/api/v1/prompts/{prompt_id}/semantic-diff/{v1}/{v2}` | GET | `Prompt.GetDiff()` |

---

## Phase 3: Ingestion — Log & Evaluation

### 3.1 DB Migration ✅
- `20260308000003_add_execution_logs_evaluations.sql`
- テーブル: `execution_logs`, `evaluations`
- トークン使用量、レイテンシ、コスト推定カラム

### 3.2 Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/logs` | GET, POST | `Log.List()`, `Log.Post()` |
| `/api/v1/logs/batch` | POST | `Log.PostBatch()` |
| `/api/v1/logs/{id}` | GET | `Log.Get()` |
| `/api/v1/evaluations` | POST | `Evaluation.Post()` |
| `/api/v1/evaluations/{id}` | GET | `Evaluation.Get()` |
| `/api/v1/logs/{log_id}/evaluations` | GET | `Evaluation.List()` |

### 3.3 Go SDK ✅
- **場所**: `apps/sdk/`
- **クライアントメソッド**:
  - `GetPromptLatest(slug)` — 最新バージョン取得
  - `GetPromptVersion(slug, version)` — 特定バージョン取得
  - `Log(entry)` — 実行ログ記録
  - `LogBatch(entries)` — バッチログ記録
  - `Evaluate(eval)` — 評価記録

---

## Phase 4: Intelligence

### 4.1 Semantic Diff ✅
- **サービス**: `services/diffservice/service.go`
- トーン検出 (formal, casual, strict, friendly, neutral)
- 変数変更追跡 (`{{variable}}` パターン)
- 長さ変化分析 + インパクトスコア (low/medium/high)
- LCS アルゴリズムによるテキスト Diff

### 4.2 Prompt Linting ✅
- **サービス**: `services/lintservice/service.go`
- **ルール (5種)**:
  1. `excessive-length` — 4000文字超過で警告
  2. `missing-output-format` — 出力形式未指定で警告
  3. `variable-check` — 未定義変数をエラー
  4. `no-vague-instruction` — 曖昧な指示を情報レベルで報告
- **スコアリング**: 100点満点 (Error: -25, Warning: -10, Info: -5)
- **エンドポイント**: `GET /prompts/{id}/versions/{v}/lint`

---

## Phase 5: Analytics

### 5.1 Analytics Queries ✅
- sqlc クエリ: バージョン別集計、プロンプト別集計、プロジェクト別集計
- 日次トレンドクエリ

### 5.2 Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/prompts/{prompt_id}/versions/{v}/analytics` | GET | `Analytics.GetVersionAnalytics()` |
| `/api/v1/prompts/{prompt_id}/analytics` | GET | `Analytics.GetPromptAnalytics()` |
| `/api/v1/prompts/{prompt_id}/trend` | GET | `Analytics.GetDailyTrend()` |
| `/api/v1/projects/{project_id}/analytics` | GET | `Analytics.GetProjectAnalytics()` |

---

## Phase 6: Consulting Chat

### 6.1 DB Migration ✅
- `20260308000004_add_consulting_tags.sql`
- テーブル: `consulting_sessions`, `consulting_messages`, `industry_configs`

### 6.2-6.4 Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/consulting/sessions` | GET, POST | `Consulting.ListSessions()`, `Consulting.PostSession()` |
| `/api/v1/consulting/sessions/{id}` | GET | `Consulting.GetSession()` |
| `/api/v1/consulting/sessions/{session_id}/messages` | GET, POST | `Consulting.ListMessages()`, `Consulting.PostMessage()` |

**追加実装済み**:
- ✅ SSE ストリーミングレスポンス (`sse.go`, `stream.go`)
- ✅ RAG パイプライン (`ragservice/` — Embed → Search → Context → Ollama)
- ✅ Intent Classifier (`intentservice/` — ルールベース EN/JP)

**未実装（今後の拡張）**:
- Citations 抽出
- ActionsTaken（チャットからのバージョン作成）

---

## Phase 7: Polish

### 7.1 Compliance Check ✅
- `POST /api/v1/industries/{slug}/compliance-check`
- 業界別ルールセット対応

### 7.2 Platform Benchmarks ✅
- `GET /api/v1/industries/{slug}/benchmarks`
- 業界別ベンチマーク取得

### 7.3 Web UI ✅
- **場所**: `apps/web/src/`
- **テンプレート (templ + HTMX)**:
  - `index.templ` — ホーム画面
  - `layout.templ` — 共通レイアウト
  - `projects.templ` — プロジェクト一覧
  - `prompts.templ` — プロンプト一覧
  - `prompt_detail.templ` — プロンプト詳細 + バージョン管理
- **HTMX パーシャル**: プロンプト作成、バージョン作成、ステータス変更
- **Material Design 3** スタイリング
- **ポート**: 3000

### 7.4 Tags ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/tags` | GET, POST | `Tag.List()`, `Tag.Post()` |
| `/api/v1/tags/{id}` | DELETE | `Tag.Delete()` |
| `/api/v1/prompts/{prompt_id}/tags` | GET, POST | `Tag.ListByPrompt()`, `Tag.AddToPrompt()` |
| `/api/v1/prompts/{prompt_id}/tags/{tag_id}` | DELETE | `Tag.RemoveFromPrompt()` |

---

## Embedding & セマンティック検索（追加実装）

### DB Migration ✅
- `20260308000005_add_pgvector_embeddings.sql`
- `prompt_versions` テーブルに `embedding real[]` カラム追加
- PL/pgSQL `cosine_similarity()` 関数（PostgreSQL native `real[]` 使用）
- 部分インデックス: `idx_prompt_versions_has_embedding`

### Embedding Client ✅
- **場所**: `apps/pkgs/embedding/client.go`
- **統合先**: Hugging Face Text Embeddings Inference (TEI) — BAAI/bge-m3 (1024次元)
- **メソッド**: `Embed()`, `EmbedOne()`, `Health()`
- **設定**: `EMBEDDING_URL` 環境変数（未設定時は無効モード）

### Embedding Service ✅
- **場所**: `api/src/services/embeddingservice/service.go`
- **機能**:
  - `EmbedVersionAsync()` — バージョン作成時に非同期（fire-and-forget）で埋め込み生成
  - `EmbedVersion()` — 同期的な埋め込み生成
  - `GenerateEmbedding()` — テキストから埋め込みベクトル生成
  - `Available()` — サービス有効判定
- **テキスト抽出**: JSONB content から "content", "text", "body", "system", "user" キーを探索

### Search Routes ✅
| エンドポイント | メソッド | ハンドラー |
|---|---|---|
| `/api/v1/search/semantic` | POST | `Search.SemanticSearch()` |
| `/api/v1/search/embedding-status` | GET | `Search.EmbeddingStatus()` |

**セマンティック検索リクエスト**:
```json
{
  "query": "検索テキスト",
  "org_id": "uuid",
  "limit": 10,
  "min_score": 0.5
}
```

### DI フロー
```
EMBEDDING_URL → embedding.Client → EmbeddingService → PromptHandler (非同期埋め込み)
                                                     → SearchHandler (検索)
```

---

## アーキテクチャ概要

### ディレクトリ構成

```
apps/
├── api/src/
│   ├── cmd/main.go              # エントリポイント、DI配線、グレースフルシャットダウン
│   ├── routes/
│   │   ├── routes.go            # Chi ルーター設定、全ハンドラー配線
│   │   ├── middleware/          # Bearer Auth, API Key Auth, Logger, CORS, RBAC, Rate Limit, Tenant
│   │   ├── requtil/             # リクエストデコード + バリデーション
│   │   ├── response/            # JSON レスポンス、AppError→HTTP マッピング
│   │   ├── tasks/               # タスク CRUD
│   │   ├── organizations/       # 組織 CRUD
│   │   ├── projects/            # プロジェクト CRUD
│   │   ├── prompts/             # プロンプト + バージョン管理 (12メソッド)
│   │   ├── logs/                # 実行ログ (POST, BATCH, GET, LIST)
│   │   ├── evaluations/         # 評価 CRUD
│   │   ├── analytics/           # 分析 (4メソッド)
│   │   ├── consulting/          # コンサルティングチャット
│   │   ├── tags/                # タグ CRUD + 関連付け
│   │   ├── industries/          # 業界設定、ベンチマーク、コンプライアンス
│   │   ├── apikeys/             # API キー管理
│   │   ├── members/             # メンバー管理
│   │   └── search/              # セマンティック検索
│   ├── domain/
│   │   ├── apperror/            # AppError インターフェース
│   │   ├── task/                # タスクドメイン
│   │   ├── organization/        # 組織ドメイン
│   │   ├── project/             # プロジェクトドメイン
│   │   ├── prompt/              # プロンプトドメイン
│   │   ├── executionlog/        # 実行ログドメイン
│   │   ├── consulting/          # コンサルティングドメイン
│   │   └── tag/                 # タグドメイン
│   ├── infra/rds/               # Repository 実装 (sqlc)
│   └── services/
│       ├── diffservice/         # Semantic + Text Diff (Redis キャッシュ対応)
│       ├── lintservice/         # Prompt Linting (5ルール)
│       ├── embeddingservice/    # 埋め込み生成・管理
│       ├── ragservice/          # RAG パイプライン (Embed → Search → Context → Ollama)
│       ├── intentservice/       # チャット意図分類 (EN/JP)
│       └── contentutil/         # テキスト抽出・変数検出
├── pkgs/
│   ├── db/                      # sqlc 生成コード、マイグレーション
│   ├── cache/                   # Redis キャッシュ (nil-safe, JSON型付き)
│   ├── embedding/               # TEI クライアント
│   ├── env/                     # 環境変数ユーティリティ
│   ├── logger/                  # 構造化ログ
│   ├── ollama/                  # Ollama LLM クライアント
│   └── testutil/                # テストヘルパー
├── web/src/                     # templ + HTMX フロントエンド
├── sdk/                         # Go SDK
├── migrate/                     # マイグレーション実行
└── iac/                         # Terraform (AWS)
```

### テーブル一覧 (13+)

| テーブル | Phase | 用途 |
|---------|-------|------|
| `tasks` | 0 | タスク管理 |
| `organizations` | 1 | 組織 |
| `users` | 1 | ユーザー |
| `organization_members` | 1 | 組織メンバー |
| `api_keys` | 1 | API キー |
| `projects` | 2 | プロジェクト |
| `prompts` | 2 | プロンプト |
| `prompt_versions` | 2+Emb | バージョン + 埋め込みベクトル |
| `execution_logs` | 3 | 実行ログ |
| `evaluations` | 3 | 評価 |
| `consulting_sessions` | 6 | コンサルセッション |
| `consulting_messages` | 6 | コンサルメッセージ |
| `industry_configs` | 6 | 業界設定 |
| `tags` | 7 | タグ |
| `prompt_tags` | 7 | プロンプト-タグ関連 |

### API エンドポイント数

- **全エンドポイント数**: 約 50
- **認証**: Bearer トークン（全 `/api/v1/` エンドポイント）
- **ヘルスチェック**: `GET /health`

---

## 技術的判断

### pgvector → native real[] への変更
PostgreSQL 18 標準イメージに pgvector 拡張がバンドルされていないため、`real[]` カラム + PL/pgSQL `cosine_similarity()` 関数で代替。HNSW インデックスは使用不可だが、部分インデックス (`WHERE embedding IS NOT NULL`) でフィルタリング性能を確保。

### 非同期埋め込み生成
バージョン作成時の応答速度を優先し、`EmbedVersionAsync()` で goroutine による fire-and-forget パターンを採用。埋め込み生成失敗はログ出力のみ。

### 環境ベース Feature Toggle
`EMBEDDING_URL` 環境変数の有無で埋め込みサービスの有効/無効を切り替え。未設定時は全セマンティック検索機能が noop。

---

## 追加実装済み

| 項目 | Phase | ステータス | 備考 |
|------|-------|-----------|------|
| SSE ストリーミング | 6 | ✅ 完了 | `sse.go` + `stream.go`、Keepalive Ping |
| RAG パイプライン | 6 | ✅ 完了 | Embed → Search → Context → Ollama Stream |
| Intent Classifier | 6 | ✅ 完了 | `intentservice/` ルールベース分類（EN/JP） |
| Python SDK | 3 | ✅ 完了 | `sdks/python/` |
| TypeScript SDK | 3 | ✅ 完了 | `sdks/typescript/` |
| RBAC ミドルウェア | 1 | ✅ 完了 | `middleware/rbac.go` owner > admin > member > viewer |
| レート制限 | 1 | ✅ 完了 | `middleware/ratelimit.go` Token Bucket (60 req/min) |
| CORS ミドルウェア | 1 | ✅ 完了 | `middleware/cors.go` CORS_ORIGINS 環境変数 |
| Redis キャッシュ | - | ✅ 完了 | `pkgs/cache/` nil-safe、Diff 結果キャッシュ (24h TTL) |
| User ドメイン | 1 | ✅ 完了 | `domain/user/` + `routes/users/` + テナントMW |
| Industry ドメイン | 6 | ✅ 完了 | `domain/industry/` 値オブジェクト |
| Web UI 拡充 | 7 | ✅ 完了 | 検索、設定、Lint/Diff 表示、Analytics API連携 |

## 未実装・今後の拡張

| 項目 | Phase | 優先度 | 備考 |
|------|-------|--------|------|
| Cognito/JWT 認証 | 1 | 高 | 本番認証（AWS Cognito 依存、現在は開発用 Bearer） |
| Citations 抽出 | 6 | 中 | RAG レスポンスからの引用元抽出 |
| ActionsTaken | 6 | 中 | チャットからのバージョン自動作成 |
| LLM Lint ルール | 4 | 中 | missing-constraints, prompt-injection-risk |
| 統計的有意性テスト | 5 | 低 | Analytics の A/B テスト有意差検定 |
| カスタム Lint ルール | 4 | 低 | ユーザー定義ルール |
| pgvector 移行 | Emb | 低 | HNSW インデックスによる高速検索 |
| バッチ集計ジョブ | 7 | 低 | 月次ベンチマーク集計 |
