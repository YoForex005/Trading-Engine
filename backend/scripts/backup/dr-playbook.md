# Disaster Recovery Playbook

## 🚨 Emergency Response Team

### Primary Contacts
- **DBA Lead**: [Contact Info]
- **DevOps Lead**: [Contact Info]
- **CTO/Engineering Manager**: [Contact Info]
- **Security Officer**: [Contact Info]

### Escalation Chain
1. On-call Engineer → DevOps Lead (15 min)
2. DevOps Lead → DBA Lead + CTO (30 min)
3. CTO → Executive Team (1 hour)

## 📋 Disaster Scenarios

### Scenario 1: Database Corruption
**RTO**: 15 minutes | **RPO**: 5 minutes

**Symptoms**:
- PostgreSQL crashes
- Data integrity errors
- Replication lag spikes
- Application 500 errors

**Recovery Steps**:
```bash
# 1. Assess damage
psql -d trading_engine -c "SELECT * FROM pg_stat_database;"

# 2. Stop application
systemctl stop trading-engine

# 3. Identify last known good backup
cd /var/backups/trading-engine
ls -lh full-* | tail -5

# 4. Restore from latest backup
/opt/trading-engine/scripts/backup/restore-full.sh \
    /var/backups/trading-engine/full-YYYYMMDD-HHMMSS/backup-*.tar.gz.gpg force

# 5. Verify restore
psql -d trading_engine -c "SELECT COUNT(*) FROM users;"
psql -d trading_engine -c "SELECT MAX(created_at) FROM transactions;"

# 6. Start application
systemctl start trading-engine

# 7. Monitor logs
tail -f /var/log/trading-engine/application.log
```

**Verification Checklist**:
- [ ] Database accessible
- [ ] Table counts match expected
- [ ] Latest transactions present
- [ ] No error logs
- [ ] Application health check passing
- [ ] User login working
- [ ] Trading functionality operational

---

### Scenario 2: Data Center Failure
**RTO**: 30 minutes | **RPO**: 5 minutes

**Symptoms**:
- Complete loss of primary datacenter
- Network connectivity lost
- All services unavailable

**Recovery Steps**:
```bash
# FAILOVER TO DR SITE

# 1. Verify primary is down
ping primary-db.trading-engine.com
curl https://api.trading-engine.com/health

# 2. Activate DR site DNS
# Update DNS to point to DR site (automated via AWS Route53 health checks)
aws route53 change-resource-record-sets \
    --hosted-zone-id Z1234567 \
    --change-batch file://dr-dns-failover.json

# 3. Restore latest backup to DR database
ssh dr-host.trading-engine.com
cd /opt/trading-engine/scripts/backup

# Download latest backup from S3
aws s3 sync s3://trading-engine-backups/backups/full/ ./latest/ \
    --exclude "*" --include "backup-$(date +%Y%m%d)*"

# Restore
./restore-full.sh ./latest/backup-*.tar.gz.gpg force

# 4. Apply WAL files for point-in-time recovery
./restore-point-in-time.sh "$(date -u +'%Y-%m-%d %H:%M:%S' -d '5 minutes ago')"

# 5. Start services
systemctl start redis
systemctl start trading-engine
systemctl start nginx

# 6. Verify
curl https://dr.trading-engine.com/health
psql -d trading_engine -c "SELECT version();"
```

**Verification Checklist**:
- [ ] DR site DNS resolving
- [ ] Database operational
- [ ] Redis cache running
- [ ] Application started
- [ ] API endpoints responding
- [ ] WebSocket connections working
- [ ] No data loss detected
- [ ] Users notified of temporary DR activation

---

### Scenario 3: Ransomware Attack
**RTO**: 60 minutes | **RPO**: 24 hours (pre-attack)

**Symptoms**:
- Encrypted files
- Ransom note present
- Unusual file modifications
- Database access denied

**Recovery Steps**:
```bash
# IMMEDIATE ACTIONS

# 1. Isolate infected systems
# Disconnect from network IMMEDIATELY
ip link set eth0 down
systemctl stop trading-engine
systemctl stop postgresql

# 2. Preserve evidence
dd if=/dev/sda of=/mnt/forensics/disk-image.dd bs=4M
tar -czf /mnt/forensics/logs-$(date +%s).tar.gz /var/log/

# 3. Notify security team
mail -s "CRITICAL: Ransomware Detected" security@trading-engine.com < incident-report.txt

# 4. Identify last clean backup (before attack)
# Review backup verification logs
grep -r "Verification completed" /var/log/trading-engine/backup/backup-verify.log

# 5. Build clean environment
# Provision new server OR wipe existing
# DO NOT RESTORE TO INFECTED SYSTEM

# 6. Restore from clean backup
./restore-full.sh /var/backups/offsite/backup-CLEAN-TIMESTAMP.tar.gz.gpg force

# 7. Change all credentials
psql -d trading_engine <<EOF
UPDATE users SET password_hash = NULL WHERE 1=1;
-- Force password resets
EOF

# Rotate API keys
./scripts/rotate-api-keys.sh

# 8. Security scan
clamscan -r /opt/trading-engine
rkhunter --check --skip-keypress

# 9. Gradual restoration
# Start with read-only mode
# Enable trading after security clearance
```

**Post-Recovery Actions**:
- [ ] Full security audit
- [ ] Patch all systems
- [ ] Review access logs
- [ ] Notify affected users
- [ ] File incident report
- [ ] Update security policies
- [ ] Conduct post-mortem

---

### Scenario 4: Accidental Data Deletion
**RTO**: 10 minutes | **RPO**: 5 minutes

**Symptoms**:
- Missing records
- User reports data loss
- Audit logs show DELETE operations

**Recovery Steps**:
```bash
# POINT-IN-TIME RECOVERY

# 1. Identify deletion time
psql -d trading_engine -c \
    "SELECT NOW() - interval '5 minutes' AS recovery_target;"

# 2. Stop writes (read-only mode)
psql -d trading_engine -c "ALTER DATABASE trading_engine SET default_transaction_read_only = on;"

# 3. Create snapshot of current state
pg_dump -d trading_engine -f /tmp/pre-restore-$(date +%s).sql

# 4. Perform point-in-time recovery
./restore-point-in-time.sh "2024-01-15 14:25:00"

# 5. Verify recovered data
psql -d trading_engine -c "SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL;"

# 6. Re-enable writes
psql -d trading_engine -c "ALTER DATABASE trading_engine SET default_transaction_read_only = off;"
```

---

### Scenario 5: Hardware Failure
**RTO**: 20 minutes | **RPO**: 5 minutes

**Recovery Steps**:
```bash
# 1. Provision replacement hardware
# Use automated provisioning or manual setup

# 2. Install base OS and dependencies
apt-get update && apt-get install -y postgresql-14 redis-server nginx

# 3. Restore application code
git clone https://github.com/trading-engine/backend.git
cd backend
make install

# 4. Restore latest backup
./scripts/backup/restore-full.sh \
    s3://trading-engine-backups/backups/full/latest.tar.gz.gpg force

# 5. Update DNS/Load Balancer
# Point traffic to new server

# 6. Monitor
watch -n 1 'systemctl status trading-engine postgresql redis'
```

---

## 🧪 DR Testing Schedule

### Monthly DR Test (First Sunday)
```bash
# Automated DR test
/opt/trading-engine/scripts/backup/dr-test.sh
```

**Test Objectives**:
1. Verify backup integrity
2. Measure restore time (RTO)
3. Validate data completeness (RPO)
4. Test failover procedures
5. Verify team readiness

### Quarterly Full DR Drill
- Complete datacenter failover simulation
- All team members participate
- Document lessons learned
- Update playbook

---

## 📊 Recovery Metrics

### RTO (Recovery Time Objective)
- **Target**: < 15 minutes
- **Measured**: From incident detection to service restoration

### RPO (Recovery Point Objective)
- **Target**: < 5 minutes
- **Measured**: Maximum acceptable data loss

### Success Criteria
- [ ] Application restored within RTO
- [ ] Data loss within RPO
- [ ] All critical functions operational
- [ ] No security compromises
- [ ] Users notified appropriately

---

## 📝 Communication Templates

### Internal Alert (Slack/Email)
```
INCIDENT: [Severity] - [Brief Description]
IMPACT: [Affected systems/users]
STATUS: [Investigating/Restoring/Resolved]
ETA: [Expected restoration time]
ACTIONS: [What we're doing]
UPDATES: [Every 15 minutes]
```

### Customer Notification
```
Subject: Service Incident - [Date/Time]

We are currently experiencing [brief description].
Our team is actively working to restore service.

Estimated Resolution: [Time]
Affected Services: [List]
Data Safety: Your data is secure and backed up.

We will update you every 30 minutes.

Status Page: https://status.trading-engine.com
```

---

## 🔧 Pre-Disaster Checklist

### Weekly
- [ ] Verify backup completion
- [ ] Test backup restoration (sample)
- [ ] Check disk space
- [ ] Review monitoring alerts

### Monthly
- [ ] Full DR test
- [ ] Update contact list
- [ ] Review and update playbook
- [ ] Audit access controls

### Quarterly
- [ ] Full DR drill
- [ ] Security audit
- [ ] Capacity planning review
- [ ] Update SLAs

---

## 📞 External Contacts

### Vendors
- **AWS Support**: 1-800-XXX-XXXX (Premium Support)
- **Database Consultant**: [Contact]
- **Security Firm**: [Contact]

### Regulatory
- **Financial Regulator**: [Contact if trading data affected]
- **Data Protection Officer**: [For GDPR incidents]

---

## 📚 Additional Resources

- Backup Verification Reports: `/var/log/trading-engine/backup/`
- Backup Inventory: `/var/backups/trading-engine/`
- S3 Backups: `s3://trading-engine-backups/`
- Monitoring Dashboard: `https://grafana.trading-engine.com/d/backups`
- Runbook Repository: `https://wiki.trading-engine.com/runbooks`

---

**Last Updated**: [Date]
**Next Review**: [Date + 3 months]
**Owner**: DevOps Team
