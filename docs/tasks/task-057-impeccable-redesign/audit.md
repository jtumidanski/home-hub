# Task 057 — Code review audit

Stage 0+1 close-out (PR 1). Two reviewers ran in parallel over `9e00203..ba26f21`; their reports follow verbatim. Blocking items (DA-021 radio, select trigger disabled opacity) were fixed in 3d5c079, 5f140d7 and 7b27b7e.

## Plan Adherence Audit — task-057 Stages 0+1 (Tasks 1–14)

**Plan Path:** docs/tasks/task-057-impeccable-redesign/plan.md
**Audit Date:** 2026-10-19
**Branch:** task-057-impeccable-redesign
**Base:** 9e00203 (merge base with origin/main)
**Scope:** Global Constraints, Stage 0 (Tasks 1–3), Stage 1 (Tasks 4–14). Stages 2–7 out of scope.

### Executive Summary

All 14 tasks are implemented. Tasks 1, 3, 4–13 are DONE. Task 2 and Task 14 are DONE with deviations that are ruled in the ledger. The plan's checkbox steps are not ticked in plan.md, so completion was judged from commits and files. Frontend `npx tsc -b` is clean and `npx vitest run` passes (132 files, 856 tests). No Global Constraint violation was found. Open items for the controller are the DA-021 radio border (below) and the close-out step 4 "Resolved" column. Step 4 is still empty for all 40 Stage 1 rows.

### Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|------------------|
| 1 | Impeccable root + PRODUCT.md | DONE | `frontend/PRODUCT.md`, `frontend/.impeccable/config.json` (the `nested-cards` entry for `.superpowers/brainstorm` is removed; the Fraunces `overused-font` and `comps/dashboard.html` `nested-cards` entries are kept). `design-audit.md` has the required headings in order, and "Impeccable root: frontend/" is the first Method line. Init run by the controller with user-confirmed answers (R1). Explicit `git add` instead of `-A` (R3). |
| 2 | Baselines | DONE (with ruling) | `design-audit.md` "## Baselines": ref 9e00203, JS gzip 371767 B, Geist latin 29400 B, commands recorded. Step 1 check excludes the Task 1 artifacts (R2). |
| 3 | Whole-app design audit | DONE | 141 rows (Stage 1: 40, 2: 7, 3: 25, 4: 15, 5: 13, 6: 20, 7: 21), all with a Stage. The page-header, empty-state and status-banner decisions are in Method. `audit-shards/` is deleted. All shards were source-only and this is recorded in Method. R5 keeps DA-041 in Stage 2. |
| 4 | Colour contrast helper | DONE | `frontend/src/lib/color-contrast.ts` and its test (commit 12f6e51). The test passes. |
| 5 | Colour tokens + token test | DONE | `frontend/src/index.css`, `src/__tests__/{design-tokens.test.ts,css-tokens.ts}` (bf95a06). The test passes. |
| 6 | Typography and fonts | DONE | package.json swaps geist for fraunces and figtree (the only dependency change). The commit records that `tnum` is present and geist removed. Font payload 56,776 B (budget 150 KB) is recorded in Baselines. The fallback `size-adjust` values come from recalled metrics (ledger minor). |
| 7 | Base layer | DONE | `index.css:35` coarse variant, `:176` `--radius: 0.75rem`, `:177-178` durations, `:131` `--ease-out`, `:314` `:focus-visible`, `:330` reduced-motion block. The base-layer tests pass. |
| 8 | Swatch module | DONE | `src/lib/constants/swatches.ts`, `swatches.test.ts`, and the token test imports it (0a95fa7). |
| 9 | No-colour-literal guard | DONE | `no-color-literals.test.ts` has PENDING and ALLOWED (empty), and the four required cases. PENDING shrank only where files became clean (Task 14 batches). It still contains calendar/ and the other Stage 2–7 directories, as designed. The red-then-restore Step 2 check is not independently verifiable. |
| 10 | Control primitives | DONE | `components/ui/{button,badge,input,select,textarea,radio-group,label}`, `button.test.tsx`, `badge.test.tsx`. `form.tsx` is unchanged (ledger observation). `ui/calendar.tsx` was added under R7. |
| 11 | Surface primitives | DONE | a486fc5, with `card.test.tsx`, `dialog.test.tsx` and updated sidebar and avatar tests. Skeleton was left unchanged and only later gained `motion-reduce:animate-none` (R15). `select.tsx` popup shadow and motion were not moved onto the overlay tokens; no commit shows them being addressed. |
| 12 | PageHeader, EmptyState, common components | DONE | `common/page-header.tsx`, `empty-state.tsx` and their tests (dae98bd). `status-banner` was added as the extra primitive that Task 3 raised, with `status-banner.test.tsx`. R8 splits some rows with Task 14. |
| 13 | Brand mark + favicon | DONE | 5937bf3, `brand-mark.test.tsx`, `public/favicon.svg`. Uses oklch fills, taken verbatim from the comp (ledger minor). R9 covers the commit trailer. |
| 14 | Resolve remaining Stage 1 findings | DONE (with rulings) | Ran as batches A1/B/C1/C2/D under R10–R17. The triage map covers DA-001..040 (the ledger's batch map plus the triage-resolved rows plus DA-023). The detector (Step 2) ran clean and no ignore entry was needed. Commits name the DA IDs. |

**Completion Rate:** 14/14 tasks (100%) implemented
**Skipped without approval:** 0
**Partial implementations:** 0 (see the residual items below)

### Global Constraints check

| Constraint | Result |
|---|---|
| No change under `services/`, `shared/` | PASS. `git diff --name-only` is empty for those paths. |
| `theme-provider.tsx` and `src/lib/dashboard/**` untouched | PASS. The diff is empty. |
| Dependencies | PASS. package.json changes only geist to figtree and fraunces, and no devDependencies are added. |
| Token light and dark parity | PASS via the design-tokens test. |
| Literal guard on migrated dirs | PASS. ALLOWED is empty, PENDING only shrank, and the guard passes. |
| Budgets | The font payload is within budget. JS gzip versus baseline ×1.10 is not re-measured here, because no build is in scope. It is a Stage 7 comparison. |
| No absolute home paths in docs | PASS for this branch's docs additions (0 added lines). Older docs from other tasks contain such paths, but they are not part of this diff. |
| Commit trailers | The trailers read Opus 5.5 (15), Sonnet 5.5 (25) and Haiku 5.5 (1) against the plan's "Opus 5.5". They accurately name the authoring model. R9 covers the Task 13 trailer only, so the controller should extend the PR-body note to all of them. |
| Status is never colour-only; targets; motion | Not checkable without rendering. Covered by tests where testable. |

### Residual items and risks (not blockers for plan adherence)

1. **Close-out step 4 not done yet.** The `Resolved` column is empty for all 40 Stage 1 rows. The plan requires a separate `docs(task-057): mark stage 1 findings resolved` commit. DA-022 (DiceBear third-party request) is deliberately report-only under R16. Mark it as "partial" or "reported", not as a plain fix, and surface it in the PR body.
2. **DA-021 radio group.** `radio-group.tsx:55` sets `border-input` and `text-primary` on a native radio. The ledger notes these are inert without a forms plugin, and only `size-5`, `accent-primary` and the focus outline take effect. The fix is partial in substance. The audit row also asked for a `min-h-11` row on mobile, and I did not verify whether that was applied.
3. **Rendered check.** None happened; 10 "not rendered — browser unavailable" rows were logged correctly. Manual checks are owed to the PR body (step 1 / step 5).
4. **Deferred minors from the ledger** are still open, as noted there: the `select.tsx` popup shadow and motion (T11), the ErrorPage h1 (T12), `calendar/` still in PENDING (T14D), event-block opacity (Task 22), and the favicon oklch fallback. No plan task claims these, so they are not adherence failures.
5. **Plan checkboxes** in plan.md remain `- [ ]`, so the plan file does not reflect completion.

### Build & Test Results

| Area | Build / typecheck | Tests | Notes |
|---|---|---|---|
| frontend | `npx tsc -b` PASS | `npx vitest run` PASS (132 files, 856 tests) | tools/verify.sh and docker deliberately not run (running separately) |

### Overall Assessment

- **Plan Adherence:** FULL (all deviations are ruled in the ledger)
- **Recommendation:** READY_TO_MERGE after the action items below

### Action Items

1. Run the close-out step 4 commit that fills the `Resolved` column for Stage 1 rows. Use the SHAs from the ledger's batch map, and mark DA-022 as partial under R16.
2. Decide DA-021. Either accept `size-5` and `accent-primary` as meeting "token control", or drop the inert `border-input` and `text-primary` classes and the test assertion that pins them.
3. Put these in the PR body: the R16 DiceBear request, the commit-trailer model names (not only R9), the 10 routes needing a manual rendered check, and the deferred minors that matter.

## Frontend Audit - task-057 Stage 0+1 close-out

- **Audit Scope:** `git diff 9e00203..HEAD -- frontend/` (62 non-test TS/TSX files changed or added, plus `frontend/src/index.css`, 38 test files)
- **Guidelines Source:** frontend-dev-guidelines skill (FE-* checklist); plan.md Global Constraints
- **Date:** 2026-10-19
- **Build:** PASS (`npm run build`, vite "built in 3.49s", chunk-size warning only)
- **Tests:** 856 passed, 0 failed (132 files; `npx vitest run`; the `--watchAll` flag is rejected by vitest 4, so the plain run was used)
- **Lint:** `npm run lint`: 0 errors, 5 warnings (event-form-dialog.tsx x4, new-dashboard-modal.tsx x1, `react-hooks/incompatible-library`)
- **Overall:** NEEDS-WORK (no build or test failures; FAIL checks below, one of which is introduced by this branch's work: DA-021 and the FE-17 gaps)

### Global Constraints (plan.md:13)

| Constraint | Status | Evidence |
|---|---|---|
| Nothing under `services/`, `shared/`, `lib/dashboard/**`, `theme-provider.tsx` modified | PASS | `git diff 9e00203..HEAD --stat` on those paths is empty |
| No new runtime deps beyond fraunces/figtree; geist removed | PASS | `frontend/package.json` diff: -geist, +figtree, +fraunces only |
| No JSON:API / hook / service / schema / type change | PASS | none of the 62 files is under `hooks/api`, `services/api`, `lib/schemas`, `types/models` |

### Judgment 1: DA-021 (radio `border-input`)

**Verdict: `border-input` is inert, and DA-021 is only partly resolved. Fix before PR.**

Evidence:
- `frontend/src/components/ui/radio-group.tsx:47-55` renders a native `<input type="radio">` with `className="size-5 shrink-0 border-input accent-primary text-primary focus-visible:outline-2 ..."`. There is no `appearance-none`.
- `grep -n 'radio\|appearance' frontend/src/index.css` returns nothing, and `@tailwindcss/forms` is not a dependency (package.json diff above). A native radio ignores `border-color` and `color` (`text-primary`) when `appearance` is `auto`, so `border-input` and `text-primary` have no visible effect.
- What does apply: `size-5` (20 px), `accent-primary` (checked fill), `focus-visible:outline-*` (outlines do render on native radios), and the label's `coarse:min-h-11`. These satisfy 3 of the 5 prescriptions in the DA-021 row (design-audit.md:84).
- Not delivered: the "decorative contrast" complaint about `border-border`. The unchecked ring is now drawn by the user agent, not from the `--input` token, so there is no token guarantee of 3:1 non-text contrast in light or dark.
- `frontend/src/components/ui/__tests__/controls.test.tsx:65` asserts `toContain("border-input")`, which passes on a class that does nothing. It tests the name, not the behaviour.
- Related: the three `type="radio"` inputs at `features/calendar/event-form-dialog.tsx:370,399,427` ("Ends" radios) are still raw native inputs (progress.md:206).

Recommended fix (one small task): give the radio a real token style (`appearance-none rounded-full border border-input bg-transparent checked:border-primary checked:bg-primary` plus an inner dot via `checked:` background or `before:`), drop the dead `text-primary`, make the test assert the effective classes, and move the event-form "Ends" radios to `RadioGroupItem`. The cheaper alternative is to drop the dead `border-input`/`text-primary`, reword the DA-021 row as "native control, accent-primary, contrast browser-defined", and accept it explicitly. Leaving it as is records DA-021 as resolved when it is not.

### Judgment 2: Triage of deferred minor findings (progress.md `minor (deferred`)

Columns: line (progress.md) | item | decision.

### Fix before PR
| Line | Item | Why |
|---|---|---|
| 165 | Task 14A1 radio `border-input`/`text-primary` inert, test asserts dead class | See Judgment 1. |
| 206 | Task 14D event-form "Ends" radios still native | Same control family as DA-021; fix together. |
| 115 | Task 10 `label.tsx` disabled uses `opacity-60` on text alone | Contrast risk on a shared primitive, one-line change. Same pattern at line 167 (`select.tsx:45` trigger `disabled:opacity-60`); fix both together. |
| 196 | Task 14C2 kebab destructive item double text class | Already fixed in 334d7fd: `dashboard-kebab-menu.tsx:145-148` merges via `cn()`. Close it. |
| 56, 54 (no code) | Task 3 row count 139 vs 140; DA index.css line cites drift | No action; note in the PR body only. |

### Defer to a named stage
| Line | Item | Defer to |
|---|---|---|
| 134 | ErrorPage has no h1 (title is an EmptyState h2) | Stage 2, Task 16 (error pages) |
| 135 | data-table empty message is an h2; heading outline untested | Stage 2/3 with page-header adoption (Task 15/16 pass) |
| 175 | LoginPage/OnboardingPage titles are CardTitle divs, not headings | Stage 2, Task 16 |
| 178 | wordmark `font-heading` untested; 390 px wrap not rendered | Stage 2, Task 15 (shell); rendered check |
| 125 | palette-drawer / config-panel `font-heading text-base` | Stage 3 (Ruling R6, already owned) |
| 204 | calendar literals event-popover:38 `bg-black/20`, calendar-grid:153,156, calendar-selection-panel:40 | Later calendar stage (Task 22 area), PENDING list owner |
| 205 | event-block time line `opacity-80` vs readable foreground; eventChipColors tests don't pin ""/"teal" | Task 22 (as recorded) |
| 73, 88 | `--calendar-busy` dark L nudge; `--ease-out` overrides Tailwind `ease-out` | Task 32 (DESIGN.md record) |
| 42, 43 | PRODUCT.md prose glance; "installed-PWA-style" wording | PR review by user; no code |
| 41 | hook path not exercised | Defer, informational |
| 126, 127 | card.tsx double edge in dark; select popup `shadow-md/duration-100` | Stage 2 polish pass (rendered check first) |
| 117, 166 | input/select no coarse enlargement beyond `coarse:min-h-11`; input `md:text-sm` 14 px on coarse tablets (iOS zoom) | Stage 2 shell/forms pass; needs device check |
| 176 | dialog title `pr-8` when `showCloseButton=false` | Stage 2 polish (cosmetic) |
| 185 | recipe-ingredients:29,31 `aria-label` on role-less svg | Stage 6 recipes pass |
| 186 | bulk-categorize label wraps Checkbox that has its own `aria-label` | Stage 6 ingredients pass |
| 136 | weather-icon `{...props}` can override `aria-hidden` | Stage 3 widgets (weather-icon consumer) |

### Defer: rendered check (no browser this session; stays deferred)
Lines 81 (size-adjust fallback layout shift), 89 (reduced-motion keyframes collapse to 0.01 ms, finish reviewer), 90 (focus offset in `--background`), 148 (favicon oklch fills in old browsers), 177 (dialog `overflow-y-auto` clipping focus outlines), 187 (skeleton height estimates IngredientsPage:200, ShoppingListsPage:123), 197 (RecipesPage:152 `h-28` skeleton).

### Defer: test-quality or tooling nits (no user impact; not blocking)
Lines 64 (parseColor hex trim), 74 (@theme-inline test coverage), 82 (@theme regex nested braces), 91 (brace indentation), 96 (TRACKER_COLORS order assertion), 106-108 (guard multi-line hex, short hex false positives, `findColorLiterals` exported from a test file), 116 (Badge children fragment), 128 (user-avatar paper-branch spy), 146-147 (brand-mark stroke guard, no RED run), 168 (checkbox disabled test tests base-ui), 188 (Link-overlay cards untested), 198 (kebab/Workout badge/export-modal skeleton untested). Defer to Stage 7 (final verification and test-hygiene pass); none gates the PR.

### File Inventory

- Components (common): brand-mark, card-action-menu, data-table, empty-state, error-card, error-page, list-filter-bar, owner-select, page-header, pull-to-refresh, weather-icon
- Components (ui): badge, button, calendar, card, checkbox, dialog, input, label, popover, radio-group, select, sidebar, skeleton, status-banner, textarea, tooltip, user-avatar
- Components (features): calendar (all-day-event-row, connection-status, event-block, event-form-dialog, event-popover, month-day-cell, month-event-chip, reauthorize-banner), dashboards/dashboard-kebab-menu, ingredients/bulk-categorize, meals (export-modal, ingredient-preview, week-grid), navigation (app-shell, mobile-drawer), recipes (cooklang-preview, planner-config-form, recipe-card, recipe-ingredients, step-segment)
- Pages: IngredientDetail, Ingredients, Login, Onboarding, RecipeDetail, RecipeForm, Recipes, ShoppingLists, WorkoutReview, WorkoutToday, dashboard-designer/zod-form
- Other: `lib/color-contrast.ts`, `lib/constants/swatches.ts`, `index.css`
- Hooks / Services / Schemas / Types: none in scope

### Anti-Pattern Checklist

Scope note: "introduced" means the offending line is in the branch's added lines; "existing" means it predates this branch but sits in a touched file.

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` | FAIL (existing, non-blocking) | `pages/dashboard-designer/zod-form.tsx:164,189,210,243,273` `name={name as any}`. Not in the diff's added lines; the Stage 1 diff there only swaps inputs for primitives. All other 61 files: zero matches. |
| FE-02 | No manual class concatenation | FAIL (existing, non-blocking) | `features/calendar/connection-status.tsx:81`, `pages/WorkoutReviewPage.tsx:243`, `pages/WorkoutTodayPage.tsx:172` use template-string `className`. None is an added line. Cheap to convert to `cn()` while the files are being touched (the first is DA-adjacent). |
| FE-03 | No API client in components | PASS | `grep 'lib/api/client'` over the 62 files: no matches |
| FE-04 | No inline Zod in components | FAIL (minor) | `features/dashboards/dashboard-kebab-menu.tsx:32` `z.object({ name: dashboardNameSchema })` wraps a shared schema, not a cross-field `.refine()`. Existing, not changed by this branch. |
| FE-05 | No spinners for content | PASS | `animate-spin` is on submit buttons (`event-form-dialog.tsx:543`, `OnboardingPage.tsx:191,244`, `RecipeFormPage.tsx:305`), a parse indicator (`cooklang-preview.tsx:21`, motion-reduce guarded), a refresh icon button (`recipe-ingredients.tsx:66`), sync icon (`connection-status.tsx:81`) and the pull-to-refresh gesture (`pull-to-refresh.tsx:34`). None is a content-loading spinner. |
| FE-06 | No hardcoded colours | FAIL (existing, owned) | `features/calendar/event-popover.tsx:38` `bg-black/20`; `features/navigation/mobile-drawer.tsx:40` `bg-black/40`. Neither is added by this branch. event-popover is on the PENDING list for the calendar stage (progress.md:204). mobile-drawer scrim is owned by Task 15 (shell). No hex or palette-class literals elsewhere in the 62 files. |
| FE-07 | No state mutation | PASS | The `.push` hits in `calendar-utils.ts:60,102,201-240,332` are local array building; `.sort` at `:187` is on a copy `[...events]`. No setState on a mutated reference. |
| FE-08 | No default exports | PASS | `grep 'export default'`: no matches |
| FE-11 | `createErrorFromUnknown` in catch | PASS | no `.catch(` in the 62 files |

### Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-09 | Tenant guard in hooks | N/A | no hook files changed |
| FE-10 | Tenant ID in query keys | N/A | no hook files changed |
| FE-12 | JSON:API model shape | N/A | no `types/models` changed |
| FE-13 | Service extends BaseService | N/A | no services changed |
| FE-14 | Query key `as const` | N/A | no hooks changed |
| FE-15 | react-hook-form + zodResolver | PASS | `features/calendar/event-form-dialog.tsx:2-3,88-89` (`useForm({ resolver: zodResolver(eventFormSchema) })`); `zod-form.tsx` keeps Controller-based fields on the passed form |
| FE-16 | Schema in `lib/schemas` with inferred type | N/A | no schemas changed |

### Styling Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-19 | Interactive elements show `cursor-pointer` | FAIL (minor) | Passing: `event-block.tsx:25`, `month-day-cell.tsx:71`, `all-day-event-row.tsx:22`, `data-table.tsx:77-82` (only when `onRowClick`), `button.tsx` base, `checkbox.tsx`, `radio-group.tsx:42`, `select.tsx` trigger, kebab items. Failing: `event-popover.tsx:38` backdrop `<div onClick={onClose}>` has no `cursor-pointer` (existing; a click-outside scrim, so low impact). `card-action-menu.tsx:22` `onClick` is a `stopPropagation` wrapper, not an affordance, and is fine. |

### Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | FAIL | Changed components with no test file: `common/list-filter-bar.tsx` (coarse class added), `common/owner-select.tsx`, `common/pull-to-refresh.tsx`, `calendar/all-day-event-row.tsx`, `calendar/connection-status.tsx`, `calendar/event-block.tsx`, `calendar/event-popover.tsx`, `calendar/month-day-cell.tsx`, `calendar/month-event-chip.tsx`, `meals/export-modal.tsx`, `meals/ingredient-preview.tsx`, `ingredients/bulk-categorize.tsx`. Most are one-class token swaps, which is why per-task reviews accepted them; `event-block`, `month-event-chip` and `all-day-event-row` take new `eventChipColors` logic whose behaviour is covered only through `calendar-utils.test.ts` (calendar-utils.test.ts +23). New primitives `status-banner`, `checkbox`, `color-contrast`, `swatches`, `brand-mark`, `page-header`, `empty-state` all have tests. |
| FE-18 | Mocks updated when services changed | N/A | no services changed |

### Summary

### Blocking (must fix before PR)
- DA-021: radio `border-input`/`text-primary` are inert on a native radio and the test asserts the dead class (`radio-group.tsx:55`, `controls.test.tsx:65`). Either implement a token-styled radio (and move `event-form-dialog.tsx:370,399,427` to it) or amend the finding and the test to describe what the control really does.
- Deferred minors to fix with it: `label.tsx` disabled `opacity-60` on text (progress.md:115) and `select.tsx:45` `disabled:opacity-60` (progress.md:167).

### Non-Blocking (should fix, existing in touched files)
- FE-01 `zod-form.tsx:164,189,210,243,273` `name as any`.
- FE-02 template-string `className` at `connection-status.tsx:81`, `WorkoutReviewPage.tsx:243`, `WorkoutTodayPage.tsx:172`: convert to `cn()` (the connection-status file was edited by this branch for DA-018).
- FE-04 `dashboard-kebab-menu.tsx:32` inline `z.object` wrapper.
- FE-06 `event-popover.tsx:38`, `mobile-drawer.tsx:40` (owned by calendar stage and Task 15 respectively).
- FE-17 test gaps listed above; the calendar chip components are the ones worth a render test (readable-foreground applied) when Task 22 touches them.
- FE-19 `event-popover.tsx:38`.
- Lint warnings: `react-hooks/incompatible-library` x5 (event-form-dialog.tsx, new-dashboard-modal.tsx), existing `watch()` usage.

## Plan Adherence Audit: task-057 Stage 2 (Tasks 15, 16)

**Branch:** task-057-impeccable-redesign-stage-2  **Range:** 548be27..617b691  **Date:** 2026-10-19

### Verdict

MOSTLY_COMPLETE, READY for close-out steps 4 and 5. No blocking items.

### Build and test

| Check | Result |
|---|---|
| `npx vitest run` (frontend) | PASS, 136 files, 879 tests |
| `npx tsc -b` | PASS |
| `npx eslint src` | 0 errors, 5 warnings (react-hooks/incompatible-library on `watch()`, pre-existing pattern) |
| Go | no changes in range |

### Task 15 (commit 76febda)

| Step | Status | Evidence |
|---|---|---|
| Remove `components/features/navigation/` from PENDING | DONE | no-color-literals.test.ts diff, -1 line |
| Tests: mobile header button name + `coarse:size-11` | DONE | mobile-header.test.tsx:15-18 |
| Tests: `aria-current="page"` on active item | DONE | nav-group.test.tsx:28-31 |
| Tests: brand mark in sidebar header and mobile header | DONE | app-shell.test.tsx:72-75, mobile-header.test.tsx:21-24 |
| Implement per design §3.6, Fraunces wordmark at text-lg | DONE | `font-heading text-lg` in app-shell.tsx:43, mobile-header.tsx:9, mobile-drawer.tsx:39. app-shell.tsx was not changed in this range; it already had the recipe from Stage 1, and the test locks it in. |
| DA-042 drawer | DONE | mobile-drawer.tsx rewritten onto base-ui Dialog (unmounts when closed, traps focus, inerts the page, returns focus). Overlay is the `bg-overlay` token (dialog.tsx:36). `bg-black` no longer appears in navigation. mobile-drawer.test.tsx added. Motion is collapsed by the global reduced-motion block (index.css:330). |
| DA-045 | DONE | "My Dashboards" is `text-xs` with no opacity (dashboards-nav-group.tsx). Badge numerals are `text-xs` (sidebar.tsx:226). One shared `groupLabelClasses` exported from nav-group.tsx. Group triggers have `coarse:min-h-11`. Active link gets `aria-current` styling via `text-sidebar-primary`. Clay marker: accepted per ruling (§3.6 "clay text + sage wash"). |
| DA-046 | DONE | `size` prop replaces `iconSize` class string. Shared `menuItemClasses` with `coarse:min-h-11`. "Sign out" casing. Destructive item stays `text-destructive` on highlight, which is `bg-accent`. Contrast on accent is not measured in a render (logged unrendered). |
| Commit message | DONE | `feat(frontend): restyle app shell and navigation` |

### Task 16 (commits d4aa410, 341db21)

| Step | Status | Evidence |
|---|---|---|
| PENDING removal for LoginPage and OnboardingPage | DONE (no-op per ruling) | Not in PENDING before this stage, and the guard passes with both scanned. |
| Tests: one `h1` with `font-heading` | DONE | LoginPage.test.tsx and OnboardingPage.test.tsx, new cases |
| Tests: controls found by `getByLabelText` | DONE | OnboardingPage.test.tsx, "labels every form control" (Account Name, Household Name, Timezone, Imperial, Metric). Login has no form controls. |
| Toaster: drop `richColors`, token classes, keep `closeButton` | DONE, with a documented deviation | App.tsx `ThemedToaster` passes `theme` from `useTheme()`, `closeButton`, `style={toasterStyle}` and `toastOptions` (lib/toast-options.ts). The plan sketch used `toast: "bg-popover …"` classNames. The implementation moves the base surface onto Sonner CSS variables (`--normal-bg` etc.) because Sonner's unlayered CSS beats Tailwind layers. This is a reasoned adaptation. Per-type borders use `border-success!` etc. and the close button is `size-6!`. The App.tsx change goes beyond "Toaster props only" (new wrapper component), but the ruling covers it, and theme-provider.tsx is untouched. |
| DA-041 | DONE | toast-options.test.ts. The per-type icons the Fix column mentions are not explicitly added. Sonner supplies default icons per type, so "not colour-only" holds. |
| DA-043 | DONE | LoginPage.tsx: anchor styled via `buttonVariants` (no nested button, tested), BrandMark, h1, `size="lg"`, `EmptyState` for no providers. The StatusBanner load-error was already present. |
| DA-044 | DONE | Brand mark, "Step N of M" label, Accept at `size="lg"`, "Create My Own" as `secondary`, timezone hint via FormDescription, spinner `motion-safe:animate-spin`. Step-count logic (2 vs 3) is a small behaviour addition but display-only. |
| DA-047 | DONE | error-page.tsx on `EmptyState` with `h1`, "Home" link, `role="status"` live region with a retry counter. protected-route.tsx skeleton has `role="status" aria-label="Loading"` and `hidden md:flex` aside. error-page.test.tsx added. dashboard-skeleton.tsx was not touched; it already had the role. |
| dialog.tsx change | Out of listed Files | `DialogTitle` only reserves `pr-8` when a close button exists, via context. This supports the drawer's sr-only title. Minor, tests added (dialog.test.tsx). |
| Commit messages | DONE | `feat(frontend): restyle login, onboarding, errors, toasts`, plus `fix(frontend): apply token styling to toasts` |

Out-of-scope deferrals (per rulings, acceptable): card dark double edge (item 126) and coarse input sizing (items 117/166) are deferred to a rendered check. Neither is a Stage 2 DA row.

### Close-out state

- Step 1: DONE. Six "not rendered, browser unavailable" rows were appended to the Accessibility verification log (617b691). None claims a render. The PR body must list these routes as manual checks.
- Step 2 (gate) and step 3 (review): not verified here. Step 3 is this review plus the frontend one.
- Steps 4 and 5: pending, as stated.

### Input for step 4 (DA row to fixing commit)

All seven Resolved cells (design-audit.md lines 104-110) are currently empty.

| Row | Fixing commit |
|---|---|
| DA-041 | d4aa410 (Toaster wrapper, richColors removed, Toaster `theme`) and 341db21 (token styling actually applied: `toasterStyle`, `!` per-type borders, 24px close). Use `341db21` as the completing commit, or list both. |
| DA-042 | 76febda |
| DA-043 | d4aa410 |
| DA-044 | d4aa410 |
| DA-045 | 76febda |
| DA-046 | 76febda |
| DA-047 | d4aa410 (error-page.tsx and its test). The protected-route.tsx skeleton part is in 76febda. List both. |

Verify via `git log` before filling in: the DA-047 split across 76febda and d4aa410 is inferred from the file diffs.

### Blocking items

None.

### Non-blocking notes

1. DA-041's Fix column mentions icons on toasts. Sonner defaults are relied upon. Confirm in the rendered check.
2. The destructive-on-accent contrast pair (DA-046) is asserted by class only, with no computed contrast. It is covered by the unrendered log row.

## Frontend Audit — task-057 Stage 2 (Shell)

- **Audit Scope:** `git diff 548be27..617b691` (25 frontend files: navigation shell, ui/sidebar, ui/dialog, LoginPage, OnboardingPage, error-page, App.tsx, toast-options, related tests)
- **Guidelines Source:** frontend-dev-guidelines skill + plan.md Global Constraints
- **Date:** 2026-10-19
- **Build:** PASS (`npm run build`, exit 0; only the chunk-size warning)
- **Tests:** 879 passed, 0 failed (136 files). Runner is vitest; `--watchAll=false` is invalid there, so `npm test -- --run` was used.
- **Overall:** NEEDS-WORK (one blocking finding)

### Blocking

1. `frontend/src/lib/toast-options.ts:20` — `closeButton: "size-6!"` pins the Sonner close button to 24px with `!important` and has no `coarse:` override. This is an icon-only control, so the constraint requires at least 44px under `coarse:`. It meets the 24px floor but fails the coarse rule. Fix: `"size-6! coarse:size-11!"`, and extend `lib/__tests__/toast-options.test.ts` to assert it.

### Non-blocking

- `frontend/src/components/features/navigation/user-menu.tsx:51-52` — the popup uses `shadow-lg` and `duration-100`. The drawer (`mobile-drawer.tsx:32`) and select use `shadow-overlay` and `duration-(--duration-*)`. Not a stated constraint, and the global reduced-motion rule (`index.css:330`) collapses it. Align the tokens when convenient (Stage 7 polish).
- `frontend/src/components/features/navigation/nav-group.tsx:48` and `dashboards-nav-group.tsx:66` — chevron uses a literal `duration-200` rather than a token. It has `motion-reduce:transition-none`, so the constraint is met.
- `frontend/src/components/common/error-page.tsx:84-86` — the live region keeps the last "Retrying… (attempt N)" text and is never cleared. It is harmless. Fine to leave.
- `frontend/src/App.tsx:1` — the `toast-options` import sits above `react`. Cosmetic.

### Checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 no `any` | PASS | Grep of all non-test changed files: no matches |
| FE-02 no class concatenation | PASS | No `className={"…" +` or template strings in classNames. The template literals found are routes and text: `App.tsx:52`, `OnboardingPage.tsx:77`, `dashboards-nav-group.tsx:51,53`, `error-page.tsx:85`. All other classes are whole literals combined through `cn()` or `buttonVariants()`, e.g. `nav-group.tsx:48`, `LoginPage.tsx:34` |
| FE-03 no direct API client | PASS | No `api/client` imports in the changed files |
| FE-04 no inline Zod | PASS | No `z.` matches |
| FE-05 no content spinners | PASS | `OnboardingPage.tsx:206,260` are submit-button spinners only, and use `motion-safe:animate-spin`. Content loading uses Skeleton (`protected-route.tsx`, `LoginPage`) |
| FE-06 no colour literals or palette classes | PASS | No palette, hex or oklch literals in the changed files. `toast-options.ts:6-9` references tokens via `var(--…)`. The guard test now scans `navigation/` (`no-color-literals.test.ts` drops it from PENDING) and passes |
| FE-07 no state mutation | PASS | No `.push(`, `.sort(` or `.splice(` in the diff |
| FE-08 no default exports | PASS | No `export default` matches |
| FE-09/10 tenant guard and keys | N/A | No hooks changed |
| FE-11 error handling | PASS | `OnboardingPage.tsx` catches use `createErrorFromUnknown` with a toast (unchanged pattern, e.g. line ~66) |
| FE-12/13/14/16 | N/A | No models, services, key factories or schemas changed |
| FE-15 forms | PASS | `OnboardingPage.tsx` still uses `useForm` + `zodResolver`; `FormDescription` was added at ~line 238 |
| FE-17 tests exist | PASS | Tests changed or added for app-shell, mobile-drawer, mobile-header, nav-group, protected-route, user-menu, dialog, error-page, toast-options, LoginPage, OnboardingPage |
| FE-18 mocks | N/A | No services changed |
| FE-19 cursor-pointer | PASS | `nav-group.tsx:26` (`groupLabelClasses`), `user-menu.tsx:16,31`. `SidebarMenuButton` renders an `<a>` or `<button>` |

### Plan Global Constraints

| Constraint | Status | Evidence |
|-----------|--------|----------|
| No colour literals or palette classes | PASS | See FE-06 |
| Whole literal Tailwind strings | PASS | See FE-02 |
| `font-heading` only at text-lg or larger, never on controls | PASS | `mobile-drawer.tsx:39` and `mobile-header.tsx:9` are `text-lg`. `LoginPage.tsx:19` and `OnboardingPage.tsx:123` are `text-2xl` on h1. `dialog.tsx:129` is `text-lg` on DialogTitle. None is on a control |
| Status never colour-only | PASS | Toast type borders (`toast-options.ts:16-19`) are supplementary, since Sonner renders a type icon. The nav active state combines a wash and clay text with `aria-current`. The destructive menu item has a LogOut icon and a "Sign out" label |
| Pointer targets at least 24px; nav items and icon-only controls at least 44px under `coarse:` | FAIL | Toast close button, see Blocking 1. Others pass: `sidebar.tsx:128,164` (`coarse:min-h-11`), `nav-group.tsx:26`, `user-menu.tsx:16,31`, and Button `icon` is `size-8 coarse:size-11` (`button.tsx:30`) |
| Motion collapses under reduced motion | PASS | Global rule at `index.css:330-337`. The chevrons also carry `motion-reduce:transition-none` |
| `theme-provider.tsx` unmodified | PASS | `git diff --stat` shows no changes. `App.tsx:56-66` only reads `useTheme()` |
| Tests updated, not deleted | PASS | Diff shows additions only, plus a single line removed from the PENDING list (the intended Stage 2 graduation) |

### Deferred-minor triage (`.superpowers/sdd/plan/stage2-deferred-minors.md`)

| Item | Decision |
|------|----------|
| DA-045 "clay marker" not added (design.md §3.6 binding) | Accept ruling. design.md specifies clay text plus a sage wash pill, which `sidebar.tsx:167` delivers |
| Ruling: DA-045 touches `sidebar.tsx` outside the plan's Files list | Accept. The edits are 2 lines and in scope |
| Ruling: Task 16 PENDING removal is a no-op | Accept |
| Ruling: `theme` via a `ThemedToaster` wrapper in App.tsx | Accept. Confirmed at `App.tsx:56-66`, and theme-provider is untouched |
| Ruling: items 126, 117, 166 need a device check | Not blocking. Put in the PR body manual checks, owner Stage 7 polish |
| Item 127 select popup | Already resolved (`select.tsx:84`). Nothing to do |
| Task 15: redundant `aria-[current=page]:text-sidebar-primary` (`nav-group.tsx:66`) | Not blocking. Remove it in Stage 7 cleanup, or keep it as a harmless fallback |
| Task 15: no active-state visual test, and none for dashboards-nav-group `aria-current` | Not blocking. Stage 7 test hardening. A jsdom class assertion would be cheap if wanted now |
| Task 15: DA-046 destructive-on-accent contrast not pinned | Not blocking. I computed the pair: light about 5.0:1 and dark about 4.5:1, so it passes AA but with thin dark-mode margin. Pin it in `design-tokens.test.ts` in Stage 7 (owner: tokens or contrast test) |
| Task 16: no shared skeleton wrapper | Not blocking. Dashboards already use role=status. Stage 7 |
| Task 16: Login load-error StatusBanner untested | Not blocking, but a cheap test. Task 16 or Stage 7 |
| Task 16: rendered toast appearance not evaluable | PR body manual check. Include the coarse close-button size once Blocking 1 is fixed |
| Task 16: `App.tsx:1` import order | Cosmetic, fix opportunistically |
| Task 16: toast tests are config-shape only | PR body manual check. Stage 7 may add a rendered test |

### Verdict

NEEDS-WORK. Build and tests pass. One blocking constraint violation: the toast close button has no 44px coarse target.

## Plan Adherence Audit — task-057 Stage 3 (Tasks 17–19)

**Branch:** task-057-impeccable-redesign-stage-3  **Range:** 93d320a..03c9210  **Date:** 2026-10-19

### Verdict

MOSTLY_COMPLETE, READY for close-out steps (rendered check, PR, Resolved cells). No blocking items. All 25 Stage 3 DA rows (DA-048..DA-072) have a fixing commit. X1-X3 are cited in commit messages but are not defined anywhere in design-audit.md or the other task docs (grep of `docs/`), so they can only be mapped by commit, not verified against a row.

### Build and test

| Check | Result |
|---|---|
| `npx vitest run src/components/features src/pages src/lib/dashboard src/__tests__` | PASS on re-run: 89 files, 523 tests. The first run showed 2 failures in 1 file under load; the immediate re-run was clean (likely timing flake, not investigated further). |
| `npx tsc -b` | PASS (no output) |
| `git diff --stat origin/main -- frontend/src/lib/dashboard` | empty (constraint held) |
| Flagless `tools/verify.sh` | passed at 03c9210 per controller; not re-run |

### Task 17 (bdba882, 6744ad7, 21ea425)

| Step | Status | Evidence |
|---|---|---|
| Cut stage branch | DONE | branch task-057-impeccable-redesign-stage-3 |
| Test iterating `widgetRegistry`, asserts `@container` root and `font-heading` title | DONE | widget-container.test.tsx imports `@/lib/dashboard/widget-registry`; extended in 6744ad7 for wrapped widgets |
| PENDING entry `components/features/dashboard-widgets/` removed | DONE | no-color-literals.test.ts diff in bdba882 |
| Implement ramp (`@container`, `font-heading text-lg @xl:text-2xl`, `@xl:` icon/list steps) | DONE | 16 adapter/summary files in bdba882. Calendar `userColor` dot kept. |
| Wrapped widgets (calendar, meals, packages, habits, weather, workouts) | DONE (beyond listed Files) | 6744ad7, 21ea425 touch `components/features/{calendar,meals,packages,trackers,weather,workouts}/*-widget.tsx`. Needed for the registry-wide test to pass; reasonable scope extension. |
| `lib/dashboard` untouched | DONE | diff stat empty |
| Commit message | DONE | `feat(frontend): container-query widget internals for kiosk legibility` |

### Task 18 (0f494d3, f6bf43c)

| Step | Status | Evidence |
|---|---|---|
| PENDING entries removed (5 pages, designer dir, dashboards dir) | DONE | no-color-literals.test.ts: 7 lines removed in 0f494d3 (dashboard-widgets line removed in bdba882); all 8 Stage 3 entries are gone from the diff vs 93d320a |
| `bg-black/30` -> `bg-overlay` | DONE | config-panel.tsx:83, palette-drawer.tsx:32; no `bg-black` remains in dashboard-designer |
| Management page uses `PageHeader` and `EmptyState` | DONE | DashboardManagementPage.tsx:24-25, 116, 128 |
| Greeting header (Fraunces text-2xl, muted date) | DONE | DashboardShell.tsx `PageHeader` at :52 with `toLocaleDateString` at :23 |
| Grid props, layout JSON, state.ts unchanged | DONE | lib/dashboard diff empty; DashboardRenderer.tsx change is 2 lines (9f3512b, `min-w-0`/`@container` wrapper) |
| Vitest src/pages, lib/dashboard, `__tests__`; tsc | DONE | see above (eslint not re-run here) |
| Commit message | DONE | `feat(frontend): restyle dashboard pages and designer` |

### Task 19

| Item | Status | Evidence |
|---|---|---|
| Each Stage 3 row fixed, named in commit message | DONE | commits 48101d9, 21ea425, ca31c47, 9f3512b name rows in the subject |
| Tests for jsdom-testable fixes | DONE | widget-error, lossy-config-badge, tasks-today, calendar-tomorrow, zod-form, DashboardShell, kebab-menu, dashboard-row, new-dashboard-modal, meal-plan-widget tests added/updated |
| Impeccable detector run (Task 14 Step 2) | NOT VERIFIED | No evidence in the commits. Controller should confirm it is part of close-out or in the verification log. |
| `lib/dashboard` diff empty | DONE | verified |
| Stage close-out procedure | PENDING | controller's next steps, not failed |

### Scope rulings (acceptable)

- DA-064/065: class fixes on the existing drawers; no shared `Sheet` primitive. Token scrim (`bg-overlay`) is in place and the hidden dead popup was removed (config-panel.tsx, -14 lines in ca31c47). The row's "Sheet primitive" Fix text is not literally followed.
- meal-plan-widget: card-wide `Link` replaced by an explicit "Open meals" icon link (03c9210) to remove a nested anchor. This is a behaviour change; list it in the PR body.

### Partial or soft items (non-blocking)

1. DA-056: 44px coarse targets applied to the drag handle, edit link and kebab (f6bf43c, 9f3512b). The row's "hide or relabel Edit on mobile" part was not done; the Edit link still renders on every viewport (dashboard-row.tsx:55-77). The designer shows a blocker on mobile (DA-072), so the user reaches an explanation rather than a dead end. Treat as DONE-with-note.
2. DA-052: fix uses `useId` and `htmlFor` (zod-form.tsx:104, 172, 204). Whether raw inputs were swapped for the `Input`/`Checkbox` primitives was not checked.
3. DA-066 "tooltip/aria-describedby" for the lossy badge was refined in 03c9210 (ids).
4. Contrast and rendered items are asserted by class only; they will be recorded as `not rendered — browser unavailable`.

### Input for step 5 (DA row to fixing commit)

| Row | Fixing commit(s) |
|---|---|
| DA-048 | bdba882 (container ramp); 6744ad7 (wrapped widgets) |
| DA-049 | 48101d9 |
| DA-050 | 48101d9 (shared `widget-error.tsx`) |
| DA-051 | 48101d9; 21ea425 (calendar dot name) |
| DA-052 | ca31c47 |
| DA-053 | ca31c47 |
| DA-054 | ca31c47 |
| DA-055 | 9f3512b |
| DA-056 | f6bf43c (coarse targets); 9f3512b (kebab) |
| DA-057 | ca31c47 |
| DA-058 | 9f3512b (0f494d3 introduced `PageHeader` on the management page) |
| DA-059 | 9f3512b (0f494d3 introduced `EmptyState`) |
| DA-060 | 9f3512b |
| DA-061 | 9f3512b |
| DA-062 | 9f3512b |
| DA-063 | 9f3512b (a11y tests in 03c9210) |
| DA-064 | ca31c47 (0f494d3 for `bg-overlay`) |
| DA-065 | ca31c47 (0f494d3 for `bg-overlay`) |
| DA-066 | 48101d9; 03c9210 |
| DA-067 | 48101d9 |
| DA-068 | 48101d9 |
| DA-069 | 48101d9; 21ea425 |
| DA-070 | 48101d9 |
| DA-071 | 9f3512b |
| DA-072 | ca31c47 (designer blocker); 9f3512b (DashboardShell schema alert, DashboardRedirect) |
| X1, X2 | 21ea425 (not defined in design-audit.md) |
| X3 | 9f3512b (not defined in design-audit.md) |

The Stage column in design-audit.md is the second-to-last cell; rows DA-048..DA-072 all have Stage = 3 and empty Resolved cells (lines 111-135). DA-045..047 belong to Stage 2.

### Blocking items

None.

## Frontend Audit — task-057 Stage 3 (Dashboards & kiosk)

- **Audit Scope:** `git diff 93d320a..HEAD -- frontend/` (49 files: 34 source, 15 tests). Covers the dashboard-widgets adapters, dashboards list/row/kebab/modal, the dashboard-designer panels, the Dashboard* pages, the six legacy widgets (calendar, meals, packages, trackers, weather, workouts), and the `no-color-literals` guard.
- **Guidelines Source:** frontend-dev-guidelines skill + plan.md Global Constraints
- **Date:** 2026-10-19
- **Build:** PASS (`npm run build`, exit 0; only the chunk-size warning)
- **Tests:** 921 passed, 0 failed (139 files). Runner is vitest, so `npm test` (`vitest run`) was used instead of `--watchAll=false`.
- **Overall:** NEEDS-WORK (two blocking findings, both cheap)

### Blocking

1. `frontend/src/components/features/trackers/habits-widget.tsx:92` and `frontend/src/components/features/workouts/workout-widget.tsx:78` — `className={\`text-sm @xl:text-xl truncate ${cond ? "…" : ""}\`}` is template-string class concatenation. Stage 3 edited both lines (it added `@xl:text-xl`). This fails FE-02 and the "whole literal class strings" constraint. Fix: `className={cn("text-sm @xl:text-xl truncate", (done || skipped) && "text-muted-foreground line-through")}`, and the same for `resolved` in workout-widget. Import `cn` if it is missing.
2. `frontend/src/components/features/meals/meal-plan-widget.tsx:105-110` — the recipe `Link` is `text-sm @xl:text-xl … truncate` with no target-size floor. It is a list-row pointer target, so the constraint requires `coarse:min-h-11` (and ≥24px otherwise). The tasks adapters already do this (`tasks-today-adapter.tsx:75`: `inline-flex min-h-6 items-center … coarse:min-h-11`). Fix: add `inline-flex min-h-6 items-center coarse:min-h-11` to the link. The existing `meal-plan-widget.test.tsx` role/name queries are unaffected. This is deferred item (a).

### Non-blocking

- `frontend/src/pages/dashboard-designer/zod-form.tsx:168,194,218,252,283` — five `name={name as any}`. FE-01 fails literally. None were introduced by Stage 3 (the diff adds none). Leave, or type the `FieldPath` generic in Stage 7.
- `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx:33` — inline `z.object({ name: dashboardNameSchema })` in a component (FE-04). Pre-existing, and it composes the shared `new-dashboard-modal.schema` field. Move `renameSchema` into that schema file when convenient.
- `frontend/src/components/features/weather/weather-widget.tsx:137-140` — `<Card onClick={navigate}>` is a clickable div. It has `cursor-pointer` (FE-19 met) but is not keyboard-operable, unlike the sibling widgets that wrap in a `Link`. Pre-existing. Convert it to a `Link` wrapper in a later stage.
- `frontend/src/components/features/meals/meal-plan-widget.tsx:50-55` — a bespoke error card (`text-destructive`, no icon) instead of the shared `WidgetError`. The text says "Failed", so it is not colour-only, but it is inconsistent with the other widgets. Swap in `WidgetError`.
- `frontend/src/__tests__/no-color-literals.test.ts:10-31` — the calendar, meals, packages, trackers, weather and workouts directories remain in PENDING, so the guard does not cover the widget files Stage 3 edited. I grepped them by hand and found no palette classes, hex values or oklch literals. Graduate them when those pages are migrated.
- FE-17 — no test files exist for the Stage 3 edits to `weather-widget`, `habits-widget`, `workout-widget`, `package-summary-widget` and `calendar-widget`. The edits are class and typography only. `meal-plan-widget`, the adapters, the dashboards components and the designer panels all have tests.
- `frontend/src/pages/DashboardDesigner.tsx:32` — `export default function`. FE-08 covers components, and this is a pre-existing page default consumed by the router. Leave.
- `frontend/src/pages/dashboard-designer/widget-chrome.tsx:42-43` — `disabled` and `aria-disabled` are both set. Redundant but harmless. See (c).

### Checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 no `any` | FAIL (pre-existing) | `zod-form.tsx:168,194,218,252,283`. No other source file matches. Non-blocking, see above |
| FE-02 no class concatenation | FAIL | `habits-widget.tsx:92`, `workout-widget.tsx:78` (Blocking 1). The other files use whole literals and `cn()`, e.g. `dashboard-row.tsx:65`, `dashboard-kebab-menu.tsx:101` |
| FE-03 no direct API client | PASS | Grep for `api/client`: no matches |
| FE-04 no inline Zod | FAIL (pre-existing) | `dashboard-kebab-menu.tsx:33` (non-blocking) |
| FE-05 no content spinners | PASS | Grep for `animate-spin`: no matches. Content loading uses Skeleton (`meal-plan-widget.tsx:54`) |
| FE-06 no colour literals | PASS | The palette/hex/oklch/rgb/hsl/`success-text` grep returns zero matches across all 34 source files. The guard test graduates the dashboard-widgets, dashboards, dashboard-designer and Dashboard* pages, and passes |
| FE-07 no state mutation | PASS | `.sort(` sites at `calendar-widget.tsx:17` (spread copy), `calendar-tomorrow-adapter.tsx:67` (sorts the fresh array from `getEventsForDateStr` / spread, never state), `reminders-today-adapter.tsx:50` and `meal-plan-widget.tsx:73` (after `.filter`/`.map`), `DashboardRenderer.tsx:23` (spread copy). No `.push`/`.splice`. None feed setState |
| FE-08 no default-export components | PASS | The only match is the page `DashboardDesigner.tsx:32` |
| FE-09/10/12/13/14/16 | N/A | No hooks, models, services, key factories or schemas changed |
| FE-11 error handling | PASS | `new-dashboard-modal.tsx:116` and `zod-form.tsx:297` are `catch {}` blocks. The former defers to the mutation-hook toasts (comment at :117), and the latter is JSON-parse fallback input handling with no async or API call. Neither swallows an API error silently |
| FE-15 forms | PASS | The rename and new-dashboard forms use `react-hook-form` + `zodResolver` (`dashboard-kebab-menu.tsx`, `new-dashboard-modal.tsx`) |
| FE-17 tests exist | PASS with gaps | See the non-blocking list. Tests were added or updated for the adapters, widget-error, lossy badge, kebab menu, row, modal, management page, shell, config-panel, palette-drawer, zod-form and meal-plan-widget |
| FE-18 mocks | N/A | No services changed |
| FE-19 cursor-pointer | PASS | The `<button>`/`<a>`/`Link` elements are native. The only clickable non-native element is `weather-widget.tsx:138` (`cursor-pointer`). The `MenuPrimitive.Item` has `cursor-pointer` (`dashboard-kebab-menu.tsx:111`) |

### Plan Global Constraints

| Constraint | Status | Evidence |
|-----------|--------|----------|
| No colour literals or palette classes | PASS | See FE-06 |
| Status tokens only (no `success-text`) | PASS | Grep for `success-text`: zero matches |
| Whole literal class strings | FAIL | See Blocking 1 |
| Fraunces (`font-heading`) at least `text-lg`, never on controls | PASS | All 19 `font-heading` uses are `text-lg` or larger on CardTitle/Title/h2 (e.g. `tasks-today-adapter.tsx:60`, `palette-drawer.tsx:42`, `DashboardManagementPage.tsx:71`). None is on a Button, Link or input. The `Link` inside CardTitle, e.g. `tasks-today-adapter.tsx:61`, inherits the heading font and is a heading link, not a control |
| Status never colour-only | PASS | `widget-error.tsx:11-14` pairs an AlertTriangle icon with text. `tasks-today-adapter.tsx:68-70` pairs the Overdue colour with an icon and a count label. The default badge uses a Star icon (`dashboard-row.tsx:62`) |
| Pointer targets at least 24px; icon-only controls, list rows and primary actions `coarse:` 44px | FAIL | Meal-plan recipe link (Blocking 2). Others pass: `dashboard-row.tsx:50,75` (`size-7 coarse:size-11`), `dashboard-kebab-menu.tsx:93` (`coarse:size-11`), the card-action chevrons (`size-6 … coarse:size-11`), the drawer close buttons (`palette-drawer.tsx:46`, `config-panel.tsx:97`), the widget-chrome `icon-sm` (Button `coarse:size-11`, `button.tsx:34`), the `sm` Button (`coarse:min-h-11`, `button.tsx:26`) |
| Motion collapses under reduced motion | PASS | The global rule at `index.css:330` plus `motion-safe:transition-colors` (`tasks-summary.tsx:54`, `reminders-summary.tsx:49`) and `prefersReducedMotion()` for the dnd transition (`dashboard-row.tsx:29`). The `transition-opacity` links on legacy widgets are collapsed by the global rule |
| Tests updated, not deleted | PASS | `git diff --diff-filter=D 93d320a..HEAD` shows no deleted test files. The only guard-test change removes PENDING entries |
| `frontend/src/lib/dashboard/**` untouched | PASS | `git diff --stat 93d320a..HEAD -- frontend/src/lib/dashboard` is empty |

### Deferred-minor triage

| Item | Decision |
|------|----------|
| (a) meal-plan recipe links lack `coarse:min-h-11` | FIX BEFORE PR. It violates a binding constraint, it is a one-line class change, and the tasks adapters set the precedent (Blocking 2) |
| (b) "Open meals" link has no explicit focus ring | LEAVE. The global `:focus-visible` rule at `index.css:314` covers it. Add the explicit ring only if you want parity with the adapter chevrons |
| (c) the disabled configure button's `title` is unreachable | LEAVE (Stage 7). `button.tsx:9` sets `disabled:pointer-events-none`, so the tooltip never fires. The state is still conveyed by `disabled` and `aria-label`, and the title is only a supplementary hint. A fix would wrap the button in a span, or put the reason into `aria-describedby` |
| (d) the Overdue heading's `text-destructive` dark-mode contrast is unpinned | LEAVE. I computed destructive against card: light about 5.4:1 (L 0.52 vs 0.995) and dark about 6.6:1 (L 0.71 vs 0.23), both above AA. `design-tokens.test.ts:12` pins destructive against background. Optionally add `["--destructive","--card"]` to that list |
| (e) `widget-error.tsx` has no retry affordance | LEAVE. Not a stated constraint. The widgets are read-only and refetch on their own query cadence. Record it as a Stage 7 or backlog UX improvement |

### Verdict

NEEDS-WORK. Build and tests pass (921/921). Two blocking constraint violations: template-string class concatenation (`habits-widget.tsx:92`, `workout-widget.tsx:78`) and the missing coarse target on the meal-plan recipe link (`meal-plan-widget.tsx:105`). Eight new non-blocking findings are listed above, four of them pre-existing. Deferred items (b) to (e) are leave, and (a) is fix-before-PR.

## Plan Adherence — task-057 Stage 4 (Food)

**Range:** cd06366..a68b611 (07b8da2 Task 20a, 67e1f3f Task 20b, a68b611 Task 21). **Branch:** task-057-impeccable-redesign-stage-4.

**Verdict: MOSTLY_COMPLETE, READY_TO_MERGE (no blocking findings).** Tasks 20 and 21 were executed; DA-073..087 are all addressed in code. Vitest over touched areas (recipes, meals, ingredients, pages, `__tests__` incl. the colour-literal guard) passes: 44 files, 235 tests. Full-suite and tsc/eslint results are from the Task 21 report (142 files / 933 tests, tsc clean, eslint 0 errors), not rerun here. The flagless `verify.sh` is the controller's gate. The plan's Task 20/21 checkboxes in plan.md and the Resolved column in design-audit.md are not yet ticked (step 4 of close-out, as in Stage 3).

### Task completion

| # | Task | Status | Evidence |
|---|---|---|---|
| 20 S1 | Cut stage branch | DONE | progress.md: branch cut from cd06366 after ancestor check |
| 20 S2 | Remove PENDING entries; guard | DONE (ruling) | `no-color-literals.test.ts` diff removes `features/ingredients/`, `features/meals/`, `features/recipes/` and `pages/{Meals,Recipes,ShoppingListDetail,WishList}Page.tsx`. Ruling: the "guard fails" step became "guard passes", because every known literal was already fixed in Stage 1 (DA-011..013). Recorded, not a deviation. |
| 20 S3 | `recipe-status.test.tsx` | DONE | `recipes/__tests__/recipe-status.test.tsx` (71 lines) pins `text-success` and `text-warning-text` on RecipeCard. Passes before and after the change, as the ruling anticipated. |
| 20 S4 | Implement mapping; PageHeader/EmptyState; `coarse:min-h-11` on shopping rows | DONE with note | All nine pages use PageHeader or EmptyState (grep). Shopping-mode rows: 20b added `coarse:min-h-11`, and Task 21 changed it to a plain `min-h-16` (`ShoppingListDetailPage.tsx:291`), so touch rows stay 64px (at or above 44px). This differs from the plan text, but it meets the intent (the 44px minimum) and has a test. Minor, see below. |
| 20 S5 | Guard, area tests, tsc, eslint | DONE | Tests rerun here; tsc/eslint per implementer reports |
| 20 S6 | Two commits `feat(frontend): restyle <area>` | DONE | 07b8da2 "restyle recipes and meals"; 67e1f3f "restyle ingredients, shopping and wish list" |
| 21 | Resolve remaining Stage 4 findings (Task 14 steps) | DONE | a68b611 fixes the 20b minors. Detector: 0 findings (regex rules only), run by the controller. The rendered check is "not rendered", disclosed. |

**Completion rate:** 7/7 (100%). **Skipped without approval:** 0. **Partial:** 0.

### DA rows

| Row | Status | Evidence |
|---|---|---|
| DA-073 | DONE | `ShoppingListDetailPage.tsx:329-330` `role="checkbox"` and `aria-checked` on a button; row `min-h-16` (:291) |
| DA-074 | DONE | `aria-label="Back to lists"` (:155); `Remove ${item.name}` (:347) |
| DA-075 | DONE (behaviour change) | `ShoppingListsPage.tsx` Rename/Delete `aria-label`s (:178, :189, :207); confirm Dialog (:249). Ruling. Name it in the PR body. |
| DA-076 | DONE | `week-selector.tsx`, test added in `week-selector.test.tsx` |
| DA-077 | DONE | `week-grid.tsx:111` sr-only "Recipe deleted"; remove button changed |
| DA-078 | DONE | `category-manager.tsx` labels |
| DA-079 | DONE | `ingredient-resolver.tsx`, `recipe-ingredients.tsx` |
| DA-080 | DONE | `recipe-selector.tsx` |
| DA-081 | DONE (ruling) | `plan-item-popover.tsx:233` `aria-pressed` on Button chips (no toggle primitive) |
| DA-082 | DONE | `RecipesPage.tsx:116,125,143` `aria-pressed` toggles |
| DA-083 | DONE | PageHeader on all 9 pages (grep count per file) |
| DA-084 | DONE | EmptyState on the page-level empties and `recipe-selector.tsx:81`; inline panel empties kept per the audit |
| DA-085 | DONE | `cook-mode.tsx:117` sr-only wake-lock text; `text-cook-*`/`size-cook-badge` utilities added in `index.css` |
| DA-086 | DONE | Nested cards flattened (RecipeFormPage keeps one inner divider, accepted by the 20a review) |
| DA-087 | DONE (ruling) | `recipe-image-field.tsx:93` uses `border-input`, because no `border-strong` token exists |

### Non-blocking findings (carry to Stage 7 or note in PR)

1. Plan text said `coarse:min-h-11`; the shipped code is `min-h-16` for shopping mode. The intent is met and tested, but the plan text and the code differ.
2. DA-075 adds a confirm dialog (behaviour change). The archived-list Delete on `ShoppingListDetailPage` and the delete on `IngredientDetailPage` remain unconfirmed (Stage 7 candidate).
3. Ingredients "uncategorized" badge sits inside the `h1`, so the count enters the heading's accessible name (disclosed).
4. `recipe-selector` and `ingredient-resolver` use `focus-visible:outline-2` without an offset.
5. The `ShoppingListsPage.test.tsx:58` delete-failure test passes before the fix and does not assert the toast. The keyboard test uses `.focus()` plus Enter, not Tab reachability.
6. `recipe-status` assertions pass both pre- and post-change (pins state per ruling).
7. The rendered check was not done (browser unavailable). That is the controller's disclosure.

### Input for step 5 (DA row to fixing commit)

| Row | Fixing commit(s) |
|---|---|
| DA-073 | 67e1f3f; a68b611 (keeps 64px rows on touch) |
| DA-074 | 67e1f3f |
| DA-075 | 67e1f3f; a68b611 (delete-failure test) |
| DA-076 | 07b8da2 |
| DA-077 | 07b8da2 |
| DA-078 | 67e1f3f |
| DA-079 | 07b8da2 |
| DA-080 | 07b8da2 |
| DA-081 | 07b8da2 |
| DA-082 | 07b8da2 |
| DA-083 | 07b8da2 (Meals, Recipes, RecipeForm, RecipeDetail); 67e1f3f (Ingredients, IngredientDetail, ShoppingLists, ShoppingListDetail, WishList) |
| DA-084 | 07b8da2 (recipe-selector, recipes, RecipesPage); 67e1f3f (Ingredients, ShoppingLists, WishList) |
| DA-085 | 07b8da2 |
| DA-086 | 07b8da2 (Meals, RecipeForm); 67e1f3f (IngredientDetail) |
| DA-087 | 07b8da2 |

## Frontend Audit — task-057 Stage 4 (Food)

- **Audit Scope:** `git diff cd06366..a68b611 -- frontend/` (25 files: 19 source, 5 tests, `index.css`). Recipes, meals, ingredients, shopping lists, wish list, cook mode.
- **Guidelines Source:** frontend-dev-guidelines skill + plan.md Global Constraints
- **Date:** 2026-10-19
- **Build:** PASS (`npm run build`, exit 0; chunk-size warning only)
- **Tests:** 933 passed, 0 failed (142 files). The runner is vitest, which rejects `--watchAll`, so `npx vitest run` was used. This includes `no-color-literals.test.ts` with the Stage 4 directories and pages removed from PENDING. ESLint could not be run: the local install fails loading its `stylish` formatter, so lint is unverified.
- **Overall:** NEEDS-WORK (one blocking finding, cheap)

### Blocking

1. `frontend/src/pages/IngredientDetailPage.tsx:213-219` — the alias-remove button is icon-only with a `h-3 w-3` (12 px) glyph and no padding, so its pointer target is about 12×12. Stage 4 edited this exact element (added the `aria-label`, line 214) and restyled its block. It breaks the Global Constraint "pointer targets ≥ 24×24, icon-only controls ≥ 44×44 under `coarse:`". Fix: add `inline-flex size-6 items-center justify-center coarse:size-11` to its className.

### Non-blocking

- FE-17 — no tests exist for `IngredientsPage` (95-line restructure), `IngredientDetailPage`, `category-manager`, `MealsPage`, `WishListPage`, `cook-mode` or `plan-item-popover`. A grep of `src` for their names in test files finds none. The edits are mostly markup, but `category-manager` gained several `aria-label`s and `icon-sm` sizes that a role/name test would pin. Stage 4 does add tests for `recipe-status`, `week-selector`, `ShoppingListsPage` and `ShoppingListDetailPage`.
- `frontend/src/pages/IngredientDetailPage.tsx:250-256` — the recipe-reference `<button>` is `text-sm` with no `min-h` floor, so it is about 20 px tall (below 24) and has no `coarse:min-h-11`. It is a list-row pointer target. Not touched by Stage 4, but sits in the restyled page.
- `frontend/src/components/features/meals/plan-item-popover.tsx:297-298,368-369` — `form: any` with an eslint-disable (FE-01). Pre-existing; the diff adds none. Leave.
- `frontend/src/components/features/meals/week-grid.tsx:79,95` — a `<td>` and a `<div>` carry `onClick` with `cursor-pointer` (FE-19 met) but are not keyboard-operable. Pre-existing; the remove `Button` inside is.
- `frontend/src/components/features/recipes/recipe-image-field.tsx:76` — `role="button"` div with `tabIndex`, `onKeyDown` and `cursor-pointer` (line 93). Compliant; noted only because it is the one non-native control.
- `frontend/src/components/features/recipes/recipe-ingredients.tsx:66` — `animate-spin` on the renormalize icon, gated on `isRenormalizing`. It is a button-bound spinner, so FE-05 allows it.

### Checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 no `any` | PASS for Stage 4 (pre-existing) | Diff adds none. Existing: `plan-item-popover.tsx:298,369` |
| FE-02 no class concatenation | PASS | No `className={\`` or `+` concatenation in changed files; `cn()` used (`ShoppingListDetailPage.tsx:290`, `recipe-image-field.tsx:92`) |
| FE-03 no direct client in components | PASS | No `@/lib/api/client` import in changed files |
| FE-04 no inline Zod | PASS | No `z.` in changed files |
| FE-05 no content spinners | PASS | Only `recipe-ingredients.tsx:66` (button-bound) |
| FE-06 no hardcoded colors | PASS | Diff grep for palette classes, hex, `rgb(`, `hsl(`, `oklch(` is empty; `no-color-literals.test.ts` passes with Stage 4 un-PENDING'd (`no-color-literals.test.ts:10-31`). New `index.css:310-330` utilities hold sizes only |
| FE-07 no state mutation | PASS | No `.push(`, `.splice(`, `.sort(` in added lines |
| FE-08 named exports | PASS | No `export default` in changed files |
| FE-09 tenant guard in hooks | PASS (n/a) | No hook files changed |
| FE-10 tenant in query keys | PASS (n/a) | No hook files changed |
| FE-11 `createErrorFromUnknown` | PASS | `category-manager.tsx:41,56,74`, `ingredient-resolver.tsx:41,50`, `IngredientDetailPage.tsx:62,72,83,96,107` each toast it |
| FE-12 to FE-16 | PASS (n/a) | No types, services, schemas or form wiring changed |
| FE-19 cursor-pointer | PASS | `week-grid.tsx:77,95`, `week-selector.tsx:56`, `IngredientsPage.tsx:198`, `ShoppingListDetailPage.tsx:292`, `recipe-image-field.tsx:93`; all other `onClick` sites are `<Button>`/`<button>` |
| FE-17 tests | WARN | See non-blocking; new tests at `recipe-status.test.tsx`, `week-selector.test.tsx`, `ShoppingListsPage.test.tsx`, `ShoppingListDetailPage.test.tsx` |
| FE-18 mocks | PASS (n/a) | No service changes |

Global Constraints spot-checks: Fraunces and `font-heading` are not on controls in the diff; Stage 4 status is icon-plus-text (`recipe-status.test.tsx`); `coarse:` floors are present on `recipe-selector.tsx:86`, `ingredient-resolver.tsx:81`, `week-selector.tsx:56`, `IngredientsPage.tsx:198`; `theme-provider.tsx`, `lib/dashboard/**` and `services/` are untouched.

### Deferred-minor triage

| Source | Finding | Verdict |
|--------|---------|---------|
| 20b (→ 21) | `ShoppingListDetailPage` shopping-mode row `coarse:min-h-11` shrank touch rows 64→44 px | Resolved by a68b611 (`ShoppingListDetailPage.tsx:291` is now `min-h-16`; test added) |
| 20b (→ 21) | `ShoppingListsPage` failed delete leaves dialog open, no feedback | Resolved in a68b611 as a test (`ShoppingListsPage.test.tsx:61-67`); the hook toasts the error. Fine |
| 21 | `ShoppingListsPage.test.tsx:58` delete-failure test passes pre-commit and does not assert the toast | Leave. It pins existing behaviour |
| 20a | `recipe-selector` / `ingredient-resolver` `focus-visible:outline-2` without `outline-offset` (`recipe-selector.tsx:86`, `ingredient-resolver.tsx:81`) | Fix before merge if convenient: add `focus-visible:outline-offset-2`, one token per line. Otherwise leave |
| 20a | `recipe-status.test` assertions pass pre-commit | Leave |
| 20a | `RecipeDetailPage` status icon inside `h1` | Leave (title remains the accessible name) |
| 20a | `RecipeFormPage:200` inner `border-b` divider | Leave (DA-086 met at one surface level) |
| 20a | DA-087 `border-input` instead of nonexistent `border-strong` | Leave (ruled) |
| 20b | keyboard test uses `.focus()`+Enter, not Tab | Leave |
| 20b | Ingredients uncategorized badge inside `h1` | Leave (disclosed) |
| 20b | archived-list Delete and `IngredientDetailPage` delete unconfirmed | Stage 7 candidate, as noted |

## Frontend Guidelines — task-057 Stage 5 (Time)

- **Scope:** `git diff origin/main...HEAD -- frontend/` (26 files; calendar feature, packages overlay, tasks/reminders cards, CalendarPage/TasksPage/RemindersPage, `lib/calendar/event-colors.ts`)
- **Build:** PASS (`npm run build`; only the pre-existing >500 kB chunk warning)
- **Tests:** 946 passed, 0 failed (146 files, `npx vitest run`; the `--watchAll=false` flag is not valid under vitest)
- **Lint:** 0 errors; 4 warnings (unused eslint-disable directives at `components/features/calendar/event-form-dialog.tsx:108,110,112,114`, file not changed in this stage)
- **Overall:** PASS (no blocking findings)

### Checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 no `any` | PASS | grep of changed files: no `: any` / `as any` |
| FE-02 no class concatenation | WARN | Template-string classNames remain in touched files, none added by this stage: `calendar-grid.tsx:80,87,138`, `connection-status.tsx:82` (pre-existing; stage only added `aria-label` at :79/:90) |
| FE-03 no direct API client in components | PASS | no `api/client` imports in changed files |
| FE-04 no inline Zod | PASS | no `z.` in changed files |
| FE-05 no content spinners | WARN | `calendar-selection-panel.tsx:46` inline `Loader2 animate-spin` on toggle pending; `connection-status.tsx:82` sync icon spin. Both pre-existing, small in-control pending indicators, not content loading |
| FE-06 no hardcoded colours | PASS | no palette classes/hex/rgb in changed non-test files; `calendar-grid.tsx:153,156` red-500 replaced by `border-primary`/`bg-primary`; `event-colors.ts` uses tokens; `no-color-literals.test.ts` now enforces calendar/reminders/tasks/pages (6 pending entries removed), and passes |
| FE-07 no state mutation | PASS | `.push/.sort` hits are on local arrays or copies (`calendar-utils.ts:186,363`, `calendar-widget.tsx:17` use `[...x].sort`), none unchanged-by-stage setState |
| FE-08 no default exports | PASS | none |
| FE-09/10 tenant guard / query keys | PASS (N/A) | no hooks changed |
| FE-11 error handling | PASS (N/A) | no `.catch(` added |
| FE-12/13/14/15/16 models, services, keys, forms, schemas | PASS (N/A) | no types, services, hooks, forms, or schemas changed |
| FE-19 cursor-pointer | PASS | `event-block.tsx:~30` chip, `all-day-event-row.tsx`, `month-day-cell.tsx:~68` (all `cursor-pointer`); `event-block.tsx:44` `span.contents` onClick is only a propagation guard, not an affordance; `calendar-grid.tsx:138` conditional `cursor-pointer` when write access; ui `Button` handles the rest |
| FE-17 tests | PASS | new `lib/calendar/__tests__/event-colors.test.ts`, `calendar-widget.test.tsx`, `event-popover.test.tsx`; Tasks/Reminders page and task-card tests updated; `eventChipColors` tests moved from calendar-utils test to its new module (not deleted) |
| FE-18 mocks | PASS (N/A) | no service interface changes |

### Global Constraints spot-check (design)

- Pointer targets: ui `Button` `sm`/`icon` variants carry `coarse:min-h-11`/`coarse:size-11` (`components/ui/button.tsx:26-35`); chips use `coarse:min-h-11` (`all-day-event-row.tsx`); reminder/task row buttons add `coarse:min-h-11 coarse:min-w-11` (`RemindersPage.tsx:156,159,164`, `TasksPage.tsx`).
- Icon-only buttons named (DA-091): `CalendarPage.tsx:298,315`, `RemindersPage.tsx:156-164`, `TasksPage.tsx` delete, `connection-status.tsx:79,90`.
- DA-089 popover is now a `Dialog` with `DialogTitle` (`event-popover.tsx`).
- Status not colour-only: package/busy chips carry text title; due dates carry "Due" label.
- Tailwind classes are literals; `eventChipColors` returns CSS-var strings via inline style, not assembled class names.

### Blocking (must fix)
- None.

### Non-Blocking (should fix)
1. FE-02: template-string classNames in `calendar-grid.tsx:80,87,138` and `connection-status.tsx:82` should use `cn()` (pre-existing).
2. FE-05: `animate-spin` at `calendar-selection-panel.tsx:46` and `connection-status.tsx:82` (pre-existing, inline control state; acceptable but not on a submit button).
3. Lint: 4 unused eslint-disable directives in `event-form-dialog.tsx:108-114` (out of stage scope).
4. Tests: no direct tests for `event-block.tsx` popover-click propagation guard (`:44`) or `month-day-cell.tsx` colour change; covered only indirectly.


## Plan Adherence — task-057 Stage 5 (Time)

**Range:** origin/main (fdabafa)..HEAD (776153b), branch task-057-impeccable-redesign-stage-5
**Verdict:** MOSTLY_COMPLETE, READY_TO_MERGE after the flagless gate passes. No blocking gaps. Two minor DA-088 sub-items (below) and two process notes.

### Task completion

| Task | Status | Evidence |
|---|---|---|
| 22 Calendar event colours | DONE (ruling: helper moved, not created) | `frontend/src/lib/calendar/event-colors.ts` (PACKAGE_EVENT_COLOR, ChipColors, eventChipColors with the planned busy/package/readableForeground order); `calendar-utils.ts` copy removed (-16 lines, no alias); tests in `lib/calendar/__tests__/event-colors.test.ts` cover every Step 2 case (busy non-owner/owner, package, #ffff00 to ink, #000000 to paper, "" and "teal" no throw) plus busy-over-package; `package-calendar-overlay.tsx` uses the sentinel and its test expects `PACKAGE_EVENT_COLOR`; all `#9ca3af` ternaries replaced (month-day-cell, event-block, all-day-event-row, month-event-chip). Commit 7a6cb14. Busy rule identical across files, so no behaviour change to note. |
| 23a Calendar (+CalendarPage) | DONE | Guard: calendar/, CalendarPage.tsx removed from PENDING (`no-color-literals.test.ts`); `#888` becomes `var(--calendar-busy)` (calendar-selection-panel.tsx:40); now-line `border-primary`/`bg-primary` (calendar-grid.tsx:153); PageHeader, `aria-label` Previous/Next period, "Today" trigger replaced by `Button` "Pick date" (behaviour/label change: name it in the PR body); aria-labels on Sync/Disconnect (connection-status.tsx:79,90); event-popover rebuilt on `Dialog` (DA-089); focus-visible ring and `transition-[filter]` on chips (DA-090); chip text-xs (DA-092). Commit fb9a87d. |
| 23b Tasks + reminders | DONE | Guard: tasks/, reminders/, TasksPage, RemindersPage removed from PENDING; PageHeader and EmptyState in both pages; `coarse:min-h-11` on row action buttons; `tabular-nums` and text-sm on due/time lines; aria-labels Snooze/Dismiss/Delete reminder and Delete task (DA-091); task-card test updated. Commit fba0231. Plan said "task/reminder rows coarse:min-h-11": applied to row buttons rather than rows, equivalent in effect. |
| 24 Remaining Stage 5 findings | DONE with caveats | Review-fix commits aaae07e (widget owner name visible, all-day chip target), f05d7b0, 162ad50 (Due copy unified), 776153b (busy-over-package test). Detector run on Stage 5 files: 0 findings (regex rules only; element rules not run, per progress.md). Caveats: (1) Step 4 convention "name the row ID in commit message" not followed; none of the 7 commits carry a DA-NNN, so the SHA table below is by file/diff mapping; (2) close-out step 1 was "not rendered" (no browser), so contrast for DA-088/093/097/099/100 is a manual check owed in the PR body. |

Tasks: 4/4 units done (22, 23a, 23b, 24). Skipped without approval: 0.

### Row findings DA-088..DA-100

All 13 rows resolved, with this exception:
- DA-088 (PARTIAL, minor, audit-permitted alternative): owner name is beside each swatch (user-legend.tsx:15-18, popover, widget after aaae07e), which satisfies the colour-only rule. Not done: the 1px `--border-strong` ring on legend/selection/CalendarPage swatches (user-legend.tsx and CalendarPage.tsx:~355 unchanged; the widget dot has `ring-1 ring-border`), and the month-cell `aria-label` (month-day-cell.tsx:65) still lists only date and count, not owners. Contrast of user-chosen hex against card therefore remains unverified for legend dots.
- DA-091, DA-095: DA-095's "44px targets when stacked at 390px" is only partly addressed (coarse:min-h-11 on Tasks/Reminders actions; no coarse sizing on CalendarPage toolbar buttons, `grep coarse CalendarPage.tsx` empty). Minor.

### Build and test

| Scope | Result |
|---|---|
| `npx vitest run` (lib/calendar, features/calendar, packages, tasks, reminders, pages, src/__tests__ incl. no-color-literals guard) | PASS, 47 files / 318 tests on re-run. A first run had 1 failing test, not reproducible on the second run while verify.sh was running concurrently (node_modules was transiently incomplete); treated as environmental. |
| `npx tsc -b` | PASS |
| `npx eslint` (calendar, lib/calendar, pages) | 0 errors, 4 pre-existing warnings (event-form-dialog unused disables) |
| Flagless `tools/verify.sh` | Not run here (in flight separately); required before merge |

### Action items
1. Merge only after flagless `tools/verify.sh` exits 0.
2. PR body: note "Today" to "Pick date" label change (DA-096) and the manual contrast check owed for DA-088/093/097/099/100.
3. Optional: add swatch ring and owner names to month-cell aria-label (DA-088), and coarse sizing on CalendarPage toolbar (DA-095).

### Row to SHA (DA-088..DA-100)

Commits do not cite DA IDs; mapping is by diff.

| Row | SHA | Note |
|---|---|---|
| DA-088 | aaae07e | Partial: owner names visible; ring and cell aria-label owners not done |
| DA-089 | fb9a87d | event-popover on Dialog |
| DA-090 | fb9a87d | focus ring on chips and cells (all-day chip target in aaae07e) |
| DA-091 | fb9a87d, fba0231 | CalendarPage/connection-status; Tasks/Reminders pages |
| DA-092 | fb9a87d | month-event-chip text-xs |
| DA-093 | 7a6cb14, fb9a87d | helper and package overlay; `#888` fallback |
| DA-094 | fb9a87d | now-line to primary |
| DA-095 | fb9a87d, fba0231 | PageHeader on all three pages (mobile 44px partial) |
| DA-096 | fb9a87d | Pick date Button |
| DA-097 | fba0231, 162ad50 | text-sm, "Due" prefix |
| DA-098 | fb9a87d, aaae07e | kiosk-scale time column; owner name |
| DA-099 | fba0231 | table secondary lines text-sm |
| DA-100 | fb9a87d | weekday/hour labels (hour label leading-tight) |
