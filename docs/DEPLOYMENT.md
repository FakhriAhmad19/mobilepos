# Deployment (Phase 9)

Production deployment for the Kasirku backend: hardened Docker image, HTTPS via
nginx, CI/CD with GitHub Actions, and database backups.

## Components

| Piece | File |
| ----- | ---- |
| Production image (non-root, healthcheck, no baked secrets) | `backend/Dockerfile.prod` |
| Production orchestration (internal DB, TLS nginx) | `docker-compose.prod.yml` |
| TLS reverse proxy (80 → 443 redirect, HSTS) | `docker/nginx/prod.conf` |
| Prod env template | `.env.prod.example` |
| CI (build, vet, unit + feature tests, image build) | `.github/workflows/ci.yml` |
| DB backup / restore | `scripts/backup-db.sh`, `scripts/restore-db.sh` |
| Staging TLS cert | `scripts/gen-self-signed-cert.sh` |

## First deploy

```bash
# 1. Configure secrets
cp .env.prod.example .env
#   edit .env: strong MYSQL passwords, APP_URL, and generate:
#     APP_KEY:    openssl rand -hex 16        (32 chars)
#     JWT_SECRET: openssl rand -base64 48

# 2. TLS certificate  ->  docker/nginx/certs/{fullchain,privkey}.pem
#    Staging / quick test (browser will warn):
./scripts/gen-self-signed-cert.sh your-domain.example
#    Production: use Let's Encrypt (see below) instead.

# 3. Build & start
docker compose -f docker-compose.prod.yml up -d --build

# 4. Migrate & seed (first time only)
docker compose -f docker-compose.prod.yml exec -T backend /www/main artisan migrate
docker compose -f docker-compose.prod.yml exec -T backend /www/main artisan db:seed   # optional demo data

# 5. Verify
curl -k https://your-domain.example/api/v1/health
```

The backend never has secrets baked into its image — `docker-compose.prod.yml`
injects them from `.env` at runtime (and pins `APP_ENV=production`,
`APP_DEBUG=false`). MySQL publishes no host port; it is reachable only on the
internal compose network.

## Production TLS with Let's Encrypt

The self-signed script is for staging only. For a real domain, issue a
certificate with certbot and drop the results in `docker/nginx/certs/` as
`fullchain.pem` / `privkey.pem` (the names `prod.conf` expects), e.g.:

```bash
certbot certonly --webroot -w /var/www/certbot -d your-domain.example
cp /etc/letsencrypt/live/your-domain.example/fullchain.pem docker/nginx/certs/
cp /etc/letsencrypt/live/your-domain.example/privkey.pem  docker/nginx/certs/
docker compose -f docker-compose.prod.yml restart nginx
```

`prod.conf` already serves `/.well-known/acme-challenge/` from `/var/www/certbot`
for webroot renewals. Reload nginx after each renewal (e.g. a certbot
`deploy-hook`). HSTS is enabled — turn it on only once HTTPS is confirmed
working, since browsers cache it.

## CI/CD

`.github/workflows/ci.yml` runs on every push to `main` and every pull request:

- **backend** job: `go build`, `go vet`, unit tests (`go test ./app/...`), and
  the feature tests (`go test ./tests/...`) against a MySQL 8.4 service
  container. Secrets (`APP_KEY`, `JWT_SECRET`) are generated per run.
- **docker** job: builds `backend/Dockerfile.prod` to catch image regressions.

Extend the `docker` job with a registry login + push step to publish images on
tagged releases when you wire up a registry.

## Backups

`scripts/backup-db.sh` dumps the database from the running `mysql` service to a
timestamped, gzipped file under `backups/` and prunes files older than
`KEEP_DAYS` (default 14). Schedule it daily via cron:

```cron
0 2 * * *  cd /opt/kasirku && ./scripts/backup-db.sh >> /var/log/kasirku-backup.log 2>&1
```

Restore with `scripts/restore-db.sh <backup.sql.gz>` (it prompts before
overwriting). `backups/` and `docker/nginx/certs/*.pem` are git-ignored so dumps
and keys never land in the repo.
