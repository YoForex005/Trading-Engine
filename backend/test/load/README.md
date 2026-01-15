# Load Testing

Load tests verify the trading platform handles concurrent users and high-frequency data.

## Prerequisites

Install k6:
```bash
brew install k6  # macOS
# For other platforms: https://k6.io/docs/get-started/installation/
```

## Running Load Tests

### WebSocket Load Test
Tests concurrent WebSocket connections and tick delivery:

```bash
cd backend/test/load
k6 run websocket_load.js
```

Override defaults:
```bash
k6 run -e WS_URL=ws://staging.example.com/ws websocket_load.js
```

### API Load Test
Tests order placement and API endpoints under load:

```bash
k6 run api_load.js
```

Override defaults:
```bash
k6 run -e API_URL=https://staging.example.com api_load.js
```

## Performance Targets

| Metric | Target |
|--------|--------|
| Concurrent WebSocket connections | 100+ |
| WebSocket tick latency (p95) | <500ms |
| WebSocket tick latency (p99) | <1000ms |
| API order placement (p95) | <200ms |
| API error rate | <5% |
| Order success rate | >95% |

## Interpreting Results

k6 outputs metrics at the end of the test:

- **http_req_duration**: Time to complete HTTP requests (p95, p99)
- **ws_connecting**: Time to establish WebSocket connections
- **ticks_received**: Total ticks delivered to all clients
- **tick_latency**: Latency from tick generation to client receipt
- **order_success_rate**: Percentage of orders successfully executed

### Example Output

```
     ✓ tick has symbol
     ✓ tick has bid
     ✓ ask > bid

     ticks_received................: 65432
     tick_latency..................: avg=245ms p(95)=420ms p(99)=780ms
     ws_connecting.................: avg=450ms p(95)=1200ms
```

## Troubleshooting

**Connection failures**: Check backend is running and WebSocket endpoint accessible
**High latency**: Check CPU/memory usage on backend, may need scaling
**Failed thresholds**: Performance below targets, investigate bottlenecks

## CI/CD Integration

Run load tests in CI/CD pipeline:

```bash
k6 run --out json=results.json websocket_load.js
```

Parse results.json to fail build if thresholds not met.
