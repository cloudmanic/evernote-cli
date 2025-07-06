#!/bin/bash

# Evernote CLI Test Runner
# This script runs all tests with coverage reporting

set -e

echo "Running all tests..."
go test ./... -v

echo ""
echo "Running tests with coverage..."
go test ./... -coverprofile=coverage.out

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out

echo ""
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "Tests completed successfully!"
echo "Coverage report generated: coverage.html"
echo "To view the coverage report, open coverage.html in a web browser"