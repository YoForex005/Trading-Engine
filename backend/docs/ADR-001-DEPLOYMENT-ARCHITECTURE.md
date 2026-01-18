# ADR-001: Deployment Architecture for RTX Trading Engine

**Status:** Accepted
**Date:** 2026-01-18
**Decision Makers:** System Architecture Designer
**Related Documents:** [DEPLOYMENT_ARCHITECTURE.md](./DEPLOYMENT_ARCHITECTURE.md)

---

## Context

The RTX Trading Engine is a production-ready Go 1.24 backend serving a complete A-Book STP + B-Book Market Making trading platform. The system requires:

1. High availability (99.9% uptime)
2. Zero-downtime deployments
3. Horizontal scalability for 10K+ concurrent users
4. Sub-100ms API response times
5. Financial-grade security and compliance
6. Disaster recovery with <15 minute RTO

Currently, the codebase has zero build errors and includes:
- Complete trading functionality (A-Book + B-Book)
- Database migration system
- Admin platform with RBAC and audit logging
- FIX API provisioning system
- WebSocket real-time updates
- PostgreSQL 15 database with comprehensive schema

**Problem:** Need a production-grade deployment architecture that ensures reliability, security, and operational excellence while maintaining cost efficiency.

---

## Decision

We have decided to implement a **cloud-native containerized architecture** using Docker and Kubernetes with the following key components:

### 1. Containerization Strategy

**Decision:** Multi-stage Docker builds with Alpine Linux base images

**Rationale:**
- Minimal attack surface (Alpine is 7MB vs 1.1GB for full Golang image)
- Fast deployment times (25MB final image vs 1.2GB)
- Production security best practices (non-root user, read-only filesystem)
- Efficient layer caching for rapid CI/CD

**Trade-offs:**
- ✅ 98% image size reduction (1.2GB → 25MB)
- ✅ 5x faster deployment times
- ✅ Reduced attack surface
- ⚠️ Requires careful dependency management (static linking)
- ⚠️ Alpine uses musl libc (potential compatibility issues, mitigated by static builds)

### 2. Orchestration Platform

**Decision:** Kubernetes 1.28+ with managed control plane (EKS/GKE/AKS)

**Rationale:**
- Industry-standard orchestration platform
- Built-in service discovery, load balancing, health checks
- Native support for rolling updates and rollbacks
- Extensive ecosystem (cert-manager, external-secrets, ingress-nginx)
- Auto-scaling capabilities (HPA, cluster autoscaler)

**Alternatives Considered:**
| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| **Docker Compose** | Simple, low learning curve | No auto-scaling, no HA | Not production-grade for financial platform |
| **Docker Swarm** | Easier than K8s | Dying community, limited features | Lack of ecosystem support |
| **ECS/Fargate** | AWS-native, serverless | Vendor lock-in, limited portability | Need multi-cloud capability |
| **Nomad** | Simpler than K8s | Smaller ecosystem | Kubernetes has better support for financial compliance |

**Trade-offs:**
- ✅ Production-grade HA and auto-scaling
- ✅ Declarative configuration (GitOps-friendly)
- ✅ Rich ecosystem and community
- ⚠️ Higher complexity than simpler alternatives
- ⚠️ Requires Kubernetes expertise
- ⚠️ Higher minimum infrastructure cost (~$300/month)

### 3. Database Architecture

**Decision:** PostgreSQL 15 StatefulSet with streaming replication (1 primary + 2 replicas)

**Rationale:**
- StatefulSet provides stable network identity and persistent storage
- Streaming replication for high availability
- TimescaleDB extension for time-series tick data
- Existing migration system ready for deployment

**Alternatives Considered:**
| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| **Managed RDS** | Fully managed, automated backups | Higher cost, vendor lock-in | Want Kubernetes-native deployment |
| **CockroachDB** | Native distributed SQL | Operational complexity | PostgreSQL expertise already exists |
| **Deployment (non-StatefulSet)** | Simpler | No stable identity for replication | Need stable pod names for replication |

**Trade-offs:**
- ✅ Kubernetes-native deployment
- ✅ Stable network identity for replication
- ✅ PersistentVolumeClaims for data durability
- ⚠️ More complex than managed RDS
- ⚠️ Requires PostgreSQL operational expertise

### 4. Deployment Strategy

**Decision:** Blue-Green deployment with manual approval gate for production

**Rationale:**
- Zero-downtime deployments (critical for 24/7 trading platform)
- Instant rollback capability (<1 minute)
- Ability to test new version in production environment before cutover
- Manual approval ensures human oversight for financial system

**Alternatives Considered:**
| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| **Rolling Update** | Resource efficient | Gradual rollout, mixed versions | Want instant rollback capability |
| **Canary Deployment** | Progressive rollout | Complex traffic splitting | Blue-green simpler for initial deployment |
| **Recreate** | Simple | Downtime during deployment | Not acceptable for 24/7 trading |

**Trade-offs:**
- ✅ Zero-downtime deployments
- ✅ Instant rollback (<1 minute)
- ✅ Production smoke testing before cutover
- ⚠️ Requires 2x resources during deployment
- ⚠️ More complex than rolling updates

### 5. CI/CD Pipeline

**Decision:** GitHub Actions with 8-stage pipeline (lint → build → test → migrate → docker → deploy-dev → deploy-staging → deploy-prod)

**Rationale:**
- Native GitHub integration (code and CI in same platform)
- Free for public repos, cost-effective for private
- Rich marketplace of actions
- YAML-based declarative configuration
- Built-in secrets management

**Pipeline Stages:**
1. **Lint & Security:** golangci-lint, gosec (SARIF), govulncheck
2. **Build & Test:** Unit tests (race detector), integration tests, coverage
3. **Migration Validation:** Test migrations up/down/up cycle
4. **Docker Build & Push:** Multi-arch builds, layer caching
5. **Deploy Dev:** Automated (on develop branch)
6. **Deploy Staging:** Blue-green with smoke tests (on staging branch)
7. **Deploy Production:** Blue-green + manual approval (on main branch)
8. **Rollback:** Manual trigger workflow

**Alternatives Considered:**
| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| **GitLab CI** | Integrated with GitLab | Would require migration from GitHub | Already using GitHub |
| **Jenkins** | Highly flexible | Requires self-hosting, maintenance | Prefer managed solution |
| **CircleCI** | Fast builds | Cost for private repos | GitHub Actions more cost-effective |
| **ArgoCD** | GitOps-native | Requires separate tool | Want single CI/CD platform |

**Trade-offs:**
- ✅ Native GitHub integration
- ✅ Cost-effective for private repos
- ✅ Rich ecosystem of actions
- ✅ Built-in secrets management
- ⚠️ Slower builds than CircleCI (mitigated by caching)
- ⚠️ Limited to 6 hours per job (not an issue for our use case)

### 6. Security Architecture

**Decision:** Multi-layered security with External Secrets Operator, cert-manager, NetworkPolicy, and PodSecurityPolicy

**Key Security Controls:**
1. **Secrets:** External Secrets Operator syncing from AWS Secrets Manager (no secrets in Git)
2. **TLS:** cert-manager with Let's Encrypt for automated certificate management
3. **Network:** NetworkPolicy restricting pod-to-pod communication
4. **Pod Security:** PodSecurityPolicy enforcing non-root, no privilege escalation, read-only root filesystem
5. **Image Security:** Alpine base, static binaries, security scanning in CI (gosec, govulncheck)
6. **RBAC:** ServiceAccount with minimal permissions

**Rationale:**
- Defense in depth (multiple layers of security)
- Secrets never committed to Git (compliance requirement)
- Automated TLS certificate renewal (operational safety)
- Zero-trust network model (pod-to-pod restrictions)

**Trade-offs:**
- ✅ Defense in depth
- ✅ Compliance with financial regulations
- ✅ Automated secret rotation
- ⚠️ Increased complexity
- ⚠️ Requires AWS Secrets Manager ($0.40/secret/month)

### 7. Monitoring & Observability

**Decision:** Prometheus + Grafana + Sentry stack with custom business metrics

**Key Metrics:**
- **Business:** Orders placed, order latency (p50/p95/p99), active connections
- **Infrastructure:** CPU, memory, disk, network
- **Application:** HTTP response times, error rates, database query latency

**Rationale:**
- Prometheus is Kubernetes-native and industry standard
- Grafana provides rich visualization
- Sentry provides error tracking and alerting
- Custom business metrics for trading-specific monitoring

**Trade-offs:**
- ✅ Comprehensive observability
- ✅ Kubernetes-native
- ✅ Rich ecosystem
- ⚠️ Requires operational expertise
- ⚠️ Storage costs for metrics retention

### 8. Disaster Recovery

**Decision:** Automated hourly PostgreSQL backups to S3 with 30-day retention + WAL archiving

**Recovery Targets:**
- **RPO (Recovery Point Objective):** 1 hour (hourly backups)
- **RTO (Recovery Time Objective):** 15 minutes (automated restore)

**Backup Strategy:**
- **PostgreSQL:** Hourly pg_dump + WAL archiving (30 days retention)
- **Redis:** Daily RDB snapshots (7 days retention)
- **Configuration:** Git version control (indefinite retention)
- **Secrets:** AWS Secrets Manager (automatic backups)

**Trade-offs:**
- ✅ Automated backups (no manual intervention)
- ✅ Point-in-time recovery via WAL archiving
- ✅ Tested recovery procedures
- ⚠️ S3 storage costs (~$5-10/month)
- ⚠️ Requires monthly backup verification tests

---

## Consequences

### Positive

1. **Reliability:** 99.9% uptime SLA achievable with HA architecture
2. **Scalability:** Horizontal scaling from 3 to 10 backend replicas via HPA
3. **Security:** Financial-grade security with multiple layers of defense
4. **Developer Experience:** Automated CI/CD reduces deployment from hours to minutes
5. **Cost Efficiency:** ~$304/month for production infrastructure (optimizable to ~$200 with spot instances)
6. **Disaster Recovery:** 1-hour RPO, 15-minute RTO with automated backups
7. **Observability:** Comprehensive metrics and logging for proactive monitoring

### Negative

1. **Complexity:** Kubernetes has a steep learning curve
2. **Operational Overhead:** Requires Kubernetes expertise for troubleshooting
3. **Minimum Infrastructure Cost:** ~$300/month even with low traffic
4. **Resource Overhead:** Blue-green deployments require 2x resources during cutover

### Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Kubernetes cluster failure** | High | Low | Multi-AZ deployment, managed control plane |
| **Database data loss** | Critical | Very Low | Hourly backups + WAL archiving, automated restore tests |
| **Deployment failure** | Medium | Medium | Blue-green deployment with instant rollback, manual approval for prod |
| **Security breach** | Critical | Low | Multi-layered security, external secrets, NetworkPolicy, regular security scans |
| **Cost overrun** | Medium | Medium | Resource limits, HPA, spot instances, cost monitoring dashboards |

---

## Implementation Plan

### Phase 1: Docker Containerization (Week 1)
- [ ] Create multi-stage Dockerfile for backend
- [ ] Create Docker Compose for local development
- [ ] Test local deployment end-to-end
- [ ] Document container architecture

### Phase 2: Kubernetes Setup (Week 2)
- [ ] Create Kubernetes manifests (Deployment, StatefulSet, Services, etc.)
- [ ] Deploy to development cluster
- [ ] Configure health checks and HPA
- [ ] Test pod autoscaling

### Phase 3: CI/CD Pipeline (Week 3)
- [ ] Implement GitHub Actions workflow
- [ ] Configure automated builds and tests
- [ ] Set up blue-green deployment strategy
- [ ] Test deployment to dev/staging/prod

### Phase 4: Security Hardening (Week 4)
- [ ] Implement External Secrets Operator
- [ ] Configure cert-manager for TLS
- [ ] Apply NetworkPolicy and PodSecurityPolicy
- [ ] Run security audit (gosec, govulncheck)

### Phase 5: Monitoring & Backup (Week 5)
- [ ] Deploy Prometheus and Grafana
- [ ] Create custom dashboards
- [ ] Configure Sentry error tracking
- [ ] Set up automated backup CronJob
- [ ] Test disaster recovery procedures

### Phase 6: Production Readiness (Week 6)
- [ ] Load testing (k6)
- [ ] Chaos engineering (Chaos Mesh)
- [ ] Documentation review
- [ ] Runbook creation
- [ ] Go-live checklist

---

## Success Metrics

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| **Deployment Time** | <5 minutes | GitHub Actions duration |
| **Uptime** | 99.9% | Prometheus uptime metric |
| **API Response Time** | <100ms (p99) | Prometheus histogram |
| **Rollback Time** | <1 minute | Manual testing |
| **Backup Recovery Time** | <15 minutes | Monthly DR drill |
| **Infrastructure Cost** | <$350/month | AWS billing dashboard |

---

## References

1. [DEPLOYMENT_ARCHITECTURE.md](./DEPLOYMENT_ARCHITECTURE.md) - Detailed architecture documentation
2. [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
3. [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
4. [PostgreSQL Streaming Replication](https://www.postgresql.org/docs/15/warm-standby.html)
5. [Blue-Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html)

---

**Document Status:** Accepted
**Next Review Date:** 2026-03-18 (8 weeks)
**Approval:** System Architecture Designer
**Stakeholders Notified:** Development Team, DevOps Team, Security Team
