# Backend Audit — tracker-service

- **Service Path:** services/tracker-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 7 packages passed, 0 failed
- **Overall:** PASS

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/entry         0.966s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/month         0.723s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/retention     1.187s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule      1.788s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/today         2.056s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem  2.864s
ok  github.com/jtumidanski/home-hub/services/tracker-service/internal/tz            3.259s
```

## Domain Checklist Results

### entry

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:35 NewBuilder, builder.go:48 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:30 |
| DOM-03 | Make(Entity) function | PASS | entity.go:45 |
| DOM-04 | Transform function | PASS | rest.go:43 |
| DOM-05 | TransformSlice function | PASS | rest.go:60 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:35 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:54,82,102,125,148 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19 |
| DOM-09 | Transform errors handled | PASS | Transform returns plain value (no error) |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:25 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to Processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with createEntry/updateEntry/deleteEntry |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:29-46 writeValidationError maps errors to 400/404 |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:22-24 |
| DOM-19 | Request models use flat structure | PASS | rest.go:26-41 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:11-33, processor_test.go:136-168 |

### schedule (internal domain, no HTTP layer)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:24 NewBuilder, builder.go:32 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:25 |
| DOM-03 | Make(Entity) function | PASS | entity.go:36 |
| DOM-04 | Transform function | FAIL | No rest.go — schedule is an internal domain with no HTTP endpoints |
| DOM-05 | TransformSlice function | FAIL | No rest.go — same as DOM-04 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:27 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:22 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11 EntityProvider |
| DOM-16 | administrator.go exists | PASS | administrator.go with CreateSnapshot |
| DOM-20 | Table-driven tests | PASS | processor_test.go:33-92 |

Note: DOM-04/05 are structural FAILs — schedule has no REST layer by design. It is consumed internally by entry, trackingitem, month, and today processors. DOM-07/08/09/12-15/17-19 are N/A.

### trackingitem

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:54 NewBuilder, builder.go:68 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:31 |
| DOM-03 | Make(Entity) function | PASS | entity.go:47 |
| DOM-04 | Transform function | PASS | rest.go:76 |
| DOM-05 | TransformSlice function | PASS | rest.go:101 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:28 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:34,53,79,116,162 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | PASS | Transform returns plain value |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:22 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers call processor only |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/softDelete |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:57-58 (404), resource.go:83-93 (400,409) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:37-39 |
| DOM-19 | Request models use flat structure | PASS | rest.go:41-60, rest.go:62-74 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:12-27 |

## Sub-Domain Checklist Results

### month

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:28-30 own Processor |
| SUB-02 | Has administrator for writes | N/A | Read-only endpoint |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | GET-only endpoints |
| SUB-04 | No manual JSON parsing | PASS | No json.NewDecoder/Unmarshal |

### today

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:31-33 own Processor |
| SUB-02 | Has administrator for writes | N/A | Read-only endpoint |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | GET-only endpoint |
| SUB-04 | No manual JSON parsing | PASS | No manual parsing |

## Summary

### Blocking (must fix)
- **schedule DOM-04**: No Transform function — no rest.go exists (internal domain by design, no HTTP surface)
- **schedule DOM-05**: No TransformSlice function — same root cause

### Non-Blocking (should fix)
- ~~trackingitem resource.go:66-67,99-100,149-150 — schedule enrichment errors silently discarded~~ — FIXED: errors now logged at Warn level
