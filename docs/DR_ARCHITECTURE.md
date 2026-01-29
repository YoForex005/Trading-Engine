# RTX Trading Engine - DR/BC Architecture

This document provides visual representations of the disaster recovery and business continuity architecture.

---

## 1. Overall DR Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        PRIMARY REGION (us-east-1)                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐       ┌──────────────┐       ┌──────────────┐       │
│  │   RTX App 1  │       │   RTX App 2  │       │   RTX App 3  │       │
│  │  (Backend)   │       │  (Backend)   │       │  (Backend)   │       │
│  └──────┬───────┘       └──────┬───────┘       └──────┬───────┘       │
│         │                      │                      │                │
│         └──────────────────────┼──────────────────────┘                │
│                                │                                        │
│                      ┌─────────▼─────────┐                             │
│                      │   Load Balancer   │                             │
│                      │     (HAProxy)     │                             │
│                      └─────────┬─────────┘                             │
│                                │                                        │
│         ┌──────────────────────┼──────────────────────┐                │
│         │                      │                      │                │
│    ┌────▼────┐          ┌─────▼──────┐        ┌─────▼──────┐         │
│    │ Primary │          │  Standby 1 │        │  Standby 2 │         │
│    │   DB    │──Stream──│     DB     │──Async─│     DB     │         │
│    │  (RW)   │   Sync   │    (RO)    │  Repl  │    (RO)    │         │
│    └────┬────┘          └────────────┘        └────────────┘         │
│         │                                                              │
│         │ WAL                                                          │
│    ┌────▼──────────────────────────────────────────────────────────┐  │
│    │                      S3 Bucket                                │  │
│    │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │  │
│    │  │ Full Backups │  │ WAL Archives │  │ Config Files │       │  │
│    │  │ (Daily)      │  │ (Real-time)  │  │ (12h)        │       │  │
│    │  └──────────────┘  └──────────────┘  └──────────────┘       │  │
│    │                     KMS Encrypted                            │  │
│    └───────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                │
                                │ Cross-Region Replication
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       SECONDARY REGION (us-west-2)                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐       ┌──────────────┐                               │
│  │   RTX App 1  │       │   RTX App 2  │                               │
│  │  (Standby)   │       │  (Standby)   │                               │
│  └──────────────┘       └──────────────┘                               │
│                                │                                        │
│                      ┌─────────▼─────────┐                             │
│                      │   Load Balancer   │                             │
│                      │     (HAProxy)     │                             │
│                      └─────────┬─────────┘                             │
│                                │                                        │
│                         ┌──────▼──────┐                                │
│                         │  Secondary  │                                │
│                         │     DB      │◄────WAL from S3                │
│                         │  (Standby)  │                                │
│                         └─────────────┘                                │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Backup Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            BACKUP PIPELINE                              │
└─────────────────────────────────────────────────────────────────────────┘

 02:00 UTC Daily              Every 6 Hours              Continuous
     │                             │                         │
     ▼                             ▼                         ▼
┌──────────┐                ┌──────────┐              ┌──────────┐
│   Full   │                │   WAL    │              │   WAL    │
│  Backup  │                │   Sync   │              │  Stream  │
│          │                │          │              │          │
│ pg_dump  │                │  rsync   │              │ archive_ │
│ --format │                │ /archive │              │ command  │
│  custom  │                │   to S3  │              │          │
└────┬─────┘                └────┬─────┘              └────┬─────┘
     │                           │                         │
     │ Compress (gzip -6)        │                         │
     ▼                           ▼                         ▼
┌──────────┐                ┌──────────┐              ┌──────────┐
│ Generate │                │ Compress │              │ Compress │
│ Checksum │                │   .gz    │              │   .gz    │
│ SHA256   │                │          │              │          │
└────┬─────┘                └────┬─────┘              └────┬─────┘
     │                           │                         │
     │ Encrypt (KMS)             │ Encrypt (KMS)           │ Encrypt (KMS)
     ▼                           ▼                         ▼
┌──────────────────────────────────────────────────────────────────┐
│                         S3 Bucket (rtx-backups)                  │
├──────────────────────────────────────────────────────────────────┤
│  /backups/full/         /wal-archive/         /config/          │
│   YYYYMMDD/             *.gz                   YYYYMMDD/        │
│                                                                  │
│  Lifecycle:             Lifecycle:             Lifecycle:       │
│  • 30d → Glacier        • 7d → Delete          • 30d → Glacier  │
│  • 365d → Delete        • Retention: 7d        • 90d → Delete   │
└──────────────────────────────────────────────────────────────────┘
     │
     │ Verification (03:00 UTC)
     ▼
┌──────────┐
│  Verify  │
│ Download │
│ Checksum │
│ pg_restore│
│  --list  │
└────┬─────┘
     │
     │ Weekly Test (Sunday 04:00 UTC)
     ▼
┌──────────┐
│  Test    │
│ Restore  │
│ to Stage │
│ Validate │
│  Data    │
└──────────┘
```

---

## 3. Restore Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           RESTORE PIPELINE                              │
└─────────────────────────────────────────────────────────────────────────┘

User Initiates Restore: ./restore-full.sh [date]
     │
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 1: Locate Backup                                           │
│  • Query S3 for backup file                                     │
│  • Select specific date or "latest"                             │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 2: Download from S3                                        │
│  • aws s3 cp s3://rtx-backups/backups/full/...                  │
│  • Download checksum file                                       │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 3: Verify Integrity                                        │
│  • SHA256 checksum validation                                   │
│  • pg_restore --list (structure check)                          │
│  ✓ Exit if --verify-only                                        │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 4: Stop Services                                           │
│  • systemctl stop rtx-backend                                   │
│  • systemctl stop rtx-websocket                                 │
│  • systemctl stop rtx-fix-gateway                               │
│  • Wait 5s for graceful shutdown                                │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 5: Terminate Connections                                   │
│  • SELECT pg_terminate_backend(...)                             │
│  • DROP DATABASE rtx                                            │
│  • CREATE DATABASE rtx                                          │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 6: Restore Data                                            │
│  • pg_restore --jobs=4 (parallel)                               │
│  • Duration: 5-15 minutes (depends on size)                     │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 7: Post-Restore                                            │
│  • REINDEX DATABASE                                             │
│  • ANALYZE (update statistics)                                  │
│  • Verify record counts                                         │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 8: Restart Services                                        │
│  • systemctl start rtx-backend                                  │
│  • curl http://localhost:7999/health                            │
│  • ✓ Success or ✗ Failure (rollback needed)                    │
└──────────────────────────────────────────────────────────────────┘

Total RTO: ~15 minutes
```

---

## 4. Failover Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         DATABASE FAILOVER                               │
└─────────────────────────────────────────────────────────────────────────┘

Primary DB Failure Detected
     │
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Alert: DatabaseDown                                             │
│  • Prometheus detects: up{job="postgresql"} == 0                │
│  • SNS → PagerDuty → On-call engineer                           │
│  • Auto-trigger: ./failover-database.sh                         │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 1: Verify Primary Down                                     │
│  • pg_isready -h primary-db (timeout)                           │
│  • Check Patroni status                                         │
│  • Confirm no split-brain                                       │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 2: Check Standby Health                                    │
│  • pg_isready -h standby1-db                                    │
│  • Check replication lag                                        │
│  • Abort if lag > 300s (unless --force)                         │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 3: Stop All Services                                       │
│  • ssh rtx-app1 "systemctl stop rtx-*"                          │
│  • ssh rtx-app2 "systemctl stop rtx-*"                          │
│  • ssh rtx-app3 "systemctl stop rtx-*"                          │
│  • Wait 30s for in-flight transactions                          │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 4: Promote Standby                                         │
│  • ssh standby1 "sudo -u postgres pg_ctl promote"               │
│  • Wait for pg_is_in_recovery() = false                         │
│  • Verify: SELECT NOT pg_is_in_recovery();                      │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 5: Update DNS                                              │
│  • Route53: db.rtx-trading.com → standby1-db                    │
│  • TTL: 60s (propagation: ~1 minute)                            │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 6: Update App Config                                       │
│  • ssh rtx-app* "sed -i 's/primary/standby1/' .env"             │
│  • Verify connection strings updated                            │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 7: Restart Services                                        │
│  • ssh rtx-app* "systemctl start rtx-backend"                   │
│  • Wait 10s for startup                                         │
│  • curl http://rtx-app*/health                                  │
└────┬─────────────────────────────────────────────────────────────┘
     ▼
┌──────────────────────────────────────────────────────────────────┐
│ Step 8: Verify                                                  │
│  • Check 3/3 apps healthy                                       │
│  • Verify order execution                                       │
│  • Monitor for 15 minutes                                       │
│  ✓ Success: Record in failover_history                          │
└──────────────────────────────────────────────────────────────────┘

Total Failover Time: 9-15 minutes
```

---

## 5. Monitoring & Alerting Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        MONITORING ARCHITECTURE                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                         DATA SOURCES                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │PostgreSQL│  │   RTX    │  │  Redis   │  │  Node    │               │
│  │ Exporter │  │ Backend  │  │ Exporter │  │ Exporter │               │
│  │ :9187    │  │  :7999   │  │  :9121   │  │  :9100   │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
│       │             │              │             │                      │
│       └─────────────┴──────────────┴─────────────┘                      │
│                             │                                           │
│                             ▼                                           │
│                    ┌─────────────────┐                                  │
│                    │   Prometheus    │                                  │
│                    │    :9090        │                                  │
│                    │                 │                                  │
│                    │ • 42 alert rules│                                  │
│                    │ • 15s scrape    │                                  │
│                    │ • 15d retention │                                  │
│                    └────┬────────────┘                                  │
│                         │                                               │
│         ┌───────────────┼───────────────┐                               │
│         │               │               │                               │
│         ▼               ▼               ▼                               │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                            │
│  │ Grafana  │   │AlertMgr  │   │CloudWatch│                            │
│  │ :3000    │   │ :9093    │   │  Metrics │                            │
│  │          │   │          │   │          │                            │
│  │ Dashboards   │ Routing  │   │ Backup   │                            │
│  │          │   │          │   │ Metrics  │                            │
│  └──────────┘   └────┬─────┘   └──────────┘                            │
│                      │                                                  │
│         ┌────────────┼────────────┐                                     │
│         │            │            │                                     │
│         ▼            ▼            ▼                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                                │
│  │PagerDuty │ │  Slack   │ │  Email   │                                │
│  │ Critical │ │ Warnings │ │   Info   │                                │
│  └──────────┘ └──────────┘ └──────────┘                                │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

Alert Severity Matrix:
┌──────────┬────────────┬─────────────┬─────────────┐
│ Severity │   Channel  │  Response   │   Examples  │
├──────────┼────────────┼─────────────┼─────────────┤
│ Critical │ PagerDuty  │   5 min     │ DB down     │
│          │   + SMS    │             │ API down    │
├──────────┼────────────┼─────────────┼─────────────┤
│ Warning  │   Slack    │   15 min    │ High lag    │
│          │            │             │ Disk 85%    │
├──────────┼────────────┼─────────────┼─────────────┤
│ Info     │   Email    │  Next day   │ Volume spike│
│          │            │             │ New feature │
└──────────┴────────────┴─────────────┴─────────────┘
```

---

## 6. Health Check System

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    HEALTH CHECK ARCHITECTURE                            │
└─────────────────────────────────────────────────────────────────────────┘

Cron: * * * * * /opt/rtx/scripts/dr/health-check.sh
  │
  ├──► Check 1: Database
  │    │ • pg_isready -h localhost
  │    │ • Query time: SELECT COUNT(*) FROM rtx_positions
  │    │ • Result: {"status":"pass","duration_ms":45}
  │
  ├──► Check 2: API
  │    │ • curl http://localhost:7999/health
  │    │ • Expected: HTTP 200
  │    │ • Result: {"status":"pass","http_code":200}
  │
  ├──► Check 3: WebSocket
  │    │ • wscat -c ws://localhost:7999/ws
  │    │ • Ping/Pong test
  │    │ • Result: {"status":"pass","duration_ms":10}
  │
  ├──► Check 4: Redis
  │    │ • redis-cli ping
  │    │ • Expected: PONG
  │    │ • Result: {"status":"pass"}
  │
  ├──► Check 5: Disk Space
  │    │ • df -h / | awk '{print $5}'
  │    │ • Threshold: 85%
  │    │ • Result: {"status":"pass","usage_percent":62}
  │
  ├──► Check 6: Replication Lag
  │    │ • SELECT pg_last_xact_replay_timestamp()
  │    │ • Threshold: 60s
  │    │ • Result: {"status":"pass","lag_seconds":2.3}
  │
  ├──► Check 7: CPU Load
  │    │ • top -bn1 | grep Cpu
  │    │ • Threshold: 80%
  │    │ • Result: {"status":"pass","load_percent":35}
  │
  └──► Check 8: Memory
       │ • free | grep Mem
       │ • Threshold: 90%
       │ • Result: {"status":"pass","usage_percent":72}
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ Aggregate Results                                               │
│  {                                                               │
│    "timestamp": "2026-01-18T15:30:00Z",                          │
│    "overall_status": "healthy",                                 │
│    "failed_checks": [],                                         │
│    "checks": [...]                                              │
│  }                                                               │
└────┬─────────────────────────────────────────────────────────────┘
     │
     ├──► Write to: /var/lib/rtx/metrics/health.json
     ├──► Insert to: health_check_history table
     └──► Send to: CloudWatch (RTX/Health/FailedChecks)
          │
          ▼
     If failures > 0:
       └──► SNS Alert: "Health check failed: [components]"
```

---

## 7. Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                 │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ Layer 1: Network Security                                              │
├─────────────────────────────────────────────────────────────────────────┤
│  • VPC with private subnets                                             │
│  • Security groups: Allow only necessary ports                          │
│  • NACLs: Additional network filtering                                  │
│  • No public internet access to databases                               │
└─────────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ Layer 2: Access Control                                                │
├─────────────────────────────────────────────────────────────────────────┤
│  • IAM roles for AWS resources (no hardcoded keys)                      │
│  • PostgreSQL role-based access control                                 │
│  • SSH key-based authentication (no passwords)                          │
│  • MFA required for admin operations                                    │
└─────────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ Layer 3: Encryption                                                    │
├─────────────────────────────────────────────────────────────────────────┤
│  • Data at rest: All backups encrypted with AWS KMS                     │
│  • Data in transit: TLS 1.3 for all connections                         │
│  • Database: PostgreSQL SSL mode required                               │
│  • EBS volumes: Encrypted                                               │
└─────────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ Layer 4: Audit & Logging                                               │
├─────────────────────────────────────────────────────────────────────────┤
│  • CloudTrail: All AWS API calls                                        │
│  • PostgreSQL: Query logging enabled                                    │
│  • Application: Audit trail in rtx_ledger                               │
│  • DR operations: Logged to backup_history, failover_history            │
└─────────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ Layer 5: Secrets Management                                            │
├─────────────────────────────────────────────────────────────────────────┤
│  • AWS Secrets Manager: API keys, passwords                             │
│  • Environment variables: Local config (.env)                           │
│  • PostgreSQL .pgpass: Database credentials                             │
│  • No secrets in code repositories                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Compliance Mapping

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      COMPLIANCE REQUIREMENTS                            │
└─────────────────────────────────────────────────────────────────────────┘

SOC 2 Type II                     Implementation
├─ CC6.1: Backup procedures   → Daily full + hourly WAL
├─ CC6.2: Recovery testing    → Weekly automated restore test
├─ CC7.1: Monitoring          → Prometheus + 42 alert rules
├─ CC7.2: Incident response   → Runbooks + escalation procedures
└─ CC9.1: Risk assessment     → Quarterly DR drills

ISO 27001                         Implementation
├─ A.12.3: Backup             → Automated backups with verification
├─ A.17.1: DR planning        → 80-page DR plan document
├─ A.17.2: Redundancy         → Multi-AZ + multi-region
└─ A.18.1: Compliance         → Audit logs + reporting

PCI DSS                           Implementation
├─ Req 9.5: Data backup       → Daily full + continuous WAL
├─ Req 10.5: Audit logs       → CloudTrail + PostgreSQL logs
├─ Req 12.10: Incident resp   → Runbooks + contact card
└─ Req 12.10.4: Testing       → Quarterly DR drills

MiFID II                          Implementation
├─ Art 16(5): Resilience      → HA architecture + failover
├─ RTS 6: BC testing          → Monthly partial + quarterly full
└─ RTS 7: Recovery            → RTO <15min, RPO <1min

GDPR                              Implementation
├─ Art 32: Security           → Encryption + access control
├─ Art 32(1)(b): Resilience   → Multi-region + redundancy
├─ Art 32(1)(c): Recovery     → Automated restore procedures
└─ Art 33: Breach notification → SNS alerts + incident response
```

---

## Summary

This DR/BC architecture provides:

- **Multi-layered redundancy** (database, application, geographic)
- **Automated recovery** (backups, failover, monitoring)
- **Comprehensive monitoring** (42 alerts, 8 health checks)
- **Security-first design** (encryption, access control, audit)
- **Regulatory compliance** (SOC 2, ISO 27001, PCI DSS, MiFID II, GDPR)

**Key Metrics:**
- RTO: 5-15 minutes
- RPO: 1 minute
- Availability: 99.95%
- Backup retention: 7 days hot, 1 year cold
- Testing: Weekly automated, quarterly manual

---

**Last Updated:** 2026-01-18
**Version:** 1.0
