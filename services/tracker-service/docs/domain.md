# Domain

## Tracking Item

### Responsibility

Manages user-defined tracking items: the named, scaled, scheduled subjects a user logs against (e.g., "Running", "Sleep quality", "Drinks"). Owns item metadata, validation, and the soft-delete lifecycle. Schedule history is delegated to the `schedule` package and looked up here when needed.

### Core Models

**Model** (`trackingitem.Model`)

| Field       | Type            |
|-------------|-----------------|
| id          | uuid.UUID       |
| tenantID    | uuid.UUID       |
| userID      | uuid.UUID       |
| name        | string          |
| scaleType   | string          |
| scaleConfig | json.RawMessage |
| color       | string          |
| sortOrder   | int             |
| createdAt   | time.Time       |
| updatedAt   | time.Time       |
| deletedAt   | *time.Time      |

All fields on Model are immutable after construction. Access is through getter methods.

### Invariants

- One active item per `(tenant_id, user_id, name)`; uniqueness is enforced where `deleted_at IS NULL`.
- `scaleType` must be one of: `sentiment`, `numeric`, `range`. Cannot be changed after creation.
- `color` must be drawn from the fixed 16-color palette (`red`, `orange`, `amber`, `yellow`, `lime`, `green`, `emerald`, `teal`, `cyan`, `blue`, `indigo`, `violet`, `purple`, `fuchsia`, `pink`, `rose`).
- `range` items require a `scaleConfig` JSON object of the form `{"min": int, "max": int}` with `min < max`.
- `sentiment` and `numeric` items have no `scaleConfig`.
- `sortOrder` is non-negative; when omitted on create it is auto-assigned to one greater than the user's current maximum.
- Soft delete sets `deletedAt` and excludes the item from active listings; historical entries and reports continue to reference it.

### Processors

**Processor** (`trackingitem.Processor`)

| Method                                                | Description                                                                  |
|-------------------------------------------------------|------------------------------------------------------------------------------|
| `List(userID)`                                        | Lists active tracking items for a user                                       |
| `Get(id)`                                             | Returns a single tracking item                                               |
| `Create(tenantID, userID, name, scaleType, ...)`      | Validates, creates the item, and writes the initial schedule snapshot        |
| `Update(id, name?, color?, schedule?, sort?, scaleConfig?)` | Partial update; a schedule change creates a new snapshot effective today |
| `Delete(id)`                                          | Soft-deletes the item                                                         |
| `GetCurrentSchedule(itemID)`                          | Returns the schedule active today for the item                                |
| `GetScheduleHistory(itemID)`                          | Returns all schedule snapshots for the item, oldest first                    |

---

## Schedule Snapshot

### Responsibility

Versions a tracking item's weekly schedule so historical month calculations remain accurate after a user changes which days an item is expected.

### Core Models

**Model** (`schedule.Model`)

| Field          | Type      |
|----------------|-----------|
| id             | uuid.UUID |
| trackingItemID | uuid.UUID |
| schedule       | []int     |
| effectiveDate  | time.Time |
| createdAt      | time.Time |

### Invariants

- `schedule` is an array of day-of-week integers (0=Sunday … 6=Saturday). An empty array means "every day".
- Schedule day values must each be in `[0,6]`.
- `(tracking_item_id, effective_date)` is unique — a single item cannot have two snapshots on the same date.
- The effective schedule for any date `D` is the snapshot with the latest `effective_date <= D`.
- An initial snapshot is written when a tracking item is created; subsequent snapshots are written on schedule change with `effective_date = today (UTC)`.

### Processors

| Method                                          | Description                                                       |
|-------------------------------------------------|-------------------------------------------------------------------|
| `GetByTrackingItemID(itemID)`                   | Lists all snapshots for an item                                   |
| `GetByTrackingItemIDs(itemIDs)`                 | Bulk-loads snapshots for multiple items (used by month summary)  |
| `GetEffectiveSchedule(itemID, date)`            | Returns the snapshot active for the given date                    |
| `CreateSnapshot(tx, itemID, schedule, date)`    | Creates a new snapshot inside a transaction                       |

---

## Entry

### Responsibility

Stores the actual logged values: one entry per tracking item per date. Enforces value validation against the item's scale type, supports the "skipped" state for scheduled rest days, and prevents future-date entries.

### Core Models

**Model** (`entry.Model`)

| Field          | Type            |
|----------------|-----------------|
| id             | uuid.UUID       |
| tenantID       | uuid.UUID       |
| userID         | uuid.UUID       |
| trackingItemID | uuid.UUID       |
| date           | time.Time       |
| value          | json.RawMessage |
| skipped        | bool            |
| note           | *string         |
| createdAt      | time.Time       |
| updatedAt      | time.Time       |

### Invariants

- `(tracking_item_id, date)` is unique — at most one entry per item per day.
- `date` must not be in the future (relative to UTC midnight).
- `note`, when set, must be ≤ 500 characters.
- A non-skipped entry must have a `value` that conforms to the parent item's `scaleType`:
  - `sentiment`: `{"rating": "positive" | "neutral" | "negative"}`
  - `numeric`: `{"count": int >= 0}`
  - `range`: `{"value": int}` where `value` ∈ `[scaleConfig.min, scaleConfig.max]`
- `skipped == true` clears `value` and `note` and is only allowed on dates that fall on the item's effective schedule for that day.

### Processors

**Processor** (`entry.Processor`)

| Method                                                                             | Description                                                                       |
|------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------|
| `CreateOrUpdate(tenantID, userID, itemID, date, value, note, scaleType, scaleConfig)` | Upserts an entry; clears any prior `skipped` flag                              |
| `Delete(itemID, date)`                                                             | Removes the entry for the given item and date                                     |
| `Skip(tenantID, userID, itemID, date, isScheduled)`                                | Marks a scheduled day as skipped; rejects unscheduled days                        |
| `RemoveSkip(itemID, date)`                                                         | Removes a previously-set skip flag                                                |
| `ListByMonth(userID, monthYYYYMM)`                                                 | Returns all entries for a user across all items for the given month               |

---

## Month

### Responsibility

Computes derived month-level data on demand: completion status, summary stats, and per-item dashboard reports. Reads from `trackingitem`, `schedule`, and `entry` and never persists its own state.

### Core Models

- `MonthSummary` — `{month, complete, completion: {expected, filled, skipped, remaining}}`
- `Report` — month-level aggregate plus a per-item `ItemReport` with scale-specific stats (`SentimentStats`, `NumericStats`, `RangeStats`).

### Invariants

- "Active items for a month" includes soft-deleted items whose `(createdAt, deletedAt)` window overlaps the month, so historical reports stay accurate.
- Expected days for an item are computed by walking each day in the item's active range within the month, looking up the applicable schedule snapshot, and counting days whose weekday matches the snapshot (or all days when the snapshot's schedule is empty).
- A month is `complete` when `expected == filled + skipped` and the month has no remaining scheduled days in the future relative to today.
- The report endpoint refuses to compute when `complete == false` (returns `ErrMonthIncomplete`).
- Numeric stats: `total`, `daily_average`, days with `count > 0`, plus min/max day and full daily series.
- Sentiment stats: positive/neutral/negative counts and `positive_ratio = positive / (positive+neutral+negative)`.
- Range stats: average, min/max day, and population standard deviation.

### Processors

**Processor** (`month.Processor`)

| Method                                  | Description                                                                |
|-----------------------------------------|----------------------------------------------------------------------------|
| `ComputeMonthSummary(userID, monthStr)` | Returns the summary plus the active items and entries used to compute it  |
| `ComputeReport(userID, monthStr)`       | Returns the dashboard report; errors with `ErrMonthIncomplete` if not done |

---

## Today

### Responsibility

Convenience read-side view that returns the items scheduled for the current day along with any entries already logged today. Pure composition over `trackingitem`, `schedule`, and `entry` — no persistence of its own.

### Behavior

- Uses the user's current effective schedule (snapshot with `effective_date <= today`) for each active item.
- Items with an empty schedule are always included.
- Entries are returned only for items in the scheduled set.
- Date is computed in UTC and truncated to midnight.
