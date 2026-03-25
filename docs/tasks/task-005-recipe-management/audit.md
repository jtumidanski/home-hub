# Plan Audit — task-005-recipe-management

**Plan Path:** docs/tasks/task-005-recipe-management/tasks.md
**Audit Date:** 2026-03-25
**Branch:** task-005
**Base Branch:** main

## Executive Summary

30 of 34 tasks were completed (88%). The core feature is fully functional — recipe-service with Cooklang parsing, all 8 API endpoints, 3 frontend pages with live preview, and full infrastructure integration. Two tasks were skipped (handler tests and frontend component tests — no test files exist for the recipe domain or recipe components). Two planned frontend files were not created (cooklang-editor.tsx, tag-input.tsx) but their functionality was absorbed into other components. The backend has 3 guideline violations around layer separation (missing administrator.go, write operations in provider.go, non-standard provider signature).

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Initialize Go module and directory structure | DONE | `services/recipe-service/go.mod`, `cmd/main.go`, `internal/config/config.go` |
| 1.2 | Add recipe-service to Go workspace | DONE | `go.work` line 6: `./services/recipe-service` |
| 1.3 | Create GORM entities and migration | DONE | `entity.go`: Entity, TagEntity, RestorationEntity with Migration() |
| 1.4 | Create domain model and builder | DONE | `model.go`: immutable with accessors; `builder.go`: fluent with invariants |
| 1.5 | Create provider — database access layer | DONE | `provider.go`: getByID, getAll, getAllTags, getDeletedByID (also contains writes — see violations) |
| 1.6 | Verify service builds and starts | DONE | `go build ./services/recipe-service/...` passes |
| 2.1 | Define parser types | DONE | `cooklang/types.go`: Segment, Step, Ingredient, Metadata, ParseResult, ParseError |
| 2.2 | Implement core parser | DONE | `cooklang/parser.go`: @ingredients, #cookware, ~timers, comments, steps, metadata blocks, sections, blockquotes, recipe references |
| 2.3 | Implement parser validation | DONE | `cooklang/parser.go`: Validate() with line/column error reporting, MaxSourceSize check |
| 2.4 | Write comprehensive parser tests | DONE | `cooklang/parser_test.go`: 28 tests covering all syntax features including full carbonara and stuffed peppers integration tests |
| 3.1 | Create processor — business logic orchestration | DONE | `processor.go`: Create, Get, List, Update, Delete, Restore, ListTags, ParseSource with metadata derivation |
| 3.2 | Create REST mappings — JSON:API serialization | DONE | `rest.go`: RestModel, RestDetailModel, RestTagModel, RestParseModel, CreateRequest, UpdateRequest, ParseRequest, RestorationRequest |
| 3.3 | Create HTTP resource — handlers and route registration | DONE | `resource.go`: 8 handlers (parse, list, create, get, update, delete, restore, tags) |
| 3.4 | Write handler tests | SKIPPED | No test files found under `services/recipe-service/internal/recipe/` — only parser tests exist |
| 3.5 | Create service documentation | DONE | `docs/domain.md`, `docs/rest.md`, `docs/storage.md` |
| 3.6 | End-to-end backend verification | DONE | Build passes, 28 tests pass |
| 4.1 | Create TypeScript types for recipes | DONE | `types/models/recipe.ts`: all interfaces including Segment, Step, Ingredient, RecipeMetadata, RecipeParseResult |
| 4.2 | Create recipe API service | DONE | `services/api/recipe.ts`: RecipeService extends BaseService with all 8 methods |
| 4.3 | Create React Query hooks | DONE | `lib/hooks/api/use-recipes.ts`: key factory, useRecipes, useRecipe, useRecipeTags, useCreateRecipe, useUpdateRecipe, useDeleteRecipe, useParseRecipe |
| 4.4 | Create Zod schemas for recipe forms | DONE | `lib/schemas/recipe.schema.ts`: recipeFormSchema with title, description, source |
| 4.5 | Create live preview hook | DONE | `lib/hooks/use-cooklang-preview.ts`: debounced server-side parse with cancellation |
| 4.6 | Create recipe list page | DONE | `pages/RecipesPage.tsx`: card list, search bar, tag filter, empty state, pull-to-refresh |
| 4.7 | Create recipe detail page | DONE | `pages/RecipeDetailPage.tsx`: ingredients + steps side-by-side, metadata, edit/delete |
| 4.8 | Create recipe create/edit page with live preview | DONE | `pages/RecipeFormPage.tsx`: form + live preview, metadata derivation display, cooklang help |
| 4.9 | Add routing and navigation | DONE | `App.tsx`: 4 routes; `app-shell.tsx` + `mobile-drawer.tsx`: UtensilsCrossed nav item |
| 4.10 | Frontend tests — components | SKIPPED | No test files found for recipe components or pages |
| 5.1 | Create Dockerfile | DONE | `services/recipe-service/Dockerfile` |
| 5.2 | Create build script + update build-all.sh | DONE | `scripts/build-recipe.sh`, `scripts/build-all.sh` updated |
| 5.3 | Add to Docker Compose | DONE | `deploy/compose/docker-compose.yml`: recipe-service container |
| 5.4 | Add nginx routing | DONE | `deploy/compose/nginx.conf`: upstream + location for /api/v1/recipes |
| 5.5 | Add K8s manifests | DONE | `deploy/k8s/recipe-service.yaml` + `ingress.yaml` updated |
| 5.6 | Add CI pipeline | DONE | `pr.yml`: recipe detection + build-recipe job + docker matrix; `main.yml`: build matrix |
| 5.7 | Update supporting scripts | DONE | `test-all.sh`, `lint-all.sh`, `ci-build.sh`, `ci-test.sh` all updated |
| 5.8 | Create Bruno collection | DONE | `bruno/recipe/`: 8 .bru files for all endpoints |

**Completion Rate:** 30/34 tasks (88%)
**Skipped without approval:** 2 (handler tests, frontend component tests)
**Partial implementations:** 0

## Skipped / Deferred Tasks

### 3.4 — Write handler tests
**Impact: Medium.** No unit tests exist for the processor, provider, or resource handlers. Only the Cooklang parser package has tests (28 tests). The processor contains significant business logic (metadata derivation, tag merging, validation orchestration) that is untested at the unit level. Risk is mitigated by the parser being well-tested and the service working end-to-end.

### 4.10 — Frontend tests — components
**Impact: Medium.** No test files were created for any recipe components or pages. The existing 277 frontend tests all pass (none are recipe-related). Key untested areas: recipe form submission flow, search/filter behavior, live preview hook debounce logic, tag filter toggle. Risk is mitigated by TypeScript strict mode catching type errors at build time.

## Developer Guidelines Compliance

### Passes

**Backend:**
- **model.go**: Immutable model with private fields and accessor methods ✓
- **entity.go**: GORM tags on entities only, TableName(), Migration(), Make(), ToEntity() ✓
- **builder.go**: Fluent builder with invariant enforcement (title + source required) ✓
- **resource.go**: Uses RegisterHandler/RegisterInputHandler, tenant from context ✓
- **rest.go**: All JSON:API types implement GetName()/GetID()/SetID() ✓
- **main.go**: Correct bootstrap pattern (logging, tracing, DB, auth, routes) ✓
- **Multi-tenancy**: tenantctx.MustFromContext used in create handler ✓

**Frontend:**
- **Types**: Proper JSON:API structure (id + type + attributes) ✓
- **Service**: Extends BaseService with tenant handling ✓
- **Hooks**: Query key factory with hierarchical keys, `as const`, tenant scoping ✓
- **Forms**: react-hook-form with zodResolver, schema in lib/schemas/ ✓
- **Error handling**: createErrorFromUnknown() + toast notifications ✓
- **Loading states**: Skeleton components for page load ✓
- **Styling**: Tailwind utilities with cn() ✓
- **Tenant context**: useTenant() properly used ✓
- **No `any` types**: TypeScript strict mode compliant ✓

### Violations

1. **Rule:** Write operations must be in administrator.go, not provider.go
   - **File:** `services/recipe-service/internal/recipe/provider.go:65-101`
   - **Issue:** Functions `create()`, `save()`, `softDelete()`, `restoreByID()`, `replaceTags()` are write operations placed in provider.go. Per the established pattern (see `services/productivity-service/internal/task/administrator.go`), write operations belong in a separate `administrator.go` file. Providers should be read-only.
   - **Severity:** medium
   - **Fix:** Extract write functions to `administrator.go`

2. **Rule:** Processor should delegate writes to administrator functions, not perform direct entity creation
   - **File:** `services/recipe-service/internal/recipe/processor.go:100-120`
   - **Issue:** The Create() method constructs Entity structs directly and calls `create()`. Per the productivity-service pattern, the processor should call administrator functions that handle entity construction. Line 251 also has a raw `p.db.WithContext(p.ctx).Create(&re)` call.
   - **Severity:** medium
   - **Fix:** Move entity construction and DB writes to administrator.go, have processor call administrator functions

3. **Rule:** Provider queries should return database.EntityProvider[T] for lazy evaluation
   - **File:** `services/recipe-service/internal/recipe/provider.go:17`
   - **Issue:** `getAll()` returns `func(db *gorm.DB) ([]Entity, int64, error)` instead of `database.EntityProvider[[]Entity]`. This breaks the lazy provider pattern and prevents composition with `model.SliceMap`.
   - **Severity:** low
   - **Fix:** Refactor to return `database.EntityProvider[[]Entity]` or accept that pagination queries need a different signature (document the deviation)

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service | PASS | PASS (28/28) | Tests only cover cooklang parser package |
| frontend | PASS | PASS (277/277) | No recipe-specific tests; all existing tests still pass |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 30/34 tasks done, 2 test tasks skipped
- **Guidelines Compliance:** MINOR_VIOLATIONS — 3 backend layer separation issues, frontend fully compliant
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **[High]** Add `administrator.go` — extract write operations (create, save, softDelete, restoreByID, replaceTags) from provider.go into administrator.go following the productivity-service pattern
2. **[High]** Refactor processor.go — delegate entity construction and DB writes to administrator functions; remove the raw `p.db.WithContext(p.ctx).Create(&re)` call on line 251
3. **[Medium]** Add handler/processor tests (task 3.4) — at minimum test the processor's Create, Update, and metadata derivation logic
4. **[Medium]** Add frontend component tests (task 4.10) — at minimum test the cooklang preview hook debounce behavior and recipe form submission
5. **[Low]** Update tasks.md — mark completed tasks as `[x]` to reflect actual state
6. **[Low]** Remove planned but unneeded files from context.md — `cooklang-editor.tsx` and `tag-input.tsx` were not created (functionality absorbed into RecipeFormPage and form fields respectively)
