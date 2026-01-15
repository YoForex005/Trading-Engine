# CI/CD Pipeline Documentation

This guide covers the continuous integration and deployment pipeline for the Trading Engine platform.

## Overview

The platform uses GitHub Actions for automated testing, building, and publishing Docker images. The CI/CD pipeline is optimized for a monorepo structure with intelligent path filtering to minimize build times.

**Key Features:**
- Automated testing on every push and pull request
- Path-based filtering (only test affected services)
- Docker image building with BuildKit caching
- Automated publishing to GitHub Container Registry
- Scheduled database backups

## GitHub Actions Workflows

The platform has three main workflows:

### 1. Backend CI/CD Workflow

**File:** `.github/workflows/backend.yml`

**Triggers:**
- Push to any branch affecting backend code:
  - `backend/**`
  - `go.mod`
  - `go.sum`
- Pull requests to main

**Jobs:**

1. **Detect Changes**
   - Uses `dorny/paths-filter@v3` for monorepo optimization
   - Checks if backend files changed
   - Skips workflow if no backend changes

2. **Test**
   - Go version: 1.21
   - Runs unit tests with race detector: `go test -race ./...`
   - Requires: Changes detected
   - Duration: ~1-2 minutes

3. **Build**
   - Builds Docker image with multi-stage Dockerfile
   - Uses BuildKit for layer caching
   - Cache type: GitHub Actions cache (type=gha)
   - Cache mode: max (exports all layers)
   - Duration: ~2-5 minutes (with cache: 30-60 seconds)

4. **Publish** (main branch only)
   - Pushes to GitHub Container Registry (ghcr.io)
   - Tags: `latest`, `sha-{git-sha}`
   - Authentication: Automatic via GITHUB_TOKEN
   - Requires: Tests pass, on main branch

**Path Filtering:**
```yaml
paths-filter:
  backend:
    - 'backend/**'
    - 'go.mod'
    - 'go.sum'
```

This saves 50-70% CI time by skipping backend tests when only frontend changes.

### 2. Frontend CI/CD Workflow

**File:** `.github/workflows/frontend.yml`

**Triggers:**
- Push to any branch affecting frontend code:
  - `clients/desktop/**`
  - `package.json`
  - `bun.lock`
- Pull requests to main

**Jobs:**

1. **Detect Changes**
   - Path filtering for frontend files
   - Skips if no frontend changes

2. **Typecheck**
   - Bun runtime for faster dependency installation
   - Runs TypeScript type checking: `bun run typecheck`
   - Ensures no type errors
   - Duration: ~30-60 seconds

3. **Test**
   - Runs Vitest tests: `bun run test`
   - Generates coverage report
   - Requires: Typecheck passes
   - Duration: ~1-2 minutes

4. **Build**
   - Builds Docker image with nginx
   - Uses BuildKit caching
   - Production build with Vite
   - Duration: ~3-5 minutes (with cache: 30-60 seconds)

5. **Publish** (main branch only)
   - Pushes to GitHub Container Registry
   - Tags: `latest`, `sha-{git-sha}`
   - Authentication: Automatic via GITHUB_TOKEN

**Path Filtering:**
```yaml
paths-filter:
  frontend:
    - 'clients/desktop/**'
    - 'package.json'
    - 'bun.lock'
```

### 3. Database Backup Workflow

**File:** `.github/workflows/backup.yml`

**Triggers:**
- Scheduled: Every 6 hours (`cron: '0 */6 * * *'`)
- Manual trigger: `workflow_dispatch`

**Jobs:**

1. **Backup**
   - Runs PostgreSQL backup script
   - Uses `pg_dump` with custom format and gzip compression
   - Uploads backup as GitHub Actions artifact
   - Retention: 30 days
   - Duration: ~1-5 minutes (depends on database size)

**Manual Trigger:**
```bash
gh workflow run backup.yml
```

## Running Workflows

### Viewing Workflow Status

**GitHub UI:**
1. Navigate to repository
2. Click "Actions" tab
3. View recent workflow runs

**GitHub CLI:**
```bash
# List recent runs
gh run list

# List runs for specific workflow
gh run list --workflow=backend.yml

# Watch live logs
gh run watch

# View logs for specific run
gh run view RUN_ID --log
```

### Triggering Workflows

**Automatic Triggers:**
Workflows trigger automatically on push to any branch:

```bash
git push origin feature-branch
```

**Manual Triggers:**
```bash
# Trigger backend workflow
gh workflow run backend.yml

# Trigger frontend workflow
gh workflow run frontend.yml

# Trigger backup
gh workflow run backup.yml

# Trigger with specific branch
gh workflow run backend.yml --ref feature-branch
```

### Monitoring Workflow Progress

**Real-time logs:**
```bash
# Watch latest workflow run
gh run watch

# View specific job logs
gh run view RUN_ID --job=test --log

# Download logs
gh run view RUN_ID --log > workflow.log
```

## Path Filtering Details

### How Path Filtering Works

The workflows use `dorny/paths-filter@v3` to determine which services changed:

**Example workflow run:**
```
1. Developer pushes change to backend/internal/core/engine.go
2. Path filter detects backend change
3. Backend workflow runs (test → build → publish)
4. Frontend workflow skipped (no frontend changes)
5. Result: 50-70% faster CI time
```

**Benefits:**
- Faster feedback loops (only test what changed)
- Reduced CI minutes usage
- Lower resource consumption
- Parallel workflow execution for independent changes

### Path Filter Configuration

**Backend paths:**
- `backend/**` - All backend code
- `go.mod` - Go dependencies
- `go.sum` - Go dependency checksums

**Frontend paths:**
- `clients/desktop/**` - All frontend code
- `package.json` - npm/bun dependencies
- `bun.lock` - Dependency lock file

**Shared changes:**
If both backend and frontend files change in one commit, both workflows run in parallel.

## Docker Image Publishing

### GitHub Container Registry (GHCR)

**Registry URL:** `ghcr.io/YOUR_ORG/`

**Images:**
- `ghcr.io/YOUR_ORG/trading-engine-backend:latest`
- `ghcr.io/YOUR_ORG/trading-engine-backend:sha-abc123`
- `ghcr.io/YOUR_ORG/trading-engine-frontend:latest`
- `ghcr.io/YOUR_ORG/trading-engine-frontend:sha-abc123`

### Image Tagging Strategy

**`:latest` tag:**
- Updated on every push to main branch
- Represents the latest stable build
- Use for development/staging environments
- **Do not use in production** (use specific versions)

**`:sha-{git-sha}` tag:**
- Unique tag for every commit
- Immutable (never overwritten)
- Allows rollback to any previous build
- **Recommended for production** (traceable to exact commit)

**Future: Semantic versioning:**
```yaml
# On git tag push
on:
  push:
    tags:
      - 'v*'

# Tag as v1.0.0
docker tag ... ghcr.io/YOUR_ORG/trading-engine-backend:v1.0.0
```

### Pulling Images

**Authentication:**
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Pull images:**
```bash
# Latest
docker pull ghcr.io/YOUR_ORG/trading-engine-backend:latest

# Specific commit
docker pull ghcr.io/YOUR_ORG/trading-engine-backend:sha-abc123

# Run pulled image
docker run ghcr.io/YOUR_ORG/trading-engine-backend:latest
```

### Image Visibility

**Public vs Private:**
- By default, images are private to repository
- Make public: Repository Settings → Packages → Change visibility

**Access control:**
- Team members: Automatic access via GitHub permissions
- CI/CD: Use GITHUB_TOKEN (automatic)
- External: Requires personal access token or GitHub App token

## BuildKit Caching

### How BuildKit Caching Works

BuildKit caches Docker layers between builds:

**Cache sources:**
1. **GitHub Actions cache** (`type=gha`)
   - Stores layers in GitHub Actions cache
   - Shared across workflow runs
   - 10GB limit per repository
   - 7-day retention for unused cache

2. **Registry cache** (alternative: `type=registry`)
   - Stores layers in Docker registry
   - Larger storage capacity
   - Requires registry push permissions

**Cache configuration:**
```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

**Cache modes:**
- `mode=min`: Only cache final image layers (faster upload)
- `mode=max`: Cache all layers including build stages (faster builds)

**Performance impact:**
- First build: ~2-5 minutes (no cache)
- Subsequent builds: ~30-60 seconds (cache hit)
- Code-only changes: ~10-30 seconds (dependency cache hit)

### Cache Invalidation

**Cache invalidates when:**
- Dependencies change (go.mod, package.json)
- Dockerfile modified
- Base image updated
- 7 days of inactivity

**Manual cache clear:**
```bash
# Clear all caches
gh cache delete --all

# Clear specific cache
gh cache list
gh cache delete CACHE_ID
```

## Secrets Configuration

### Required Secrets

**For image publishing (automatic):**
- `GITHUB_TOKEN`: Provided automatically by GitHub Actions
- Permissions: Write packages

**For database backups:**
- `DB_PASSWORD`: PostgreSQL password for backup script
  - Add in: Repository Settings → Secrets and variables → Actions → New repository secret

### Adding Secrets

**Via GitHub UI:**
1. Navigate to repository
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `DB_PASSWORD`
5. Value: Your database password
6. Click "Add secret"

**Via GitHub CLI:**
```bash
gh secret set DB_PASSWORD
# Paste password when prompted
```

**Using secrets in workflows:**
```yaml
env:
  DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
```

## Best Practices

### Workflow Optimization

**1. Cache dependencies:**
```yaml
- uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

**2. Run tests in parallel:**
```yaml
strategy:
  matrix:
    go-version: [1.21, 1.22]
  parallel: 2
```

**3. Fail fast:**
```yaml
strategy:
  fail-fast: true
```

**4. Use specific action versions:**
```yaml
# Good: Pinned to specific version
- uses: actions/checkout@v4

# Bad: Using latest (unpredictable)
- uses: actions/checkout@latest
```

### Testing Strategy

**Run on every push:**
- Unit tests (fast, comprehensive)
- Type checking (TypeScript)
- Linting (code quality)

**Run on main branch:**
- Integration tests (slower)
- End-to-end tests (slowest)
- Security scans

**Run on release:**
- Performance tests
- Load tests
- Full regression suite

### Security Best Practices

**1. Pin action versions:**
```yaml
uses: docker/build-push-action@v5  # Specific version
```

**2. Review permissions:**
```yaml
permissions:
  contents: read
  packages: write
```

**3. Scan images:**
```yaml
- name: Scan image
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ghcr.io/YOUR_ORG/trading-engine-backend:latest
```

**4. Never commit secrets:**
- Use GitHub Secrets
- Use environment variables
- Never hardcode credentials

## Local Testing

### Testing Workflows Locally

Use `act` to run GitHub Actions locally:

**Install act:**
```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

**Run workflow locally:**
```bash
# Run backend tests
act -j test

# Run with secrets
act -j test -s DB_PASSWORD=test

# Run specific workflow
act -W .github/workflows/backend.yml

# Dry run (list jobs)
act -l
```

**Limitations:**
- Some features not supported (GitHub cache, GHCR push)
- Docker-in-Docker required
- Slower than GitHub-hosted runners

## Troubleshooting

### Workflow Fails to Trigger

**Check trigger conditions:**
```yaml
on:
  push:
    paths:
      - 'backend/**'
```

Ensure your changes match the path filter.

**Manual trigger:**
```bash
gh workflow run backend.yml
```

### Tests Fail in CI but Pass Locally

**Common causes:**
1. Missing environment variables
2. Different Go/Node versions
3. Timing issues (tests assume fast execution)
4. Missing dependencies

**Solutions:**
1. Add required secrets in repository settings
2. Pin versions in workflow (e.g., `go-version: 1.21`)
3. Increase test timeouts
4. Add dependency installation step

### Docker Build Fails

**Check build logs:**
```bash
gh run view RUN_ID --job=build --log
```

**Common issues:**
1. Dockerfile syntax error
2. Missing dependencies
3. Cache corruption
4. Out of disk space

**Solutions:**
```bash
# Clear cache
gh cache delete --all

# Rebuild without cache
docker build --no-cache ...
```

### Image Push Fails

**Check permissions:**
```yaml
permissions:
  packages: write  # Required for GHCR push
```

**Verify authentication:**
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Check package visibility:**
Repository Settings → Packages → Verify permissions

### Backup Workflow Fails

**Check secret configuration:**
```bash
gh secret list
```

Ensure `DB_PASSWORD` is set.

**Test backup manually:**
```bash
docker-compose exec backend bash /app/scripts/backup-db.sh
```

## Deployment Pipeline (Future)

When ready for production, add deployment jobs:

### Staging Deployment

```yaml
deploy-staging:
  runs-on: ubuntu-latest
  needs: [test, build]
  if: github.ref == 'refs/heads/main'
  steps:
    - name: Deploy to staging
      run: |
        kubectl set image deployment/backend \
          backend=ghcr.io/YOUR_ORG/trading-engine-backend:sha-${{ github.sha }}
```

### Production Deployment

```yaml
deploy-production:
  runs-on: ubuntu-latest
  needs: [deploy-staging]
  environment: production  # Requires manual approval
  steps:
    - name: Deploy to production
      run: |
        kubectl set image deployment/backend \
          backend=ghcr.io/YOUR_ORG/trading-engine-backend:sha-${{ github.sha }}
```

**Features:**
- Manual approval gate for production
- Automated rollback on failure
- Health check validation
- Blue-green or rolling deployment

## Metrics and Monitoring

### Workflow Metrics

**GitHub Insights:**
- Actions → Usage
- View CI minutes usage
- Track workflow run times
- Monitor success/failure rates

**Recommended tracking:**
- Build time trends
- Test duration
- Cache hit rate
- Deployment frequency
- Mean time to recovery (MTTR)

### Optimization Targets

**CI Pipeline:**
- Unit tests: <2 minutes
- Build time: <5 minutes (cold), <1 minute (cached)
- Total pipeline: <10 minutes
- Cache hit rate: >80%

**Deployment:**
- Staging deployment: <5 minutes
- Production deployment: <10 minutes (with health checks)
- Rollback time: <2 minutes

## Related Documentation

- [Docker Deployment Guide](./DOCKER.md)
- [Local Development Guide](./LOCAL_DEV.md)
- [Operations Runbook](./OPERATIONS.md)
- [Monitoring Guide](./MONITORING.md)
- Backend Workflow: `.github/workflows/backend.yml`
- Frontend Workflow: `.github/workflows/frontend.yml`
- Backup Workflow: `.github/workflows/backup.yml`
