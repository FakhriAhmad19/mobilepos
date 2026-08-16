#!/usr/bin/env bash
# Kasirku Mobile — one-command dev bootstrap.
# Creates env files (+ secrets), builds & starts Docker, migrates and seeds.
# Safe to re-run: it won't overwrite an existing .env, and migrate/seed are idempotent.
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
echo "▶ Kasirku dev setup — $ROOT"

# --- 1. Root .env (docker-compose) --------------------------------------
if [ ! -f .env ]; then
  cp .env.example .env
  echo "  ✓ created .env from .env.example"
else
  echo "  • .env already exists — leaving it"
fi

# --- 2. Backend .env (+ generated secrets) ------------------------------
if [ ! -f backend/.env ]; then
  APP_KEY="$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32)"
  JWT_SECRET="$(openssl rand -base64 48 | tr -d '\n' | tr '/+' '_-')"
  sed -e "s#^APP_KEY=.*#APP_KEY=${APP_KEY}#" \
      -e "s#^JWT_SECRET=.*#JWT_SECRET=${JWT_SECRET}#" \
      -e "s#^DB_HOST=.*#DB_HOST=mysql#" \
      -e "s#^DB_PASSWORD=.*#DB_PASSWORD=kasirku_secret_change_me#" \
      backend/.env.example > backend/.env
  echo "  ✓ created backend/.env with generated APP_KEY & JWT_SECRET"
else
  echo "  • backend/.env already exists — leaving it"
fi

# --- 3. Build & start the stack -----------------------------------------
echo "▶ Building & starting Docker (mysql + backend + nginx)…"
docker compose up -d --build

# --- 4. Wait for the backend to answer ----------------------------------
echo "▶ Waiting for the API to come up…"
for i in $(seq 1 60); do
  if curl -fsS http://localhost:3000/api/v1/ping >/dev/null 2>&1; then
    echo "  ✓ API is up"
    break
  fi
  sleep 2
  [ "$i" = "60" ] && { echo "  ✗ API did not come up in time"; docker compose logs backend | tail -20; exit 1; }
done

# --- 5. Migrate & seed ---------------------------------------------------
echo "▶ Running migrations…"
docker compose exec -T backend /www/main artisan migrate
echo "▶ Seeding database…"
docker compose exec -T backend /www/main artisan db:seed

# --- 6. Done -------------------------------------------------------------
echo
echo "✅ Ready. Health check:"
curl -fsS http://localhost:3000/api/v1/health || true
echo
echo
echo "   API:    http://localhost:3000/api/v1"
echo "   Login:  admin@kasirku.test / password"
echo
echo "   Mobile app:  cd mobile && cp -n .env.example .env && npm install && npm run web"
