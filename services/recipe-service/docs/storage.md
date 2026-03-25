# Recipe Service — Storage

Schema: `recipe`

## Tables

### recipes

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, indexed |
| household_id | UUID | NOT NULL, indexed |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| source | TEXT | NOT NULL |
| servings | INT | nullable |
| prep_time_minutes | INT | nullable |
| cook_time_minutes | INT | nullable |
| source_url | VARCHAR(2048) | nullable |
| deleted_at | TIMESTAMP | nullable, indexed |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### recipe_tags

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL, FK |
| tag | VARCHAR(100) | NOT NULL |

Unique index on (recipe_id, tag).

### recipe_restorations

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL, FK |
| restored_at | TIMESTAMP | NOT NULL |

## Migrations

Managed via GORM AutoMigrate on startup. No separate SQL migration files.
