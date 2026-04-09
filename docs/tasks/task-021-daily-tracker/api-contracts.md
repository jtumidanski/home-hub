# Daily Tracker — API Contracts

## Base Path

`/api/v1/trackers`

---

## 1. Tracking Items

### POST `/api/v1/trackers`

Create a new tracking item.

**Request:**

```json
{
  "data": {
    "type": "trackers",
    "attributes": {
      "name": "Running",
      "scale_type": "sentiment",
      "scale_config": null,
      "schedule": [1, 2, 4, 5, 6],
      "color": "blue",
      "sort_order": 1
    }
  }
}
```

Schedule values: 0 = Sunday, 1 = Monday, ..., 6 = Saturday. Empty array `[]` means every day.

Color palette: `red`, `orange`, `amber`, `yellow`, `lime`, `green`, `emerald`, `teal`, `cyan`, `blue`, `indigo`, `violet`, `purple`, `fuchsia`, `pink`, `rose`. Required on create.

Scale config by type:
- `sentiment`: `null` (no config needed)
- `numeric`: `null` (no config needed)
- `range`: `{"min": 0, "max": 100}`

**Response (201):**

```json
{
  "data": {
    "id": "uuid",
    "type": "trackers",
    "attributes": {
      "name": "Running",
      "scale_type": "sentiment",
      "scale_config": null,
      "schedule": [1, 2, 4, 5, 6],
      "color": "blue",
      "sort_order": 1,
      "schedule_history": [
        {"schedule": [1, 2, 4, 5, 6], "effective_date": "2026-04-01"}
      ],
      "created_at": "2026-04-01T00:00:00Z",
      "updated_at": "2026-04-01T00:00:00Z"
    }
  }
}
```

**Errors:**
- `409` — name already exists for this user
- `400` — invalid scale_type, invalid schedule values, invalid color, range config missing min/max

---

### GET `/api/v1/trackers`

List all active tracking items for the current user.

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "type": "trackers",
      "attributes": {
        "name": "Running",
        "scale_type": "sentiment",
        "scale_config": null,
        "schedule": [1, 2, 4, 5, 6],
        "color": "blue",
        "sort_order": 1,
        "created_at": "2026-04-01T00:00:00Z",
        "updated_at": "2026-04-01T00:00:00Z"
      }
    }
  ]
}
```

---

### GET `/api/v1/trackers/{id}`

Get a single tracking item.

**Response (200):** Same shape as single item in list response.

**Errors:**
- `404` — item not found or not owned by user

---

### PATCH `/api/v1/trackers/{id}`

Update a tracking item. Only provided fields are updated. `scale_type` cannot be changed.

**Request:**

```json
{
  "data": {
    "type": "trackers",
    "attributes": {
      "name": "Trail Running",
      "color": "emerald",
      "schedule": [1, 2, 4, 5],
      "sort_order": 2
    }
  }
}
```

**Response (200):** Updated item.

When `schedule` is updated, a new schedule snapshot is created with `effective_date` = today. The `schedule` field in the response always reflects the current active schedule. The `schedule_history` field shows all snapshots.

**Errors:**
- `400` — attempted to change scale_type
- `404` — item not found
- `409` — name conflict

---

### DELETE `/api/v1/trackers/{id}`

Soft-delete a tracking item. Entries are preserved — the item and its entries remain visible in historical month reports for months where the item was active.

**Response:** `204 No Content`

**Errors:**
- `404` — item not found

---

## 2. Today Quick-Entry

### GET `/api/v1/trackers/today`

Get today's scheduled tracking items with their current entries (if any).

**Response (200):**

```json
{
  "data": {
    "type": "tracker-today",
    "attributes": {
      "date": "2026-04-02"
    },
    "relationships": {
      "items": {
        "data": [
          {
            "id": "uuid",
            "type": "trackers",
            "attributes": {
              "name": "Food",
              "scale_type": "sentiment",
              "scale_config": null,
              "color": "green",
              "sort_order": 1
            }
          },
          {
            "id": "uuid",
            "type": "trackers",
            "attributes": {
              "name": "Running",
              "scale_type": "sentiment",
              "scale_config": null,
              "color": "blue",
              "sort_order": 4
            }
          }
        ]
      },
      "entries": {
        "data": [
          {
            "id": "uuid",
            "type": "tracker-entries",
            "attributes": {
              "tracking_item_id": "uuid",
              "date": "2026-04-02",
              "value": {"rating": "positive"},
              "skipped": false,
              "note": "felt great today",
              "scheduled": true
            }
          }
        ]
      }
    }
  }
}
```

Only items scheduled for today (based on current schedule snapshot) are included. Entries array contains only items already logged today — missing items have no entry in the array.

---

## 3. Entries

### PUT `/api/v1/trackers/{id}/entries/{YYYY-MM-DD}`

Create or update an entry. Idempotent — creates if absent, updates if exists.

**Request (sentiment):**

```json
{
  "data": {
    "type": "tracker-entries",
    "attributes": {
      "value": {
        "rating": "positive"
      },
      "note": "felt great after the run"
    }
  }
}
```

**Request (numeric):**

```json
{
  "data": {
    "type": "tracker-entries",
    "attributes": {
      "value": {
        "count": 3
      },
      "note": "birthday party"
    }
  }
}
```

**Request (range):**

```json
{
  "data": {
    "type": "tracker-entries",
    "attributes": {
      "value": {
        "value": 72
      },
      "note": null
    }
  }
}
```

The `note` field is optional (omit or null to clear). Max 500 characters.

**Response (200 or 201):**

```json
{
  "data": {
    "id": "uuid",
    "type": "tracker-entries",
    "attributes": {
      "tracking_item_id": "uuid",
      "date": "2026-04-01",
      "value": {"rating": "positive"},
      "skipped": false,
      "note": "felt great after the run",
      "scheduled": true,
      "created_at": "2026-04-01T12:00:00Z",
      "updated_at": "2026-04-01T12:00:00Z"
    }
  }
}
```

The `scheduled` field is computed (not stored) — indicates whether this date falls on the item's schedule.

**Errors:**
- `400` — future date, value doesn't match scale type, range value out of bounds
- `404` — tracking item not found

---

### DELETE `/api/v1/trackers/{id}/entries/{YYYY-MM-DD}`

Delete an entry.

**Response:** `204 No Content`

---

### PUT `/api/v1/trackers/{id}/entries/{YYYY-MM-DD}/skip`

Mark a scheduled day as skipped. No request body required.

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "type": "tracker-entries",
    "attributes": {
      "tracking_item_id": "uuid",
      "date": "2026-04-01",
      "value": null,
      "skipped": true,
      "scheduled": true,
      "created_at": "2026-04-01T12:00:00Z",
      "updated_at": "2026-04-01T12:00:00Z"
    }
  }
}
```

**Errors:**
- `400` — date is not on the item's schedule (can only skip scheduled days)

---

### DELETE `/api/v1/trackers/{id}/entries/{YYYY-MM-DD}/skip`

Remove skip marker from a day (reverts to unfilled).

**Response:** `204 No Content`

---

### GET `/api/v1/trackers/entries?month={YYYY-MM}`

Get all entries for the current user for a given month. Returns entries across all tracking items.

**Query params:**
- `month` (required): `YYYY-MM` format

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "type": "tracker-entries",
      "attributes": {
        "tracking_item_id": "uuid",
        "date": "2026-04-01",
        "value": {"rating": "positive"},
        "skipped": false,
        "note": "felt great after the run",
        "scheduled": true,
        "created_at": "...",
        "updated_at": "..."
      }
    }
  ]
}
```

---

## 4. Monthly Views

### GET `/api/v1/trackers/months/{YYYY-MM}`

Get month summary with items, entries, and completion status.

**Response (200):**

```json
{
  "data": {
    "type": "tracker-months",
    "attributes": {
      "month": "2026-04",
      "complete": false,
      "completion": {
        "expected": 52,
        "filled": 40,
        "skipped": 3,
        "remaining": 9
      }
    },
    "relationships": {
      "items": {
        "data": [
          {
            "id": "uuid",
            "type": "trackers",
            "attributes": {
              "name": "Running",
              "scale_type": "sentiment",
              "scale_config": null,
              "color": "blue",
              "sort_order": 1,
              "active_from": "2026-04-01",
              "active_until": null,
              "schedule_snapshots": [
                {"schedule": [1, 2, 4, 5, 6], "effective_date": "2026-03-01"},
                {"schedule": [1, 2, 4, 5], "effective_date": "2026-04-15"}
              ]
            }
          }
        ]
      },
      "entries": {
        "data": [
          {
            "id": "uuid",
            "type": "tracker-entries",
            "attributes": {
              "tracking_item_id": "uuid",
              "date": "2026-04-01",
              "value": {"rating": "positive"},
              "skipped": false,
              "note": null,
              "scheduled": true
            }
          }
        ]
      }
    }
  }
}
```

`active_from` / `active_until` are derived from the item's `created_at` / `deleted_at` to determine which days in the month count as expected. `schedule_snapshots` are included so the client can determine which days were scheduled during which date ranges within the month.

---

### GET `/api/v1/trackers/months/{YYYY-MM}/report`

Get computed dashboard report for a completed month.

**Response (200):**

```json
{
  "data": {
    "type": "tracker-reports",
    "attributes": {
      "month": "2026-04",
      "summary": {
        "total_items": 6,
        "completion_rate": 0.94,
        "skip_rate": 0.06,
        "total_expected": 52,
        "total_filled": 49,
        "total_skipped": 3
      },
      "items": [
        {
          "tracking_item_id": "uuid",
          "name": "Running",
          "scale_type": "sentiment",
          "stats": {
            "expected_days": 10,
            "filled_days": 8,
            "skipped_days": 2,
            "positive": 6,
            "neutral": 1,
            "negative": 1,
            "positive_ratio": 0.75,
            "daily_values": [
              {"date": "2026-04-01", "rating": "positive"},
              {"date": "2026-04-02", "rating": "neutral"}
            ]
          }
        },
        {
          "tracking_item_id": "uuid",
          "name": "Drinks",
          "scale_type": "numeric",
          "stats": {
            "expected_days": 30,
            "filled_days": 30,
            "skipped_days": 0,
            "total": 23,
            "daily_average": 0.77,
            "days_with_entries_above_zero": 15,
            "days_with_entries_above_zero_pct": 0.50,
            "min": {"date": "2026-04-05", "count": 0},
            "max": {"date": "2026-04-12", "count": 4},
            "daily_values": [
              {"date": "2026-04-01", "count": 2},
              {"date": "2026-04-02", "count": 0}
            ]
          }
        },
        {
          "tracking_item_id": "uuid",
          "name": "Sleep Quality",
          "scale_type": "range",
          "stats": {
            "expected_days": 30,
            "filled_days": 30,
            "skipped_days": 0,
            "average": 72.3,
            "min": {"date": "2026-04-08", "value": 35},
            "max": {"date": "2026-04-20", "value": 95},
            "std_dev": 14.2,
            "daily_values": [
              {"date": "2026-04-01", "value": 72},
              {"date": "2026-04-02", "value": 68}
            ]
          }
        }
      ]
    }
  }
}
```

**Errors:**
- `400` — month is not complete
