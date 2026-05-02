# API Contracts — Kiosk Dashboard Widgets

This document captures only the contract change(s) introduced by this feature. Final selection between mechanism (1) and (2) is deferred to `design.md`; both are documented so the implementer can pick one without re-litigating the shape.

## Mechanism (1): Per-Key Seed

### `POST /api/v1/dashboards/seed`

Existing endpoint, body extended with an optional `key`.

**Request body** (JSON:API style, matching existing dashboard-service conventions):

```json
{
  "data": {
    "type": "dashboards",
    "attributes": {
      "name": "Kiosk",
      "key": "kiosk",
      "layout": {
        "version": 1,
        "widgets": [ /* ... */ ]
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | yes | 1–80 chars after trim, same rules as create |
| `key` | string | no | 1–40 chars, lowercase, hyphen-separated (`^[a-z][a-z0-9-]{0,39}$`). Idempotency key per `(tenant, household)`. Omit to preserve legacy "any-household-dashboard counts" behavior. |
| `layout` | object | yes | Same `Layout` shape validated by `layout.Validate` |

**Response**: `SeedResult` JSON:API document. Unchanged shape.

**Idempotency semantics**:

- `key` provided + no row exists with `(tenant_id, household_id, seed_key=key)` → create row, return `created: true` and the new dashboard.
- `key` provided + row already exists with that `(tenant, household, seed_key)` → return `created: false` and the existing dashboard (just that one) in `existing[]`.
- `key` omitted → preserves the existing "no-op when any household-scoped dashboard exists" behavior. Used by deployed clients that haven't been updated.

**Validation errors**: standard `422` with code `validation.invalid_field` for name/key/layout violations. Code `validation.invalid_field` with detail pointing to `/data/attributes/key` for malformed keys.

**Transactional safety**: the existing Postgres advisory lock keyed on `(tenant_id, household_id)` already serializes concurrent calls. The per-key uniqueness is enforced by a partial unique index (see `data-model.md`).

### Frontend Orchestration

On dashboard-list load (already happens today in `DashboardsLandingPage` / wherever the listing query fires):

1. Run the existing first-run check.
2. Issue **two** seed calls in parallel (both idempotent):
   - `{ name: "Home", key: "home", layout: <home seed> }`
   - `{ name: "Kiosk", key: "kiosk", layout: <kiosk seed> }`
3. Refetch the dashboard list once both resolve.

Existing households: the second call creates the "Kiosk" row on first load post-deploy, then is a no-op forever after.

---

## Mechanism (2): Frontend-Orchestrated With Preference Flag

No change to `POST /api/v1/dashboards/seed`.

### `account-service` Preferences

A new key in the existing key/value preferences:

| Key | Type | Default | Purpose |
|---|---|---|---|
| `kiosk_dashboard_seeded` | `boolean` | `false` | Frontend-set flag indicating the Kiosk dashboard seeding flow has run for this `(tenant, user, household)` |

Read via the existing `GET /api/v1/preferences`, set via the existing `PATCH /api/v1/preferences`. No new endpoint, no schema migration (the preferences store is already key/value).

### Frontend Orchestration

On dashboard-list load:

1. Existing flow: if no household-scoped dashboards visible, call `/seed` for "Home". Unchanged.
2. New flow (runs after step 1 completes successfully):
   - Read `preferences.kiosk_dashboard_seeded` for the current `(user, household)`.
   - If `true`, do nothing.
   - If `false` or unset:
     - Call `POST /api/v1/dashboards` with the Kiosk layout, `name: "Kiosk"`, `sort_order: 1`, `scope: "household"`.
     - On success (201), `PATCH /api/v1/preferences` to set `kiosk_dashboard_seeded: true`.
     - On 409 / conflict (already exists with that name), still set the preference and move on.

### Trade-offs

| Aspect | Mechanism (1) | Mechanism (2) |
|---|---|---|
| Backend changes | New column + index + processor + handler | None |
| Frontend changes | Two seed calls | One create call + preference read/write |
| "Delete is permanent" | Enforced by row-existence | Enforced by preference flag |
| Future seeded dashboards | Easy — just add a key | Each one needs a new preference flag |
| Failure modes | Standard server-side idempotency | Preference write must happen *after* create succeeds; otherwise re-creation can recur if the create succeeded but the preference write failed |

`design.md` should pick one explicitly with a short rationale.
