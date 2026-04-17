
# Migration Plan

This task is a **breaking API change**. The frontend build that drops the `X-Timezone` header and starts sending `?date=` / `?start=`+`?end=` parameters cannot ship before every affected backend service has been updated to accept those parameters, or users hitting the old backend during the deploy window will see 400s.

## Two options

### Option A — Transitional accept-both window (recommended)

Ship in three PRs, strictly ordered.

**PR 1 — Backend: accept the new params, keep old behavior as fallback.**
- Each affected endpoint accepts `?date=YYYY-MM-DD` (or `?start`/`?end`) when present.
- When the parameter is absent, fall back to the current `X-Timezone` / UTC behavior. No 400 yet.
- Ship and deploy to production. No client-visible change.

**PR 2 — Frontend: start sending the new params.**
- Every hook call sends the date parameter.
- Query keys include the date.
- Remove `X-Timezone` injection from `client.ts`.
- Widgets use the `useLocalDate` poller so the date refreshes past midnight.
- Ship and deploy. At this point, no request in production relies on the fallback.

**PR 3 — Backend: make the params required and delete the tz resolution.**
- Parameters become required. Missing/malformed → 400.
- Delete `internal/tz/` packages in tracker, workout, calendar services.
- Delete productivity's `resolveTimezone`.
- Delete `accountBaseURL` constructor wiring in `cmd/main.go` for affected services.
- Update `rest.md` for each service.

This is the safest ordering. At no point is there a backend + frontend version mismatch that returns 400s to real users.

### Option B — Big-bang single PR

Everything in one PR. The docker-compose and k8s manifests deploy all services atomically (or close enough) — the brief window where nginx routes a request to an old replica while a new frontend build is already cached client-side is the only real risk.

Viable for this repo because:
- Solo deploy (no staged rollouts in prod).
- The user base is tiny (household usage).
- A refresh fixes any cached-frontend-hits-new-backend mismatch.

Choose Option B only if the user explicitly prefers the single-PR route for simplicity and accepts the ~minutes-long risk window.

## Recommendation

Option A. Three small PRs are easier to review, easier to revert, and guarantee zero user-facing 400s. The intermediate state (backend accepts both, frontend still sends old) is a normal state the code already handles today.

## Deploy gating checklist for PR 3

Before merging PR 3, verify:

- PR 2 has been deployed to production for at least one hour, and all active browser sessions have refreshed to the new frontend bundle (check that `X-Timezone` headers have stopped appearing in access logs — or grep for any 400s on `?date=` parse failures, which would indicate stale clients).
- No Bruno collections or external scripts still hit the affected endpoints without the new params. Scan `bruno/` and update collections in the same PR.
- Nobody on the team has pending local branches that depend on the transitional backend behavior.

## Rollback plan

- **PR 1 rollback:** benign. Old behavior still works.
- **PR 2 rollback:** restore `X-Timezone` injection + old hook signatures. Backend still accepts `?date=` (from PR 1) so there's no coordination needed — just revert the frontend.
- **PR 3 rollback:** restore the `internal/tz/` packages and the accept-both fallback. This is the most painful rollback; do it only if PR 3 introduced a regression that can't be forward-fixed quickly.
