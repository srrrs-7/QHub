# Database Package

このパッケージは、PostgreSQL データベースのスキーマ、マイグレーション、および生成されたクエリコードを管理します。

## ディレクトリ構造

```
db/
├── atlas.hcl           # Atlas マイグレーション設定
├── sqlc.yaml           # sqlc コード生成設定
├── schema/             # データベーススキーマ定義（SQL）
│   └── tasks.sql
├── queries/            # SQLクエリ定義（sqlc用）
│   └── tasks.sql
├── migrations/         # バージョン管理されたマイグレーションファイル
│   └── .gitkeep
└── db/                 # sqlc生成コード
    ├── db.go
    ├── models.go
    ├── querier.go
    └── tasks.sql.go
```

## ツール

- **[Atlas](https://atlasgo.io/)**: データベーススキーママイグレーションツール
- **[sqlc](https://sqlc.dev/)**: SQLからタイプセーフなGoコードを生成

## 環境変数

データベース接続は `.devcontainer/compose.override.yaml` で定義された環境変数から構築されます:

- `DB_HOST`: データベースホスト（例: `db`）
- `DB_PORT`: データベースポート（例: `5432`）
- `DB_DBNAME`: データベース名（例: `test`）
- `DB_USERNAME`: データベースユーザー（例: `postgres`）
- `DB_PASSWORD`: データベースパスワード

## PostgreSQL への接続

### devcontainer 内から接続

```bash
# 接続文字列を使用
psql "postgresql://postgres:postgres@db:5432/test"

# または個別オプションで指定
psql -h db -p 5432 -U postgres -d test
```

### ホストマシンから接続

```bash
# 接続文字列を使用
psql "postgresql://postgres:postgres@localhost:5432/test"

# または個別オプションで指定
psql -h localhost -p 5432 -U postgres -d test

# パスワード入力なし（環境変数使用）
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d test
```

### JDBC 接続文字列

```
jdbc:postgresql://localhost:5432/test
ユーザー名: postgres
パスワード: postgres
```

## PostgreSQL コマンド (psql)

### よく使うメタコマンド

```sql
-- データベース情報
\l              -- データベース一覧
\c dbname       -- データベースに接続

-- テーブル・スキーマ
\dt             -- テーブル一覧
\dt+            -- テーブル一覧（詳細情報付き）
\dt *.*         -- すべてのスキーマのテーブル一覧
\d tablename    -- テーブル構造を表示

-- その他のオブジェクト
\dv             -- ビュー一覧
\di             -- インデックス一覧
\ds             -- シーケンス一覧
\df             -- 関数一覧
\du             -- ユーザー/ロール一覧

-- ヘルプ・終了
\?              -- メタコマンド一覧
\h              -- SQL コマンドのヘルプ
\q              -- psql 終了
```

### SQL クエリでテーブル一覧を取得

```sql
-- カレントスキーマのテーブル一覧
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public';

-- テーブル名とスキーマ
SELECT schemaname, tablename
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema');

-- テーブルのサイズを確認
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## マイグレーションワークフロー

### 1. スキーマファイルの編集

`schema/` ディレクトリ内のSQLファイルを編集して、希望するデータベーススキーマを定義します。

```sql
-- schema/tasks.sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    -- ...
);
```

### 2. マイグレーションの生成

スキーマファイルと現在のデータベースの差分からマイグレーションを生成:

```bash
make atlas-diff
```

または、空のマイグレーションファイルを作成:

```bash
make atlas-new NAME=add_users_table
```

### 3. マイグレーションの確認

生成されたマイグレーションファイルを `migrations/` ディレクトリで確認します。

### 4. マイグレーションの適用

```bash
# 確認プロンプトあり
make atlas-apply

# 自動承認（注意して使用）
make atlas-apply-auto
```

### 5. マイグレーションステータスの確認

```bash
make atlas-status
```

## Atlas コマンド一覧

| コマンド | 説明 |
|---------|------|
| `make atlas-new NAME=<name>` | 新しいマイグレーションファイルを作成 |
| `make atlas-diff` | スキーマの差分からマイグレーションを生成 |
| `make atlas-apply` | 保留中のマイグレーションを適用 |
| `make atlas-apply-auto` | マイグレーションを自動承認で適用 |
| `make atlas-status` | マイグレーションステータスを表示 |
| `make atlas-hash` | マイグレーションディレクトリを再ハッシュ |
| `make atlas-lint` | マイグレーションファイルをリント |
| `make atlas-validate` | マイグレーションディレクトリの整合性を検証 |
| `make atlas-inspect` | データベーススキーマを検査 |
| `make atlas-schema-apply` | スキーマを直接適用（開発用） |
| `make atlas-clean` | 開発用データベースをクリーン/リセット |
| `make atlas-help` | Atlas コマンドのヘルプを表示 |

## sqlc コマンド

| コマンド | 説明 |
|---------|------|
| `make sqlc-gen` | クエリからGoコードを生成 |
| `make sqlc-compile` | SQLクエリをコンパイル（検証） |
| `make sqlc-verify` | 生成されたコードを検証 |

## 環境の切り替え

デフォルトは `local` 環境ですが、`ATLAS_ENV` 変数で変更可能:

```bash
make atlas-apply ATLAS_ENV=docker
make atlas-apply ATLAS_ENV=ci
```

利用可能な環境:
- `local`: ローカル開発（デフォルト）
- `docker`: Docker Compose環境
- `ci`: CI/CD環境

## 開発データベースについて

Atlas は差分生成や検証のために開発用データベース (`test_dev`) を使用します。
このデータベースは Atlas が自由に操作できる必要があります。

## ベストプラクティス

1. **スキーマファーストアプローチ**: `schema/` ディレクトリでスキーマを定義し、`atlas-diff` でマイグレーションを生成
2. **マイグレーションのレビュー**: 適用前に生成されたマイグレーションを必ず確認
3. **バージョン管理**: すべてのマイグレーションファイルをGitにコミット
4. **手動編集の禁止**: 適用後のマイグレーションファイルは編集しない
5. **リント実行**: `make atlas-lint` で問題を早期発見

## トラブルシューティング

### マイグレーションの整合性エラー

```bash
make atlas-hash
```

### データベース接続エラー

環境変数が正しく設定されているか確認:

```bash
echo $DB_HOST $DB_PORT $DB_DBNAME
```

### スキーマの完全なリセット（開発用）

```bash
make atlas-clean
make atlas-apply
```
