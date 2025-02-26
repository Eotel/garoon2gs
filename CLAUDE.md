# Garoon2GS Project Guidelines

## Build & Test Commands
- Build: `go build .`
- Test all: `go test ./...`
- Test specific file: `go test -v ./path/to/file_test.go`
- Test specific function: `go test -v -run TestFunctionName`
- Format code: `go fmt ./...`
- Lint: `go vet ./...`

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

## Project Structure
- `/cmd` - Command executables
- `/internal` - Internal packages
- `/users`, `/organizations` - Domain-specific packages