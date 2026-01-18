# E2E Tests for RTX Trading Engine

End-to-end tests using Playwright to test complete workflows from a user perspective.

## Quick Start

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install

# Make sure backend is running
# cd ../../backend && go run cmd/server/main.go

# Run all tests
npm test

# Or use the helper script
./run_tests.sh all
```

## Test Files

### 1. trading_workflow_test.js
Tests complete trading workflows:
- User authentication
- Market order placement
- Limit order workflow
- Stop order workflow
- Position management
- Order modification (SL/TP, breakeven, trailing stop)
- Multiple concurrent orders
- Risk calculator integration
- Historical data retrieval
- Error handling
- Performance benchmarks

### 2. admin_workflow_test.js
Tests administrative operations:
- Broker configuration
- Execution mode switching (A-Book/B-Book)
- LP management (enable/disable, status)
- FIX session control
- Symbol management
- Account operations (deposit, withdraw, bonus, adjustment)
- Transaction ledger viewing
- Password reset
- Complete admin scenarios

## Test Structure

Each test file follows this pattern:

```javascript
test.describe('Test Category', () => {
  let api;

  test.beforeEach(async () => {
    // Setup
    api = new TradingAPI();
  });

  test('Test Name', async () => {
    // Test implementation
    expect(result).toBe(expected);
  });
});
```

## Running Tests

### All Tests
```bash
npm test
# or
./run_tests.sh all
```

### Specific Test Category
```bash
npm run test:trading      # Trading workflow tests
npm run test:admin        # Admin workflow tests
```

### Individual Tests
```bash
# Run specific test by name
npx playwright test -g "Complete Trading Workflow"

# Run specific file
npx playwright test trading_workflow_test.js
```

### Debug Mode
```bash
# Open debug UI
npm run test:debug

# Run with visible browser
npm run test:headed

# Open Playwright UI
./run_tests.sh ui
```

### Test Reports
```bash
# View HTML report
npm run test:report

# Generate custom report
npx playwright test --reporter=html
```

## Test Configuration

Configuration is in `playwright.config.js`:

- **Timeout**: 30 seconds per test
- **Retries**: 0 (2 in CI)
- **Workers**: 1 (sequential execution)
- **Base URL**: http://localhost:7999
- **Reporters**: HTML + List
- **Screenshots**: On failure
- **Videos**: On failure
- **Traces**: On first retry

## Test Helpers

### TradingAPI Class
Helper class for API interactions:

```javascript
const api = new TradingAPI();

// Login
const data = await api.login('username', 'password');

// Place order
const result = await api.placeOrder({
  symbol: 'EURUSD',
  side: 'BUY',
  volume: 0.1,
  type: 'MARKET'
});

// Get positions
const positions = await api.getPositions();

// Close position
await api.closePosition(tradeId);
```

### AdminAPI Class
Helper class for admin operations:

```javascript
const admin = new AdminAPI();

// Get configuration
const config = await admin.getConfig();

// Toggle execution mode
await admin.setExecutionMode('ABOOK');

// Manage LPs
const lps = await admin.listLPs();
await admin.toggleLP('oanda', true);

// Account operations
await admin.deposit('test-user', 1000, 'BANK_TRANSFER', 'Test');
```

## Test Data

### Test Account
- **Username**: test-user
- **Password**: password123
- **Initial Balance**: $10,000

### Test Symbols
- EURUSD
- GBPUSD
- USDJPY
- XAUUSD

### API Base URL
- **Local**: http://localhost:7999
- **WebSocket**: ws://localhost:7999/ws

## Expected Test Results

All tests should pass with the backend running:

```
  Trading Platform E2E Tests
    ✓ Complete Trading Workflow: Login → Place Order → Close Position
    ✓ Limit Order Workflow
    ✓ Multiple Orders Workflow
    ✓ Order Modification Workflow
    ✓ Risk Calculator Integration
    ✓ Historical Data Retrieval
    ✓ Error Handling: Invalid Orders
    ✓ Concurrent Order Placement

  Admin Configuration Tests
    ✓ Complete Admin Workflow: Configure → Enable LPs → Test Execution
    ✓ Broker Configuration Management
    ✓ Execution Mode Toggle
    ✓ LP Management
    ✓ Account Management
    ✓ Symbol Management

  All tests passed!
```

## Troubleshooting

### Backend Not Running
```bash
Error: Backend server is not running

Solution:
cd ../../backend
go run cmd/server/main.go
```

### Port Already in Use
```bash
Error: Port 7999 already in use

Solution:
# Find and kill the process
lsof -ti:7999 | xargs kill -9
```

### Dependencies Not Installed
```bash
Error: Cannot find module '@playwright/test'

Solution:
npm install
npx playwright install
```

### Tests Timing Out
```bash
# Increase timeout in playwright.config.js
timeout: 60 * 1000  // 60 seconds
```

### Browser Not Found
```bash
Error: Executable doesn't exist

Solution:
npx playwright install
```

## Best Practices

1. **Keep tests independent**: Each test should be self-contained
2. **Use descriptive names**: Test names should explain what is being tested
3. **Clean up resources**: Always close connections and reset state
4. **Handle async properly**: Use await for all async operations
5. **Add meaningful assertions**: Verify expected behavior explicitly
6. **Log progress**: Use console.log for debugging
7. **Handle errors gracefully**: Catch and log errors for debugging

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run E2E Tests
  run: |
    cd tests/e2e
    npm install
    npx playwright install --with-deps
    npm test
```

### Docker
```dockerfile
FROM mcr.microsoft.com/playwright:v1.40.0
COPY tests/e2e /tests
WORKDIR /tests
RUN npm install
CMD ["npm", "test"]
```

## Performance Expectations

- **Order Placement**: <2000ms
- **API Response**: <500ms
- **WebSocket Latency**: <100ms
- **Test Execution**: <30s per test

## Contributing

When adding new E2E tests:
1. Follow existing patterns
2. Add to appropriate test file
3. Update this README
4. Ensure tests pass locally
5. Add meaningful assertions
6. Include error handling

## Support

For issues or questions:
1. Check test logs
2. Verify backend is running
3. Check network connectivity
4. Review Playwright documentation
5. Check test reports for details
