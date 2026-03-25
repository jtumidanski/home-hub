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

All services use environment variables. See `.env.example` for the full list with defaults.

| Variable | Used by | Required |
|---|---|---|
| `DB_HOST` | all services | yes |
| `DB_PORT` | all services | no (default: `5432`) |
| `DB_USER` | all services | yes |
| `DB_PASSWORD` | all services | yes |
| `DB_NAME` | all services | yes |
| `PORT` | all services | no (default: `8080`) |
| `JWKS_URL` | all except auth-service | no (default: internal cluster URL) |
| `JWT_PRIVATE_KEY` | auth-service | yes |
| `JWT_KEY_ID` | auth-service | no (default: `home-hub-1`) |
| `OIDC_CLIENT_ID` | auth-service | yes |
| `OIDC_CLIENT_SECRET` | auth-service | yes |
| `OIDC_ISSUER_URL` | auth-service | no (default: `https://accounts.google.com`) |
| `REFRESH_INTERVAL_MINUTES` | weather-service | no (default: `15`) |
| `CACHE_TTL_MINUTES` | weather-service | no (default: `15`) |

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

Manifests live in `deploy/k8s/`. Plain YAML only. All resources use the `home-hub` namespace. Ingress uses path prefixes with TLS termination.

### Applying manifests

```sh
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secrets.yaml
kubectl apply -f deploy/k8s/
```

### Secrets

Copy `deploy/k8s/secrets.example.yaml` to `deploy/k8s/secrets.yaml` and fill in real values. The example uses `stringData` so values are plain text (K8s base64-encodes them internally). Do not commit `secrets.yaml`.

Two Secret objects are required:
- `db-credentials` — database connection details
- `auth-secrets` — JWT key, OIDC client credentials, redirect URI

### TLS setup (self-signed for development)

The ingress expects a TLS secret named `home-hub-tls`. For local/dev testing with a self-signed certificate:

```sh
# 1. Generate a self-signed cert into deploy/k8s/tls/
mkdir -p deploy/k8s/tls
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deploy/k8s/tls/tls.key -out deploy/k8s/tls/tls.crt \
  -subj "/CN=homehub.tumidanski.me"

# 2. Create the K8s TLS secret
kubectl -n home-hub create secret tls home-hub-tls \
  --cert=deploy/k8s/tls/tls.crt --key=deploy/k8s/tls/tls.key
```

The `deploy/k8s/tls/` directory is gitignored. Keep the files around to recreate the secret if needed.

Your browser will show a certificate warning — accept it to proceed.

Ensure `OIDC_REDIRECT_URI` in `secrets.yaml` and the Google Cloud Console both use the `https://` scheme to match.

For production, replace the self-signed cert with cert-manager + Let's Encrypt.

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
