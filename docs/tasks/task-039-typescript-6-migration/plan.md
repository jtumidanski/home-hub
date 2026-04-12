# TypeScript 6.0 Migration — Implementation Plan

Last Updated: 2026-04-12

## Executive Summary

Migrate the Home Hub frontend from TypeScript 5.9 to 6.0, resolving all breaking changes (deprecated `baseUrl`, tightened module resolution), fixing an implicit `any` type error, and cleaning up pre-existing eslint errors/warnings to establish a zero-warning baseline. This is a toolchain-only change with no runtime behavior modifications. The eslint 9→10 upgrade is explicitly excluded.

## Current State Analysis

- **TypeScript**: `~5.9.3` in `frontend/package.json`
- **tsconfig.json**: Uses `baseUrl: "."` with `paths` for `@/*` alias. References `tsconfig.app.json` and `tsconfig.node.json`.
- **tsconfig.app.json**: Also has `baseUrl: "."` and `paths`. Uses `moduleResolution: "bundler"`, `module: "ESNext"`, `target: "ES2023"`. Includes `src/` only. Types limited to `["vite/client"]`.
- **tsconfig.node.json**: No `baseUrl`, no `paths`. Includes only `vite.config.ts`.
- **vite.config.ts**: Has its own `@` alias via `path.resolve(__dirname, "./src")` — unaffected by tsconfig changes.
- **eslint.config.js**: Flat config with `@eslint/js`, `typescript-eslint@^8.57.0`, `react-hooks`, `react-refresh`. No per-rule suppressions currently.
- **Pre-existing lint issues**: 3 errors (2 setState-in-effect, 1 unused var), ~8 warnings (react-hooks/incompatible-library + exhaustive-deps).

## Proposed Future State

- TypeScript `~6.0.0` with zero `tsc -b` errors
- No `baseUrl` in any tsconfig (future-proofed for TS 7.0 removal)
- All four checks pass: `tsc -b`, `vite build`, `vitest run`, `eslint .`
- Zero eslint errors, zero unacknowledged warnings
- Clean branch off `main` (no eslint 10 changes from Renovate PR)

---

## Phase 1: Branch Setup & Version Bump

**Goal**: Create a clean branch, bump TypeScript, establish error baseline.

### Task 1.1 — Create branch and bump TypeScript
- Create branch `task-039-typescript-6` off `main`
- Update `typescript` from `~5.9.3` to `~6.0.0` in `frontend/package.json`
- Run `npm install` — verify clean resolution (no `--legacy-peer-deps`)
- Verify `typescript-eslint@^8.57.0` peer dep compatibility (allows `<6.1.0`)
- **Effort**: S
- **Acceptance**: `npm install` exits 0 with no peer dep warnings for typescript

### Task 1.2 — Capture error baseline
- Run `tsc -b`, `vite build`, `vitest run`, `eslint .`
- Document exact error counts and locations
- **Effort**: S
- **Acceptance**: Baseline errors documented, no surprises beyond PRD expectations

---

## Phase 2: Remove Deprecated `baseUrl`

**Goal**: Eliminate TS5101 deprecation errors without breaking path aliases.

### Task 2.1 — Remove `baseUrl` from tsconfig files
- Remove `"baseUrl": "."` from `frontend/tsconfig.json`
- Remove `"baseUrl": "."` from `frontend/tsconfig.app.json`
- Keep `paths` mapping `"@/*": ["./src/*"]` in both files (works without `baseUrl` under `moduleResolution: "bundler"`)
- **Effort**: S
- **Depends on**: 1.1
- **Acceptance**: `tsc -b` no longer reports TS5101; `@/*` imports still resolve in IDE and build

### Task 2.2 — Verify Vite alias unaffected
- Confirm `vite build` succeeds — Vite's alias is path.resolve-based, independent of tsconfig `baseUrl`
- **Effort**: S
- **Depends on**: 2.1
- **Acceptance**: `vite build` exits 0

---

## Phase 3: Fix Module Resolution Errors (TS2305)

**Goal**: Resolve ~40 TS2305 errors where `@testing-library/react` exports are not found.

### Task 3.1 — Investigate and fix testing-library type resolution
- Check if `@testing-library/dom` needs to be an explicit devDependency (TS 6.0 may not resolve transitive types)
- If insufficient, check if `tsconfig.app.json` `types` array needs `@testing-library/jest-dom`
- If test files need different resolution: consider a `tsconfig.test.json` extending `tsconfig.app.json` with test-specific `types` entries, and update tsconfig.json references
- **Effort**: M
- **Depends on**: 2.1
- **Risk**: Medium — investigation may reveal unexpected resolution behavior
- **Acceptance**: `tsc -b` reports zero TS2305 errors for testing-library imports

---

## Phase 4: Fix Type Errors

**Goal**: Fix remaining compilation errors.

### Task 4.1 — Fix implicit `any` on `btn` parameter
- File: `src/pages/__tests__/HouseholdMembersPage.test.tsx:203`
- The `.filter((btn) => ...)` callback needs an explicit type: `HTMLElement`
- **Effort**: S
- **Depends on**: 3.1
- **Acceptance**: `tsc -b` exits 0 with zero errors

---

## Phase 5: Fix Pre-existing ESLint Issues

**Goal**: Zero errors, zero unacknowledged warnings from `eslint .`

### Task 5.1 — Fix setState-in-effect errors (2 instances)

**calendar-grid.tsx:357-368** — `RangeEditor` component:
- `useState` initialized from `initial` prop, then `useEffect` syncs on `[itemId, date, initial, min, max]`
- This is a controlled-value-from-props pattern. Options:
  - Use a `key` prop on the component to reset state when identity changes
  - Or derive state computation without effect (preferred if feasible)

**today-view.tsx:165-178** — `RangeInput` component:
- Same pattern: `useState` from `currentValue`, `useEffect` syncs on `[itemId, date, currentValue, min, max]`
- Same fix approach as above

- **Effort**: M
- **Depends on**: None (can be done in parallel with phases 2-4)
- **Acceptance**: `eslint .` reports zero errors for these files

### Task 5.2 — Fix unused variable error
- File: `src/pages/DataRetentionPage.tsx:150`
- Replace `catch (e)` with `catch (_e)` or `catch` (no binding)
- **Effort**: S
- **Depends on**: None
- **Acceptance**: `eslint .` reports zero errors for this file

### Task 5.3 — Suppress react-hooks/incompatible-library warnings
- ~7 instances in files using `useReactTable()` and `form.watch()`:
  - `src/components/common/data-table.tsx` (useReactTable)
  - `src/pages/RecipeFormPage.tsx` (form.watch)
  - `src/components/features/meals/plan-item-popover.tsx` (form.watch)
  - `src/components/features/calendar/event-form-dialog.tsx` (form.watch)
- These are library-level React Compiler compatibility warnings, not code bugs
- Add `// eslint-disable-next-line react-hooks/incompatible-library -- library-level incompatibility with React Compiler; no actionable fix` comments
- **Effort**: S
- **Depends on**: None
- **Acceptance**: `eslint .` reports zero warnings for this rule

### Task 5.4 — Investigate exhaustive-deps warning
- File: `src/components/features/tracker/calendar-grid.tsx:79` (inside `entryMap` useMemo)
- Check if the `entries` dependency is correctly specified
- Fix if genuine missing dep; suppress with justification if intentional
- **Effort**: S
- **Depends on**: None
- **Acceptance**: `eslint .` reports zero warnings for this location

---

## Phase 6: Final Verification

**Goal**: All four checks pass cleanly.

### Task 6.1 — Run full verification suite
- `tsc -b` — 0 errors
- `vite build` — succeeds
- `vitest run` — all tests pass, no new failures
- `eslint .` — 0 errors, 0 unacknowledged warnings
- **Effort**: S
- **Depends on**: All previous phases
- **Acceptance**: All four commands exit 0

### Task 6.2 — Create PR
- PR off `main`, title: "feat(frontend): migrate to TypeScript 6.0"
- Verify branch does not include eslint 10 changes
- **Effort**: S
- **Depends on**: 6.1
- **Acceptance**: PR created, CI passes

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Module resolution fix requires more than adding `@testing-library/dom` | Medium | Medium | Budget time for tsconfig.test.json approach; worst case is a separate test config |
| `paths` without `baseUrl` doesn't work as expected | Low | Low | `moduleResolution: "bundler"` documents this as supported; Vite alias is independent |
| react-hooks/incompatible-library warnings can't be suppressed cleanly | Low | Low | Per-line eslint-disable is always available as fallback |
| setState-in-effect refactor introduces subtle state bugs | Low | Medium | Existing tests cover these components; run vitest after changes |

## Success Metrics

- All four verification commands exit 0
- TypeScript version is `~6.0.0`
- No `baseUrl` in any tsconfig
- No runtime behavior changes
- Clean, reviewable PR with focused commits

## Required Resources

- Node.js / npm (already available)
- All changes confined to `frontend/` directory
- No backend service impact
- No infrastructure changes
