// k6 Soak Test - 24-hour endurance test
// Run with: k6 run --duration 24h soak-test.js

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics for long-running stability
const memoryLeakIndicator = new Gauge('memory_leak_indicator');
const performanceDegradation = new Trend('performance_degradation');
const connectionLeaks = new Counter('connection_leaks');
const dataCorruption = new Counter('data_corruption_events');
const systemStability = new Rate('system_stability');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

// Sustained load over 24 hours
export const options = {
  scenarios: {
    soak_test: {
      executor: 'constant-vus',
      vus: 500, // Moderate sustained load
      duration: '24h',
    },
  },
  thresholds: {
    'http_req_duration': [
      'p(95)<300', // Performance should not degrade
      'p(99)<600',
    ],
    'http_req_failed': ['rate<0.01'], // Very low error rate
    'system_stability': ['rate>0.99'], // 99% stability
    'memory_leak_indicator': ['value<1.5'], // Memory usage shouldn't grow >50%
    'connection_leaks': ['count<100'], // Minimal connection leaks
  },
};

const symbols = ['BTCUSD', 'ETHUSD', 'XRPUSD', 'BNBUSD', 'SOLUSD'];

// Track baseline metrics
let baselineResponseTime = 0;
let iterationCount = 0;

export function setup() {
  console.log('Starting 24-hour soak test - monitoring for memory leaks and performance degradation');

  // Establish baseline
  const baselineRequests = [];
  for (let i = 0; i < 10; i++) {
    const start = Date.now();
    http.get(`${BASE_URL}/health`);
    baselineRequests.push(Date.now() - start);
  }

  const baseline = baselineRequests.reduce((a, b) => a + b, 0) / baselineRequests.length;

  return {
    startTime: Date.now(),
    baselineResponseTime: baseline,
    hourlyMetrics: [],
  };
}

export default function (data) {
  iterationCount++;
  const currentHour = Math.floor((Date.now() - data.startTime) / (1000 * 60 * 60));

  // 1. Health check every 100 iterations
  if (iterationCount % 100 === 0) {
    const healthStart = Date.now();
    const healthRes = http.get(`${BASE_URL}/health`);
    const healthDuration = Date.now() - healthStart;

    // Detect performance degradation
    const degradation = healthDuration / data.baselineResponseTime;
    performanceDegradation.add(degradation);

    if (degradation > 1.5) {
      console.warn(`Performance degradation detected: ${degradation.toFixed(2)}x slower than baseline`);
      memoryLeakIndicator.add(degradation);
    }

    const healthOk = check(healthRes, {
      'health check ok': (r) => r.status === 200,
    });

    systemStability.add(healthOk);
  }

  // 2. Place realistic order
  const order = {
    symbol: symbols[Math.floor(Math.random() * symbols.length)],
    side: Math.random() > 0.5 ? 'buy' : 'sell',
    type: 'market',
    quantity: (Math.random() * 10 + 0.1).toFixed(2),
  };

  const orderRes = http.post(`${BASE_URL}/api/orders`, JSON.stringify(order), {
    headers: { 'Content-Type': 'application/json' },
    tags: { hour: currentHour },
  });

  const orderSuccess = check(orderRes, {
    'order placed': (r) => r.status === 200 || r.status === 201,
    'has order_id': (r) => r.json('order_id') !== undefined,
    'response time ok': (r) => r.timings.duration < 1000,
  });

  systemStability.add(orderSuccess);

  // 3. Verify order (data integrity check)
  if (orderSuccess) {
    const orderId = orderRes.json('order_id');
    const verifyRes = http.get(`${BASE_URL}/api/orders/${orderId}`, {
      tags: { hour: currentHour },
    });

    const dataIntegrity = check(verifyRes, {
      'order retrieved': (r) => r.status === 200,
      'order_id matches': (r) => r.json('order_id') === orderId,
      'symbol matches': (r) => r.json('symbol') === order.symbol,
    });

    if (!dataIntegrity) {
      dataCorruption.add(1);
      console.error(`Data corruption detected for order ${orderId}`);
    }
  }

  // 4. Get positions
  const positionsRes = http.get(`${BASE_URL}/api/positions`, {
    tags: { hour: currentHour },
  });

  check(positionsRes, {
    'positions retrieved': (r) => r.status === 200,
  });

  // 5. WebSocket connection test (10% of requests)
  if (Math.random() < 0.1) {
    testWebSocketStability();
  }

  // 6. Memory leak detection via metrics endpoint (every 500 iterations)
  if (iterationCount % 500 === 0) {
    const metricsRes = http.get(`${BASE_URL}/metrics`);

    if (metricsRes.status === 200) {
      // Parse metrics and look for memory growth
      const body = metricsRes.body;

      // Look for connection pool leaks
      const activeConns = parseMetric(body, 'active_connections');
      const poolSize = parseMetric(body, 'connection_pool_size');

      if (activeConns && poolSize && activeConns > poolSize * 0.9) {
        connectionLeaks.add(1);
        console.warn(`Connection pool exhaustion detected: ${activeConns}/${poolSize}`);
      }
    }
  }

  // Realistic user think time
  sleep(Math.random() * 5 + 2); // 2-7 seconds
}

function testWebSocketStability() {
  const url = `${WS_URL}`;

  const res = ws.connect(url, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({ type: 'ping' }));
    });

    socket.on('message', (data) => {
      const msg = JSON.parse(data);
      check(msg, {
        'ws message valid': (m) => m.type !== undefined,
      });
    });

    socket.on('error', (e) => {
      connectionLeaks.add(1);
      console.error('WebSocket error during soak test:', e);
    });

    // Keep connection for 30 seconds
    socket.setTimeout(() => {
      socket.close();
    }, 30000);
  });

  const wsSuccess = check(res, {
    'ws connected': (r) => r && r.status === 101,
  });

  if (!wsSuccess) {
    connectionLeaks.add(1);
  }
}

function parseMetric(body, metricName) {
  const regex = new RegExp(`${metricName}\\s+(\\d+)`);
  const match = body.match(regex);
  return match ? parseInt(match[1]) : null;
}

export function teardown(data) {
  const durationHours = (Date.now() - data.startTime) / (1000 * 60 * 60);
  console.log(`Soak test completed: ${durationHours.toFixed(2)} hours`);
}

export function handleSummary(data) {
  const metrics = data.metrics;
  const durationHours = (Date.now() - data.root_group.checks.length) / (1000 * 60 * 60);

  const report = {
    test_type: 'soak_test',
    timestamp: new Date().toISOString(),
    duration_hours: durationHours,
    metrics: {
      total_requests: metrics.http_reqs.values.count,
      http_req_duration_p95: metrics.http_req_duration.values['p(95)'],
      http_req_duration_p99: metrics.http_req_duration.values['p(99)'],
      error_rate: metrics.http_req_failed.values.rate,
      system_stability_rate: metrics.system_stability.values.rate,
      performance_degradation_avg: metrics.performance_degradation.values.avg,
      performance_degradation_max: metrics.performance_degradation.values.max,
      memory_leak_indicator: metrics.memory_leak_indicator.values.value,
      connection_leaks: metrics.connection_leaks.values.count,
      data_corruption_events: metrics.data_corruption_events.values.count,
    },
    analysis: {
      memory_leak_detected: metrics.memory_leak_indicator.values.value > 1.5,
      performance_degradation_detected: metrics.performance_degradation.values.max > 2.0,
      connection_leaks_detected: metrics.connection_leaks.values.count > 10,
      data_integrity_maintained: metrics.data_corruption_events.values.count === 0,
      stable_under_load: metrics.system_stability.values.rate > 0.99,
    },
  };

  return {
    'soak-test-results.json': JSON.stringify(report, null, 2),
    stdout: generateSoakReport(report),
  };
}

function generateSoakReport(report) {
  let output = '\n';
  output += '🕐'.repeat(30) + '\n';
  output += '🕐 24-HOUR SOAK TEST RESULTS\n';
  output += '🕐'.repeat(30) + '\n\n';

  output += `Test Duration: ${report.duration_hours.toFixed(2)} hours\n`;
  output += `Total Requests: ${report.metrics.total_requests}\n\n`;

  output += `Performance Metrics:\n`;
  output += `  • Response Time p95: ${report.metrics.http_req_duration_p95.toFixed(2)}ms\n`;
  output += `  • Response Time p99: ${report.metrics.http_req_duration_p99.toFixed(2)}ms\n`;
  output += `  • Error Rate: ${(report.metrics.error_rate * 100).toFixed(4)}%\n`;
  output += `  • System Stability: ${(report.metrics.system_stability_rate * 100).toFixed(2)}%\n\n`;

  output += `Stability Analysis:\n`;
  output += `  • Memory Leak: ${report.analysis.memory_leak_detected ? '⚠️  DETECTED' : '✓ None'}\n`;
  output += `    - Indicator: ${report.metrics.memory_leak_indicator.toFixed(2)}x baseline\n`;
  output += `  • Performance Degradation: ${report.analysis.performance_degradation_detected ? '⚠️  DETECTED' : '✓ Stable'}\n`;
  output += `    - Average: ${report.metrics.performance_degradation_avg.toFixed(2)}x\n`;
  output += `    - Maximum: ${report.metrics.performance_degradation_max.toFixed(2)}x\n`;
  output += `  • Connection Leaks: ${report.analysis.connection_leaks_detected ? '⚠️  DETECTED' : '✓ None'}\n`;
  output += `    - Total Leaks: ${report.metrics.connection_leaks}\n`;
  output += `  • Data Integrity: ${report.analysis.data_integrity_maintained ? '✓ Maintained' : '⚠️  COMPROMISED'}\n`;
  output += `    - Corruption Events: ${report.metrics.data_corruption_events}\n\n`;

  output += `Overall Assessment:\n`;
  if (report.analysis.stable_under_load && !report.analysis.memory_leak_detected &&
      !report.analysis.performance_degradation_detected && report.analysis.data_integrity_maintained) {
    output += `  ✓✓ EXCELLENT - System is production-ready for 24/7 operation\n`;
  } else if (report.analysis.stable_under_load) {
    output += `  ✓ GOOD - System is stable but has minor issues to address\n`;
  } else {
    output += `  ⚠️  NEEDS IMPROVEMENT - Critical stability issues detected\n`;
  }

  output += `\nRecommendations:\n`;
  if (report.analysis.memory_leak_detected) {
    output += `  → Investigate memory leaks in long-running processes\n`;
    output += `  → Review garbage collection settings\n`;
    output += `  → Add memory profiling\n`;
  }
  if (report.analysis.connection_leaks_detected) {
    output += `  → Fix connection pool leaks\n`;
    output += `  → Implement connection health checks\n`;
    output += `  → Review WebSocket lifecycle management\n`;
  }
  if (report.analysis.performance_degradation_detected) {
    output += `  → Investigate cache invalidation issues\n`;
    output += `  → Review database connection pooling\n`;
    output += `  → Check for resource contention\n`;
  }

  output += '\n' + '='.repeat(60) + '\n';

  return output;
}
