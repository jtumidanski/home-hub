# Sidebar Reorganization — Context

Last Updated: 2026-03-25

## Key Files

| File | Role | Changes Needed |
|------|------|---------------|
| `frontend/src/components/features/navigation/app-shell.tsx` | Desktop sidebar + layout shell | Remove inline navItems, import shared config, render groups + new footer |
| `frontend/src/components/features/navigation/mobile-drawer.tsx` | Mobile slide-out drawer | Remove inline navItems, import shared config, render groups + new footer |
| `frontend/src/components/features/navigation/mobile-header.tsx` | Mobile top bar | No changes expected |
| `frontend/src/components/ui/` | shadcn component directory | Add Collapsible + DropdownMenu (or equivalent) |

## New Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/components/features/navigation/nav-config.ts` | Shared navigation group/item definitions |
| `frontend/src/components/features/navigation/nav-group.tsx` | Collapsible group component (label + chevron + indented items) |
| `frontend/src/components/features/navigation/user-menu.tsx` | User info popover with theme toggle + sign out |
| `frontend/src/lib/hooks/use-nav-group-state.ts` | Custom hook for localStorage-persisted collapsed/expanded state |

## Key Decisions

1. **Component library**: Project uses `@base-ui/react`, not `@radix-ui/react`. shadcn CLI may generate radix imports — must verify and adjust if needed.
2. **Collapsible approach**: If @base-ui/react lacks a direct Collapsible primitive, use a simple `div` with CSS `max-height` / `grid-template-rows` transition and manual `aria-expanded` management.
3. **DropdownMenu vs Popover**: For the user menu, prefer whatever @base-ui/react provides. The PRD says "popover menu" — a Menu component with proper keyboard navigation and focus management is ideal.
4. **localStorage key**: Use `home-hub:sidebar-groups` to namespace and avoid collisions.
5. **Active route expansion**: Determined at render time by matching `location.pathname` against group items' `to` paths. This overrides persisted collapsed state for the active group only.

## Nav Group Configuration

```typescript
// Proposed shape
interface NavItem {
  to: string;
  icon: LucideIcon;
  label: string;
  end?: boolean;
}

interface NavGroup {
  key: string;        // localStorage key segment
  label: string;
  items: NavItem[];
}

// Groups (Settings excluded — it goes in footer)
const navGroups: NavGroup[] = [
  { key: "home",         label: "Home",         items: [dashboard] },
  { key: "productivity", label: "Productivity",  items: [tasks, reminders] },
  { key: "lifestyle",    label: "Lifestyle",     items: [recipes, weather] },
  { key: "management",   label: "Management",    items: [households] },
];

// Footer nav item
const settingsItem: NavItem = { to: "/app/settings", icon: Settings, label: "Settings", end: false };
```

## Dependencies Between Tasks

```
Phase 1 (foundation) → Phase 2 (groups) → Phase 4 (mobile parity)
Phase 1 (foundation) → Phase 3 (footer) → Phase 4 (mobile parity)
Phase 4 (mobile parity) → Phase 5 (polish)
```

Phases 2 and 3 can be done in parallel after Phase 1.

## Current Nav Items Comparison

| Item | Desktop (app-shell) | Mobile (mobile-drawer) |
|------|-------------------|----------------------|
| Dashboard | Yes | Yes |
| Tasks | Yes | Yes |
| Reminders | Yes | Yes |
| Recipes | Yes | Yes |
| Weather | Yes | **Missing** |
| Households | Yes | Yes |
| Settings | Yes | Yes |

The shared config will fix the missing Weather item on mobile.
