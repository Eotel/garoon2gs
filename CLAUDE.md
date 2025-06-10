# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Garoon2GS Project Guidelines

## Project Overview
Garoon2GS synchronizes schedules from Cybozu Garoon to Google Sheets, tracking employee attendance and work locations. It fetches events from Garoon's REST API and updates spreadsheets based on configurable mappings.

## Build & Test Commands
- Build: `go build .`
- Test all: `go test ./...`
- Test specific file: `go test -v ./path/to/file_test.go`
- Test specific function: `go test -v -run TestFunctionName`
- Format code: `go fmt ./...`
- Lint: `go vet ./...`
- Build for release: `make release` (builds for all platforms)
- Build and sign for macOS: `make release-mac` (includes notarization)
- Install git hooks: `./scripts/install-hooks.sh`

## Architecture Overview

### Core Flow
1. `garoon2gs.go` loads environment configuration and user mappings
2. For each user, it fetches events from Garoon API (`internal/client`)
3. Events are categorized (normal/holiday/outing) based on menu types
4. Results are written to Google Sheets using date-to-sheet mappings

### Key Components
- **internal/client**: Garoon API client with authentication and pagination
- **internal/mapping**: CSV-based user-to-column mapping loader
- **schedule_writer.go**: Core logic for determining work status and updating sheets
- **sheet_mapper.go**: Maps dates to sheet names (monthly organization)
- **users/, organizations/**: Domain packages for Garoon entities

### Configuration Strategy
- Primary: Environment variables loaded from `.env` files
- Secondary: CSV files for user and sheet mappings
- Configuration directory: Executable location or current directory
- Required: Garoon credentials, Google service account, spreadsheet ID

## Code Style Guidelines
- Imports: Standard library first, then third-party packages, then local packages
- Error handling: Use detailed error messages with `fmt.Errorf` and context
- Comments: Use Japanese for user-facing documentation, English for code internals
- File structure: Package main for executables, internal packages for implementation
- Naming: CamelCase for exported names, camelCase for internal, use Japanese names where appropriate
- Error messages in Japanese for end users, English for development
- Environment variables for configuration
- Use pointers for optional values
- Test table pattern with descriptive test case names

## Testing Patterns
- Table-driven tests with `testCases` structure
- Mock environment variables using `t.Setenv()`
- Test both success and error cases
- Use descriptive test names in Japanese when testing user-facing features

## Key Dependencies
- `github.com/joho/godotenv`: Environment management
- `golang.org/x/crypto/pkcs12`: Certificate handling
- `google.golang.org/api/sheets/v4`: Google Sheets API

## Project Structure
- `/cmd` - Command executables
- `/internal` - Internal packages
- `/users`, `/organizations` - Domain-specific packages