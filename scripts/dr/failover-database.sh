#!/bin/bash
###############################################################################
# RTX Trading Engine - Database Failover Script
# Purpose: Promote standby to primary during disaster
# Usage: ./failover-database.sh [--force]
###############################################################################

set -euo pipefail

STANDBY_HOST="${RTX_STANDBY_HOST:-standby1-db}"
PRIMARY_HOST="${RTX_PRIMARY_HOST:-primary-db}"
FORCE_MODE="${1:-}"
LOG_FILE="/var/log/rtx/failover-$(date +%Y%m%d_%H%M%S).log"

mkdir -p "$(dirname "$LOG_FILE")"

log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $1" | tee -a "$LOG_FILE"
}

error_exit() {
    log "ERROR: $1"
    aws sns publish \
        --topic-arn "${RTX_SNS_TOPIC}" \
        --subject "RTX Database Failover FAILED" \
        --message "Failover failed: $1"
    exit 1
}

log "=== Starting Database Failover ==="
log "Primary: $PRIMARY_HOST"
log "Standby: $STANDBY_HOST"

# Step 1: Verify primary is down
log "Step 1/8: Verifying primary database status"
if pg_isready -h "$PRIMARY_HOST" -q && [ "$FORCE_MODE" != "--force" ]; then
    error_exit "Primary database is still accessible. Use --force to override."
fi
log "Primary database confirmed down or force mode enabled"

# Step 2: Check standby status
log "Step 2/8: Checking standby database health"
if ! pg_isready -h "$STANDBY_HOST" -q; then
    error_exit "Standby database is not accessible"
fi

# Check replication lag
LAG=$(ssh admin@"$STANDBY_HOST" "sudo -u postgres psql -t -c \"SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()));\"" | xargs)
log "Current replication lag: ${LAG}s"

if (( $(echo "$LAG > 300" | bc -l) )); then
    log "WARNING: Replication lag is high (${LAG}s). Some data loss may occur."
    if [ "$FORCE_MODE" != "--force" ]; then
        error_exit "Replication lag too high. Use --force to proceed anyway."
    fi
fi

# Step 3: Stop services on all application servers
log "Step 3/8: Stopping RTX services on all application servers"
for APP_SERVER in rtx-app1 rtx-app2 rtx-app3; do
    ssh admin@"$APP_SERVER" "sudo systemctl stop rtx-backend rtx-websocket rtx-fix-gateway" || \
        log "WARNING: Failed to stop services on $APP_SERVER"
done

# Step 4: Wait for in-flight transactions
log "Step 4/8: Waiting for in-flight transactions to complete (30s)"
sleep 30

# Step 5: Promote standby to primary
log "Step 5/8: Promoting standby to primary"
ssh admin@"$STANDBY_HOST" "sudo -u postgres pg_ctl promote -D /var/lib/postgresql/14/main" || \
    error_exit "Failed to promote standby"

# Wait for promotion
log "Waiting for promotion to complete..."
for i in {1..30}; do
    if ssh admin@"$STANDBY_HOST" "sudo -u postgres psql -t -c 'SELECT NOT pg_is_in_recovery();'" | grep -q "t"; then
        log "Promotion successful after ${i}s"
        break
    fi
    if [ $i -eq 30 ]; then
        error_exit "Promotion timeout after 30s"
    fi
    sleep 1
done

# Step 6: Update DNS/Load Balancer
log "Step 6/8: Updating DNS to point to new primary"
aws route53 change-resource-record-sets \
    --hosted-zone-id "${RTX_HOSTED_ZONE_ID}" \
    --change-batch "{
        \"Changes\": [{
            \"Action\": \"UPSERT\",
            \"ResourceRecordSet\": {
                \"Name\": \"db.rtx-trading.com\",
                \"Type\": \"CNAME\",
                \"TTL\": 60,
                \"ResourceRecords\": [{\"Value\": \"$STANDBY_HOST\"}]
            }
        }]
    }" || error_exit "DNS update failed"

log "DNS updated - TTL 60s means clients will switch within 1 minute"

# Step 7: Update application configuration
log "Step 7/8: Updating application database configuration"
for APP_SERVER in rtx-app1 rtx-app2 rtx-app3; do
    ssh admin@"$APP_SERVER" "sed -i 's/$PRIMARY_HOST/$STANDBY_HOST/g' /opt/rtx/.env" || \
        log "WARNING: Failed to update config on $APP_SERVER"
done

# Step 8: Restart services
log "Step 8/8: Restarting RTX services"
for APP_SERVER in rtx-app1 rtx-app2 rtx-app3; do
    ssh admin@"$APP_SERVER" "sudo systemctl start rtx-backend" || \
        log "WARNING: Failed to start rtx-backend on $APP_SERVER"
done

# Wait and verify
log "Waiting 10s for services to stabilize..."
sleep 10

# Health check
log "Performing health checks..."
HEALTHY=0
for APP_SERVER in rtx-app1 rtx-app2 rtx-app3; do
    if ssh admin@"$APP_SERVER" "curl -f http://localhost:7999/health" 2>/dev/null; then
        log "✓ Health check passed on $APP_SERVER"
        HEALTHY=$((HEALTHY + 1))
    else
        log "✗ Health check failed on $APP_SERVER"
    fi
done

if [ $HEALTHY -eq 0 ]; then
    error_exit "All health checks failed after failover"
fi

# Record failover event
log "Recording failover event in database"
psql -h "$STANDBY_HOST" -U postgres -d rtx -c "
INSERT INTO failover_history (
    failover_date, from_host, to_host, replication_lag_seconds, services_restarted, status
) VALUES (
    NOW(), '$PRIMARY_HOST', '$STANDBY_HOST', $LAG, $HEALTHY, 'completed'
);" || log "WARNING: Failed to record failover event"

# Success notification
log "=== Failover Completed Successfully ==="
log "New primary: $STANDBY_HOST"
log "Services healthy: $HEALTHY/3"
log "Replication lag at failover: ${LAG}s"

aws sns publish \
    --topic-arn "${RTX_SNS_TOPIC}" \
    --subject "RTX Database Failover SUCCESS" \
    --message "Database failover completed. New primary: $STANDBY_HOST. Healthy services: $HEALTHY/3. Lag: ${LAG}s"

log ""
log "NEXT STEPS:"
log "1. Investigate root cause of primary failure"
log "2. Rebuild failed primary as new standby"
log "3. Monitor new primary for 24 hours"
log "4. Update documentation and runbooks"

exit 0
