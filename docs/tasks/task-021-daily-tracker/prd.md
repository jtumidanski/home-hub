# Daily Tracker — Product Requirements Document

Version: v2
Status: Draft
Created: 2026-04-02
Updated: 2026-04-02
---

## 1. Overview

Daily Tracker is a personal habit and wellness tracking feature that gives each user a customizable monthly calendar view for logging daily metrics. Users define their own tracking items — each with a name, rating scale, and weekly schedule — then fill in entries day-by-day throughout the month.

The feature serves users who want a lightweight, structured way to monitor recurring habits (exercise, nutrition, sleep, alcohol intake, etc.) without the overhead of a full journaling or fitness app. The calendar-grid format makes it easy to see patterns at a glance and stay accountable.

When all expected entries for a month are filled in (or explicitly skipped), the month transitions to a completed state and displays a read-only dashboard with aggregate statistics, ratios, and visualizations.

## 2. Goals

Primary goals:
- Let users define personalized tracking items with flexible rating scales and per-day-of-week schedules
- Provide an intuitive monthly calendar grid for logging and editing entries
- Allow backfilling of past months so forgetting a day doesn't lock the user out
- Generate a meaningful monthly dashboard report when all entries are complete
- Keep the data model user-scoped (not household-scoped) for personal use

Non-goals:
- Sharing or comparing trackers between household members (future)
- Multi-month trend reports or historical comparisons (future)
- Push notifications or reminders to log entries (future)
- Import/export of tracker data (future)
- Streaks, gamification, or achievement badges (future)

## 3. User Stories

- As a user, I want to create a tracking item (e.g., "Running") with a specific scale and weekly schedule so that I can define what I track and when
- As a user, I want to edit or delete my tracking items so that I can adjust my tracking setup over time
- As a user, I want to see a monthly calendar grid showing all my tracking items and their entries so that I can log my day at a glance
- As a user, I want to tap a cell in the calendar to log or update an entry for a specific item on a specific day
- As a user, I want to skip a scheduled day (e.g., rest day exception) so that it doesn't block month completion
- As a user, I want to log an entry on an unscheduled day (one-off exception) so that I can capture activity that happened outside my normal routine
- As a user, I want to backfill entries for any past month that still has missing entries
- As a user, I want to see a dashboard report for any completed month with aggregate stats and visualizations
- As a user, I want to see at a glance which months are complete and which still need entries
- As a user, I want a "today" quick-entry view showing only today's scheduled items so I can log my day fast
- As a user, I want to add an optional note to any entry for context (e.g., "legs felt heavy", "birthday party")

## 4. Functional Requirements

### 4.1 Tracking Item Management

- Users can create tracking items with the following properties:
  - **Name**: free-text, max 100 characters, unique per user
  - **Scale type**: one of `sentiment`, `numeric`, `range`
  - **Scale config** (varies by type):
    - `sentiment`: 3-point scale (positive / neutral / negative), no additional config
    - `numeric`: unbounded non-negative integer counter, no additional config
    - `range`: min and max values (e.g., 0–100), defined at creation
  - **Schedule**: set of days-of-week when this item is expected (e.g., Mon, Tue, Thu, Fri, Sat). Empty schedule means "every day"
  - **Color**: selected from a fixed palette, used in the calendar grid and dashboard charts
  - **Sort order**: integer for display ordering in the calendar grid
- Users can update name, color, schedule, sort order, and scale config of existing items
- Users can delete tracking items; this soft-deletes the item but preserves entries for historical reports
- Scale type cannot be changed after creation (would invalidate existing entries)
- When a tracking item's schedule is updated, a new schedule snapshot is created effective from the current date; previous months retain their original schedule for completion calculations

### 4.2 Entry Logging

- Users can create, update, and delete entries for any tracking item on any date
- Entry value must conform to the item's scale type:
  - `sentiment`: one of `positive`, `neutral`, `negative`
  - `numeric`: non-negative integer
  - `range`: integer within the item's configured min–max
- Entries on scheduled days are "expected"; entries on unscheduled days are "bonus" (one-off exceptions)
- Users can mark a scheduled day as **skipped** — this counts as fulfilled for completion purposes
  - A skipped entry stores no value, just the skipped flag
- Entries can be created or modified for any date in the past or present; future dates are not allowed
- Entries may include an optional **note** (free-text, max 500 characters) for context

### 4.3 Monthly Calendar View

- Default view shows the current month
- Users can navigate to any past month
- The calendar grid displays:
  - Rows: one per tracking item (sorted by sort order)
  - Columns: one per day of the month
  - Cells: show the entry value (icon/number), empty if not yet logged, dimmed if not a scheduled day
  - Scheduled but unfilled cells are visually distinct (highlighted / outlined)
  - Skipped cells show a skip indicator
- Tapping a cell opens an inline editor or modal to log/update/skip the entry
- The view header shows month completion progress (e.g., "47/52 entries filled")

### 4.4 Today Quick-Entry

- Dedicated view showing only today's scheduled tracking items in a vertical list
- Each item shows its name, scale type input, and optional note field
- Items already logged today show their current value (editable)
- Unfilled items are visually prominent
- Accessible from the tracker landing page (e.g., "Log Today" button) or as the default mobile view
- Does not replace the full calendar grid — it's a convenience shortcut

### 4.5 Month Completion

- A month is **complete** when every expected entry is either filled or skipped, and the month has no remaining scheduled days in the future
- Expected entries = for each active tracking item, the days in the month matching its **snapshotted schedule** for that month
- Items created mid-month: only days from the creation date onward count as expected
- Items deleted mid-month: only days up to the deletion date count as expected; deleted items still appear in historical reports for months where they were active
- Completion is computed dynamically, not stored — recalculated on each view
- Past months with missing entries remain editable indefinitely

### 4.6 Schedule Snapshots

- Each tracking item's schedule is versioned via snapshots
- A snapshot records the schedule (days-of-week) and an `effective_date`
- When a tracking item is created, its initial schedule becomes the first snapshot
- When a user updates the schedule, a new snapshot is created with `effective_date` = today; previous snapshots are preserved
- Month completion and report calculations use the snapshot that was active during each day of the month
- For a given date, the applicable schedule is the snapshot with the latest `effective_date` on or before that date

### 4.7 Monthly Dashboard Report

- Available for any completed month
- Displayed automatically when viewing a completed month (replaces the calendar grid)
- Users can toggle back to the calendar view to see/edit individual entries even on a completed month
- Report widgets (computed server-side):

  **General:**
  - Overall completion rate (filled entries / expected entries, excluding skips)
  - Skip rate (skipped / expected)
  - Total tracking items active during the month

  **Per sentiment item:**
  - Positive / neutral / negative counts and percentages
  - Positive-to-total ratio (e.g., "Running: 8/10 positive days")
  - Calendar heatmap (color-coded by sentiment)

  **Per numeric item:**
  - Total sum for the month (e.g., "Total drinks: 23")
  - Daily average
  - Percentage of days with entries > 0
  - Min / max day
  - Bar chart of daily values

  **Per range item:**
  - Monthly average (e.g., "Avg sleep quality: 72")
  - Min / max / standard deviation
  - Trend line (daily values across the month)

## 5. API Surface

Base path: `/api/v1/trackers`

### 5.1 Tracking Items

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/trackers` | Create tracking item |
| GET | `/api/v1/trackers` | List all tracking items for current user |
| GET | `/api/v1/trackers/{id}` | Get single tracking item |
| PATCH | `/api/v1/trackers/{id}` | Update tracking item |
| DELETE | `/api/v1/trackers/{id}` | Soft-delete tracking item |

### 5.2 Today Quick-Entry

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/trackers/today` | Get today's scheduled items with current entries |

### 5.3 Entries

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/v1/trackers/{id}/entries/{date}` | Create or update entry for item on date |
| DELETE | `/api/v1/trackers/{id}/entries/{date}` | Delete entry |
| PUT | `/api/v1/trackers/{id}/entries/{date}/skip` | Mark scheduled day as skipped |
| DELETE | `/api/v1/trackers/{id}/entries/{date}/skip` | Remove skip marker |
| GET | `/api/v1/trackers/entries?month={YYYY-MM}` | Get all entries for user for a month |

Date format: `YYYY-MM-DD`

### 5.4 Monthly Views

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/trackers/months/{YYYY-MM}` | Get month summary: items, entries, completion status |
| GET | `/api/v1/trackers/months/{YYYY-MM}/report` | Get computed report/dashboard data for a completed month |

The month summary endpoint returns:
- All active tracking items for the month
- All entries for the month
- Completion stats (expected, filled, skipped, remaining)
- `complete` boolean flag

The report endpoint returns 400 if the month is not complete.

### 5.5 Error Cases

| Status | Condition |
|--------|-----------|
| 400 | Invalid scale value for item type |
| 400 | Future date on entry |
| 400 | Report requested for incomplete month |
| 404 | Tracking item not found or not owned by user |
| 409 | Duplicate tracking item name for user |

## 6. Data Model

### Schema: `tracker`

#### `tracker.tracking_items`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, indexed |
| user_id | UUID | NOT NULL, indexed |
| name | VARCHAR(100) | NOT NULL |
| scale_type | VARCHAR(20) | NOT NULL, one of: sentiment, numeric, range |
| scale_config | JSONB | nullable, stores range min/max |
| color | VARCHAR(20) | NOT NULL, value from fixed palette |
| sort_order | INTEGER | NOT NULL, default 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |
| deleted_at | TIMESTAMP | nullable (soft delete) |

Unique constraint: `(tenant_id, user_id, name)` where `deleted_at IS NULL`

Note: The `schedule` column has been removed from this table. Schedules are tracked via the `schedule_snapshots` table to preserve historical accuracy.

#### `tracker.schedule_snapshots`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tracking_item_id | UUID | NOT NULL, FK → tracking_items.id |
| schedule | JSONB | NOT NULL, array of day-of-week integers (0=Sun, 6=Sat), empty array = every day |
| effective_date | DATE | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |

Unique constraint: `(tracking_item_id, effective_date)`

Index: `(tracking_item_id, effective_date)` for lookups

When a tracking item is created, an initial snapshot is inserted with `effective_date` = creation date. When the schedule is updated via PATCH, a new snapshot is inserted with `effective_date` = today. To determine the schedule for any given date, find the snapshot with the latest `effective_date <= date`.

#### `tracker.tracking_entries`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, indexed |
| user_id | UUID | NOT NULL, indexed |
| tracking_item_id | UUID | NOT NULL, FK → tracking_items.id |
| date | DATE | NOT NULL |
| value | JSONB | nullable (null when skipped) |
| skipped | BOOLEAN | NOT NULL, default false |
| note | VARCHAR(500) | nullable |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Unique constraint: `(tracking_item_id, date)`

Index: `(tenant_id, user_id, date)` for month queries

The `value` field stores scale-appropriate data:
- sentiment: `{"rating": "positive"}` / `{"rating": "neutral"}` / `{"rating": "negative"}`
- numeric: `{"count": 3}`
- range: `{"value": 72}`

## 7. Service Impact

| Service | Impact |
|---------|--------|
| **tracker-service** (new) | New Go microservice following standard service pattern. Owns `tracker` schema. Domains: tracking item management, entry logging, month computation, report generation. |
| **frontend** | New pages: tracking item setup, monthly calendar grid, monthly dashboard report. New sidebar entry. |
| **Ingress / nginx** | New route: `/api/v1/trackers` → tracker-service |
| **Docker Compose** | New service entry for tracker-service |
| **Kubernetes** | New deployment, service, ingress rule for tracker-service |
| **CI/CD** | New build/test/publish pipeline for tracker-service |

No cross-service API calls required. Tracker-service is fully self-contained.

## 8. Non-Functional Requirements

- **Performance**: Month summary endpoint must return in < 200ms for a user with up to 20 tracking items
- **Data limits**: Max 50 tracking items per user; no limit on entries (bounded by items x days)
- **Security**: All endpoints require valid JWT. All queries scoped by tenant_id + user_id extracted from JWT claims. No user can access another user's tracking data.
- **Observability**: Standard structured logging with request_id, user_id, tenant_id. OpenTelemetry tracing.
- **Persistence**: GORM with AutoMigrate. UUID primary keys. PostgreSQL.
- **Multi-tenancy**: Data scoped by tenant_id and user_id (not household_id). Tenant callbacks applied per standard shared module pattern.

## 9. Open Questions

1. Should the report endpoint allow partial/incomplete months with a warning, or strictly require completion? (Current spec: strictly requires completion)
2. Should the calendar grid support week-start preference (Monday vs Sunday)? (Current spec: deferred to frontend preference)

## 10. Acceptance Criteria

- [ ] User can create a tracking item with name, scale type, scale config, color, schedule, and sort order
- [ ] User can list, update, and soft-delete tracking items
- [ ] Scale type cannot be changed after creation
- [ ] User can log entries conforming to the item's scale type for any past or present date
- [ ] User can skip a scheduled day; skip counts toward completion
- [ ] User can log entries on unscheduled days (one-off exceptions)
- [ ] Month summary returns all items, entries, and correct completion status
- [ ] Month with all expected entries filled/skipped is marked complete
- [ ] Items created or deleted mid-month only count expected entries for their active range
- [ ] Report endpoint returns aggregate stats for completed months
- [ ] Report endpoint returns 400 for incomplete months
- [ ] All data is scoped by tenant_id + user_id; no cross-user access
- [ ] Frontend displays monthly calendar grid with entry editing
- [ ] Frontend displays dashboard report for completed months
- [ ] Past months with missing entries remain editable
- [ ] Schedule changes create new snapshots; past months use their historical schedule
- [ ] Soft-deleted items still appear in reports for months where they were active
- [ ] Today quick-entry view shows only today's scheduled items with current values
- [ ] Entries support optional note field (max 500 characters)
- [ ] Current month can complete once all remaining scheduled days are in the past and filled/skipped
- [ ] With zero tracking items, the calendar grid renders empty (no onboarding wizard)
- [ ] Standard observability: structured logging, tracing
