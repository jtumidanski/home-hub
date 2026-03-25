# Sidebar Reorganization — Task Checklist

Last Updated: 2026-03-25

## Phase 1: Foundation — Install Components + Shared Config [S]

- [ ] 1.1 Install shadcn Collapsible component (verify @base-ui/react compatibility)
- [ ] 1.2 Install shadcn DropdownMenu component (or equivalent Menu/Popover from @base-ui/react)
- [ ] 1.3 Create `nav-config.ts` with shared NavGroup/NavItem types and group definitions
- [ ] 1.4 Create `use-nav-group-state.ts` hook for localStorage-persisted collapsed/expanded state
- [ ] 1.5 Verify new components render without errors in isolation

**Acceptance**: Shared config exports groups matching PRD table. Hook reads/writes localStorage. New UI primitives available.

## Phase 2: Collapsible Nav Groups — Desktop Sidebar [M]

- [ ] 2.1 Create `NavGroup` component (collapsible header with chevron + indented nav items)
- [ ] 2.2 Integrate NavGroup into `app-shell.tsx`, replacing flat nav list with grouped rendering
- [ ] 2.3 Implement active-route auto-expansion (group containing active NavLink is always open)
- [ ] 2.4 Add chevron rotation animation (CSS transition on expand/collapse)
- [ ] 2.5 Add ARIA attributes (`aria-expanded`, `aria-controls`) to group headers
- [ ] 2.6 Verify collapsed/expanded state persists across page navigation and browser refresh

**Acceptance**: Desktop sidebar shows 4 labeled collapsible groups. Active group always expanded. State persists in localStorage.

## Phase 3: Footer Restructure — Settings + User Menu [M]

- [ ] 3.1 Remove Settings from navGroups; render it as standalone nav item in sidebar footer above user container
- [ ] 3.2 Create `UserMenu` component (user info container that opens a popover/dropdown on click)
- [ ] 3.3 Add theme toggle item to UserMenu popover (Moon/Sun icon + label based on current theme)
- [ ] 3.4 Add sign out item to UserMenu popover (LogOut icon + destructive text)
- [ ] 3.5 Remove standalone theme toggle and sign out buttons from footer
- [ ] 3.6 Ensure popover has keyboard navigation and Escape-to-close
- [ ] 3.7 Integrate UserMenu into desktop sidebar footer

**Acceptance**: Settings in footer above user info. Tapping user info opens popover with theme toggle + sign out. Old standalone buttons removed.

## Phase 4: Mobile Drawer Parity [M]

- [ ] 4.1 Refactor `mobile-drawer.tsx` to import shared nav config (removes duplicated navItems)
- [ ] 4.2 Render collapsible NavGroups in mobile drawer (same component as desktop)
- [ ] 4.3 Add Settings nav item to mobile drawer footer
- [ ] 4.4 Add UserMenu to mobile drawer footer
- [ ] 4.5 Ensure tapping a nav item still closes the drawer
- [ ] 4.6 Verify popover works correctly within the drawer (positioning, z-index)
- [ ] 4.7 Confirm Weather nav item now appears on mobile (previously missing)

**Acceptance**: Mobile drawer is structurally identical to desktop sidebar. Nav item tap closes drawer. Weather present. No duplicate nav config.

## Phase 5: Polish and Verification [S]

- [ ] 5.1 Keyboard navigation audit: tab through all groups, items, and popover
- [ ] 5.2 Screen reader check: verify ARIA labels and roles are correct
- [ ] 5.3 Test collapsed state with active route change (navigate to item in collapsed group)
- [ ] 5.4 Test localStorage clear / first-visit (all groups should default to expanded)
- [ ] 5.5 Verify no visual layout shift on initial load from localStorage read
- [ ] 5.6 Cross-check all 12 PRD acceptance criteria
- [ ] 5.7 Run frontend build (`npm run build`) — no errors or warnings

**Acceptance**: All PRD acceptance criteria met. No regressions. Clean build.
