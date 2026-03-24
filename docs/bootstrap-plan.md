# Home Hub — Bootstrap Implementation Plan

This document defines the milestone-based bootstrap plan for the Home Hub monorepo.

The goal is to start with a mature structure from the beginning, including:

- CI/CD
- deployment manifests
- dependency management
- endpoint testing
- strict service boundaries
- versioned APIs

Core components:

- frontend
- auth-service
- account-service
- productivity-service
- shared Go modules
- docker-compose local runtime
- k3s deployment manifests
- GitHub Actions
- Renovate
- Bruno

All APIs are versioned from the start.

---

## Milestone 1 — Monorepo Scaffold

Create the base repository layout.

### Structure

    go.work

    frontend/

    services/
      auth-service/
      account-service/
      productivity-service/

    shared/go/
      auth/
      http/
      logging/
      testing/

    deploy/
      k8s/
      compose/

    scripts/

    docs/

    bruno/

    .github/workflows/

    renovate.json

Each service:

    go.mod
    cmd/
    internal/
    migrations/

Frontend:

    React
    Vite
    ShadCN

Docs:

    README.md
    docs/architecture.md
    docs/development.md
    docs/decisions.md
    docs/bootstrap-plan.md

Acceptance Criteria:

- repo builds
- go.work resolves modules
- frontend runs
- services compile
- renovate.json exists
- bruno exists
- scripts exists

---

## Milestone 2 — Common Service Baseline

Create shared patterns.

Requirements:

- env-based config
- Logrus structured logging
- OpenTelemetry tracing
- request id middleware
- net/http
- Gorm
- UUID generated in app
- migrations on startup
- /healthz
- /readyz
- JWT validation helper

Acceptance Criteria:

- services share bootstrap
- logs structured
- traces initialized
- migrations run automatically

---

## Milestone 3 — CI Foundation

Scripts:

    scripts/build-all.sh
    scripts/test-all.sh
    scripts/lint-all.sh
    scripts/build-auth.sh
    scripts/build-account.sh
    scripts/build-productivity.sh
    scripts/build-frontend.sh
    scripts/local-up.sh
    scripts/local-down.sh
    scripts/ci-build.sh
    scripts/ci-test.sh

PR workflow:

- detect changed paths
- build impacted
- test impacted
- lint impacted
- build frontend
- build docker
- validate yaml

Main workflow:

- build images
- push GHCR
- tag main
- tag sha

Acceptance:

- PR blocked if failing
- main publishes snapshots

---

## Milestone 4 — Auth Service

Features:

- Google OIDC
- asymmetric JWT
- JWKS endpoint
- refresh sessions
- cookies

Endpoints:

    /api/v1/auth/providers
    /api/v1/auth/login/{provider}
    /api/v1/auth/callback/{provider}
    /api/v1/auth/token/refresh
    /api/v1/auth/logout
    /api/v1/auth/.well-known/jwks.json
    /api/v1/users/me

Schema:

    auth.user
    auth.external_identity
    auth.oidc_provider
    auth.refresh_token

Acceptance:

- login works
- repeat login works
- refresh works
- logout works
- JWKS works

---

## Milestone 5 — Account Service

Schema:

    account.tenant
    account.household
    account.membership
    account.preference

Endpoints:

    /api/v1/tenants
    /api/v1/households
    /api/v1/memberships
    /api/v1/preferences
    /api/v1/contexts/current

Rules:

- owner creates households
- roles per household
- context persisted
- theme persisted

Acceptance:

- tenant create
- household create
- preference works
- context works

---

## Milestone 6 — Frontend Auth + Onboarding

Routes:

    /login
    /onboarding
    /app
    /app/tasks
    /app/reminders
    /app/settings
    /app/households

Bootstrap:

    GET /api/v1/users/me
    GET /api/v1/contexts/current

Features:

- ShadCN
- theme toggle
- household switch
- onboarding

Acceptance:

- login works
- onboarding works
- app loads
- theme saved
- context saved

---

## Milestone 7 — Productivity Service

Schema:

    productivity.task
    productivity.task_restoration
    productivity.reminder
    productivity.reminder_snooze
    productivity.reminder_dismissal

Endpoints:

    /api/v1/tasks
    /api/v1/tasks/restorations
    /api/v1/reminders
    /api/v1/reminders/snoozes
    /api/v1/reminders/dismissals
    /api/v1/summary/tasks
    /api/v1/summary/reminders
    /api/v1/summary/dashboard

Rules:

- soft delete
- restore window
- fixed snooze
- derived summary
- tenant scoped
- household scoped

Acceptance:

- CRUD works
- restore works
- snooze works
- summary works

---

## Milestone 8 — Frontend Productivity UI

Pages:

- dashboard
- tasks
- reminders
- settings
- households

Acceptance:

- tasks UI works
- reminders UI works
- dashboard works
- include works
- context works

---

## Milestone 9 — Local Environment

Compose:

- nginx
- frontend
- auth
- account
- productivity

External infra via env.

Acceptance:

- local-up works
- routing works
- login works
- frontend works

---

## Milestone 10 — k3s Deployment

Plain YAML:

    deploy/k8s/
      auth-service.yaml
      account-service.yaml
      productivity-service.yaml
      frontend.yaml
      ingress.yaml

Acceptance:

- deploy works
- ingress works
- secrets external
- health checks used

---

## Milestone 11 — Bruno

    bruno/
      auth/
      account/
      productivity/
      environments/

Must cover:

- identity
- tenant
- household
- preference
- context
- tasks
- reminders
- summary

Acceptance:

- endpoints testable

---

## Milestone 12 — Renovate + Maturity

`renovate.json` must support:

- Go
- npm
- actions
- docker

Acceptance:

- renovate PRs open
- docs accurate
- CI strict

---

## Execution Order

1. Scaffold
2. Baseline
3. CI
4. Auth
5. Account
6. Frontend auth
7. Productivity
8. Frontend UI
9. Local
10. k8s
11. Bruno
12. Renovate
