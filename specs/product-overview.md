# Product Overview

## Product Name

PromptLab (仮称)

## Vision

LLM プロンプトのバージョン管理・品質トラッキング・改善サイクルを一元化し、
蓄積されたナレッジを業界特化コンサルチャットとして提供する SaaS プラットフォーム。

## Delivery Model

- **OSS コア**: バージョン管理、ログ ingestion、基本評価、CLI/SDK — セルフホスト可能
- **Cloud SaaS**: Semantic Diff、Prompt Linting、Consulting Chat、チーム機能 — 有料

## Target Users

| ペルソナ | ニーズ |
|---------|--------|
| LLM アプリ開発者 | プロンプトのバージョン管理、回帰テスト、品質メトリクス |
| プロダクトマネージャー | 品質ダッシュボード、改善トレンドの可視化、業界ベンチマーク |
| ドメインエキスパート | 業界コンプライアンスチェック、業界テンプレート |

## Feature Map

### Layer 1: Prompt Management Platform (基盤)

#### F1. Prompt Version Control (プロンプトバージョン管理)

- プロンプトを作成・編集し、変更ごとにイミュータブルなバージョンを自動生成
- system prompt / user prompt template / combined に対応
- テンプレート変数（`{{variable}}`）のサポート
- Semantic Diff: テキスト差分 + LLM による意味的変更サマリー

#### F2. Prompt Lifecycle (ライフサイクル管理)

- Draft → Review → Production → Archived の状態遷移
- Production バージョンのみが SDK から取得可能（デフォルト）
- ロールバック（過去バージョンを Production に再昇格）

#### F3. Execution Log Ingestion (実行ログ取り込み)

- 外部システムから SDK / API 経由でログを送信・保存
- SDK: Go, Python, TypeScript を提供
- 各ログは使用されたプロンプトバージョンと紐付け
- メタデータ（モデル名、トークン数、レイテンシ、コスト等）を記録

#### F4. Quality Evaluation (品質評価)

- 実行ログに対してスコアリング（手動 / 自動）
- 評価基準をカスタマイズ可能（正確性、有用性、安全性など）
- バージョン単位で品質メトリクスを集計

#### F5. Prompt Linting (プロンプト静的解析)

- ルールベース: 変数未使用/未定義、出力形式未指定、長さ超過
- LLM ベース: 曖昧な指示、矛盾する制約の検出
- カスタムルール定義

#### F6. Analytics Dashboard (分析ダッシュボード)

- バージョン間の品質比較
- 時系列トレンド（品質、コスト、レイテンシ）
- コスト分析（トークン使用量、API コスト推定）
- 統計的有意差検定

### Layer 2: Industry Consulting Chat (業界特化コンサル)

#### F7. Consulting Chat (コンサルチャット)

- 組織のプロンプトデータに基づくパーソナライズされたアドバイス
- プロンプト改善提案の生成
- 改善提案からの直接バージョン作成

#### F8. Industry Knowledge Base (業界ナレッジベース)

- 業界別のベストプラクティス、規制要件、テンプレート
- 対応業界: ヘルスケア、法律、金融、カスタマーサポート、教育、EC/小売
- カスタムナレッジの追加

#### F9. Compliance Check (コンプライアンスチェック)

- 業界規制に対するプロンプトの準拠チェック
- HIPAA、GDPR、金融規制等のルールセット
- 準拠率スコアと改善提案

#### F10. Benchmark Comparison (ベンチマーク比較)

- 匿名化されたクロス組織の統計（オプトイン）
- 業界別の品質ベンチマーク
- 自組織 vs 業界中央値の比較

### Shared: Collaboration (コラボレーション)

#### F11. Multi-Tenancy (マルチテナント)

- Organization（チーム）/ 個人
- ロールベースアクセス制御: Owner / Admin / Member / Viewer

#### F12. SDK & API

- Go / Python / TypeScript SDK
- API Key 認証（SDK 用）+ JWT 認証（Web UI 用）
- レート制限

## Non-Goals (MVP スコープ外)

- LLM API の直接プロキシ（Portkey のようなゲートウェイ機能）
- リアルタイム A/B テスト自動振り分け
- Prompt Registry（公開マーケットプレイス）— Phase 3 以降
- モバイルアプリ
- 自動プロンプト最適化（AI が自動でプロンプトを書き換える機能）

## Technical Constraints

- Go monorepo（Go 1.25, workspaces）
- PostgreSQL 18 + Atlas migrations + sqlc
- templ + HTMX フロントエンド
- AWS インフラ（ECS Fargate, Aurora Serverless v2）
- TDD 必須（カバレッジ 80% 以上）
- OSS コアは Apache 2.0 or MIT ライセンス
