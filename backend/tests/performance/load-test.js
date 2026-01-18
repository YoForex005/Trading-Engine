// k6 Load Test - 1000 concurrent users placing orders
// Run with: k6 run --vus 1000 --duration 5m load-test.js

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const orderPlacementRate = new Rate('order_placement_success');
const orderExecutionTime = new Trend('order_execution_time');
const wsMessageLatency = new Trend('websocket_message_latency');
const wsConnections = new Counter('websocket_connections');
const apiErrors = new Counter('api_errors');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

// Test thresholds - fail if not met
export const options = {
  scenarios: {
    // Ramping load test
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 200 },  // Ramp up to 200 users
        { duration: '5m', target: 500 },  // Ramp up to 500 users
        { duration: '5m', target: 1000 }, // Ramp up to 1000 users
        { duration: '10m', target: 1000 }, // Stay at 1000 users
        { duration: '3m', target: 0 },    // Ramp down
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<200', 'p(99)<500'], // 95% under 200ms, 99% under 500ms
    'http_req_failed': ['rate<0.01'], // Error rate under 1%
    'order_placement_success': ['rate>0.99'], // 99% success rate
    'order_execution_time': ['p(95)<50', 'p(99)<100'], // Target <50ms p95
    'websocket_message_latency': ['p(95)<10', 'p(99)<20'], // Target <10ms p95
  },
};

// Test data
const symbols = ['BTCUSD', 'ETHUSD', 'XRPUSD', 'BNBUSD', 'SOLUSD'];
const orderTypes = ['market', 'limit', 'stop_loss', 'take_profit'];
const sides = ['buy', 'sell'];

// Helper function to generate random order
function generateOrder() {
  return {
    symbol: symbols[Math.floor(Math.random() * symbols.length)],
    side: sides[Math.floor(Math.random() * sides.length)],
    type: orderTypes[Math.floor(Math.random() * orderTypes.length)],
    quantity: (Math.random() * 10 + 0.01).toFixed(2),
    price: orderTypes[Math.floor(Math.random() * orderTypes.length)] === 'market'
      ? undefined
      : (Math.random() * 50000 + 1000).toFixed(2),
  };
}

// Setup function - runs once per VU
export function setup() {
  // Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check passed': (r) => r.status === 200,
  });

  return { startTime: Date.now() };
}

// Main test function
export default function (data) {
  const userId = `user_${__VU}`;

  // 1. Login/Authentication
  const loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    username: userId,
    password: 'test_password',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const authSuccess = check(loginRes, {
    'login successful': (r) => r.status === 200,
    'received token': (r) => r.json('token') !== undefined,
  });

  if (!authSuccess) {
    apiErrors.add(1);
    sleep(1);
    return;
  }

  const token = loginRes.json('token');
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 2. Get Account Info
  const accountRes = http.get(`${BASE_URL}/api/account`, { headers });
  check(accountRes, {
    'account info retrieved': (r) => r.status === 200,
  });

  // 3. Get Market Data
  const marketRes = http.get(`${BASE_URL}/api/market/prices`, { headers });
  check(marketRes, {
    'market data retrieved': (r) => r.status === 200,
  });

  // Think time - simulate user reading data
  sleep(Math.random() * 2 + 1); // 1-3 seconds

  // 4. Place Order
  const order = generateOrder();
  const orderStartTime = Date.now();

  const orderRes = http.post(`${BASE_URL}/api/orders`, JSON.stringify(order), { headers });

  const orderSuccess = check(orderRes, {
    'order placed successfully': (r) => r.status === 200 || r.status === 201,
    'order has ID': (r) => r.json('order_id') !== undefined,
  });

  orderPlacementRate.add(orderSuccess);
  if (orderSuccess) {
    orderExecutionTime.add(Date.now() - orderStartTime);
  } else {
    apiErrors.add(1);
  }

  // 5. Get Order Status
  if (orderSuccess) {
    const orderId = orderRes.json('order_id');
    const orderStatusRes = http.get(`${BASE_URL}/api/orders/${orderId}`, { headers });
    check(orderStatusRes, {
      'order status retrieved': (r) => r.status === 200,
    });
  }

  // 6. Get Positions
  const positionsRes = http.get(`${BASE_URL}/api/positions`, { headers });
  check(positionsRes, {
    'positions retrieved': (r) => r.status === 200,
  });

  // Think time before next action
  sleep(Math.random() * 3 + 2); // 2-5 seconds

  // 7. WebSocket Connection (30% of users)
  if (Math.random() < 0.3) {
    testWebSocket(token);
  }

  // Realistic user think time
  sleep(Math.random() * 5 + 3); // 3-8 seconds
}

// WebSocket testing function
function testWebSocket(token) {
  const url = `${WS_URL}?token=${token}`;

  const res = ws.connect(url, {
    headers: { 'Authorization': `Bearer ${token}` },
  }, function (socket) {
    wsConnections.add(1);

    socket.on('open', () => {
      // Subscribe to market data
      const subscribeMsg = JSON.stringify({
        type: 'subscribe',
        channels: ['prices', 'orders', 'positions'],
      });
      socket.send(subscribeMsg);
    });

    socket.on('message', (data) => {
      const receiveTime = Date.now();
      const msg = JSON.parse(data);

      // Calculate latency if message has timestamp
      if (msg.timestamp) {
        const latency = receiveTime - msg.timestamp;
        wsMessageLatency.add(latency);
      }

      // Verify message structure
      check(msg, {
        'valid message type': (m) => m.type !== undefined,
        'has data': (m) => m.data !== undefined,
      });
    });

    socket.on('error', (e) => {
      console.error('WebSocket error:', e);
      apiErrors.add(1);
    });

    // Keep connection alive for 10-30 seconds
    socket.setTimeout(() => {
      socket.close();
    }, Math.random() * 20000 + 10000);
  });

  check(res, {
    'websocket connected': (r) => r && r.status === 101,
  });
}

// Teardown function
export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`Test completed in ${duration} seconds`);
}

// Generate summary report
export function handleSummary(data) {
  return {
    'summary.json': JSON.stringify(data),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, config) {
  const indent = config.indent || '';
  const colors = config.enableColors || false;

  let summary = '\n' + indent + '█ Load Test Summary\n';
  summary += indent + '─'.repeat(50) + '\n\n';

  // Add metrics
  const metrics = data.metrics;

  summary += indent + 'HTTP Performance:\n';
  summary += indent + `  • Requests: ${metrics.http_reqs.values.count}\n`;
  summary += indent + `  • Duration p95: ${metrics.http_req_duration.values['p(95)']}ms\n`;
  summary += indent + `  • Duration p99: ${metrics.http_req_duration.values['p(99)']}ms\n`;
  summary += indent + `  • Failed: ${(metrics.http_req_failed.values.rate * 100).toFixed(2)}%\n\n`;

  summary += indent + 'Order Performance:\n';
  summary += indent + `  • Success Rate: ${(metrics.order_placement_success.values.rate * 100).toFixed(2)}%\n`;
  summary += indent + `  • Execution p95: ${metrics.order_execution_time.values['p(95)']}ms\n`;
  summary += indent + `  • Execution p99: ${metrics.order_execution_time.values['p(99)']}ms\n\n`;

  summary += indent + 'WebSocket Performance:\n';
  summary += indent + `  • Connections: ${metrics.websocket_connections.values.count}\n`;
  summary += indent + `  • Latency p95: ${metrics.websocket_message_latency.values['p(95)']}ms\n`;
  summary += indent + `  • Latency p99: ${metrics.websocket_message_latency.values['p(99)']}ms\n\n`;

  summary += indent + `Total Errors: ${metrics.api_errors.values.count}\n`;

  return summary;
}
