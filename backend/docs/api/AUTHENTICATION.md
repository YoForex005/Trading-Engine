# Authentication Documentation

## Overview

The RTX Trading Engine uses **JWT (JSON Web Tokens)** for authentication. All authenticated endpoints require a valid JWT token in the Authorization header.

## Authentication Flow

```
┌──────────┐                 ┌──────────┐                 ┌──────────┐
│  Client  │                 │   API    │                 │   Auth   │
└────┬─────┘                 └────┬─────┘                 └────┬─────┘
     │                            │                            │
     │  POST /login               │                            │
     ├────────────────────────────►                            │
     │  { username, password }    │                            │
     │                            │  Validate credentials      │
     │                            ├────────────────────────────►
     │                            │                            │
     │                            │  Generate JWT token        │
     │                            │◄────────────────────────────┤
     │  { token, user }           │                            │
     │◄────────────────────────────┤                            │
     │                            │                            │
     │  GET /api/account/summary  │                            │
     │  Authorization: Bearer ... │                            │
     ├────────────────────────────►                            │
     │                            │  Verify token              │
     │                            ├────────────────────────────►
     │                            │                            │
     │  { accountSummary }        │                            │
     │◄────────────────────────────┤                            │
```

## Login Endpoint

### POST /login

Authenticate user and receive JWT token.

**Request:**
```http
POST /login HTTP/1.1
Host: localhost:7999
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**Response (Success):**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjAiLCJ1c2VybmFtZSI6ImFkbWluIiwicm9sZSI6IkFETUlOIiwiZXhwIjoxNzA1OTI0ODAwfQ.signature",
  "user": {
    "id": "0",
    "username": "admin",
    "role": "ADMIN"
  }
}
```

**Response (Failure):**
```http
HTTP/1.1 401 Unauthorized
Content-Type: text/plain

Unauthorized
```

## Account Types

### 1. Admin Account

**Username:** `admin`
**Password:** `password` (default, should be changed)
**Role:** `ADMIN`

**Permissions:**
- Full access to all endpoints
- LP management (`/admin/lps/*`)
- FIX session management (`/admin/fix/*`)
- Account management (`/admin/accounts`, `/admin/deposit`, etc.)
- Configuration changes (`/api/config`)

### 2. Trader Account

**Username:** Account ID or username
**Role:** `TRADER`

**Permissions:**
- Order placement (`/api/orders/*`, `/order`)
- Position management (`/api/positions/*`)
- Account information (`/api/account/*`)
- Market data (`/ticks`, `/ohlc`)
- Risk calculations (`/risk/*`)

## JWT Token Structure

### Token Payload

```json
{
  "id": "1",
  "username": "trader001",
  "role": "TRADER",
  "exp": 1705924800
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | User ID |
| `username` | string | Username |
| `role` | string | User role (ADMIN or TRADER) |
| `exp` | number | Expiration timestamp (Unix) |

### Token Lifetime

- **Default expiration:** 24 hours
- **Refresh:** Re-authenticate via `/login`
- **No refresh tokens:** Must login again after expiration

## Using JWT Tokens

### HTTP Headers

Include the JWT token in the `Authorization` header with the `Bearer` scheme:

```http
GET /api/account/summary HTTP/1.1
Host: localhost:7999
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### JavaScript Example

```javascript
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';

const response = await fetch('http://localhost:7999/api/account/summary', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
```

### Python Example

```python
import requests

token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'

response = requests.get(
    'http://localhost:7999/api/account/summary',
    headers={
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json'
    }
)

data = response.json()
```

### curl Example

```bash
curl -X GET http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

## Complete Authentication Flow Examples

### JavaScript/TypeScript

```typescript
class RTXAuthClient {
  private token: string | null = null;
  private baseURL = 'http://localhost:7999';

  async login(username: string, password: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    if (!response.ok) {
      throw new Error('Login failed');
    }

    const data = await response.json();
    this.token = data.token;

    // Store token securely
    localStorage.setItem('rtx_token', this.token);
  }

  async request(endpoint: string, options: RequestInit = {}): Promise<Response> {
    if (!this.token) {
      throw new Error('Not authenticated');
    }

    const headers = {
      'Authorization': `Bearer ${this.token}`,
      'Content-Type': 'application/json',
      ...options.headers
    };

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers
    });

    // Handle token expiration
    if (response.status === 401) {
      this.token = null;
      localStorage.removeItem('rtx_token');
      throw new Error('Token expired, please login again');
    }

    return response;
  }

  isAuthenticated(): boolean {
    return this.token !== null;
  }

  logout(): void {
    this.token = null;
    localStorage.removeItem('rtx_token');
  }

  // Restore token from storage
  restoreToken(): boolean {
    const stored = localStorage.getItem('rtx_token');
    if (stored) {
      this.token = stored;
      return true;
    }
    return false;
  }
}

// Usage
const client = new RTXAuthClient();

// Login
await client.login('admin', 'password');

// Make authenticated requests
const response = await client.request('/api/account/summary');
const account = await response.json();

// Logout
client.logout();
```

### Python

```python
import requests
from typing import Optional, Dict, Any
import json

class RTXAuthClient:
    def __init__(self, base_url: str = 'http://localhost:7999'):
        self.base_url = base_url
        self.token: Optional[str] = None
        self.user: Optional[Dict[str, Any]] = None

    def login(self, username: str, password: str) -> Dict[str, Any]:
        response = requests.post(
            f'{self.base_url}/login',
            json={'username': username, 'password': password},
            timeout=10
        )

        if response.status_code != 200:
            raise Exception('Login failed')

        data = response.json()
        self.token = data['token']
        self.user = data['user']

        return data

    def request(
        self,
        method: str,
        endpoint: str,
        **kwargs
    ) -> requests.Response:
        if not self.token:
            raise Exception('Not authenticated')

        headers = kwargs.pop('headers', {})
        headers['Authorization'] = f'Bearer {self.token}'

        response = requests.request(
            method,
            f'{self.base_url}{endpoint}',
            headers=headers,
            **kwargs
        )

        # Handle token expiration
        if response.status_code == 401:
            self.token = None
            self.user = None
            raise Exception('Token expired, please login again')

        response.raise_for_status()
        return response

    def get(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('GET', endpoint, **kwargs)

    def post(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('POST', endpoint, **kwargs)

    def is_authenticated(self) -> bool:
        return self.token is not None

    def logout(self):
        self.token = None
        self.user = None

# Usage
client = RTXAuthClient()

# Login
client.login('admin', 'password')

# Make authenticated requests
response = client.get('/api/account/summary')
account = response.json()

# Place order
response = client.post('/api/orders/market', json={
    'symbol': 'EURUSD',
    'side': 'BUY',
    'volume': 0.1
})

# Logout
client.logout()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type RTXAuthClient struct {
    BaseURL string
    Token   string
    User    *User
    Client  *http.Client
}

type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

func NewRTXAuthClient(baseURL string) *RTXAuthClient {
    return &RTXAuthClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *RTXAuthClient) Login(username, password string) error {
    reqBody := LoginRequest{
        Username: username,
        Password: password,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return err
    }

    resp, err := c.Client.Post(
        c.BaseURL+"/login",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("login failed: %d", resp.StatusCode)
    }

    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return err
    }

    c.Token = loginResp.Token
    c.User = &loginResp.User

    return nil
}

func (c *RTXAuthClient) Request(method, endpoint string, body io.Reader) (*http.Response, error) {
    if c.Token == "" {
        return nil, fmt.Errorf("not authenticated")
    }

    req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }

    // Handle token expiration
    if resp.StatusCode == http.StatusUnauthorized {
        c.Token = ""
        c.User = nil
        return resp, fmt.Errorf("token expired, please login again")
    }

    return resp, nil
}

func (c *RTXAuthClient) IsAuthenticated() bool {
    return c.Token != ""
}

func (c *RTXAuthClient) Logout() {
    c.Token = ""
    c.User = nil
}

// Usage
func main() {
    client := NewRTXAuthClient("http://localhost:7999")

    // Login
    if err := client.Login("admin", "password"); err != nil {
        panic(err)
    }

    // Make authenticated request
    resp, err := client.Request("GET", "/api/account/summary", nil)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // Logout
    client.Logout()
}
```

## Token Validation

The server validates tokens on every request:

1. **Extract token** from Authorization header
2. **Verify signature** using secret key
3. **Check expiration** (`exp` claim)
4. **Extract user info** (id, username, role)

**Invalid token responses:**

```http
HTTP/1.1 401 Unauthorized
Content-Type: text/plain

Unauthorized
```

## Security Best Practices

### 1. Store Tokens Securely

**Browser:**
```javascript
// ✅ Good: Use localStorage or sessionStorage
localStorage.setItem('rtx_token', token);

// ❌ Bad: Don't store in cookies without HttpOnly/Secure flags
document.cookie = `token=${token}`; // Vulnerable to XSS
```

**Mobile Apps:**
```javascript
// Use secure storage
import * as SecureStore from 'expo-secure-store';

await SecureStore.setItemAsync('rtx_token', token);
const token = await SecureStore.getItemAsync('rtx_token');
```

### 2. Handle Token Expiration

```javascript
async function requestWithRetry(endpoint, options) {
  try {
    return await client.request(endpoint, options);
  } catch (error) {
    if (error.message.includes('Token expired')) {
      // Redirect to login
      window.location.href = '/login';
    }
    throw error;
  }
}
```

### 3. Secure Transport (HTTPS)

**Always use HTTPS in production:**

```javascript
// ✅ Production
const baseURL = 'https://api.rtxtrading.com';

// ❌ Development only
const baseURL = 'http://localhost:7999';
```

### 4. Password Security

**Never log or store passwords:**

```javascript
// ❌ Bad
console.log('Password:', password);
localStorage.setItem('password', password);

// ✅ Good
// Only send password during login, never store
```

### 5. Token Rotation (Future)

**Planned:** Refresh token mechanism

```javascript
// Future implementation
async function refreshToken(refreshToken) {
  const response = await fetch('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refreshToken })
  });

  const { token } = await response.json();
  return token;
}
```

## CORS Configuration

The server allows cross-origin requests:

```javascript
// Server-side (already configured)
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
```

**Preflight requests:**

```http
OPTIONS /api/orders/market HTTP/1.1
Host: localhost:7999
Origin: http://localhost:3000
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Authorization, Content-Type
```

## Testing Authentication

### Using curl

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.token')

echo "Token: $TOKEN"

# Use token
curl -X GET http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer $TOKEN"
```

### Using Postman

1. **Login:**
   - Method: POST
   - URL: `http://localhost:7999/login`
   - Body (JSON):
     ```json
     {
       "username": "admin",
       "password": "password"
     }
     ```
   - Extract token from response

2. **Set Authorization:**
   - Go to Authorization tab
   - Type: Bearer Token
   - Token: Paste JWT token

3. **Make Requests:**
   - All subsequent requests will include the token

## Troubleshooting

### 401 Unauthorized

**Problem:** All requests return 401

**Causes:**
1. Token not included in header
2. Token expired
3. Invalid token format
4. Incorrect Bearer scheme

**Solution:**
```javascript
// ✅ Correct
headers: {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
}

// ❌ Wrong
headers: {
  'Authorization': 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' // Missing 'Bearer '
}
```

### Login Always Fails

**Problem:** Cannot login even with correct credentials

**Causes:**
1. Incorrect username/password
2. Account disabled
3. Server not running

**Solution:**
1. Verify credentials (default: admin/password)
2. Check server logs
3. Ensure server is running on port 7999

## Future Enhancements

- ✅ **Current:** JWT authentication
- 🔄 **Planned:** Refresh tokens
- 🔄 **Planned:** Two-factor authentication (2FA)
- 🔄 **Planned:** OAuth2 support
- 🔄 **Planned:** API keys for programmatic access
- 🔄 **Planned:** Session management (concurrent login limits)

## Summary

| Feature | Status | Description |
|---------|--------|-------------|
| JWT Auth | ✅ Live | Token-based authentication |
| Admin Account | ✅ Live | Full system access |
| Trader Account | ✅ Live | Trading and account access |
| Token Expiration | ✅ Live | 24-hour lifetime |
| Refresh Tokens | 🔄 Planned | Automatic token refresh |
| 2FA | 🔄 Planned | Two-factor authentication |
| OAuth2 | 🔄 Planned | Third-party auth |
| API Keys | 🔄 Planned | Programmatic access |
