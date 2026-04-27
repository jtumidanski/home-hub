# REST API

All `/api/v1/dashboards*` endpoints require JWT authentication and return
JSON:API payloads. Tenant, household, and user are derived from the JWT.
Routes are registered in `internal/dashboard/resource.go:InitializeRoutes`.

Internal endpoints under `/internal/retention/...` are mounted by
`shared/go/retention.MountInternalEndpoints` and guarded by the shared
`X-Internal-Token` header; they are not part of the user-facing API surface.

## Endpoints

| Verb   | Path                                | Purpose                                                    |
|--------|-------------------------------------|------------------------------------------------------------|
| GET    | `/api/v1/dashboards`                | List dashboards visible to the caller                      |
| POST   | `/api/v1/dashboards`                | Create a dashboard (scope = `household` \| `user`)         |
| POST   | `/api/v1/dashboards/seed`           | Idempotent first-run seed of a household dashboard         |
| PATCH  | `/api/v1/dashboards/order`          | Bulk reorder within a single scope                         |
| GET    | `/api/v1/dashboards/{id}`           | Fetch a single dashboard by id                             |
| PATCH  | `/api/v1/dashboards/{id}`           | Update name / layout / sortOrder                           |
| DELETE | `/api/v1/dashboards/{id}`           | Delete a dashboard                                         |
| POST   | `/api/v1/dashboards/{id}/promote`   | Promote a user-scoped dashboard to household scope         |
| POST   | `/api/v1/dashboards/{id}/copy-to-mine` | Deep-copy a household dashboard into caller's user scope |

Resource type is `dashboards`. See `internal/dashboard/rest.go:RestModel` for
the serialized attribute shape:

| Attribute     | Type           |
|---------------|----------------|
| name          | string         |
| scope         | string (`household` \| `user`, derived — see domain.md) |
| sortOrder     | int            |
| layout        | object         |
| schemaVersion | int            |
| createdAt     | string (RFC3339) |
| updatedAt     | string (RFC3339) |

### Reorder body

`PATCH /api/v1/dashboards/order` deliberately uses a plain JSON body, not
JSON:API-wrapped, because it is a bulk action rather than a single-resource
mutation (see `resource.go:reorderHandler` comment).

```json
{ "data": [ { "id": "<uuid>", "sortOrder": 0 }, ... ] }
```

All referenced ids must be visible to the caller and share a single scope
(all household or all caller-owned user rows). Mixed scope → 400
`dashboard.mixed_scope`.

### Seed response

`POST /api/v1/dashboards/seed` returns either:

- 201 with a single `dashboards` resource (created a new row), or
- 200 with an array of visible `dashboards` resources (existing rows — the UI
  picks one). `seedHandler` in `resource.go` branches on `res.Created`.

## Error Codes

Error objects follow JSON:API (`server.WriteJSONAPIError`). Every error listed
here carries a machine-readable `code` plus, where applicable, a
`source.pointer` into the request body.

### Layout validation (`internal/layout/validator.go`)

Raised by Create, Update, and Seed when the layout payload fails validation.
All layout errors return **HTTP 422** and a pointer into
`/data/attributes/layout/...`. See `resource.go:writeLayoutError`.

| Code                                  | Meaning                                                   |
|---------------------------------------|-----------------------------------------------------------|
| `layout.unsupported_schema_version`   | `version` ≠ 1                                             |
| `layout.widget_count_exceeded`        | More than 40 widgets                                      |
| `layout.widget_unknown_type`          | Widget `type` not in the shared allowlist                 |
| `layout.widget_bad_geometry`          | `x < 0 \|\| y < 0 \|\| w < 1 \|\| h < 1 \|\| x+w > 12`    |
| `layout.widget_bad_id`                | Widget `id` missing or not a UUID                         |
| `layout.widget_duplicate_id`          | Two widgets share the same `id`                           |
| `layout.config_too_large`             | Widget `config` > 4 KiB                                   |
| `layout.config_too_deep`              | Widget `config` nesting depth > 5                         |
| `layout.config_not_object`            | Widget `config` present but not a JSON object             |
| `layout.payload_too_large`            | Whole layout document > 64 KiB                            |
| `layout.malformed`                    | JSON parse error on envelope or config                    |

### Dashboard-level (`internal/dashboard/resource.go`)

| Code                            | HTTP | Raised in                                   | Meaning                                                   |
|---------------------------------|------|---------------------------------------------|-----------------------------------------------------------|
| `dashboard.invalid_scope`       | 400  | `createHandler`                             | `scope` attribute not in {`household`, `user`}            |
| `dashboard.name_invalid`        | 422  | `createHandler`, `updateHandler`, `seedHandler` | `name` empty after trim or > 80 chars                |
| `dashboard.mixed_scope`         | 400  | `reorderHandler`                            | Reorder batch mixes household and user rows               |
| `dashboard.already_household`   | 409  | `promoteHandler`                            | Row is already household-scoped                           |
| `dashboard.not_copyable`        | 400  | `copyToMineHandler`                         | `copy-to-mine` target is user-scoped                      |

Other responses:

| HTTP | Condition                                                                    |
|------|------------------------------------------------------------------------------|
| 400  | Unparseable reorder body or invalid UUID in reorder entry                    |
| 403  | Non-owner attempted to update / delete / promote a user-scoped row           |
| 404  | Row not visible to caller (missing, cross-tenant, cross-household, or other user's) |
| 500  | Unclassified processor / DB errors                                           |

## Consumed Events

The service subscribes to the shared Kafka bus for user-lifecycle events
(`cmd/main.go`):

| Setting       | Value / env                                                        |
|---------------|--------------------------------------------------------------------|
| Brokers       | `BOOTSTRAP_SERVERS` (comma-separated)                              |
| Topic         | `home-hub.user.lifecycle` (`EVENT_TOPIC_USER_LIFECYCLE`)           |
| Consumer group| `dashboard-service` (`KAFKA_CONSUMER_GROUP`)                       |
| Envelope      | `shared/go/events.Envelope`                                        |
| Handled types | `USER_DELETED` → `UserDeletedEvent`                                |

On `USER_DELETED`, the handler at `internal/events/handler.go` deletes every
user-scoped dashboard where `(tenant_id, user_id)` matches the event, then
returns `nil` so the consumer commits the offset. Malformed envelopes and
unknown types are logged and swallowed rather than re-read forever.

## Auth

- All `/api/v1/dashboards*` routes sit behind the JWT middleware from
  `shared/go/auth.Middleware` (wired in `cmd/main.go`). Tenant, household, and
  user ids come from the authenticated context via
  `tenantctx.MustFromContext`.
- `/internal/retention/purge` and `/internal/retention/runs` are mounted by
  `shared/go/retention.MountInternalEndpoints` and guarded by
  `X-Internal-Token` (shared service-to-service secret).
- `/metrics` is the Prometheus scrape endpoint registered alongside retention.
