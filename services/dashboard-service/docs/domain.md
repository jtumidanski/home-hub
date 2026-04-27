# Domain

## Dashboard

### Responsibility

A named, ordered collection of widgets a household or individual user pins to the
home screen. The service owns layout documents (widget positions + configs) and
the scope rules that decide who can see and edit each row. It does not render
widgets — widget data is fetched directly by the frontend from the owning
service.

### Core Models

**Model** (`dashboard.Model`, `internal/dashboard/model.go`)

| Field         | Type           |
|---------------|----------------|
| id            | uuid.UUID      |
| tenantID      | uuid.UUID      |
| householdID   | uuid.UUID      |
| userID        | *uuid.UUID     |
| name          | string         |
| sortOrder     | int            |
| layout        | datatypes.JSON |
| schemaVersion | int            |
| createdAt     | time.Time      |
| updatedAt     | time.Time      |

All fields are immutable after construction. Access is through getter methods.
`Model.IsHouseholdScoped()` reports `userID == nil`.

The GORM entity (`internal/dashboard/entity.go`) carries the same fields plus a
composite index `idx_dashboards_scope` on `(tenant_id, household_id, user_id)`.

### Scope Derivation

Scope is not stored on the row. It is derived from `user_id`:

- `user_id IS NULL` → `"household"` (visible to every household member)
- `user_id = <uuid>` → `"user"` (visible only to that user)

The derivation lives in `rest.go:Transform`. The Create handler takes the inbound
`scope` string and maps it back to a nullable `user_id` in
`processor.go:Processor.Create` (`"household"` → nil, `"user"` → callerUserID,
anything else → `ErrInvalidScope`).

### Invariants

- `tenantID`, `householdID`, and non-empty trimmed `name` are required
  (`builder.go:Build`; `ErrTenantRequired`, `ErrHouseholdRequired`,
  `ErrNameRequired`).
- `name` max length 80 chars after trim (`ErrNameTooLong`).
- `schemaVersion` defaults to 1 and matches `shared.LayoutSchemaVersion`.
- Layout must be a valid JSON document per `internal/layout/validator.go` (see
  below).
- List / GetByID filter rows to those visible to the caller:
  `provider.go:visibleToCaller` — `(tenant_id, household_id) AND (user_id IS NULL
  OR user_id = caller)`. Cross-tenant, cross-household, or other users' rows
  surface as `ErrNotFound`.
- Edit rights (`processor.go:requireEditable`): household rows are editable by
  any member of the household; user rows only by the owner. Non-owner edits of a
  user row return `ErrForbidden`.
- Scope is immutable via `Update`; the only user→household transition is
  `Promote`, and there is no household→user transition.

### Widget Allowlist

The 9 known widget types are registered in `shared/go/dashboard/types.go`
(`WidgetTypes` + `IsKnownWidgetType`):

`weather`, `tasks-summary`, `reminders-summary`, `overdue-summary`,
`meal-plan-today`, `calendar-today`, `packages-summary`, `habits-today`,
`workout-today`.

A parity fixture lives at `shared/go/dashboard/fixtures/widget-types.json` and is
asserted against the Go map. The mirrored TypeScript allowlist at
`frontend/src/lib/dashboard/widget-types.ts` is also asserted against the same
fixture (lands in Phase K).

### Layout Validation

`internal/layout/validator.go` — a pure function (no DB, no HTTP) returning a
stable `ValidationError{Code, Pointer, Message}`. Rules:

| Rule                                                | Code                                  |
|-----------------------------------------------------|---------------------------------------|
| `version == shared.LayoutSchemaVersion` (= 1)       | `CodeUnsupportedSchemaVersion`        |
| `len(widgets) ≤ shared.MaxWidgets` (= 40)           | `CodeWidgetCountExceeded`             |
| Widget `type` in allowlist                          | `CodeWidgetUnknownType`               |
| `x ≥ 0`, `y ≥ 0`, `w ≥ 1`, `h ≥ 1`, `x+w ≤ GridColumns` (= 12) | `CodeWidgetBadGeometry`   |
| Widget `id` is a non-nil UUID                       | `CodeWidgetBadID`                     |
| Widget `id` unique within document                  | `CodeWidgetDuplicateID`               |
| Raw payload ≤ `shared.MaxLayoutBytes` (= 64 KiB)    | `CodePayloadTooLarge`                 |
| Widget `config` ≤ `shared.MaxWidgetConfigBytes` (= 4 KiB) | `CodeConfigTooLarge`            |
| Widget `config` nesting depth ≤ `shared.MaxWidgetConfigDepth` (= 5) | `CodeConfigTooDeep`   |
| Widget `config` must be a JSON object (when non-empty) | `CodeConfigNotObject`              |
| Parse errors on envelope or config                  | `CodeMalformed`                       |

Each `ValidationError` carries a JSON Pointer into the request body so the REST
layer can surface a `source.pointer` for the offending widget/field.

### Processors

**Processor** (`dashboard.Processor`, `internal/dashboard/processor.go`)

| Method                                                         | Description                                           |
|----------------------------------------------------------------|-------------------------------------------------------|
| `List(tenantID, householdID, callerUserID)`                    | Visible dashboards, ordered by sort_order then created_at |
| `GetByID(id, tenantID, householdID, callerUserID)`             | Single dashboard if visible, else `ErrNotFound`       |
| `Create(tenantID, householdID, callerUserID, CreateAttrs)`     | Insert with scope + name + layout validation         |
| `Update(id, tenantID, householdID, callerUserID, UpdateAttrs)` | Patch name/layout/sortOrder; re-runs validation     |
| `Delete(id, tenantID, householdID, callerUserID)`              | Remove row (edit rights required)                    |
| `Reorder(tenantID, householdID, callerUserID, []ReorderPair)`  | Bulk sort_order update, single-scope only           |
| `Promote(id, tenantID, householdID, callerUserID)`             | user-scoped → household-scoped (owner only)          |
| `CopyToMine(id, tenantID, householdID, callerUserID)`          | Deep-copy a household dashboard into caller's scope  |
| `Seed(tenantID, householdID, callerUserID, name, layout)`      | First-run creation, idempotent + race-safe          |

### Promote / CopyToMine Semantics

**Promote** (`processor.go:Promote`). Sets `user_id = NULL` via a raw UPDATE so
the nil reliably writes across dialects (GORM's `Updates` map skips nil
entries). Requires the caller to own the row. Returns `ErrAlreadyHousehold` if
already household-scoped. No household→user transition exists.

**CopyToMine** (`processor.go:CopyToMine`). Only household dashboards are
copyable — a user-scoped source returns `ErrNotCopyable`. Deep-copies the row
into the caller's user scope:

1. Parses the layout and regenerates every widget `id` via
   `regenerateWidgetIDs` so the two rows are independent.
2. Names the new row `"<original name> (mine)"`, truncated to 80 chars so it
   fits the name-length invariant.
3. Appends at `max(sortOrder)+1` within the caller's user scope.

### Seed

`processor.go:Seed` ensures at least one household-scoped dashboard exists for
the given (tenant, household). Idempotent and race-safe:

- Serializes concurrent seeders via `SELECT pg_advisory_xact_lock(?)` inside a
  transaction (see `acquireSeedLock` + `seedLockKey` in storage.md).
- On existing rows, returns `SeedResult{Created: false, Existing: [...]}` where
  `Existing` is `visibleToCaller` so the UI can pick one. On first call, inserts
  the supplied name + layout and returns `SeedResult{Created: true, Dashboard}`.

### Errors

Exported sentinel errors (`processor.go`):

- `ErrInvalidScope` — 400 `dashboard.invalid_scope`
- `ErrNotFound` — 404
- `ErrForbidden` — 403
- `ErrMixedScope` — 400 `dashboard.mixed_scope` (Reorder batch mixes household
  and user rows)
- `ErrAlreadyHousehold` — 409 `dashboard.already_household`
- `ErrNotCopyable` — 400 `dashboard.not_copyable`

Name errors (`builder.go`): `ErrNameRequired`, `ErrNameTooLong` — both surface
as 422 `dashboard.name_invalid`.

---

## Retention Category

### Responsibility

Registers `dashboard.dashboards` with the shared retention framework
(`internal/retention/handlers.go`). Plumbing only — `Reap` is intentionally a
no-op in v1 and the compiled default is 0 days (never purge). Scope is
household-level; `DiscoverScopes` returns every distinct `(tenant_id,
household_id)` with dashboards.

An `AuditTrim` handler under `system.retention_audit` trims this service's own
`retention_runs` table, matching the pattern shared by every retention-using
service.

---

## User-Lifecycle Cascade

### Responsibility

Consumes `UserDeletedEvent` from the shared Kafka bus and hard-deletes every
user-scoped dashboard for the affected `(tenant, user)`. Implemented in
`internal/events/handler.go` (`Handler.Dispatch`) and wired as the single
Kafka-message handler in `cmd/main.go`.

- Events arrive on topic `home-hub.user.lifecycle` (env
  `EVENT_TOPIC_USER_LIFECYCLE`, default matching). Consumer group
  `dashboard-service`.
- The envelope is `shared/go/events.Envelope{Type, Version, Payload}`. Only
  `TypeUserDeleted` is handled; unknown types are logged and silently ignored.
- The delete is `WHERE tenant_id = ? AND user_id = ?`, so it is naturally
  idempotent — a second delivery matches zero rows.
- Malformed envelopes and malformed `UserDeletedEvent` payloads are logged and
  swallowed (return `nil`) so the consumer commits the offset rather than
  re-reading forever.
