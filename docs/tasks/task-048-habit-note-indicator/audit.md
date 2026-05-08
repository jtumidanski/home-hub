# Audit — task-048-habit-note-indicator

**Branch:** `worktree-task-048-habit-note-indicator`
**PR:** https://github.com/jtumidanski/home-hub/pull/130
**Base SHA:** `a13b78334745c31648051bc4973f9f5eac5e065d`
**Head SHA:** `dd31f6be1f535e700f8b6c9386ac03fbac5f592b`
**Date:** 2026-05-08

## Plan Adherence

**Verdict:** FULL — 12/12 tasks faithfully implemented. **READY_TO_MERGE.**

| # | Task | Status | Evidence |
|---|------|--------|----------|
| 1 | Add `Tooltip` primitive | DONE | `frontend/src/components/ui/tooltip.tsx:1-54` (commit `82b7695`); commit `dd31f6b` removed unused `React` import (build-mode `tsc -b` flags it) |
| 2 | Mount `TooltipProvider` in `App.tsx` | DONE | `frontend/src/App.tsx:8` import; `:61-128` `<TenantProvider>` → `<TooltipProvider delay={200}>` → `<Toaster />` + `<Routes>` (commit `2f56a53`) |
| 3 | Add solid color-token maps | DONE | `frontend/src/components/features/tracker/calendar-grid.tsx:34-54` (commit `2a3ad0d`) |
| 4 | Test scaffold + smoke | DONE | `…/__tests__/calendar-grid.test.tsx:1-122`; flat `MonthItemInfo` fixture + all-weekday schedule (deviation authorized — actual runtime shape, see Deviations); commit `60a4097` |
| 5 | Indicator (border + flange) — TDD | DONE | `calendar-grid.tsx:347` `hasNote`; `:374-382` `triggerClassName`; `:388-396` flange `<span>`; `:402-409` Tooltip wrap; test `calendar-grid.test.tsx:125-144` (commit `c7d9a44`) |
| 6 | Tooltip preview on hover — TDD | DONE | `calendar-grid.test.tsx:146-172` uses `fireEvent.focus` + `findByText` + `whitespace-pre-wrap` (plan-authorized jsdom fallback) (commit `aff61a1`) |
| 7 | `cursor-pointer` test | DONE | `calendar-grid.test.tsx:229-255`; cursor baked into `triggerClassName` at `calendar-grid.tsx:375` (commit `b731b0c`) |
| 8 | `aria-label` test | DONE | `calendar-grid.test.tsx:257-290`; aria-label composition at `calendar-grid.tsx:364-372` (commit `d5db072`) |
| 9 | Negative-path tests | DONE | `calendar-grid.test.tsx:174-226` (no border/flange, no tooltip on no-note, no popover-trigger on future) (commit `e18a440`) |
| 10 | `patterns-styling.md` cursor section | DONE | `.claude/skills/frontend-dev-guidelines/resources/patterns-styling.md:230-236` (commit `2f88226`) |
| 11 | FE-19 Styling Checklist row | DONE | `.claude/agents/frontend-guidelines-reviewer.md:102-106` (commit `c23ddd5`) |
| 12 | Final verification | DONE | See Build & Test Results below |

### Build & Test Results

| Surface | Result |
|---------|--------|
| `npm test -- --run` | **PASS** — 92 files, 606 tests, 0 fail |
| `npx tsc --noEmit` | **PASS** — clean |
| `npm run build` (`tsc -b && vite build`) | **PASS** — built in ~750ms |
| `npm run lint` | **PASS-with-pre-existing** — 11 errors all in untouched files (`use-cooklang-preview.ts:35`, `DashboardDesigner.tsx:56-57`, `WorkoutReviewPage.test.tsx:55-56`); zero diff vs base |
| Manual browser smoke (PRD §10) | **N/A** — operator-only, not in audit scope |

### Deviations from Plan Literal

All four deviations are pre-authorized by the plan's own fallback notes or are mechanical fixes:

1. **`MonthItemInfo` fixture is flat (no nested `attributes`)** — `calendar-grid.test.tsx:71-87`. Matches the actual runtime shape consumed at `calendar-grid.tsx:105` (`relationships.items.data` is read as flat `MonthItemInfo`).
2. **Schedule set to `[0,1,2,3,4,5,6]` instead of plan literal `[]`** — behavioral parity preserved; explicit array clarifies intent.
3. **Locator uses `aria-label?.startsWith("Run, May N")` instead of `data-slot="popover-trigger"`** — plan §5 Step 3 explicitly authorizes a fallback locator. Necessary because Base UI's `<TooltipTrigger render={...} />` rewrites `data-slot` to `tooltip-trigger` on note-bearing cells.
4. **`fireEvent.focus` instead of `pointerEnter`/`mouseEnter`** — plan §6 explicitly authorizes this jsdom fallback.
5. **`React` import removed from `tooltip.tsx`** (commit `dd31f6b`) — `tsc -b` (build mode used by `npm run build`) flags it as unused; behavior unchanged.

None materially deviate from intent.

## Frontend Guidelines (FE-* Checklist)

**Verdict:** PASS — zero blocking, zero non-blocking issues.

### Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | `grep -nE ': any\|as any'` zero matches across changed files. `as unknown as` casts at `calendar-grid.tsx:110, 194, 286` are pre-existing; test fixtures use them only to narrow `MonthSummaryResponse`/`TrackerEntry` shapes. |
| FE-02 | No manual class concatenation | PASS | All conditional classes via `cn(...)`: `calendar-grid.tsx:374-382` (`triggerClassName`), `:389-395` (flange), `tooltip.tsx:41-44`. |
| FE-03 | No direct API client calls in components | PASS | Data flows through hooks at `calendar-grid.tsx:9-10`. |
| FE-04 | No inline Zod schemas | PASS | No new forms in this PR. |
| FE-05 | No spinners for content loading | PASS | Loading uses `<Skeleton />` at `calendar-grid.tsx:133`. |
| FE-06 | No hardcoded colors (semantic only) | PASS-with-context | New `colorBorderSolid` (`:34-43`) and `colorBgSolid` (`:45-54`) maps use raw Tailwind palette names by design — these index a per-habit user-configurable color, not a theme color. The pre-existing `colorBg` (`:16-25`) uses the identical pattern. Semantic-token fallbacks at `:381` (`?? "border-foreground/30"`) and `:393` (`?? "bg-foreground/30"`) correctly handle unknown colors. |
| FE-07 | No state mutation | PASS | `[...items].sort(...)` at `calendar-grid.tsx:135` is copy-then-sort. |
| FE-08 | No default exports for components | PASS | All named exports (`tooltip.tsx:54`, `calendar-grid.tsx:95`). |
| FE-09 | Tenant guard in hooks | N/A | No new hooks added. |
| FE-10 | Tenant ID in query keys | N/A | No new query keys. |
| FE-11 | Error handling with `createErrorFromUnknown` | N/A | No new async surfaces. |
| **FE-19** | **`cursor-pointer` on interactive elements** | **PASS** | This PR introduces the rule. `cursor-pointer` baked into `triggerClassName` at `calendar-grid.tsx:375`, applied to both Tooltip-wrapped (`:404`) and bare (`:408`) trigger paths. Future-cell `<span>`s at `:340-342` are non-interactive — correctly omitted. Asserted in test at `calendar-grid.test.tsx:229-255`. |

### Architecture Checklist

| ID | Check | Status |
|----|-------|--------|
| FE-12 | JSON:API model shape | N/A — no new models |
| FE-13 | Service extends `BaseService` | N/A — no service changes |
| FE-14 | Query key factory uses `as const` | N/A — no new query keys |
| FE-15 | Forms use `react-hook-form` + `zodResolver` | N/A — no forms |
| FE-16 | Schema in `lib/schemas/` with inferred type | N/A — no schemas |

### Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | `calendar-grid.test.tsx` adds 8 specs across `note indicator`, `cursor affordance`, `aria-label` describe blocks. |
| FE-18 | Mocks updated when services changed | PASS | Module-boundary mocks at `calendar-grid.test.tsx:9-31`. No service changes in this PR. |

### Tooltip Primitive Checks

| Check | Status | Evidence |
|-------|--------|----------|
| Mirrors existing `popover.tsx` shape | PASS | `tooltip.tsx` named exports, `data-slot="…"` attributes, `Popup`/`Positioner`/`Portal` composition. |
| Uses `cn()` for className merging | PASS | `tooltip.tsx:41-44`. |
| Provider mounted at root | PASS | `App.tsx:62` wraps `<Routes>` in `<TooltipProvider delay={200}>`. |
| No semantic-token violations | PASS | `bg-popover text-popover-foreground ring-foreground/10` at `tooltip.tsx:42`. |

## Summary

### Blocking (must fix)

None.

### Non-Blocking (should fix)

None.

### Overall

**READY_TO_MERGE.** The implementation is faithful to the plan, exercises the new FE-19 rule it introduces, ships a clean Tooltip primitive mirroring the existing popover wrapper, and passes every objective gate (606 tests, tsc, build). The only outstanding step is operator-only manual browser smoke (PRD §10) which is documented in the PR test plan.
