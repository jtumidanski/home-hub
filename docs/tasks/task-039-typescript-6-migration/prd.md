# TypeScript 6.0 Migration — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-12
---

## 1. Overview

The Home Hub frontend currently uses TypeScript 5.9. TypeScript 6.0 introduces breaking changes to module resolution behavior and deprecates the `baseUrl` compiler option (with removal planned in TS 7.0). A Renovate PR (`renovate/major-npm-dependencies`) flagged this as a pending major update.

This task migrates the frontend to TypeScript 6.0 on a clean branch off `main`, fixing all compilation errors, resolving module resolution changes, and cleaning up pre-existing lint issues to establish a clean baseline going forward.

The eslint 9→10 upgrade from the same Renovate PR is explicitly excluded — it is blocked by `eslint-plugin-react-hooks` not yet having a stable release with eslint 10 peer support.

## 2. Goals

Primary goals:
- Upgrade `typescript` from `~5.9.3` to `~6.0.0`
- Fix all `tsc -b` compilation errors introduced by TS 6.0
- Remove deprecated `baseUrl` usage from all tsconfig files (future-proof for TS 7.0)
- Fix pre-existing eslint errors and warnings for a clean lint baseline
- All four checks pass cleanly: `tsc -b`, `vite build`, `vitest run`, `eslint .`

Non-goals:
- Upgrading eslint to v10 (blocked externally)
- Changing any runtime behavior — this is a toolchain-only change
- Upgrading any other dependencies beyond what TS 6.0 requires

## 3. User Stories

- As a developer, I want to be on the latest TypeScript major version so that I have access to new language features and stay within the supported version window.
- As a developer, I want zero lint errors on `main` so that CI lint checks are meaningful and new issues are immediately visible.

## 4. Functional Requirements

### 4.1 TypeScript Version Bump

- Update `typescript` in `frontend/package.json` from `~5.9.3` to `~6.0.0`
- Run `npm install` and verify clean dependency resolution (no `--legacy-peer-deps` needed)
- Verify `typescript-eslint@^8.57.0` is compatible (its peer dep allows `typescript >=4.8.4 <6.1.0`)

### 4.2 Remove Deprecated `baseUrl`

TS 6.0 deprecates `baseUrl` (error TS5101). Rather than suppressing with `"ignoreDeprecations": "6.0"`, remove `baseUrl` properly.

Files requiring changes:
- `frontend/tsconfig.json` — remove `baseUrl`, keep `paths` with `"@/*": ["./src/*"]`
- `frontend/tsconfig.app.json` — remove `baseUrl`, keep `paths` with `"@/*": ["./src/*"]`

The `paths` mapping `"@/*": ["./src/*"]` works without `baseUrl` when `moduleResolution` is `"bundler"` (which `tsconfig.app.json` already uses). Verify that Vite's path alias in `vite.config.ts` is unaffected.

### 4.3 Fix Module Resolution Errors

TS 6.0 tightens module resolution, causing ~40 TS2305 errors where `@testing-library/react` exports (`screen`, `waitFor`, `fireEvent`) are not found.

Investigate and resolve by:
1. Checking if `@testing-library/react` type definitions need updating
2. Checking if `tsconfig.app.json` module resolution settings need adjustment for test files (currently tests in `src/` are included via `"include": ["src"]`)
3. Checking if a separate `tsconfig.test.json` or types configuration is needed
4. Checking if `@testing-library/dom` needs to be added as an explicit dependency (it's a transitive dep that TS 6.0 may no longer resolve implicitly)

### 4.4 Fix Implicit `any` Error

- `src/pages/__tests__/HouseholdMembersPage.test.tsx:204` — parameter `btn` has implicit `any` type
- Add an explicit type annotation (likely `HTMLElement`)

### 4.5 Fix Pre-existing ESLint Errors (3 errors)

**setState-in-effect errors (2):**
- `src/components/features/tracker/calendar-grid.tsx:362` — `setLocal(initial)` called synchronously in effect
- `src/components/features/tracker/today-view.tsx:172` — `setLocal(currentValue)` called synchronously in effect

These need to be refactored to avoid synchronous setState inside effects (e.g., derive state from props/deps instead of syncing via effect, or use a ref).

**Unused variable error (1):**
- `src/pages/DataRetentionPage.tsx:150` — `'e'` is defined but never used

Remove or prefix with underscore per convention.

### 4.6 Fix Pre-existing ESLint Warnings (8 warnings)

All are `react-hooks/incompatible-library` warnings from React Hook Form's `watch()` and TanStack Table's `useReactTable()`. These are informational warnings from the React Compiler's compatibility checks.

If these cannot be resolved without significant refactoring (likely the case — they're library-level incompatibilities), suppress them with targeted eslint-disable comments that reference the specific rule and explain why.

One additional `react-hooks/exhaustive-deps` warning at `calendar-grid.tsx:79` should be investigated and fixed if a genuine missing dependency, or suppressed with justification if intentional.

## 5. API Surface

No API changes — this is a toolchain-only migration.

## 6. Data Model

No data model changes.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| frontend/ | tsconfig changes, package.json update, test file fixes, lint fixes |
| All others | None |

## 8. Non-Functional Requirements

- `npm install` must resolve cleanly without `--legacy-peer-deps` or `--force`
- `tsc -b` must exit 0 with no errors
- `vite build` must produce a working production bundle
- `vitest run` must pass all existing tests with no new failures
- `eslint .` must report 0 errors and 0 warnings (or only intentionally suppressed warnings)

## 9. Open Questions

1. Does TS 6.0's module resolution change require `@testing-library/dom` as an explicit devDependency, or is a tsconfig adjustment sufficient? (To be determined during implementation.)
2. Are the `react-hooks/incompatible-library` warnings suppressible at the eslint config level (per-rule disable for specific libraries) or do they need per-file comments? (To be determined during implementation.)

## 10. Acceptance Criteria

- [ ] `typescript` version is `~6.0.0` in package.json
- [ ] `npm install` resolves cleanly (no `--legacy-peer-deps`, no peer dep warnings for typescript)
- [ ] No `baseUrl` in any tsconfig file
- [ ] `@/*` path alias works in both IDE and build
- [ ] `tsc -b` exits with 0 errors
- [ ] `vite build` succeeds
- [ ] `vitest run` — all tests pass, no new failures
- [ ] `eslint .` — 0 errors, 0 unacknowledged warnings
- [ ] No runtime behavior changes — app functions identically before and after
- [ ] Branch is off `main`, does not include eslint 10 changes
