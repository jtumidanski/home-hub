# Home Hub — Architecture

## 1. Overview

Home Hub is a multi-tenant household productivity platform implemented as a microservice-based system.

The system is designed to run:

- locally using Docker Compose
- in Kubernetes (k3s)
- behind a reverse proxy using path-prefix routing

The architecture emphasizes:

- strict service boundaries
- versioned APIs
- JSON:API resource modeling
- environment-driven configuration
- reproducible builds
- mature CI/CD from the beginning

Core services:

- frontend
- auth-service
- account-service
- productivity-service
- recipe-service
- weather-service
- calendar-service
- package-service

Shared modules provide common functionality but do not contain business logic.

Each service maintains its own documentation under `docs/` per the DOCS.md contract.

## 2. High Level Architecture

```
Browser → Ingress / Reverse Proxy → Services
```

Routing:

```
/ -> frontend
/api/v1/auth -> auth-service
/api/v1/users -> auth-service
/api/v1/tenants -> account-service
/api/v1/households -> account-service
/api/v1/memberships -> account-service
/api/v1/preferences -> account-service
/api/v1/contexts -> account-service
/api/v1/tasks -> productivity-service
/api/v1/reminders -> productivity-service
/api/v1/summary -> productivity-service
/api/v1/recipes -> recipe-service
/api/v1/ingredients -> recipe-service
/api/v1/meals -> recipe-service
/api/v1/weather -> weather-service
/api/v1/calendar -> calendar-service
/api/v1/packages -> package-service
```

All services are stateless except for database persistence. Authentication is handled by auth-service. Authorization is enforced by each service using JWT claims.

## 3. Service Responsibilities

### 3.1 frontend

Responsibilities:

- UI rendering
- API orchestration
- session bootstrap
- theme switching
- household switching
- JSON:API client

Technology: React, Vite, ShadCN, static build, nginx container. No server-side rendering.

### 3.2 auth-service

Responsibilities:

- OIDC login
- JWT issuance
- refresh token sessions
- JWKS endpoint
- user identity
- external identity mapping

Schema: `auth.users`, `auth.external_identities`, `auth.oidc_providers`, `auth.refresh_tokens`

Auth model:

- asymmetric JWT
- short-lived access token
- refresh token stored server-side

JWKS endpoint: `/api/v1/auth/.well-known/jwks.json`

Downstream services validate JWT using JWKS.

### 3.3 account-service

Responsibilities:

- tenant management
- household management
- membership roles
- user preferences
- active context

Schema: `account.tenants`, `account.households`, `account.memberships`, `account.preferences`

Rules:

- one tenant per user initially
- multiple households allowed
- roles per household
- preference per tenant per user
- context persisted

Endpoint: `/api/v1/contexts/current`

### 3.4 productivity-service

Responsibilities:

- tasks
- reminders
- summary projections

Schema: `productivity.tasks`, `productivity.task_restorations`, `productivity.reminders`, `productivity.reminder_snoozes`, `productivity.reminder_dismissals`

Rules:

- soft delete
- restore window
- fixed snooze durations
- derived summary
- tenant scoped
- household scoped

Summary endpoints return single resources.

### 3.5 recipe-service

Responsibilities:

- recipe CRUD
- Cooklang source parsing
- tag management
- soft delete with restore window

Schema: `recipe.recipes`, `recipe.recipe_tags`, `recipe.recipe_restorations`

Rules:

- recipes stored as Cooklang plain-text, parsed server-side
- Cooklang syntax validated on create and update
- tags normalized to lowercase, deduplicated
- soft delete with 3-day restore window
- tenant scoped, household scoped

### 3.6 calendar-service

Responsibilities:

- Google Calendar OAuth connection management
- background event synchronization via polling
- encrypted OAuth token storage (AES-256-GCM)
- household calendar event aggregation
- privacy masking for private/confidential events
- user color assignment per household

Schema: `calendar.calendar_connections`, `calendar.calendar_sources`, `calendar.calendar_events`, `calendar.calendar_oauth_states`

Rules:

- separate OAuth flow from auth-service (calendar.readonly scope)
- incremental sync via Google sync tokens per source calendar
- background sync at configurable interval (default 15 minutes)
- sync staggered with random jitter to avoid burst traffic
- exponential backoff on Google API 429/5xx
- manual sync rate-limited to once per 5 minutes
- private/confidential events masked as "Busy" for non-owners
- tenant scoped, household scoped

### 3.7 package-service

Responsibilities:

- household package tracking across USPS, UPS, FedEx
- carrier auto-detection from tracking numbers
- background status polling with adaptive intervals
- package lifecycle management (archive, stale detection, cleanup)
- privacy controls per package

Schema: `package.packages`, `package.tracking_events`

Rules:

- polls carrier APIs via OAuth 2.0 client credentials
- per-carrier daily rate budgets (USPS: 1000, UPS: 250, FedEx: 500)
- adaptive polling: 30m default, 15m for out-for-delivery
- stale after 14 days with no status change
- auto-archive delivered packages after 7 days
- hard-delete archived packages after 30 days
- max 25 active packages per household
- private packages redacted for non-owners
- manual refresh rate-limited to once per 5 minutes
- tenant scoped, household scoped

### 3.8 weather-service

Responsibilities:

- current weather conditions
- 7-day forecast
- geocoding place search
- weather data caching
- background cache refresh

Schema: `weather.weather_caches`

Rules:

- fetches from Open-Meteo public API (no API key)
- caches weather data per household in PostgreSQL (JSONB)
- background ticker refreshes cache at configurable interval
- cache self-heals on coordinate/unit mismatch
- tenant scoped, household scoped

## 4. API Design

### 4.1 Versioning

All endpoints are versioned under `/api/v1/...`. Versioning is required from the start.

### 4.2 Resource Style

JSON:API-style resources:

```
/tasks
/tasks/{id}
/tasks/restorations
/reminders
/reminders/snoozes
/reminders/dismissals
```

Summary:

```
/summary/tasks
/summary/reminders
/summary/dashboard
```

Context: `/contexts/current`

### 4.3 Includes

Supported for summary endpoints. Example:

```
/summary/dashboard?include=tasks,reminders
```

## 5. Authentication Model

Flow:

1. frontend redirects to auth-service
2. auth-service performs OIDC login
3. auth-service issues JWT
4. frontend stores cookies
5. downstream services validate JWT

Tokens:

- short-lived access token
- refresh token stored server-side

JWKS used for validation.

## 6. Multi-Tenancy Model

Hierarchy:

```
User → Tenant → Household → Membership
```

All data must include `tenant_id` and `household_id`. No global data.

## 7. Configuration Model

All services configured via environment variables:

```
DB_HOST
DB_USER
DB_PASSWORD
OIDC_CLIENT_ID
OIDC_SECRET
JWT_PRIVATE_KEY
```

Secrets provided externally. No config files required.

## 8. Persistence

Each service owns its schema: `auth.*`, `account.*`, `productivity.*`, `recipe.*`, `weather.*`, `calendar.*`, `package.*`. No cross-service tables.

Migrations:

- per service
- run on startup via GORM AutoMigrate in each domain's `entity.go`
- forward-only acceptable
- no separate SQL migration files

ORM: GORM. IDs: UUID, generated in application.

## 9. Logging

Standard: Logrus, structured JSON.

Every request should include: `request_id`, `user_id`, `tenant_id`, `household_id`.

## 10. Tracing

OpenTelemetry used. Required: trace id, span id, request id correlation. Metrics optional.

## 11. Local Runtime

Local environment uses:

- docker-compose
- nginx reverse proxy
- path prefix routing

External infrastructure: postgres, redis (future). Config via `.env`.

## 12. Kubernetes Runtime

Deployment: plain YAML, no helm required.

Images:

```
ghcr.io/<owner>/home-hub-auth
ghcr.io/<owner>/home-hub-account
ghcr.io/<owner>/home-hub-productivity
ghcr.io/<owner>/home-hub-recipe
ghcr.io/<owner>/home-hub-weather
ghcr.io/<owner>/home-hub-calendar
ghcr.io/<owner>/home-hub-package
ghcr.io/<owner>/home-hub-frontend
```

Tags: `main`, `sha`. Ingress uses path prefixes.

## 13. CI/CD

GitHub Actions.

PR: build impacted, test impacted, lint impacted.

Main: publish snapshot images.

Scripts used for builds.

## 14. Dependency Management

Renovate enabled. Supports: Go modules, npm, GitHub Actions, Docker. Automerge disabled.

## 15. API Testing

Bruno collections under `bruno/` (`auth/`, `account/`, `productivity/`, `recipe/`, `packages/`, `environments/`). Used for manual endpoint testing.

## 16. Shared Modules

Shared Go modules live under `shared/go/`:

| Module   | Purpose                                                                              |
| -------- | ------------------------------------------------------------------------------------ |
| auth     | JWT validation and auth middleware                                                   |
| database | GORM connection, migration orchestration, tenant callbacks                           |
| http     | HTTP utilities                                                                       |
| logging  | Logrus structured logging                                                            |
| model    | shared domain model types                                                            |
| server   | HTTP server lifecycle, handler registration, JSON:API response helpers, health checks, middleware, tracing |
| tenant   | tenant context extraction and validation                                             |
| testing  | test helpers and fixtures                                                            |

No business logic in shared modules.

## 17. Service Code Pattern

Each service domain follows a consistent file structure:

| File           | Purpose                            |
| -------------- | ---------------------------------- |
| `model.go`     | immutable domain model with accessors |
| `entity.go`    | GORM entity with `Migration()` function |
| `builder.go`   | fluent builder enforcing invariants |
| `processor.go` | pure business logic                |
| `provider.go`  | lazy database access               |
| `resource.go`  | route registration and HTTP handlers |
| `rest.go`      | JSON:API resource mappings         |

Details in DOCS.md contract and per-service documentation.

## 18. Service Documentation

Each service maintains its own documentation per the DOCS.md contract:

```
services/<service>/docs/domain.md
services/<service>/docs/rest.md
services/<service>/docs/storage.md
```

These are the authoritative source for domain logic, REST endpoints, and database schema per service.

## 19. Design Principles

- strict service boundaries
- versioned APIs
- stateless services
- env-driven config
- per-service schema
- JSON:API style
- minimal shared code
- reproducible builds
- CI from start
- deployment from start
