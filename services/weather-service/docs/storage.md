# Storage

All tables are created in the PostgreSQL `weather` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### weather_caches

| Column        | Type             | Constraints                           |
|---------------|------------------|---------------------------------------|
| id            | UUID             | PRIMARY KEY                           |
| tenant_id     | UUID             | NOT NULL, INDEX                       |
| household_id  | UUID             | NOT NULL, UNIQUE INDEX                |
| latitude      | DOUBLE PRECISION | NOT NULL                              |
| longitude     | DOUBLE PRECISION | NOT NULL                              |
| units         | TEXT             | NOT NULL                              |
| current_data  | JSONB            | NOT NULL                              |
| forecast_data | JSONB            | NOT NULL                              |
| fetched_at    | TIMESTAMP        | NOT NULL                              |
| created_at    | TIMESTAMP        | NOT NULL                              |
| updated_at    | TIMESTAMP        | NOT NULL                              |

## JSONB Structures

### current_data

```json
{
  "temperature": 72.5,
  "weatherCode": 2,
  "summary": "Partly Cloudy",
  "icon": "cloud-sun"
}
```

### forecast_data

```json
[
  {
    "date": "2026-03-25",
    "highTemperature": 78.0,
    "lowTemperature": 55.0,
    "weatherCode": 2,
    "summary": "Partly Cloudy",
    "icon": "cloud-sun"
  }
]
```

## Indexes

| Table          | Index Name            | Columns      | Type   |
|----------------|-----------------------|--------------|--------|
| weather_caches | idx_weather_household | household_id | UNIQUE |
| weather_caches | (auto)                | tenant_id    | INDEX  |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: weather_caches.
