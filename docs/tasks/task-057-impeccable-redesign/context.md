# Task 057 — Context

Read with `plan.md` (what to do) and `design.md` (why). This file is the
orientation map: where things are, what was decided, and what not to touch.

## Artifacts in this folder

| File | Role |
|---|---|
| `prd.md` | Requirements (FR-1…FR-22, NFRs, acceptance criteria) |
| `design.md` | Approved direction "The Kitchen Table": tokens, type, shape, motion, architecture, stages |
| `comps/dashboard.html` | Approved rendered comp; the finish reviewer checks against it |
| `comps/brand-mark-options.html` | Brand mark options; `#mB` "Window light" was chosen |
| `plan.md` | 33 tasks across Stages 0–7 |
| `design-audit.md` | Created in Task 1; Impeccable audit findings `DA-NNN` plus baselines and the a11y log |
| `audit.md` | Code-review findings (reviewer agents), distinct from `design-audit.md` |

## Key files (frontend/)

| Path | Why it matters |
|---|---|
| `src/index.css` | Only source of tokens: `@theme inline`, `:root`, `.dark`, base layer. Today: all-grey OKLCH, Geist, `--radius: 0.625rem`, `outline-ring/50` focus |
| `src/components/ui/*` | 16 shadcn/base-ui primitives (badge, button, calendar, card, dialog, form, input, label, popover, radio-group, select, sidebar, skeleton, textarea, tooltip, user-avatar). Focus is `focus-visible:ring-3 ring-ring/50` today |
| `src/components/common/*` | brand-mark, card-action-menu, dashboard-skeleton, data-table, error-card, error-page (exports `Error404Page`), list-filter-bar, owner-select, pull-to-refresh, weather-icon. New: page-header, empty-state |
| `src/components/features/navigation/*` | The shell: app-shell, mobile-header, mobile-drawer, nav-group, dashboards-nav-group, user-menu |
| `src/App.tsx` | Routes; `<Toaster richColors closeButton />` from sonner |
| `src/components/providers/theme-provider.tsx` | Do not modify (FR-8) |
| `src/lib/dashboard/**` | Do not modify at all; its tests must pass unmodified |
| `src/types/models/tracker.ts` | `TrackerColor`: 16 Tailwind hue names, persisted by the backend. Names stay; visuals move to `--swatch-*` |
| `src/types/models/calendar.ts` | `userColor: string` (hex from API) is user data, a documented exception |
| `public/favicon.svg` | Must stay in sync with `brand-mark.tsx` |
| `vite.config.ts` | vitest: jsdom, globals, setup `src/test/setup.ts`, include `src/**/*.test.{ts,tsx}` |
| `.impeccable/config.json` (repo root, **untracked today**) | Detector ignores; Task 1 moves it and commits it |

## Literal hotspots at task start (palette classes + colour literals, excl. tests)

tracker ≈139, calendar 26, pages 15 (IngredientsPage, Recipe*, Workout*, designer `bg-black/30`), recipes 14, packages 7, common 5 (brand-mark SVG), ui 3 (`bg-black/30` dialog, `text-white` avatar), meals 2, workouts 2, ingredients/navigation/trackers 1 each; dashboards, dashboard-widgets, households, reminders, tasks, weather, workout 0. Specific known literals: `#9ca3af` busy (month-day-cell, event-block, all-day-event-row, month-event-chip), `#0d9488` package overlay (`features/packages/package-calendar-overlay.tsx`), four duplicated `Record<TrackerColor,string>` maps in tracker (create/edit dialogs, today-view, tracker-setup).

Tests that assert colour classes today: `features/tracker/__tests__/calendar-grid.test.tsx` (`bg-blue-500`), `features/packages/__tests__/package-calendar-overlay.test.ts` (`#0d9488`).

## Decisions made in the plan (beyond design.md)

- **PR mechanics:** PR 1 = Stage 0 + 1 on `task-057-impeccable-redesign`. Stages 2–7 each on `task-057-impeccable-redesign-stage-N`, cut from `origin/main` after the previous PR merges, in the same worktree. Strictly sequential (shared `PENDING` list and `design-audit.md`).
- **New tokens not named in design.md:** `--warning-text` (ochre legible as text/icon; light ochre fill fails 3:1 as an icon), `--overlay` (dialog/drawer backdrop, replaces `bg-black/30`), `--calendar-busy-foreground`, `--calendar-package-foreground`, `--on-color-ink`/`--on-color-paper` (theme-invariant, equal to `INK`/`PAPER` in `color-contrast.ts`), `--elevation-card`/`--elevation-overlay` (raw values behind `shadow-card`/`shadow-overlay`, named to avoid a circular `@theme` mapping).
- **Calendar colour:** `eventChipColors()` in `src/lib/calendar/event-colors.ts`; package events use the sentinel `PACKAGE_EVENT_COLOR = "var(--calendar-package)"` as their `userColor`.
- **Literal guard scope:** `components/features/**` and `pages/**` only; `PENDING` is a list of dir/file prefixes relative to `src/`, emptied stage by stage; Stage 7 asserts it is empty.
- **Recipe step segments:** ingredient → `text-primary`, cookware → `text-info`, timer → `text-success` (text-safe tokens, not chart tokens).
- **Baseline ref:** `9e00203` (main at task start; the branch has no frontend changes before Stage 1).
- **Impeccable root:** decided empirically in Task 1 (`frontend/` preferred) and recorded as the first line of `design-audit.md` `## Method`.

## Dependencies

- Add: `@fontsource-variable/fraunces`, `@fontsource-variable/figtree`. Remove: `@fontsource-variable/geist` (kept only if Figtree lacks `tnum`). Nothing else.
- Tools: Impeccable plugin (`/impeccable:impeccable` `init` / `audit`; agents `impeccable-finish-reviewer`, `impeccable-documenter`). `tea` for PRs. `scripts/local-up.sh` for rendered checks. `tools/verify.sh` (flagless) at each stage close-out, via `task-verifier`.

## Risks to watch

- Rendered checks need a browser tool. If none is available, record "not rendered" and hand those checks to the user in the PR body. Never claim a rendered check that did not happen.
- Font budget: Fraunces variable may exceed its share; fallback is static `@fontsource/fraunces` 600.
- Stage 1 restyles every page at once while feature pages still carry palette classes. That interim inconsistency is accepted (design.md §6); stages should merge in quick succession.
- Container-query widget tests must mock data hooks the same way the existing `dashboard-widgets/__tests__` do.
