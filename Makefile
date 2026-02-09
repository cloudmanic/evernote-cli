# Copyright 2026. All rights reserved.
# Date: 2026-02-08

BINARY_NAME=evernote-cli
MODULE=$(shell go list -m)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test test-verbose coverage clean fmt vet lint install cross-build help

## build: Build the binary for the current platform
build:
	go build -o $(BINARY_NAME) .

## install: Build and install the binary to GOPATH/bin
install:
	go install .

## test: Run all tests
test:
	go test ./...

## test-verbose: Run all tests with verbose output
test-verbose:
	go test -v ./...

## coverage: Run tests with coverage and generate HTML report
coverage:
	go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "HTML report: coverage.html"

## fmt: Format all Go source files
fmt:
	go fmt ./...

## vet: Run Go vet on all packages
vet:
	go vet ./...

## lint: Run fmt and vet together
lint: fmt vet

## clean: Remove build artifacts and coverage files
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

## cross-build: Build for all release platforms
cross-build: clean
	GOOS=linux   GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o $(BINARY_NAME)-windows-arm64.exe .

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
