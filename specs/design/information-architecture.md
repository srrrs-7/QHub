# Information Architecture

## Navigation Model

### 構造

```
PromptLab
├── Organization Switcher (Top App Bar)
├── Navigation Rail (Left)
│   ├── Prompts (default)
│   ├── Logs
│   ├── Analytics
│   ├── Chat (Layer 2)
│   └── Settings
└── Content Area
    ├── Breadcrumb
    ├── Page Header
    └── Main Content
```

### Navigation Rail

M3 Navigation Rail: アイコン + ラベルの縦型ナビゲーション。

```
┌──────────┐
│  [Logo]  │
│          │
│ 📝       │  ← Prompts（デフォルト選択）
│ Prompts  │
│          │
│ 📋       │  ← Execution Logs
│ Logs     │
│          │
│ 📊       │  ← Analytics Dashboard
│ Analytics│
│          │
│ 💬       │  ← Consulting Chat (Layer 2)
│ Chat     │
│          │
│          │
│          │
│ ⚙️       │  ← Settings（下部固定）
│ Settings │
└──────────┘
 80px wide
```

---

## URL 設計

RESTful でリソース階層を反映。HTMX はパーシャルURL (`/partials/*`) を使用。

### ページURL

```
/                                           → ダッシュボード（リダイレクト: /prompts）
/login                                      → ログイン
/orgs/{org_slug}/projects                   → プロジェクト一覧
/orgs/{org_slug}/projects/{project_slug}    → プロジェクト詳細

# Prompt Management
/orgs/{org_slug}/projects/{project_slug}/prompts                    → プロンプト一覧
/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}      → プロンプト詳細（バージョン一覧）
/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}/v/{n}            → バージョン詳細
/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}/v/{n}/diff/{m}   → バージョン差分 (n vs m)

# Execution Logs
/orgs/{org_slug}/logs                       → ログ一覧
/orgs/{org_slug}/logs/{log_id}              → ログ詳細

# Analytics
/orgs/{org_slug}/analytics                  → アナリティクスダッシュボード

# Consulting Chat (Layer 2)
/orgs/{org_slug}/chat                       → チャットセッション一覧
/orgs/{org_slug}/chat/{session_id}          → チャットセッション

# Settings
/orgs/{org_slug}/settings                   → 組織設定
/orgs/{org_slug}/settings/members           → メンバー管理
/orgs/{org_slug}/settings/api-keys          → APIキー管理
```

### HTMX パーシャルURL

```
/partials/projects                                      → プロジェクトリスト
/partials/prompts?project_id={id}                       → プロンプトリスト
/partials/prompts/{prompt_id}/versions                  → バージョンリスト
/partials/prompts/{prompt_id}/versions/{n}              → バージョン詳細パネル
/partials/prompts/{prompt_id}/versions/{n}/diff/{m}     → 差分ビュー
/partials/prompts/{prompt_id}/versions/create           → バージョン作成フォーム結果
/partials/prompts/{prompt_id}/versions/{n}/status       → ステータス更新結果
/partials/logs?page={n}&prompt_id={id}                  → ログリスト（ページネーション）
/partials/chat/{session_id}/messages                    → チャットメッセージ（SSE）
/partials/status                                        → ステータスメッセージ（OOB）
```

---

## Page Hierarchy & Content Structure

### 1. プロンプト一覧ページ（メインビュー）

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
│ [OrgSwitcher ▼]  PromptLab    [Search]  [Avatar ▼]  │
├────┬────────────────────────────────────────────────┤
│    │ Breadcrumb: MyOrg > MyProject > Prompts        │
│ N  │                                                 │
│ a  │ [+ New Prompt]                  [Filter ▼]     │
│ v  │                                                 │
│    │ ┌─────────────────────────────────────────────┐│
│ R  │ │ system-prompt-v2             system  prod   ││
│ a  │ │ Main system prompt           v12 ● live     ││
│ i  │ ├─────────────────────────────────────────────┤│
│ l  │ │ user-onboarding              user    draft  ││
│    │ │ Onboarding flow template     v3 ○ draft     ││
│    │ ├─────────────────────────────────────────────┤│
│    │ │ combined-support             combined prod  ││
│    │ │ Customer support prompt      v8 ● live      ││
│    │ └─────────────────────────────────────────────┘│
│    │                                                 │
│    │ Showing 3 of 12 prompts                        │
└────┴────────────────────────────────────────────────┘
```

### 2. プロンプト詳細ページ（Master-Detail）

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬──────────────────────┬─────────────────────────┤
│    │ system-prompt-v2     │ Version 12 (Production) │
│ N  │                      │                          │
│ a  │ Version History:     │ Status: [● Production]  │
│ v  │                      │ Author: user@example.com│
│    │ v12 ● Production     │ Created: 2026-03-08     │
│ R  │ v11 ■ Archived    ←─│                          │
│ a  │ v10 ■ Archived       │ Content:                │
│ i  │ v9  ■ Archived       │ ┌─────────────────────┐│
│ l  │ v8  ■ Archived       │ │ You are a helpful   ││
│    │ v7  ■ Archived       │ │ assistant that...    ││
│    │ ...                  │ │                      ││
│    │                      │ └─────────────────────┘│
│    │ [+ New Version]      │                          │
│    │                      │ Variables: [topic, lang] │
│    │                      │                          │
│    │                      │ [Diff with v11] [Edit]  │
│    │                      │ [→ Review] [Archive]    │
└────┴──────────────────────┴─────────────────────────┘
```

### 3. バージョン差分ページ

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬────────────────────────────────────────────────┤
│    │ Diff: v11 → v12                                │
│ N  │                                                 │
│ a  │ [Tab: Text Diff] [Tab: Semantic Diff]          │
│ v  │                                                 │
│    │ ┌──────────────────┬──────────────────────────┐│
│ R  │ │ v11 (Archived)   │ v12 (Production)         ││
│ a  │ │                  │                           ││
│ i  │ │ You are a        │ You are a                ││
│ l  │ │-helpful          │+knowledgeable            ││
│    │ │ assistant that   │ assistant that            ││
│    │ │-answers          │+provides detailed         ││
│    │ │-questions.       │+answers with sources.     ││
│    │ │                  │                           ││
│    │ └──────────────────┴──────────────────────────┘│
│    │                                                 │
│    │ Semantic Summary:                               │
│    │ • Tone: neutral → authoritative                │
│    │ • Scope: general → detailed with citations     │
│    │ • Specificity: +42%                            │
└────┴────────────────────────────────────────────────┘
```

### 4. ログ一覧ページ

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬────────────────────────────────────────────────┤
│    │ Execution Logs                                  │
│ N  │                                                 │
│ a  │ [Filter: Prompt ▼] [Filter: Status ▼] [Search] │
│ v  │                                                 │
│    │ ┌──────┬──────────┬────────┬───────┬──────────┐│
│ R  │ │ Time │ Prompt   │Version │Tokens │ Score    ││
│ a  │ ├──────┼──────────┼────────┼───────┼──────────┤│
│ i  │ │12:34 │sys-v2    │ v12    │ 1,234 │ ★★★★☆  ││
│ l  │ │12:33 │user-onb  │ v3     │   456 │ ★★★☆☆  ││
│    │ │12:30 │sys-v2    │ v12    │ 2,100 │ ★★★★★  ││
│    │ │12:28 │combined  │ v8     │   890 │ ★★★★☆  ││
│    │ └──────┴──────────┴────────┴───────┴──────────┘│
│    │                                                 │
│    │ ← 1 2 3 ... 42 →                              │
└────┴────────────────────────────────────────────────┘
```

### 5. アナリティクスダッシュボード

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬────────────────────────────────────────────────┤
│    │ Analytics                [Period: 30d ▼]       │
│ N  │                                                 │
│ a  │ ┌─────────────┐ ┌─────────────┐ ┌────────────┐│
│ v  │ │ Avg Score   │ │ Total Logs  │ │ Avg Cost   ││
│    │ │   4.2 ★     │ │   12,345    │ │  $0.032    ││
│ R  │ │   +8% ↑     │ │  +23% ↑     │ │  -12% ↓   ││
│ a  │ └─────────────┘ └─────────────┘ └────────────┘│
│ i  │                                                 │
│ l  │ ┌──────────────────────────────────────────────┐│
│    │ │ Quality Trend (line chart)                   ││
│    │ │    ___/\___/\                                ││
│    │ │   /         \_                               ││
│    │ │  /            \___                           ││
│    │ └──────────────────────────────────────────────┘│
│    │                                                 │
│    │ ┌───────────────────┐ ┌───────────────────────┐│
│    │ │ Top Prompts       │ │ Version Performance   ││
│    │ │ by quality score  │ │ comparison (bar)      ││
│    │ └───────────────────┘ └───────────────────────┘│
└────┴────────────────────────────────────────────────┘
```

### 6. コンサルティングチャット (Layer 2)

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬──────────────┬─────────────────────────────────┤
│    │ Sessions     │ Healthcare Prompt Optimization   │
│ N  │              │                                  │
│ a  │ ● Current    │ ┌──────────────────────────────┐│
│ v  │ ○ Mar 7      │ │ 🤖 Based on your execution  ││
│    │ ○ Mar 5      │ │ logs, I notice your system   ││
│ R  │ ○ Mar 1      │ │ prompt lacks HIPAA-specific  ││
│ a  │              │ │ guardrails. Here's what I     ││
│ i  │              │ │ recommend:                    ││
│ l  │              │ │                               ││
│    │              │ │ 1. Add PHI detection...       ││
│    │ [+ New Chat] │ │ 2. Include disclaimer...      ││
│    │              │ │                               ││
│    │              │ │ Sources: [Log #1234] [KB:HC]  ││
│    │              │ │ [Apply to sys-prompt-v2 →]    ││
│    │              │ └──────────────────────────────┘│
│    │              │                                  │
│    │              │ ┌────────────────────────┐ [Send]│
│    │              │ │ Type your question...  │      │
│    │              │ └────────────────────────┘      │
└────┴──────────────┴─────────────────────────────────┘
```

### 7. 設定ページ

```
┌─────────────────────────────────────────────────────┐
│ Top App Bar                                          │
├────┬────────────────────────────────────────────────┤
│    │ Settings                                        │
│ N  │                                                 │
│ a  │ [Tab: General] [Tab: Members] [Tab: API Keys]  │
│ v  │                                                 │
│    │ ── Members ──────────────────────────────────── │
│ R  │                                                 │
│ a  │ [Invite Member]                                 │
│ i  │                                                 │
│ l  │ ┌──────────────────┬────────┬─────────────────┐│
│    │ │ user@example.com │ Owner  │                  ││
│    │ │ dev@example.com  │ Admin  │ [Role ▼] [Remove]││
│    │ │ pm@example.com   │ Member │ [Role ▼] [Remove]││
│    │ └──────────────────┴────────┴─────────────────┘│
└────┴────────────────────────────────────────────────┘
```

---

## State Management

### URL-Driven State

全ページ状態はURLで表現。ブラウザの戻る/進むが自然に動作。

```
/orgs/my-org/projects/my-proj/prompts?type=system&sort=updated
```

### HTMX State

- ページ遷移: `hx-get` + `hx-push-url="true"` （URL更新あり）
- パーシャル更新: `hx-get` + `hx-target="#panel"` （URL更新なし）
- フォーム送信: `hx-post` + `hx-swap="outerHTML"`

### Server State

テンプレートレンダリング時にサーバーが全状態を保持。クライアントは状態を持たない。

```go
// Handler がコンテキストから全状態を構築
func PromptDetailHandler(apiClient *client.APIClient) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        prompt := apiClient.GetPrompt(ctx, projectID, slug)
        versions := apiClient.ListVersions(ctx, prompt.ID)
        templates.PromptDetail(prompt, versions).Render(ctx, w)
    }
}
```
