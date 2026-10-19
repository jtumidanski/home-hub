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
