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

### canonical_ingredients

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| display_name | VARCHAR(255) | nullable |
| unit_family | VARCHAR(20) | nullable |
| category_id | UUID | nullable |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### canonical_ingredient_aliases

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| canonical_ingredient_id | UUID | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |

### recipe_ingredients

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| recipe_id | UUID | NOT NULL |
| raw_name | VARCHAR(255) | NOT NULL |
| raw_quantity | VARCHAR(100) | nullable |
| raw_unit | VARCHAR(100) | nullable |
| position | INT | NOT NULL |
| canonical_ingredient_id | UUID | nullable |
| canonical_unit | VARCHAR(50) | nullable |
| normalization_status | VARCHAR(30) | NOT NULL, default 'unresolved' |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### recipe_planner_configs

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL |
| classification | VARCHAR(50) | nullable |
| servings_yield | INT | nullable |
| eat_within_days | INT | nullable |
| min_gap_days | INT | nullable |
| max_consecutive_days | INT | nullable |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### plan_weeks

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| starts_on | DATE | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| locked | BOOL | NOT NULL, default false |
| created_by | UUID | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### plan_items

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| plan_week_id | UUID | NOT NULL |
| day | DATE | NOT NULL |
| slot | VARCHAR(20) | NOT NULL |
| recipe_id | UUID | NOT NULL |
| serving_multiplier | DECIMAL(5,2) | nullable |
| planned_servings | INT | nullable |
| notes | TEXT | nullable |
| position | INT | NOT NULL, default 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

### recipe_audit_events

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| entity_type | VARCHAR(50) | NOT NULL |
| entity_id | UUID | NOT NULL |
| action | VARCHAR(50) | NOT NULL |
| actor_id | UUID | NOT NULL |
| metadata | JSONB | nullable |
| created_at | TIMESTAMP | NOT NULL |

## Relationships

- `recipe_tags.recipe_id` references `recipes.id` with `ON DELETE CASCADE`
- `recipe_restorations.recipe_id` references `recipes.id`
- `canonical_ingredient_aliases.canonical_ingredient_id` references `canonical_ingredients.id` with `ON DELETE CASCADE`
- `recipe_ingredients.recipe_id` references `recipes.id`
- `recipe_ingredients.canonical_ingredient_id` references `canonical_ingredients.id`
- `recipe_planner_configs.recipe_id` references `recipes.id`
- `plan_items.plan_week_id` references `plan_weeks.id`
- `plan_items.recipe_id` references `recipes.id`

## Indexes

| Table | Index | Columns | Type |
|-------|-------|---------|------|
| recipes | idx_recipe_tenant_household | tenant_id, household_id | composite |
| recipes | idx_recipe_soft_delete | deleted_at | single |
| recipe_tags | idx_tag_recipe_unique | recipe_id, tag | unique |
| recipe_tags | idx_tag_value | tag | single |
| canonical_ingredients | idx_canonical_ingredient_tenant_name | tenant_id, name | unique |
| canonical_ingredients | idx_canonical_ingredient_category | category_id | single |
| canonical_ingredient_aliases | idx_alias_tenant_name | tenant_id, name | unique |
| canonical_ingredient_aliases | idx_alias_canonical_ingredient | canonical_ingredient_id | single |
| recipe_ingredients | idx_recipe_ingredient_tenant_household | tenant_id, household_id | composite |
| recipe_ingredients | idx_recipe_ingredient_recipe | recipe_id | single |
| recipe_ingredients | idx_recipe_ingredient_canonical | canonical_ingredient_id | single |
| recipe_planner_configs | idx_planner_config_recipe | recipe_id | unique |
| plan_weeks | idx_plan_week_unique | tenant_id, household_id, starts_on | unique |
| plan_weeks | idx_plan_week_tenant_household | tenant_id, household_id | composite |
| plan_items | idx_plan_item_plan_week | plan_week_id | single |
| plan_items | idx_plan_item_recipe | recipe_id | single |
| recipe_audit_events | idx_audit_tenant_entity | tenant_id, entity_type, entity_id | composite |
| recipe_audit_events | idx_audit_tenant_action | tenant_id, action | composite |
| recipe_audit_events | idx_audit_created_at | created_at | single |

## Migration Rules

Managed via GORM AutoMigrate on startup. No separate SQL migration files. Migrates all entity types: `recipe.Entity`, `recipe.TagEntity`, `recipe.RestorationEntity`, `ingredient.Entity`, `ingredient.AliasEntity`, `normalization.Entity`, `planner.Entity`, `plan.Entity`, `planitem.Entity`, `audit.Entity`.

Note: The foreign key constraint `fk_canonical_ingredients_category` is explicitly dropped during ingredient migration because categories are managed by the external category-service. The `category_id` column remains as an opaque UUID reference.
