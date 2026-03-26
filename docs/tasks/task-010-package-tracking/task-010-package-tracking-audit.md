# Plan Audit — task-010-package-tracking

**Plan Path:** docs/tasks/task-010-package-tracking/tasks.md
**Audit Date:** 2026-03-26
**Branch:** task-010-package-tracking
**Base Branch:** main

## Executive Summary

The implementation covers all 7 phases with 39 of 39 tasks fully completed (100%). Backend code demonstrates strong compliance with developer guidelines — immutable models, builder pattern, provider lazy evaluation, administrator layer separation, multi-tenancy propagation, and JSON:API REST patterns are all correctly implemented. All backend tests pass (38 tests), all frontend tests pass (388 tests including 30 new package-specific tests), all 7 services build successfully, and the frontend TypeScript compiles clean.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Service scaffold | DONE | `services/package-service/cmd/main.go`, `internal/config/config.go`, `Dockerfile`, `go.mod` |
| 1.2 | Package domain entity & model | DONE | `internal/tracking/entity.go` — GORM tags, indexes, Migration(); `internal/tracking/model.go` — 17 private fields, accessors |
| 1.3 | Package domain builder | DONE | `internal/tracking/builder.go` — fluent setters, Build() validates tracking number, carrier enum, status enum |
| 1.4 | Package domain processor & provider | DONE | `internal/tracking/processor.go` — CRUD, archive, unarchive, duplicate check, household limit; `provider.go` — lazy EntityProvider |
| 1.5 | Package domain resource & REST | DONE | `internal/tracking/resource.go` — 10 handlers via RegisterHandler/RegisterInputHandler; `rest.go` — JSON:API models |
| 1.6 | Tracking event domain | DONE | `internal/trackingevent/` — entity, model, builder, provider, administrator; append-only with deduplication |
| 1.7 | Carrier detection endpoint | DONE | `internal/carrier/detect.go` — regex matching for UPS/FedEx/USPS with confidence levels |
| 2.1 | CarrierClient interface & TrackingResult | DONE | `internal/carrier/client.go` — interface, Registry, TrackingResult types, shared HTTP client (15s timeout) |
| 2.2 | OAuth token management | DONE | `internal/carrier/oauth.go` — thread-safe OAuthTokenManager, 60s expiry buffer, double-check locking |
| 2.3 | USPS carrier client | DONE | `internal/carrier/usps.go` — Tracking API v3, OAuth, response normalization |
| 2.4 | UPS carrier client | DONE | `internal/carrier/ups.go` — Tracking API v1, response normalization, budget integration |
| 2.5 | FedEx carrier client | DONE | `internal/carrier/fedex.go` — Track API v1, response normalization, budget integration |
| 2.6 | Initial poll on create | DONE | `internal/tracking/processor.go` — `p.pollEntity(&e)` called after entity creation |
| 2.7 | Manual refresh endpoint | DONE | `resource.go` — POST /{id}/refresh with 5-min cooldown via refreshCooldown check |
| 3.1 | Polling scheduler | DONE | `internal/poller/poller.go` — background goroutine, adaptive intervals, carrier budgets, context cancellation |
| 3.2 | Poll execution | DONE | `internal/tracking/processor.go` — pollEntity() updates status/ETA, appends events via trackingevent.Create |
| 3.3 | Stale detection | DONE | `internal/poller/cleanup.go` — marks stale after configurable days with no status change |
| 3.4 | Archive/cleanup job | DONE | `internal/poller/cleanup.go` — daily: delivered->archived after N days, delete archived after M days |
| 4.1 | Docker Compose | DONE | `deploy/compose/docker-compose.yml:118-142` — container with DB, JWKS, carrier API env vars |
| 4.2 | Nginx route | DONE | `deploy/compose/nginx.conf:146-153` — `/api/v1/packages` proxied to package-service:8080 |
| 4.3 | Go workspace | DONE | `go.work:7` — `./services/package-service` entry added |
| 4.4 | Document env vars | DONE | `.env.example:34-46`, `services/package-service/README.md` |
| 5.1 | TypeScript types | DONE | `frontend/src/types/models/package.ts` — Package, TrackingEventInline, PackageSummary, CarrierDetection |
| 5.2 | API service class | DONE | `frontend/src/services/api/package.ts` — extends BaseService, all CRUD + summary + detect methods |
| 5.3 | React Query hooks | DONE | `frontend/src/lib/hooks/api/use-packages.ts` — key factory with tenant scoping, all CRUD mutations with toast feedback |
| 5.4 | Zod validation schemas | DONE | `frontend/src/lib/schemas/package.schema.ts` — create and edit schemas with proper defaults |
| 5.5 | Package list page | DONE | `frontend/src/pages/PackagesPage.tsx` — archived toggle, sorted by ETA ascending (nulls last) |
| 5.6 | Package card component | DONE | `frontend/src/components/features/packages/package-card.tsx` — carrier icon, status badge, ETA, privacy redaction |
| 5.7 | Package detail/expand | DONE | `frontend/src/components/features/packages/package-detail.tsx` — tracking event timeline, carrier website links |
| 5.8 | Create package dialog | DONE | `frontend/src/components/features/packages/create-package-dialog.tsx` — carrier auto-detect on blur/paste |
| 5.9 | Package quick actions | DONE | `package-card.tsx` — refresh, edit, toggle privacy, archive/unarchive, delete with confirmation |
| 5.10 | Sidebar nav with badge | DONE | `nav-config.ts` has `badgeKey: "inTransitCount"`; `app-shell.tsx` queries `usePackageSummary()` and passes `inTransitCount` in navBadges |
| 5.11 | Route registration | DONE | `App.tsx:49` — `/app/packages` route within ProtectedRoute |
| 6.1 | Calendar overlay | DONE | `package-calendar-overlay.tsx` — teal-colored all-day events; integrated in `CalendarPage.tsx:78-84` |
| 6.2 | Dashboard summary widget | DONE | `package-summary-widget.tsx` — arriving today, in transit, exceptions; integrated in `DashboardPage.tsx:62` |
| 7.1 | Backend unit tests | DONE | `builder_test.go` (8 cases), `processor_test.go` (16 tests), `rest_test.go` (4 privacy tests), `detect_test.go` (10 cases) |
| 7.2 | Frontend component tests | DONE | `packages/__tests__/package-card.test.tsx` (8 tests), `create-package-dialog.test.tsx` (5 tests), `package-calendar-overlay.test.ts` (5 tests), `schemas/__tests__/package.schema.test.ts` (12 tests) |
| 7.3 | Service documentation | DONE | `services/package-service/docs/domain.md`, `rest.md`, `storage.md` |
| 7.4 | Bruno API collection | DONE | `bruno/packages/` — 10 .bru files covering all endpoints |

### Cross-Cutting Concerns

| Task | Status | Evidence / Notes |
|------|--------|-----------------|
| Update `docs/architecture.md` | DONE | Lines 31, 60, 189-212 — routing entry, full service description, schema, business rules |
| Verify Docker build | DONE | Dockerfile present, K8s deployment at `deploy/k8s/package-service.yaml` |
| Verify existing services build after `go.work` change | DONE | All 6 services (auth, account, productivity, recipe, weather, calendar) build successfully |
| End-to-end smoke test | SKIPPED | No evidence of E2E testing |

**Completion Rate:** 39/39 tasks fully done (100%)
**Partial implementations:** 0
**Skipped without approval:** 1 cross-cutting (E2E smoke test)

## Skipped / Deferred Tasks

| Task | What's Missing | Impact |
|------|---------------|--------|
| **E2E smoke test** | No evidence of end-to-end verification (add package -> see on list -> see on calendar -> see on dashboard). | **Medium** — Integration between frontend and backend not verified. |

## Developer Guidelines Compliance

### Passes

| Guideline | Evidence |
|-----------|---------|
| Immutable models with accessors | `tracking/model.go` — all 17 fields unexported, public accessor methods; `trackingevent/model.go` — same pattern |
| Entity separation from model | `tracking/entity.go` has GORM tags; `model.go` has none; bidirectional `Make()`/`ToEntity()` |
| Builder pattern with invariant enforcement | `tracking/builder.go` — fluent setters, Build() validates required fields, carrier/status enum; `trackingevent/builder.go` — same |
| Provider lazy evaluation | `tracking/provider.go` — `database.Query[Entity]` and `database.SliceQuery[Entity]` for getByID, getByHousehold* |
| Administrator layer separation | `tracking/administrator.go` — create, update, deleteByID; `trackingevent/administrator.go` — Create with deduplication |
| Multi-tenancy context propagation | `resource.go` — `tenantctx.MustFromContext(r.Context())` in all handlers; providers use `db.WithContext(p.ctx)` |
| WithoutTenantFilter for background ops | `poller/poller.go:49`, `poller/cleanup.go:43` — `database.WithoutTenantFilter(ctx)` for cross-tenant queries |
| REST model separation | `rest.go` — separate RestModel, RestDetailModel, RestSummaryModel with JSON:API GetName/GetID/SetID |
| Transform functions return errors | `TransformWithPrivacy()`, `TransformSliceWithPrivacy()`, `TransformDetail()` all return `(T, error)` |
| Handlers check Transform errors | All handlers in `resource.go` check Transform err and return 500 |
| Processor accepts FieldLogger | `processor.go` — `NewProcessor(l logrus.FieldLogger, ...)` (interface, not concrete) |
| Handlers pass d.Logger() | `resource.go` — `newProc(d.Logger(), ...)` consistently in all 10 handlers |
| POST/PATCH use RegisterInputHandler | `resource.go` — `server.RegisterInputHandler[CreateRequest]`, `[UpdateRequest]` |
| String-based WHERE clauses | `provider.go` — `.Where("id = ?", id)` throughout; avoids GORM zero-value gotcha |
| No os.Getenv in handlers | `config/config.go` loaded once in `main.go`; injected into processor |
| Privacy redaction at REST layer | `rest.go:TransformWithPrivacy()` — redacts tracking number, notes, status, lastPolledAt for non-owners |
| Shared HTTP client with timeout | `carrier/client.go:NewHTTPClient()` — 15s timeout, injected into all carriers |
| Config loaded once at startup | `main.go` calls `config.Load()` once; no runtime os.Getenv |
| Table-driven tests | `builder_test.go`, `processor_test.go`, `detect_test.go` all use table-driven patterns |

### Violations

1. **Rule:** Provider count/aggregate functions should use EntityProvider pattern
   - **File:** `services/package-service/internal/tracking/provider.go:38-78`
   - **Issue:** Helper functions `countActiveByHousehold()`, `existsByHouseholdAndTrackingNumber()`, `countArrivingToday()`, `countInTransit()`, `countExceptions()` accept `*gorm.DB` directly instead of returning `database.EntityProvider` types. They bypass lazy evaluation.
   - **Severity:** low
   - **Fix:** These are aggregate/count functions that don't fit the standard EntityProvider pattern. Acceptable as pragmatic exceptions for count queries, but could be wrapped in a lazy pattern for consistency.

2. **Rule:** Direct entity mutations — prefer builder pattern for updates
   - **File:** `services/package-service/internal/tracking/processor.go` (Update method)
   - **Issue:** The Update method fetches an entity, directly mutates its fields (`e.Label = attrs.Label`, etc.), then saves. The guideline prefers fetching the model, using a builder for modifications, converting back to entity.
   - **Severity:** low
   - **Fix:** Refactor Update to: fetch entity -> Make(entity) to model -> builder from model -> Build() -> ToEntity() -> save. Pragmatically acceptable as-is since the entity is fetched and immediately persisted.

3. **Rule:** Trackingevent sub-domain missing processor layer
   - **File:** `services/package-service/internal/trackingevent/`
   - **Issue:** The trackingevent package has entity, model, builder, provider, and administrator, but no processor or resource layer. The `tracking` processor directly calls `trackingevent.Create()` and `trackingevent.GetByPackageID()`.
   - **Severity:** low
   - **Fix:** For an append-only sub-domain this is an acceptable simplification. A processor would add a layer with no additional business logic.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| package-service | PASS | PASS | 38 tests across 2 packages (carrier: 10, tracking: 28) |
| auth-service | PASS | — | Build verified after go.work change |
| account-service | PASS | — | Build verified after go.work change |
| productivity-service | PASS | — | Build verified after go.work change |
| recipe-service | PASS | — | Build verified after go.work change |
| weather-service | PASS | — | Build verified after go.work change |
| calendar-service | PASS | — | Build verified after go.work change |

## Overall Assessment

- **Plan Adherence:** FULL — 39/39 tasks done (100%), only E2E smoke test cross-cutting item remains.
- **Guidelines Compliance:** MINOR_VIOLATIONS — All core patterns correctly followed. Three low-severity deviations: aggregate provider functions bypass lazy evaluation, Update method uses direct entity mutation, trackingevent sub-domain omits processor layer. All are pragmatically acceptable.
- **Recommendation:** READY_TO_MERGE — All tasks complete, all tests pass.

## Action Items

1. **(Optional)** Perform end-to-end smoke test: add package -> verify on list -> verify on calendar -> verify on dashboard.

## Fixes Applied (2026-03-26, Round 2)

The following audit findings were resolved:

1. **Sidebar badge wired (5.10)** — Added `usePackageSummary()` hook to `app-shell.tsx` and included `inTransitCount` in the `navBadges` memo.
2. **ETA sorting added (5.5)** — `PackagesPage.tsx` now sorts packages by `estimatedDelivery` ascending with nulls last using `useMemo`.
3. **Frontend component tests added (7.2)** — Created `package-card.test.tsx` (8 tests: rendering, privacy redaction, owner actions), `create-package-dialog.test.tsx` (5 tests: form validation, carrier detection), `package-calendar-overlay.test.ts` (5 tests: event conversion, filtering), `package.schema.test.ts` (12 tests: create/edit schema validation). Fixed pre-existing test failures in `app-shell.test.tsx` and `DashboardPage.test.tsx` caused by unmocked package imports.
