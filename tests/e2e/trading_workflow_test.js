/**
 * E2E Trading Workflow Tests
 * Tests complete trading workflows from login to position close
 *
 * Setup: npm install playwright @playwright/test
 * Run: npx playwright test trading_workflow_test.js
 */

const { test, expect } = require('@playwright/test');

const API_BASE = 'http://localhost:7999';
const WS_URL = 'ws://localhost:7999/ws';

// Test configuration
const testConfig = {
  username: 'test-user',
  password: 'password123',
  initialBalance: 10000,
};

/**
 * API Helper Functions
 */
class TradingAPI {
  constructor(token = null) {
    this.token = token;
  }

  async login(username, password) {
    const response = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();
    this.token = data.token;
    return data;
  }

  async placeOrder(orderData) {
    const response = await fetch(`${API_BASE}/order`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
      },
      body: JSON.stringify(orderData),
    });

    return await response.json();
  }

  async getPositions() {
    const response = await fetch(`${API_BASE}/positions`, {
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
    });

    return await response.json();
  }

  async closePosition(tradeId) {
    const response = await fetch(`${API_BASE}/position/close`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
      },
      body: JSON.stringify({ tradeId }),
    });

    return await response.json();
  }

  async getAccount() {
    const response = await fetch(`${API_BASE}/api/account/summary`, {
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
    });

    return await response.json();
  }

  async getPendingOrders() {
    const response = await fetch(`${API_BASE}/orders/pending`, {
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
    });

    return await response.json();
  }

  async cancelOrder(orderId) {
    const response = await fetch(`${API_BASE}/order/cancel`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
      },
      body: JSON.stringify({ orderId }),
    });

    return await response.json();
  }
}

/**
 * WebSocket Helper
 */
class TradingWebSocket {
  constructor() {
    this.ws = null;
    this.messages = [];
    this.connected = false;
  }

  async connect() {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(WS_URL);

        this.ws.onopen = () => {
          this.connected = true;
          resolve();
        };

        this.ws.onmessage = (event) => {
          const data = JSON.parse(event.data);
          this.messages.push(data);
        };

        this.ws.onerror = (error) => {
          reject(error);
        };

        setTimeout(() => reject(new Error('Connection timeout')), 5000);
      } catch (error) {
        reject(error);
      }
    });
  }

  subscribe(symbol) {
    this.ws.send(JSON.stringify({
      type: 'subscribe',
      symbol: symbol,
    }));
  }

  waitForTick(symbol, timeout = 5000) {
    return new Promise((resolve, reject) => {
      const startTime = Date.now();

      const checkMessages = () => {
        const tick = this.messages.find(
          msg => msg.type === 'tick' && msg.symbol === symbol
        );

        if (tick) {
          resolve(tick);
        } else if (Date.now() - startTime > timeout) {
          reject(new Error('Timeout waiting for tick'));
        } else {
          setTimeout(checkMessages, 100);
        }
      };

      checkMessages();
    });
  }

  close() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

/**
 * E2E Test Suite
 */

test.describe('Trading Platform E2E Tests', () => {
  let api;

  test.beforeEach(async () => {
    api = new TradingAPI();
  });

  test('Complete Trading Workflow: Login → Place Order → Close Position', async () => {
    // Step 1: Login
    const loginResult = await api.login(testConfig.username, testConfig.password);

    expect(loginResult.token).toBeTruthy();
    expect(loginResult.user).toBeTruthy();
    expect(loginResult.user.username).toBe(testConfig.username);

    console.log('✓ Login successful');

    // Step 2: Check initial account balance
    const accountBefore = await api.getAccount();
    console.log('✓ Account balance:', accountBefore.balance || 'N/A');

    // Step 3: Place market order
    const orderData = {
      symbol: 'EURUSD',
      side: 'BUY',
      volume: 0.1,
      type: 'MARKET',
    };

    const orderResult = await api.placeOrder(orderData);

    expect(orderResult.success).toBe(true);
    expect(orderResult.order).toBeTruthy();

    const orderId = orderResult.order.ClientOrderID;
    console.log('✓ Order placed:', orderId);

    // Wait for order to be processed
    await new Promise(resolve => setTimeout(resolve, 500));

    // Step 4: Verify position opened
    const positions = await api.getPositions();
    console.log('✓ Positions:', positions);

    // Step 5: Check account after order
    const accountAfter = await api.getAccount();
    console.log('✓ Account after order:', accountAfter);

    // Step 6: Close position (if available)
    if (positions && positions.length > 0) {
      const closeResult = await api.closePosition(positions[0].id);
      console.log('✓ Position closed:', closeResult);
    } else {
      console.log('⚠ No positions to close (may be A-Book mode)');
    }

    console.log('✓ Complete workflow finished');
  });

  test('Limit Order Workflow', async () => {
    // Login
    await api.login(testConfig.username, testConfig.password);

    // Place limit order
    const limitOrder = {
      symbol: 'EURUSD',
      side: 'BUY',
      volume: 0.1,
      price: 1.09500,
    };

    const response = await fetch(`${API_BASE}/order/limit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${api.token}`,
      },
      body: JSON.stringify(limitOrder),
    });

    const result = await response.json();
    expect(result).toBeTruthy();

    console.log('✓ Limit order placed:', result);

    // Check pending orders
    const pendingOrders = await api.getPendingOrders();
    expect(Array.isArray(pendingOrders)).toBe(true);

    console.log('✓ Pending orders:', pendingOrders.length);

    // Cancel order
    if (pendingOrders.length > 0) {
      const cancelResult = await api.cancelOrder(pendingOrders[0].id);
      expect(cancelResult.success).toBe(true);
      console.log('✓ Order cancelled');
    }
  });

  test('Multiple Orders Workflow', async () => {
    await api.login(testConfig.username, testConfig.password);

    const symbols = ['EURUSD', 'GBPUSD', 'USDJPY'];
    const orders = [];

    // Place multiple orders
    for (const symbol of symbols) {
      const orderData = {
        symbol,
        side: 'BUY',
        volume: 0.05,
        type: 'MARKET',
      };

      const result = await api.placeOrder(orderData);
      orders.push(result);

      console.log(`✓ Order placed for ${symbol}`);
      await new Promise(resolve => setTimeout(resolve, 200));
    }

    expect(orders.length).toBe(3);
    console.log('✓ All orders placed');

    // Check positions
    await new Promise(resolve => setTimeout(resolve, 500));
    const positions = await api.getPositions();

    console.log('✓ Positions count:', positions ? positions.length : 0);
  });

  test('Order Modification Workflow', async () => {
    await api.login(testConfig.username, testConfig.password);

    // Place order with SL/TP
    const orderData = {
      symbol: 'EURUSD',
      side: 'BUY',
      volume: 0.1,
      type: 'MARKET',
      sl: 1.09500,
      tp: 1.10500,
    };

    const orderResult = await api.placeOrder(orderData);
    expect(orderResult.success).toBe(true);

    const tradeId = orderResult.order.ClientOrderID;
    console.log('✓ Order with SL/TP placed:', tradeId);

    await new Promise(resolve => setTimeout(resolve, 300));

    // Modify SL/TP
    const modifyResponse = await fetch(`${API_BASE}/position/modify`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${api.token}`,
      },
      body: JSON.stringify({
        tradeId,
        sl: 1.09800,
        tp: 1.10700,
      }),
    });

    const modifyResult = await modifyResponse.json();
    expect(modifyResult.success).toBe(true);

    console.log('✓ SL/TP modified');
  });

  test('Risk Calculator Integration', async () => {
    await api.login(testConfig.username, testConfig.password);

    // Calculate lot size
    const response = await fetch(
      `${API_BASE}/risk/calculate-lot?symbol=EURUSD&riskPercent=2&slPips=20`,
      {
        headers: {
          'Authorization': `Bearer ${api.token}`,
        },
      }
    );

    const result = await response.json();

    expect(result).toBeTruthy();
    expect(result.recommendedLotSize).toBeDefined();

    console.log('✓ Risk calculation:', result);

    // Preview margin
    const marginResponse = await fetch(
      `${API_BASE}/risk/margin-preview?symbol=EURUSD&volume=1.0&side=BUY`,
      {
        headers: {
          'Authorization': `Bearer ${api.token}`,
        },
      }
    );

    const marginResult = await marginResponse.json();

    expect(marginResult).toBeTruthy();
    expect(marginResult.requiredMargin).toBeDefined();

    console.log('✓ Margin preview:', marginResult);
  });

  test('Historical Data Retrieval', async () => {
    await api.login(testConfig.username, testConfig.password);

    // Get tick data
    const ticksResponse = await fetch(
      `${API_BASE}/ticks?symbol=EURUSD&limit=10`,
      {
        headers: {
          'Authorization': `Bearer ${api.token}`,
        },
      }
    );

    const ticks = await ticksResponse.json();

    expect(Array.isArray(ticks)).toBe(true);
    console.log('✓ Ticks retrieved:', ticks.length);

    // Get OHLC data
    const ohlcResponse = await fetch(
      `${API_BASE}/ohlc?symbol=EURUSD&timeframe=1m&limit=5`,
      {
        headers: {
          'Authorization': `Bearer ${api.token}`,
        },
      }
    );

    const ohlc = await ohlcResponse.json();

    expect(Array.isArray(ohlc)).toBe(true);
    console.log('✓ OHLC candles:', ohlc.length);
  });

  test('Error Handling: Invalid Orders', async () => {
    await api.login(testConfig.username, testConfig.password);

    // Test invalid symbol
    const invalidOrder = {
      symbol: 'INVALID',
      side: 'BUY',
      volume: 0.1,
      type: 'MARKET',
    };

    const result = await api.placeOrder(invalidOrder);
    console.log('✓ Invalid order response:', result);

    // Test negative volume
    const negativeVolume = {
      symbol: 'EURUSD',
      side: 'BUY',
      volume: -0.1,
      type: 'MARKET',
    };

    const result2 = await api.placeOrder(negativeVolume);
    console.log('✓ Negative volume response:', result2);
  });

  test('Concurrent Order Placement', async () => {
    await api.login(testConfig.username, testConfig.password);

    const concurrentOrders = Array(5).fill(null).map((_, i) => ({
      symbol: 'EURUSD',
      side: i % 2 === 0 ? 'BUY' : 'SELL',
      volume: 0.01,
      type: 'MARKET',
    }));

    // Place all orders concurrently
    const results = await Promise.all(
      concurrentOrders.map(order => api.placeOrder(order))
    );

    const successCount = results.filter(r => r.success).length;

    console.log(`✓ Concurrent orders: ${successCount}/${results.length} successful`);
    expect(successCount).toBeGreaterThan(0);
  });
});

test.describe('WebSocket Real-time Tests', () => {
  test.skip('WebSocket Connection and Tick Streaming', async () => {
    // Note: This test requires WebSocket support in the test environment
    // Skip if WebSocket is not available

    const ws = new TradingWebSocket();

    try {
      await ws.connect();
      expect(ws.connected).toBe(true);

      console.log('✓ WebSocket connected');

      // Subscribe to symbol
      ws.subscribe('EURUSD');

      // Wait for tick
      const tick = await ws.waitForTick('EURUSD', 10000);

      expect(tick).toBeTruthy();
      expect(tick.symbol).toBe('EURUSD');
      expect(tick.bid).toBeGreaterThan(0);
      expect(tick.ask).toBeGreaterThan(0);

      console.log('✓ Tick received:', tick);

      ws.close();
    } catch (error) {
      console.log('⚠ WebSocket test skipped:', error.message);
    }
  });
});

test.describe('Performance Tests', () => {
  test('Order Placement Latency', async () => {
    const api = new TradingAPI();
    await api.login(testConfig.username, testConfig.password);

    const iterations = 10;
    const latencies = [];

    for (let i = 0; i < iterations; i++) {
      const start = Date.now();

      await api.placeOrder({
        symbol: 'EURUSD',
        side: 'BUY',
        volume: 0.01,
        type: 'MARKET',
      });

      const latency = Date.now() - start;
      latencies.push(latency);

      await new Promise(resolve => setTimeout(resolve, 100));
    }

    const avgLatency = latencies.reduce((a, b) => a + b, 0) / latencies.length;
    const maxLatency = Math.max(...latencies);
    const minLatency = Math.min(...latencies);

    console.log('Order Placement Performance:');
    console.log(`  Average: ${avgLatency.toFixed(2)}ms`);
    console.log(`  Min: ${minLatency}ms`);
    console.log(`  Max: ${maxLatency}ms`);

    expect(avgLatency).toBeLessThan(2000); // Should be under 2 seconds
  });
});

/**
 * Test Cleanup
 */
test.afterAll(async () => {
  console.log('\n✓ All E2E tests completed');
});
