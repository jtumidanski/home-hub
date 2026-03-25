# Plan Audit — task-001-init

**Plan Path:** docs/tasks/task-001-init/tasks.md
**Audit Date:** 2026-03-25
**Branch:** bug-fixes
**Base Branch:** main

## Executive Summary

103 of 107 tasks are completed (96.3%). The 4 incomplete tasks are manual verification steps (9.4, 9.5, 12.2, 12.3) that require a running environment or GitHub admin access. All 3 backend services and the frontend build successfully, and all Go tests pass. The frontend has 1 failing test (255/256 pass) caused by a test mock not updated after a bug-fix commit added a `slug` field. Code on the `bug-fixes` branch follows backend and frontend developer guidelines with minor observations noted below.

---

## Task Completion

### Phase 1 — Monorepo Scaffold

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Create top-level go.work | DONE | `go.work` references 3 services + 8 shared modules |
| 1.2 | Create services/auth-service skeleton | DONE | `services/auth-service/cmd/main.go`, `internal/` packages |
| 1.3 | Create services/account-service skeleton | DONE | `services/account-service/cmd/main.go`, `internal/` packages |
| 1.4 | Create services/productivity-service skeleton | DONE | `services/productivity-service/cmd/main.go`, `internal/` packages |
| 1.5 | Create shared/go/auth module | DONE | `shared/go/auth/go.mod`, `auth.go`, `jwks.go` |
| 1.6 | Create shared/go/http module | DONE | `shared/go/http/go.mod`, `http.go` |
| 1.7 | Create shared/go/logging module | DONE | `shared/go/logging/go.mod`, `logging.go`, `tracing.go` |
| 1.8 | Create shared/go/testing module | DONE | `shared/go/testing/go.mod`, `testing.go` |
| 1.9 | Initialize frontend (React + Vite + ShadCN) | DONE | `frontend/package.json`, React 19, Vite 8, ShadCN components in `components/ui/` |
| 1.10 | Create deploy/compose/ and deploy/k8s/ directories | DONE | `deploy/compose/docker-compose.yml`, `deploy/k8s/*.yaml` |
| 1.11 | Create scripts/ directory | DONE | 11 scripts (build-*.sh, test-all.sh, lint-all.sh, local-up.sh, local-down.sh, ci-*.sh) |
| 1.12 | Create bruno/ directory structure | DONE | `bruno/auth/`, `bruno/account/`, `bruno/productivity/`, `bruno/environments/` |
| 1.13 | Verify `go work sync` succeeds | DONE | `go.work.sum` exists; all services build successfully |

### Phase 2 — Common Service Baseline

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 2.1 | Implement env-based config loading | DONE | `services/*/internal/config/config.go` reads env vars |
| 2.2 | Implement Logrus structured JSON logging | DONE | `shared/go/logging/logging.go` |
| 2.3 | Implement OpenTelemetry tracing initialization | DONE | `shared/go/logging/tracing.go` |
| 2.4 | Implement request ID middleware | DONE | `shared/go/server/middleware.go` |
| 2.5 | Implement GORM database connection setup | DONE | `shared/go/database/database.go` with schema isolation |
| 2.6 | Implement UUID generation in application layer | DONE | Builders generate UUIDs; entity.go files have no UUID generation |
| 2.7 | Implement startup migration runner | DONE | `database.SetMigrations()` + `entity.Migration()` per domain |
| 2.8 | Implement /healthz and /readyz endpoints | DONE | `shared/go/server/health.go` |
| 2.9 | Implement JWT validation helper | DONE | `shared/go/auth/auth.go` with JWKS validation |
| 2.10 | Implement tenant context middleware | DONE | `shared/go/auth/auth.go:Middleware`, `shared/go/tenant/context.go` |
| 2.11 | Implement service bootstrap pattern | DONE | All 3 services follow config → DB → migrations → middleware → routes → serve |

### Phase 3 — CI Foundation

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 3.1 | Create build scripts | DONE | `scripts/build-all.sh`, `build-auth.sh`, `build-account.sh`, `build-productivity.sh`, `build-frontend.sh` |
| 3.2 | Create test scripts | DONE | `scripts/test-all.sh` |
| 3.3 | Create lint scripts | DONE | `scripts/lint-all.sh` |
| 3.4 | Create local-up.sh and local-down.sh | DONE | `scripts/local-up.sh`, `scripts/local-down.sh` (with --env-file fix on this branch) |
| 3.5 | Create PR GitHub Actions workflow | DONE | `.github/workflows/pr.yml` with change detection |
| 3.6 | Create main branch GitHub Actions workflow | DONE | `.github/workflows/main.yml` builds + pushes to GHCR |
| 3.7 | Create Dockerfiles for each service | DONE | `services/auth-service/Dockerfile`, `services/account-service/Dockerfile`, `services/productivity-service/Dockerfile` |

### Phase 4 — Auth Service

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 4.1 | Create auth schema migrations | DONE | `user.Migration`, `externalidentity.Migration`, `oidcprovider.Migration`, `refreshtoken.Migration` in entity.go files |
| 4.2 | Implement GORM entities | DONE | Entity structs with GORM tags in each domain's entity.go |
| 4.3 | Implement domain models | DONE | `model.go` + `builder.go` in user, externalidentity, oidcprovider, refreshtoken |
| 4.4 | Implement OIDC provider management | DONE | `internal/oidcprovider/` package, `internal/oidc/oidc.go` |
| 4.5 | Implement OIDC login flow | DONE | `internal/authflow/processor.go` |
| 4.6 | Implement OIDC callback | DONE | `internal/authflow/processor.go` |
| 4.7 | Implement asymmetric JWT issuance | DONE | `internal/jwt/issuer.go` |
| 4.8 | Implement JWKS endpoint | DONE | `internal/jwt/jwks.go` |
| 4.9 | Implement refresh token sessions | DONE | `internal/refreshtoken/` package with provider, processor, administrator |
| 4.10 | Implement HTTP-only cookie handling | DONE | Cookie management in authflow processor |
| 4.11 | Implement token refresh | DONE | `POST /auth/token/refresh` in resource.go |
| 4.12 | Implement logout | DONE | `POST /auth/logout` in resource.go |
| 4.13 | Implement GET /users/me | DONE | `internal/user/resource.go` |
| 4.14 | Implement GET /auth/providers | DONE | `internal/oidcprovider/resource.go` |
| 4.15 | Write unit tests | DONE | Tests in authflow, externalidentity, jwt, oidc, oidcprovider, refreshtoken, user packages |

### Phase 5 — Account Service

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 5.1 | Create account schema migrations | DONE | `tenant.Migration`, `household.Migration`, `membership.Migration`, `preference.Migration` |
| 5.2 | Implement GORM entities | DONE | Entity structs in each domain's entity.go |
| 5.3 | Implement domain models | DONE | model.go + builder.go per domain |
| 5.4 | Implement tenant CRUD | DONE | `internal/tenant/` package with resource.go |
| 5.5 | Implement household CRUD | DONE | `internal/household/` package with resource.go |
| 5.6 | Implement membership management | DONE | `internal/membership/` package with resource.go |
| 5.7 | Implement preference management | DONE | `internal/preference/` package with resource.go |
| 5.8 | Implement GET /contexts/current | DONE | `internal/appcontext/context.go` + `resource.go` |
| 5.9 | Implement context includes support | DONE | `internal/appcontext/rest.go` with include parameters |
| 5.10 | Implement fallback when active household is invalid | DONE | Logic in `appcontext/context.go` resolves safely |
| 5.11 | Write unit tests | DONE | Tests in appcontext, household, membership, preference, tenant packages |

### Phase 6 — Frontend Auth + Onboarding

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 6.1 | Set up React Router with routes | DONE | `App.tsx` defines /login, /onboarding, /app/* routes |
| 6.2 | Implement JSON:API client | DONE | `lib/api/client.ts` with dedup, caching, retry |
| 6.3 | Implement login page | DONE | `pages/LoginPage.tsx` |
| 6.4 | Implement auth bootstrap | DONE | `lib/hooks/api/use-auth.ts`, `components/providers/auth-provider.tsx` |
| 6.5 | Implement onboarding flow | DONE | `pages/OnboardingPage.tsx` with 2-step tenant→household flow |
| 6.6 | Implement ShadCN theme toggle | DONE | `lib/hooks/use-theme-toggle.ts`, `components/providers/theme-provider.tsx` |
| 6.7 | Implement household switcher | DONE | `components/features/households/household-switcher.tsx` |
| 6.8 | Implement protected route wrapper | DONE | `components/features/navigation/protected-route.tsx` |
| 6.9 | Implement app shell | DONE | `components/features/navigation/app-shell.tsx` |

### Phase 7 — Productivity Service

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 7.1 | Create productivity schema migrations | DONE | `task.Migration`, `restoration.Migration`, `reminder.Migration`, `snooze.Migration`, `dismissal.Migration` |
| 7.2 | Implement GORM entities | DONE | Entity structs in each domain/subdomain entity.go |
| 7.3 | Implement domain models | DONE | model.go + builder.go per domain (task, reminder, restoration, snooze, dismissal) |
| 7.4 | Implement task CRUD | DONE | `internal/task/` package with resource.go |
| 7.5 | Implement task restoration | DONE | `internal/task/restoration/` subdomain |
| 7.6 | Implement reminder CRUD | DONE | `internal/reminder/` package with resource.go |
| 7.7 | Implement reminder snooze | DONE | `internal/reminder/snooze/` subdomain |
| 7.8 | Implement reminder dismissal | DONE | `internal/reminder/dismissal/` subdomain |
| 7.9 | Implement task summary | DONE | `internal/summary/processor.go` |
| 7.10 | Implement reminder summary | DONE | `internal/summary/processor.go` |
| 7.11 | Implement dashboard summary | DONE | `internal/summary/resource.go` |
| 7.12 | Implement include support for summary endpoints | DONE | `internal/summary/rest.go` |
| 7.13 | Write unit tests | DONE | Tests in reminder, dismissal, snooze, summary, task, restoration packages |

### Phase 8 — Frontend Productivity UI

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 8.1 | Implement dashboard page | DONE | `pages/DashboardPage.tsx` |
| 8.2 | Implement tasks list page | DONE | `pages/TasksPage.tsx` with data table, filters, create/delete |
| 8.3 | Implement task detail/edit view | DONE | Task update mutation in `lib/hooks/api/use-tasks.ts`, inline edit in table |
| 8.4 | Implement task restore UI | DONE | `useRestoreTask()` hook |
| 8.5 | Implement reminders list page | DONE | `pages/RemindersPage.tsx` with data table |
| 8.6 | Implement reminder detail/edit view | DONE | Reminder update mutation in `lib/hooks/api/use-reminders.ts` |
| 8.7 | Implement snooze UI | DONE | Snooze button in RemindersPage, `useSnoozeReminder()` hook |
| 8.8 | Implement dismiss UI | DONE | Dismiss button in RemindersPage, `useDismissReminder()` hook |
| 8.9 | Implement settings page | DONE | `pages/SettingsPage.tsx` with profile and appearance sections |
| 8.10 | Implement households management page | DONE | `pages/HouseholdsPage.tsx` with data table and create dialog |

### Phase 9 — Local Environment

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 9.1 | Create docker-compose.yml | DONE | `deploy/compose/docker-compose.yml` |
| 9.2 | Create nginx.conf | DONE | `deploy/compose/nginx.conf` |
| 9.3 | Create .env.example | DONE | `.env.example` with DB, JWT, OIDC vars |
| 9.4 | Verify end-to-end login flow locally | SKIPPED | Manual verification task — not checked in tasks.md |
| 9.5 | Verify all API endpoints work through proxy | SKIPPED | Manual verification task — not checked in tasks.md |

### Phase 10 — k3s Deployment

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 10.1 | Create k8s manifests for auth-service | DONE | `deploy/k8s/auth-service.yaml` |
| 10.2 | Create k8s manifests for account-service | DONE | `deploy/k8s/account-service.yaml` |
| 10.3 | Create k8s manifests for productivity-service | DONE | `deploy/k8s/productivity-service.yaml` |
| 10.4 | Create k8s manifests for frontend | DONE | `deploy/k8s/frontend.yaml` |
| 10.5 | Create Ingress YAML | DONE | `deploy/k8s/ingress.yaml` |
| 10.6 | Document secret management approach | DONE | Secrets referenced as env vars in manifests, not embedded |

### Phase 11 — Bruno Collections

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 11.1 | Create auth collection | DONE | `bruno/auth/` with 5 requests |
| 11.2 | Create account collection | DONE | `bruno/account/` with 6 requests |
| 11.3 | Create productivity collection | DONE | `bruno/productivity/` with 7 requests |
| 11.4 | Create environment files | DONE | `bruno/environments/Local.bru` |

### Phase 12 — Renovate + Maturity

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 12.1 | Create renovate.json | DONE | `renovate.json` with Go, npm, GitHub Actions, Docker groups |
| 12.2 | Verify all docs are accurate | SKIPPED | Manual verification task — not checked in tasks.md |
| 12.3 | Verify CI is strict (branch protection) | SKIPPED | Requires GitHub admin configuration — not checked in tasks.md |

---

**Completion Rate:** 103/107 tasks (96.3%)
**Skipped without approval:** 4 (all are manual verification tasks requiring a running environment or admin access)
**Partial implementations:** 0

## Skipped / Deferred Tasks

| Task | What's Missing | Impact |
|------|---------------|--------|
| 9.4 Verify end-to-end login flow locally | Requires running Docker Compose with valid OIDC credentials and a browser | Cannot confirm the full auth flow works end-to-end without manual testing |
| 9.5 Verify all API endpoints through proxy | Requires running Docker Compose and executing Bruno collections | Cannot confirm nginx routing is correct without integration testing |
| 12.2 Verify docs accuracy | Manual cross-referencing of docs vs code | Docs may drift from implementation over time |
| 12.3 Verify CI branch protection | Requires GitHub repository admin settings | PR merge without CI passing is possible until configured |

All 4 skipped tasks are inherently manual verification steps that cannot be automated in code. Their absence does not indicate missing functionality — the underlying artifacts (docker-compose, nginx.conf, docs, CI workflows) all exist.

---

## Developer Guidelines Compliance

### Changes Audited

The `bug-fixes` branch contains 3 commits with 23 files changed (122 insertions, 23 deletions). Changes are bug fixes, not new features.

### Passes

- **Immutable models**: No model mutations introduced; all domain models remain immutable with accessor methods.
- **Entity separation**: GORM tags only on entity.go files; models are clean.
- **Builder pattern**: Builders exist for all domains with invariant enforcement.
- **Processor/Provider layer separation**: Bug fixes in `appcontext/context.go` correctly use processors (`membership.NewProcessor`, `tenant.NewProcessor`) rather than direct DB access.
- **REST JSON:API compliance**: `oidcprovider/rest.go` correctly adds `Slug` field to `RestModel` and populates via `Transform()`.
- **Multi-tenancy context**: `database.WithoutTenantFilter(ctx)` used appropriately in `appcontext/context.go` for the bootstrapping query, with clear comment explaining why.
- **SetID empty-string guard**: All `SetID` methods across services consistently handle empty strings (JSON:API creates may not include an ID).
- **No anti-patterns in new code**: No handlers calling providers directly, no business logic in handlers, no `logrus.StandardLogger()` misuse.
- **Frontend component patterns**: `cursor-pointer` added to shared UI components (Button, Select) using Tailwind.
- **Frontend routing**: Root `/` redirect to `/app` uses React Router `<Navigate>` correctly.
- **Named exports**: All frontend components use named exports.

### Violations

- **Rule:** Frontend test must match production code
- **File:** `frontend/src/pages/__tests__/LoginPage.test.tsx:62-63`
- **Issue:** Test mock provides `{ id: "google", type: "auth-providers", attributes: { displayName: "Google" } }` but `LoginPage.tsx:29` now uses `provider.attributes.slug` (added in commit `7b53cbc`). The mock is missing the `slug` attribute, causing the login URL to render as `/api/v1/auth/login/undefined?redirect=%2Fapp`.
- **Severity:** medium
- **Fix:** Add `slug: "google"` to the test mock's attributes: `{ id: "google", type: "auth-providers", attributes: { slug: "google", displayName: "Google" } }`

---

- **Rule:** Security — Header override for tenant context should be scoped
- **File:** `shared/go/auth/auth.go:130-143`
- **Issue:** The auth middleware allows `X-Tenant-ID` and `X-Household-ID` headers to override JWT claims when claims are nil UUID. While this is necessary for onboarding (where JWT doesn't yet contain tenant/household), the override could allow an authenticated user to impersonate a different tenant if their JWT has nil values. The code has a comment explaining the purpose, and the guard (`== uuid.Nil`) limits the scope, but there is no validation that the header-provided tenant actually belongs to the authenticated user.
- **Severity:** low (only applies when JWT has nil tenant, which is a narrow window during onboarding)
- **Fix:** Consider adding a note or TODO for future hardening — validate that the user has membership in the header-supplied tenant. This is acceptable for MVP.

---

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| auth-service | PASS | PASS | 8 packages tested, all pass |
| account-service | PASS | PASS | 5 packages tested, all pass |
| productivity-service | PASS | PASS | 6 packages tested, all pass |
| frontend | PASS | FAIL (1/256) | `LoginPage.test.tsx > renders correct login URLs` — mock missing `slug` attribute |

---

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 103/107 tasks done; 4 skipped tasks are all manual verification steps
- **Guidelines Compliance:** MINOR_VIOLATIONS — 1 test mock not updated after code change, 1 low-severity security observation
- **Recommendation:** NEEDS_FIXES — Fix the failing test before merge

## Action Items

1. **Fix LoginPage test mock** — Add `slug: "google"` to the mock provider attributes in `frontend/src/pages/__tests__/LoginPage.test.tsx:63` so the test matches the updated `AuthProvider` type.
2. **Consider tenant validation hardening** (optional, non-blocking) — Add a comment or future ticket to validate that header-supplied `X-Tenant-ID` belongs to the authenticated user in `shared/go/auth/auth.go`.
3. **Complete manual verification tasks** (9.4, 9.5, 12.2, 12.3) when the local environment is available — these are operational readiness checks, not code defects.
