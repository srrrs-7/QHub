# Positioning Strategy

## Positioning Statement

**PromptLab は、プロンプトのバージョン管理と品質評価を基盤に、
蓄積データと業界知識を組み合わせたコンサルチャットで
プロンプト改善サイクルを加速する OSS ベースの SaaS プラットフォーム。**

---

## 戦略: OSS コア + 日本市場先行 + データフライホイール

### Why OSS

| 理由 | 説明 |
|------|------|
| 市場の現実 | Langfuse (20k stars), Agenta, Pezzo, Promptfoo が OSS。プロプラのみでは採用されない |
| コミュニティ | OSS でコントリビューター・アーリーアダプターを獲得 |
| 信頼性 | セルフホスト可能 = データの安全性。エンタープライズの導入障壁を下げる |
| 差別化 | OSS コアは無料。**LLM を使う機能 (Semantic Diff, Lint, Chat) を有料**にすることで明確な境界 |

### Why 日本市場先行

| 理由 | 説明 |
|------|------|
| 競合ゼロ | 主要競合は全て英語圏。日本語ファーストのツールは存在しない |
| 参入障壁 | ローカライゼーション（UI, ドキュメント, サポート）が自然な参入障壁 |
| PMF 検証 | 小さい市場で素早く PMF を検証し、グローバル展開の基盤にする |
| 業界特化 | 日本の業界特有の規制・慣習をナレッジベース化することで独自の価値 |

### Why データフライホイール

| 理由 | 説明 |
|------|------|
| 参入障壁 | データが溜まるほどコンサルが賢くなり、後発が追いつけない |
| ネットワーク効果 | 匿名ベンチマーク参加組織が増えるほど、全員にとっての価値が上がる |
| LTV 向上 | コンサルに依存するほど解約しにくくなる |

---

## OSS / Cloud の機能境界

```
┌─────────────────────────────────┬────────────────────────────────┐
│          OSS Core               │         Cloud (SaaS)           │
│         (Apache 2.0)            │         (Proprietary)          │
│         無料                    │         有料                    │
├─────────────────────────────────┼────────────────────────────────┤
│ Layer 1 基盤:                   │ Layer 1 拡張 (LLM 利用):       │
│ ・Prompt CRUD + Versioning      │ ・Semantic Diff                │
│ ・Lifecycle (draft→prod)        │ ・Prompt Linting (LLM)         │
│ ・Execution Log Ingestion       │ ・Advanced Analytics           │
│ ・Basic Evaluation (manual)     │                                │
│ ・Basic Analytics               │ Layer 2 全体:                  │
│ ・Text Diff                     │ ・Consulting Chat              │
│ ・CLI + SDK (Go/Py/TS)          │ ・Industry Knowledge Base      │
│ ・Basic Web UI                  │ ・Compliance Check             │
│ ・API Key / Bearer Auth         │ ・Platform Benchmarks          │
│ ・Docker Compose self-host      │                                │
│                                 │ Enterprise:                    │
│                                 │ ・SSO / SAML                   │
│                                 │ ・Audit Log                    │
│                                 │ ・SLA                          │
│                                 │ ・Managed Hosting              │
└─────────────────────────────────┴────────────────────────────────┘
```

**境界の原則**: LLM を使う機能 = 有料。LLM を使わない機能 = OSS。
（LLM 利用にはコストがかかるため、自然な課金境界になる）

---

## 価格戦略

| Plan | 価格 | 対象 | 含まれる機能 |
|------|------|------|-------------|
| **OSS** | 無料 | 個人/実験 | Layer 1 基盤全機能。セルフホスト |
| **Pro** | $20/user/月 | 小規模チーム | OSS + Semantic Diff + Lint + チャット (50 sessions/day) |
| **Team** | $50/user/月 | 中規模チーム | Pro + Industry KB + Compliance + Benchmarks + 高レート |
| **Enterprise** | 要問合せ | 大企業 | Team + SSO + Audit + SLA + セルフホストサポート |

**競合比較**:
| | PromptLab | Braintrust | PromptLayer | Langfuse Cloud |
|---|---|---|---|---|
| 無料プラン | OSS (無制限) | 5 users, 1M spans | 5 users, 2.5K req | 50K units |
| 有料開始 | $20/user | $249/月 (5 users) | $50/user | ~$50/月〜 |
| Consulting Chat | o | x | x | x |

---

## Go-to-Market

### Phase 1: OSS + 開発者コミュニティ (Month 1-3)

**目標**: GitHub Stars 500+、初期ユーザー 50 名

1. OSS として GitHub に公開 (Layer 1 コア)
2. Go SDK を先行リリース
3. 技術ブログ:
   - 「プロンプトの Semantic Diff を Go で実装した」
   - 「なぜ自然言語にはコード管理ツールが使えないのか」
4. Zenn / connpass / Twitter で日本の LLM 開発者にリーチ

### Phase 2: 早期ユーザー + コンサルチャット (Month 4-6)

**目標**: 有料ユーザー 10 社、NPS 40+

1. 日本の AI スタートアップ 5-10 社に Pro プランを無料提供
2. コンサルチャット (Layer 2) をベータリリース
3. 業界ナレッジ: まず customer_support と healthcare
4. ケーススタディ作成
5. Product Hunt ローンチ

### Phase 3: 有料化 + グローバル (Month 7-12)

**目標**: ARR $100K、英語圏ユーザー開始

1. Pro / Team プランを正式リリース
2. Python / TypeScript SDK リリース
3. 英語 UI / ドキュメント追加
4. Compliance Check (HIPAA, GDPR) で海外ニーズに対応
5. Platform Benchmarks 開始（オプトイン組織 20+ で統計的意味が出る）

---

## リスクと対策

| リスク | 対策 |
|--------|------|
| Langfuse がコンサル機能を追加 | データフライホイールで先行。業界特化の深さで差別化 |
| Anthropic/OpenAI がコンソールに類似機能内蔵 | OSS + セルフホストで「ベンダーロックイン回避」を訴求 |
| OSS のメンテナンス負荷 | コア機能に絞り、LLM 依存機能は Cloud のみ |
| 日本市場が小さすぎる | Phase 3 でグローバル展開。日本は PMF 検証の場 |
| コンサルチャットの品質 | Citations 必須でハルシネーション防止。評価スコア付きで品質管理 |
