# QHub

プロンプトとバージョンを管理するシステム。Go モノレポ + PostgreSQL + templ/HTMX。

## アーキテクチャ

```
apps/
  api/    → バックエンド API (Go, chi router, :8080)
  web/    → フロントエンド (templ + HTMX, M3 Design, :3000)
  cli/    → CLI ツール「qhub」(Cobra)
  pkgs/   → 共有パッケージ (db, env, logger, testutil)
  iac/    → Terraform インフラ (AWS: ECS, Aurora, Cognito, CloudFront, WAF)
  migrate/ → マイグレーション用コンテナ
```

**クリーンアーキテクチャ**: `routes → domain ← infra`（ドメイン層は外部依存なし）

**技術スタック**:
- Go 1.26, Go Workspaces
- PostgreSQL 18, Redis, ElasticMQ
- Atlas (スキーマファーストマイグレーション) + sqlc (型安全クエリ生成)
- templ + HTMX (SSR フロントエンド、JS 不要)
- Docker Compose (ローカル開発)
- Dev Containers (開発環境)

## セットアップ

### Dev Container (推奨)

VS Code の Dev Containers 拡張を使用:

```bash
cp .devcontainer/compose.override.yaml.example .devcontainer/compose.override.yaml
# VS Code: "Reopen in Container"
```

### ローカル開発

```bash
# サービス起動
docker compose up -d          # 全サービス (api, web, db, cache, queue)
docker compose up -d db       # DB のみ

# サーバー起動
make run-api                  # API サーバー (:8080)
make run-web                  # Web サーバー (:3000, 要 make templ-gen)
make run-all                  # マイグレーション + API + Web

# CLI
make build-cli                # bin/qhub にビルド
make run-cli ARGS="prompt list --project <id>"
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
make atlas-apply              # マイグレーション適用
make atlas-status             # マイグレーション状態確認
make sqlc-gen                 # SQL クエリから Go コード生成

# フロントエンド
make templ-gen                # .templ → Go コード生成
make templ-watch              # ファイル監視モード

# Terraform
make tf-fmt                   # .tf ファイルフォーマット
```

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
