# Shopping List — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-31

---

## 1. Overview

The Shopping List feature lets household members create, manage, and shop from categorized lists of items. Lists support both ad-hoc items (e.g., "laundry detergent", "paper towels") and bulk-imported ingredients from a meal plan's consolidated ingredient list.

Users can maintain multiple named lists simultaneously (e.g., "Costco Run", "Weekly Groceries"). Each list has two interaction modes: **edit mode** for adding, removing, and organizing items, and **shopping mode** for checking items off while at the store. When shopping is complete, the user archives the list, making it read-only and visible in a history view for eventual cleanup.

This feature introduces two new services: a **category-service** that extracts the existing ingredient category concept into a shared domain consumed by both recipe-service and shopping-service, and a **shopping-service** that owns shopping list lifecycle and items.

## 2. Goals

Primary goals:
- Allow users to create and manage multiple named shopping lists per household
- Support adding freeform items with quantity, name, and category
- Provide a one-action import of consolidated ingredients from an existing meal plan (append to list)
- Offer a focused shopping mode for checking items off by category
- Enable archiving completed lists with read-only history view
- Extract ingredient categories into a shared category-service usable by both recipe-service and shopping-service

Non-goals:
- Real-time collaborative editing or live sync between household members
- Price tracking, budgeting, or cost estimation
- Store-specific lists or aisle mapping beyond category sort order
- Recurring or template shopping lists
- Automatic list generation without user action
- Push notifications or reminders

## 3. User Stories

- As a household member, I want to create a named shopping list so that I can organize purchases for different stores or trips.
- As a household member, I want to add ad-hoc items (with freeform quantity and a category) to my list so that I can track non-food purchases alongside groceries.
- As a household member, I want to import ingredients from a meal plan into my shopping list so that I don't have to manually re-enter everything.
- As a household member, I want imported ingredients appended to my existing list (not replacing it) so that I can combine multiple sources.
- As a household member, I want items organized by category with a configurable sort order so that I can shop efficiently by store section.
- As a household member, I want to enter shopping mode so that I can check items off as I pick them up, without accidentally editing the list.
- As a household member, I want to finish shopping and archive the list so that I start fresh next time while keeping a history.
- As a household member, I want to view archived lists (read-only) so that I can reference past purchases.
- As a household member, I want any member of my household to see and use the same shopping lists so that we stay coordinated.

## 4. Functional Requirements

### 4.1 Category Service (Extraction)

- Extract the existing `ingredient_categories` table and logic from recipe-service into a new category-service.
- The category-service owns CRUD for categories: list, create, update (name, sort_order), delete.
- Categories are scoped by tenant_id (consistent with current implementation).
- Auto-seed default categories on first access per tenant (same 11 defaults as today).
- Recipe-service migrates to consume category-service via HTTP for category data instead of owning the table.
- Shopping-service consumes category-service for item categorization.
- Add non-food default categories to the seed list: "Household", "Personal Care", "Baby & Kids", "Pet Supplies".

### 4.2 Shopping List Management

- Users can create a shopping list with a required name.
- Users can rename an active list.
- Users can delete an active list (hard delete, not archive).
- Users can view all active lists for their household, sorted by most recently updated.
- Users can view archived lists for their household, sorted by archived date descending.
- Lists are scoped to tenant_id + household_id.

### 4.3 Shopping List Items

- Each item has: name (required), quantity (freeform string, optional), category_id (optional reference to category-service), checked (boolean, default false), position (integer for ordering within category).
- Users can add items to an active list.
- Users can update an item's name, quantity, category, or position.
- Users can remove items from an active list.
- Items are displayed grouped by category (using category sort_order), with uncategorized items last.

### 4.4 Meal Plan Import

- From a meal plan detail view, users can select "Add to Shopping List" and choose a target shopping list.
- The system calls recipe-service's existing `/plans/{planId}/ingredients` endpoint to get consolidated ingredients.
- Each consolidated ingredient becomes a shopping list item with:
  - name = display_name (or canonical name)
  - quantity = formatted as "{quantity} {unit}" string (freeform)
  - category_id = resolved from the ingredient's category via category-service
- Imported items are appended to the list; existing items are not modified or deduplicated.
- No back-reference to the source meal plan is stored.

### 4.5 Shopping Mode

- Users can enter shopping mode on an active list.
- In shopping mode, users can check/uncheck items.
- Checked items visually move to the bottom of their category group (or a separate "In Cart" section).
- Shopping mode does not allow adding, removing, or editing items — only checking/unchecking.
- Users can "uncheck all" items to reset the list for re-shopping.
- The check state persists across sessions (stored server-side).

### 4.6 Archiving

- Users can "Finish Shopping" on an active list, which transitions it to archived status.
- Archived lists are read-only: no item modifications, no checking/unchecking.
- Archived lists appear in a separate history view.
- Archived lists retain all items and their checked/unchecked state at time of archival.
- Users can unarchive a list, returning it to active status (clears archived_at).
- Archived lists can be deleted by the user from the history view.
- No automatic cleanup is implemented in v1; cleanup is manual.

## 5. API Surface

### 5.1 Category Service

Base: `/api/v1/categories`

The category-service exposes the same contract as the current recipe-service ingredient-categories endpoints, re-mounted at a service-specific path.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/categories` | List all categories for tenant (sorted by sort_order) |
| POST | `/categories` | Create a category |
| PATCH | `/categories/{id}` | Update category name and/or sort_order |
| DELETE | `/categories/{id}` | Delete category (consumers handle orphaned references) |

JSON:API type: `"categories"`

**Response attributes:** `name`, `sort_order`, `created_at`, `updated_at`

Note: The `ingredient_count` computed field is dropped from the category-service response. Recipe-service can compute this locally if needed.

### 5.2 Shopping Service

Base: `/api/v1/shopping`

#### Lists

| Method | Path | Description |
|--------|------|-------------|
| GET | `/lists` | List shopping lists. Query param: `status=active\|archived` (default: active) |
| POST | `/lists` | Create a shopping list |
| GET | `/lists/{id}` | Get list detail with all items |
| PATCH | `/lists/{id}` | Update list name |
| DELETE | `/lists/{id}` | Delete list (active or archived) |
| POST | `/lists/{id}/archive` | Archive list (finish shopping) |
| POST | `/lists/{id}/unarchive` | Unarchive list (return to active) |

JSON:API type: `"shopping-lists"`

**Create request attributes:** `name` (required, string)

**Response attributes:** `name`, `status` (active/archived), `item_count`, `checked_count`, `archived_at` (nullable), `created_at`, `updated_at`

#### Items

| Method | Path | Description |
|--------|------|-------------|
| POST | `/lists/{id}/items` | Add item to list |
| PATCH | `/lists/{id}/items/{itemId}` | Update item |
| DELETE | `/lists/{id}/items/{itemId}` | Remove item |
| PATCH | `/lists/{id}/items/{itemId}/check` | Toggle checked state |
| POST | `/lists/{id}/items/uncheck-all` | Uncheck all items on list |

JSON:API type: `"shopping-items"`

**Create/Update request attributes:** `name` (string), `quantity` (string, optional), `category_id` (UUID, optional), `position` (int, optional)

**Check request attributes:** `checked` (boolean)

**Response attributes:** `name`, `quantity`, `category_id`, `category_name`, `category_sort_order`, `checked`, `position`, `created_at`, `updated_at`

Note: `category_name` and `category_sort_order` are denormalized onto items at creation time (fetched from category-service). This avoids runtime cross-service calls for every list render.

#### Import

| Method | Path | Description |
|--------|------|-------------|
| POST | `/lists/{id}/import/meal-plan` | Import ingredients from a meal plan |

**Request attributes:** `plan_id` (UUID, required)

The shopping-service calls recipe-service's `/api/v1/meals/plans/{planId}/ingredients` to fetch consolidated ingredients, transforms them into shopping items, and appends them to the list.

## 6. Data Model

### 6.1 Category Service

Schema: `category`

#### `categories`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| name | VARCHAR(100) | NOT NULL |
| sort_order | INT | NOT NULL, DEFAULT 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Unique index: `(tenant_id, name)`

### 6.2 Shopping Service

Schema: `shopping`

#### `shopping_lists`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'active' |
| archived_at | TIMESTAMP | NULL |
| created_by | UUID | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Index: `(tenant_id, household_id, status)`

#### `shopping_items`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| list_id | UUID | NOT NULL, FK → shopping_lists.id CASCADE |
| name | VARCHAR(255) | NOT NULL |
| quantity | VARCHAR(100) | NULL |
| category_id | UUID | NULL |
| category_name | VARCHAR(100) | NULL |
| category_sort_order | INT | NULL, DEFAULT 0 |
| checked | BOOLEAN | NOT NULL, DEFAULT false |
| position | INT | NOT NULL, DEFAULT 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Index: `(list_id, checked, category_sort_order, position)`

Note: `category_name` and `category_sort_order` are denormalized snapshots from the category-service at item creation/update time. This avoids cross-service joins and ensures list rendering is self-contained.

## 7. Service Impact

### 7.1 New: category-service

- New microservice owning the `category` schema.
- Implements CRUD for categories with tenant scoping.
- Seeds default categories (existing 11 + 4 new non-food categories).
- Added to docker-compose, nginx routing, go.work.

### 7.2 New: shopping-service

- New microservice owning the `shopping` schema.
- Implements shopping list and item CRUD, archiving, and meal plan import.
- Depends on category-service (fetch category details at item creation).
- Depends on recipe-service (fetch consolidated ingredients for import).
- Added to docker-compose, nginx routing, go.work.

### 7.3 Modified: recipe-service

- Remove `ingredient/category/` domain package (model, entity, builder, processor, provider, resource, rest).
- Remove category-related migration from recipe-service startup.
- Remove `/ingredient-categories` route registration.
- Add HTTP client to call category-service for category data where needed (e.g., consolidated ingredient response enrichment).
- Migration: Drop `ingredient_categories` table from `recipe` schema after data migration to `category` schema.
- Update `canonical_ingredients.category_id` to be an opaque UUID reference (no FK constraint, since the data now lives in another service's database).

### 7.4 Modified: frontend

- Update ingredient category API calls to point to category-service base URL.
- New shopping list pages and components.
- Add "Add to Shopping List" action on meal plan detail page.
- New navigation entry for Shopping Lists.

### 7.5 Modified: nginx

- Add routing rules for `/api/v1/categories` → category-service.
- Add routing rules for `/api/v1/shopping` → shopping-service.

## 8. Non-Functional Requirements

- **Performance:** List detail with items should render in <200ms for lists up to 200 items. Category-service responses should be cacheable (5-min stale time on frontend).
- **Multi-tenancy:** All data tenant-scoped via JWT context and GORM callbacks, consistent with existing services.
- **Household scoping:** Shopping lists scoped to household_id. Categories remain tenant-scoped (shared across households within a tenant).
- **Data integrity:** Shopping item category fields are denormalized snapshots. If a category is renamed or deleted in category-service, existing shopping items retain their snapshot values. New items will reflect the current state.
- **Service-to-service communication:** Shopping-service calls recipe-service and category-service by forwarding the user's JWT from the originating request. No service-to-service token infrastructure needed since all cross-service calls are user-initiated.
- **Observability:** Standard structured logging and tracing consistent with other services.
- **Migration safety:** Category extraction requires a data migration. Recipe-service must continue to function during the transition. Recommended approach: deploy category-service first, migrate data, then update recipe-service to consume it.

## 9. Open Questions

- Should there be a limit on the number of active lists per household? (Suggestion: no hard limit in v1, revisit if needed.)
- Should category renames in category-service propagate to existing shopping items, or is the snapshot approach sufficient? (Current spec: snapshot, no propagation.)
- Should the history view support filtering or search, or is a simple chronological list sufficient for v1?

## 10. Acceptance Criteria

- [ ] Category-service is deployed and serves CRUD for categories with tenant scoping.
- [ ] Default categories include both food and non-food options.
- [ ] Recipe-service no longer owns categories; it consumes category-service.
- [ ] Existing ingredient categorization continues to work after migration.
- [ ] Shopping-service supports creating, renaming, and deleting named lists.
- [ ] Users can add, edit, and remove items with freeform quantity and optional category.
- [ ] Items display grouped by category sort order with uncategorized items last.
- [ ] Users can import consolidated ingredients from a meal plan, appending to an existing list.
- [ ] Shopping mode allows checking/unchecking items without editing.
- [ ] "Finish Shopping" archives a list, making it read-only.
- [ ] Archived lists can be unarchived back to active status.
- [ ] "Uncheck All" resets all items on a list.
- [ ] Archived lists are visible in a history view and can be deleted.
- [ ] All data is scoped to tenant and household.
- [ ] Any household member can view and interact with the household's shopping lists.
- [ ] Frontend has navigation entry, list management page, list detail/shopping page, and history view.
