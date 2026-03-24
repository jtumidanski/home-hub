# Home Hub — Development Guide

This document describes how to develop, run, and test the Home Hub monorepo.

It covers:

- local setup
- building services
- running the frontend
- running docker-compose
- adding a service
- adding migrations
- running tests
- running CI locally
- endpoint testing with Bruno

This guide assumes:

- Go installed
- Node installed
- Docker installed
- Bash available
- Git installed

---

## 1. Repository Layout

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
      compose/
      k8s/

    scripts/

    docs/

    bruno/

    .github/workflows/

    renovate.json

Each service is a separate Go module.

Shared code lives under shared/go.

Frontend lives under frontend.

---

## 2. Required Tools

Required:

- Go
- Node
- npm
- Docker
- Bash
- Git

Recommended:

- VSCode or GoLand
- Bruno
- kubectl

---

## 3. Go Workspace

The repo uses go.work.

To sync modules:

    go work sync

To build all modules:

    ./scripts/build-all.sh

---

## 4. Environment Variables

All services use environment variables.

Example:

    DB_HOST
    DB_USER
    DB_PASSWORD
    OIDC_CLIENT_ID
    OIDC_SECRET
    JWT_PRIVATE_KEY

Local development uses `.env`.

Do not commit real secrets.

---

## 5. Running Services Locally

Build everything:

    ./scripts/build-all.sh

Run compose:

    ./scripts/local-up.sh

Stop compose:

    ./scripts/local-down.sh

Compose runs:

- nginx
- frontend
- auth-service
- account-service
- productivity-service

Database is external.

---

## 6. Frontend Development

Frontend lives in:

    frontend/

Install:

    npm install

Run dev server:

    npm run dev

Build:

    npm run build

Frontend uses:

- React
- Vite
- ShadCN

No server-side rendering.

---

## 7. Building Services

Build all:

    ./scripts/build-all.sh

Build one:

    ./scripts/build-auth.sh
    ./scripts/build-account.sh
    ./scripts/build-productivity.sh

Each service must compile independently.

---

## 8. Testing

Run all tests:

    ./scripts/test-all.sh

Tests must exist for:

- new packages
- new handlers
- new business logic

Only unit tests required.

Integration tests optional.

---

## 9. Linting

Run lint:

    ./scripts/lint-all.sh

Backend lint:

- golangci-lint

Frontend lint:

- eslint

Formatting must pass before merge.

---

## 10. Adding a New Service

Create:

    services/new-service/

Required structure:

    go.mod
    cmd/
    internal/
    migrations/

Add to go.work.

Add Dockerfile.

Add compose entry.

Add k8s manifest.

Add CI detection rules.

---

## 11. Adding a Shared Package

Create:

    shared/go/<name>/

Add go.mod.

Add to go.work.

Do not put business logic in shared.

Shared is for:

- auth helpers
- http helpers
- logging
- testing

---

## 12. Migrations

Each service owns its schema.

Location:

    services/<service>/migrations/

Migrations run on startup.

Migrations must be:

- forward safe
- idempotent

Do not modify old migrations.

---

## 13. API Changes

Rules:

- API is versioned
- do not break v1
- add fields instead of removing
- keep JSON:API style

Endpoints:

    /api/v1/...

---

## 14. Endpoint Testing

Bruno collections live in:

    bruno/

Run Bruno to test endpoints.

Collections:

    auth
    account
    productivity

Environment files:

    local
    prod

Do not commit secrets.

---

## 15. CI Behavior

Pull request:

- build impacted services
- run tests
- run lint
- build frontend
- build docker

Main branch:

- publish snapshot images
- tag main
- tag sha

CI must pass before merge.

---

## 16. Docker Images

Images:

    ghcr.io/<owner>/home-hub-auth
    ghcr.io/<owner>/home-hub-account
    ghcr.io/<owner>/home-hub-productivity
    ghcr.io/<owner>/home-hub-frontend

Tags:

    main
    sha

Images built in CI.

---

## 17. Kubernetes Deployment

Manifests:

    deploy/k8s/

Plain YAML only.

Ingress uses path prefixes.

Secrets external.

---

## 18. Renovate

Renovate runs automatically.

Supports:

- Go
- npm
- GitHub Actions
- Docker

Do not manually update versions unless needed.

---

## 19. Logging

Use Logrus.

Logs must include:

    request_id
    user_id
    tenant_id
    household_id

No plain text logs.

---

## 20. Tracing

OpenTelemetry enabled.

Required:

- trace id
- span id

Metrics optional.

---

## 21. General Rules

- keep services small
- keep APIs versioned
- keep schemas separate
- do not bypass auth
- do not bypass tenant scope
- do not commit secrets
- keep CI green
