# Ingredient Normalization — Task Checklist

Last Updated: 2026-03-27

## Phase 1: Database & Domain Foundations

### 1.1 Canonical Ingredient Domain [M]
- [ ] Create `internal/ingredient/model.go` — immutable model (id, tenantID, name, displayName, unitFamily, aliases, aliasCount, usageCount)
- [ ] Create `internal/ingredient/entity.go` — CanonicalIngredientEntity + CanonicalIngredientAliasEntity, Migration(), Make()
- [ ] Create `internal/ingredient/builder.go` — builder with validation (name required, unitFamily enum)
- [ ] Create `internal/ingredient/provider.go` — getByID, getByName, getByAlias, search, getUsageCount

### 1.2 Recipe Ingredient / Normalization Domain [M]
- [ ] Create `internal/normalization/model.go` — RecipeIngredient model (raw + normalized fields, status enum)
- [ ] Create `internal/normalization/entity.go` — RecipeIngredientEntity, Migration(), Make()
- [ ] Create `internal/normalization/builder.go` — builder with validation
- [ ] Create `internal/normalization/provider.go` — getByRecipeID, getByCanonicalIngredientID, bulkCreate, bulkUpdate, deleteByRecipeID
- [ ] Create `internal/normalization/unit_registry.go` — static unit map (count/weight/volume families)

### 1.3 Planner Config Domain [S]
- [ ] Create `internal/planner/model.go` — PlannerConfig model
- [ ] Create `internal/planner/entity.go` — PlannerConfigEntity, Migration(), Make()
- [ ] Create `internal/planner/builder.go` — builder with validation
- [ ] Create `internal/planner/provider.go` — getByRecipeID, upsert

### 1.4 Audit Domain [S]
- [ ] Create `internal/audit/entity.go` — AuditEventEntity, Migration()
- [ ] Create `internal/audit/emitter.go` — Emit() function

## Phase 2: Normalization Pipeline

### 2.1 Core Pipeline [M]
- [ ] Create `internal/normalization/processor.go` — NormalizeIngredients() with 4-step pipeline
- [ ] Implement exact canonical match (lowercased, trimmed)
- [ ] Implement alias match
- [ ] Implement text normalization (strip plural 's', collapse whitespace, strip articles) + re-match
- [ ] Implement unit normalization via static registry lookup

### 2.2 Reconciliation [M]
- [ ] Implement ReconcileIngredients() — match by raw_name, preserve manually_confirmed, re-normalize others
- [ ] Handle new ingredients (run full pipeline)
- [ ] Handle removed ingredients (delete records)
- [ ] Update position values to new ordering

### 2.3 Manual Correction [S]
- [ ] Implement ResolveIngredient() — update status to manually_confirmed, set canonical reference
- [ ] Implement alias learning — create alias when saveAsAlias is true
- [ ] Validate alias doesn't conflict with existing canonical name

### 2.4 Re-Normalization [S]
- [ ] Implement Renormalize() — re-run pipeline for unresolved only
- [ ] Preserve manually_confirmed statuses
- [ ] Return summary of status changes

## Phase 3: Planner Configuration & Readiness

### 3.1 Planner Processor [S]
- [ ] Create `internal/planner/processor.go` — CreateOrUpdate planner config
- [ ] Implement readiness computation (classification + servings check)
- [ ] Return plannerReady boolean + plannerIssues array

## Phase 4: Audit Events

### 4.1 Wire Audit Emission [S]
- [ ] Emit recipe.created, recipe.updated, recipe.deleted, recipe.restored from recipe processor
- [ ] Emit recipe.renormalized from normalization processor
- [ ] Emit normalization.corrected from normalization processor
- [ ] Emit ingredient.alias_created from normalization processor

## Phase 5: API Wiring

### 5.1 Canonical Ingredient Endpoints [L]
- [ ] Create `internal/ingredient/rest.go` — JSON:API request/response models
- [ ] Create `internal/ingredient/resource.go` — route registration
- [ ] Implement GET /ingredients (list with search + pagination)
- [ ] Implement POST /ingredients (create)
- [ ] Implement GET /ingredients/:id (detail with aliases)
- [ ] Implement PATCH /ingredients/:id (update)
- [ ] Implement DELETE /ingredients/:id (with reference check)
- [ ] Implement POST /ingredients/:id/aliases (add alias with conflict check)
- [ ] Implement DELETE /ingredients/:id/aliases/:aliasId (remove alias)
- [ ] Implement GET /ingredients/:id/recipes (usage list, paginated)
- [ ] Implement POST /ingredients/:id/reassign (reassign references + delete)

### 5.2 Normalization Endpoints [M]
- [ ] Create `internal/normalization/rest.go` — JSON:API models for resolve
- [ ] Create `internal/normalization/resource.go` — route registration
- [ ] Implement POST /recipes/:id/ingredients/:ingredientId/resolve
- [ ] Implement POST /recipes/:id/renormalize

### 5.3 Modify Recipe Endpoints [L]
- [ ] Extend rest.go RestDetailModel — add ingredients[], plannerConfig, plannerReady, plannerIssues
- [ ] Extend rest.go RestModel — add plannerReady, classification, resolvedIngredients, totalIngredients
- [ ] Extend rest.go CreateRequest/UpdateRequest — add plannerConfig
- [ ] Extend rest.go RestParseModel — add normalization[] per ingredient
- [ ] Modify resource.go createHandler — trigger normalization, include in response
- [ ] Modify resource.go updateHandler — trigger reconciliation if source changed, handle plannerConfig
- [ ] Modify resource.go getHandler — include normalized ingredients, planner config, readiness
- [ ] Modify resource.go listHandler — add filter params, join ingredient counts
- [ ] Modify resource.go parseHandler — add normalization status to preview
- [ ] Modify provider.go getAll — support new filter params + ingredient count subqueries
- [ ] Implement suggestion search for resolve dropdown (ILIKE prefix + usage count order)

### 5.4 Service Wiring [S]
- [ ] Update cmd/main.go — add migrations for all 4 new domains
- [ ] Update cmd/main.go — register ingredient + normalization routes
- [ ] Update nginx/ingress config — route /api/v1/ingredients to recipe-service

### 5.5 Build & Test [M]
- [ ] All tests pass: `go test ./... -count=1`
- [ ] Service builds: `go build ./...`
- [ ] Docker build succeeds

## Phase 6: Frontend

### 6.1 API Service & Types [M]
- [ ] Create `services/frontend/src/services/api/ingredient.ts` — canonical ingredient CRUD + aliases
- [ ] Add types for CanonicalIngredient, CanonicalIngredientAlias, RecipeIngredient (with normalization fields)
- [ ] Add types for PlannerConfig, PlannerReadiness
- [ ] Extend recipe types — add ingredients[], plannerConfig, plannerReady, plannerIssues to detail
- [ ] Extend recipe list types — add plannerReady, classification, resolvedIngredients, totalIngredients
- [ ] Add resolve + renormalize methods to recipe service

### 6.2 React Query Hooks [M]
- [ ] Create `use-ingredients.ts` — useIngredients, useIngredient, useCreateIngredient, useUpdateIngredient, useDeleteIngredient, useReassignIngredient
- [ ] Add alias hooks — useAddAlias, useRemoveAlias
- [ ] Add useIngredientRecipes hook
- [ ] Create `use-ingredient-normalization.ts` — useResolveIngredient, useRenormalize
- [ ] Extend `use-recipes.ts` — add plannerReady, classification, normalizationStatus filter params

### 6.3 Normalization Panel Components [L]
- [ ] Create `ingredient-normalization-panel.tsx` — per-ingredient status display, summary badge, re-normalize button
- [ ] Create `ingredient-resolver.tsx` — search/select dropdown, inline create canonical, save-as-alias checkbox
- [ ] Color-coded status indicators (green: matched/confirmed, yellow: unresolved)

### 6.4 Planner Components [S]
- [ ] Create `planner-config-form.tsx` — collapsible section, classification dropdown, numeric inputs
- [ ] Create `planner-ready-badge.tsx` — green/yellow badge with tooltip for issues

### 6.5 Recipe Detail Page Updates [M]
- [ ] Add normalization panel below ingredients list
- [ ] Add planner readiness badge in header
- [ ] Add planner config display section
- [ ] Add re-normalize button

### 6.6 Recipe Form Page Updates [M]
- [ ] Add collapsible planner settings section
- [ ] Add normalization review panel after preview
- [ ] Wire normalization status into live preview from parse endpoint
- [ ] Handle inline ingredient resolution (resolve API calls independent of form save)

### 6.7 Recipe List Page Updates [M]
- [ ] Add planner-ready badge to recipe cards
- [ ] Add classification label to recipe cards
- [ ] Add normalization summary ("5/8 resolved") to recipe cards
- [ ] Add filter controls for plannerReady, classification, normalizationStatus

### 6.8 Canonical Ingredient Pages [L]
- [ ] Create `IngredientsPage.tsx` — searchable, paginated list (name, display name, unit family, alias count, usage count)
- [ ] Create `IngredientDetailPage.tsx` — edit form (display name, unit family)
- [ ] Implement alias management on detail page (add/remove aliases)
- [ ] Implement linked recipe list on detail page (paginated, clickable)
- [ ] Implement reassign-and-delete flow (dialog with ingredient selector)
- [ ] Implement empty state guidance for empty registry

### 6.9 Navigation & Routing [S]
- [ ] Add "Ingredients" sub-item under Recipes in nav-config.ts
- [ ] Add routes for /app/ingredients and /app/ingredients/:id
- [ ] Verify responsive layout on all new/modified pages
