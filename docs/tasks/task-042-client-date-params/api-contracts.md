
# API Contracts

Concrete per-endpoint specs. Error bodies follow JSON:API.

## Shared error shape

```json
{
  "errors": [{
    "status": "400",
    "title": "Invalid request",
    "detail": "query parameter 'date' is required and must be YYYY-MM-DD"
  }]
}
```

`detail` is the only field that varies; it always names the offending parameter and the expected format.

---

## tracker-service

### `GET /trackers/today`

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `date` | yes | `YYYY-MM-DD` | The calendar day to fetch. |

**400 cases**
- `date` missing or empty.
- `date` doesn't match `YYYY-MM-DD`.
- `date` parses but represents an invalid calendar date (e.g. `2026-02-30`).

**Response** (unchanged from current shape; `attributes.date` echoes the input)

```json
{
  "data": {
    "type": "tracking-today",
    "id": "2026-04-16",
    "attributes": { "date": "2026-04-16" },
    "relationships": {
      "items": { "data": [...] },
      "entries": { "data": [...] }
    }
  }
}
```

### `GET /trackers/months/:month` (ripple)

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `today` | yes | `YYYY-MM-DD` | Used to compute past/future gating in the month view. |

**400 cases:** same pattern as above on `today`.

### `PUT /trackers/:item/entries/:date` (ripple — if Open Question 1a chosen)

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `today` | yes | `YYYY-MM-DD` | Used to reject future-dated entries (entry `date` > `today`). |

---

## workout-service

### `GET /workouts/today`

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `date` | yes | `YYYY-MM-DD` | The calendar day to fetch. |

**Response** (`attributes.date` echoes the input; `weekStartDate` and `dayOfWeek` derived server-side from the ISO week containing `date`)

```json
{
  "data": {
    "type": "today",
    "id": "2026-04-16",
    "attributes": {
      "date": "2026-04-16",
      "weekStartDate": "2026-04-13",
      "dayOfWeek": 3,
      "isRestDay": false,
      "items": [...]
    }
  }
}
```

---

## productivity-service

### `GET /summary/tasks`

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `date` | yes | `YYYY-MM-DD` | Defines the boundary for "completed today" and "overdue". |

**Semantics**
- `completedTodayCount` counts tasks with `completed_at >= <date>T00:00:00Z`.
- `overdueCount` counts pending tasks with `due_on < <date>`.

### `GET /summary/dashboard`

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `date` | yes | `YYYY-MM-DD` | Same semantics as `/summary/tasks`. |

### `GET /summary/reminders`

**Unchanged.** No date parameter. Reminder counts are timestamp-based (`CURRENT_TIMESTAMP`) and not affected by the timezone bug.

---

## calendar-service

### `GET /calendar/events`

**Query parameters**

| Name | Required | Format | Notes |
|---|---|---|---|
| `start` | yes | RFC 3339 | Inclusive start of the event window. |
| `end` | yes | RFC 3339 | Exclusive end of the event window. |

**400 cases**
- Either parameter missing.
- Either parameter fails `time.Parse(time.RFC3339, v)`.
- `end <= start`.

**Behavioral change from today:** the fallback to "now in resolved timezone, plus 7 days" when parameters are omitted is removed. The dashboard widget already supplies both — no frontend change beyond ensuring it continues to do so.

---

## Shared parser

New helper in `shared/go/http/params.go`:

```go
// ParseDateParam parses a YYYY-MM-DD query parameter. Returns an error that
// includes the parameter name and the raw value so handlers can render a
// useful 400. The returned time.Time is anchored to midnight UTC — callers
// use it as a calendar day, not an instant.
func ParseDateParam(r *http.Request, name string) (time.Time, error)
```

Error format (returned via `fmt.Errorf`):

```
query parameter 'date' is required and must be YYYY-MM-DD (got: "")
query parameter 'date' must be YYYY-MM-DD (got: "2026-13-40")
```

Handlers convert these errors to 400 JSON:API responses via the existing `server.WriteError` helper.
