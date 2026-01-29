# Trading Engine Documentation

Welcome to the RTX Trading Engine documentation. This comprehensive guide covers architecture, API usage, deployment, and trading concepts.

## Documentation Structure

### 1. Architecture Documentation
- [System Overview](./architecture/system-overview.md) - High-level architecture and component diagrams
- [Data Flow](./architecture/data-flow.md) - Data flow diagrams and message routing
- [Database Schema](./architecture/database-schema.md) - Ledger and data storage specifications
- [API Architecture](./architecture/api-architecture.md) - REST API design and patterns
- [WebSocket Protocol](./architecture/websocket-protocol.md) - Real-time communication specification

### 2. API Documentation
- [OpenAPI Specification](./api/openapi.yaml) - Complete OpenAPI 3.0 specification
- [Authentication Guide](./api/authentication.md) - JWT authentication and authorization
- [Endpoints Reference](./api/endpoints.md) - All REST endpoints with examples
- [Error Codes](./api/error-codes.md) - HTTP status codes and error handling
- [Rate Limiting](./api/rate-limiting.md) - API rate limits and quotas

### 3. Developer Guides
- [Setup Guide](./developer/setup.md) - Installation and development environment
- [Code Organization](./developer/code-organization.md) - Project structure and conventions
- [Testing Guide](./developer/testing.md) - Unit, integration, and E2E testing
- [Deployment Guide](./developer/deployment.md) - Production deployment procedures
- [Contributing](./developer/contributing.md) - Contribution guidelines

### 4. Admin User Guide
- [Feature Toggles](./admin/feature-toggles.md) - Managing execution modes and features
- [Client Management](./admin/client-management.md) - Account creation and management
- [Risk Parameters](./admin/risk-parameters.md) - Configuring leverage, margins, and limits
- [Reporting](./admin/reporting.md) - Analytics and transaction reporting
- [Troubleshooting](./admin/troubleshooting.md) - Common issues and solutions

### 5. Trading Concepts
- [Execution Models](./concepts/execution-models.md) - A-Book, B-Book, and C-Book explained
- [Order Types](./concepts/order-types.md) - Market, limit, stop, and stop-limit orders
- [Position Management](./concepts/position-management.md) - Opening, closing, and modifying positions
- [P&L Calculation](./concepts/pnl-calculation.md) - Profit and loss methodology
- [Risk Management](./concepts/risk-management.md) - Margin, leverage, and risk controls

## Quick Start

### For Developers
```bash
# Clone the repository
git clone <repository-url>
cd trading-engine

# Install dependencies (Go 1.19+)
cd backend
go mod download

# Run the server
go run cmd/server/main.go
```

### For Administrators
1. Access the admin panel at `http://localhost:7999/admin`
2. Configure execution mode (A-Book/B-Book)
3. Set up liquidity provider connections
4. Create client accounts and set risk parameters

### For API Users
```bash
# Login and get JWT token
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo-user","password":"password"}'

# Get account summary
curl http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## System Requirements

### Minimum Requirements
- Go 1.19 or higher
- 2 GB RAM
- 1 CPU core
- 10 GB disk space

### Recommended Requirements
- Go 1.20 or higher
- 4 GB RAM
- 2 CPU cores
- 50 GB disk space
- SSD storage for tick data

## Support

- **Issues**: Submit issues on GitHub
- **Email**: support@rtxtrading.com
- **Documentation**: https://docs.rtxtrading.com

## License

Copyright 2024 RTX Trading. All rights reserved.
