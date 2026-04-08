# Tasks — Task 024: Consolidated Ingredient Categories

Last Updated: 2026-04-08

Tracking checklist. See `plan.md` for full descriptions and `context.md` for code pointers.

## Phase 1 — Diagnosis (BLOCKS Phase 2)

- [~] **1.1** Add throwaway info-level logging around `processor.go:131` `ListCategories` call (success count + error). Deploy/run against failing environment. _(S)_ — **Skipped:** no deployed-env access; the Phase 3 permanent error log replaces this. See `notes.md`.
- [~] **1.2** Run `SELECT id, name, category_id FROM canonical_ingredients WHERE tenant_id='<tenant>' LIMIT 20` against deployed recipe-service DB to rule out NULL `category_id`. _(S)_ — **Skipped:** no deployed-DB access. Suspect 2 ruled out by static reasoning (admin UI shows categories correctly). See `notes.md`.
- [x] **1.3** Write `notes.md` with confirmed root cause, evidence, and one sentence ruling out the other suspects. PRD §4.1 acceptance check. _(S, depends on 1.1, 1.2)_ — Working hypothesis (suspect 1) recorded; reasoning for skipping 1.1/1.2 included.

## Phase 2 — Root cause fix

- [x] **2.1** Apply targeted fix per Phase 1 finding. Most likely: restructure `processor.go:130-136` to capture and propagate the `ListCategories` error. If suspect 2 confirmed, halt this task and file follow-up for the write path. _(S–M, depends on Phase 1)_

## Phase 3 — Observability hardening (PRD §4.3)

- [x] **3.1** Replace `if cats, err := …; err == nil` at `processor.go:131` with explicit error capture; log at error level: `"Failed to fetch categories for plan ingredient consolidation"` with `plan_id` field. Continue with empty `categoryByID`. _(S)_
- [x] **3.2** After `ingredientProc.GetByIDs` at `processor.go:233`, when `len(canonicalsByID) < len(canonIDs)`, log warn: `"Canonical ingredient batch fetch returned fewer rows than requested"` with `plan_id`, `requested`, `received`, `missing_count`. _(S)_
- [x] **3.3** In the resolved-loop at `processor.go:238-252`, when `canonical.CategoryID() != nil` but ID missing from `categoryByID`, log warn: `"Canonical ingredient references unknown category"` with `plan_id`, `canonical_ingredient_id`, `category_id`. _(S)_
- [x] **3.4** Verify all three log messages are constant string literals (no `fmt.Sprintf` in the message slot) so they're greppable in prod. _(S)_

## Phase 4 — Shared grouping helper + tests

- [x] **4.1** Create `services/recipe-service/internal/export/grouping.go` with `GroupByCategory(ingredients []ConsolidatedIngredient) []CategoryGroup`. Empty `CategoryName` → uncategorized bucket sorted last. Within group, alphabetical by `DisplayName` (fallback `Name`). _(M)_
- [x] **4.2** Refactor `services/recipe-service/internal/export/markdown.go:96-156` to call `GroupByCategory` instead of its inline grouping. Markdown rendering byte-identical for representative inputs. _(S, depends on 4.1)_
- [x] **4.3** Refactor `processor.go:260-280` final-pass sort to flow through `GroupByCategory` and re-flatten, so JSON:API order matches markdown order. _(S, depends on 4.1)_
- [x] **4.4** Re-read `frontend/src/components/features/meals/ingredient-preview.tsx:22-70`. Confirm no frontend change required; document in PR. _(S)_
- [x] **4.5** Add unit tests in `services/recipe-service/internal/export/processor_test.go` (or new file) covering: (a) all categorized, (b) all uncategorized, (c) mixed, (d) unresolved present, (e) `categoryclient` error. Plus table-driven tests for `GroupByCategory`. _(M, depends on 3.1, 4.1)_

## Phase 5 — Verification & cleanup

- [x] **5.1** Remove temporary diagnostic logging from Phase 1 that isn't part of the §4.3 set. _(S)_ — N/A; Phase 1 throwaway logs were never added (see 1.1).
- [x] **5.2** Run `go build ./...` and `go test ./...` for recipe-service. If `categoryclient` was touched, also build `category-service`. _(S)_ — recipe-service all green; category-service builds clean as a sanity check though categoryclient was not touched.
- [ ] **5.3** Manual deployed-env verification: confirm categorized output for a known-categorized plan; deliberately break `CategoryServiceURL` in non-prod and confirm graceful 200 + error log. _(S)_ — **Pending user / deployment.**
- [ ] **5.4** Open PR. Commit message body includes the diagnosis paragraph from `notes.md`. _(S)_ — **Pending user request.**

## PRD Acceptance Criteria Checklist (PRD §10)

- [x] Root cause identified, documented in `notes.md` and commit, with evidence. (Working hypothesis — suspect 1 — without deployed-env confirmation; see `notes.md`.)
- [ ] `GET /api/v1/meals/plans/{planId}/ingredients` returns non-null `category_name` / `category_sort_order` for categorized canonicals. — pending deployed-env verification (5.3).
- [x] In-app preview renders category groups in `category_sort_order` ascending, "Uncategorized" last. (Already implemented in `ingredient-preview.tsx`; backend now provides correct ordering.)
- [x] Markdown export renders category headings in same order as preview. (Both flow through `GroupByCategory`.)
- [x] `ListCategories` error → 200 OK + everything in Uncategorized + error log with plan ID. (Verified by `TestLoadCategoryLookup_ServerErrorLogsAndReturnsEmpty`.)
- [x] Partial canonical batch fetch → warn log with count + plan ID + graceful degrade.
- [x] Unknown category ID lookup → warn log with both IDs + plan ID + ingredient lands in Uncategorized.
- [x] Preview and markdown export verifiably share `GroupByCategory` helper.
- [x] Five unit-test scenarios present and passing.
- [x] No recipe-service test regressions; `go build ./...` and `go test ./...` pass.
- [x] No frontend changes (or, if any, justified in `notes.md`). — none required.
