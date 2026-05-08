# User Menu Settings Move — Design

Version: v1
Status: Approved-pending-review
Created: 2026-05-08

Companion to: `prd.md`

---

## 1. Scope

Frontend-only UI consolidation. Three files change in
`frontend/src/components/features/navigation/`:

- `user-menu.tsx` — gain a Settings item, lose the theme-toggle item.
- `app-shell.tsx` — drop the desktop sidebar's Settings footer link.
- `mobile-drawer.tsx` — drop the mobile drawer's Settings footer link.

Plus the test file `__tests__/app-shell.test.tsx`, which currently asserts the
removed Settings footer link and the removed Dark Mode menu item — both
assertions go away. A new `user-menu.test.tsx` will cover the new pop-over
contents.

`nav-config.ts` is touched only as a consumer relocation: `settingsNavItem`
remains exported and structurally unchanged; the only difference is which file
imports it.

No backend, schema, contract, or build-pipeline changes.

## 2. Architecture

The change reshapes a small piece of one component tree:

```
AppShell                            MobileDrawer
└── <aside>                         └── <div role="dialog">
    ├── <nav> (groups)                  ├── <nav> (groups)
    ├── (REMOVED) Settings NavLink      ├── (REMOVED) Settings NavLink
    └── UserMenu                        └── UserMenu (onAction=onClose)
        └── MenuPrimitive.Popup             └── MenuPrimitive.Popup
            ├── (NEW) Settings item             ├── (NEW) Settings item
            ├── (REMOVED) Theme toggle          ├── (REMOVED) Theme toggle
            └── Sign Out                        └── Sign Out
```

Both surfaces converge on a single source of truth — `UserMenu` — for the
user-scoped actions. The two outer surfaces (sidebar, drawer) become simpler
and their footers shrink to a single block (`UserMenu` only, no bordered
Settings link above it).

`UserMenu` keeps the same external contract: `onAction?: () => void` (called
after every item activation; the drawer wires this to `onClose`) and
`iconSize?: string` (passed through to both icons inside the pop-over).

## 3. Resolved Open Questions

The PRD listed three open questions for the design phase. Each is resolved here.

### 3.1 Navigation primitive — `useNavigate`

**Decision:** Use react-router's `useNavigate` hook imperatively from the
Settings menu item's `onClick`, matching the imperative pattern the existing
Sign Out item already uses with `logout.mutate()`.

**Why not a `NavLink` rendered through `MenuPrimitive.Item`'s `render` prop?**
Two reasons:

1. **Pattern symmetry with the sibling item.** Sign Out is imperative
   (`logout.mutate(); onAction?.()`). A declarative `<NavLink>` for Settings
   would create two different interaction models inside the same five-line
   pop-over — readers would have to context-switch between them.
2. **`onAction?.()` ordering is explicit.** The mobile drawer relies on
   `onAction` firing after activation so it can close itself. With
   `useNavigate`, the sequence is one expression: `navigate(...); onAction?.()`.
   With a `NavLink` render prop, closing the drawer would have to hook into
   `onClick` separately — the same imperative work, but split across two
   layers.

The cost of `useNavigate` is one extra hook call inside `UserMenu`. The
component already calls four hooks (`useAuth`, `useThemeToggle`, `useLogout`,
plus the menu primitive's internal state); adding `useNavigate` and dropping
`useThemeToggle` keeps that count flat.

```tsx
const navigate = useNavigate();
// ...
<MenuPrimitive.Item
  className="..." // same neutral styling currently applied to the theme toggle
  onClick={() => {
    navigate(settingsNavItem.to);
    onAction?.();
  }}
>
  <Settings className={iconSize} />
  {settingsNavItem.label}
</MenuPrimitive.Item>
```

### 3.2 Visual separator — none

**Decision:** No `MenuPrimitive.Separator` (or hand-rolled `<hr>`) between
Settings and Sign Out.

**Why:** The two items are already visually differentiated by Sign Out's
`text-destructive` color treatment, which is preserved verbatim. Today's
pop-over has no separator between Theme and Sign Out and the destructive
styling alone has been sufficient. Adding a separator now would introduce a
new visual element that the PRD explicitly says is out of scope ("Changing the
visual style or layout of the user-profile pop-over beyond the necessary
additions/removals"). If a future task wants to harden destructive-action
affordances (e.g., a confirmation step or visual rule), it can do so as its
own change.

### 3.3 Active-state styling inside the pop-over — none

**Decision:** The Settings item does not carry any active/current-route
styling when the user is on `/app/settings`.

**Why:** The pop-over closes the moment the user activates an item, so an
active-state badge would only be visible during the brief interval between
opening the menu and clicking. Sidebar/drawer route highlights are the
correct surface for "you are here" feedback; the pop-over is for actions, not
location indication. This also matches the Sign Out item, which has no
"active" concept.

## 4. Behavior Details

### 4.1 Item ordering and focus

The `MenuPrimitive.Popup` will render Settings first, then Sign Out. base-ui's
menu places initial focus on the first focusable item when the pop-over opens
via keyboard, so Settings receives focus on open — matching what the
theme-toggle item does today (it is currently the first item).

Mouse-open behavior is unchanged: focus does not auto-jump to an item on a
pointer-driven open; arrow keys still move between Settings and Sign Out.

### 4.2 `onAction?.()` semantics

Both items continue to call `onAction?.()` after their primary effect. In the
sidebar (`AppShell`), `UserMenu` is rendered with no props, so `onAction` is
undefined and the optional call is a no-op. In the mobile drawer,
`onAction={onClose}` ensures the drawer dismisses after Settings is activated
(the user has already been navigated away, so leaving the drawer open would
be a stale UI state) and after Sign Out is activated (the user is being
logged out; the drawer closing is essential since the protected route will
re-render).

### 4.3 Removed theme toggle — fallout

Removing the theme item makes three imports in `user-menu.tsx` unused:
`Moon`, `Sun`, and `useThemeToggle`. All three are deleted. The
`useThemeToggle` *hook* itself is untouched and remains in use by
`SettingsPage`.

The `theme` variable read from `useThemeToggle()` was used for the
"Light Mode"/"Dark Mode" label flip and the icon swap; both go away with the
item.

### 4.4 Sidebar and drawer footer simplification

Both surfaces today have two footer blocks separated by `border-t`:

```
┌──────────────┐
│ NavLink: Settings (border-t, p)
├──────────────┤
│ UserMenu        (border-t)
└──────────────┘
```

After the change, each becomes a single block:

```
┌──────────────┐
│ UserMenu        (border-t)
└──────────────┘
```

The `border-t` on the surviving `UserMenu` wrapper stays — it provides the
visual separation from the scrolling nav region above. No new wrapper is
introduced.

### 4.5 Imports cleanup

Each of the three modified files loses imports:

| File | Removed imports |
|------|------------------|
| `user-menu.tsx` | `Moon`, `Sun` (from `lucide-react`); `useThemeToggle` (from `@/lib/hooks/use-theme-toggle`) |
| `user-menu.tsx` | (added) `Settings` (from `lucide-react`); `useNavigate` (from `react-router-dom`); `settingsNavItem` (from `@/components/features/navigation/nav-config`) |
| `app-shell.tsx` | `Settings` (from `lucide-react`); `settingsNavItem` (from nav-config); `NavLink` if no other code path uses it |
| `mobile-drawer.tsx` | `Settings` (from `lucide-react`); `settingsNavItem` (from nav-config); `NavLink` if no other code path uses it |

`NavLink` removal must be verified file-by-file; both files use it only for
the Settings footer link, so both removals should land. `Link` (used for the
"Home Hub" branding in `app-shell.tsx`) is unaffected.

## 5. Alternatives Considered

### Alt A — Render Settings as `MenuPrimitive.Item` + `<NavLink>`

base-ui menu items support a `render` prop, e.g.
`<MenuPrimitive.Item render={<NavLink to={...} />}>...</MenuPrimitive.Item>`.
This is more "react-router idiomatic" — middle-click and Cmd-click would
open the route in a new tab, which `useNavigate` cannot do.

**Rejected because** the user-menu pop-over is a dismissible action surface,
not a primary navigation list. The desktop sidebar and the in-app routes
already accommodate "open in new tab" behavior for users who want it. Mixing
declarative and imperative items in the same five-item pop-over would harm
readability for a benefit (new-tab open from a pop-over) that nothing in the
PRD or user feedback asks for.

### Alt B — Keep the theme toggle, just add Settings

Leaves three items in the pop-over: Theme, Settings, Sign Out.

**Rejected** by the PRD itself ("Remove the Light/Dark theme toggle entry
from the user-profile pop-over"). Recording the rejection here only to make
explicit that the design did not silently undo a PRD decision.

### Alt C — Move Settings into a sub-menu (e.g., a "Account" group with sub-items)

base-ui supports nested menus. Could host Settings, Theme, Sign Out, and
future items under one expandable group.

**Rejected because** the pop-over has two items. Sub-menus add structural
complexity without payoff at this size, and the PRD explicitly limits scope
to "necessary additions/removals."

## 6. Testing Strategy

Two test files are touched.

### 6.1 `app-shell.test.tsx` — update

Today the file has two tests that the change invalidates:

- `"renders sidebar with navigation links"` — asserts
  `screen.getByText("Settings")`. After the change, "Settings" text no
  longer appears anywhere in the rendered sidebar (the pop-over is closed by
  default). This assertion is removed.
- `"calls toggleTheme when theme button is clicked in user menu"` — opens
  the pop-over and clicks Dark Mode. After the change, no Dark Mode item
  exists. This test is deleted.

Tests that remain unchanged:

- household switcher renders
- user info renders
- outlet renders
- Sign Out item still calls `logout.mutate()` (the existing assertion is
  preserved verbatim; the click target is still the Sign Out menu item)
- null-user case still renders nothing

The `useThemeToggle` mock can stay in the file (harmless if unused) or be
removed; preference is to remove it since the file no longer references the
hook, keeping the mock surface honest.

### 6.2 `user-menu.test.tsx` — new

A new colocated test file under
`frontend/src/components/features/navigation/__tests__/user-menu.test.tsx`
covers `UserMenu` directly. Mocks `useAuth`, `useLogout`, and react-router's
`useNavigate`. Cases:

1. **Pop-over opens with both items** — open the menu, assert that exactly
   two `menuitem` roles exist and they read "Settings" then "Sign Out".
2. **Settings item navigates and triggers `onAction`** — render with
   `onAction={mockOnAction}`, click Settings, assert `mockNavigate` was
   called with `/app/settings` and `mockOnAction` was called.
3. **Sign Out item logs out and triggers `onAction`** — same pattern,
   assert `mockLogoutMutate` and `mockOnAction`.
4. **No theme toggle item** — assert `screen.queryByRole("menuitem", { name: /dark mode|light mode/i })` returns `null`.
5. **Null user renders nothing** — preserves the existing early-return guard.

This file owns the pop-over contract going forward, which lets the
`AppShell` test file shrink to "the sidebar wraps a `UserMenu`" rather than
re-asserting `UserMenu` internals.

### 6.3 What is intentionally not tested

- Visual styling (Tailwind classes) — covered by manual review.
- base-ui menu internals (focus ring, escape-to-close) — owned by the
  library, not by this change.
- `useThemeToggle` behavior — out of scope; the hook is untouched and the
  Settings page's tests cover it.
- Multi-tenancy — the component does not take a tenant input.

## 7. Risks and Edge Cases

### 7.1 Drawer closes before navigation completes

`onClose` (the drawer dismiss) runs synchronously after `navigate(...)`. In
react-router, `navigate(path)` updates the history stack synchronously and
the route render happens on the next React commit. The drawer's `onClose`
unmounting and the route render race against each other on the same commit
batch — not a problem in practice (react-router handles this case throughout
the codebase, e.g., the existing nav-group `onItemClick={onClose}` pattern
in `mobile-drawer.tsx`). Mitigation: order is `navigate` first, then
`onAction?.()`, mirroring the order used by the Sign Out item.

### 7.2 Already on `/app/settings` when activated

`navigate("/app/settings")` from `/app/settings` is a no-op route-wise but
still triggers a render. The pop-over still closes (via `onAction`); the
mobile drawer still closes. No special-casing needed.

### 7.3 Pop-over stays open if `navigate` throws

`useNavigate` does not throw under normal use; the only realistic failure is
the hook being called outside a `<Router>`, which is a build-level mistake
that would crash the app on first render, not silently leave the pop-over
open. No mitigation required.

### 7.4 Stale references to the old footer link

The PRD mandates removing imports and the bordered container. A grep across
the modified files for `settingsNavItem` (in `app-shell.tsx`,
`mobile-drawer.tsx`) and for `Settings` (the icon import) is the lint of
record. CI's TypeScript build will fail on unused imports if any are missed,
which is the desired safety net.

## 8. Acceptance-Criteria Mapping

Every PRD acceptance-criterion line maps to a place in the design:

| PRD AC | Design coverage |
|--------|------------------|
| Sidebar footer no longer has Settings link | §2 architecture diff; §4.4 footer simplification |
| Drawer footer no longer has Settings link | §2 architecture diff; §4.4 footer simplification |
| Pop-over contains, in order: Settings, Sign Out | §4.1 item ordering and focus |
| Settings activates → `/app/settings`, no full reload | §3.1 `useNavigate` choice; §4.2 `onAction` semantics |
| Settings inside drawer also closes drawer | §4.2 `onAction` semantics |
| Sign Out still triggers logout flow | §2; §4.2 (sign out path unchanged) |
| Theme toggle absent from pop-over | §3.1; §4.3 fallout |
| Theme toggle on Settings page still works | out-of-scope guarantee §1; `useThemeToggle` untouched §4.3 |
| Pop-over operable via keyboard and screen reader | §4.1 (base-ui handles keyboard; aria roles preserved) |
| `frontend` builds with no unused-import warnings | §4.5 imports cleanup |
| Existing nav tests updated; pop-over coverage exists | §6.1, §6.2 |
| `npm test` passes | §6.1, §6.2 |
| No backend services modified | §1 scope statement |

## 9. Out of Scope (restated)

- Replacing or restyling Sign Out's destructive treatment.
- Adding a new "Profile" or "Account" item to the pop-over.
- Reorganizing `nav-config.ts`'s group structure.
- Adding telemetry on Settings access.
- Touching `useThemeToggle` or the Settings page itself.
- Any non-frontend change.

## 10. Implementation Sequencing (for the Plan Phase)

The plan phase will sequence this, but the natural ordering is:

1. Update `user-menu.tsx` (most isolated change, has its own test file).
2. Add `user-menu.test.tsx` and verify it green.
3. Update `app-shell.tsx` and its test file.
4. Update `mobile-drawer.tsx`.
5. Run `npm run build` and `npm test` in `frontend/`.

No service rebuilds, no Docker work, no migrations.
