# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication and searching notes.

## Initial Setup

Run `evernote-cli init` and follow the prompts to provide your Evernote developer client ID and secret. The command then opens a browser to authenticate and stores the resulting token together with your credentials in `~/.config/evernote/auth.json`.

## Searching

Search notes with:

```bash
evernote-cli search "your query"
```

Use `--json` to output the raw JSON returned by the API.

## Development

### Running Tests

To run all tests:

```bash
go test ./...
```

To run tests with coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Or use the test script:

```bash
./test.sh
```

### Test Structure

The project includes comprehensive unit tests for all components:

- `main_test.go` - Tests for the main function
- `cmd/root_test.go` - Tests for configuration management and authentication
- `cmd/auth_test.go` - Tests for OAuth2 authentication components
- `cmd/init_test.go` - Tests for initialization command logic
- `cmd/search_test.go` - Tests for search command functionality

Tests focus on testing individual functions and components in isolation, using mocking for external dependencies like HTTP calls and file I/O.

