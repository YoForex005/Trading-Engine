# Setup and Installation Guide

This guide will help you set up the RTX Trading Engine development environment.

## Prerequisites

### Required Software
- **Go**: Version 1.19 or higher
- **Git**: For version control
- **Text Editor/IDE**: VS Code, GoLand, or your preferred editor
- **curl**: For API testing (included on most systems)

### Optional Tools
- **Postman**: For API testing
- **WebSocket Client**: For testing real-time feeds
- **Docker**: For containerized deployment (future)

## System Requirements

### Minimum
- **OS**: Linux, macOS, or Windows
- **RAM**: 2 GB
- **CPU**: 1 core
- **Disk**: 10 GB free space

### Recommended
- **OS**: Linux (Ubuntu 20.04+) or macOS
- **RAM**: 4 GB or more
- **CPU**: 2+ cores
- **Disk**: 50 GB free space (SSD preferred)

## Installation Steps

### 1. Install Go

#### macOS (via Homebrew)
```bash
brew install go
```

#### Linux (Ubuntu/Debian)
```bash
# Download and install
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Windows
Download and run the installer from https://go.dev/dl/

#### Verify Installation
```bash
go version
# Should output: go version go1.21.0 ...
```

### 2. Clone the Repository

```bash
# Clone the project
git clone <repository-url>
cd trading-engine

# Navigate to backend
cd backend
```

### 3. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify
```

### 4. Configure Environment

#### Create Configuration Files

**4.1. LP Configuration** (`backend/data/lp_config.json`)
```json
{
  "version": "1.0",
  "lps": [
    {
      "id": "binance",
      "name": "Binance",
      "type": "CRYPTO",
      "enabled": true,
      "wsUrl": "wss://stream.binance.com:9443/ws",
      "symbols": ["BTCUSD", "ETHUSD", "BNBUSD", "SOLUSD", "XRPUSD"]
    },
    {
      "id": "oanda",
      "name": "OANDA",
      "type": "FOREX",
      "enabled": false,
      "apiKey": "YOUR_OANDA_API_KEY",
      "accountId": "YOUR_OANDA_ACCOUNT_ID"
    }
  ],
  "lastModified": 0
}
```

**4.2. FIX Session Configuration** (`backend/fix/config/yofx1.cfg`)
```ini
[DEFAULT]
ConnectionType=initiator
ReconnectInterval=30
FileStorePath=backend/fixstore
FileLogPath=backend/fixlogs
StartTime=00:00:00
EndTime=23:59:59
UseDataDictionary=Y
DataDictionary=backend/fix/config/FIX44.xml
ResetOnLogon=Y
ResetOnLogout=Y
ResetOnDisconnect=Y

[SESSION]
BeginString=FIX.4.4
SenderCompID=YOUR_SENDER_ID
TargetCompID=YOFX
HeartBtInt=30
SocketConnectHost=fix.yofx.com
SocketConnectPort=1234
```

#### Set Environment Variables (Optional)

```bash
# Create .env file
cat > .env << EOF
# Server Configuration
SERVER_PORT=7999

# Execution Mode (BBOOK or ABOOK)
EXECUTION_MODE=BBOOK

# Broker Configuration
BROKER_NAME=RTX Trading
DEFAULT_LEVERAGE=100
DEFAULT_BALANCE=5000.00

# OANDA Configuration (if using)
OANDA_API_KEY=your-api-key-here
OANDA_ACCOUNT_ID=your-account-id

# Logging
LOG_LEVEL=info
EOF
```

### 5. Directory Structure Setup

```bash
# Create required directories
mkdir -p backend/data/ohlc
mkdir -p backend/data/ticks
mkdir -p backend/fixstore
mkdir -p backend/fixlogs

# Set permissions (Linux/macOS)
chmod 755 backend/data
chmod 755 backend/fixstore
chmod 755 backend/fixlogs
```

### 6. Build the Application

```bash
# Build binary
cd backend
go build -o server cmd/server/main.go

# Verify build
./server --version  # (if version flag is implemented)
ls -lh server       # Check binary size
```

### 7. Run the Server

```bash
# Run directly
go run cmd/server/main.go

# Or run the built binary
./server
```

Expected output:
```
╔═══════════════════════════════════════════════════════════╗
║          RTX Trading - Backend v3.0                ║
║        BBOOK Mode + OANDA LP                 ║
╚═══════════════════════════════════════════════════════════╝
[B-Book] Demo account created: RTX-000001 with $5000.00
[Hub] Pipeline check: 1000 ticks received...
═══════════════════════════════════════════════════════════
  SERVER READY - B-BOOK TRADING ENGINE
═══════════════════════════════════════════════════════════
  HTTP API:    http://localhost:7999
  WebSocket:   ws://localhost:7999/ws
```

### 8. Verify Installation

#### Test Health Endpoint
```bash
curl http://localhost:7999/health
# Expected: OK
```

#### Test Login
```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo-user","password":"password"}'
```

Expected response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "demo-user",
    "role": "USER",
    "accountId": 1
  }
}
```

#### Test WebSocket Connection
```bash
# Install wscat (if not already installed)
npm install -g wscat

# Connect to WebSocket
wscat -c ws://localhost:7999/ws

# You should see real-time price updates
```

## Development Workflow

### 1. Running with Hot Reload

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

Run with hot reload:
```bash
cd backend
air
```

### 2. Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/core/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Code Formatting

```bash
# Format all Go files
go fmt ./...

# Run linter (install golangci-lint first)
golangci-lint run

# Fix linting issues automatically
golangci-lint run --fix
```

### 4. Building for Production

```bash
# Build with optimizations
go build -ldflags="-s -w" -o server cmd/server/main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o server-linux cmd/server/main.go
GOOS=darwin GOARCH=amd64 go build -o server-mac cmd/server/main.go
GOOS=windows GOARCH=amd64 go build -o server.exe cmd/server/main.go
```

## IDE Setup

### VS Code

#### Recommended Extensions
```bash
# Install Go extension
code --install-extension golang.go

# Install REST Client
code --install-extension humao.rest-client

# Install YAML support
code --install-extension redhat.vscode-yaml
```

#### Settings (`.vscode/settings.json`)
```json
{
  "go.useLanguageServer": true,
  "go.toolsManagement.autoUpdate": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

#### Launch Configuration (`.vscode/launch.json`)
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/backend/cmd/server",
      "env": {},
      "args": []
    }
  ]
}
```

### GoLand / IntelliJ IDEA

1. Open project directory
2. Go to **File → Settings → Go → GOROOT**
3. Set GOROOT to your Go installation
4. Go to **Run → Edit Configurations**
5. Add new **Go Build** configuration:
   - **Name**: Run Server
   - **Run kind**: File
   - **Files**: `cmd/server/main.go`
   - **Working directory**: `$ProjectFileDir$/backend`

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 7999
lsof -i :7999

# Kill the process
kill -9 <PID>

# Or change port in main.go
# Line: http.ListenAndServe(":8000", nil)
```

### Module Errors
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download

# Update dependencies
go get -u ./...
go mod tidy
```

### Permission Denied (Linux/macOS)
```bash
# Give execute permission
chmod +x server

# Or run with sudo (not recommended)
sudo ./server
```

### OANDA Connection Failed
- Verify API key and account ID
- Check network connectivity
- Ensure API key has proper permissions
- Review OANDA API documentation

### FIX Connection Issues
- Verify SenderCompID and TargetCompID
- Check host and port
- Ensure FIX session is approved by provider
- Review FIX logs in `backend/fixlogs/`

## Next Steps

1. Read the [Code Organization](code-organization.md) guide
2. Explore the [API Documentation](../api/endpoints.md)
3. Review [Trading Concepts](../concepts/execution-models.md)
4. Check the [Testing Guide](testing.md)

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Go Modules Reference](https://go.dev/ref/mod)
- [WebSocket RFC](https://datatracker.ietf.org/doc/html/rfc6455)
- [FIX Protocol](https://www.fixtrading.org/)

## Support

If you encounter issues:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review logs in `backend/logs/`
3. Search existing GitHub issues
4. Create a new issue with:
   - OS and Go version
   - Error messages
   - Steps to reproduce
