# Plan Audit — task-049-user-menu-settings-move

**Plan Path:** docs/tasks/task-049-user-menu-settings-move/plan.md
**Audit Date:** 2026-05-08
**Branch:** task-049-user-menu-settings-move
**Base Branch:** main
**Diff Range:** 38e6c39..f8383d8

## Executive Summary

All 6 tasks (and every sub-step) in the plan were faithfully implemented. The five expected frontend files changed exactly as the plan specifies; `nav-config.ts`, `use-theme-toggle.ts`, and the Settings page were correctly left untouched. `npm test` passes (610/610 across 93 files) and `npm run build` succeeds with no unused-import warnings. The branch is ready for code review and merge.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|------------------|
| 1 | Add `user-menu.test.tsx` covering new pop-over contract (failing first) | DONE | Commit `e29fc86`; file created at `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx` (96 lines), content matches plan §1 byte-for-byte (verified lines 1-96 vs plan lines 33-130). All five required cases present: null user (line 49), order check (line 55), no theme toggle (line 65), Settings navigates + onAction (line 74), Sign Out + onAction (line 86). |
| 2 | Update `user-menu.tsx` to make new tests pass | DONE | Commit `2360992`; `user-menu.tsx:1-77` matches the plan's §2 replacement exactly. Imports added: `Settings` (lucide-react, line 2), `useNavigate` (line 3), `settingsNavItem` (line 6). Imports removed: `Moon`, `Sun`, `useThemeToggle` (verified absent). Settings item is first (lines 52-61) with neutral styling and `navigate(settingsNavItem.to); onAction?.()` ordering. Sign Out item (lines 62-71) is preserved verbatim with destructive styling. |
| 3 | Update `app-shell.test.tsx` to drop obsolete assertions | DONE | Commit `b67028a`; `app-shell.test.tsx` Change A: `mockToggleTheme` declaration and `useThemeToggle` mock block both removed. Change B: `expect(screen.getByText("Settings")).toBeInTheDocument();` no longer present in `"renders sidebar with navigation links"` (line 66). Change C: `"calls toggleTheme when theme button is clicked in user menu"` test fully removed; the sign-out test (lines 91-99) remains intact. |
| 4 | Remove Settings footer block from `app-shell.tsx` | DONE | Commit `c5595c0`; `app-shell.tsx:53-55` now contains only `<div className="border-t"><UserMenu /></div>`. The bordered NavLink Settings block is gone. Imports trimmed: `NavLink` removed (line 2 now `Outlet, Link`); `Settings` icon import line deleted; `settingsNavItem` removed from nav-config import (line 10). Plan permitted dropping `cn` if no remaining usage; `grep "cn(" app-shell.tsx` returns zero hits and `cn` was correctly dropped from imports. |
| 5 | Remove Settings footer block from `mobile-drawer.tsx` | DONE | Commit `f8383d8`; `mobile-drawer.tsx:95-97` now contains only `<div className="border-t"><UserMenu onAction={onClose} iconSize="h-5 w-5" /></div>`. The bordered NavLink Settings block is gone. Imports trimmed: `NavLink` import line deleted entirely; `Settings` removed from lucide-react import (line 2: `import { X } from "lucide-react"`); `settingsNavItem` removed from nav-config import (line 8). `cn` correctly retained — used at lines 40 and 55. |
| 6 | Final verification (build + tests) | DONE | Re-run during this audit: `npm test` reports 610/610 tests passing across 93 files (11.55s); `npm run build` succeeds cleanly with no unused-import warnings or TS errors. Diff smoke check: 5 frontend files changed (matches plan §6.3 expected list); `nav-config.ts`, `use-theme-toggle.ts`, and `pages/settings/` are absent from diff as required. |

**Completion Rate:** 6/6 tasks (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

None. Every task and sub-step in the plan has corresponding evidence in the diff and source.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| frontend | PASS | PASS | `npm run build` (tsc -b && vite build) succeeded; 1.2 MB main bundle warning is pre-existing and unrelated to this task. `npm test` passes 610/610 tests across 93 files in 11.55s. |

## PRD Acceptance Criteria Coverage (prd.md §10)

| AC | Status | Evidence |
|----|--------|----------|
| Sidebar footer no longer contains "Settings" link | MET | `app-shell.tsx:53-55` shows only `UserMenu`; bordered NavLink block removed in `c5595c0`. |
| Drawer footer no longer contains "Settings" link | MET | `mobile-drawer.tsx:95-97` shows only `UserMenu`; bordered NavLink block removed in `f8383d8`. |
| Pop-over contains, in order, Settings then Sign Out (and nothing else) | MET | `user-menu.tsx:52-71`; tested by `user-menu.test.tsx:55-63` (`renders only Settings and Sign Out items… in that order`). |
| Activating Settings navigates to `/app/settings` without full reload | MET | `user-menu.tsx:54-56` calls `navigate(settingsNavItem.to)` (= `/app/settings`); tested by `user-menu.test.tsx:74-84`. |
| Activating Settings inside drawer also closes drawer | MET | `user-menu.tsx:54-56` calls `onAction?.()` after navigate; `mobile-drawer.tsx:96` wires `onAction={onClose}`. |
| Activating Sign Out still triggers logout | MET | `user-menu.tsx:62-71` preserved verbatim; tested by both `user-menu.test.tsx:86-95` and `app-shell.test.tsx:91-99`. |
| Theme toggle absent from pop-over | MET | `user-menu.tsx` has no `useThemeToggle`/`Moon`/`Sun` imports or items; tested by `user-menu.test.tsx:65-72`. |
| Theme toggle on Settings page still works | MET | `frontend/src/lib/hooks/use-theme-toggle.ts` and `frontend/src/pages/settings/` are absent from the diff. |
| Pop-over fully operable via keyboard / screen reader | MET | `MenuPrimitive.Item` (base-ui) provides `menuitem` role and arrow-key navigation; `user-menu.test.tsx` queries by `role="menuitem"`, confirming roles are exposed. |
| `frontend` builds with no unused-import warnings introduced | MET | `npm run build` clean during audit re-run; `cn` correctly dropped from `app-shell.tsx`, retained in `mobile-drawer.tsx`. |
| Existing nav tests updated; pop-over coverage exists | MET | `app-shell.test.tsx` updated (commit `b67028a`); new `user-menu.test.tsx` added (commit `e29fc86`). |
| `npm test` passes in frontend | MET | 610/610 tests pass during audit re-run. |
| No backend services modified | MET | `git diff --name-only main...HEAD` shows only `docs/tasks/...` and `frontend/src/components/features/navigation/` files. |

All 13 PRD acceptance criteria are met.

## Design Coverage Map (design.md §8)

Every row in the design's acceptance-criteria mapping table has implementation evidence:

| Design AC | Implementation evidence |
|-----------|------------------------|
| Sidebar footer simplification | `app-shell.tsx:53-55` (commit `c5595c0`) |
| Drawer footer simplification | `mobile-drawer.tsx:95-97` (commit `f8383d8`) |
| Item ordering Settings → Sign Out | `user-menu.tsx:52-71` (commit `2360992`); test `user-menu.test.tsx:55-63` |
| `useNavigate` choice for Settings activation | `user-menu.tsx:3,18,55` |
| `onAction?.()` semantics in drawer | `user-menu.tsx:56,66` (called after primary effect); `mobile-drawer.tsx:96` |
| Sign Out path unchanged | `user-menu.tsx:62-71` (byte-for-byte preserved) |
| Theme toggle removed from pop-over | `user-menu.tsx` (no `Moon`/`Sun`/`useThemeToggle` imports); tested `user-menu.test.tsx:65-72` |
| Theme toggle on Settings page untouched | `use-theme-toggle.ts` and `pages/settings/` absent from diff |
| Keyboard / screen reader operability | base-ui `MenuPrimitive.Item` (lines 52, 62) provides standard menu roles |
| Imports cleanup | `app-shell.tsx:1-12`, `mobile-drawer.tsx:1-11`, `user-menu.tsx:1-8` all match plan-specified import lists |
| Nav tests updated; pop-over coverage exists | `__tests__/app-shell.test.tsx` modified; `__tests__/user-menu.test.tsx` added |
| `npm test` green | 610/610 pass |
| No backend changes | Diff scope limited to `frontend/src/components/features/navigation/` |

All 13 design coverage rows verified.

## Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

Notes for review:
- Plan §6.3 stated the expected diff would have 7 entries (4 doc files + 3 component files + 1 test edit + 1 new test). Actual diff has 9 entries (4 docs + 5 frontend files): the test for `app-shell` shows up because the plan correctly modified it in Task 3, and the new `user-menu.test.tsx` shows up too. This matches the plan's intent; the §6.3 list is a slight undercount but the intent (don't touch `nav-config.ts`, don't touch `use-theme-toggle.ts`, don't touch `pages/settings/`) is fully respected.
- Commit ordering follows the plan exactly: tests-first (Task 1, 3) precede their implementation counterparts (Task 2, 4), matching TDD.

## Action Items

None. The plan was implemented faithfully, builds and tests are green, and all PRD/design acceptance criteria are satisfied.

---

# Frontend Guidelines Audit — task-049-user-menu-settings-move

- **Audit Scope:** TypeScript/React files changed in `38e6c39..f8383d8` (navigation feature only)
- **Guidelines Source:** `.claude/skills/frontend-dev-guidelines`
- **Date:** 2026-05-08
- **Build:** PASS (per plan-adherence reviewer above; auditor environment could not re-run — see note)
- **Tests:** PASS — 610/610 (per plan-adherence reviewer above)
- **Overall:** PASS

## Build & Test Note

The plan-adherence reviewer (section above) independently confirmed `npm run build` and `npm test` (610/610 tests across 93 files) pass on this branch. This guidelines audit could not independently re-run those commands because the auditor's WSL shell exposes only Windows `node` (`/mnt/c/Program Files/nodejs/`), but `node_modules` was installed for Linux x64 — the rolldown native binding mismatch (`@rolldown/binding-win32-x64-msvc` missing) prevents direct execution from this shell. The build/test gate is therefore taken from the section above; if those numbers change on a subsequent re-run, this audit's PASS status is contingent on them remaining green.

## File Inventory

- `frontend/src/components/features/navigation/user-menu.tsx` — Component (modified)
- `frontend/src/components/features/navigation/app-shell.tsx` — Component (modified, delete-only)
- `frontend/src/components/features/navigation/mobile-drawer.tsx` — Component (modified, delete-only)
- `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx` — Test (new)
- `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx` — Test (modified)

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | `grep -nE ': any\|as any'` over all five files: zero matches. |
| FE-02 | No manual class concatenation | PASS | All `className` values use `cn(...)` helper. `user-menu.tsx:25-27,45-50`; `mobile-drawer.tsx:40-43,55-58`. No `+` operator or template-string concatenation in any `className` prop. |
| FE-03 | No direct API client calls in components | PASS | No imports from `@/lib/api/client` in any changed file. Auth flows go through the `useLogout()` hook (`user-menu.tsx:5,17`). |
| FE-04 | No inline Zod schemas | PASS | No `z.object(` / `z.string(` / etc. in any changed file. |
| FE-05 | No spinners for content loading | PASS | No `animate-spin` in any changed file. |
| FE-06 | No hardcoded colors | PASS (in scope) | The only raw-color match is `bg-black/40` at `mobile-drawer.tsx:41`, but `git show 38e6c39:.../mobile-drawer.tsx` line 42 confirms it pre-existed and is unchanged by this task — outside the diff hunks and out of scope for a task-049 audit. |
| FE-07 | No state mutation | PASS | No `.push(`, `.splice(`, or `.sort(` followed by `setState` in changed files. State use is read-only `useState` (`app-shell.tsx:15`) and `useRef` (`mobile-drawer.tsx:19`); test files use mock factories only. |
| FE-08 | No default exports for components | PASS | All components are named exports: `export function UserMenu` (`user-menu.tsx:15`), `export function AppShell` (`app-shell.tsx:14`), `export function MobileDrawer` (`mobile-drawer.tsx:18`). |
| FE-09 | Tenant guard in hooks | N/A | No hooks added/modified. `UserMenu` consumes `useAuth` and `useLogout`, neither of which is a tenant-scoped resource hook. |
| FE-10 | Tenant ID in query keys | N/A | No query key factories added/modified. |
| FE-11 | Error handling with `createErrorFromUnknown` | N/A | No new `.catch(` async paths in changed files. `logout.mutate()` (`user-menu.tsx:65`) delegates error handling to the `useLogout` hook. |

## Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | N/A | No model definitions added; the `user.attributes.{displayName,email,...}` access pattern (`user-menu.tsx:30-38`) correctly consumes the existing `{ id, attributes }` shape. |
| FE-13 | Service extends `BaseService` | N/A | No services touched. |
| FE-14 | Query key factory uses `as const` | N/A | No query key factories touched. |
| FE-15 | Forms use `react-hook-form` + `zodResolver` | N/A | No forms in scope. |
| FE-16 | Schema in `lib/schemas/` with inferred type | N/A | No schemas in scope. |

## Styling Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-19 | Interactive elements show `cursor-pointer` | PASS | The new clickable surfaces in `user-menu.tsx` all carry `cursor-pointer`: trigger at `user-menu.tsx:26`, Settings menuitem at `user-menu.tsx:53`, Sign Out menuitem at `user-menu.tsx:63`. The `MobileDrawer` overlay (`mobile-drawer.tsx:39-46`) is an `aria-hidden="true"` dismiss surface (pre-existing, unchanged by task-049). The drawer panel itself (`mobile-drawer.tsx:49-58`) is the dialog container, not a click target. The close button (`mobile-drawer.tsx:63-65`) is a native `<Button>` and inherits the browser pointer. No new clickable non-`<button>`/non-`<a>` elements were introduced. |

## Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | New `user-menu.test.tsx` covers the five contract behaviours of the modified `UserMenu`: render-nothing without user (`__tests__/user-menu.test.tsx:49-53`), pop-over item set & order (`:55-63`), absence of theme toggle (`:65-72`), Settings navigation + `onAction` (`:74-84`), Sign Out logout + `onAction` (`:86-95`). `app-shell.test.tsx` was correctly pruned: dropped assertion for the dismissed sidebar `Settings` link and the obsolete `toggleTheme`-on-menu-click test; sidebar render + sign-out tests are preserved (`__tests__/app-shell.test.tsx:66-73,91-99`). `MobileDrawer` had only delete-only changes (Settings footer link removed, no behavioural change to other paths) — no new dedicated test required. |
| FE-18 | Mocks updated when services changed | PASS | `vi.mock("@/lib/hooks/use-theme-toggle", ...)` and the `mockToggleTheme` declaration were correctly removed from `app-shell.test.tsx` because `AppShell` no longer imports `useThemeToggle`. Confirmed by `grep -rn "useThemeToggle\|use-theme-toggle" frontend/src/components/features/navigation/`: zero matches. The new `user-menu.test.tsx` registers correct mocks for `useAuth` (`:10-12`), `useLogout` (`:14-16`), and `useNavigate` (`:18-21`) — exactly the three external dependencies imported by `user-menu.tsx:4-6`. |

## Frontend Audit Summary

### Blocking (must fix)
- None.

### Non-Blocking (should fix)
- None within the FE-* checklist.

### Notes
- All FE-* checks that apply to this scope pass with cited evidence. FE-09–FE-16 are correctly N/A: this task is a navigation refactor with no API, hook, schema, form, or model changes.
- Pre-existing `bg-black/40` at `mobile-drawer.tsx:41` is noted for completeness but is outside the diff and not introduced by task-049; flagging it would be manufacturing a finding.

