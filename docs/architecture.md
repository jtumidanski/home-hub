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

Shared modules provide common functionality but do not contain business logic.

---

## 2. High Level Architecture

Browser → Ingress / Reverse Proxy → Services

Routing:

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

All services are stateless except for database persistence.

Authentication is handled by auth-service.

Authorization is enforced by each service using JWT claims.

---

## 3. Service Responsibilities

### 3.1 frontend

Responsibilities:

- UI rendering
- API orchestration
- session bootstrap
- theme switching
- household switching
- JSON:API client

Technology:

- React
- Vite
- ShadCN
- static build
- nginx container

No server-side rendering.

---

### 3.2 auth-service

Responsibilities:

- OIDC login
- JWT issuance
- refresh token sessions
- JWKS endpoint
- user identity
- external identity mapping

Schema:

    auth.user
    auth.external_identity
    auth.oidc_provider
    auth.refresh_token

Auth model:

- asymmetric JWT
- short-lived access token
- refresh token stored server-side

JWKS endpoint:

    /api/v1/auth/.well-known/jwks.json

Downstream services validate JWT using JWKS.

---

### 3.3 account-service

Responsibilities:

- tenant management
- household management
- membership roles
- user preferences
- active context

Schema:

    account.tenant
    account.household
    account.membership
    account.preference

Rules:

- one tenant per user initially
- multiple households allowed
- roles per household
- preference per tenant per user
- context persisted

Endpoint:

    /api/v1/contexts/current

---

### 3.4 productivity-service

Responsibilities:

- tasks
- reminders
- summary projections

Schema:

    productivity.task
    productivity.task_restoration
    productivity.reminder
    productivity.reminder_snooze
    productivity.reminder_dismissal

Rules:

- soft delete
- restore window
- fixed snooze durations
- derived summary
- tenant scoped
- household scoped

Summary endpoints return single resources.

---

## 4. API Design

### 4.1 Versioning

All endpoints are versioned.

    /api/v1/...

Versioning is required from the start.

---

### 4.2 Resource Style

JSON:API-style resources.

Examples:

    /tasks
    /tasks/{id}
    /tasks/restorations
    /reminders
    /reminders/snoozes
    /reminders/dismissals

Summary:

    /summary/tasks
    /summary/reminders
    /summary/dashboard

Context:

    /contexts/current

---

### 4.3 Includes

Supported for summary endpoints.

Example:

    /summary/dashboard?include=tasks,reminders

---

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

---

## 6. Multi-Tenancy Model

Hierarchy:

    User
      -> Tenant
           -> Household
                -> Membership

All data must include:

    tenant_id
    household_id

No global data.

---

## 7. Configuration Model

All services configured via environment variables.

Examples:

    DB_HOST
    DB_USER
    DB_PASSWORD
    OIDC_CLIENT_ID
    OIDC_SECRET
    JWT_PRIVATE_KEY

Secrets provided externally.

No config files required.

---

## 8. Persistence

Each service owns its schema.

    auth.*
    account.*
    productivity.*

No cross-service tables.

Migrations:

- per service
- run on startup
- forward-only acceptable

ORM:

- Gorm

IDs:

- UUID
- generated in application

---

## 9. Logging

Standard:

- Logrus
- structured JSON

Every request should include:

    request_id
    user_id
    tenant_id
    household_id

---

## 10. Tracing

OpenTelemetry used.

Required:

- trace id
- span id
- request id correlation

Metrics optional.

---

## 11. Local Runtime

Local environment uses:

- docker-compose
- nginx reverse proxy
- path prefix routing

Infra external:

- postgres
- redis (future)

Config via `.env`.

---

## 12. Kubernetes Runtime

Deployment:

- plain YAML
- no helm required

Images:

    ghcr.io/<owner>/home-hub-auth
    ghcr.io/<owner>/home-hub-account
    ghcr.io/<owner>/home-hub-productivity
    ghcr.io/<owner>/home-hub-frontend

Tags:

    main
    sha

Ingress uses path prefixes.

---

## 13. CI/CD

GitHub Actions.

PR:

- build impacted
- test impacted
- lint impacted

Main:

- publish snapshot images

Scripts used for builds.

---

## 14. Dependency Management

Renovate enabled.

Supports:

- Go modules
- npm
- GitHub Actions
- Docker

Automerge disabled.

---

## 15. API Testing

Bruno collections.

    bruno/
      auth/
      account/
      productivity/
      environments/

Used for manual endpoint testing.

---

## 16. Design Principles

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
