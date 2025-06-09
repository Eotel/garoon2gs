# list_users - Garoonユーザー一覧取得ツール

## 概要

`list_users`は、Garoonに登録されているすべてのユーザー情報を取得して表示するコマンドラインツールです。ユーザーマッピングファイル（`user_mapping.csv`）の作成時に必要なユーザーIDと名前の確認に使用します。

## 使い方

### 基本的な使い方

```bash
./list_users
```

### 出力形式

ツールはユーザー情報をJSON形式で表示します：

```json
{
  "users": [
    {
      "id": "12345",
      "name": "山田太郎"
    },
    {
      "id": "67890",
      "name": "田中花子"
    },
    ...
  ]
}
```

### CSVファイル用の形式で出力

user_mapping.csv形式で出力する場合：

```bash
./list_users | jq -r '.users[] | [.id, .name] | @csv' > user_mapping_draft.csv
```

## 設定

`list_users`は、メインツールと同じ`.env`ファイルを使用します。以下の環境変数が必要です：

- `GAROON_BASE_URL` - GaroonのベースURL
- `GAROON_USERNAME` - Garoonのユーザー名（Basic認証の場合）
- `GAROON_PASSWORD` - Garoonのパスワード（Basic認証の場合）
- `CLIENT_CERT_PATH` - クライアント証明書のパス（証明書認証の場合）
- `CLIENT_CERT_PASSWORD` - クライアント証明書のパスワード（証明書認証の場合）

## 実行例

```bash
# 環境変数を設定して実行
$ ./list_users

{
  "users": [
    {
      "id": "1",
      "name": "Administrator"
    },
    {
      "id": "2",
      "name": "山田太郎"
    },
    {
      "id": "3",
      "name": "田中花子"
    },
    {
      "id": "4",
      "name": "佐藤次郎"
    },
    ...
  ]
}
```

## 注意事項

- 大量のユーザーが存在する場合、API制限により時間がかかることがあります
- 権限によっては一部のユーザー情報が取得できない場合があります
- 取得したユーザーIDは、`user_mapping.csv`で使用します

## トラブルシューティング

### 認証エラーが発生する場合

1. `.env`ファイルの認証情報が正しいか確認してください
2. Garoonの管理者権限またはユーザー情報の閲覧権限があるか確認してください

### ユーザーが表示されない場合

1. APIのアクセス権限を確認してください
2. Garoonのシステム管理者に問い合わせてください

## 関連ドキュメント

- [メインツールのユーザーマニュアル](../garoon2gs/user_manual.md)
- [ユーザーマッピングの設定方法](../garoon2gs/user_manual.md#ユーザーマッピングuser_mappingcsv)
