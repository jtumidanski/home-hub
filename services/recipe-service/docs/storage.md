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

### recipe_images

At most one row per recipe.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| recipe_id | UUID | NOT NULL, unique |
| full_key | VARCHAR(512) | NOT NULL |
| thumb_key | VARCHAR(512) | NOT NULL |
| content_type | VARCHAR(50) | NOT NULL |
| width | INT | NOT NULL |
| height | INT | NOT NULL |
| full_bytes | INT | NOT NULL |
| thumb_bytes | INT | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

`width`/`height` are the dimensions of the `full` rendition. Replace semantics: a replace deletes the old row and inserts a new one in one transaction, so every replace gets a new `id` (which changes the `v` URL parameter and the ETag) while `created_at` is carried forward from the replaced row. The old rendition keys are enqueued in `recipe_image_object_deletions` in the same transaction.

### recipe_image_object_deletions

Transactional outbox for object deletions.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| household_id | UUID | NOT NULL |
| object_key | VARCHAR(512) | NOT NULL |
| attempts | INT | NOT NULL, default 0 |
| last_error | TEXT | nullable |
| created_at | TIMESTAMP | NOT NULL |

Drain semantics:

- Rows are written in the same transaction that removes or replaces an image row (replace, delete, retention purge).
- After commit, the request drains its own rows immediately (`DrainKeys`).
- A background loop runs every 5 minutes and drains up to 100 rows per run, oldest first (`ObjectReaper.Run`). It uses a tenant-less context so the tenant callback does not filter the outbox.
- A delete that returns "not found" counts as success and the row is removed.
- Any other failure increments `attempts`, stores `last_error`, and leaves the row for the next run. Retries are unbounded; `attempts` and `last_error` exist for diagnosis.
- Concurrent drainers are safe: repeated object deletes are no-ops and repeated row deletes affect nothing.

## Relationships

- `recipe_tags.recipe_id` references `recipes.id` with `ON DELETE CASCADE`
- `recipe_restorations.recipe_id` references `recipes.id`
- `canonical_ingredient_aliases.canonical_ingredient_id` references `canonical_ingredients.id` with `ON DELETE CASCADE`
- `recipe_ingredients.recipe_id` references `recipes.id`
- `recipe_ingredients.canonical_ingredient_id` references `canonical_ingredients.id`
- `recipe_planner_configs.recipe_id` references `recipes.id`
- `plan_items.plan_week_id` references `plan_weeks.id`
- `plan_items.recipe_id` references `recipes.id`
- `recipe_images.recipe_id` references a recipe by id with no database foreign key; the image row is removed by the retention purge and by the replace/delete endpoints

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
| recipe_images | idx_recipe_image_recipe | recipe_id | unique |
| recipe_images | idx_recipe_image_tenant_household | tenant_id, household_id | composite |
| recipe_image_object_deletions | idx_recipe_image_deletion_created | created_at | single |

## Migration Rules

Managed via GORM AutoMigrate on startup. No separate SQL migration files. Migrates all entity types: `recipe.Entity`, `recipe.TagEntity`, `recipe.RestorationEntity`, `ingredient.Entity`, `ingredient.AliasEntity`, `normalization.Entity`, `planner.Entity`, `plan.Entity`, `planitem.Entity`, `audit.Entity`, `recipeimage.Entity`, `recipeimage.ObjectDeletionEntity`.

Note: The foreign key constraint `fk_canonical_ingredients_category` is explicitly dropped during ingredient migration because categories are managed by the external category-service. The `category_id` column remains as an opaque UUID reference.

## Object Storage (recipe images)

Image bytes live in an S3-compatible object store (RustFS in compose and k8s), accessed through the `minio-go` client in `internal/objectstore`. Configuration is the `S3_*` variables described in the README.

### Object keys

```
<tenant_id>/<household_id>/recipes/<recipe_id>/<image_id>/full.jpg
<tenant_id>/<household_id>/recipes/<recipe_id>/<image_id>/thumb.jpg
```

Every key component is a UUID, so user input never reaches a key.

### Bucket requirements

- The bucket must be private (no anonymous access); images are served only through recipe-service.
- In production the bucket must pre-exist. The service does not create it (compose creates it with the `rustfs-init` job for local development).
- The access key should be scoped to `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` and `s3:ListBucket` on that bucket.

### Memory bounds

Image decoding is bounded by a semaphore of 2 concurrent uploads (`processSem`). A 40 MP PNG decodes to roughly 160 MB, so the worst case is about 340 MB of transient memory for decoding plus the 10 MiB upload buffers. Plan the recipe-service memory limit accordingly.

### Benchmark

`BenchmarkProcess12MP` (12 MP source to both renditions; budget 2 s per image), re-run 2026-10-14:

```
cd services/recipe-service && go test ./internal/recipeimage/imaging/ -run=^$ -bench=BenchmarkProcess12MP -benchtime=5x
BenchmarkProcess12MP-24    5    540229533 ns/op    0.36 MB/s
```

Result: about 0.54 s/op. Machine: x86_64, AMD Ryzen 9 5900X 12-Core Processor (24 threads), WSL2 (Linux 5.15 microsoft-standard).
### Backups

Image objects are outside the Postgres backups. Backing up the bucket is the operator's responsibility; a Postgres restore without a matching bucket restore leaves `recipe_images` rows pointing at missing objects (served as 404).

### Known orphan path

A crash between the object `Put` and the database commit leaves objects with no row and no outbox entry. Upload logs `storing recipe image renditions` with `recipe_id` and `image_id` before the `Put`; an `image_id` with that line but no `recipe image stored` line (and no row in `recipe_images`) identifies the orphaned `<image_id>/` objects, which can be deleted by hand.
