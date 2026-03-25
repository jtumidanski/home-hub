# Sidebar Reorganization — UX Flow

## Desktop Sidebar Layout

```
┌──────────────────────┐
│  Home Hub             │  ← App title / link
├──────────────────────┤
│  [Household Switcher] │
├──────────────────────┤
│                       │
│  ▾ Home               │  ← Collapsible group header
│    ○ Dashboard        │
│                       │
│  ▾ Productivity       │
│    ○ Tasks            │
│    ○ Reminders        │
│                       │
│  ▸ Lifestyle          │  ← Collapsed group
│                       │
│  ▾ Management         │
│    ○ Households       │
│                       │
├──────────────────────┤
│  ○ Settings           │  ← Footer nav item
├──────────────────────┤
│  John Doe          ▾  │  ← Tappable user container
│  john@example.com     │
└──────────────────────┘
```

## User Menu Popover (on user container tap)

```
┌──────────────────────┐
│  John Doe          ▾  │
│  john@example.com     │
├──────────────────────┤
│  ☽  Dark Mode         │
│  ⊳  Sign Out          │  ← destructive color
└──────────────────────┘
```

The popover anchors to the user container and opens upward (toward the top of the screen) since the container is at the bottom of the sidebar.

## Mobile Drawer Layout

Identical structure to the desktop sidebar, rendered inside the existing slide-in drawer panel. Tapping any nav item closes the drawer. The user menu popover works the same way.

## Collapsible Group Behavior

1. **Default state**: All groups expanded on first visit (no localStorage entry)
2. **Toggle**: Clicking/tapping a group header toggles that group
3. **Active override**: If a group contains the currently active route, it is forced expanded regardless of persisted state
4. **Persistence**: On each toggle, the full set of collapsed group keys is written to localStorage under a single key (e.g., `sidebar-collapsed-groups`)
5. **Restore**: On mount, collapsed state is read from localStorage and applied
