# Sidebar Reorganization — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25
---

## 1. Overview

The sidebar navigation currently presents all pages as a flat list. As the application grows with new pages, this list will become unwieldy and hard to scan. This feature reorganizes the sidebar into collapsible groups that categorize pages by domain, making navigation more scalable and intuitive.

Additionally, the footer area is restructured: the Settings nav item moves to the footer (above the user info container), and the theme toggle and sign out actions are consolidated into a popover menu triggered by tapping the user name/email container. These changes apply consistently to both the desktop sidebar and mobile drawer.

## 2. Goals

Primary goals:
- Organize nav items into labeled, collapsible groups for better scalability
- Persist collapsed/expanded state across navigation
- Move Settings to the sidebar footer as a standalone nav item
- Consolidate theme toggle and sign out into a user menu popover
- Maintain consistent behavior between desktop sidebar and mobile drawer

Non-goals:
- Adding new pages or routes
- Sidebar collapse to icons-only rail mode
- Customizable or reorderable sidebar items
- Search within the sidebar
- Profile page or user account management

## 3. User Stories

- As a user, I want navigation items grouped by category so that I can quickly find the page I need as more pages are added
- As a user, I want to collapse sidebar groups I don't use often so that the sidebar stays uncluttered
- As a user, I want my collapsed/expanded group preferences to persist as I navigate between pages
- As a user, I want to access theme toggle and sign out from a single user menu so that the sidebar footer is cleaner
- As a user, I want Settings in the footer area so that utility navigation is visually separated from feature navigation

## 4. Functional Requirements

### 4.1 Navigation Groups

The sidebar nav items are organized into the following groups:

| Group        | Items                |
|-------------|----------------------|
| Home        | Dashboard            |
| Productivity| Tasks, Reminders     |
| Lifestyle   | Recipes, Weather     |
| Management  | Households           |

- Each group displays a label header and contains its nav items indented beneath it
- Groups use the shadcn Collapsible component with a chevron indicator showing expanded/collapsed state
- Clicking the group header toggles the group open/closed
- The group containing the currently active route is always expanded regardless of persisted state
- Collapsed/expanded state is persisted in `localStorage` and survives page navigation and browser refresh

### 4.2 Settings in Footer

- The Settings nav item is removed from the main nav groups
- It appears in the sidebar footer area, directly above the user info container
- It renders as a standard nav item (icon + label) with the same active-state styling as group nav items

### 4.3 User Menu Popover

- The user info container (display name + email) in the sidebar footer becomes a tappable/clickable element
- Tapping it opens a popover menu (using shadcn DropdownMenu) anchored to the container
- The popover contains exactly two items:
  1. **Theme toggle** — shows "Dark Mode" or "Light Mode" with Moon/Sun icon based on current theme
  2. **Sign Out** — with LogOut icon, styled with destructive text color
- The standalone theme toggle button and sign out button currently in the footer are removed
- On mobile, the popover works the same way within the drawer

### 4.4 Shared Navigation Config

- A single nav configuration (groups, items, icons, routes) is defined once and consumed by both the desktop sidebar and mobile drawer
- This eliminates the current duplication of `navItems` arrays between `app-shell.tsx` and `mobile-drawer.tsx`

### 4.5 Mobile Drawer Parity

- The mobile drawer mirrors the desktop sidebar structure exactly:
  - Collapsible groups with the same grouping and labels
  - Settings in the footer above the user container
  - User menu popover for theme toggle and sign out
- Tapping a nav item still closes the drawer (existing behavior preserved)

## 5. API Surface

No API changes. This is a frontend-only feature.

## 6. Data Model

No data model changes. Collapsed/expanded group state is stored client-side in `localStorage`.

## 7. Service Impact

| Service   | Impact |
|-----------|--------|
| Frontend  | Primary changes — sidebar components restructured, new shadcn components installed |

No backend services are affected.

## 8. Non-Functional Requirements

- **Performance**: Collapsible animations should be smooth (use CSS transitions via shadcn Collapsible). localStorage reads on mount should not cause visible layout shifts.
- **Accessibility**: Collapsible group headers must be keyboard-navigable and use appropriate ARIA attributes (`aria-expanded`, `aria-controls`). The user menu popover must support keyboard navigation and Escape to close.
- **Responsiveness**: All changes must work correctly on both desktop (sidebar) and mobile (drawer) viewports.

## 9. Open Questions

None — all decisions have been confirmed.

## 10. Acceptance Criteria

- [ ] Nav items are organized into four labeled groups: Home, Productivity, Lifestyle, Management
- [ ] Each group is collapsible via its header with a chevron indicator
- [ ] Collapsed/expanded state persists in localStorage across navigation and page refresh
- [ ] The group containing the active route is always expanded
- [ ] Settings appears as a nav item in the sidebar footer, above the user container
- [ ] Tapping the user name/email container opens a popover with theme toggle and sign out
- [ ] The standalone theme toggle and sign out buttons are removed from the footer
- [ ] Desktop sidebar and mobile drawer have identical structure and behavior
- [ ] Nav config is defined once and shared between desktop and mobile
- [ ] shadcn Collapsible and DropdownMenu components are installed and used
- [ ] All interactions are keyboard-accessible with proper ARIA attributes
- [ ] No regressions in existing navigation (routing, active states, drawer open/close)
