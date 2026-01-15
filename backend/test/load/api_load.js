import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const ordersPlaced = new Counter('orders_placed');
const orderSuccessRate = new Rate('order_success_rate');
const orderDuration = new Trend('order_duration');

// Load test configuration
export const options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp to 20 users
    { duration: '2m', target: 50 },    // Ramp to 50 users
    { duration: '3m', target: 50 },    // Sustain 50 users
    { duration: '30s', target: 100 },  // Spike to 100
    { duration: '1m', target: 100 },   // Sustain spike
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<200'],      // 95% of requests under 200ms
    'http_req_duration': ['p(99)<500'],      // 99% under 500ms
    'http_req_failed': ['rate<0.05'],        // Error rate below 5%
    'order_success_rate': ['rate>0.95'],     // 95%+ orders succeed
    'orders_placed': ['count>1000'],         // Place >1000 orders total
  },
};

// Test data
const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'BTCUSD', 'ETHUSD'];
const sides = ['buy', 'sell'];

// Setup function (runs once at start)
export function setup() {
  // Login to get auth token
  const baseURL = __ENV.API_URL || 'http://localhost:8080';
  const loginRes = http.post(`${baseURL}/api/auth/login`, JSON.stringify({
    username: 'loadtest',
    password: 'loadtest123'
  }), {
    headers: { 'Content-Type': 'application/json' }
  });

  const token = loginRes.json('token');
  return { baseURL, token };
}

// Main test function
export default function (data) {
  const { baseURL, token } = data;

  // Random order parameters
  const symbol = symbols[Math.floor(Math.random() * symbols.length)];
  const side = sides[Math.floor(Math.random() * sides.length)];
  const lots = (Math.random() * 0.9 + 0.1).toFixed(2); // 0.1 to 1.0 lots

  const payload = JSON.stringify({
    symbol,
    type: 'market',
    side,
    lots: parseFloat(lots)
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    tags: { name: 'PlaceOrder' }
  };

  // Place order
  const startTime = Date.now();
  const res = http.post(`${baseURL}/api/orders`, payload, params);
  const duration = Date.now() - startTime;

  orderDuration.add(duration);

  // Verify response
  const success = check(res, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'response has order ID': (r) => r.json('id') !== undefined,
    'order executed': (r) => r.json('status') === 'filled',
  });

  if (success) {
    ordersPlaced.add(1);
    orderSuccessRate.add(1);
  } else {
    orderSuccessRate.add(0);
    console.error(`VU ${__VU}: Order failed - ${res.status} - ${res.body}`);
  }

  // Get account balance (read operation)
  http.get(`${baseURL}/api/account`, {
    headers: { 'Authorization': `Bearer ${token}` },
    tags: { name: 'GetAccount' }
  });

  // Get open positions
  http.get(`${baseURL}/api/positions`, {
    headers: { 'Authorization': `Bearer ${token}` },
    tags: { name: 'GetPositions' }
  });

  // Think time (simulate user reading/deciding)
  sleep(Math.random() * 2 + 1); // 1-3 seconds
}

// Teardown function
export function teardown(data) {
  console.log('API load test complete');
}
