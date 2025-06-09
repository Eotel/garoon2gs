# list_organizations - Garoon組織一覧取得ツール

## 概要

`list_organizations`は、Garoonに登録されているすべての組織情報を取得して表示するコマンドラインツールです。組織構造の確認や、特定の組織に所属するユーザーの把握に使用します。

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

### 特定の組織のユーザーを表示

組織IDを指定して、その組織に所属するユーザーを表示：

```bash
./list_organizations --org 5
```

## 設定

`list_organizations`は、メインツールと同じ`.env`ファイルを使用します。以下の環境変数が必要です：

- `GAROON_BASE_URL` - GaroonのベースURL
- `GAROON_USERNAME` - Garoonのユーザー名（Basic認証の場合）
- `GAROON_PASSWORD` - Garoonのパスワード（Basic認証の場合）
- `CLIENT_CERT_PATH` - クライアント証明書のパス（証明書認証の場合）
- `CLIENT_CERT_PASSWORD` - クライアント証明書のパスワード（証明書認証の場合）

## コマンドラインオプション

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `--org` | 特定の組織IDのユーザー情報を表示 | なし |

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

### 組織のユーザーを表示

```bash
$ ./list_organizations --org 3

{
  "users": [
    {
      "id": "123",
      "name": "山田太郎"
    },
    {
      "id": "456",
      "name": "田中花子"
    },
    {
      "id": "789",
      "name": "佐藤次郎"
    }
  ]
}
```

## 活用例

### 部署ごとのユーザーマッピング作成

特定の部署のユーザーのみをGaroon2GSで同期したい場合：

1. `list_organizations --org [部署ID]`で部署のユーザーを確認
2. 表示されたユーザーIDを使って`user_mapping.csv`を作成
3. メインツールで同期を実行

### 組織構造の可視化

組織の階層構造を確認して、適切な権限設定や管理体制の把握に活用できます。

## 注意事項

- 大規模な組織構造の場合、表示に時間がかかることがあります
- ユーザー表示オプションを使用すると、API呼び出し回数が増加します
- 権限によっては一部の組織情報が取得できない場合があります

## トラブルシューティング

### 組織が表示されない場合

1. Garoonの組織情報閲覧権限があるか確認してください
2. `.env`ファイルの設定が正しいか確認してください

### ユーザーが表示されない場合

1. `--org`オプションで正しい組織IDを指定しているか確認してください
2. 該当組織のユーザー情報閲覧権限があるか確認してください

## 関連ドキュメント

- [メインツールのユーザーマニュアル](../garoon2gs/user_manual.md)
- [list_usersツール](../list_users/README.md)
