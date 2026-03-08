# QHub

プロンプトとバージョンを管理し、実行ログ・評価・コンサルティング・セマンティック検索まで提供するプラットフォーム。

## アーキテクチャ

```
apps/
  api/      → バックエンド API (Go, chi router, :8080)
  web/      → フロントエンド (templ + HTMX, M3 Design, :3000)
  cli/      → CLI ツール「qhub」(Cobra)
  sdk/      → Go SDK クライアント
  pkgs/     → 共有パッケージ (db, env, logger, testutil)
  iac/      → Terraform インフラ (AWS: ECS, Aurora, Cognito, CloudFront, WAF)
  migrate/  → マイグレーション用コンテナ
```

**クリーンアーキテクチャ**: `routes → domain ← infra`, `routes → services → domain`

### サービス層 (`apps/api/src/services/`)

ハンドラーとリポジトリの間で、複雑なビジネスロジックを担当するサービス群:

| サービス | 説明 |
|---------|------|
| **diffservice** | プロンプトバージョン間の差分分析。**セマンティック差分** (文字数変化、変数の追加/削除、トーンシフト検出、具体性変化) と **テキスト差分** (LCS アルゴリズムによる行単位の added/removed/equal 比較) の2種類を提供。結果は DB に永続化。 |
| **lintservice** | プロンプトの品質分析。4つのルール (`excessive-length`: 4000文字超過、`missing-output-format`: 出力形式未指定、`variable-check`: 未宣言変数の検出、`no-vague-instruction`: 曖昧な表現の検出) を適用し、severity (error/warning/info) に基づく 0〜100 のスコアを算出。結果は DB に永続化。 |
| **embeddingservice** | プロンプトバージョンのベクトルエンベディング生成・保存。TEI (Text Embeddings Inference, BAAI/bge-m3) を使用。`EMBEDDING_URL` 未設定時は全メソッドが no-op で動作。非同期 (fire-and-forget) と同期の両モードに対応。セマンティック検索のクエリエンベディング生成にも使用。 |
| **contentutil** | 共有ユーティリティ。JSON RawMessage からテキスト抽出、`{{variable}}` パターンの変数検出。 |

**技術スタック**:
- Go 1.26, Go Workspaces (5モジュール: api, pkgs, web, cli, sdk)
- PostgreSQL 18, Redis, ElasticMQ
- Text Embeddings Inference (BAAI/bge-m3) + Ollama (ホストマシン)
- Atlas (スキーマファーストマイグレーション) + sqlc (型安全クエリ生成)
- templ + HTMX (SSR フロントエンド、JS 不要)
- Dev Containers (PostgreSQL, Redis, ElasticMQ, TEI を自動起動)

## セットアップ

### Dev Container (推奨)

VS Code の Dev Containers 拡張を使用:

```bash
cp .devcontainer/compose.override.yaml.example .devcontainer/compose.override.yaml
# VS Code: "Reopen in Container"
```

Dev Container は db, cache, queue, embedding サービスを自動起動します。

### ローカル開発

```bash
# サービス起動
docker compose up -d          # 全サービス (api, web, db, cache, queue)
docker compose up -d db       # DB のみ

# サーバー起動
make run-api                  # API サーバー (:8080)
make run-web                  # Web サーバー (:3000, 要 make templ-gen)
make run-all                  # マイグレーション + API + Web

# CLI / SDK
make build-cli                # bin/qhub にビルド
make run-cli ARGS="prompt list --project <id>"
```

### Ollama (ホストマシン)

セマンティック検索やエンベディングに Ollama を使用:

```bash
make ollama-health            # 接続確認
make ollama-models            # モデル一覧
make ollama-pull-embed        # エンベディングモデル取得
make ollama-embed TEXT="hello" # エンベディング生成
```

## 開発コマンド

```bash
# 品質チェック
make check                    # fmt + vet + lint + cspell + test (CI と同等)
make test                     # 全モジュールテスト (要 DB, atlas-apply 自動実行)
make fmt                      # go fmt
make vet                      # 静的解析
make lint                     # golangci-lint

# 単体テスト実行
cd apps/api && go test -run TestGetHandler ./src/routes/tasks/

# カバレッジ
cd apps/api && go test -coverprofile=c.out ./... && go tool cover -func=c.out

# データベース
make atlas-diff NAME=<name>   # スキーマ差分からマイグレーション生成
make atlas-apply              # マイグレーション適用 (ATLAS_ENV=docker で Docker 環境)
make atlas-status             # マイグレーション状態確認
make sqlc-gen                 # SQL クエリから Go コード生成

# フロントエンド
make templ-gen                # .templ → Go コード生成
make templ-watch              # ファイル監視モード

# Terraform
make tf-fmt                   # .tf ファイルフォーマット
```

## API 概要

全エンドポイントは `/api/v1` プレフィックス、Bearer 認証必須 (`/health` を除く)。

### コアリソース

| リソース | エンドポイント | 説明 |
|----------|--------------|------|
| Organization | `/organizations` | 組織管理 |
| Project | `/organizations/{org_id}/projects` | プロジェクト管理 (組織配下) |
| Prompt | `/projects/{project_id}/prompts` | プロンプト管理 (プロジェクト配下) |
| Version | `/prompts/{prompt_id}/versions` | バージョン管理 (ステータス変更、lint、diff) |
| Tag | `/tags`, `/prompts/{prompt_id}/tags` | タグ管理・プロンプトへのタグ付け |

### 実行・評価

| リソース | エンドポイント | 説明 |
|----------|--------------|------|
| Execution Log | `/logs`, `/logs/batch` | プロンプト実行ログ (バッチ登録対応) |
| Evaluation | `/evaluations`, `/logs/{log_id}/evaluations` | 実行結果の評価 |

### コンサルティング・業界設定

| リソース | エンドポイント | 説明 |
|----------|--------------|------|
| Consulting | `/consulting/sessions`, `.../messages` | コンサルティングセッション・メッセージ |
| Industry | `/industries` | 業界設定・ベンチマーク・コンプライアンスチェック |

### インテリジェンス

| リソース | エンドポイント | 説明 |
|----------|--------------|------|
| Semantic Diff | `/prompts/{id}/semantic-diff/{v1}/{v2}` | バージョン間のセマンティック差分 |
| Text Diff | `/prompts/{id}/versions/{v}/text-diff` | テキストベースの差分 |
| Lint | `/prompts/{id}/versions/{v}/lint` | プロンプト品質スコア (0-100) |
| Analytics | `/prompts/{id}/analytics`, `/projects/{id}/analytics` | 利用統計・トレンド |
| Search | `/search/semantic` | セマンティック検索 |

### 組織管理

| リソース | エンドポイント | 説明 |
|----------|--------------|------|
| Member | `/organizations/{org_id}/members` | メンバー管理 (招待・ロール変更) |
| API Key | `/organizations/{org_id}/api-keys` | APIキー管理 |

## CI/CD

### CI

GitHub Actions で Dev Container 内で実行 (push/PR to main):

```
make vet → make atlas-apply → make test
```

### CD

| ワークフロー | トリガー | 説明 |
|-------------|---------|------|
| `cd-api.yml` | 手動 (workflow_dispatch) | CI 確認 → DB マイグレーション → API デプロイ (ECS) |
| `cd-web.yml` | 手動 (workflow_dispatch) | Web フロントエンドデプロイ |
| `cd-migrate.yml` | cd-api から呼出 | データベースマイグレーション実行 |

**認証**: AWS OIDC (長期クレデンシャル不要) + RDS IAM Authentication

**環境**: `dev` / `stg` / `prd` (GitHub Environments で管理)

```
Database Migration → Build & Push to ECR → Deploy to ECS
```

### GitHub Environments 設定

リポジトリの **Settings** → **Environments** で `dev`, `stg`, `prd` を作成し、以下の変数を設定:

| Variable | 説明 | 例 |
|----------|------|-----|
| `AWS_REGION` | AWS リージョン | `ap-northeast-1` |
| `AWS_ROLE_ARN` | OIDC 用 IAM Role ARN | `arn:aws:iam::123456789012:role/github-actions-role` |
| `DB_HOST` | RDS エンドポイント | `test.xxxx.ap-northeast-1.rds.amazonaws.com` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_DBNAME` | データベース名 | `myapp` |
| `DB_USERNAME` | IAM 認証用ユーザー名 | `app_user` |
| `ECR_REPOSITORY_API` | ECR リポジトリ名 | `myapp-api` |
| `CONTAINER_NAME_API` | タスク定義内のコンテナ名 | `api` |
| `ECS_SERVICE_API` | ECS サービス名 | `myapp-api-service` |
| `ECS_CLUSTER` | ECS クラスター名 | `myapp-cluster` |

**本番環境 (prd) の保護** (推奨):
- Required reviewers (承認者設定)
- Wait timer (待機時間)
- Deployment branches: `main` のみ

### AWS 設定

#### OIDC Provider

```
Provider URL: https://token.actions.githubusercontent.com
Audience: sts.amazonaws.com
```

#### IAM Role 権限

- ECR: `GetAuthorizationToken`, `BatchCheckLayerAvailability`, `GetDownloadUrlForLayer`, `BatchGetImage`, `PutImage`, `InitiateLayerUpload`, `UploadLayerPart`, `CompleteLayerUpload`
- ECS: `DescribeTaskDefinition`, `RegisterTaskDefinition`, `UpdateService`, `DescribeServices`
- RDS: `rds-db:connect` (IAM Database Authentication)
- IAM: `PassRole` (ECS タスクロール用)

<details>
<summary>Trust Policy 例</summary>

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:<owner>/<repo>:*"
        }
      }
    }
  ]
}
```

</details>

### 必要なファイル

| ファイル | 説明 |
|---------|------|
| `.aws/task-definition-api.json` | ECS タスク定義テンプレート |
