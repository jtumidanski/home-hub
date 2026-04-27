# Plan Audit — task-045-dashboard-designer

**Plan Path:** `docs/tasks/task-045-dashboard-designer/plan.md`
**Audit Date:** 2026-04-27
**Branch:** `feature/task-045-dashboard-designer` (worktree at `.worktrees/task-045-dashboard-designer`)
**Base Branch:** `main`
**Commits ahead of main:** 69

## Executive Summary

All 65 tasks across phases A–O are implemented with code, tests, and (where relevant) documentation evidence. The plan checkboxes were never marked, but git history walks Phase A → Phase O cleanly, with a final cutover commit (`7ec75fc feat(frontend): cut over /app to DashboardRedirect and remove legacy DashboardPage`) followed by stabilization fixes. Every Go module under `shared/go/{dashboard,events,kafka,retention}` and both touched services (`dashboard-service`, `account-service`) build and test green. The frontend `tsc -b`, `vite build`, and `vitest run` (518 tests) all pass. Recommendation: READY_TO_MERGE.

## Task Completion

| # | Task | Status | Evidence |
|---|------|--------|----------|
| 1 | `shared/go/dashboard` widget allowlist + caps | DONE | `shared/go/dashboard/types.go:1-30`, fixture `shared/go/dashboard/fixtures/widget-types.json`, test `shared/go/dashboard/types_test.go` (passes) |
| 2 | `shared/go/events` envelope + UserDeletedEvent | DONE | `shared/go/events/events.go:1-30`, `events_test.go` (passes); commit `215f2ff` |
| 3 | `shared/go/kafka/producer` retry wrapper | DONE | `shared/go/kafka/producer/producer.go:1-30`, tests pass; commit `ca1c547` |
| 4 | `shared/go/kafka/consumer` Manager | DONE | `shared/go/kafka/consumer/consumer.go:1-30`, expanded tests (commit `51a25aa`) |
| 5 | Retention `dashboard.dashboards` category | DONE | `shared/go/retention/category.go:27,53,70`, `category_test.go:46-53` |
| 6 | dashboard-service skeleton | DONE | `services/dashboard-service/{go.mod,Dockerfile,cmd/main.go,internal/config/config.go}` |
| 7 | `appcontext` helper | NOT_APPLICABLE | Service doesn't need a re-export wrapper; tenant context flows through `shared/go/tenant` directly (plan permitted skipping) |
| 8 | CI/build smoke | DONE | `go build ./services/dashboard-service/...` succeeds in audit |
| 9 | Validator happy path + version rule | DONE | `services/dashboard-service/internal/layout/validator.go:55`, `CodePayloadTooLarge` etc. |
| 10 | Widget count cap + unknown-type | DONE | commit `3e72a40`; tests in `validator_test.go` |
| 11 | Geometry + ID rules | DONE | commit `bb8e02c` |
| 12 | Config size + depth rules | DONE | commit `a3e2128` |
| 13 | Payload-size cap edge case | DONE | commit `53b9073` |
| 14 | Entity + migration | DONE | `services/dashboard-service/internal/dashboard/entity.go`; wired in `cmd/main.go:25` (`SetMigrations(dashboard.Migration)`) |
| 15 | Immutable model + builder | DONE | `dashboard/model.go`, `builder.go`, `builder_test.go` (passes) |
| 16 | Provider (read-side) | DONE | `dashboard/provider.go` |
| 17 | Administrator (write-side) | DONE | `dashboard/administrator.go` |
| 18 | Processor + Create | DONE | `processor.go:407` `Create`, processor_test.go pass |
| 19 | List + Get | DONE | `processor.go:365,383` |
| 20 | Update + Delete | DONE | `processor.go:305,339` |
| 21 | Bulk reorder (single-scope) | DONE | `processor.go:239` `Reorder` |
| 22 | Promote | DONE | `processor.go:207` `Promote` |
| 23 | Copy-to-mine | DONE | `processor.go:57` `CopyToMine`; `regenerateWidgetIDs` at line 191 |
| 24 | Seed with advisory lock | DONE | `processor.go:111` `Seed`, `acquireSeedLock` line 171 uses `pg_advisory_xact_lock` |
| 25 | REST models | DONE | `dashboard/rest.go:10-72`; `Transform` at line 74 |
| 26 | List + Get handlers | DONE | `dashboard/resource.go:25,44,69`; wired in `cmd/main.go:53` |
| 27 | Create/Update/Delete handlers | DONE | `resource.go:95,134,178` |
| 28 | Reorder handler | DONE | `resource.go:203` |
| 29 | Promote + CopyToMine handlers | DONE | `resource.go:255,290` |
| 30 | Seed handler | DONE | `resource.go:321` |
| 31 | Retention wire + noop reaper | DONE | `services/dashboard-service/internal/retention/{wire.go,handlers.go}`; wired in `cmd/main.go:46` (commit `9782a47`) |
| 32 | UserDeletedEvent consumer handler | DONE | `services/dashboard-service/internal/events/handler.go` + `handler_test.go` (passes) |
| 33 | Wire consumer in main.go | DONE | `cmd/main.go:36-42` (`kcons.New(...).Run(ctx)`) |
| 34 | account-service household_preferences entity/model | DONE | `services/account-service/internal/householdpreference/{entity.go,model.go,builder.go,builder_test.go}` |
| 35 | household_preferences processor + REST | DONE | `householdpreference/{processor.go,resource.go,rest.go,rest_test.go}` (passes); routes wired in `account-service/cmd/main.go:70` |
| 36 | UserDeletedEvent producer + internal endpoint | DONE | `services/account-service/internal/userlifecycle/{resource.go,resource_test.go}` (passes); wired `cmd/main.go:58` |
| 37 | Account-service Kafka config | DONE | `services/account-service/internal/config/config.go:17,41` (`TopicUserLifecycle`, `KafkaBrokers`) |
| 38 | docker-compose + nginx | DONE | `deploy/compose/docker-compose.yml:152` (`hh-dashboard` service), `nginx.conf:143` upstream |
| 39 | local-up.sh + .env | DONE | `scripts/local-up.sh` is generic `docker compose up`; compose env vars in `docker-compose.yml:169-171` (commit `411517e` documented Kafka env vars). No `.env.example` in repo to update. |
| 40 | Service docs | DONE | `services/dashboard-service/docs/{domain,rest,storage}.md`; `account-service/docs/domain.md` mentions `household_preferences` (lines 220, 351); `docs/architecture.md:591-600` mentions kafka producer/consumer + `dashboard-service` consumer group |
| 41 | Frontend layout schema + types | DONE | `frontend/src/lib/dashboard/{widget-types.ts,schema.ts}` + tests in `__tests__/widget-registry.test.ts` and parity test |
| 42 | parseConfig tolerant-read | DONE | `frontend/src/lib/dashboard/parse-config.ts` (commit `1a8a9fd`) |
| 43 | Widget registry shape + first 3 widgets | DONE | `frontend/src/lib/dashboard/widget-registry.ts`; widgets in `widgets/{weather,tasks-summary,reminders-summary}.ts` |
| 44 | Remaining 6 widgets | DONE | `widgets/{overdue-summary,meal-plan-today,calendar-today,packages-summary,habits-today,workout-today}.ts` (commit `2843ade`) |
| 45 | Seed layout module | DONE | `frontend/src/lib/dashboard/seed-layout.ts` + `__tests__/seed-layout.test.ts` |
| 46 | API client + query hooks | DONE | `services/api/dashboard.ts`, `household-preferences.ts`; `lib/hooks/api/use-dashboards.ts`, `use-household-preferences.ts` |
| 47 | DashboardShell | DONE | `frontend/src/pages/DashboardShell.tsx` + test; commit `6ecdf92` |
| 48 | DashboardRenderer (CSS Grid) | DONE | `pages/DashboardRenderer.tsx`; tests `DashboardRenderer.test.tsx` |
| 49 | UnknownWidgetPlaceholder + LossyConfigBadge | DONE | `components/features/dashboard-widgets/{unknown-widget-placeholder,lossy-config-badge}.tsx` + tests |
| 50 | DashboardRedirect | DONE | `pages/DashboardRedirect.tsx` + `__tests__/DashboardRedirect.test.tsx` |
| 51 | Extract DashboardSkeleton | DONE | `frontend/src/components/common/dashboard-skeleton.tsx:3` (commit `08a9a9b`) |
| 52 | Pull-to-refresh | DONE | `pages/DashboardRenderer.tsx:8,41,86`; test `DashboardRenderer.refresh.test.tsx` |
| 53 | Wire routes in App.tsx | DONE | `frontend/src/App.tsx:74-83` plus `pages/DashboardsIndexRedirect.tsx`. Note: legacy `/app` was deferred to Task 65; here it now also redirects (post-cutover). |
| 54 | DashboardsNavGroup | DONE | `components/features/navigation/dashboards-nav-group.tsx:118` + test |
| 55 | New-dashboard modal + kebab | DONE | `components/features/dashboards/{new-dashboard-modal,dashboard-kebab-menu}.tsx` + tests |
| 56 | Drag-reorder in sidebar | DONE | `frontend/package.json:17-19` adds `@dnd-kit/{core,sortable,utilities}`; commit `251e20a` |
| 57 | Install react-grid-layout | DONE | `frontend/package.json:32,55` |
| 58 | Designer reducer | DONE | `pages/dashboard-designer/state.ts` + `__tests__/state.test.ts` |
| 59 | DesignerGrid | DONE | `pages/dashboard-designer/designer-grid.tsx:2,95` (uses `react-grid-layout/legacy`, `draggableHandle=".widget-drag-handle"`) |
| 60 | Widget edit chrome | DONE | `pages/dashboard-designer/widget-chrome.tsx:19` + test |
| 61 | Palette drawer | DONE | `pages/dashboard-designer/palette-drawer.tsx` (uses `@base-ui/react/dialog` rather than shadcn `Sheet`, but the behavior — drawer listing every `widgetRegistry` entry, drag-add path emitting fresh UUID and dispatching `add`) matches the spec. |
| 62 | Config panel + ZodForm | DONE | `pages/dashboard-designer/{config-panel,zod-form}.tsx` + tests |
| 63 | Save / Discard / dirty guard | DONE | `pages/DashboardDesigner.tsx:33,56,59,71`; `pages/dashboard-designer/use-unsaved-guard.ts` + test |
| 64 | Below-tablet blocker + tooltip | DONE | `pages/DashboardDesigner.tsx:34-54` (`useMobile()` blocker); `pages/DashboardShell.tsx:12,44-48` (disabled Edit + tooltip); `__tests__/mobile-blocker.test.tsx` |
| 65 | Cutover (delete DashboardPage) | DONE | `App.tsx:74` `<Route index element={<DashboardRedirect />} />`; `grep -r DashboardPage frontend/src` only finds an unrelated doc-comment in `seed-layout.ts:9` (referencing the original layout). Parity test at `pages/__tests__/dashboard-renderer-parity.test.tsx` (commit `0d2a501`) |

**Completion Rate:** 64/65 DONE, 1/65 NOT_APPLICABLE (Task 7 — plan explicitly permits skipping)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

- **Task 7 (`appcontext` helper):** marked NOT_APPLICABLE. The plan itself says "If no such package exists in the reference services, skip this task." `services/dashboard-service/internal/appcontext/` does not exist; the service uses `shared/go/tenant` directly (verified in `cmd/main.go` and `internal/dashboard/processor.go`). No impact.

- **Task 39 (`.env.example`):** No `.env.example` was added because none exists in `deploy/compose/` or repo root in this codebase; the plan made this conditional ("If a `.env.example` exists ... add placeholders"). Compose-level env-var defaults (`BOOTSTRAP_SERVERS`, `EVENT_TOPIC_USER_LIFECYCLE`, `KAFKA_CONSUMER_GROUP`) are present in `deploy/compose/docker-compose.yml:69,169-171`.

## Build & Test Results

| Service / Module | Build | Tests | Notes |
|------------------|-------|-------|-------|
| `shared/go/dashboard` | PASS | PASS | `go test ./...` |
| `shared/go/events` | PASS | PASS | |
| `shared/go/kafka` | PASS | PASS | producer + consumer (commit `51a25aa` expanded coverage) |
| `shared/go/retention` | PASS | PASS | new `CatDashboardDashboards` covered |
| `services/dashboard-service` | PASS | PASS | `dashboard`, `events`, `layout` packages all pass; `cmd/config/retention` have no test files |
| `services/account-service` | PASS | PASS | `householdpreference`, `userlifecycle`, etc. all green |
| `frontend` (tsc -b + vite build) | PASS | — | `npx tsc -b` clean; `npx vite build` produces dist/ |
| `frontend` (vitest run) | — | PASS | 72 files, 518 tests, all green |

## Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

Every required artifact in the plan exists in the worktree, every affected Go module compiles and tests green, every frontend test passes, and the production build succeeds. The two tasks not implemented (Task 7 appcontext, Task 39 `.env.example`) are explicitly conditional in the plan text and the conditions for skipping are met. The Phase O cutover (`7ec75fc`) deletes legacy `DashboardPage.tsx` cleanly — `grep -r DashboardPage frontend/src` returns only an unrelated doc-comment.

Minor stylistic note (not a fix item): the palette drawer uses `@base-ui/react/dialog` rather than the shadcn `Sheet` named in the plan; behavior is equivalent.

## Action Items

None blocking. Optional cleanup if desired:

1. The `seed-layout.ts:9` doc-comment still says "the legacy `DashboardPage`" — update wording to reference the renderer if the original phrasing is now misleading. Cosmetic only.
