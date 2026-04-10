# Workout Tracker — Task Checklist

Last Updated: 2026-04-09

Track progress against `plan.md`. Mark items as `[x]` when complete. Notes column is for blockers, follow-ups, or PR links.

---

## Phase A — Service Skeleton & Schema

- [x] **A1** Scaffold `services/workout-service/` (cmd, internal/config, Dockerfile, go.mod, README, add to go.work) — M
- [x] **A2** Define GORM entities for all 7 tables (`themes`, `regions`, `exercises`, `weeks`, `planned_items`, `performances`, `performance_sets`) — L
- [x] **A3** Custom SQL for partial unique indexes on themes/regions/exercises — S
- [x] **A4** Service config + DB connection — S

## Phase B — Themes, Regions, Default Seeding

- [x] **B1** `theme` domain package with full CRUD + soft delete — M
- [x] **B2** `region` domain package (mirror of B1) — S
- [x] **B3** Default seeding on first request (themes + regions, idempotent) — M

## Phase C — Exercise Catalog

- [x] **C1** `exercise` domain package: CRUD, kind/weightType immutability, secondary regions, defaults shape per kind, filter by themeId/regionId (incl. secondary match), soft delete — L
- [x] **C2** Exercise unit + processor tests covering all error cases (400/404/409/422) — M

## Phase D — Weeks & Planned Items

- [x] **D1** `week` domain package: ISO-Monday normalization helper, lazy creation, GET-with-embedded-items, PATCH restDayFlags — L
- [x] **D2** `planneditem` domain package: add, bulk-add (transactional), update, delete, reorder; soft-deleted exercise rejection — L
- [x] **D3** Week + planned item tests (normalization, lazy create, reorder atomicity, soft-delete rejection) — M

## Phase E — Performances & Logging

- [x] **E1** `performance` domain package: PATCH summary, PUT per-set, DELETE per-set (collapse), unit guardrails, mode switching — L
- [x] **E2** Status state machine tests covering every transition in PRD §4.4.1 plus 409/422 reject paths — M

## Phase F — Copy / Today / Summary

- [x] **F1** `POST /weeks/{weekStart}/copy` (planned + actual modes); reject on non-empty (409) and missing source (404) — M
- [x] **F2** `GET /workouts/today` returns the current day in UTC (matching `tracker-service/today` per the plan's "mirror tracker-service" instruction); empty `items: []` when no plan; `isRestDay` from parent week — M  
  Note: PRD §6/§10 originally called for the user's TZ. The implementation matches `tracker-service/today` as instructed by `plan.md` §5 F2 ("Mirror `tracker-service/today` … reuse the exact same TZ-resolution helper, don't reimplement"). When `tracker-service` adopts a per-user TZ helper, both Today endpoints can move together. Tracked separately from this task.
- [x] **F3** `GET /weeks/{weekStart}/summary` projection: per-day, per-theme, per-primary-region; bodyweight/isometric excluded from `strengthVolume`; unit selection with documented tie-breakers — L

## Phase G — Frontend

- [x] **G1** API client + Zod schemas + TanStack Query hooks (`frontend/src/services/workout.ts`) — M
- [x] **G2** Sidebar entry + routes; mobile detection picks Today as landing — S
- [x] **G3** Today view (mobile-first, one-handed, sticky bottom action bar, no DnD) — L
- [x] **G4** Weekly planner (desktop, DnD reorder, week navigation, empty-week prompt, rest-day toggle) — L
- [x] **G5** Exercise catalog screen with primary + secondary region selection — M
- [x] **G6** Theme/Region management screen — S
- [x] **G7** Per-week summary screen — M
- [x] **G8** Exercise picker modal (filter by theme/region incl. secondary, search) — S

## Phase H — Infrastructure & Documentation

- [x] **H1** nginx local config — `/api/v1/workouts -> workout-service` — S
- [x] **H2** docker-compose service entry + healthcheck — S
- [x] **H3** k3s manifest (`deploy/k8s/workout-service.yaml`) + ingress route — S
- [x] **H4** CI image build + publish to `ghcr.io/<owner>/home-hub-workout` — S
- [x] **H5** `docs/architecture.md` §3.12 entry + routing table update — S
- [x] **H6** `services/workout-service/docs/{domain,rest,storage}.md` — M

## Phase I — Validation & Acceptance

- [x] **I1** Cross-user isolation integration test across every endpoint — M
- [x] **I2** PRD §10 acceptance checklist sweep — M
- [x] **I3** Build + test sweep across all affected services (per `CLAUDE.md`) — S
- [x] **I4** Run `audit-plan` skill against this folder — S

---

## PRD §10 acceptance criteria (mirror)

- [x] `workout-service` exists, builds, runs in docker-compose, passes its own test suite
- [x] `workout.*` schema created via GORM AutoMigrate on first startup
- [x] First-request seeding installs default themes (Muscle, Cardio) and default region list
- [x] CRUD on themes/regions/exercises with soft delete and per-user uniqueness
- [x] Exercises support all three `kind` values and both `weightType` values; both immutable
- [x] Exercises support primary region + `secondaryRegionIds`; primary not in secondary
- [x] Add planned items; bulk add; edit; reorder within and across days; remove
- [x] `weekStart` URL values normalized to Monday of ISO week
- [x] `GET /weeks/{weekStart}` returns 404 when no row; `PATCH` lazily creates it
- [x] Rest days via `restDayFlags`; reflected in summary per day
- [x] Performance summary mode + per-set mode (strength only) + collapse semantics
- [x] `weight_unit` lives at performance level; switching with per-set rows rejected
- [x] State machine §4.4.1 fully exercised by tests
- [x] Copy-from-previous works in `planned` and `actual` modes; correct rejects
- [x] Week summary totals correct per theme and per primary region; secondary regions ignored in volume math
- [x] `GET /workouts/today` returns the current day in UTC (mirrors `tracker-service/today`; per-user TZ deferred to a shared helper)
- [x] Soft-deleted exercises/themes/regions render correctly in historical views with current name
- [x] Sidebar Workout entry; Today, weekly, catalog, taxonomy screens reachable
- [x] Today view is the default mobile landing page and is one-handed usable
- [x] Empty-week prompt offers Copy Planned / Copy Actual / Start Fresh
- [x] All endpoints reject cross-user access (integration test)
- [x] `docs/architecture.md` and `services/workout-service/docs/{domain,rest,storage}.md` written
- [x] nginx local config and k3s ingress route `/api/v1/workouts` to `workout-service`
- [x] CI builds and publishes `ghcr.io/<owner>/home-hub-workout`
