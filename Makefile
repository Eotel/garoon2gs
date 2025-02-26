# Makefile for Garoon2GS

# Variables
BINARY_NAME=garoon2gs
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

# Installation text templates
define INSTALL_INTEL
Garoon2GS (Intel Mac version)

Installation:
1. Copy garoon2gs to any directory
2. Run "chmod +x garoon2gs" in terminal to make it executable
3. Right-click and select "Open" for first run

Usage:
See README.md for details.
endef
export INSTALL_INTEL

define INSTALL_SILICON
Garoon2GS (Apple Silicon version)

Installation:
1. Copy garoon2gs to any directory
2. Run "chmod +x garoon2gs" in terminal to make it executable
3. Right-click and select "Open" for first run

Usage:
See README.md for details.
endef
export INSTALL_SILICON

# Build targets
.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# Cross compile
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

# macOS code signing
.PHONY: sign-macos
sign-macos:
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime dist/${BINARY_NAME}_macos_arm64

# macOS release package and notarization
.PHONY: notarize-macos
notarize-macos:
	# Clean up
	rm -rf dist/release
	mkdir -p dist/release
	
	# Make binaries executable
	chmod +x dist/${BINARY_NAME}_macos_amd64
	chmod +x dist/${BINARY_NAME}_macos_arm64
	
	# Code sign binaries
	@echo "Signing binaries with secure options..."
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_arm64
	
	# Add README and configuration files to release directory
	@echo "Creating release package contents..."
	cp README.md dist/release/README.md
	cp .env.sample dist/release/
	cp sheet_mapping.csv dist/release/
	cp user_mapping.csv dist/release/
	
	# Package for Intel Mac
	@echo "Creating Intel Mac (x86_64) release package..."
	mkdir -p dist/release/intel
	cp dist/${BINARY_NAME}_macos_amd64 dist/release/intel/${BINARY_NAME}
	@echo "$${INSTALL_INTEL}" > dist/release/intel/INSTALL.txt
	
	# Package for Apple Silicon
	@echo "Creating Apple Silicon (arm64) release package..."
	mkdir -p dist/release/apple_silicon
	cp dist/${BINARY_NAME}_macos_arm64 dist/release/apple_silicon/${BINARY_NAME}
	@echo "$${INSTALL_SILICON}" > dist/release/apple_silicon/INSTALL.txt
	
	# Create disk image for Intel Mac
	@echo "Creating disk image for Intel Mac..."
	hdiutil create -volname "Garoon2GS-Intel" -srcfolder dist/release/intel -ov -format UDZO dist/release/${BINARY_NAME}-intel.dmg
	
	# Create disk image for Apple Silicon
	@echo "Creating disk image for Apple Silicon..."
	hdiutil create -volname "Garoon2GS-AppleSilicon" -srcfolder dist/release/apple_silicon -ov -format UDZO dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# Sign disk images
	@echo "Signing disk images..."
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/release/${BINARY_NAME}-intel.dmg
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# Submit Intel Mac DMG for notarization
	@echo "Submitting Intel Mac disk image for notarization..."
	xcrun notarytool submit dist/release/${BINARY_NAME}-intel.dmg --apple-id "${APPLE_ID}" --password "${APPLE_APP_PWD}" --team-id "${APPLE_TEAM_ID}" --wait
	
	# Submit Apple Silicon DMG for notarization
	@echo "Submitting Apple Silicon disk image for notarization..."
	xcrun notarytool submit dist/release/${BINARY_NAME}-apple_silicon.dmg --apple-id "${APPLE_ID}" --password "${APPLE_APP_PWD}" --team-id "${APPLE_TEAM_ID}" --wait
	
	# Staple notarization tickets to DMGs
	@echo "Stapling notarization tickets to disk images..."
	xcrun stapler staple -v dist/release/${BINARY_NAME}-intel.dmg
	xcrun stapler staple -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	# Verify stapled disk images
	@echo "Verifying stapled disk images..."
	xcrun stapler validate -v dist/release/${BINARY_NAME}-intel.dmg
	xcrun stapler validate -v dist/release/${BINARY_NAME}-apple_silicon.dmg
	
	@echo ""
	@echo "==========================================="
	@echo "Release package creation complete\!"
	@echo "Notarized disk images are available at:"
	@echo "Intel Mac: dist/release/${BINARY_NAME}-intel.dmg"
	@echo "Apple Silicon: dist/release/${BINARY_NAME}-apple_silicon.dmg"
	@echo "==========================================="

# Test
.PHONY: test
test:
	go test -v ./...

# Linter
.PHONY: lint
lint:
	go vet ./...

# Format
.PHONY: fmt
fmt:
	go fmt ./...

# Cleanup
.PHONY: clean
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/

# Release
.PHONY: release
release: clean build-all
	goreleaser release --rm-dist

# Test release (no upload)
.PHONY: release-test
release-test: clean
	goreleaser release --skip-publish --rm-dist

# Development cycle
.PHONY: dev
dev: fmt lint test build
	@echo "Development cycle completed"

# Help
.PHONY: help
help:
	@echo "Garoon2GS Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build binary for current platform"
	@echo "  build-all     - Build binaries for all platforms"
	@echo "  sign-macos    - Sign macOS binaries"
	@echo "  notarize-macos - Notarize macOS binaries"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  release       - Release with GoReleaser"
	@echo "  release-test  - Test GoReleaser release (no publish)"
	@echo "  dev           - Run format, lint, test, build"
	@echo ""
	@echo "Environment variables:"
	@echo "  APPLE_DEVELOPER_ID - Developer ID for macOS signing (e.g. 'Developer ID Application: Your Name')"
	@echo "  APPLE_ID        - Apple ID email (for notarization)"
	@echo "  APPLE_TEAM_ID   - Apple Developer Team ID (for notarization)"
	@echo "  APPLE_APP_PWD   - App-specific password (for notarization, generated in Apple ID settings)"
