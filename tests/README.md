# Integration Tests

Comprehensive integration test suite for the Trading Engine platform.

## Overview

This test suite provides end-to-end integration tests for both backend (Go) and frontend (TypeScript/React) components.

## Backend Tests (Go)

### Location
```
backend/internal/api/handlers/analytics_test.go
```

### Running Backend Tests

```bash
# Run all integration tests
cd backend
go test ./internal/api/handlers -v

# Run specific test
go test ./internal/api/handlers -v -run TestIntegrationRoutingMetrics

# Run with coverage
go test ./internal/api/handlers -v -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

## Frontend Tests (TypeScript/React)

### Location
```
clients/desktop/src/components/__tests__/
```

### Running Frontend Tests

```bash
# Navigate to desktop client
cd clients/desktop

# Run all tests in watch mode
bun test

# Run tests once (CI mode)
bun run test:run

# Run tests with UI
bun run test:ui

# Run with coverage
bun run test:coverage
```

## Test Coverage Goals

- Backend: > 80% coverage
- Frontend: > 75% coverage
- Critical paths: 100% coverage
