#!/bin/bash
# Run backup verification on the latest backup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
fi

BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"

echo "Finding latest backup to verify..."

# Find latest full backup
LATEST_BACKUP=$(find "$BACKUP_ROOT" -name "full-*/backup-*.tar.gz*" -type f | sort -r | head -1)

if [[ -z "$LATEST_BACKUP" ]]; then
    echo "ERROR: No backups found to verify"
    exit 1
fi

echo "Latest backup: $LATEST_BACKUP"
echo "Starting verification..."

"$SCRIPT_DIR/backup-verify.sh" "$LATEST_BACKUP"

exit $?
