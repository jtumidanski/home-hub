# Backend Audit — package-service

- **Service Path:** services/package-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 3 packages passed, 0 failed
- **Overall:** PASS

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/package-service/internal/carrier        0.955s
ok  github.com/jtumidanski/home-hub/services/package-service/internal/tracking       1.059s
ok  github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent  1.572s
```

## Domain Checklist Results

### tracking

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:67 NewBuilder, builder.go:87 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:45 |
| DOM-03 | Make(Entity) function | PASS | entity.go:67 |
| DOM-04 | Transform function | PASS | rest.go:164 |
| DOM-05 | TransformSlice function | PASS | rest.go:192 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:44,51 logrus.FieldLogger |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:48,99,124,163,206,232,259,294,339 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:23-24 |
| DOM-09 | Transform errors handled | PASS | resource.go:75-79,108-112,146-150,187-191,242-246,276-280,368-372 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:21 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11-12 returns EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | resource.go:312-327 detectCarrierHandler delegates to proc.DetectCarrier() + TransformDetection() |
| DOM-14 | Handlers don't call providers directly | PASS | All through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No db ops in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/deleteByID |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:57-67 (400,409,422), resource.go:170-176 (404,403), resource.go:345-361 (429) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:29-31, rest.go:40-42, rest.go:78-80, rest.go:140-149, rest.go:160-162 |
| DOM-19 | Request models use flat structure | PASS | rest.go:131-138, rest.go:152-158 |
| DOM-20 | Table-driven tests | PASS | builder_test.go:21-157, processor_test.go:97-129 |

### trackingevent (internal domain, no HTTP layer)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:27 NewBuilder, builder.go:37 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:36 |
| DOM-03 | Make(Entity) function | PASS | entity.go:49 |
| DOM-04 | Transform function | N/A | No rest.go — transforms in tracking/rest.go:111 |
| DOM-05 | TransformSlice function | N/A | Transforms in tracking/rest.go:122 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9-13 database.SliceQuery |
| DOM-16 | administrator.go exists | PASS | administrator.go with Create |
| DOM-20 | Table-driven tests | PASS | builder_test.go:11-104 with t.Run |

## Summary

### Blocking (must fix)
- None — all blocking issues resolved.

### Non-Blocking (should fix)
- None remaining.

## Fixes Applied (2026-04-12)

1. **DOM-13 fixed**: Moved carrier detection logic from `detectCarrierHandler` into `Processor.DetectCarrier()` method and added `TransformDetection()` in rest.go. Handler now follows the standard processor delegation pattern.
