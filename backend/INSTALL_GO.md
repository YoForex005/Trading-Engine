# Installing Go on WSL2 Ubuntu

## Quick Install (Copy & Paste)

```bash
# Download Go 1.21.6 (or check https://go.dev/dl/ for latest)
cd ~
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# Remove old installation (if exists)
sudo rm -rf /usr/local/go

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add to PATH permanently
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc

# Reload shell configuration
source ~/.bashrc

# Verify installation
go version

# Should output: go version go1.21.6 linux/amd64
```

## Verify Installation

```bash
# Check Go version
go version

# Check Go environment
go env

# Important paths:
# GOROOT: /usr/local/go (where Go is installed)
# GOPATH: ~/go (where packages are downloaded)
# GOBIN: ~/go/bin (where installed binaries go)
```

## Set Up Go Workspace

```bash
# Create Go workspace directories
mkdir -p ~/go/{bin,pkg,src}

# Verify GOPATH
go env GOPATH
# Should show: /home/your-username/go
```

## Install Essential Go Tools

```bash
# Static analysis tool
go install honnef.co/go/tools/cmd/staticcheck@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Code formatter
go install golang.org/x/tools/cmd/goimports@latest

# Verify tools installed
which staticcheck
which golangci-lint
which goimports
```

## Configure VS Code (if using)

Install Go extension:
```
code --install-extension golang.go
```

VS Code `settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"]
}
```

## Test Go Installation

```bash
# Create test program
cat > /tmp/hello.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, RTX5!")
}
EOF

# Compile and run
cd /tmp
go run hello.go

# Should output: Hello, RTX5!

# Clean up
rm /tmp/hello.go
```

## Initialize RTX5 Backend

```bash
cd /mnt/d/Tading\ engine/Trading-Engine/backend

# Download all dependencies
go mod download

# Verify module integrity
go mod verify

# Build the server
go build -o bin/server ./cmd/server/

# Should complete without errors
```

## Common Issues

### Issue: "command not found: go"

**Fix**: PATH not updated. Run:
```bash
source ~/.bashrc
# or restart terminal
```

### Issue: "permission denied" when extracting

**Fix**: Use sudo:
```bash
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
```

### Issue: "go: cannot find main module"

**Fix**: You're not in a Go module directory. Run:
```bash
cd /mnt/d/Tading\ engine/Trading-Engine/backend
# This directory contains go.mod
```

### Issue: Old Go version

**Fix**: Completely remove old version first:
```bash
sudo rm -rf /usr/local/go
# Then reinstall latest version
```

## Version Management (Optional)

If you need multiple Go versions, use a version manager:

### Option 1: goenv
```bash
# Install goenv
git clone https://github.com/syndbg/goenv.git ~/.goenv

# Add to PATH
echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(goenv init -)"' >> ~/.bashrc
source ~/.bashrc

# Install specific version
goenv install 1.21.6
goenv global 1.21.6
```

### Option 2: Manual
```bash
# Download multiple versions to different directories
sudo tar -C /opt -xzf go1.21.6.linux-amd64.tar.gz
sudo mv /opt/go /opt/go1.21.6

# Switch versions by changing PATH
export PATH=/opt/go1.21.6/bin:$PATH
```

## Upgrade Go

```bash
# Download new version
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz

# Remove old
sudo rm -rf /usr/local/go

# Install new
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Verify
go version
```

## Performance Tuning (Optional)

Add to `~/.bashrc`:
```bash
# Go build cache location
export GOCACHE="$HOME/.cache/go-build"

# Enable Go modules (default in Go 1.16+)
export GO111MODULE=on

# Proxy for faster downloads (optional)
export GOPROXY=https://proxy.golang.org,direct

# Private module access (if needed)
export GOPRIVATE=github.com/your-org/*
```

## Verify Backend Build

```bash
cd /mnt/d/Tading\ engine/Trading-Engine/backend

# Run build verification
./scripts/verify_build.sh

# Expected output:
# ✓ Go installation
# ✓ Go module integrity
# ✓ Build cmd/server
# ✓ go vet analysis
# ... etc
```

## Uninstall Go

```bash
# Remove Go installation
sudo rm -rf /usr/local/go

# Remove from PATH (edit ~/.bashrc)
# Delete lines containing:
# export PATH=$PATH:/usr/local/go/bin

# Reload shell
source ~/.bashrc

# Clean workspace (optional)
rm -rf ~/go
```

---

**Next Steps**:
1. Run `./scripts/verify_build.sh` to check backend health
2. Review `QUICK_FIX_GUIDE.md` for build fixes
3. Read `BUILD_QUALITY_REPORT.md` for detailed analysis
