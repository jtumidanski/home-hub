# TypeScript 6.0 Migration — Context

Last Updated: 2026-04-12

## Key Files

### Configuration
- `frontend/package.json` — TypeScript version (`~5.9.3` → `~6.0.0`), devDependencies
- `frontend/tsconfig.json` — Root config with `baseUrl` and `paths`, references app + node configs
- `frontend/tsconfig.app.json` — App compilation config: `moduleResolution: "bundler"`, `baseUrl`, `paths`, types `["vite/client"]`, includes `src/`
- `frontend/tsconfig.node.json` — Node/Vite config, no `baseUrl`, includes only `vite.config.ts`
- `frontend/vite.config.ts` — Vite config with `@` path alias (path.resolve-based, independent of tsconfig)
- `frontend/eslint.config.js` — Flat eslint config, no per-rule suppressions currently

### Files Requiring Code Changes
- `frontend/src/components/features/tracker/calendar-grid.tsx` — Lines 357-368: setState-in-effect (RangeEditor). Line 79: exhaustive-deps warning (entryMap useMemo)
- `frontend/src/components/features/tracker/today-view.tsx` — Lines 165-178: setState-in-effect (RangeInput)
- `frontend/src/pages/DataRetentionPage.tsx` — Line 150: unused `e` in catch
- `frontend/src/pages/__tests__/HouseholdMembersPage.test.tsx` — Line 203: implicit `any` on `btn` parameter
- `frontend/src/components/common/data-table.tsx` — Line 27: useReactTable incompatible-library warning
- `frontend/src/pages/RecipeFormPage.tsx` — Line 42: form.watch incompatible-library warning
- `frontend/src/components/features/meals/plan-item-popover.tsx` — Line 212: form.watch incompatible-library warning
- `frontend/src/components/features/calendar/event-form-dialog.tsx` — Line 88: form.watch incompatible-library warning

## Key Decisions

1. **Remove `baseUrl` instead of suppressing** — TS 6.0 deprecates `baseUrl` (TS5101). Rather than using `"ignoreDeprecations": "6.0"`, we remove it outright. The `paths` mapping works without `baseUrl` when `moduleResolution` is `"bundler"`.

2. **Exclude eslint 10 upgrade** — The Renovate PR bundles eslint 9→10, but `eslint-plugin-react-hooks` doesn't have stable eslint 10 peer support yet. We take only the TypeScript 6.0 changes on a clean branch from `main`.

3. **Suppress library compatibility warnings** — `react-hooks/incompatible-library` warnings for `useReactTable()` and `form.watch()` are library-level React Compiler incompatibilities. These cannot be fixed in our code — suppress with eslint-disable comments.

4. **setState-in-effect fix approach** — Both `RangeEditor` (calendar-grid) and `RangeInput` (today-view) use the same pattern: useState from prop + useEffect to sync. Preferred fix: key-based reset or derived computation without effect.

## Dependencies

### Package Compatibility
- `typescript-eslint@^8.57.0` — Peer dep `typescript >=4.8.4 <6.1.0` ✓ compatible with TS 6.0
- `@testing-library/react@^16.3.2` — May need `@testing-library/dom` as explicit dep for TS 6.0 module resolution
- `vite@^8.0.1`, `vitest@^4.1.1`, `@vitejs/plugin-react@^6.0.1` — Should be unaffected

### Open Questions (from PRD)
1. Does TS 6.0 module resolution require `@testing-library/dom` as explicit devDep, or is a tsconfig adjustment sufficient? → To be determined in Phase 3
2. Can `react-hooks/incompatible-library` warnings be suppressed at eslint config level? → To be determined in Phase 5

## Branch Strategy
- New branch `task-039-typescript-6` off `main`
- Does NOT branch from `renovate/major-npm-dependencies` (which includes eslint 10 changes)
- PR targets `main`
