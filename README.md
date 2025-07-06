# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication, searching notes, listing notebooks, and listing tags.

## Installation

### Homebrew (macOS)

The easiest way to install evernote-cli on macOS is using [Homebrew](https://brew.sh/):

```bash
# Add the evernote-cli tap
brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli

# Install evernote-cli
brew install evernote-cli
```

Alternatively, you can use the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/cloudmanic/evernote-cli/main/install-homebrew.sh | bash
```

#### How Homebrew Installation Works

The Homebrew formula downloads pre-built macOS binaries from the [GitHub Releases](https://github.com/cloudmanic/evernote-cli/releases) page. The formula automatically selects the correct binary for your Mac:

- **Intel Macs**: Downloads `evernote-cli-darwin-amd64.tar.gz`
- **Apple Silicon Macs**: Downloads `evernote-cli-darwin-arm64.tar.gz`

The formula installs the binary as `evernote-cli` in your `$PATH`, making it available from any terminal.

#### Updating via Homebrew

To update to the latest version:

```bash
brew update
brew upgrade evernote-cli
```

#### Uninstalling via Homebrew

To remove evernote-cli:

```bash
brew uninstall evernote-cli
brew untap cloudmanic/evernote-cli  # Optional: remove the tap
```

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

### Homebrew Formula Maintenance

The project includes a Homebrew formula located in `Formula/evernote-cli.rb` that allows users to install evernote-cli via `brew install`.

#### How the Homebrew Formula Works

The formula downloads pre-built macOS binaries from GitHub releases:

1. **Automatic Binary Selection**: The formula detects the user's Mac architecture (Intel or Apple Silicon) and downloads the appropriate binary
2. **Latest Release**: Uses GitHub's "latest" release endpoint to always get the most recent version
3. **Simple Installation**: Extracts and installs the binary to the user's `$PATH`

#### Formula Structure

```ruby
class EvernoteCli < Formula
  desc "CLI tool for interacting with Evernote"
  homepage "https://github.com/cloudmanic/evernote-cli"
  license "MIT"
  version "latest"

  # Downloads appropriate binary based on architecture
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/cloudmanic/evernote-cli/releases/latest/download/evernote-cli-darwin-amd64.tar.gz"
  elsif OS.mac? && Hardware::CPU.arm?
    url "https://github.com/cloudmanic/evernote-cli/releases/latest/download/evernote-cli-darwin-arm64.tar.gz"
  end

  def install
    # Install binary with correct name based on architecture
    if Hardware::CPU.intel?
      bin.install "evernote-cli-darwin-amd64" => "evernote-cli"
    elsif Hardware::CPU.arm?
      bin.install "evernote-cli-darwin-arm64" => "evernote-cli"
    end
  end

  test do
    # Verify installation works
    output = shell_output("#{bin}/evernote-cli --help")
    assert_match "A CLI tool to interact with Evernote", output
  end
end
```

#### Using the Formula

Users can install evernote-cli via Homebrew in two ways:

1. **Direct tap installation**:
   ```bash
   brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli
   brew install evernote-cli
   ```

2. **Installation script**:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/cloudmanic/evernote-cli/main/install-homebrew.sh | bash
   ```

#### Testing the Formula

To test the formula locally:

```bash
# Test the formula syntax
brew install --build-from-source --verbose --debug Formula/evernote-cli.rb

# Or audit the formula
brew audit --strict Formula/evernote-cli.rb
```

#### Formula Maintenance

The formula is designed to work automatically with the existing release process:

- **No manual updates needed**: The formula uses the "latest" release endpoint
- **Automatic architecture detection**: Works on both Intel and Apple Silicon Macs
- **Follows Homebrew best practices**: Uses proper naming, testing, and structure

When new releases are created by the GitHub Actions workflow, users can update by running:
```bash
brew update
brew upgrade evernote-cli
```

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

