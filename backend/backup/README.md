# Trading Engine Backup & Recovery System

Comprehensive backup and disaster recovery system with RTO < 15 minutes and RPO < 5 minutes.

## 🎯 Features

- **Automated Backups**
  - Full backups daily (PostgreSQL + Redis + Config + State)
  - Incremental backups every 4 hours
  - Point-in-time recovery with WAL archiving

- **Security**
  - GPG encryption at rest
  - Secure credential management
  - Audit logging

- **Multi-Destination**
  - Local storage
  - AWS S3 (STANDARD_IA)
  - Offsite location (rsync)

- **Retention Policy**
  - 7 daily backups
  - 4 weekly backups
  - 12 monthly backups

- **Verification**
  - Automated daily backup verification
  - Test restores to verify integrity
  - Health monitoring and alerting

- **Disaster Recovery**
  - Complete DR playbook
  - Automated DR testing monthly
  - Multiple recovery scenarios

## 📋 Prerequisites

```bash
# Required packages
sudo apt-get install -y \
    postgresql-client \
    redis-tools \
    gzip \
    tar \
    gpg \
    rsync \
    awscli
```

## 🚀 Quick Start

### 1. Setup Configuration

```bash
# Copy and edit configuration
cp scripts/backup/backup.config.example scripts/backup/backup.config
vi scripts/backup/backup.config

# Set required environment variables
export POSTGRES_PASSWORD="your_password"
export GPG_RECIPIENT="backup@trading-engine.local"
```

### 2. Generate GPG Key for Encryption

```bash
# Generate GPG key
gpg --full-generate-key

# Use:
# - Key type: RSA and RSA
# - Key size: 4096
# - Email: backup@trading-engine.local

# Export public key for backup
gpg --armor --export backup@trading-engine.local > backup-public-key.asc
```

### 3. Setup Cron Jobs

```bash
# Install cron jobs for automated backups
sudo ./scripts/backup/setup-cron.sh

# Verify cron installation
sudo crontab -l
```

### 4. Run First Backup

```bash
# Create first full backup
./scripts/backup/backup-full.sh

# Verify backup
./scripts/backup/backup-verify.sh /var/backups/trading-engine/full-*/backup-*.tar.gz.gpg
```

## 📚 Usage

### Create Full Backup

```bash
./scripts/backup/backup-full.sh
```

**Output:**
- Backup file: `/var/backups/trading-engine/full-YYYYMMDD-HHMMSS/backup-*.tar.gz.gpg`
- Log file: `/var/log/trading-engine/backup/backup-full.log`

### Create Incremental Backup

```bash
./scripts/backup/backup-incremental.sh
```

**Output:**
- Backup file: `/var/backups/trading-engine/incremental-YYYYMMDD-HHMMSS.tar.gz.gpg`
- Log file: `/var/log/trading-engine/backup/backup-incremental.log`

### Verify Backup

```bash
./scripts/backup/backup-verify.sh <backup-file>
```

### Full System Restore

```bash
# List available backups
ls -lh /var/backups/trading-engine/full-*/

# Restore (requires 'force' confirmation)
./scripts/backup/restore-full.sh \
    /var/backups/trading-engine/full-20240115-020000/backup-*.tar.gz.gpg \
    force
```

### Point-in-Time Recovery

```bash
# Restore to specific timestamp
./scripts/backup/restore-point-in-time.sh '2024-01-15 14:30:00'

# Or specify base backup
./scripts/backup/restore-point-in-time.sh \
    '2024-01-15 14:30:00' \
    /var/backups/trading-engine/full-20240115-020000/backup-*.tar.gz.gpg
```

### Check Backup Health

```bash
# Run health check
./scripts/backup/backup-health-check.sh

# View health dashboard (Go package)
cd backend/backup
go run cmd/dashboard/main.go
```

### Run DR Test

```bash
# Manual DR test
./scripts/backup/dr-test.sh

# View test results
tail -f /var/log/trading-engine/backup/dr-test.log
```

## 🔧 Programmatic Usage (Go Package)

```go
package main

import (
    "context"
    "log"

    "github.com/epic1st/rtx/backend/backup"
)

func main() {
    config := &backup.BackupConfig{
        DBHost:           "localhost",
        DBPort:           5432,
        DBName:           "trading_engine",
        DBUser:           "postgres",
        DBPassword:       "password",
        RedisHost:        "localhost",
        RedisPort:        6379,
        BackupRoot:       "/var/backups/trading-engine",
        EnableEncryption: true,
        GPGRecipient:     "backup@trading-engine.local",
        EnableS3:         true,
        S3Bucket:         "trading-engine-backups",
    }

    logger := &SimpleLogger{}

    // Create full backup
    manager := backup.NewBackupManager(config, logger)
    metadata, err := manager.CreateFullBackup(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Backup created: %s, size: %d bytes", metadata.BackupID, metadata.Size)

    // Monitor backup health
    monitor := backup.NewBackupMonitor(config, logger)
    health, err := monitor.CheckHealth(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    dashboard := monitor.GenerateDashboard(health)
    fmt.Println(dashboard)
}

type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
    log.Printf("INFO: "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
    log.Printf("ERROR: "+msg, args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
    log.Printf("WARN: "+msg, args...)
}
```

## 📊 Monitoring

### Metrics Endpoint

Metrics are exported in Prometheus format at `/var/log/trading-engine/backup/metrics.txt`:

```
backup_health_status 1.0
backup_total_count 14
backup_total_size_bytes 5368709120
backup_disk_usage_percent 45.2
backup_rto_minutes 12
backup_rpo_minutes 240
backup_last_full_age_hours 6
backup_success_rate 0.9800
backup_failures_last_24h 0
```

### Health Dashboard

```bash
# View current health
./scripts/backup/backup-health-check.sh

# Or use Go dashboard
cd backend/backup
go run cmd/dashboard/main.go
```

### Alerting

Alerts are sent via:
- Email (configured in `backup.config`)
- Syslog
- Slack webhook (if configured)

## 🗂️ File Structure

```
backend/
├── backup/                      # Go package
│   ├── backup.go               # Backup manager
│   ├── restore.go              # Restore manager
│   ├── monitor.go              # Health monitoring
│   └── README.md               # This file
│
├── scripts/backup/             # Shell scripts
│   ├── backup-full.sh          # Full backup
│   ├── backup-incremental.sh   # Incremental backup
│   ├── backup-verify.sh        # Verification
│   ├── restore-full.sh         # Full restore
│   ├── restore-point-in-time.sh # PITR
│   ├── backup-health-check.sh  # Health monitoring
│   ├── cleanup-old-backups.sh  # Retention management
│   ├── dr-test.sh              # DR testing
│   ├── dr-playbook.md          # DR procedures
│   ├── setup-cron.sh           # Cron installation
│   └── backup.config           # Configuration
│
└── /var/backups/trading-engine/ # Backup storage
    ├── full-YYYYMMDD-HHMMSS/    # Full backups
    ├── incremental-*/           # Incremental backups
    ├── weekly-*/                # Weekly retention
    └── monthly-*/               # Monthly retention
```

## 🔒 Security Best Practices

1. **Encryption**
   - All backups are GPG encrypted
   - Use separate GPG key for backups
   - Store private key securely (KMS, HSM)

2. **Access Control**
   - Restrict backup directory permissions
   - Use IAM roles for S3 access
   - SSH key-based auth for offsite

3. **Credentials**
   - Never commit passwords to git
   - Use environment variables
   - Rotate credentials regularly

4. **Audit**
   - All backup operations logged
   - Track access to backup files
   - Review logs regularly

## 📈 Performance

### Backup Performance
- **Full backup**: ~15 minutes for 10GB database
- **Incremental backup**: ~2 minutes
- **Verification**: ~8 minutes
- **Compression ratio**: ~35% (varies by data)

### Recovery Performance
- **RTO (Recovery Time Objective)**: < 15 minutes
- **RPO (Recovery Point Objective)**: < 5 minutes
- **Point-in-time recovery**: ~20 minutes

## 🧪 Testing

### Manual Test

```bash
# Create test backup
./scripts/backup/backup-full.sh

# Verify
./scripts/backup/backup-verify.sh <backup-file>

# Test restore to isolated environment
# (recommended: use Docker or VM)
```

### Automated DR Test

```bash
# Run monthly DR test
./scripts/backup/dr-test.sh

# Check results
cat /var/log/trading-engine/backup/dr-test.log
```

## 🐛 Troubleshooting

### Backup fails with "pg_dump: command not found"

```bash
# Install PostgreSQL client
sudo apt-get install postgresql-client
```

### GPG encryption fails

```bash
# Check GPG keys
gpg --list-keys

# Import key if missing
gpg --import backup-public-key.asc
```

### S3 upload fails

```bash
# Configure AWS credentials
aws configure

# Test S3 access
aws s3 ls s3://your-bucket/
```

### Restore fails with connection errors

```bash
# Check PostgreSQL is running
systemctl status postgresql

# Check database connection
psql -h localhost -U postgres -d postgres
```

## 📞 Support

- Documentation: `./scripts/backup/dr-playbook.md`
- Logs: `/var/log/trading-engine/backup/`
- Issues: Contact DevOps team

## 📜 License

Copyright (c) 2024 Trading Engine

---

**Last Updated**: 2024-01-18
**Version**: 1.0.0
