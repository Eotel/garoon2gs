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

File Placement:
- The .env file, CSV files, service account JSON, and client certificates
  must be placed in the same directory as the executable
- Rename .env.sample to .env and configure settings as needed
- Make sure sheet_mapping.csv and user_mapping.csv are properly configured

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

File Placement:
- The .env file, CSV files, service account JSON, and client certificates
  must be placed in the same directory as the executable
- Rename .env.sample to .env and configure settings as needed
- Make sure sheet_mapping.csv and user_mapping.csv are properly configured

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

# Unsigned release (build only without signing or notarization)
.PHONY: release-unsigned
release-unsigned: clean build-all
	@echo "Creating unsigned release packages..."
	mkdir -p dist/release
	
	# Copy release files
	cp README.md dist/release/
	cp .env.sample dist/release/
	cp sheet_mapping.csv dist/release/
	cp user_mapping.csv dist/release/
	cp LICENSE dist/release/
	
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
	
	@echo ""
	@echo "==========================================="
	@echo "Unsigned release package creation complete!"
	@echo "Disk images are available at:"
	@echo "Intel Mac: dist/release/${BINARY_NAME}-intel.dmg"
	@echo "Apple Silicon: dist/release/${BINARY_NAME}-apple_silicon.dmg"
	@echo "==========================================="

# Official release with code signing and notarization
.PHONY: release
release: clean build-all
	@echo "Creating signed and notarized release packages..."
	mkdir -p dist/release
	
	# Make binaries executable
	chmod +x dist/${BINARY_NAME}_macos_amd64
	chmod +x dist/${BINARY_NAME}_macos_arm64
	
	# Code sign binaries
	@echo "Signing binaries with secure options..."
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_amd64
	codesign -s "${APPLE_DEVELOPER_ID}" --force --timestamp --options runtime -v dist/${BINARY_NAME}_macos_arm64
	
	# Copy release files
	cp README.md dist/release/
	cp .env.sample dist/release/
	cp sheet_mapping.csv dist/release/
	cp user_mapping.csv dist/release/
	cp LICENSE dist/release/
	
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
	
	# Submit DMGs for notarization
	@echo "Submitting disk images for notarization..."
	xcrun notarytool submit dist/release/${BINARY_NAME}-intel.dmg --apple-id "${APPLE_ID}" --password "${APPLE_APP_PWD}" --team-id "${APPLE_TEAM_ID}" --wait
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
	@echo "Signed and notarized release complete!"
	@echo "Distribution-ready disk images are available at:"
	@echo "Intel Mac: dist/release/${BINARY_NAME}-intel.dmg"
	@echo "Apple Silicon: dist/release/${BINARY_NAME}-apple_silicon.dmg"
	@echo "==========================================="

# Release with GoReleaser (requires GitHub token)
.PHONY: release-github
release-github: clean build-all
	goreleaser release --clean

# Test release (no upload)
.PHONY: release-test
release-test: clean
	goreleaser release --snapshot --clean

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
	@echo "  release       - Build, sign and notarize full release (requires Apple Developer ID)"
	@echo "  release-unsigned - Create unsigned release packages"
	@echo "  release-github - Release with GoReleaser (requires GitHub token)"
	@echo "  release-test  - Test GoReleaser release (no publish)"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  dev           - Run format, lint, test, build"
	@echo ""
	@echo "Environment variables:"
	@echo "  APPLE_DEVELOPER_ID - Developer ID for macOS signing (e.g. 'Developer ID Application: Your Name')"
	@echo "  APPLE_ID        - Apple ID email (for notarization)"
	@echo "  APPLE_TEAM_ID   - Apple Developer Team ID (for notarization)"
	@echo "  APPLE_APP_PWD   - App-specific password (for notarization, generated in Apple ID settings)"