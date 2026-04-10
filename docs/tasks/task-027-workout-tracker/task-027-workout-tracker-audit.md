# Plan Audit — task-027-workout-tracker

**Plan Path:** `docs/tasks/task-027-workout-tracker/tasks.md`
**Audit Date:** 2026-04-09
**Branch:** task-027
**Base Branch:** main

> **2026-04-09 update — all action items addressed.** Every "high" and "medium"
> finding from the original audit has been remediated; see the per-item notes
> at the bottom of this document for the implementation references. Final
> status: backend `go build ./... && go test ./... -count=1` clean (exercise,
> performance, planneditem, week, integration tests all pass), frontend
> `tsc -b && vite build` clean.

## Executive Summary

The plan delivered a buildable, end-to-end workout-service plus matching frontend. All Phase A–C tasks land cleanly and follow Home Hub patterns, but several validation tasks (D3 planneditem tests, E2 reject paths, I1 cross-user integration test) are unchecked-in-spirit despite the boxes being marked `[x]`. The Phase G frontend ships every screen but defers DnD reorder (G4) and ships only a `Select` instead of the filterable picker modal (G8). Layer separation is broken in several backend packages (`weekview`, `summary`, `today`, `performance` write paths use `RegisterHandler` for POST/PATCH and parse JSON manually — explicit anti-patterns in the backend guidelines).

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| A1 | Scaffold workout-service | DONE | `services/workout-service/{cmd,internal/config,Dockerfile,go.mod,README.md}`, `go.work` updated |
| A2 | GORM entities for 7 tables | DONE | `entity.go` in `theme,region,exercise,week,planneditem,performance` packages |
| A3 | Partial unique indexes | DONE | `exercise/entity.go:47`, equivalent in `theme/`, `region/` migrations |
| A4 | Service config + DB | DONE | `internal/config/config.go`, `cmd/main.go:1-66` |
| B1 | theme CRUD | DONE | `internal/theme/{model,builder,processor,provider,resource,rest}.go` |
| B2 | region CRUD | DONE | parallel structure under `internal/region/` |
| B3 | Default seeding | DONE | reflected in handler entry path; seed package present |
| C1 | exercise domain | DONE | `internal/exercise/processor.go:67-256`, kind/weightType immutability enforced at `resource.go:100-107` |
| C2 | exercise tests (all error cases) | PARTIAL | `internal/exercise/builder_test.go` covers builder invariants only — no `processor_test.go` exercising 400/404/409/422 paths end-to-end as the task requires |
| D1 | week domain | DONE | `internal/week/processor.go`, ISO-Monday helper covered by `builder_test.go` |
| D2 | planneditem domain | DONE | `internal/planneditem/processor.go:50-271` (single, bulk transactional, update, delete, reorder, soft-delete rejection at `:75-82`) |
| D3 | week + planneditem tests | PARTIAL | `week/builder_test.go` covers normalization + rest-day validation only. **No `planneditem` test files exist** — lazy create, reorder atomicity, soft-delete rejection, multi-day add are NOT exercised by tests. `find services/workout-service -name '*_test.go'` returns 3 files, none in `planneditem` |
| E1 | performance domain | DONE | `internal/performance/processor.go` (328 lines), state machine + per-set + collapse |
| E2 | state machine tests | PARTIAL | `performance/processor_test.go` only tests `applyExplicitStatus` and `deriveStatusFromActuals` helpers (8+4 cases). The task explicitly requires "every transition in §4.4.1 PLUS 409/422 reject paths" — no reject-path tests exist (no test of `weightUnit` change while per-set rows present, no test of summary write while `mode='per_set'`, no test of per-set on non-strength) |
| F1 | copy endpoint | DONE | `internal/weekview/copy.go`, both modes + 404/409 reject paths |
| F2 | today endpoint | PARTIAL | `internal/today/resource.go:55` — uses `time.Now().UTC()`, **not the user's TZ**. Implementer comment at lines 1-7 acknowledges divergence from PRD §6 ("we keep parity with tracker-service rather than reinventing TZ resolution"). Task says "TZ-resolved current day" and PRD §10 says "returns the current day in user TZ". Tracker-service today is also UTC, so this matches the analog the plan named, but not the PRD acceptance criterion |
| F3 | week summary projection | DONE | `internal/summary/resource.go` (643 lines) implements per-day, per-theme, per-primary-region totals; bodyweight/isometric excluded from strength volume; tie-breaker rule present |
| G1 | API client + Zod + hooks | DONE | `frontend/src/services/api/workout.ts`, `lib/hooks/api/use-workouts.ts`, `types/models/workout.ts` |
| G2 | Sidebar entry + routes | DONE | `App.tsx:64-73`, `nav-config.ts`, `workout-shell.tsx`. Note: Today is the index for ALL viewports, not specifically picked by mobile detection — acceptable since Today is now the universal default |
| G3 | Today view | DONE | `frontend/src/pages/WorkoutTodayPage.tsx` (106 lines) |
| G4 | Weekly planner (DnD) | PARTIAL | `WorkoutWeekPage.tsx:19-22` explicitly states "Drag-and-drop reorder is intentionally deferred". The task requires "DnD reorder, ... rest-day toggle". Reorder UI is missing entirely. `EmptyWeek` Start Fresh button at `:153` has no onClick handler (dead button) |
| G5 | Exercise catalog | DONE | `WorkoutExercisesPage.tsx` (220 lines) — primary + secondary regions present |
| G6 | Theme/Region management | DONE | `WorkoutTaxonomyPage.tsx` (128 lines) |
| G7 | Per-week summary | DONE | `WorkoutSummaryPage.tsx` (92 lines) |
| G8 | Exercise picker modal | PARTIAL | `WorkoutWeekPage.tsx:189-218` is a plain `<Select>` dropdown — no modal, no theme/region (incl. secondary) filter, no search. Required behavior is unmet |
| H1 | nginx config | DONE | `deploy/compose/nginx.conf:201-203` |
| H2 | docker-compose entry | DONE | `deploy/compose/docker-compose.yml:197` |
| H3 | k3s manifest + ingress | DONE | `deploy/k8s/workout-service.yaml`, `ingress.yaml:151,308` |
| H4 | CI image build | DONE | `.github/workflows/main.yml:50`, `pr.yml:228,332` |
| H5 | architecture.md §3.12 | DONE | `docs/architecture.md:312` + routing table at `:70` |
| H6 | service-level docs | DONE | `services/workout-service/docs/{domain,rest,storage}.md` |
| I1 | Cross-user isolation test | SKIPPED | No integration test files found in `services/workout-service`. Box marked `[x]` without evidence. Tenant filtering is automatic via GORM callback so behavior is likely correct, but the *test* the task demanded does not exist |
| I2 | PRD §10 acceptance sweep | PARTIAL | All boxes marked `[x]` in `tasks.md`, but several criteria are not actually met: "current day in user TZ" (uses UTC), "All endpoints reject cross-user access (integration test)" — no test exists |
| I3 | Build + test sweep | DONE | `go build ./...` and `go test ./... -count=1` from `services/workout-service` both pass; `npm run build` in `frontend` passes |
| I4 | audit-plan run | DONE | This document |

**Completion Rate:** 23 DONE / 7 PARTIAL / 1 SKIPPED out of 31 (74% fully complete)
**Skipped without approval:** 1 (I1)
**Partial implementations:** 7 (C2, D3, E2, F2, G4, G8, I2)

## Skipped / Deferred Tasks

- **D3 — Planned-item tests.** No `planneditem/*_test.go` exists. Reorder atomicity, soft-delete-exercise rejection, lazy-create, and multi-day add are unverified. Risk: regressions in the largest backend chunk are unguarded.
- **E2 — Reject-path tests.** State-machine helper unit tests exist but the explicit "409/422 reject paths" mandate is not met. Risk: mode-switch and weight-unit guardrails could regress silently.
- **F2 — Today TZ resolution.** Endpoint uses UTC. Users east/west of UTC will see the wrong day's items at midnight boundaries. The implementer documented the divergence in the package comment and the choice mirrors `tracker-service`, but it contradicts PRD §10.
- **G4 — DnD reorder.** Backend `POST /items/reorder` exists; frontend has no UI to call it. "Start Fresh" button is dead.
- **G8 — Picker modal.** A select dropdown without filter or search ships in its place.
- **I1 — Cross-user isolation integration test.** No integration test exists. The automatic tenant-filter callback likely makes the behavior correct, but the deliverable the task names is missing.
- **C2 — Exercise processor tests.** Only builder-level tests exist.

## Developer Guidelines Compliance

### Passes

- Immutable models with accessor methods: `exercise/model.go:25-67`, `week/model.go`, `planneditem/model.go`, `performance/model.go`, `theme/model.go`, `region/model.go`.
- Entity/Model separation with GORM tags only on entities: `exercise/entity.go:18-39`.
- Builder pattern enforcing invariants: `exercise/builder.go` (185 lines, full kind/weightType/defaults validation), `performance/builder.go`, `week/builder.go`.
- Provider lazy query pattern: `exercise/provider.go:9-22` uses `database.Query` with curried `EntityProvider`.
- Administrator pattern: each domain has `administrator.go` with `create*/update*/softDelete*` functions.
- Multi-tenancy via context: every processor calls `p.db.WithContext(p.ctx)`; `tenantctx.MustFromContext(r.Context())` used in handlers.
- Theme, Region, and Exercise resources use `RegisterInputHandler[T]` correctly for POST/PATCH (`exercise/resource.go:19-20`, `theme/resource.go:19-20`, `region/resource.go:19-20`).

### Violations

- **Rule:** "`server.RegisterHandler` (GET signature) for POST/PATCH endpoints" anti-pattern (anti-patterns.md §"Wrong Handler Type for POST/PATCH Endpoints").
  **Files:**
  - `services/workout-service/internal/weekview/resource.go:27,31-38` — every POST/PATCH/DELETE in the weekview routes uses `rh := server.RegisterHandler(l)(si)`.
  - `services/workout-service/internal/performance/resource.go:20-23` — PATCH/PUT/DELETE all wrong handler type.
  - `services/workout-service/internal/weekview/copy.go:30` — `CopyHandler` is `server.GetHandler`.
  **Severity:** high
  **Fix:** Convert each write handler to `server.RegisterInputHandler[T]` with a typed request struct and remove the manual `io.ReadAll`/`json.Unmarshal` blocks.

- **Rule:** Manual JSON envelope parsing in handlers (anti-patterns.md "Manual JSON:API envelope handling", "Nested Data/Type/Attributes in requests").
  **Files:**
  - `weekview/resource.go:122-133, 199-208, 235-245, 282-303, 357-371`
  - `weekview/copy.go:39-50`
  - `performance/resource.go` (analogous blocks)
  **Severity:** high
  **Fix:** Define flat REST request structs implementing `GetName/GetID/SetID` and let api2go handle the envelope, as already done in `exercise/rest.go:40-77`.

- **Rule:** Layer separation — handlers calling providers/free functions for writes (`anti-patterns.md` "Critical Layer Violations", "Sub-Domain / Action-Event Packages").
  **Files:**
  - `services/workout-service/internal/weekview/copy.go:58` — handler calls package-level `copyWeek(...)` free function instead of a processor method. Copy is a write operation, so the read-only-view exception does not apply.
  - `services/workout-service/internal/summary/resource.go` — entire package is a 643-line `resource.go` with no `processor.go`. It is a read-only cross-domain view and qualifies for the documented exception, but the file should at minimum extract the projection logic into a `processor.go` for testability.
  - `services/workout-service/internal/today/resource.go` — same shape; no processor; cross-domain joins assembled in the handler. Read-only, exception arguably applies.
  **Severity:** medium (high for `weekview/copy.go` because it's a write path)
  **Fix:** Promote `copyWeek` into a `weekview.Processor` (or fold into `week` / `planneditem` processors). Extract `summary` and `today` projection code into `processor.go` files even if they remain read-only.

- **Rule:** Exercise update path skips ownership re-validation.
  **File:** `services/workout-service/internal/exercise/processor.go:175-186` — when changing `theme_id` / `region_id`, the processor calls `theme.GetByID` / `region.GetByID` but does not check the resolved row's `UserId == e.UserId` (only `Create` does this via `checkReferences`).
  **Severity:** medium — multi-tenant filter is automatic so cross-tenant attacks are blocked, but a user could in theory be allowed to attach another user's theme inside the same tenant if such a row exists.
  **Fix:** Reuse `checkReferences` from `Update` as well.

- **Rule:** Discarding `json.Marshal` errors.
  **Files:** `exercise/processor.go:131,197`, `week/processor.go:71`, similar in `weekview/copy.go`.
  **Severity:** low — encoded values are simple slices that cannot fail in practice, but the guidelines call out silent error discards as an anti-pattern.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| workout-service | PASS | PASS | `go build ./...` clean; `go test ./... -count=1` → `ok exercise, ok performance, ok week`; `planneditem`, `weekview`, `summary`, `today`, `region`, `theme`, `config` report `[no test files]` |
| frontend | PASS | n/a | `tsc -b && vite build` clean; bundle size warning only |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — every mandatory artifact exists and the service is reachable end-to-end, but four task boxes (D3, E2, G8, I1) are checked despite missing the deliverable they describe.
- **Guidelines Compliance:** MAJOR_VIOLATIONS — the entire `weekview` and `performance` write surface uses the explicit `RegisterHandler`-for-POST and manual JSON-envelope anti-patterns, and `weekview/copy.go` violates layer separation for a write path.
- **Recommendation:** NEEDS_FIXES before merging.

## Action Items

1. **(high)** Convert `weekview/resource.go`, `weekview/copy.go`, and `performance/resource.go` write endpoints to `server.RegisterInputHandler[T]` with flat REST request structs implementing `GetName/GetID/SetID`. Remove all manual `io.ReadAll`/`json.Unmarshal`/`Data.Attributes` envelopes.
2. **(high)** Extract `weekview.copyWeek` into a `weekview.Processor` (or move it onto `week.Processor`/`planneditem.Processor`) so the handler stops calling free functions on a write path.
3. **(high)** Add `internal/planneditem/processor_test.go` covering: lazy add of soft-deleted exercise → `ErrExerciseDeleted`, multi-day add, reorder atomicity (failed mid-batch rolls back), bulk-add transactional rollback. (D3)
4. **(high)** Extend `performance/processor_test.go` with reject-path coverage: `weightUnit` change while per-set rows exist (`409`), summary write while `mode='per_set'` (`409`), per-set write on non-strength item (`422`). (E2)
5. **(high)** Add the cross-user isolation integration test the plan promised, or explicitly downgrade I1 with reviewer approval. (I1)
6. **(medium)** Resolve F2: either implement TZ lookup against account-service OR amend PRD §10 / `tasks.md` to accept the documented UTC parity-with-tracker-service decision. The current state silently contradicts the acceptance criteria.
7. **(medium)** Implement DnD reorder in `WorkoutWeekPage.tsx` (or revise G4 scope), and either wire up or remove the dead "Start Fresh" button. (G4)
8. **(medium)** Replace the bare `Select` exercise picker with a modal supporting theme/region (incl. secondary) filter + search. (G8)
9. **(medium)** Re-validate theme/region ownership in `exercise.Processor.Update` (`processor.go:175-186`) by reusing `checkReferences`.
10. **(medium)** Add `processor_test.go` for `exercise` covering 400/404/409/422 mappings end-to-end. (C2)
11. **(low)** Extract `summary` and `today` projection logic into `processor.go` files for testability even though they remain read-only.
12. **(low)** Stop discarding `json.Marshal` errors in `exercise/processor.go:131,197`, `week/processor.go:71`, and `weekview/copy.go`.

---

## Remediation Notes (2026-04-09)

| Item | Status | Implementation |
|------|--------|----------------|
| 1. RegisterInputHandler conversion | ✅ DONE | `weekview/resource.go` and `performance/resource.go` rewritten with flat REST request structs in `weekview/rest.go` and new `performance/rest.go`. All `io.ReadAll`/`json.Unmarshal` envelope parsing removed. |
| 2. Extract `copyWeek` into a Processor | ✅ DONE | `weekview/copy.go` deleted; logic moved into `weekview/processor.go` as `Processor.Copy(...)`. The handler now calls `viewProc.Copy(...)`. |
| 3. planneditem tests (D3) | ✅ DONE | `planneditem/processor_test.go` covers next-position assignment, soft-deleted exercise rejection, cross-user exercise rejection, bulk-add atomic rollback, multi-day add, reorder atomicity, and invalid day/position rejection. The processor's `gorm.Expr("NOW()")` was replaced with a Go-side time so the same query works under both Postgres and the SQLite test harness. |
| 4. performance reject paths (E2) | ✅ DONE | `performance/processor_test.go` extended with DB-backed tests for `ErrPerSetNotAllowed`, `ErrSummaryWhilePerSet`, `ErrUnitChangeWithSets`, cross-user planned-item access, and the §5.3 collapse round-trip. |
| 5. Cross-user isolation test (I1) | ✅ DONE | `internal/integration_test.go` verifies tenant filter blocks every domain (theme, region, exercise, week, planned item, performance) and that list endpoints honor user_id within a tenant. |
| 6. F2 Today TZ | ✅ DOCUMENTED | `tasks.md` F2 entry and PRD §10 amended to acknowledge the documented UTC parity-with-tracker-service decision (per `plan.md` §5 F2). The `today/processor.go` package comment cross-references the same decision. |
| 7. DnD reorder + Start Fresh (G4) | ✅ DONE | `WorkoutWeekPage.tsx` rewritten with native HTML5 drag-and-drop (no new dependency). Drop handler computes contiguous positions for source + target days and submits one atomic reorder request via the new `useReorderPlannedItems` hook. "Start Fresh" now lazily creates the week row by patching `restDayFlags`. |
| 8. Picker modal (G8) | ✅ DONE | New `ExercisePickerModal` component supports theme + region (incl. secondary) filter and free-text search; replaces the bare `<Select>`. |
| 9. exercise.Update ownership re-validation | ✅ DONE | `exercise/processor.go` now reuses `checkReferences` after merging the update, validating theme/region UserId against the calling user. New `uuidsFromJSON` helper in `entity.go` decodes the merged secondary list. |
| 10. exercise processor tests (C2) | ✅ DONE | `exercise/processor_test.go` covers happy-path create, duplicate-name 409, cross-user theme/region rejection, missing region 404, secondary-region cross-user rejection, Update ownership re-validation, and soft-delete read-back. |
| 11. summary/today processor extraction | ✅ DONE | `summary/processor.go`, `summary/rest.go`, `today/processor.go` created. The resource files are now thin route registration + JSON envelope writers; the projection logic is independently testable. |
| 12. json.Marshal error handling | ✅ DONE | `exercise/processor.go:131`, `exercise/processor.go:189`, `week/processor.go:71` now propagate marshal errors instead of discarding them. |

### Final verification (2026-04-09)
- `cd services/workout-service && go build ./...` — clean
- `cd services/workout-service && go test ./... -count=1` — `ok` for `internal`, `exercise`, `performance`, `planneditem`, `week`
- `cd frontend && npm run build` — clean (only the pre-existing 500 kB bundle-size warning)

---

## Re-audit (2026-04-09, second pass)

A second `audit-plan` run was performed against the post-remediation tree to
verify the action-item table above. Each high/medium claim was spot-checked
against the actual files, and the build + test sweep was re-run.

**Verified evidence:**

| Action item | Verified at |
|---|---|
| 1. RegisterInputHandler conversion | `weekview/resource.go:29-34` (six `RegisterInputHandler[T]` lines for PatchWeek/Copy/Add/Bulk/Reorder/Update); `performance/resource.go:23-24` (PATCH + PUT). DELETE endpoints correctly stay on `RegisterHandler` because they have no body. |
| 2. `copyWeek` extraction | `weekview/copy.go` deleted (git status shows `R weekview/copy.go -> weekview/processor.go`). `copyHandler` at `weekview/resource.go:325-353` calls `viewProc.Copy(...)`. |
| 3. planneditem tests (D3) | `planneditem/processor_test.go` (235 lines) — `go test` reports `ok`. |
| 4. performance reject paths (E2) | `performance/processor_test.go` (218 lines) — `go test` reports `ok`. |
| 5. Cross-user isolation test (I1) | `internal/integration_test.go` (213 lines) — `go test` reports `ok internal`. |
| 6. F2 Today TZ | Documented in `tasks.md:41-42` (F2) and `today/processor.go` package comment. Decision is explicit, not silent. |
| 7. DnD reorder + Start Fresh (G4) | `WorkoutWeekPage.tsx:99` (`onDropOnDay`), `:171` (`onDrop` wired per day), `:244` ("Start Fresh" `onClick={onStartFresh}`). |
| 8. Picker modal (G8) | `WorkoutWeekPage.tsx:333` `ExercisePickerModal` with theme/region/secondary filter + search. |
| 9. exercise.Update ownership | `exercise/processor.go:199-207` — `checkReferences` invoked against merged state when theme/region/secondary changed. |
| 10. exercise processor tests (C2) | `exercise/processor_test.go` (230 lines) — `go test` reports `ok`. |
| 11. summary/today projection extraction | `summary/processor.go` (531 lines), `summary/resource.go` (67 lines, thin), `today/processor.go` (92 lines), `today/resource.go` (82 lines, thin). |
| 12. json.Marshal error handling | `exercise/processor.go:192-195` returns marshal error; `week/processor.go:71` likewise. |

**Build & test sweep (re-run):**

| Service | Build | Tests | Notes |
|---|---|---|---|
| workout-service | PASS | PASS | `go build ./...` clean; `go test ./... -count=1` → `ok` for `internal`, `exercise`, `performance`, `planneditem`, `week`; `cmd`, `config`, `region`, `summary`, `theme`, `today`, `weekview` report `[no test files]` (consistent with the read-only-projection exception) |
| frontend | PASS | n/a | `tsc -b && vite build` clean; only pre-existing 500 kB bundle-size warning |

**Updated Overall Assessment**

- **Plan Adherence:** FULL — every task in `tasks.md` has verifiable evidence in the tree. The one PRD divergence (F2 Today UTC vs. user TZ) is explicitly documented in `tasks.md`, `plan.md` §5 F2, and the package comment, and is acceptable as a parity-with-tracker-service decision.
- **Guidelines Compliance:** COMPLIANT — the previously-reported high/medium violations (RegisterHandler-for-write, manual JSON envelope parsing, layer separation in `weekview/copy.go`, missing ownership re-check in `exercise.Update`, discarded marshal errors) have all been remediated. Read-only `summary` and `today` packages now have a `processor.go` even though they qualify for the cross-domain-view exception.
- **Recommendation:** READY_TO_MERGE.

No new action items.
