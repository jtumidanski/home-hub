# Ingredient Category Grouping — Task Checklist

Last Updated: 2026-03-31

---

## Phase 1: Backend — Category Domain & Migration

- [ ] **1.1** Create `internal/ingredient/category/model.go` — immutable Model with id, tenantID, name, sortOrder, ingredientCount, timestamps (S)
- [ ] **1.2** Create `internal/ingredient/category/builder.go` — fluent builder with validation (S)
- [ ] **1.3** Create `internal/ingredient/category/entity.go` — GORM Entity, Migration(), Make() converter, unique index on (tenant_id, name) (S)
- [ ] **1.4** Create `internal/ingredient/category/provider.go` — GetByID, GetByTenantID, GetByName, CountIngredientsByCategory (M)
- [ ] **1.5** Create `internal/ingredient/category/processor.go` — List (with auto-seed), Create, Update, Delete (M)
- [ ] **1.6** Create `internal/ingredient/category/rest.go` — RestModel with JSON:API mapping (S)
- [ ] **1.7** Create `internal/ingredient/category/resource.go` — GET/POST/PATCH/DELETE routes and handlers (M)
- [ ] **1.8** Modify `internal/ingredient/entity.go` — add CategoryId FK column with index (S)
- [ ] **1.9** Modify `internal/ingredient/model.go` — add categoryID, categoryName fields and accessors (S)
- [ ] **1.10** Modify `internal/ingredient/builder.go` — add WithCategoryID method (S)
- [ ] **1.11** Register category.Migration in cmd/main.go before ingredient.Migration (S)
- [ ] **1.12** Build recipe-service and verify migrations run cleanly (S)

## Phase 2: Backend — Ingredient & Export Modifications

- [ ] **2.1** Modify `internal/ingredient/provider.go` — join ingredient_categories to populate categoryName on list/detail queries (M)
- [ ] **2.2** Modify `internal/ingredient/processor.go` — accept categoryID in Create and Update, validate category exists and belongs to tenant (M)
- [ ] **2.3** Modify `internal/ingredient/rest.go` — add CategoryId, CategoryName to REST models and create/update request shapes (S)
- [ ] **2.4** Modify `internal/ingredient/resource.go` — parse categoryId from request bodies, pass to processor (S)
- [ ] **2.5** Add BulkCategorize to `internal/ingredient/processor.go` — validate + update in single transaction (M)
- [ ] **2.6** Add POST /ingredients/bulk-categorize route to `internal/ingredient/resource.go` (S)
- [ ] **2.7** Modify `internal/export/processor.go` — add CategoryName, CategorySortOrder to ConsolidatedIngredient, populate during consolidation (M)
- [ ] **2.8** Modify `internal/export/resource.go` + `rest.go` — add category fields to RestIngredientModel (S)
- [ ] **2.9** Modify `internal/export/markdown.go` — group ingredients by category with ## headers, Uncategorized at end (M)
- [ ] **2.10** Build recipe-service and run full test suite (S)

## Phase 3: Frontend — Types, API Services, Hooks

- [ ] **3.1** Add IngredientCategory type to `types/models/ingredient.ts` (S)
- [ ] **3.2** Add categoryId, categoryName to CanonicalIngredientListAttributes and DetailAttributes (S)
- [ ] **3.3** Add category_name, category_sort_order to PlanIngredientAttributes in `types/models/meal-plan.ts` (S)
- [ ] **3.4** Add category CRUD + bulkCategorize methods to `services/api/ingredient.ts` (S)
- [ ] **3.5** Create `lib/hooks/api/use-ingredient-categories.ts` — query and mutation hooks (M)
- [ ] **3.6** Update `lib/hooks/api/use-ingredients.ts` type references (S)

## Phase 4: Frontend — UI Components

- [ ] **4.1** Create category management UI — list, create, rename, delete categories (M)
- [ ] **4.2** Add category selector to IngredientDetailPage.tsx (S)
- [ ] **4.3** Show category badge on IngredientsPage.tsx ingredient cards (S)
- [ ] **4.4** Add category filter to IngredientsPage.tsx (S)
- [ ] **4.5** Show uncategorized ingredient count on IngredientsPage.tsx (S)
- [ ] **4.6** Create bulk category assignment UI — multi-select, filter, assign (L)
- [ ] **4.7** Modify ingredient-preview.tsx — group by category with section headers (M)
- [ ] **4.8** Frontend build verification (S)

## Phase 5: Integration & Verification

- [ ] **5.1** Docker build recipe-service (S)
- [ ] **5.2** Docker build frontend (S)
- [ ] **5.3** End-to-end test: create categories, assign to ingredients, verify plan ingredient grouping (M)
- [ ] **5.4** End-to-end test: markdown export with category headers (S)
- [ ] **5.5** End-to-end test: delete category, verify ingredients become uncategorized (S)
- [ ] **5.6** End-to-end test: bulk categorize (S)
- [ ] **5.7** Verify backward compatibility — plans with no categorized ingredients still work (S)
