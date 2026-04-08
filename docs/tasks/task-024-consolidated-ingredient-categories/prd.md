# Consolidated Ingredient Categories — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-08

---

## 1. Overview

The meal planner's "consolidated ingredients" view aggregates every recipe's ingredients across a plan week into a single shopping-style list. The frontend (`frontend/src/components/features/meals/ingredient-preview.tsx`) is already written to render this list grouped by ingredient category, and the backend (`services/recipe-service/internal/export/processor.go::ConsolidateIngredients`) is already written to fetch categories from `category-service` via `categoryclient` and to attach `CategoryName` / `CategorySortOrder` to each consolidated ingredient. The end-to-end pipeline exists.

In practice, in deployed environments the rendered list shows every ingredient as a single ungrouped block (everything falling into "Uncategorized"), even though those same canonical ingredients are visibly categorized everywhere else (canonical ingredient admin, shopping list). No errors are visible in the browser console, suggesting the failure is silent and server-side.

This task is a bug fix: diagnose the root cause of the silent category drop in `ConsolidateIngredients`, fix it, and harden the code path so future failures of the same shape are loud rather than silent. The fix must cover both the in-app preview and the markdown export, since they share the same backend producer.

## 2. Goals

Primary goals:
- Restore category grouping in the meal plan consolidated ingredients preview, matching the categorization shown elsewhere (canonical ingredient admin, shopping list).
- Apply the same category grouping to the meal plan markdown export so its sections mirror the in-app preview.
- Ensure failures in the category lookup or canonical ingredient batch fetch are logged with enough detail to diagnose future incidents, instead of silently degrading to "Uncategorized".
- Reach root cause — do not patch the symptom by hardcoding categories or bypassing `category-service`.

Non-goals:
- No redesign of the consolidated ingredient UI beyond what's needed to render category groups.
- No changes to how categories are assigned to canonical ingredients (covered by task-016).
- No changes to the unit-family consolidation / quantity merging logic.
- No new categorization for unresolved ingredients (free-text ingredients without a canonical match continue to fall under "Uncategorized").
- No new category assignment UI or workflow.

## 3. User Stories

- As a household meal planner, I want the consolidated ingredient list for a week to be grouped by category (Produce, Dairy, Pantry, …) so that I can shop section-by-section in the store, the same way I already shop from the shopping list.
- As a household meal planner, I want the markdown export of a meal plan to have the same category grouping as the in-app preview, so the printed/shared version is just as useful.
- As a household meal planner, I want ingredients with no category (e.g., free-text or uncategorized canonical ingredients) to appear in a single "Uncategorized" group at the bottom of the list, so they aren't lost.
- As a developer on call, I want category-lookup or canonical-ingredient batch-fetch failures inside `ConsolidateIngredients` to produce clear server logs, so I can diagnose silent degradation quickly the next time something breaks.

## 4. Functional Requirements

### 4.1 Diagnosis (must precede the fix)

Before writing the fix, the implementer MUST identify which of the following is the actual root cause in the deployed environment, and document the finding in the task's commit message and/or `notes.md`:

1. `categoryclient.ListCategories` is returning an error (network, 401, malformed body) and the error is being silently swallowed at `processor.go:131` (`if cats, err := …; err == nil`).
2. `ingredientProc.GetByIDs` (the canonical ingredients batch fetch) is returning an empty / partial map, so `acc.categoryName` never gets assigned even though `categoryByID` is populated.
3. Canonical ingredients in the deployed DB have `category_id IS NULL` despite the admin UI appearing to have set them — i.e., a write-path bug elsewhere, not a read-path bug in `ConsolidateIngredients`. (If this is the case, raise a follow-up task; do not retroactively re-categorize data as part of this task.)
4. `categoryByID` lookup key mismatch — e.g., the canonical ingredient stores a category ID that doesn't appear in the response from `category-service` (cross-tenant pollution, deleted category, etc.).
5. Some other cause discovered during investigation.

The acceptance check for diagnosis is a one-paragraph written explanation of *which* of the above (or other) causes was confirmed, with evidence (logs, DB query, repro steps).

### 4.2 Fix

Once diagnosed, the implementer MUST fix the root cause such that:
- A meal plan whose canonical ingredients have categories assigned in the recipe-service DB and whose categories exist in category-service for the same tenant produces a `GET /meals/plans/{planId}/ingredients` response where the relevant items have non-null `category_name` and `category_sort_order`.
- The frontend `IngredientPreview` component renders those items grouped by category, in `category_sort_order` ascending, with a final "Uncategorized" group when applicable. (No frontend changes are anticipated; the existing grouping logic at `ingredient-preview.tsx:22-70` already handles this. If the diagnosis reveals a frontend-side bug, the frontend MAY be modified.)

### 4.3 Observability hardening

Independent of which root cause turns out to be load-bearing, the following defensive logging MUST be added inside `ConsolidateIngredients`:

- The error branch of `p.catClient.ListCategories(...)` MUST log at error level with the error and a stable message (e.g., `"Failed to fetch categories for plan ingredient consolidation"`), including the plan ID. Today this branch is not even reached because of the `err == nil` guard — the structure must change so the error is captured.
- Cases where `len(canonicalsByID) < len(canonIDs)` MUST log at warn level with the count of missing canonical IDs and the plan ID. (Today they are silently dropped.)
- Cases where `acc.categoryName == ""` for a resolved ingredient whose canonical record has a non-nil `CategoryID()` but the ID is missing from `categoryByID` MUST log at warn level with the canonical ingredient ID, the missing category ID, and the plan ID — this catches the cross-tenant / deleted-category cases.

The log messages MUST be stable strings suitable for grepping in production logs. The fix MUST NOT change the response status code on these failures: the endpoint continues to return `200 OK` with whatever ingredients it could compute, with the affected items falling into "Uncategorized" as today (per question 6 in the design discussion).

### 4.4 Markdown export parity

The markdown export at `exportMarkdownHandler` (`services/recipe-service/internal/plan/resource.go:312` → `internal/export/markdown.go`) MUST also render the consolidated ingredient list grouped by category, with the same ordering and "Uncategorized" fallback as the in-app preview. Implementation detail: both call paths can — and should — share the same grouping/sorting helper so they cannot drift in the future.

The markdown format for a category section is:

```
### Produce
- 2 cup carrot
- 1 head lettuce

### Dairy
- 1 gallon milk

### Uncategorized
- 1 tbsp special sauce
```

(Heading level may be adjusted to match the surrounding markdown export style — the requirement is that there is one heading per category and ingredients appear under their heading in the same display order as the in-app preview.)

### 4.5 Unresolved ingredients

Ingredients that fall into the `unresolved` slice in `ConsolidateIngredients` (no canonical match) MUST continue to render under "Uncategorized" in both the preview and the markdown export. They MUST NOT be dropped and MUST NOT be promoted into a category.

## 5. API Surface

No new endpoints. No request/response schema changes are required by this task — `RestIngredientModel` already exposes `category_name` and `category_sort_order` (`services/recipe-service/internal/export/resource.go:11-23`), and the frontend type `PlanIngredient` already consumes them (`frontend/src/types/models/meal-plan.ts`).

Existing endpoints affected (behavior only, not shape):

- `GET /api/v1/meals/plans/{planId}/ingredients` — `category_name` / `category_sort_order` are now reliably populated for ingredients whose canonical record is categorized.
- `GET /api/v1/meals/plans/{planId}/export/markdown` — body is now grouped by category section.

If diagnosis (§4.1) reveals that the root cause requires forwarding additional headers from `categoryclient` to `category-service`, those headers are an internal contract change between two services and do not affect external API consumers.

## 6. Data Model

No schema changes. No migrations.

The fix relies on existing data:
- `canonical_ingredients.category_id` (recipe-service DB) — already exists, populated by the canonical ingredient admin UI (task-016).
- Categories owned by `category-service` per tenant — already exist.

If diagnosis (§4.1, option 3) reveals that `canonical_ingredients.category_id` is not actually being populated by some write path despite the UI appearing to set it, that finding MUST be raised as a separate follow-up task and not silently fixed by this one.

## 7. Service Impact

- **recipe-service** — Primary changes: `internal/export/processor.go` (the diagnosed fix + observability hardening + extracting a shared grouping helper), `internal/export/markdown.go` (rendering category sections), and possibly `internal/categoryclient/client.go` (only if diagnosis reveals a bug in how the client builds the request). New unit tests covering the categorized output and the "Uncategorized" fallback.
- **category-service** — No changes anticipated. Touched only if diagnosis reveals an issue in the category endpoint contract (e.g., missing field, pagination behavior).
- **frontend** — No changes anticipated. The grouping logic in `frontend/src/components/features/meals/ingredient-preview.tsx:22-70` is already correct. Touched only if diagnosis (§4.1) points the finger at a frontend-side bug.

## 8. Non-Functional Requirements

- **Performance** — `ConsolidateIngredients` already batches its DB calls. The fix MUST NOT introduce per-ingredient round-trips to category-service. The single existing `ListCategories` call per plan is acceptable; do not multiply it.
- **Multi-tenancy** — The `access_token` cookie carries the tenant claim, and `categoryclient` already forwards it. The fix MUST preserve tenant isolation: a plan in tenant A must never group its ingredients using categories from tenant B. If diagnosis reveals tenant bleed in either direction, that becomes a P0 sub-issue of this task.
- **Observability** — Per §4.3, three new log statements with stable messages. No new metrics required, but if a counter for "ingredients silently uncategorized due to category fetch failure" is trivial to add via the existing logging infra, it is welcomed.
- **Backward compatibility** — Response shape is unchanged. Clients that ignored `category_name` continue to work; clients that already use it (the existing frontend) start receiving non-null values.
- **Error handling at the boundary** — Per design question 6, category-service unreachability degrades gracefully to "everything in Uncategorized" plus a server-side error log. The plan ingredients endpoint MUST NOT return 5xx because of a category-service blip.

## 9. Open Questions

- None at spec time. If diagnosis (§4.1) reveals a finding that requires a product decision (e.g., "the canonical ingredients in the deployed DB really do have NULL category_id — should we backfill?"), the implementer should pause and raise it before continuing.

## 10. Acceptance Criteria

- [ ] Root cause of the deployed-environment failure has been identified, documented in writing (commit message or `notes.md` in this task folder), and supported by concrete evidence.
- [ ] In a deployed-like environment with at least one canonical ingredient assigned to a category, calling `GET /api/v1/meals/plans/{planId}/ingredients` for a plan that uses that ingredient returns it with non-null `category_name` and `category_sort_order` in the JSON:API response.
- [ ] The meal plan ingredient preview UI renders ingredients grouped under category headings, with categories ordered by `category_sort_order` ascending, and a single "Uncategorized" group at the bottom containing both unresolved ingredients and resolved ingredients whose canonical record has no category.
- [ ] The meal plan markdown export (`GET .../export/markdown`) renders ingredients grouped under category headings using the same ordering as the preview.
- [ ] When `categoryclient.ListCategories` returns an error, the endpoint still returns `200 OK`, every ingredient appears in "Uncategorized", and the server logs an error message containing the plan ID and the underlying error.
- [ ] When the canonical ingredient batch fetch returns fewer rows than requested, a warning is logged with the count and plan ID, and the missing rows degrade to "Uncategorized" in the response (no panic, no 5xx).
- [ ] When a canonical ingredient references a `category_id` that is not present in the `categoryclient.ListCategories` response, a warning is logged with both IDs and the plan ID, and that ingredient appears in "Uncategorized".
- [ ] The preview and the markdown export use a shared grouping/sorting helper so they cannot drift; this is verified by code review.
- [ ] Unit tests cover: (a) all ingredients categorized, (b) all ingredients uncategorized, (c) mixed categorized + uncategorized, (d) unresolved ingredients present, (e) `categoryclient` returning an error.
- [ ] No regressions in existing recipe-service tests; `go build ./...` and `go test ./...` pass for recipe-service.
- [ ] No frontend changes were required (or, if they were, the diagnosis paragraph explains why and the change is minimal).
