# Auth Flow Resilience — UX Flow

## Current Flow (Broken)

```
User opens app
  → ProtectedRoute renders skeleton
  → GET /users/me (fails 500, retries 3x with backoff: ~7s)
  → Eventually resolves or errors
  → If error: user stays on skeleton until React Query gives up
  → If success: GET /context (potentially same 500 retry issue)
  → Finally routes to login, onboarding, or dashboard

User's session expires while on dashboard
  → Background queries (tasks, reminders) get 401
  → Queries silently fail or retry
  → User sees stale/broken UI with no indication
  → No redirect to login, no token refresh attempted
```

## Target Flow

### App Bootstrap (Happy Path)
```
User opens app
  → ProtectedRoute renders skeleton
  → GET /users/me (1 retry max, 500ms delay)
  → Success → GET /context (1 retry max, 500ms delay)
  → Success → Route to /app (dashboard)
  Total: < 2 seconds
```

### App Bootstrap (500 on First Try)
```
User opens app
  → ProtectedRoute renders skeleton
  → GET /users/me → 500
  → Retry after 500ms → Success
  → GET /context → Success
  → Route to /app (dashboard)
  Total: < 3 seconds
```

### App Bootstrap (Unauthenticated)
```
User opens app
  → ProtectedRoute renders skeleton
  → GET /users/me → 401
  → Interceptor attempts POST /auth/token/refresh
  → Refresh fails (no valid refresh token)
  → Redirect to /login
  Total: < 2 seconds
```

### Session Expiry During Use
```
User is on dashboard, access token expires (15 min)
  → Background query (e.g., GET /tasks) → 401
  → Interceptor pauses request
  → POST /auth/token/refresh (refresh token still valid, 7-day TTL)
  → Refresh succeeds → new access token cookie set
  → Original request retried with new token → succeeds
  → User never notices
```

### Session Fully Expired
```
User returns after days, both tokens expired
  → Any API call → 401
  → Interceptor attempts POST /auth/token/refresh → 401
  → Interceptor sets isRedirecting = true
  → Clears React Query cache
  → Redirects to /login
  → Other in-flight 401s see isRedirecting flag, do nothing
  → User sees login page (single clean redirect, no toast storm)
```

## 401 Interceptor Sequence

```
API response received
  │
  ├─ Status !== 401 → normal flow
  │
  └─ Status === 401
       │
       ├─ isRedirecting? → reject (silent)
       │
       ├─ Is this the refresh endpoint itself? → redirect to /login
       │
       ├─ refreshPromise exists? → await it
       │     ├─ resolved true → retry original request
       │     └─ resolved false → reject (silent, redirect already triggered)
       │
       └─ No refresh in progress
             → set refreshPromise = authService.refreshToken()
             ├─ Success → clear refreshPromise, retry original request
             └─ Failure → set isRedirecting, clear cache, redirect to /login
```
