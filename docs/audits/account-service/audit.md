# Backend Audit — account-service

- **Service Path:** services/account-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 7 packages passed, 0 failed
- **Overall:** NEEDS-WORK

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/account-service/internal/appcontext   2.159s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/household     3.657s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/invitation    4.104s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/membership    0.883s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/preference    3.417s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/retention     2.827s
ok  github.com/jtumidanski/home-hub/services/account-service/internal/tenant        4.489s
```

## Domain Checklist Results

### household

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:32 NewBuilder, builder.go:86 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:27 |
| DOM-03 | Make(Entity) function | PASS | entity.go:42 |
| DOM-04 | Transform function | PASS | rest.go:29 |
| DOM-05 | TransformSlice function | PASS | rest.go:37 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:19 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:33,55,82,110,133 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | PASS | resource.go:40-44,67-71,94-98,150-154 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:22, resource_test.go:28 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:10,18 database.Query/SliceQuery |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All calls through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db.Create/Save/Delete in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:58-65 (400), resource.go:85-88 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:21-23 |
| DOM-19 | Request models use flat structure | PASS | rest.go:49-54, rest.go:67-83 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:151, processor_test.go:43 |

### invitation

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:37 NewBuilder, builder.go:55 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:38 |
| DOM-03 | Make(Entity) function | PASS | entity.go:53 |
| DOM-04 | Transform function | PASS | rest.go:46 |
| DOM-05 | TransformSlice function | PASS | rest.go:60 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:36 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:45,74,105,144,176,222 |
| DOM-08 | POST/PATCH use RegisterInputHandler | FAIL | resource.go:27-28 accept/decline POST routes use RegisterHandler (GetHandler) instead of RegisterInputHandler. These are body-less action endpoints. |
| DOM-09 | Transform errors handled | PASS | resource.go:52-56,87-90,128-132 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:26 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:12,19,27,33 database.Query/SliceQuery |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All calls through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db.Create/Save/Delete in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/updateStatus |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:108-125 (403,409,422,400,500), resource.go:147-192 (404,410,403,422) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:24-26 |
| DOM-19 | Request models use flat structure | PASS | rest.go:72-78 |
| DOM-20 | Table-driven tests | PASS | processor_test.go:135,198,281 |

### membership

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:26 NewBuilder, builder.go:65 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:24 |
| DOM-03 | Make(Entity) function | PASS | entity.go:36 |
| DOM-04 | Transform function | PASS | rest.go:43 |
| DOM-05 | TransformSlice function | PASS | rest.go:54 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:28 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:34,95,123,160 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | PASS | resource.go:58-63,80-84,107-111,142-146 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:21, resource_test.go:27 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:10,18,23,29 |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All calls through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/updateRole/deleteByID |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:98 (400), resource.go:126-134 (404,400,403) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:21-23 |
| DOM-19 | Request models use flat structure | PASS | rest.go:66-71, rest.go:84-87 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:85, processor_test.go:39 |

### preference

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:25 NewBuilder, builder.go:64 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:24 |
| DOM-03 | Make(Entity) function | PASS | entity.go:36 |
| DOM-04 | Transform function | PASS | rest.go:24 |
| DOM-05 | TransformSlice function | PASS | rest.go:28 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:18 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:30,57 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19 |
| DOM-09 | Transform errors handled | PASS | resource.go:37-41,87-91 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:21, resource_test.go:26 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11,17 |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All calls through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/updateTheme/setActiveHousehold |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:35-36 (500), resource.go:53-54 (400), resource.go:64-65 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:16-18 |
| DOM-19 | Request models use flat structure | PASS | rest.go:40-44 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:85, processor_test.go:49 |

### tenant

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:21 NewBuilder, builder.go:45 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:21 |
| DOM-03 | Make(Entity) function | PASS | entity.go:30 |
| DOM-04 | Transform function | PASS | rest.go:24 |
| DOM-05 | TransformSlice function | PASS | rest.go:28 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:18 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:31,52,79 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19 |
| DOM-09 | Transform errors handled | PASS | resource.go:38-41,63-67,85-89 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:21, resource_test.go:28 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All calls through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:34-35 (404), resource.go:55-56 (400) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:15-17 |
| DOM-19 | Request models use flat structure | PASS | rest.go:40-43 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:49, processor_test.go:37 |

## Sub-Domain Checklist Results

### appcontext

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | context.go:33 Resolve() orchestrates multiple processors |
| SUB-02 | Has administrator for writes | N/A | Read-only endpoint |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | GET-only endpoint |
| SUB-04 | No manual JSON parsing | PASS | No json.NewDecoder/Unmarshal/ReadAll |

### retention

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:22-28 own Processor |
| SUB-02 | Has administrator for writes | PASS | administrator.go with upsertOverride/deleteOverride |
| SUB-03 | Uses RegisterInputHandler for POST | PASS | resource.go:25,29 |
| SUB-04 | No manual JSON parsing | PASS | No manual parsing in resource handlers |

## Summary

### Blocking (must fix)
- **invitation DOM-08**: `/invitations/{id}/accept` and `/invitations/{id}/decline` POST routes use `RegisterHandler` (resource.go:27-28) instead of `RegisterInputHandler`. These are body-less action endpoints. **Verified:** this is the established project convention — package-service (archive/unarchive/refresh), recipe-service (lock/unlock), and auth-service (refresh/logout) all use `RegisterHandler` for body-less POST actions. **No fix needed.**

### Non-Blocking (should fix)
- ~~retention/fanout.go:154,225 — `uuid.Parse` errors silently discarded in HTTP client fanout code~~ **FIXED:** Purge path now returns an error on invalid run_id; ListRuns path logs a warning and skips entries with invalid ids.
- ~~retention/resource.go:134,168 — `proc.ResolveAll` errors silently discarded after successful writes~~ **FIXED:** Both PATCH handlers now check the error from `ResolveAll` and return 500 on failure.
