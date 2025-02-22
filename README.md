# Garoon to Google Sheets

GaroonのスケジュールをGoogle Sheetsに同期するツールです。複数ユーザーの予定を取得し、指定されたスプレッドシートに書き込みます。

## 機能

- 複数ユーザーのスケジュール取得と書き込み
- スプレッドシートのヘッダー名に基づく列マッピング
- 月別シートの自動マッピング
- 休暇・外出などのイベント種別の自動判定
- クライアント証明書認証対応

## 必要条件

- Go 1.24以上
- Garoonアカウント
- Google Cloud Platform のサービスアカウント
- 書き込み先のGoogle Spreadsheet

## インストール

```bash
git clone https://github.com/eotel/garoon2gs.git
cd garoon2gs
go build .
```

## 設定

1. `.env.sample`を`.env`にコピーし、必要な情報を設定します：

```env
GAROON_BASE_URL="https://<your-subdomain>.cybozu.com/g"
GAROON_USERNAME="<your-username>"
GAROON_PASSWORD="<your-password>"
SPREADSHEET_ID="<your-spreadsheet-id>"
GOOGLE_SERVICE_ACCOUNT_FILE="<your-service-account-file>.json"
HOLIDAY_MENUS='["休み", "週休", "祝休日", "年次休暇"]'
OUTING_MENUS='["外出", "出張", "視察", "訪問"]'
NORMAL_PLACE="渋谷"
SHEET_MAPPING_PATH="sheet_mapping.csv"
HEADER_ROW=7
DATE_COL=A
USER_MAPPING_PATH="user_mapping.csv"
```

2. `sheet_mapping.csv`でシート名のマッピングを設定：

```csv
month,sheet_name
2025-01,R6年度_1月
2025-02,R6年度_2月
...
```

3. `user_mapping.csv`でユーザーと列のマッピングを設定：

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

# 開発用（環境変数を.env.devから読み込む）
./garoon2gs -env dev
```

## スプレッドシートの要件

- ヘッダー行に各ユーザーの名前が設定されていること
- DATE列に日付が入力されていること
- ユーザー列には以下の値が書き込まれます：
    - 通常勤務：指定された勤務地
    - 休暇："週休"
    - 外出・出張：指定されたメニューに応じて"外出"など

## 注意事項

- スプレッドシートのアクセス権限を適切に設定してください
- Garoonの認証情報は安全に管理してください
- クライアント証明書が必要な環境では適切に設定してください

## ライセンス

MITライセンス