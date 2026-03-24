# Backend Audit — productivity-service

- **Service Path:** `services/productivity-service`
- **Guidelines Source:** CLAUDE.md + backend-dev-guidelines skill
- **Last Updated:** 2026-03-24

---

## 1. Executive Summary

The productivity-service provides task management, reminders, and dashboard summaries. Its two primary domains (`task`, `reminder`) follow core architectural patterns well — immutable models, proper entity mapping, GORM-based providers, and JSON:API transport via `server.RegisterHandler`. However, the three sub-domain packages (`taskrestoration`, `remindersnooze`, `reminderdismissal`) contain significant guideline violations: business logic in handlers, manual JSON parsing, direct entity creation bypassing the processor layer, and use of `server.RegisterHandler` (GET) for POST endpoints instead of `server.RegisterInputHandler`.

Key gaps across the service include: **no builder.go** in any domain, **eager provider execution** instead of lazy evaluation, **missing tenant callback registration** in tests, **no `TransformSlice` functions**, and **no `ToEntity()` methods** on models. The primary domains need incremental fixes; the sub-domains need structural refactoring to comply with layer separation rules.

**Overall Status: needs-work**

---

## 2. Current State Analysis

### Structure

```
services/productivity-service/
├── cmd/main.go                          # Entry point, route registration
├── internal/
│   ├── config/config.go                 # Environment-based configuration
│   ├── task/                            # Primary domain: task CRUD
│   │   ├── model.go, entity.go, provider.go, administrator.go
│   │   ├── processor.go, processor_test.go, resource.go, rest.go
│   ├── reminder/                        # Primary domain: reminder CRUD
│   │   ├── model.go, entity.go, provider.go, administrator.go
│   │   ├── processor.go, processor_test.go, resource.go, rest.go
│   ├── taskrestoration/                 # Sub-domain: task restore action
│   │   ├── entity.go, administrator.go, resource.go
│   ├── remindersnooze/                  # Sub-domain: reminder snooze action
│   │   ├── entity.go, resource.go
│   ├── reminderdismissal/               # Sub-domain: reminder dismiss action
│   │   ├── entity.go, resource.go
│   └── summary/                         # Cross-domain: dashboard summaries
│       └── resource.go
├── docs/                                # Service documentation
├── go.mod, go.sum, Dockerfile, README.md
```

### Key Responsibilities
- **task**: Full CRUD with soft-delete, restore (time-windowed), status transitions (pending/completed)
- **reminder**: Full CRUD with dismiss/snooze actions, active-state computation
- **taskrestoration/remindersnooze/reminderdismissal**: Action-event recording sub-domains
- **summary**: Read-only aggregation across task and reminder domains

---

## 3. Findings by Check ID

### ARCH-001: Immutable Domain Models

**Status: PASS**

Both `task/model.go` and `reminder/model.go` use private fields with public accessor methods. No setters exist. All state changes produce new instances via `Make(Entity)`.

**Evidence:**
- `task/model.go:9-39` — private fields, value-receiver getters
- `reminder/model.go:9-42` — private fields, value-receiver getters

---

### ARCH-002: Entity Pattern

**Status: WARN**

Entities have GORM tags, `TableName()`, `Migration()`, and `Make(Entity) (Model, error)`. However, **no `ToEntity()` method exists** on either Model, which the guidelines specify.

**Evidence:**
- `task/entity.go:30-38` — `Make` present, no `ToEntity()`
- `reminder/entity.go:27-34` — `Make` present, no `ToEntity()`

**Recommended Action:** Add `ToEntity()` methods to `task.Model` and `reminder.Model`.
**Effort:** S

---

### ARCH-003: Builder Pattern

**Status: FAIL**

No `builder.go` exists in any domain package. Guidelines require fluent builders that enforce invariants during construction. Currently models are only created via `Make(Entity)`.

**Evidence:**
- No `builder.go` in `task/`, `reminder/`, or any sub-domain package

**Recommended Action:** Create `builder.go` with fluent API and `Build()` validation for `task` and `reminder` domains.
**Effort:** M

---

### ARCH-004: Processor Pattern

**Status: WARN**

Processors exist for `task` and `reminder` as struct-based types with methods, which is acceptable. They correctly use `p.db.WithContext(p.ctx)`. However:
- Processors call administrator functions directly (correct) but also call provider functions that execute eagerly (see ARCH-005).
- Sub-domain packages (`taskrestoration`, `remindersnooze`, `reminderdismissal`) have **no processor at all** — business logic lives directly in handlers.

**Evidence:**
- `task/processor.go:28-88` — struct-based processor, correct `WithContext` usage
- `reminder/processor.go:22-79` — same pattern
- `taskrestoration/resource.go:35-71` — business logic in handler
- `remindersnooze/resource.go:38-74` — business logic in handler
- `reminderdismissal/resource.go:36-69` — business logic in handler

**Recommended Action:** Create processors for sub-domain packages to encapsulate business logic.
**Effort:** M

---

### ARCH-005: Provider Pattern (Lazy Evaluation)

**Status: WARN**

Providers execute queries eagerly: they call `db.Where(...).First(...)` immediately and wrap results in `model.FixedProvider` or `model.ErrorProvider`. Guidelines specify providers should use `database.Query`/`database.SliceQuery` for deferred (lazy) evaluation.

Additionally, `countByStatus`, `countOverdue`, `countCompletedToday`, `countDueNow`, `countUpcoming`, `countSnoozed` are plain functions returning `(int64, error)` — not providers at all.

**Evidence:**
- `task/provider.go:9-18` — `getByID` executes query eagerly, returns `FixedProvider`
- `task/provider.go:20-33` — `getAll` same pattern
- `reminder/provider.go:9-18` — `getByID` same pattern
- `task/provider.go:46-66` — count functions are not providers

**Recommended Action:** Migrate to `database.Query`/`database.SliceQuery` for lazy evaluation.
**Effort:** M

---

### ARCH-006: Resource / Route Registration

**Status: PASS** (primary domains) / **FAIL** (sub-domains)

Primary domains (`task`, `reminder`) correctly use `server.RegisterHandler` and `server.RegisterInputHandler`. Sub-domains misuse the pattern:

- `taskrestoration/resource.go` uses `server.RegisterHandler` (GetHandler) for a POST endpoint, then manually reads/parses request body
- `remindersnooze/resource.go` — same issue
- `reminderdismissal/resource.go` — same issue

**Evidence:**
- `task/resource.go:16-27` — correct route registration
- `taskrestoration/resource.go:31` — `RegisterHandler` used for POST
- `remindersnooze/resource.go:34` — same
- `reminderdismissal/resource.go:32` — same

**Recommended Action:** Refactor sub-domains to use `server.RegisterInputHandler` with proper request DTOs.
**Effort:** M

---

### ARCH-007: REST / JSON:API DTOs

**Status: WARN**

REST models correctly implement `GetName()`, `GetID()`, `SetID()`. ID fields marked `json:"-"`. However:
- **No `TransformSlice` function** in `task/rest.go` or `reminder/rest.go` (list handlers inline the loop)
- `task/rest.go` `CreateRequest` has a `Status` field that is ignored by `createHandler` (always sets "pending")

**Evidence:**
- `task/rest.go:22-24` — correct JSON:API interface
- `task/rest.go:40-51` — `CreateRequest` includes unused `Status` field
- `task/resource.go:39-43` — inline transform loop instead of `TransformSlice`
- `reminder/resource.go:39-43` — same

**Recommended Action:** Add `TransformSlice` functions. Remove unused `Status` from `CreateRequest`.
**Effort:** S

---

### ARCH-008: Multi-Tenancy Context

**Status: WARN**

Handlers correctly use `tenantctx.MustFromContext(r.Context())`. Create functions correctly receive `tenantID`. Update/delete do not receive `tenantID` (correct). Providers don't take `tenantID` (correct).

However, **tests do not call `database.RegisterTenantCallbacks(l, db)`** when using SQLite in-memory DBs. This is explicitly required by guidelines.

**Evidence:**
- `task/processor_test.go:14-22` — `setupTestDB` creates SQLite DB without registering tenant callbacks
- `reminder/processor_test.go:14-22` — same

**Recommended Action:** Add `database.RegisterTenantCallbacks(l, db)` to all test `setupTestDB` functions.
**Effort:** S

---

### ARCH-009: Layer Separation

**Status: FAIL** (sub-domains)

Sub-domain handlers violate layer separation by:
1. Reading raw request body and parsing JSON manually (`io.ReadAll` + `json.Unmarshal`)
2. Creating entities directly in handlers (bypassing processor/administrator layers)
3. Containing business logic (restore window validation is in processor, but entity creation is in handler)

The `summary` package accesses other domain processors directly from handlers, which is acceptable per the anti-patterns exception for read-only cross-domain handlers.

**Evidence:**
- `taskrestoration/resource.go:40-41` — `io.ReadAll(r.Body)` in handler
- `taskrestoration/resource.go:61` — direct `create()` call from handler
- `remindersnooze/resource.go:43-44` — manual body reading
- `remindersnooze/resource.go:68` — direct `db.WithContext(r.Context()).Create(&e)` in handler
- `reminderdismissal/resource.go:41-42` — manual body reading
- `reminderdismissal/resource.go:63` — direct `db.Create` in handler

**Recommended Action:** Refactor sub-domains to use proper request DTOs with `RegisterInputHandler`, and move entity creation into processors or administrators.
**Effort:** L

---

### ARCH-010: Manual JSON:API Envelope Handling

**Status: FAIL**

Sub-domains manually parse nested `Data.Relationships` and `Data.Attributes` structures — an explicit anti-pattern in the guidelines ("manual JSON:API envelope handling" and "nested Data/Type/Attributes in requests").

**Evidence:**
- `taskrestoration/resource.go:73-92` — `extractTaskRelationship` parses nested JSON:API envelope
- `remindersnooze/resource.go:76-102` — `extractSnoozeRequest` parses nested envelope
- `reminderdismissal/resource.go:71-90` — `extractReminderRelationship` parses nested envelope

**Recommended Action:** Use flat request DTOs with `server.RegisterInputHandler` for automatic deserialization.
**Effort:** M

---

### ARCH-011: Error Handling & Logging

**Status: WARN**

Handlers return generic error messages. None use `d.Logger().WithError(err)` for structured error logging as recommended. Transform errors are silently ignored (`_, _ := Transform(m)`).

**Evidence:**
- `task/resource.go:35` — `server.WriteError(w, http.StatusInternalServerError, "Error", "")` — no logging
- `task/resource.go:65` — `rest, _ := Transform(m)` — error silently ignored
- `reminder/resource.go:34-36` — same patterns

**Recommended Action:** Add `d.Logger().WithError(err)` logging before error responses. Handle transform errors.
**Effort:** S

---

### ARCH-012: Testing Conventions

**Status: WARN**

Tests exist for `task` and `reminder` processors covering core flows (create, update, complete, reopen, soft-delete, restore, snooze, dismiss). However:
- Not table-driven (guidelines recommend table-driven tests)
- Missing tenant callback registration (see ARCH-008)
- No tests for sub-domain packages
- No tests for REST layer (status code mapping, JSON:API compliance)

**Evidence:**
- `task/processor_test.go` — 6 test functions, individual test style
- `reminder/processor_test.go` — 5 test functions, individual test style
- No `*_test.go` in `taskrestoration/`, `remindersnooze/`, `reminderdismissal/`, `summary/`

**Recommended Action:** Convert to table-driven tests, add tenant callbacks, add sub-domain and REST tests.
**Effort:** M

---

### ARCH-013: Logger Usage in Handlers

**Status: WARN**

Handlers create processors with `logrus.StandardLogger()` instead of using the `d.Logger()` from `HandlerDependency`, which includes trace context and structured fields.

**Evidence:**
- `task/resource.go:32` — `NewProcessor(logrus.StandardLogger(), r.Context(), db)`
- `task/resource.go:59` — same
- `reminder/resource.go:32` — same
- All handler functions use `logrus.StandardLogger()`

**Recommended Action:** Use `d.Logger()` (cast to `*logrus.Logger` if needed) from `HandlerDependency` instead of `logrus.StandardLogger()`.
**Effort:** S

---

## 4. Structural Gaps

| Expected File | task | reminder | taskrestoration | remindersnooze | reminderdismissal | summary |
|---|---|---|---|---|---|---|
| model.go | Y | Y | N | N | N | N/A |
| entity.go | Y | Y | Y | Y | Y | N/A |
| builder.go | **N** | **N** | N/A | N/A | N/A | N/A |
| processor.go | Y | Y | **N** | **N** | **N** | N/A |
| provider.go | Y | Y | N/A | N/A | N/A | N/A |
| administrator.go | Y | Y | Y | N | N | N/A |
| resource.go | Y | Y | Y | Y | Y | Y |
| rest.go | Y | Y | N* | N* | N* | N* |

*Sub-domains define RestModel inline in resource.go rather than in a separate rest.go file.

---

## 5. Blocking Issues

1. **ARCH-009 / ARCH-010: Sub-domain handlers contain business logic, manual JSON parsing, and direct DB writes.** This is a critical layer separation violation. Three packages (`taskrestoration`, `remindersnooze`, `reminderdismissal`) bypass the processor/administrator pattern entirely from the handler layer.

2. **ARCH-006: Sub-domains use `RegisterHandler` (GET) for POST endpoints.** This bypasses automatic input deserialization and forces manual `io.ReadAll`/`json.Unmarshal`.

---

## 6. Non-Blocking Issues

1. **ARCH-003**: Missing `builder.go` for `task` and `reminder` domains
2. **ARCH-002**: Missing `ToEntity()` on domain models
3. **ARCH-005**: Providers execute eagerly instead of lazily
4. **ARCH-007**: Missing `TransformSlice` functions; unused `Status` field in `CreateRequest`
5. **ARCH-008**: Tests missing `database.RegisterTenantCallbacks`
6. **ARCH-011**: No structured error logging in handlers; transform errors silently ignored
7. **ARCH-012**: Tests not table-driven; no sub-domain or REST layer tests
8. **ARCH-013**: Handlers use `logrus.StandardLogger()` instead of `d.Logger()`

---

## 7. Inputs for /dev-docs

| Objective | Priority |
|---|---|
| Refactor sub-domain packages (taskrestoration, remindersnooze, reminderdismissal) to use proper processor layer and `RegisterInputHandler` | P0 |
| Add builder.go with fluent API and validation for task and reminder domains | P1 |
| Migrate providers to lazy evaluation using `database.Query`/`database.SliceQuery` | P1 |
| Add `ToEntity()` methods to domain models | P2 |
| Add `TransformSlice` functions and clean up unused request fields | P2 |
| Fix test infrastructure (tenant callbacks, table-driven style, sub-domain coverage) | P1 |
| Replace `logrus.StandardLogger()` with `d.Logger()` in handlers | P2 |
| Add structured error logging with `d.Logger().WithError(err)` | P2 |

---

## 8. Notes / Ambiguities

- **summary package**: Accesses `task.NewProcessor` and `reminder.NewProcessor` directly from handlers. This is a cross-domain read-only handler, which the anti-patterns doc explicitly allows as an exception. Marked acceptable.
- **Sub-domain packages** (taskrestoration, remindersnooze, reminderdismissal) are action-event recorders. Their entity-only nature (no domain model needed) may warrant a lighter pattern than full domain packages. However, they currently violate core handler-thinness and layer-separation rules regardless.
- **Provider count functions** (`countByStatus`, `countOverdue`, etc.) don't follow the provider pattern at all — they are plain functions returning `(int64, error)`. It's unclear if the guidelines expect count/aggregate queries to also use the provider pattern.
- The `task.CreateRequest` includes a `Status` field that is never used (processor always sets "pending"). This could confuse API consumers.
