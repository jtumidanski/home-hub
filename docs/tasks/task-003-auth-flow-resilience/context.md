# Auth Flow Resilience — Context

Last Updated: 2026-03-25

---

## Key Files

### Files to Modify

| File | Role | Changes |
|------|------|---------|
| `frontend/src/lib/api/client.ts` | HTTP client with retry, dedup, caching | Add 401 interceptor, single-flight refresh, redirect guard |
| `frontend/src/services/api/auth.ts` | Auth service (login URLs, /me, refresh, logout) | Accept `RequestOptions` on `getMe()` |
| `frontend/src/services/api/account.ts` | Account service (context, tenants, households) | Accept `RequestOptions` on `getContext()` |
| `frontend/src/lib/hooks/api/use-auth.ts` | React Query hook for `/users/me` | Pass `{ maxRetries: 1, retryDelay: 500 }` to service |
| `frontend/src/lib/hooks/api/use-context.ts` | React Query hook for `/contexts/current` | Pass `{ maxRetries: 1, retryDelay: 500 }` to service |
| `frontend/src/components/providers/query-provider.tsx` | React Query client config | Add retry filter for auth errors, global `onError` |

### Files for Reference (read-only)

| File | Why |
|------|-----|
| `frontend/src/components/providers/auth-provider.tsx` | Understand bootstrap flow (`useMe` → `useAppContext`) |
| `frontend/src/components/features/navigation/protected-route.tsx` | Understand routing guards |
| `frontend/src/lib/api/errors.ts` | Error classification (`isAuthError`, `requiresAuthentication`) |
| `frontend/src/context/tenant-context.tsx` | Understand tenant context lifecycle |

### Backend Reference

| File | Why |
|------|-----|
| `services/auth-service/internal/resource/resource.go` | Refresh endpoint implementation, cookie setting |
| `services/auth-service/internal/authflow/processor.go` | `HandleRefresh()` — token rotation logic |
| `shared/go/auth/auth.go` | Auth middleware — when/why 401s are returned |

## Key Decisions

1. **Interceptor lives in ApiClient, not React Query** — ensures all HTTP calls (including non-query mutations and manual fetches) get 401 handling uniformly.

2. **Single-flight refresh via shared promise** — `private refreshPromise: Promise<boolean> | null` pattern. First 401 creates the promise; concurrent 401s await it.

3. **Redirect guard via boolean flag** — `private isRedirecting = false` prevents cascade effects once we've committed to redirecting.

4. **Refresh endpoint excluded from interception** — if `POST /auth/token/refresh` itself returns 401, we go straight to redirect (no infinite loop).

5. **Bootstrap retries threaded via RequestOptions** — service methods accept optional `RequestOptions`, hooks pass `{ maxRetries: 1, retryDelay: 500 }`. Clean separation of concerns.

6. **React Query retry filter** — `retry` callback checks error type; returns `false` for auth errors so React Query doesn't independently retry 401s that the interceptor already handled.

## Dependencies Between Tasks

```
Phase 1 (401 Interceptor)
  ├── Phase 2 (Bootstrap Retries) — independent, can be parallel
  ├── Phase 3 (React Query Safety Net) — independent, can be parallel
  └── Phase 4 (Testing) — depends on all above
```

## Token Lifecycle Reference

| Token | Storage | TTL | Path |
|-------|---------|-----|------|
| Access token | httpOnly cookie `access_token` | 15 minutes | `/` |
| Refresh token | httpOnly cookie `refresh_token` | 7 days | `/api/v1/auth` |

- Access token is a JWT validated by downstream services via JWKS
- Refresh token is a random string, hashed (SHA256) and stored in `auth.refresh_tokens` table
- Token refresh rotates both tokens (old refresh token revoked, new pair issued)
