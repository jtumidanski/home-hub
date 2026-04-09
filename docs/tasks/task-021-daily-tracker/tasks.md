# Daily Tracker — Task Checklist

Last Updated: 2026-04-08

---

## Phase 1: Service Scaffold & Tracking Items [L]

- [x] 1.1 Create `services/tracker-service/` directory structure (cmd/, internal/, docs/)
- [x] 1.2 Create `go.mod` with shared module dependencies and `replace` directives
- [x] 1.3 Add tracker-service to `go.work`
- [x] 1.4 Create `internal/config/config.go` with DB config (schema: "tracker"), port, JWKS URL
- [x] 1.5 Create `cmd/main.go` — logger, tracing, DB connect, auth validator, server setup
- [x] 1.6 Create `trackingitem/model.go` — immutable model with all fields (id, tenantId, userId, name, scaleType, scaleConfig, color, sortOrder, createdAt, updatedAt, deletedAt)
- [x] 1.7 Create `trackingitem/builder.go` — validation: name required (max 100), valid scale type, valid color from palette, range config validation, schedule validation (days 0-6)
- [x] 1.8 Create `trackingitem/entity.go` — GORM entity with table `tracker.tracking_items`, Migration function, Entity<->Model conversion
- [x] 1.9 Create `schedule/model.go` — immutable model (id, trackingItemId, schedule, effectiveDate, createdAt)
- [x] 1.10 Create `schedule/builder.go` — validation: schedule days 0-6, effective date required
- [x] 1.11 Create `schedule/entity.go` — GORM entity with table `tracker.schedule_snapshots`, Migration function, unique constraint on (tracking_item_id, effective_date)
- [x] 1.12 Create `schedule/provider.go` — getByTrackingItemId, getEffectiveSchedule(trackingItemId, date)
- [x] 1.13 Create `schedule/administrator.go` — create snapshot
- [x] 1.14 Create `trackingitem/provider.go` — getById, listByUser, getByName (for uniqueness check)
- [x] 1.15 Create `trackingitem/administrator.go` — create, update, softDelete
- [x] 1.16 Create `trackingitem/processor.go` — Create (with initial schedule snapshot), List, Get, Update (with schedule snapshot on schedule change), Delete (soft)
- [x] 1.17 Create `trackingitem/rest.go` — RestModel, RestDetailModel (with schedule_history), CreateRequest, UpdateRequest, transform functions
- [x] 1.18 Create `trackingitem/resource.go` — InitializeRoutes, POST/GET/GET{id}/PATCH/DELETE handlers
- [x] 1.19 Wire trackingitem routes in `cmd/main.go`
- [x] 1.20 Unit tests for builder validation, processor logic
- [x] 1.21 Verify service compiles and tracking item CRUD works end-to-end

**Acceptance:** POST creates item + initial schedule snapshot. GET returns items with current schedule. PATCH updates fields; schedule changes create new snapshot. DELETE soft-deletes. 409 on duplicate name. Scale type immutable after creation.

---

## Phase 2: Entry Logging [M]

- [x] 2.1 Create `entry/model.go` — immutable model (id, tenantId, userId, trackingItemId, date, value, skipped, note, createdAt, updatedAt)
- [x] 2.2 Create `entry/builder.go` — validation: date not in future, value matches scale type, note max 500 chars, range value within bounds
- [x] 2.3 Create `entry/entity.go` — GORM entity with table `tracker.tracking_entries`, Migration function, unique constraint on (tracking_item_id, date)
- [x] 2.4 Create `entry/provider.go` — getByItemAndDate, listByUserAndMonth, listByItemAndDateRange
- [x] 2.5 Create `entry/administrator.go` — create, update, delete
- [x] 2.6 Create `entry/processor.go` — CreateOrUpdate (upsert), Delete, Skip, RemoveSkip, ListByMonth
- [x] 2.7 Create `entry/rest.go` — RestModel (with computed `scheduled` field), EntryRequest, transform functions
- [x] 2.8 Create `entry/resource.go` — InitializeRoutes, PUT entry, DELETE entry, PUT skip, DELETE skip, GET entries by month
- [x] 2.9 Wire entry routes in `cmd/main.go`
- [x] 2.10 Unit tests for value validation per scale type, skip logic, date validation

**Acceptance:** PUT creates/updates entries with correct value validation. Skip marks scheduled days as skipped. DELETE removes entries. GET by month returns all user entries. Future dates rejected. Range values validated against bounds.

---

## Phase 3: Month Computation & Reports [L]

- [x] 3.1 Create `month/model.go` — MonthSummary model (month, complete, completion stats), ReportModel (summary stats, per-item stats), per-scale-type stat models
- [x] 3.2 Create `month/processor.go` — ComputeMonthSummary: determine active items for month, compute expected days per item using schedule snapshots, calculate completion. ComputeReport: aggregate statistics per item by scale type
- [x] 3.3 Implement schedule-aware expected day calculation: for each item, for each day of month within item's active range, check if the day matches the applicable schedule snapshot
- [x] 3.4 Implement completion logic: complete = all expected entries filled or skipped AND no future scheduled days remain in the month
- [x] 3.5 Implement sentiment report stats: positive/neutral/negative counts, positive ratio, daily values
- [x] 3.6 Implement numeric report stats: total, daily average, days > 0, min/max day, daily values
- [x] 3.7 Implement range report stats: average, min/max, std deviation, daily values
- [x] 3.8 Create `month/rest.go` — MonthSummaryRest, ReportRest, transform functions
- [x] 3.9 Create `month/resource.go` — InitializeRoutes, GET month summary, GET month report
- [x] 3.10 Wire month routes in `cmd/main.go`
- [x] 3.11 Unit tests: completion calculation with mid-month create/delete, schedule changes, all-filled, partial-fill, all-skipped, mixed scenarios *(see month/processor_test.go)*
- [x] 3.12 Unit tests: report statistics accuracy with known inputs *(see month/processor_test.go)*

**Acceptance:** Month summary returns correct completion stats. Items created/deleted mid-month only count active range. Schedule snapshots correctly determine expected days. Report returns accurate per-item stats. Report returns 400 for incomplete months.

---

## Phase 4: Today Quick-Entry [S]

- [x] 4.1 Add today logic to trackingitem or entry processor: get items scheduled for today based on current schedule snapshot, pair with today's entries
- [x] 4.2 Create today REST model and transform
- [x] 4.3 Create today resource handler (GET /api/v1/trackers/today)
- [x] 4.4 Wire today route in `cmd/main.go`
- [x] 4.5 Unit test: correct items returned for day of week, entries paired correctly *(see today/processor_test.go)*

**Acceptance:** Returns only items scheduled for today. Includes existing entries for today. Empty array for items with no entry yet.

---

## Phase 5: Infrastructure & Deployment [M]

- [x] 5.1 Create `services/tracker-service/Dockerfile` following existing pattern
- [x] 5.2 Add tracker-service to `deploy/compose/docker-compose.yml`
- [x] 5.3 Add `/api/v1/trackers` route to `deploy/compose/nginx.conf`
- [x] 5.4 Create `deploy/k8s/tracker-service.yaml` (Deployment + Service)
- [x] 5.5 Add `/api/v1/trackers` ingress rule to `deploy/k8s/ingress.yaml`
- [x] 5.6 Add CI/CD workflow for tracker-service (build, test, lint, publish)
- [x] 5.7 Verify Docker build succeeds
- [x] 5.8 Verify local compose stack starts with tracker-service
- [x] 5.9 Verify nginx routing works end-to-end

**Acceptance:** Service builds as Docker image. Compose stack includes tracker-service. Nginx routes `/api/v1/trackers` to service. K8s manifests ready for deployment.

---

## Phase 6: Frontend — Tracking Item Management [M]

- [x] 6.1 Add "Tracker" sidebar entry under personal section
- [x] 6.2 Create tracker page route and layout
- [x] 6.3 Create API client hooks for tracking item CRUD (useTrackers, useCreateTracker, useUpdateTracker, useDeleteTracker)
- [x] 6.4 Build tracking item list component with color dots, scale type, schedule display
- [x] 6.5 Build tracking item creation form: name, color palette picker, scale type dropdown, conditional scale config (range min/max), day-of-week toggle buttons, sort order
- [x] 6.6 Build tracking item edit form (same as create, scale_type disabled)
- [x] 6.7 Build delete confirmation dialog
- [x] 6.8 Zod validation schemas for create/update forms

**Acceptance:** User can create items with all fields. Edit updates name, color, schedule, sort order, scale config. Delete soft-deletes with confirmation. Scale type locked after creation. Color palette rendered visually.

---

## Phase 7: Frontend — Monthly Calendar Grid [XL]

- [x] 7.1 Create API client hooks for month summary (useMonthSummary) and entries (useEntries)
- [x] 7.2 Build month navigation component (prev/next arrows, current month display, no future months)
- [x] 7.3 Build calendar grid layout: rows = items (sorted), columns = days of month
- [x] 7.4 Render cells by scale type: sentiment icons, numeric values, range values
- [x] 7.5 Style cells: dimmed for unscheduled days, highlighted/outlined for scheduled-but-unfilled, skip indicator
- [x] 7.6 Build completion progress bar in header (e.g., "47/52 entries filled")
- [x] 7.7 Build cell entry editor (popover/modal): sentiment selector, numeric +/- input, range slider
- [x] 7.8 Add skip/clear buttons to entry editor (skip only on scheduled days)
- [x] 7.9 Add optional note field to entry editor
- [x] 7.10 Wire PUT/DELETE entry mutations with optimistic updates
- [x] 7.11 Auto-switch to dashboard view when month is complete
- [x] 7.12 Handle empty state (no tracking items)

**Acceptance:** Calendar grid renders all items x days. Cells show correct values/states. Entry editor allows logging, updating, skipping, clearing. Month navigation works. Progress shown. Completed months auto-switch to dashboard (with toggle back).

---

## Phase 8: Frontend — Today Quick-Entry [M]

- [x] 8.1 Create API client hook for today endpoint (useTrackerToday)
- [x] 8.2 Build today quick-entry page with vertical item list
- [x] 8.3 Build inline editors per scale type: sentiment buttons, numeric +/-, range slider
- [x] 8.4 Build collapsible note field per item
- [x] 8.5 Show already-logged values as editable, unfilled items as prominent
- [x] 8.6 Add progress counter ("3/5 logged today")
- [x] 8.7 Add "Log Today" button on tracker landing page linking to quick-entry
- [x] 8.8 Add "Calendar" link from today view to monthly grid

**Acceptance:** Shows only today's scheduled items. Inline editing works for all scale types. Notes editable. Progress counter accurate. Navigation between views works.

---

## Phase 9: Frontend — Monthly Dashboard Report [L]

- [x] 9.1 Create API client hook for month report (useMonthReport)
- [x] 9.2 Build report layout with general summary cards (completion rate, skip rate, total items)
- [x] 9.3 Build sentiment item stats: positive/neutral/negative bar, ratio display
- [x] 9.4 Build numeric item stats: total, average, dry days percentage, min/max
- [x] 9.5 Build range item stats: average, min/max, std dev
- [x] 9.6 Build mini visualizations: sentiment color bars, numeric sparkline/bar chart, range trend line
- [x] 9.7 Add "[Calendar]" toggle button to switch back to grid view
- [x] 9.8 Handle 400 error gracefully (incomplete month — should not normally occur due to auto-switch logic)

**Acceptance:** Dashboard displays for completed months. All stat widgets render correctly per scale type. Toggle to calendar view works. Visualizations render from daily_values data.

---

## Phase 10: Testing, Documentation & Polish [M]

- [x] 10.1 Integration tests for tracking item CRUD (including uniqueness, soft delete)
- [x] 10.2 Integration tests for entry CRUD (including value validation, skip, date constraints)
- [x] 10.3 Integration tests for month summary completion calculation
- [x] 10.4 Integration tests for report generation with known data
- [x] 10.5 Write `services/tracker-service/docs/domain.md`
- [x] 10.6 Write `services/tracker-service/docs/rest.md`
- [x] 10.7 Write `services/tracker-service/docs/storage.md`
- [x] 10.8 Write `services/tracker-service/README.md`
- [x] 10.9 Update `docs/architecture.md` with tracker-service entry
- [x] 10.10 Verify all 20 PRD acceptance criteria pass *(walked PRD §10; see commit message)*
- [ ] 10.11 Verify month summary endpoint < 200ms with 20 tracking items *(not measured against live data; pending pre-launch perf pass)*

**Acceptance:** All tests pass. Service documentation complete per DOCS.md contract. Architecture docs updated. All PRD acceptance criteria verified.
