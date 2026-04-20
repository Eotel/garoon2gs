# Makefileガイド

Garoon2GSプロジェクトのMakefileコマンドの使い方を説明します。

## 利用可能なコマンド

### ビルド関連

#### `make build`

現在のプラットフォーム向けにバイナリをビルドします。

```bash
make build
```

#### `make build-list-organizations`

`list_organizations` ツールをビルドします。

```bash
make build-list-organizations
```

#### `make build-list-users`

`list_users` ツールをビルドします。

```bash
make build-list-users
```

#### `make build-tools`

`list_organizations` と `list_users` の両方のツールをビルドします。

```bash
make build-tools
```

#### `make build-all-with-tools`

メインアプリとすべてのツールを現在のプラットフォーム向けにビルドします。

```bash
make build-all-with-tools
```

#### `make build-all`

すべてのサポートプラットフォーム向けにバイナリをビルドします。

```bash
make build-all
```

#### `make build-all-tools`

メインアプリと`list_organizations`、`list_users`のツールをすべてのサポートプラットフォーム向けにビルドします。

```bash
make build-all-tools
```

生成されるバイナリ：

- `dist/garoon2gs_darwin_amd64` - macOS (Intel)
- `dist/garoon2gs_darwin_arm64` - macOS (Apple Silicon)
- `dist/garoon2gs_linux_amd64` - Linux (64bit)
- `dist/garoon2gs_linux_386` - Linux (32bit)
- `dist/garoon2gs_windows_amd64.exe` - Windows (64bit)
- `dist/garoon2gs_windows_386.exe` - Windows (32bit)

### リリース関連

#### `make release`

全プラットフォーム向けにリリースビルドを作成します。

```bash
make release
```

#### `make release-mac`

macOS向けに署名・公証済みのリリースを作成します。

```bash
make release-mac
```

要件：

- Apple Developer証明書
- Xcodeツール
- 有効なApple Developer ID

#### `make dmg`

macOS向けのDMGインストーラーを作成します。

```bash
make dmg
```

### テスト・品質管理

#### `make test`

すべてのテストを実行します。

```bash
make test
```

#### `make test-verbose`

詳細な出力でテストを実行します。

```bash
make test-verbose
```

#### `make coverage`

テストカバレッジレポートを生成します。

```bash
make coverage
```

HTML形式のレポートが`coverage.html`として生成されます。

#### `make lint`

コードの静的解析を実行します。

```bash
make lint
```

#### `make fmt`

コードをフォーマットします。

```bash
make fmt
```

### 開発支援

#### `make dev`

開発用ビルドを作成して実行します。

```bash
make dev
```

#### `make run`

ビルドして即座に実行します。

```bash
make run
```

#### `make install-hooks`

Git の `pre-commit` / `pre-push` hooks をインストールします。

```bash
make install-hooks
```

### クリーンアップ

#### `make clean`

ビルド成果物とキャッシュをクリーンアップします。

```bash
make clean
```

削除されるもの：

- `dist/`ディレクトリ
- `coverage.out`, `coverage.html`
- ビルドキャッシュ

## Makefile変数

### VERSION

リリースバージョンを指定します。

```bash
make release VERSION=1.2.3
```

### GOOS / GOARCH

特定のプラットフォーム向けにビルドします。

```bash
make build GOOS=linux GOARCH=amd64
```

### BUILD_FLAGS

追加のビルドフラグを指定します。

```bash
make build BUILD_FLAGS="-ldflags '-X main.Version=dev'"
```

## よく使うコマンドの組み合わせ

### 開発時のワークフロー

```bash
# コードの変更後
make fmt         # フォーマット
make lint        # Lintチェック
make test        # テスト実行
make dev         # 開発ビルドと実行
```

### リリース前の確認

```bash
# すべてのチェックを実行
make fmt
make lint
make test
make coverage    # カバレッジ確認
make build-all   # 全プラットフォームビルド確認
```

### リリース作成

```bash
# バージョンを指定してリリース
make release VERSION=1.2.3

# macOS向け署名済みリリース
make release-mac VERSION=1.2.3
```

## トラブルシューティング

### ビルドエラー

1. **依存関係の問題**

   ```bash
   go mod download
   go mod tidy
   ```

2. **キャッシュの問題**

   ```bash
   make clean
   go clean -cache
   ```

### 署名エラー（macOS）

1. **証明書が見つからない**

   ```bash
   security find-identity -v -p codesigning
   ```

2. **権限の問題**

   ```bash
   sudo xcode-select --reset
   ```

## カスタマイズ

### 新しいターゲットの追加

`Makefile`に新しいターゲットを追加：

```makefile
.PHONY: my-target
my-target:
 @echo "Running my custom target"
 # カスタムコマンド
```

### プラットフォームの追加

`BUILD_TARGETS`に新しいプラットフォームを追加：

```makefile
BUILD_TARGETS += \
 GOOS=freebsd GOARCH=amd64 \
 GOOS=openbsd GOARCH=amd64
```
