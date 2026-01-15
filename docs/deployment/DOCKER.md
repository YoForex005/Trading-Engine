# Docker Deployment Guide

This guide covers building, running, and publishing Docker images for the Trading Engine platform.

## Overview

The platform uses multi-stage Docker builds to create minimal, production-ready container images:

- **Backend**: Go trading engine (~15-20MB final image)
- **Frontend**: React/Vite UI with nginx (~40-60MB final image)

Both images are optimized for production deployment with security hardening and minimal attack surface.

## Building Images

### Backend Image

Build the Go trading engine image:

```bash
cd backend
docker build -t trading-engine-backend:latest .
```

**Build Details:**
- Base: `golang:1.24-alpine` (build stage) → `gcr.io/distroless/static:nonroot` (runtime)
- Final size: ~15-20MB
- Security: Non-root user, no shell, distroless base
- Features: Static binary with optimizations, includes data/ and db/ directories

**Build Flags:**
- `CGO_ENABLED=0`: Pure Go binary (no C dependencies)
- `-trimpath`: Removes file system paths for security
- `-ldflags="-s -w"`: Strips debug info and symbol table

### Frontend Image

Build the React frontend image:

```bash
cd clients/desktop
docker build -t trading-engine-frontend:latest .
```

**Build Details:**
- Base: `node:20-alpine` (build stage) → `nginx:1.25-alpine` (runtime)
- Final size: ~40-60MB
- Security: Nginx production configuration with security headers
- Features: Bun package manager, Vite optimization, SPA routing

## Running Containers

### Backend

Run the trading engine backend:

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL=postgresql://user:pass@host:5432/trading_engine \
  -e REDIS_URL=redis:6379 \
  -e JWT_SECRET=your-secure-secret \
  -e OANDA_API_KEY=your-api-key \
  -e OANDA_ACCOUNT_ID=your-account-id \
  trading-engine-backend:latest
```

**Required Environment Variables:**
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: Secret key for JWT token signing
- `OANDA_API_KEY`: OANDA API key for market data
- `OANDA_ACCOUNT_ID`: OANDA account ID

**Optional Environment Variables:**
- `LOG_LEVEL`: Logging level (debug, info, warn, error) - default: info
- `PORT`: HTTP port - default: 8080

### Frontend

Run the React frontend:

```bash
docker run -p 80:80 trading-engine-frontend:latest
```

The frontend is pre-built with the API URL configured at build time. To customize:

```bash
docker build \
  --build-arg VITE_API_URL=http://your-backend:8080 \
  -t trading-engine-frontend:latest \
  clients/desktop
```

## Registry Publishing

### GitHub Container Registry (GHCR)

The CI/CD pipeline automatically publishes images to GitHub Container Registry on the main branch.

**Manual Publishing:**

1. **Login to GHCR:**
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

2. **Tag images:**
```bash
# Backend
docker tag trading-engine-backend:latest ghcr.io/YOUR_ORG/trading-engine-backend:v1.0.0
docker tag trading-engine-backend:latest ghcr.io/YOUR_ORG/trading-engine-backend:latest

# Frontend
docker tag trading-engine-frontend:latest ghcr.io/YOUR_ORG/trading-engine-frontend:v1.0.0
docker tag trading-engine-frontend:latest ghcr.io/YOUR_ORG/trading-engine-frontend:latest
```

3. **Push to registry:**
```bash
docker push ghcr.io/YOUR_ORG/trading-engine-backend:v1.0.0
docker push ghcr.io/YOUR_ORG/trading-engine-backend:latest
docker push ghcr.io/YOUR_ORG/trading-engine-frontend:v1.0.0
docker push ghcr.io/YOUR_ORG/trading-engine-frontend:latest
```

### Alternative Registries

For Docker Hub or other registries:

```bash
# Docker Hub
docker tag trading-engine-backend:latest your-username/trading-engine-backend:v1.0.0
docker push your-username/trading-engine-backend:v1.0.0

# AWS ECR
docker tag trading-engine-backend:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/trading-engine-backend:v1.0.0
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/trading-engine-backend:v1.0.0
```

## Best Practices

### Version Tags

**DO:**
- Use semantic versioning: `v1.0.0`, `v1.2.3`
- Tag with git SHA: `sha-abc123`
- Use specific versions in production

**DON'T:**
- Use `:latest` in production deployments
- Overwrite existing version tags

### Resource Limits

Always set resource limits in production:

```bash
docker run \
  --memory=512m \
  --cpus=1.0 \
  -p 8080:8080 \
  trading-engine-backend:latest
```

Recommended limits:
- Backend: 512MB-1GB memory, 1-2 CPUs
- Frontend: 128MB-256MB memory, 0.5-1 CPU

### Health Checks

Use Docker health checks for orchestration:

```bash
docker run \
  --health-cmd="curl -f http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=5s \
  --health-retries=3 \
  -p 8080:8080 \
  trading-engine-backend:latest
```

### Networking

For multi-container deployments, use Docker networks:

```bash
# Create network
docker network create trading-network

# Run containers on network
docker run -d --name db --network trading-network postgres:16-alpine
docker run -d --name redis --network trading-network redis:7-alpine
docker run -d --name backend --network trading-network \
  -e DATABASE_URL=postgresql://postgres:password@db:5432/trading_engine \
  -e REDIS_URL=redis:6379 \
  trading-engine-backend:latest
```

## Security

### Image Security

The images include multiple security features:

1. **Distroless Base (Backend):**
   - No shell access
   - Minimal attack surface (~2MB base image)
   - No package manager or unnecessary tools
   - Security updates via base image rebuilds

2. **Non-Root User:**
   - Both images run as non-root user
   - Backend: `nonroot` user from distroless
   - Frontend: nginx runs as `nginx` user

3. **No Secrets in Images:**
   - All secrets passed via environment variables
   - Never bake credentials into images
   - Use Docker secrets or environment injection

### Secret Management

**DO:**
```bash
# Pass secrets via environment variables
docker run -e JWT_SECRET=$JWT_SECRET ...

# Or use Docker secrets (Swarm/Kubernetes)
docker secret create jwt_secret jwt_secret.txt
docker run --secret jwt_secret ...
```

**DON'T:**
```bash
# Never hardcode secrets in Dockerfile
ENV JWT_SECRET=hardcoded-secret  # BAD!

# Never build secrets into image
COPY secret.key /app/  # BAD!
```

### Security Headers (Frontend)

The nginx configuration includes security headers:

- `X-Frame-Options: SAMEORIGIN` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - Enables XSS protection

## Troubleshooting

### Backend Won't Start

**Check logs:**
```bash
docker logs container-name
```

**Common issues:**
- Database connection failed: Verify `DATABASE_URL` is correct
- Redis connection failed: Verify `REDIS_URL` is accessible
- Missing environment variables: Check all required env vars are set

### Frontend 502 Bad Gateway

**Check nginx status:**
```bash
docker exec container-name nginx -t  # Test nginx config
docker logs container-name  # Check nginx logs
```

**Common issues:**
- Backend not accessible: Ensure backend is running and accessible
- Incorrect API URL: Rebuild with correct `VITE_API_URL`

### Image Too Large

**Check image size:**
```bash
docker images | grep trading-engine
```

**Optimize build:**
- Ensure multi-stage build is working correctly
- Check for unnecessary files in build context
- Use `.dockerignore` to exclude unnecessary files

### Build Cache Issues

**Clear build cache:**
```bash
docker builder prune -a
```

**Force rebuild without cache:**
```bash
docker build --no-cache -t trading-engine-backend:latest .
```

## Performance Optimization

### Build Time

Optimize build time with BuildKit:

```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Build with cache
docker build \
  --cache-from=ghcr.io/YOUR_ORG/trading-engine-backend:latest \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  -t trading-engine-backend:latest .
```

### Layer Caching

The Dockerfiles are optimized for layer caching:

1. Dependencies copied first (go.mod, package.json)
2. Dependencies installed (cached layer)
3. Source code copied (changes frequently)
4. Application built

This means dependency changes trigger full rebuilds, but code changes only rebuild the final layers.

## Related Documentation

- [Local Development Guide](./LOCAL_DEV.md)
- [CI/CD Pipeline](./CI_CD.md)
- [Operations Runbook](./OPERATIONS.md)
- Backend Dockerfile: `backend/Dockerfile`
- Frontend Dockerfile: `clients/desktop/Dockerfile`
- Docker Compose: `docker-compose.yml`
