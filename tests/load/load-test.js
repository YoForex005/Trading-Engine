// ============================================
// k6 Load Testing Script for Trading Engine
// ============================================

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const orderExecutionTime = new Trend('order_execution_time');
const wsConnectionTime = new Trend('websocket_connection_time');
const successfulOrders = new Counter('successful_orders');
const failedOrders = new Counter('failed_orders');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 500 },   // Ramp up to 500 users
    { duration: '5m', target: 500 },   // Stay at 500 users
    { duration: '2m', target: 1000 },  // Ramp up to 1000 users
    { duration: '10m', target: 1000 }, // Stay at 1000 users
    { duration: '2m', target: 5000 },  // Spike test to 5000 users
    { duration: '5m', target: 5000 },  // Maintain spike
    { duration: '5m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'], // Error rate < 1%
    errors: ['rate<0.05'], // Custom error rate < 5%
    order_execution_time: ['p(95)<300', 'p(99)<500'],
    websocket_connection_time: ['p(95)<1000'],
  },
  ext: {
    loadimpact: {
      projectID: 3579709,
      name: 'Trading Engine Load Test'
    }
  }
};

// Environment configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:7999';
const WS_URL = __ENV.WS_URL || 'ws://localhost:7999/ws';

// Test data
const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'BTCUSD', 'ETHUSD'];
const sides = ['BUY', 'SELL'];

// ============================================
// Authentication
// ============================================
function login() {
  const loginPayload = JSON.stringify({
    username: 'demo-user',
    password: 'password'
  });

  const loginRes = http.post(`${BASE_URL}/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login successful': (r) => r.status === 200,
    'received token': (r) => r.json('token') !== '',
  }) || errorRate.add(1);

  return loginRes.json('token');
}

// ============================================
// Main Test Scenario
// ============================================
export default function() {
  const token = login();

  if (!token) {
    errorRate.add(1);
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };

  // Test 1: Get Account Summary
  const accountRes = http.get(`${BASE_URL}/api/account/summary`, { headers });
  check(accountRes, {
    'account summary status 200': (r) => r.status === 200,
    'has balance': (r) => r.json('balance') !== undefined,
  }) || errorRate.add(1);

  sleep(1);

  // Test 2: Get Market Data
  const symbol = symbols[Math.floor(Math.random() * symbols.length)];
  const ticksRes = http.get(`${BASE_URL}/ticks?symbol=${symbol}&limit=100`, { headers });
  check(ticksRes, {
    'ticks status 200': (r) => r.status === 200,
    'received ticks': (r) => Array.isArray(r.json()),
  }) || errorRate.add(1);

  sleep(1);

  // Test 3: Place Market Order
  const orderPayload = JSON.stringify({
    symbol: symbol,
    side: sides[Math.floor(Math.random() * sides.length)],
    volume: 0.01 + Math.random() * 0.09, // 0.01 to 0.1 lots
  });

  const orderStart = Date.now();
  const orderRes = http.post(`${BASE_URL}/api/orders/market`, orderPayload, { headers });
  const orderDuration = Date.now() - orderStart;

  orderExecutionTime.add(orderDuration);

  const orderSuccess = check(orderRes, {
    'order status 200': (r) => r.status === 200,
    'order has position': (r) => r.json('position') !== undefined,
    'order executed quickly': (r) => orderDuration < 500,
  });

  if (orderSuccess) {
    successfulOrders.add(1);
  } else {
    failedOrders.add(1);
    errorRate.add(1);
  }

  sleep(2);

  // Test 4: Get Open Positions
  const positionsRes = http.get(`${BASE_URL}/api/positions`, { headers });
  check(positionsRes, {
    'positions status 200': (r) => r.status === 200,
    'positions is array': (r) => Array.isArray(r.json()),
  }) || errorRate.add(1);

  sleep(1);

  // Test 5: Get OHLC Data
  const ohlcRes = http.get(`${BASE_URL}/ohlc?symbol=${symbol}&timeframe=1m&limit=100`, { headers });
  check(ohlcRes, {
    'ohlc status 200': (r) => r.status === 200,
    'received ohlc data': (r) => Array.isArray(r.json()),
  }) || errorRate.add(1);

  sleep(2);

  // Test 6: WebSocket Connection (for 20% of users)
  if (Math.random() < 0.2) {
    testWebSocket();
  }
}

// ============================================
// WebSocket Test
// ============================================
function testWebSocket() {
  const wsStart = Date.now();

  const res = ws.connect(WS_URL, function(socket) {
    socket.on('open', () => {
      const wsDuration = Date.now() - wsStart;
      wsConnectionTime.add(wsDuration);

      // Subscribe to market data
      socket.send(JSON.stringify({
        type: 'subscribe',
        symbols: symbols.slice(0, 3)
      }));
    });

    socket.on('message', (data) => {
      try {
        const msg = JSON.parse(data);
        check(msg, {
          'ws message has type': (m) => m.type !== undefined,
          'ws message has symbol': (m) => m.symbol !== undefined || m.type === 'subscribed',
        }) || errorRate.add(1);
      } catch (e) {
        errorRate.add(1);
      }
    });

    socket.on('error', (e) => {
      console.error('WebSocket error:', e);
      errorRate.add(1);
    });

    // Keep connection open for 30 seconds
    socket.setTimeout(() => {
      socket.close();
    }, 30000);
  });

  check(res, {
    'ws connection successful': (r) => r && r.status === 101,
  }) || errorRate.add(1);
}

// ============================================
// Stress Test Scenario
// ============================================
export function stressTest() {
  const token = login();

  if (!token) {
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };

  // Rapid fire orders
  for (let i = 0; i < 10; i++) {
    const symbol = symbols[Math.floor(Math.random() * symbols.length)];
    const orderPayload = JSON.stringify({
      symbol: symbol,
      side: sides[Math.floor(Math.random() * sides.length)],
      volume: 0.01,
    });

    const orderStart = Date.now();
    const orderRes = http.post(`${BASE_URL}/api/orders/market`, orderPayload, { headers });
    const orderDuration = Date.now() - orderStart;

    orderExecutionTime.add(orderDuration);

    if (orderRes.status === 200) {
      successfulOrders.add(1);
    } else {
      failedOrders.add(1);
      errorRate.add(1);
    }

    sleep(0.1); // Very short sleep for stress
  }
}

// ============================================
// Smoke Test Scenario
// ============================================
export function smokeTest() {
  const endpoints = [
    '/health',
    '/api/config',
  ];

  endpoints.forEach((endpoint) => {
    const res = http.get(`${BASE_URL}${endpoint}`);
    check(res, {
      [`${endpoint} status 200`]: (r) => r.status === 200,
    }) || errorRate.add(1);
  });
}

// ============================================
// Setup and Teardown
// ============================================
export function setup() {
  // Check if service is available
  const healthCheck = http.get(`${BASE_URL}/health`);

  if (healthCheck.status !== 200) {
    throw new Error('Service is not healthy');
  }

  console.log('Service is healthy, starting load test...');

  return {
    startTime: new Date().toISOString()
  };
}

export function teardown(data) {
  console.log(`Test completed. Started at: ${data.startTime}`);
  console.log(`Successful orders: ${successfulOrders.value}`);
  console.log(`Failed orders: ${failedOrders.value}`);
}

// ============================================
// Custom Scenarios
// ============================================
export const scenarios = {
  default: {
    executor: 'ramping-vus',
    exec: 'default',
    stages: options.stages,
  },
  stress: {
    executor: 'ramping-vus',
    exec: 'stressTest',
    startTime: '30m',
    stages: [
      { duration: '2m', target: 100 },
      { duration: '5m', target: 100 },
      { duration: '2m', target: 0 },
    ],
  },
  smoke: {
    executor: 'constant-vus',
    exec: 'smokeTest',
    vus: 1,
    duration: '1m',
  },
};
