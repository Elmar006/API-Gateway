# API Gateway

A production-grade API Gateway with a desktop control plane.

The repository ships **two** components and a single Docker Compose orchestration:

| Component | Path | Stack | Purpose |
|---|---|---|---|
| **Backend** | [`backend/`](./backend) | Go 1.25, chi, pgx/pgxpool, Prometheus, golang-migrate | Reverse-proxy gateway with dynamic routing from PostgreSQL, JWT-protected admin API, sharded IP rate-limiter, async batch logging, health-checks, metrics, graceful shutdown |
| **Client** | [`client/`](./client) | Tauri 2 + React 18 + TypeScript + Tailwind + shadcn/ui + React Flow + Zustand + Framer Motion | Native desktop control plane (RouteFlow) with dark-glass UI, interactive routing graph, dynamic inspector, file import/export and system notifications via Tauri |

Both components are independent runtimes connected by HTTP. The **backend** exposes a public proxy on port `8080` and a JWT-protected admin API on port `9090`; the **client** consumes the admin API and renders the topology graph, route inspector, and configuration panels.

---

## Quick start (Docker Compose)

The included [`docker-compose.yml`](./docker-compose.yml) spins up the full stack — `postgres`, the Go gateway, and a containerised web build of the desktop client behind nginx — with a single command.

```bash
# 1. Configure secrets
cp .env.example .env
# Edit .env: set JWT_SECRET (>=16 chars), ADMIN_USERNAME, ADMIN_PASSWORD,
#            DB_PASSWORD, etc.

# 2. Build and start everything
docker compose up -d --build

# 3. Verify
curl -i http://localhost:8080/health      # public proxy
curl -i http://localhost:9090/ready       # admin API + DB ping
open  http://localhost:8000/              # client web bundle
```

Bootstrapping behaviour:

* `postgres` — Postgres 16 with a named volume (`apigw-pg-data`) and `pg_isready` healthcheck.
* `backend` — waits for postgres healthy, applies `golang-migrate` migrations from the embedded FS, creates the bootstrap admin user from `ADMIN_USERNAME` / `ADMIN_PASSWORD`, and starts both servers.
* `client` — multi-stage build: Vite produces the web bundle, nginx serves it on port `80` (mapped to host `8000`) and proxies `/admin/*` to `backend:9090` inside the docker network.

Stop and clean up everything (including the database volume):

```bash
docker compose down -v
```

---

## Components

### Backend ([`backend/`](./backend))

* Two HTTP servers: public proxy (`PORT`, default `:8080`) and admin API (`ADMIN_PORT`, default `:9090`).
* Dynamic routes stored in PostgreSQL, matched by an in-process trie router with `{param}` and `*` wildcards; matched parameters are propagated to upstreams as `X-Path-Param-<name>` headers.
* Async request log pipeline: ring buffer + `pgx.CopyFrom` batches.
* Token-bucket rate limiter sharded by FNV32a of the client IP (16 shards) with TTL cleanup.
* Prometheus metrics (`/metrics` on the admin port, JWT-protected).
* Health checks (`/health`, `/live`, `/ready`).
* Graceful shutdown of both servers via `errgroup` + `signal.NotifyContext`.
* Migrations embedded via `embed.FS`, executed on startup; CLI `--migrate up|down|down N`.
* Full unit + integration test suite (testcontainers-postgres). See [`backend/README.md`](./backend/README.md) for the full reference.

### Client — RouteFlow ([`client/`](./client))

* Native desktop application built on **Tauri 2** (Rust shell + WebView).
* Custom borderless title bar with drag region, global search, notifications, account chip, and OS-styled window controls.
* Three-column layout: navigation sidebar with environment switcher, interactive routing canvas, dynamic inspector.
* Routing canvas is a fully restyled **React Flow v12** graph with custom glass nodes (sources, gateway, routes, services), accent-glowing curved edges, hover/select highlighting, custom minimap, and floating toolbar.
* Inspector built on **React Hook Form**: editable fields for path, method, target, load-balancing strategy, retries, timeout, healthcheck thresholds, rate limits, headers, policies. Tabs for Overview / Rules / Middleware / Metrics with custom SVG sparklines.
* Tauri integration: native open/save dialogs, JSON/YAML import-export of routing configuration, system notifications, native filesystem access, custom window controls.
* See [`client/README.md`](./client/README.md) for the full reference and packaging instructions (`npm run tauri:build`).

---

## Environment variables

The Docker Compose stack reads its configuration from a single `.env` file at the repository root (see [`.env.example`](./.env.example)). Required variables:

| Variable | Description |
|---|---|
| `JWT_SECRET` | ≥ 16 character secret used to sign admin and user JWTs |
| `ADMIN_USERNAME` | Bootstrap admin login (created on first start when `admin_users` is empty) |
| `ADMIN_PASSWORD` | Bootstrap admin password (hashed with bcrypt cost 12) |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | PostgreSQL credentials |

Optional tuning knobs (rate limit TTL, log buffer/batch sizes, max body, TLS, port mappings) are listed in [`.env.example`](./.env.example) with sane defaults. The full backend reference is in [`backend/.env.example`](./backend/.env.example).

---

## Repository layout

```
API-Gateway/
├── backend/                     # Go 1.25 API gateway (production-grade)
│   ├── cmd/main.go
│   ├── internal/                # domain · application · infrastructure · presentation
│   ├── migrations/              # SQL migrations executed via golang-migrate
│   ├── Makefile                 # build · test · lint · migrate · run
│   ├── Dockerfile               # multi-stage distroless image
│   └── README.md
├── client/                      # Tauri + React desktop control plane (RouteFlow)
│   ├── src/                     # React UI: layout, components, graph, panels, store
│   ├── src-tauri/               # Rust shell, capabilities, icons
│   ├── Dockerfile               # Vite build → nginx (browser preview)
│   ├── nginx.conf
│   └── README.md
├── docker-compose.yml           # postgres + backend + client
├── .env.example                 # docker-compose secrets / configuration
├── .gitignore
└── README.md                    # ← you are here
```

---

## Local development without Docker

If you want to iterate on a single component without Docker:

```bash
# Backend
cd backend
cp .env.example .env             # set secrets
make run                         # applies migrations and starts both servers

# Client (browser dev mode, no native window)
cd client
npm install
npm run dev                      # http://localhost:1420

# Client (full native desktop window via Tauri)
cd client
npm install
npm run tauri:dev
```

Linux prerequisites for native Tauri builds:

```bash
sudo apt install pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev \
    libayatana-appindicator3-dev librsvg2-dev libssl-dev libsoup-3.0-dev
```

For full per-component documentation see [`backend/README.md`](./backend/README.md) and [`client/README.md`](./client/README.md).

---

## License

Proprietary. Internal use only.
