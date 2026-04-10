# workout-service — REST

All endpoints under `/api/v1/workouts/...`. Tenant + user are extracted from
the JWT. JSON:API style envelopes throughout. The full request/response
shapes are in `docs/tasks/task-027-workout-tracker/api-contracts.md`.

## Themes

- `GET    /api/v1/workouts/themes`
- `POST   /api/v1/workouts/themes`
- `PATCH  /api/v1/workouts/themes/{id}`
- `DELETE /api/v1/workouts/themes/{id}`

Default themes (`Muscle`, `Cardio`) are installed on first request.

## Regions

- `GET    /api/v1/workouts/regions`
- `POST   /api/v1/workouts/regions`
- `PATCH  /api/v1/workouts/regions/{id}`
- `DELETE /api/v1/workouts/regions/{id}`

Default region list is installed on first request.

## Exercises

- `GET    /api/v1/workouts/exercises?themeId=&regionId=` — `regionId` matches primary OR secondary
- `POST   /api/v1/workouts/exercises`
- `PATCH  /api/v1/workouts/exercises/{id}` — `kind` and `weightType` are immutable (422 if attempted)
- `DELETE /api/v1/workouts/exercises/{id}` — soft delete

## Weeks

- `GET    /api/v1/workouts/weeks/{weekStart}` — 404 if no row exists; lazy create only happens on mutation
- `PATCH  /api/v1/workouts/weeks/{weekStart}` — currently only `restDayFlags` is patchable; lazily creates the week
- `POST   /api/v1/workouts/weeks/{weekStart}/copy` — `mode: planned|actual`. Returns 409 on non-empty target, 404 on no source

`weekStart` is any `YYYY-MM-DD`; the server normalizes to the Monday of the ISO week.

## Planned items

- `POST   /api/v1/workouts/weeks/{weekStart}/items` — single add; defaults from the exercise when omitted
- `POST   /api/v1/workouts/weeks/{weekStart}/items/bulk` — atomic batch
- `PATCH  /api/v1/workouts/weeks/{weekStart}/items/{itemId}`
- `DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}`
- `POST   /api/v1/workouts/weeks/{weekStart}/items/reorder` — atomic (day, position) reassignment

## Performances

- `PATCH  /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance` — summary mode; status state machine per PRD §4.4.1
- `PUT    /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets` — strength only; switches to per_set mode
- `DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets` — collapses sets back to summary

## Composite reads

- `GET    /api/v1/workouts/today` — the current day in UTC, with embedded items + performances
- `GET    /api/v1/workouts/weeks/{weekStart}/summary` — per-day, per-theme, per-primary-region totals

## Error mapping

| Code | When |
| --- | --- |
| 400 | invalid IDs, validation failures (kind/unit/numeric/shape) |
| 404 | resource not found, no source week for copy |
| 409 | duplicate name; per-set ↔ summary guardrails; non-empty copy target |
| 422 | kind/weightType immutability; planning a soft-deleted exercise; per-set on a non-strength item |
