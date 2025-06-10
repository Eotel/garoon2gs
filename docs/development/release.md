# リリースプロセス

Garoon2GSのリリース手順について説明します。

## リリースの種類

### セマンティックバージョニング

Garoon2GSは[セマンティックバージョニング](https://semver.org/lang/ja/)に従います：

- **メジャー (X.0.0)**: 後方互換性のない変更
- **マイナー (0.X.0)**: 後方互換性のある機能追加
- **パッチ (0.0.X)**: 後方互換性のあるバグ修正

## リリース準備

### 1. ブランチの準備

```bash
# mainブランチを最新に
git checkout main
git pull origin main

# リリースブランチの作成
git checkout -b release/v1.2.3
```

### 2. バージョン番号の更新

`version.go`ファイルを更新：

```go
// version.go
package main

const Version = "1.2.3"
```

### 3. CHANGELOGの更新

`CHANGELOG.md`を更新：

```markdown
# Changelog

## [1.2.3] - 2025-05-14

### Added
- 新機能の説明

### Changed
- 変更点の説明

### Fixed
- 修正したバグの説明

### Deprecated
- 非推奨となった機能

### Removed
- 削除された機能

### Security
- セキュリティ関連の修正
```

### 4. ドキュメントの更新

必要に応じて以下を更新：
- README.md
- docs/配下のドキュメント
- API仕様書

## ビルドとテスト

### 1. 全テストの実行

```bash
# ユニットテスト
go test ./...

# 統合テスト
go test -tags=integration ./...

# レースコンディションのチェック
go test -race ./...

# Lintチェック
golangci-lint run
```

### 2. ビルドの確認

```bash
# 全プラットフォーム向けビルド
make build-all

# ビルド結果の確認
ls -la dist/
```

### 3. 動作確認

```bash
# ローカルでの動作確認
./dist/garoon2gs_darwin_amd64 --version

# 基本的な動作確認
./dist/garoon2gs_darwin_amd64
```

## リリースの作成

### 1. コミットとタグ

```bash
# 変更をコミット
git add .
git commit -m "Release version 1.2.3"

# タグの作成
git tag -a v1.2.3 -m "Release version 1.2.3"

# プッシュ
git push origin release/v1.2.3
git push origin v1.2.3
```

### 2. プルリクエストの作成

```bash
# GitHub CLIを使用
gh pr create --base main --title "Release v1.2.3" --body "Release version 1.2.3

## Changes
- Feature 1
- Bug fix 2

## Checklist
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated"
```

### 3. マージとリリース

1. プルリクエストのレビューを受ける
2. 承認後、mainブランチにマージ
3. GitHub Releasesページで新しいリリースを作成

## GitHub Releasesの作成

### 1. リリースノートの作成

```markdown
# Garoon2GS v1.2.3

## 🎉 新機能
- ユーザーごとの同期設定が可能に
- エラー通知機能を追加

## 🐛 バグ修正
- 特定の日付でクラッシュする問題を修正
- 認証エラーのハンドリングを改善

## 📝 変更点
- パフォーマンスの向上（約30%高速化）
- ログ出力の改善

## 📦 ダウンロード

### macOS
- [garoon2gs_darwin_amd64.tar.gz](link) - Intel Mac
- [garoon2gs_darwin_arm64.tar.gz](link) - Apple Silicon

### Windows
- [garoon2gs_windows_amd64.zip](link) - 64bit
- [garoon2gs_windows_386.zip](link) - 32bit

### Linux
- [garoon2gs_linux_amd64.tar.gz](link) - 64bit
- [garoon2gs_linux_386.tar.gz](link) - 32bit

## インストール方法

詳細は[ユーザーマニュアル](docs/garoon2gs/user_manual.md)を参照してください。
```

### 2. アセットのアップロード

```bash
# GitHub CLIを使用したリリースの作成
gh release create v1.2.3 \
  --title "Garoon2GS v1.2.3" \
  --notes-file RELEASE_NOTES.md \
  dist/*
```

## macOS向けの署名と公証

### 1. コード署名

```bash
# 署名用の証明書IDを確認
security find-identity -v -p codesigning

# バイナリに署名
codesign --force --sign "Developer ID Application: Your Name (XXXXXXXXXX)" \
  --options runtime \
  --timestamp \
  dist/garoon2gs_darwin_amd64
```

### 2. 公証（Notarization）

```bash
# zipファイルの作成
zip -j garoon2gs_darwin_amd64.zip dist/garoon2gs_darwin_amd64

# 公証のアップロード
xcrun notarytool submit garoon2gs_darwin_amd64.zip \
  --apple-id "your-apple-id@example.com" \
  --password "app-specific-password" \
  --team-id "XXXXXXXXXX" \
  --wait

# 公証結果の確認
xcrun notarytool log <submission-id> \
  --apple-id "your-apple-id@example.com" \
  --password "app-specific-password"
```

### 3. DMGの作成（オプション）

```bash
# DMGの作成
make dmg

# DMGにも署名
codesign --force --sign "Developer ID Application: Your Name (XXXXXXXXXX)" \
  garoon2gs.dmg
```

## 自動化されたリリースプロセス

### GitHub Actionsを使用した自動化

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build all platforms
      run: make build-all
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: dist/*
        draft: false
        prerelease: false
        generate_release_notes: true
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## リリース後の作業

### 1. 動作確認

- 各プラットフォームでダウンロードして動作確認
- インストール手順の確認
- アップグレード手順の確認

### 2. アナウンス

- プロジェクトのWebサイトを更新
- ユーザーへの通知（メール、Slack等）
- ソーシャルメディアでの告知

### 3. 次バージョンの準備

```bash
# developブランチの作成/更新
git checkout -b develop
git merge main

# バージョン番号を次の開発版に更新
# version.go を "1.2.4-dev" に更新
```

## ホットフィックス

緊急の修正が必要な場合：

```bash
# 現在のリリースタグから分岐
git checkout -b hotfix/v1.2.4 v1.2.3

# 修正を実施
# ...

# バージョンを更新（パッチバージョンのみ）
# version.go を "1.2.4" に更新

# コミットとタグ
git commit -am "Hotfix: critical bug fix"
git tag -a v1.2.4 -m "Hotfix release v1.2.4"

# リリース
git push origin v1.2.4
```

## チェックリスト

リリース前の最終確認：

- [ ] 全テストがパスしている
- [ ] ドキュメントが最新
- [ ] CHANGELOGが更新されている
- [ ] バージョン番号が正しい
- [ ] ビルドが全プラットフォームで成功
- [ ] 基本的な動作確認完了
- [ ] リリースノートの準備完了
- [ ] 署名と公証（macOS）完了