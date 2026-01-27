.PHONY: help test test-unit test-integration test-coverage test-race test-bench clean-test build run migrate lint fmt vet

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Display this help message
	@echo "PLI Agent Management System - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Test commands
test: ## Run all tests
	@./scripts/run-tests.sh all

test-unit: ## Run unit tests only
	@./scripts/run-tests.sh unit

test-integration: ## Run integration tests
	@./scripts/run-tests.sh integration

test-coverage: ## Generate test coverage report
	@./scripts/run-tests.sh coverage

test-race: ## Run tests with race detection
	@./scripts/run-tests.sh race

test-bench: ## Run benchmark tests
	@./scripts/run-tests.sh bench

clean-test: ## Clean test artifacts
	@./scripts/run-tests.sh clean

# Build commands
build: ## Build the application
	@echo "Building PLI Agent API..."
	@go build -o bin/pli-agent-api ./main.go
	@echo "Build complete: bin/pli-agent-api"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/pli-agent-api-linux ./main.go
	@echo "Build complete: bin/pli-agent-api-linux"

# Run commands
run: ## Run the application
	@echo "Running PLI Agent API..."
	@go run main.go

run-dev: ## Run with hot reload (requires air)
	@echo "Running with hot reload..."
	@air

# Database commands
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@go run main.go migrate up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@go run main.go migrate down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@go run main.go migrate status

# Code quality commands
lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Dependency commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

# Clean commands
clean: clean-test ## Clean build artifacts and test files
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t pli-agent-api:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 pli-agent-api:latest

docker-compose-up: ## Start docker-compose services
	@echo "Starting docker-compose services..."
	@docker-compose up -d

docker-compose-down: ## Stop docker-compose services
	@echo "Stopping docker-compose services..."
	@docker-compose down

# Documentation commands
docs: ## Generate documentation
	@echo "Generating documentation..."
	@godoc -http=:6060
	@echo "Documentation server running at http://localhost:6060"

# Git commands
git-status: ## Show git status
	@git status

git-log: ## Show recent git log
	@git log --oneline -10

# Quick commands
quick-check: fmt vet test-unit ## Run quick checks (format, vet, unit tests)
	@echo "Quick check complete!"

full-check: fmt vet lint test test-coverage ## Run full checks (all quality and tests)
	@echo "Full check complete!"

# Installation commands
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@echo "Tools installed!"
