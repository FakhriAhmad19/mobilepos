# Kasirku Mobile

Mobile POS & Inventory Management System — a React Native (Expo) app backed by a
Goravel (Go) REST API and MySQL. See the PRD for the full product scope.

This repository is the **Phase 1 foundation** (PRD §21): a runnable monorepo
skeleton wired end-to-end. Feature modules (auth, master data, inventory, POS,
reporting) are built out in later phases.

## Architecture

```
React Native (Expo, TS)  ──HTTPS/JSON──►  Goravel API (Go)  ──►  MySQL
                                              ▲
                                            nginx (reverse proxy)
```

## Repository layout

```
.
├── backend/            Goravel REST API (Go 1.25) — MySQL, JWT-ready, /api/v1
├── mobile/             Expo + TypeScript app (React Navigation, TanStack Query,
│                       Zustand, React Hook Form + Zod, Axios, SecureStore)
├── docker/nginx/       Reverse-proxy config
├── docker-compose.yml  mysql + backend + nginx
└── .env.example        Root env consumed by docker-compose
```

## Prerequisites

- Docker + Docker Compose (for backend + MySQL)
- Node.js 20+ and npm (for the mobile app)
- No local Go toolchain required — the backend builds inside Docker

## Quick start — backend + database

```bash
cp .env.example .env          # adjust credentials (must match backend/.env)
docker compose up --build
```

Then verify the API is live:

```bash
curl http://localhost:3000/api/v1/health
```

Expected response:

```json
{
  "success": true,
  "message": "Service healthy",
  "data": { "service": "KasirkuMobile", "env": "local", "database": "up", "timestamp": "..." },
  "meta": {}
}
```

The API is also reachable through nginx at `http://localhost:8080/api/v1/health`.

### API surface

| Method | Path                          | Auth        | Description                          |
| ------ | ----------------------------- | ----------- | ------------------------------------ |
| GET    | `/api/v1/ping`                | public      | Liveness — never touches the DB      |
| GET    | `/api/v1/health`              | public      | Liveness + MySQL connectivity check  |
| POST   | `/api/v1/auth/login`          | public      | Authenticate, returns JWT + user     |
| GET    | `/api/v1/auth/me`             | bearer      | Current authenticated user           |
| POST   | `/api/v1/auth/refresh`        | bearer      | Issue a fresh token                  |
| POST   | `/api/v1/auth/logout`         | bearer      | Invalidate the current token         |
| POST   | `/api/v1/auth/change-password`| bearer      | Change own password                  |
| GET    | `/api/v1/admin/ping`          | bearer+admin| Example role-gated route (RBAC)      |
| —      | `/api/v1/{categories,units,products,suppliers,warehouses,customers}` | bearer | Master-data CRUD (§8.4–§8.8) |

Master-data resources expose `GET` (list — `?search=`, `?page=`, `?per_page=`,
plus `?status=` / `?category_id=` filters), `GET /{id}`, `POST`, `PUT /{id}`,
`DELETE /{id}` (soft delete). Reads are open to any authenticated user; writes
are restricted — `admin`+`warehouse` for products/categories/units/suppliers/
warehouses, `admin`+`cashier` for customers.

### Inventory (Phase 5)

| Method | Path                                   | Auth              | Description                              |
| ------ | -------------------------------------- | ----------------- | ---------------------------------------- |
| GET    | `/api/v1/inventory`                    | bearer            | Stock levels (`?warehouse_id`, `?search`, `?low_stock=1`) |
| GET    | `/api/v1/stock-movements`              | bearer            | Ledger (`?product_id`, `?type`, `?date_from/to`) |
| GET/POST | `/api/v1/stock-ins`                  | read / admin+wh   | Goods receiving (§8.10)                  |
| GET/POST | `/api/v1/stock-outs`                 | read / admin+wh   | Non-sale removal (§8.11)                 |
| GET/POST | `/api/v1/stock-adjustments`          | read / admin+wh   | Correction to a counted value (§8.12)    |
| GET/POST | `/api/v1/stock-opnames`              | read / admin+wh   | Physical count (§8.13)                   |
| POST   | `/api/v1/stock-opnames/{id}/complete`  | admin+wh          | Apply counted differences                |

Every stock change runs in a DB transaction and writes an immutable
`stock_movements` row (PRD §8.14, §18). Removals that would drive stock below
zero are rejected (`422`) and rolled back atomically.

### POS / transactions (Phase 6)

| Method | Path                          | Auth            | Description                              |
| ------ | ----------------------------- | --------------- | ---------------------------------------- |
| GET    | `/api/v1/payment-methods`     | bearer          | Active tender types (`?active=1`)        |
| GET    | `/api/v1/pos/scan/{barcode}`  | bearer          | Barcode → product (§8.17), `404` if none |
| POST   | `/api/v1/orders`              | admin+cashier   | Atomic checkout (§8.19) → receipt        |
| GET    | `/api/v1/orders`              | bearer          | Transaction history (§8.21)              |
| GET    | `/api/v1/orders/{id}`         | bearer          | Order + items + payments (receipt, §8.22)|
| POST   | `/api/v1/orders/{id}/void`    | admin           | Cancel + restock (RETURN movement)       |

Checkout is server-authoritative (prices come from the product record), computes
totals + change, validates payment ≥ total, and reduces stock via `SALE`
movements — order, items, payments and stock changes all commit or roll back
together. The mobile app has a full POS tab (search → cart → cash payment →
receipt) and a Transactions tab with tap-through receipts.

### Dashboard & reports (Phase 7)

| Method | Path                          | Auth   | Description                                   |
| ------ | ----------------------------- | ------ | --------------------------------------------- |
| GET    | `/api/v1/dashboard`           | bearer | KPIs, 7-day sales series, top products (§8.23)|
| GET    | `/api/v1/reports/sales`       | admin  | Daily sales over a range (`?date_from/to`)    |
| GET    | `/api/v1/reports/products`    | admin  | Best/worst sellers (`?sort=best\|worst`)      |
| GET    | `/api/v1/reports/inventory`   | admin  | Stock valuation + low-stock list              |
| GET    | `/api/v1/reports/cashiers`    | admin  | Transactions & sales per cashier              |

Aggregations are computed in SQL (excluding soft-deleted and cancelled orders).
The mobile Beranda tab shows the KPIs, a 7-day sales bar chart, a low-stock
alert, and the top-products list.

Auth is JWT via the `Authorization: Bearer <token>` header. Roles (`admin`,
`cashier`, `warehouse`) are enforced server-side by middleware (PRD §8.2).
Demo logins (all password `password`): `admin@kasirku.test`,
`cashier@kasirku.test`, `warehouse@kasirku.test`.

## Quick start — mobile app

```bash
cd mobile
cp .env.example .env          # point EXPO_PUBLIC_API_URL at your API
npm install                   # already installed if you just scaffolded
npm run ios                   # or: npm run android / npm run web
```

On a physical device, set `EXPO_PUBLIC_API_URL` in `mobile/.env` to your
machine's LAN IP (e.g. `http://192.168.1.20:3000/api/v1`) — `localhost` from the
device points at the device itself.

The app boots to a placeholder login (form validation is real; authentication is
stubbed until Phase 3), then a dashboard that live-checks backend + DB health.

## Configuration notes

- **Database**: MySQL with the `app_` table prefix (PRD §12), configured in
  `backend/config/database.go`.
- **Secrets**: `backend/.env` holds `APP_KEY` and `JWT_SECRET` (generated, gitignored).
  Regenerate with `openssl` if needed. Keep `backend/.env` DB credentials in sync
  with the root `.env` MySQL credentials.
- **Response envelope**: all API responses follow the PRD §11 shape
  (`success` / `message` / `data` / `meta`, or `errors` on failure). The mobile
  API client normalises these in `mobile/src/api/client.ts`.

## Development phases

Phase 1 (this repo) — foundation. Next up per the PRD: Phase 2 database/ERD,
Phase 3 authentication, Phase 4 master data, Phase 5 inventory, Phase 6 POS,
Phase 7 dashboard/reporting.
