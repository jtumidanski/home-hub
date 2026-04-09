# Workout Tracker — API Contracts

All endpoints under `/api/v1/workouts/...`. Served by `workout-service`. JSON:API style. Tenant and user are extracted from the JWT — no path or query parameters carry them.

Standard error envelope: `{ "errors": [ { "status": "400", "code": "...", "title": "...", "detail": "..." } ] }`.

---

## 1. Themes

### 1.1 List themes

`GET /api/v1/workouts/themes`

Response `200`:

```json
{
  "data": [
    {
      "type": "themes",
      "id": "9c3...",
      "attributes": {
        "name": "Muscle",
        "sortOrder": 0,
        "createdAt": "2026-04-09T12:00:00Z",
        "updatedAt": "2026-04-09T12:00:00Z"
      }
    }
  ]
}
```

### 1.2 Create theme

`POST /api/v1/workouts/themes`

```json
{
  "data": {
    "type": "themes",
    "attributes": { "name": "Mobility", "sortOrder": 2 }
  }
}
```

`201` returns the created theme. `409` if name already exists for this user.

### 1.3 Update theme

`PATCH /api/v1/workouts/themes/{id}`

Body: partial `attributes` (`name`, `sortOrder`). `200` returns the updated theme.

### 1.4 Delete theme

`DELETE /api/v1/workouts/themes/{id}` — soft delete. `204`.

---

## 2. Regions

Same shape as themes. Endpoints: `GET|POST /api/v1/workouts/regions`, `PATCH|DELETE /api/v1/workouts/regions/{id}`.

---

## 3. Exercises

### 3.1 List exercises

`GET /api/v1/workouts/exercises?themeId={id}&regionId={id}`

Both query params optional. Returns active exercises only. `regionId` matches if the region is the primary OR appears in `secondaryRegionIds`.

```json
{
  "data": [
    {
      "type": "exercises",
      "id": "f4a...",
      "attributes": {
        "name": "Machine Chest Press",
        "kind": "strength",
        "weightType": "free",
        "themeId": "9c3...",
        "regionId": "1b8...",
        "secondaryRegionIds": ["2c4...", "3d5..."],
        "defaults": {
          "sets": 3,
          "reps": 10,
          "weight": 100,
          "weightUnit": "lb"
        },
        "notes": null,
        "createdAt": "...",
        "updatedAt": "..."
      }
    }
  ]
}
```

`defaults` shape varies by `kind`:

- `strength`: `{ sets, reps, weight, weightUnit }`
- `isometric`: `{ sets, durationSeconds, weight?, weightUnit? }` — no reps; weight is optional (weighted planks)
- `cardio`: `{ durationSeconds, distance, distanceUnit }`

Any field inside `defaults` may be `null` or omitted.

For `weightType: "bodyweight"` strength items, `weight` represents *added* weight only (e.g., dip belt). Volume math contributes `sets × reps` only.

### 3.2 Create exercise

`POST /api/v1/workouts/exercises`

```json
{
  "data": {
    "type": "exercises",
    "attributes": {
      "name": "Machine Chest Press",
      "kind": "strength",
      "weightType": "free",
      "themeId": "9c3...",
      "regionId": "1b8...",
      "secondaryRegionIds": ["2c4...", "3d5..."],
      "defaults": { "sets": 3, "reps": 10, "weight": 100, "weightUnit": "lb" },
      "notes": "Seat height 4"
    }
  }
}
```

`weightType` defaults to `free` if omitted. `secondaryRegionIds` defaults to `[]` if omitted.

`201` returns the created exercise. Errors:
- `400` invalid defaults; primary `regionId` also appears in `secondaryRegionIds`
- `409` duplicate name
- `404` theme, primary region, or any secondary region not found

### 3.3 Update exercise

`PATCH /api/v1/workouts/exercises/{id}` — partial. `kind` and `weightType` cannot change; `422` if attempted. `regionId` and `secondaryRegionIds` are mutable.

### 3.4 Delete exercise

`DELETE /api/v1/workouts/exercises/{id}` — soft delete. `204`.

---

## 4. Weeks

### 4.1 Get week

`GET /api/v1/workouts/weeks/{weekStart}`

`weekStart` is any `YYYY-MM-DD` date. The server normalizes to the Monday of the ISO week containing that date. Clients may pass any day-of-week.

Returns `404 Not Found` when no week row exists for the user. The frontend interprets this as "show the empty-week prompt."

For a populated week:

```json
{
  "data": {
    "type": "weeks",
    "id": "8aa...",
    "attributes": {
      "weekStartDate": "2026-04-06",
      "restDayFlags": [5, 6],
      "items": [
        {
          "id": "p1...",
          "dayOfWeek": 0,
          "position": 0,
          "exerciseId": "f4a...",
          "exerciseName": "Machine Chest Press",
          "exerciseDeleted": false,
          "kind": "strength",
          "weightType": "free",
          "planned": {
            "sets": 3,
            "reps": 10,
            "weight": 100,
            "weightUnit": "lb"
          },
          "performance": {
            "status": "partial",
            "mode": "summary",
            "weightUnit": "lb",
            "actuals": {
              "sets": 2,
              "reps": 10,
              "weight": 100
            },
            "sets": null,
            "notes": null,
            "updatedAt": "2026-04-06T18:32:00Z"
          },
          "notes": null
        }
      ]
    }
  }
}
```

When `mode == "per_set"`, `actuals` is `null` and `sets` contains an array (each set inherits `weightUnit` from the parent performance):

```json
"sets": [
  { "setNumber": 1, "reps": 10, "weight": 100 },
  { "setNumber": 2, "reps": 8,  "weight": 105 },
  { "setNumber": 3, "reps": 6,  "weight": 110 }
]
```

Item shape per `kind`:

- `strength`: `planned` has `sets/reps/weight/weightUnit`; `actuals` has `sets/reps/weight`
- `isometric`: `planned` has `sets/durationSeconds`, optional `weight/weightUnit`; `actuals` has `sets/durationSeconds`, optional `weight`
- `cardio`: `planned` has `durationSeconds/distance/distanceUnit`; `actuals` has `durationSeconds/distance/distanceUnit`

### 4.1a Patch week

`PATCH /api/v1/workouts/weeks/{weekStart}`

```json
{ "data": { "attributes": { "restDayFlags": [5, 6] } } }
```

Lazily creates the week row if it doesn't exist. `200` returns the full week resource. Currently only `restDayFlags` is patchable.

### 4.2 Copy from previous week

`POST /api/v1/workouts/weeks/{weekStart}/copy`

```json
{ "data": { "attributes": { "mode": "planned" } } }
```

`mode` is `"planned"` or `"actual"`.

- `201` — week created and populated; returns the populated week resource (same shape as 4.1)
- `404` — no prior week with planned items exists for this user
- `409` — target week already has planned items

### 4.3 Add planned item

`POST /api/v1/workouts/weeks/{weekStart}/items`

```json
{
  "data": {
    "type": "planned-items",
    "attributes": {
      "exerciseId": "f4a...",
      "dayOfWeek": 0,
      "position": 0,
      "planned": { "sets": 3, "reps": 10, "weight": 100, "weightUnit": "lb" },
      "notes": null
    }
  }
}
```

If `planned` is omitted, defaults are seeded from the exercise. If `position` is omitted, it is assigned to `(max position for that day) + 1`. Creates the week row lazily if needed.

`201` returns the created planned item (same shape as inside `weeks.attributes.items`). `422` if exercise is soft-deleted.

### 4.3a Bulk add planned items

`POST /api/v1/workouts/weeks/{weekStart}/items/bulk`

```json
{
  "data": {
    "attributes": {
      "items": [
        { "exerciseId": "f4a...", "dayOfWeek": 0, "position": 0 },
        { "exerciseId": "g5b...", "dayOfWeek": 0, "position": 1 },
        { "exerciseId": "h6c...", "dayOfWeek": 2, "position": 0,
          "planned": { "sets": 4, "reps": 8, "weight": 135, "weightUnit": "lb" } }
      ]
    }
  }
}
```

Atomic — all items insert in a single transaction or none do. Each item follows the same defaulting rules as the single-item endpoint. `201` returns the full week resource. Errors apply per-item but cause the whole request to fail.

### 4.4 Update planned item

`PATCH /api/v1/workouts/weeks/{weekStart}/items/{itemId}`

Partial: `dayOfWeek`, `position`, `planned.*`, `notes`. `200` returns the updated item.

### 4.5 Delete planned item

`DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}` — also deletes its performance and per-set rows. `204`.

### 4.6 Reorder planned items

`POST /api/v1/workouts/weeks/{weekStart}/items/reorder`

```json
{
  "data": {
    "attributes": {
      "items": [
        { "itemId": "p1...", "dayOfWeek": 0, "position": 0 },
        { "itemId": "p2...", "dayOfWeek": 0, "position": 1 },
        { "itemId": "p3...", "dayOfWeek": 2, "position": 0 }
      ]
    }
  }
}
```

Atomically applies all updates in a single transaction. `200` returns the full week resource.

---

## 5. Performance

### 5.1 Update summary performance

`PATCH /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance`

```json
{
  "data": {
    "attributes": {
      "status": "done",
      "weightUnit": "lb",
      "actuals": { "sets": 3, "reps": 10, "weight": 100 },
      "notes": "Felt strong"
    }
  }
}
```

Any field optional. Setting `actuals` while in `per_set` mode is rejected with `409` — caller must collapse to summary first via `DELETE .../performance/sets`. Changing `weightUnit` while per-set rows exist is also rejected with `409`.

`actuals` shape varies by item kind:
- `strength`: `{ sets, reps, weight }`
- `isometric`: `{ sets, durationSeconds, weight? }`
- `cardio`: `{ durationSeconds, distance, distanceUnit }`

Status derivation when the client sends actuals without a status:
- previous status was `pending` → server sets `partial`
- previous status was `partial` → stays `partial`
- previous status was `done` → stays `done` (the user is correcting, not retracting)
- previous status was `skipped` → server sets `partial` and clears the skip

Explicit status transitions per the state machine in PRD §4.4.1:
- `status: "done"` from any state → `done` (preserves actuals)
- `status: "skipped"` from any state → `skipped` (clears actuals)
- `status: "pending"` from `done` → `partial` if actuals exist, else `pending`
- `status: "pending"` from `skipped` → `pending` (unskip)

### 5.2 Replace per-set entries

`PUT /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets`

```json
{
  "data": {
    "attributes": {
      "weightUnit": "lb",
      "sets": [
        { "reps": 10, "weight": 100 },
        { "reps": 8,  "weight": 105 },
        { "reps": 6,  "weight": 110 }
      ]
    }
  }
}
```

`setNumber` is assigned by the server in array order, starting at 1. Switches the performance into `per_set` mode and clears any prior summary `actuals`. The single `weightUnit` applies to every set. Rejected with `422` for `isometric` and `cardio` items.

`200` returns the updated performance.

### 5.3 Collapse per-set into summary

`DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets`

Switches the performance back to `summary` mode. Server collapses per-set rows into the summary row using:

- `actualSets = count of sets`
- `actualReps = max reps across sets`
- `actualWeight = max weight across sets`

`weightUnit` on the performance is preserved. `200` returns the updated performance.

---

## 6. Today

`GET /api/v1/workouts/today`

Returns the current day's planned items with embedded performances. The current day is computed in the user's local time zone (from account preferences). This is the primary mobile entry point.

```json
{
  "data": {
    "type": "today",
    "id": "2026-04-09",
    "attributes": {
      "date": "2026-04-09",
      "weekStartDate": "2026-04-06",
      "dayOfWeek": 3,
      "isRestDay": false,
      "items": [
        /* same item shape as inside weeks.attributes.items */
      ]
    }
  }
}
```

`isRestDay` is `true` if today's day-of-week appears in the parent week's `restDayFlags`. Returns an empty `items` array (not `404`) when there are no planned items today.

---

## 7. Week Summary

`GET /api/v1/workouts/weeks/{weekStart}/summary`

```json
{
  "data": {
    "type": "week-summaries",
    "id": "2026-04-06",
    "attributes": {
      "weekStartDate": "2026-04-06",
      "restDayFlags": [5, 6],
      "totalPlannedItems": 18,
      "totalPerformedItems": 16,
      "totalSkippedItems": 1,
      "byDay": [
        {
          "dayOfWeek": 0,
          "isRestDay": false,
          "items": [
            {
              "itemId": "p1...",
              "exerciseName": "Machine Chest Press",
              "status": "done",
              "planned": { "sets": 3, "reps": 10, "weight": 100, "weightUnit": "lb" },
              "actualSummary": { "sets": 3, "reps": 10, "weight": 100, "weightUnit": "lb" }
            }
          ]
        }
      ],
      "byTheme": [
        {
          "themeId": "9c3...",
          "themeName": "Muscle",
          "itemCount": 16,
          "strengthVolume": { "value": 24500, "unit": "lb" },
          "cardio": null
        },
        {
          "themeId": "ab1...",
          "themeName": "Cardio",
          "itemCount": 2,
          "strengthVolume": null,
          "cardio": {
            "totalDurationSeconds": 3600,
            "totalDistance": { "value": 6.2, "unit": "mi" }
          }
        }
      ],
      "byRegion": [
        {
          "regionId": "1b8...",
          "regionName": "Chest",
          "itemCount": 4,
          "strengthVolume": { "value": 6000, "unit": "lb" },
          "cardio": null
        }
      ]
    }
  }
}
```

`strengthVolume.value = Σ (sets × reps × weight)` for `strength` items in the group, after converting all weights to a single unit. `bodyweight` strength items are excluded from `strengthVolume` (we don't know absolute load); they contribute to `itemCount` only. `isometric` items are also excluded from `strengthVolume`; they contribute to `itemCount` only.

The chosen weight unit is the user's most-used unit in the week (ties broken in favor of `lb`). `byTheme[*].cardio.totalDistance.unit` is the user's most-used cardio distance unit in the week (ties broken in favor of `mi`).

**Region totals only count the primary `regionId`** of each exercise. `secondaryRegionIds` are ignored in volume math to prevent double-counting. (They are still used for filtering on `GET /exercises?regionId=...`.)

`actualSummary` reflects the canonical summary of the performance regardless of mode — for `per_set` performances, it is the same collapse rule used in §5.3, but is computed on the fly and not persisted.
