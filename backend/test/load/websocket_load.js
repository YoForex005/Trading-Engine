import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// Custom metrics
const ticksReceived = new Counter('ticks_received');
const tickLatency = new Trend('tick_latency');
const connectionFailures = new Counter('connection_failures');

// Load test configuration
export const options = {
  stages: [
    { duration: '1m', target: 50 },    // Ramp up to 50 concurrent users
    { duration: '3m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 for 5 minutes (sustained load)
    { duration: '1m', target: 200 },   // Spike to 200 users
    { duration: '2m', target: 200 },   // Sustain spike
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    // Performance thresholds - test fails if not met
    'ticks_received': ['count>50000'],       // Should receive >50k ticks total
    'tick_latency': ['p(95)<500'],           // 95% of ticks arrive within 500ms
    'tick_latency': ['p(99)<1000'],          // 99% within 1 second
    'connection_failures': ['count<10'],     // Fewer than 10 connection failures
    'ws_connecting': ['p(95)<2000'],         // 95% of connections establish within 2s
  },
};

// Main test function - runs for each virtual user
export default function () {
  const url = __ENV.WS_URL || 'ws://localhost:8080/ws';
  const params = { tags: { name: 'WebSocketLoadTest' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      console.log(`VU ${__VU}: Connected`);

      // Subscribe to multiple symbols
      socket.send(JSON.stringify({
        type: 'subscribe',
        symbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'BTCUSD', 'ETHUSD']
      }));

      // Send heartbeat every 10 seconds
      socket.setInterval(() => {
        socket.send(JSON.stringify({ type: 'heartbeat' }));
      }, 10000);
    });

    socket.on('message', (data) => {
      try {
        const message = JSON.parse(data);

        if (message.type === 'tick') {
          ticksReceived.add(1);

          // Calculate latency if timestamp present
          if (message.timestamp) {
            const now = Date.now();
            const messageTime = new Date(message.timestamp).getTime();
            const latency = now - messageTime;
            tickLatency.add(latency);
          }

          // Validate tick data
          check(message, {
            'tick has symbol': (m) => m.symbol !== undefined,
            'tick has bid': (m) => m.bid !== undefined && m.bid > 0,
            'tick has ask': (m) => m.ask !== undefined && m.ask > 0,
            'ask > bid': (m) => m.ask > m.bid,
          });
        }
      } catch (e) {
        console.error(`VU ${__VU}: Parse error - ${e}`);
      }
    });

    socket.on('error', (e) => {
      console.error(`VU ${__VU}: WebSocket error - ${e.error()}`);
      connectionFailures.add(1);
    });

    socket.on('close', () => {
      console.log(`VU ${__VU}: Disconnected`);
    });

    // Stay connected for 60 seconds
    socket.setTimeout(() => {
      console.log(`VU ${__VU}: Closing connection after 60s`);
      socket.close();
    }, 60000);
  });

  check(res, {
    'WebSocket connection successful': (r) => r && r.status === 101,
  });

  if (!res || res.status !== 101) {
    connectionFailures.add(1);
    console.error(`VU ${__VU}: Connection failed with status ${res ? res.status : 'unknown'}`);
  }
}

// Teardown function (runs once at end)
export function teardown(data) {
  console.log('Load test complete - check metrics for results');
}
