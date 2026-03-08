# Web Frontend

Go + templ + HTMX で構築されたタスク管理アプリケーションのフロントエンドです。

## 概要

- **フレームワーク**: [go-chi/chi](https://github.com/go-chi/chi) ルーター
- **テンプレートエンジン**: [templ](https://github.com/a-h/templ) - Go用の型安全なHTMLテンプレート
- **インタラクション**: [HTMX](https://htmx.org/) - JavaScriptなしでAJAX通信
- **ポート**: 3000

## ディレクトリ構成

```
apps/web/
├── .images/
│   └── Dockerfile       # マルチステージビルド (builder → runner)
├── src/
│   ├── cmd/
│   │   └── main.go      # エントリポイント、HTTPサーバー起動
│   ├── client/
│   │   ├── api.go       # バックエンドAPI通信クライアント
│   │   └── types.go     # API レスポンス型定義
│   ├── handlers/
│   │   ├── tasks.go     # メインページハンドラー
│   │   └── partials.go  # HTMX パーシャルハンドラー
│   ├── routes/
│   │   └── routes.go    # ルーティング設定
│   └── templates/
│       ├── layout.templ       # 共通レイアウト (HTML/CSS)
│       ├── index.templ        # メインページ
│       └── components/
│           ├── add_form.templ   # タスク追加フォーム
│           ├── task_list.templ  # タスク一覧
│           ├── task_item.templ  # タスクアイテム
│           └── status.templ     # ステータスメッセージ
├── go.mod
├── go.sum
└── README.md
```

## 開発環境セットアップ

### 前提条件

- Go 1.25.4+
- templ CLI

### templ CLI のインストール

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

### 依存関係のインストール

```bash
cd apps/web
go mod download
```

## 開発方法

### 1. テンプレートの生成

`.templ` ファイルを編集した後、Go コードを生成する必要があります:

```bash
# 一度だけ生成
make templ-gen

# ファイル変更を監視して自動生成 (開発時推奨)
make templ-watch
```

### 2. ローカル実行

**Docker Compose を使用 (推奨)**:

```bash
# リポジトリルートで実行
docker compose up -d
```

`http://localhost:3000` でアクセス可能になります。

**直接実行する場合**:

```bash
# バックエンドAPIが起動している必要があります
cd apps/web/src
templ generate
go run ./cmd
```

### 環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `PORT` | `3000` | HTTPサーバーのポート |
| `API_BASE_URL` | `http://localhost:8080` | バックエンドAPIのベースURL |

## アーキテクチャ

### リクエストフロー

```
Browser → Web Server (port 3000) → Backend API (port 8080) → Database
           ↓
         templ SSR
           ↓
         HTML Response
```

### HTMX パーシャル

HTMXを使用してページ全体をリロードせずに部分更新を行います:

| エンドポイント | メソッド | 説明 |
|---------------|---------|------|
| `/` | GET | メインページ (初回ロード) |
| `/health` | GET | ヘルスチェック |
| `/partials/tasks` | GET | タスク一覧の更新 |
| `/partials/tasks` | POST | 新規タスク追加 |

### API クライアント

`src/client/api.go` でバックエンドAPIとの通信を行います:

```go
// タスク一覧取得
tasks, err := apiClient.ListTasks(ctx)

// タスク作成
task, err := apiClient.CreateTask(ctx, CreateTaskRequest{
    Title:       "タイトル",
    Description: "説明",
})
```

## テンプレート開発

### templ の基本構文

```go
// コンポーネント定義
templ MyComponent(name string) {
    <div class="component">
        <p>Hello, { name }!</p>
    </div>
}

// 条件分岐
if condition {
    <p>True</p>
} else {
    <p>False</p>
}

// ループ
for _, item := range items {
    @ItemComponent(item)
}

// 子コンポーネントの埋め込み
@Layout() {
    <main>{ children... }</main>
}
```

### HTMX 属性

```html
<!-- GET リクエストでタスク一覧を更新 -->
<button
    hx-get="/partials/tasks"
    hx-target="#tasks-container"
    hx-swap="outerHTML"
>
    Refresh
</button>

<!-- POST リクエストでタスク追加 -->
<form
    hx-post="/partials/tasks"
    hx-target="#tasks-container"
    hx-swap="outerHTML"
>
    ...
</form>
```

## ビルド

### ローカルビルド

```bash
cd apps/web/src
templ generate
go build -o web ./cmd
./web
```

### Docker イメージビルド

```bash
# リポジトリルートから実行
docker build -f apps/web/.images/Dockerfile -t web .
```

### Docker イメージの実行

```bash
docker run -p 3000:3000 \
  -e API_BASE_URL=http://host.docker.internal:8080 \
  web
```

## テスト

```bash
cd apps/web
go test ./...
```

## コード品質

```bash
# フォーマット
make fmt

# 静的解析
make vet

# templ フォーマット
make templ-fmt
```

## トラブルシューティング

### テンプレートの変更が反映されない

`.templ` ファイルを編集した後は `templ generate` を実行してください:

```bash
make templ-gen
```

### APIに接続できない

1. バックエンドAPIが起動しているか確認:
   ```bash
   curl http://localhost:8080/health
   ```

2. `API_BASE_URL` 環境変数が正しく設定されているか確認

### HTMX が動作しない

ブラウザの開発者ツールでネットワークタブを確認し、パーシャルエンドポイントへのリクエストが正しく送信されているか確認してください。

## 関連ドキュメント

- [templ ドキュメント](https://templ.guide/)
- [HTMX ドキュメント](https://htmx.org/docs/)
- [go-chi/chi ドキュメント](https://go-chi.io/)
