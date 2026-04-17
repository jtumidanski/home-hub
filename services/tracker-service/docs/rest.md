# REST API

All endpoints are prefixed with `/api/v1`. All endpoints require JWT authentication. Request and response bodies use JSON:API format. Tenant and user context are derived from the JWT — there is no household scope on this service. Date parameters use `YYYY-MM-DD`; month parameters use `YYYY-MM`.

## Endpoints

### GET /api/v1/trackers

Lists active tracking items for the current user.

**Response:** JSON:API array of `trackers` resources.

| Attribute    | Type    |
|--------------|---------|
| name         | string  |
| scale_type   | string  |
| scale_config | object  |
| schedule     | int[]   |
| color        | string  |
| sort_order   | int     |
| created_at   | string  |
| updated_at   | string  |

`schedule` is the item's currently effective schedule (array of day-of-week integers, `0`=Sunday … `6`=Saturday; empty array means every day).

---

### POST /api/v1/trackers

Creates a tracking item and writes its initial schedule snapshot.

**Request:** JSON:API `trackers` resource.

| Attribute    | Type   | Required | Notes                                                              |
|--------------|--------|----------|--------------------------------------------------------------------|
| name         | string | yes      | ≤ 100 chars, unique per user (case-sensitive)                      |
| scale_type   | string | yes      | One of `sentiment`, `numeric`, `range`                             |
| scale_config | object | cond.    | Required for `range`: `{"min": int, "max": int}`, `min < max`     |
| schedule     | int[]  | no       | Day-of-week integers (0–6); empty/omitted = every day              |
| color        | string | yes      | Must be from the fixed palette                                     |
| sort_order   | int    | no       | Auto-assigned when omitted or `0`                                  |

**Response:** 201 Created. JSON:API `trackers` resource including `schedule_history`.

**Error Conditions:**

| Status | Condition                                                              |
|--------|------------------------------------------------------------------------|
| 400    | Validation failure (name length, scale type, color, range config, schedule day) |
| 409    | Duplicate name for this user                                           |

---

### GET /api/v1/trackers/{id}

Returns a single tracking item with its complete `schedule_history`.

**Response:** JSON:API `trackers` resource.

| Attribute        | Type     |
|------------------|----------|
| name             | string   |
| scale_type       | string   |
| scale_config     | object   |
| schedule         | int[]    |
| color            | string   |
| sort_order       | int      |
| schedule_history | object[] |
| created_at       | string   |
| updated_at       | string   |

Each `schedule_history` entry has `{schedule: int[], effective_date: "YYYY-MM-DD"}`.

**Error Conditions:**

| Status | Condition                       |
|--------|---------------------------------|
| 404    | Tracking item not found         |

---

### PATCH /api/v1/trackers/{id}

Partial update. Any field may be omitted; sending `schedule` creates a new snapshot effective today.

**Request:** JSON:API `trackers` resource. All attributes optional except as noted.

| Attribute    | Type    | Notes                                                                       |
|--------------|---------|-----------------------------------------------------------------------------|
| name         | string  | Subject to the same uniqueness/length rules as create                       |
| color        | string  | Must be from the palette                                                    |
| schedule     | int[]   | Creates a new schedule snapshot effective today                             |
| sort_order   | int     | Non-negative                                                                |
| scale_config | object  | For `range` items only; `min < max`                                         |
| scale_type   | string  | **Forbidden** — sending any value returns 400 (`scale type cannot be changed`) |

**Response:** JSON:API `trackers` resource with updated state and `schedule_history`.

**Error Conditions:**

| Status | Condition                                                       |
|--------|-----------------------------------------------------------------|
| 400    | Attempted scale_type change, or any other validation failure   |
| 404    | Tracking item not found                                         |
| 409    | Duplicate name for this user                                    |

---

### DELETE /api/v1/trackers/{id}

Soft-deletes a tracking item. Existing entries and historical reports continue to reference it.

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition                |
|--------|--------------------------|
| 404    | Tracking item not found |

---

### GET /api/v1/trackers/today?date=YYYY-MM-DD

Returns the items scheduled for the supplied calendar day along with any entries the user has logged on that date.

The client computes `date` from the household's local timezone and sends it as a required query parameter. The server does not infer "today" from a header or its own clock.

**Parameters:**

| Name | In    | Type   | Required | Format     |
|------|-------|--------|----------|------------|
| date | query | string | yes      | YYYY-MM-DD |

**Response:** A `tracker-today` resource with two relationships:

- `items` — array of `trackers` summaries (id, name, scale_type, scale_config, color, sort_order)
- `entries` — array of `tracker-entries` for the same items, where present

Top-level attribute: `date` (the supplied date, `YYYY-MM-DD`).

If the user has no scheduled items on the supplied date, both arrays are empty.

**Error Conditions:**

| Status | Condition                                          |
|--------|----------------------------------------------------|
| 400    | Missing, empty, or malformed `date` parameter       |

---

### PUT /api/v1/trackers/{id}/entries/{date}?today=YYYY-MM-DD

Creates or updates an entry for the given item on the given date. If a prior entry exists for the same `(id, date)`, it is updated and any prior `skipped` flag is cleared.

The `today` query parameter is required: it carries the client's local "today" so the server can reject future-dated entries without resolving a timezone.

**Parameters:**

| Name  | In    | Type   | Required | Format     |
|-------|-------|--------|----------|------------|
| today | query | string | yes      | YYYY-MM-DD |

**Request:** JSON:API `tracker-entries` resource.

| Attribute | Type   | Required | Notes                                                                       |
|-----------|--------|----------|-----------------------------------------------------------------------------|
| value     | object | yes      | Must conform to the item's `scale_type` (see below)                         |
| note      | string | no       | ≤ 500 chars                                                                 |

Value shapes:

- sentiment: `{"rating": "positive" | "neutral" | "negative"}`
- numeric: `{"count": int >= 0}`
- range: `{"value": int}` within `[scale_config.min, scale_config.max]`

**Response:** 201 Created on insert, 200 OK on update. JSON:API `tracker-entries` resource.

| Attribute        | Type    |
|------------------|---------|
| tracking_item_id | string  |
| date             | string  |
| value            | object  |
| skipped          | boolean |
| note             | string  |
| scheduled        | boolean |
| created_at       | string  |
| updated_at       | string  |

`scheduled` reflects whether `date` matches the item's schedule effective for that date.

**Error Conditions:**

| Status | Condition                                                              |
|--------|------------------------------------------------------------------------|
| 400    | Future date, malformed date, missing/invalid value, note too long      |
| 404    | Tracking item not found                                                |

---

### DELETE /api/v1/trackers/{id}/entries/{date}

Deletes the entry for the given item and date. Idempotent — returns 204 even when no entry existed.

**Response:** 204 No Content.

---

### PUT /api/v1/trackers/{id}/entries/{date}/skip?today=YYYY-MM-DD

Marks a scheduled day as skipped. The skip counts toward month completion. Only allowed on dates that fall on the item's effective schedule.

The `today` query parameter is required so the server can reject skip attempts for future dates.

**Parameters:**

| Name  | In    | Type   | Required | Format     |
|-------|-------|--------|----------|------------|
| today | query | string | yes      | YYYY-MM-DD |

**Response:** JSON:API `tracker-entries` resource with `skipped: true` and `value: null`.

**Error Conditions:**

| Status | Condition                                              |
|--------|--------------------------------------------------------|
| 400    | Future date, malformed date, or date is unscheduled    |
| 404    | Tracking item not found                                |

---

### DELETE /api/v1/trackers/{id}/entries/{date}/skip

Clears a previously-set skip flag for the given item and date. Idempotent.

**Response:** 204 No Content.

---

### GET /api/v1/trackers/entries?month=YYYY-MM

Lists all entries for the current user across all items for the given month.

**Parameters:**

| Name  | In    | Type   | Required |
|-------|-------|--------|----------|
| month | query | string | yes      |

**Response:** JSON:API array of `tracker-entries` resources (same shape as `PUT /entries/{date}`).

**Error Conditions:**

| Status | Condition                          |
|--------|------------------------------------|
| 400    | Missing or malformed month         |

---

### GET /api/v1/trackers/months/{YYYY-MM}?today=YYYY-MM-DD

Returns the month summary: active items, entries, and completion status.

The `today` query parameter is required and is used to gate past/future cells in the month view.

**Parameters:**

| Name  | In    | Type   | Required | Format     |
|-------|-------|--------|----------|------------|
| today | query | string | yes      | YYYY-MM-DD |

**Response:** A `tracker-months` resource. Top-level attributes:

| Attribute  | Type    |
|------------|---------|
| month      | string  |
| complete   | boolean |
| completion | object  |

`completion` is `{expected, filled, skipped, remaining}` (all integers).

The response also embeds:

- `items` — `MonthItemInfo` per item active in the month, including `active_from`, `active_until`, and the schedule snapshots that overlap the month.
- `entries` — all `tracker-entries` for the month.

**Error Conditions:**

| Status | Condition                            |
|--------|--------------------------------------|
| 400    | Month is not in `YYYY-MM` format     |

---

### GET /api/v1/trackers/months/{YYYY-MM}/report?today=YYYY-MM-DD

Returns the computed dashboard report for a completed month. Refuses to compute for in-progress months. Requires `today` to determine month completion against the client's local date.

**Parameters:**

| Name  | In    | Type   | Required | Format     |
|-------|-------|--------|----------|------------|
| today | query | string | yes      | YYYY-MM-DD |

**Response:** A `tracker-reports` resource.

| Attribute | Type     |
|-----------|----------|
| month     | string   |
| summary   | object   |
| items     | object[] |

`summary`: `{total_items, completion_rate, skip_rate, total_expected, total_filled, total_skipped}`

Each `items` entry: `{tracking_item_id, name, scale_type, stats}` where `stats` is one of:

- **sentiment**: `{expected_days, filled_days, skipped_days, positive, neutral, negative, positive_ratio, daily_values: [{date, rating}]}`
- **numeric**: `{expected_days, filled_days, skipped_days, total, daily_average, days_with_entries_above_zero, days_with_entries_above_zero_pct, min, max, daily_values: [{date, count}]}`
- **range**: `{expected_days, filled_days, skipped_days, average, min, max, std_dev, daily_values: [{date, value}]}`

**Error Conditions:**

| Status | Condition                                |
|--------|------------------------------------------|
| 400    | Month is not in `YYYY-MM` format         |
| 400    | Month is not yet complete                |
