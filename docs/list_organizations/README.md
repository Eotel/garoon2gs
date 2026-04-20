# list_organizations - Garoon組織一覧取得ツール

## 概要

`list_organizations`は、Garoonに登録されているすべての組織情報を取得して表示するコマンドラインツールです。組織構造の確認に使用します。

## 使い方

### 基本的な使い方

```bash
./list_organizations
```

### 出力形式

ツールは組織情報をJSON形式で出力します：

```json
{
  "organizations": [
    {
      "id": "1",
      "name": "営業部",
      "code": "sales"
    },
    {
      "id": "2",
      "name": "営業1課",
      "code": "sales1"
    },
    {
      "id": "3",
      "name": "営業2課",
      "code": "sales2"
    },
    {
      "id": "4",
      "name": "開発部",
      "code": "dev"
    },
    {
      "id": "5",
      "name": "システム開発課",
      "code": "sysdev"
    },
    ...
  ]
}
```

## 設定

`list_organizations`は、メインツールと同じ`.env`ファイルを使用します。Garoon REST API の実行には、以下のいずれかの認証設定が必要です：

- `GAROON_BASE_URL` - GaroonのベースURL
- `GAROON_AUTH_TYPE` - `password` または `oauth`。省略時は `password`
- `GAROON_USERNAME` - Garoonのユーザー名（`password` 認証時）
- `GAROON_PASSWORD` - Garoonのパスワード（`password` 認証時）
- `GAROON_BEARER_TOKEN` - OAuthアクセストークン（`oauth` 認証時）
- `CLIENT_CERT_PATH` - クライアント証明書のパス（IPアドレス制限環境の場合）
- `CLIENT_CERT_PASSWORD` - クライアント証明書のパスワード（IPアドレス制限環境の場合）

## 実行例

### 全組織を表示

```bash
$ ./list_organizations

{
  "organizations": [
    {
      "id": "1",
      "name": "株式会社Example",
      "code": "example"
    },
    {
      "id": "2",
      "name": "営業本部",
      "code": "sales_hq"
    },
    {
      "id": "3",
      "name": "東京営業部",
      "code": "tokyo_sales"
    },
    {
      "id": "4",
      "name": "大阪営業部",
      "code": "osaka_sales"
    },
    {
      "id": "5",
      "name": "開発本部",
      "code": "dev_hq"
    },
    {
      "id": "6",
      "name": "プロダクト開発部",
      "code": "product_dev"
    },
    {
      "id": "7",
      "name": "インフラ部",
      "code": "infra"
    }
  ]
}
```

## 活用例

### 部署ごとのユーザーマッピング作成

特定の部署のユーザーのみをGaroon2GSで同期したい場合：

1. `list_organizations`で部署IDを確認
2. `list_users --org [部署ID]`で部署のユーザーを確認
3. 表示されたユーザーIDを使って`user_mapping.csv`を作成
4. メインツールで同期を実行

### 組織構造の可視化

組織の階層構造を確認して、適切な権限設定や管理体制の把握に活用できます。

## 注意事項

- 大規模な組織構造の場合、表示に時間がかかることがあります
- 権限によっては一部の組織情報が取得できない場合があります

## トラブルシューティング

### 組織が表示されない場合

1. Garoonの組織情報閲覧権限があるか確認してください
2. `.env`ファイルの設定が正しいか確認してください

## 関連ドキュメント

- [メインツールのユーザーマニュアル](../garoon2gs/user_manual.md)
- [list_usersツール](../list_users/README.md)
