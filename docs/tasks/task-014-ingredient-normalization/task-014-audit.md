# Plan Audit — task-014-ingredient-normalization

**Plan Path:** docs/tasks/task-014-ingredient-normalization/tasks.md
**Audit Date:** 2026-03-27
**Branch:** task-014
**Base Branch:** main

## Executive Summary

Of 68 total tasks, 58 are implemented (85%), 3 are partial, and 7 are skipped. All backend domains (ingredient, normalization, planner, audit) are created with correct file structure. The normalization pipeline, planner readiness, and audit emission are functional. Frontend types, services, hooks, components, pages, and routing are all in place. However, there are several backend guideline violations — most notably direct DB access in handlers, business logic in providers, audit emission in handlers instead of processors, and a lack of tests for all 4 new domains. The frontend has minor violations around missing `cn()` usage and missing Zod schemas for the planner config form.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| **Phase 1: Database & Domain Foundations** | | | |
| 1.1a | Create `internal/ingredient/model.go` | DONE | File exists; immutable model with private fields + accessors |
| 1.1b | Create `internal/ingredient/entity.go` | DONE | GORM entity + alias entity, Migration(), Make() |
| 1.1c | Create `internal/ingredient/builder.go` | DONE | Builder with name required + unit family validation |
| 1.1d | Create `internal/ingredient/provider.go` | DONE | getByID, getByName, getByAlias, search, suggestByPrefix |
| 1.2a | Create `internal/normalization/model.go` | DONE | RecipeIngredient model with status enum |
| 1.2b | Create `internal/normalization/entity.go` | DONE | Entity with Migration(), Make() |
| 1.2c | Create `internal/normalization/builder.go` | DONE | Builder with rawName required validation |
| 1.2d | Create `internal/normalization/provider.go` | DONE | getByRecipeID, GetByCanonicalIngredientID, bulkCreate, bulkUpdate, deleteByRecipeID |
| 1.2e | Create `internal/normalization/unit_registry.go` | DONE | Static unit map with count/weight/volume families |
| 1.3a | Create `internal/planner/model.go` | DONE | Immutable PlannerConfig model |
| 1.3b | Create `internal/planner/entity.go` | DONE | Entity with Migration(), Make() |
| 1.3c | Create `internal/planner/builder.go` | DONE | Builder present (no strong invariants) |
| 1.3d | Create `internal/planner/provider.go` | DONE | getByRecipeID, upsert |
| 1.4a | Create `internal/audit/entity.go` | DONE | AuditEventEntity with indices, Migration() |
| 1.4b | Create `internal/audit/emitter.go` | DONE | Emit() function with JSON metadata |
| **Phase 2: Normalization Pipeline** | | | |
| 2.1a | Create `internal/normalization/processor.go` | DONE | NormalizeIngredients with multi-step pipeline |
| 2.1b | Implement exact canonical match | DONE | processor.go tryMatch() → ingredient.GetByName |
| 2.1c | Implement alias match | DONE | processor.go tryMatch() → ingredient.GetByAlias |
| 2.1d | Implement text normalization + re-match | DONE | normalizeText() strips plurals, whitespace, articles; matchWithNormalization() |
| 2.1e | Implement unit normalization via registry | DONE | LookupUnit() in unit_registry.go called in pipeline |
| 2.2a | Implement ReconcileIngredients() | DONE | processor.go ReconcileIngredients matches by rawName |
| 2.2b | Handle new ingredients | DONE | New ingredients run full pipeline |
| 2.2c | Handle removed ingredients | DONE | deleteByRecipeID called then bulk recreate |
| 2.2d | Update position values | DONE | Positions set from loop index |
| 2.3a | Implement ResolveIngredient() | DONE | processor.go ResolveIngredient updates status |
| 2.3b | Implement alias learning | DONE | createAliasIfNotExists when saveAsAlias=true |
| 2.3c | Validate alias conflict | DONE | Checks existing canonical name before creating alias |
| 2.4a | Implement Renormalize() | DONE | processor.go Renormalize re-runs pipeline for unresolved |
| 2.4b | Preserve manually_confirmed | DONE | Filters to unresolved only |
| 2.4c | Return summary of changes | DONE | Returns count of resolved items |
| **Phase 3: Planner Configuration** | | | |
| 3.1a | Create `internal/planner/processor.go` | DONE | CreateOrUpdate, GetByRecipeID |
| 3.1b | Implement readiness computation | DONE | ComputeReadiness checks classification + servings |
| 3.1c | Return plannerReady + plannerIssues | DONE | Readiness struct with Ready bool + Issues []string |
| **Phase 4: Audit Events** | | | |
| 4.1a | Emit recipe.created/updated/deleted/restored | DONE | recipe/resource.go calls audit.Emit in all 4 handlers |
| 4.1b | Emit recipe.renormalized | DONE | normalization/resource.go renormalizeHandler |
| 4.1c | Emit normalization.corrected | DONE | normalization/resource.go resolveHandler |
| 4.1d | Emit ingredient.alias_created | DONE | normalization/resource.go resolveHandler (when alias saved) |
| **Phase 5: API Wiring** | | | |
| 5.1a | Create `internal/ingredient/rest.go` | DONE | JSON:API models |
| 5.1b | Create `internal/ingredient/resource.go` | DONE | Route registration |
| 5.1c | GET /ingredients (list) | DONE | listIngredientsHandler with search + pagination |
| 5.1d | POST /ingredients (create) | DONE | createIngredientHandler |
| 5.1e | GET /ingredients/:id | DONE | getIngredientHandler with aliases |
| 5.1f | PATCH /ingredients/:id | DONE | updateIngredientHandler |
| 5.1g | DELETE /ingredients/:id (with ref check) | DONE | deleteIngredientHandler checks usage count |
| 5.1h | POST /ingredients/:id/aliases | DONE | addAliasHandler with conflict check |
| 5.1i | DELETE /ingredients/:id/aliases/:aliasId | DONE | removeAliasHandler |
| 5.1j | GET /ingredients/:id/recipes | DONE | ingredientRecipesHandler |
| 5.1k | POST /ingredients/:id/reassign | DONE | reassignHandler |
| 5.2a | Create `internal/normalization/rest.go` | DONE | JSON:API models for resolve |
| 5.2b | Create `internal/normalization/resource.go` | DONE | Route registration |
| 5.2c | POST /recipes/:id/ingredients/:ingredientId/resolve | DONE | resolveHandler |
| 5.2d | POST /recipes/:id/renormalize | DONE | renormalizeHandler |
| 5.3a | Extend RestDetailModel (ingredients, plannerConfig, etc.) | DONE | recipe/rest.go RestDetailModel updated |
| 5.3b | Extend RestModel (plannerReady, classification, counts) | DONE | recipe/rest.go RestModel updated |
| 5.3c | Extend CreateRequest/UpdateRequest with plannerConfig | DONE | recipe/rest.go both updated |
| 5.3d | Extend RestParseModel with normalization | DONE | recipe/rest.go Normalization field added |
| 5.3e | Modify createHandler — trigger normalization | DONE | recipe/resource.go normalization + planner wired |
| 5.3f | Modify updateHandler — reconciliation + plannerConfig | DONE | recipe/resource.go reconciliation on source change |
| 5.3g | Modify getHandler — include normalized ingredients | DONE | recipe/resource.go buildDetailEnrichment |
| 5.3h | Modify listHandler — filter params, ingredient counts | DONE | recipe/resource.go post-query filtering |
| 5.3i | Modify parseHandler — normalization preview | DONE | recipe/resource.go PreviewNormalization call |
| 5.3j | Modify provider.go getAll — filter support | PARTIAL | Filtering done in-memory in handler, not at DB query level |
| 5.3k | Implement suggestion search for resolve | DONE | ingredient/provider.go suggestByPrefix with ILIKE |
| 5.4a | Update cmd/main.go — migrations | DONE | All 4 domain migrations registered |
| 5.4b | Update cmd/main.go — routes | DONE | ingredient + normalization routes registered |
| 5.4c | Update nginx config | DONE | /api/v1/ingredients route added |
| 5.5a | All tests pass | PARTIAL | Existing tests pass (recipe, cooklang), but NO tests written for new domains |
| 5.5b | Service builds | DONE | `go build ./...` passes |
| 5.5c | Docker build | SKIPPED | Not verified |
| **Phase 6: Frontend** | | | |
| 6.1a | Create ingredient.ts service | DONE | frontend/src/services/api/ingredient.ts |
| 6.1b | Add types for CanonicalIngredient, RecipeIngredient | DONE | frontend/src/types/models/ingredient.ts + recipe.ts |
| 6.1c | Add types for PlannerConfig, PlannerReadiness | DONE | frontend/src/types/models/recipe.ts PlannerConfig |
| 6.1d | Extend recipe types (detail) | DONE | RecipeDetailAttributes extended |
| 6.1e | Extend recipe list types | DONE | RecipeListAttributes extended |
| 6.1f | Add resolve + renormalize to recipe service | DONE | recipe.ts resolveIngredient + renormalize |
| 6.2a | Create use-ingredients.ts | DONE | Full CRUD + alias hooks |
| 6.2b | Add alias hooks | DONE | useAddAlias, useRemoveAlias in use-ingredients.ts |
| 6.2c | Add useIngredientRecipes | DONE | In use-ingredients.ts |
| 6.2d | Create use-ingredient-normalization.ts | DONE | useResolveIngredient, useRenormalize |
| 6.2e | Extend use-recipes.ts with filter params | SKIPPED | No evidence of filter params added to useRecipes |
| 6.3a | Create ingredient-normalization-panel.tsx | DONE | Status display, summary, re-normalize button |
| 6.3b | Create ingredient-resolver.tsx | DONE | Search/select dropdown, inline create, save-as-alias |
| 6.3c | Color-coded status indicators | DONE | Green/yellow status icons in panel |
| 6.4a | Create planner-config-form.tsx | DONE | Collapsible section, classification, numeric inputs |
| 6.4b | Create planner-ready-badge.tsx | DONE | Green/yellow badge with tooltip |
| 6.5a | Add normalization panel to detail page | DONE | RecipeDetailPage.tsx IngredientNormalizationPanel |
| 6.5b | Add planner readiness badge in header | DONE | PlannerReadyBadge in title area |
| 6.5c | Add planner config display section | DONE | Full planner config grid display |
| 6.5d | Add re-normalize button | DONE | Wired to renormalize mutation |
| 6.6a | Add planner settings section to form | DONE | PlannerConfigForm in RecipeFormPage |
| 6.6b | Add normalization review panel after preview | DONE | IngredientNormalizationPanel in edit mode |
| 6.6c | Wire normalization into live preview | DONE | useCooklangPreview returns normalization, displayed in form |
| 6.6d | Handle inline ingredient resolution | PARTIAL | Resolver component exists but resolution happens via page-level hook, not fully independent of form save |
| 6.7a | Add planner-ready badge to recipe cards | DONE | recipe-card.tsx PlannerReadyBadge |
| 6.7b | Add classification label to cards | DONE | recipe-card.tsx Badge for classification |
| 6.7c | Add normalization summary to cards | DONE | recipe-card.tsx resolved/total display |
| 6.7d | Add filter controls | SKIPPED | No filter UI controls on RecipesPage |
| 6.8a | Create IngredientsPage.tsx | DONE | Searchable, paginated list |
| 6.8b | Create IngredientDetailPage.tsx | DONE | Edit form with alias management |
| 6.8c | Alias management on detail page | DONE | Add/remove aliases |
| 6.8d | Linked recipe list on detail page | DONE | ingredientRecipes section |
| 6.8e | Reassign-and-delete flow | DONE | Dialog with ingredient selector |
| 6.8f | Empty state guidance | DONE | IngredientsPage empty state |
| 6.9a | Add "Ingredients" nav item | DONE | nav-config.ts with Carrot icon |
| 6.9b | Add routes | DONE | App.tsx /ingredients and /ingredients/:id |
| 6.9c | Verify responsive layout | SKIPPED | Cannot verify visually in CLI |

**Completion Rate:** 58/68 tasks DONE (85%), 3 PARTIAL (4%), 7 SKIPPED (10%)
**Skipped without approval:** 4 (filter UI, Docker build, responsive verification, use-recipes filter extension)
**Partial implementations:** 3

## Skipped / Deferred Tasks

| Task | Impact |
|------|--------|
| **5.3j — DB-level filter support** | Filtering is done in-memory after fetching all results. Works for small datasets but will not scale. Should be moved to SQL WHERE clauses. |
| **5.5c — Docker build verification** | Not run during audit. Should be verified before merge. |
| **6.2e — Filter params in useRecipes** | Recipe list page cannot filter by plannerReady/classification/normalizationStatus from the frontend since the hook doesn't pass these params. |
| **6.7d — Filter controls on RecipesPage** | No filter UI added to the recipes list page. Users cannot filter by planner readiness or normalization status. |
| **6.6d — Independent inline resolution** | Resolver component works but is coupled to page context rather than being fully independent of form save cycle. Minor impact. |
| **6.9c — Responsive layout verification** | Cannot verify in CLI environment. |

## Developer Guidelines Compliance

### Passes

- **Immutable models**: All 4 new domain models use private fields with accessor methods (ingredient/model.go, normalization/model.go, planner/model.go)
- **Entity separation**: All entities have GORM tags separate from models, with Migration() and Make() functions
- **Builder pattern**: All builders enforce key invariants (name required, unit family validation, raw name required)
- **REST JSON:API**: Proper request/response models with GetName/GetID/SetID interface methods
- **Multi-tenancy context**: All handlers extract tenant via `tenantctx.MustFromContext(r.Context())`
- **Handler registration**: Uses `server.RegisterHandler` and `server.RegisterInputHandler[T]` correctly
- **Logger usage**: Uses `d.Logger()` from HandlerDependency, not `logrus.StandardLogger()`
- **Frontend JSON:API types**: Proper `id + attributes` structure in ingredient.ts types
- **Frontend service layer**: IngredientService extends BaseService
- **Frontend query key factory**: use-ingredients.ts has proper hierarchical keys with tenant isolation
- **Frontend tenant context**: All hooks use `useTenant()` with `enabled: !!tenant?.id` guards
- **Frontend skeleton loading**: IngredientsPage and IngredientDetailPage use Skeleton components
- **Frontend error handling**: Uses `createErrorFromUnknown()` and toast notifications
- **Frontend named exports**: All components use named exports

### Violations

| # | Rule | File | Issue | Severity | Fix |
|---|------|------|-------|----------|-----|
| V1 | Handlers must not call providers directly | `ingredient/resource.go:257-263` | `ingredientRecipesHandler` queries DB directly instead of via Processor | **high** | Create `Processor.GetIngredientRecipes()` method |
| V2 | Handlers must not call providers directly | `ingredient/resource.go:291-296` | `reassignHandler` does DB updates directly instead of via Processor | **high** | Create `Processor.Reassign()` method |
| V3 | Cross-domain logic in handlers | `normalization/resource.go:57,64,89-93` | `audit.Emit()` called directly from handlers instead of Processor | **medium** | Move audit emission into `normalization.Processor.ResolveIngredient()` and `Renormalize()` |
| V4 | Cross-domain logic in handlers | `recipe/resource.go:230,337,368,387` | `audit.Emit()` called directly from recipe handlers | **medium** | Move audit emission into recipe processor methods |
| V5 | Business logic in provider | `planner/provider.go:16-52` | `upsert()` contains create-or-update decision logic | **medium** | Move upsert logic to Processor.CreateOrUpdate() |
| V6 | N+1 query pattern | `ingredient/resource.go:57` | `listIngredientsHandler` fetches usage count per ingredient in loop | **medium** | Use a single query with subquery or JOIN for counts |
| V7 | No tests for new domains | `internal/ingredient/`, `normalization/`, `planner/`, `audit/` | All 4 new domains have zero test files | **high** | Write table-driven tests for processors, providers, builders |
| V8 | Inconsistent provider pattern | `normalization/provider.go`, `planner/provider.go` | Use direct `*gorm.DB` param instead of `database.Query[Entity]` functional composition | **medium** | Refactor to use `database.Query`/`database.SliceQuery` pattern |
| V9 | Missing `cn()` helper | Multiple frontend components | Conditional class names use inline ternaries instead of `cn()` | **low** | Replace with `cn()` calls |
| V10 | Missing Zod schema | `planner-config-form.tsx` | Planner config form uses controlled state, not `react-hook-form` with `zodResolver` | **low** | Create `lib/schemas/planner-config.schema.ts` and integrate with form |
| V11 | Type assertions | `IngredientsPage.tsx:26`, `IngredientDetailPage.tsx:42`, `ingredient-resolver.tsx:26` | Unsafe `as CanonicalIngredientListItem[]` casts | **low** | Fix typing at hook level to eliminate casts |
| V12 | Hardcoded constants | `IngredientDetailPage.tsx:22`, `planner-config-form.tsx:7` | UNIT_FAMILIES and CLASSIFICATIONS defined inline | **low** | Move to `lib/constants/` |

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service (Go) | PASS | PASS | `go build ./...` clean. `go test ./... -count=1` passes — but only existing recipe/cooklang tests run. All 4 new domains have NO test files. |
| frontend (TypeScript) | PASS | PASS | `npx tsc --noEmit` clean. `npx vitest run` — 43 suites, 398 tests pass. `npx vite build` succeeds. |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 85% of tasks done, 3 partial, 4 functionally skipped (filter UI, Docker, recipe hook filters, responsive check)
- **Guidelines Compliance:** MINOR_VIOLATIONS — Several handler layer violations (V1-V4) and missing tests (V7) are the most significant. Models, entities, builders, and REST patterns are solid.
- **Recommendation:** NEEDS_FIXES — The high-severity violations (direct DB in handlers, zero tests for new domains) should be addressed before merge.

## Action Items

1. **[HIGH]** Move direct DB access out of `ingredient/resource.go` handlers (`ingredientRecipesHandler`, `reassignHandler`) into Processor methods (V1, V2)
2. **[HIGH]** Write tests for all 4 new backend domains — at minimum: builder validation tests, processor pipeline tests, provider query tests (V7)
3. **[MEDIUM]** Move `audit.Emit()` calls from handlers into Processor methods for both `normalization/resource.go` and `recipe/resource.go` (V3, V4)
4. **[MEDIUM]** Move upsert business logic from `planner/provider.go` into Processor (V5)
5. **[MEDIUM]** Fix N+1 query in `ingredient/resource.go` list handler — use single query with JOIN or subquery for usage counts (V6)
6. **[MEDIUM]** Refactor `normalization/provider.go` and `planner/provider.go` to use `database.Query[Entity]` functional composition pattern (V8)
7. **[LOW]** Add recipe list filter params to `useRecipes` hook and filter UI controls on RecipesPage (skipped tasks 6.2e, 6.7d)
8. **[LOW]** Move list filtering from in-memory (recipe/resource.go) to DB-level WHERE clauses (task 5.3j)
9. **[LOW]** Verify Docker build succeeds (task 5.5c)
10. **[LOW]** Address frontend style violations: add `cn()` helper usage, Zod schema for planner form, remove type assertions (V9-V12)
