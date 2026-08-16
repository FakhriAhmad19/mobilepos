# Kasirku Mobile — Project Guide (for Claude Code)

Mobile POS & inventory system built from a PRD. This file is the context handoff
so any Claude Code session (including a fresh cloud/mobile one) can continue work.

## Stack & layout

- `backend/` — Goravel (Go 1.25) REST API, MySQL (all tables use the `app_`
  prefix), JWT auth + RBAC. Runs in Docker.
- `mobile/` — Expo SDK 57 + TypeScript app (React Navigation, TanStack Query,
  Zustand, React Hook Form + Zod, Axios, expo-secure-store).
- `docker-compose.yml` — mysql + backend + nginx.
- `docs/DATABASE.md` — ERD + schema catalogue.
- `README.md` — full API reference and run instructions.

## Running it (fresh environment)

Secrets (`.env`, `backend/.env`) are gitignored, so a fresh clone needs them
generated once. The bootstrap script does everything:

```bash
./scripts/dev-setup.sh
```

It creates env files (+ generates `APP_KEY` and `JWT_SECRET`), builds and starts
Docker, then runs `artisan migrate` and `db:seed`. Verify with:

```bash
curl http://localhost:3000/api/v1/health
```

Demo logins (all password `password`): `admin@kasirku.test`,
`cashier@kasirku.test`, `warehouse@kasirku.test`.

## Phase status (per the PRD)

| Phase | Scope | Status |
| ----- | ----- | ------ |
| 1 | Foundation (monorepo, Docker, MySQL) | ✅ done |
| 2 | Database (22 `app_` tables, models, seeders) | ✅ done |
| 3 | Auth (JWT login/me/refresh/logout, RBAC middleware) | ✅ done |
| 4 | Master data CRUD (categories, units, products, suppliers, customers, warehouses) | ✅ done |
| 5 | Inventory (stock, stock-in/out, adjustment, opname, movement ledger) | ✅ done |
| 6 | POS/checkout (cart, barcode, atomic checkout, receipt, history, void) | ✅ done |
| 7 | Dashboard & reports (KPIs, 7-day chart, sales/products/inventory/cashier) | ✅ done |
| 8 | Hardening (validation, error handling, security, **tests**, perf) | 🚧 in progress |
| 9 | Deployment (prod Docker, HTTPS, CI/CD, backups) | ⬜ pending |

Each completed phase was verified end-to-end with live curl/db checks. The
backend currently boots with ~66 routes.

## Conventions & gotchas (learned the hard way)

- **Build via Docker.** The backend is designed to build inside Docker
  (`docker compose build backend`); a local Go toolchain is optional. Migrations,
  models and seeders are compiled into the binary — **rebuild the image after any
  Go change** before running `artisan`.
- **Run artisan:** `docker compose exec -T backend /www/main artisan <cmd>`.
- **Goravel v1.18 `http.Middleware` is an interface** `{ Handle(Context); Signature() string }`,
  NOT a func. Middleware are structs (see `app/http/middleware/`).
- **Facades:** use `github.com/goravel/framework/facades` (has `Auth`, `Orm`,
  `Hash`, `Config`, `Route`). The app-local `app/facades` wrappers lack `Hash`.
- **Validation rules:** stick to well-known names (`required`, `email`,
  `max_len`, `min_len`). An unknown rule makes `ctx.Request().Validate` return an
  error (→ 500), not a 422.
- **Routing:** don't put a static segment as a sibling of a `{id}` param at the
  same depth (e.g. `products/barcode` next to `products/{id}`) — risks a router
  panic at boot. Barcode lookup lives at `pos/scan/{barcode}` for this reason.
- **DB:** MySQL, `app_` table prefix (set in `backend/config/database.go`). In
  raw SQL, prefix table names and add `deleted_at IS NULL` / `status='completed'`
  filters manually.
- **Money** columns are `DECIMAL(15,2)`; prices are server-authoritative at
  checkout (never trust client prices).

## The stock engine (core invariant)

Every stock change goes through `services.ApplyMovement(tx, MovementInput)`
inside a `facades.Orm().Transaction(...)`: it row-locks/creates the stock row,
applies a signed delta, enforces `stock >= 0` (`ErrInsufficientStock`), and
writes an immutable `app_stock_movements` ledger row. Stock-in/out, adjustment,
opname, POS `SALE`, and order `void` (RETURN) all use it. Insufficient stock →
`422` and the whole transaction rolls back. Preserve this pattern.

## RBAC matrix

- Reads (lists/show/dashboard): any authenticated user.
- Master-data writes: `admin`+`warehouse` (products/categories/units/suppliers/
  warehouses), `admin`+`cashier` (customers).
- Inventory writes (stock-in/out/adjustment/opname): `admin`+`warehouse`.
- Checkout: `admin`+`cashier`. Void: `admin`. Reports: `admin`.

## Response envelope (PRD §11)

Success: `{ success, message, data, meta }`. Error:
`{ success, message, errors? }`. Helpers in `app/http/controllers/helpers.go`
and `auth_controller.go`.

## Phase 8 (hardening) — progress

Landed so far:

- **Security headers** — `middleware.SecureHeaders()` (registered as global
  middleware in `routes/api.go`) sets `X-Content-Type-Options`, `X-Frame-Options`,
  `Referrer-Policy`, `X-XSS-Protection`, and a locked-down CSP on every response.
- **Login rate limiting** — `middleware.Throttle(name, max, window)` (in-memory
  fixed-window, per client IP) guards `POST /auth/login` at 10 req/min → `429`
  with a `Retry-After` header.
- **Consistent errors** — `facades.Route().Recover(...)` turns any unhandled
  panic into the PRD `{success:false,message}` envelope + a logged error, instead
  of a leaked stack trace.
- **Testability refactor** — checkout money math extracted to pure functions in
  `app/services/pricing.go` (`LineSubtotal`, `GrandTotal`, `PaymentCovers`).
- **DB index (perf)** — migration `20260816000001_add_report_indexes` adds a
  composite `app_orders(status, created_at)` index. Every report and the
  order-history date filter scan orders by `status='completed' AND created_at
  BETWEEN …`; the table previously had `status` alone, so the date range still
  meant scanning all completed orders.
- **Validation coverage** — reviewed all write endpoints: master-data
  controllers use `ctx.Request().Validate` (required/max_len rules); the
  inventory/POS controllers (order, stock-in/out/adjustment/opname) validate
  manually via `Bind` + explicit existence/quantity checks. No gaps found.
- **Tests**
  - Unit (no DB, run anywhere): `app/services/pricing_test.go`,
    `app/http/middleware/rate_limit_test.go`,
    `app/http/controllers/helpers_test.go` (pagination math) — `go test ./app/...`.
  - Feature (need MySQL): `tests/feature/hardening_test.go` covers auth
    (login success/invalid/validation), security headers, RBAC (cashier can't
    write master data, warehouse can't checkout, unauth → 401), and **checkout
    atomicity** (rollback leaves stock untouched and no order row).
    `tests/feature/inventory_test.go` adds the stock-engine flows: void returns
    stock via a RETURN movement (admin-only), stock-in increments stock and
    writes a STOCK_IN ledger row, barcode scan lookup (hit/404), and
    insufficient-payment rejection. Both share one suite/token cache. Run with
    the dev stack up: `docker compose up -d mysql && (cd backend && go test ./tests/...)`.

Local build/test note: `go mod tidy` was run to complete `go.sum` for host builds
(it also swapped stale postgres indirect deps for the mysql ones actually used).
`go build ./...`, `go vet ./...`, and the unit tests all pass on the host; the
feature tests compile (`go test -c ./tests/feature/`) and run once a MySQL is
reachable.

Still open for Phase 8: broader input sanitisation/validation coverage on
master-data writes, a DB index review for report/lookup hot paths, and expanding
feature-test coverage. Then Phase 9 (deployment).
