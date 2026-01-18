// k6 Spike Test - Sudden load increase
// Run with: k6 run spike-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const spikeRecovery = new Trend('spike_recovery_time');
const errorsDuringSpike = new Counter('spike_errors');
const recoverySuccess = new Rate('recovery_success');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Multiple spike scenarios
export const options = {
  scenarios: {
    // Normal load baseline
    normal_load: {
      executor: 'constant-vus',
      vus: 100,
      duration: '20m',
      startTime: '0s',
    },

    // First spike - moderate
    spike_1: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 2000 },  // Sudden spike
        { duration: '1m', target: 2000 },   // Hold
        { duration: '10s', target: 0 },     // Drop
      ],
      startTime: '3m',
    },

    // Second spike - severe
    spike_2: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5s', target: 5000 },   // Very sudden spike
        { duration: '2m', target: 5000 },   // Hold longer
        { duration: '10s', target: 0 },     // Drop
      ],
      startTime: '8m',
    },

    // Third spike - extreme
    spike_3: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '3s', target: 10000 },  // Extreme spike
        { duration: '30s', target: 10000 }, // Short hold
        { duration: '5s', target: 0 },      // Fast drop
      ],
      startTime: '15m',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'spike_errors': ['count<1000'], // Allow some errors during spikes
    'recovery_success': ['rate>0.95'], // 95% recovery success
  },
};

const symbols = ['BTCUSD', 'ETHUSD', 'XRPUSD', 'BNBUSD', 'SOLUSD'];

export function setup() {
  console.log('Starting spike test - testing sudden load increases');
  return {
    startTime: Date.now(),
    spikeEvents: [],
  };
}

export default function (data) {
  const executorName = __ENV.SCENARIO || 'unknown';
  const isSpike = executorName.startsWith('spike_');
  const startTime = Date.now();

  // Generate order
  const order = {
    symbol: symbols[Math.floor(Math.random() * symbols.length)],
    side: Math.random() > 0.5 ? 'buy' : 'sell',
    type: 'market',
    quantity: (Math.random() * 10 + 0.1).toFixed(2),
  };

  const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify(order), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '30s',
    tags: { scenario: executorName },
  });

  const responseTime = Date.now() - startTime;

  const success = check(res, {
    'status is 200-201': (r) => r.status >= 200 && r.status < 300,
    'response time acceptable': (r) => r.timings.duration < 2000,
  });

  // Track errors during spike
  if (isSpike && !success) {
    errorsDuringSpike.add(1);
  }

  // Track recovery after spike
  if (isSpike && success && responseTime < 500) {
    recoverySuccess.add(1);
    spikeRecovery.add(responseTime);
  } else if (isSpike && !success) {
    recoverySuccess.add(0);
  }

  // Realistic behavior
  if (isSpike) {
    sleep(0.05); // Minimal sleep during spike
  } else {
    sleep(Math.random() * 2 + 1); // Normal user behavior
  }
}

export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`Spike test completed in ${duration} seconds`);
}

export function handleSummary(data) {
  const metrics = data.metrics;

  const report = {
    test_type: 'spike_test',
    timestamp: new Date().toISOString(),
    spike_events: [
      {
        name: 'Moderate Spike',
        target_vus: 2000,
        duration: '1m',
        start_time: '3m',
      },
      {
        name: 'Severe Spike',
        target_vus: 5000,
        duration: '2m',
        start_time: '8m',
      },
      {
        name: 'Extreme Spike',
        target_vus: 10000,
        duration: '30s',
        start_time: '15m',
      },
    ],
    metrics: {
      http_req_duration_p95: metrics.http_req_duration.values['p(95)'],
      http_req_duration_p99: metrics.http_req_duration.values['p(99)'],
      spike_errors: metrics.spike_errors.values.count,
      recovery_success_rate: metrics.recovery_success.values.rate,
      avg_recovery_time: metrics.spike_recovery.values.avg,
      total_requests: metrics.http_reqs.values.count,
    },
    analysis: {
      handles_moderate_spike: metrics.spike_errors.values.count < 100,
      handles_severe_spike: metrics.spike_errors.values.count < 500,
      handles_extreme_spike: metrics.spike_errors.values.count < 1000,
      fast_recovery: metrics.spike_recovery.values.avg < 200,
    },
  };

  return {
    'spike-test-results.json': JSON.stringify(report, null, 2),
    stdout: generateSpikeReport(report),
  };
}

function generateSpikeReport(report) {
  let output = '\n';
  output += '⚡'.repeat(30) + '\n';
  output += '⚡ SPIKE TEST RESULTS\n';
  output += '⚡'.repeat(30) + '\n\n';

  output += `Spike Events Tested:\n`;
  report.spike_events.forEach((spike, idx) => {
    output += `  ${idx + 1}. ${spike.name}\n`;
    output += `     • Target: ${spike.target_vus} concurrent users\n`;
    output += `     • Duration: ${spike.duration}\n`;
    output += `     • Start Time: ${spike.start_time}\n`;
  });
  output += '\n';

  output += `Performance During Spikes:\n`;
  output += `  • Response Time p95: ${report.metrics.http_req_duration_p95.toFixed(2)}ms\n`;
  output += `  • Response Time p99: ${report.metrics.http_req_duration_p99.toFixed(2)}ms\n`;
  output += `  • Errors During Spikes: ${report.metrics.spike_errors}\n`;
  output += `  • Recovery Success Rate: ${(report.metrics.recovery_success_rate * 100).toFixed(2)}%\n`;
  output += `  • Average Recovery Time: ${report.metrics.avg_recovery_time.toFixed(2)}ms\n\n`;

  output += `System Resilience:\n`;
  output += `  • Moderate Spike (2K users): ${report.analysis.handles_moderate_spike ? '✓ Handled' : '✗ Failed'}\n`;
  output += `  • Severe Spike (5K users): ${report.analysis.handles_severe_spike ? '✓ Handled' : '✗ Failed'}\n`;
  output += `  • Extreme Spike (10K users): ${report.analysis.handles_extreme_spike ? '✓ Handled' : '⚠ Degraded'}\n`;
  output += `  • Fast Recovery: ${report.analysis.fast_recovery ? '✓ Yes (<200ms)' : '⚠ Slow'}\n\n`;

  output += `Recommendations:\n`;
  if (!report.analysis.handles_moderate_spike) {
    output += `  ⚠️  Critical: Cannot handle moderate traffic spikes\n`;
    output += `  → Implement auto-scaling\n`;
    output += `  → Add request queuing\n`;
    output += `  → Review connection pool settings\n`;
  } else if (!report.analysis.handles_severe_spike) {
    output += `  ⚠️  Warning: Struggles with severe spikes\n`;
    output += `  → Implement circuit breakers\n`;
    output += `  → Add rate limiting\n`;
    output += `  → Consider CDN for static content\n`;
  } else if (!report.analysis.handles_extreme_spike) {
    output += `  ✓ Good resilience for normal spikes\n`;
    output += `  → Consider load shedding for extreme cases\n`;
  } else {
    output += `  ✓✓ Excellent spike resilience\n`;
    output += `  → System handles all spike scenarios well\n`;
  }

  output += '\n' + '='.repeat(60) + '\n';

  return output;
}
