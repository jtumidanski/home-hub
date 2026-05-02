# Kiosk Dashboard Widgets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add five new dashboard widgets (`tasks-today`, `reminders-today`, `weather-tomorrow`, `calendar-tomorrow`, `tasks-tomorrow`), extend `meal-plan-today` with a `view: "today-detail"` config, and seed a second household dashboard ("Kiosk") that arranges them into a 4-column kiosk-style layout — without reviving the deprecated `apps/kiosk` app.

**Architecture:** Three edit zones. (1) Cross-service widget allowlist in `shared/go/dashboard/types.go` + the TS mirror — additive only. (2) Backend per-key seeding via a new `seed_key` column on `dashboards` plus a reworked `Processor.Seed`. (3) Frontend widget definitions/adapters using existing query hooks (no new APIs), plus a new `kiosk-seed-layout.ts` and a `DashboardRedirect.tsx` that issues two seed calls gated by the new `kiosk_dashboard_seeded` boolean on `household_preferences`.

**Tech Stack:** Go 1.22+ (GORM + Postgres + sqlite-in-memory tests), TypeScript + React + Vite + Vitest + React Testing Library, TanStack React Query v5, Zod, JSON:API via api2go.

---

## Pre-flight

- [ ] **Verify clean baseline.** From repo root, run `git status --short` and confirm only `docs/tasks/task-046-kiosk-dashboard-widgets/` is touched (the task folder). Run `go test ./shared/go/dashboard/... ./services/dashboard-service/... ./services/account-service/...` and `pnpm --filter frontend test --run` and confirm both pass before starting.

---

## Phase A — Cross-service widget allowlist

### Task A1: Update Go widget-type fixture and parity test

**Files:**
- Modify: `shared/go/dashboard/fixtures/widget-types.json`
- Modify: `shared/go/dashboard/types_test.go:11-18`

- [ ] **Step 1: Extend the parity test to assert the new types are recognized.**

Edit `shared/go/dashboard/types_test.go:11-22` so the type list in `TestIsKnownWidgetType` includes the five new strings:

```go
func TestIsKnownWidgetType(t *testing.T) {
	for _, typ := range []string{
		"weather", "tasks-summary", "reminders-summary", "overdue-summary",
		"meal-plan-today", "calendar-today", "packages-summary",
		"habits-today", "workout-today",
		"tasks-today", "reminders-today",
		"weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow",
	} {
		if !IsKnownWidgetType(typ) {
			t.Fatalf("expected %q to be known", typ)
		}
	}
	if IsKnownWidgetType("foo") {
		t.Fatalf("expected unknown type to be rejected")
	}
}
```

- [ ] **Step 2: Run the test — expect failure.**

Run: `go test ./shared/go/dashboard/ -run TestIsKnownWidgetType`
Expected: FAIL with `expected "tasks-today" to be known`.

- [ ] **Step 3: Add the five strings to the Go map.**

In `shared/go/dashboard/types.go`, replace the `WidgetTypes` map literal with:

```go
var WidgetTypes = map[string]struct{}{
	"weather":           {},
	"tasks-summary":     {},
	"reminders-summary": {},
	"overdue-summary":   {},
	"meal-plan-today":   {},
	"calendar-today":    {},
	"packages-summary":  {},
	"habits-today":      {},
	"workout-today":     {},
	"tasks-today":       {},
	"reminders-today":   {},
	"weather-tomorrow":  {},
	"calendar-tomorrow": {},
	"tasks-tomorrow":    {},
}
```

- [ ] **Step 4: Update the JSON fixture.**

Replace the contents of `shared/go/dashboard/fixtures/widget-types.json` with the alphabetised superset:

```json
[
  "calendar-today",
  "calendar-tomorrow",
  "habits-today",
  "meal-plan-today",
  "overdue-summary",
  "packages-summary",
  "reminders-summary",
  "reminders-today",
  "tasks-summary",
  "tasks-today",
  "tasks-tomorrow",
  "weather",
  "weather-tomorrow",
  "workout-today"
]
```

- [ ] **Step 5: Re-run all the dashboard package tests.**

Run: `go test ./shared/go/dashboard/...`
Expected: all tests PASS (`TestIsKnownWidgetType`, `TestWidgetTypesParityFixture`, `TestLayoutConstants`).

- [ ] **Step 6: Commit.**

```bash
git add shared/go/dashboard/types.go shared/go/dashboard/fixtures/widget-types.json shared/go/dashboard/types_test.go
git commit -m "feat(dashboard): allowlist 5 task-046 widget types in shared/go/dashboard"
```

### Task A2: Update TS allowlist + fixture

**Files:**
- Modify: `frontend/src/lib/dashboard/widget-types.ts`
- Modify: `frontend/src/lib/dashboard/fixtures/widget-types.json`
- Test: `frontend/src/lib/dashboard/__tests__/widget-types.test.ts`

- [ ] **Step 1: Read the existing widget-types test to know the shape it expects.**

Run: `cat frontend/src/lib/dashboard/__tests__/widget-types.test.ts`

(If the test asserts the array length or specific entries, update it first to expect 14.)

- [ ] **Step 2: Add the 5 new strings to the TS allowlist.**

Replace `frontend/src/lib/dashboard/widget-types.ts` with:

```ts
export const WIDGET_TYPES = [
  "weather",
  "tasks-summary",
  "reminders-summary",
  "overdue-summary",
  "meal-plan-today",
  "calendar-today",
  "packages-summary",
  "habits-today",
  "workout-today",
  "tasks-today",
  "reminders-today",
  "weather-tomorrow",
  "calendar-tomorrow",
  "tasks-tomorrow",
] as const;

export type WidgetType = (typeof WIDGET_TYPES)[number];

export function isKnownWidgetType(t: string): t is WidgetType {
  return (WIDGET_TYPES as readonly string[]).includes(t);
}

export const LAYOUT_SCHEMA_VERSION = 1;
export const GRID_COLUMNS = 12;
export const MAX_WIDGETS = 40;
```

- [ ] **Step 3: Mirror the JSON fixture so the parity test the TS suite uses sees the same set.**

Replace `frontend/src/lib/dashboard/fixtures/widget-types.json` with the same 14-entry alphabetised array used in Task A1 step 4.

- [ ] **Step 4: Update or add the TS parity assertion.**

Open `frontend/src/lib/dashboard/__tests__/widget-types.test.ts`. If a parity test exists comparing `WIDGET_TYPES` against the fixture, ensure it passes. If `widget-types.test.ts` only asserts `isKnownWidgetType` behavior, append:

```ts
import widgetTypesFixture from "@/lib/dashboard/fixtures/widget-types.json";

it("WIDGET_TYPES matches the parity fixture", () => {
  expect([...WIDGET_TYPES].sort()).toEqual([...widgetTypesFixture].sort());
});
```

- [ ] **Step 5: Run the TS allowlist tests.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widget-types.test.ts`
Expected: PASS.

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/lib/dashboard/widget-types.ts frontend/src/lib/dashboard/fixtures/widget-types.json frontend/src/lib/dashboard/__tests__/widget-types.test.ts
git commit -m "feat(dashboard): allowlist 5 task-046 widget types in frontend mirror"
```

---

## Phase B — Backend per-key seeding (`dashboard-service`)

### Task B1: Add `SeedKey` column + partial unique index + brownfield backfill

**Files:**
- Modify: `services/dashboard-service/internal/dashboard/entity.go`
- Test: `services/dashboard-service/internal/dashboard/processor_test.go` (new test added in Task B2)

- [ ] **Step 1: Add the field and migrations to the entity.**

Replace the contents of `services/dashboard-service/internal/dashboard/entity.go` with:

```go
package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Entity struct {
	Id            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	HouseholdId   uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	UserId        *uuid.UUID     `gorm:"type:uuid;index:idx_dashboards_scope"`
	Name          string         `gorm:"type:varchar(80);not null"`
	SortOrder     int            `gorm:"not null;default:0"`
	Layout        datatypes.JSON `gorm:"type:jsonb;not null"`
	SchemaVersion int            `gorm:"not null;default:1"`
	SeedKey       *string        `gorm:"column:seed_key;type:varchar(40)"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
}

func (Entity) TableName() string { return "dashboards" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_dashboards_household_partial
		ON dashboards (tenant_id, household_id) WHERE user_id IS NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboards_seed_key
		ON dashboards (tenant_id, household_id, seed_key) WHERE seed_key IS NOT NULL`).Error; err != nil {
		return err
	}
	// Brownfield backfill — claim existing seeded "Home" rows so the
	// updated client's idempotent home-seed call is a no-op for them.
	return db.Exec(`UPDATE dashboards
		SET seed_key = 'home'
		WHERE seed_key IS NULL
		  AND user_id IS NULL
		  AND sort_order = 0
		  AND name = 'Home'`).Error
}
```

- [ ] **Step 2: Confirm the package still compiles.**

Run: `go build ./services/dashboard-service/...`
Expected: success.

- [ ] **Step 3: Confirm existing tests still pass (entity build only — processor changes come next).**

Run: `go test ./services/dashboard-service/internal/dashboard/ -run TestProcessorCreateHousehold`
Expected: PASS.

- [ ] **Step 4: Commit.**

```bash
git add services/dashboard-service/internal/dashboard/entity.go
git commit -m "feat(dashboard-service): add seed_key column and brownfield backfill"
```

### Task B2: Rework `Processor.Seed` for optional `seedKey`

**Files:**
- Modify: `services/dashboard-service/internal/dashboard/processor.go`
- Modify: `services/dashboard-service/internal/dashboard/processor_test.go`
- Modify: `services/dashboard-service/internal/dashboard/resource.go` (handler call site only)

- [ ] **Step 1: Write a failing test that seeds with key="home" twice and expects idempotency by key.**

Append to `services/dashboard-service/internal/dashboard/processor_test.go`:

```go
func TestProcessorSeedByKeyIdempotent(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)
	homeKey := "home"

	first, err := p.Seed(tid, hid, uid, "Home", &homeKey, layoutJSON)
	require.NoError(t, err)
	require.True(t, first.Created)

	second, err := p.Seed(tid, hid, uid, "Home", &homeKey, layoutJSON)
	require.NoError(t, err)
	require.False(t, second.Created, "second seed with same key must not create")
	require.Len(t, second.Existing, 1)
	require.Equal(t, first.Dashboard.Id(), second.Existing[0].Id())
}

func TestProcessorSeedDistinctKeysCoexist(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)
	homeKey, kioskKey := "home", "kiosk"

	home, err := p.Seed(tid, hid, uid, "Home", &homeKey, layoutJSON)
	require.NoError(t, err)
	require.True(t, home.Created)

	kiosk, err := p.Seed(tid, hid, uid, "Kiosk", &kioskKey, layoutJSON)
	require.NoError(t, err)
	require.True(t, kiosk.Created, "distinct seedKey must not collide with home")
	require.NotEqual(t, home.Dashboard.Id(), kiosk.Dashboard.Id())
}

func TestProcessorSeedNilKeyPreservesLegacyBehavior(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	first, err := p.Seed(tid, hid, uid, "Home", nil, layoutJSON)
	require.NoError(t, err)
	require.True(t, first.Created)

	// Legacy nil-key seed: any household-scoped row already exists -> no-op.
	second, err := p.Seed(tid, hid, uid, "Home", nil, layoutJSON)
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Len(t, second.Existing, 1)
}
```

Also update the existing `TestProcessorSeedIdempotent` (lines 337-354) and `TestProcessorSeedRace` (lines 356-383) call sites to pass `nil` for the new arg:

```go
// In TestProcessorSeedIdempotent:
first, err := p.Seed(tid, hid, uid, "Home", nil, layoutJSON)
// ...
second, err := p.Seed(tid, hid, uid, "Home", nil, layoutJSON)

// In TestProcessorSeedRace inner goroutine:
res, err := p.Seed(tid, hid, uid, "Home", nil, layoutJSON)
```

- [ ] **Step 2: Run the new tests — expect compile failure.**

Run: `go test ./services/dashboard-service/internal/dashboard/ -run TestProcessorSeed`
Expected: FAIL — `Seed` signature mismatch (4 args expected, 5 given).

- [ ] **Step 3: Update `Processor.Seed` to accept the seed key.**

In `services/dashboard-service/internal/dashboard/processor.go`, replace the `Seed` method (lines 111-166) and the lock helpers (171-187) with:

```go
// Seed ensures at least one household-scoped dashboard exists for the given
// (tenant, household). Two modes:
//
//   - seedKey == nil  legacy behavior: idempotent only if any household-scoped
//                     row already exists. Used by clients that haven't been
//                     updated to send a key.
//   - seedKey != nil  idempotent per (tenant, household, *seedKey). Insert if
//                     no row with that key exists; otherwise return the row.
//                     The partial unique index `idx_dashboards_seed_key` is the
//                     race backstop.
func (p *Processor) Seed(
	tenantID, householdID, callerUserID uuid.UUID,
	name string,
	seedKey *string,
	layoutJSON json.RawMessage,
) (SeedResult, error) {
	name = trimName(name)
	if err := validateNameLen(name); err != nil {
		return SeedResult{}, err
	}
	if _, err := layout.Validate(layoutJSON); err != nil {
		return SeedResult{}, err
	}

	var out SeedResult
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		if err := p.acquireSeedLock(tx, tenantID, householdID, seedKey); err != nil {
			return err
		}
		if seedKey != nil {
			// Per-key path: look for a row already claimed by this key.
			var existing Entity
			err := tx.Where("tenant_id = ? AND household_id = ? AND seed_key = ?",
				tenantID, householdID, *seedKey).First(&existing).Error
			if err == nil {
				m, mErr := Make(existing)
				if mErr != nil {
					return mErr
				}
				out.Existing = []Model{m}
				out.Created = false
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			key := *seedKey
			e := Entity{
				TenantId:      tenantID,
				HouseholdId:   householdID,
				UserId:        nil,
				Name:          name,
				SortOrder:     0,
				Layout:        datatypes.JSON(layoutJSON),
				SchemaVersion: 1,
				SeedKey:       &key,
			}
			saved, sErr := insert(tx, e)
			if sErr != nil {
				return sErr
			}
			m, mErr := Make(saved)
			if mErr != nil {
				return mErr
			}
			out.Dashboard = m
			out.Created = true
			return nil
		}

		// Legacy nil-key path: any household-scoped row counts.
		count, err := countHouseholdScoped(tx, tenantID, householdID)
		if err != nil {
			return err
		}
		if count > 0 {
			list, err := visibleToCaller(tenantID, householdID, callerUserID)(tx)()
			if err != nil {
				return err
			}
			for _, e := range list {
				m, mErr := Make(e)
				if mErr != nil {
					return mErr
				}
				out.Existing = append(out.Existing, m)
			}
			out.Created = false
			return nil
		}
		e := Entity{
			TenantId:      tenantID,
			HouseholdId:   householdID,
			UserId:        nil,
			Name:          name,
			SortOrder:     0,
			Layout:        datatypes.JSON(layoutJSON),
			SchemaVersion: 1,
		}
		saved, err := insert(tx, e)
		if err != nil {
			return err
		}
		m, err := Make(saved)
		if err != nil {
			return err
		}
		out.Dashboard = m
		out.Created = true
		return nil
	})
	return out, err
}

// acquireSeedLock takes a Postgres advisory xact lock keyed on
// (tenant, household[, seedKey]) so concurrent seeders serialize on the same
// key. Sqlite tests rely on the surrounding transaction.
func (p *Processor) acquireSeedLock(tx *gorm.DB, tenantID, householdID uuid.UUID, seedKey *string) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var key int64
	if seedKey == nil {
		key = seedLockKey(tenantID, householdID)
	} else {
		key = seedLockKeyForKey(tenantID, householdID, *seedKey)
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", key).Error
}

func seedLockKey(tenantID, householdID uuid.UUID) int64 {
	var combined [32]byte
	copy(combined[:16], tenantID[:])
	copy(combined[16:], householdID[:])
	sum := sha256.Sum256(combined[:])
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

func seedLockKeyForKey(tenantID, householdID uuid.UUID, key string) int64 {
	combined := make([]byte, 32+len(key))
	copy(combined[:16], tenantID[:])
	copy(combined[16:32], householdID[:])
	copy(combined[32:], []byte(key))
	sum := sha256.Sum256(combined)
	return int64(binary.BigEndian.Uint64(sum[:8]))
}
```

- [ ] **Step 4: Update the REST handler call site to pass `nil` for now.**

In `services/dashboard-service/internal/dashboard/resource.go:326`, change:

```go
res, err := proc.Seed(t.Id(), t.HouseholdId(), t.UserId(), input.Name, input.Layout)
```

to:

```go
res, err := proc.Seed(t.Id(), t.HouseholdId(), t.UserId(), input.Name, nil, input.Layout)
```

(Task B3 will read the key from the request body.)

- [ ] **Step 5: Run all dashboard-service tests.**

Run: `go test ./services/dashboard-service/...`
Expected: all tests PASS, including the four new ones.

- [ ] **Step 6: Commit.**

```bash
git add services/dashboard-service/internal/dashboard/processor.go services/dashboard-service/internal/dashboard/processor_test.go services/dashboard-service/internal/dashboard/resource.go
git commit -m "feat(dashboard-service): rework Processor.Seed to accept optional seed key"
```

### Task B3: Accept `key` on the seed REST endpoint

**Files:**
- Modify: `services/dashboard-service/internal/dashboard/rest.go`
- Modify: `services/dashboard-service/internal/dashboard/resource.go`
- Modify: `services/dashboard-service/internal/dashboard/rest_test.go`

- [ ] **Step 1: Write a failing test that POSTs `attributes.key = "kiosk"` and expects 201 + the row carries `seed_key = 'kiosk'`.**

Read `services/dashboard-service/internal/dashboard/rest_test.go` to understand the test scaffolding (HTTP harness, headers, JSON:API helpers). Add at the bottom (using the existing helpers — `serverInfoForTest`, `authHeaders`, etc., whatever the file uses):

```go
func TestSeedHandlerWithKioskKey(t *testing.T) {
	db := setupTestDB(t) // or whatever test DB helper rest_test.go uses
	srv := newTestServer(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()

	body := []byte(`{
	  "data": {
	    "type": "dashboards",
	    "attributes": {
	      "name": "Kiosk",
	      "key": "kiosk",
	      "layout": {"version":1,"widgets":[]}
	    }
	  }
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboards/seed", bytes.NewReader(body))
	req = withTenantContext(req, tid, hid, uid)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var saved Entity
	require.NoError(t, db.Where("tenant_id = ? AND household_id = ?", tid, hid).First(&saved).Error)
	require.NotNil(t, saved.SeedKey)
	require.Equal(t, "kiosk", *saved.SeedKey)
}

func TestSeedHandlerRejectsMalformedKey(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()

	body := []byte(`{
	  "data": {
	    "type": "dashboards",
	    "attributes": {
	      "name": "X",
	      "key": "Has Spaces",
	      "layout": {"version":1,"widgets":[]}
	    }
	  }
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboards/seed", bytes.NewReader(body))
	req = withTenantContext(req, tid, hid, uid)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code, rr.Body.String())
	require.Contains(t, rr.Body.String(), "/data/attributes/key")
}
```

(If `rest_test.go` uses a different helper name for `newTestServer` / `withTenantContext`, mirror what other tests in that file use.)

- [ ] **Step 2: Run the new tests — expect failure (`SeedRequest.Key` doesn't exist).**

Run: `go test ./services/dashboard-service/internal/dashboard/ -run TestSeedHandlerWith`
Expected: FAIL with compile error or 422 vs 201 mismatch.

- [ ] **Step 3: Add `Key` to `SeedRequest`.**

In `services/dashboard-service/internal/dashboard/rest.go:65-73`, replace the `SeedRequest` definition with:

```go
type SeedRequest struct {
	Name   string          `json:"name"`
	Key    *string         `json:"key,omitempty"`
	Layout json.RawMessage `json:"layout"`
}

func (SeedRequest) GetName() string      { return "dashboards" }
func (SeedRequest) GetID() string        { return "" }
func (*SeedRequest) SetID(_ string) error { return nil }
```

- [ ] **Step 4: Validate the key in the handler and pass it to the processor.**

In `services/dashboard-service/internal/dashboard/resource.go`, add a regex constant near the top (before `InitializeRoutes`):

```go
var seedKeyRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}$`)
```

Add `"regexp"` to the imports if missing.

Replace the `seedHandler` body so it validates and forwards the key:

```go
func seedHandler(db *gorm.DB) server.InputHandler[SeedRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input SeedRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			// Treat empty string as omitted; reject malformed values.
			var key *string
			if input.Key != nil && *input.Key != "" {
				if !seedKeyRegex.MatchString(*input.Key) {
					server.WriteJSONAPIError(w, http.StatusUnprocessableEntity,
						"validation.invalid_field", "Invalid key",
						"key must match ^[a-z][a-z0-9-]{0,39}$",
						"/data/attributes/key")
					return
				}
				k := *input.Key
				key = &k
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			res, err := proc.Seed(t.Id(), t.HouseholdId(), t.UserId(), input.Name, key, input.Layout)
			if err != nil {
				var ve layout.ValidationError
				if errors.As(err, &ve) {
					writeLayoutError(w, ve)
					return
				}
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
					server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, "dashboard.name_invalid", "Invalid name", err.Error(), "/data/attributes/name")
					return
				}
				d.Logger().WithError(err).Error("seed dashboard")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			if res.Created {
				rest, err := Transform(res.Dashboard)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
				return
			}
			out := make([]RestModel, 0, len(res.Existing))
			for _, m := range res.Existing {
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				out = append(out, rest)
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(out)
		}
	}
}
```

- [ ] **Step 5: Re-run all dashboard-service tests.**

Run: `go test ./services/dashboard-service/...`
Expected: every test PASS (including the two new REST tests).

- [ ] **Step 6: Commit.**

```bash
git add services/dashboard-service/internal/dashboard/rest.go services/dashboard-service/internal/dashboard/resource.go services/dashboard-service/internal/dashboard/rest_test.go
git commit -m "feat(dashboard-service): accept optional seed key on POST /dashboards/seed"
```

---

## Phase C — `kiosk_dashboard_seeded` flag (`account-service`)

### Task C1: Add the boolean column and its accessor methods

**Files:**
- Modify: `services/account-service/internal/householdpreference/entity.go`
- Modify: `services/account-service/internal/householdpreference/builder.go`
- Modify: `services/account-service/internal/householdpreference/model.go`

- [ ] **Step 1: Add the column to the entity.**

Replace the relevant section of `services/account-service/internal/householdpreference/entity.go` (lines 10-22) with:

```go
type Entity struct {
	Id                    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId              uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	UserId                uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	HouseholdId           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	DefaultDashboardId    *uuid.UUID `gorm:"type:uuid"`
	KioskDashboardSeeded  bool       `gorm:"column:kiosk_dashboard_seeded;not null;default:false"`
	CreatedAt             time.Time  `gorm:"not null"`
	UpdatedAt             time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "household_preferences" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                   m.id,
		TenantId:             m.tenantID,
		UserId:               m.userID,
		HouseholdId:          m.householdID,
		DefaultDashboardId:   m.defaultDashboardID,
		KioskDashboardSeeded: m.kioskDashboardSeeded,
		CreatedAt:            m.createdAt,
		UpdatedAt:            m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetHouseholdID(e.HouseholdId).
		SetDefaultDashboardID(e.DefaultDashboardId).
		SetKioskDashboardSeeded(e.KioskDashboardSeeded).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
```

- [ ] **Step 2: Read the existing `model.go` and `builder.go` to mirror their style.**

Run: `cat services/account-service/internal/householdpreference/model.go services/account-service/internal/householdpreference/builder.go`

- [ ] **Step 3: Add the field, getter, and builder method.**

In `services/account-service/internal/householdpreference/model.go`, add `kioskDashboardSeeded bool` to the `Model` struct alongside `defaultDashboardID`, plus a `KioskDashboardSeeded() bool` getter.

In `services/account-service/internal/householdpreference/builder.go`, add a parallel `kioskDashboardSeeded bool` field to the builder, a `SetKioskDashboardSeeded(v bool) *Builder` method, and ensure `Build()` carries the field through.

(Mirror the exact pattern used by `defaultDashboardID` — same struct positions, same method signatures.)

- [ ] **Step 4: Compile check.**

Run: `go build ./services/account-service/...`
Expected: success.

- [ ] **Step 5: Commit.**

```bash
git add services/account-service/internal/householdpreference/
git commit -m "feat(account-service): add kiosk_dashboard_seeded column to household_preferences"
```

### Task C2: Expose the field on the existing `GET /household-preferences` response

**Files:**
- Modify: `services/account-service/internal/householdpreference/rest.go`
- Modify: `services/account-service/internal/householdpreference/rest_test.go`

- [ ] **Step 1: Add a failing test asserting the new field is present in the GET response.**

In `services/account-service/internal/householdpreference/rest_test.go`, locate `TestGetHouseholdPreferencesAutoCreates` and add a new test below it (using the same setup helpers):

```go
func TestGetHouseholdPreferencesIncludesKioskFlag(t *testing.T) {
	// Re-use whatever harness TestGetHouseholdPreferencesAutoCreates uses.
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/household-preferences", nil)
	req = withTenantContext(req, tid, hid, uid)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	require.Contains(t, rr.Body.String(), `"kiosk_dashboard_seeded":false`)
}
```

(Use the exact harness names that `TestGetHouseholdPreferencesAutoCreates` uses — read the test file to get them.)

- [ ] **Step 2: Run the new test — expect failure.**

Run: `go test ./services/account-service/internal/householdpreference/ -run TestGetHouseholdPreferencesIncludesKioskFlag`
Expected: FAIL — body does not contain the expected JSON key.

- [ ] **Step 3: Update the RestModel to expose the field.**

Replace the `RestModel` and `Transform` in `services/account-service/internal/householdpreference/rest.go` (lines 9-31):

```go
type RestModel struct {
	Id                   uuid.UUID  `json:"-"`
	DefaultDashboardId   *uuid.UUID `json:"default_dashboard_id"`
	KioskDashboardSeeded bool       `json:"kiosk_dashboard_seeded"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "householdPreferences" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:                   m.Id(),
		DefaultDashboardId:   m.DefaultDashboardID(),
		KioskDashboardSeeded: m.KioskDashboardSeeded(),
		CreatedAt:            m.CreatedAt(),
		UpdatedAt:            m.UpdatedAt(),
	}, nil
}
```

- [ ] **Step 4: Re-run the test.**

Run: `go test ./services/account-service/internal/householdpreference/ -run TestGetHouseholdPreferencesIncludesKioskFlag`
Expected: PASS.

- [ ] **Step 5: Run all account-service tests to confirm no regressions.**

Run: `go test ./services/account-service/...`
Expected: PASS.

- [ ] **Step 6: Commit.**

```bash
git add services/account-service/internal/householdpreference/rest.go services/account-service/internal/householdpreference/rest_test.go
git commit -m "feat(account-service): expose kiosk_dashboard_seeded on household_preferences GET"
```

### Task C3: Add `PATCH /household-preferences/{id}/kiosk-seeded` sub-route

**Files:**
- Modify: `services/account-service/internal/householdpreference/administrator.go`
- Modify: `services/account-service/internal/householdpreference/processor.go`
- Modify: `services/account-service/internal/householdpreference/resource.go`
- Modify: `services/account-service/internal/householdpreference/rest_test.go`

- [ ] **Step 1: Add a failing test that PATCHes the new sub-route and verifies the column flips to true.**

Append to `rest_test.go`:

```go
func TestMarkKioskSeededFlipsFlag(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()

	// Auto-create the row via GET.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/household-preferences", nil)
	getReq = withTenantContext(getReq, tid, hid, uid)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, getReq)
	require.Equal(t, http.StatusOK, rr.Code)

	// Read back the id from the GET response.
	var listResp struct {
		Data []struct {
			Id string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &listResp))
	require.Len(t, listResp.Data, 1)
	id := listResp.Data[0].Id

	patchReq := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+id+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":true}`)))
	patchReq = withTenantContext(patchReq, tid, hid, uid)
	patchReq.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, patchReq)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	require.Contains(t, rr.Body.String(), `"kiosk_dashboard_seeded":true`)

	// Idempotency: a second PATCH stays true.
	patchReq2 := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+id+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":true}`)))
	patchReq2 = withTenantContext(patchReq2, tid, hid, uid)
	patchReq2.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, patchReq2)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
}
```

- [ ] **Step 2: Run — expect 404 (route not registered).**

Run: `go test ./services/account-service/internal/householdpreference/ -run TestMarkKioskSeeded`
Expected: FAIL with 404.

- [ ] **Step 3: Add a SQL helper to flip the flag.**

Append to `services/account-service/internal/householdpreference/administrator.go`:

```go
func markKioskSeeded(db *gorm.DB, id uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	if err := db.Exec(
		"UPDATE household_preferences SET kiosk_dashboard_seeded = TRUE, updated_at = ? WHERE id = ?",
		now, id,
	).Error; err != nil {
		return Entity{}, err
	}
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
```

- [ ] **Step 4: Add the processor method.**

Append to `services/account-service/internal/householdpreference/processor.go`:

```go
// MarkKioskSeeded sets kiosk_dashboard_seeded = TRUE for the given row. Idempotent.
func (p *Processor) MarkKioskSeeded(id uuid.UUID) (Model, error) {
	e, err := markKioskSeeded(p.db.WithContext(p.ctx), id)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
```

- [ ] **Step 5: Add the route + handler.**

In `services/account-service/internal/householdpreference/resource.go`, add a route registration inside `InitializeRoutes` (right after the existing PATCH `/household-preferences/{id}` line):

```go
api.HandleFunc("/household-preferences/{id}/kiosk-seeded",
	rh("MarkKioskSeeded", markKioskSeededHandler(db))).Methods(http.MethodPatch)
```

Then append the handler:

```go
func markKioskSeededHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Body is plain JSON, not JSON:API: {"value": true}. We accept
				// only `true`; the flag is write-once-true.
				var body struct {
					Value bool `json:"value"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !body.Value {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "expected {\"value\":true}")
					return
				}
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.MarkKioskSeeded(id)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					d.Logger().WithError(err).Error("mark kiosk seeded")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
```

Add `"encoding/json"` and `"errors"` to imports if missing.

- [ ] **Step 6: Re-run the test.**

Run: `go test ./services/account-service/internal/householdpreference/ -run TestMarkKioskSeeded`
Expected: PASS.

- [ ] **Step 7: Run all account-service tests.**

Run: `go test ./services/account-service/...`
Expected: PASS.

- [ ] **Step 8: Commit.**

```bash
git add services/account-service/internal/householdpreference/
git commit -m "feat(account-service): add PATCH /household-preferences/:id/kiosk-seeded"
```

---

## Phase D — Frontend type + service plumbing

### Task D1: `seedDashboard` accepts optional `key`

**Files:**
- Modify: `frontend/src/services/api/dashboard.ts`
- Modify: `frontend/src/lib/hooks/api/use-dashboards.ts`

- [ ] **Step 1: Add the `key` parameter to `seedDashboard`.**

Replace the `seedDashboard` method in `frontend/src/services/api/dashboard.ts` (lines 70-82) with:

```ts
seedDashboard(
  tenant: { id: string },
  name: string,
  layout: Layout,
  key?: string,
): Promise<ApiResponse<Dashboard> | ApiListResponse<Dashboard>> {
  this.setTenant(tenant);
  const attributes: Record<string, unknown> = { name, layout };
  if (key !== undefined) attributes.key = key;
  return api.post<ApiResponse<Dashboard> | ApiListResponse<Dashboard>>(`/dashboards/seed`, {
    data: { type: "dashboards", attributes },
  });
},
```

- [ ] **Step 2: Forward `key` through `useSeedDashboard`.**

Replace `useSeedDashboard` in `frontend/src/lib/hooks/api/use-dashboards.ts` (lines 163-176) with:

```ts
export function useSeedDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ name, layout, key }: { name: string; layout: Layout; key?: string }) =>
      dashboardService.seedDashboard(tenant!, name, layout, key),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to seed dashboard"));
    },
  });
}
```

- [ ] **Step 3: Type-check.**

Run: `pnpm --filter frontend tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit.**

```bash
git add frontend/src/services/api/dashboard.ts frontend/src/lib/hooks/api/use-dashboards.ts
git commit -m "feat(frontend): seedDashboard + useSeedDashboard accept optional key"
```

### Task D2: HouseholdPreferences type adds `kioskDashboardSeeded`

**Files:**
- Modify: `frontend/src/types/models/dashboard.ts`
- Modify: `frontend/src/services/api/household-preferences.ts`
- Modify: `frontend/src/lib/hooks/api/use-household-preferences.ts`

- [ ] **Step 1: Extend `HouseholdPreferencesAttributes`.**

In `frontend/src/types/models/dashboard.ts`, replace the `HouseholdPreferencesAttributes` interface with:

```ts
export interface HouseholdPreferencesAttributes {
  defaultDashboardId: string | null;
  kioskDashboardSeeded: boolean;
  createdAt: string;
  updatedAt: string;
}
```

The Go REST shape uses `kiosk_dashboard_seeded` (snake_case). Confirm whether the existing API client transforms snake_case → camelCase. Inspect `frontend/src/services/api/base.ts` and `frontend/src/services/api/household-preferences.ts`. If a transformer exists, no further work is needed; if not, transform inline in the household-preferences service.

- [ ] **Step 2: If needed, add a transform in `household-preferences.ts`.**

Read `frontend/src/services/api/household-preferences.ts`. If the `getPreferences` method returns the raw response, map `kiosk_dashboard_seeded` to `kioskDashboardSeeded` in the response transform. If the existing `default_dashboard_id` field is already mapped to `defaultDashboardId`, replicate the same code path.

- [ ] **Step 3: Add a `markKioskSeeded` service method.**

Append to the `HouseholdPreferencesService` class in `frontend/src/services/api/household-preferences.ts`:

```ts
markKioskSeeded(tenant: { id: string }, id: string) {
  this.setTenant(tenant);
  return api.patch<ApiResponse<HouseholdPreferences>>(
    `/household-preferences/${id}/kiosk-seeded`,
    { value: true },
  );
}
```

(Mirror import patterns already used in this file. If `ApiResponse` and `HouseholdPreferences` aren't imported, add them.)

- [ ] **Step 4: Add a `useMarkKioskSeeded` hook.**

Append to `frontend/src/lib/hooks/api/use-household-preferences.ts`:

```ts
export function useMarkKioskSeeded() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => householdPreferencesService.markKioskSeeded(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: householdPreferencesKeys.all(tenant, household) });
    },
    // Silent failure mode: redirect doesn't toast — DashboardRedirect retries
    // on next load if the PATCH didn't stick. See plan Task H1.
  });
}
```

- [ ] **Step 5: Type-check.**

Run: `pnpm --filter frontend tsc --noEmit`
Expected: clean.

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/types/models/dashboard.ts frontend/src/services/api/household-preferences.ts frontend/src/lib/hooks/api/use-household-preferences.ts
git commit -m "feat(frontend): expose kioskDashboardSeeded + add useMarkKioskSeeded mutation"
```

### Task D3: `useLocalDateOffset` hook + `getLocalDateStrOffset` util

**Files:**
- Modify: `frontend/src/lib/date-utils.ts`
- Create: `frontend/src/lib/hooks/use-local-date-offset.ts`
- Create: `frontend/src/lib/hooks/__tests__/use-local-date-offset.test.ts`

- [ ] **Step 1: Write the failing date-utils test.**

Create `frontend/src/lib/__tests__/date-utils-offset.test.ts`:

```ts
import { describe, it, expect, vi, afterEach } from "vitest";
import { getLocalDateStrOffset } from "@/lib/date-utils";

describe("getLocalDateStrOffset", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns the next-day date in the given timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    expect(getLocalDateStrOffset("America/New_York", 1)).toBe("2026-05-02");
  });

  it("offset 0 matches today's date in the same timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    expect(getLocalDateStrOffset("UTC", 0)).toBe("2026-05-01");
  });

  it("crosses month boundaries correctly", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-31T12:00:00Z"));
    expect(getLocalDateStrOffset("UTC", 1)).toBe("2026-02-01");
  });
});
```

- [ ] **Step 2: Run — expect failure (`getLocalDateStrOffset` not exported).**

Run: `pnpm --filter frontend test --run src/lib/__tests__/date-utils-offset.test.ts`
Expected: FAIL.

- [ ] **Step 3: Implement `getLocalDateStrOffset`.**

Append to `frontend/src/lib/date-utils.ts`:

```ts
/**
 * Returns the date as YYYY-MM-DD that is `offsetDays` after today in the given
 * timezone. Day-string arithmetic (not timestamp arithmetic) — DST-safe.
 */
export function getLocalDateStrOffset(tz: string | undefined, offsetDays: number): string {
  const resolved = tz || Intl.DateTimeFormat().resolvedOptions().timeZone;
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: resolved,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const year = Number(parts.find((p) => p.type === "year")!.value);
  const month = Number(parts.find((p) => p.type === "month")!.value);
  const day = Number(parts.find((p) => p.type === "day")!.value);

  // Construct a Date at noon UTC on the resolved local date so adding days
  // never crosses a DST cliff in the source timezone, then re-format in tz.
  const anchor = new Date(Date.UTC(year, month - 1, day + offsetDays, 12, 0, 0));
  const out = new Intl.DateTimeFormat("en-CA", {
    timeZone: resolved,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(anchor);
  const y = out.find((p) => p.type === "year")!.value;
  const m = out.find((p) => p.type === "month")!.value;
  const d = out.find((p) => p.type === "day")!.value;
  return `${y}-${m}-${d}`;
}
```

- [ ] **Step 4: Run the date-utils test.**

Run: `pnpm --filter frontend test --run src/lib/__tests__/date-utils-offset.test.ts`
Expected: PASS.

- [ ] **Step 5: Add the hook.**

Create `frontend/src/lib/hooks/use-local-date-offset.ts`:

```ts
import { useCallback, useSyncExternalStore } from "react";
import { getLocalDateStrOffset } from "@/lib/date-utils";

const POLL_MS = 60_000;

/**
 * Returns the calendar date as YYYY-MM-DD `offsetDays` after today in the
 * given IANA timezone, polling every 60s to catch midnight transitions.
 * Mirrors `useLocalDate` but with an offset.
 */
export function useLocalDateOffset(tz: string | undefined, offsetDays: number): string {
  const subscribe = useCallback(
    (notify: () => void) => {
      const id = window.setInterval(notify, POLL_MS);
      return () => window.clearInterval(id);
    },
    [],
  );
  const getSnapshot = useCallback(
    () => getLocalDateStrOffset(tz, offsetDays),
    [tz, offsetDays],
  );
  return useSyncExternalStore(subscribe, getSnapshot);
}
```

- [ ] **Step 6: Add a smoke test for the hook.**

Create `frontend/src/lib/hooks/__tests__/use-local-date-offset.test.ts`:

```ts
import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";

describe("useLocalDateOffset", () => {
  afterEach(() => vi.useRealTimers());

  it("returns tomorrow's date in the given timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    const { result } = renderHook(() => useLocalDateOffset("UTC", 1));
    expect(result.current).toBe("2026-05-02");
  });
});
```

- [ ] **Step 7: Run both new tests.**

Run: `pnpm --filter frontend test --run src/lib/__tests__/date-utils-offset.test.ts src/lib/hooks/__tests__/use-local-date-offset.test.ts`
Expected: PASS.

- [ ] **Step 8: Commit.**

```bash
git add frontend/src/lib/date-utils.ts frontend/src/lib/hooks/use-local-date-offset.ts frontend/src/lib/hooks/__tests__/use-local-date-offset.test.ts frontend/src/lib/__tests__/date-utils-offset.test.ts
git commit -m "feat(frontend): add useLocalDateOffset + getLocalDateStrOffset"
```

---

## Phase E — New widget definitions and adapters

Each widget pair (definition + adapter) is a single task with its own commit. Tests live in `frontend/src/components/features/dashboard-widgets/__tests__/`.

### Task E1: `tasks-today` widget definition

**Files:**
- Create: `frontend/src/lib/dashboard/widgets/tasks-today.ts`
- Test: `frontend/src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`

- [ ] **Step 1: Write the failing schema test.**

Create `frontend/src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { tasksTodayWidget } from "@/lib/dashboard/widgets/tasks-today";

describe("tasks-today widget definition", () => {
  it("declares the registry metadata", () => {
    expect(tasksTodayWidget.type).toBe("tasks-today");
    expect(tasksTodayWidget.displayName).toBe("Today's Tasks");
    expect(tasksTodayWidget.dataScope).toBe("household");
    expect(tasksTodayWidget.defaultSize).toEqual({ w: 4, h: 4 });
    expect(tasksTodayWidget.minSize).toEqual({ w: 3, h: 2 });
    expect(tasksTodayWidget.maxSize).toEqual({ w: 6, h: 8 });
  });

  it("default config sets includeCompleted to true", () => {
    expect(tasksTodayWidget.defaultConfig).toEqual({ includeCompleted: true });
  });

  it("schema accepts a custom title and rejects long titles", () => {
    expect(tasksTodayWidget.configSchema.safeParse({ title: "Custom" }).success).toBe(true);
    expect(tasksTodayWidget.configSchema.safeParse({ title: "x".repeat(81) }).success).toBe(false);
  });

  it("schema applies includeCompleted default", () => {
    const parsed = tasksTodayWidget.configSchema.parse({});
    expect(parsed.includeCompleted).toBe(true);
  });
});
```

- [ ] **Step 2: Run the test — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`
Expected: FAIL.

- [ ] **Step 3: Create the widget definition.**

Create `frontend/src/lib/dashboard/widgets/tasks-today.ts`:

```ts
import { z } from "zod";
import { TasksTodayAdapter } from "@/components/features/dashboard-widgets/tasks-today-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  includeCompleted: z.boolean().default(true),
});

type Cfg = z.infer<typeof schema>;

export const tasksTodayWidget: WidgetDefinition<Cfg> = {
  type: "tasks-today",
  displayName: "Today's Tasks",
  description: "Overdue plus today's incomplete tasks",
  component: TasksTodayAdapter,
  configSchema: schema,
  defaultConfig: { includeCompleted: true },
  defaultSize: { w: 4, h: 4 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 8 },
  dataScope: "household",
};
```

(The adapter doesn't exist yet — Task E2 adds it. The widget definition will fail to import until then; that's expected. The schema test does not import the adapter so it can pass.)

- [ ] **Step 4: Run the schema test.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`
Expected: PASS (it imports the widget but TS allows the unresolved adapter import as long as the file compiles — if it fails on the import, proceed straight to E2 and re-run).

- [ ] **Step 5: Commit (no commit yet — pair with Task E2 below).**

(No commit; the adapter file is required for the build.)

### Task E2: `tasks-today` adapter component

**Files:**
- Create: `frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx`
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx`

- [ ] **Step 1: Write the failing adapter test.**

Create `frontend/src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TasksTodayAdapter } from "@/components/features/dashboard-widgets/tasks-today-adapter";

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useTasks: vi.fn(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));
vi.mock("@/lib/hooks/use-local-date", () => ({
  useLocalDate: () => "2026-05-01",
}));

import { useTasks } from "@/lib/hooks/api/use-tasks";

const renderAdapter = (config = { includeCompleted: true }) =>
  render(
    <MemoryRouter>
      <TasksTodayAdapter config={config} />
    </MemoryRouter>,
  );

const task = (
  id: string,
  title: string,
  status: "pending" | "completed",
  dueOn?: string,
  completedAt?: string,
) => ({
  id,
  type: "tasks",
  attributes: { title, status, dueOn, completedAt, rolloverEnabled: false, createdAt: "", updatedAt: "" },
});

describe("TasksTodayAdapter", () => {
  it("renders skeleton while loading", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({ data: undefined, isLoading: true, isError: false });
    const { container } = renderAdapter();
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });

  it("renders overdue and today's incomplete tasks", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Pay bill", "pending", "2026-04-30"),         // overdue
        task("b", "Walk dog", "pending", "2026-05-01"),         // today
        task("c", "Old done", "completed", "2026-04-29", "2026-04-29T09:00:00Z"),
      ] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText("Pay bill")).toBeInTheDocument();
    expect(screen.getByText("Walk dog")).toBeInTheDocument();
    expect(screen.getByText(/Overdue \(1\)/)).toBeInTheDocument();
  });

  it("shows the all-completed fallback when only completed-today tasks exist", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Done early", "completed", "2026-05-01", "2026-05-01T08:00:00Z"),
      ] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText(/All tasks completed/i)).toBeInTheDocument();
  });

  it("shows empty copy when nothing is due", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText(/No tasks for today/i)).toBeInTheDocument();
  });

  it("renders error banner without crashing", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    });
    renderAdapter();
    expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run — expect failure (adapter doesn't exist).**

Run: `pnpm --filter frontend test --run src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement the adapter.**

Create `frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx`:

```tsx
// Read-only widget — no mutations. See PRD §4.1.
import { Link } from "react-router-dom";
import { useTasks } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckSquare, AlertTriangle, ChevronRight } from "lucide-react";
import type { Task } from "@/types/models/task";

export interface TasksTodayConfig {
  title?: string | undefined;
  includeCompleted: boolean;
}

export function TasksTodayAdapter({ config }: { config: TasksTodayConfig }) {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading, isError } = useTasks();

  const title = config.title?.trim() || "Today's Tasks";

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-28" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load tasks</p></CardContent>
      </Card>
    );
  }

  const all = (data?.data ?? []) as Task[];
  const overdue = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn && t.attributes.dueOn < today);
  const todayTasks = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn === today);
  const completedToday = all.filter((t) =>
    t.attributes.status === "completed" &&
    t.attributes.completedAt &&
    t.attributes.completedAt.slice(0, 10) === today,
  );

  const showAllCompleted =
    config.includeCompleted &&
    todayTasks.length === 0 &&
    overdue.length === 0 &&
    completedToday.length > 0;

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/tasks" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/tasks"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent className="space-y-3">
        {overdue.length > 0 && (
          <section>
            <h4 className="text-xs font-medium text-destructive flex items-center gap-1">
              <AlertTriangle className="h-3 w-3" />
              Overdue ({overdue.length})
            </h4>
            <ul className="mt-1 space-y-1">
              {overdue.map((t) => (
                <li key={t.id} className="text-sm">
                  <Link to="/app/tasks" className="hover:underline">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        )}
        {todayTasks.length > 0 ? (
          <section>
            <h4 className="text-xs font-medium text-muted-foreground">Today</h4>
            <ul className="mt-1 space-y-1">
              {todayTasks.map((t) => (
                <li key={t.id} className="text-sm">
                  <Link to="/app/tasks" className="hover:underline">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        ) : showAllCompleted ? (
          <section className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground">All tasks completed!</p>
            <ul className="mt-1 space-y-1 opacity-70">
              {completedToday.map((t) => (
                <li key={t.id}>{t.attributes.title}</li>
              ))}
            </ul>
          </section>
        ) : overdue.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CheckSquare className="h-5 w-5" />
            <p className="text-sm">No tasks for today</p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 4: Run the adapter test.**

Run: `pnpm --filter frontend test --run src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 5: Run the schema test from E1 (now that the adapter exists).**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`
Expected: PASS.

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/lib/dashboard/widgets/tasks-today.ts frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx frontend/src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts frontend/src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx
git commit -m "feat(frontend): add tasks-today widget"
```

### Task E3: `reminders-today` widget definition + adapter

**Files:**
- Create: `frontend/src/lib/dashboard/widgets/reminders-today.ts`
- Create: `frontend/src/components/features/dashboard-widgets/reminders-today-adapter.tsx`
- Test: `frontend/src/lib/dashboard/__tests__/widgets/reminders-today.schema.test.ts`
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/reminders-today-adapter.test.tsx`

- [ ] **Step 1: Schema test.**

Create `frontend/src/lib/dashboard/__tests__/widgets/reminders-today.schema.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { remindersTodayWidget } from "@/lib/dashboard/widgets/reminders-today";

describe("reminders-today widget definition", () => {
  it("declares the registry metadata", () => {
    expect(remindersTodayWidget.type).toBe("reminders-today");
    expect(remindersTodayWidget.dataScope).toBe("household");
    expect(remindersTodayWidget.defaultSize).toEqual({ w: 3, h: 4 });
    expect(remindersTodayWidget.minSize).toEqual({ w: 3, h: 2 });
    expect(remindersTodayWidget.maxSize).toEqual({ w: 6, h: 8 });
    expect(remindersTodayWidget.defaultConfig).toEqual({ limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 5 }).success).toBe(true);
  });
});
```

- [ ] **Step 2: Adapter test.**

Create `frontend/src/components/features/dashboard-widgets/__tests__/reminders-today-adapter.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { RemindersTodayAdapter } from "@/components/features/dashboard-widgets/reminders-today-adapter";

vi.mock("@/lib/hooks/api/use-reminders", () => ({ useReminders: vi.fn() }));

import { useReminders } from "@/lib/hooks/api/use-reminders";

const reminder = (id: string, title: string, scheduledFor: string, active = true) => ({
  id, type: "reminders",
  attributes: { title, scheduledFor, active, createdAt: "", updatedAt: "" },
});

describe("RemindersTodayAdapter", () => {
  it("renders the active list capped by limit", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        reminder("1", "Take meds", "2026-05-01T08:00:00Z"),
        reminder("2", "Standup", "2026-05-01T09:00:00Z"),
        reminder("3", "Inactive", "2026-05-01T10:00:00Z", false),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText("Take meds")).toBeInTheDocument();
    expect(screen.getByText("Standup")).toBeInTheDocument();
    expect(screen.queryByText("Inactive")).not.toBeInTheDocument();
  });

  it("shows empty copy when no active reminders", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No active reminders/i)).toBeInTheDocument();
  });

  it("respects the limit", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        reminder("1", "A", "2026-05-01T08:00:00Z"),
        reminder("2", "B", "2026-05-01T09:00:00Z"),
        reminder("3", "C", "2026-05-01T10:00:00Z"),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 2 }} /></MemoryRouter>);
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });
});
```

- [ ] **Step 3: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/reminders-today.schema.test.ts src/components/features/dashboard-widgets/__tests__/reminders-today-adapter.test.tsx`
Expected: FAIL.

- [ ] **Step 4: Create the widget definition.**

Create `frontend/src/lib/dashboard/widgets/reminders-today.ts`:

```ts
import { z } from "zod";
import { RemindersTodayAdapter } from "@/components/features/dashboard-widgets/reminders-today-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const remindersTodayWidget: WidgetDefinition<Cfg> = {
  type: "reminders-today",
  displayName: "Active Reminders",
  description: "List of currently active reminders",
  component: RemindersTodayAdapter,
  configSchema: schema,
  defaultConfig: { limit: 5 },
  defaultSize: { w: 3, h: 4 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 8 },
  dataScope: "household",
};
```

- [ ] **Step 5: Create the adapter.**

Create `frontend/src/components/features/dashboard-widgets/reminders-today-adapter.tsx`:

```tsx
// Read-only widget — no mutations. See PRD §4.2.
import { Link } from "react-router-dom";
import { useReminders } from "@/lib/hooks/api/use-reminders";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Bell, ChevronRight } from "lucide-react";
import type { Reminder } from "@/types/models/reminder";

export interface RemindersTodayConfig {
  title?: string | undefined;
  limit: number;
}

function formatRelative(iso: string): string {
  const ms = new Date(iso).getTime() - Date.now();
  if (Math.abs(ms) < 60_000) return "Now";
  const mins = Math.round(ms / 60_000);
  if (Math.abs(mins) < 60) return ms > 0 ? `in ${mins} min` : `${-mins} min ago`;
  const hours = Math.round(mins / 60);
  if (Math.abs(hours) < 24) return ms > 0 ? `in ${hours}h` : `${-hours}h ago`;
  const days = Math.round(hours / 24);
  return ms > 0 ? `in ${days}d` : `${-days}d ago`;
}

export function RemindersTodayAdapter({ config }: { config: RemindersTodayConfig }) {
  const { data, isLoading, isError } = useReminders();
  const title = config.title?.trim() || "Active Reminders";

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-32" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load reminders</p></CardContent>
      </Card>
    );
  }

  const reminders = ((data?.data ?? []) as Reminder[])
    .filter((r) => r.attributes.active)
    .sort((a, b) => a.attributes.scheduledFor.localeCompare(b.attributes.scheduledFor))
    .slice(0, config.limit);

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/reminders" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/reminders"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {reminders.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Bell className="h-5 w-5" />
            <p className="text-sm">No active reminders</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {reminders.map((r) => (
              <li key={r.id} className="flex items-baseline justify-between gap-2 text-sm">
                <span className="truncate">{r.attributes.title}</span>
                <span className="text-xs text-muted-foreground shrink-0">{formatRelative(r.attributes.scheduledFor)}</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 6: Re-run the tests.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/reminders-today.schema.test.ts src/components/features/dashboard-widgets/__tests__/reminders-today-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 7: Commit.**

```bash
git add frontend/src/lib/dashboard/widgets/reminders-today.ts frontend/src/components/features/dashboard-widgets/reminders-today-adapter.tsx frontend/src/lib/dashboard/__tests__/widgets/reminders-today.schema.test.ts frontend/src/components/features/dashboard-widgets/__tests__/reminders-today-adapter.test.tsx
git commit -m "feat(frontend): add reminders-today widget"
```

### Task E4: `weather-tomorrow` widget definition + adapter

**Files:**
- Create: `frontend/src/lib/dashboard/widgets/weather-tomorrow.ts`
- Create: `frontend/src/components/features/dashboard-widgets/weather-tomorrow-adapter.tsx`
- Test: `frontend/src/lib/dashboard/__tests__/widgets/weather-tomorrow.schema.test.ts`
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/weather-tomorrow-adapter.test.tsx`

- [ ] **Step 1: Schema test.**

Create `frontend/src/lib/dashboard/__tests__/widgets/weather-tomorrow.schema.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { weatherTomorrowWidget } from "@/lib/dashboard/widgets/weather-tomorrow";

describe("weather-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(weatherTomorrowWidget.type).toBe("weather-tomorrow");
    expect(weatherTomorrowWidget.dataScope).toBe("household");
    expect(weatherTomorrowWidget.defaultSize).toEqual({ w: 3, h: 2 });
    expect(weatherTomorrowWidget.minSize).toEqual({ w: 2, h: 2 });
    expect(weatherTomorrowWidget.maxSize).toEqual({ w: 6, h: 3 });
  });

  it("schema accepts null/imperial/metric for units", () => {
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "imperial" }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "metric" }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: null }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "kelvin" }).success).toBe(false);
  });

  it("default config has units=null", () => {
    expect(weatherTomorrowWidget.defaultConfig).toEqual({ units: null });
  });
});
```

- [ ] **Step 2: Adapter test.**

Create `frontend/src/components/features/dashboard-widgets/__tests__/weather-tomorrow-adapter.test.tsx`:

```tsx
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeatherTomorrowAdapter } from "@/components/features/dashboard-widgets/weather-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-weather", () => ({ useWeatherForecast: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useWeatherForecast } from "@/lib/hooks/api/use-weather";

const daily = (date: string, hi: number, lo: number, unit = "F") => ({
  id: date, type: "weather-daily",
  attributes: { date, highTemperature: hi, lowTemperature: lo, temperatureUnit: unit, summary: "Sunny", icon: "sun", weatherCode: 0, hourlyForecast: [] },
});

describe("WeatherTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

  it("renders tomorrow's high/low", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useWeatherForecast as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        daily("2026-05-01", 70, 50),
        daily("2026-05-02", 75, 55),
      ] },
      isLoading: false, isError: false,
    });
    render(<WeatherTomorrowAdapter config={{ units: null }} />);
    expect(screen.getByText("75°F")).toBeInTheDocument();
    expect(screen.getByText("55°F")).toBeInTheDocument();
  });

  it("shows fallback when tomorrow is missing from forecast", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useWeatherForecast as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [daily("2026-05-01", 70, 50)] },
      isLoading: false, isError: false,
    });
    render(<WeatherTomorrowAdapter config={{ units: null }} />);
    expect(screen.getByText(/Tomorrow's forecast not available/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 3: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/weather-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/weather-tomorrow-adapter.test.tsx`
Expected: FAIL.

- [ ] **Step 4: Create the definition.**

Create `frontend/src/lib/dashboard/widgets/weather-tomorrow.ts`:

```ts
import { z } from "zod";
import { WeatherTomorrowAdapter } from "@/components/features/dashboard-widgets/weather-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  units: z.enum(["imperial", "metric"]).nullable().default(null),
});

type Cfg = z.infer<typeof schema>;

export const weatherTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "weather-tomorrow",
  displayName: "Tomorrow's Weather",
  description: "Tomorrow's high and low",
  component: WeatherTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { units: null },
  defaultSize: { w: 3, h: 2 },
  minSize: { w: 2, h: 2 },
  maxSize: { w: 6, h: 3 },
  dataScope: "household",
};
```

- [ ] **Step 5: Create the adapter.**

Create `frontend/src/components/features/dashboard-widgets/weather-tomorrow-adapter.tsx`:

```tsx
// Read-only widget — no mutations. See PRD §4.3.
import { useWeatherForecast } from "@/lib/hooks/api/use-weather";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Cloud } from "lucide-react";

export interface WeatherTomorrowConfig {
  units: "imperial" | "metric" | null;
}

export function WeatherTomorrowAdapter({ config: _config }: { config: WeatherTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { data, isLoading, isError } = useWeatherForecast();

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-4 w-24" data-slot="skeleton" /></CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-1" data-slot="skeleton" />
          <Skeleton className="h-3 w-16" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load forecast</p></CardContent>
      </Card>
    );
  }

  const entry = (data?.data ?? []).find((d) => d.attributes.date === tomorrow);
  if (!entry) {
    return (
      <Card className="h-full">
        <CardHeader><CardTitle className="text-sm font-medium">Tomorrow</CardTitle></CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Tomorrow's forecast not available</p>
        </CardContent>
      </Card>
    );
  }

  const unit = entry.attributes.temperatureUnit;
  return (
    <Card className="h-full">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">Tomorrow</CardTitle>
        <Cloud className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-2">
          <span className="text-2xl font-bold">{entry.attributes.highTemperature}°{unit}</span>
          <span className="text-sm text-muted-foreground">/ {entry.attributes.lowTemperature}°{unit}</span>
        </div>
        <p className="text-xs text-muted-foreground">{entry.attributes.summary}</p>
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 6: Re-run tests.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/weather-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/weather-tomorrow-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 7: Commit.**

```bash
git add frontend/src/lib/dashboard/widgets/weather-tomorrow.ts frontend/src/components/features/dashboard-widgets/weather-tomorrow-adapter.tsx frontend/src/lib/dashboard/__tests__/widgets/weather-tomorrow.schema.test.ts frontend/src/components/features/dashboard-widgets/__tests__/weather-tomorrow-adapter.test.tsx
git commit -m "feat(frontend): add weather-tomorrow widget"
```

### Task E5: `calendar-tomorrow` widget definition + adapter

**Files:**
- Create: `frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts`
- Create: `frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx`
- Test: `frontend/src/lib/dashboard/__tests__/widgets/calendar-tomorrow.schema.test.ts`
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx`

- [ ] **Step 1: Schema test.**

```ts
// frontend/src/lib/dashboard/__tests__/widgets/calendar-tomorrow.schema.test.ts
import { describe, it, expect } from "vitest";
import { calendarTomorrowWidget } from "@/lib/dashboard/widgets/calendar-tomorrow";

describe("calendar-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(calendarTomorrowWidget.type).toBe("calendar-tomorrow");
    expect(calendarTomorrowWidget.dataScope).toBe("household");
    expect(calendarTomorrowWidget.defaultSize).toEqual({ w: 4, h: 3 });
    expect(calendarTomorrowWidget.defaultConfig).toEqual({ includeAllDay: true, limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(calendarTomorrowWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(calendarTomorrowWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
  });
});
```

- [ ] **Step 2: Adapter test.**

```tsx
// frontend/src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CalendarTomorrowAdapter } from "@/components/features/dashboard-widgets/calendar-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-calendar", () => ({ useCalendarEvents: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";

const event = (id: string, title: string, startTime: string, endTime: string, allDay = false) => ({
  id, type: "calendar-events",
  attributes: { title, startTime, endTime, allDay, userColor: "#000", userDisplayName: "Me" },
});

describe("CalendarTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

  it("renders tomorrow's events sorted with all-day first", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useCalendarEvents as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        event("a", "Lunch",   "2026-05-02T17:00:00Z", "2026-05-02T18:00:00Z"),
        event("b", "Holiday", "2026-05-02T00:00:00Z", "2026-05-02T23:59:59Z", true),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><CalendarTomorrowAdapter config={{ includeAllDay: true, limit: 5 }} /></MemoryRouter>);
    const items = screen.getAllByRole("listitem");
    expect(items[0]).toHaveTextContent("Holiday");
    expect(items[1]).toHaveTextContent("Lunch");
  });

  it("shows empty state", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useCalendarEvents as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><CalendarTomorrowAdapter config={{ includeAllDay: true, limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No events tomorrow/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 3: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/calendar-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx`
Expected: FAIL.

- [ ] **Step 4: Create the definition.**

```ts
// frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts
import { z } from "zod";
import { CalendarTomorrowAdapter } from "@/components/features/dashboard-widgets/calendar-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  includeAllDay: z.boolean().default(true),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const calendarTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "calendar-tomorrow",
  displayName: "Tomorrow's Calendar",
  description: "Tomorrow's events",
  component: CalendarTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { includeAllDay: true, limit: 5 },
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 6 },
  dataScope: "household",
};
```

- [ ] **Step 5: Create the adapter.**

```tsx
// frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx
// Read-only widget — no mutations. See PRD §4.4.
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CalendarDays, ChevronRight } from "lucide-react";
import type { CalendarEvent } from "@/types/models/calendar";

export interface CalendarTomorrowConfig {
  includeAllDay: boolean;
  limit: number;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function tomorrowRange(tomorrow: string): { start: string; end: string } {
  // tomorrow is "YYYY-MM-DD" in the household timezone; render the full local
  // day as ISO timestamps for the calendar query.
  const [y, m, d] = tomorrow.split("-").map(Number);
  const start = new Date(y, m - 1, d, 0, 0, 0, 0).toISOString();
  const end = new Date(y, m - 1, d, 23, 59, 59, 999).toISOString();
  return { start, end };
}

export function CalendarTomorrowAdapter({ config }: { config: CalendarTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { start, end } = useMemo(() => tomorrowRange(tomorrow), [tomorrow]);
  const { data, isLoading, isError } = useCalendarEvents(start, end);

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-32" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load calendar</p></CardContent>
      </Card>
    );
  }

  const all = (data?.data ?? []) as CalendarEvent[];
  const filtered = all.filter((e) => config.includeAllDay || !e.attributes.allDay);
  const sorted = filtered.sort((a, b) => {
    if (a.attributes.allDay && !b.attributes.allDay) return -1;
    if (!a.attributes.allDay && b.attributes.allDay) return 1;
    return a.attributes.startTime.localeCompare(b.attributes.startTime);
  });
  const visible = sorted.slice(0, config.limit);
  const remainder = sorted.length - visible.length;

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/calendar" className="hover:underline">Tomorrow</Link>
        </CardTitle>
        <CardAction><Link to="/app/calendar"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {visible.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CalendarDays className="h-5 w-5" />
            <p className="text-sm">No events tomorrow</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {visible.map((e) => (
              <li key={e.id} className="flex items-start gap-2 text-sm">
                <span className="text-xs font-medium text-muted-foreground w-16 shrink-0 pt-0.5">
                  {e.attributes.allDay ? "All Day" : formatTime(e.attributes.startTime)}
                </span>
                <span className="flex items-center gap-1.5 min-w-0">
                  <span className="h-2 w-2 rounded-full shrink-0" style={{ backgroundColor: e.attributes.userColor }} />
                  <span className="truncate">{e.attributes.title}</span>
                </span>
              </li>
            ))}
            {remainder > 0 && (
              <li className="text-xs text-muted-foreground">+{remainder} more</li>
            )}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 6: Re-run tests.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/calendar-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 7: Commit.**

```bash
git add frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx frontend/src/lib/dashboard/__tests__/widgets/calendar-tomorrow.schema.test.ts frontend/src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx
git commit -m "feat(frontend): add calendar-tomorrow widget"
```

### Task E6: `tasks-tomorrow` widget definition + adapter

**Files:**
- Create: `frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts`
- Create: `frontend/src/components/features/dashboard-widgets/tasks-tomorrow-adapter.tsx`
- Test: `frontend/src/lib/dashboard/__tests__/widgets/tasks-tomorrow.schema.test.ts`
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/tasks-tomorrow-adapter.test.tsx`

- [ ] **Step 1: Schema test.**

```ts
// frontend/src/lib/dashboard/__tests__/widgets/tasks-tomorrow.schema.test.ts
import { describe, it, expect } from "vitest";
import { tasksTomorrowWidget } from "@/lib/dashboard/widgets/tasks-tomorrow";

describe("tasks-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(tasksTomorrowWidget.type).toBe("tasks-tomorrow");
    expect(tasksTomorrowWidget.dataScope).toBe("household");
    expect(tasksTomorrowWidget.defaultSize).toEqual({ w: 3, h: 3 });
    expect(tasksTomorrowWidget.defaultConfig).toEqual({ limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(tasksTomorrowWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(tasksTomorrowWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
  });
});
```

- [ ] **Step 2: Adapter test.**

```tsx
// frontend/src/components/features/dashboard-widgets/__tests__/tasks-tomorrow-adapter.test.tsx
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TasksTomorrowAdapter } from "@/components/features/dashboard-widgets/tasks-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-tasks", () => ({ useTasks: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useTasks } from "@/lib/hooks/api/use-tasks";

const task = (id: string, title: string, dueOn: string, status: "pending"|"completed" = "pending") => ({
  id, type: "tasks",
  attributes: { title, status, dueOn, rolloverEnabled: false, createdAt: "", updatedAt: "" },
});

describe("TasksTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

  it("renders incomplete tasks due tomorrow capped by limit", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Pay bill", "2026-05-02"),
        task("b", "Old", "2026-04-28"),
        task("c", "Done", "2026-05-02", "completed"),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><TasksTomorrowAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText("Pay bill")).toBeInTheDocument();
    expect(screen.queryByText("Old")).not.toBeInTheDocument();
    expect(screen.queryByText("Done")).not.toBeInTheDocument();
  });

  it("renders empty state", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><TasksTomorrowAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No tasks for tomorrow/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 3: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/tasks-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/tasks-tomorrow-adapter.test.tsx`
Expected: FAIL.

- [ ] **Step 4: Create the definition.**

```ts
// frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts
import { z } from "zod";
import { TasksTomorrowAdapter } from "@/components/features/dashboard-widgets/tasks-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const tasksTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "tasks-tomorrow",
  displayName: "Tomorrow's Tasks",
  description: "Incomplete tasks due tomorrow",
  component: TasksTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { limit: 5 },
  defaultSize: { w: 3, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 6 },
  dataScope: "household",
};
```

- [ ] **Step 5: Create the adapter.**

```tsx
// frontend/src/components/features/dashboard-widgets/tasks-tomorrow-adapter.tsx
// Read-only widget — no mutations. See PRD §4.5.
import { Link } from "react-router-dom";
import { useTasks } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckSquare, ChevronRight } from "lucide-react";
import type { Task } from "@/types/models/task";

export interface TasksTomorrowConfig {
  title?: string | undefined;
  limit: number;
}

export function TasksTomorrowAdapter({ config }: { config: TasksTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { data, isLoading, isError } = useTasks();
  const title = config.title?.trim() || "Tomorrow's Tasks";

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-28" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load tasks</p></CardContent>
      </Card>
    );
  }

  const all = (data?.data ?? []) as Task[];
  const due = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn === tomorrow);
  const visible = due.slice(0, config.limit);
  const remainder = due.length - visible.length;

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/tasks" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/tasks"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {visible.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CheckSquare className="h-5 w-5" />
            <p className="text-sm">No tasks for tomorrow</p>
          </div>
        ) : (
          <ul className="space-y-1">
            {visible.map((t) => (
              <li key={t.id} className="text-sm">
                <Link to="/app/tasks" className="hover:underline">{t.attributes.title}</Link>
              </li>
            ))}
            {remainder > 0 && <li className="text-xs text-muted-foreground">+{remainder} more</li>}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 6: Re-run tests.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/tasks-tomorrow.schema.test.ts src/components/features/dashboard-widgets/__tests__/tasks-tomorrow-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 7: Commit.**

```bash
git add frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts frontend/src/components/features/dashboard-widgets/tasks-tomorrow-adapter.tsx frontend/src/lib/dashboard/__tests__/widgets/tasks-tomorrow.schema.test.ts frontend/src/components/features/dashboard-widgets/__tests__/tasks-tomorrow-adapter.test.tsx
git commit -m "feat(frontend): add tasks-tomorrow widget"
```

### Task E7: Register all five widgets in the registry

**Files:**
- Modify: `frontend/src/lib/dashboard/widget-registry.ts`
- Modify: `frontend/src/lib/dashboard/__tests__/widget-registry.test.ts` (extend assertions)

- [ ] **Step 1: Extend the registry test.**

Read `frontend/src/lib/dashboard/__tests__/widget-registry.test.ts`. Append:

```ts
it("contains the 5 task-046 widgets", () => {
  for (const t of ["tasks-today", "reminders-today", "weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow"] as const) {
    expect(findWidget(t)).toBeDefined();
  }
});
```

- [ ] **Step 2: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widget-registry.test.ts`
Expected: FAIL — `findWidget("tasks-today")` returns undefined.

- [ ] **Step 3: Register all five definitions.**

Replace the imports + array in `frontend/src/lib/dashboard/widget-registry.ts` with:

```ts
import type { ComponentType } from "react";
import type { z } from "zod";
import type { WidgetType } from "@/lib/dashboard/widget-types";
import { weatherWidget } from "@/lib/dashboard/widgets/weather";
import { tasksSummaryWidget } from "@/lib/dashboard/widgets/tasks-summary";
import { remindersSummaryWidget } from "@/lib/dashboard/widgets/reminders-summary";
import { overdueSummaryWidget } from "@/lib/dashboard/widgets/overdue-summary";
import { mealPlanTodayWidget } from "@/lib/dashboard/widgets/meal-plan-today";
import { calendarTodayWidget } from "@/lib/dashboard/widgets/calendar-today";
import { packagesSummaryWidget } from "@/lib/dashboard/widgets/packages-summary";
import { habitsTodayWidget } from "@/lib/dashboard/widgets/habits-today";
import { workoutTodayWidget } from "@/lib/dashboard/widgets/workout-today";
import { tasksTodayWidget } from "@/lib/dashboard/widgets/tasks-today";
import { remindersTodayWidget } from "@/lib/dashboard/widgets/reminders-today";
import { weatherTomorrowWidget } from "@/lib/dashboard/widgets/weather-tomorrow";
import { calendarTomorrowWidget } from "@/lib/dashboard/widgets/calendar-tomorrow";
import { tasksTomorrowWidget } from "@/lib/dashboard/widgets/tasks-tomorrow";

export type WidgetDefinition<TConfig> = {
  type: WidgetType;
  displayName: string;
  description: string;
  component: ComponentType<{ config: TConfig }>;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize: { w: number; h: number };
  dataScope: "household" | "user";
};

export type AnyWidgetDefinition = WidgetDefinition<unknown>;

export const widgetRegistry: readonly AnyWidgetDefinition[] = [
  weatherWidget,
  tasksSummaryWidget,
  remindersSummaryWidget,
  overdueSummaryWidget,
  mealPlanTodayWidget,
  calendarTodayWidget,
  packagesSummaryWidget,
  habitsTodayWidget,
  workoutTodayWidget,
  tasksTodayWidget,
  remindersTodayWidget,
  weatherTomorrowWidget,
  calendarTomorrowWidget,
  tasksTomorrowWidget,
] as unknown as readonly AnyWidgetDefinition[];

export function findWidget(type: string): AnyWidgetDefinition | undefined {
  return widgetRegistry.find((w) => w.type === type);
}
```

- [ ] **Step 4: Run — expect PASS.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widget-registry.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/lib/dashboard/widget-registry.ts frontend/src/lib/dashboard/__tests__/widget-registry.test.ts
git commit -m "feat(frontend): register 5 task-046 widgets in registry"
```

---

## Phase F — `meal-plan-today` view extension

### Task F1: Extend the Zod schema with `view`

**Files:**
- Modify: `frontend/src/lib/dashboard/widgets/meal-plan-today.ts`
- Modify: `frontend/src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts` (create if missing)

- [ ] **Step 1: Write the failing schema test.**

Create `frontend/src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { mealPlanTodayWidget } from "@/lib/dashboard/widgets/meal-plan-today";

describe("meal-plan-today widget extension", () => {
  it("schema accepts view: today-detail", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 3, view: "today-detail" }).success).toBe(true);
  });

  it("schema accepts view: list", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 1, view: "list" }).success).toBe(true);
  });

  it("schema rejects unknown view", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 1, view: "grid" }).success).toBe(false);
  });

  it("default view is list when omitted", () => {
    const parsed = mealPlanTodayWidget.configSchema.parse({ horizonDays: 1 });
    expect(parsed.view).toBe("list");
  });

  it("default config still has horizonDays:1 and view:list", () => {
    expect(mealPlanTodayWidget.defaultConfig).toEqual({ horizonDays: 1, view: "list" });
  });
});
```

- [ ] **Step 2: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts`
Expected: FAIL.

- [ ] **Step 3: Update the schema and default config.**

Replace `frontend/src/lib/dashboard/widgets/meal-plan-today.ts` with:

```ts
import { z } from "zod";
import { MealPlanAdapter } from "@/components/features/dashboard-widgets/meal-plan-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  horizonDays: z.union([z.literal(1), z.literal(3), z.literal(7)]).default(1),
  view: z.enum(["list", "today-detail"]).default("list"),
});

type Cfg = z.infer<typeof schema>;

export const mealPlanTodayWidget: WidgetDefinition<Cfg> = {
  type: "meal-plan-today",
  displayName: "Meal Plan",
  description: "Upcoming planned meals",
  component: MealPlanAdapter,
  configSchema: schema,
  defaultConfig: { horizonDays: 1, view: "list" },
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 6 },
  dataScope: "household",
};
```

- [ ] **Step 4: Run — expect PASS.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit (don't push yet — adapter pending).**

```bash
git add frontend/src/lib/dashboard/widgets/meal-plan-today.ts frontend/src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts
git commit -m "feat(frontend): add view config to meal-plan-today widget"
```

### Task F2: `MealPlanTodayDetail` component

**Files:**
- Create: `frontend/src/components/features/meals/meal-plan-today-detail.tsx`
- Create: `frontend/src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx`

- [ ] **Step 1: Write the failing rendering test.**

Create `frontend/src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { MealPlanTodayDetail } from "@/components/features/meals/meal-plan-today-detail";

vi.mock("@/lib/hooks/api/use-meals", () => ({
  usePlans: vi.fn(),
  usePlan: vi.fn(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";

const mockPlan = (items: Array<{ id: string; day: string; slot: string; recipe_id: string; recipe_title: string }>) => {
  (usePlans as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
    data: { data: [{ id: "p1" }] },
    isLoading: false, isError: false,
  });
  (usePlan as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
    data: { data: { attributes: { items } } },
    isLoading: false, isError: false,
  });
};

describe("MealPlanTodayDetail", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
  });
  afterEach(() => vi.useRealTimers());

  it("renders today's full B/L/D + N follow-up days of dinners", () => {
    mockPlan([
      { id: "1", day: "2026-05-01", slot: "breakfast", recipe_id: "r1", recipe_title: "Toast" },
      { id: "2", day: "2026-05-01", slot: "lunch",     recipe_id: "r2", recipe_title: "Salad" },
      { id: "3", day: "2026-05-01", slot: "dinner",    recipe_id: "r3", recipe_title: "Tacos" },
      { id: "4", day: "2026-05-02", slot: "dinner",    recipe_id: "r4", recipe_title: "Pasta" },
      { id: "5", day: "2026-05-02", slot: "lunch",     recipe_id: "r5", recipe_title: "SkipThis" },
      { id: "6", day: "2026-05-03", slot: "dinner",    recipe_id: "r6", recipe_title: "Stir fry" },
      { id: "7", day: "2026-05-04", slot: "dinner",    recipe_id: "r7", recipe_title: "Beyond horizon" },
    ]);
    render(<MemoryRouter><MealPlanTodayDetail horizonDays={3} /></MemoryRouter>);

    // Today section: all populated slots
    expect(screen.getByText("Toast")).toBeInTheDocument();
    expect(screen.getByText("Salad")).toBeInTheDocument();
    expect(screen.getByText("Tacos")).toBeInTheDocument();

    // Next-N section: only dinners for the next 3 days
    expect(screen.getByText("Pasta")).toBeInTheDocument();
    expect(screen.getByText("Stir fry")).toBeInTheDocument();
    expect(screen.queryByText("SkipThis")).not.toBeInTheDocument();
    expect(screen.queryByText("Beyond horizon")).not.toBeInTheDocument();
  });

  it("collapses to today-only when horizonDays is 1", () => {
    mockPlan([
      { id: "1", day: "2026-05-01", slot: "dinner", recipe_id: "r1", recipe_title: "Tacos" },
      { id: "2", day: "2026-05-02", slot: "dinner", recipe_id: "r2", recipe_title: "Pasta" },
    ]);
    render(<MemoryRouter><MealPlanTodayDetail horizonDays={1} /></MemoryRouter>);
    expect(screen.getByText("Tacos")).toBeInTheDocument();
    expect(screen.queryByText("Pasta")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run — expect failure.**

Run: `pnpm --filter frontend test --run src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx`
Expected: FAIL.

- [ ] **Step 3: Implement the component.**

Create `frontend/src/components/features/meals/meal-plan-today-detail.tsx`:

```tsx
import { Link } from "react-router-dom";
import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";
import { useTenant } from "@/context/tenant-context";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { UtensilsCrossed, ChevronRight } from "lucide-react";
import { getLocalWeekStart, getLocalTodayStr, getLocalDateStrOffset } from "@/lib/date-utils";
import type { Slot } from "@/types/models/meal-plan";

const SLOT_ORDER: Slot[] = ["breakfast", "lunch", "dinner", "snack", "side"];
const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

export interface MealPlanTodayDetailProps {
  horizonDays: 1 | 3 | 7;
}

export function MealPlanTodayDetail({ horizonDays }: MealPlanTodayDetailProps) {
  const { household } = useTenant();
  const tz = household?.attributes.timezone;
  const monday = getLocalWeekStart(tz);
  const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, "0")}-${String(monday.getDate()).padStart(2, "0")}`;
  const todayStr = getLocalTodayStr(tz);

  const { data: plansData, isLoading: plansLoading, isError: plansError } = usePlans({ starts_on: weekStart });
  const planId = plansData?.data?.[0]?.id ?? null;
  const { data: planDetail, isLoading: detailLoading, isError: detailError } = usePlan(planId);

  if (plansLoading || (planId !== null && detailLoading)) {
    return (
      <Card><CardHeader><Skeleton className="h-5 w-28" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (plansError || detailError) {
    return (
      <Card className="border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load meal plan</p></CardContent>
      </Card>
    );
  }

  const items = planDetail?.data?.attributes?.items ?? [];
  const today = items
    .filter((i) => i.day === todayStr)
    .sort((a, b) => SLOT_ORDER.indexOf(a.slot) - SLOT_ORDER.indexOf(b.slot));

  // Build follow-up dinner rows for days [today+1 .. today+horizonDays-1] when
  // horizonDays > 1. (horizonDays=1 means today only.)
  const followUps: Array<{ day: string; item: typeof items[number] | null }> = [];
  for (let n = 1; n < horizonDays; n++) {
    const day = getLocalDateStrOffset(tz, n);
    const dinner = items.find((i) => i.day === day && i.slot === "dinner") ?? null;
    followUps.push({ day, item: dinner });
  }

  return (
    <Link to="/app/meals" className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader>
          <CardTitle className="text-sm font-medium">Meals</CardTitle>
          <CardAction><ChevronRight className="h-4 w-4 text-muted-foreground" /></CardAction>
        </CardHeader>
        <CardContent className="space-y-3">
          <section>
            <h4 className="text-xs font-medium text-muted-foreground">Today</h4>
            {today.length === 0 ? (
              <div className="flex items-center gap-2 text-muted-foreground mt-1">
                <UtensilsCrossed className="h-4 w-4" />
                <p className="text-sm">No meals planned</p>
              </div>
            ) : (
              <ul className="mt-1 space-y-1">
                {today.map((item) => (
                  <li key={item.id} className="flex items-baseline gap-2">
                    <span className="text-xs font-medium text-muted-foreground w-16 shrink-0">{SLOT_LABELS[item.slot]}</span>
                    <span className="text-sm truncate">{item.recipe_title}</span>
                  </li>
                ))}
              </ul>
            )}
          </section>
          {followUps.length > 0 && (
            <section>
              <h4 className="text-xs font-medium text-muted-foreground">Next {followUps.length} {followUps.length === 1 ? "day" : "days"}</h4>
              <ul className="mt-1 space-y-1">
                {followUps.map((f) => (
                  <li key={f.day} className="flex items-baseline gap-2">
                    <span className="text-xs font-medium text-muted-foreground w-16 shrink-0">{f.day.slice(5)}</span>
                    <span className="text-sm truncate">{f.item?.recipe_title ?? "—"}</span>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
```

- [ ] **Step 4: Re-run the test.**

Run: `pnpm --filter frontend test --run src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/components/features/meals/meal-plan-today-detail.tsx frontend/src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx
git commit -m "feat(frontend): add MealPlanTodayDetail component"
```

### Task F3: Branch the meal-plan adapter on `view`

**Files:**
- Modify: `frontend/src/components/features/dashboard-widgets/meal-plan-adapter.tsx`

- [ ] **Step 1: Write the failing test.**

Create `frontend/src/components/features/dashboard-widgets/__tests__/meal-plan-adapter.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { MealPlanAdapter } from "@/components/features/dashboard-widgets/meal-plan-adapter";

vi.mock("@/components/features/meals/meal-plan-widget", () => ({
  MealPlanWidget: () => <div>LIST_VIEW</div>,
}));
vi.mock("@/components/features/meals/meal-plan-today-detail", () => ({
  MealPlanTodayDetail: () => <div>DETAIL_VIEW</div>,
}));

describe("MealPlanAdapter", () => {
  it("renders MealPlanWidget for view='list'", () => {
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 1, view: "list" }} /></MemoryRouter>);
    expect(screen.getByText("LIST_VIEW")).toBeInTheDocument();
  });

  it("renders MealPlanTodayDetail for view='today-detail'", () => {
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 3, view: "today-detail" }} /></MemoryRouter>);
    expect(screen.getByText("DETAIL_VIEW")).toBeInTheDocument();
  });

  it("absent view defaults to list", () => {
    // Adapter receives whatever the registry produces; with the new schema
    // default the absence path is exercised by Zod, but we also assert the
    // adapter handles a config object lacking view (defensive).
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 1 } as any} /></MemoryRouter>);
    expect(screen.getByText("LIST_VIEW")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run — expect failure (no branching today).**

Run: `pnpm --filter frontend test --run src/components/features/dashboard-widgets/__tests__/meal-plan-adapter.test.tsx`
Expected: FAIL — both tests render LIST_VIEW.

- [ ] **Step 3: Branch the adapter.**

Replace `frontend/src/components/features/dashboard-widgets/meal-plan-adapter.tsx` with:

```tsx
import { MealPlanWidget } from "@/components/features/meals/meal-plan-widget";
import { MealPlanTodayDetail } from "@/components/features/meals/meal-plan-today-detail";

export interface MealPlanAdapterConfig {
  horizonDays: 1 | 3 | 7;
  view?: "list" | "today-detail";
}

export function MealPlanAdapter({ config }: { config: MealPlanAdapterConfig }) {
  if (config.view === "today-detail") {
    return <MealPlanTodayDetail horizonDays={config.horizonDays} />;
  }
  return <MealPlanWidget horizonDays={config.horizonDays} />;
}
```

- [ ] **Step 4: Re-run.**

Run: `pnpm --filter frontend test --run src/components/features/dashboard-widgets/__tests__/meal-plan-adapter.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/components/features/dashboard-widgets/meal-plan-adapter.tsx frontend/src/components/features/dashboard-widgets/__tests__/meal-plan-adapter.test.tsx
git commit -m "feat(frontend): branch meal-plan-adapter on view config"
```

---

## Phase G — Kiosk seed layout + parity test

### Task G1: `kiosk-seed-layout.ts` + tests

**Files:**
- Create: `frontend/src/lib/dashboard/kiosk-seed-layout.ts`
- Create: `frontend/src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts`

- [ ] **Step 1: Write failing tests.**

Create `frontend/src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { layoutSchema } from "@/lib/dashboard/schema";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";

describe("kioskSeedLayout", () => {
  it("passes layoutSchema validation", () => {
    expect(layoutSchema.safeParse(kioskSeedLayout()).success).toBe(true);
  });

  it("every widget type is in the registry", () => {
    for (const w of kioskSeedLayout().widgets) {
      expect(findWidget(w.type), `missing registry entry for ${w.type}`).toBeDefined();
    }
  });

  it("every widget satisfies x + w <= GRID_COLUMNS", () => {
    for (const w of kioskSeedLayout().widgets) {
      expect(w.x + w.w, `widget ${w.type} overflows`).toBeLessThanOrEqual(GRID_COLUMNS);
    }
  });

  it("each invocation produces fresh UUIDs", () => {
    const a = kioskSeedLayout().widgets.map((w) => w.id);
    const b = kioskSeedLayout().widgets.map((w) => w.id);
    expect(a.filter((id) => b.includes(id))).toEqual([]);
  });

  it("places the expected widgets in the four columns", () => {
    const types = kioskSeedLayout().widgets.map((w) => w.type);
    expect(types).toEqual([
      "weather", "meal-plan-today", "tasks-today",
      "calendar-today",
      "weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow",
      "reminders-today",
    ]);
  });

  it("meal-plan-today seeded with view: today-detail and horizonDays: 3", () => {
    const meal = kioskSeedLayout().widgets.find((w) => w.type === "meal-plan-today")!;
    expect(meal.config).toMatchObject({ view: "today-detail", horizonDays: 3 });
  });
});
```

- [ ] **Step 2: Run — expect failure (file doesn't exist).**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts`
Expected: FAIL.

- [ ] **Step 3: Create the seed layout.**

Create `frontend/src/lib/dashboard/kiosk-seed-layout.ts`:

```ts
import type { Layout } from "@/lib/dashboard/schema";

function uuid(): string {
  return (crypto as Crypto).randomUUID();
}

/**
 * Seeds the "Kiosk" dashboard — a 4-column kiosk-style composition.
 * Columns each total h:12 grid units; per-widget heights tuned per design §6.2.
 */
export function kioskSeedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      // Column 1 (x:0, w:3, h-total:12)
      { id: uuid(), type: "weather",          x: 0, y: 0,  w: 3, h: 3,  config: { units: "imperial", location: null } },
      { id: uuid(), type: "meal-plan-today",  x: 0, y: 3,  w: 3, h: 5,  config: { horizonDays: 3, view: "today-detail" } },
      { id: uuid(), type: "tasks-today",      x: 0, y: 8,  w: 3, h: 4,  config: { includeCompleted: true } },
      // Column 2 (x:3, w:3, h-total:12)
      { id: uuid(), type: "calendar-today",   x: 3, y: 0,  w: 3, h: 12, config: { horizonDays: 1, includeAllDay: true } },
      // Column 3 (x:6, w:3, h-total:12)
      { id: uuid(), type: "weather-tomorrow", x: 6, y: 0,  w: 3, h: 3,  config: { units: null } },
      { id: uuid(), type: "calendar-tomorrow",x: 6, y: 3,  w: 3, h: 5,  config: { includeAllDay: true, limit: 5 } },
      { id: uuid(), type: "tasks-tomorrow",   x: 6, y: 8,  w: 3, h: 4,  config: { limit: 5 } },
      // Column 4 (x:9, w:3, h-total:12)
      { id: uuid(), type: "reminders-today",  x: 9, y: 0,  w: 3, h: 12, config: { limit: 10 } },
    ],
  };
}
```

- [ ] **Step 4: Re-run.**

Run: `pnpm --filter frontend test --run src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/lib/dashboard/kiosk-seed-layout.ts frontend/src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts
git commit -m "feat(frontend): add kiosk-seed-layout"
```

### Task G2: Update `dashboard-renderer-parity.test.tsx` to consider both layouts

**Files:**
- Modify: `frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx`

- [ ] **Step 1: Replace the parity assertions.**

Open `frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx` and replace the third `it("widget-type allowlist matches seed layout exactly", …)` block (and the first `it("seedLayout yields one widget per registry type", …)`) with the broader check:

```tsx
import { describe, it, expect } from "vitest";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { widgetRegistry, findWidget } from "@/lib/dashboard/widget-registry";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";

describe("dashboard parity — seed layouts against widget registry", () => {
  it("union of seedLayout + kioskSeedLayout covers every registry type", () => {
    const seeded = new Set([
      ...seedLayout().widgets.map((w) => w.type),
      ...kioskSeedLayout().widgets.map((w) => w.type),
    ]);
    for (const def of widgetRegistry) {
      expect(seeded.has(def.type), `${def.type} appears in registry but not in any seed layout`).toBe(true);
    }
  });

  it("every seeded widget resolves via findWidget", () => {
    for (const w of [...seedLayout().widgets, ...kioskSeedLayout().widgets]) {
      expect(findWidget(w.type), `missing registry entry for ${w.type}`).toBeDefined();
    }
  });

  it("WIDGET_TYPES is fully covered by the union of seed layouts", () => {
    const seeded = new Set([
      ...seedLayout().widgets.map((w) => w.type),
      ...kioskSeedLayout().widgets.map((w) => w.type),
    ]);
    for (const t of WIDGET_TYPES) {
      expect(seeded.has(t), `allowlist contains ${t} but neither seed layout includes it`).toBe(true);
    }
  });
});
```

- [ ] **Step 2: Run — expect PASS.**

Run: `pnpm --filter frontend test --run src/pages/__tests__/dashboard-renderer-parity.test.tsx`
Expected: PASS.

- [ ] **Step 3: Commit.**

```bash
git add frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx
git commit -m "test(frontend): broaden dashboard parity test to union of seed layouts"
```

---

## Phase H — `DashboardRedirect.tsx` orchestration

### Task H1: Two-seed orchestration with kiosk preference gate

**Files:**
- Modify: `frontend/src/pages/DashboardRedirect.tsx`

- [ ] **Step 1: Inspect the existing redirect tests so the new orchestration stays compatible.**

Run: `cat frontend/src/pages/__tests__/DashboardRedirect.test.tsx`

(Read all assertions; the new orchestration must not break them.)

- [ ] **Step 2: Replace the redirect component.**

Replace `frontend/src/pages/DashboardRedirect.tsx` with:

```tsx
import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useDashboards, useSeedDashboard } from "@/lib/hooks/api/use-dashboards";
import {
  useHouseholdPreferences,
  useMarkKioskSeeded,
} from "@/lib/hooks/api/use-household-preferences";
import { useTenant } from "@/context/tenant-context";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { getErrorMessage } from "@/lib/api/errors";
import type { Dashboard } from "@/types/models/dashboard";

export function DashboardRedirect() {
  const navigate = useNavigate();
  const { tenant, household } = useTenant();
  const prefsQuery = useHouseholdPreferences();
  const dashboardsQuery = useDashboards();
  const homeSeed = useSeedDashboard();
  const kioskSeed = useSeedDashboard();
  const markKioskSeeded = useMarkKioskSeeded();
  const seededOnce = useRef(false);

  const { data: prefs, isError: prefsErr, error: prefsError } = prefsQuery;
  const { data: dashboards, refetch, isError: listErr, error: listError } = dashboardsQuery;

  useEffect(() => {
    if (!prefs || !dashboards) return;
    if (seededOnce.current) return;
    seededOnce.current = true;

    const list = dashboards.data;
    const prefRow = prefs.data[0];
    const kioskFlag = prefRow?.attributes.kioskDashboardSeeded ?? false;
    const prefId = prefRow?.id ?? null;

    const homeNeeded = !list.some((d) => d.attributes.scope === "household");
    const hasKioskRow = list.some(
      (d) => d.attributes.scope === "household" && d.attributes.name === "Kiosk",
    );
    const kioskNeeded = !kioskFlag && !hasKioskRow;

    const seedHome = homeNeeded
      ? homeSeed.mutateAsync({ name: "Home", layout: seedLayout(), key: "home" })
      : Promise.resolve(null);
    const seedKiosk = kioskNeeded
      ? kioskSeed.mutateAsync({ name: "Kiosk", layout: kioskSeedLayout(), key: "kiosk" })
      : Promise.resolve(null);

    Promise.allSettled([seedHome, seedKiosk]).then(async () => {
      const refreshed = (await refetch()).data?.data ?? list;
      const kioskNow = refreshed.some(
        (d) => d.attributes.scope === "household" && d.attributes.name === "Kiosk",
      );
      // Set the preference once we observe the row, regardless of who created it.
      if (kioskNow && !kioskFlag && prefId) {
        markKioskSeeded.mutate(prefId);
      }
      // Resolve target dashboard:
      const pref = prefRow?.attributes.defaultDashboardId ?? null;
      if (pref && refreshed.some((d) => d.id === pref)) {
        navigate(`/app/dashboards/${pref}`, { replace: true });
        return;
      }
      const householdDash = refreshed.find((d) => d.attributes.scope === "household");
      if (householdDash) {
        navigate(`/app/dashboards/${householdDash.id}`, { replace: true });
        return;
      }
      const userDash = refreshed.find((d) => d.attributes.scope === "user");
      if (userDash) {
        navigate(`/app/dashboards/${userDash.id}`, { replace: true });
        return;
      }
      // Fall through if we somehow have no dashboards even after seeding.
    });
  }, [prefs, dashboards, navigate, homeSeed, kioskSeed, markKioskSeeded, refetch]);

  if (!tenant?.id || !household?.id) {
    return <DashboardMessage title="No household selected" body="Pick a household to view its dashboard." />;
  }
  if (prefsErr || listErr || homeSeed.isError || kioskSeed.isError) {
    const msg =
      getErrorMessage(prefsError, "") ||
      getErrorMessage(listError, "") ||
      getErrorMessage(homeSeed.error, "") ||
      getErrorMessage(kioskSeed.error, "") ||
      "The dashboard service is unavailable.";
    return <DashboardMessage title="Couldn't load dashboards" body={msg} />;
  }
  return <DashboardSkeleton />;
}

function DashboardMessage({ title, body }: { title: string; body: string }) {
  return (
    <div className="p-8 max-w-xl mx-auto">
      <h1 className="text-xl font-semibold mb-2">{title}</h1>
      <p className="text-muted-foreground">{body}</p>
    </div>
  );
}

// Suppress unused-import lint hint for the unused Dashboard type if not used.
export type _Dashboard = Dashboard;
```

- [ ] **Step 3: Type-check.**

Run: `pnpm --filter frontend tsc --noEmit`
Expected: clean.

- [ ] **Step 4: Commit (tests will be updated next task).**

```bash
git add frontend/src/pages/DashboardRedirect.tsx
git commit -m "feat(frontend): seed Kiosk alongside Home with preference gate"
```

### Task H2: Update DashboardRedirect tests for the new scenarios

**Files:**
- Modify: `frontend/src/pages/__tests__/DashboardRedirect.test.tsx`

- [ ] **Step 1: Read the existing test file to understand the mock helpers.**

Run: `cat frontend/src/pages/__tests__/DashboardRedirect.test.tsx`

- [ ] **Step 2: Add four scenario tests.**

Add these tests inside the existing `describe` block (replicating whatever mock-helper style is in place — `vi.mock` of `use-dashboards`, `use-household-preferences`, `useNavigate`, etc.):

```tsx
it("brand-new household: fires both seed calls then navigates to Home", async () => {
  // Mock useDashboards: empty list initially, then post-refetch returns Home + Kiosk.
  // Mock useHouseholdPreferences: kioskDashboardSeeded=false, prefId="P1".
  // Mock useSeedDashboard.mutateAsync to resolve with a created dashboard.
  // Render DashboardRedirect, await effect, assert:
  //   - homeSeed called with {name:"Home", key:"home"}
  //   - kioskSeed called with {name:"Kiosk", key:"kiosk"}
  //   - markKioskSeeded called with prefId
  //   - navigate("/app/dashboards/<homeId>")
});

it("brownfield household: only kiosk seed fires; home seed is skipped", async () => {
  // Mock useDashboards: returns Home only initially, Home+Kiosk after refetch.
  // Mock prefs: kioskDashboardSeeded=false.
  // Assert home seed NOT called, kiosk seed called, markKioskSeeded called.
});

it("returning household with both already seeded: no seed calls fire", async () => {
  // Mock useDashboards: Home + Kiosk visible.
  // Mock prefs: kioskDashboardSeeded=true.
  // Assert no seed mutations fire.
});

it("user previously deleted Kiosk: preference flag prevents re-seed", async () => {
  // Mock useDashboards: Home only.
  // Mock prefs: kioskDashboardSeeded=true.
  // Assert kiosk seed NOT called even though Kiosk row is missing.
});
```

(Flesh out each `it` body using the existing test's mocking idioms — `(useSeedDashboard as Mock).mockReturnValue({ mutateAsync: vi.fn().mockResolvedValue(...), ... })` and so on.)

- [ ] **Step 3: Run — expect PASS.**

Run: `pnpm --filter frontend test --run src/pages/__tests__/DashboardRedirect.test.tsx`
Expected: all scenarios PASS.

- [ ] **Step 4: Commit.**

```bash
git add frontend/src/pages/__tests__/DashboardRedirect.test.tsx
git commit -m "test(frontend): cover four seeding scenarios in DashboardRedirect"
```

---

## Phase I — Final integration & verification

### Task I1: Full-suite green sweep

- [ ] **Step 1: Run all frontend tests.**

Run: `pnpm --filter frontend test --run`
Expected: PASS.

- [ ] **Step 2: Type-check.**

Run: `pnpm --filter frontend tsc --noEmit`
Expected: clean.

- [ ] **Step 3: Lint.**

Run: `pnpm --filter frontend lint`
Expected: clean. If any new file complains, fix the warnings (no `// @ts-ignore` shortcuts; usually means an unused import or missing return type).

- [ ] **Step 4: Run all Go tests.**

From repo root:

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 5: Smoke-test locally.**

Run: `scripts/local-up.sh`

Open the frontend, log in to a brand-new tenant, and confirm both "Home" and "Kiosk" dashboards appear in the sidebar. Click into the Kiosk dashboard and confirm all eight widgets render. Delete the Kiosk dashboard, refresh, and confirm it does NOT come back.

(If `scripts/local-up.sh` fails or the smoke fails, debug rather than skip.)

- [ ] **Step 6: Commit any incidental fixups discovered during the sweep.**

```bash
git status
git add <whatever the sweep produced>
git commit -m "chore: sweep cleanups for task-046"
```

- [ ] **Step 7: Summarize.**

Stop. Report:
- Total commits produced.
- Confirm every PRD §10 acceptance criterion is met (cross-reference against PRD).
- Hand off to `superpowers:requesting-code-review`.

---

## Self-Review

Before declaring the plan complete, the planning agent ran these checks:

**1. Spec coverage.** Each PRD §10 acceptance criterion maps to a task:

| Criterion | Task(s) |
|---|---|
| `tasks-today` palette + render + read-only + header link | E1, E2 |
| `reminders-today` list capped by `limit` + header link | E3 |
| `weather-tomorrow` H/L using household unit + fallback | E4 |
| `calendar-tomorrow` sorted, all-day first, +N more | E5 |
| `tasks-tomorrow` filtered by tomorrow + limit | E6 |
| `meal-plan-today` `view: "today-detail"` (and `list` byte-for-byte) | F1, F2, F3 |
| Allowlist additions in Go and TS, parity test passes | A1, A2, G2 |
| Brand-new household sees both dashboards | H1 |
| Brownfield households receive Kiosk on next load | B1 (backfill), H1 |
| Deleting Kiosk is permanent | C1, C2, C3, H1 |
| Independent loading/empty/error states per widget | E1–E6 (each adapter) |
| Unit tests for schemas + rendering | E1–E6, F1, F2 |
| Updated parity test passes | G2 |
| No regressions on Home dashboard | F1 (Zod default `view:"list"`), G2 (parity), I1 (suite) |

**2. Placeholder scan.** Searched the plan for "TBD", "TODO", "fill in", "similar to". One acceptable softer placeholder remains: Task H2 step 2 says "Flesh out each `it` body using the existing test's mocking idioms" — this is intentional because the test file's mock plumbing is ad-hoc and varies per repo; the agent must read it before duplicating. The four scenario behaviours are spelled out explicitly.

**3. Type consistency.** Cross-checked: `tasksTodayWidget.defaultConfig` (E1) shape matches what TasksTodayAdapter (E2) reads. `useSeedDashboard` accepts `{name, layout, key?}` (D1) and that exact shape is passed in DashboardRedirect (H1). `markKioskSeeded` takes `(tenant, id)` everywhere it's used.
