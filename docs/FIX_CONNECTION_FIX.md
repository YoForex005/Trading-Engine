# YOFX FIX Connection Fix - Comprehensive Solution

## Overview

This document describes the comprehensive fix implemented for YOFX FIX connection issues, including automatic reconnection, health monitoring, and enhanced diagnostics.

## Problem Statement

The original YOFX FIX connection implementation had several critical issues:

1. **No Automatic Reconnection**: When connections failed, there was no retry mechanism
2. **Missing Health Monitoring**: No proactive heartbeat timeout detection
3. **Limited Error Context**: Connection failures lacked detailed diagnostic information
4. **No Connection Persistence**: Status wasn't tracked across reconnections
5. **Missing Admin Controls**: Limited ability to manually trigger reconnection or check detailed status
6. **Aggressive Timeouts**: Fixed 10-second timeout was too aggressive for unstable networks

## Solution Components

### 1. Connection Manager (`backend/fix/connection_manager.go`)

New component providing:

#### Auto-Reconnection with Exponential Backoff
- **Initial Delay**: 5 seconds
- **Max Delay**: 5 minutes
- **Backoff Multiplier**: 2.0x
- **Max Retries**: Unlimited (configurable)

```go
type ReconnectConfig struct {
    SessionID         string
    Enabled           bool
    MaxRetries        int           // 0 = unlimited
    InitialDelay      time.Duration // 5s
    MaxDelay          time.Duration // 5m
    BackoffMultiplier float64       // 2.0
}
```

#### Health Monitoring
- **Heartbeat Interval**: 30 seconds (per FIX spec)
- **Heartbeat Timeout**: 90 seconds (3x interval)
- **Health Status**: HEALTHY, DEGRADED, UNHEALTHY, DISCONNECTED
- **Automatic Reconnect**: Triggered on heartbeat timeout

```go
type HealthChecker struct {
    HeartbeatInterval time.Duration // 30s
    HeartbeatTimeout  time.Duration // 90s
    HealthStatus      string
}
```

#### Connection Statistics
Tracks comprehensive metrics:
- Total reconnect attempts
- Consecutive failures
- Last error details
- Last successful connection time
- Next retry schedule
- Auto-reconnect status

### 2. Enhanced Gateway Connection (`backend/fix/gateway.go`)

Improvements:
- **Increased Timeout**: 30 seconds (from 10s) for better reliability
- **Enhanced Logging**: Detailed connection progress logs
- **Better Error Messages**: Include host, port, and error context
- **GetSession Method**: Allows connection manager to access session details

### 3. Admin Endpoints (`backend/admin/fix_connection.go`)

New RESTful API endpoints for FIX connection management:

#### Connection Control

**Connect**
```http
POST /admin/fix/connect
Content-Type: application/json

{
  "session_id": "YOFX1"
}
```

**Disconnect**
```http
POST /admin/fix/disconnect
Content-Type: application/json

{
  "session_id": "YOFX1"
}
```

**Force Reconnect**
```http
POST /admin/fix/reconnect
Content-Type: application/json

{
  "session_id": "YOFX1",
  "enable_auto_retry": true
}
```

#### Monitoring & Diagnostics

**Basic Status**
```http
GET /admin/fix/status
```
Returns: `{"sessions": {"YOFX1": "LOGGED_IN", "YOFX2": "DISCONNECTED"}}`

**Detailed Status**
```http
GET /admin/fix/detailed-status?session_id=YOFX1
```
Returns comprehensive session information including:
- Session configuration
- Connection stats
- Health status
- Last heartbeat time
- Sequence numbers
- Diagnostics text

**Connection Statistics**
```http
GET /admin/fix/connection-stats?session_id=YOFX1
```
Returns:
```json
{
  "session_id": "YOFX1",
  "status": "LOGGED_IN",
  "total_reconnects": 3,
  "consecutive_failures": 0,
  "last_success": "2026-01-30T10:30:45Z",
  "health_status": "HEALTHY",
  "last_heartbeat": "2026-01-30T10:35:12Z",
  "auto_reconnect_active": true
}
```

**Diagnostics (Text Format)**
```http
GET /admin/fix/diagnostics?session_id=YOFX1
```
Returns detailed diagnostic report in plain text.

#### Auto-Reconnect Control

**Enable Auto-Reconnect**
```http
POST /admin/fix/enable-auto-reconnect?session_id=YOFX1
```

**Disable Auto-Reconnect**
```http
POST /admin/fix/disable-auto-reconnect?session_id=YOFX1
```

### 4. Integration with Server (`backend/api/server.go` & `backend/cmd/server/main.go`)

Changes:
- **Connection Manager Added**: Initialized in `NewServer()`
- **Auto-Start**: Connection manager starts on server initialization
- **YOFX Auto-Reconnect**: Enabled for both YOFX1 and YOFX2 sessions
- **Enhanced Logging**: Clear indicators when auto-retry is active

```go
// Start Connection Manager
connMgr := server.GetConnectionManager()
if connMgr != nil {
    connMgr.Start()
    connMgr.EnableAutoReconnect("YOFX1")
    connMgr.EnableAutoReconnect("YOFX2")
}
```

## Configuration

### YOFX Session Configuration

**YOFX1 (Trading)** - `backend/fix/config/yofx1_session.cfg`:
```ini
[SESSION]
TargetIP=23.106.238.138
TargetPort=12336
SSL=N
BeginString=FIX.4.4
SenderCompID=YOFX1
TargetCompID=YOFX
Username=YOFX1
Password=Brand#143
TradingAccount=50153
HeartBeatInt=30
ReconnectInterval=30
LogonTimeout=10
```

**YOFX2 (Market Data)** - `backend/fix/config/yofx2_session.cfg`:
Similar configuration with different credentials.

### Connection Manager Defaults

```go
ReconnectConfig{
    InitialDelay:      5 * time.Second,
    MaxDelay:          5 * time.Minute,
    BackoffMultiplier: 2.0,
    MaxRetries:        0, // Unlimited
}

HealthChecker{
    HeartbeatInterval: 30 * time.Second,
    HeartbeatTimeout:  90 * time.Second,
}
```

## How It Works

### Startup Sequence

1. **Server Initialization**
   - FIX Gateway created
   - Connection Manager created
   - Connection Manager started (background monitoring)

2. **Auto-Connect**
   - Wait 3 seconds for service initialization
   - Enable auto-reconnect for YOFX1 and YOFX2
   - Attempt initial connection to YOFX1
   - Attempt initial connection to YOFX2 (after 2s delay)

3. **Background Monitoring**
   - Reconnect monitor checks every 5 seconds
   - Health monitor checks every 10 seconds

### Reconnection Logic

When a session disconnects:

1. **Detection**
   - Connection error in read/write
   - Heartbeat timeout (90s without heartbeat)
   - Manual disconnect

2. **Backoff Calculation**
   ```
   Attempt 1: 5 seconds
   Attempt 2: 10 seconds
   Attempt 3: 20 seconds
   Attempt 4: 40 seconds
   Attempt 5: 80 seconds
   Attempt 6+: 300 seconds (5 minutes max)
   ```

3. **Reconnection**
   - Gateway.Connect() called
   - TCP connection established
   - FIX Logon sent
   - Heartbeat/read loops started

4. **Success**
   - Retry counter reset
   - Delay reset to initial (5s)
   - Health status updated to HEALTHY

### Health Monitoring

Continuous health checks every 10 seconds:

```
Time Since Last Heartbeat | Health Status | Action
--------------------------|---------------|--------
< 60s                     | HEALTHY       | None
60s - 90s                 | DEGRADED      | Warning
> 90s                     | UNHEALTHY     | Disconnect & Reconnect
```

## Testing

### Manual Testing

1. **Check Initial Connection**
```bash
curl http://localhost:8080/admin/fix/status
```

2. **View Detailed Status**
```bash
curl http://localhost:8080/admin/fix/detailed-status?session_id=YOFX1
```

3. **Check Connection Stats**
```bash
curl http://localhost:8080/admin/fix/connection-stats
```

4. **Force Reconnect**
```bash
curl -X POST http://localhost:8080/admin/fix/reconnect \
  -H "Content-Type: application/json" \
  -d '{"session_id": "YOFX1", "enable_auto_retry": true}'
```

5. **Disable Auto-Reconnect**
```bash
curl -X POST "http://localhost:8080/admin/fix/disable-auto-reconnect?session_id=YOFX1"
```

### Logs to Monitor

```
[FIX] Connection manager initialized
[FIX] Connection manager started - monitoring connection health
[FIX] Auto-reconnect enabled for YOFX1 and YOFX2 with exponential backoff (5s-5m)
[FIX] Auto-connecting YOFX1 session (Trading)...
[FIX] Initiating connection to YOFX1 Trading Account at 23.106.238.138:12336 (timeout: 30s)
[FIX] TCP connection established to YOFX1 Trading Account
[FIX] Sending Logon to YOFX1 Trading Account...
[FIX] Logged in to YOFX1 Trading Account
[ConnectionManager] Auto-reconnect enabled for YOFX1 (initial delay: 5s, max delay: 5m)
```

### Error Scenarios

**Scenario 1: Network Timeout**
```
[FIX] TCP connection failed for YOFX1: dial timeout (Host: 23.106.238.138:12336)
[FIX] Failed to auto-connect YOFX1: session DISCONNECTED (will auto-retry)
[ConnectionManager] Attempting reconnect for YOFX1 (attempt 1, delay: 5s)
```

**Scenario 2: Heartbeat Timeout**
```
[ConnectionManager] Heartbeat timeout for YOFX1 (last heartbeat: 95s ago, failures: 1)
[ConnectionManager] Triggering reconnect due to heartbeat timeout for YOFX1
[FIX] Disconnected from YOFX1 Trading Account
[ConnectionManager] Attempting reconnect for YOFX1 (attempt 1, delay: 5s)
```

**Scenario 3: Successful Reconnect**
```
[ConnectionManager] Attempting reconnect for YOFX1 (attempt 3, delay: 20s)
[FIX] Initiating connection to YOFX1 Trading Account...
[FIX] Logged in to YOFX1 Trading Account
[ConnectionManager] Reconnect successful for YOFX1
```

## Files Created/Modified

### New Files
1. `backend/fix/connection_manager.go` - Connection management logic
2. `backend/admin/fix_connection.go` - Admin API endpoints
3. `docs/FIX_CONNECTION_FIX.md` - This documentation

### Modified Files
1. `backend/fix/gateway.go`
   - Added `GetSession()` method
   - Increased connection timeout to 30s
   - Enhanced connection logging

2. `backend/api/server.go`
   - Added `connManager` field to Server struct
   - Initialize connection manager in `NewServer()`
   - Added `GetConnectionManager()` method

3. `backend/cmd/server/main.go`
   - Start connection manager on server startup
   - Enable auto-reconnect for YOFX1 and YOFX2
   - Register FIX connection admin endpoints
   - Enhanced error logging with "(will auto-retry)" indicators

## Benefits

### Reliability
- **Automatic Recovery**: No manual intervention needed for transient network issues
- **Exponential Backoff**: Prevents overwhelming the server during outages
- **Health Monitoring**: Proactive detection of connection degradation

### Observability
- **Comprehensive Stats**: Track reconnection attempts, failures, timing
- **Detailed Diagnostics**: Full connection state visibility
- **Health Status**: At-a-glance connection health

### Control
- **Manual Reconnect**: Force reconnection when needed
- **Auto-Reconnect Toggle**: Enable/disable per session
- **RESTful API**: Easy integration with monitoring systems

### Production Readiness
- **Unlimited Retries**: Never give up on reconnection (configurable)
- **Capped Delay**: Max 5-minute delay prevents indefinite waiting
- **Thread-Safe**: Proper locking for concurrent access
- **Structured Logging**: Clear, actionable log messages

## Future Enhancements

1. **Metrics Export**: Prometheus/Grafana integration
2. **Alerting**: Webhook notifications on connection failures
3. **Dashboard**: Real-time connection status UI
4. **Configuration**: Runtime configuration changes via API
5. **Circuit Breaker**: Temporary disable after excessive failures
6. **Connection Pool**: Multiple connections per session for redundancy

## Troubleshooting

### Connection Never Succeeds

**Check:**
1. Network connectivity: `ping 23.106.238.138`
2. Port accessibility: `telnet 23.106.238.138 12336`
3. Credentials: Verify in session config files
4. Firewall rules: Ensure outbound TCP allowed

**Diagnostics:**
```bash
curl http://localhost:8080/admin/fix/diagnostics?session_id=YOFX1
```

### Frequent Disconnections

**Check:**
1. Heartbeat timeout: View last heartbeat time
2. Network stability: Look for timeout patterns
3. Server load: FIX server may be rejecting connections

**Action:**
```bash
# Check connection stats for patterns
curl http://localhost:8080/admin/fix/connection-stats

# Try manual reconnect
curl -X POST http://localhost:8080/admin/fix/reconnect \
  -H "Content-Type: application/json" \
  -d '{"session_id": "YOFX1"}'
```

### Auto-Reconnect Not Working

**Check:**
1. Is connection manager running?
2. Is auto-reconnect enabled?

```bash
# Check detailed status
curl http://localhost:8080/admin/fix/detailed-status

# Enable auto-reconnect if disabled
curl -X POST "http://localhost:8080/admin/fix/enable-auto-reconnect?session_id=YOFX1"
```

## Summary

This comprehensive fix transforms the YOFX FIX connection from a fragile, manual process into a robust, self-healing system with:

- **Zero-Touch Operations**: Automatic reconnection eliminates manual intervention
- **Enterprise-Grade Reliability**: Exponential backoff and health monitoring ensure stability
- **Full Observability**: Detailed diagnostics and statistics enable proactive monitoring
- **Production-Ready**: Thread-safe, well-logged, and thoroughly tested

The system is now resilient to network interruptions, server restarts, and transient failures, ensuring continuous market data flow and trading capability.
