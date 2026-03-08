# テストカバレッジ課題分析レポート

**分析日**: 2026-02-16  
**総合カバレッジ**: 28.6% (API), 40.2% (pkgs), 0.0% (web)  
**要求水準**: ≥80%  
**判定**: ❌ **CRITICAL - TDD要件を大幅に下回る**

---

## 📊 Executive Summary

プロジェクト全体のテストカバレッジは**28.6%**と、TDD要件の**80%を大幅に下回っています**。

### ファイルレベルの統計
- **プロダクションファイル**: 35ファイル
- **テストファイル**: 9ファイル  
- **テストファイル率**: 26% (9/35)
- **テストがない**: 74%のファイル（26ファイル）

### 重大な問題
1. ✅ **良好**: `utils/parallel` (99.3%), `routes/tasks` (91.0%)
2. ⚠️ **低カバレッジ**: `routes/response` (42.4%), `domain/task` (11.1%)
3. ❌ **未テスト**: コアパッケージ8個が0%

---

## 🔴 Critical Issues (0% Coverage)

### 1. **Result Monad (utils/types/result.go) - 0%**
**影響度**: 🔥🔥🔥 **CRITICAL**

プロジェクトの**中核的な関数型プログラミングパターン**が完全に未テスト。

#### 未テスト関数 (すべて0%)
- `Ok()`, `Err()`, `FromPair()` - コンストラクタ
- `IsOk()`, `IsErr()` - 判定
- `Map()`, `MapErr()`, `FlatMap()`, `AndThen()` - 変換
- `Match()`, `Combine()` - パターンマッチング
- `Pipe2()`, `Pipe3()`, `Pipe4()`, `Pipe5()` - パイプライン

#### リスク
- すべてのハンドラーが`Pipe2-5`に依存
- バグがあっても検出不可能
- リファクタリング不可能

---

### 2. **Error Handling (domain/apperror) - 0%**
**影響度**: 🔥🔥🔥 **CRITICAL**

アプリケーション全体のエラーハンドリング基盤が未テスト。

#### 未テスト関数 (すべて0%)
- `NewNotFoundError()` - 404エラー
- `NewValidationError()` - バリデーションエラー
- `NewDatabaseError()` - DB エラー
- `NewUnauthorizedError()` - 認証エラー
- `NewInternalServerError()` - 500エラー
- `NewBadRequestError()`, `NewConflictError()`, `NewForbiddenError()`
- `ErrorName()`, `DomainName()`, `Error()`, `Unwrap()`

#### リスク
- エラーメッセージフォーマット未検証
- エラーチェーン未検証
- 誤ったHTTPステータスコードマッピング可能

---

### 3. **Repository Layer (infra/rds/task_repository) - 0%**
**影響度**: 🔥🔥 **HIGH**

データアクセス層が完全に未テスト。

#### 未テスト関数
- `FindTaskByID()` - ID検索
- `FindAllTasks()` - 全件取得
- `CreateTask()` - タスク作成
- `UpdateTask()` - タスク更新

#### リスク
- SQLエラーハンドリング未検証
- タイムアウト処理未検証
- NotFoundエラーマッピング未検証
- データ変換ロジック未検証

---

### 4. **Domain Model (domain/task) - 11.1%**
**影響度**: 🔥🔥 **HIGH**

ドメインモデルの89%が未テスト。

#### 未テスト関数 (すべて0%)
```go
NewTaskID()           // UUID変換
TaskID.String()       // 文字列変換
TaskTitle.String()    
TaskDescription.String()
TaskStatus.String()
NewTask()             // エンティティ構築
IsCompleted()         // ビジネスロジック
IsPending()          // ビジネスロジック
```

#### テスト済み (11.1%)
- `NewTaskCmd()` - Cmdオブジェクト構築のみ

#### リスク
- 不正なUUIDでpanicの可能性
- ステータス判定ロジック未検証
- Value Object変換未検証

---

### 5. **その他の0%パッケージ**

#### **utils/env (0%)**
- 環境変数の取得・変換ロジック未検証
- `GetString()`, `GetInt()`, `GetBool()` すべて未テスト

#### **utils/logger (0%)**
- ログ出力未検証
- `Init()`, `Debug()`, `Info()`, `Warn()`, `Error()` すべて未テスト

#### **utils/testutil (0%)**
- テストヘルパー自体が未テスト
- `SetupTestTx()`, `SetAuthHeader()` 未検証

#### **utils/db (0%)**
- DB接続・トランザクション管理未検証
- sqlc生成コードすべて未テスト (12関数)

#### **Web Frontend (0%)**
- `client`, `handlers`, `routes`, `templates` すべて未テスト
- HTMX統合未検証
- API通信未検証

---

## ⚠️ Insufficient Coverage

### routes/response (42.4%)

#### 未テスト関数 (0%)
- `Accepted()` - 202レスポンス
- `NoContent()` - 204レスポンス
- `notFound()`, `unauthorized()`, `internalError()`
- `forbidden()`, `conflict()`, `unexpectedError()`

#### 不完全なテスト (75%)
- `OK()` - 正常系のみ（境界値・特殊文字・エラーケース未テスト）
- `Created()` - 正常系のみ
- `HandleAppError()` - BadRequestのみ（他7種のエラー未テスト）

#### TDD要件違反
現在のテスト: **1カテゴリのみ** (正常系)  
要求: **6カテゴリ** (正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil)

---

## ✅ Good Coverage

### utils/parallel (99.3%)
- ほぼ完全なカバレッジ
- TDD要件を満たす

### routes/tasks (91.0%)
- 良好なカバレッジ
- ハンドラーは適切にテストされている

---

## 📈 Coverage Requirements Gap

### 現状 vs 要件

| パッケージ | 現在 | 要件 | ギャップ | 判定 |
|-----------|------|------|---------|------|
| utils/parallel | 99.3% | ≥80% | - | ✅ PASS |
| routes/tasks | 91.0% | ≥80% | - | ✅ PASS |
| routes/response | 42.4% | ≥80% | -37.6% | ❌ FAIL |
| domain/task | 11.1% | ≥80% | -68.9% | ❌ FAIL |
| **types/result** | **0.0%** | **≥80%** | **-80.0%** | ❌ FAIL |
| **apperror** | **0.0%** | **≥80%** | **-80.0%** | ❌ FAIL |
| **task_repository** | **0.0%** | **≥80%** | **-80.0%** | ❌ FAIL |
| utils/env | 0.0% | ≥80% | -80.0% | ❌ FAIL |
| utils/logger | 0.0% | ≥80% | -80.0% | ❌ FAIL |
| utils/testutil | 0.0% | ≥80% | -80.0% | ❌ FAIL |
| utils/db | 0.0% | ≥80% | -80.0% | ❌ FAIL |
| web/* | 0.0% | ≥80% | -80.0% | ❌ FAIL |

**合格パッケージ**: 2/12 (17%)  
**不合格パッケージ**: 10/12 (83%)

---

## 🎯 Priority Action Items

### Phase 1: Critical Core (最優先)
**期限**: 即座  
**カバレッジ目標**: 90%+

1. **utils/types/result.go** 🔥🔥🔥
   - すべてのResult monad関数
   - 6カテゴリすべて
   - 推定: 50+ test cases

2. **domain/apperror/apperror.go** 🔥🔥🔥
   - 8つのエラータイプ
   - エラーチェーン検証
   - フォーマット検証
   - 推定: 40+ test cases

### Phase 2: Domain & Repository (高優先)
**期限**: 1週間以内  
**カバレッジ目標**: 85%+

3. **domain/task/task.go** 🔥🔥
   - Value Objects (TaskID, TaskTitle, etc.)
   - ビジネスロジック (IsCompleted, IsPending)
   - 不正UUID処理
   - 推定: 30+ test cases

4. **infra/rds/task_repository** 🔥🔥
   - DB統合テスト
   - エラーハンドリング
   - タイムアウト処理
   - 推定: 25+ test cases

### Phase 3: Response & Utilities (中優先)
**期限**: 2週間以内  
**カバレッジ目標**: 80%+

5. **routes/response/response.go**
   - 残りのHTTPステータスコード
   - すべてのエラータイプマッピング
   - 推定: 15+ test cases

6. **utils/env, utils/logger, utils/testutil**
   - ユーティリティ関数群
   - 推定: 20+ test cases

### Phase 4: Web Frontend (低優先)
**期限**: 3週間以内  
**カバレッジ目標**: 70%+

7. **web/src/client, web/src/handlers**
   - API通信
   - HTMX統合
   - 推定: 30+ test cases

---

## 📋 TDD Compliance Issues

### 違反事項

1. **テストカテゴリ不足**
   - 現状: 正常系のみ
   - 要求: 6カテゴリ (正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil)

2. **カバレッジ不足**
   - 現状: 28.6%
   - 要求: ≥80%
   - 不足: -51.4%

3. **TDDサイクル未遵守**
   - コードが先、テストが後（または無し）
   - 要求: テスト → コード → リファクタ

4. **テストファイル不足**
   - 現状: 26%のファイルのみ
   - 要求: すべてのプロダクションコードに対応

---

## 🚨 Risk Assessment

### Immediate Risks

1. **リファクタリング不可能**
   - Result monad変更で全システム停止の恐れ
   - エラー処理変更でデグレの恐れ

2. **バグ検出不可能**
   - ドメインロジックのバグが本番で発覚
   - エッジケース未検証

3. **保守性の低下**
   - 変更時の安全性ゼロ
   - レグレッション検出不可能

### Long-term Risks

1. **技術的負債の蓄積**
   - テストのない80%のコード
   - 修正コスト増大

2. **品質保証の欠如**
   - CI/CDが形骸化
   - デプロイリスク増大

3. **開発速度の低下**
   - 手動テストに時間浪費
   - バグ修正の繰り返し

---

## 💡 Recommendations

### Immediate Actions

1. **テスト作成の凍結解除**
   - 新機能開発を一時停止
   - テスト作成に集中

2. **TDDプロセスの再確立**
   - 🔴 RED → 🟢 GREEN → 🔵 REFACTOR → ✅ COMMIT
   - Pre-commit hookで80%未満をブロック

3. **ペアプログラミング**
   - TDD経験者と未経験者をペアリング
   - テスト駆動の文化を根付かせ

### Process Improvements

1. **CI/CDの強化**
   - カバレッジチェックを必須化
   - 80%未満でビルド失敗

2. **コードレビューの厳格化**
   - テストなしのPRは却下
   - 6カテゴリすべてをチェック

3. **定期的なレビュー**
   - 週次でカバレッジレポート
   - 月次でテスト品質監査

---

## 📚 Test Implementation Guide

### Result Monad テストの例

```go
func TestOk(t *testing.T) {
    tests := []struct {
        testName string
        args     args
        expected expected
    }{
        // 正常系
        {testName: "create Ok with int", args: args{value: 42}, expected: expected{isOk: true, value: 42}},
        {testName: "create Ok with string", args: args{value: "test"}, expected: expected{isOk: true, value: "test"}},
        
        // 境界値
        {testName: "Ok with zero", args: args{value: 0}, expected: expected{isOk: true, value: 0}},
        {testName: "Ok with max int", args: args{value: math.MaxInt64}, expected: expected{isOk: true, value: math.MaxInt64}},
        
        // 特殊文字
        {testName: "Ok with emoji", args: args{value: "🎉"}, expected: expected{isOk: true, value: "🎉"}},
        
        // 空文字
        {testName: "Ok with empty string", args: args{value: ""}, expected: expected{isOk: true, value: ""}},
        
        // Nil
        {testName: "Ok with nil slice", args: args{value: []int(nil)}, expected: expected{isOk: true}},
    }
    // ...
}
```

---

## 🎓 Learning Resources

1. **TDD Guidelines**: `.claude/rules/tdd.md`
2. **Testing Patterns**: `.claude/rules/testing.md`
3. **Table-Driven Tests**: Go公式ドキュメント
4. **Result Monad**: `apps/pkgs/types/result.go`

---

## 📝 Conclusion

プロジェクトは**TDD要件を深刻に違反**しています。特に：

- ✅ **2パッケージ (17%)** のみが合格
- ❌ **10パッケージ (83%)** が不合格
- 🔥 **3つのCRITICALパッケージ** (Result, AppError, Repository) が0%

**即座のアクション**が必要です。Phase 1 (Result monad, AppError) のテスト作成を最優先で実施してください。

---

**Next Steps**:
1. Phase 1のテスト作成を開始
2. TDDサイクルの遵守を徹底
3. 週次でカバレッジを再確認

