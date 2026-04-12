# TypeScript 6.0 Migration Plan

## Phase 1: Setup

1. Create a new branch off `main`
2. Bump `typescript` from `~5.9.3` to `~6.0.0` in `frontend/package.json`
3. Run `npm install` and verify clean resolution
4. Run `tsc -b`, `vite build`, `vitest run`, `eslint .` to capture the full error baseline

## Phase 2: tsconfig — Remove `baseUrl`

1. Remove `baseUrl` from `frontend/tsconfig.json`
2. Remove `baseUrl` from `frontend/tsconfig.app.json`
3. Verify `paths` mapping `"@/*": ["./src/*"]` still resolves correctly without `baseUrl` (it should — `moduleResolution: "bundler"` supports rootless paths)
4. Verify Vite path alias in `vite.config.ts` is unaffected
5. Run `tsc -b` — confirm TS5101 deprecation error is gone

## Phase 3: Fix Module Resolution (TS2305 errors)

This is the largest area of work. TS 6.0 tightens how types are resolved from packages.

Investigation steps:
1. Check if `@testing-library/react`'s package.json `exports` / `types` field changed resolution behavior under TS 6.0
2. Try adding `@testing-library/dom` as an explicit devDependency (it provides the underlying types that `@testing-library/react` re-exports)
3. If that's insufficient, check if adjusting `moduleResolution` or adding `"types"` entries helps
4. If test files need different resolution than app files, consider a `tsconfig.test.json` that extends `tsconfig.app.json` with test-specific settings

Expected fix: Most likely adding `@testing-library/dom` as an explicit dep or adjusting the types configuration. TS 6.0 may have stopped implicitly resolving types through transitive dependencies.

## Phase 4: Fix Type Errors

1. `HouseholdMembersPage.test.tsx:204` — add explicit type to `btn` parameter

## Phase 5: Fix Pre-existing ESLint Issues

### Errors

1. **calendar-grid.tsx:362** — `setLocal(initial)` in effect
   - Investigate if this can be replaced with initializing state from props directly, or using a key-based reset pattern
2. **today-view.tsx:172** — `setLocal(currentValue)` in effect
   - Same approach — derive from props or use controlled state pattern
3. **DataRetentionPage.tsx:150** — unused `e` variable
   - Replace with `_e` or remove if the catch block doesn't need the error

### Warnings

1. **`react-hooks/incompatible-library`** (7 instances) — these are React Compiler compatibility warnings for `useReactTable()` and `form.watch()`. Since these are library-level limitations (not code bugs), the appropriate fix is eslint-disable comments with explanatory notes, or a config-level rule adjustment if the warnings are too noisy.
2. **`react-hooks/exhaustive-deps`** at `calendar-grid.tsx:79` — investigate the missing dependency. Fix if genuine, suppress with reason if intentional.

## Phase 6: Verification

Run all four checks and confirm:
- `tsc -b` — 0 errors
- `vite build` — builds successfully  
- `vitest run` — all tests pass
- `eslint .` — 0 errors, 0 warnings (or only intentionally suppressed)

## Risk Notes

- **Low risk**: `baseUrl` removal is straightforward since `moduleResolution: "bundler"` already supports `paths` without it
- **Medium risk**: Module resolution changes may require more investigation than expected if the fix isn't simply adding `@testing-library/dom`
- **Low risk**: ESLint fixes are isolated to specific files with clear error messages
