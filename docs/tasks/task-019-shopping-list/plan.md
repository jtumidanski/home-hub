# Task 019: Shopping List — Implementation Plan

Last Updated: 2026-03-31

---

## Executive Summary

Implement a Shopping List feature spanning two new backend services (category-service, shopping-service), modifications to recipe-service, and a full frontend feature. The category-service extracts the existing ingredient category domain into a shared service. The shopping-service provides list management, item CRUD, shopping mode, archiving, and meal plan import. The frontend adds list management, shopping mode, history views, and a meal plan import action.

---

## Current State Analysis

### What Exists
- **recipe-service** owns `ingredient/category/` domain with full CRUD, auto-seeding of 11 default categories, and tenant scoping
- **recipe-service** has a consolidated ingredients endpoint at `GET /meals/plans/{planId}/ingredients` used by the export/meal-plan feature
- **Frontend** has category management UI under `components/features/ingredients/category-manager.tsx` and hooks in `lib/hooks/api/use-ingredient-categories.ts`
- **Frontend** calls categories at `/api/v1/ingredient-categories`
- **nginx** routes `/api/v1/ingredient-categories` to recipe-service
- All services follow the DDD pattern: `entity.go`, `model.go`, `builder.go`, `processor.go`, `provider.go`, `resource.go`, `rest.go`

### What Needs to Change
- Extract category domain from recipe-service into standalone category-service
- Recipe-service must drop its category table and consume category-service via HTTP
- Two new services added to docker-compose, nginx, go.work, CI
- Frontend category API calls re-pointed to new base URL `/api/v1/categories`
- Full shopping list UI built from scratch

---

## Proposed Future State

```
Browser -> nginx -> /api/v1/categories  -> category-service (new)
                    /api/v1/shopping    -> shopping-service (new)
                    /api/v1/meals       -> recipe-service (modified)
                    /api/v1/ingredients -> recipe-service (modified)
```

- **category-service**: Owns `category` schema, serves CRUD for categories, seeds 15 defaults (11 food + 4 non-food)
- **shopping-service**: Owns `shopping` schema, provides list lifecycle, item CRUD, shopping mode, archiving, meal plan import
- **recipe-service**: No longer owns categories; consumes category-service via HTTP for enrichment
- **Frontend**: New shopping list feature pages, updated category API calls

---

## Implementation Phases

### Phase 1: Category Service (Backend)

Build the new category-service as a standalone microservice, closely mirroring the existing recipe-service category domain.

**1.1 Scaffold category-service** [M]
- Create `services/category-service/` directory structure following existing service pattern
- `cmd/main.go` entry point
- `Dockerfile` (copy from existing service, adjust paths)
- `go.mod` with dependencies on shared modules
- `internal/config/` for environment configuration
- Add to `go.work`
- **Acceptance:** Service compiles, starts, and passes health check

**1.2 Implement category domain** [L]
- `internal/category/entity.go` — GORM entity with `(tenant_id, name)` unique index
- `internal/category/model.go` — Immutable domain model with accessors
- `internal/category/builder.go` — Builder with validation
- `internal/category/processor.go` — List, Create, Update, Delete + seedDefaults (15 categories)
- `internal/category/provider.go` — Data access functions
- `internal/category/resource.go` — HTTP handlers for GET, POST, PATCH, DELETE
- `internal/category/rest.go` — JSON:API resource type `"categories"`
- **Acceptance:** All CRUD operations work, auto-seeds 15 defaults on first access per tenant

**1.3 Add category-service to infrastructure** [S]
- Add to `deploy/compose/docker-compose.yml`
- Add nginx route: `/api/v1/categories` -> category-service
- **Acceptance:** Category-service reachable at `/api/v1/categories` through nginx

**1.4 Data migration strategy** [M]
- Category-service seeds defaults on first access (same as recipe-service does today)
- Recipe-service's existing `ingredient_categories` data maps 1:1 to category-service
- Migration approach: category-service auto-seeds, recipe-service updated to consume it
- No manual data migration SQL needed since categories are tenant-seeded on demand
- **Acceptance:** Categories accessible via new service for any tenant that accesses the endpoint

**1.5 Unit tests for category domain** [M]
- Builder tests for validation edge cases
- Processor tests for CRUD logic and seeding
- **Acceptance:** Tests pass, cover happy path and error cases

### Phase 2: Recipe Service Migration

Modify recipe-service to consume category-service instead of owning categories.

**2.1 Add HTTP client for category-service** [M]
- Create `internal/categoryclient/` package in recipe-service
- Client fetches category list from category-service, forwarding JWT
- Used by export processor and ingredient enrichment
- **Acceptance:** Recipe-service can fetch categories from category-service

**2.2 Remove category domain from recipe-service** [M]
- Delete `internal/ingredient/category/` package entirely
- Remove category route registration from recipe-service
- Update `canonical_ingredients.category_id` — keep as opaque UUID, remove FK constraint
- Remove category migration from recipe-service startup
- Update export processor to use category client instead of local provider
- **Acceptance:** Recipe-service compiles without category domain, ingredient CRUD still works, consolidated ingredients endpoint still returns category data

**2.3 Update recipe-service tests** [M]
- Update any tests that reference the category domain
- Ensure export/consolidation tests work with mocked category client
- **Acceptance:** All recipe-service tests pass

### Phase 3: Shopping Service (Backend)

Build the shopping-service as a new microservice.

**3.1 Scaffold shopping-service** [M]
- Create `services/shopping-service/` directory structure
- `cmd/main.go`, `Dockerfile`, `go.mod`
- `internal/config/` for environment configuration
- Add to `go.work`
- **Acceptance:** Service compiles, starts, passes health check

**3.2 Implement shopping list domain** [L]
- `internal/list/entity.go` — GORM entity for `shopping_lists`
- `internal/list/model.go` — Domain model with status (active/archived)
- `internal/list/builder.go` — Builder with validation
- `internal/list/processor.go` — Create, Rename, Delete, Archive, Unarchive, List (by status)
- `internal/list/provider.go` — Data access with tenant + household scoping
- `internal/list/resource.go` — HTTP handlers for all list endpoints
- `internal/list/rest.go` — JSON:API resource type `"shopping-lists"`
- **Acceptance:** Full list lifecycle works: create, rename, delete, archive, unarchive, list by status

**3.3 Implement shopping item domain** [L]
- `internal/item/entity.go` — GORM entity for `shopping_items` with cascade delete
- `internal/item/model.go` — Domain model with denormalized category fields
- `internal/item/builder.go` — Builder with validation
- `internal/item/processor.go` — Add, Update, Remove, Check/Uncheck, UncheckAll
- `internal/item/provider.go` — Data access scoped to list
- `internal/item/resource.go` — HTTP handlers for item endpoints
- `internal/item/rest.go` — JSON:API resource type `"shopping-items"`
- **Acceptance:** Item CRUD works, check/uncheck works, items grouped by category sort order

**3.4 Implement meal plan import** [M]
- `internal/import/` package (or method on list processor)
- HTTP client to call recipe-service `/meals/plans/{planId}/ingredients`
- HTTP client to call category-service for category resolution
- Transform consolidated ingredients to shopping items
- Append to target list
- `POST /lists/{id}/import/meal-plan` endpoint
- **Acceptance:** Importing a meal plan appends all consolidated ingredients as shopping items with correct category data

**3.5 Add shopping-service to infrastructure** [S]
- Add to `deploy/compose/docker-compose.yml`
- Add nginx routes: `/api/v1/shopping` -> shopping-service
- **Acceptance:** Shopping-service reachable at `/api/v1/shopping` through nginx

**3.6 Unit tests for shopping domains** [L]
- Builder tests for list and item
- Processor tests for all operations
- Import logic tests with mocked HTTP responses
- **Acceptance:** Tests pass, cover CRUD, archiving, checking, import

### Phase 4: Frontend — Category Migration

Update frontend to use new category-service endpoints.

**4.1 Update category API service** [S]
- Update `services/api/ingredient.ts` category methods to call `/categories` instead of `/ingredient-categories`
- Update query keys in `lib/hooks/api/query-keys.ts`
- Update types if JSON:API type changes from `"ingredient-categories"` to `"categories"`
- **Acceptance:** Ingredient category management works against new category-service, no regressions

### Phase 5: Frontend — Shopping List Feature

Build the complete shopping list UI.

**5.1 Types and API service layer** [M]
- `types/models/shopping.ts` — ShoppingList, ShoppingItem interfaces
- `services/api/shopping.ts` — ShoppingService with all endpoint methods
- **Acceptance:** Type-safe API service covers all shopping endpoints

**5.2 React Query hooks** [M]
- `lib/hooks/api/use-shopping-lists.ts` — useShoppingLists, useShoppingList, useCreateList, useUpdateList, useDeleteList, useArchiveList, useUnarchiveList
- `lib/hooks/api/use-shopping-items.ts` — useAddItem, useUpdateItem, useRemoveItem, useCheckItem, useUncheckAll
- `lib/hooks/api/use-shopping-import.ts` — useImportMealPlan
- Update `query-keys.ts` with shopping keys
- **Acceptance:** All hooks work with proper cache invalidation

**5.3 Shopping list management page** [L]
- `pages/ShoppingListsPage.tsx` — List view of active shopping lists
- Create list dialog/form
- Rename, delete actions
- Sort by most recently updated
- **Acceptance:** Users can view, create, rename, and delete shopping lists

**5.4 Shopping list detail / edit mode** [L]
- `pages/ShoppingListDetailPage.tsx` — List detail with items
- Add item form (name, quantity, category dropdown)
- Edit/remove item actions
- Items grouped by category with uncategorized last
- Position/reorder support
- **Acceptance:** Users can manage items on a list, items display grouped by category

**5.5 Shopping mode** [L]
- Shopping mode toggle on list detail page
- Check/uncheck items
- Checked items move to bottom of category group or "In Cart" section
- No add/edit/remove in shopping mode
- "Uncheck All" action
- **Acceptance:** Users can enter shopping mode, check items, see them move, uncheck all

**5.6 Archive and history** [M]
- "Finish Shopping" action on active list → archive
- `pages/ShoppingHistoryPage.tsx` — Archived lists view
- Archived list detail (read-only)
- Unarchive action
- Delete from history
- **Acceptance:** Full archive lifecycle works, history view shows archived lists

**5.7 Meal plan import integration** [M]
- "Add to Shopping List" button on meal plan detail page
- Dialog to select target shopping list
- Calls import endpoint
- Success feedback with item count
- **Acceptance:** Users can import meal plan ingredients into a shopping list from the plan detail page

**5.8 Navigation** [S]
- Add "Shopping" entry to main navigation
- Route configuration for `/shopping`, `/shopping/:id`, `/shopping/history`
- **Acceptance:** Shopping pages accessible from navigation

### Phase 6: CI/CD and Documentation

**6.1 CI pipeline updates** [S]
- Add category-service and shopping-service to GitHub Actions build/test/lint
- Add Docker image builds for both services
- **Acceptance:** CI runs for both new services on PRs and main

**6.2 Service documentation** [M]
- `services/category-service/docs/domain.md`, `rest.md`, `storage.md`
- `services/shopping-service/docs/domain.md`, `rest.md`, `storage.md`
- Update `docs/architecture.md` with new services
- **Acceptance:** Documentation follows existing service doc pattern

---

## Risk Assessment and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Category extraction breaks ingredient workflows | High | Medium | Deploy category-service first, verify seeding, then migrate recipe-service |
| Cross-service calls add latency to import | Medium | Medium | Denormalize category data on shopping items; import is a one-time action per list |
| Stale category snapshots on shopping items | Low | High | Accepted by design — snapshots are sufficient for shopping context |
| Large meal plan imports slow | Low | Low | Imports are async from user perspective; batch insert items |
| Recipe-service category removal breaks frontend | High | Low | Phase 4 (frontend migration) should deploy alongside Phase 2 |

---

## Success Metrics

- All acceptance criteria from PRD Section 10 pass
- No regression in recipe-service ingredient/meal plan functionality
- Category-service serves existing + new default categories correctly
- Shopping list round-trip (create -> add items -> shop -> archive) works end-to-end
- Meal plan import correctly transforms and appends all ingredients
- All new services pass CI (build, test, lint)

---

## Required Resources and Dependencies

### Service Dependencies
- **category-service**: PostgreSQL (shared DB, own schema), shared Go modules
- **shopping-service**: PostgreSQL, shared Go modules, category-service (HTTP), recipe-service (HTTP for import)
- **recipe-service**: category-service (HTTP, replaces local domain)

### External Dependencies
- No new external APIs or third-party services
- No new infrastructure beyond existing PostgreSQL instance

### Shared Module Usage
- `shared/go/auth` — JWT validation middleware
- `shared/go/database` — GORM connection, migration, tenant callbacks
- `shared/go/server` — HTTP server lifecycle, JSON:API helpers
- `shared/go/tenant` — Tenant context extraction
- `shared/go/logging` — Structured logging
- `shared/go/model` — Shared domain types (if applicable)

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Category Service | L | None |
| Phase 2: Recipe Service Migration | L | Phase 1 |
| Phase 3: Shopping Service | XL | Phase 1 |
| Phase 4: Frontend Category Migration | S | Phase 1 |
| Phase 5: Frontend Shopping Feature | XL | Phase 3, Phase 4 |
| Phase 6: CI/CD and Documentation | M | Phase 1, Phase 3 |

Phases 1, 3, and 4 can partially overlap. Phase 2 requires Phase 1 complete. Phase 5 requires Phase 3 complete.
