# Shopping Domain

## List

### Responsibility

Represents a named shopping list scoped to a tenant and household. Lists track their status and aggregate item counts.

### Core Models

**Model** (`internal/list/model.go`)

| Field | Type | Description |
|---|---|---|
| id | UUID | Unique identifier |
| tenantID | UUID | Owning tenant |
| householdID | UUID | Owning household |
| name | string | List name |
| status | string | `active` or `archived` |
| archivedAt | *time.Time | Timestamp when archived, nil if active |
| createdBy | UUID | User who created the list |
| itemCount | int | Total items in the list (computed) |
| checkedCount | int | Checked items in the list (computed) |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

The model is immutable. Fields are private with getter methods. `WithCounts` returns a new model with updated item and checked counts.

### Invariants

- Name is required
- Name must not exceed 255 characters
- Default status is `active`
- `IsArchived()` returns true when status is `archived`

### State Transitions

- **active -> archived**: Sets status to `archived` and records `archivedAt` timestamp
- **archived -> active**: Sets status to `active` and clears `archivedAt`
- Archiving an already-archived list returns `ErrAlreadyArchived`
- Unarchiving a non-archived list returns `ErrNotArchived`
- Updating an archived list returns `ErrArchived`

### Processors

**Processor** (`internal/list/processor.go`)

| Method | Description |
|---|---|
| List(status) | Returns lists filtered by status (defaults to `active`), with item counts |
| Get(id) | Returns a single list with item counts |
| Create(tenantID, householdID, userID, name) | Creates a new active list |
| Update(id, name) | Updates list name; rejects if archived |
| Delete(id) | Hard-deletes a list |
| Archive(id) | Transitions list from active to archived |
| Unarchive(id) | Transitions list from archived to active |
| GetWithItems(id) | Returns a list with all its items |
| AddItem(listID, input, authHeader) | Validates list is not archived, enriches category from category service, creates item |
| UpdateItem(listID, itemID, input, authHeader) | Validates list is not archived, enriches category from category service, updates item |
| RemoveItem(listID, itemID) | Validates list is not archived, deletes item |
| CheckItem(listID, itemID, checked) | Validates list is not archived, sets item checked state |
| UncheckAllItems(listID) | Validates list is not archived, unchecks all items in list |
| ImportFromMealPlan(listID, planID, authHeader) | Validates list is not archived, fetches ingredients from recipe service, resolves categories, creates items |

---

## Item

### Responsibility

Represents a single item on a shopping list. Items have optional quantity, category association, checked state, and sort position.

### Core Models

**Model** (`internal/item/model.go`)

| Field | Type | Description |
|---|---|---|
| id | UUID | Unique identifier |
| listID | UUID | Parent shopping list |
| name | string | Item name |
| quantity | *string | Optional quantity description |
| categoryID | *UUID | Optional category reference |
| categoryName | *string | Denormalized category name |
| categorySortOrder | *int | Denormalized category sort order |
| checked | bool | Whether the item is checked off |
| position | int | Sort position within the list |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

The model is immutable with private fields and getter methods.

### Invariants

- Name is required
- Name must not exceed 255 characters

### Processors

**Processor** (`internal/item/processor.go`)

| Method | Description |
|---|---|
| GetByListID(listID) | Returns all items for a list, ordered by category sort order (nulls last), position, created_at |
| Add(input) | Creates a new unchecked item; auto-assigns position if not provided |
| Update(id, input) | Updates item fields; supports clearing category via `ClearCategory` flag |
| Delete(id) | Hard-deletes an item |
| Check(id, checked) | Sets the checked state of an item |
| UncheckAll(listID) | Bulk-unchecks all items for a list |
