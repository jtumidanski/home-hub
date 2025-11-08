# Home Hub — Architecture Overview

_Last updated: 2025-11-03_

---

## 1. High-Level Overview

**Home Hub** is a modular, microservice-based home information platform designed to power kiosk-style displays and administrative interfaces for shared households.

It delivers:
- Household dashboards (tasks, meals, weather, calendar, reminders)
- Multi-tenant and multi-household access control
- Google-based authentication
- Modular domain services with strong boundaries
- Strict data isolation (Row-Level Security, or RLS)
- Poll-only kiosk UI (no websockets)
- Portable deployment (Docker, Kubernetes, Helm)

---

## 2. Core Principles

- **Microservices by domain:** each domain (tasks, meals, weather, reminders, etc.) owns its own database schema, API, and migrator.
- **Gateway-based access:** the `gateway` service is the only public API entry point.
- **RLS everywhere:** tenant and household isolation are enforced in the database layer.
- **Single-origin local dev:** via `nginx` proxy → `http://localhost:3000`.
- **Mono-repo structure:** all services, UIs, and libraries coexist under one repo for unified CI/CD.
- **Stateless APIs:** all state persisted in Postgres per service; workers handle time-based logic.
- **Polling-based UI updates:** predictable, resilient kiosk design with local cache.

---

## 3. Repository Structure

```
home-hub/
  apps/
    gateway/                 # Public API
    dashboard-composer/      # Dashboard aggregation
    svc-users/               # Tenants, households, members, invites, devices
    svc-calendar/            # Google Calendar integration
    svc-weather/             # Weather forecast + caching
    svc-tasks/               # Tasks CRUD + rollover
    svc-meals/               # Recipes + weekly planner
    svc-reminders/           # Reminders CRUD + snooze/dismiss
    workers/                 # Domain-specific background jobs
    migrators/               # One per service (AutoMigrate + RLS + SQL)
    kiosk/                   # React/Tailwind (Next.js)
    admin/                   # React/Tailwind (Next.js)
  docker-compose.yml
  nginx.conf
  Taskfile.yml
  flake.nix
  go.work
  package.json
```

---

## 4. Services and Boundaries

| Service | Responsibility | Database Schema | Notes |
|----------|----------------|-----------------|-------|
| **gateway** | Entry point, auth, routing, token verification | None | Stateless |
| **dashboard-composer** | Aggregates data from other services | None | Read-only aggregation |
| **svc-users** | Tenants, households, invites, devices | `users`, `tenants`, `households`, `devices` | Defines household timezone |
| **svc-calendar** | Calendar sync and storage | `calendars`, `events`, `tokens` | 5-minute sync worker |
| **svc-weather** | Weather retrieval and caching | `locations`, `current`, `forecast` | 30-minute refresh |
| **svc-tasks** | Tasks CRUD and rollover | `tasks`, `audit_log` | Daily rollover worker |
| **svc-meals** | Recipes and weekly planner | `recipes`, `plan_weeks`, `plan_items` | Sunday planner worker |
| **svc-reminders** | Reminders CRUD and snooze | `reminders`, `audit_log` | 1-minute sweep worker |

---

## 5. Workers and Jobs

| Worker | Schedule | Description |
|--------|-----------|-------------|
| `calendar-sync` | Every 5m | Sync Google Calendar events |
| `weather-sync` | Every 30m | Refresh weather forecast |
| `tasks-rollover` | 00:01 local | Rollover incomplete tasks |
| `meals-planner` | Sunday 08:00 local | Generate weekly plan |
| `reminders-sweep` | Every 1m | Manage reminders and quiet hours |

All workers are headless and communicate with their owning service DBs.

---

## 6. Communication Flow

1. **Kiosk/Admin → Gateway**
   Authenticated requests via Google OIDC tokens.

2. **Gateway → Services**
   Gateway injects headers:
   - `X-Tenant-ID`
   - `X-Household-ID`
   - `X-User-ID`
   - A short-lived **service JWT** signed by Gateway.

3. **Service → Database**
   Sets `app.tenant_id` and `app.household_id` in session.
   Queries automatically filtered by RLS.

4. **Dashboard-composer → Services**
   Fan-out calls, aggregates data, caches results briefly.

---

## 7. Local Development Stack

- **Frontend:** kiosk (`5173`), admin (`5174`)
- **Backend:** gateway (`8080`)
- **Proxy:** nginx (`3000`)
- **Databases:** one Postgres per service
- **Compose command:**
  ```bash
  task dev
  # http://localhost:3000/kiosk/
  # http://localhost:3000/admin/
  # API via http://localhost:3000/api/
  ```

---

## 8. CI/CD

1. Detect changed services using Turborepo + Go workspace graph.
2. Lint, typecheck, unit test per service.
3. Launch ephemeral Postgres → run `migrator-*`.
4. Integration tests verify RLS.
5. Build/push Docker images (`svc-tasks:v1.2.3`, etc.).
6. Deploy changed services, then migrators.
7. Run E2E (kiosk/admin flows) → promote to production.

---

## 9. Key Non-Functional Requirements

| Area | Goal |
|------|------|
| **Dashboard load (Pi 4)** | < 2 seconds |
| **API latency (P99)** | < 300 ms |
| **Offline cache window** | 1 day |
| **RLS leakage** | Zero cross-tenant reads |
| **Audit retention** | 90 days |
| **Frontend polling interval** | 30 seconds |
| **Quiet hours (reminders)** | 00:00–05:00 |

---

## 10. Deployment Summary

| Layer | Description |
|--------|-------------|
| **Ingress (NGINX / Traefik)** | Handles `/api`, `/kiosk`, `/admin` |
| **Gateway** | Auth, routing, aggregation |
| **Domain Services** | Isolated per schema |
| **Workers** | Background jobs per domain |
| **Postgres (schema per service)** |
| **Redis / Cache** | Composer or weather caching |
| **Grafana / Prometheus** | Metrics + uptime |

---
