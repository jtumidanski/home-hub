# Home Hub

A multi-tenant household productivity platform built as a microservice system in Go with a React frontend.

## Architecture

Home Hub runs behind a reverse proxy with path-prefix routing. All services are stateless except for database persistence. Authentication uses asymmetric JWT with OIDC login, validated downstream via JWKS.

| Component | Description |
|---|---|
| **frontend** | React + Vite + shadcn/ui, served as a static build via nginx |
| **auth-service** | OIDC login, JWT issuance, refresh tokens, JWKS endpoint |
| **account-service** | Tenants, households, memberships, preferences, active context |
| **productivity-service** | Tasks, reminders, summary projections |
| **recipe-service** | Recipe storage and retrieval using Cooklang format |
| **calendar-service** | Google Calendar integration with per-user OAuth and household-merged view |
| **package-service** | Package delivery tracking across USPS, UPS, and FedEx with background polling |
| **weather-service** | Weather forecasts, current conditions, and geocoding via Open-Meteo |

Shared Go libraries under `shared/go/` provide auth, database, HTTP, logging, server lifecycle, tenant context, and test utilities. No business logic lives in shared code.

All APIs are versioned (`/api/v1/...`) and follow JSON:API resource style.

## Getting Started

Prerequisites: Go, Node, npm, Docker, Git.

```sh
# Build all services
./scripts/build-all.sh

# Start the stack (nginx + frontend + all services)
./scripts/local-up.sh

# Stop
./scripts/local-down.sh
```

The app is available at `http://localhost:8080`. Database is external — configure connection details in `.env` (see `.env.example`).

## Project Structure

```
frontend/                   React frontend
services/
  auth-service/             Authentication and identity
  account-service/          Households, tenants, memberships
  productivity-service/     Tasks and reminders
  recipe-service/           Recipe management with Cooklang parsing
  calendar-service/         Google Calendar integration
  package-service/          Package delivery tracking
  weather-service/          Weather forecasts and geocoding
shared/go/                  Shared Go libraries
deploy/
  compose/                  Docker Compose setup
  k8s/                      Kubernetes manifests
scripts/                    Build, test, and lint scripts
docs/                       Architecture and development guides
bruno/                      API test collections
```

## Development

```sh
./scripts/test-all.sh       # Run all tests
./scripts/lint-all.sh       # Run all linters
```

Frontend dev server: `cd frontend && npm install && npm run dev`

See [docs/development.md](docs/development.md) for the full development guide and [docs/architecture.md](docs/architecture.md) for system design details.

## Documentation

Each service maintains its own docs per the [DOCS.md](DOCS.md) contract:

- `services/<service>/docs/domain.md` — domain logic and invariants
- `services/<service>/docs/rest.md` — REST endpoint specs
- `services/<service>/docs/storage.md` — database schema
