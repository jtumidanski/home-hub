# Auth Flow Resilience — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25
---

## 1. Overview

The frontend has two auth flow issues that degrade the user experience. First, when the backend returns 500s during app bootstrap (the `/users/me` and `/context` calls), the API client retries each request 3 times with exponential backoff (1s → 2s → 4s), causing users to stare at a loading skeleton for up to 7+ seconds before being routed to login or onboarding. Second, when a user's 15-minute access token expires, background polling queries silently accumulate 401 errors with no mechanism to refresh the token or redirect to login.

The backend already supports token refresh via `POST /auth/token/refresh` using a 7-day refresh token cookie, but the frontend never calls it. This task adds a global 401 interceptor with transparent token refresh, reduces retry overhead on auth-critical bootstrap requests, and ensures expired sessions are handled gracefully.

## 2. Goals

Primary goals:
- Add a global 401 response interceptor that attempts token refresh before redirecting to `/login`
- Reduce time-to-decision on app bootstrap by limiting retries on `/users/me` and `/context`
- Ensure multiple simultaneous 401s produce a single refresh attempt and (if needed) a single redirect
- Cancel/abort in-flight queries when a session is determined to be dead

Non-goals:
- Backend changes to auth or session management
- Proactive session timeout warnings ("session expires in X minutes")
- Changes to the onboarding wizard flow itself
- Offline or network-down handling

## 3. User Stories

- As a user opening the app for the first time, I want to reach login or onboarding within 1-2 seconds, not 7+, so the app feels responsive.
- As a user whose session has expired, I want the app to silently refresh my token so I don't notice the expiration at all.
- As a user whose refresh token has also expired, I want to be redirected to login immediately instead of seeing broken UI with background errors.

## 4. Functional Requirements

### 4.1 Global 401 Interceptor

- When any API response returns 401, the interceptor MUST attempt a single token refresh (`POST /auth/token/refresh`) before treating the session as dead.
- If the refresh succeeds, the original failed request MUST be transparently retried with the new credentials.
- If the refresh fails (401 or any error), the interceptor MUST:
  1. Clear React Query cache
  2. Redirect to `/login`
- The interceptor MUST be implemented in the API client layer (`frontend/src/lib/api/client.ts`) so it applies to all requests uniformly.

### 4.2 Single-Flight Refresh

- If multiple requests fail with 401 simultaneously, only ONE refresh attempt MUST be made.
- All other 401'd requests MUST wait for the single refresh attempt to resolve, then either retry (on success) or abort (on failure).
- Implementation: a shared promise reference (e.g., `private refreshPromise: Promise<boolean> | null`) that concurrent callers await.

### 4.3 Reduced Bootstrap Retries

- The `/users/me` and `/context` API calls MUST use a maximum of 1 retry with a short delay (~500ms) instead of the default 3 retries with exponential backoff.
- This applies at the HTTP client level (`fetchWithRetry`), not the React Query level (which already has `retry: false` for these queries).
- The `useMe()` and `useAppContext()` hooks should pass `{ maxRetries: 1, retryDelay: 500 }` as request options to the underlying service calls.

### 4.4 Auth Bootstrap Request Options

- `authService.getMe()` MUST be called with `{ maxRetries: 1, retryDelay: 500 }`.
- `accountService.getContext()` MUST be called with `{ maxRetries: 1, retryDelay: 500 }`.
- Other API calls retain the default retry behavior (3 retries, exponential backoff).

### 4.5 Redirect Deduplication

- Once a redirect to `/login` is triggered, no further redirects or error toasts should fire from other in-flight queries.
- A simple boolean guard (e.g., `private isRedirecting = false`) in the API client is sufficient.

## 5. API Surface

No new endpoints. Existing endpoints used:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/users/me` | Bootstrap: identify current user |
| GET | `/api/v1/account/context` | Bootstrap: get tenant/household context |
| POST | `/api/v1/auth/token/refresh` | Refresh expired access token using refresh cookie |

## 6. Data Model

No data model changes. All changes are frontend-only.

## 7. Service Impact

### Frontend (`frontend/`)

| File | Change |
|------|--------|
| `src/lib/api/client.ts` | Add 401 interceptor with single-flight refresh, redirect guard, and refresh retry logic |
| `src/services/api/auth.ts` | Add `RequestOptions` parameter to `getMe()` |
| `src/services/api/account.ts` | Add `RequestOptions` parameter to `getContext()` |
| `src/lib/hooks/api/use-auth.ts` | Pass `{ maxRetries: 1, retryDelay: 500 }` to `getMe()` |
| `src/lib/hooks/api/use-context.ts` | Pass `{ maxRetries: 1, retryDelay: 500 }` to `getContext()` |
| `src/components/providers/query-provider.tsx` | Optionally add global `onError` for React Query as a safety net |

### Backend

No changes required.

## 8. Non-Functional Requirements

- **Performance**: Auth bootstrap (from page load to routed state) should complete in under 2 seconds under normal conditions, and under 3 seconds when the first attempt to `/users/me` or `/context` fails with a 500.
- **Security**: The refresh token cookie is httpOnly and path-restricted; the interceptor must use the existing `authService.refreshToken()` which sends cookies automatically via `credentials: "include"`.
- **Observability**: Failed refresh attempts should log to the browser console for debugging.
- **Multi-tenancy**: No tenant-scoped changes; auth and refresh are tenant-agnostic (Pattern C).

## 9. Open Questions

None — all questions resolved during spec collaboration.

## 10. Acceptance Criteria

- [ ] When access token expires, background queries trigger a transparent token refresh and retry successfully without user-visible errors.
- [ ] When both access and refresh tokens are expired, the user is redirected to `/login` within 1 second of the first 401.
- [ ] Multiple simultaneous 401s result in exactly one refresh attempt and (if failed) one redirect.
- [ ] App bootstrap with healthy backend completes in under 2 seconds.
- [ ] App bootstrap when `/users/me` returns a single 500 before succeeding completes in under 3 seconds.
- [ ] No duplicate redirect or toast storms when session expires while multiple queries are polling.
- [ ] Existing auth flows (login, logout, onboarding) continue to work unchanged.
- [ ] All existing frontend tests pass.
