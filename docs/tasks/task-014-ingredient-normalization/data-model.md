# Ingredient Normalization — Data Model

## Entity Relationship Diagram

```
recipes (existing)
  │
  ├──< recipe_ingredients (new)
  │       │
  │       └──> canonical_ingredients (new)
  │                │
  │                └──< canonical_ingredient_aliases (new)
  │
  ├──< recipe_planner_configs (new, 1:1)
  │
  ├──< recipe_tags (existing)
  │
  └──< recipe_restorations (existing)

recipe_audit_events (new, standalone)
```

## New Tables

### `recipe_ingredients`

Persisted parsed ingredients with raw and normalized values. Re-created on every recipe source update.

```sql
CREATE TABLE recipe_ingredients (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    household_id    UUID NOT NULL,
    recipe_id       UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    raw_name        VARCHAR(255) NOT NULL,
    raw_quantity    VARCHAR(100),
    raw_unit        VARCHAR(100),
    position        INT NOT NULL,
    canonical_ingredient_id UUID REFERENCES canonical_ingredients(id) ON DELETE SET NULL,
    canonical_unit  VARCHAR(50),
    normalization_status VARCHAR(30) NOT NULL DEFAULT 'unresolved',
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);

CREATE INDEX idx_recipe_ingredient_recipe ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredient_tenant ON recipe_ingredients(tenant_id, household_id);
CREATE INDEX idx_recipe_ingredient_canonical ON recipe_ingredients(canonical_ingredient_id);
```

**normalization_status values:**
- `matched` — exact canonical name match
- `alias_matched` — matched via alias lookup
- `unresolved` — no match found
- `manually_confirmed` — user manually assigned a canonical ingredient

**Lifecycle:**
- Created during recipe create (after Cooklang parse)
- On recipe source update, **reconciled** (not blindly deleted and re-created):
  - New parsed ingredients matched to existing records by `raw_name` (lowercased)
  - `manually_confirmed` records carry forward their canonical assignment when the raw name still exists
  - Other statuses are re-normalized against the current registry
  - Records for ingredients no longer in the source are deleted
  - New ingredients not matching any existing record run normalization from scratch
  - Position values are updated to reflect new ordering
- `canonical_ingredient_id` and `normalization_status` updated by manual correction endpoint
- `canonical_ingredient_id` and `normalization_status` updated by re-normalize endpoint (only for `unresolved` ingredients; `manually_confirmed` are preserved)
- `canonical_ingredient_id` updated to target by reassign endpoint (`POST /ingredients/:id/reassign`)
- `canonical_ingredient_id` set to NULL if referenced canonical ingredient is deleted (ON DELETE SET NULL)

### `canonical_ingredients`

Tenant-scoped ingredient registry. Each entry represents one logical ingredient.

```sql
CREATE TABLE canonical_ingredients (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    display_name    VARCHAR(255),
    unit_family     VARCHAR(20),
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_canonical_ingredient_tenant_name ON canonical_ingredients(tenant_id, name);
CREATE INDEX idx_canonical_ingredient_tenant ON canonical_ingredients(tenant_id);
```

**unit_family values:** `count`, `weight`, `volume`, or NULL (unspecified)

**Constraints:**
- `(tenant_id, name)` is unique — no two canonical ingredients in the same tenant can share a name
- `name` is stored normalized: lowercase, trimmed, collapsed whitespace

### `canonical_ingredient_aliases`

Alternate names that map to a canonical ingredient. Used during normalization to resolve variations.

```sql
CREATE TABLE canonical_ingredient_aliases (
    id                      UUID PRIMARY KEY,
    tenant_id               UUID NOT NULL,
    canonical_ingredient_id UUID NOT NULL REFERENCES canonical_ingredients(id) ON DELETE CASCADE,
    name                    VARCHAR(255) NOT NULL,
    created_at              TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_alias_tenant_name ON canonical_ingredient_aliases(tenant_id, name);
CREATE INDEX idx_alias_canonical ON canonical_ingredient_aliases(canonical_ingredient_id);
```

**Constraints:**
- `(tenant_id, name)` is unique — an alias name cannot point to two different canonical ingredients
- Alias name must not conflict with any canonical ingredient name in the same tenant
- `name` is stored normalized: lowercase, trimmed

### `recipe_planner_configs`

Planner-specific metadata per recipe. One-to-one relationship with recipes.

```sql
CREATE TABLE recipe_planner_configs (
    id                  UUID PRIMARY KEY,
    recipe_id           UUID NOT NULL UNIQUE REFERENCES recipes(id) ON DELETE CASCADE,
    classification      VARCHAR(50),
    servings_yield      INT,
    eat_within_days     INT,
    min_gap_days        INT,
    max_consecutive_days INT,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_planner_config_recipe ON recipe_planner_configs(recipe_id);
```

**classification values:** Free-form but UI presents common options: `breakfast`, `lunch`, `dinner`, `snack`, `side`

### `recipe_audit_events`

Write-only audit log for recipe-service actions. No foreign keys — entity may be deleted after event is recorded.

```sql
CREATE TABLE recipe_audit_events (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id   UUID NOT NULL,
    action      VARCHAR(50) NOT NULL,
    actor_id    UUID NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMP NOT NULL
);

CREATE INDEX idx_audit_entity ON recipe_audit_events(tenant_id, entity_type, entity_id);
CREATE INDEX idx_audit_action ON recipe_audit_events(tenant_id, action);
CREATE INDEX idx_audit_time ON recipe_audit_events(created_at);
```

## Unit Registry (Static, In-Memory)

Not a database table. Defined as a Go map in `normalization/unit_registry.go`.

```go
// UnitFamily represents a category of measurement units.
type UnitFamily string

const (
    UnitFamilyCount  UnitFamily = "count"
    UnitFamilyWeight UnitFamily = "weight"
    UnitFamilyVolume UnitFamily = "volume"
)

// CanonicalUnit represents a normalized unit identity.
type CanonicalUnit struct {
    Name   string     // e.g., "gram", "cup", "each"
    Family UnitFamily // e.g., "weight", "volume", "count"
}
```

**Mapping (raw string -> CanonicalUnit):**

| Raw Strings | Canonical Name | Family |
|-------------|---------------|--------|
| each, piece, pcs, count, whole | each | count |
| clove, cloves | clove | count |
| head, heads | head | count |
| bunch, bunches | bunch | count |
| sprig, sprigs | sprig | count |
| stalk, stalks | stalk | count |
| slice, slices | slice | count |
| pinch, pinches | pinch | count |
| dash, dashes | dash | count |
| g, gram, grams | gram | weight |
| kg, kilogram, kilograms | kilogram | weight |
| oz, ounce, ounces | ounce | weight |
| lb, pound, pounds | pound | weight |
| ml, milliliter, milliliters | milliliter | volume |
| l, liter, liters | liter | volume |
| tsp, teaspoon, teaspoons | teaspoon | volume |
| tbsp, tablespoon, tablespoons | tablespoon | volume |
| cup, cups | cup | volume |
| fl oz, fluid ounce, fluid ounces | fluid ounce | volume |

Units not in this map resolve to `nil` canonical unit (raw unit is still preserved).

## GORM Entity Patterns

All new entities follow existing project conventions:

- UUID primary keys (generated via `uuid.New()`)
- `TableName()` method on each entity
- `Migration()` function per domain package using `db.AutoMigrate(...)`
- `Make(entity) (Model, error)` function to convert entity -> domain model
- `ToEntity()` method on model to convert domain model -> entity
- Tenant scoping via GORM callbacks (consistent with existing services)
