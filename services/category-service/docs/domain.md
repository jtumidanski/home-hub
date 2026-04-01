# Category

## Responsibility

The category domain manages named, ordered groupings used to classify items (e.g. grocery ingredients) within a tenant. It enforces name uniqueness per tenant, validates input constraints, and auto-seeds a default set of categories when a tenant has none.

## Core Models

### Model

Immutable domain model with private fields and getter accessors.

| Field     | Type      |
|-----------|-----------|
| id        | uuid.UUID |
| tenantID  | uuid.UUID |
| name      | string    |
| sortOrder | int       |
| createdAt | time.Time |
| updatedAt | time.Time |

### Builder

Fluent builder for constructing a validated `Model`.

**Setter methods** (chainable): `SetId`, `SetTenantID`, `SetName`, `SetSortOrder`, `SetCreatedAt`, `SetUpdatedAt`.

`Build()` returns `(Model, error)` after applying validation rules.

## Invariants

- Name must not be empty (`ErrNameRequired`).
- Name must not exceed 100 characters (`ErrNameTooLong`).
- Sort order must be >= 0 (`ErrInvalidSortOrder`).
- Name must be unique within a tenant (`ErrDuplicateName`).

## Processors

### Processor

Constructed with a `*gorm.DB` reference.

#### List

Returns all categories for a tenant ordered by sort order ascending. If the tenant has no categories, default categories are seeded within a transaction before returning.

**Default categories** (seeded in order): Produce, Meats & Seafood, Dairy & Eggs, Bakery & Bread, Pantry & Dry Goods, Frozen, Beverages, Snacks & Sweets, Condiments & Sauces, Spices & Seasonings, Other, Household, Personal Care, Baby & Kids, Pet Supplies.

#### Create

Trims whitespace from the name, validates via the builder, checks for duplicate names within the tenant, assigns the next sort order (max + 1), and persists a new category.

#### Update

Fetches the existing category by ID. Optionally updates name (with trim, validation, and uniqueness check allowing the same ID) and/or sort order (with non-negative validation). Persists changes.

#### Delete

Verifies the category exists by ID, then deletes it. No cascade behavior.
