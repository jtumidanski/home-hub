# Auth Flow Resilience — Implementation Plan

Last Updated: 2026-03-25

---

## Executive Summary

This task fixes two frontend auth flow issues: slow app bootstrap caused by aggressive HTTP retries on auth-critical requests, and missing 401 handling that leaves users on broken pages when their session expires. The fix is entirely frontend — adding a 401 interceptor with transparent token refresh to the API client, reducing retry configuration for bootstrap requests, and wiring up a global error handler in React Query as a safety net.

## Current State Analysis

### What exists today

1. **API Client** (`frontend/src/lib/api/client.ts`):
   - `fetchWithRetry` retries all 5xx responses 3 times with exponential backoff (1s → 2s → 4s)
   - No special handling for 401 responses — they are thrown as `AppError` with type `"auth"`
   - Request deduplication exists but is unrelated to auth
   - `credentials: "include"` is already set on all requests (cookies sent automatically)

2. **Auth Bootstrap** (`frontend/src/components/providers/auth-provider.tsx`):
   - Calls `useMe()` → on success calls `useAppContext()`
   - Both React Query hooks have `retry: false`, but the underlying `api.get()` still retries 500s 3x at the HTTP level
   - `isLoading` blocks rendering until both resolve

3. **Token Refresh** (`frontend/src/services/api/auth.ts`):
   - `authService.refreshToken()` exists and calls `POST /auth/token/refresh`
   - **Nothing in the codebase calls it** — the method is unused

4. **React Query** (`frontend/src/components/providers/query-provider.tsx`):
   - Default `retry: 3` for queries — means non-bootstrap queries retry 401s at the React Query level too
   - No global `onError` handler

5. **Error Classification** (`frontend/src/lib/api/errors.ts`):
   - 401/403 → `"auth"` type, helper `requiresAuthentication()` exists but is unused

### Root causes

| Symptom | Root Cause |
|---------|-----------|
| Slow bootstrap on 500s | `fetchWithRetry` defaults to 3 retries with exponential backoff; `useMe`/`useAppContext` don't override |
| No redirect on 401 | No 401 interceptor exists; `authService.refreshToken()` is never called |
| Background 401 accumulation | React Query retries 401s 3 times (default), no global error handler redirects |

## Proposed Future State

1. **API client intercepts 401s globally**: Any 401 triggers a single-flight token refresh. On success, the original request retries transparently. On failure, user is redirected to `/login` once.
2. **Bootstrap is fast**: `/users/me` and `/context` use 1 retry with 500ms delay instead of 3 retries with exponential backoff.
3. **React Query doesn't retry auth errors**: Global query config skips retries for 401s. Global `onError` acts as a safety net for any auth errors that slip through.

## Implementation Phases

### Phase 1: API Client — 401 Interceptor (Core)

The highest-value change. All other phases depend on the interceptor being in place.

### Phase 2: Reduced Bootstrap Retries

Independent of Phase 1 but logically related. Threads `RequestOptions` through the service and hook layers for `/users/me` and `/context`.

### Phase 3: React Query Safety Net

Adds a global retry filter and `onError` handler to prevent React Query from retrying or silently swallowing auth errors.

### Phase 4: Testing

Unit tests for the interceptor logic and integration-style tests for the auth flow.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Refresh loop: 401 → refresh → 401 → refresh → ... | Medium | High | Exclude the refresh endpoint itself from the interceptor; clear `refreshPromise` on failure |
| Race between interceptor redirect and React Query error handler | Low | Medium | `isRedirecting` guard checked in both layers |
| Breaking existing login/logout flows | Low | High | Login page doesn't call authenticated endpoints; logout already redirects via `onSettled` |
| Refresh token endpoint returns non-401 error (e.g., 500) | Low | Medium | Treat any refresh failure as session-dead; redirect to login |

## Success Metrics

- App bootstrap < 2s on healthy backend
- App bootstrap < 3s with one 500 on `/users/me`
- Transparent token refresh with zero user-visible interruption when access token expires
- Single redirect on full session expiry (no toast storms, no duplicate navigations)
- All existing tests pass

## Required Resources and Dependencies

- **No backend changes** — refresh endpoint already exists and works
- **No new dependencies** — uses existing React Query, react-router-dom, and the API client

## Timeline Estimates

| Phase | Effort | Estimate |
|-------|--------|----------|
| Phase 1: 401 Interceptor | M | Core logic + tests |
| Phase 2: Bootstrap Retries | S | Threading options through 4 files |
| Phase 3: React Query Safety Net | S | Config change + global handler |
| Phase 4: Testing | M | Unit + integration tests |
| **Total** | **M-L** | |
