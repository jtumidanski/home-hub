# Dashboard Designer — Data Model

## 1. Table: `dashboard.dashboards`

GORM entity (sketch — follow `services/*/internal/*/entity.go` conventions):

```go
type Entity struct {
    ID             uuid.UUID       `gorm:"type:uuid;primaryKey"`
    TenantID       uuid.UUID       `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
    HouseholdID    uuid.UUID       `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
    UserID         *uuid.UUID      `gorm:"type:uuid;index:idx_dashboards_scope"`
    Name           string          `gorm:"size:80;not null"`
    SortOrder      int             `gorm:"not null;default:0"`
    Layout         datatypes.JSON  `gorm:"type:jsonb;not null"`
    SchemaVersion  int             `gorm:"not null;default:1"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

Indexes:
- `idx_dashboards_scope` on `(tenant_id, household_id, user_id)` — covers every list query.
- Partial index on `(tenant_id, household_id) WHERE user_id IS NULL` for the household list fast-path.

Constraints:
- `name` length validated in application (1–80 after trim).
- `layout` size enforced in application; Postgres jsonb can hold larger but we cap at 64 KB.

## 2. Layout JSON schema (version 1)

```
{
  "version": 1,
  "widgets": [
    {
      "id":     "<uuid>",
      "type":   "<registry-id>",
      "x":      <int>=0>,
      "y":      <int>=0>,
      "w":      <int>=1>,
      "h":      <int>=1>,
      "config": { /* widget-specific, opaque to backend */ }
    }
  ]
}
```

Backend validates:
- `version == 1`
- `widgets` length 0–40
- `id` is a UUID, unique within this layout
- `type` in the registry allowlist (§4 below)
- `x,y,w,h` integers with `x>=0, y>=0, w>=1, h>=1, x+w<=12`
- `config` is an object, serialized size ≤ 4 KB, max nesting depth 5
- Full `layout` serialized ≤ 64 KB

Backend does NOT validate:
- `config` shape (frontend owns that via Zod)
- Overlap between widgets (grid engine on frontend prevents, and renderer tolerates)
- Positioning optimality

## 3. Frontend widget registry

The frontend maintains a single module `frontend/src/lib/dashboard/widget-registry.ts` that exports an array of:

```ts
type WidgetDefinition<TConfig> = {
  type: string;
  displayName: string;
  description: string;
  component: React.ComponentType<{ config: TConfig }>;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize: { w: number; h: number };
  // dataScope is informational only — every widget fetches what it needs
  dataScope: "household" | "user";
};
```

The backend imports a string-only allowlist (see §4). Backend and frontend agree on the set of `type` ids through a shared constant file (can live in `shared/go/dashboard` for Go-side and be hand-mirrored in TS — the set is small enough that a plain file is fine).

## 4. Widget registry (v1)

| `type` | Data scope | `defaultW × defaultH` | `minW × minH` | `maxW × maxH` | Config (v1) |
|---|---|---|---|---|---|
| `weather` | household | 12×3 | 6×2 | 12×4 | `{ units: "imperial" \| "metric", location?: { lat: number, lon: number, label: string } \| null }` |
| `tasks-summary` | household | 4×2 | 3×2 | 6×2 | `{ status: "pending" \| "overdue" \| "completed", title?: string }` |
| `reminders-summary` | household | 4×2 | 3×2 | 6×2 | `{ filter: "active" \| "snoozed" \| "upcoming", title?: string }` |
| `overdue-summary` | household | 4×2 | 3×2 | 6×2 | `{ title?: string }` |
| `meal-plan-today` | household | 4×3 | 3×3 | 6×5 | `{ horizonDays: 1 \| 3 \| 7 }` |
| `calendar-today` | household | 6×3 | 4×3 | 12×6 | `{ horizonDays: 1 \| 3 \| 7, includeAllDay: boolean }` |
| `packages-summary` | household | 4×3 | 3×2 | 6×4 | `{ title?: string }` |
| `habits-today` | user | 4×3 | 3×3 | 6×5 | `{ title?: string }` |
| `workout-today` | user | 4×3 | 3×3 | 6×5 | `{ title?: string }` |

All title fields are ≤ 80 chars. `location.label` ≤ 200 chars. Strings must be plain text (frontend sanitizes on input; renderer does not `dangerouslySetInnerHTML`).

## 5. Default-dashboard preference

Stored in `account-service`'s existing preferences store:

- Key: `default_dashboard_id`
- Type: string (uuid) or null
- Scope: per `(tenant_id, user_id, household_id)` — same scope that preferences already use

Resolution (frontend-only logic):
1. Read preference.
2. If set AND `GET /api/v1/dashboards/{id}` succeeds, use it.
3. Else use the first household-scoped dashboard visible to the caller (by `sort_order`, then `created_at`).
4. Else use the first user-scoped dashboard.
5. Else trigger the seeding flow and use the returned dashboard.

Clearing on delete:
- When the user deletes their currently-preferred dashboard, the frontend clears the preference in the same interaction.
- Cross-user deletes (household dashboard deleted by another member) are handled by fallback at read time — no back-references to clean up.

## 6. Seed template

The seed template (sent by the frontend in `POST /dashboards/seed`) replicates the current `DashboardPage`:

Row 1 (weather banner, full width):
- `weather` @ `(x=0, y=0, w=12, h=3)`, default units

Row 2 (tasks / reminders / overdue, 3 columns):
- `tasks-summary` @ `(0, 3, 4, 2)` config `{ status: "pending", title: "Pending Tasks" }`
- `reminders-summary` @ `(4, 3, 4, 2)` config `{ filter: "active", title: "Active Reminders" }`
- `overdue-summary` @ `(8, 3, 4, 2)` config `{ title: "Overdue" }`

Row 3 (meals / habits / packages, 3 columns, equal height):
- `meal-plan-today` @ `(0, 5, 4, 3)` config `{ horizonDays: 1 }`
- `habits-today` @ `(4, 5, 4, 3)` config `{}`
- `packages-summary` @ `(8, 5, 4, 3)` config `{}`

Row 4 (calendar / workout, 2 columns):
- `calendar-today` @ `(0, 8, 6, 3)` config `{ horizonDays: 1, includeAllDay: true }`
- `workout-today` @ `(6, 8, 6, 3)` config `{}`

This is a total of 9 widgets, within the cap. Cells are chosen to make the parity obvious; the designer is free to change them later.
