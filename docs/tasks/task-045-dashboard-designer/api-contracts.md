# Dashboard Designer — API Contracts

All endpoints use JSON:API and the same auth/tenant middleware as existing services. Tenant, household, and user are derived from the JWT + `X-Household-Id` header per existing conventions.

Resource type: `dashboards`.

## List

```
GET /api/v1/dashboards
```

Returns all dashboards the caller can see for the current household: every household-scoped row, plus the caller's own user-scoped rows.

Response (200):
```json
{
  "data": [
    {
      "type": "dashboards",
      "id": "3f1c...",
      "attributes": {
        "name": "Home",
        "scope": "household",
        "sortOrder": 0,
        "layout": { "version": 1, "widgets": [ /* ... */ ] },
        "schemaVersion": 1,
        "createdAt": "2026-04-23T10:00:00Z",
        "updatedAt": "2026-04-23T10:00:00Z"
      }
    }
  ]
}
```

Notes:
- `scope` is a derived attribute in the response: `"household"` when `user_id IS NULL`, `"user"` otherwise. It is NOT stored as a column.
- `layout` is always included (dashboards are small enough that lazy-loading adds no value).

## Fetch one

```
GET /api/v1/dashboards/{id}
```

404 if not visible to caller.

## Create

```
POST /api/v1/dashboards
```

Request:
```json
{
  "data": {
    "type": "dashboards",
    "attributes": {
      "name": "Weekend",
      "scope": "household",
      "layout": { "version": 1, "widgets": [] }
    }
  }
}
```

- `scope` required, either `"household"` or `"user"`.
- `layout` optional; defaults to `{ "version": 1, "widgets": [] }` when omitted.
- `sortOrder` optional; defaults to `max(sort_order) + 1` within the target scope for the caller.
- `schemaVersion` is server-assigned; client input is ignored.

Response (201): full resource.

Errors:
- 400 JSON:API validation for missing/invalid `name`, invalid `scope`.
- 422 for layout validation failures (see §4.9 of PRD). Error objects should have stable codes:
    - `layout.widget_count_exceeded`
    - `layout.widget_unknown_type`
    - `layout.widget_bad_geometry`
    - `layout.config_too_large`
    - `layout.payload_too_large`

## Update

```
PATCH /api/v1/dashboards/{id}
```

Accepts any subset of `name`, `layout`, `sortOrder`. Unspecified fields are unchanged. Same validation rules as create.

## Delete

```
DELETE /api/v1/dashboards/{id}
```

204 on success. 404 if not visible. Allowed even if the target is the caller's default — the preference is cleared as a side effect (service-to-service call to account-service, or a null check on read in the frontend fallback; see PRD §4.6 — the frontend fallback is the simpler v1 implementation).

## Bulk reorder

```
PATCH /api/v1/dashboards/order
```

Request:
```json
{
  "data": [
    { "id": "3f1c...", "sortOrder": 0 },
    { "id": "7ab9...", "sortOrder": 1 }
  ]
}
```

- All ids in the request must be in the caller's visibility set AND in a single scope (either all household-scoped or all caller-user-scoped).
- Mixed-scope requests return 400.
- Each id is updated atomically in one transaction.

Response (200): new list in the new order.

## Promote (user → household)

```
POST /api/v1/dashboards/{id}/promote
```

- Caller must be the owner of the user-scoped dashboard.
- Clears `user_id`. Keeps `id`, `name`, `layout`.
- Returns the updated resource.
- 409 if the dashboard is already household-scoped.

## Copy to mine

```
POST /api/v1/dashboards/{id}/copy-to-mine
```

- Source must be a household-scoped dashboard visible to the caller.
- Creates a new row with a new id, `user_id = caller`, `name` = `"{source.name} (mine)"`, same `layout`.
- `sort_order` = `max(sort_order) + 1` within the caller's user scope (appended at end).
- Returns the new resource.

## Seed

```
POST /api/v1/dashboards/seed
```

Request:
```json
{
  "data": {
    "type": "dashboards",
    "attributes": {
      "name": "Home",
      "layout": { "version": 1, "widgets": [ /* frontend-supplied seed */ ] }
    }
  }
}
```

- Idempotent: if the caller's household already has at least one household-scoped dashboard, returns 200 with the existing list (JSON:API list response, not a single resource).
- On first call, creates the household-scoped dashboard from the supplied payload and returns 201 with a single-resource body.
- The frontend owns the seed contents so the backend stays widget-agnostic; the backend still enforces §4.9 validation on the submitted layout.

## Error envelope

All error responses follow JSON:API:
```json
{
  "errors": [
    {
      "status": "422",
      "code": "layout.widget_unknown_type",
      "title": "Unknown widget type",
      "detail": "widget[3].type 'foo' is not in the registry",
      "source": { "pointer": "/data/attributes/layout/widgets/3/type" }
    }
  ]
}
```
