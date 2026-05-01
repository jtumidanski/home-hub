# Data Model — Kiosk Dashboard Widgets

## Summary

This feature is almost entirely additive on the **frontend** (new widget definitions and adapters) plus a five-entry expansion of the cross-service widget allowlist. Whether it touches the database at all depends on the seeding mechanism chosen in `design.md`.

## Widget Allowlist Additions

The shared widget-type registry gains five entries.

### Go (source of truth for backend validation)

`shared/go/dashboard/types.go`:

```go
var WidgetTypes = map[string]struct{}{
    // existing...
    "weather":           {},
    "tasks-summary":     {},
    "reminders-summary": {},
    "overdue-summary":   {},
    "meal-plan-today":   {},
    "calendar-today":    {},
    "packages-summary":  {},
    "habits-today":      {},
    "workout-today":     {},
    // new in task-046:
    "tasks-today":       {},
    "reminders-today":   {},
    "weather-tomorrow":  {},
    "calendar-tomorrow": {},
    "tasks-tomorrow":    {},
}
```

### TypeScript (mirrors the Go map)

`frontend/src/lib/dashboard/widget-types.ts`:

```ts
export const WIDGET_TYPES = [
  // existing...
  "weather",
  "tasks-summary",
  "reminders-summary",
  "overdue-summary",
  "meal-plan-today",
  "calendar-today",
  "packages-summary",
  "habits-today",
  "workout-today",
  // new in task-046:
  "tasks-today",
  "reminders-today",
  "weather-tomorrow",
  "calendar-tomorrow",
  "tasks-tomorrow",
] as const;
```

### Parity Fixture

The existing `widget-types.json` parity fixture (referenced from `services/dashboard-service/docs/domain.md`) gets the same five entries appended; the parity test asserts both lists match it.

## Widget Config Schemas (Frontend Zod)

### `tasks-today`

```ts
z.object({
  title: z.string().max(80).optional(),
  includeCompleted: z.boolean().default(true),
});
```

### `reminders-today`

```ts
z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});
```

### `weather-tomorrow`

```ts
z.object({
  units: z.enum(["imperial", "metric"]).nullable().default(null),
});
```

The `null` default means "inherit from household preference" — match whatever pattern the existing `weather` widget uses; if it stores `units: "imperial"` literally, mirror that.

### `calendar-tomorrow`

```ts
z.object({
  includeAllDay: z.boolean().default(true),
  limit: z.number().int().min(1).max(10).default(5),
});
```

### `tasks-tomorrow`

```ts
z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});
```

### `meal-plan-today` (extension)

Existing schema gains a `view` field:

```ts
z.object({
  horizonDays: z.union([z.literal(1), z.literal(3), z.literal(7)]).default(1),
  view: z.enum(["list", "today-detail"]).default("list"),  // NEW
});
```

Existing layouts persisted before this change have no `view` field; Zod's default makes them parse as `"list"`. No migration of stored configs required.

## Possible Schema Change (Mechanism 1 from §4.9.2)

Only if `design.md` selects mechanism (1):

### `dashboard.dashboards` — add `seed_key` column

```sql
ALTER TABLE dashboards
  ADD COLUMN seed_key VARCHAR(40);

CREATE UNIQUE INDEX idx_dashboards_seed_key
  ON dashboards (tenant_id, household_id, seed_key)
  WHERE seed_key IS NOT NULL;
```

| Column | Type | Nullable | Notes |
|---|---|---|---|
| `seed_key` | `VARCHAR(40)` | yes | `NULL` for user-created dashboards; populated only by the seed flow. Format: `^[a-z][a-z0-9-]{0,39}$`. |

Migration is additive — existing rows have `seed_key = NULL` and are unaffected.

The partial unique index ensures one row per `(tenant, household, key)` for seeded dashboards while leaving user-created ones unconstrained (multiple "Kiosk"-named user-created dashboards are still allowed; only the seeded one occupies the `seed_key='kiosk'` slot).

The GORM `Entity` gains:

```go
type Entity struct {
    // existing...
    SeedKey *string `gorm:"column:seed_key;type:varchar(40)"`
}
```

The seed processor sets it on insert; nothing else touches it. `DELETE` is unaffected — deleting a seeded dashboard removes the row and its `seed_key`, freeing the slot for a future re-seed if a user explicitly re-runs seeding (which is not exposed in this feature, but the data model permits it).

## Possible Preference Addition (Mechanism 2 from §4.9.2)

Only if `design.md` selects mechanism (2):

| Service | Preference key | Type | Default | Scope |
|---|---|---|---|---|
| `account-service` | `kiosk_dashboard_seeded` | boolean | `false` | per `(tenant_id, user_id, household_id)` |

No schema migration — the preferences store is already a key/value bag. The frontend treats unset and `false` identically.

## Seeded Layout Fixtures

### `frontend/src/lib/dashboard/seed-layout.ts` (existing)

Unchanged in this feature. Continues to produce the "Home" layout with the original nine widgets.

### `frontend/src/lib/dashboard/kiosk-seed-layout.ts` (new)

Mirrors `seedLayout()` shape; produces the four-column kiosk-style layout described in PRD §4.9.1. Each invocation generates fresh widget UUIDs.

```ts
export function kioskSeedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      // Column 1: weather + meals (today-detail) + tasks-today
      { id: uuid(), type: "weather",          x: 0, y: 0, w: 3, h: 4, config: { units: "imperial", location: null } },
      { id: uuid(), type: "meal-plan-today",  x: 0, y: 4, w: 3, h: 4, config: { horizonDays: 3, view: "today-detail" } },
      { id: uuid(), type: "tasks-today",      x: 0, y: 8, w: 3, h: 4, config: {} },
      // Column 2: calendar-today, tall
      { id: uuid(), type: "calendar-today",   x: 3, y: 0, w: 3, h: 12, config: { horizonDays: 1, includeAllDay: true } },
      // Column 3: tomorrow stack
      { id: uuid(), type: "weather-tomorrow", x: 6, y: 0, w: 3, h: 3,  config: {} },
      { id: uuid(), type: "calendar-tomorrow",x: 6, y: 3, w: 3, h: 5,  config: {} },
      { id: uuid(), type: "tasks-tomorrow",   x: 6, y: 8, w: 3, h: 4,  config: {} },
      // Column 4: reminders-today, tall
      { id: uuid(), type: "reminders-today",  x: 9, y: 0, w: 3, h: 12, config: {} },
    ],
  };
}
```

Heights are illustrative — final tuning is done in design (PRD §9 question 2). Widget instance counts and `w`/`x` columns are stable.

## Validation

- All five new widget types pass `IsKnownWidgetType` once added.
- Per-widget config payloads: each schema's serialized JSON stays well under the 4 KB cap.
- Total kiosk seed-layout payload: 8 widgets × small configs ≈ a couple hundred bytes, well under the 64 KB cap.
- Grid bounds: every widget satisfies `x + w ≤ 12`.
