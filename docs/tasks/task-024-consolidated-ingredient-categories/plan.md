# Plan — Task 024: Consolidated Ingredient Categories

Last Updated: 2026-04-08

## Executive Summary

The meal-planner consolidated-ingredient view (in-app preview + markdown export) is supposed to render ingredients grouped by category. Frontend grouping logic, backend producer, and JSON:API surface are all already wired end-to-end. In deployed environments, however, every ingredient falls into a single "Uncategorized" block silently — no errors, no logs.

This task is a diagnose-then-fix bug task targeting `services/recipe-service/internal/export/processor.go::ConsolidateIngredients`. Three structural defects in that function make the failure silent and the root cause hard to identify:

1. The `ListCategories` error path is hidden behind an `if err == nil` guard at `processor.go:131` — there is **no log line at all** when category fetch fails.
2. Missing canonical-ingredient rows are silently dropped.
3. Canonical→category-ID mismatches (cross-tenant, deleted category) are silently dropped.

The work has three deliverables:
1. **Diagnosis** — confirm which of the four candidate root causes (or some other) is the actual deployed-env failure mode, with evidence captured in `notes.md`.
2. **Fix** — repair the root cause without bypassing `category-service` or hardcoding categories.
3. **Hardening** — add three stable, greppable log statements so the next time this regresses it is loud, plus extract a shared grouping/sorting helper so the in-app preview and markdown export cannot drift.

The endpoint must continue to return `200 OK` on category-service blips, falling back to "Uncategorized" rather than 5xx-ing the meal planner.

## Current State Analysis

### What works
- **Frontend grouping** (`frontend/src/components/features/meals/ingredient-preview.tsx:22-70`) builds `CategoryGroup[]`, sorts by `category_sort_order`, and appends a final "Uncategorized" group. No bug here.
- **JSON:API serialization** (`services/recipe-service/internal/export/resource.go:11-67`) — `RestIngredientModel` and `TransformIngredient` already expose `category_name` and `category_sort_order` as nullable fields and only populate them when non-empty.
- **Wiring** — `categoryclient` is constructed in `services/recipe-service/cmd/main.go:43,54` and passed into `plan.InitializeRoutes`. Non-nil at runtime when `CategoryServiceURL` is configured.
- **Auth** — `category-service` derives tenancy directly from the JWT claims in the `access_token` cookie (`services/category-service/cmd/main.go:33-38`, `shared/go/auth/auth.go:114-147`). The cookie is the only header `categoryclient` needs to forward and it already does. **Tenant forwarding is not the bug.**
- **Sorting in processor.go** — `processor.go:260-280` already sorts resolved ingredients by `categorySortOrder` then alphabetically, with empty-category placed last.
- **Markdown export grouping** — Despite what the PRD/context say in §4.4 / §C, `services/recipe-service/internal/export/markdown.go:96-156` already groups the shopping list by category with an "Uncategorized" tail. **The §4.4 parity gap is therefore ALREADY closed in code**, but the grouping logic is duplicated across three places (`processor.go` sort + `markdown.go` group + `ingredient-preview.tsx` group). Extracting a shared backend helper is still in scope to prevent drift.

### What is broken
- `services/recipe-service/internal/export/processor.go:130-136`:
  ```go
  if p.catClient != nil && pd.AccessToken != "" {
      if cats, err := p.catClient.ListCategories(pd.AccessToken); err == nil {
          for _, c := range cats {
              categoryByID[c.ID] = catInfo{name: c.Name, sortOrder: c.SortOrder}
          }
      }
  }
  ```
  The `err == nil` guard discards the error completely. There is no log, no metric, no trace. If `category-service` returns 401, EOF, a malformed body, or anything non-2xx, `categoryByID` is empty, every canonical's category lookup misses, and every ingredient ends up in "Uncategorized". The deployed-environment symptom matches this exactly.
- `processor.go:233-237`: `canonicalsByID` short-rows are silently truncated. The `len(canonicalsByID) < len(canonIDs)` case does not log.
- `processor.go:246-251`: When `canonical.CategoryID() != nil` but the ID isn't in `categoryByID` (cross-tenant, deleted category, race), the ingredient just stays uncategorized with no log.

### Diagnosis suspects (ranked by likelihood — from `context.md`)
1. `ListCategories` returning an error in deployed env (network, 401, decode failure). **Most likely.** The fix here is also the fix.
2. `canonical_ingredients.category_id` is NULL in the deployed DB despite the admin UI appearing to have set it. Indicates a write-path bug elsewhere; would NOT be fixed by this task — would spawn a follow-up.
3. Cross-tenant / deleted-category ID mismatch. Less likely but cheap to catch.
4. `ingredientProc.GetByIDs` returning a partial map. Least likely.

## Proposed Future State

After this task:
- `ConsolidateIngredients` calls `ListCategories` defensively, capturing and **logging** any error at error level, then continuing with an empty `categoryByID` so the endpoint stays `200 OK`.
- Missing canonical rows produce a warn log including the missing count and plan ID.
- Resolved ingredients whose canonical references an unknown category ID produce a warn log including the canonical ID, the missing category ID, and the plan ID.
- A single shared helper, e.g. `export.GroupByCategory(ingredients []ConsolidatedIngredient) []CategoryGroup`, owns the grouping/sorting rules. Both the markdown export and the JSON:API path use it (the JSON:API path can still flatten back to a slice for transport, but the sort order is centralized).
- Unit tests cover all-categorized, all-uncategorized, mixed, unresolved-only, and `categoryclient`-error cases.
- A diagnosis paragraph is captured in `notes.md` in this task folder, identifying which of the four candidates was confirmed and the evidence.
- Frontend untouched (unless diagnosis pivots, in which case the change is documented).

## Implementation Phases

### Phase 1 — Diagnosis (must complete before Phase 2)

**Goal:** confirm the actual root cause in the deployed environment.

#### Task 1.1 — Add throwaway diagnostic logging (S)
Temporarily wrap the `ListCategories` call in `processor.go:131` to log on **both** branches at info level — count of categories returned on success, full error on failure — and deploy/run against the failing environment.
- **Acceptance:** server logs after a failing meal-plan ingredient request show either (a) a `ListCategories` error message + tenant context, or (b) a non-zero category count and a separate datapoint pointing the finger elsewhere.
- **Note:** these temporary logs will be replaced by the permanent ones in Phase 3; this exists so the implementer can decide whether they're chasing suspect 1 vs 2/3/4.

#### Task 1.2 — Cross-check the data layer (S)
Run a `SELECT id, name, category_id FROM canonical_ingredients WHERE tenant_id = '<tenant>' LIMIT 20;` against the deployed recipe-service DB to rule out suspect 2 (NULL category_id despite UI assignment).
- **Acceptance:** confirmation that `category_id` is populated (or, if not, this task halts and a follow-up is filed for the write path).
- **Dependencies:** read access to deployed recipe-service DB.

#### Task 1.3 — Capture evidence and write the diagnosis paragraph (S)
Create `docs/tasks/task-024-consolidated-ingredient-categories/notes.md` with:
- Which suspect was confirmed (1/2/3/4/other).
- Concrete evidence: log excerpt, DB row sample, request/response trace, or similar.
- One sentence on why the other suspects were ruled out.
- **Acceptance:** `notes.md` exists with all of the above. PRD §4.1 acceptance check satisfied.
- **Dependencies:** Tasks 1.1 and 1.2.

### Phase 2 — Root-cause fix

**Goal:** make the deployed-env failure mode produce a categorized response.

#### Task 2.1 — Apply the targeted fix (S–M, depending on diagnosis)
Implementation depends on what Phase 1 found. Likely candidates:
- **If suspect 1 (most likely):** restructure `processor.go:130-136` so the error is captured into a variable, logged at error level (see Task 3.1), and execution continues with an empty `categoryByID`. If the underlying error is from `categoryclient.ListCategories` itself (e.g., bad URL build, missing header, decode panic), fix that in `services/recipe-service/internal/categoryclient/client.go` as well.
- **If suspect 3:** the fix is the warn log (Task 3.3) plus, if applicable, a category-service contract correction.
- **If suspect 4:** investigate `ingredient.Processor.GetByIDs` for chunking/limit bugs.
- **If suspect 2:** stop. Raise a follow-up task; this task only does observability hardening (Phase 3) for the current code path.
- **Acceptance:** in the failing environment, `GET /api/v1/meals/plans/{planId}/ingredients` for a known-categorized plan returns ingredients with non-null `category_name` / `category_sort_order` for canonicals that have a category assigned.
- **Dependencies:** Phase 1.

### Phase 3 — Observability hardening (independent of which suspect was confirmed)

**Goal:** PRD §4.3. Make the next failure of this shape loud.

#### Task 3.1 — Log `ListCategories` errors (S)
Restructure `processor.go:130-136` so the error from `ListCategories` is captured in a named variable and logged at **error** level with a stable message and the plan ID. Suggested message: `"Failed to fetch categories for plan ingredient consolidation"`. Logger fields: `plan_id`, `error`.
- **Acceptance:**
  - Code no longer uses the `if … err == nil` swallowing pattern.
  - Unit test (Task 4.5) injects a category-client error and asserts the log line is emitted.
  - The endpoint still returns `200 OK` and every ingredient appears in "Uncategorized".

#### Task 3.2 — Log partial canonical-ingredient batch fetch (S)
After the `canonicalsByID, err := ingredientProc.GetByIDs(canonIDs)` call at `processor.go:233`, when `len(canonicalsByID) < len(canonIDs)`, log at **warn** level. Suggested message: `"Canonical ingredient batch fetch returned fewer rows than requested"`. Logger fields: `plan_id`, `requested`, `received`, `missing_count`.
- **Acceptance:** code path triggers the warn log; processing continues; missing rows degrade to "Uncategorized".

#### Task 3.3 — Log unknown category ID lookups (S)
Inside the `for canonID, acc := range resolved` loop at `processor.go:238-252`, when `canonical.CategoryID() != nil` but the ID is missing from `categoryByID`, log at **warn** level. Suggested message: `"Canonical ingredient references unknown category"`. Logger fields: `plan_id`, `canonical_ingredient_id`, `category_id`.
- **Acceptance:** code path triggers the warn log; ingredient ends up in "Uncategorized".

#### Task 3.4 — Verify log message stability for grep (S)
Ensure the three new messages are string literals (not format strings with interpolated values), so production-log greps work. The variable parts must live in structured logger fields, not in the message itself.
- **Acceptance:** code review confirms each `WithError(err).WithField(...).Error(...)` / `Warn(...)` uses a constant message string.

### Phase 4 — Shared grouping helper

**Goal:** PRD §4.4 last paragraph. Eliminate the markdown/preview drift risk.

#### Task 4.1 — Extract `GroupByCategory` helper (M)
Create a new exported helper in `services/recipe-service/internal/export/` (e.g., `grouping.go`) with a signature roughly:
```go
type CategoryGroup struct {
    Name        string // "" means uncategorized
    SortOrder   int
    Ingredients []ConsolidatedIngredient
}

func GroupByCategory(ingredients []ConsolidatedIngredient) []CategoryGroup
```
The helper:
- Groups by `CategoryName` (empty string → uncategorized bucket).
- Sorts groups by `SortOrder` ascending; uncategorized always last.
- Within each group, sorts ingredients alphabetically by `DisplayName` (falling back to `Name`).
- **Acceptance:** function exists with godoc; matches existing markdown.go and processor.go sort order; covered by unit tests (Task 4.5).

#### Task 4.2 — Refactor `markdown.go` to use the helper (S)
Replace the inline grouping/sorting block at `services/recipe-service/internal/export/markdown.go:96-156` with a call to `GroupByCategory`. Keep the markdown rendering identical (heading levels, "_(unresolved)_" suffix for unresolved, etc.).
- **Acceptance:** `markdown.go` no longer contains its own `categoryGroup` type or sort. `go test ./services/recipe-service/...` still passes. Markdown output is byte-identical for representative inputs (verify with a snapshot-style test if convenient).
- **Dependencies:** Task 4.1.

#### Task 4.3 — Refactor `processor.go` final pass to use the helper (S)
The current `sort.Slice(sortedAccums, ...)` block at `processor.go:260-280` orders the *flat* response slice by category sort order then display name. After the helper exists, the simplest path is to build the flat slice as today, then run it through `GroupByCategory` and re-flatten. This guarantees parity with `markdown.go`.
- **Acceptance:** processor returns ingredients in the same order as `markdown.go` renders them. Existing recipe-service tests still pass.
- **Dependencies:** Task 4.1.

#### Task 4.4 — Confirm frontend stays untouched (S)
Re-read `frontend/src/components/features/meals/ingredient-preview.tsx:22-70` to confirm it tolerates the now-correct ordering. No code change expected.
- **Acceptance:** documented in PR description that no frontend change was required (or, if one was, justification in `notes.md`).

#### Task 4.5 — Unit tests (M)
New test file `services/recipe-service/internal/export/processor_test.go` (or extend an existing one) covering:
- (a) all ingredients categorized — grouped in sort order, alphabetical within group.
- (b) all ingredients uncategorized — single trailing group, no panics.
- (c) mixed categorized + uncategorized — uncategorized last.
- (d) unresolved ingredients present — they appear in the uncategorized group, not dropped.
- (e) `categoryclient` returning an error — every ingredient lands in uncategorized, error log emitted, response is non-empty.
- New helper `GroupByCategory` gets table-driven tests for ordering edge cases (empty input, single group, ties on sort order).
- **Acceptance:** `go test ./services/recipe-service/...` passes; the five scenarios are visibly named in the test output.
- **Dependencies:** Tasks 3.1, 4.1.

### Phase 5 — Verification & cleanup

#### Task 5.1 — Remove temporary diagnostic logging (S)
Remove any throwaway logs added during Phase 1 that aren't part of the permanent §4.3 set.
- **Acceptance:** only the three §4.3 log statements remain.

#### Task 5.2 — Build and test all affected services (S)
Per CLAUDE.md: `go build ./...` and `go test ./...` for recipe-service. If `categoryclient` was touched, also build `category-service`. Verify Docker builds if any shared library was touched (none anticipated here).
- **Acceptance:** all green.

#### Task 5.3 — Manual deployed-env verification (S)
Re-run the failing meal-plan ingredients request in the deployed environment. Confirm categorized output. Confirm error/warn paths log if intentionally broken (e.g., point `CategoryServiceURL` at a bad host briefly).
- **Acceptance:** PRD §10 acceptance criteria items 2, 3, 4, 5 are demonstrated end-to-end.

#### Task 5.4 — PR with diagnosis paragraph in commit message (S)
Per PRD §4.1, the diagnosis must be in the commit message and/or `notes.md`. Include both: `notes.md` for archival, commit message body for git history.
- **Acceptance:** PR opened, commit message includes the diagnosis paragraph.

## Risk Assessment & Mitigation

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | Phase 1 reveals suspect 2 (NULL category_id in DB), meaning the visible bug is upstream of `ConsolidateIngredients` | Medium | Stops this task short of its primary goal | Phase 3 hardening still ships independently. File a follow-up task immediately. |
| R2 | The three new log statements are too noisy in normal operation (e.g., uncategorized ingredients are common and the warn at 3.3 fires constantly) | Medium | Log spam, alert fatigue | The §4.3 logs only fire on **error** conditions (fetch failure, missing rows, dangling FK) — none of which should be common. If 3.3 fires for every uncategorized ingredient, the condition is wrong; gate it on `canonical.CategoryID() != nil` (already specified). |
| R3 | Refactoring `processor.go` final-pass sort to use `GroupByCategory` changes ordering for some edge case and breaks an existing test | Low | Failing tests, possible UI glitch | Run the existing recipe-service test suite; snapshot the markdown output before/after for a representative plan. |
| R4 | The category-service error in deployed env is intermittent, making it hard to confirm the fix actually fixed it (vs. the env recovering) | Medium | False sense of resolution | Verify by deliberately pointing `CategoryServiceURL` at a bad host in a non-prod env and confirming the new error log AND graceful 200-with-uncategorized fallback. |
| R5 | Frontend tolerates the new ordering but renders empty groups or duplicate "Uncategorized" headers if the backend response shape shifts subtly | Low | UI regression | Read `ingredient-preview.tsx:22-70` after Task 4.3 and assert it consumes only `category_name`/`category_sort_order` (which are unchanged). |
| R6 | The PRD's §4.4 markdown parity work is already done in code, so the implementer might think there's nothing to do here and skip Phase 4 | Medium | Drift risk persists | Phase 4 explicitly calls out that the helper extraction is the goal, not the grouping itself. Plan/tasks documents are explicit. |

## Success Metrics

- **Functional:** every ingredient with a categorized canonical record now renders under its category heading in both the in-app preview and the markdown export, in deployed env.
- **Observability:** all three §4.3 log lines exist as constant strings, are unit-tested, and can be grepped from production logs.
- **Resilience:** killing `category-service` (or pointing `CategoryServiceURL` at a bad host) does NOT 5xx the meal-plan ingredients endpoint — it returns 200 with everything in "Uncategorized" and emits the §4.3 error log.
- **No drift:** preview and markdown export use the same `GroupByCategory` helper. Diff is verifiable in code review.
- **Test coverage:** the five scenarios in Task 4.5 are present and passing.
- **No regressions:** `go test ./services/recipe-service/...` is green.

## Required Resources & Dependencies

- Read access to a deployed recipe-service DB (Phase 1).
- Ability to read deployed service logs (Phase 1, Phase 5.3).
- Optional: ability to deploy a recipe-service build with throwaway diagnostic logging to the failing environment, OR a local repro path mimicking the failure (e.g., bad `CategoryServiceURL`).
- No new dependencies. No schema changes. No new endpoints.

## Timeline / Effort

Per CLAUDE.md, no calendar estimates. Effort sizing only:

| Phase | Effort | Notes |
|-------|--------|-------|
| Phase 1 — Diagnosis | S–M | Bottleneck is access to deployed env, not code. |
| Phase 2 — Fix | S–M | Likely a few-line restructure if suspect 1 holds. |
| Phase 3 — Observability | S | Three log statements, mechanical. |
| Phase 4 — Shared helper + tests | M | Extract, two refactors, five-scenario test suite. |
| Phase 5 — Verify & ship | S | Build, test, deploy-check, PR. |
| **Total** | **M** | Single PR if diagnosis confirms suspect 1; possibly split if suspect 2 forces a follow-up. |

## Out of Scope (per PRD §2 non-goals)

- UI redesign of consolidated ingredients beyond rendering category groups (already correct).
- Changes to canonical-ingredient → category assignment (task-016 territory).
- Changes to unit-family / quantity merging.
- New categorization for unresolved free-text ingredients.
- Backfilling NULL category_ids in the DB.
