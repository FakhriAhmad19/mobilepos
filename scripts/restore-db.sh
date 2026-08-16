#!/usr/bin/env bash
# Kasirku Mobile — MySQL restore (PRD Phase 9).
# Restores a gzipped dump produced by scripts/backup-db.sh into the running
# `mysql` compose service. THIS OVERWRITES the current database.
#
# Usage:
#   ./scripts/restore-db.sh backups/kasirku-kasirku-20260816-020000.sql.gz
#
# Env:
#   COMPOSE_FILE  compose file to target (default: docker-compose.prod.yml)
set -euo pipefail

cd "$(dirname "$0")/.."

DUMP="${1:-}"
if [ -z "$DUMP" ] || [ ! -f "$DUMP" ]; then
  echo "Usage: $0 <backup.sql.gz>" >&2
  exit 1
fi

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"

if [ -f .env ]; then
  set -a; . ./.env; set +a
fi
DB_NAME="${MYSQL_DATABASE:-kasirku}"
DB_USER="${MYSQL_USER:-kasirku}"
DB_PASS="${MYSQL_PASSWORD:?MYSQL_PASSWORD not set (put it in .env)}"

echo "⚠ This will OVERWRITE database '${DB_NAME}' from ${DUMP}."
read -r -p "  Type 'yes' to continue: " CONFIRM
[ "$CONFIRM" = "yes" ] || { echo "Aborted."; exit 1; }

echo "▶ Restoring '${DB_NAME}'…"
gunzip -c "$DUMP" | docker compose -f "$COMPOSE_FILE" exec -T mysql \
  mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

echo "✓ Restore complete"
