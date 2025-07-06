#!/bin/bash

# Homebrew Tap Installation Script for evernote-cli
# This script helps users install evernote-cli via Homebrew

set -e

echo "🍺 Evernote CLI Homebrew Installation"
echo "====================================="
echo ""

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew is not installed!"
    echo ""
    echo "Please install Homebrew first by running:"
    echo '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
    echo ""
    echo "Then run this script again."
    exit 1
fi

echo "✅ Homebrew is installed"

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "⚠️  This formula is designed for macOS only."
    echo "For other platforms, please download binaries from:"
    echo "https://github.com/cloudmanic/evernote-cli/releases"
    exit 1
fi

echo "✅ Running on macOS"

# Add the tap
echo ""
echo "📦 Adding the evernote-cli tap..."
brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli

# Install evernote-cli
echo ""
echo "⬇️  Installing evernote-cli..."
brew install evernote-cli

echo ""
echo "🎉 Installation complete!"
echo ""
echo "You can now use evernote-cli by running:"
echo "  evernote-cli --help"
echo ""
echo "To get started:"
echo "  evernote-cli init"
echo ""
echo "For more information, visit:"
echo "  https://github.com/cloudmanic/evernote-cli"