# Backend Audit — account-service

- **Service Path:** `services/account-service`
- **Guidelines Source:** CLAUDE.md + backend-dev-guidelines skill
- **Last Updated:** 2026-03-24

---

## 1. Executive Summary

The account-service manages tenants, households, memberships, preferences, and an aggregated application context. It follows many of the prescribed patterns (immutable models, entity/Make separation, provider/administrator split, processor orchestration, JSON:API REST models). However, it has several compliance gaps:

- **No `builder.go` in any domain** — models are constructed without fluent builders or invariant validation.
- **Manual tenant filtering in providers** — multiple providers accept `tenantID` as a parameter and filter manually, bypassing GORM's automatic tenant callbacks.
- **Manual JSON:API body parsing in handlers** — membership and preference handlers use `io.ReadAll`/`json.Unmarshal` instead of `server.RegisterInputHandler[T]`.
- **Cross-domain business logic in handlers** — household's create handler directly calls `membership.NewProcessor`, which should be in the processor layer.
- **`logrus.StandardLogger()` used instead of `d.Logger()`** — handlers pass a global logger to processors instead of the context-aware dependency logger.
- **Test DBs missing `RegisterTenantCallbacks`** — all four test files create SQLite databases without registering tenant callbacks.
- **Dead code** — `tenant/provider.go` contains an unused `getByUserID` function.

Overall status: **needs-work**. The core architecture is sound but several guideline violations need addressing before new feature work.

---

## 2. Current State Analysis

### Directory Structure
```
services/account-service/
├── cmd/main.go
├── Dockerfile
├── go.mod / go.sum
├── internal/
│   ├── appcontext/       (aggregation domain — no persistence)
│   │   ├── context.go    (Resolve orchestration)
│   │   ├── resource.go   (GET /contexts/current)
│   │   └── rest.go       (JSON:API with relationships)
│   ├── config/config.go
│   ├── household/        (CRUD domain)
│   │   ├── model.go, entity.go, processor.go, provider.go
│   │   ├── administrator.go, resource.go, rest.go
│   │   └── processor_test.go
│   ├── membership/       (CRUD domain)
│   │   ├── model.go, entity.go, processor.go, provider.go
│   │   ├── administrator.go, resource.go, rest.go
│   │   └── processor_test.go
│   ├── preference/       (CRUD domain)
│   │   ├── model.go, entity.go, processor.go, provider.go
│   │   ├── administrator.go, resource.go, rest.go
│   │   └── processor_test.go
│   └── tenant/           (CRUD domain)
│       ├── model.go, entity.go, processor.go, provider.go
│       ├── administrator.go, resource.go, rest.go
│       └── processor_test.go
```

### Key Responsibilities
- **tenant**: Multi-tenant identity (no tenant_id column — it IS the tenant)
- **household**: Homes/locations within a tenant
- **membership**: User-to-household role bindings
- **preference**: Per-user settings (theme, active household)
- **appcontext**: Aggregated read-only view of user's full context

---

## 3. Findings by Check ID

### ARCH-001: Layer Separation (Handler → Processor → Provider/Administrator)

**Status: FAIL**

| Evidence | Issue |
|----------|-------|
| `household/resource.go:60-61` | `createHandler` calls `membership.NewProcessor` directly — cross-domain business logic in handler |
| `membership/resource.go:49` | `createHandler` returns `server.GetHandler` instead of `server.InputHandler[CreateRequest]` |
| `membership/resource.go:54-56` | Manual `io.ReadAll` + `jsonapi.Unmarshal` instead of `server.RegisterInputHandler` |
| `membership/resource.go:79` | `updateHandler` returns `server.GetHandler`, manually reads body |
| `preference/resource.go:42` | `updateHandler` returns `server.GetHandler`, manually reads body |

The household create handler embeds membership creation logic that belongs in `household/processor.go`. The membership and preference handlers bypass `server.RegisterInputHandler[T]` for write operations, manually parsing JSON:API envelopes — a documented anti-pattern.

### ARCH-002: Immutable Models

**Status: PASS**

All five domains define models with private fields and public accessor methods. No public mutable fields exist.

| Evidence | Note |
|----------|------|
| `household/model.go:9-17` | Private fields with value-receiver accessors |
| `membership/model.go:9-17` | Private fields with value-receiver accessors |
| `preference/model.go:9-17` | Private fields with value-receiver accessors |
| `tenant/model.go:9-14` | Private fields with value-receiver accessors |

### ARCH-003: Builder Pattern

**Status: FAIL**

No `builder.go` file exists in any domain. Models are constructed directly in `Make()` and administrator functions without fluent builders or invariant validation.

| Evidence | Note |
|----------|------|
| All domains | Missing `builder.go` — no validation on model construction |
| `household/entity.go:24-30` | `Make()` constructs model directly without validation |
| `household/administrator.go:10-25` | `create()` builds entity inline without builder |

### ARCH-004: Provider Pattern (Read Operations)

**Status: WARN**

Providers follow the curried function pattern correctly but several manually filter by `tenant_id`, which should be automatic via GORM callbacks.

| Evidence | Issue |
|----------|-------|
| `household/provider.go:20-21` | `getByTenantID(tenantID uuid.UUID)` — accepts tenantID parameter; should rely on automatic filtering |
| `membership/provider.go:20-23` | `getByUserAndTenant(userID, tenantID)` — manual `tenant_id` WHERE |
| `preference/provider.go:9-12` | `getByTenantAndUser(tenantID, userID)` — manual `tenant_id` WHERE |
| `tenant/provider.go:9-18` | `getByID` — correct (tenant entity has no tenant_id) |

### ARCH-005: Administrator Pattern (Write Operations)

**Status: PASS**

All write operations are properly separated into `administrator.go` files. Create functions accept `tenantID` where needed, update/delete do not (except where tenant_id isn't applicable, like tenant domain itself).

| Evidence | Note |
|----------|------|
| `household/administrator.go` | `create` takes tenantID, `update` does not — correct |
| `membership/administrator.go` | `create` takes tenantID, `updateRole`/`deleteByID` do not — correct |
| `preference/administrator.go` | `create` takes tenantID, others do not — correct |
| `tenant/administrator.go` | No tenantID (is the tenant itself) — correct |

### ARCH-006: REST/JSON:API Compliance

**Status: WARN**

REST models follow the JSON:API interface pattern correctly. However, request model handling has issues.

| Evidence | Issue |
|----------|-------|
| `membership/resource.go:126-155` | `extractRelationships` manually parses JSON:API envelope — anti-pattern |
| `preference/resource.go:86-104` | `extractActiveHouseholdRelationship` manually parses JSON:API envelope |
| `preference/resource.go:101` | Uses `json.Unmarshal(nil, nil)` to generate an error — fragile |
| All rest.go files | Missing `TransformSlice` function — handlers use inline loops instead |

### ARCH-007: Multi-Tenancy Context

**Status: WARN**

Tenant context is extracted correctly via `tenantctx.MustFromContext(r.Context())`. However, several providers manually filter by tenant_id instead of relying on automatic GORM callback filtering (see ARCH-004). Additionally, all handlers pass `logrus.StandardLogger()` instead of `d.Logger()` to processors, losing context-aware logging.

| Evidence | Issue |
|----------|-------|
| `household/resource.go:32,52,77` | `logrus.StandardLogger()` instead of `d.Logger()` |
| `membership/resource.go:33,67,95,115` | `logrus.StandardLogger()` instead of `d.Logger()` |
| `preference/resource.go:30,59` | `logrus.StandardLogger()` instead of `d.Logger()` |
| `tenant/resource.go:30,45,65` | `logrus.StandardLogger()` instead of `d.Logger()` |
| `appcontext/resource.go:25` | `logrus.StandardLogger()` instead of `d.Logger()` |

### ARCH-008: Testing

**Status: WARN**

Each domain has processor tests with reasonable coverage. However:

| Evidence | Issue |
|----------|-------|
| `household/processor_test.go:13-21` | `setupTestDB` missing `database.RegisterTenantCallbacks(l, db)` |
| `membership/processor_test.go:13-21` | Same — missing tenant callbacks |
| `preference/processor_test.go:13-21` | Same — missing tenant callbacks |
| `tenant/processor_test.go:13-21` | Same — missing tenant callbacks |
| All domains | No builder tests (builders don't exist) |
| All domains | No REST/handler tests |
| All domains | Tests use simple assertions, not table-driven tests |

### ARCH-009: Dead Code

**Status: WARN**

| Evidence | Issue |
|----------|-------|
| `tenant/provider.go:20-31` | `getByUserID` is defined but never called anywhere |

### ARCH-010: Processor Logger Type

**Status: WARN**

All processors accept `*logrus.Logger` (concrete type) instead of `logrus.FieldLogger` (interface). The handlers then pass `logrus.StandardLogger()` which compounds the issue — context-aware structured logging is lost.

| Evidence | Issue |
|----------|-------|
| `household/processor.go:13,18` | `l *logrus.Logger` — should be `logrus.FieldLogger` |
| `membership/processor.go:13,18` | Same |
| `preference/processor.go:13,18` | Same |
| `tenant/processor.go:13,18` | Same |
| `appcontext/context.go:25` | `Resolve` accepts `*logrus.Logger` — should be `logrus.FieldLogger` |

### ARCH-011: Error Handling in Handlers

**Status: WARN**

Multiple handlers silently discard Transform errors with `_`, which could mask data conversion issues.

| Evidence | Issue |
|----------|-------|
| `household/resource.go:41` | `rm, _ := Transform(m)` — error discarded in list handler |
| `household/resource.go:63,83,103` | Transform errors discarded |
| `membership/resource.go:42,73,101` | Transform errors discarded |
| `preference/resource.go:36,79` | Transform errors discarded |
| `tenant/resource.go:36,52,71` | Transform errors discarded |
| `appcontext/rest.go:88,89,100,105` | `TransformContext` discards all sub-Transform errors |

### ARCH-012: Dockerfile

**Status: WARN**

| Evidence | Issue |
|----------|-------|
| `Dockerfile:1` | Uses `golang:1.26-alpine` — version 1.26 does not exist; guidelines specify Go 1.24+ |

---

## 4. Structural Gaps

| Expected | Status | Notes |
|----------|--------|-------|
| `builder.go` per domain | **MISSING** (all 4 domains) | No fluent builders, no invariant validation |
| `TransformSlice` in rest.go | **MISSING** (all 4 domains) | Inline loops used in handlers |
| Handler tests | **MISSING** | No REST layer tests |
| Provider/administrator tests | **MISSING** | Only processor tests exist |
| Table-driven tests | **MISSING** | Simple sequential assertions used |

---

## 5. Blocking Issues

1. **ARCH-001**: Membership and preference handlers manually parse JSON:API bodies instead of using `server.RegisterInputHandler[T]`. This bypasses automatic tenant parsing and tracing.
2. **ARCH-001**: Household create handler contains cross-domain business logic (membership creation) that belongs in the processor layer.
3. **ARCH-003**: No builder.go in any domain — models can be constructed in invalid states.
4. **ARCH-008**: Test DBs missing `RegisterTenantCallbacks` — tenant filtering is not tested.

---

## 6. Non-Blocking Issues

1. **ARCH-004**: Providers manually filter by tenant_id where GORM callbacks should handle it.
2. **ARCH-007/010**: All handlers use `logrus.StandardLogger()` instead of `d.Logger()` — structured logging with request context is lost.
3. **ARCH-009**: Dead code in `tenant/provider.go` (`getByUserID`).
4. **ARCH-011**: Transform errors silently discarded in handlers.
5. **ARCH-006**: Missing `TransformSlice` functions (cosmetic, inline loops work).
6. **ARCH-012**: Dockerfile references non-existent Go 1.26.
7. **ARCH-008**: Tests are not table-driven; no handler or provider-level tests.

---

## 7. Inputs for /dev-docs

| Objective | Priority |
|-----------|----------|
| Add `builder.go` with fluent builders and invariant validation for all 4 domains | P0 |
| Refactor membership/preference handlers to use `server.RegisterInputHandler[T]` | P0 |
| Move membership auto-creation from household handler to household processor | P0 |
| Add `database.RegisterTenantCallbacks` to all test `setupTestDB` functions | P0 |
| Remove manual tenant_id filtering from providers (rely on GORM callbacks) | P1 |
| Replace `logrus.StandardLogger()` with `d.Logger()` in all handlers | P1 |
| Change processor logger type from `*logrus.Logger` to `logrus.FieldLogger` | P1 |
| Remove dead `getByUserID` from tenant/provider.go | P1 |
| Add `TransformSlice` to each domain's rest.go | P2 |
| Handle Transform errors instead of discarding with `_` | P2 |
| Add table-driven tests and expand test coverage (handlers, providers) | P2 |
| Fix Dockerfile Go version to valid release (1.24) | P2 |

---

## 8. Notes / Ambiguities

1. **appcontext domain**: This is an aggregation domain with no persistence of its own. It cross-cuts tenant, household, membership, and preference. The guidelines don't explicitly address this pattern. The `Resolve()` function creates processors for multiple domains, which is valid orchestration but doesn't follow the standard single-domain processor pattern. Marking as acceptable but worth documenting.

2. **Tenant entity has no `tenant_id` column**: This is correct — the tenant IS the root entity. GORM tenant callbacks are a no-op for this entity. The `getByUserID` provider retrieves all tenants without filtering, which may be intentional for a simplified v1 (per the comment), but the function is unused.

3. **`preference/resource.go:101`**: `json.Unmarshal(nil, nil)` is used as a hack to produce a non-nil error. This is fragile and should use `errors.New()`.

4. **Household update handler reuses `CreateRequest`**: `updateHandler` on line 89 uses `server.InputHandler[CreateRequest]` — should have a separate `UpdateRequest` type for semantic clarity and to support partial updates.

5. **Membership `createHandler` uses `server.GetHandler`**: This is a POST handler using the GET handler type signature, then manually reading the body. This is a structural mismatch that bypasses the framework's input handling.
