# Frontend Audit — branch `fix/ui-cleanup-favicon-sidebar`

- **Audit Scope:** `git diff main...HEAD -- frontend/` (UI cleanup: tomorrow-widget tz fix, shadcn-style sidebar primitives, favicon/brand mark, header spacing)
- **Guidelines Source:** frontend-dev-guidelines skill (FE-* checklist)
- **Date:** 2026-07-09
- **Build:** COULD NOT RUN (environment) — `tsc -b` PASS
- **Tests:** COULD NOT RUN (environment)
- **Overall:** PASS (with 2 non-blocking notes)

## Build & Test Results

`npm run build` = `tsc -b && vite build`. The `vite build` and `vitest run` steps cannot execute in this
environment: `frontend/node_modules` was installed for **musl** (Alpine/Docker — `@rolldown/binding-linux-x64-musl`
present) and is **root-owned**, so on this glibc WSL2 host rolldown fails with
`Cannot find module '@rolldown/binding-linux-x64-gnu'` and `npm install` of the gnu binding is rejected with EACCES.
This is a host/`node_modules` platform mismatch, not a defect in the branch.

The type-check half of the build was run directly and passes:

```
$ npx tsc -b --force
TSC_EXIT=0
```

No TypeScript errors — strict mode (incl. `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`) is satisfied.
Tests should be run in CI / a glibc install before merge to close the objective gate.

## File Inventory

- `frontend/public/favicon.svg` — Other (static asset)
- `frontend/src/components/common/brand-mark.tsx` — Component (common, presentational) [NEW]
- `frontend/src/components/ui/sidebar.tsx` — Component (ui primitive) [NEW]
- `frontend/src/components/features/calendar/calendar-utils.ts` — Other (pure util)
- `frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx` — Component (feature)
- `frontend/src/components/features/navigation/nav-config.ts` — Other (config/util)
- `frontend/src/components/features/navigation/app-shell.tsx` — Component (feature)
- `frontend/src/components/features/navigation/nav-group.tsx` — Component (feature)
- `frontend/src/components/features/navigation/dashboards-nav-group.tsx` — Component (feature)
- `frontend/src/components/features/navigation/mobile-header.tsx` — Component (feature)
- `frontend/src/components/features/navigation/mobile-drawer.tsx` — Component (feature)
- `frontend/src/components/features/workout/workout-shell.tsx` — Component (feature)
- `frontend/src/pages/TrackerPage.tsx` — Page
- `frontend/src/components/features/dashboard-widgets/__tests__/calendar-tomorrow-adapter.test.tsx` — Test
- `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx` — Test

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | grep of all changed files: zero `: any` / `as any` / `<any>`. sidebar.tsx uses `Record<string, unknown>` + typed casts for cloneElement child (sidebar.tsx:968-969). |
| FE-02 | No manual class concatenation | PASS | No `+`/template concat in `className`. `cn()` used throughout (e.g. sidebar.tsx:965, nav-group.tsx). Static class strings assigned to consts (`groupLabelClasses`) are literals, not concatenation. |
| FE-03 | No direct API client calls in components | PASS | No `@/lib/api/client` imports in changed components. calendar-tomorrow-adapter uses `useCalendarEvents` hook (line 4). |
| FE-04 | No inline Zod schemas | PASS | No `z.` usage in scope. |
| FE-05 | No spinners for content loading | PASS | calendar-tomorrow-adapter loading uses `<Skeleton>` (lines 45-54). No `animate-spin` in scope. |
| FE-06 | No hardcoded colors | PASS (1 note) | `bg-black/40` scrim at mobile-drawer.tsx:40 — idiomatic modal overlay, pre-existing (only the adjacent ternary changed), acceptable. Brand hex colors in brand-mark.tsx / favicon.svg are SVG `fill` attributes on a fixed-brand logo, not theme classNames — out of FE-06 scope. `style={{backgroundColor: e.attributes.userColor}}` (adapter:98) is per-event user data. |
| FE-07 | No state mutation | PASS | adapter `.sort()` (line 68) operates on freshly-built arrays returned by `getEventsForDateStr` (new `allDay`/`timed` arrays), not React Query cache or props. nav-group-state uses `new Set(prev)` immutably. |
| FE-08 | No default exports for components | PASS | All exports named (`export function BrandMark`, `Sidebar`, `SidebarProvider`, etc.). Zero `export default` in scope. |
| FE-09 | Tenant guard in hooks | PASS (N/A) | No hook files changed. adapter reads `household?.attributes.timezone` (optional-chained, adapter:41,66) and delegates fetch to existing `useCalendarEvents`. |
| FE-10 | Tenant ID in query keys | PASS (N/A) | No query key factories changed. |
| FE-11 | Error handling | PASS | adapter surfaces fetch failure via `isError` error card (lines 57-63). No new `.catch` in scope. |

## Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | PASS | `Dashboard`/`CalendarEvent` used as `{id, attributes}` (dashboards-nav-group.tsx:53-57, adapter:93-99). No new models. |
| FE-13 | Service extends BaseService | PASS (N/A) | No service files changed. |
| FE-14 | Query key factory `as const` | PASS (N/A) | No key factories changed. |
| FE-15 | Forms use RHF + zodResolver | PASS (N/A) | No forms in scope. |
| FE-16 | Schema in lib/schemas with inferred type | PASS (N/A) | No schemas in scope. |
| FE-19 | Interactive elements show `cursor-pointer` | PASS | Collapsible group triggers carry `cursor-pointer` in `groupLabelClasses` (nav-group.tsx / dashboards-nav-group.tsx). `SidebarTrigger` and `UserMenu` use base `<Button>` (already `cursor-pointer`). `SidebarMenuButton` renders `<a>`/`<button>` (native affordance). Link wrappers in app-shell/adapter are `<a>` elements. |

## Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PARTIAL | Strong: adapter has a new tz-regression test (calendar-tomorrow-adapter.test.tsx:42-59); app-shell test updated. Gap: new `sidebar.tsx` primitive (context/`asChild` cloneElement/mobile toggle) has no direct test, and the drawer open/close now wired through `SidebarProvider`/`useSidebar` is not exercised (mocks stub it out). Non-blocking. |
| FE-18 | Mocks updated when interfaces changed | FAIL (non-blocking) | app-shell.test.tsx:30-38 still mocks `MobileHeader` with `onMenuOpen` and `MobileDrawer` with `open`, but both real components dropped those props (now use `useSidebar()`). Mocks are stale; tests pass only because no test drives the mobile open/close path. |

## Detailed Notes on Requested Focus Areas

- **`React.cloneElement` asChild pattern (sidebar.tsx:957-987):** Sound. Guards with `React.isValidElement(children)`, merges `className` via `cn(classes, childClassName)`, sets `data-active` used by the `data-[active=true]:` cva variants, and typing avoids `any` (`Record<string, unknown>`). One minor limitation: in the `asChild` branch the remaining `...props` are not forwarded to the cloned child — harmless because all call sites pass only `asChild`/`size`/`isActive`/`children`.
- **Sidebar tokens:** all referenced tokens (`sidebar-ring`, `sidebar-accent`, `sidebar-accent-foreground`, `sidebar-foreground`, `bg-sidebar`, `border`) are defined for light and dark in `src/index.css:11-18,76-83,110-117`. No hardcoded theme colors.
- **`isNavItemActive` (nav-config.ts:617-620):** Correct and an improvement — non-`end` items now match `to` or `to/...` but not sibling string-prefixes (fixes `/app/dashboards` vs `/app/dashboards/123`). `NavGroup` uses it consistently for both group-open detection and per-item active state.
- **Tomorrow widget tz fix:** widened UTC query window (adapter:33-35) + client-side `getEventsForDateStr` selection is correct; all-day events matched by inclusive date-string range, timed events by tz-aware window (calendar-utils.ts). Backed by a real regression test for a west-of-UTC household.
- **Accessibility:** `SidebarTrigger` has `sr-only` label plus caller `aria-label` (mobile-header.tsx); mobile drawer retains Escape handling and `aria-hidden` scrim; brand SVGs carry `role="img"` + `aria-label`.

## Summary

### Blocking (must fix)
- None.

### Non-Blocking (should fix)
- **FE-18** — app-shell.test.tsx:30-38: update the `MobileHeader`/`MobileDrawer` mocks; they reference removed props (`onMenuOpen`, `open`) and no longer reflect the real `useSidebar()` interface.
- **FE-17** — add coverage for the mobile drawer open/close now driven by `SidebarProvider`/`useSidebar` (currently untested end-to-end), and optionally a focused test for `SidebarMenuButton`'s `asChild` cloneElement behavior.
- **Objective gate** — run `npm run build` and `npm test` on a glibc/CI install before merge; they could not be executed here due to a musl-vs-glibc `node_modules` platform mismatch. `tsc -b` passes.
