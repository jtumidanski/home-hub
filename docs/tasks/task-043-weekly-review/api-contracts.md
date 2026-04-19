# Weekly Review — API Contracts

This document specifies the exact request / response shapes for the endpoints touched by this task. JSON:API envelope conventions follow the project standard and the existing summary endpoint.

---

## 1. `GET /api/v1/workouts/weeks/{weekStart}/summary` — Extended

Existing endpoint; two new top-level attributes and one extended nested shape.

### Path parameters
- `weekStart` — `YYYY-MM-DD`. Normalized server-side to the Monday of its ISO week.

### Response — `200 OK`

```json
{
  "data": {
    "type": "workoutWeekSummary",
    "id": "2026-04-13",
    "attributes": {
      "weekStartDate": "2026-04-13",
      "restDayFlags": [6],
      "totalPlannedItems": 12,
      "totalPerformedItems": 9,
      "totalSkippedItems": 1,

      "previousPopulatedWeek": "2026-04-06",
      "nextPopulatedWeek": null,

      "byDay": [
        {
          "dayOfWeek": 0,
          "isRestDay": false,
          "items": [
            {
              "itemId": "0e9a...",
              "exerciseName": "Bench Press",
              "status": "done",
              "planned": {
                "sets": 3,
                "reps": 10,
                "weight": 135,
                "weightUnit": "lb"
              },
              "actualSummary": {
                "sets": 3,
                "reps": 10,
                "weight": 140,
                "weightUnit": "lb",
                "sets": [
                  { "setNumber": 1, "reps": 10, "weight": 135 },
                  { "setNumber": 2, "reps": 10, "weight": 140 },
                  { "setNumber": 3, "reps": 8,  "weight": 145 }
                ]
              }
            }
          ]
        }
      ],

      "byTheme":  [ /* unchanged */ ],
      "byRegion": [ /* unchanged */ ]
    }
  }
}
```

### New field semantics

- `previousPopulatedWeek` — `string | null`. Most recent week start strictly earlier than `weekStart` that has at least one planned item. `null` when none.
- `nextPopulatedWeek` — `string | null`. Soonest week start strictly later than `weekStart` that has at least one planned item. `null` when none.
- `byDay[].items[].actualSummary.sets` — `Array<{ setNumber: int, reps: int, weight: number }>`. Present only when the underlying performance is in `per_set` mode; absent otherwise. The parallel scalar `sets` (the summary count) and `reps` / `weight` (the collapsed max-based fields) continue to be emitted alongside the array, per the existing projection at services/workout-service/internal/summary/processor.go:487–505.

> Naming note: the collapsed scalar `sets` is an integer count and the per-set array is also named `sets`. In the existing projection the collapsed fields are sibling map keys without a wrapping object, so the two cannot coexist in the same JSON map. Implementation plan: rename the per-set array key to `setRows` in the actual response to avoid the collision. The example above uses `sets` for readability; the implementation is free to finalize naming.

### Errors
- `400` — malformed `weekStart`
- `404` — no week row exists

---

## 2. `GET /api/v1/workouts/weeks/nearest` — New

Lightweight navigation helper. Returns the nearest week (in the requested direction) that has planned items for the current user.

### Query parameters
- `reference` — `YYYY-MM-DD`, required. Normalized to the Monday of its ISO week server-side.
- `direction` — `"prev"` or `"next"`, required.

### Response — `200 OK`

```json
{
  "data": {
    "type": "workoutWeekPointer",
    "id": "2026-04-06",
    "attributes": {
      "weekStartDate": "2026-04-06"
    }
  }
}
```

### Errors
- `400` — missing or invalid parameter
- `404` — no populated week exists in that direction for this user

### Semantics
- Authenticated via the standard JWT middleware; tenant and user are derived from the JWT.
- Cross-user access is impossible: the query is always filtered by `user_id`.
- Strictly earlier / strictly later: a populated week matching `reference` itself is never returned.
- Reuses the `planned_items INNER JOIN weeks` pattern in services/workout-service/internal/week/provider.go:26 (for `prev`) and a new symmetric helper (for `next`).

---

## 3. No changes

- `GET /api/v1/workouts/weeks/{weekStart}` — unchanged
- `PATCH /api/v1/workouts/weeks/{weekStart}` — unchanged
- `POST /api/v1/workouts/weeks/{weekStart}/copy` — unchanged
- All planned-item and performance endpoints — unchanged
