# Garoon to Google Sheets

GaroonのスケジュールをGoogle Sheetsに同期するツールです。複数ユーザーの予定を取得し、指定されたスプレッドシートに書き込みます。

## 機能

- 複数ユーザーのスケジュール取得と書き込み
- スプレッドシートのヘッダー名に基づく列マッピング
- 月別シートの自動マッピング
- 休暇・外出などのイベント種別の自動判定
- IPアドレス制限環境向けのクライアント証明書ファイル読み込み対応
- 過去日付の上書き防止機能

## 必要条件

- Go 1.24以上
- Garoonアカウント
- Google Cloud Platform のサービスアカウント
- 書き込み先のGoogle Spreadsheet

## インストール

### バイナリをダウンロード

[リリースページ](https://github.com/eotel/garoon2gs/releases)から、ご利用のプラットフォーム用のバイナリをダウンロードしてください。

### ソースからビルド

```bash
# リポジトリをクローン
git clone https://github.com/eotel/garoon2gs.git
cd garoon2gs

# 通常ビルド
make build

# 全プラットフォーム向けにビルド
make build-all

# バージョン情報の確認
./garoon2gs -version
```

## 設定

1. `.env.sample`を`.env`にコピーし、必要な情報を設定します：

```env
GAROON_BASE_URL="https://<your-subdomain>.cybozu.com/g"
GAROON_USERNAME="<your-username>"
GAROON_PASSWORD="<your-password>"
SPREADSHEET_ID="<your-spreadsheet-id>"
GOOGLE_SERVICE_ACCOUNT_FILE="<your-service-account-file>.json"
# CLIENT_CERT_PATH="<your-client-cert-path>.pfx"
# CLIENT_CERT_PASSWORD="<your-client-cert-password>"
HOLIDAY_MENUS='["休み", "週休", "祝休日", "年次休暇"]'
OUTING_MENUS='["外出", "出張", "視察", "訪問"]'
NORMAL_PLACE="渋谷"
SHEET_MAPPING_PATH="sheet_mapping.csv"
HEADER_ROW=7
DATE_COL=A
USER_MAPPING_PATH="user_mapping.csv"
```

主な設定項目：

- `GAROON_BASE_URL`: Garoon のベース URL
- `GAROON_USERNAME` / `GAROON_PASSWORD`: Garoon のパスワード認証情報
- `SPREADSHEET_ID`: 書き込み先 Google スプレッドシート ID
- `GOOGLE_SERVICE_ACCOUNT_FILE`: Google サービスアカウント JSON のファイル名
- `SHEET_MAPPING_PATH`: 月とシート名の対応 CSV
- `HEADER_ROW`: 名前が並んでいるヘッダー行番号
- `DATE_COL`: 日付列の列名
- `USER_MAPPING_PATH`: Garoon ユーザー ID とヘッダー名の対応 CSV
- `HOLIDAY_MENUS`: 休暇扱いにする Garoon メニュー名の JSON 配列
- `OUTING_MENUS`: 外出扱いにする Garoon メニュー名の JSON 配列
- `NORMAL_PLACE`: 予定がない日に書き込む通常勤務地

2. `sheet_mapping.csv` でシート名のマッピングを設定します。
実装上の形式は `month,sheet_name` で、月は `YYYY-MM` です。

```csv
month,sheet_name
2025-01,R6年度_1月
2025-02,R6年度_2月
...
```

3. `user_mapping.csv` で Garoon ユーザー ID とスプレッドシートのヘッダー名を対応付けます。
ヘッダーは `user_id,name` 固定です。

```csv
user_id,name
2,田中
3,佐々木
...
```

## 使用方法

```bash
# 実行
./garoon2gs
```

利用可能なオプションは `-version` のみです。

```bash
./garoon2gs -version
```

設定ファイルは、まず実行ファイルと同じディレクトリの `.env` を探し、見つからない場合はカレントディレクトリの `.env` を読み込みます。

実行時は、現在月の1日から 3 か月先の月末までの予定を取得します。

## 開発者向け設定

### Git Hooks

コミット前に自動的にコードの整形と静的解析を実行するGit Hooksを設定できます。

```bash
# Git Hooksをインストール
./scripts/install-hooks.sh
```

これにより、コミット前に以下のチェックが実行されます：
- `go fmt` によるコードの整形
- `go vet` による静的解析
- `golangci-lint` による高度な静的解析（インストールされている場合）

## スプレッドシートの要件

- `HEADER_ROW` で指定した行に各ユーザーの名前が設定されていること
- `DATE_COL` で指定した列に `1,2,3,...` の日番号が入力されていること
- `user_mapping.csv` の `name` がヘッダーの表示名と完全一致していること
- ユーザー列には以下の値が書き込まれます
- 通常勤務: `NORMAL_PLACE` の値
- 休暇: `週休`
- 外出・出張: `外出`

## 注意事項

- スプレッドシートのアクセス権限を適切に設定してください
- Google サービスアカウントに対象スプレッドシートの編集権限を付与してください
- Garoonの認証情報は安全に管理してください
- Garoon REST API の認証は現行実装ではパスワード認証です
- `CLIENT_CERT_PATH` と `CLIENT_CERT_PASSWORD` は IP アドレス制限環境で追加指定するためのものです

## ライセンス

MITライセンス
