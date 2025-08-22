# Makefile for lxc-go-cli

# Version configuration
VERSION_MAJOR := 1
VERSION_MINOR := $(shell git rev-parse --short=4 HEAD 2>/dev/null || echo "dev")
VERSION_PATCH := $(shell date +%Y%m%d%H%M%S)
VERSION := $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)

# Build flags
LDFLAGS := -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)

.PHONY: test test-unit test-integration test-all coverage clean build help version

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Unit tests (default - no external dependencies)
test: test-unit ## Run unit tests (default)

test-unit: ## Run unit tests only (no LXC required)
	@echo "Running unit tests..."
	go test ./...

# Integration tests (use mocks by default)
test-integration: ## Run integration tests only (uses mocks, no LXC required)
	@echo "Running integration tests with mocks..."
	go test -tags=integration ./...

# All tests
test-all: ## Run both unit and integration tests
	@echo "Running all tests..."
	go test -tags=integration ./...

# Coverage reports
coverage: ## Run unit tests with coverage
	@echo "Running unit tests with coverage..."
	go test ./... -cover

coverage-detailed: ## Run unit tests with detailed coverage report
	@echo "Running unit tests with detailed coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-integration: ## Run integration tests with coverage
	@echo "Running integration tests with coverage..."
	go test -tags=integration ./... -cover

# Real LXC tests (require LXC installation)
test-real-lxc: ## Run integration tests with real LXC (requires LXC)
	@echo "Running integration tests with real LXC..."
	LXC_REAL=1 go test -tags=integration ./internal/helpers -v

# Fast test (short tests only)
test-fast: ## Run fast tests only (mocks, no integration)
	@echo "Running fast tests..."
	go test -short ./...

# Helpers coverage (specific target for the helpers package)
coverage-helpers: ## Run coverage tests for helpers package specifically
	@echo "Running helpers coverage tests..."
	go test ./internal/helpers -cover -coverprofile=helpers_coverage.out
	go tool cover -html=helpers_coverage.out -o helpers_coverage.html
	@echo "Helpers coverage report generated: helpers_coverage.html"

# Build
build: ## Build the binary with version information
	@echo "Building lxc-go-cli version $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o lxc-go-cli .

# Show version info
version: ## Show version information that would be built
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# Clean
clean: ## Clean build artifacts and test files
	@echo "Cleaning..."
	rm -f lxc-go-cli
	rm -f coverage.out coverage.html

# Verify (for CI)
verify: test lint ## Run tests and linting for CI

lint: ## Run go vet and other linting
	@echo "Running linting..."
	go vet ./...
	go fmt ./...

# Development helpers
dev-setup: ## Install development dependencies
	@echo "Setting up development environment..."
	go mod tidy
	go mod download

.DEFAULT_GOAL := help
