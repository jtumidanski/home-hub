# Dashboard Designer — Design

Version: v1
Status: Approved
Created: 2026-04-23

Companion to `prd.md`, `api-contracts.md`, `data-model.md`, `ux-flow.md`. This document
records the architectural decisions made during Phase 2 and how the pieces fit
together. The PRD answers *what* and *why*; this doc answers *how*.

---

## 1. Summary of decisions

Eight architectural decisions were resolved in brainstorming, each chosen from 2–4
explicit alternatives. The short form:

| # | Decision | Chosen | Why (one-liner) |
|---|---|---|---|
| 1 | `default_dashboard_id` storage | New typed `household_preferences` table in `account-service`, scoped by `(tenant, user, household)` | Matches existing typed-preference style; current `preference` table has no household dimension. |
| 2 | Account-delete cascade mechanism | Kafka event bus (`UserDeletedEvent` on `EVENT_TOPIC_USER_LIFECYCLE`); new `shared/go/kafka` lib modeled on atlas-ms conventions; dashboard-service consumes | Loose coupling; reusable for future cascades (household-delete, etc.). |
| 3 | Widget-type allowlist single source of truth | Hand-mirrored constants in `shared/go/dashboard/types.go` and `frontend/src/lib/dashboard/widget-types.ts`, enforced by parity tests on both sides | Set is ~9 items, grows slowly; cross-stack change is already required per new widget (component + Zod + default config). |
| 4 | Grid library | `react-grid-layout`, code-split into the designer bundle only | Built for this exact use case; renderer path stays library-free (CSS Grid). |
| 5 | Renderer vs. designer page structure | Two nested routes (`/dashboards/:id` and `/dashboards/:id/edit`) sharing a `DashboardShell` parent that owns the dashboard query | Natural code splitting; back-button restores view mode; each component has a single responsibility. |
| 6 | Seed endpoint race safety | Postgres advisory lock (`pg_advisory_xact_lock`) keyed by `hash(tenant_id, household_id, "dashboard_seed")` | No schema pollution; scoped to one rare operation. |
| 7 | Config schema evolution | Tolerant-read / strict-write: merge `defaultConfig` under raw, `safeParse`, fall back to `defaultConfig` + `lossy: true` flag surfaced as a small badge in view mode | Survives schema drift without migration infra; loud enough to notice, soft enough to not break dashboards. |
| 8 | Dirty-state guard + legacy redirect | `useBlocker` + `beforeunload` for unsaved changes; dedicated `<DashboardRedirect>` component for `/app/dashboard` (classic React Router — loader API unavailable) | Matches the current non-data-router setup. |

Kafka infrastructure: the existing k3s-managed broker is treated like Postgres — an
external dependency with `BOOTSTRAP_SERVERS` env. No docker-compose broker container.

---

## 2. Backend architecture

### 2.1 New service `dashboard-service`

Layout follows the standard Home Hub service pattern:

```
services/dashboard-service/
├── cmd/main.go
├── Dockerfile
├── go.mod
├── docs/{domain.md, rest.md, storage.md}
└── internal/
    ├── appcontext/                  tenant/household/user extraction (copy existing pattern)
    ├── config/                      env loading
    ├── dashboard/                   primary domain
    │   ├── entity.go                GORM row matching data-model.md §1
    │   ├── model.go                 immutable domain model + builder
    │   ├── builder.go
    │   ├── provider.go              read-side query functions
    │   ├── administrator.go         write-side mutation functions
    │   ├── processor.go             orchestrates provider + administrator + events
    │   ├── processor_test.go
    │   ├── resource.go              route registration
    │   ├── rest.go                  JSON:API request/response shapes
    │   └── rest_test.go
    ├── layout/                      layout-JSON validation (pure, no DB)
    │   ├── validator.go
    │   └── validator_test.go
    ├── events/                      Kafka consumer
    │   ├── consumer.go
    │   └── handler_user_deleted.go
    └── retention/                   registers the `dashboards` category
        └── wire.go
```

Owns schema `dashboard` with one table `dashboard.dashboards` exactly as in
`data-model.md §1`. The scope index `(tenant_id, household_id, user_id)` plus the
partial index on `(tenant_id, household_id) WHERE user_id IS NULL` back every
list query. Tenant scoping via existing middleware; DB via `shared/go/database`.

Nine endpoints under `/api/v1/dashboards` per `api-contracts.md`: list, get,
create, update, delete, bulk reorder, promote, copy-to-mine, seed.

### 2.2 New shared packages

**`shared/go/dashboard`** — widget-type allowlist + layout-schema constants.

```go
package dashboard

const (
    LayoutSchemaVersion   = 1
    MaxWidgets            = 40
    MaxLayoutBytes        = 64 * 1024
    MaxWidgetConfigBytes  = 4 * 1024
    MaxWidgetConfigDepth  = 5
    GridColumns           = 12
)

var WidgetTypes = map[string]struct{}{
    "weather": {}, "tasks-summary": {}, "reminders-summary": {},
    "overdue-summary": {}, "meal-plan-today": {}, "calendar-today": {},
    "packages-summary": {}, "habits-today": {}, "workout-today": {},
}

func IsKnownWidgetType(t string) bool { _, ok := WidgetTypes[t]; return ok }
```

A Go parity test asserts the sorted set matches a committed fixture; a TS parity
test asserts the same against the same fixture. Divergence fails both builds.

**`shared/go/kafka`** — producer/consumer wrapper on `segmentio/kafka-go`,
following atlas-ms conventions. Includes: `producer.Produce` +
`producer.SingleMessageProvider`, header decorators for tenant + request-id
propagation, a retry wrapper, and `consumer.Manager` + `consumer.Register` for
topic handlers. Topic names come from env vars
(e.g. `EVENT_TOPIC_USER_LIFECYCLE`). Broker address from `BOOTSTRAP_SERVERS`.

**`shared/go/events`** — cross-service domain event envelopes.

```go
type UserDeletedEvent struct {
    TenantID  uuid.UUID `json:"tenantId"`
    UserID    uuid.UUID `json:"userId"`
    DeletedAt time.Time `json:"deletedAt"`
}
```

Versioned via the envelope's `Type` field (atlas convention); `UserDeletedEvent`
is `Type: "USER_DELETED"`, keyed on the 32-bit hash of `userID` so all events for
a user land on one partition.

### 2.3 `account-service` changes

#### `household_preferences` table

New typed entity at `services/account-service/internal/householdpreference/`,
mirroring the existing `preference` package style:

```go
type Entity struct {
    Id                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TenantId           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tup"`
    UserId             uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tup"`
    HouseholdId        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tup"`
    DefaultDashboardId *uuid.UUID `gorm:"type:uuid"`
    CreatedAt          time.Time
    UpdatedAt          time.Time
}

func (Entity) TableName() string { return "household_preferences" }
```

Two JSON:API endpoints under `/api/v1/household-preferences`:

- `GET` — returns the caller's row for the active household; auto-creates an
  empty row if none exists, so the shape is always present.
- `PATCH` — accepts `defaultDashboardId: string | null`.

JSON:API response uses resource type `householdPreferences` and carries the row's
`id` plus `defaultDashboardId`, `createdAt`, `updatedAt`.

#### `UserDeletedEvent` production

`account-service`'s user-delete flow (whatever the existing handler is) gains two
things inside the same transactional unit as the user row delete:

1. `DELETE FROM household_preferences WHERE tenant_id = $1 AND user_id = $2` —
   owned locally, no event needed.
2. After commit, best-effort produce `UserDeletedEvent` on
   `EVENT_TOPIC_USER_LIFECYCLE` via the `shared/go/kafka` retry wrapper. On
   final failure, log a warning with tenant/user ids; no outbox in v1.

### 2.4 Dashboard-service consumer

`services/dashboard-service/internal/events/handler_user_deleted.go` registers
with the shared-kafka `consumer.Manager` on `EVENT_TOPIC_USER_LIFECYCLE`. On
receipt of a `USER_DELETED` envelope it runs:

```sql
DELETE FROM dashboard.dashboards
WHERE tenant_id = $1 AND user_id = $2
```

Idempotent by construction. Emits `dashboard_user_delete_cascade_total`.

### 2.5 Layout validator

`services/dashboard-service/internal/layout/validator.go` is a pure function:

```go
func Validate(raw json.RawMessage) (Layout, error)
```

Enforces PRD §4.9 rules and returns structured `layout.ValidationError` values
with the stable codes from `api-contracts.md`:

- `layout.widget_count_exceeded`
- `layout.widget_unknown_type`
- `layout.widget_bad_geometry`
- `layout.config_too_large`
- `layout.payload_too_large`

`resource.go` maps each code to a JSON:API error object with a proper source
pointer (e.g. `/data/attributes/layout/widgets/3/type`). Validation runs on
create, update, and seed.

Explicitly **not** validated at the backend: the inside shape of `config`
(frontend Zod owns that); overlap between widgets (frontend grid engine
prevents, renderer tolerates).

### 2.6 Seed endpoint race safety

`POST /api/v1/dashboards/seed` runs the critical section inside one transaction
guarded by a Postgres advisory lock keyed per household:

```go
lockKey := packHash(tenantID, householdID)  // int64 derived from both UUIDs
tx := db.Begin()
tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey)
// re-check count(*) of household-scoped rows
// if > 0: rollback or commit; return existing list with 200
// if == 0: insert the seeded row, commit; return new row with 201
```

`pg_advisory_xact_lock` releases on commit or rollback — no manual release
required. No new columns, no new tables.

### 2.7 Promote and copy-to-mine

- `POST /{id}/promote` (user → household): owner-only single UPDATE clearing
  `user_id`. 409 if the dashboard is already household-scoped.
- `POST /{id}/copy-to-mine` (household → user): source must be a visible
  household-scoped dashboard. Creates a new row with a new `id`,
  `user_id = caller`, `name = source.name + " (mine)"`,
  `sort_order = max(sort_order) + 1 WHERE user_id = caller`, and a **deep copy**
  of `source.layout`. During the copy, every widget instance's `id` is
  regenerated (new UUIDs) so future per-instance features stay clean. Done on
  the backend, not the frontend, so the client never has to know.

### 2.8 Observability

Metrics (Prometheus):

- `dashboard_save_total{result}` — counter
- `dashboard_save_duration_seconds` — histogram
- `dashboard_validation_failure_total{code}` — counter labeled by the error code
- `dashboard_seed_total{result}` — counter (`created` | `existed`)
- `dashboard_user_delete_cascade_total` — counter

Structured logs on every mutation include `tenant_id`, `household_id`,
`user_id`, `dashboard_id`.

### 2.9 Infrastructure

Kafka is treated like Postgres — an externally-managed platform component.
The k3s deployment at `kafka-broker.kafka.svc.cluster.local:9092` (and
`192.168.23.230:9092` from host/docker-compose) is the shared broker.

- `dashboard-service` and `account-service` receive `BOOTSTRAP_SERVERS` and the
  per-domain topic env var `EVENT_TOPIC_USER_LIFECYCLE=home-hub.user.lifecycle`.
- Auto-create topics is on at the broker, so services do not assert topics
  exist at startup.
- `kafka-ui.bee.tumidanski` is available for manual inspection during
  development and integration tests.
- `scripts/local-up.sh` exports `BOOTSTRAP_SERVERS` into the compose env; no
  Kafka container is added to `docker-compose.yml` — the broker is assumed
  to be running externally. A new compose block for `dashboard-service`
  itself (image, env, port, DB wait) is added, matching the existing
  per-service blocks.
- Ingress / nginx — `/api/v1/dashboards*` routes to the new service.
  `/api/v1/household-preferences*` stays on the account-service ingress rule
  (no new ingress change needed beyond the existing prefix routing).
- CI — dashboard-service joins the dynamic docker-build matrix automatically
  (zero-config with the recent aggregator work).

---

## 3. Frontend architecture

### 3.1 Widget registry module

```
frontend/src/lib/dashboard/
├── widget-types.ts               hand-mirrored string allowlist + parity test
├── widget-registry.ts            array of WidgetDefinition<T>
├── parse-config.ts               tolerant-read helper
├── seed-layout.ts                "Home" seed JSON (data-model.md §6)
├── schema.ts                     Zod schemas for Layout + WidgetInstance
└── widgets/
    ├── weather.ts
    ├── tasks-summary.ts
    ├── reminders-summary.ts
    ├── overdue-summary.ts
    ├── meal-plan-today.ts
    ├── calendar-today.ts
    ├── packages-summary.ts
    ├── habits-today.ts
    └── workout-today.ts
```

`WidgetDefinition<TConfig>` — see `data-model.md §3` for the canonical shape.
`widgetRegistry` is typed `readonly WidgetDefinition<unknown>[]` externally so
consumers iterate without a per-type existential escape hatch; internally each
entry is constructed with its real `TConfig`.

`parseConfig` implements tolerant-read:

```ts
export function parseConfig<T>(def: WidgetDefinition<T>, raw: unknown):
  { config: T; lossy: boolean }
{
  const merged = { ...def.defaultConfig, ...(isObject(raw) ? raw : {}) };
  const result = def.configSchema.safeParse(merged);
  return result.success
    ? { config: result.data, lossy: false }
    : { config: def.defaultConfig, lossy: true };
}
```

### 3.2 Widget component reuse

The nine registered widgets reuse the existing feature components rendered
by today's `DashboardPage`. Six widgets map directly:

- `weather` → `WeatherWidget`
- `meal-plan-today` → `MealPlanWidget` (accepts new optional `horizonDays` prop)
- `calendar-today` → `CalendarWidget` (accepts new `horizonDays`, `includeAllDay` props)
- `habits-today` → `HabitsWidget`
- `workout-today` → `WorkoutWidget`
- `packages-summary` → `PackageSummaryWidget`

Three widgets are new thin wrappers that encapsulate the inline `<Card>` blocks
currently living inside `DashboardPage.tsx`:

- `tasks-summary` — new component; uses `useTaskSummary`; accepts
  `{ status: "pending" | "overdue" | "completed", title?: string }`.
- `reminders-summary` — new component; uses `useReminderSummary`; accepts
  `{ filter: "active" | "snoozed" | "upcoming", title?: string }`.
- `overdue-summary` — new component; optional `title`; internally equivalent
  to `tasks-summary` with `status: "overdue"`.

Where an existing widget's data hook hardcodes a value that the PRD's config
exposes (e.g. `MealPlanWidget` always fetching 1 day), the hook gains a
parameter and the widget wires it through. Widgets with no real config still
accept an optional `title` prop.

### 3.3 Routes

`frontend/src/App.tsx` changes:

```tsx
// removed: <Route index element={<DashboardPage />} />
<Route path="dashboard" element={<DashboardRedirect />} />              // legacy
<Route path="dashboards" element={<DashboardsIndexRedirect />} />       // → default
<Route path="dashboards/:dashboardId" element={<DashboardShell />}>
  <Route index element={<DashboardRenderer />} />
  <Route path="edit" element={<DashboardDesigner />} />                 // React.lazy
</Route>
```

- `<DashboardShell>` — parent route that owns `useDashboard(dashboardId)` and
  the shared page header. Renders `<Outlet/>`. Caches the dashboard across the
  view ↔ edit transition.
- `<DashboardRenderer>` — CSS Grid based, no `react-grid-layout` import on this
  path. Grid math: `grid-template-columns: repeat(12, 1fr)` with per-widget
  inline-style `grid-column: span Xw; grid-row: span Yh`. Below the `md`
  breakpoint, widgets stack in a single column sorted by `(y, x)`.
- `<DashboardDesigner>` — React.lazy-loaded so the renderer bundle does not
  pull in `react-grid-layout`.
- `<DashboardRedirect>` — reads `household-preferences.defaultDashboardId` +
  the dashboards list; applies the PRD §4.6 fallback order; triggers seed
  when the household has zero household-scoped dashboards; `navigate(...,
  { replace: true })`. Shows `DashboardSkeleton` while deciding.
- `<DashboardsIndexRedirect>` — thin wrapper that runs the same fallback
  order as `<DashboardRedirect>` from the `/dashboards` bare path.

### 3.4 Designer interactions

Built around `react-grid-layout`'s `<GridLayout>`:

- **Palette drawer** — right-side shadcn `Sheet`; lists every registry entry
  (name + description + icon). Entries are draggable; dropping on the grid
  calls `GridLayout.onDrop` with the widget's `defaultSize` and `defaultConfig`.
- **Per-widget edit chrome** — an absolute-positioned overlay inside each
  grid item: drag handle (top-left), gear icon (top-right), trash icon
  (top-right). Trash is non-destructive (no confirm; save is explicit).
- **Config panel** — a Sheet opened by the gear icon. Renders a form from
  that widget's Zod schema using `react-hook-form` + `zodResolver`. A small
  recursive `<ZodForm>` component handles enum/boolean/string/number/nested
  object per `ux-flow.md`. `Apply` validates + merges into the draft; `Cancel`
  drops widget-local edits; `Reset to defaults` loads `defaultConfig`.
- **Grid bounds** — `cols={12}`, `isBounded={true}`, `compactType="vertical"`
  (collisions push downward); per-widget `minW/minH/maxW/maxH` come from the
  registry.

### 3.5 Designer state

A single `useReducer` in `<DashboardDesigner>`:

```ts
type DraftState = {
  name: string;
  layout: Layout;                    // version: 1, widgets: WidgetInstance[]
  dirty: boolean;
  selectedWidgetId: string | null;
  paletteOpen: boolean;
};

type DraftAction =
  | { type: "rename";          name: string }
  | { type: "move-or-resize";  widgets: WidgetInstance[] }
  | { type: "add";             widget: WidgetInstance }
  | { type: "remove";          id: string }
  | { type: "update-config";   id: string; config: unknown }
  | { type: "reset";           server: Dashboard }
  | { type: "saved";           server: Dashboard };
```

`dirty` flips on any mutation and is cleared by `saved` and `reset`.
`react-grid-layout`'s `onLayoutChange` is batched into a single
`move-or-resize` action, isolating the reducer from the library's callback
shape. Save is one `PATCH /dashboards/{id}` with `{ name, layout }`.

### 3.6 Dirty-state guard

- **In-app**: `useBlocker` from `react-router-dom`. While
  `draft.dirty && !savedIntent`, navigation is blocked and an `AlertDialog`
  offers Save / Discard / Keep editing. Save triggers the mutation then
  allows nav; Discard clears dirty; Keep editing cancels.
- **Tab close**: a `beforeunload` listener installed *only while dirty*,
  uninstalled on save or discard. Shows the browser's native confirm.

### 3.7 Sidebar integration

`<DashboardsNavGroup>` replaces the static "Dashboard" entry. Lists every
dashboard from `useDashboards()`:

- Grouped: household-scoped first, user-scoped second, divider between.
- Per entry: icon (home vs. person), name, kebab (Rename / Set as default /
  Delete / Promote | Copy-to-mine). Active highlighted via `useParams()`.
- `+ New dashboard` row at the bottom opens a modal: name, scope radio,
  optional "Copy of" dropdown. Create → navigate to
  `/app/dashboards/{newId}/edit`.
- Drag-to-reorder within each scope via `@dnd-kit/sortable` (lighter than
  `react-grid-layout` for vertical lists). Release persists via
  `PATCH /dashboards/order` with a single-scope payload.

### 3.8 First-load seeding

`<DashboardRedirect>` and `<DashboardsIndexRedirect>` both run:

```
list = GET /dashboards
if list has zero household-scoped dashboards:
  resp = POST /dashboards/seed { name: "Home", layout: seedLayout() }
  // 201 = created new; 200 = idempotent no-op, re-read list
  navigate to the first household dashboard
else:
  apply default-preference resolution (PRD §4.6)
```

`seedLayout()` lives in `lib/dashboard/seed-layout.ts` and is an exact
translation of `data-model.md §6` (nine widgets, precise cells, precise
configs). A frontend test renders it through `<DashboardRenderer>` and
asserts widget order/props match the legacy `DashboardPage`.

### 3.9 `DashboardPage.tsx` deletion

The legacy page is deleted once the seeded renderer demonstrates parity.
Before deletion, its three inline `<Card>` blocks are extracted into
`components/features/dashboard-widgets/{tasks-summary,reminders-summary,
overdue-summary}.tsx` — the new registry's widgets.

### 3.10 Responsive behavior

- **Renderer** — always responsive. At `≥ md` (768px): true 12-column
  CSS Grid. At `< md`: single-column stack, order by `(y, x)`.
- **Designer** — below `md`, the route renders a blocker screen
  ("Editing is only available on tablet-or-wider screens") + a "View only"
  button linking back to view mode. The `Edit` button in the view-mode header
  is disabled with a tooltip below 768px.

### 3.11 Schema-version gate

If a loaded dashboard's `schema_version` exceeds the frontend constant
`LAYOUT_SCHEMA_VERSION` (currently 1), the renderer shows a full-page empty
state ("This dashboard was made with a newer version of the app — please
refresh your browser") and mounts no widgets. Prevents mystery failures when
the backend ships v2 ahead of a user's cached frontend.

### 3.12 Unknown widget types

When a persisted layout references a `type` string not present in the
frontend registry (e.g. the widget was removed after a downgrade, or the
frontend cache is older than the backend), the renderer mounts a labeled
placeholder component (`<UnknownWidgetPlaceholder type="foo" />`) in that
grid cell rather than crashing. The placeholder shows the literal type
string and a short explanation ("Widget type 'foo' is no longer available").
In edit mode, the same placeholder renders inside the draggable chrome —
the user can delete or relocate the instance but cannot configure it (the
gear icon is disabled with a tooltip).

Backend validation (§2.5) rejects new / updated layouts referencing unknown
types, so this code path only executes for layouts that were valid at
persistence time but are unknown to *this* client build.

### 3.13 Lossy-config badge

When `parseConfig` returns `lossy: true` for a widget, the renderer decorates
that widget with a small badge icon in its chrome and a hover tooltip
explaining the config was normalized. In edit mode, opening the gear panel
re-normalizes; saving clears the flag (it is a render-time derivation, not
persisted).

---

## 4. Cross-cutting concerns

### 4.1 Retention framework

`services/dashboard-service/internal/retention/wire.go` registers a retention
category named `dashboards` per the `shared/go/retention` pattern used by
package-service and others. In v1 **no window handlers** are registered — the
PRD explicitly allows zero-windowed retention. `account-service`'s retention
policy config gains a `dashboards` category entry with a default window of 0
(never auto-purge), so future time-based policies land without a schema change.

Account-delete cascade is handled by the Kafka consumer, not retention.
Retention is for time-windowed cleanup; cascade is for referential integrity.

### 4.2 Testing strategy

**Backend — `dashboard-service`**

- `internal/layout/validator_test.go` — every PRD §4.9 rule gets happy + failure
  cases with the stable error codes asserted.
- `internal/dashboard/processor_test.go` — create / update / delete / promote /
  copy-to-mine / reorder logic against a mocked administrator + provider.
- `internal/dashboard/rest_test.go` — integration tests using the existing
  service-test harness (Postgres + the service's router):
    - scope visibility (user-scoped rows invisible to other members);
    - household-scoped editability across members;
    - promote + copy-to-mine semantics (new id, widget instance-ids regenerated);
    - seed idempotency across two sequential calls;
    - seed race across two parallel goroutines (advisory lock);
    - reorder with mixed-scope rejected (400);
    - every §4.9 validation path returns the right JSON:API error code.
- `internal/events/handler_user_deleted_test.go` — synthetic event → DELETE
  verified; idempotent on replay.

**Backend — `account-service`**

- CRUD + visibility tests for `household_preferences` (mirror existing
  `preference` tests).
- User-delete flow test: preferences rows removed inside the same Tx; a
  `UserDeletedEvent` is produced to the expected topic.

**Frontend**

- `widget-types.test.ts` — parity against a shared fixture also asserted by the
  Go test.
- Renderer parity snapshot — `<DashboardRenderer>` with `seedLayout()`
  matches the legacy `DashboardPage`'s widget layout by data (not pixels).
- `parse-config.test.ts` — tolerant-read fallback behavior.
- Reducer tests for `DraftState`/`DraftAction` — no DOM needed.
- One integration-level test per designer interaction (add/move/resize/config
  apply/save) using `@testing-library/react` against the real reducer and a
  mocked grid callback.

Playwright / E2E is out of scope for v1; the repo does not currently run it.

### 4.3 Migrations

- `account-service` — `household_preferences` via the existing AutoMigrate chain.
- `dashboard-service` — `dashboards` table + scope index + partial index via
  AutoMigrate.
- `shared/go/retention` — `sr.MigrateRuns(db)` runs as it does in other services.
- No data backfill; seeding is per-household and runs lazily on first visit.

### 4.4 Rollout

Pure additive; no feature flag needed. Sequence:

1. Ship `dashboard-service` + `household_preferences` + Kafka plumbing,
   wired but with no frontend caller. Silent landing.
2. Ship frontend registry + renderer + routes + seeding, with `/app/dashboard`
   still routed to the old `DashboardPage`.
3. Manually verify parity in staging: seed a household, visit
   `/app/dashboards/{id}`, compare against the old page.
4. Flip `/app/dashboard` to `<DashboardRedirect>`; delete `DashboardPage.tsx`.
5. Verify in staging again; merge.

Steps 1–2 can be separate commits; step 4 is the cutover and is reversible by
a single-file revert.

### 4.5 Documentation

Per the DOCS.md contract:

- `services/dashboard-service/docs/domain.md` — entity, scoping, validation.
- `services/dashboard-service/docs/rest.md` — endpoint reference.
- `services/dashboard-service/docs/storage.md` — table, indexes, advisory lock.
- `services/account-service/docs/domain.md` — updated with `household_preferences`.
- `docs/architecture.md` — short section on the new Kafka event bus,
  `shared/go/kafka`, and the `UserDeletedEvent` flow, so future cross-service
  cascades follow the same pattern.

### 4.6 Risks and mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Kafka crash between user-delete commit and event produce → orphan dashboards | Low | Accepted per decision #2. Graduate to transactional outbox when a second event type appears. |
| `react-grid-layout` maintenance stalls | Low-medium | Persisted JSON is library-agnostic; swap is contained to the designer bundle. |
| Widget registry drift between Go and TS | Medium | Parity tests on both sides against a committed fixture. |
| Tampered client submits huge `config` | Low | Backend caps in `layout.Validate`. |
| Two tabs seeding simultaneously | Low | Advisory lock + idempotent return. |
| Renderer crashes on unknown widget type (post-downgrade) | Low | Registry lookup returns undefined → placeholder component. |
| User deletes their currently-open dashboard | Medium | Post-delete, frontend clears local belief about default and re-runs resolution; navigates to fallback. |
| `household_preferences` row leaks after household delete | Low | Out of scope for v1; follow-up using the same Kafka pattern with `HouseholdDeletedEvent`. |
| `beforeunload` prompts misfiring on saved state | Low | Listener is installed only while `dirty === true`. |
| Single-partition topic limits consumer scale-out later | Low | Fine for v1 (one consumer). Future: pre-create topics with more partitions; consumer key is `userID`, safe to partition by. |

### 4.7 Deferred to the plan phase

The following are implementation details small enough for the plan phase to
decide without more architectural context:

- Exact `<ZodForm>` implementation (recursive discriminated-union walk vs.
  per-widget hand-written form). The interface is settled; the implementation
  is a plan-phase call per widget.
- Whether sidebar drag-reorder uses `@dnd-kit/sortable` or an equivalent
  simpler library (interchangeable for this use case).
- Prometheus scrape config changes for the new service (mechanical, follows
  the existing pattern).
- The exact sr.NewPolicyClient wiring for `dashboard-service` if the
  account-service policy endpoint needs a new category entry shipped
  alongside.

---

## 5. Acceptance mapping

Each PRD §10 criterion maps to the design's provisions:

| PRD acceptance | Design provision |
|---|---|
| dashboard-service exists with schema + endpoints + tests | §2.1, §4.2 |
| docker-compose / ingress / build pipeline wired | §2.9 |
| Empty → seed flow on fresh household | §3.8 |
| Seeded "Home" visually matches current DashboardPage | §3.8 parity test |
| DashboardPage.tsx deleted, legacy route redirects | §3.9, §3.3 |
| Household + user scoped creation | §2.1 create endpoint |
| Sidebar lists dashboards grouped by scope | §3.7 |
| Drag-reorder persists via bulk endpoint | §3.7 |
| Edit mode chrome on tablet+ | §3.4, §3.10 |
| Palette add with defaults | §3.4 |
| Resize enforces min/max + 12-col | §3.4 |
| Zod-rendered config panel with inline validation | §3.4 |
| Save / Discard / unsaved-warn | §3.6 |
| All 9 widgets render via registry | §3.1, §3.2 |
| User-scoped invisibility + household editability | §2.1, §4.2 |
| Default-dashboard deletion clears preference | §3.8, PRD §4.6 |
| copy-to-mine sort_order = max+1 | §2.7 |
| Account delete cascade | §2.3, §2.4 |
| Responsive renderer + disabled edit below md | §3.10 |
| Unknown widget type placeholder | §3.12 |
| Backend rejections with stable codes | §2.5 |
| Retention category registered | §4.1 |
| Per-service docs files exist | §4.5 |
