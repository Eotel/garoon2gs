# Makefile for Garoon2GS

# 変数の定義
BINARY_NAME=garoon2gs
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

# ビルドターゲット
.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# クロスコンパイル
.PHONY: build-all
build-all:
	# MacOS (Intel)
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_macos_amd64 .
	# MacOS (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_macos_arm64 .
	# Linux
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_linux_amd64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}_windows_amd64.exe .

# macOS のバイナリに署名
.PHONY: sign-macos
sign-macos:
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime dist/${BINARY_NAME}_macos_arm64

# macOSリリース用パッケージの作成と公証
.PHONY: notarize-macos
notarize-macos:
	# クリーンアップして始める
	rm -rf dist/release
	mkdir -p dist/release
	
	# バイナリを実行可能にする
	chmod +x dist/${BINARY_NAME}_macos_amd64
	chmod +x dist/${BINARY_NAME}_macos_arm64
	
	# バイナリをコード署名
	@echo "Signing binaries with secure options..."
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_arm64
	
	# インストール手順などを含めたREADMEをリリースディレクトリに追加
	@echo "Creating release package contents..."
	cp README.md dist/release/README.md
	cp .env.sample dist/release/
	cp sheet_mapping.csv dist/release/
	cp user_mapping.csv dist/release/
	
	# Intel Mac用パッケージング
	@echo "Creating Intel Mac (x86_64) release package..."
	mkdir -p dist/release/intel
	cp dist/${BINARY_NAME}_macos_amd64 dist/release/intel/${BINARY_NAME}
	cat > dist/release/intel/INSTALL.txt << EOF
Garoon2GS (Intel Mac版)

インストール手順:
1. ${BINARY_NAME}を任意のディレクトリにコピー
2. ターミナルで「chmod +x ${BINARY_NAME}」を実行して実行権限を付与
3. 初回実行時は右クリック→「開く」を選択して実行

使用方法:
詳細はREADME.mdをご参照ください。
EOF
	
	# Apple Silicon用パッケージング
	@echo "Creating Apple Silicon (arm64) release package..."
	mkdir -p dist/release/apple_silicon
	cp dist/${BINARY_NAME}_macos_arm64 dist/release/apple_silicon/${BINARY_NAME}
	cat > dist/release/apple_silicon/INSTALL.txt << EOF
Garoon2GS (Apple Silicon版)

インストール手順:
1. ${BINARY_NAME}を任意のディレクトリにコピー
2. ターミナルで「chmod +x ${BINARY_NAME}」を実行して実行権限を付与
3. 初回実行時は右クリック→「開く」を選択して実行

使用方法:
詳細はREADME.mdをご参照ください。
EOF
	
	# リリースパッケージ用のディスクイメージ作成（Intel用）
	@echo "Creating disk image for Intel Mac..."
	hdiutil create -volname "Garoon2GS-Intel" -srcfolder dist/release/intel -ov -format UDZO dist/release/${BINARY_NAME}-intel.dmg
	
	# リリースパッケージ用のディスクイメージ作成（Apple Silicon用）
	@echo "Creating disk image for Apple Silicon..."
	hdiutil create -volname "Garoon2GS-AppleSilicon" -srcfolder dist/release/apple_silicon -ov -format UDZO dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# ディスクイメージに署名
	@echo "Signing disk images..."
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/release/${BINARY_NAME}-intel.dmg
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# DMGを公証サービスに送信（Intel）
	@echo "Submitting Intel Mac disk image for notarization..."
	xcrun notarytool submit dist/release/${BINARY_NAME}-intel.dmg --apple-id "${APPLE_ID}" --password "${APPLE_APP_PWD}" --team-id "${APPLE_TEAM_ID}" --wait
	
	# DMGを公証サービスに送信（Apple Silicon）
	@echo "Submitting Apple Silicon disk image for notarization..."
	xcrun notarytool submit dist/release/${BINARY_NAME}-apple_silicon.dmg --apple-id "${APPLE_ID}" --password "${APPLE_APP_PWD}" --team-id "${APPLE_TEAM_ID}" --wait
	
	# 公証されたDMGにステープル
	@echo "Stapling notarization tickets to disk images..."
	xcrun stapler staple -v dist/release/${BINARY_NAME}-intel.dmg
	xcrun stapler staple -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# ステープルの検証
	@echo "Verifying stapled disk images..."
	xcrun stapler validate -v dist/release/${BINARY_NAME}-intel.dmg
	xcrun stapler validate -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	@echo ""
	@echo "==========================================="
	@echo "リリースパッケージの作成が完了しました！"
	@echo "公証済みのディスクイメージが以下の場所にあります:"
	@echo "Intel Mac版: dist/release/${BINARY_NAME}-intel.dmg"
	@echo "Apple Silicon版: dist/release/${BINARY_NAME}-apple_silicon.dmg"
	@echo "==========================================="

# テスト
.PHONY: test
test:
	go test -v ./...

# リンター
.PHONY: lint
lint:
	go vet ./...

# フォーマット
.PHONY: fmt
fmt:
	go fmt ./...

# クリーンアップ
.PHONY: clean
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/

# リリース
.PHONY: release
release: clean build-all
	goreleaser release --rm-dist

# リリースのテスト（アップロードなし）
.PHONY: release-test
release-test: clean
	goreleaser release --skip-publish --rm-dist

# 開発サイクル用のコマンド
.PHONY: dev
dev: fmt lint test build
	@echo "Development cycle completed"

# ヘルプ
.PHONY: help
help:
	@echo "Garoon2GS Makefile"
	@echo ""
	@echo "使用可能なターゲット:"
	@echo "  build         - 現在のプラットフォーム用のバイナリをビルド"
	@echo "  build-all     - 全プラットフォーム用のバイナリをビルド"
	@echo "  sign-macos    - macOS用バイナリに署名"
	@echo "  notarize-macos - macOS用バイナリをAppleの公証サービスに送信"
	@echo "  test          - テストを実行"
	@echo "  lint          - linterを実行"
	@echo "  fmt           - コードをフォーマット"
	@echo "  clean         - ビルド成果物をクリーンアップ"
	@echo "  release       - GoReleaserでリリース"
	@echo "  release-test  - GoReleaserでリリースをテスト（パブリッシュなし）"
	@echo "  dev           - フォーマット、リント、テスト、ビルドを実行"
	@echo ""
	@echo "環境変数:"
	@echo "  APPLE_DEVELOPER_ID - macOS署名用のDeveloper ID（例: 'Developer ID Application: Your Name'）"
	@echo "  APPLE_ID        - Apple IDのメールアドレス（公証用）"
	@echo "  APPLE_TEAM_ID   - Apple Developer TeamのID（公証用）"
	@echo "  APPLE_APP_PWD   - App固有のパスワード（公証用、アカウント設定から生成）"