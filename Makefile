# Makefile for go-repo-manager
.PHONY: help build test lint clean install format deps check-deps run

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=go-repo-manager
BINARY_PATH=./$(BINARY_NAME)

# Go related variables
GOFILES=$(wildcard *.go)

# Install dependencies
deps: ## Install project dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies installed."

# Check for outdated dependencies
deps-check: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	@go list -u -m all

# Update dependencies
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Build the application
build: deps ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_PATH).exe ./cmd
	@echo "Binary built at $(BINARY_PATH).exe"

# Build for current platform
build-local: deps ## Build for current platform
	@echo "Building $(BINARY_NAME) for current platform..."
	@go build -o $(BINARY_NAME).exe ./cmd
	@echo "Binary built as $(BINARY_NAME).exe"

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format Go code
format: ## Format Go code using gofmt and goimports
	@echo "Formatting Go code..."
	@gofmt -s -w .
	@goimports -w . 2>/dev/null || "$(shell go env GOPATH)/bin/goimports.exe" -w .
	@echo "Code formatted."


# Check if golangci-lint is installed
check-linter-deps: ## Check if required tools are installed
	@echo "Checking dependencies..."
	@echo "GOPATH: $(shell go env GOPATH)"
	@golangci-lint version > /dev/null 2>&1 || echo "golangci-lint not found in PATH, will use GOPATH/bin version"
	@echo "Dependencies check completed."

# Run linter
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run 2>/dev/null || "$(shell go env GOPATH)/bin/golangci-lint.exe" run

# Fix linter issues automatically where possible
lint-fix: ## Fix linter issues automatically
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix 2>/dev/null || "$(shell go env GOPATH)/bin/golangci-lint.exe" run --fix

# Install golangci-lint
install-linter: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "golangci-lint installed successfully."

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@if exist $(BINARY_NAME).exe del $(BINARY_NAME).exe
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html
	@go clean

# Install the binary to GOPATH/bin
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd
	@echo "$(BINARY_NAME) installed successfully."

# Run the application with default arguments
run: build-local ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME).exe --help


# Development workflow: format, lint, test, build
dev: format lint test build ## Run full development workflow (format, lint, test, build)
	@echo "Development workflow completed successfully!"

# CI workflow: check formatting, lint, test
ci: deps ## Run CI workflow (deps, lint, test)
	@echo "Running CI workflow..."
	@echo "Checking if code is formatted..."
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make format'" && exit 1)
	@echo "Running linter..."
	@golangci-lint run 2>/dev/null || "$(shell go env GOPATH)/bin/golangci-lint.exe" run
	@echo "Running tests..."
	@go test -v ./...
	@echo "CI workflow completed successfully!"

# Security scan
security: ## Run security scan using gosec
	@echo "Running security scan..."
	@gosec ./...

# Generate documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all > docs.txt
	@echo "Documentation generated in docs.txt"
