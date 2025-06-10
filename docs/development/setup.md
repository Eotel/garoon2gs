# 開発環境セットアップガイド

Garoon2GSの開発環境を構築する手順を説明します。

## 前提条件

### 必須ソフトウェア

- **Go**: 1.19以上
- **Git**: 2.0以上
- **Make**: GNU Make 3.81以上

### 推奨ソフトウェア

- **Docker**: テスト環境の構築用
- **VS Code**: 推奨エディタ
- **golangci-lint**: コード品質チェック

## セットアップ手順

### 1. リポジトリのクローン

```bash
git clone https://github.com/Eotel/garoon2gs.git
cd garoon2gs
```

### 2. Go環境の確認

```bash
# Goのバージョン確認
go version

# Go環境の確認
go env
```

### 3. 依存関係のインストール

```bash
# Goモジュールの依存関係をダウンロード
go mod download

# 依存関係の検証
go mod verify
```

### 4. Git Hooksのセットアップ

```bash
# Git hooks インストールスクリプトの実行
./scripts/install-hooks.sh
```

これにより以下のhooksが設定されます：
- `pre-commit`: コードフォーマットとlintチェック

### 5. 開発用ツールのインストール

```bash
# golangci-lint（macOS）
brew install golangci-lint

# golangci-lint（その他）
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## 開発用設定ファイル

### 1. `.env`ファイルの作成

```bash
# サンプルファイルをコピー
cp .env.sample .env.development

# 開発用の設定を編集
vim .env.development
```

開発用の`.env`設定例：

```bash
# Garoon設定（開発環境）
GAROON_BASE_URL="https://dev.cybozu.com/g"
GAROON_USERNAME="testuser"
GAROON_PASSWORD="testpass"

# Google Sheets設定（テスト用シート）
SPREADSHEET_ID="test-spreadsheet-id"
GOOGLE_SERVICE_ACCOUNT_FILE="test-service-account.json"

# デバッグ設定
DEBUG=true
LOG_LEVEL=debug
```

### 2. テスト用マッピングファイル

```bash
# テスト用ユーザーマッピング
cat > user_mapping_test.csv << EOF
1,テストユーザー1
2,テストユーザー2
EOF

# テスト用シートマッピング
cat > sheet_mapping_test.csv << EOF
2025,1,2025年1月
2025,2,2025年2月
EOF
```

## VS Code設定

### 推奨拡張機能

`.vscode/extensions.json`:
```json
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "streetsidesoftware.code-spell-checker",
    "wayou.vscode-todo-highlight"
  ]
}
```

### ワークスペース設定

`.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.formatOnSave": true,
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "go.testFlags": ["-v"],
  "go.testTimeout": "10s"
}
```

### デバッグ設定

`.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch garoon2gs",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}",
      "envFile": "${workspaceFolder}/.env.development",
      "args": []
    },
    {
      "name": "Launch list_users",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/list_users",
      "envFile": "${workspaceFolder}/.env.development"
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.v"]
    }
  ]
}
```

## ビルドとテスト

### ローカルビルド

```bash
# 開発用ビルド
go build -o garoon2gs .

# デバッグ情報付きビルド
go build -gcflags="all=-N -l" -o garoon2gs .

# 全プラットフォーム向けビルド
make build-all
```

### テストの実行

```bash
# 全テストの実行
go test ./...

# 詳細出力付きテスト
go test -v ./...

# カバレッジ付きテスト
go test -cover ./...

# 特定のテストのみ実行
go test -v -run TestScheduleWriter ./...
```

### コード品質チェック

```bash
# フォーマット
go fmt ./...

# より高度なフォーマット
goimports -w .

# Lint
golangci-lint run

# 脆弱性チェック
go list -json -m all | nancy sleuth
```

## 開発フロー

### 1. 機能開発

```bash
# 新機能用のブランチを作成
git checkout -b feature/new-feature

# 開発
# ... コードを編集 ...

# テストの実行
go test ./...

# コミット前のチェック
make pre-commit
```

### 2. デバッグ

```bash
# デバッグログを有効にして実行
DEBUG=true ./garoon2gs

# 特定のユーザーのみでテスト
./garoon2gs --users 1
```

### 3. プロファイリング

```bash
# CPUプロファイル
go run . -cpuprofile=cpu.prof

# メモリプロファイル
go run . -memprofile=mem.prof

# プロファイルの確認
go tool pprof cpu.prof
```

## トラブルシューティング

### 依存関係の問題

```bash
# モジュールキャッシュをクリア
go clean -modcache

# 依存関係を再ダウンロード
go mod download
```

### ビルドエラー

```bash
# ビルドキャッシュをクリア
go clean -cache

# 詳細なビルドログ
go build -v -x .
```

## 便利なMakeターゲット

```bash
# 開発用のよく使うコマンド
make dev        # ビルドして実行
make test       # テスト実行
make lint       # Lintチェック
make fmt        # コードフォーマット
make clean      # ビルド成果物のクリーンアップ
```

## 次のステップ

- [テストガイド](./testing.md) - テストの書き方
- [アーキテクチャ](./architecture.md) - システム設計の理解
- [リリースプロセス](./release.md) - リリース手順