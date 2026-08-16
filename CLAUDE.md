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
| 8 | Hardening (validation, error handling, security, **tests**, perf) | ⬜ next |
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

## Next up — Phase 8 (hardening)

Suggested focus: backend tests (Goravel test suite for auth, checkout atomicity,
RBAC), consistent error handling, rate limiting on `/auth/login`, secure headers,
input sanitisation, and a few DB indexes review. Then Phase 9 (deployment).
