# Storage

## Tables

### categories

Schema: `category`

| Column     | Type         | Constraints          |
|------------|--------------|----------------------|
| id         | uuid         | PRIMARY KEY          |
| tenant_id  | uuid         | NOT NULL             |
| name       | varchar(100) | NOT NULL             |
| sort_order | int          | DEFAULT 0            |
| created_at | timestamp    | NOT NULL             |
| updated_at | timestamp    | NOT NULL             |

## Relationships

None. The categories table has no foreign key constraints.

## Indexes

| Name                       | Columns          | Unique |
|----------------------------|------------------|--------|
| `idx_category_tenant_name` | tenant_id, name  | yes    |

## Migration Rules

Migrations are executed via GORM AutoMigrate on service startup.
