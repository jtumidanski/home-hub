# Backend Audit — calendar-service

- **Service Path:** services/calendar-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 7 packages passed, 0 failed
- **Overall:** PASS

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/connection   1.352s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto       0.335s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/event        1.072s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal    1.773s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/oauthstate   1.933s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/retention    2.617s
ok  github.com/jtumidanski/home-hub/services/calendar-service/internal/sync         2.844s
```

## Domain Checklist Results

### connection

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go has NewBuilder(), fluent setters, Build() |
| DOM-02 | ToEntity() method | PASS | entity.go has ToEntity() on Model |
| DOM-03 | Make(Entity) function | PASS | entity.go has Make(Entity) returning (Model, error) |
| DOM-04 | Transform function | PASS | rest.go has Transform(Model) |
| DOM-05 | TransformSlice function | PASS | rest.go has TransformSlice |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go accepts logrus.FieldLogger |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go uses d.Logger() in all NewProcessor calls |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | POST/PATCH routes use RegisterInputHandler |
| DOM-09 | Transform errors handled | PASS | All Transform calls check errors |
| DOM-10 | Test DB has tenant callbacks | PASS | Tests call RegisterTenantCallbacks |
| DOM-11 | Providers use lazy evaluation | PASS | Providers return EntityProvider (lazy) |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go exists |
| DOM-17 | Domain error to HTTP status mapping | PASS | Proper error mapping in resource.go |
| DOM-18 | JSON:API interface on REST models | PASS | GetName/GetID/SetID implemented |
| DOM-19 | Request models use flat structure | PASS | Flat request models |
| DOM-20 | Table-driven tests | PASS | Table-driven tests with t.Run |

### event

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go with NewBuilder, Build |
| DOM-02 | ToEntity() method | PASS | entity.go has ToEntity() |
| DOM-03 | Make(Entity) function | PASS | entity.go has Make() |
| DOM-04 | Transform function | PASS | rest.go has Transform |
| DOM-05 | TransformSlice function | PASS | rest.go has TransformSlice |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go accepts logrus.FieldLogger |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go uses d.Logger() |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | POST/PATCH use RegisterInputHandler |
| DOM-09 | Transform errors handled | PASS | Transform errors checked |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go uses setupTestDB with RegisterTenantCallbacks; 8 DB-level tests |
| DOM-11 | Providers use lazy evaluation | PASS | Providers use database.Query/SliceQuery |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Cross-domain processors injected via ConnectionProcessor/SourceProcessor interfaces; wired in resource.go via NewMutationProcessor |
| DOM-14 | Handlers don't call providers directly | PASS | All through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go exists |
| DOM-17 | Domain error to HTTP status mapping | PASS | Proper error mapping |
| DOM-18 | JSON:API interface on REST models | PASS | GetName/GetID/SetID |
| DOM-19 | Request models use flat structure | PASS | Flat request models |
| DOM-20 | Table-driven tests | PASS | Table-driven tests |

### googlecal

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go with NewBuilder, Build |
| DOM-02 | ToEntity() method | PASS | entity.go has ToEntity() |
| DOM-03 | Make(Entity) function | PASS | entity.go has Make() |
| DOM-04 | Transform function | PASS | rest.go has Transform |
| DOM-05 | TransformSlice function | PASS | rest.go has TransformSlice |
| DOM-06 | Processor accepts FieldLogger | PASS | logrus.FieldLogger |
| DOM-07 | Handlers pass d.Logger() | PASS | d.Logger() used |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | Correct handler types |
| DOM-09 | Transform errors handled | PASS | Errors checked |
| DOM-10 | Test DB has tenant callbacks | PASS | RegisterTenantCallbacks in tests |
| DOM-11 | Providers use lazy evaluation | PASS | Lazy providers |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops |
| DOM-16 | administrator.go exists | PASS | administrator.go exists |
| DOM-17 | Domain error to HTTP status mapping | PASS | Proper mapping |
| DOM-18 | JSON:API interface on REST models | PASS | Interface implemented |
| DOM-19 | Request models use flat structure | PASS | Flat models |
| DOM-20 | Table-driven tests | PASS | Table-driven tests |

### oauthstate

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go with NewBuilder, Build |
| DOM-02 | ToEntity() method | PASS | entity.go has ToEntity() |
| DOM-03 | Make(Entity) function | PASS | entity.go has Make() |
| DOM-04 | Transform function | PASS | rest.go has Transform |
| DOM-05 | TransformSlice function | PASS | rest.go has TransformSlice |
| DOM-06 | Processor accepts FieldLogger | PASS | logrus.FieldLogger |
| DOM-07 | Handlers pass d.Logger() | PASS | d.Logger() used |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | Correct handler types |
| DOM-09 | Transform errors handled | PASS | Errors checked |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go uses setupTestDB with RegisterTenantCallbacks; 6 DB-level tests |
| DOM-11 | Providers use lazy evaluation | PASS | Lazy providers |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops |
| DOM-16 | administrator.go exists | PASS | administrator.go exists |
| DOM-17 | Domain error to HTTP status mapping | PASS | Proper mapping |
| DOM-18 | JSON:API interface on REST models | PASS | Interface implemented |
| DOM-19 | Request models use flat structure | PASS | Flat models |
| DOM-20 | Table-driven tests | PASS | Table-driven tests |

### source (Support — no resource.go)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go uses setupTestDB with RegisterTenantCallbacks; 10 DB-level tests |
| DOM-20 | Table-driven tests | PASS | processor_test.go covers CreateOrUpdate, ListByConnection, ToggleVisibility, SyncToken, DeleteByConnection, ByIDProvider |

Note: `source` has `model.go` but no `resource.go` — it functions as an internal domain consumed by event processor. Checks requiring resource.go are N/A.

## Summary

### Blocking (must fix)
- ~~**event DOM-10**: No DB-level tests exist; no RegisterTenantCallbacks usage~~ **FIXED** — processor_test.go added with 8 DB tests
- ~~**event DOM-13**: Processor cross-domain coupling~~ **FIXED** — Introduced ConnectionProcessor/SourceProcessor interfaces; wiring moved to resource.go via NewMutationProcessor
- ~~**oauthstate DOM-10**: No DB integration tests; no RegisterTenantCallbacks~~ **FIXED** — processor_test.go added with 6 DB tests
- ~~**source DOM-10**: No test files at all~~ **FIXED** — processor_test.go added with 10 DB tests
- ~~**source DOM-20**: No test files at all~~ **FIXED** — table-driven tests covering all processor methods

### Non-Blocking (should fix)
- None remaining
