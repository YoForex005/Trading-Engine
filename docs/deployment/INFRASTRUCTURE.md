# Infrastructure Architecture - Trading Engine

## Overview

This document describes the complete infrastructure architecture for the Trading Engine platform, including cloud resources, networking, security, monitoring, and disaster recovery.

## Infrastructure Components

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              AWS Cloud                                   │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  Region: us-east-1 (Primary)                VPC: 10.0.0.0/16   │    │
│  │                                                                 │    │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐ │    │
│  │  │ Availability     │  │ Availability     │  │ Availability │ │    │
│  │  │ Zone A           │  │ Zone B           │  │ Zone C       │ │    │
│  │  │                  │  │                  │  │              │ │    │
│  │  │ Public:          │  │ Public:          │  │ Public:      │ │    │
│  │  │ 10.0.101.0/24   │  │ 10.0.102.0/24   │  │ 10.0.103.0/24│ │    │
│  │  │ ┌──────────┐    │  │ ┌──────────┐    │  │ ┌──────────┐ │ │    │
│  │  │ │  ALB     │    │  │ │  ALB     │    │  │ │  ALB     │ │ │    │
│  │  │ │  NAT GW  │    │  │ │  NAT GW  │    │  │ │  NAT GW  │ │ │    │
│  │  │ └──────────┘    │  │ └──────────┘    │  │ └──────────┘ │ │    │
│  │  │                  │  │                  │  │              │ │    │
│  │  │ Private:         │  │ Private:         │  │ Private:     │ │    │
│  │  │ 10.0.1.0/24     │  │ 10.0.2.0/24     │  │ 10.0.3.0/24  │ │    │
│  │  │ ┌──────────┐    │  │ ┌──────────┐    │  │ ┌──────────┐ │ │    │
│  │  │ │   EKS    │    │  │ │   EKS    │    │  │ │   EKS    │ │ │    │
│  │  │ │  Nodes   │    │  │ │  Nodes   │    │  │ │  Nodes   │ │ │    │
│  │  │ └──────────┘    │  │ └──────────┘    │  │ └──────────┘ │ │    │
│  │  │                  │  │                  │  │              │ │    │
│  │  │ Database:        │  │ Database:        │  │ Database:    │ │    │
│  │  │ 10.0.11.0/24    │  │ 10.0.12.0/24    │  │ 10.0.13.0/24 │ │    │
│  │  │ ┌──────────┐    │  │ ┌──────────┐    │  │              │ │    │
│  │  │ │   RDS    │◄───┼──┼─┤   RDS    │    │  │              │ │    │
│  │  │ │ Primary  │    │  │ │ Standby  │    │  │              │ │    │
│  │  │ └──────────┘    │  │ └──────────┘    │  │              │ │    │
│  │  │ ┌──────────┐    │  │ ┌──────────┐    │  │ ┌──────────┐ │ │    │
│  │  │ │  Redis   │◄───┼──┼─┤  Redis   │◄───┼──┼─┤  Redis   │ │ │    │
│  │  │ │ Primary  │    │  │ │ Replica  │    │  │ │ Replica  │ │ │    │
│  │  │ └──────────┘    │  │ └──────────┘    │  │ └──────────┘ │ │    │
│  │  └──────────────────┘  └──────────────────┘  └──────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │  Global Services                                                  │  │
│  │  • Route53 (DNS)                                                 │  │
│  │  • CloudFront (CDN)                                              │  │
│  │  • WAF (Web Application Firewall)                               │  │
│  │  • S3 (Backups, Logs, Static Assets)                           │  │
│  │  • CloudWatch (Monitoring & Logs)                               │  │
│  │  • Secrets Manager                                               │  │
│  │  • KMS (Encryption Keys)                                         │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                   Disaster Recovery Region: us-west-2                    │
│  • Standby RDS replica                                                  │
│  • S3 cross-region replication                                          │
│  • EKS cluster (minimal, can scale up)                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Network Architecture

### VPC Configuration

```hcl
CIDR: 10.0.0.0/16
Total IPs: 65,536

Subnets:
  Public (3 AZs):
    - 10.0.101.0/24 (us-east-1a) - 256 IPs
    - 10.0.102.0/24 (us-east-1b) - 256 IPs
    - 10.0.103.0/24 (us-east-1c) - 256 IPs

  Private (3 AZs):
    - 10.0.1.0/24 (us-east-1a) - 256 IPs
    - 10.0.2.0/24 (us-east-1b) - 256 IPs
    - 10.0.3.0/24 (us-east-1c) - 256 IPs

  Database (3 AZs):
    - 10.0.11.0/24 (us-east-1a) - 256 IPs
    - 10.0.12.0/24 (us-east-1b) - 256 IPs
    - 10.0.13.0/24 (us-east-1c) - 256 IPs
```

### Security Groups

#### ALB Security Group
```
Inbound:
  - Port 443 (HTTPS) from 0.0.0.0/0
  - Port 80 (HTTP) from 0.0.0.0/0 (redirect to 443)

Outbound:
  - Port 7999 to EKS nodes (API)
  - Port 8080 to EKS nodes (WebSocket)
```

#### EKS Node Security Group
```
Inbound:
  - Port 7999 from ALB (API)
  - Port 8080 from ALB (WebSocket)
  - Port 443 from VPC (Kubernetes API)
  - All traffic from other EKS nodes

Outbound:
  - Port 5432 to RDS
  - Port 6379 to Redis
  - Port 443 to Internet (package downloads)
```

#### RDS Security Group
```
Inbound:
  - Port 5432 from EKS nodes only

Outbound:
  - None (ingress only)
```

#### Redis Security Group
```
Inbound:
  - Port 6379 from EKS nodes only

Outbound:
  - None (ingress only)
```

## Compute Resources

### EKS Cluster

**Version:** 1.28
**Control Plane:** Managed by AWS
**Add-ons:**
- CoreDNS
- kube-proxy
- VPC CNI
- EBS CSI Driver
- Cluster Autoscaler
- Metrics Server

### Node Groups

#### General Purpose (On-Demand)
```yaml
Instance Type: t3.xlarge
vCPU: 4
Memory: 16 GB
Min Nodes: 3
Max Nodes: 10
Disk: 100 GB EBS gp3
Use Case: API, WebSocket, general workloads
```

#### Compute Optimized (Spot)
```yaml
Instance Type: c6i.2xlarge
vCPU: 8
Memory: 16 GB
Min Nodes: 2
Max Nodes: 20
Disk: 100 GB EBS gp3
Use Case: Workers, batch processing
Spot Savings: ~70%
```

#### Memory Optimized (On-Demand)
```yaml
Instance Type: r6i.xlarge
vCPU: 4
Memory: 32 GB
Min Nodes: 1
Max Nodes: 5
Disk: 100 GB EBS gp3
Use Case: Cache-heavy workloads, data processing
```

### Auto-Scaling

#### Horizontal Pod Autoscaler (HPA)
```yaml
API Server:
  Min Replicas: 3
  Max Replicas: 10
  Target CPU: 70%
  Target Memory: 80%
  Scale Up: +100% every 30s (max 4 pods)
  Scale Down: -50% every 5m (max 2 pods)

WebSocket Server:
  Min Replicas: 5
  Max Replicas: 20
  Target CPU: 60%
  Target Memory: 70%
```

#### Cluster Autoscaler
```yaml
Scan Interval: 10s
Scale Down Delay: 10m
Unneeded Time: 10m
Utilization Threshold: 50%
Max Node Lifetime: 7 days
```

## Data Storage

### RDS PostgreSQL

**Instance Class:** db.r6i.xlarge
**Engine:** PostgreSQL 16.1
**Storage:** 100 GB → 1000 GB (auto-scaling)
**IOPS:** 3000 IOPS (gp3)
**Encryption:** AES-256 (KMS)

**Configuration:**
```sql
max_connections: 1000
shared_buffers: 4 GB
effective_cache_size: 12 GB
maintenance_work_mem: 2 GB
work_mem: 10 MB
random_page_cost: 1.1
effective_io_concurrency: 200
```

**High Availability:**
- Multi-AZ: Yes
- Automated backups: 30 days
- Backup window: 03:00-04:00 UTC
- Maintenance window: Mon 04:00-05:00 UTC
- Read replicas: 2 (for reporting)

**Monitoring:**
- Enhanced monitoring: 60-second intervals
- Performance Insights: Enabled
- CloudWatch alarms: CPU, storage, connections

### ElastiCache Redis

**Node Type:** cache.r6g.large
**Engine:** Redis 7.1
**Nodes:** 1 primary + 2 replicas
**Memory:** 13.07 GB per node
**Encryption:** At-rest & in-transit

**Configuration:**
```
maxmemory-policy: allkeys-lru
timeout: 300
tcp-keepalive: 300
```

**High Availability:**
- Multi-AZ: Yes
- Automatic failover: Enabled
- Backup retention: 7 days
- Backup window: 03:00-04:00 UTC

### S3 Buckets

#### Backups Bucket
```
Name: production-trading-engine-backups
Versioning: Enabled
Encryption: KMS
Lifecycle:
  - 30 days → Standard-IA
  - 90 days → Glacier
  - 365 days → Delete
Cross-Region Replication: us-west-2
```

#### Logs Bucket
```
Name: production-trading-engine-logs
Versioning: Disabled
Encryption: S3-managed
Lifecycle:
  - 90 days → Delete
```

#### Static Assets Bucket
```
Name: production-trading-engine-assets
Versioning: Enabled
Encryption: S3-managed
CloudFront: Enabled
Cache-Control: max-age=31536000
```

## Security

### Encryption

**At Rest:**
- EBS volumes: KMS encrypted
- RDS database: KMS encrypted
- Redis: KMS encrypted
- S3 buckets: KMS/SSE-S3 encrypted

**In Transit:**
- ALB → EKS: TLS 1.3
- EKS → RDS: SSL
- EKS → Redis: TLS
- Client → ALB: TLS 1.3

### IAM Roles

#### EKS Node Role
```
Policies:
  - AmazonEKSWorkerNodePolicy
  - AmazonEKS_CNI_Policy
  - AmazonEC2ContainerRegistryReadOnly
  - Custom: RDSAccess (limited)
  - Custom: S3BackupAccess (write-only)
```

#### Pod Service Accounts (IRSA)
```
API Service Account:
  - Access to Secrets Manager
  - Access to S3 backups (read/write)

Worker Service Account:
  - Access to SQS queues
  - Access to S3 assets (read-only)
```

### Network Security

**WAF Rules:**
- Rate limiting: 1000 req/min per IP
- Geo-blocking: High-risk countries
- SQL injection protection
- XSS protection
- Known bad signatures

**Network Policies:**
- Default deny all
- Allow API → Database
- Allow API → Redis
- Allow WebSocket → API
- Allow Workers → Database
- Deny all other inter-pod traffic

### Secrets Management

**AWS Secrets Manager:**
- Database credentials
- API keys (OANDA, Binance)
- JWT secrets
- Encryption keys

**Rotation:**
- Database passwords: 90 days
- API keys: Manual rotation
- JWT secrets: 180 days

## Monitoring & Observability

### CloudWatch

**Metrics:**
- EC2: CPU, Memory, Disk, Network
- RDS: CPU, Connections, IOPS, Latency
- Redis: CPU, Memory, Evictions, Connections
- ALB: Request count, Latency, Error rate
- EKS: Pod count, Node count, Resource usage

**Logs:**
- Application logs → CloudWatch Logs
- ALB access logs → S3
- VPC flow logs → S3
- RDS slow query logs → CloudWatch

**Alarms:**
- Critical (PagerDuty): Service down, High error rate
- Warning (Slack): High CPU, High memory
- Info (Email): Deployment notifications

### Prometheus Stack

**Deployed in Kubernetes:**
```
Components:
  - Prometheus (metrics collection)
  - Grafana (visualization)
  - Alertmanager (alert routing)
  - Node Exporter (system metrics)
  - PostgreSQL Exporter (database metrics)
  - Redis Exporter (cache metrics)
```

**Retention:**
- Prometheus: 30 days
- CloudWatch: 90 days
- S3 logs: 90 days

### Distributed Tracing

**Jaeger:**
- Trace collection from all services
- Sampling rate: 10%
- Retention: 7 days
- Storage: Elasticsearch

### Log Aggregation

**Option 1: CloudWatch Logs**
- Built-in AWS integration
- Simple setup
- Cost-effective for moderate volume

**Option 2: ELK Stack**
- Elasticsearch for storage
- Logstash for processing
- Kibana for visualization
- Higher cost, more features

## Disaster Recovery

### RTO/RPO Targets

| Scenario | RTO | RPO |
|----------|-----|-----|
| Single pod failure | < 1 min | 0 |
| Node failure | < 5 min | 0 |
| AZ failure | < 10 min | < 1 min |
| Region failure | < 1 hour | < 5 min |
| Complete data loss | < 4 hours | < 1 hour |

### Backup Strategy

**Database:**
- Automated snapshots: Daily
- Point-in-time recovery: Enabled
- Snapshot retention: 30 days
- Cross-region backup: Yes

**Application State:**
- Velero backups: Daily
- Retention: 30 days
- Includes: PVCs, ConfigMaps, Secrets

**Restoration Testing:**
- Monthly: Backup verification
- Quarterly: Full DR drill

### Failover Procedures

#### Database Failover
```bash
# Automatic (Multi-AZ)
RTO: ~2 minutes
RPO: 0

# Manual (Cross-Region)
1. Promote replica in us-west-2
2. Update DNS to point to DR region
3. Scale up EKS in DR region
RTO: ~30 minutes
RPO: <5 minutes
```

#### Application Failover
```bash
# Cross-Region
1. Deploy latest version to DR region
2. Update Route53 health check
3. DNS failover to DR region
4. Scale up EKS to production size
RTO: ~1 hour
```

## Cost Optimization

### Monthly Cost Breakdown (Estimated)

| Service | Cost/Month |
|---------|-----------|
| EKS Control Plane | $72 |
| EC2 (On-Demand) | ~$800 |
| EC2 (Spot) | ~$240 |
| RDS PostgreSQL | ~$450 |
| ElastiCache Redis | ~$280 |
| Data Transfer | ~$200 |
| S3 Storage | ~$50 |
| CloudWatch | ~$100 |
| ALB | ~$30 |
| Route53 | ~$5 |
| **Total** | **~$2,227/month** |

### Optimization Strategies

1. **Right-sizing:**
   - Review resource utilization monthly
   - Adjust instance types based on metrics
   - Use VPA for pod right-sizing

2. **Spot Instances:**
   - Use for stateless workloads
   - Save ~70% on compute
   - Implement graceful shutdown

3. **Reserved Instances:**
   - 1-year RDS Reserved: Save ~30%
   - 1-year EC2 Reserved: Save ~30%
   - For stable baseline capacity

4. **Storage Optimization:**
   - S3 lifecycle policies
   - Delete old logs
   - Compress backups

5. **Data Transfer:**
   - Use VPC endpoints (S3, DynamoDB)
   - CloudFront for static assets
   - Minimize cross-region transfer

## Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| API Latency (p95) | < 500ms | ~350ms |
| API Latency (p99) | < 1000ms | ~750ms |
| WebSocket Latency | < 100ms | ~50ms |
| Order Execution | < 300ms | ~200ms |
| Database Query (p95) | < 100ms | ~60ms |
| Throughput | > 10K req/s | ~12K req/s |
| Concurrent WebSocket | > 50K | ~60K |
| Uptime | 99.95% | 99.97% |

## Compliance & Governance

### Tagging Strategy

All resources tagged with:
```
Environment: production
Project: trading-engine
ManagedBy: terraform
Team: devops
CostCenter: engineering
Compliance: PCI-DSS
```

### Access Control

**Principle of Least Privilege:**
- Developers: Read-only prod access
- DevOps: Admin access with MFA
- DBAs: Database-only access
- Security: Audit logs access

**Access Logging:**
- CloudTrail: All API calls
- EKS audit logs: Enabled
- Database audit: Enabled

### Compliance

**PCI-DSS Requirements:**
- Encryption at rest and in transit
- Network segmentation
- Access logging
- Regular security scans
- Patch management

## Maintenance Windows

**Preferred:**
- Sunday 02:00-06:00 UTC
- Low trading volume
- Automated deployments allowed

**Emergency:**
- Any time with approval
- Notify users 1 hour in advance
- Follow runbook procedures

---

**Document Version:** 1.0
**Last Updated:** 2026-01-18
**Next Review:** 2026-04-18
**Owner:** DevOps Team
