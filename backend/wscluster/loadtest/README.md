# WebSocket Load Testing Tool

Load testing tool for WebSocket cluster designed to simulate 100,000+ concurrent connections.

## Features

- Configurable connection rate and total connections
- Message throughput testing
- Latency measurements (P50, P95, P99)
- Connection and message success rates
- Real-time status reporting

## Installation

```bash
cd loadtest
go build -o loadtest
```

## Usage

### Basic Load Test

```bash
# Test with 1,000 connections
./loadtest -url ws://localhost:8080/ws -connections 1000

# Test with 10,000 connections at 500/sec
./loadtest -url ws://localhost:8080/ws -connections 10000 -rate 500

# Test for 5 minutes with 50,000 connections
./loadtest -url ws://localhost:8080/ws -connections 50000 -duration 5m
```

### Full Parameters

```bash
./loadtest \
  -url ws://localhost:8080/ws \
  -connections 100000 \
  -rate 1000 \
  -duration 10m \
  -messages 100 \
  -interval 1s \
  -payload 256 \
  -verbose
```

### Parameters

- `-url` - WebSocket endpoint URL (default: ws://localhost:8080/ws)
- `-connections` - Total number of connections (default: 1000)
- `-rate` - Connections per second (default: 100)
- `-duration` - Test duration (default: 60s)
- `-messages` - Messages per connection (default: 100)
- `-interval` - Interval between messages (default: 1s)
- `-payload` - Message payload size in bytes (default: 100)
- `-verbose` - Verbose logging (default: false)

## Test Scenarios

### Scenario 1: Connection Stress Test
Test cluster's ability to handle rapid connection establishment:

```bash
./loadtest -connections 50000 -rate 2000 -messages 10
```

### Scenario 2: Message Throughput Test
Test message processing capacity:

```bash
./loadtest -connections 10000 -messages 1000 -interval 100ms
```

### Scenario 3: Sustained Load Test
Test stability under sustained load:

```bash
./loadtest -connections 20000 -duration 30m -messages 10000 -interval 1s
```

### Scenario 4: 100k Connection Test
Full-scale test (requires cluster setup):

```bash
# Run from multiple machines to reach 100k+
# Machine 1:
./loadtest -connections 25000 -duration 10m

# Machine 2:
./loadtest -connections 25000 -duration 10m

# Machine 3:
./loadtest -connections 25000 -duration 10m

# Machine 4:
./loadtest -connections 25000 -duration 10m
```

## Sample Output

```
Starting load test with configuration:
  Target URL: ws://localhost:8080/ws
  Total connections: 10000
  Connection rate: 500/sec
  Test duration: 1m0s
  Messages per connection: 100
  Message interval: 1s
  Payload size: 100 bytes

Status: 2500 active connections, 125000 messages sent, 124980 messages received
Status: 5000 active connections, 250000 messages sent, 249850 messages received
Status: 7500 active connections, 375000 messages sent, 374700 messages received
Status: 10000 active connections, 500000 messages sent, 499500 messages received

========================================
Load Test Results
========================================

Test Configuration:
  Target URL:           ws://localhost:8080/ws
  Total connections:    10000
  Test duration:        1m2.456s

Connection Statistics:
  Successful:           10000
  Failed:               0
  Success rate:         100.00%
  Connection latency:
    Min:     5.234ms
    Max:     45.123ms
    Average: 12.456ms
    P50:     10.234ms
    P95:     25.678ms
    P99:     38.901ms

Message Statistics:
  Sent:                 1000000
  Received:             998500
  Failed:               0
  Messages/sec:         16025.64
  Message latency:
    Min:     0.234ms
    Max:     15.678ms
    Average: 2.345ms
    P50:     1.890ms
    P95:     5.678ms
    P99:     8.901ms

Throughput:
  Connections/sec:      160.26
  Messages/sec:         16025.64
  Throughput:           3.05 MB/s

========================================
```

## Monitoring During Test

### Watch Cluster Metrics
```bash
# In another terminal, monitor cluster status
watch -n 1 'redis-cli KEYS "ws:cluster:node:*" | xargs -I {} redis-cli GET {}'
```

### Check Node Load
```bash
# SSH to cluster nodes
ssh node1 "top -b -n 1 | head -20"
```

### Monitor Redis
```bash
redis-cli INFO stats | grep instantaneous
```

## Performance Tips

1. **Increase File Descriptors**: Required for 100k+ connections
   ```bash
   ulimit -n 200000
   ```

2. **Tune OS Parameters**:
   ```bash
   # Increase TCP buffer sizes
   sudo sysctl -w net.core.rmem_max=16777216
   sudo sysctl -w net.core.wmem_max=16777216
   sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
   sudo sysctl -w net.ipv4.tcp_wmem="4096 16384 16777216"

   # Increase connection tracking
   sudo sysctl -w net.netfilter.nf_conntrack_max=1000000
   sudo sysctl -w net.nf_conntrack_max=1000000
   ```

3. **Use Multiple Test Machines**: Distribute load across machines to avoid client-side bottlenecks.

4. **Gradual Ramp-Up**: Start with low connection rates and gradually increase.

## Troubleshooting

### "Too many open files" error
```bash
ulimit -n 200000
```

### High CPU on test machine
Reduce connection rate or run from multiple machines.

### Connection timeouts
Increase handshake timeout in code or reduce connection rate.

### Memory issues
Reduce payload size or number of messages per connection.

## Automated Testing

### Run Test Suite
```bash
#!/bin/bash
# test-suite.sh

echo "Running WebSocket load test suite..."

# Test 1: Connection stress
echo "Test 1: Connection stress (50k connections)"
./loadtest -connections 50000 -rate 2000 -messages 10 > results-test1.txt

# Test 2: Message throughput
echo "Test 2: Message throughput (10k connections, 1k messages each)"
./loadtest -connections 10000 -messages 1000 -interval 100ms > results-test2.txt

# Test 3: Sustained load
echo "Test 3: Sustained load (20k connections, 10 minutes)"
./loadtest -connections 20000 -duration 10m > results-test3.txt

echo "Test suite complete. Check results-test*.txt files."
```

## License

MIT
