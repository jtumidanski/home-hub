# Retention API Contracts

All endpoints are JSON:API and live on the existing `/api/v1` prefix in account-service unless noted. Internal endpoints live on each service's internal port and require a service-to-service token.

## GET /api/v1/retention-policies

Returns the fully-resolved retention policy for the caller's active household and personal scope.

**Auth:** authenticated user.

**Response 200**

```json
{
  "data": {
    "type": "retention-policies",
    "id": "{tenant_id}",
    "attributes": {
      "household": {
        "id": "{household_id}",
        "categories": {
          "productivity.completed_tasks":              { "days": 365, "source": "default" },
          "productivity.deleted_tasks_restore_window": { "days": 30,  "source": "household" },
          "recipe.deleted_recipes_restore_window":     { "days": 30,  "source": "default" },
          "recipe.restoration_audit":                  { "days": 90,  "source": "default" },
          "calendar.past_events":                      { "days": 365, "source": "default" },
          "package.archive_window":                    { "days": 7,   "source": "default" },
          "package.archived_delete_window":            { "days": 30,  "source": "default" },
          "system.retention_audit":                    { "days": 180, "source": "default" }
        }
      },
      "user": {
        "id": "{user_id}",
        "categories": {
          "tracker.entries":                           { "days": 730,  "source": "default" },
          "tracker.deleted_items_restore_window":      { "days": 30,   "source": "user" },
          "workout.performances":                      { "days": 1825, "source": "default" },
          "workout.deleted_catalog_restore_window":    { "days": 30,   "source": "default" }
        }
      }
    }
  }
}
```

`source` is one of `default`, `tenant`, `household`, `user`.

## PATCH /api/v1/retention-policies/household/:household_id

Set or clear household-scoped overrides. Body is a sparse map; `null` clears an override.

**Auth:** household admin role for `:household_id`.

**Request**

```json
{
  "data": {
    "type": "retention-policies",
    "attributes": {
      "categories": {
        "productivity.completed_tasks": 180,
        "calendar.past_events": null
      }
    }
  }
}
```

**Response 200** — same shape as `GET`.

**Errors**

- `400` — value below 1 or above 3650; soft-delete restore window above 365; unknown category.
- `403` — caller is not an admin of the target household.
- `404` — household does not exist.

## PATCH /api/v1/retention-policies/user

Set or clear user-scoped overrides for the calling user. Same body shape as the household variant.

**Auth:** authenticated user; `scope_id` is implied to be the caller.

**Errors:** `400` on bounds / unknown category.

## POST /api/v1/retention-policies/purge

Trigger an immediate purge for one category. Account-service authorizes the request and forwards it to the owning service. Set `dry_run: true` to get the counts that *would* be deleted without removing any rows.

**Request**

```json
{
  "data": {
    "type": "retention-purges",
    "attributes": {
      "category": "productivity.completed_tasks",
      "scope": "household",
      "dry_run": false
    }
  }
}
```

`dry_run` defaults to `false`. The settings UI sends `dry_run: true` to preview row counts before confirming a window-shrinking edit.

**Response 202**

```json
{
  "data": {
    "type": "retention-purges",
    "id": "{run_id}",
    "attributes": {
      "category": "productivity.completed_tasks",
      "scope": "household",
      "scope_id": "{household_id}",
      "status": "accepted"
    }
  }
}
```

**Errors**

- `400` — unknown category.
- `403` — caller lacks permission for the requested scope (non-admin requesting household scope, or scope mismatch on user).
- `429` — rate limit: one manual purge per (tenant, category) per 60s.
- `503` — owning service unreachable; the request is not queued.

## GET /api/v1/retention-runs

Paginated audit feed across services for the caller's tenant, filtered to scopes the caller can see (their household + their personal scope).

**Query params**

- `category` — optional filter.
- `trigger` — optional, `scheduled` | `manual`.
- `limit` — default 20, max 100.
- `cursor` — opaque pagination cursor.

**Response 200**

```json
{
  "data": [
    {
      "type": "retention-runs",
      "id": "{run_id}",
      "attributes": {
        "service": "productivity-service",
        "category": "productivity.completed_tasks",
        "scope": "household",
        "scope_id": "{household_id}",
        "trigger": "scheduled",
        "scanned": 1284,
        "deleted": 47,
        "started_at": "2026-04-11T03:00:00Z",
        "finished_at": "2026-04-11T03:00:01Z",
        "error": null
      }
    }
  ],
  "meta": { "next_cursor": "..." }
}
```

## POST /internal/retention/purge (per service)

Internal endpoint exposed by each reaper-owning service. Not routed through the public gateway. Supports `?dry_run=true` for preview.

**Auth:** service-to-service token (existing internal auth pattern).

**Request**

```json
{
  "tenant_id": "...",
  "scope_kind": "household",
  "scope_id": "...",
  "category": "productivity.completed_tasks",
  "dry_run": false
}
```

**Response 200**

```json
{
  "run_id": "...",
  "scanned": 1284,
  "deleted": 47,
  "dry_run": false,
  "duration_ms": 812
}
```

When `dry_run` is `true`, the reaper performs the full scan and cascade walk inside a transaction that is rolled back at the end. `deleted` reflects what *would* have been removed; no rows are touched. The audit row is still written with `dry_run = true` for visibility.

**Errors**

- `400` — category not owned by this service.
- `409` — another reaper is currently holding the advisory lock for this `(tenant, category)`; caller may retry.
- `503` — policy could not be loaded and no cache available; no rows touched.

## Category ownership map

| Category | Owning service |
|---|---|
| `productivity.completed_tasks` | productivity-service |
| `productivity.deleted_tasks_restore_window` | productivity-service |
| `recipe.deleted_recipes_restore_window` | recipe-service |
| `recipe.restoration_audit` | recipe-service |
| `tracker.entries` | tracker-service |
| `tracker.deleted_items_restore_window` | tracker-service |
| `workout.performances` | workout-service |
| `workout.deleted_catalog_restore_window` | workout-service |
| `calendar.past_events` | calendar-service |
| `package.archive_window` | package-service |
| `package.archived_delete_window` | package-service |
| `system.retention_audit` | each reaper-owning service (self-reaping) |
