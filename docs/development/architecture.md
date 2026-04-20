# システムアーキテクチャ

Garoon2GSの内部構造と設計思想について説明します。

## 全体構成

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│   Garoon    │────▶│  Garoon2GS   │────▶│ Google Sheets  │
│  REST API   │     │              │     │     API        │
└─────────────┘     └──────────────┘     └────────────────┘
                           │
                    ┌──────┴──────┐
                    │   Mapping   │
                    │    Files    │
                    └─────────────┘
```

## コンポーネント構成

### 1. Main Application (`garoon2gs.go`)

アプリケーションのエントリーポイント。主な責務：
- 環境設定の読み込み
- 各コンポーネントの初期化
- メイン処理フローの制御

### 2. Internal Packages

#### `internal/client`
Garoon APIとの通信を担当：
- HTTP通信の管理
- 認証処理（パスワード認証、IPアドレス制限環境向けクライアント証明書）
- ページネーション処理
- エラーハンドリング

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
    auth       AuthMethod
}
```

#### `internal/mapping`
CSVファイルからのマッピング情報読み込み：
- ユーザーIDと列名の対応
- エラーハンドリング
- データ検証

### 3. Domain Packages

#### `users`
ユーザー関連の処理：
- ユーザー一覧の取得
- ユーザー情報の整形

#### `organizations`
組織関連の処理：
- 組織一覧の取得

### 4. Core Logic

#### `schedule_writer.go`
スケジュールデータの書き込みロジック：
- イベントの分類（通常/休暇/外出）
- Google Sheets APIとの通信
- セル位置の計算

#### `sheet_mapper.go`
日付とシート名のマッピング：
- 月別シートの管理
- 日本の会計年度対応

## データフロー

1. **設定読み込み**
   ```
   .env → Environment Variables
   CSV Files → Memory Maps
   ```

2. **データ取得**
   ```
   For each user:
     Garoon API → Events List
     Filter by date range
     Group by date
   ```

3. **データ変換**
   ```
   Events → Status (通常/休み/外出)
   Priority: Holiday > Outing > Normal
   ```

4. **データ書き込み**
   ```
   Date → Sheet Name (via mapping)
   User → Column (via mapping)
   Status → Cell Value
   ```

## エラーハンドリング戦略

### 1. Fail-Fast原則
重要な設定エラーは即座に終了：
- 必須環境変数の不足
- 認証情報の不正
- マッピングファイルの不在

### 2. 部分的な失敗の許容
個別ユーザーの処理失敗は継続：
- 特定ユーザーのAPI呼び出し失敗
- 特定セルの書き込み失敗

### 3. エラーメッセージの国際化
- エンドユーザー向け：日本語
- 開発者向け：英語

## 並行処理

現在の実装はシーケンシャル処理ですが、将来の拡張点：

```go
// 将来の実装案
var wg sync.WaitGroup
for _, user := range users {
    wg.Add(1)
    go func(u User) {
        defer wg.Done()
        processUser(u)
    }(user)
}
wg.Wait()
```

## 設定管理

### 環境変数の優先順位
1. システム環境変数
2. .envファイル（実行ファイルと同じディレクトリ）
3. .envファイル（カレントディレクトリ）

### 設定の検証
起動時に全ての必須設定を検証：
```go
func validateConfig(cfg *Config) error {
    if cfg.GaroonBaseURL == "" {
        return errors.New("GAROON_BASE_URL is required")
    }
    // 他の検証...
}
```

## セキュリティ考慮事項

1. **認証情報の保護**
   - パスワードはメモリ内でのみ保持
   - ログ出力時はマスキング

2. **通信の暗号化**
   - HTTPS通信の強制
   - 証明書検証の実施

3. **権限の最小化**
   - 必要最小限のAPIスコープ
   - 読み取り専用の操作を優先

## パフォーマンス最適化

### 現在の実装
- APIコール数：ユーザー数 × 1回
- バッチサイズ：100イベント/リクエスト

### 最適化の余地
1. **キャッシング**
   - ユーザー情報のキャッシュ
   - シートマッピングのキャッシュ

2. **バッチ処理**
   - 複数ユーザーのイベントを一括取得
   - Google Sheets APIのバッチ更新

3. **差分更新**
   - 前回実行時からの変更のみ更新
   - タイムスタンプベースの同期

## 拡張ポイント

### 1. プラグインアーキテクチャ
```go
type EventProcessor interface {
    Process(event Event) Status
}

type StatusWriter interface {
    Write(date time.Time, user User, status Status) error
}
```

### 2. 設定のDB化
- SQLiteでのローカル設定管理
- Web UIでの設定変更

### 3. 通知機能
- 処理完了通知
- エラー通知
- Slack/Teams連携
