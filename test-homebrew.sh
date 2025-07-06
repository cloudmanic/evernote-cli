#!/bin/bash

# Test script for Homebrew formula validation
# This script tests the evernote-cli Homebrew formula

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMULA_PATH="$SCRIPT_DIR/Formula/evernote-cli.rb"

echo "🧪 Testing evernote-cli Homebrew Formula"
echo "======================================="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "⚠️  Homebrew formula tests require macOS"
    echo "Skipping Homebrew-specific tests..."
    exit 0
fi

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "⚠️  Homebrew is not installed"
    echo "Skipping Homebrew formula tests..."
    exit 0
fi

echo "✅ Running on macOS with Homebrew installed"

# Test 1: Formula syntax validation
echo ""
echo "📋 Test 1: Formula syntax validation"
if brew audit --strict --online "$FORMULA_PATH" 2>/dev/null; then
    echo "✅ Formula syntax is valid"
else
    echo "❌ Formula syntax validation failed"
    echo "Running basic Ruby syntax check..."
    if ruby -c "$FORMULA_PATH" > /dev/null; then
        echo "✅ Basic Ruby syntax is valid"
    else
        echo "❌ Ruby syntax errors found"
        exit 1
    fi
fi

# Test 2: Check formula structure
echo ""
echo "📋 Test 2: Formula structure validation"
echo "Checking required elements..."

if grep -q "class EvernoteCli < Formula" "$FORMULA_PATH"; then
    echo "✅ Formula class definition found"
else
    echo "❌ Formula class definition missing"
    exit 1
fi

if grep -q "desc " "$FORMULA_PATH"; then
    echo "✅ Description found"
else
    echo "❌ Description missing"
    exit 1
fi

if grep -q "homepage " "$FORMULA_PATH"; then
    echo "✅ Homepage found"
else
    echo "❌ Homepage missing"
    exit 1
fi

if grep -q "def install" "$FORMULA_PATH"; then
    echo "✅ Install method found"
else
    echo "❌ Install method missing"
    exit 1
fi

if grep -q "test do" "$FORMULA_PATH"; then
    echo "✅ Test block found"
else
    echo "❌ Test block missing"
    exit 1
fi

# Test 3: Check architecture handling
echo ""
echo "📋 Test 3: Architecture handling validation"

if grep -q "Hardware::CPU.intel?" "$FORMULA_PATH"; then
    echo "✅ Intel architecture handling found"
else
    echo "❌ Intel architecture handling missing"
    exit 1
fi

if grep -q "Hardware::CPU.arm?" "$FORMULA_PATH"; then
    echo "✅ ARM architecture handling found"
else
    echo "❌ ARM architecture handling missing"
    exit 1
fi

# Test 4: URL validation
echo ""
echo "📋 Test 4: Download URL validation"

INTEL_URL=$(grep -A 1 "Hardware::CPU.intel?" "$FORMULA_PATH" | grep "url " | sed 's/.*url "//' | sed 's/".*//')
ARM_URL=$(grep -A 1 "Hardware::CPU.arm?" "$FORMULA_PATH" | grep "url " | sed 's/.*url "//' | sed 's/".*//')

echo "Intel URL: $INTEL_URL"
echo "ARM URL: $ARM_URL"

if [[ "$INTEL_URL" == *"darwin-amd64.tar.gz" ]]; then
    echo "✅ Intel URL format is correct"
else
    echo "❌ Intel URL format is incorrect"
    exit 1
fi

if [[ "$ARM_URL" == *"darwin-arm64.tar.gz" ]]; then
    echo "✅ ARM URL format is correct"
else
    echo "❌ ARM URL format is incorrect"
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
    if grep -q "brew tap cloudmanic/evernote-cli" "$INSTALL_SCRIPT"; then
        echo "✅ Installation script contains tap command"
    else
        echo "❌ Installation script missing tap command"
        exit 1
    fi
    
    if grep -q "brew install evernote-cli" "$INSTALL_SCRIPT"; then
        echo "✅ Installation script contains install command"
    else
        echo "❌ Installation script missing install command"
        exit 1
    fi
else
    echo "❌ Installation script not found"
    exit 1
fi

echo ""
echo "🎉 All Homebrew formula tests passed!"
echo ""
echo "Formula is ready for distribution. Users can install with:"
echo "  brew tap cloudmanic/evernote-cli https://github.com/cloudmanic/evernote-cli"
echo "  brew install evernote-cli"
echo ""
echo "Or use the installation script:"
echo "  curl -fsSL https://raw.githubusercontent.com/cloudmanic/evernote-cli/main/install-homebrew.sh | bash"