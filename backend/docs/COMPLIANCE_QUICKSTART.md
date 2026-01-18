# Compliance System Quick Start Guide

## 5-Minute Setup

### 1. Run Database Migration

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go run cmd/migrate/main.go up
```

This creates all compliance tables including:
- Audit log with tamper-proof hashing
- Best execution report tables
- Order routing report tables
- Compliance alerts and rules

### 2. Configure Environment

```bash
# .env file
COMPLIANCE_ENABLED=true
AUDIT_RETENTION_YEARS=7
COMPLIANCE_ADMIN_ONLY=true
COMPLIANCE_MIFID_II=true
COMPLIANCE_SEC_RULE_606=true
```

### 3. Start Server

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go run cmd/server/main.go
```

### 4. Test Endpoints

```bash
# Get admin token first
TOKEN=$(curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Best Execution Report (MiFID II)
curl "http://localhost:7999/api/compliance/best-execution?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=json" \
  -H "Authorization: Bearer $TOKEN" | jq

# Order Routing Report (SEC Rule 606)
curl "http://localhost:7999/api/compliance/order-routing?quarter=Q1&year=2026&format=json" \
  -H "Authorization: Bearer $TOKEN" | jq

# Audit Trail
curl "http://localhost:7999/api/compliance/audit-trail?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=json" \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Key Features

### Tamper-Proof Audit Trail

Every audit entry contains:
- SHA-256 hash of the entry
- Hash of previous entry (blockchain pattern)
- Automatic chain verification

```sql
-- Verify audit chain integrity
SELECT * FROM verify_audit_chain('orders', 1000);
```

### Automatic Audit Logging

All database changes are automatically audited via triggers:
- User inserts, updates, deletes
- Order lifecycle events
- Position changes
- Account modifications

### 7-Year Retention

Audit logs are retained for 7 years as required by:
- MiFID II
- SEC regulations
- Dodd-Frank Act

Automatic archival to cold storage after retention period.

## CSV Export Examples

### Best Execution Report CSV

```bash
curl "http://localhost:7999/api/compliance/best-execution?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o best_execution_report.csv
```

### Order Routing Report CSV

```bash
curl "http://localhost:7999/api/compliance/order-routing?quarter=Q1&year=2026&format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o order_routing_Q1_2026.csv
```

### Audit Trail CSV

```bash
curl "http://localhost:7999/api/compliance/audit-trail?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&entity_type=order&format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit_trail.csv
```

## Compliance Alerts

View open alerts:

```sql
SELECT * FROM v_open_compliance_alerts;
```

Configure alert rules:

```sql
INSERT INTO compliance_alert_rules (rule_name, rule_type, alert_severity, threshold_config) VALUES
  ('Custom Alert', 'execution_quality', 'warning', '{"fill_rate_threshold": 90.0}');
```

## Regulatory Filing

Track regulatory submissions:

```sql
INSERT INTO regulatory_filings (filing_type, filing_period, regulator, filing_deadline)
VALUES ('RTS 27 Best Execution', 'Q1 2026', 'ESMA', '2026-04-30 23:59:59');
```

## Testing

Run compliance tests:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go test ./internal/api/handlers -run TestCompliance -v
```

## Production Checklist

- [ ] Database migrations applied
- [ ] Environment variables configured
- [ ] Admin authentication enabled
- [ ] Archive storage configured (S3/Glacier)
- [ ] Scheduled cleanup jobs configured
- [ ] Alert emails configured
- [ ] Tamper-proof logging verified
- [ ] Chain verification tested
- [ ] Backup strategy documented
- [ ] Disaster recovery tested

## Common Use Cases

### Generate Monthly Best Execution Report

```bash
# First day of next month
MONTH_START="2026-01-01T00:00:00Z"
MONTH_END="2026-01-31T23:59:59Z"

curl "http://localhost:7999/api/compliance/best-execution?start_time=$MONTH_START&end_time=$MONTH_END&format=pdf" \
  -H "Authorization: Bearer $TOKEN" \
  -o best_execution_january_2026.pdf
```

### Generate Quarterly Order Routing Report

```bash
# On regulatory deadline
curl "http://localhost:7999/api/compliance/order-routing?quarter=Q1&year=2026&format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o sec_606_Q1_2026.csv
```

### Export Audit Trail for Investigation

```bash
# Specific time period and entity type
curl "http://localhost:7999/api/compliance/audit-trail?start_time=2026-01-15T00:00:00Z&end_time=2026-01-15T23:59:59Z&entity_type=order&format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit_investigation_jan15.csv
```

### Manual Audit Entry

```bash
# Record manual administrative action
curl -X POST http://localhost:7999/api/compliance/audit-log \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin-001",
    "action": "MANUAL_ADJUSTMENT",
    "entity_type": "account",
    "entity_id": "account-123",
    "details": {
      "reason": "Customer service credit",
      "amount": 100.00,
      "approved_by": "supervisor-456"
    }
  }'
```

## Troubleshooting

### Issue: Reports show no data

**Solution**: Check date range and ensure orders exist in that period

```sql
SELECT COUNT(*) FROM orders WHERE created_at BETWEEN '2026-01-01' AND '2026-01-31';
```

### Issue: Audit chain verification fails

**Solution**: Investigate tampering or database corruption

```sql
SELECT * FROM verify_audit_chain('orders', 1000) WHERE is_valid = false;
```

### Issue: CSV export is empty

**Solution**: Verify admin authentication and permissions

```bash
# Check token validity
curl http://localhost:7999/api/account/summary -H "Authorization: Bearer $TOKEN"
```

## Support Resources

- Full Documentation: [COMPLIANCE_SYSTEM.md](./COMPLIANCE_SYSTEM.md)
- API Reference: [API.md](./API.md)
- Database Schema: [migrations/008_add_compliance_reporting.sql](../migrations/008_add_compliance_reporting.sql)
- Test Suite: [internal/api/handlers/compliance_test.go](../internal/api/handlers/compliance_test.go)

## Next Steps

1. Review full documentation in `COMPLIANCE_SYSTEM.md`
2. Configure scheduled jobs for report generation
3. Set up S3/Glacier for long-term archival
4. Configure email notifications for alerts
5. Test disaster recovery procedures
6. Document compliance procedures for your organization
