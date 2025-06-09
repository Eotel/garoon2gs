# 設定リファレンス

Garoon2GSの詳細な設定方法について説明します。

## 環境変数一覧

### 必須環境変数

| 環境変数 | 説明 | 例 |
|----------|------|-----|
| `GAROON_BASE_URL` | GaroonのベースURL | `https://example.cybozu.com/g` |
| `SPREADSHEET_ID` | Google SheetsのスプレッドシートID | `1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms` |
| `GOOGLE_SERVICE_ACCOUNT_FILE` | Googleサービスアカウントの認証ファイルパス | `garoon2gs-service-account.json` |
| `HOLIDAY_MENUS` | 休暇として扱うイベントメニューのJSON配列 | `["休み", "週休", "年次休暇"]` |
| `OUTING_MENUS` | 外出として扱うイベントメニューのJSON配列 | `["外出", "出張", "訪問"]` |
| `NORMAL_PLACE` | 通常勤務の場所 | `渋谷` |
| `SHEET_MAPPING_PATH` | シートマッピングCSVファイルのパス | `sheet_mapping.csv` |
| `HEADER_ROW` | スプレッドシートのヘッダー行番号 | `7` |
| `DATE_COL` | スプレッドシートの日付列 | `A` |
| `USER_MAPPING_PATH` | ユーザーマッピングCSVファイルのパス | `user_mapping.csv` |

### 認証関連環境変数

#### Basic認証の場合

| 環境変数 | 説明 | 例 |
|----------|------|-----|
| `GAROON_USERNAME` | Garoonのユーザー名 | `admin` |
| `GAROON_PASSWORD` | Garoonのパスワード | `password123` |

#### クライアント証明書認証の場合

| 環境変数 | 説明 | 例 |
|----------|------|-----|
| `CLIENT_CERT_PATH` | クライアント証明書（PFX/PKCS#12形式）のパス | `client-cert.pfx` |
| `CLIENT_CERT_PASSWORD` | クライアント証明書のパスワード | `cert-password` |

## マッピングファイル

### user_mapping.csv

Garoonのユーザーとスプレッドシートの列を対応付けるファイルです。

#### 形式

```csv
user_id,header_name
12345,山田太郎
67890,田中花子
11111,佐藤次郎
```

- `user_id`: Garoonのユーザー識別子（数値）
- `header_name`: スプレッドシートのヘッダーに記載されている名前

#### 作成方法

1. `list_users`コマンドでユーザー一覧を取得
2. 必要なユーザーのIDと名前を抽出
3. CSVファイルとして保存

### sheet_mapping.csv

年月とスプレッドシートのシート名を対応付けるファイルです。

#### 形式

```csv
year,month,sheet_name
2025,1,2025年1月
2025,2,2025年2月
2025,3,2025年3月
2025,4,2025年4月
2025,5,2025年5月
2025,6,2025年6月
2025,7,2025年7月
2025,8,2025年8月
2025,9,2025年9月
2025,10,2025年10月
2025,11,2025年11月
2025,12,2025年12月
```

- `year`: 年（4桁）
- `month`: 月（1-12）
- `sheet_name`: 対応するシート名

## イベントメニューの設定

### HOLIDAY_MENUS

休暇として扱うイベントのメニュー名を定義します。

```json
["休み", "週休", "祝休日", "年次休暇", "時間休暇", "夏季休暇", "年末年始休暇", "振休", "代休", "その他休暇"]
```

これらのメニューのイベントがある場合、セルに「休み」と記入されます。

### OUTING_MENUS

外出として扱うイベントのメニュー名を定義します。

```json
["外出", "出張", "視察", "訪問"]
```

これらのメニューのイベントがある場合、セルに「外出」と記入されます。

### 優先順位

1. 休暇イベント（HOLIDAY_MENUS）
2. 外出イベント（OUTING_MENUS）
3. 通常勤務（NORMAL_PLACE）

複数のイベントがある場合は、上記の優先順位で判定されます。

## 設定ファイルの配置

### 検索順序

Garoon2GSは以下の順序で設定ファイルを検索します：

1. 実行ファイルと同じディレクトリ
2. カレントディレクトリ

### ディレクトリ構成例

```
garoon2gs/
├── garoon2gs           # 実行ファイル
├── .env                # 環境設定
├── user_mapping.csv    # ユーザーマッピング
├── sheet_mapping.csv   # シートマッピング
└── service-account.json # Google認証ファイル
```

## 高度な設定

### カスタムイベントメニュー

組織独自のイベントメニューがある場合は、適切なカテゴリに追加してください：

```bash
# カスタム休暇メニューの追加例
HOLIDAY_MENUS='["休み", "週休", "特別休暇", "リフレッシュ休暇"]'

# カスタム外出メニューの追加例
OUTING_MENUS='["外出", "出張", "客先訪問", "セミナー参加"]'
```

### 複数環境の管理

開発環境と本番環境で異なる設定を使用する場合：

```bash
# 開発環境
cp .env.development .env
./garoon2gs

# 本番環境
cp .env.production .env
./garoon2gs
```