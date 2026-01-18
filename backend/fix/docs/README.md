# FIX API Configuration

This directory contains FIX protocol session configurations for connecting to YoFx liquidity provider.

## Sessions Overview

### YOFX1 - Trading Operations
**Purpose**: Order placement, execution, and trading operations
- **Trading Account**: 50153
- **SenderCompID**: YOFX1
- **TargetCompID**: YOFX
- **Protocol**: FIX 4.4
- **SSL**: No (unencrypted connection)

### YOFX2 - Market Data Feed
**Purpose**: Market data feeds only (quotes, prices, order book)
- **Trading Account**: 50153
- **SenderCompID**: YOFX2
- **TargetCompID**: YOFX
- **Protocol**: FIX 4.4
- **SSL**: No (unencrypted connection)

## Files

### `sessions.json`
JSON format configuration for both YOFX1 and YOFX2 sessions.

## Usage

Before using this configuration:

1. **Contact your manager** to obtain:
   - Target IP address
   - Target Port number

2. **Update the configuration files** with the actual IP and port values.

3. **Security Notes**:
   - Keep credentials secure
   - Do not commit actual IP/Port to public repositories
   - Consider using environment variables for sensitive data
   - SSL is currently disabled (SSL=N) - ensure this meets your security requirements

## Integration

To integrate this session into the FIX gateway:

```go
// Add YOFX1 session to the NewFIXGateway() function in gateway.go
"YOFX1": {
    ID:           "YOFX1",
    Name:         "YOFX Trading Account",
    Host:         "YOUR_TARGET_IP",
    Port:         YOUR_TARGET_PORT,
    SenderCompID: "YOFX1",
    TargetCompID: "YOFX",
    Status:       "DISCONNECTED",
},
```

## Connection Details

**T4B FIX Server:**
- **IP**: 23.106.238.138
- **Port**: 12336
- **SSL**: No

## Session Parameters

### YOFX1 (Trading)
| Parameter | Value | Description |
|-----------|-------|-------------|
| BeginString | FIX.4.4 | FIX protocol version |
| SenderCompID | YOFX1 | Trading session ID |
| TargetCompID | YOFX | Counterparty ID |
| Username | YOFX1 | Authentication username |
| Password | Brand#143 | Authentication password |
| Trading Account | 50153 | Account number |
| HeartBeatInt | 30 | Heartbeat interval (seconds) |

### YOFX2 (Market Data)
| Parameter | Value | Description |
|-----------|-------|-------------|
| BeginString | FIX.4.4 | FIX protocol version |
| SenderCompID | YOFX2 | Market data session ID |
| TargetCompID | YOFX | Counterparty ID |
| Username | YOFX2 | Authentication username |
| Password | Brand#143 | Authentication password |
| Trading Account | 50153 | Account number |
| HeartBeatInt | 30 | Heartbeat interval (seconds) |

## Testing

### Test YOFX1 (Full Features)
```bash
cd backend/fix
go run test_all_features.go
```

### Test YOFX2 (Market Data Only)
```bash
cd backend/fix
go run test_yofx2_marketdata.go
```

## Network Troubleshooting

If you experience connection timeouts:

1. **Check ISP Port Blocking**: Some ISPs block non-standard ports
   ```bash
   nc -zv 23.106.238.138 12336
   ```

2. **Use SSH Tunnel**: If port is blocked, use the provided tunnel script
   ```bash
   ./ssh_tunnel.sh user@your-vps-ip
   ```

3. **VPN**: Connect through VPN to bypass ISP restrictions

## Next Steps

1. Resolve network connectivity to port 12336
2. Test YOFX2 market data feed connection
3. Test YOFX1 trading operations connection
4. Implement proper error handling and reconnection logic
5. Set up production monitoring
