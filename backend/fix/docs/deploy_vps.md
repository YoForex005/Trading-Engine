# Deploy Trading Engine on VPS

## Why VPS?
Your ISP is blocking port 12336. A VPS (cloud server) doesn't have this restriction and provides:
- 24/7 uptime for trading
- Low latency to T4B servers
- No ISP port blocking
- Professional production setup

## Quick Setup Options

### Option A: DigitalOcean (Recommended - $4/month)

1. Sign up: https://www.digitalocean.com/
2. Create Droplet:
   - Image: Ubuntu 22.04
   - Plan: Basic $4/month (1GB RAM)
   - Region: Choose closest to T4B server (likely EU/UK)
   - Add your SSH key

3. Connect to your droplet:
   ```bash
   ssh root@YOUR_DROPLET_IP
   ```

4. Install Go and deploy:
   ```bash
   # Install Go
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin

   # Clone your trading engine (or scp the files)
   mkdir -p /opt/trading-engine
   cd /opt/trading-engine

   # Copy your backend files here
   # Then build and run:
   go build -o server ./cmd/server/
   ./server
   ```

### Option B: Vultr ($5/month)
- Similar to DigitalOcean
- https://www.vultr.com/

### Option C: AWS Lightsail ($3.50/month)
- https://aws.amazon.com/lightsail/
- Includes 3 months free

## Test Connectivity First

Before deploying, SSH into the VPS and test:

```bash
# Test if port 12336 is reachable from VPS
nc -zv 23.106.238.138 12336

# Should show: Connection to 23.106.238.138 12336 port [tcp/*] succeeded!
```

## Alternative: SSH Tunnel (If you already have a VPS)

If you have ANY server that can reach port 12336, create a tunnel:

```bash
# On your local Mac:
ssh -L 12336:23.106.238.138:12336 user@your-vps-ip -N

# Then in gateway.go, change Host to:
# Host: "127.0.0.1"  (localhost, tunneled through SSH)
```

This forwards local port 12336 through your VPS to T4B.
