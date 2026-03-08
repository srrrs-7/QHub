# PromptLab Design System

Material Design 3 (M3) をベースに、templ + HTMX のサーバーサイドレンダリング環境に最適化したデザインシステム。

## Design Philosophy

### 原則

1. **Content First** — プロンプトテキストが主役。UIはコンテンツを邪魔しない
2. **Progressive Disclosure** — 必要な情報だけを段階的に表示
3. **Instant Feedback** — HTMX による即座のUI更新。ページ遷移なし
4. **Accessible by Default** — WCAG 2.1 AA準拠。キーボード操作完全対応
5. **Dense but Readable** — 開発者向けの情報密度。ただし読みやすさを犠牲にしない

### M3 適用方針

- CSS Custom Properties でテーマトークンを管理
- JavaScript コンポーネントは使わない（HTMX + CSS のみ）
- M3 の Color / Typography / Shape / Elevation / Motion をCSS実装
- サーバーサイドで状態管理、クライアントは表示に専念

---

## Color System

### Tonal Palette（M3 Dynamic Color ベース）

Primary: Indigo（プロンプト管理のプロフェッショナル感）

```css
:root {
  /* Primary */
  --md-sys-color-primary: #4355B9;
  --md-sys-color-on-primary: #FFFFFF;
  --md-sys-color-primary-container: #DEE0FF;
  --md-sys-color-on-primary-container: #00105C;

  /* Secondary */
  --md-sys-color-secondary: #5B5D72;
  --md-sys-color-on-secondary: #FFFFFF;
  --md-sys-color-secondary-container: #E0E1F9;
  --md-sys-color-on-secondary-container: #181A2C;

  /* Tertiary */
  --md-sys-color-tertiary: #77536D;
  --md-sys-color-on-tertiary: #FFFFFF;
  --md-sys-color-tertiary-container: #FFD7F1;
  --md-sys-color-on-tertiary-container: #2D1228;

  /* Error */
  --md-sys-color-error: #BA1A1A;
  --md-sys-color-on-error: #FFFFFF;
  --md-sys-color-error-container: #FFDAD6;
  --md-sys-color-on-error-container: #410002;

  /* Surface (M3 Tonal Surface) */
  --md-sys-color-surface: #FFFBFF;
  --md-sys-color-on-surface: #1B1B21;
  --md-sys-color-surface-variant: #E3E1EC;
  --md-sys-color-on-surface-variant: #46464F;
  --md-sys-color-surface-container-lowest: #FFFFFF;
  --md-sys-color-surface-container-low: #F6F2FA;
  --md-sys-color-surface-container: #F0ECF4;
  --md-sys-color-surface-container-high: #EAE7EF;
  --md-sys-color-surface-container-highest: #E5E1E9;

  /* Outline */
  --md-sys-color-outline: #777680;
  --md-sys-color-outline-variant: #C7C5D0;

  /* Inverse */
  --md-sys-color-inverse-surface: #303036;
  --md-sys-color-inverse-on-surface: #F3EFF7;
  --md-sys-color-inverse-primary: #BAC3FF;
}
```

### Semantic Colors（ドメイン固有）

```css
:root {
  /* Version Status */
  --color-status-draft: #6750A4;       /* Primary variant */
  --color-status-review: #E8A317;      /* Amber */
  --color-status-production: #1B873B;  /* Green */
  --color-status-archived: #777680;    /* Outline */

  /* Prompt Type */
  --color-type-system: #4355B9;        /* Primary */
  --color-type-user: #77536D;          /* Tertiary */
  --color-type-combined: #5B5D72;      /* Secondary */

  /* Diff */
  --color-diff-added: #D4EDDA;
  --color-diff-removed: #F8D7DA;
  --color-diff-changed: #FFF3CD;

  /* Evaluation Score */
  --color-score-excellent: #1B873B;    /* 90-100 */
  --color-score-good: #4CAF50;         /* 70-89 */
  --color-score-fair: #E8A317;         /* 50-69 */
  --color-score-poor: #BA1A1A;         /* 0-49 */
}
```

### Dark Theme

```css
@media (prefers-color-scheme: dark) {
  :root {
    --md-sys-color-primary: #BAC3FF;
    --md-sys-color-on-primary: #0C1F92;
    --md-sys-color-primary-container: #2C3EA0;
    --md-sys-color-on-primary-container: #DEE0FF;

    --md-sys-color-surface: #1B1B21;
    --md-sys-color-on-surface: #E5E1E9;
    --md-sys-color-surface-container-lowest: #131318;
    --md-sys-color-surface-container-low: #1B1B21;
    --md-sys-color-surface-container: #201F26;
    --md-sys-color-surface-container-high: #2A2930;
    --md-sys-color-surface-container-highest: #35343B;

    --md-sys-color-outline: #918F9A;
    --md-sys-color-outline-variant: #46464F;
  }
}
```

---

## Typography

### Type Scale（M3 ベース）

システムフォント優先。コード表示は等幅フォント。

```css
:root {
  /* Font Families */
  --md-sys-typescale-font: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Noto Sans JP', sans-serif;
  --md-sys-typescale-font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;

  /* Display */
  --md-sys-typescale-display-large: 400 2.25rem/2.75rem var(--md-sys-typescale-font);   /* 36/44 */
  --md-sys-typescale-display-medium: 400 2.813rem/2.25rem var(--md-sys-typescale-font);  /* 45/52 → 不使用、情報密度優先 */

  /* Headline */
  --md-sys-typescale-headline-large: 400 2rem/2.5rem var(--md-sys-typescale-font);       /* 32/40 */
  --md-sys-typescale-headline-medium: 400 1.75rem/2.25rem var(--md-sys-typescale-font);  /* 28/36 */
  --md-sys-typescale-headline-small: 400 1.5rem/2rem var(--md-sys-typescale-font);       /* 24/32 */

  /* Title */
  --md-sys-typescale-title-large: 400 1.375rem/1.75rem var(--md-sys-typescale-font);     /* 22/28 */
  --md-sys-typescale-title-medium: 500 1rem/1.5rem var(--md-sys-typescale-font);         /* 16/24 */
  --md-sys-typescale-title-small: 500 0.875rem/1.25rem var(--md-sys-typescale-font);     /* 14/20 */

  /* Body */
  --md-sys-typescale-body-large: 400 1rem/1.5rem var(--md-sys-typescale-font);           /* 16/24 */
  --md-sys-typescale-body-medium: 400 0.875rem/1.25rem var(--md-sys-typescale-font);     /* 14/20 */
  --md-sys-typescale-body-small: 400 0.75rem/1rem var(--md-sys-typescale-font);          /* 12/16 */

  /* Label */
  --md-sys-typescale-label-large: 500 0.875rem/1.25rem var(--md-sys-typescale-font);     /* 14/20 */
  --md-sys-typescale-label-medium: 500 0.75rem/1rem var(--md-sys-typescale-font);        /* 12/16 */
  --md-sys-typescale-label-small: 500 0.688rem/1rem var(--md-sys-typescale-font);        /* 11/16 */

  /* Code（プロンプト表示専用） */
  --md-sys-typescale-code-large: 400 0.938rem/1.5rem var(--md-sys-typescale-font-mono);  /* 15/24 */
  --md-sys-typescale-code-medium: 400 0.813rem/1.25rem var(--md-sys-typescale-font-mono);/* 13/20 */
}
```

---

## Shape System

### Corner Radius（M3 Shape Scale）

```css
:root {
  --md-sys-shape-none: 0;
  --md-sys-shape-extra-small: 4px;
  --md-sys-shape-small: 8px;
  --md-sys-shape-medium: 12px;
  --md-sys-shape-large: 16px;
  --md-sys-shape-extra-large: 28px;
  --md-sys-shape-full: 9999px;
}
```

### 用途マッピング

| コンポーネント | Shape | 値 |
|---|---|---|
| Chip, Badge | Full | 9999px |
| Button | Full | 9999px |
| Text Field | Extra Small | 4px (top only) |
| Card | Medium | 12px |
| Dialog | Extra Large | 28px |
| Navigation Rail | Large | 16px (indicator) |
| FAB | Large | 16px |

---

## Elevation System

M3 では shadow ではなく tonal elevation（Surface Container の色変化）を推奨。

```css
:root {
  /* Elevation Levels → Surface Container mapping */
  --md-sys-elevation-0: var(--md-sys-color-surface);
  --md-sys-elevation-1: var(--md-sys-color-surface-container-low);
  --md-sys-elevation-2: var(--md-sys-color-surface-container);
  --md-sys-elevation-3: var(--md-sys-color-surface-container-high);
  --md-sys-elevation-4: var(--md-sys-color-surface-container-highest);

  /* Shadow（hover/drag時のみ使用） */
  --md-sys-elevation-shadow-1: 0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.08);
  --md-sys-elevation-shadow-2: 0 3px 6px rgba(0,0,0,0.12), 0 2px 4px rgba(0,0,0,0.08);
  --md-sys-elevation-shadow-3: 0 8px 16px rgba(0,0,0,0.12), 0 4px 8px rgba(0,0,0,0.08);
}
```

### 用途マッピング

| コンポーネント | Elevation | 色 |
|---|---|---|
| Page Background | 0 | surface |
| Card (default) | 1 | surface-container-low |
| Navigation Rail | 0 | surface |
| Top App Bar | 2 | surface-container |
| Dialog | 3 | surface-container-high |
| Prompt Editor | 1 | surface-container-low |
| Code Block | 2 | surface-container |

---

## Spacing & Grid

### Spacing Scale（4dp ベース）

```css
:root {
  --space-0: 0;
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-8: 32px;
  --space-10: 40px;
  --space-12: 48px;
  --space-16: 64px;
}
```

### Layout Grid

```
Desktop (≥1240px):  Navigation Rail (80px) + Content (fluid) + Detail Panel (400px optional)
Tablet  (600-1239): Navigation Rail (80px) + Content (fluid)
Mobile  (<600px):   Bottom Navigation + Content (full width)
```

```css
.app-layout {
  display: grid;
  grid-template-columns: 80px 1fr;
  grid-template-rows: auto 1fr;
  height: 100vh;
}

.app-layout--with-detail {
  grid-template-columns: 80px 1fr 400px;
}

@media (max-width: 599px) {
  .app-layout {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr auto;
  }
}
```

---

## Motion

### Transition Tokens

```css
:root {
  /* Duration */
  --md-sys-motion-duration-short1: 50ms;
  --md-sys-motion-duration-short2: 100ms;
  --md-sys-motion-duration-short3: 150ms;
  --md-sys-motion-duration-short4: 200ms;
  --md-sys-motion-duration-medium1: 250ms;
  --md-sys-motion-duration-medium2: 300ms;
  --md-sys-motion-duration-medium4: 400ms;
  --md-sys-motion-duration-long2: 500ms;

  /* Easing */
  --md-sys-motion-easing-standard: cubic-bezier(0.2, 0, 0, 1);
  --md-sys-motion-easing-emphasized: cubic-bezier(0.2, 0, 0, 1);
  --md-sys-motion-easing-emphasized-decelerate: cubic-bezier(0.05, 0.7, 0.1, 1);
  --md-sys-motion-easing-emphasized-accelerate: cubic-bezier(0.3, 0, 0.8, 0.15);
}
```

### HTMX Transition Classes

```css
/* HTMX swap animation */
.htmx-swapping { opacity: 0; transition: opacity var(--md-sys-motion-duration-short3) var(--md-sys-motion-easing-emphasized-accelerate); }
.htmx-settling { opacity: 1; transition: opacity var(--md-sys-motion-duration-medium1) var(--md-sys-motion-easing-emphasized-decelerate); }
.htmx-added    { opacity: 0; transition: opacity var(--md-sys-motion-duration-medium2) var(--md-sys-motion-easing-standard); }

/* Loading indicator */
.htmx-indicator { display: none; }
.htmx-request .htmx-indicator { display: inline-flex; }
.htmx-request.htmx-indicator { display: inline-flex; }
```

---

## Iconography

Material Symbols Outlined（Variable Font、CDN から読み込み）

```html
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" />
```

### 主要アイコンマッピング

| 機能 | アイコン名 | 用途 |
|---|---|---|
| プロンプト | `description` | プロンプト一覧 |
| バージョン | `history` | バージョン履歴 |
| 差分 | `compare_arrows` | Semantic Diff |
| 実行ログ | `receipt_long` | ログ一覧 |
| 評価 | `analytics` | 品質スコア |
| Lint | `rule` | Lintルール |
| チャット | `chat` | コンサルティング |
| プロジェクト | `folder` | プロジェクト |
| 組織 | `business` | 組織管理 |
| API Key | `key` | キー管理 |
| 設定 | `settings` | 設定 |
| ステータス: Draft | `edit_note` | 下書き |
| ステータス: Review | `rate_review` | レビュー中 |
| ステータス: Production | `rocket_launch` | 本番 |
| ステータス: Archived | `archive` | アーカイブ |

---

## Accessibility

### 要件

- **コントラスト比**: 4.5:1 以上（テキスト）、3:1 以上（大テキスト・UI要素）
- **フォーカス**: 全インタラクティブ要素にvisible focus ring
- **キーボード**: Tab / Shift+Tab / Enter / Escape で全操作可能
- **ARIA**: ランドマーク、ライブリージョン、ステート属性を適切に設定
- **Motion**: `prefers-reduced-motion` で アニメーション無効化

```css
/* Focus ring (M3 style) */
:focus-visible {
  outline: 3px solid var(--md-sys-color-primary);
  outline-offset: 2px;
  border-radius: var(--md-sys-shape-extra-small);
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### HTMX ARIA パターン

```html
<!-- ローディング状態 -->
<div id="content" aria-busy="false"
     hx-get="/partials/prompts"
     hx-indicator="#loading"
     hx-on::before-request="this.setAttribute('aria-busy','true')"
     hx-on::after-settle="this.setAttribute('aria-busy','false')">
</div>
<span id="loading" class="htmx-indicator" role="status" aria-label="Loading">
  <span class="material-symbols-outlined spin">progress_activity</span>
</span>

<!-- ライブリージョン（ステータスメッセージ） -->
<div id="status" role="status" aria-live="polite" aria-atomic="true"></div>
```
