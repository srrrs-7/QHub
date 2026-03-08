# go-fp-devcontainer

## CI/CD

### CD Pipeline

CD パイプラインは GitHub Actions で実行され、以下のフローでデプロイします:

```
Database Migration → Build & Push to ECR → Deploy to ECS
```

#### トリガー

| イベント | デプロイ先 |
|---------|-----------|
| Push to `main` | dev |
| Manual (workflow_dispatch) | dev / stg / prd から選択 |

#### 手動デプロイ

1. GitHub リポジトリの **Actions** タブを開く
2. 左メニューから **CD** を選択
3. **Run workflow** をクリック
4. 環境を選択して **Run workflow** を実行

### GitHub Environments 設定

CD パイプラインを動作させるために、GitHub Environments の設定が必要です。

#### 1. Environments の作成

1. リポジトリの **Settings** → **Environments** に移動
2. **New environment** をクリック
3. 以下の3つの環境を作成:
   - `dev`
   - `stg`
   - `prd`

#### 2. Environment Variables の設定

各環境で **Add environment variable** から以下を設定:

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

#### 3. 本番環境 (prd) の保護設定 (推奨)

`prd` 環境には以下の保護ルールを設定することを推奨:

1. **Settings** → **Environments** → **prd** を開く
2. **Deployment protection rules** セクションで設定:
   - **Required reviewers**: 承認者を追加（デプロイ前に承認が必要）
   - **Wait timer**: デプロイまでの待機時間（例: 5分）
3. **Deployment branches and tags** で `main` ブランチのみに制限

### 必要なファイル

| ファイル | 説明 |
|---------|------|
| `.aws/task-definition-api.json` | ECS タスク定義テンプレート |

### AWS 側の設定

#### OIDC Provider

GitHub Actions から AWS にアクセスするために、IAM OIDC Provider を作成:

```
Provider URL: https://token.actions.githubusercontent.com
Audience: sts.amazonaws.com
```

#### IAM Role

OIDC 用の IAM Role に必要な権限:

- `ecr:GetAuthorizationToken`
- `ecr:BatchCheckLayerAvailability`
- `ecr:GetDownloadUrlForLayer`
- `ecr:BatchGetImage`
- `ecr:PutImage`
- `ecr:InitiateLayerUpload`
- `ecr:UploadLayerPart`
- `ecr:CompleteLayerUpload`
- `ecs:DescribeTaskDefinition`
- `ecs:RegisterTaskDefinition`
- `ecs:UpdateService`
- `ecs:DescribeServices`
- `rds-db:connect` (IAM Database Authentication 用)
- `iam:PassRole` (ECS タスクロール用)

Trust Policy 例:

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
# QHub
# QHub
