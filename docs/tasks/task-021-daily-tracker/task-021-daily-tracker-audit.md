# Plan Audit — task-021-daily-tracker

**Plan Path:** docs/tasks/task-021-daily-tracker/tasks.md
**Audit Date:** 2026-04-08
**Branch:** task-021
**Base Branch:** main

## Executive Summary

Daily tracker service is functionally complete: 99/101 task items checked (98%), backend builds clean, and all unit/processor tests pass. Two tasks remain unchecked and explicitly noted in the plan: a unit test for the today endpoint (4.5) and a measured perf check (10.11). However, the implementation has **significant guideline violations** in the cross-domain handlers — `today/`, `entry/`, and `month/` resources call providers directly, contain business logic, and emit JSON:API envelopes by hand. The `today` package has no `processor.go` at all. These are repeated, named anti-patterns from the project's backend guidelines and need rework before merge.

## Task Completion

| #     | Task                                                                 | Status   | Evidence / Notes |
|-------|----------------------------------------------------------------------|----------|------------------|
| 1.1–1.21 | Service scaffold, trackingitem domain (model/builder/entity/provider/administrator/processor/rest/resource), wiring, unit tests | DONE | services/tracker-service/internal/trackingitem/* + cmd/main.go:47 |
| 2.1–2.10 | Entry domain end-to-end + unit tests                              | DONE     | internal/entry/*; processor_test.go covers value validation, skip, date constraints |
| 3.1–3.12 | Month summary + report processor, schedule-aware expected days, scale-type stats, tests | DONE | internal/month/processor.go + processor_test.go |
| 4.1   | Today logic added to processor                                       | PARTIAL  | No `today/processor.go` exists. Logic lives entirely inside `today/resource.go:27-165` (handler) — see violation below |
| 4.2   | Today REST model + transform                                        | PARTIAL  | No `rest.go`; anonymous structs declared inline in handler (today/resource.go:41-60) |
| 4.3   | Today resource handler                                              | DONE     | today/resource.go:27 |
| 4.4   | Today route wired                                                   | DONE     | cmd/main.go:44 |
| 4.5   | Unit test: correct items for day-of-week, entries paired           | SKIPPED  | No test files in internal/today/; explicitly unchecked |
| 5.1–5.9 | Dockerfile, compose, nginx, k8s, ingress, CI workflow            | DONE     | Dockerfile, deploy/compose/docker-compose.yml +17, deploy/compose/nginx.conf +10, deploy/k8s/tracker-service.yaml, deploy/k8s/ingress.yaml +7, .github/workflows/{main,pr}.yml |
| 6.1–6.8 | Frontend tracker management UI + zod schemas                      | DONE     | frontend/src/components/features/tracker/*, frontend/src/lib/schemas/tracker.schema.ts |
| 7.1–7.12 | Calendar grid, month nav, cell editor, optimistic mutations, complete-month auto-switch, empty state | DONE | frontend/src/components/features/tracker/calendar-grid.tsx + use-trackers.ts |
| 8.1–8.8 | Today quick-entry view, inline editors, progress, navigation     | DONE     | frontend/src/components/features/tracker/today-view.tsx |
| 9.1–9.8 | Monthly dashboard report                                          | DONE     | frontend/src/components/features/tracker/month-report.tsx |
| 10.1–10.4 | Integration tests for tracking item / entry / month summary / report | DONE (in spirit) | Implemented as processor-level tests against in-memory SQLite (trackingitem/processor_test.go, entry/processor_test.go, month/processor_test.go). Exercise full domain through GORM but not via HTTP. Acceptable as integration coverage for processors. |
| 10.5–10.9 | Service docs (domain.md, rest.md, storage.md, README.md) and architecture.md update | DONE | services/tracker-service/docs/*, docs/architecture.md (+31/-?) |
| 10.10 | All 20 PRD acceptance criteria walked                              | DONE     | Per commit message 3c46f2c |
| 10.11 | Month summary endpoint < 200ms with 20 tracking items              | SKIPPED  | Not measured; explicitly noted as deferred to pre-launch perf pass |

**Completion Rate:** 99/101 tasks checked (98%)
**Skipped (unchecked):** 2 (4.5, 10.11)
**Partial implementations:** 2 (4.1, 4.2 — done in handler instead of processor/REST layer)

## Skipped / Deferred Tasks

- **4.5 — Today unit test:** No tests exist in `internal/today/`. Day-of-week filtering and entry pairing are unverified. Risk: regression in scheduled-item filtering would not be caught. Compounded by 4.1/4.2 being done in the handler — this code is not even reachable from a processor unit test today.
- **10.11 — Perf check:** Plan acknowledges deferral to a pre-launch perf pass; acceptable to leave as a follow-up but should be tracked.

## Developer Guidelines Compliance

### Passes

- Immutable models with private fields and accessor methods (trackingitem/model.go:10-34, entry/model.go, schedule/model.go).
- Builder pattern enforces invariants (trackingitem/builder.go, entry/builder.go) with table-driven unit tests (builder_test.go).
- Entity files contain GORM tags only; Make/ToEntity round-trip via Builder (trackingitem/entity.go:47-61).
- Providers use `database.Query` / `database.SliceQuery` lazy curried form (trackingitem/provider.go:9-37).
- `trackingitem/resource.go` is a clean reference: handlers call processor only, use `RegisterInputHandler[T]` for POST/PATCH, `d.Logger()` for trace context, and `server.Marshal*Response` for envelope.
- `entry/processor.go` and `month/processor.go` correctly thread tenant context and contain pure business logic.
- Multi-tenancy via `tenantctx.MustFromContext` in handlers; `database.Connect` already wires tenant callbacks (cmd/main.go:26).
- Dropping the `tracker.` schema prefix from `TableName()` matches project convention (schema is set on the connection's search_path) — verified in config.go:23 and commit 3c46f2c.

### Violations

- **Rule:** "Handlers calling provider functions directly" / "Cross-domain business logic in handlers"
  **File:** `services/tracker-service/internal/today/resource.go:34, 71, 109`
  **Issue:** Handler calls `trackingitem.GetAllByUser(...)`, `schedule.GetEffectiveSchedule(...)`, and `entry.GetByItemAndDate(...)` directly. The whole "items scheduled for today + paired entries" computation lives in the HTTP handler. There is no `today/processor.go`.
  **Severity:** high
  **Fix:** Create `internal/today/processor.go` (or add a `Today(userID, date)` method to `trackingitem` or a new orchestration package). Move the day-of-week filter and entry pairing into it. Handler should only call the processor and marshal the result.

- **Rule:** "Manual JSON:API envelope handling" / "Nested Data/Type/Attributes in requests"
  **File:** `today/resource.go:129-162`, `month/resource.go:105-118`
  **Issue:** Both handlers build a hand-rolled `struct { Data struct { Type, Attributes, Relationships ... } }` and call `json.Marshal` + `w.Write` directly, bypassing api2go and `server.Marshal*Response`. The transform layer should produce a typed REST model implementing `GetName/GetID/SetID` and let `server.MarshalResponse` emit the envelope.
  **Severity:** high
  **Fix:** Define a real `RestModel` for `tracker-today` and `tracker-month-summary`/`tracker-reports` with JSON:API interface methods, then route through `server.MarshalResponse[…]`.

- **Rule:** "Handlers calling provider functions directly" / "Direct entity creation in handlers"
  **File:** `services/tracker-service/internal/entry/resource.go:32-52, 61, 120, 132, 190`
  **Issue:** `isScheduledForDate` is a private helper inside the resource file that calls `schedule.GetEffectiveSchedule(...)` directly from the handler path. `createOrUpdateHandler` and `skipHandler` also call `trackingitem.GetByID(...)` (provider) directly to validate item existence and read `ScaleType`/`ScaleConfig`. This is cross-domain orchestration in a handler.
  **Severity:** high
  **Fix:** Push "is the entry scheduled for that date?" and "load tracking item context to validate value" into `entry/processor.go` (it can take a `trackingitem.Provider`/`schedule.Provider` or fetch via curried providers internally). Handlers should not import sibling domain providers.

- **Rule:** "Handlers calling provider functions directly"
  **File:** `services/tracker-service/internal/month/resource.go:54`
  **Issue:** `monthSummaryHandler` calls `schedule.GetByTrackingItemIDs(...)` directly to build a schedule snapshot map after the processor returns. This snapshot lookup belongs inside `month.Processor.ComputeMonthSummary` (which already loads schedule snapshots elsewhere).
  **Severity:** medium
  **Fix:** Have `ComputeMonthSummary` return the snapshots-by-item map (or fold the transform into the processor) so the handler does not query providers.

- **Rule:** "Discarding Transform errors with `_`" / write errors ignored
  **File:** `today/resource.go:159`, `month/resource.go:78, 114, 117`
  **Issue:** `b, _ := json.Marshal(result)` and `w.Write(b)` (ignored). Marshalling errors are swallowed; if a future field breaks JSON encoding, the response will silently 200 with truncated bytes.
  **Severity:** low
  **Fix:** Either route through `server.Marshal*Response` (which logs) or check the marshal error and log via `d.Logger()`.

- **Rule:** "Sub-Domain / Action-Event Packages — Must have a processor.go (or use the parent domain's processor)"
  **File:** `services/tracker-service/internal/today/`
  **Issue:** Package contains only `resource.go`; no model/processor/rest. The guideline explicitly says: "If the sub-domain is simple enough that a standalone processor adds no value, fold the action into the parent domain's processor as a method instead of creating a separate package with layer violations." Today does neither.
  **Severity:** high
  **Fix:** Add `today/processor.go` and `today/rest.go`, or move the today endpoint into `trackingitem` as a processor method.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| tracker-service | PASS | PASS | `go build ./...` clean; `go test ./... -count=1` → ok for entry, month, trackingitem (3 packages with tests). Today and schedule have no tests. |
| (all other workspace modules) | PASS | not run | `go build ./...` clean for every module in `go.work`; this branch only adds the tracker service so other services were not exercised by tests. |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 98% of tasks checked. Only 4.5 and 10.11 unchecked, both flagged in the file. However, 4.1/4.2 are checked but the underlying intent ("add to processor", "create REST model") was not met.
- **Guidelines Compliance:** MAJOR_VIOLATIONS — `trackingitem` is exemplary, but `today`, `entry`, and `month` resources break layer separation, manual envelope, and sub-domain rules in ways the anti-patterns doc names explicitly.
- **Recommendation:** NEEDS_FIXES — backend rework on the three cross-domain handlers before merge. Frontend and infra are merge-ready.

## Action Items

1. Create `services/tracker-service/internal/today/processor.go` containing `Today(userID, date)` that returns scheduled items + paired entries. Add `today/rest.go` with a proper `RestModel` (`GetName/GetID/SetID`). Reduce `today/resource.go` to a thin handler that calls the processor and `server.MarshalResponse`.
2. Add a `today/processor_test.go` covering: items scheduled for the requested day-of-week are returned, items not scheduled are excluded, existing entries for today are paired correctly, items with empty schedule (everyday) are included. Closes task 4.5.
3. Move `isScheduledForDate` and `trackingitem.GetByID` calls out of `entry/resource.go` into `entry/processor.go`. Handlers should not import `trackingitem` or `schedule` providers.
4. Move the `schedule.GetByTrackingItemIDs` lookup in `month/resource.go:54` into `month.Processor.ComputeMonthSummary` so the handler does not query providers. Have the processor return the data the transform needs.
5. Replace hand-rolled JSON envelopes in `today/resource.go:129-162` and `month/resource.go:105-118` with `server.MarshalResponse[…]` over typed REST models.
6. Stop discarding marshal/write errors — log via `d.Logger()` or use the `server.Marshal*` helpers.
7. Schedule the perf measurement for task 10.11 (month summary < 200ms with 20 items) before launch and check it off.
