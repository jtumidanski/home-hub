# Ingredient Normalization — Implementation Plan

Last Updated: 2026-03-27

## Executive Summary

Task-014 extends the recipe-service with ingredient normalization, a canonical ingredient registry, planner configuration, and audit events. This is a large feature spanning 4 new backend domains, modifications to the existing recipe domain, 2 new frontend pages, and extensions to 3 existing frontend pages. The work is structured in 6 phases: database/domain foundations, normalization pipeline, planner config, audit events, API wiring, and frontend.

## Current State Analysis

**Backend (recipe-service)**:
- Single `recipe` domain under `internal/recipe/`
- Cooklang parser produces `ParseResult` with aggregated `Ingredient` structs (name, quantity, unit)
- Ingredients are ephemeral — parsed on every read, never persisted
- No canonical ingredient concept, no normalization, no planner metadata
- Standard domain file pattern: model/entity/builder/processor/provider/resource/rest
- Database: `recipe.recipes`, `recipe.recipe_tags`, `recipe.recipe_restorations`
- Provider uses `administrator.go` for write operations, `provider.go` for reads

**Frontend**:
- `RecipesPage` — grid with search/tag filtering
- `RecipeDetailPage` — two-column layout (ingredients sidebar + steps)
- `RecipeFormPage` — source editor with live Cooklang preview (300ms debounce)
- `use-recipes.ts` — React Query hooks with query key factory
- `recipe.ts` — API service extending BaseService
- Navigation in `nav-config.ts` — Recipes under "lifestyle" group

## Proposed Future State

- 4 new backend domains: `ingredient`, `normalization`, `planner`, `audit`
- 5 new database tables: `canonical_ingredients`, `canonical_ingredient_aliases`, `recipe_ingredients`, `recipe_planner_configs`, `recipe_audit_events`
- Recipe create/update triggers normalization pipeline synchronously
- Parse preview includes normalization status
- Recipe detail/list responses enriched with ingredient status, planner config, planner readiness
- New canonical ingredient management pages in frontend
- Normalization review panel on recipe detail and form pages
- Planner settings section on recipe form
- Planner-ready badges on recipe list and detail

## Implementation Phases

### Phase 1: Database & Domain Foundations (Backend)

Build the four new domains with their data models, entities, builders, and providers. No business logic yet — just the persistence layer and type system.

**1.1 Canonical Ingredient Domain (`internal/ingredient/`)**
- `model.go` — Immutable model: id, tenantID, name, displayName, unitFamily, aliases[], aliasCount, usageCount, createdAt, updatedAt
- `entity.go` — CanonicalIngredientEntity + CanonicalIngredientAliasEntity, Migration(), Make() conversion
- `builder.go` — Builder with validation (name required, unitFamily must be valid enum or empty)
- `provider.go` — getByID, getByName, getByAlias, search (ILIKE + pagination), getUsageCount, getAllByTenantID
- Effort: **M**

**1.2 Recipe Ingredient / Normalization Domain (`internal/normalization/`)**
- `model.go` — RecipeIngredient model: id, tenantID, householdID, recipeID, rawName, rawQuantity, rawUnit, position, canonicalIngredientID, canonicalUnit, normalizationStatus, createdAt, updatedAt
- `entity.go` — RecipeIngredientEntity, Migration(), Make()
- `builder.go` — Builder with validation (rawName required, position >= 0, status must be valid enum)
- `provider.go` — getByRecipeID, getByCanonicalIngredientID, bulkCreate, bulkUpdate, deleteByRecipeID
- `unit_registry.go` — Static map of unit strings to canonical unit identities (count/weight/volume families)
- Effort: **M**

**1.3 Planner Config Domain (`internal/planner/`)**
- `model.go` — PlannerConfig model: id, recipeID, classification, servingsYield, eatWithinDays, minGapDays, maxConsecutiveDays
- `entity.go` — PlannerConfigEntity, Migration(), Make()
- `builder.go` — Builder with validation (classification must be valid enum if set)
- `provider.go` — getByRecipeID, upsert
- Effort: **S**

**1.4 Audit Domain (`internal/audit/`)**
- `entity.go` — AuditEventEntity, Migration()
- `emitter.go` — Emit(tenantID, entityType, entityID, action, actorID, metadata) function
- Effort: **S**

### Phase 2: Normalization Pipeline (Backend)

Core business logic for automatic ingredient normalization.

**2.1 Normalization Processor (`internal/normalization/processor.go`)**
- `NormalizeIngredients(tenantID, parsedIngredients[]) → RecipeIngredient[]` — main pipeline
- Pipeline steps: exact match → alias match → text normalization → unresolved
- Text normalization helpers: strip trailing 's', collapse whitespace, strip leading articles ("a ", "an ", "the ")
- Unit normalization: lookup raw unit in static registry, resolve to canonical unit + family
- Effort: **M**

**2.2 Reconciliation Logic**
- `ReconcileIngredients(tenantID, existingIngredients[], newParsedIngredients[]) → RecipeIngredient[]`
- Match by raw_name (lowercased, trimmed)
- Preserve manually_confirmed statuses, re-normalize others
- Remove ingredients no longer in source
- Update positions
- Effort: **M**

**2.3 Manual Correction**
- `ResolveIngredient(ingredientID, canonicalIngredientID, saveAsAlias) → RecipeIngredient`
- Updates status to `manually_confirmed`
- Optionally creates alias via ingredient domain
- Effort: **S**

**2.4 Re-Normalization**
- `Renormalize(recipeID) → summary` — re-runs pipeline for unresolved ingredients only
- Preserves manually_confirmed
- Returns count of status changes
- Effort: **S**

### Phase 3: Planner Configuration & Readiness (Backend)

**3.1 Planner Config Processor (`internal/planner/processor.go`)**
- Create/update planner config via upsert
- Readiness computation: check classification set + servings present
- Return plannerReady boolean + plannerIssues array
- Effort: **S**

### Phase 4: Audit Events (Backend)

**4.1 Wire Audit Emission**
- Emit events from recipe processor (create, update, delete, restore)
- Emit from normalization processor (corrected, alias_created, renormalized)
- Effort: **S**

### Phase 5: API Wiring (Backend)

Connect all domains to HTTP endpoints and modify existing recipe endpoints.

**5.1 Canonical Ingredient Endpoints (`internal/ingredient/resource.go`, `rest.go`)**
- CRUD routes: GET/POST /ingredients, GET/PATCH/DELETE /ingredients/:id
- Alias routes: POST/DELETE /ingredients/:id/aliases, /ingredients/:id/aliases/:aliasId
- Recipe usage: GET /ingredients/:id/recipes
- Reassign + delete: POST /ingredients/:id/reassign
- JSON:API request/response models
- Effort: **L**

**5.2 Normalization Endpoints (`internal/normalization/resource.go`, `rest.go`)**
- POST /recipes/:id/ingredients/:ingredientId/resolve
- POST /recipes/:id/renormalize
- JSON:API models for resolve request/response
- Effort: **M**

**5.3 Modify Recipe Endpoints**
- `resource.go` — Wire normalization into create/update handlers; add planner config to update
- `rest.go` — Extend RestDetailModel with ingredients[], plannerConfig, plannerReady, plannerIssues
- Extend RestModel (list) with plannerReady, classification, resolvedIngredients, totalIngredients
- Parse handler — add normalization status to parse response
- List handler — add filter params (plannerReady, classification, normalizationStatus)
- `provider.go` — Update list query with new filters and joins for ingredient counts
- Effort: **L**

**5.4 Service Wiring (cmd/main.go)**
- Run migrations for all 4 new domains
- Register new route groups
- Effort: **S**

### Phase 6: Frontend

**6.1 API Service & Types**
- New `ingredient.ts` service — CRUD for canonical ingredients, aliases
- New types for canonical ingredients, recipe ingredients with normalization, planner config
- Extend recipe types with ingredients[], plannerConfig, plannerReady, plannerIssues
- Extend recipe list types with plannerReady, classification, resolvedIngredients, totalIngredients
- Effort: **M**

**6.2 React Query Hooks**
- `use-ingredients.ts` — useIngredients, useIngredient, useCreateIngredient, useUpdateIngredient, useDeleteIngredient, useReassignIngredient, useAddAlias, useRemoveAlias, useIngredientRecipes
- `use-ingredient-normalization.ts` — useResolveIngredient, useRenormalize
- Extend `use-recipes.ts` — add filter params support
- Effort: **M**

**6.3 Ingredient Normalization Panel**
- `ingredient-normalization-panel.tsx` — Shows per-ingredient status with color coding, summary badge, re-normalize button
- `ingredient-resolver.tsx` — Search/select dropdown for resolving unresolved ingredients, inline create, save-as-alias checkbox
- Used on both RecipeDetailPage (read-only status + resolve actions) and RecipeFormPage (inline corrections)
- Effort: **L**

**6.4 Planner Config Components**
- `planner-config-form.tsx` — Collapsible form section with classification dropdown, numeric inputs
- `planner-ready-badge.tsx` — Status badge (green ready, yellow incomplete with tooltip)
- Effort: **S**

**6.5 Recipe Detail Page Updates**
- Add normalization panel below ingredients
- Add planner readiness badge
- Add planner config display section
- Add re-normalize button
- Effort: **M**

**6.6 Recipe Form Page Updates**
- Add planner settings collapsible section
- Add normalization review panel after live preview
- Wire normalization status into live preview (from parse endpoint)
- Effort: **M**

**6.7 Recipe List Page Updates**
- Add planner-ready badge and classification to recipe cards
- Add normalization summary ("5/8 resolved") to cards
- Add filter controls: planner readiness, classification, normalization completeness
- Effort: **M**

**6.8 Canonical Ingredient Pages**
- `IngredientsPage.tsx` — Searchable, paginated list with name, display name, unit family, alias count, usage count
- `IngredientDetailPage.tsx` — Edit form, alias management (add/remove), linked recipe list, reassign-and-delete
- Empty state guidance
- Effort: **L**

**6.9 Navigation**
- Add "Ingredients" sub-item under Recipes in sidebar nav-config
- Add routes to router config
- Effort: **S**

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Recipe create/update latency from synchronous normalization | Medium | Low | Pipeline is DB lookups only (no external calls), target <50ms for 50 ingredients. Index canonical_ingredients on (tenant_id, name) |
| Reconciliation edge cases (ingredient renames vs. new ingredients) | Medium | Medium | Match by raw_name only (simple, deterministic). Accept that renamed ingredients lose their manual confirmation |
| Large recipe list queries slow with ingredient count joins | Medium | Low | Use subquery or precomputed counts. Recipe list only needs summary counts, not full ingredient data |
| Frontend complexity of inline ingredient resolution | Low | Medium | Keep resolver as self-contained component. Use React Query mutations for immediate persistence |
| Unit registry completeness | Low | Medium | Start with comprehensive static list. Unmatched units get null canonical unit — doesn't block functionality |

## Success Metrics

- All acceptance criteria from PRD section 10 pass
- `go test ./... -count=1` passes for recipe-service
- `go build ./...` succeeds for recipe-service
- Docker build succeeds for recipe-service and frontend
- Normalization pipeline adds <50ms to recipe create/update
- All endpoints enforce tenant_id scoping (verified via tests)

## Dependencies

- Existing recipe-service (task-005) — must be on main branch (merged)
- Shared Go modules: auth, database, server, tenant
- Frontend: shadcn/ui components, TanStack React Query, react-hook-form, Zod

## Estimated Total Effort

- **Backend**: ~XL (4 new domains + modifications to recipe domain)
- **Frontend**: ~XL (2 new pages + 3 page modifications + 6+ new components)
- **Total**: ~XXL — recommend splitting implementation across multiple sessions
