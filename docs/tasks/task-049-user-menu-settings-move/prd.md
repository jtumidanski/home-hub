# User Menu Settings Move — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-05-08
---

## 1. Overview

The application's primary navigation currently exposes a dedicated "Settings" link
in the footer area of both the desktop sidebar (`app-shell.tsx`) and the mobile
drawer (`mobile-drawer.tsx`). Immediately below that link sits the user-profile
control (`UserMenu`), whose pop-over today contains exactly two actions: a
Light/Dark theme toggle and Sign Out.

This task consolidates those surfaces. "Settings" moves into the user-profile
pop-over alongside Sign Out, where users already look for account-level actions.
The Light/Dark theme toggle is dropped from the pop-over because the same control
already exists on the Settings page itself, which is now one click deeper.

The change is UI-only and frontend-scoped. The `/app/settings` route, the
`SettingsPage` component, the underlying `useThemeToggle` hook, and the auth
provider/logout mutation are all untouched. No backend, contract, or storage
changes are required.

## 2. Goals

Primary goals:
- Remove the dedicated "Settings" footer link from the desktop sidebar and the
  mobile drawer.
- Add a "Settings" entry inside the user-profile pop-over, positioned above the
  "Sign Out" entry.
- Remove the Light/Dark theme toggle entry from the user-profile pop-over.
- Preserve a single source of truth for the Settings route metadata (path,
  label, icon) so it is not duplicated.
- Keep keyboard, focus, and screen-reader behavior of `UserMenu` working.

Non-goals:
- Redesigning the Settings page or any of its sub-screens.
- Moving or modifying the theme toggle that lives on the Settings page.
- Changing the visual style or layout of the user-profile pop-over beyond the
  necessary additions/removals.
- Reordering items inside the pop-over beyond placing "Settings" above
  "Sign Out".
- Changing routing, auth, or theme persistence behavior.
- Backend changes of any kind.

## 3. User Stories

- As a signed-in user, I want a single, predictable place to access my account
  settings, so that I don't have to scan a long sidebar to find them.
- As a signed-in user, I want the user-profile pop-over to focus on identity-
  scoped actions (Settings, Sign Out), so that account controls feel grouped.
- As a signed-in user on mobile, I want the drawer footer to stop wasting a row
  on a Settings link that overlaps with the profile menu, so that the drawer is
  more compact.
- As a signed-in user, I no longer need a quick theme toggle in the pop-over
  because I can change the theme on the Settings page itself.

## 4. Functional Requirements

### 4.1 User pop-over content (`user-menu.tsx`)

- The `<MenuPrimitive.Popup>` rendered for the user profile MUST contain the
  following items, in this order, top to bottom:
  1. **Settings** — navigates the user to `/app/settings` when activated.
  2. **Sign Out** — invokes the existing logout mutation.
- The Light/Dark theme toggle item (the one that currently calls
  `toggleTheme()` and renders a Sun/Moon icon) MUST be removed from the
  pop-over.
- The "Settings" item MUST:
  - Use `Settings` icon from `lucide-react`.
  - Use the same `MenuPrimitive.Item` styling currently applied to the theme
    toggle item (i.e., neutral foreground, hover/focus accent backgrounds —
    NOT the destructive styling reserved for Sign Out).
  - Trigger client-side navigation to `/app/settings` (using react-router's
    navigation API; not a full page reload).
  - Call `onAction?.()` after navigating, mirroring the existing pattern used
    by the Sign Out item, so that the mobile drawer continues to close itself
    when an item inside the pop-over is selected.
- The "Sign Out" item, its destructive styling, and its existing behavior MUST
  be preserved.
- Imports of `useThemeToggle`, `Moon`, and `Sun` SHOULD be removed from
  `user-menu.tsx` once they are unused.
- The `iconSize` prop MUST continue to apply to both pop-over items so the
  mobile drawer can pass `h-5 w-5`.

### 4.2 Desktop sidebar (`app-shell.tsx`)

- The footer block that renders a `NavLink` to `settingsNavItem.to` MUST be
  removed.
- The bordered container that wrapped that link MUST also be removed (so there
  is no empty bordered region left in the sidebar).
- The `UserMenu` component MUST remain rendered in the sidebar footer in its
  current position.
- Imports of `Settings` (icon) and `settingsNavItem` MUST be removed from
  `app-shell.tsx` if no longer referenced.

### 4.3 Mobile drawer (`mobile-drawer.tsx`)

- The footer block that renders a `NavLink` to `settingsNavItem.to` MUST be
  removed.
- The bordered container that wrapped that link MUST also be removed.
- `UserMenu` MUST remain rendered with the existing `onAction={onClose}` and
  `iconSize="h-5 w-5"` props so that selecting "Settings" or "Sign Out" inside
  the pop-over still closes the drawer.
- Imports of `Settings` (icon) and `settingsNavItem` MUST be removed from
  `mobile-drawer.tsx` if no longer referenced.

### 4.4 Nav config (`nav-config.ts`)

- The exported `settingsNavItem` (path, label, icon) MUST remain available as
  the single source of truth for the Settings entry, so that `user-menu.tsx`
  can import it instead of hard-coding the route, label, and icon.
- The visual nav-group structure (`navGroups`) MUST NOT change.

### 4.5 Theme toggle reachability

- After this task, the only user-facing theme toggle in the application MUST
  be the one that already exists inside `SettingsPage`.
- The `useThemeToggle` hook itself MUST NOT be modified or removed; other
  callers (Settings page, tests) keep using it unchanged.

### 4.6 Keyboard, focus, and accessibility

- The user-profile pop-over MUST continue to open with the same trigger
  affordance and keyboard behavior provided by `@base-ui/react/menu`.
- After the change, focus MUST land on the first item ("Settings") when the
  pop-over opens, matching prior behavior of the first menu item receiving
  focus.
- Both items inside the pop-over MUST be reachable and activatable via
  keyboard (arrow keys + Enter/Space) and via mouse/touch.

## 5. API Surface

No backend or HTTP API changes. No new client-side service modules.

The only "surface" change is in the React component contract for `UserMenu`,
which remains backward-compatible:

- `UserMenu` continues to accept the same props it does today
  (`onAction?: () => void`, `iconSize?: string`). No new props are required.

Internally, `UserMenu` will gain a dependency on react-router's `useNavigate`
(or equivalent navigation primitive already used elsewhere in the app) to drive
navigation to `/app/settings` from the new menu item. The exact navigation
primitive will be chosen during the design phase to match conventions used
elsewhere in this codebase.

## 6. Data Model

No data-model changes. No new tables, columns, migrations, or persisted
preferences.

The application's existing theme persistence (driven by `useThemeToggle` and
whatever storage it uses today) is unaffected.

## 7. Service Impact

| Service / surface           | Change                                                       |
|-----------------------------|--------------------------------------------------------------|
| `frontend` UI               | Modified (see Functional Requirements 4.1–4.4)               |
| All Go services             | No change                                                    |
| Database                    | No change                                                    |
| Build / deploy pipelines    | No change                                                    |

Files expected to change in `frontend/`:

- `frontend/src/components/features/navigation/user-menu.tsx` — modify
- `frontend/src/components/features/navigation/app-shell.tsx` — modify
- `frontend/src/components/features/navigation/mobile-drawer.tsx` — modify
- `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx` —
  update to reflect removed Settings footer link and any new pop-over content
- Possibly new/updated test for `UserMenu` to cover the Settings item

`nav-config.ts` is expected to stay structurally unchanged (still exports
`settingsNavItem`), although `settingsNavItem` will now be consumed by
`user-menu.tsx` rather than by the desktop sidebar / mobile drawer footers.

## 8. Non-Functional Requirements

### 8.1 Performance
- No measurable change. One fewer `NavLink` per layout; one extra
  `MenuPrimitive.Item` per pop-over open. Bundle size delta is negligible.

### 8.2 Security
- No change. No new endpoints, no new permissions, no new data flows.

### 8.3 Accessibility
- The pop-over MUST remain operable via keyboard alone (Tab to trigger, Enter
  to open, arrow keys to move between items, Enter/Space to activate).
- Focus order MUST be: Settings → Sign Out.
- No item MUST rely on color alone to convey its meaning; text labels stay
  required as today.

### 8.4 Responsiveness / multi-surface
- Desktop sidebar and mobile drawer MUST both reflect the change. There MUST
  NOT be a state where only one surface has been updated.
- Mobile drawer MUST continue to close itself after a pop-over item is
  activated, via the existing `onAction` callback wiring.

### 8.5 Multi-tenancy
- `UserMenu` already renders inside the authenticated app shell that resolves
  the current tenant context; this task does not interact with tenancy in any
  new way.

### 8.6 Observability
- No logging or analytics changes are required. If the project later wants to
  track Settings access, that is out of scope for this task.

## 9. Open Questions

1. **Navigation primitive** — Should the Settings menu item navigate via
   react-router's `useNavigate` from inside `UserMenu`, or should it render an
   `<a>`/`NavLink` styled as a `MenuPrimitive.Item`? Both patterns exist in the
   codebase; the design phase should pick the one already used by similar menu
   items (or fall back to `useNavigate` for consistency with the existing
   `logout.mutate()` pattern, which is also imperative).
2. **Visual separator** — Should there be a divider rule between Settings and
   Sign Out? Defaulting to "no divider" (matching today's pop-over, which has
   none between theme toggle and Sign Out), but the design phase may overrule
   this if the destructive Sign Out warrants a stronger visual break once it
   is no longer the only "different" item.
3. **Active-state styling** — The pop-over does not currently indicate when the
   user is already on `/app/settings`. Defaulting to "no active-state styling
   inside the pop-over" — the route highlight is a sidebar concern, and the
   pop-over is dismissed once activated. Design phase may revisit if it feels
   wrong.

## 10. Acceptance Criteria

- [ ] On desktop, the sidebar footer no longer contains a "Settings" link;
      only the user-profile control remains in the footer.
- [ ] On mobile, the drawer footer no longer contains a "Settings" link;
      only the user-profile control remains in the footer.
- [ ] Clicking/tapping the user-profile control opens a pop-over containing,
      in order: Settings, Sign Out (and nothing else).
- [ ] Activating "Settings" navigates the user to `/app/settings` without a
      full page reload.
- [ ] Activating "Settings" inside the mobile drawer also closes the drawer.
- [ ] Activating "Sign Out" still triggers the existing logout flow.
- [ ] The Light/Dark theme toggle no longer appears in the user-profile
      pop-over on either desktop or mobile.
- [ ] The theme toggle on `/app/settings` continues to work unchanged.
- [ ] The pop-over is fully operable via keyboard (open, navigate items,
      activate) and via screen reader (each item has a reachable, labelled
      role).
- [ ] `frontend` builds with no unused-import warnings introduced by this
      change in the modified files.
- [ ] Existing navigation tests are updated; tests that asserted the presence
      of the Settings footer link no longer assert it, and at least one test
      asserts that the Settings item is reachable inside the user pop-over.
- [ ] `npm test` passes in `frontend/`.
- [ ] No backend services are modified.
