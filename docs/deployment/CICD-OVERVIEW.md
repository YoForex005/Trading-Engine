# CI/CD Pipeline Overview - Trading Engine

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         GitHub Repository                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │   main     │  │  develop   │  │  staging   │  │  feature/* │   │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘   │
└────────┼────────────────┼────────────────┼────────────────┼──────────┘
         │                │                │                │
         ▼                ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      GitHub Actions Workflows                        │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │  Stage 1: Code Quality & Security                        │       │
│  │  • Linting (golangci-lint)                              │       │
│  │  • Formatting check (go fmt)                             │       │
│  │  • Security scan (gosec, CodeQL)                         │       │
│  │  • Dependency scan (govulncheck)                         │       │
│  │  • Secret detection (TruffleHog, Gitleaks)              │       │
│  │  Duration: ~3-5 minutes                                  │       │
│  └──────────────────────────────────────────────────────────┘       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │  Stage 2: Testing                                        │       │
│  │  • Unit tests (Go 1.23, 1.24)                           │       │
│  │  • Integration tests (PostgreSQL, Redis)                 │       │
│  │  • Code coverage (threshold: 70%)                        │       │
│  │  • Performance benchmarks                                │       │
│  │  Duration: ~5-8 minutes                                  │       │
│  └──────────────────────────────────────────────────────────┘       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │  Stage 3: Build & Push                                   │       │
│  │  • Multi-stage Docker builds                             │       │
│  │  • Multi-architecture (amd64, arm64)                     │       │
│  │  • Container security scan (Trivy)                       │       │
│  │  • Push to GHCR                                          │       │
│  │  Services: api, websocket, fix-gateway, worker          │       │
│  │  Duration: ~8-12 minutes                                 │       │
│  └──────────────────────────────────────────────────────────┘       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │  Stage 4: Deploy                                         │       │
│  │  • Staging (develop branch)                              │       │
│  │  • Production (release tag, blue-green)                  │       │
│  │  • Helm deployment                                       │       │
│  │  • Smoke tests                                           │       │
│  │  • Load tests (staging only)                             │       │
│  │  Duration: ~10-15 minutes                                │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster (EKS)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │   API    │  │WebSocket │  │   FIX    │  │  Worker  │           │
│  │  Pods    │  │  Pods    │  │ Gateway  │  │   Pods   │           │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘           │
└─────────────────────────────────────────────────────────────────────┘
```

## Pipeline Stages

### 1. Code Quality & Security (3-5 minutes)

**Runs on:** All branches, all commits

```yaml
Jobs:
  - lint
  - security
```

**Checks:**
- Go code formatting (`go fmt`)
- Linting with golangci-lint
- Static security analysis (gosec)
- Dependency vulnerabilities (govulncheck)
- Secret detection (TruffleHog, Gitleaks)
- License compliance

**Failure Action:** Block PR merge, notify developer

### 2. Testing (5-8 minutes)

**Runs on:** All branches, all commits

```yaml
Jobs:
  - test (matrix: Go 1.23, 1.24)
  - integration
  - benchmark (PRs only)
```

**Tests:**
- Unit tests with race detection
- Integration tests with PostgreSQL & Redis
- Code coverage (minimum 70%)
- Performance regression tests

**Artifacts:**
- Coverage reports (Codecov)
- Benchmark comparisons
- Test results

**Failure Action:** Block PR merge, detailed report

### 3. Build & Push (8-12 minutes)

**Runs on:** main, develop, staging branches after tests pass

```yaml
Jobs:
  - build (matrix: api, websocket, fix-gateway, worker)
```

**Process:**
1. Multi-stage Docker build
2. Build for linux/amd64 and linux/arm64
3. Security scan with Trivy
4. Push to GitHub Container Registry
5. Tag with branch name and SHA

**Artifacts:**
- Docker images
- Security scan results

**Failure Action:** Notify DevOps team, block deployment

### 4. Deployment (10-15 minutes)

**Runs on:**
- Staging: develop branch
- Production: release tags

```yaml
Jobs:
  - deploy-staging (develop branch)
  - deploy-production (release tags)
```

**Deployment Strategy:**

#### Staging Deployment (Automated)
```bash
trigger: push to develop
strategy: rolling update
duration: ~5 minutes
smoke tests: yes
load tests: yes
rollback: automatic on failure
```

#### Production Deployment (Blue-Green)
```bash
trigger: release tag
strategy: blue-green
approval: required
duration: ~15 minutes
smoke tests: yes
monitoring: 5 minutes before cutover
rollback: manual/automatic
```

### 5. Load Testing (5-10 minutes)

**Runs on:** Staging after deployment

```yaml
Job: load-test
```

**Configuration:**
- Tool: k6
- Virtual users: 100 → 1000 → 5000
- Duration: 30 minutes
- Scenarios: REST API, WebSocket, order execution

**Thresholds:**
- p95 latency < 500ms
- p99 latency < 1000ms
- Error rate < 1%
- Order execution < 300ms

**Failure Action:** Alert team, block production promotion

## Environment Strategy

### Development
- **Trigger:** Every commit to feature branches
- **Tests:** Full test suite
- **Deployment:** No automatic deployment
- **Purpose:** Developer validation

### Staging
- **Trigger:** Merge to develop branch
- **Tests:** Full suite + integration tests
- **Deployment:** Automatic
- **Purpose:** Pre-production testing
- **URL:** https://staging.trading-engine.com
- **Data:** Anonymized production snapshot

### Production
- **Trigger:** Release tag (v*.*.*)
- **Tests:** Full suite + load tests
- **Deployment:** Blue-green with approval
- **Purpose:** Live customer traffic
- **URL:** https://api.trading-engine.com
- **Data:** Live production data

## Deployment Strategies

### Rolling Update (Staging)

```yaml
strategy:
  type: RollingUpdate
  maxSurge: 1        # 1 extra pod during update
  maxUnavailable: 0  # No downtime
```

**Process:**
1. Create new pod with new version
2. Wait for health check (30s)
3. Route traffic to new pod
4. Terminate old pod
5. Repeat for all pods

**Pros:** Simple, gradual rollout
**Cons:** Mixed versions during deployment
**Rollback:** Fast (< 2 minutes)

### Blue-Green (Production)

```yaml
environments:
  blue:  # Currently serving traffic
    replicas: 5
    version: v1.0.0
  green: # New version being deployed
    replicas: 5
    version: v1.1.0
```

**Process:**
1. Deploy to inactive environment (green)
2. Run smoke tests on green
3. Switch 5% traffic to green (canary)
4. Monitor for 5 minutes
5. Switch 100% traffic to green
6. Monitor for 10 minutes
7. Scale down blue to 1 replica

**Pros:** Zero downtime, instant rollback
**Cons:** Double resources during deployment
**Rollback:** Instant (switch back to blue)

### Canary Deployment (Optional)

```yaml
phases:
  - 5% traffic  → monitor 5 min
  - 25% traffic → monitor 10 min
  - 50% traffic → monitor 15 min
  - 100% traffic → complete
```

**Use case:** High-risk changes, major versions

## Security Scanning

### Daily Security Scan

**Schedule:** 2 AM UTC daily

```yaml
Scans:
  - Dependency vulnerabilities
  - Container image CVEs
  - SAST (gosec)
  - Secret detection
  - IaC security (tfsec, Checkov)
  - License compliance
```

**Severity Levels:**
- **Critical:** Immediate notification, block deployments
- **High:** Notification, fix within 7 days
- **Medium:** Track in backlog
- **Low:** Informational

### DAST (Dynamic Testing)

**Schedule:** Weekly on staging

```yaml
Tool: OWASP ZAP
Target: https://staging.trading-engine.com
Tests:
  - SQL injection
  - XSS
  - CSRF
  - Authentication bypass
  - API security
```

## Monitoring & Observability

### Deployment Tracking

```bash
# Every deployment creates:
- Deployment annotation in Grafana
- Alert silence during deployment
- Rollout status dashboard
- Performance baseline snapshot
```

### Key Metrics Post-Deployment

**Monitor for 1 hour:**
- Error rate (< 0.1%)
- Latency p95 (< 500ms)
- Throughput (baseline ±10%)
- Database queries (no N+1)
- Memory usage (stable)
- CPU usage (baseline ±20%)

### Automatic Rollback Triggers

```yaml
conditions:
  - error_rate > 5%
  - latency_p95 > 2000ms
  - pod_restarts > 3
  - health_check_failures > 5
action: automatic rollback
notification: PagerDuty alert
```

## Infrastructure as Code

### Terraform Workflow

```bash
# On infrastructure changes:
1. terraform plan (automated)
2. Review plan in PR
3. Approve PR
4. terraform apply (automated on merge)
5. Update documentation
```

### Managed Resources

- VPC and networking
- EKS cluster
- RDS PostgreSQL
- ElastiCache Redis
- S3 buckets
- IAM roles
- Security groups
- Load balancers

## Performance Benchmarks

### Target Metrics

| Metric | Target | Measured |
|--------|--------|----------|
| Build time | < 15 min | ~12 min |
| Test time | < 10 min | ~8 min |
| Deploy time (staging) | < 5 min | ~4 min |
| Deploy time (prod) | < 20 min | ~15 min |
| Total pipeline | < 40 min | ~35 min |

### Optimization Strategies

1. **Caching:**
   - Go module cache
   - Docker layer cache
   - Build artifact cache

2. **Parallelization:**
   - Matrix builds
   - Parallel test suites
   - Concurrent deployments

3. **Resource Allocation:**
   - Large GitHub runners
   - Dedicated build nodes

## Troubleshooting

### Pipeline Failures

#### Lint Failures
```bash
Error: golangci-lint failed
Fix: Run locally: golangci-lint run ./...
Prevention: Pre-commit hooks
```

#### Test Failures
```bash
Error: Tests failed on Go 1.24
Fix: Run locally: go test ./...
Check: Version-specific issues
```

#### Build Failures
```bash
Error: Docker build failed
Fix: Test locally: docker build -f Dockerfile .
Check: Dependency availability
```

#### Deployment Failures
```bash
Error: Helm upgrade failed
Fix: Check logs: kubectl logs -n production
Rollback: helm rollback trading-engine
```

### Common Issues

**1. Slow builds**
- Check cache hit rate
- Review Docker layer optimization
- Parallel job configuration

**2. Flaky tests**
- Identify intermittent failures
- Add retries for network tests
- Review test isolation

**3. Deployment timeout**
- Check pod startup time
- Review health check configuration
- Increase timeout threshold

## Best Practices

### Commit Messages
```
feat: add WebSocket reconnection logic
fix: resolve memory leak in order processor
perf: optimize database query for positions
docs: update API documentation
```

### Branch Strategy
```
main       → production (protected)
develop    → staging (protected)
staging    → pre-production testing
feature/*  → development work
hotfix/*   → emergency fixes
```

### Version Tagging
```
v1.2.3
 │ │ │
 │ │ └─ Patch (bug fixes)
 │ └─── Minor (new features, backwards compatible)
 └───── Major (breaking changes)
```

### Deployment Windows

**Preferred:**
- Tuesday-Thursday, 10 AM - 2 PM EST
- Low trading volume periods

**Avoid:**
- Market open/close hours
- Fridays (limited support window)
- Major holidays
- High-volume trading events

## Metrics & Reporting

### Daily Report
- Deployment count
- Success rate
- Average pipeline duration
- Test coverage trends

### Weekly Report
- Failed deployments analysis
- Security vulnerability summary
- Performance trends
- Resource utilization

### Monthly Report
- Deployment frequency
- MTTR (Mean Time To Recovery)
- Change failure rate
- Lead time for changes

## Continuous Improvement

### Quarterly Reviews
- Pipeline performance optimization
- Tool updates and upgrades
- Process refinement
- Team retrospective

### Automation Goals
- Reduce manual steps
- Improve test coverage
- Faster feedback loops
- Enhanced observability

---

**Last Updated:** 2026-01-18
**Version:** 1.0
**Owner:** DevOps Team
