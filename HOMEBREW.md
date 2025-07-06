# Homebrew Tap Documentation

This repository serves as both the source code for evernote-cli and a Homebrew tap for easy installation on macOS.

## For Users

### Quick Installation

```bash
# One-liner installation
curl -fsSL https://raw.githubusercontent.com/cloudmanic/evernote-cli/main/install-homebrew.sh | bash
```

### Manual Installation

```bash
# Add the tap
brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli

# Install evernote-cli
brew install evernote-cli

# Verify installation
evernote-cli --help
```

### Updating

```bash
brew update
brew upgrade evernote-cli
```

### Uninstalling

```bash
brew uninstall evernote-cli
brew untap cloudmanic/evernote-cli  # Optional: remove the tap
```

## For Maintainers

### How the Tap Works

This repository uses the "repository as tap" approach, where the same GitHub repository contains both:

1. **Source Code**: Go source files, tests, and documentation
2. **Homebrew Formula**: Located in `Formula/evernote-cli.rb`

When users run `brew tap cloudmanic/evernote-cli`, Homebrew clones this repository and looks for formulas in the `Formula/` directory.

### Formula Maintenance

The formula is designed to work automatically with the existing release process:

- **No manual updates needed**: Uses GitHub's "latest" release endpoint
- **Automatic architecture detection**: Works on both Intel and Apple Silicon Macs
- **Binary distribution**: Downloads pre-built binaries instead of building from source

### Testing

Run the validation script to check the formula:

```bash
./validate-homebrew.sh
```

On macOS with Homebrew installed, you can also run:

```bash
./test-homebrew.sh
```

### Release Process

The existing GitHub Actions workflow automatically:

1. Builds binaries for macOS (Intel and Apple Silicon)
2. Creates GitHub releases with downloadable archives
3. The Homebrew formula automatically uses these releases

No additional steps are needed for Homebrew distribution when cutting releases.

### Formula Structure

The formula (`Formula/evernote-cli.rb`) includes:

- **Architecture Detection**: Automatically selects Intel or ARM binary
- **Download URLs**: Points to GitHub release assets
- **Installation Logic**: Extracts and installs the binary
- **Tests**: Verifies the installation works correctly

### Troubleshooting

If users report installation issues:

1. **Check release assets**: Ensure the latest release has the expected macOS binaries
2. **Verify URLs**: Check that the download URLs in the formula match the actual release assets
3. **Test locally**: Run the validation and test scripts
4. **Check architecture**: Ensure both Intel and ARM binaries are available

### Future Improvements

Potential enhancements:

1. **Semantic Versioning**: Update the release process to use semantic versions
2. **Checksums**: Add SHA256 checksums to the formula for better security
3. **Official Tap**: Consider submitting to homebrew-core for wider distribution
4. **Cask Support**: If a GUI version is ever created, consider Homebrew Cask

## Resources

- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Best Practices](https://docs.brew.sh/Homebrew-and-Python)