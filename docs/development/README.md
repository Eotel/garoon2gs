# 開発者向けドキュメント

Garoon2GSの開発に関する情報をまとめています。

## 📚 ドキュメント一覧

- [TODO](./TODO.md) - 既知の問題と修正提案
- [アーキテクチャ](./architecture.md) - システム設計と内部構造
- [開発環境セットアップ](./setup.md) - 開発環境の構築方法
- [テストガイド](./testing.md) - テストの書き方と実行方法
- [リリースプロセス](./release.md) - リリース手順

## 🛠️ 開発の始め方

1. リポジトリをクローン
   ```bash
   git clone https://github.com/Eotel/garoon2gs.git
   cd garoon2gs
   ```

2. 依存関係のインストール
   ```bash
   go mod download
   ```

3. Git hooksのセットアップ
   ```bash
   ./scripts/install-hooks.sh
   ```

4. テストの実行
   ```bash
   go test ./...
   ```

## 📝 コーディング規約

[CLAUDE.md](../../CLAUDE.md)に記載されているコーディング規約に従ってください。

## 🤝 貢献方法

1. Issueを作成して問題や提案を共有
2. フォークしてブランチを作成
3. コードを修正してテストを追加
4. プルリクエストを送信

詳細は[CONTRIBUTING.md](../../CONTRIBUTING.md)を参照してください。