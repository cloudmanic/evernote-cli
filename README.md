# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication, searching notes, listing notebooks, and listing tags.

## Installation

### Download Pre-built Binaries

The easiest way to install evernote-cli is to download a pre-built binary from the [GitHub Releases](https://github.com/cloudmanic/evernote-cli/releases) page.

Available platforms:
- **Linux**: x86_64 and ARM64
- **macOS**: Intel and Apple Silicon 
- **Windows**: x86_64 and ARM64

Download the appropriate archive for your platform, extract it, and run the binary. On Unix systems, you may need to make it executable:
```bash
chmod +x evernote-cli
```

### Build from Source

If you have Go installed, you can build from source:
```bash
git clone https://github.com/cloudmanic/evernote-cli.git
cd evernote-cli
go build .
```

## Initial Setup

Run `evernote-cli init` and follow the prompts to provide your Evernote developer client ID and secret. The command then opens a browser to authenticate and stores the resulting token together with your credentials in `~/.config/evernote/auth.json`.

## Searching

Search notes with:

```bash
evernote-cli search "your query"
```

Use `--json` to output the raw JSON returned by the API.

## Listing Notebooks

List all available notebooks with:

```bash
evernote-cli notebooks
```

This will display a formatted list of notebooks with their names and GUIDs. Use `--json` to output the raw JSON returned by the API.

## Listing Tags

List all available tags with:

```bash
evernote-cli tags
```

This will display a formatted list of tags with their names and GUIDs. Use `--json` to output the raw JSON returned by the API.

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

### Releases

This project uses GitHub Actions to automatically build and release binaries for multiple platforms whenever code is merged to the main branch. 

The release workflow:
1. Builds binaries for Linux, macOS, and Windows (both x86_64 and ARM64)
2. Creates compressed archives for distribution
3. Automatically creates a GitHub release with all binaries attached
4. Provides download instructions in the release notes

To skip creating a release for a particular commit, include `[skip release]` in the commit message.

### Test Structure

The project includes comprehensive unit tests for all components:

- `main_test.go` - Tests for the main function
- `cmd/root_test.go` - Tests for configuration management and authentication
- `cmd/auth_test.go` - Tests for OAuth2 authentication components
- `cmd/init_test.go` - Tests for initialization command logic
- `cmd/search_test.go` - Tests for search command functionality
- `cmd/notebooks_test.go` - Tests for notebooks command functionality
- `cmd/tags_test.go` - Tests for tags command functionality

Tests focus on testing individual functions and components in isolation, using mocking for external dependencies like HTTP calls and file I/O.

