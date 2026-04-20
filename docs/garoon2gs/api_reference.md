# Garoon API リファレンス

Garoon2GSが使用するGaroon REST APIの詳細情報です。

## API概要

Garoon2GSは、Garoon REST API v1を使用してスケジュールとユーザー情報を取得します。

### 認証方式

- **パスワード認証**: ユーザー名とパスワードを使用
- **クライアント証明書追加設定**: IPアドレス制限環境で PKCS#12 形式の証明書を追加指定

## 使用するAPIエンドポイント

### 1. スケジュールイベント取得API

#### エンドポイント

```http
GET /api/v1/schedule/events
```

#### パラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| rangeStart | datetime | ✓ | 取得期間の開始日時（ISO8601形式） |
| rangeEnd | datetime | ✓ | 取得期間の終了日時（ISO8601形式） |
| target | integer | ✓ | 対象ユーザーのID |
| limit | integer | - | 取得件数の上限（デフォルト: 100） |
| offset | integer | - | 取得開始位置（ページネーション用） |

#### レスポンス例

```json
{
  "events": [
    {
      "id": "12345",
      "subject": "定例会議",
      "start": {
        "dateTime": "2025-05-14T10:00:00+09:00"
      },
      "end": {
        "dateTime": "2025-05-14T11:00:00+09:00"
      },
      "eventType": "REGULAR",
      "eventMenu": "打合",
      "isAllDay": false,
      "attendees": [
        {
          "id": "1",
          "name": "山田太郎",
          "type": "USER"
        }
      ]
    }
  ],
  "hasNext": true
}
```

### 2. ユーザー一覧取得API

#### ユーザーエンドポイント

```http
GET /api/v1/base/users
```

#### ユーザーパラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| limit | integer | - | 取得件数の上限（デフォルト: 100、最大: 1000） |
| offset | integer | - | 取得開始位置（ページネーション用） |

#### ユーザーレスポンス例

```json
{
  "users": [
    {
      "id": "1",
      "code": "yamada",
      "name": "山田太郎",
      "email": "yamada@example.com",
      "phone": "03-1234-5678"
    },
    {
      "id": "2",
      "code": "tanaka",
      "name": "田中花子",
      "email": "tanaka@example.com",
      "phone": "03-1234-5679"
    }
  ],
  "hasNext": true
}
```

### 3. 組織一覧取得API

#### 組織エンドポイント

```http
GET /api/v1/base/organizations
```

#### 組織パラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| limit | integer | - | 取得件数の上限（デフォルト: 100、最大: 1000） |
| offset | integer | - | 取得開始位置（ページネーション用） |

#### 組織レスポンス例

```json
{
  "organizations": [
    {
      "id": "1",
      "code": "sales",
      "name": "営業部",
      "parentOrganization": null
    },
    {
      "id": "2",
      "code": "sales1",
      "name": "営業1課",
      "parentOrganization": "1"
    }
  ],
  "hasNext": false
}
```

### 4. 組織内ユーザー取得API

#### 組織ユーザーエンドポイント

```http
GET /api/v1/base/organizations/{organizationId}/users
```

#### 組織ユーザーパラメータ

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| organizationId | integer | ✓ | 組織ID（URLパス内） |
| limit | integer | - | 取得件数の上限 |
| offset | integer | - | 取得開始位置 |

## イベントタイプとメニュー

### eventType（イベントタイプ）

| 値 | 説明 |
|-----|------|
| REGULAR | 通常予定 |
| REPEAT | 繰り返し予定 |
| ALL_DAY | 終日予定 |

### eventMenu（イベントメニュー）

組織ごとにカスタマイズ可能です。一般的な例：

- 通常業務: `打合`, `会議`, `作業`
- 休暇関連: `休み`, `年次休暇`, `週休`
- 外出関連: `外出`, `出張`, `訪問`

## エラーハンドリング

### HTTPステータスコード

| コード | 説明 | 対処法 |
|--------|------|--------|
| 401 | 認証エラー | 認証情報を確認 |
| 403 | 権限エラー | アクセス権限を確認 |
| 404 | リソースが見つからない | URLやIDを確認 |
| 429 | レート制限 | しばらく待ってリトライ |
| 500 | サーバーエラー | Garoon管理者に連絡 |

### エラーレスポンス例

```json
{
  "code": "INVALID_REQUEST",
  "message": "The request parameters are invalid.",
  "errors": [
    {
      "code": "INVALID_VALUE",
      "message": "rangeStart must be before rangeEnd"
    }
  ]
}
```

## レート制限

- 1分あたり60リクエストまで
- 制限を超えると429エラーが返される
- `X-RateLimit-Remaining`ヘッダーで残りリクエスト数を確認可能

## ベストプラクティス

### 1. ページネーション

大量のデータを取得する際は、必ずページネーションを使用：

```go
offset := 0
limit := 100
for {
    response := fetchUsers(offset, limit)
    // データ処理
    if !response.HasNext {
        break
    }
    offset += limit
}
```

### 2. エラーリトライ

一時的なエラーに対してはリトライ処理を実装：

```go
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    response, err := apiCall()
    if err == nil {
        return response
    }
    if i < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(i+1))
    }
}
```

### 3. 日時の扱い

- 常にタイムゾーン付きのISO8601形式を使用
- 日本時間の場合: `+09:00`を付与
- 例: `2025-05-14T10:00:00+09:00`

## 参考リンク

- [Cybozu Developer Network - Garoon REST API](https://developer.cybozu.io/hc/ja/articles/360000503586)
- [Garoon REST API仕様書](https://developer.cybozu.io/hc/ja/categories/200157760)
