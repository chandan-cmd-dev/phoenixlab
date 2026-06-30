#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

echo "============================================"
echo "  Phoenix Lab — Safe Deploy"
echo "============================================"
echo ""

echo "==> Step 1: Creating pre-deploy database backup..."
if docker compose ps db --status running -q 2>/dev/null | grep -q .; then
    bash "$SCRIPT_DIR/backup.sh"
else
    echo "    WARNING: Database container is not running — skipping backup"
    echo "    If this is first deploy, this is expected."
fi
echo ""

echo "==> Step 2: Rebuilding app container..."
docker compose build app
echo ""

echo "==> Step 3: Restarting app container (DB untouched)..."
docker compose up -d --no-deps app
echo ""

echo "==> Step 4: Waiting for app to be healthy..."
sleep 3
for i in $(seq 1 10); do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/auth/login 2>/dev/null || echo "000")
    if [ "$HTTP_CODE" = "200" ]; then
        echo "    App is healthy (HTTP 200 on /auth/login)"
        break
    fi
    if [ "$i" = "10" ]; then
        echo "    ERROR: App did not become healthy after 10 attempts"
        echo "    Check logs: docker compose logs app --tail 50"
        exit 1
    fi
    echo "    Attempt $i/10 — HTTP $HTTP_CODE, retrying in 3s..."
    sleep 3
done
echo ""

echo "==> Step 5: Migration log (last 20 lines):"
docker compose logs app --tail 20 2>&1 | grep -i "migration" || echo "    (no migration output found)"
echo ""

echo "============================================"
echo "  Deploy complete. Database was NOT touched."
echo "============================================"
