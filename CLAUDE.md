# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Development Commands

### Building
```bash
# Build for current platform
go build .

# Cross-platform builds (handled by GitHub Actions, but can be done locally)
GOOS=linux GOARCH=amd64 go build -o evernote-cli-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o evernote-cli-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o evernote-cli-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o evernote-cli-windows-amd64.exe
```

### Testing
```bash
# Run all tests with coverage
./test.sh

# Run a single test
go test -v -run TestFunctionName ./cmd/

# Run tests for a specific package
go test -v ./cmd/
```

### Linting and Type Checking
```bash
# Go has built-in formatting and vetting
go fmt ./...
go vet ./...
```

## Architecture Overview

### Command Structure
This is a CLI application built with Cobra framework. Each command lives in the `cmd/` directory and follows this pattern:

1. **Command files** (`cmd/[command].go`): Contains the command implementation
2. **Test files** (`cmd/[command]_test.go`): One test file per command file
3. **Registration**: Commands self-register via `init()` functions

### Authentication Flow
- OAuth2-based authentication stored in `~/.config/evernote/auth.json`
- Local HTTP server on port 8080 handles OAuth callbacks
- Token validation happens on each API call
- Browser-based flow with fallback URL display

### API Integration Pattern
- All API calls go through helper functions in `cmd/root.go`
- Consistent error handling with wrapped errors
- JSON output support via `--json` flag on all commands
- Custom Evernote API endpoint (not standard SDK)

### Testing Approach
- Mock HTTP clients for API calls
- Test both success and error cases
- Verify JSON formatting and command metadata
- One test file per source file (Go convention)

### Key Design Decisions
1. **Configuration Storage**: JSON file in user's config directory
2. **Error Handling**: All errors are wrapped with context
3. **Output Format**: Structured text by default, raw JSON with `--json` flag
4. **Command Pattern**: Each command is isolated and testable
5. **No External SDK**: Direct API calls instead of Evernote SDK

### Adding New Commands
1. Create new file in `cmd/` directory
2. Define command with `cobra.Command` struct
3. Implement command logic
4. Add `init()` function to register with root command
5. Create corresponding test file with comprehensive tests
6. Follow existing patterns for API calls and error handling

### Release Process
- Commits to main branch trigger automatic releases
- Use `[skip release]` in commit message to skip
- Releases include binaries for 6 platforms
- Homebrew formula automatically uses latest release