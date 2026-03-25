# Recipe Management — Data Model

## Schema: `recipe`

All tables live in the `recipe` schema, owned exclusively by `recipe-service`.

## Entity Relationship

```
recipe.recipes 1──* recipe.recipe_tags
recipe.recipes 1──* recipe.recipe_restorations
```

## Tables

### recipe.recipes

The core recipe table. Stores recipe metadata and raw Cooklang source. Parsed ingredients and steps are derived at read time, not stored.

| Column | Go Type | GORM Type | Constraints | Notes |
|--------|---------|-----------|-------------|-------|
| id | uuid.UUID | `type:uuid;primaryKey` | PK | Application-generated |
| tenant_id | uuid.UUID | `type:uuid;not null;index:idx_recipe_tenant_household` | NOT NULL | From JWT claims |
| household_id | uuid.UUID | `type:uuid;not null;index:idx_recipe_tenant_household` | NOT NULL | From JWT claims |
| title | string | `type:varchar(255);not null` | NOT NULL | |
| description | *string | `type:text` | nullable | |
| source | string | `type:text;not null` | NOT NULL | Raw Cooklang text |
| servings | *int | `type:int` | nullable | |
| prep_time_minutes | *int | `type:int` | nullable | |
| cook_time_minutes | *int | `type:int` | nullable | |
| source_url | *string | `type:varchar(2048)` | nullable | |
| created_at | time.Time | `autoCreateTime` | NOT NULL | |
| updated_at | time.Time | `autoUpdateTime` | NOT NULL | |
| deleted_at | *time.Time | `index:idx_recipe_soft_delete` | nullable | Soft delete |

**Indexes:**
- `idx_recipe_tenant_household` — `(tenant_id, household_id)` — primary query path
- `idx_recipe_soft_delete` — `(tenant_id, household_id, deleted_at)` — filter active recipes
- `idx_recipe_title_search` — `(tenant_id, household_id, title)` — title search support

### recipe.recipe_tags

Join table for recipe tags. Tags are stored as normalized (lowercase, trimmed) strings.

| Column | Go Type | GORM Type | Constraints | Notes |
|--------|---------|-----------|-------------|-------|
| id | uuid.UUID | `type:uuid;primaryKey` | PK | Application-generated |
| recipe_id | uuid.UUID | `type:uuid;not null;index:idx_tag_recipe` | NOT NULL, FK | References recipes.id |
| tag | string | `type:varchar(100);not null` | NOT NULL | Normalized lowercase |

**Indexes:**
- `idx_tag_recipe_unique` — `(recipe_id, tag)` UNIQUE — prevent duplicate tags per recipe
- `idx_tag_value` — `(tag)` — for tag listing and filtering

**GORM relationship:** `Recipe` has many `RecipeTags`, with `OnDelete: CASCADE`.

### recipe.recipe_restorations

Tracks restoration of soft-deleted recipes (consistent with productivity-service pattern).

| Column | Go Type | GORM Type | Constraints | Notes |
|--------|---------|-----------|-------------|-------|
| id | uuid.UUID | `type:uuid;primaryKey` | PK | Application-generated |
| recipe_id | uuid.UUID | `type:uuid;not null` | NOT NULL, FK | References recipes.id |
| restored_at | time.Time | `not null` | NOT NULL | |

## GORM Entity Mapping

```go
type Entity struct {
    ID              uuid.UUID       `gorm:"type:uuid;primaryKey"`
    TenantID        uuid.UUID       `gorm:"type:uuid;not null;index:idx_recipe_tenant_household"`
    HouseholdID     uuid.UUID       `gorm:"type:uuid;not null;index:idx_recipe_tenant_household"`
    Title           string          `gorm:"type:varchar(255);not null"`
    Description     *string         `gorm:"type:text"`
    Source          string          `gorm:"type:text;not null"`
    Servings        *int            `gorm:"type:int"`
    PrepTimeMinutes *int            `gorm:"type:int"`
    CookTimeMinutes *int            `gorm:"type:int"`
    SourceURL       *string         `gorm:"type:varchar(2048)"`
    Tags            []TagEntity     `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE"`
    CreatedAt       time.Time       `gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
    DeletedAt       gorm.DeletedAt  `gorm:"index:idx_recipe_soft_delete"`
}

type TagEntity struct {
    ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
    RecipeID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_tag_recipe_unique"`
    Tag      string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_tag_recipe_unique;index:idx_tag_value"`
}

type RestorationEntity struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
    RecipeID   uuid.UUID `gorm:"type:uuid;not null"`
    RestoredAt time.Time `gorm:"not null"`
}
```

## Migration

Migrations run via GORM AutoMigrate on startup in `entity.go`, consistent with all other services. No separate SQL migration files.

```go
func Migration(db *gorm.DB) error {
    return db.AutoMigrate(&Entity{}, &TagEntity{}, &RestorationEntity{})
}
```
