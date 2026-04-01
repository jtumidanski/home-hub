# Shopping Storage

## Tables

### shopping_lists

| Column | Type | Constraints |
|---|---|---|
| id | uuid | Primary key |
| tenant_id | uuid | Not null |
| household_id | uuid | Not null |
| name | varchar(255) | Not null |
| status | varchar(20) | Not null, default `active` |
| archived_at | timestamp | Nullable |
| created_by | uuid | Not null |
| created_at | timestamp | Not null |
| updated_at | timestamp | Not null |

### shopping_items

| Column | Type | Constraints |
|---|---|---|
| id | uuid | Primary key |
| list_id | uuid | Not null |
| name | varchar(255) | Not null |
| quantity | varchar(100) | Nullable |
| category_id | uuid | Nullable |
| category_name | varchar(100) | Nullable |
| category_sort_order | int | Nullable |
| checked | bool | Not null, default `false` |
| position | int | Not null, default `0` |
| created_at | timestamp | Not null |
| updated_at | timestamp | Not null |

## Relationships

- `shopping_items.list_id` references a `shopping_lists.id`

## Indexes

| Index | Table | Columns |
|---|---|---|
| idx_shopping_list_tenant_household_status | shopping_lists | tenant_id, household_id, status |
| idx_shopping_item_list_checked_sort | shopping_items | list_id, checked, position, category_sort_order |

## Migration Rules

Migrations are managed by GORM AutoMigrate, executed at service startup for both `shopping_lists` and `shopping_items` entities.
