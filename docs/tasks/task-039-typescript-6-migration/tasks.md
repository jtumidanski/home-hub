# TypeScript 6.0 Migration — Task Checklist

Last Updated: 2026-04-12

## Phase 1: Branch Setup & Version Bump

- [ ] **1.1** Create branch `task-039-typescript-6` off `main` [S]
- [ ] **1.1** Bump `typescript` from `~5.9.3` to `~6.0.0` in package.json [S]
- [ ] **1.1** Run `npm install` — verify clean resolution [S]
- [ ] **1.2** Run `tsc -b`, `vite build`, `vitest run`, `eslint .` — capture error baseline [S]

## Phase 2: Remove Deprecated `baseUrl`

- [ ] **2.1** Remove `baseUrl` from `frontend/tsconfig.json` [S]
- [ ] **2.1** Remove `baseUrl` from `frontend/tsconfig.app.json` [S]
- [ ] **2.1** Verify `@/*` path alias still resolves (tsc + IDE) [S]
- [ ] **2.2** Verify `vite build` succeeds [S]

## Phase 3: Fix Module Resolution Errors (TS2305)

- [ ] **3.1** Investigate `@testing-library/react` TS2305 errors [M]
- [ ] **3.1** Apply fix (add `@testing-library/dom` as explicit dep, or tsconfig.test.json, or types adjustment) [M]
- [ ] **3.1** Verify zero TS2305 errors from `tsc -b` [S]

## Phase 4: Fix Type Errors

- [ ] **4.1** Add explicit `HTMLElement` type to `btn` in `HouseholdMembersPage.test.tsx:203` [S]

## Phase 5: Fix Pre-existing ESLint Issues

### Errors
- [ ] **5.1** Fix setState-in-effect in `calendar-grid.tsx:357-368` (RangeEditor) [M]
- [ ] **5.1** Fix setState-in-effect in `today-view.tsx:165-178` (RangeInput) [M]
- [ ] **5.2** Fix unused `e` in `DataRetentionPage.tsx:150` [S]

### Warnings
- [ ] **5.3** Suppress `react-hooks/incompatible-library` in `data-table.tsx` [S]
- [ ] **5.3** Suppress `react-hooks/incompatible-library` in `RecipeFormPage.tsx` [S]
- [ ] **5.3** Suppress `react-hooks/incompatible-library` in `plan-item-popover.tsx` [S]
- [ ] **5.3** Suppress `react-hooks/incompatible-library` in `event-form-dialog.tsx` [S]
- [ ] **5.4** Investigate/fix `exhaustive-deps` warning at `calendar-grid.tsx:79` [S]

## Phase 6: Final Verification

- [ ] **6.1** `tsc -b` — 0 errors [S]
- [ ] **6.1** `vite build` — succeeds [S]
- [ ] **6.1** `vitest run` — all tests pass [S]
- [ ] **6.1** `eslint .` — 0 errors, 0 warnings [S]
- [ ] **6.2** Create PR targeting `main` [S]
