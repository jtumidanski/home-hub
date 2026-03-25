# Home Hub — Development Guide

This document describes how to develop, run, and test the Home Hub monorepo. It covers local setup, building services, running the frontend, docker-compose, adding services/migrations, running tests, CI, and endpoint testing with Bruno.

**Prerequisites:** Go, Node, npm, Docker, Bash, Git.

## 1. Repository Layout

```
go.work

frontend/

services/
  auth-service/
  account-service/
  productivity-service/

shared/go/
  auth/
  database/
  http/
  logging/
  model/
  server/
  tenant/
  testing/

deploy/
  compose/
  k8s/

scripts/
dev/
docs/
bruno/
.github/workflows/
renovate.json
```

Each service is a separate Go module. Shared code lives under `shared/go`. Frontend lives under `frontend/`.

## 2. Required Tools

**Required:** Go, Node, npm, Docker, Bash, Git

**Recommended:** VSCode or GoLand, Bruno, kubectl

## 3. Go Workspace

The repo uses `go.work`.

```sh
go work sync         # sync modules
./scripts/build-all.sh  # build all modules
```

## 4. Environment Variables

All services use environment variables:

```
DB_HOST
DB_USER
DB_PASSWORD
OIDC_CLIENT_ID
OIDC_SECRET
JWT_PRIVATE_KEY
```

Local development uses `.env`. Do not commit real secrets.

## 5. Running Services Locally

```sh
./scripts/build-all.sh   # build everything
./scripts/local-up.sh    # run compose
./scripts/local-down.sh  # stop compose
```

Compose runs: nginx, frontend, auth-service, account-service, productivity-service. Database is external.

## 6. Frontend Development

Frontend lives in `frontend/`.

```sh
npm install    # install dependencies
npm run dev    # run dev server
npm run build  # production build
```

Uses React, Vite, ShadCN. No server-side rendering.

## 7. Building Services

```sh
./scripts/build-all.sh          # build all
./scripts/build-auth.sh         # build auth-service
./scripts/build-account.sh      # build account-service
./scripts/build-productivity.sh # build productivity-service
```

Each service must compile independently.

## 8. Testing

```sh
./scripts/test-all.sh  # run all tests
```

Tests must exist for new packages, handlers, and business logic. Only unit tests required; integration tests optional.

## 9. Linting

```sh
./scripts/lint-all.sh  # run all linters
```

- Backend: golangci-lint
- Frontend: eslint

Formatting must pass before merge.

## 10. Adding a New Service

1. Create `services/new-service/` with required structure:

   ```
   go.mod
   cmd/
   internal/
   migrations/
   ```

2. Add to `go.work`
3. Add Dockerfile
4. Add compose entry
5. Add k8s manifest
6. Add CI detection rules

## 11. Adding a Shared Package

1. Create `shared/go/<name>/`
2. Add `go.mod`
3. Add to `go.work`

Do not put business logic in shared. Shared is for: auth helpers, database helpers, http helpers, logging, model types, server lifecycle, tenant context, testing.

## 12. Migrations

Each service owns its schema. Migrations are defined in each domain's `entity.go` via GORM AutoMigrate. There are no separate SQL migration files.

Migrations run on startup and must be:

- forward safe
- idempotent

Do not modify old entity migration definitions.

## 13. API Changes

Rules:

- API is versioned (`/api/v1/...`)
- do not break v1
- add fields instead of removing
- keep JSON:API style

## 14. Endpoint Testing

Bruno collections live in `bruno/` with collections for `auth`, `account`, `productivity` and environment files for `local` and `prod`.

Run Bruno to test endpoints. Do not commit secrets.

## 15. CI Behavior

**Pull request:** build impacted services, run tests, run lint, build frontend, build docker.

**Main branch:** publish snapshot images, tag `main`, tag `sha`.

CI must pass before merge.

## 16. Docker Images

```
ghcr.io/<owner>/home-hub-auth
ghcr.io/<owner>/home-hub-account
ghcr.io/<owner>/home-hub-productivity
ghcr.io/<owner>/home-hub-frontend
```

Tags: `main`, `sha`. Images built in CI.

## 17. Kubernetes Deployment

Manifests live in `deploy/k8s/`. Plain YAML only. Ingress uses path prefixes. Secrets external.

## 18. Renovate

Renovate runs automatically. Supports: Go, npm, GitHub Actions, Docker. Do not manually update versions unless needed.

## 19. Logging

Use Logrus with structured JSON. Logs must include: `request_id`, `user_id`, `tenant_id`, `household_id`. No plain text logs.

## 20. Tracing

OpenTelemetry enabled. Required: trace id, span id. Metrics optional.

## 21. Service Documentation

Each service maintains documentation per the DOCS.md contract:

```
services/<service>/docs/domain.md
services/<service>/docs/rest.md
services/<service>/docs/storage.md
```

Update documentation when adding or modifying domains, endpoints, or schema. See DOCS.md for the full documentation contract.

## 22. General Rules

- keep services small
- keep APIs versioned
- keep schemas separate
- do not bypass auth
- do not bypass tenant scope
- do not commit secrets
- keep CI green
