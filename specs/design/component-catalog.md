# Component Catalog

templ コンポーネントの仕様カタログ。M3 準拠、HTMX 駆動。

## Component Tree

```
templates/
├── layout.templ              # ベースレイアウト（HTML shell）
├── app_layout.templ          # アプリレイアウト（NavRail + Content）
├── pages/
│   ├── login.templ
│   ├── projects.templ
│   ├── prompt_list.templ
│   ├── prompt_detail.templ
│   ├── version_detail.templ
│   ├── version_diff.templ
│   ├── log_list.templ
│   ├── log_detail.templ
│   ├── analytics.templ
│   ├── chat.templ
│   └── settings.templ
└── components/
    ├── navigation/
    │   ├── top_app_bar.templ
    │   ├── nav_rail.templ
    │   ├── breadcrumb.templ
    │   └── org_switcher.templ
    ├── prompt/
    │   ├── prompt_card.templ
    │   ├── version_list.templ
    │   ├── version_item.templ
    │   ├── version_badge.templ
    │   ├── prompt_editor.templ
    │   ├── diff_viewer.templ
    │   └── variable_chips.templ
    ├── log/
    │   ├── log_table.templ
    │   ├── log_row.templ
    │   └── log_detail_panel.templ
    ├── analytics/
    │   ├── stat_card.templ
    │   ├── chart_container.templ
    │   └── score_badge.templ
    ├── chat/
    │   ├── session_list.templ
    │   ├── message_bubble.templ
    │   ├── chat_input.templ
    │   └── citation_link.templ
    ├── settings/
    │   ├── member_table.templ
    │   └── api_key_card.templ
    └── shared/
        ├── button.templ
        ├── icon_button.templ
        ├── text_field.templ
        ├── chip.templ
        ├── dialog.templ
        ├── empty_state.templ
        ├── pagination.templ
        ├── status_message.templ
        ├── loading.templ
        └── tab_bar.templ
```

---

## Shared Components

### Button

M3 Button: Filled / Outlined / Text / Tonal / FAB

```templ
// Props
templ Button(label string, opts ButtonOpts) {
    <button
        class={ "btn", opts.Variant.Class() }
        if opts.HxGet != "" { hx-get={ opts.HxGet } }
        if opts.HxPost != "" { hx-post={ opts.HxPost } }
        if opts.HxTarget != "" { hx-target={ opts.HxTarget } }
        if opts.Disabled { disabled }
    >
        if opts.Icon != "" {
            <span class="material-symbols-outlined btn__icon">{ opts.Icon }</span>
        }
        <span class="btn__label">{ label }</span>
    </button>
}
```

```css
.btn {
  display: inline-flex; align-items: center; gap: var(--space-2);
  padding: 10px 24px; border: none; border-radius: var(--md-sys-shape-full);
  font: var(--md-sys-typescale-label-large); cursor: pointer;
  transition: background var(--md-sys-motion-duration-short3) var(--md-sys-motion-easing-standard);
}
.btn--filled {
  background: var(--md-sys-color-primary); color: var(--md-sys-color-on-primary);
}
.btn--filled:hover { box-shadow: var(--md-sys-elevation-shadow-1); }
.btn--outlined {
  background: transparent; color: var(--md-sys-color-primary);
  border: 1px solid var(--md-sys-color-outline);
}
.btn--text { background: transparent; color: var(--md-sys-color-primary); padding: 10px 12px; }
.btn--tonal {
  background: var(--md-sys-color-secondary-container); color: var(--md-sys-color-on-secondary-container);
}
.btn--fab {
  width: 56px; height: 56px; padding: 0; justify-content: center;
  border-radius: var(--md-sys-shape-large);
  background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container);
  box-shadow: var(--md-sys-elevation-shadow-2);
}
```

### TextField

M3 Outlined Text Field

```templ
templ TextField(name, label string, opts TextFieldOpts) {
    <div class={ "text-field", templ.KV("text-field--error", opts.Error != "") }>
        <input
            type={ opts.Type }
            name={ name }
            id={ name }
            class="text-field__input"
            placeholder=" "
            value={ opts.Value }
            if opts.Required { required }
        />
        <label for={ name } class="text-field__label">{ label }</label>
        if opts.Error != "" {
            <span class="text-field__error">{ opts.Error }</span>
        }
        if opts.Helper != "" {
            <span class="text-field__helper">{ opts.Helper }</span>
        }
    </div>
}
```

```css
.text-field { position: relative; margin-bottom: var(--space-4); }
.text-field__input {
  width: 100%; padding: 16px; font: var(--md-sys-typescale-body-large);
  border: 1px solid var(--md-sys-color-outline); border-radius: var(--md-sys-shape-extra-small);
  background: transparent; color: var(--md-sys-color-on-surface);
  transition: border-color var(--md-sys-motion-duration-short3);
}
.text-field__input:focus { outline: none; border: 2px solid var(--md-sys-color-primary); padding: 15px; }
.text-field__label {
  position: absolute; left: 12px; top: 16px; font: var(--md-sys-typescale-body-large);
  color: var(--md-sys-color-on-surface-variant); pointer-events: none;
  transition: all var(--md-sys-motion-duration-short3) var(--md-sys-motion-easing-standard);
  background: var(--md-sys-color-surface); padding: 0 4px;
}
.text-field__input:focus + .text-field__label,
.text-field__input:not(:placeholder-shown) + .text-field__label {
  top: -8px; font: var(--md-sys-typescale-body-small); color: var(--md-sys-color-primary);
}
.text-field--error .text-field__input { border-color: var(--md-sys-color-error); }
.text-field--error .text-field__label { color: var(--md-sys-color-error); }
.text-field__error { font: var(--md-sys-typescale-body-small); color: var(--md-sys-color-error); margin-top: var(--space-1); }
```

### Chip

M3 Assist / Filter / Input Chip

```templ
templ Chip(label string, opts ChipOpts) {
    <span class={ "chip", opts.Variant.Class(), templ.KV("chip--selected", opts.Selected) }>
        if opts.Icon != "" {
            <span class="material-symbols-outlined chip__icon">{ opts.Icon }</span>
        }
        <span class="chip__label">{ label }</span>
        if opts.Removable {
            <button class="chip__remove" aria-label="Remove">
                <span class="material-symbols-outlined">close</span>
            </button>
        }
    </span>
}
```

### Dialog

M3 Dialog（HTMX 対応）

```templ
templ Dialog(title string) {
    <div class="dialog-overlay" id="dialog-overlay"
         hx-on:click="htmx.remove('#dialog-overlay')">
        <dialog class="dialog" open
                hx-on:click="event.stopPropagation()">
            <h2 class="dialog__title">{ title }</h2>
            <div class="dialog__content">
                { children... }
            </div>
        </dialog>
    </div>
}
```

### StatusMessage (OOB)

```templ
templ StatusMessage(msg string, isError bool) {
    <div id="status" role="status" aria-live="polite"
         class={ "status-msg", templ.KV("status-msg--error", isError) }
         hx-swap-oob="true">
        <span class="material-symbols-outlined">
            if isError { "error" } else { "check_circle" }
        </span>
        { msg }
    </div>
}
```

### EmptyState

```templ
templ EmptyState(icon, title, description string) {
    <div class="empty-state">
        <span class="material-symbols-outlined empty-state__icon">{ icon }</span>
        <h3 class="empty-state__title">{ title }</h3>
        <p class="empty-state__description">{ description }</p>
        { children... }
    </div>
}
```

### Pagination

```templ
templ Pagination(current, total int, baseURL string) {
    <nav class="pagination" aria-label="Pagination">
        if current > 1 {
            <a class="pagination__btn" hx-get={ fmt.Sprintf("%s?page=%d", baseURL, current-1) }
               hx-target="#content" hx-push-url="true">
                <span class="material-symbols-outlined">chevron_left</span>
            </a>
        }
        <span class="pagination__info">{ fmt.Sprintf("%d / %d", current, total) }</span>
        if current < total {
            <a class="pagination__btn" hx-get={ fmt.Sprintf("%s?page=%d", baseURL, current+1) }
               hx-target="#content" hx-push-url="true">
                <span class="material-symbols-outlined">chevron_right</span>
            </a>
        }
    </nav>
}
```

---

## Domain Components

### VersionBadge

バージョンのステータスを色付きバッジで表示。

```templ
templ VersionBadge(status string) {
    <span class={ "version-badge", "version-badge--" + status }>
        <span class="material-symbols-outlined version-badge__icon">
            switch status {
                case "draft":      "edit_note"
                case "review":     "rate_review"
                case "production": "rocket_launch"
                case "archived":   "archive"
            }
        </span>
        { status }
    </span>
}
```

```css
.version-badge {
  display: inline-flex; align-items: center; gap: var(--space-1);
  padding: 2px 10px; border-radius: var(--md-sys-shape-full);
  font: var(--md-sys-typescale-label-medium); text-transform: capitalize;
}
.version-badge--draft      { background: #F3ECFF; color: var(--color-status-draft); }
.version-badge--review     { background: #FFF8E1; color: var(--color-status-review); }
.version-badge--production { background: #E8F5E9; color: var(--color-status-production); }
.version-badge--archived   { background: #F5F5F5; color: var(--color-status-archived); }
.version-badge__icon { font-size: 16px; }
```

### PromptCard

プロンプト一覧に表示されるカード。

```templ
templ PromptCard(p Prompt) {
    <a class="prompt-card" href={ templ.SafeURL(p.DetailURL()) }
       hx-get={ p.DetailPartialURL() } hx-target="#content" hx-push-url={ p.DetailURL() }>
        <div class="prompt-card__header">
            <span class="material-symbols-outlined prompt-card__type-icon">
                switch p.PromptType {
                    case "system":   "memory"
                    case "user":     "person"
                    case "combined": "join"
                }
            </span>
            <div class="prompt-card__titles">
                <h3 class="prompt-card__name">{ p.Name }</h3>
                <span class="prompt-card__slug">{ p.Slug }</span>
            </div>
            @VersionBadge(p.ProductionStatus())
        </div>
        if p.Description != "" {
            <p class="prompt-card__description">{ p.Description }</p>
        }
        <div class="prompt-card__meta">
            <span>v{ strconv.Itoa(p.LatestVersion) }</span>
            <span>{ p.PromptType }</span>
        </div>
    </a>
}
```

### VersionItem

バージョン履歴リスト内の1アイテム。

```templ
templ VersionItem(v Version, isSelected bool) {
    <li class={ "version-item", templ.KV("version-item--selected", isSelected) }
        hx-get={ v.DetailPartialURL() }
        hx-target="#version-detail"
        hx-swap="innerHTML">
        <div class="version-item__number">v{ strconv.Itoa(v.VersionNumber) }</div>
        @VersionBadge(string(v.Status))
        <div class="version-item__meta">
            <span class="version-item__desc">{ v.ChangeDescription }</span>
            <time class="version-item__date">{ v.CreatedAt.Format("Jan 2") }</time>
        </div>
    </li>
}
```

### DiffViewer

2バージョン間のテキスト差分表示。

```templ
templ DiffViewer(left, right DiffContent) {
    <div class="diff-viewer">
        <div class="diff-viewer__tabs" role="tablist">
            <button role="tab" class="diff-tab diff-tab--active"
                    hx-get={ left.TextDiffURL() } hx-target="#diff-content">Text Diff</button>
            <button role="tab" class="diff-tab"
                    hx-get={ left.SemanticDiffURL() } hx-target="#diff-content">Semantic Diff</button>
        </div>
        <div id="diff-content" class="diff-viewer__content">
            <div class="diff-side diff-side--left">
                <div class="diff-side__header">v{ strconv.Itoa(left.Version) } ({ left.Status })</div>
                <pre class="diff-side__code">{ left.Content }</pre>
            </div>
            <div class="diff-side diff-side--right">
                <div class="diff-side__header">v{ strconv.Itoa(right.Version) } ({ right.Status })</div>
                <pre class="diff-side__code">{ right.Content }</pre>
            </div>
        </div>
    </div>
}
```

### PromptEditor

プロンプト内容のエディタ（textarea + プレビュー）。

```templ
templ PromptEditor(content string, variables []Variable) {
    <div class="prompt-editor">
        <div class="prompt-editor__toolbar">
            <button class="btn btn--text" onclick="document.getElementById('editor').focus()">
                <span class="material-symbols-outlined">edit</span> Edit
            </button>
            <div class="prompt-editor__variables">
                for _, v := range variables {
                    @Chip(fmt.Sprintf("{{%s}}", v.Name), ChipOpts{Icon: "data_object"})
                }
            </div>
        </div>
        <textarea id="editor" name="content" class="prompt-editor__textarea"
                  rows="12" spellcheck="true">{ content }</textarea>
    </div>
}
```

```css
.prompt-editor__textarea {
  width: 100%; padding: var(--space-4);
  font: var(--md-sys-typescale-code-large);
  background: var(--md-sys-color-surface-container);
  border: 1px solid var(--md-sys-color-outline-variant);
  border-radius: var(--md-sys-shape-small);
  resize: vertical; min-height: 200px;
  color: var(--md-sys-color-on-surface);
}
.prompt-editor__textarea:focus {
  outline: none; border-color: var(--md-sys-color-primary);
  box-shadow: 0 0 0 1px var(--md-sys-color-primary);
}
```

### ChatMessage (SSE対応)

```templ
templ ChatMessage(msg Message) {
    <div class={ "chat-msg", "chat-msg--" + msg.Role }>
        if msg.Role == "assistant" {
            <span class="material-symbols-outlined chat-msg__avatar">smart_toy</span>
        }
        <div class="chat-msg__body">
            <div class="chat-msg__content">{ msg.Content }</div>
            if len(msg.Citations) > 0 {
                <div class="chat-msg__citations">
                    for _, c := range msg.Citations {
                        @CitationLink(c)
                    }
                </div>
            }
            if len(msg.Actions) > 0 {
                <div class="chat-msg__actions">
                    for _, a := range msg.Actions {
                        @Button(a.Label, ButtonOpts{Variant: Tonal, HxPost: a.URL, HxTarget: "#status"})
                    }
                </div>
            }
        </div>
    </div>
}
```

### StatCard（アナリティクス）

```templ
templ StatCard(label, value, trend string, isPositive bool) {
    <div class="stat-card">
        <span class="stat-card__label">{ label }</span>
        <span class="stat-card__value">{ value }</span>
        <span class={ "stat-card__trend", templ.KV("stat-card__trend--up", isPositive), templ.KV("stat-card__trend--down", !isPositive) }>
            if isPositive { "↑" } else { "↓" }
            { trend }
        </span>
    </div>
}
```

---

## Layout Components

### AppLayout

全ページ共通のシェルレイアウト。

```templ
templ AppLayout(nav NavContext) {
    @Layout() {
        <div class="app-layout">
            @TopAppBar(nav)
            @NavRail(nav.ActiveItem)
            <main class="app-content" id="content">
                { children... }
            </main>
        </div>
    }
}
```

### TopAppBar

M3 Small Top App Bar。

```templ
templ TopAppBar(nav NavContext) {
    <header class="top-app-bar">
        <div class="top-app-bar__start">
            <a href="/" class="top-app-bar__logo">PromptLab</a>
        </div>
        <div class="top-app-bar__center">
            @Breadcrumb(nav.Breadcrumbs)
        </div>
        <div class="top-app-bar__end">
            <button class="icon-btn" aria-label="Search">
                <span class="material-symbols-outlined">search</span>
            </button>
            @OrgSwitcher(nav.CurrentOrg, nav.Organizations)
            <button class="icon-btn avatar" aria-label="Account">
                { string(nav.User.Name[0]) }
            </button>
        </div>
    </header>
}
```

### NavRail

```templ
templ NavRail(active string) {
    <nav class="nav-rail" aria-label="Main navigation">
        @navRailItem("description", "Prompts", "/prompts", active == "prompts")
        @navRailItem("receipt_long", "Logs", "/logs", active == "logs")
        @navRailItem("analytics", "Analytics", "/analytics", active == "analytics")
        @navRailItem("chat", "Chat", "/chat", active == "chat")
        <div class="nav-rail__spacer"></div>
        @navRailItem("settings", "Settings", "/settings", active == "settings")
    </nav>
}

templ navRailItem(icon, label, href string, active bool) {
    <a class={ "nav-rail__item", templ.KV("nav-rail__item--active", active) }
       href={ templ.SafeURL(href) }
       hx-get={ href } hx-target="#content" hx-push-url="true">
        <div class="nav-rail__indicator">
            <span class="material-symbols-outlined">{ icon }</span>
        </div>
        <span class="nav-rail__label">{ label }</span>
    </a>
}
```

```css
.nav-rail {
  width: 80px; display: flex; flex-direction: column; align-items: center;
  padding: var(--space-3) 0; gap: var(--space-3);
  background: var(--md-sys-color-surface);
  border-right: 1px solid var(--md-sys-color-outline-variant);
}
.nav-rail__item {
  display: flex; flex-direction: column; align-items: center; gap: var(--space-1);
  text-decoration: none; color: var(--md-sys-color-on-surface-variant);
}
.nav-rail__indicator {
  width: 56px; height: 32px; display: flex; align-items: center; justify-content: center;
  border-radius: var(--md-sys-shape-large);
  transition: background var(--md-sys-motion-duration-short3);
}
.nav-rail__item:hover .nav-rail__indicator { background: var(--md-sys-color-surface-container); }
.nav-rail__item--active .nav-rail__indicator {
  background: var(--md-sys-color-secondary-container);
  color: var(--md-sys-color-on-secondary-container);
}
.nav-rail__label { font: var(--md-sys-typescale-label-medium); }
.nav-rail__spacer { flex: 1; }
```
