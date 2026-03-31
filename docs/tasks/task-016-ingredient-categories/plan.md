# Ingredient Category Grouping — Implementation Plan

Last Updated: 2026-03-31

---

## Executive Summary

Add product-category grouping to canonical ingredients so the meal planner's consolidated ingredient list renders grouped by grocery-aisle section. The feature spans a new `ingredient/category` backend domain, modifications to the existing `ingredient` and `export` domains, a database migration, and frontend changes across the ingredients pages, meal plan ingredient preview, and markdown export.

---

## Current State Analysis

### Backend (recipe-service)

- **Ingredient domain** (`internal/ingredient/`): Full CRUD for canonical ingredients with aliases, unit families, reassignment, and usage tracking. No category concept.
- **Normalization domain** (`internal/normalization/`): Maps raw recipe ingredient names to canonical ingredients. Tracks resolution status. Does not carry category info.
- **Export domain** (`internal/export/`): `ConsolidateIngredients` pipeline accumulates quantities by canonical ingredient. `GenerateMarkdown` renders a flat ingredient list. Neither groups by category.
- **Plan domain** (`internal/plan/`): `GET /meals/plans/{planId}/ingredients` returns `RestIngredientModel` without category fields.
- **Migrations**: GORM AutoMigrate in each domain's `entity.go`, registered in `cmd/main.go`.

### Frontend

- **IngredientsPage**: Card list with search, create, delete. No category column.
- **IngredientDetailPage**: Shows name, display name, unit family, aliases, recipes. No category selector.
- **ingredient-preview.tsx**: Flat list of consolidated ingredients. No grouping.
- **export-modal.tsx**: Renders markdown from backend. Backend currently produces flat list.
- **Types**: `PlanIngredientAttributes` has no category fields. `CanonicalIngredientListAttributes` has no category field.
- **API services**: `IngredientService` has no category methods. `MealsService.getIngredients` returns flat list.

---

## Proposed Future State

1. New `ingredient/category` domain in recipe-service with full CRUD and auto-seeding
2. `canonical_ingredients` table gains nullable `category_id` FK
3. Ingredient CRUD endpoints include category info in responses and accept category assignment
4. New `POST /ingredients/bulk-categorize` endpoint
5. `ConsolidateIngredients` output carries category name and sort order
6. `GET /meals/plans/{planId}/ingredients` response includes category fields
7. Markdown export groups ingredients under category headers
8. Frontend ingredient preview groups by category with section headers
9. Frontend ingredients page shows category, detail page has category selector
10. New category management UI and bulk-edit UI

---

## Implementation Phases

### Phase 1: Backend — Category Domain & Migration

Create the new `ingredient/category` domain and modify the ingredient entity.

#### 1.1 Category Domain — Model, Builder, Entity (Effort: M)

Create `internal/ingredient/category/` with:

- **model.go**: Immutable `Model` with fields: `id`, `tenantID`, `name`, `sortOrder`, `ingredientCount`, `createdAt`, `updatedAt`. Accessors for all fields.
- **builder.go**: Fluent builder. Validates: name required (max 100 chars), sortOrder >= 0.
- **entity.go**: GORM `Entity` struct mapping to `ingredient_categories` table. Unique index on `(TenantId, Name)`. `Migration()` function. `Make(Entity) Model` converter.

Acceptance criteria:
- Model is immutable with getter methods
- Builder rejects empty name and negative sort order
- GORM AutoMigrate creates `ingredient_categories` table with correct schema
- Unique constraint on (tenant_id, name)

#### 1.2 Category Domain — Provider & Processor (Effort: M)

- **provider.go**: Functional query builders — `GetByID`, `GetByTenantID` (list all for tenant), `GetByName` (tenant + name lookup), `CountIngredientsByCategory` (join query for ingredient_count).
- **processor.go**: Business logic — `List` (with auto-seed on empty), `Create`, `Update`, `Delete`. Auto-seed logic: if `List` returns empty for tenant, insert default categories in a transaction, then return them.

Default categories and sort order:
1. Produce
2. Meats & Seafood
3. Dairy & Eggs
4. Bakery & Bread
5. Pantry & Dry Goods
6. Frozen
7. Beverages
8. Snacks & Sweets
9. Condiments & Sauces
10. Spices & Seasonings
11. Other

Acceptance criteria:
- First `List` call for a tenant auto-seeds defaults and returns them
- Subsequent `List` calls return existing categories (no re-seed)
- `Delete` works even when ingredients reference the category (FK handles SET NULL)
- `Create` auto-assigns next sort_order value

#### 1.3 Category Domain — REST Resource (Effort: M)

- **rest.go**: `RestModel` with id, name, sortOrder, ingredientCount, timestamps.
- **resource.go**: Register routes and handlers:
  - `GET /ingredient-categories` — list
  - `POST /ingredient-categories` — create
  - `PATCH /ingredient-categories/{id}` — update
  - `DELETE /ingredient-categories/{id}` — delete

Acceptance criteria:
- All endpoints return JSON:API format
- Tenant scoping enforced on all operations
- 404 on missing category, 409 on duplicate name

#### 1.4 Ingredient Entity Migration (Effort: S)

Modify `internal/ingredient/entity.go`:
- Add `CategoryId *uuid.UUID` field to `Entity` with `gorm:"index"` tag
- GORM AutoMigrate will add the column
- Add foreign key constraint to `ingredient_categories` with ON DELETE SET NULL

Modify `internal/ingredient/model.go`:
- Add `categoryID *uuid.UUID` and `categoryName string` fields with accessors

Modify `internal/ingredient/builder.go`:
- Add `WithCategoryID(*uuid.UUID)` method

Register category migration in `cmd/main.go` (must run before ingredient migration due to FK).

Acceptance criteria:
- `canonical_ingredients.category_id` column exists, nullable
- FK constraint with ON DELETE SET NULL
- Index on `category_id`
- Existing ingredients unaffected (null category)

---

### Phase 2: Backend — Ingredient & Export Modifications

#### 2.1 Ingredient CRUD — Category Support (Effort: M)

Modify `internal/ingredient/`:

- **provider.go**: Join `ingredient_categories` to populate `categoryName` on queries. Add `GetByCategoryID` for listing ingredients in a category.
- **processor.go**: Accept optional `categoryID` in `Create` and `Update`. Validate that category exists and belongs to same tenant before assigning.
- **rest.go**: Add `CategoryId` and `CategoryName` to `RestModel` and `RestDetailModel`. Add `CategoryId` to create/update request payloads.
- **resource.go**: Parse `categoryId` from create/update requests. Pass to processor.

Acceptance criteria:
- `POST /ingredients` accepts optional `categoryId`, validates it, persists it
- `PATCH /ingredients/{id}` accepts `categoryId` (including null to clear)
- `GET /ingredients` response includes `category_id` and `category_name`
- `GET /ingredients/{id}` response includes `category_id` and `category_name`
- Invalid/wrong-tenant category ID returns 422

#### 2.2 Bulk Categorize Endpoint (Effort: M)

Add to `internal/ingredient/`:

- **processor.go**: `BulkCategorize(tenantID, ingredientIDs []uuid.UUID, categoryID uuid.UUID)` — validate category, update all ingredients in a single transaction.
- **resource.go**: `POST /ingredients/bulk-categorize` — parse JSON:API request, call processor, return 204.

Acceptance criteria:
- Updates up to 200 ingredients in one request
- Single transaction — all succeed or all fail
- Validates category belongs to tenant
- Returns 204 on success
- Returns 404 if category not found, 422 if any ingredient ID is invalid

#### 2.3 Export Pipeline — Category Grouping (Effort: M)

Modify `internal/export/`:

- **processor.go**: `ConsolidatedIngredient` struct gains `CategoryName string` and `CategorySortOrder int` fields. During consolidation, look up category from canonical ingredient. Unresolved/uncategorized get empty category name and max sort order.
- **resource.go** / **rest.go**: `RestIngredientModel` gains `CategoryName *string` and `CategorySortOrder *int` fields.
- **markdown.go**: Group ingredients by category. Render `## CategoryName` headers. Sort groups by `CategorySortOrder`. Uncategorized group at end under `## Uncategorized`. Omit empty categories.

Acceptance criteria:
- `GET /meals/plans/{planId}/ingredients` returns category_name and category_sort_order per ingredient
- Unresolved ingredients have null category_name and null sort_order
- Markdown export has category section headers
- Categories in markdown sorted by sort_order
- Uncategorized section at end
- Empty categories omitted

---

### Phase 3: Frontend — Types, API Services, Hooks

#### 3.1 Category Types & API Service (Effort: S)

- **types/models/ingredient.ts**: Add `IngredientCategory` type. Add `categoryId` and `categoryName` to `CanonicalIngredientListAttributes` and `CanonicalIngredientDetailAttributes`. Add `category_name` and `category_sort_order` to `PlanIngredientAttributes`.
- **services/api/ingredient.ts**: Add methods: `listCategories`, `createCategory`, `updateCategory`, `deleteCategory`, `bulkCategorize`.

Acceptance criteria:
- All new types match JSON:API response shapes
- API service methods call correct endpoints with correct payloads

#### 3.2 React Query Hooks (Effort: S)

- **lib/hooks/api/use-ingredient-categories.ts** (new): Query hooks for `useCategories`, mutation hooks for `useCreateCategory`, `useUpdateCategory`, `useDeleteCategory`, `useBulkCategorize`. Query key factory: `categoryKeys`.
- **lib/hooks/api/use-ingredients.ts**: Update type references for new category fields.

Acceptance criteria:
- Hooks invalidate appropriate query keys on mutations
- `useBulkCategorize` invalidates both ingredient lists and category lists (ingredient_count changes)

---

### Phase 4: Frontend — UI Components

#### 4.1 Category Management UI (Effort: M)

New section accessible from the ingredients area (e.g., tab or settings gear on IngredientsPage):

- List categories with name, sort order, ingredient count
- Create new category (name input)
- Rename category (inline edit or modal)
- Delete category (with confirmation showing affected ingredient count)
- Reorder not in scope (non-goal per PRD)

Acceptance criteria:
- CRUD operations work with optimistic updates
- Delete confirmation shows "X ingredients will become uncategorized"
- Categories display in sort_order sequence

#### 4.2 Ingredient Detail — Category Selector (Effort: S)

Modify `IngredientDetailPage.tsx`:

- Add category selector (dropdown/combobox) populated from `useCategories()`
- On change, call `useUpdateIngredient()` with new `categoryId`
- Show "Uncategorized" when no category set

Acceptance criteria:
- Category selector shows all tenant categories
- Changing category persists immediately
- "None" option clears category

#### 4.3 Ingredients Page — Category Display (Effort: S)

Modify `IngredientsPage.tsx`:

- Show category name as a badge/subtitle on each ingredient card
- Add category filter dropdown (filter client-side or add query param)
- Show uncategorized count badge to encourage categorization

Acceptance criteria:
- Category visible on ingredient cards
- Filter by category works
- Uncategorized count displayed prominently

#### 4.4 Bulk Category Assignment UI (Effort: L)

New UI (modal or dedicated view) for bulk category assignment:

- Table/list of ingredients with checkboxes
- Filter: uncategorized only, by category, search by name
- Select target category from dropdown
- Apply button assigns category to all selected ingredients
- Show count of selected ingredients
- Show count of uncategorized ingredients as prompt

Acceptance criteria:
- Can select multiple ingredients and assign category in one action
- Filters work correctly (uncategorized, by category, search)
- Uses `useBulkCategorize` mutation
- Updates UI optimistically or refetches on success
- Handles up to 200 ingredients

#### 4.5 Ingredient Preview — Category Grouping (Effort: M)

Modify `ingredient-preview.tsx`:

- Group ingredients by `category_name`
- Render category section headers
- Sort groups by `category_sort_order`
- Uncategorized/null group at end
- Within each group, ingredients sorted alphabetically
- Empty categories not shown

Acceptance criteria:
- Ingredients visually grouped under category headers
- Correct sort order (aisle order)
- Uncategorized at bottom
- Alphabetical within groups

#### 4.6 Export Modal — No Changes Needed (Effort: —)

The export modal renders markdown from the backend. Since Phase 2.3 modifies the backend markdown output to include category headers, the export modal will automatically display grouped ingredients. The raw markdown view and rendered preview will both reflect the grouping.

No frontend changes needed here.

---

## Dependency Graph

```
Phase 1.1 (Category Domain)
  ↓
Phase 1.4 (Ingredient Entity Migration) ← depends on 1.1 for FK target
  ↓
Phase 1.2 (Category Provider/Processor) ← depends on 1.1
  ↓
Phase 1.3 (Category REST) ← depends on 1.2
  ↓
Phase 2.1 (Ingredient CRUD Category) ← depends on 1.4
Phase 2.2 (Bulk Categorize) ← depends on 1.4
Phase 2.3 (Export Pipeline) ← depends on 2.1
  ↓
Phase 3.1 (Frontend Types/API) ← depends on 2.x (needs API contracts)
Phase 3.2 (Frontend Hooks) ← depends on 3.1
  ↓
Phase 4.1–4.5 (Frontend UI) ← depends on 3.2
```

Phases 1.1–1.4 are sequential. Phases 2.1, 2.2 can be parallel. Phase 2.3 depends on 2.1. Phase 3 depends on Phase 2. Phase 4 tasks are mostly independent of each other but all depend on Phase 3.

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Auto-seed race condition (two concurrent requests for new tenant) | Duplicate categories | Low | Use unique constraint; catch constraint violation in seed logic and re-query |
| GORM AutoMigrate FK ordering | Migration failure | Medium | Register category migration before ingredient migration in main.go |
| Large bulk-categorize request timeout | 500 error | Low | Single UPDATE WHERE id IN (...) query; 200 IDs is well within DB limits |
| Category deletion orphans UI state | Stale UI | Medium | Invalidate ingredient queries on category delete; FK SET NULL handles DB |
| Markdown export regression | Broken shopping lists | Medium | Unit test markdown output with categorized and uncategorized ingredients |

---

## Success Metrics

- All 13 acceptance criteria from the PRD pass
- Consolidated ingredient list renders grouped in < 200ms additional latency
- Bulk-categorize handles 200 ingredients without timeout
- Markdown export includes category headers in correct order
- Zero regression in existing ingredient/export functionality

---

## Required Resources

- recipe-service Go codebase
- Frontend React codebase
- PostgreSQL (GORM AutoMigrate)
- Docker for build verification
