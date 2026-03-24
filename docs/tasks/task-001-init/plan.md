# task-001-init — Implementation Plan

Last Updated: 2026-03-24

---

## Executive Summary

Bootstrap the Home Hub monorepo from zero to a fully functional multi-tenant household productivity platform. This covers 12 milestones: repo scaffold, shared service baseline, CI/CD, three backend services (auth, account, productivity), frontend (auth + productivity UI), local Docker Compose environment, k3s deployment manifests, Bruno API collections, and Renovate configuration.

The design phase is complete — PRD, architecture, API spec, schema, bootstrap plan, and development guide all exist. No code exists yet.

---

## Current State Analysis

**What exists:**
- Product Requirements Document (PRD + Appendix)
- Architecture document
- Bootstrap plan (12 milestones defined)
- Database schema specification
- Full API specification (JSON:API)
- Development guide
- Decisions log (template only)
- Claude development skills and commands

**What does not exist:**
- Any Go code
- Any frontend code
- go.work or go.mod files
- Docker Compose or k8s manifests
- CI/CD workflows
- Bruno collections
- Scripts
- renovate.json

---

## Proposed Future State

A working monorepo with:
- 3 Go microservices (auth, account, productivity) with full CRUD
- React/Vite/ShadCN frontend with auth flow, onboarding, and productivity UI
- Shared Go modules (auth, http, logging, testing)
- PostgreSQL with per-service schemas and startup migrations
- Docker Compose local environment with nginx reverse proxy
- k3s deployment manifests
- GitHub Actions CI (PR + main branch)
- Bruno API collections for all endpoints
- Renovate dependency management

---

## Implementation Phases

### Phase 1 — Monorepo Scaffold (Milestone 1)

Establish the directory structure, Go workspace, and frontend skeleton.

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 1.1 | Create top-level go.work | S | go.work references all service and shared modules |
| 1.2 | Create services/auth-service skeleton (go.mod, cmd/, internal/) | S | `go build ./...` succeeds |
| 1.3 | Create services/account-service skeleton | S | `go build ./...` succeeds |
| 1.4 | Create services/productivity-service skeleton | S | `go build ./...` succeeds |
| 1.5 | Create shared/go/auth module | S | go.mod exists, compiles |
| 1.6 | Create shared/go/http module | S | go.mod exists, compiles |
| 1.7 | Create shared/go/logging module | S | go.mod exists, compiles |
| 1.8 | Create shared/go/testing module | S | go.mod exists, compiles |
| 1.9 | Initialize frontend (React + Vite + ShadCN) | M | `npm run dev` serves app, `npm run build` succeeds |
| 1.10 | Create deploy/compose/ and deploy/k8s/ directories | S | Directories exist with placeholder files |
| 1.11 | Create scripts/ directory with placeholder scripts | S | Scripts exist and are executable |
| 1.12 | Create bruno/ directory structure | S | auth/, account/, productivity/, environments/ exist |
| 1.13 | Verify `go work sync` succeeds | S | All modules resolve |

**Dependencies:** None — this is the starting point.

---

### Phase 2 — Common Service Baseline (Milestone 2)

Build shared infrastructure that all services use.

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 2.1 | Implement env-based config loading (shared/go/http or per-service) | M | Services read DB_HOST, DB_USER, etc. from environment |
| 2.2 | Implement Logrus structured JSON logging (shared/go/logging) | M | Logs output JSON with request_id, user_id, tenant_id, household_id |
| 2.3 | Implement OpenTelemetry tracing initialization | M | Trace ID and span ID present in request context |
| 2.4 | Implement request ID middleware | S | Every request gets a unique request_id in context and logs |
| 2.5 | Implement GORM database connection setup | M | Services connect to PostgreSQL, schemas are isolated |
| 2.6 | Implement UUID generation in application layer | S | IDs are UUIDs generated before insert |
| 2.7 | Implement startup migration runner | M | Migrations in services/*/migrations/ run automatically on boot |
| 2.8 | Implement /healthz and /readyz endpoints | S | Return 200 when service is healthy/ready |
| 2.9 | Implement JWT validation helper (shared/go/auth) | M | Validates JWT against JWKS, extracts claims |
| 2.10 | Implement tenant context middleware | M | Extracts tenant_id, household_id from JWT claims into request context |
| 2.11 | Implement service bootstrap pattern (main.go template) | M | Standardized startup: config → DB → migrations → middleware → routes → serve |

**Dependencies:** Phase 1 complete.

---

### Phase 3 — CI Foundation (Milestone 3)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 3.1 | Create build scripts (build-all, build-auth, build-account, build-productivity, build-frontend) | M | Each script builds its target successfully |
| 3.2 | Create test scripts (test-all) | S | Runs go test for all services and shared modules |
| 3.3 | Create lint scripts (lint-all) | S | Runs golangci-lint + eslint |
| 3.4 | Create local-up.sh and local-down.sh | S | Starts/stops Docker Compose |
| 3.5 | Create PR GitHub Actions workflow | L | Detects changed paths, builds/tests/lints impacted services, builds Docker images |
| 3.6 | Create main branch GitHub Actions workflow | M | Builds and publishes snapshot images to GHCR |
| 3.7 | Create Dockerfiles for each service and frontend | M | Multi-stage builds, images are minimal |

**Dependencies:** Phase 1 complete. Phase 2 partially (services must compile).

---

### Phase 4 — Auth Service (Milestone 4)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 4.1 | Create auth schema migrations (users, external_identities, oidc_providers, refresh_tokens) | M | Tables created on startup in auth schema |
| 4.2 | Implement GORM entities for auth tables | M | Entities map to schema correctly |
| 4.3 | Implement domain models (User, ExternalIdentity, OIDCProvider, RefreshToken) | M | Immutable models with builders |
| 4.4 | Implement OIDC provider management (Google) | L | Provider configuration loads, redirects work |
| 4.5 | Implement OIDC login flow (GET /auth/login/{provider}) | L | Redirects to provider with correct state |
| 4.6 | Implement OIDC callback (GET /auth/callback/{provider}) | XL | Exchanges code, resolves/creates user, maps external identity |
| 4.7 | Implement asymmetric JWT issuance | L | Access tokens signed with RS256, short-lived |
| 4.8 | Implement JWKS endpoint (GET /auth/.well-known/jwks.json) | M | Returns public keys in JWKS format |
| 4.9 | Implement refresh token sessions | L | Server-side storage, rotation, revocation |
| 4.10 | Implement HTTP-only cookie handling | M | Access + refresh tokens in secure HTTP-only cookies |
| 4.11 | Implement token refresh (POST /auth/token/refresh) | M | Rotates refresh, issues new access token |
| 4.12 | Implement logout (POST /auth/logout) | S | Revokes refresh session, clears cookies |
| 4.13 | Implement GET /users/me | S | Returns identity-only user resource (JSON:API) |
| 4.14 | Implement GET /auth/providers | S | Returns enabled providers list |
| 4.15 | Write unit tests for auth domain logic | M | Token generation, validation, session management tested |

**Dependencies:** Phase 2 complete.

---

### Phase 5 — Account Service (Milestone 5)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 5.1 | Create account schema migrations (tenants, households, memberships, preferences) | M | Tables created on startup in account schema |
| 5.2 | Implement GORM entities | M | Entities map to schema correctly |
| 5.3 | Implement domain models (Tenant, Household, Membership, Preference) | M | Immutable models with invariants |
| 5.4 | Implement tenant CRUD (GET/POST /tenants, GET /tenants/{id}) | M | One tenant per user initially |
| 5.5 | Implement household CRUD (GET/POST /households, GET/PATCH /households/{id}) | M | Households scoped to tenant, owner creates |
| 5.6 | Implement membership management (GET/POST /memberships, PATCH/DELETE /memberships/{id}) | L | Roles enforced, unique per household+user |
| 5.7 | Implement preference management (GET/PATCH /preferences) | M | Theme + active household, one per user per tenant |
| 5.8 | Implement GET /contexts/current | L | Derives resolved context with tenant, household, role, theme, memberships |
| 5.9 | Implement context includes support | M | ?include=tenant,activeHousehold,preference,memberships |
| 5.10 | Implement fallback when active household is invalid | M | Returns selection-required state or resolves safely |
| 5.11 | Write unit tests for account domain logic | M | Tenant/household/membership invariants tested |

**Dependencies:** Phase 2 complete. Can parallelize with Phase 4.

---

### Phase 6 — Frontend Auth + Onboarding (Milestone 6)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 6.1 | Set up React Router with routes (/login, /onboarding, /app/*) | M | Routes render correct components |
| 6.2 | Implement JSON:API client | M | Typed client for data/attributes/relationships responses |
| 6.3 | Implement login page (provider selection, redirect) | M | Clicking provider initiates OIDC flow |
| 6.4 | Implement auth bootstrap (GET /users/me + GET /contexts/current) | M | App determines auth state on load |
| 6.5 | Implement onboarding flow (create tenant → create household) | L | New users guided through setup |
| 6.6 | Implement ShadCN theme toggle (light/dark) | M | Theme persisted via PATCH /preferences |
| 6.7 | Implement household switcher | M | Switches active household via PATCH /preferences, refetches context |
| 6.8 | Implement protected route wrapper | S | Unauthenticated users redirected to /login |
| 6.9 | Implement app shell (sidebar/nav, layout) | M | Consistent layout for /app/* routes |

**Dependencies:** Phase 4 and Phase 5 complete (APIs must exist).

---

### Phase 7 — Productivity Service (Milestone 7)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 7.1 | Create productivity schema migrations (tasks, task_restorations, reminders, reminder_snoozes, reminder_dismissals) | M | Tables created on startup in productivity schema |
| 7.2 | Implement GORM entities | M | Entities map to schema correctly |
| 7.3 | Implement domain models (Task, Reminder, etc.) | M | Immutable models, status transitions, soft delete |
| 7.4 | Implement task CRUD (GET/POST /tasks, GET/PATCH/DELETE /tasks/{id}) | L | Filters, sorts, soft delete, status transitions |
| 7.5 | Implement task restoration (POST /tasks/restorations) | M | 3-day window enforced, clears deleted_at |
| 7.6 | Implement reminder CRUD (GET/POST /reminders, GET/PATCH/DELETE /reminders/{id}) | L | Filters, sorts, active state derived |
| 7.7 | Implement reminder snooze (POST /reminders/snoozes) | M | Fixed durations (10/30/60), snoozed_until computed |
| 7.8 | Implement reminder dismissal (POST /reminders/dismissals) | M | Updates dismissal state |
| 7.9 | Implement task summary (GET /summary/tasks) | M | Pending, completed today, overdue counts derived |
| 7.10 | Implement reminder summary (GET /summary/reminders) | M | Due now, upcoming, snoozed counts derived |
| 7.11 | Implement dashboard summary (GET /summary/dashboard) | M | Aggregates task + reminder summaries, includes household info |
| 7.12 | Implement include support for summary endpoints | M | ?include=tasks, ?include=reminders |
| 7.13 | Write unit tests for productivity domain logic | M | Status transitions, restore window, snooze durations tested |

**Dependencies:** Phase 2 complete. Can parallelize with Phases 4-5.

---

### Phase 8 — Frontend Productivity UI (Milestone 8)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 8.1 | Implement dashboard page | L | Shows task/reminder summary, household name |
| 8.2 | Implement tasks list page | L | Filter by status, sort by due date, create/edit/delete |
| 8.3 | Implement task detail/edit view | M | Update title, notes, due date, status |
| 8.4 | Implement task restore UI | S | Restore soft-deleted tasks within window |
| 8.5 | Implement reminders list page | L | Filter by active, sort by scheduled time |
| 8.6 | Implement reminder detail/edit view | M | Update title, notes, scheduled time |
| 8.7 | Implement snooze UI | S | Select from 10/30/60 minute options |
| 8.8 | Implement dismiss UI | S | One-click dismiss |
| 8.9 | Implement settings page | M | Theme, household management links |
| 8.10 | Implement households management page | M | List households, create new (if owner) |

**Dependencies:** Phase 6 and Phase 7 complete.

---

### Phase 9 — Local Environment (Milestone 9)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 9.1 | Create docker-compose.yml (nginx, frontend, auth, account, productivity) | L | `docker compose up` starts all services |
| 9.2 | Create nginx.conf with path-prefix routing | M | Routes match architecture doc |
| 9.3 | Create .env.example | S | All required env vars documented |
| 9.4 | Verify end-to-end login flow locally | M | Browser → login → callback → onboarding → dashboard |
| 9.5 | Verify all API endpoints work through proxy | M | Bruno collections pass against local env |

**Dependencies:** Phases 4-8 complete.

---

### Phase 10 — k3s Deployment (Milestone 10)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 10.1 | Create k8s Deployment + Service YAML for auth-service | M | Deploys, health checks work |
| 10.2 | Create k8s Deployment + Service YAML for account-service | M | Deploys, health checks work |
| 10.3 | Create k8s Deployment + Service YAML for productivity-service | M | Deploys, health checks work |
| 10.4 | Create k8s Deployment + Service YAML for frontend | M | Deploys, serves UI |
| 10.5 | Create Ingress YAML with path-prefix routing | M | Routes match architecture doc |
| 10.6 | Document secret management approach | S | Secrets external, not in manifests |

**Dependencies:** Phase 9 complete (local env validates the services work).

---

### Phase 11 — Bruno Collections (Milestone 11)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 11.1 | Create auth collection (providers, login, callback, refresh, logout, JWKS, users/me) | M | All auth endpoints testable |
| 11.2 | Create account collection (tenants, households, memberships, preferences, contexts/current) | M | All account endpoints testable |
| 11.3 | Create productivity collection (tasks, restorations, reminders, snoozes, dismissals, summaries) | M | All productivity endpoints testable |
| 11.4 | Create environment files (local, prod) | S | Variables configured for each env |

**Dependencies:** Services must have working endpoints. Can build incrementally alongside Phases 4-7.

---

### Phase 12 — Renovate + Maturity (Milestone 12)

| # | Task | Effort | Acceptance Criteria |
|---|------|--------|-------------------|
| 12.1 | Create renovate.json (Go, npm, GitHub Actions, Docker) | S | Renovate opens PRs for dependency updates |
| 12.2 | Verify all docs are accurate against implementation | M | Architecture, API, schema docs match code |
| 12.3 | Verify CI is strict (PR must pass before merge) | S | Branch protection configured |

**Dependencies:** All prior phases complete.

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| OIDC provider integration complexity | High | Medium | Start with Google only, abstract provider interface |
| JWT/JWKS setup across services | High | Medium | Build shared/go/auth early, test with integration tests |
| Cross-service context resolution (no direct DB access) | Medium | Medium | Account service context endpoint must be reliable; frontend caches |
| Migration ordering (schemas must exist before services boot) | Medium | Low | Each service creates its own schema on startup |
| Docker Compose networking between services | Medium | Low | Use service names as hostnames, test early |
| Scope creep on frontend | Medium | Medium | Stick to PRD non-goals list strictly |

---

## Success Metrics

- All 3 backend services build and pass tests independently
- Frontend builds and serves
- End-to-end login → onboarding → dashboard flow works locally
- All API endpoints match the API specification
- Docker Compose `local-up.sh` starts a working system
- k8s manifests deploy successfully to k3s
- CI passes on PR and publishes images on main
- Bruno collections cover all endpoints
- No cross-schema foreign keys
- All rows include tenant_id (and household_id where applicable)

---

## Required Resources and Dependencies

**External:**
- PostgreSQL instance (local or Docker)
- Google Cloud OIDC credentials (for auth)
- GHCR access for image publishing
- k3s cluster for deployment testing

**Internal:**
- Go 1.22+
- Node 20+
- Docker
- golangci-lint
- Bruno

---

## Timeline Estimates

| Phase | Effort |
|-------|--------|
| 1. Scaffold | S |
| 2. Baseline | L |
| 3. CI | L |
| 4. Auth | XL |
| 5. Account | L |
| 6. Frontend Auth | L |
| 7. Productivity | XL |
| 8. Frontend UI | XL |
| 9. Local Env | M |
| 10. k3s | M |
| 11. Bruno | M |
| 12. Renovate | S |
