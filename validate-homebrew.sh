#!/bin/bash

# Basic validation script for Homebrew formula (platform independent)
# This script tests the evernote-cli Homebrew formula structure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMULA_PATH="$SCRIPT_DIR/Formula/evernote-cli.rb"

echo "🧪 Basic Homebrew Formula Validation"
echo "===================================="
echo ""

# Test 1: Check if formula file exists
echo "📋 Test 1: Formula file existence"
if [[ -f "$FORMULA_PATH" ]]; then
    echo "✅ Formula file exists: $FORMULA_PATH"
else
    echo "❌ Formula file not found: $FORMULA_PATH"
    exit 1
fi

# Test 2: Basic Ruby syntax check
echo ""
echo "📋 Test 2: Ruby syntax validation"
if command -v ruby > /dev/null; then
    if ruby -c "$FORMULA_PATH" > /dev/null 2>&1; then
        echo "✅ Ruby syntax is valid"
    else
        echo "❌ Ruby syntax errors found"
        ruby -c "$FORMULA_PATH"
        exit 1
    fi
else
    echo "⚠️  Ruby not available, skipping syntax check"
fi

# Test 3: Check formula structure
echo ""
echo "📋 Test 3: Formula structure validation"

required_elements=(
    "class EvernoteCli < Formula"
    "desc "
    "homepage "
    "license "
    "def install"
    "test do"
    "Hardware::CPU.intel?"
    "Hardware::CPU.arm?"
    "bin.install"
)

for element in "${required_elements[@]}"; do
    if grep -q "$element" "$FORMULA_PATH"; then
        echo "✅ Found: $element"
    else
        echo "❌ Missing: $element"
        exit 1
    fi
done

# Test 4: Check URL patterns
echo ""
echo "📋 Test 4: Download URL validation"

if grep -q "evernote-cli-darwin-amd64.tar.gz" "$FORMULA_PATH"; then
    echo "✅ Intel macOS URL pattern found"
else
    echo "❌ Intel macOS URL pattern missing"
    exit 1
fi

if grep -q "evernote-cli-darwin-arm64.tar.gz" "$FORMULA_PATH"; then
    echo "✅ ARM macOS URL pattern found"
else
    echo "❌ ARM macOS URL pattern missing"
    exit 1
fi

# Test 5: Installation script validation
echo ""
echo "📋 Test 5: Installation script validation"

INSTALL_SCRIPT="$SCRIPT_DIR/install-homebrew.sh"
if [[ -f "$INSTALL_SCRIPT" ]]; then
    echo "✅ Installation script exists"
    
    if [[ -x "$INSTALL_SCRIPT" ]]; then
        echo "✅ Installation script is executable"
    else
        echo "❌ Installation script is not executable"
        exit 1
    fi
    
    # Check script content
    required_script_elements=(
        "brew tap cloudmanic/evernote-cli"
        "brew install evernote-cli"
        "command -v brew"
        "evernote-cli --help"
    )
    
    for element in "${required_script_elements[@]}"; do
        if grep -q "$element" "$INSTALL_SCRIPT"; then
            echo "✅ Installation script contains: $element"
        else
            echo "❌ Installation script missing: $element"
            exit 1
        fi
    done
else
    echo "❌ Installation script not found"
    exit 1
fi

# Test 6: Check test script
echo ""
echo "📋 Test 6: Test script validation"

TEST_SCRIPT="$SCRIPT_DIR/test-homebrew.sh"
if [[ -f "$TEST_SCRIPT" ]]; then
    echo "✅ Test script exists"
    
    if [[ -x "$TEST_SCRIPT" ]]; then
        echo "✅ Test script is executable"
    else
        echo "❌ Test script is not executable"
        exit 1
    fi
else
    echo "❌ Test script not found"
    exit 1
fi

echo ""
echo "🎉 All basic validation tests passed!"
echo ""
echo "The Homebrew formula appears to be correctly structured."
echo "On macOS with Homebrew installed, users can install with:"
echo ""
echo "  # Add the tap"
echo "  brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli"
echo ""
echo "  # Install evernote-cli"
echo "  brew install evernote-cli"
echo ""
echo "Or use the one-liner installation script:"
echo "  curl -fsSL https://raw.githubusercontent.com/cloudmanic/evernote-cli/main/install-homebrew.sh | bash"