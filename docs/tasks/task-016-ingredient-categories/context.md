# Ingredient Category Grouping — Context

Last Updated: 2026-03-31

---

## Key Files — Backend (recipe-service)

### New Domain
- `services/recipe-service/internal/ingredient/category/` — New directory for category domain (model, builder, entity, processor, provider, resource, rest)

### Ingredient Domain (modify)
- `services/recipe-service/internal/ingredient/model.go` — Add categoryID, categoryName fields
- `services/recipe-service/internal/ingredient/builder.go` — Add WithCategoryID method
- `services/recipe-service/internal/ingredient/entity.go` — Add CategoryId FK column, migration
- `services/recipe-service/internal/ingredient/processor.go` — Accept categoryID in Create/Update, add BulkCategorize
- `services/recipe-service/internal/ingredient/provider.go` — Join category table in queries
- `services/recipe-service/internal/ingredient/resource.go` — Add bulk-categorize route, parse categoryId
- `services/recipe-service/internal/ingredient/rest.go` — Add CategoryId/CategoryName to REST models

### Export Domain (modify)
- `services/recipe-service/internal/export/processor.go` — Add category fields to ConsolidatedIngredient, carry through pipeline
- `services/recipe-service/internal/export/markdown.go` — Group by category headers
- `services/recipe-service/internal/export/resource.go` — Add category fields to RestIngredientModel

### Entry Point
- `services/recipe-service/cmd/main.go` — Register category.Migration before ingredient.Migration

---

## Key Files — Frontend

### Types (modify)
- `frontend/src/types/models/ingredient.ts` — Add IngredientCategory type, category fields to existing types
- `frontend/src/types/models/meal-plan.ts` — Add category_name, category_sort_order to PlanIngredientAttributes

### API Services (modify/new)
- `frontend/src/services/api/ingredient.ts` — Add category CRUD methods, bulkCategorize method

### Hooks (modify/new)
- `frontend/src/lib/hooks/api/use-ingredient-categories.ts` — New file: category query/mutation hooks
- `frontend/src/lib/hooks/api/use-ingredients.ts` — Update types for category fields

### Pages (modify)
- `frontend/src/pages/IngredientsPage.tsx` — Show category badge, filter by category, uncategorized count
- `frontend/src/pages/IngredientDetailPage.tsx` — Add category selector dropdown

### Components (modify/new)
- `frontend/src/components/features/meals/ingredient-preview.tsx` — Group by category with section headers
- `frontend/src/components/features/ingredients/` — New: category management UI, bulk-edit UI

### Routing
- `frontend/src/App.tsx` — May need new route for category management (or inline in ingredients page)

---

## Key Decisions

1. **Category domain as sub-package of ingredient**: `internal/ingredient/category/` rather than top-level `internal/category/` — categories are tightly coupled to ingredients.

2. **Auto-seed on first GET**: Default categories are seeded when `List` returns empty for a tenant, not during tenant creation (recipe-service doesn't know about tenant creation).

3. **No custom sort order editing**: PRD explicitly marks this as a non-goal. Sort order is fixed at creation time (defaults follow aisle order).

4. **Category on canonical ingredient only**: Not on recipe ingredients or normalization records. Category resolution happens at consolidation time by joining through the canonical ingredient.

5. **SET NULL on category delete**: Standard FK cascade. Deleting a category makes affected ingredients "Uncategorized" — no user confirmation needed beyond showing the count.

6. **Export modal needs no frontend changes**: Backend markdown output already flows through to the rendered preview. Category headers in markdown are automatically displayed.

---

## Dependencies

- No cross-service dependencies — all changes are within recipe-service and frontend
- No shared module changes needed
- No new external dependencies or packages
- GORM AutoMigrate handles schema changes

---

## Existing Patterns to Follow

- **Domain structure**: model.go → entity.go → builder.go → processor.go → provider.go → resource.go → rest.go
- **Immutable models**: Private fields with public accessors
- **Fluent builders**: `NewBuilder().WithName(...).WithTenantID(...).Build()`
- **Functional providers**: `GetByID(id) func(*gorm.DB) func() (Entity, error)`
- **JSON:API responses**: `server.MarshalResponse` / `server.MarshalSliceResponse`
- **Tenant scoping**: `tenantctx.MustFromContext(r.Context())` in every handler
- **Migration registration**: `database.SetMigrations(...)` in cmd/main.go
- **React Query keys**: Factory pattern with `all > lists > details > detail(id)`
- **API service pattern**: Extends `BaseService`, methods accept `Tenant` first param
