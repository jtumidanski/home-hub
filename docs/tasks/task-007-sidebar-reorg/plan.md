# Sidebar Reorganization — Implementation Plan

Last Updated: 2026-03-25

## Executive Summary

Reorganize the flat sidebar navigation into collapsible groups, move Settings to the footer, and consolidate theme toggle + sign out into a user menu popover. All changes apply to both desktop sidebar and mobile drawer, driven by a single shared nav configuration.

## Current State Analysis

### Navigation Structure
- **Desktop sidebar** (`app-shell.tsx`): 7 flat nav items (Dashboard, Tasks, Reminders, Recipes, Weather, Households, Settings) defined inline as `navItems` array
- **Mobile drawer** (`mobile-drawer.tsx`): 6 flat nav items (same minus Weather — likely a bug) with a duplicated `navItems` array
- Both use `NavLink` with identical active/inactive styling patterns

### Footer Area
- Theme toggle button (Moon/Sun icon + label)
- Sign out button (LogOut icon + destructive text)
- User info section (display name + email, truncated)
- Desktop and mobile have the same layout with slightly different padding/sizing

### Component Library
- shadcn/ui with `@base-ui/react` (not radix-ui)
- **Missing components**: Collapsible, DropdownMenu — need to be added
- Existing: Button, Card, Dialog, Form, Input, Label, RadioGroup, Select, Skeleton, Textarea, Badge

### Key Issues
1. Nav items duplicated between desktop and mobile
2. Mobile drawer is missing Weather nav item
3. No grouping or hierarchy in navigation
4. Footer has 3 separate elements consuming vertical space

## Proposed Future State

### Grouped Navigation
```
Home
  Dashboard
Productivity
  Tasks
  Reminders
Lifestyle
  Recipes
  Weather
Management
  Households
---footer---
Settings (nav item)
[User Name / Email] → popover (theme toggle, sign out)
```

### Architecture
```
nav-config.ts          — shared group/item definitions
nav-group.tsx          — collapsible group component (label + chevron + items)
user-menu.tsx          — user info container + popover (theme toggle, sign out)
app-shell.tsx          — consumes nav-config, renders groups + footer
mobile-drawer.tsx      — consumes nav-config, renders groups + footer
```

### State Management
- Collapsed/expanded state stored in `localStorage` under a single key (e.g., `sidebar-groups`)
- Custom hook `useNavGroupState` manages read/write/toggle
- Group containing active route is force-expanded regardless of persisted state

## Implementation Phases

### Phase 1: Foundation (Install Components + Shared Config)
Install missing shadcn components and create the shared navigation configuration.

### Phase 2: Collapsible Groups
Build the NavGroup component and integrate it into the desktop sidebar.

### Phase 3: Footer Restructure
Move Settings to footer, build UserMenu popover, remove standalone buttons.

### Phase 4: Mobile Drawer Parity
Apply all changes to mobile drawer using the same shared components.

### Phase 5: Polish and Verification
Accessibility audit, keyboard navigation, visual consistency, edge cases.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| @base-ui/react Collapsible API differs from radix-ui docs | Medium | Medium | Check @base-ui/react docs; may need manual Collapsible with CSS transitions |
| DropdownMenu not available in @base-ui/react | Medium | Medium | Use @base-ui/react Menu or Popover component instead |
| localStorage state conflicts with active-route expansion | Low | Low | Clear precedence rule: active route group always expanded |
| Mobile drawer popover positioning issues | Medium | Low | Test on small viewports; use `side="top"` anchoring |

## Success Metrics

- All 12 acceptance criteria from PRD pass
- No regressions in existing navigation (routing, active states, drawer open/close)
- Keyboard navigation works for all new interactive elements
- Desktop and mobile have identical logical structure

## Dependencies

- @base-ui/react Collapsible component (or CSS-based alternative)
- @base-ui/react Menu/Popover component for user menu
- lucide-react ChevronDown/ChevronRight icons for group indicators
