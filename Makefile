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
	codesign -s "${APPLE_DEVELOPER_ID}" dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" dist/${BINARY_NAME}_macos_arm64

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