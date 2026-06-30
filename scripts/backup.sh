#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="$(cd "$(dirname "$0")/.." && pwd)/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/phoenixlab_${TIMESTAMP}.sql"

mkdir -p "$BACKUP_DIR"

echo "==> Backing up Phoenix Lab database..."
docker compose exec -T db pg_dump -U tracker -d phoenixlab --no-owner --no-acl > "$BACKUP_FILE"

if [ -s "$BACKUP_FILE" ]; then
    SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "==> Backup complete: $BACKUP_FILE ($SIZE)"
else
    echo "==> ERROR: Backup file is empty — check if the database is running"
    rm -f "$BACKUP_FILE"
    exit 1
fi

cd "$BACKUP_DIR"
ls -1t phoenixlab_*.sql 2>/dev/null | tail -n +31 | xargs -r rm -f
echo "==> Retained latest 30 backups"
