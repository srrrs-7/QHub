# Implementation Roadmap

## Phase Overview

| Phase | Focus | 内容 |
|-------|-------|------|
| 1 | Foundation | 認証、組織、ユーザー、API Key |
| 2 | Core | プロンプト、バージョン、ライフサイクル |
| 3 | Ingestion | 実行ログ、評価、SDK |
| 4 | Intelligence | Semantic Diff、Prompt Linting |
| 5 | Analytics | ダッシュボード、バージョン比較 |
| 6 | Consulting | チャットエンジン、業界ナレッジ |
| 7 | Polish | コンプライアンス、ベンチマーク、Web UI |

---

## Phase 1: Foundation

### 1.1 DB Migration
- [ ] organizations, users, organization_members, api_keys テーブル
- [ ] sqlc クエリ定義

### 1.2 Domain Layer
- [ ] Organization (値オブジェクト、エンティティ、Repository interface)
- [ ] User (値オブジェクト、エンティティ、Repository interface)
- [ ] MemberRole enum
- [ ] ApiKey ドメイン (hash, prefix, validation)

### 1.3 Infrastructure
- [ ] OrganizationRepository 実装
- [ ] UserRepository 実装
- [ ] ApiKey hashing / lookup

### 1.4 Middleware
- [ ] API Key 認証ミドルウェア
- [ ] テナントコンテキストミドルウェア
- [ ] RBAC ミドルウェア
- [ ] レート制限ミドルウェア

### 1.5 Routes
- [ ] Organization CRUD ハンドラー
- [ ] Member 管理ハンドラー
- [ ] API Key 管理ハンドラー

---

## Phase 2: Core — Prompt & Version

### 2.1 DB Migration
- [ ] projects, prompts, prompt_versions テーブル

### 2.2 Domain Layer
- [ ] Project ドメイン
- [ ] Prompt ドメイン (PromptType, PromptContent, PromptVariables)
- [ ] PromptVersion ドメイン (イミュータブル、Status 遷移)
- [ ] バージョン自動インクリメントロジック
- [ ] ライフサイクル遷移ルール (draft→review→production→archived)

### 2.3 Infrastructure & Routes
- [ ] Project CRUD
- [ ] Prompt CRUD
- [ ] PromptVersion 作成・一覧・取得
- [ ] ステータス変更 API
- [ ] テキスト Diff API

---

## Phase 3: Ingestion — Log & Evaluation

### 3.1 DB Migration
- [ ] execution_logs, evaluations テーブル

### 3.2 Domain & Infrastructure
- [ ] ExecutionLog ドメイン (TokenUsage 値オブジェクト)
- [ ] Evaluation ドメイン
- [ ] ログ ingestion ハンドラー (POST /logs)
- [ ] バッチ ingestion ハンドラー (POST /logs/batch)
- [ ] 評価 CRUD

### 3.3 SDK
- [ ] Go SDK: Client, Log(), GetPromptLatest()
- [ ] Python SDK: 同上
- [ ] TypeScript SDK: 同上

---

## Phase 4: Intelligence

### 4.1 Semantic Diff
- [ ] LLM クライアント (Anthropic API)
- [ ] SemanticDiff 生成サービス
- [ ] キャッシュ戦略 (Redis + DB)
- [ ] GET /prompts/{id}/semantic-diff/{v1}...{v2} エンドポイント

### 4.2 Prompt Linting
- [ ] ルールベース lint エンジン
  - [ ] variable-unused / variable-undefined
  - [ ] define-output-format
  - [ ] excessive-length
  - [ ] language-consistency
- [ ] LLM ベース lint
  - [ ] no-vague-instruction
  - [ ] missing-constraints
  - [ ] prompt-injection-risk
- [ ] GET /prompts/{id}/versions/{v}/lint エンドポイント

---

## Phase 5: Analytics

### 5.1 Analytics Queries
- [ ] バージョン別集計 SQL (executions, avg_score, avg_latency, tokens)
- [ ] 時系列トレンド SQL (daily/weekly/monthly)
- [ ] コスト分析 SQL (token 使用量 → 推定コスト)

### 5.2 API
- [ ] GET /prompts/{id}/analytics
- [ ] GET /prompts/{id}/versions/{v}/analytics
- [ ] GET /projects/{id}/analytics

### 5.3 Web UI (templ + HTMX)
- [ ] ダッシュボード画面
- [ ] バージョン比較画面
- [ ] トレンドグラフ

---

## Phase 6: Consulting Chat

### 6.1 DB Migration
- [ ] industry_configs, consulting_sessions, consulting_messages, platform_benchmarks

### 6.2 Chat Engine
- [ ] Knowledge Retriever (Org Data + Industry KB)
- [ ] Intent Classifier
- [ ] LLM オーケストレーション (RAG パイプライン)
- [ ] SSE ストリーミングレスポンス
- [ ] Citations 抽出・構造化
- [ ] ActionsTaken (チャットからのバージョン作成)

### 6.3 Industry Knowledge Base
- [ ] 業界別ナレッジ JSON (初期: healthcare, customer_support)
- [ ] IndustryConfig 管理 API
- [ ] CustomKnowledge 追加機能

### 6.4 API & UI
- [ ] セッション CRUD
- [ ] メッセージ送信 (SSE)
- [ ] チャット UI (templ + HTMX + SSE)

---

## Phase 7: Polish

### 7.1 Compliance Check
- [ ] 業界別ルールセット (HIPAA, GDPR, 金融規制)
- [ ] POST /industries/{industry}/compliance-check

### 7.2 Platform Benchmarks
- [ ] バッチ集計ジョブ (月次)
- [ ] オプトイン管理
- [ ] GET /industries/{industry}/benchmarks

### 7.3 Web UI
- [ ] プロンプト管理画面 (一覧、詳細、エディタ)
- [ ] ログ一覧・詳細画面
- [ ] 組織設定画面
- [ ] Semantic Diff 表示コンポーネント
- [ ] Lint 結果表示

### 7.4 Tags
- [ ] tags, prompt_tags テーブル
- [ ] タグ CRUD + タグ付け API

---

## MVP Scope (Phase 1-3 + 4 partial + 6 basic)

MVP で最低限必要な機能:

**Layer 1 (必須)**:
- Organization + User + API Key (Phase 1)
- Prompt + Version + Lifecycle (Phase 2)
- Execution Log + Evaluation (Phase 3)
- Go SDK (Phase 3)
- 基本テキスト Diff (Phase 2)

**Layer 2 (差別化)**:
- 基本コンサルチャット — 組織データに基づく改善提案 (Phase 6 の一部)
- Semantic Diff (Phase 4.1)

**後回し**:
- Prompt Linting (Phase 4.2)
- Analytics Dashboard (Phase 5)
- Compliance Check (Phase 7.1)
- Platform Benchmarks (Phase 7.2)
- Python/TS SDK (Phase 3 の一部)
- 完全な Web UI (Phase 7.3)
