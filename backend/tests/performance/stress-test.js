// k6 Stress Test - Find breaking point
// Run with: k6 run stress-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const systemBreakpoint = new Counter('system_breakpoint_reached');
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Aggressive ramping to find breaking point
export const options = {
  scenarios: {
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 500 },   // Warm up
        { duration: '2m', target: 1000 },  // Increase load
        { duration: '2m', target: 2000 },  // Push harder
        { duration: '2m', target: 3000 },  // Push to limits
        { duration: '2m', target: 5000 },  // Breaking point
        { duration: '2m', target: 7000 },  // Beyond breaking point
        { duration: '2m', target: 10000 }, // Maximum stress
        { duration: '5m', target: 0 },     // Recovery
      ],
      gracefulRampDown: '1m',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<1000'], // Allow degradation, but not total failure
    'errors': ['rate<0.1'], // Allow up to 10% errors to find breaking point
  },
};

const symbols = ['BTCUSD', 'ETHUSD', 'XRPUSD'];

export function setup() {
  console.log('Starting stress test - finding system breaking point');
  return { startTime: Date.now() };
}

export default function () {
  const startTime = Date.now();

  // Aggressive order placement - no think time
  const order = {
    symbol: symbols[Math.floor(Math.random() * symbols.length)],
    side: Math.random() > 0.5 ? 'buy' : 'sell',
    type: 'market',
    quantity: (Math.random() * 5 + 0.1).toFixed(2),
  };

  const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify(order), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s', // Allow longer timeout for stressed system
  });

  const success = check(res, {
    'status is 200-201': (r) => r.status >= 200 && r.status < 300,
    'response time < 5s': (r) => r.timings.duration < 5000,
  });

  if (!success) {
    errorRate.add(1);

    // Detect breaking point
    if (res.status >= 500 || res.timings.duration > 10000) {
      systemBreakpoint.add(1);
      console.log(`Breaking point detected at VU: ${__VU}, Status: ${res.status}, Duration: ${res.timings.duration}ms`);
    }
  } else {
    errorRate.add(0);
  }

  responseTime.add(Date.now() - startTime);

  // Minimal sleep to maintain maximum pressure
  sleep(0.1);
}

export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`\n${'='.repeat(60)}`);
  console.log(`Stress test completed in ${duration} seconds`);
  console.log(`Breaking point analysis complete`);
  console.log(`${'='.repeat(60)}\n`);
}

export function handleSummary(data) {
  const metrics = data.metrics;

  // Calculate breaking point
  let breakingPointVUs = 0;
  if (metrics.system_breakpoint_reached && metrics.system_breakpoint_reached.values.count > 0) {
    // Estimate based on timeline
    breakingPointVUs = estimateBreakingPoint(data);
  }

  const report = {
    test_type: 'stress_test',
    timestamp: new Date().toISOString(),
    breaking_point_vus: breakingPointVUs,
    max_vus_tested: 10000,
    metrics: {
      http_req_duration_p95: metrics.http_req_duration.values['p(95)'],
      http_req_duration_p99: metrics.http_req_duration.values['p(99)'],
      error_rate: metrics.errors.values.rate,
      total_requests: metrics.http_reqs.values.count,
      total_errors: metrics.errors.values.count,
    },
    system_behavior: {
      graceful_degradation: metrics.errors.values.rate < 0.5,
      catastrophic_failure: metrics.errors.values.rate >= 0.5,
    },
  };

  return {
    'stress-test-results.json': JSON.stringify(report, null, 2),
    stdout: generateStressReport(report),
  };
}

function estimateBreakingPoint(data) {
  // Analyze VU progression and error rates to estimate breaking point
  // This is a simplified estimation
  const errorRate = data.metrics.errors.values.rate;

  if (errorRate < 0.05) return 10000; // System handled max load
  if (errorRate < 0.10) return 7000;
  if (errorRate < 0.20) return 5000;
  if (errorRate < 0.30) return 3000;
  if (errorRate < 0.50) return 2000;
  return 1000;
}

function generateStressReport(report) {
  let output = '\n';
  output += '█'.repeat(60) + '\n';
  output += '█ STRESS TEST RESULTS\n';
  output += '█'.repeat(60) + '\n\n';

  output += `Breaking Point Analysis:\n`;
  output += `  • Estimated Breaking Point: ~${report.breaking_point_vus} concurrent users\n`;
  output += `  • Max VUs Tested: ${report.max_vus_tested}\n`;
  output += `  • System Behavior: ${report.system_behavior.graceful_degradation ? 'Graceful Degradation ✓' : 'Catastrophic Failure ✗'}\n\n`;

  output += `Performance Metrics:\n`;
  output += `  • Response Time p95: ${report.metrics.http_req_duration_p95.toFixed(2)}ms\n`;
  output += `  • Response Time p99: ${report.metrics.http_req_duration_p99.toFixed(2)}ms\n`;
  output += `  • Error Rate: ${(report.metrics.error_rate * 100).toFixed(2)}%\n`;
  output += `  • Total Requests: ${report.metrics.total_requests}\n`;
  output += `  • Total Errors: ${report.metrics.total_errors}\n\n`;

  output += `Recommendations:\n`;
  if (report.breaking_point_vus < 5000) {
    output += `  ⚠️  System capacity below target (5000 users)\n`;
    output += `  → Consider horizontal scaling\n`;
    output += `  → Optimize database queries\n`;
    output += `  → Implement caching layer\n`;
  } else if (report.breaking_point_vus < 7000) {
    output += `  ✓ System capacity acceptable\n`;
    output += `  → Minor optimizations recommended\n`;
  } else {
    output += `  ✓✓ Excellent system capacity\n`;
    output += `  → System performs well under stress\n`;
  }

  output += '\n' + '='.repeat(60) + '\n';

  return output;
}
