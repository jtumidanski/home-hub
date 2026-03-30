# Meals v2 — Data Model

## New Tables

Both tables live in the `recipe` schema alongside existing recipe tables.

### plan_week

Weekly meal plan for a household.

```
┌─────────────────────────────────────────────┐
│ plan_week                                   │
├─────────────────────────────────────────────┤
│ id              UUID          PK            │
│ tenant_id       UUID          NOT NULL      │
│ household_id    UUID          NOT NULL      │
│ starts_on       DATE          NOT NULL      │
│ name            VARCHAR(255)  NOT NULL      │
│ locked          BOOLEAN       NOT NULL [F]  │
│ created_by      UUID          NOT NULL      │
│ created_at      TIMESTAMP     NOT NULL      │
│ updated_at      TIMESTAMP     NOT NULL      │
├─────────────────────────────────────────────┤
│ UNIQUE (tenant_id, household_id, starts_on) │
│ INDEX  (tenant_id, household_id)            │
└─────────────────────────────────────────────┘
```

Notes:
- `starts_on` can be any date (week start day is a frontend preference, not a backend constraint)
- `name` auto-generated as "Week of {Month Day, Year}" on creation, overridable
- `created_by` is the user_id from JWT claims
- `locked` defaults to false

### plan_item

A single recipe placement within a weekly plan.

```
┌──────────────────────────────────────────────────┐
│ plan_item                                        │
├──────────────────────────────────────────────────┤
│ id                  UUID          PK             │
│ plan_week_id        UUID          NOT NULL [FK]  │
│ day                 DATE          NOT NULL       │
│ slot                VARCHAR(20)   NOT NULL       │
│ recipe_id           UUID          NOT NULL [FK]  │
│ serving_multiplier  DECIMAL(5,2)  nullable       │
│ planned_servings    INTEGER       nullable        │
│ notes               TEXT          nullable        │
│ position            INTEGER       NOT NULL [0]   │
│ created_at          TIMESTAMP     NOT NULL       │
│ updated_at          TIMESTAMP     NOT NULL       │
├──────────────────────────────────────────────────┤
│ FK plan_week_id → plan_week.id ON DELETE CASCADE │
│ FK recipe_id → recipes.id (NO CASCADE)            │
│ INDEX (plan_week_id)                             │
│ INDEX (recipe_id)                                │
└──────────────────────────────────────────────────┘
```

Notes:
- `day` must fall within `plan_week.starts_on` to `plan_week.starts_on + 6 days` — enforced in processor
- `slot` constrained to: `breakfast`, `lunch`, `dinner`, `snack`, `side` — enforced in processor and builder
- `position` orders items within the same day+slot (allows multiple recipes per slot, e.g., main + side for dinner)
- `serving_multiplier` and `planned_servings` are both nullable; see PRD §4.3 for precedence rules
- CASCADE delete on `plan_week_id` ensures items are removed when a plan is deleted
- NO CASCADE on `recipe_id` — plan items survive recipe soft-deletion; the item remains with a `recipe_deleted` indicator computed at read time

## Relationships

```
plan_week 1 ──── * plan_item
                      │
                      └──── 1 recipe
                                │
                                └──── * recipe_ingredient (normalized)
                                          │
                                          └──── 0..1 canonical_ingredient
```

## Existing Tables Used (No Modifications)

| Table | Usage |
|-------|-------|
| recipes | Validate recipe_id, read title/servings for display and export |
| recipe_ingredients | Read normalized ingredient data for consolidation |
| canonical_ingredients | Read display_name and unit_family for export formatting |
| recipe_planner_configs | Read classification and servings_yield for display |
| recipe_audit_events | Store audit events with entity_type "plan" |

## GORM Entity Mapping

### PlanWeekEntity

```go
type Entity struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    TenantID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_plan_week_unique,priority:1"`
    HouseholdID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_plan_week_unique,priority:2"`
    StartsOn    time.Time `gorm:"type:date;not null;uniqueIndex:idx_plan_week_unique,priority:3"`
    Name        string    `gorm:"type:varchar(255);not null"`
    Locked      bool      `gorm:"not null;default:false"`
    CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt   time.Time `gorm:"not null"`
    UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "plan_weeks" }
```

### PlanItemEntity

```go
type Entity struct {
    ID                uuid.UUID        `gorm:"type:uuid;primaryKey"`
    PlanWeekID        uuid.UUID        `gorm:"type:uuid;not null;index"`
    Day               time.Time        `gorm:"type:date;not null"`
    Slot              string           `gorm:"type:varchar(20);not null"`
    RecipeID          uuid.UUID        `gorm:"type:uuid;not null;index;constraint:OnDelete:NO ACTION"`
    ServingMultiplier *float64           `gorm:"type:decimal(5,2)"`
    PlannedServings   *int             `gorm:"type:integer"`
    Notes             *string          `gorm:"type:text"`
    Position          int              `gorm:"not null;default:0"`
    CreatedAt         time.Time        `gorm:"not null"`
    UpdatedAt         time.Time        `gorm:"not null"`
}

func (Entity) TableName() string { return "plan_items" }
```

## Migration

Both tables are added via GORM AutoMigrate on service startup, following the existing pattern in `entity.go` files:

```go
func Migration(db *gorm.DB) error {
    return db.AutoMigrate(&Entity{})
}
```

The plan domain's migration is registered in `main.go` alongside existing domain migrations via `database.SetMigrations(...)`.
