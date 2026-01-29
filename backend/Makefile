# RTX Trading Engine - Makefile
# Convenient shortcuts for common test commands

.PHONY: help test test-fast test-unit test-integration test-e2e test-load coverage report clean docker-up docker-down

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m

help: ## Show this help message
	@echo "$(BLUE)RTX Trading Engine - Test Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

test: ## Run all tests (default mode)
	@echo "$(GREEN)Running all tests...$(NC)"
	./scripts/test/run-all-tests.sh

test-fast: ## Run all tests in parallel mode
	@echo "$(GREEN)Running all tests (parallel mode)...$(NC)"
	./scripts/test/run-all-tests.sh --parallel

test-failfast: ## Run tests and stop on first failure
	@echo "$(YELLOW)Running tests (fail-fast mode)...$(NC)"
	./scripts/test/run-all-tests.sh --fail-fast

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	./scripts/test/run-backend-tests.sh

test-integration: ## Run integration tests only
	@echo "$(GREEN)Running integration tests...$(NC)"
	./scripts/test/run-integration-tests.sh

test-e2e: ## Run end-to-end tests only
	@echo "$(GREEN)Running E2E tests...$(NC)"
	./scripts/test/run-e2e-tests.sh

test-load: ## Run performance/load tests
	@echo "$(GREEN)Running load tests...$(NC)"
	./scripts/test/run-load-tests.sh

coverage: ## Verify coverage threshold
	@echo "$(GREEN)Verifying coverage...$(NC)"
	./scripts/test/verify-coverage.sh

report: ## Generate test report
	@echo "$(GREEN)Generating test report...$(NC)"
	./scripts/test/generate-test-report.sh
	@echo "$(GREEN)Opening report...$(NC)"
	@open test-reports/index.html 2>/dev/null || xdg-open test-reports/index.html 2>/dev/null || echo "Please open test-reports/index.html manually"

coverage-report: ## Open coverage report in browser
	@echo "$(GREEN)Opening coverage report...$(NC)"
	@open test-reports/coverage/coverage.html 2>/dev/null || xdg-open test-reports/coverage/coverage.html 2>/dev/null || echo "Please open test-reports/coverage/coverage.html manually"

benchmark-report: ## Open benchmark report in browser
	@echo "$(GREEN)Opening benchmark report...$(NC)"
	@open test-reports/benchmarks/report.html 2>/dev/null || xdg-open test-reports/benchmarks/report.html 2>/dev/null || echo "Please open test-reports/benchmarks/report.html manually"

clean: ## Clean test reports and cache
	@echo "$(YELLOW)Cleaning test artifacts...$(NC)"
	rm -rf test-reports/
	go clean -testcache
	@echo "$(GREEN)Clean complete!$(NC)"

docker-up: ## Start test services (Redis, etc.)
	@echo "$(GREEN)Starting test services...$(NC)"
	docker-compose -f docker-compose.test.yml up -d
	@echo "$(GREEN)Waiting for services to be ready...$(NC)"
	@sleep 5
	@echo "$(GREEN)Services ready!$(NC)"

docker-down: ## Stop test services
	@echo "$(YELLOW)Stopping test services...$(NC)"
	docker-compose -f docker-compose.test.yml down
	@echo "$(GREEN)Services stopped!$(NC)"

docker-logs: ## View test service logs
	docker-compose -f docker-compose.test.yml logs -f

verify: ## Quick verification that scripts work
	@echo "$(GREEN)Verifying test scripts...$(NC)"
	./scripts/test/test-example.sh

ci: ## Run tests in CI mode (no Docker services)
	@echo "$(GREEN)Running tests in CI mode...$(NC)"
	CI=true ./scripts/test/run-all-tests.sh --skip-services

build: ## Build the server
	@echo "$(GREEN)Building server...$(NC)"
	go build -o bin/server ./cmd/server

run: ## Run the server
	@echo "$(GREEN)Starting server...$(NC)"
	go run ./cmd/server

dev: ## Run server in development mode
	@echo "$(GREEN)Starting server in development mode...$(NC)"
	./dev_runner.sh

lint: ## Run Go linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run ./...

format: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	gofmt -s -w .

tidy: ## Tidy Go modules
	@echo "$(GREEN)Tidying modules...$(NC)"
	go mod tidy

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	go mod download

all: clean deps test report ## Clean, download deps, run tests, generate report
	@echo "$(GREEN)All tasks complete!$(NC)"

pre-commit: format lint test-unit ## Run before committing (format, lint, unit tests)
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

pre-push: clean test ## Run before pushing (clean, all tests)
	@echo "$(GREEN)Pre-push checks passed!$(NC)"

# Quality gates
quality: lint test coverage ## Run quality gates (lint, test, coverage)
	@echo "$(GREEN)Quality gates passed!$(NC)"

# Performance profiling
profile-cpu: ## Profile CPU usage
	@echo "$(GREEN)Profiling CPU...$(NC)"
	go tool pprof test-reports/benchmarks/cpu.prof

profile-mem: ## Profile memory usage
	@echo "$(GREEN)Profiling memory...$(NC)"
	go tool pprof test-reports/benchmarks/mem.prof

# Quick stats
stats: ## Show test statistics
	@echo "$(BLUE)Test Statistics:$(NC)"
	@echo ""
	@if [ -f test-reports/coverage/coverage-percent.txt ]; then \
		echo "  Coverage: $$(cat test-reports/coverage/coverage-percent.txt)%"; \
	else \
		echo "  Coverage: Not available (run 'make test' first)"; \
	fi
	@echo ""
	@if [ -f test-reports/test-run.log ]; then \
		echo "  Last test run: $$(stat -f %Sm test-reports/test-run.log)"; \
	else \
		echo "  Last test run: Never"; \
	fi
	@echo ""

# Watch mode (requires entr)
watch-test: ## Watch files and run tests on changes (requires entr)
	@echo "$(GREEN)Watching for changes...$(NC)"
	find . -name '*.go' | entr -c make test-unit

# Install development tools
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)Tools installed!$(NC)"
