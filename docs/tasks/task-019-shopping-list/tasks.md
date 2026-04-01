# Task 019: Shopping List — Task Checklist

Last Updated: 2026-03-31

---

## Phase 1: Category Service (Backend)

- [ ] 1.1 Scaffold category-service (cmd, Dockerfile, go.mod, config, go.work) [M]
- [ ] 1.2 Implement category domain (entity, model, builder, processor, provider, resource, rest) [L]
- [ ] 1.3 Add category-service to docker-compose and nginx [S]
- [ ] 1.4 Verify data migration — auto-seed 15 defaults, tenant scoping works [M]
- [ ] 1.5 Unit tests for category domain [M]

## Phase 2: Recipe Service Migration

- [ ] 2.1 Add HTTP client for category-service in recipe-service [M]
- [ ] 2.2 Remove category domain from recipe-service [M]
- [ ] 2.3 Update recipe-service tests [M]

## Phase 3: Shopping Service (Backend)

- [ ] 3.1 Scaffold shopping-service (cmd, Dockerfile, go.mod, config, go.work) [M]
- [ ] 3.2 Implement shopping list domain (entity, model, builder, processor, provider, resource, rest) [L]
- [ ] 3.3 Implement shopping item domain (entity, model, builder, processor, provider, resource, rest) [L]
- [ ] 3.4 Implement meal plan import (HTTP clients, transform, append) [M]
- [ ] 3.5 Add shopping-service to docker-compose and nginx [S]
- [ ] 3.6 Unit tests for shopping domains [L]

## Phase 4: Frontend — Category Migration

- [ ] 4.1 Update category API calls to `/categories`, update types and query keys [S]

## Phase 5: Frontend — Shopping List Feature

- [ ] 5.1 Types and API service layer (ShoppingService, models) [M]
- [ ] 5.2 React Query hooks (lists, items, import) [M]
- [ ] 5.3 Shopping list management page (list, create, rename, delete) [L]
- [ ] 5.4 Shopping list detail / edit mode (items, grouped by category) [L]
- [ ] 5.5 Shopping mode (check/uncheck, in-cart section, uncheck all) [L]
- [ ] 5.6 Archive and history (finish shopping, history page, unarchive, delete) [M]
- [ ] 5.7 Meal plan import integration (button on plan detail, list picker dialog) [M]
- [ ] 5.8 Navigation and routing [S]

## Phase 6: CI/CD and Documentation

- [ ] 6.1 Add category-service and shopping-service to CI pipeline [S]
- [ ] 6.2 Service documentation (domain.md, rest.md, storage.md for both services) [M]

---

## Verification Checklist (PRD Acceptance Criteria)

- [ ] Category-service deployed and serves CRUD with tenant scoping
- [ ] Default categories include both food and non-food options (15 total)
- [ ] Recipe-service no longer owns categories; consumes category-service
- [ ] Existing ingredient categorization works after migration
- [ ] Shopping-service supports creating, renaming, and deleting named lists
- [ ] Users can add, edit, and remove items with freeform quantity and optional category
- [ ] Items display grouped by category sort order, uncategorized last
- [ ] Users can import consolidated ingredients from a meal plan, appending to list
- [ ] Shopping mode allows checking/unchecking items without editing
- [ ] "Finish Shopping" archives a list, making it read-only
- [ ] Archived lists can be unarchived back to active status
- [ ] "Uncheck All" resets all items on a list
- [ ] Archived lists visible in history view and can be deleted
- [ ] All data scoped to tenant and household
- [ ] Any household member can view and interact with household's shopping lists
- [ ] Frontend has navigation entry, list management, list detail/shopping, and history views
