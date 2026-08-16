#!/usr/bin/env bash
# Kasirku Mobile — MySQL backup (PRD Phase 9).
# Dumps the database from the running `mysql` compose service to a timestamped,
# gzipped file and prunes old backups.
#
# Usage:
#   ./scripts/backup-db.sh                 # uses docker-compose.prod.yml
#   COMPOSE_FILE=docker-compose.yml ./scripts/backup-db.sh
#
# Env:
#   BACKUP_DIR   destination directory (default: ./backups)
#   KEEP_DAYS    delete backups older than this many days (default: 14)
#
# Schedule daily via cron, e.g.:
#   0 2 * * *  cd /opt/kasirku && ./scripts/backup-db.sh >> /var/log/kasirku-backup.log 2>&1
set -euo pipefail

cd "$(dirname "$0")/.."

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
KEEP_DAYS="${KEEP_DAYS:-14}"

# Load DB credentials from the root .env (same file compose uses).
if [ -f .env ]; then
  set -a; . ./.env; set +a
fi
DB_NAME="${MYSQL_DATABASE:-kasirku}"
DB_USER="${MYSQL_USER:-kasirku}"
DB_PASS="${MYSQL_PASSWORD:?MYSQL_PASSWORD not set (put it in .env)}"

mkdir -p "$BACKUP_DIR"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$BACKUP_DIR/kasirku-${DB_NAME}-${STAMP}.sql.gz"

echo "▶ Backing up '${DB_NAME}' → ${OUT}"
docker compose -f "$COMPOSE_FILE" exec -T mysql \
  mysqldump --single-transaction --quick --routines --triggers \
    -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" \
  | gzip > "$OUT"

# Fail loudly if the dump produced an empty/broken file.
if [ ! -s "$OUT" ]; then
  echo "✗ Backup file is empty — aborting" >&2
  rm -f "$OUT"
  exit 1
fi

echo "✓ Backup complete ($(du -h "$OUT" | cut -f1))"

echo "▶ Pruning backups older than ${KEEP_DAYS} days"
find "$BACKUP_DIR" -name 'kasirku-*.sql.gz' -type f -mtime "+${KEEP_DAYS}" -print -delete
echo "✓ Done"
