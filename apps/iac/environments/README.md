# Environments

各環境のTerraform設定ディレクトリです。環境ごとに異なるデフォルト値と構成を持ちます。

## 環境一覧

| 環境 | 用途 | 月額コスト目安 |
|------|------|---------------|
| **dev** | 開発・検証 | ~$41 |
| **stg** | ステージング・本番前検証 | ~$200-400 |
| **prd** | 本番 | ~$500-2000+ |

## 環境別構成比較

### ネットワーク (VPC)

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| VPC CIDR | `10.0.0.0/16` | `10.1.0.0/16` | `10.2.0.0/16` |
| AZ数 | 2 | 2 | 3 |
| NAT Gateway | 1 (single) | 1 (single) | 3 (per AZ) |
| VPC Flow Logs | OFF | ON | ON |
| Interface Endpoints | OFF | ON | ON |

### データベース (Aurora/RDS)

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| タイプ | RDS db.t4g.micro | Aurora Serverless v2 | Aurora Serverless v2 |
| インスタンス数 | 1 | 2 | 3 |
| 最小ACU | - | 0.5 | 2 |
| 最大ACU | - | 8 | 64 |
| バックアップ保持 | 1日 | 14日 | 35日 |
| 削除保護 | OFF | ON | ON |
| 最終スナップショット | スキップ | 作成 | 作成 |

### コンピューティング (ECS Fargate)

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| CPU | 256 (0.25 vCPU) | 512 (0.5 vCPU) | 1024 (1 vCPU) |
| メモリ | 512 MB | 1024 MB | 2048 MB |
| タスク数 | 1 | 2 | 3 |
| オートスケール | OFF | ON (2-4) | ON (3-20) |
| Fargate Spot | ON | OFF | OFF |
| ECS Exec | ON | ON | OFF |

### セキュリティ

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| WAF | OFF | ON | ON |
| WAF Rate Limit | - | 2000 req/5min | 5000 req/5min |
| ALB 削除保護 | OFF | ON | ON |
| Cognito MFA | OPTIONAL | OPTIONAL | ON (必須) |
| Cognito 高度なセキュリティ | OFF | AUDIT | ENFORCED |
| Cognito 削除保護 | OFF | ON | ON |

### CDN (CloudFront)

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| Price Class | PriceClass_200 | PriceClass_200 | PriceClass_All |
| ドメイン | なし (ALB直接) | 必須 | 必須 |

### Terraform Backend

| 設定 | dev | stg | prd |
|------|-----|-----|-----|
| Backend | ローカル | S3 | S3 |
| 状態ロック | なし | DynamoDB (推奨) | DynamoDB (必須) |

## 使用方法

### dev環境

```bash
cd apps/iac/environments/dev
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvars を編集

terraform init
terraform plan
terraform apply
```

### stg/prd環境

```bash
cd apps/iac/environments/stg  # または prd
cp backend.hcl.example backend.hcl
cp terraform.tfvars.example terraform.tfvars
# 両方のファイルを編集

terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

## アーキテクチャ図

```
                                    ┌─────────────────────────────────────┐
                                    │            Internet                 │
                                    └──────────────────┬──────────────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        ▼                        │
                              │    ┌─────────────────────────────────────┐      │
                              │    │   CloudFront (stg/prd only)         │      │
                              │    │   + WAF (stg/prd only)              │      │
                              │    └──────────────────┬──────────────────┘      │
                              │                       │                         │
                              │    ┌──────────────────┼──────────────────┐      │
                              │    │   Public Subnets                    │      │
                              │    │  ┌──────────────┴──────────────┐    │      │
                              │    │  │     ALB (all environments)  │    │      │
                              │    │  └──────────────┬──────────────┘    │      │
                              │    │                 │                   │      │
                              │    │  ┌──────────────┴──────────────┐    │      │
                              │    │  │      NAT Gateway            │    │      │
                              │    │  │  dev: 1, stg: 1, prd: 3    │    │      │
                              │    │  └──────────────┬──────────────┘    │      │
                              │    └─────────────────┼───────────────────┘      │
                              │                      │                          │
                              │    ┌─────────────────┼───────────────────┐      │
                              │    │   Private Subnets                   │      │
                              │    │  ┌──────────────┴──────────────┐    │      │
                              │    │  │     ECS Fargate             │    │      │
                              │    │  │  dev: 1 task (Spot)         │    │      │
                              │    │  │  stg: 2-4 tasks             │    │      │
                              │    │  │  prd: 3-20 tasks            │    │      │
                              │    │  └──────────────┬──────────────┘    │      │
                              │    └─────────────────┼───────────────────┘      │
                              │                      │                          │
                              │    ┌─────────────────┼───────────────────┐      │
                              │    │   Database Subnets                  │      │
                              │    │  ┌──────────────┴──────────────┐    │      │
                              │    │  │   dev: RDS db.t4g.micro     │    │      │
                              │    │  │   stg: Aurora (2 instances) │    │      │
                              │    │  │   prd: Aurora (3 instances) │    │      │
                              │    │  └─────────────────────────────┘    │      │
                              │    └─────────────────────────────────────┘      │
                              │                                                 │
                              │                    VPC                          │
                              └─────────────────────────────────────────────────┘
```

## コスト最適化のポイント

### dev環境 (~$41/月)

- **RDS db.t4g.micro**: Aurora の代わりに最小RDSインスタンス
- **Fargate Spot**: 最大70%のコスト削減（中断の可能性あり）
- **Interface Endpoints無効**: NAT Gateway経由でAWSサービスにアクセス
- **VPC Flow Logs無効**: ログ収集コスト削減
- **WAF無効**: セキュリティコスト削減
- **ローカルBackend**: S3バケット不要

### stg環境

- **Single NAT Gateway**: AZ冗長性を犠牲にコスト削減
- **Aurora最小ACU 0.5**: アイドル時のコスト最小化

### prd環境

- **フル冗長構成**: 3 AZ、複数インスタンス
- **高可用性優先**: コストよりも可用性を重視

## 環境間の昇格フロー

```
dev (開発) → stg (検証) → prd (本番)
     │            │            │
     │            │            └── 本番トラフィック
     │            └── 本番同等の構成でテスト
     └── 機能開発・単体テスト
```

## 注意事項

1. **OIDC Provider**: 同一AWSアカウント内で1つのみ作成。dev環境で作成し、stg/prdでは既存ARNを参照。

2. **ドメイン**: stg/prd環境ではドメイン設定が必須。dev環境はALB直接アクセス可能。

3. **削除保護**: stg/prd環境では各種リソースの削除保護が有効。意図しない削除を防止。

4. **ECS Exec**: prd環境では無効。デバッグが必要な場合は一時的に有効化。
