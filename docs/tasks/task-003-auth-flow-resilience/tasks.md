# Auth Flow Resilience — Task Checklist

Last Updated: 2026-03-25

---

## Phase 1: API Client — 401 Interceptor

- [x] **1.1** Add `refreshPromise` and `isRedirecting` private fields to `ApiClient`
- [x] **1.2** Add `onAuthFailure` callback registration to `ApiClient`
- [x] **1.3** Implement `attemptRefresh()` method (raw fetch, not via api.post)
- [x] **1.4** Implement `handleUnauthorized()` method with single-flight refresh
- [x] **1.5** Wire 401 interception into `get`, `post`, `put`, `patch`, `delete`, `upload`, `download` methods
- [x] **1.6** Add `resetAuthState()` public method

## Phase 2: Reduced Bootstrap Retries

- [x] **2.1** Add `RequestOptions` parameter to `authService.getMe()`
- [x] **2.2** Add `RequestOptions` parameter to `accountService.getContext()`
- [x] **2.3** Pass `{ maxRetries: 1, retryDelay: 500 }` in `useMe()` hook
- [x] **2.4** Pass `{ maxRetries: 1, retryDelay: 500 }` in `useAppContext()` hook

## Phase 3: React Query Safety Net

- [x] **3.1** Add retry filter to exclude auth errors from React Query retries
- [x] **3.2** Wire `onAuthFailure` callback from QueryProvider to clear cache

## Phase 4: Testing

- [x] **4.1** Unit tests for 401 interceptor logic (10 test cases)
- [x] **4.2** Existing `use-auth.test.tsx` passes (no changes needed — mock covers new options)
- [x] **4.3** Existing `use-context.test.tsx` passes (no changes needed — mock covers new options)
- [x] **4.4** All 267 tests pass across 29 test files

## Final Verification

- [ ] **5.1** Manual smoke test: fresh login → onboarding → dashboard
- [ ] **5.2** Manual smoke test: wait for access token expiry → verify transparent refresh
- [ ] **5.3** Manual smoke test: clear all cookies → verify redirect to login
- [x] **5.4** TypeScript type check passes
- [x] **5.5** Production build succeeds
