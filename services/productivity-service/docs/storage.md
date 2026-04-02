# Storage

All tables are created in the PostgreSQL `productivity` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### tasks

| Column              | Type      | Constraints              |
|---------------------|-----------|--------------------------|
| id                  | UUID      | PRIMARY KEY              |
| tenant_id           | UUID      | NOT NULL                 |
| household_id        | UUID      | NOT NULL                 |
| title               | TEXT      | NOT NULL                 |
| notes               | TEXT      |                          |
| status              | TEXT      | NOT NULL, DEFAULT 'pending' |
| due_on              | DATE      | NULLABLE                 |
| rollover_enabled    | BOOLEAN   | NOT NULL, DEFAULT false  |
| owner_user_id       | UUID      | NULLABLE                 |
| completed_at        | TIMESTAMP | NULLABLE                 |
| completed_by_user_id | UUID     | NULLABLE                 |
| deleted_at          | TIMESTAMP | NULLABLE                 |
| created_at          | TIMESTAMP | NOT NULL                 |
| updated_at          | TIMESTAMP | NOT NULL                 |

### reminders

| Column             | Type      | Constraints |
|--------------------|-----------|-------------|
| id                 | UUID      | PRIMARY KEY |
| tenant_id          | UUID      | NOT NULL    |
| household_id       | UUID      | NOT NULL    |
| title              | TEXT      | NOT NULL    |
| notes              | TEXT      |             |
| scheduled_for      | TIMESTAMP | NOT NULL    |
| owner_user_id      | UUID      | NULLABLE    |
| last_dismissed_at  | TIMESTAMP | NULLABLE    |
| last_snoozed_until | TIMESTAMP | NULLABLE    |
| created_at         | TIMESTAMP | NOT NULL    |
| updated_at         | TIMESTAMP | NOT NULL    |

### reminder_dismissals

| Column             | Type      | Constraints |
|--------------------|-----------|-------------|
| id                 | UUID      | PRIMARY KEY |
| tenant_id          | UUID      | NOT NULL    |
| household_id       | UUID      | NOT NULL    |
| reminder_id        | UUID      | NOT NULL    |
| created_by_user_id | UUID      | NOT NULL    |
| created_at         | TIMESTAMP | NOT NULL    |

### reminder_snoozes

| Column             | Type      | Constraints |
|--------------------|-----------|-------------|
| id                 | UUID      | PRIMARY KEY |
| tenant_id          | UUID      | NOT NULL    |
| household_id       | UUID      | NOT NULL    |
| reminder_id        | UUID      | NOT NULL    |
| duration_minutes   | INTEGER   | NOT NULL    |
| snoozed_until      | TIMESTAMP | NOT NULL    |
| created_by_user_id | UUID      | NOT NULL    |
| created_at         | TIMESTAMP | NOT NULL    |

### task_restorations

| Column             | Type      | Constraints |
|--------------------|-----------|-------------|
| id                 | UUID      | PRIMARY KEY |
| tenant_id          | UUID      | NOT NULL    |
| household_id       | UUID      | NOT NULL    |
| task_id            | UUID      | NOT NULL    |
| created_by_user_id | UUID      | NOT NULL    |
| created_at         | TIMESTAMP | NOT NULL    |

## Relationships

- `tasks.tenant_id` references a tenant (external).
- `tasks.household_id` references a household (external).
- `tasks.owner_user_id` references a user (external, nullable).
- `tasks.completed_by_user_id` references a user (external, nullable).
- `reminders.tenant_id` references a tenant (external).
- `reminders.household_id` references a household (external).
- `reminders.owner_user_id` references a user (external, nullable).
- `reminder_dismissals.reminder_id` references a reminder.
- `reminder_dismissals.created_by_user_id` references a user (external).
- `reminder_snoozes.reminder_id` references a reminder.
- `reminder_snoozes.created_by_user_id` references a user (external).
- `task_restorations.task_id` references a task.
- `task_restorations.created_by_user_id` references a user (external).

## Indexes

| Table                | Index Name | Columns       | Type  |
|----------------------|------------|---------------|-------|
| tasks                | (auto)     | tenant_id     | INDEX |
| tasks                | (auto)     | household_id  | INDEX |
| tasks                | (auto)     | deleted_at    | INDEX |
| reminders            | (auto)     | tenant_id     | INDEX |
| reminders            | (auto)     | household_id  | INDEX |
| reminders            | (auto)     | scheduled_for | INDEX |
| reminder_dismissals  | (auto)     | tenant_id     | INDEX |
| reminder_dismissals  | (auto)     | household_id  | INDEX |
| reminder_snoozes     | (auto)     | tenant_id     | INDEX |
| reminder_snoozes     | (auto)     | household_id  | INDEX |
| task_restorations    | (auto)     | tenant_id     | INDEX |
| task_restorations    | (auto)     | household_id  | INDEX |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: tasks, task_restorations, reminders, reminder_snoozes, reminder_dismissals.
