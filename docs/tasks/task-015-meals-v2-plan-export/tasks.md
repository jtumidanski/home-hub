# Meals v2 — Task Checklist

Last Updated: 2026-03-29

---

## Phase 1: Backend Domain Models [Effort: L]

### 1.1 Plan Week Domain

- [x] **1.1.1** Create `internal/plan/model.go` — Immutable PlanWeek model (id, tenantID, householdID, startsOn, name, locked, createdBy, createdAt, updatedAt) with getters
- [x] **1.1.2** Create `internal/plan/entity.go` — GORM entity for `plan_weeks` table with unique index (tenant_id, household_id, starts_on), Make() and ToEntity() converters, Migration()
- [x] **1.1.3** Create `internal/plan/builder.go` — Builder with validation (startsOn required, name auto-generated if empty)
- [x] **1.1.4** Create `internal/plan/provider.go` — Query functions: getByID, getAll with ListFilters (starts_on filter, pagination), getByHouseholdAndStartsOn (uniqueness check)
- [x] **1.1.5** Create `internal/plan/administrator.go` — CRUD: create, update (name only), setLocked, delete (if needed)
- [x] **1.1.6** Create `internal/plan/processor.go` — Business logic: Create, Get, List, UpdateName, Lock, Unlock, Duplicate; audit emission; locked-state guard

### 1.2 Plan Item Domain

- [x] **1.2.1** Create `internal/planitem/model.go` — Immutable PlanItem model (id, planWeekID, day, slot, recipeID, servingMultiplier as `*float64`, plannedServings, notes, position, createdAt, updatedAt) with getters; Slot enum constants
- [x] **1.2.2** Create `internal/planitem/entity.go` — GORM entity for `plan_items` table with indexes (plan_week_id, recipe_id), FK constraint (no cascade on recipe_id, cascade on plan_week_id), Make() and ToEntity(), Migration()
- [x] **1.2.3** Create `internal/planitem/builder.go` — Builder with validation (day within week range, valid slot enum, recipe_id required)
- [x] **1.2.4** Create `internal/planitem/provider.go` — Query functions: getByPlanWeekID, getByID, getRecipeUsage (last_used_date + usage_count per recipe for a household)
- [x] **1.2.5** Create `internal/planitem/administrator.go` — CRUD: create (auto-assign position as MAX+1 for day+slot when not provided), update, delete
- [x] **1.2.6** Create `internal/planitem/processor.go` — Business logic: AddItem, UpdateItem, RemoveItem; validate recipe exists and is active; audit emission

### 1.3 Registration

- [x] **1.3.1** Register plan and planitem migrations in `cmd/main.go` via `database.SetMigrations()`
- [x] **1.3.2** Verify GORM auto-migration creates tables correctly (build + test)

**Acceptance:** Both tables migrate successfully. Plan and item CRUD works at processor level (no REST yet).

---

## Phase 2: Backend REST API [Effort: L]

### 2.1 Plan REST Endpoints

- [x] **2.1.1** Create `internal/plan/resource.go` — JSON:API resource types: RestListModel (with item_count), RestDetailModel (with items array), CreateRequest, UpdateRequest, DuplicateRequest; GetName(), GetID(), SetID(); transform functions
- [x] **2.1.2** Create `internal/plan/rest.go` — InitializeRoutes under `/meals/plans`; handlers for: POST create, GET list, GET detail, PATCH update, POST lock, POST unlock, POST duplicate
- [x] **2.1.3** Plan detail handler — Enrich items with recipe metadata (title, servings, classification, recipe_deleted) by reading from recipe and planner domains
- [x] **2.1.4** Plan list handler — Include item_count per plan; support starts_on filter and pagination

### 2.2 Plan Item REST Endpoints

- [x] **2.2.1** Create `internal/planitem/resource.go` — JSON:API resource types: RestModel, CreateRequest, UpdateRequest; transform functions
- [x] **2.2.2** Create `internal/planitem/rest.go` — Handlers under `/meals/plans/{planId}/items`: POST add, PATCH update, DELETE remove
- [x] **2.2.3** Wire item handlers into plan route initialization (or separate InitializeRoutes)

### 2.3 Recipe Usage Enhancement

- [x] **2.3.1** Add `include_usage=true` query param support to existing recipe list endpoint
- [x] **2.3.2** When include_usage=true, join plan_items to compute last_used_date and usage_count per recipe
- [x] **2.3.3** Add usage fields to recipe RestModel and transform function

### 2.4 Audit Events

- [x] **2.4.1** Emit audit events in plan processor: plan.created, plan.updated, plan.locked, plan.unlocked, plan.duplicated
- [x] **2.4.2** Emit audit events in planitem processor: plan.item_added, plan.item_updated, plan.item_removed
- [x] **2.4.3** Include relevant metadata (recipe_id for item events, source_plan_id for duplicate)

### 2.5 Build Verification

- [x] **2.5.1** Verify recipe-service builds successfully with all new packages
- [x] **2.5.2** Verify Docker build passes

**Acceptance:** All plan and item endpoints return correct JSON:API responses. Audit events fire. Recipe list supports include_usage. Build passes.

---

## Phase 3: Backend Export [Effort: M]

### 3.1 Ingredient Consolidation

- [x] **3.1.1** Create `internal/export/processor.go` — ExportProcessor with dependencies on plan, planitem, normalization, ingredient, recipe, planner processors
- [x] **3.1.2** Implement `ConsolidateIngredients()` — For each plan item: read recipe_ingredients, apply serving multiplier, group by (canonical_ingredient_id, canonical_unit), sum quantities; list unresolved individually
- [x] **3.1.3** Implement serving scaling logic: planned_servings > serving_multiplier > default 1.0; effective_multiplier = planned_servings / servings_yield when applicable
- [x] **3.1.4** Handle edge cases: missing recipe_ingredients (fall back to raw Cooklang via `recipe.Processor.ParseSource()`), deleted recipes (use last-known normalized data), zero-quantity omission

### 3.2 Markdown Generation

- [x] **3.2.1** Create `internal/export/markdown.go` — Generate markdown from plan + consolidated ingredients
- [x] **3.2.2** Format: heading with plan name, day sections (calendar order, only days with items), slot ordering (breakfast, lunch, dinner, snack, side), servings override notation, consolidated ingredient list, notes section for unresolved

### 3.3 Export REST Endpoints

- [x] **3.3.1** Add `GET /meals/plans/{planId}/export/markdown` handler — returns text/markdown content type, raw markdown body
- [x] **3.3.2** Add `GET /meals/plans/{planId}/ingredients` handler — returns JSON:API collection of consolidated ingredients
- [x] **3.3.3** Emit plan.exported audit event on markdown export
- [x] **3.3.4** Wire export routes into plan route initialization

### 3.4 Build Verification

- [x] **3.4.1** Verify recipe-service builds successfully with export package
- [x] **3.4.2** Verify Docker build passes

**Acceptance:** Markdown export produces correctly formatted output. Ingredient consolidation merges matching canonical ingredients, lists unresolved separately. Serving scaling applied correctly.

---

## Phase 4: Infrastructure [Effort: S]

- [x] **4.1** Add nginx location block for `/api/v1/meals` → `http://recipe-service:8080` in `deploy/compose/nginx.conf`
- [x] **4.2** Update `docs/architecture.md` routing table with new meals route
- [x] **4.3** Verify nginx routes requests correctly to recipe-service

**Acceptance:** API requests to `/api/v1/meals/plans` route through to recipe-service.

---

## Phase 5: Frontend [Effort: XL]

### 5.1 API Layer

- [x] **5.1.1** Create `services/api/meals.ts` — MealsService extending BaseService: listPlans, getPlan, createPlan, updatePlan, lockPlan, unlockPlan, duplicatePlan, addItem, updateItem, removeItem, exportMarkdown, getIngredients
- [x] **5.1.2** Create `lib/hooks/api/use-meals.ts` — TanStack Query hooks: usePlans, usePlan, useCreatePlan, useUpdatePlan, useLockPlan, useUnlockPlan, useDuplicatePlan, useAddPlanItem, useUpdatePlanItem, useRemovePlanItem, useExportMarkdown, usePlanIngredients; cache key factory
- [x] **5.1.3** Create `lib/schemas/meals.schema.ts` — Zod schemas: planFormSchema, planItemFormSchema
- [x] **5.1.4** Add TypeScript types for Plan, PlanItem, PlanIngredient models in `types/models/`

### 5.2 Planner Page Shell

- [x] **5.2.1** Create `pages/MealsPage.tsx` — Main planner page with layout: header with week selector + actions, grid area, recipe selector + ingredient preview below
- [x] **5.2.2** Add route `/app/meals` in `App.tsx`
- [x] **5.2.3** Add "Meal Planner" entry to navigation config in `nav-config.ts` under Lifestyle group

### 5.3 Week Navigation

- [x] **5.3.1** Week selector component with previous/next arrows (step by 7 days)
- [x] **5.3.2** Auto-load existing plan for selected week or show empty grid
- [x] **5.3.3** Configurable week start day (default Monday, stored in frontend preference)

### 5.4 Weekly Grid

- [x] **5.4.1** Create `components/features/meals/week-grid.tsx` — 7-day rows x slot columns grid displaying plan items
- [x] **5.4.2** Each cell shows recipe title, classification badge, servings info
- [x] **5.4.3** Click empty cell to open recipe selector with auto-classification filter for that slot
- [x] **5.4.4** Click existing item to open edit popover (change day, slot, servings, notes)
- [x] **5.4.5** Remove button (x) on each item with confirmation for saved plans
- [x] **5.4.6** Read-only mode when plan is locked (greyed out, no interactions)
- [x] **5.4.7** Deleted recipe warning indicator on affected items

### 5.5 Recipe Selector

- [x] **5.5.1** Create `components/features/meals/recipe-selector.tsx` — Recipe search panel reusing existing recipe list API
- [x] **5.5.2** Search by name (debounced input)
- [x] **5.5.3** Classification filter dropdown (auto-applied from slot click, clearable)
- [x] **5.5.4** Display: recipe name, classification, servings_yield, last_used_date, usage_count
- [x] **5.5.5** Click recipe to open add-item popover (day, slot, servings, notes)

### 5.6 Plan Item Popover

- [x] **5.6.1** Create `components/features/meals/plan-item-popover.tsx` — Popover for adding/editing plan items
- [x] **5.6.2** Fields: day picker (constrained to plan week), slot selector, serving multiplier OR planned servings toggle, notes textarea
- [x] **5.6.3** Validation with Zod schema
- [x] **5.6.4** Save triggers API call (create or update)

### 5.7 Ingredient Preview

- [x] **5.7.1** Create `components/features/meals/ingredient-preview.tsx` — Panel showing consolidated ingredient list
- [x] **5.7.2** Auto-refresh when plan items change
- [x] **5.7.3** Show unresolved ingredient count with warning
- [x] **5.7.4** Display format: quantity + unit + name

### 5.8 Plan Actions

- [x] **5.8.1** Save button — creates plan via `POST /plans` if new (eager creation), then persists items via separate `POST /plans/{id}/items` calls
- [x] **5.8.2** Lock/unlock toggle button with lock icon
- [x] **5.8.3** Duplicate button — opens date picker for target week, calls duplicate API
- [x] **5.8.4** Export button — triggers export and opens modal

### 5.9 Export Modal

- [x] **5.9.1** Create `components/features/meals/export-modal.tsx` — Modal with rendered markdown preview
- [x] **5.9.2** Copy to clipboard button
- [x] **5.9.3** Download as .md file button
- [x] **5.9.4** Raw markdown text display option

### 5.10 Build Verification

- [x] **5.10.1** Verify frontend builds successfully
- [x] **5.10.2** Verify all new components render without errors

**Acceptance:** Full planner workflow works end-to-end: create plan, add recipes to slots, adjust servings, preview ingredients, lock plan, export markdown.

---

## Phase 6: Integration & Polish [Effort: M]

- [x] **6.1** End-to-end test: create plan, add items, lock, export markdown — verify output format
- [x] **6.2** Test duplicate plan to another week — verify day offsets correct
- [x] **6.3** Test deleted recipe handling — verify warning indicator and consolidation includes last-known data
- [x] **6.4** Test serving scaling — verify planned_servings precedence over serving_multiplier
- [x] **6.5** Test locked plan rejection — verify 409 on all modification attempts
- [x] **6.6** Test recipe usage metadata — verify last_used_date and usage_count in recipe list
- [x] **6.7** Verify all audit events emit correctly
- [x] **6.8** Verify role-based access (viewer can read, cannot write)
- [x] **6.9** Full Docker build verification (recipe-service + frontend)
