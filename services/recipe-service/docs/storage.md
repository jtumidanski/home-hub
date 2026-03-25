# Recipe Service — Storage

Schema: `recipe`

## Tables

### recipes

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| source | TEXT | NOT NULL |
| servings | INT | nullable |
| prep_time_minutes | INT | nullable |
| cook_time_minutes | INT | nullable |
| source_url | VARCHAR(2048) | nullable |
| deleted_at | TIMESTAMP | nullable |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### recipe_tags

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL |
| tag | VARCHAR(100) | NOT NULL |

### recipe_restorations

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL |
| restored_at | TIMESTAMP | NOT NULL |

## Relationships

- `recipe_tags.recipe_id` references `recipes.id` with `ON DELETE CASCADE`
- `recipe_restorations.recipe_id` references `recipes.id`

## Indexes

| Table | Index | Columns | Type |
|-------|-------|---------|------|
| recipes | idx_recipe_tenant_household | tenant_id, household_id | composite |
| recipes | idx_recipe_soft_delete | deleted_at | single |
| recipe_tags | idx_tag_recipe_unique | recipe_id, tag | unique |
| recipe_tags | idx_tag_value | tag | single |

## Migration Rules

Managed via GORM AutoMigrate on startup. No separate SQL migration files. Migrates `Entity`, `TagEntity`, and `RestorationEntity`.
