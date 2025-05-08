# Garoon2GS ユーザーマニュアル

## 目次
1. [概要](#概要)
2. [前提条件](#前提条件)
3. [インストール方法](#インストール方法)
4. [設定方法](#設定方法)
   - [環境変数](#環境変数)
   - [マッピングファイル](#マッピングファイル)
5. [認証情報の設定](#認証情報の設定)
   - [Garoon認証](#garoon認証)
   - [Google Sheets認証](#google-sheets認証)
6. [スプレッドシートの形式](#スプレッドシートの形式)
7. [実行方法](#実行方法)
8. [トラブルシューティング](#トラブルシューティング)
9. [Garoon APIリファレンス](#garoon-apiリファレンス)

## 概要

Garoon2GSは、サイボウズGaroonのスケジュールデータをGoogle Sheetsに自動的に転記するツールです。複数ユーザーの予定情報をGaroonから抽出し、構造化されたスプレッドシートに書き込みます。休暇、外出、通常勤務などの異なるイベントタイプを適切にマッピングして表示します。

## 前提条件

- Garoonのアカウント（管理者権限または対象ユーザーのスケジュール閲覧権限が必要）
- Google Cloud Platformのプロジェクトとサービスアカウント
- Google Sheets APIの有効化
- 適切な形式のスプレッドシート（詳細は[スプレッドシートの形式](#スプレッドシートの形式)セクションを参照）

## インストール方法

### バイナリからのインストール

1. [リリースページ](https://github.com/Eotel/garoon2gs/releases)から、お使いのプラットフォーム（Windows、macOS、Linux）に合ったバイナリをダウンロードします。
2. ダウンロードしたファイルを解凍します。
3. 必要に応じて、実行ファイルをPATHの通ったディレクトリに配置します。

### ソースからのビルド

```bash
git clone https://github.com/Eotel/garoon2gs.git
cd garoon2gs
make build
```

ビルドされたバイナリは`bin`ディレクトリに生成されます。

## 設定方法

### 環境変数

Garoon2GSは`.env`ファイルから設定を読み込みます。`.env.sample`ファイルをコピーして`.env`ファイルを作成し、必要な情報を設定してください。

```
GAROON_BASE_URL="https://<your-subdomain>.cybozu.com/g"
GAROON_USERNAME="<your-username>"
GAROON_PASSWORD="<your-password>"
SPREADSHEET_ID="<your-spreadsheet-id>"
GOOGLE_SERVICE_ACCOUNT_FILE="<your-service-account-file>.json"
#CLIENT_CERT_PATH="<your-client-cert-path>.pfx"
#CLIENT_CERT_PASSWORD="<your-client-cert-password>"
HOLIDAY_MENUS='["休み", "週休", "祝休日", "年次休暇", "時間休暇", "夏季休暇", "年末年始休暇", "振休", "代休", "その他休暇"]'
OUTING_MENUS='["外出", "出張", "視察", "訪問"]'
NORMAL_PLACE="渋谷"
SHEET_MAPPING_PATH="sheet_mapping.csv"
HEADER_ROW=7
DATE_COL=A
USER_MAPPING_PATH="user_mapping.csv"
```

各環境変数の説明：

| 環境変数 | 説明 | 必須 |
|----------|------|------|
| GAROON_BASE_URL | GaroonのベースURL | ✓ |
| GAROON_USERNAME | Garoonのユーザー名 | ✓（クライアント証明書認証を使用しない場合） |
| GAROON_PASSWORD | Garoonのパスワード | ✓（クライアント証明書認証を使用しない場合） |
| SPREADSHEET_ID | Google SheetsのスプレッドシートID | ✓ |
| GOOGLE_SERVICE_ACCOUNT_FILE | Google Cloud Platformのサービスアカウントキーファイルのパス | ✓ |
| CLIENT_CERT_PATH | クライアント証明書（PFX形式）のパス | ✓（クライアント証明書認証を使用する場合） |
| CLIENT_CERT_PASSWORD | クライアント証明書のパスワード | ✓（クライアント証明書認証を使用する場合） |
| HOLIDAY_MENUS | 休暇として扱うイベントメニューのJSON配列 | ✓ |
| OUTING_MENUS | 外出として扱うイベントメニューのJSON配列 | ✓ |
| NORMAL_PLACE | 通常勤務の場所（例：「渋谷」） | ✓ |
| SHEET_MAPPING_PATH | シートマッピングCSVファイルのパス | ✓ |
| HEADER_ROW | ヘッダー行の番号（1から始まる） | ✓ |
| DATE_COL | 日付列のアルファベット（A, B, C, ...） | ✓ |
| USER_MAPPING_PATH | ユーザーマッピングCSVファイルのパス | ✓ |

### マッピングファイル

#### シートマッピング（sheet_mapping.csv）

月ごとのシート名を定義するCSVファイルです。以下の形式で作成してください：

```csv
year,month,sheet_name
2025,1,2025年1月
2025,2,2025年2月
...
```

#### ユーザーマッピング（user_mapping.csv）

GaroonのユーザーIDとスプレッドシートの列を対応付けるCSVファイルです。以下の形式で作成してください：

```csv
user_id,header_name
12345,伊藤
67890,田中
...
```

- `user_id`: GaroonのユーザーID
- `header_name`: スプレッドシートのヘッダーに表示されるユーザー名

## 認証情報の設定

### Garoon認証

#### ユーザー名/パスワード認証

`.env`ファイルに以下の情報を設定します：

```
GAROON_BASE_URL="https://<your-subdomain>.cybozu.com/g"
GAROON_USERNAME="<your-username>"
GAROON_PASSWORD="<your-password>"
```

#### クライアント証明書認証

クライアント証明書認証を使用する場合は、`.env`ファイルに以下の情報を設定します：

```
GAROON_BASE_URL="https://<your-subdomain>.cybozu.com/g"
CLIENT_CERT_PATH="<your-client-cert-path>.pfx"
CLIENT_CERT_PASSWORD="<your-client-cert-password>"
```

### Google Sheets認証

1. [Google Cloud Console](https://console.cloud.google.com/)でプロジェクトを作成します。
2. Google Sheets APIを有効化します。
3. サービスアカウントを作成し、キーファイル（JSON形式）をダウンロードします。
4. `.env`ファイルに以下の情報を設定します：

```
SPREADSHEET_ID="<your-spreadsheet-id>"
GOOGLE_SERVICE_ACCOUNT_FILE="<your-service-account-file>.json"
```

5. スプレッドシートの共有設定で、サービスアカウントのメールアドレスに編集権限を付与します。

## スプレッドシートの形式

Garoon2GSは、以下の形式のスプレッドシートを前提としています：

1. 各月ごとに別のシートがあり、シート名は`sheet_mapping.csv`で定義されています。
2. ヘッダー行（`HEADER_ROW`で指定）には、ユーザー名が含まれています。
3. 日付列（`DATE_COL`で指定）には、日付が入力されています。
4. 各ユーザーの列は、`user_mapping.csv`で定義されています。

例：

```
      | A | B | C | ... | J    | K    | L    |
------+---+---+---+-----+------+------+------+
1     | 2025年 | 5月 | STUDIO A | ... | 伊藤 | 乙戸 | 島田 |
...   |   |   |   |     |      |      |      |
7     | 1 | 木 |   |     | 週休 |      |      |
8     | 2 | 金 |   |     | 週休 |      |      |
...   |   |   |   |     |      |      |      |
```

## 実行方法

設定ファイルを準備した後、以下のコマンドでGaroon2GSを実行します：

```bash
./garoon2gs
```

デフォルトでは、現在の月から3ヶ月先までのスケジュールを取得します。特定の期間を指定する場合は、以下のオプションを使用します：

```bash
./garoon2gs --start-date 2025-01-01 --end-date 2025-12-31
```

特定のユーザーのみを対象とする場合は、以下のオプションを使用します：

```bash
./garoon2gs --users 12345,67890
```

## トラブルシューティング

### よくある問題と解決策

1. **認証エラー**
   - Garoonの認証情報が正しいか確認してください。
   - クライアント証明書のパスワードが正しいか確認してください。
   - Google Sheetsのサービスアカウントに適切な権限が付与されているか確認してください。

2. **シートが見つからない**
   - `sheet_mapping.csv`の設定が正しいか確認してください。
   - スプレッドシートIDが正しいか確認してください。

3. **ユーザーの列が見つからない**
   - `user_mapping.csv`の設定が正しいか確認してください。
   - スプレッドシートのヘッダー行に該当するユーザー名が含まれているか確認してください。

4. **予定が一部しか入力されない**
   - 過去の日付はスキップされる仕様になっています。現在の月の日付でも、実行日より前の日付はスキップされます。
   - 詳細は`docs/TODO.md`を参照してください。

## Garoon APIリファレンス

Garoon2GSは、Garoon REST APIを使用してスケジュールデータを取得しています。APIの詳細については、以下のリソースを参照してください：

- [Garoon REST API リファレンス](https://developer.cybozu.io/hc/ja/articles/360000503586-Garoon-REST-API%E4%B8%80%E8%A6%A7)
- [Garoon スケジュールAPI](https://developer.cybozu.io/hc/ja/articles/360000440583-Garoon-REST-API-%E3%82%B9%E3%82%B1%E3%82%B8%E3%83%A5%E3%83%BC%E3%83%AB)
- [Garoon ユーザーAPI](https://developer.cybozu.io/hc/ja/articles/360000503606-Garoon-REST-API-%E3%83%A6%E3%83%BC%E3%82%B6%E3%83%BC)

### APIの使い方

Garoon2GSでは、主に以下のAPIエンドポイントを使用しています：

1. **スケジュール取得**
   ```
   GET /api/v1/schedule/events?rangeStart={start}&rangeEnd={end}&target={user_id}
   ```

2. **ユーザー情報取得**
   ```
   GET /api/v1/base/users?offset={offset}&limit={limit}
   ```

3. **組織情報取得**
   ```
   GET /api/v1/base/organizations?offset={offset}&limit={limit}
   ```

APIの詳細な使用方法については、[Garoon API開発者サイト](https://developer.cybozu.io/hc/ja/categories/200157760-Garoon)を参照してください。
