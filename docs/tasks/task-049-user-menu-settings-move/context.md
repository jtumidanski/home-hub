# Task-049 Context

Companion to `prd.md`, `design.md`, `plan.md`. Read this first; it summarizes the small handful of files this task touches and the conventions the plan assumes.

## Scope

Frontend-only UI consolidation. Move "Settings" out of the desktop sidebar / mobile drawer footers and into the user-profile pop-over (`UserMenu`), and drop the Light/Dark theme toggle from that pop-over. No backend, schema, or contract changes.

## Files Involved

All paths are relative to the worktree root `/.worktrees/task-049-user-menu-settings-move/`.

| File | Role | Change |
|------|------|--------|
| `frontend/src/components/features/navigation/user-menu.tsx` | Pop-over component (the only user-action menu) | Modify: add Settings item, remove theme toggle item, drop unused imports, add `useNavigate`, `Settings`, `settingsNavItem` imports |
| `frontend/src/components/features/navigation/app-shell.tsx` | Desktop sidebar layout | Modify: delete the bordered `NavLink` footer that linked to Settings; remove unused `Settings` icon import; remove `settingsNavItem` from the import line; check if `NavLink` is still used (it isn't elsewhere in this file — also remove) |
| `frontend/src/components/features/navigation/mobile-drawer.tsx` | Mobile drawer layout | Modify: delete the bordered `NavLink` footer that linked to Settings; remove unused `Settings` icon import; remove `settingsNavItem` import; remove `NavLink` import (no other use) |
| `frontend/src/components/features/navigation/nav-config.ts` | Source of truth for nav metadata | Unchanged. `settingsNavItem` (path `/app/settings`, label `Settings`, icon `Settings` from lucide-react) keeps its current shape and stays exported. Only its consumer changes. |
| `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx` | Existing sidebar test | Modify: drop the `expect(screen.getByText("Settings"))` assertion in `"renders sidebar with navigation links"`; delete the entire `"calls toggleTheme when theme button is clicked in user menu"` test; remove the `useThemeToggle` mock since the file no longer needs it |
| `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx` | New colocated test | Create: covers the pop-over contract — pop-over has only Settings + Sign Out items in that order, Settings calls `useNavigate("/app/settings")` and `onAction`, Sign Out calls `logout.mutate()` and `onAction`, no theme item exists, null user renders nothing |

## Key Conventions

- **Test runner:** `vitest` (not `jest`). Run from `frontend/` with `npm test` (which runs `vitest run`). Single-file: `npm test -- src/components/features/navigation/__tests__/user-menu.test.tsx`.
- **Imports:** `useNavigate` is imported from `react-router-dom`. The codebase already uses this pattern in `recipe-card.tsx`, `new-dashboard-modal.tsx`, `dashboard-kebab-menu.tsx`, and `weather-widget.tsx`. Match it.
- **Icon prop convention:** `iconSize` is a string of Tailwind classes (e.g. `"h-4 w-4"` default, `"h-5 w-5"` for the mobile drawer). Pass it through to the `Settings` icon inside the new menu item exactly the same way the current theme-toggle item passes it to `Moon`/`Sun`.
- **Pop-over item styling:** `MenuPrimitive.Item` already has two distinct class strings in the file:
  - Neutral: `flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground`
  - Destructive (Sign Out): `flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm text-destructive outline-none select-none hover:bg-accent focus:bg-accent`

  Settings uses the **neutral** styling. Sign Out keeps the destructive styling.
- **`onAction?.()` ordering:** Both items perform their primary effect first (`navigate(...)` for Settings, `logout.mutate()` for Sign Out), then call `onAction?.()`. Mirror the existing Sign Out pattern.

## Decisions Already Locked (from design.md)

1. **Settings navigation primitive: `useNavigate`** (imperative, called inside `onClick`). Rejected `<NavLink>` via `MenuPrimitive.Item`'s `render` prop to keep symmetry with the imperative Sign Out item.
2. **No separator** between Settings and Sign Out. The destructive color treatment on Sign Out is the visual differentiator.
3. **No active-state styling** inside the pop-over for Settings. The pop-over is an action surface; route-highlight is a sidebar concern.
4. **`useThemeToggle` hook untouched.** Only `user-menu.tsx`'s consumption of it goes away. The Settings page keeps using it.

## Out-of-Scope Files (explicitly do NOT modify)

- `frontend/src/lib/hooks/use-theme-toggle.ts`
- Anything under `frontend/src/pages/settings/` (the Settings page itself and its sub-screens)
- `frontend/src/components/providers/auth-provider.tsx`
- `frontend/src/lib/hooks/api/use-auth.ts` (logout mutation)
- Any Go service code
- Any infrastructure / Docker / migration / CI files

## Dependencies

No new npm packages. `react-router-dom` (`useNavigate`), `lucide-react` (`Settings`), `@base-ui/react/menu` are all already in `package.json`.

## Verification Targets

Before reporting completion, run from `frontend/`:

1. `npm run build` — must succeed with no unused-import warnings introduced in the modified files.
2. `npm test` — full vitest run must pass.

Per `CLAUDE.md` and saved feedback, **always run tests in addition to the build**; a passing `tsc -b` is not sufficient evidence the change is correct.

## Risks

The only non-obvious behavior risk is the `navigate(...); onAction?.()` ordering inside the mobile drawer. `onAction` is wired to `onClose`, which dismisses the drawer. The order-of-operations risk is covered explicitly in `design.md` §7.1 — `navigate` first, then `onAction?.()`, mirroring Sign Out. Do not reverse the order.
