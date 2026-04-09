# Workout Tracker — Implementation Plan

Last Updated: 2026-04-09
Status: Draft, awaiting approval

---

## 1. Executive Summary

Build a new `workout-service` Go microservice that lets a single user plan, log, and review weekly strength + cardio workouts. Add a corresponding "Workout" UI section to the frontend (mobile-first Today view + desktop weekly planner) and wire the new service into nginx, k3s, docker-compose, CI, and `docs/architecture.md`.

The PRD (`prd.md`), data model (`data-model.md`), and API contracts (`api-contracts.md`) are already approved and authoritative — this plan translates them into ordered, sized tasks. No scope additions.

## 2. Current State Analysis

- **Services in repo**: `account, auth, calendar, category, package, productivity, recipe, shopping, tracker, weather`. No workout service exists.
- **Closest analog**: `tracker-service` — also user-scoped, has a `today` endpoint, uses `entry/schedule/trackingitem/today/month` domain packages following the standard `model/entity/builder/processor/provider/resource/rest.go` layout. `workout-service` should mirror this pattern.
- **Shared infra**: `shared/` provides `auth`, `database`, `server`, `tenant` modules; `category-service` and `tracker-service` already establish patterns for partial unique indexes (`WHERE deleted_at IS NULL`) via `db.Exec` after AutoMigrate, and for default seeding on first request.
- **Deploy**: nginx config at `deploy/compose/nginx.conf`; per-service k8s manifests in `deploy/k8s/`; central ingress at `deploy/k8s/ingress.yaml`. CI builds + publishes per-service images to `ghcr.io/<owner>/home-hub-<service>`.
- **Frontend**: Vite + React + TS + shadcn/ui + TanStack Query + react-hook-form + Zod, in `frontend/src/`. New section will live alongside the existing Tracker entry in the sidebar.
- **Date/time**: user time zone preference is read from account-service (existing pattern in `tracker-service` today endpoint). Workout service should reuse the same lookup pattern.

## 3. Proposed Future State

- New `services/workout-service/` directory containing a Go module with the standard layout, owning the `workout.*` Postgres schema.
- Domains under `internal/`: `theme`, `region`, `exercise`, `week`, `planneditem`, `performance`, `summary`, `today`, `seed`, `config`.
- Routes mounted under `/api/v1/workouts/...` per `api-contracts.md`.
- Frontend Workout section: sidebar entry, Today page (mobile-default), Weekly planner page, Exercise catalog page, Theme/Region management page, Per-week summary page.
- nginx + ingress + docker-compose + k8s manifests + CI pipeline updated to deploy `workout-service` and route `/api/v1/workouts` to it.
- `docs/architecture.md` §3.12 entry; `services/workout-service/docs/{domain,rest,storage}.md` written.

## 4. Implementation Phases

The phases are ordered so that backend foundation lands before any feature endpoints, all backend lands before frontend, and infrastructure plumbing happens in parallel where it doesn't block.

### Phase A — Service Skeleton & Schema
Stand up an empty but buildable `workout-service` with config, server bootstrap, health endpoints, GORM migrations for all tables, and the partial unique indexes.

### Phase B — Taxonomy Domains (Themes, Regions)
CRUD + soft delete + default seeding for themes and regions. Establishes the seeding pattern used by every later domain.

### Phase C — Exercise Catalog
CRUD with kind/weightType immutability, secondary regions, defaults shape per kind, soft delete, filtering.

### Phase D — Weeks & Planned Items
Lazy week resolution + ISO-Monday normalization, single + bulk add, edit, delete, reorder, week GET with embedded items + performances, restDayFlags PATCH.

### Phase E — Performances & Logging
Summary logging, status state machine (incl. server derivation), per-set mode switch (PUT/DELETE), unit guardrails.

### Phase F — Copy-From-Previous & Today & Summary
Copy endpoint (planned + actual modes); Today endpoint with TZ resolution; WeekSummary projection with per-theme/per-primary-region totals and unit selection.

### Phase G — Frontend
Sidebar entry; Today view; Weekly planner; Exercise catalog; Theme/Region management; Per-week summary; Empty-week prompt; Exercise picker.

### Phase H — Infrastructure & Documentation
nginx local config; k3s ingress route; docker-compose service entry; k8s manifest; CI image build; `docs/architecture.md` §3.12; service-level `docs/{domain,rest,storage}.md`.

### Phase I — Validation & Acceptance
End-to-end tests against the PRD acceptance criteria, cross-user isolation tests, build/test sweep across all affected services, audit-plan run.

## 5. Detailed Tasks

Effort scale: S ≤ 0.5 day, M ~1 day, L 2–3 days, XL >3 days. Tasks within a phase are listed in execution order. Dependencies are noted where they cross phases.

### Phase A — Service Skeleton & Schema

**A1. Scaffold `services/workout-service/`** — M
- Copy structure from `tracker-service`: `cmd/main.go`, `internal/config`, `Dockerfile`, `go.mod`, `README.md`.
- Wire shared `auth`, `database`, `server`, `tenant` modules.
- Add to `go.work`.
- Acceptance: `go build ./...` from the service dir succeeds; `cmd/main.go` starts an HTTP server with `/health` and `/ready`.

**A2. Define GORM entities for all 7 tables** — L
- One `entity.go` per domain package (`theme`, `region`, `exercise`, `week`, `planneditem`, `performance`, `performance_sets` lives under `performance`).
- Match column types/nullability from `data-model.md` exactly.
- `secondary_region_ids` as `datatypes.JSON` or `pq.StringArray`-style — verify approach used elsewhere in the repo for jsonb arrays.
- Acceptance: `AutoMigrate` runs against a fresh Postgres and creates all 7 tables in `workout` schema.

**A3. Custom SQL for partial unique indexes** — S
- Run `db.Exec` for the three `UNIQUE ... WHERE deleted_at IS NULL` indexes on `themes`, `regions`, `exercises` after AutoMigrate.
- Pattern: copy from `category-service` / `tracker-service`.
- Acceptance: indexes exist after a clean boot; duplicate-name insert against an active row returns `23505`.

**A4. Service config + DB connection** — S
- `internal/config` reads `WORKOUT_DB_*` env vars; default Postgres dsn for compose.
- Acceptance: service boots in compose pointing at the existing Postgres instance.

### Phase B — Themes & Regions

**B1. `theme` domain package** — M (depends on A2)
- Files: `model.go` (immutable Model), `entity.go` (already from A2), `builder.go`, `processor.go`, `provider.go`, `resource.go`, `rest.go`.
- Endpoints: `GET/POST /workouts/themes`, `PATCH/DELETE /workouts/themes/{id}`.
- Soft delete; uniqueness per `(tenant, user, name)` enforced by partial index → `409` mapping.
- Acceptance: full CRUD via curl; soft-deleted themes excluded from list.

**B2. `region` domain package** — S (mirrors B1)
- Same shape as themes.
- Acceptance: parity with B1 endpoints.

**B3. Default seeding on first request** — M
- New `internal/seed` package; called from a middleware or per-request guard hooked into theme/region/exercise/week handlers.
- Idempotent: only seeds when `themes` and `regions` are both empty for the `(tenant, user)`.
- Seeds: themes `Muscle, Cardio`; regions `Chest, Shoulders, Back, Biceps, Triceps, Core, Legs, Glutes, Full Body, Other`, sort_order 0..n.
- Acceptance: a brand-new user sees the default lists on their first `GET /themes` call.

### Phase C — Exercises

**C1. `exercise` domain package** — L (depends on B)
- All endpoints in api-contracts §3.
- Validations: `kind` ∈ enum and immutable; `weight_type` ∈ enum and immutable; `defaults` shape matches `kind`; primary `regionId` not in `secondaryRegionIds`; theme/all referenced regions exist and belong to the same user.
- `regionId` filter must match primary OR secondary region (use `WHERE region_id = ? OR secondary_region_ids @> ?::jsonb`).
- Soft delete.
- Acceptance: all error cases in §3.7 (`400`, `404`, `409`, `422`) reproducible; happy-path CRUD round-trips.

**C2. Exercise unit + processor tests** — M
- Cover: kind/weightType immutability, secondary region validation, default shape per kind, filter behavior including secondary-region match.
- Acceptance: `go test ./internal/exercise/...` green.

### Phase D — Weeks & Planned Items

**D1. `week` domain package** — L (depends on C)
- ISO-Monday normalization helper (`weekStart` URL → Monday DATE in user TZ); single source of truth used by every week endpoint.
- Lazy creation: `GET /weeks/{weekStart}` returns 404 if no row; `PATCH /weeks/{weekStart}` (rest_day_flags) creates lazily.
- `GET` returns the week with embedded `items` joined to exercises (including soft-deleted) and embedded performances + per-set rows. Render `exerciseName` + `exerciseDeleted` from the join.
- Acceptance: round-trips week resource; 404 on empty; PATCH lazy create; date normalization works for arbitrary day-of-week input.

**D2. `planneditem` domain package** — L
- Endpoints: single add, bulk add (transactional), update, delete, reorder.
- Position auto-assign when omitted (`max+1` for the day).
- Reject adding soft-deleted exercise → `422`.
- Reorder is single-transaction over the supplied list.
- Acceptance: all §4.3–§4.6 endpoints; bulk add atomicity verified by injecting a failing item.

**D3. Week + planned item tests** — M
- Cover normalization, lazy create, reorder atomicity, soft-delete-exercise rejection, multi-day add.
- Acceptance: tests green.

### Phase E — Performances & Logging

**E1. `performance` domain package** — L (depends on D)
- 1:1 with planned item; row created on first write or status change away from `pending`.
- `PATCH .../performance` — summary actuals, status, weightUnit, notes; status state machine per PRD §4.4.1; auto-derive `partial`/`done` when only actuals are sent; reject summary writes while `mode='per_set'` with `409`; reject `weightUnit` change while per-set rows exist with `409`.
- `PUT .../performance/sets` — replace per-set rows, switch mode to `per_set`, clear summary actuals, reject for non-strength items with `422`. Server assigns `setNumber` from array order.
- `DELETE .../performance/sets` — collapse using `count` / `max reps` / `max weight`, switch back to `summary`.
- Acceptance: all transitions in §4.4.1 round-trip; mode switching round-trips both ways with the documented collapse semantics.

**E2. Performance state machine tests** — M
- Table-driven test enumerating every transition in §4.4.1 plus the reject paths (`409`, `422`).
- Acceptance: all transitions exercised.

### Phase F — Copy / Today / Summary

**F1. Copy-from-previous endpoint** — M (depends on D)
- `POST /weeks/{weekStart}/copy` with `mode: planned|actual`.
- Find most recent prior week with at least one planned item for the user.
- Planned mode: copy `planned_items` rows verbatim, new ids, new `week_id`.
- Actual mode: copy structure but populate `planned_*` from the source week's actuals when present (collapse per-set → max-weight × max-reps × count); fall back to source planned values when no actuals.
- Reject when target week already has planned items → `409`. Reject when no source → `404`.
- Acceptance: both modes round-trip; conflict cases produce expected codes.

**F2. Today endpoint** — M
- `GET /workouts/today`.
- Resolve today's date in the user's TZ via account-service preferences (mirror `tracker-service/today`).
- Compute the parent week (Monday), find or 404-implicit-empty the items for that day.
- Returns empty `items: []` when no plan, not 404.
- Acceptance: returns the correct day's items across TZ boundaries; `isRestDay` flag set from parent week.

**F3. Week summary projection** — L
- `GET /weeks/{weekStart}/summary` builds the §4.6 / §7 shape.
- Strength volume = `Σ sets × reps × weight` per group, after unit conversion. `bodyweight` and `isometric` items contribute to `itemCount` only, never `strengthVolume`.
- Per-region totals only count primary `region_id`.
- Unit selection: most-used unit in the week, ties → `lb` for weight, `mi` for distance.
- `actualSummary` collapses per-set on the fly.
- Acceptance: a fixture week with mixed strength/cardio/isometric/bodyweight produces the expected totals.

### Phase G — Frontend

**G1. API client + Zod schemas** — M
- New `frontend/src/services/workout.ts` with typed clients for all endpoints.
- Zod schemas mirroring `api-contracts.md` shapes (incl. discriminated union per `kind`).
- TanStack Query hooks per endpoint.
- Acceptance: typecheck green; query keys consistent with the rest of the app.

**G2. Sidebar entry + routing** — S
- Add Workout entry under Tracker in the sidebar.
- Routes: `/workouts/today` (default on mobile), `/workouts/week/:weekStart`, `/workouts/exercises`, `/workouts/taxonomy`, `/workouts/week/:weekStart/summary`.
- Mobile detection chooses Today as the landing page.
- Acceptance: navigation works; deep links parse `weekStart`.

**G3. Today view (mobile-first)** — L (depends on G1)
- Renders the day's items with status, large tap targets, sticky bottom action bar.
- Per-item logging UI in summary mode; "Track per set" affordance for strength.
- Mark item done/skipped; "Mark day complete" bulk action.
- One-handed usable; no drag-and-drop.
- Acceptance: usable on a 375×667 viewport; all status transitions reachable.

**G4. Weekly planner view (desktop-first)** — L
- 7 day columns; planned items sortable by drag-and-drop; back/forward week navigation.
- Empty-week prompt: Copy Planned / Copy Actual / Start Fresh.
- Rest day toggle per day.
- Acceptance: full CRUD on planned items; both copy modes work; rest day toggle persists.

**G5. Exercise catalog screen** — M
- List + filter by theme/region; create/edit modal supporting kind, weightType, primary + secondary regions, defaults shape per kind.
- Soft delete with confirm.
- Acceptance: a created exercise appears in the picker and the catalog list.

**G6. Theme/Region management screen** — S
- CRUD + reorder; soft delete.
- Acceptance: changes immediately reflected in catalog and picker.

**G7. Per-week summary screen** — M
- Renders `WeekSummary` shape: per-day breakdown, per-theme totals, per-region totals.
- Acceptance: matches a manually computed fixture.

**G8. Exercise picker modal** — S
- Used inside the weekly planner; client-side filter by theme/region (incl. secondary) and search.
- Acceptance: filter narrowing works across theme + region simultaneously.

### Phase H — Infrastructure & Documentation

**H1. nginx local config** — S
- Add `/api/v1/workouts -> workout-service:PORT` to `deploy/compose/nginx.conf`.
- Acceptance: `curl localhost/api/v1/workouts/themes` reaches the service in compose.

**H2. docker-compose service entry + Postgres provisioning** — S
- Add `workout-service` to `deploy/compose/docker-compose.yml` with healthcheck, depends-on Postgres, env wiring.
- Schema is auto-created by AutoMigrate; reuse existing Postgres instance.
- Acceptance: `up.sh` brings the service up healthy.

**H3. k3s manifest + ingress route** — S
- New `deploy/k8s/workout-service.yaml`.
- Add `/api/v1/workouts` rule to `deploy/k8s/ingress.yaml`.
- Acceptance: manifests pass `kubectl apply --dry-run=client`.

**H4. CI: image build + publish** — S
- Add the service to whatever pipeline already builds the others (likely a matrix in `.github/workflows/...`).
- Image: `ghcr.io/<owner>/home-hub-workout`.
- Acceptance: CI green on a feature branch.

**H5. `docs/architecture.md` §3.12** — S
- New entry covering responsibilities, schema, key rules (no household scope, lazy weeks, ISO-Monday, soft-delete read-through, primary-region-only volume math).
- Update routing table.
- Acceptance: doc renders, links resolve.

**H6. Service-level docs** — M
- Write `services/workout-service/docs/{domain.md, rest.md, storage.md}` matching the depth of `tracker-service/docs/`.
- Acceptance: parity in shape with the existing services' docs.

### Phase I — Validation & Acceptance

**I1. Cross-user isolation integration test** — M
- Spin two users; verify no theme/region/exercise/week/item/performance from user A is reachable as user B.
- Acceptance: every endpoint covered.

**I2. PRD acceptance checklist sweep** — M
- Walk every box in PRD §10; mark each pass/fail.
- Acceptance: all boxes checked.

**I3. Build + test sweep across all affected services** — S
- Per `CLAUDE.md`: run builds and tests for ALL affected services; expect fix-and-rebuild cycles for any shared changes.
- Verify Docker builds for any shared lib changes (none expected, but verify).
- Acceptance: all green.

**I4. audit-plan skill run** — S
- Run `audit-plan` to verify nothing was skipped or deferred without approval.
- Acceptance: clean audit report.

## 6. Risk Assessment & Mitigation

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | Per-set ↔ summary mode switching introduces data loss bugs | M | M | Lock the collapse semantics in unit tests early (E2); always require explicit `DELETE .../sets` before summary writes |
| R2 | Time-zone resolution drifts from `tracker-service` Today behavior | M | M | Reuse the exact same TZ-resolution helper, don't reimplement |
| R3 | ISO-Monday normalization differs between client and server, causing duplicate weeks | M | H | Server is the only authority — client never normalizes; one helper, one test fixture covering Sun/Mon/mid-week |
| R4 | Partial unique indexes silently not created → duplicate names possible | L | H | Explicit assertion in startup check or smoke test that the indexes exist |
| R5 | `secondary_region_ids` jsonb query approach not portable | L | M | Confirm Postgres-only deployment; use `@>` operator and a GIN index if list filter perf becomes an issue |
| R6 | Frontend mobile UX fails one-handed test | M | M | Manually validate on a 375×667 viewport before merging G3 |
| R7 | Copy-from-actual semantics surprise the user when collapsing per-set rows | L | L | Document the `max × count` rule in the empty-week prompt UI |
| R8 | CI image matrix not auto-picking up new service | M | M | Confirm existing matrix includes the service entry; do not assume |
| R9 | Soft-deleted FK joins return null in some ORMs | L | M | Use explicit `Unscoped()` (GORM) or raw SQL on the read path |
| R10 | Volume-math unit selection ambiguity for ties | L | L | Tie-breaker fixed in code (`lb` / `mi`) and tested |

## 7. Success Metrics

- All PRD §10 acceptance boxes checked.
- `workout-service` builds + tests pass; image published; service reachable through nginx (compose) and ingress (k3s).
- p95 server-side latency for `GET /weeks/{weekStart}` and `GET /workouts/today` ≤ 100ms on a typical week (≤ 30 items).
- Zero cross-user data leaks in I1.
- audit-plan returns clean.

## 8. Required Resources & Dependencies

- **External services**: existing Postgres instance (compose + k3s); existing account-service for user TZ.
- **Shared libs**: `shared/{auth,database,server,tenant}` — no changes anticipated.
- **Frontend deps**: existing shadcn/ui, TanStack Query, react-hook-form, Zod — no new packages anticipated.
- **CI**: existing image-build pipeline matrix.
- **Skills**: `backend-dev-guidelines`, `frontend-dev-guidelines`, `audit-plan`, `service-doc`.

## 9. Timeline (effort, not calendar)

Total effort estimate: ~21 task-units (1 unit ≈ 1 dev-day on a single-person stream).

| Phase | Effort | Notes |
|-------|--------|-------|
| A — Skeleton & Schema | 2 | Mostly mechanical |
| B — Themes/Regions + Seeding | 1.5 | |
| C — Exercises | 2 | Validation-heavy |
| D — Weeks & Planned Items | 3 | Largest backend chunk |
| E — Performances | 2 | State-machine-heavy |
| F — Copy / Today / Summary | 3 | Summary projection is the long pole |
| G — Frontend | 5 | Two large views (Today, Weekly) dominate |
| H — Infra & Docs | 1.5 | |
| I — Validation | 1 | |

Per `CLAUDE.md`, no calendar dates are committed.
