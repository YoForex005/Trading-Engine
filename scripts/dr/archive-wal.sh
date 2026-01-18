#!/bin/bash
###############################################################################
# RTX Trading Engine - WAL Archive Command
# Purpose: Archive WAL segments for PITR (called by PostgreSQL archive_command)
# Usage: archive_command = '/opt/rtx/scripts/archive-wal.sh %p %f'
###############################################################################

set -euo pipefail

WAL_PATH="$1"
WAL_FILE="$2"
ARCHIVE_DIR="/var/lib/postgresql/14/archive"
S3_BUCKET="${RTX_BACKUP_BUCKET:-rtx-backups}"

# Ensure archive directory exists
mkdir -p "$ARCHIVE_DIR"

# Compress and copy WAL file
gzip -c "$WAL_PATH" > "${ARCHIVE_DIR}/${WAL_FILE}.gz"

# Immediately upload to S3 for real-time archiving
aws s3 cp \
    "${ARCHIVE_DIR}/${WAL_FILE}.gz" \
    "s3://${S3_BUCKET}/wal-archive/${WAL_FILE}.gz" \
    --sse aws:kms \
    --sse-kms-key-id "${RTX_KMS_KEY}" \
    --quiet

# Return 0 for success (PostgreSQL expects this)
exit 0
