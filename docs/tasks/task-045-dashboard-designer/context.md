# Task 045 — Dashboard Designer — Executor Context

> Quick reference for agents executing `plan.md`. Keep close while working.

## What we're building

A new `dashboard-service` microservice plus frontend designer/renderer that
turns the hand-coded `DashboardPage` into a user-designable, registry-driven,
JSON-backed dashboard system. Multiple dashboards per household (+ private
per-user), sidebar-integrated, drag/resize/configure grid designer.

Read these first, in order:

1. `docs/tasks/task-045-dashboard-designer/prd.md` — what + why + acceptance
2. `docs/tasks/task-045-dashboard-designer/design.md` — how (architecture)
3. `docs/tasks/task-045-dashboard-designer/api-contracts.md` — endpoint shapes
4. `docs/tasks/task-045-dashboard-designer/data-model.md` — schema + registry + seed
5. `docs/tasks/task-045-dashboard-designer/ux-flow.md` — screens + interactions

`CLAUDE.md` (project rules) and `docs/superpowers-integration.md` are also
required reading before starting.

## Key files to reference (existing patterns)

These are the cleanest reference implementations for each pattern you'll repeat:

| Pattern | Reference file |
|---|---|
| Service layout | `services/package-service/` (full service, including retention) |
| Entity + migration | `services/package-service/internal/tracking/entity.go` |
| Immutable model + builder | `services/package-service/internal/tracking/model.go` + `builder.go` |
| Provider / Administrator / Processor | `services/package-service/internal/tracking/{provider,administrator,processor}.go` |
| REST / JSON:API resource | `services/package-service/internal/tracking/{rest,resource}.go` |
| Retention wiring | `services/package-service/internal/retention/{wire,handlers}.go` |
| cmd/main bootstrap | `services/package-service/cmd/main.go` |
| Dockerfile template | `services/package-service/Dockerfile` |
| Typed preference (account-service) | `services/account-service/internal/preference/*` |
| Retention shared lib | `shared/go/retention/` |
| Atlas kafka library (inspiration only — copy style, not module) | `/home/tumidanski/source/atlas-ms/atlas/libs/atlas-kafka/` |
| Frontend page + widgets | `frontend/src/pages/DashboardPage.tsx` |
| Sidebar navigation | `frontend/src/components/features/navigation/{nav-config.ts,nav-group.tsx,app-shell.tsx}` |
| Tanstack Query hooks pattern | `frontend/src/lib/hooks/api/use-packages.ts` |
| Service API client pattern | `frontend/src/services/api/package.ts` + `base.ts` |
| Feature widgets | `frontend/src/components/features/{weather,packages,meals,calendar,trackers,workouts}/*.tsx` |

## Architectural decisions (from design.md §1)

| # | Choice | Why |
|---|---|---|
| 1 | `default_dashboard_id` lives in a **new** `household_preferences` table in `account-service` (scoped by tenant + user + household) | Current `preference` table lacks household dimension |
| 2 | Account-delete cascade via Kafka (`shared/go/kafka`, `UserDeletedEvent` on `EVENT_TOPIC_USER_LIFECYCLE`) | Loose coupling; reusable for future cascades |
| 3 | Widget-type allowlist mirrored between Go (`shared/go/dashboard`) and TS (`frontend/src/lib/dashboard/widget-types.ts`) with parity test | Set is tiny (9 items); cross-stack change already required per widget |
| 4 | Grid library: `react-grid-layout`, code-split into designer only; renderer uses plain CSS Grid | Renderer stays lean |
| 5 | Routes: `/dashboards/:id` (renderer) and `/dashboards/:id/edit` (designer) share a `DashboardShell` parent | Natural code splitting, back-button works |
| 6 | Seed race safety via Postgres `pg_advisory_xact_lock(hash(tenant, household))` | No schema pollution |
| 7 | Config schema evolution: tolerant-read / strict-write (`safeParse` with `defaultConfig` fallback + `lossy: true` badge) | Survives drift without migration infra |
| 8 | Dirty-state guard: `useBlocker` + `beforeunload`-only-while-dirty; `<DashboardRedirect>` for legacy `/app/dashboard` | Matches current non-data-router setup |

## Non-obvious gotchas

- **Kafka broker is external.** Treat like Postgres. No docker-compose container. `BOOTSTRAP_SERVERS` env var drives connection. Auto-create topics is enabled at the broker.
- **CLAUDE.md rule:** "When refactoring shared types, prefer straightforward moves over re-exporting type aliases." Apply when extracting the three inline `<Card>` blocks from `DashboardPage.tsx` into widget components.
- **Backend never validates `config` internals.** Only size/depth. Frontend Zod owns config shape. Do not add Go structs for individual widget configs.
- **Backend never validates widget overlap.** The frontend grid engine prevents it; the renderer tolerates it if something slips through.
- **`scope` is derived, not stored.** `user_id IS NULL` → `"household"`, else `"user"`. The column does not exist.
- **Widget-instance IDs are regenerated** during `copy-to-mine` (backend, not frontend — see design §2.7).
- **No existing account-deletion flow** in `account-service` as of 2026-04-23. The plan adds a thin internal endpoint (`POST /api/v1/internal/users/{id}/deleted`) that account-service's future delete flow will call to do local cleanup + emit `UserDeletedEvent`. Don't attempt to design a full delete flow.
- **Retention category key:** per shared/go/retention naming convention, use `dashboard.dashboards` (not just `dashboards`). Category constant: `CatDashboardDashboards`. Default window = 0 (never auto-purge) — the plumbing exists but the reaper is a no-op.
- **Frontend file layout override:** artifacts go under `docs/tasks/task-045-dashboard-designer/` per CLAUDE.md, NOT under `docs/superpowers/plans/`.
- **Legacy `/app` route:** currently renders `DashboardPage`. After cutover, `/app` index becomes `<DashboardRedirect>`.
- **`scripts/local-up.sh`** builds everything via compose. Add new env vars there too.
- **Frontend stack:** React 19, Vite 8, TanStack Query 5, react-hook-form 7 + zod 4, Tailwind 4, shadcn-style `components/ui/*`.
- **React-grid-layout is not yet a dependency** — add it in the Designer phase.
- **CI** docker-build matrix is dynamic (recent aggregator work) — new service joins automatically when its `Dockerfile` exists.

## Environment variables to add

`dashboard-service`:
```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, PORT=8080
JWKS_URL=http://auth-service:8080/api/v1/auth/.well-known/jwks.json
BOOTSTRAP_SERVERS
EVENT_TOPIC_USER_LIFECYCLE=home-hub.user.lifecycle
ACCOUNT_SERVICE_URL=http://account-service:8080
INTERNAL_SERVICE_TOKEN
RETENTION_INTERVAL=6h
```

`account-service` (new):
```
BOOTSTRAP_SERVERS
EVENT_TOPIC_USER_LIFECYCLE=home-hub.user.lifecycle
```

## Test philosophy

- Every task uses TDD: write the failing test first, confirm it fails for the right reason, implement, confirm it passes, commit.
- Backend: unit + integration. `rest_test.go` uses the existing service-test harness (Postgres + real router).
- Frontend: Vitest + RTL. Reducer tests are pure. Integration tests use a mocked grid callback.
- Playwright / E2E is out of scope for v1.

## Rollout (from design §4.4) — follow this sequence

1. Ship backend + infra wired but with no frontend caller (silent landing).
2. Ship frontend registry + renderer + routes + seeding, **with `/app/dashboard` still routed to the old `DashboardPage`**.
3. Manual staging parity check.
4. Flip `/app` index to `<DashboardRedirect>`; delete `DashboardPage.tsx`.
5. Final staging verification; merge.

The plan preserves this order — do NOT short-circuit and delete `DashboardPage` before the parity snapshot test passes.
