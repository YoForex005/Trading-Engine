# Rate Limiting Documentation

## Overview

The RTX Trading Engine implements rate limiting to ensure fair usage and system stability. Rate limits are applied per IP address and per authenticated user.

**Current Status:** Rate limiting is planned but not yet implemented in the codebase.

**Future Implementation:** Token bucket algorithm with Redis backend for distributed rate limiting.

## Rate Limit Tiers

### Public Endpoints (Unauthenticated)

| Endpoint Pattern | Limit | Window | Burst |
|-----------------|-------|--------|-------|
| `/health` | 100 | 1 minute | 10 |
| `/login` | 5 | 1 minute | 2 |
| `/docs` | 20 | 1 minute | 5 |

### Trader Endpoints (Authenticated)

| Endpoint Pattern | Limit | Window | Burst |
|-----------------|-------|--------|-------|
| `/api/orders/*` | 100 | 1 minute | 20 |
| `/api/positions/*` | 100 | 1 minute | 20 |
| `/api/account/*` | 60 | 1 minute | 10 |
| `/ticks` | 30 | 1 minute | 5 |
| `/ohlc` | 30 | 1 minute | 5 |
| Market data queries | 120 | 1 minute | 30 |

### Admin Endpoints

| Endpoint Pattern | Limit | Window | Burst |
|-----------------|-------|--------|-------|
| `/admin/*` | 200 | 1 minute | 50 |
| `/admin/lps/*` | 30 | 1 minute | 10 |
| `/admin/fix/*` | 20 | 1 minute | 5 |

### WebSocket

| Connection Type | Limit | Description |
|----------------|-------|-------------|
| Connections per IP | 10 | Max concurrent connections |
| Messages per second | 100 | Outbound message rate |
| Reconnect attempts | 5 per minute | Prevent flood |

## Response Headers

When rate limiting is active, all responses include these headers:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705838445
X-RateLimit-Window: 60
```

**Header Descriptions:**

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Total requests allowed in window | 100 |
| `X-RateLimit-Remaining` | Requests remaining in current window | 95 |
| `X-RateLimit-Reset` | Unix timestamp when limit resets | 1705838445 |
| `X-RateLimit-Window` | Rate limit window in seconds | 60 |

## Rate Limit Exceeded Response

**HTTP Status:** 429 Too Many Requests

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "limit": 100,
    "window": "1m",
    "retryAfter": 45,
    "resetAt": 1705838445
  }
}
```

**Headers:**
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705838445
Retry-After: 45
Content-Type: application/json
```

## Implementation Examples

### JavaScript/TypeScript

```typescript
interface RateLimitHeaders {
  limit: number;
  remaining: number;
  reset: number;
  window: number;
}

class APIClient {
  private rateLimits: Map<string, RateLimitHeaders> = new Map();

  async request(endpoint: string, options: RequestInit): Promise<Response> {
    const response = await fetch(endpoint, options);

    // Parse rate limit headers
    const limit = parseInt(response.headers.get('X-RateLimit-Limit') || '0');
    const remaining = parseInt(response.headers.get('X-RateLimit-Remaining') || '0');
    const reset = parseInt(response.headers.get('X-RateLimit-Reset') || '0');
    const window = parseInt(response.headers.get('X-RateLimit-Window') || '60');

    this.rateLimits.set(endpoint, { limit, remaining, reset, window });

    if (response.status === 429) {
      const retryAfter = parseInt(response.headers.get('Retry-After') || '60');
      throw new RateLimitError(retryAfter, this.rateLimits.get(endpoint)!);
    }

    return response;
  }

  getRemainingRequests(endpoint: string): number {
    return this.rateLimits.get(endpoint)?.remaining ?? Infinity;
  }

  getResetTime(endpoint: string): Date | null {
    const reset = this.rateLimits.get(endpoint)?.reset;
    return reset ? new Date(reset * 1000) : null;
  }
}

class RateLimitError extends Error {
  constructor(
    public retryAfter: number,
    public rateLimitInfo: RateLimitHeaders
  ) {
    super(`Rate limit exceeded. Retry after ${retryAfter}s`);
  }
}

// Usage
const client = new APIClient();

try {
  const response = await client.request('http://localhost:7999/api/orders/market', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: JSON.stringify(orderData)
  });

  console.log(`Remaining requests: ${client.getRemainingRequests('/api/orders/market')}`);
} catch (error) {
  if (error instanceof RateLimitError) {
    console.log(`Rate limited. Retry in ${error.retryAfter}s`);
    await sleep(error.retryAfter * 1000);
  }
}
```

### Python

```python
import time
import requests
from typing import Optional, Dict
from datetime import datetime

class RateLimitInfo:
    def __init__(self, limit: int, remaining: int, reset: int, window: int):
        self.limit = limit
        self.remaining = remaining
        self.reset = reset
        self.window = window

    @property
    def reset_datetime(self) -> datetime:
        return datetime.fromtimestamp(self.reset)

    @property
    def seconds_until_reset(self) -> int:
        return max(0, self.reset - int(time.time()))

class RateLimitError(Exception):
    def __init__(self, retry_after: int, info: RateLimitInfo):
        self.retry_after = retry_after
        self.info = info
        super().__init__(f"Rate limit exceeded. Retry after {retry_after}s")

class RTXClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.rate_limits: Dict[str, RateLimitInfo] = {}

    def request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        url = f"{self.base_url}{endpoint}"
        headers = kwargs.pop('headers', {})
        headers['Authorization'] = f'Bearer {self.token}'

        response = requests.request(method, url, headers=headers, **kwargs)

        # Parse rate limit headers
        if 'X-RateLimit-Limit' in response.headers:
            info = RateLimitInfo(
                limit=int(response.headers.get('X-RateLimit-Limit', 0)),
                remaining=int(response.headers.get('X-RateLimit-Remaining', 0)),
                reset=int(response.headers.get('X-RateLimit-Reset', 0)),
                window=int(response.headers.get('X-RateLimit-Window', 60))
            )
            self.rate_limits[endpoint] = info

        if response.status_code == 429:
            retry_after = int(response.headers.get('Retry-After', 60))
            raise RateLimitError(retry_after, self.rate_limits[endpoint])

        response.raise_for_status()
        return response

    def get_remaining_requests(self, endpoint: str) -> Optional[int]:
        info = self.rate_limits.get(endpoint)
        return info.remaining if info else None

    def wait_for_reset(self, endpoint: str):
        info = self.rate_limits.get(endpoint)
        if info:
            wait_time = info.seconds_until_reset
            if wait_time > 0:
                print(f"Waiting {wait_time}s for rate limit reset...")
                time.sleep(wait_time)

# Usage
client = RTXClient('http://localhost:7999', token)

try:
    response = client.request('POST', '/api/orders/market', json=order_data)
    remaining = client.get_remaining_requests('/api/orders/market')
    print(f"Remaining requests: {remaining}")
except RateLimitError as e:
    print(f"Rate limited. Retry in {e.retry_after}s")
    time.sleep(e.retry_after)
```

### Go

```go
package main

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
)

type RateLimitInfo struct {
    Limit     int
    Remaining int
    Reset     int64
    Window    int
}

type RateLimitError struct {
    RetryAfter int
    Info       RateLimitInfo
}

func (e *RateLimitError) Error() string {
    return fmt.Sprintf("rate limit exceeded, retry after %ds", e.RetryAfter)
}

type Client struct {
    BaseURL    string
    Token      string
    HTTPClient *http.Client
    RateLimits map[string]RateLimitInfo
}

func NewClient(baseURL, token string) *Client {
    return &Client{
        BaseURL:    baseURL,
        Token:      token,
        HTTPClient: &http.Client{Timeout: 10 * time.Second},
        RateLimits: make(map[string]RateLimitInfo),
    }
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set("Authorization", "Bearer "+c.Token)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }

    // Parse rate limit headers
    if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
        info := RateLimitInfo{}
        info.Limit, _ = strconv.Atoi(limit)
        info.Remaining, _ = strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
        reset, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
        info.Reset = reset
        info.Window, _ = strconv.Atoi(resp.Header.Get("X-RateLimit-Window"))

        c.RateLimits[req.URL.Path] = info
    }

    // Handle rate limit
    if resp.StatusCode == http.StatusTooManyRequests {
        retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
        return resp, &RateLimitError{
            RetryAfter: retryAfter,
            Info:       c.RateLimits[req.URL.Path],
        }
    }

    return resp, nil
}

func (c *Client) GetRemainingRequests(path string) int {
    if info, ok := c.RateLimits[path]; ok {
        return info.Remaining
    }
    return -1
}

func (c *Client) WaitForReset(path string) {
    if info, ok := c.RateLimits[path]; ok {
        now := time.Now().Unix()
        if info.Reset > now {
            waitTime := time.Duration(info.Reset-now) * time.Second
            fmt.Printf("Waiting %v for rate limit reset...\n", waitTime)
            time.Sleep(waitTime)
        }
    }
}
```

## Best Practices

### 1. Monitor Rate Limit Headers

Always check `X-RateLimit-Remaining` before making requests:

```javascript
if (client.getRemainingRequests('/api/orders/market') < 10) {
  console.warn('Approaching rate limit!');
  await throttle();
}
```

### 2. Implement Exponential Backoff

```javascript
async function withRetry(fn, maxRetries = 3) {
  let delay = 1000;

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (error instanceof RateLimitError) {
        delay = error.retryAfter * 1000;
      }

      if (i === maxRetries - 1) throw error;

      await sleep(delay);
      delay *= 2; // Exponential backoff
    }
  }
}
```

### 3. Use Request Queuing

```typescript
class RequestQueue {
  private queue: Array<() => Promise<any>> = [];
  private processing = false;
  private requestsPerSecond = 10;

  async add<T>(fn: () => Promise<T>): Promise<T> {
    return new Promise((resolve, reject) => {
      this.queue.push(async () => {
        try {
          const result = await fn();
          resolve(result);
        } catch (error) {
          reject(error);
        }
      });

      this.process();
    });
  }

  private async process() {
    if (this.processing || this.queue.length === 0) return;

    this.processing = true;

    while (this.queue.length > 0) {
      const fn = this.queue.shift()!;
      await fn();
      await sleep(1000 / this.requestsPerSecond);
    }

    this.processing = false;
  }
}

// Usage
const queue = new RequestQueue();
const result = await queue.add(() => placeOrder(orderData));
```

### 4. Cache Responses

Reduce API calls by caching responses:

```javascript
class CachedAPIClient {
  private cache = new Map();
  private cacheTTL = 5000; // 5 seconds

  async get(endpoint) {
    const cached = this.cache.get(endpoint);
    if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
      return cached.data;
    }

    const data = await this.fetch(endpoint);
    this.cache.set(endpoint, { data, timestamp: Date.now() });
    return data;
  }
}
```

### 5. Batch Operations

Use bulk endpoints when available:

```javascript
// ❌ Bad: Multiple requests
for (const positionId of positionIds) {
  await closePosition(positionId);
}

// ✅ Good: Single batch request
await closePositionsBulk(positionIds);
```

## Rate Limit Bypass for Admin

Admin accounts have higher rate limits:

| Endpoint | Trader Limit | Admin Limit |
|----------|--------------|-------------|
| `/api/orders/*` | 100/min | 500/min |
| `/api/positions/*` | 100/min | 500/min |
| `/admin/*` | N/A | 1000/min |

## IP-Based Rate Limiting

Rate limits are tracked per IP address for unauthenticated endpoints:

- Public endpoints: 100 requests per minute per IP
- Login attempts: 5 per minute per IP (brute force protection)

## Distributed Rate Limiting

**Future Implementation:** Redis-backed distributed rate limiting for multi-server deployments.

```javascript
// Pseudocode for distributed rate limiting
class RedisRateLimiter {
  async checkLimit(key, limit, window) {
    const current = await redis.incr(key);

    if (current === 1) {
      await redis.expire(key, window);
    }

    if (current > limit) {
      const ttl = await redis.ttl(key);
      throw new RateLimitError(ttl);
    }

    return {
      limit,
      remaining: limit - current,
      reset: Date.now() + (await redis.ttl(key)) * 1000
    };
  }
}
```

## WebSocket Rate Limiting

WebSocket connections have separate rate limits:

### Connection Limits

- **Max connections per IP:** 10
- **Reconnect attempts:** 5 per minute
- **Connection duration:** Unlimited

### Message Limits

- **Outbound messages:** 100 per second (server to client)
- **Inbound messages:** 10 per second (client to server)

**Enforcement:** Slow consumers are automatically disconnected.

## Summary

| Aspect | Current | Planned |
|--------|---------|---------|
| Implementation | ❌ Not implemented | ✅ Token bucket |
| Storage | N/A | Redis |
| Per-user limits | ❌ | ✅ |
| Per-IP limits | ❌ | ✅ |
| Burst capacity | ❌ | ✅ |
| Headers | ❌ | ✅ |
| WebSocket limits | ❌ | ✅ |

**Note:** Rate limiting is currently planned but not enforced. Production deployment will include full rate limiting implementation.
