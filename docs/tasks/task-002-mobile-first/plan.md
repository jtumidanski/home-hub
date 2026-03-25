# Mobile-First Responsive UI — Implementation Plan

Last Updated: 2026-03-25

---

## Executive Summary

Retrofit the Home Hub frontend to support mobile as a first-class citizen. The current desktop-only layout (fixed 256px sidebar, DataTable-based lists) becomes unusable below ~768px. This plan introduces a hamburger-driven navigation drawer, card-based mobile data views, a full-screen household selector, pull-to-refresh, and touch-friendly sizing — all while preserving the existing desktop experience unchanged.

The work is entirely frontend (React, Tailwind, TypeScript). No backend or API changes. No new major dependencies. The implementation is structured in 5 phases, ordered to deliver value incrementally: foundation/hooks first, then navigation, then page-by-page responsive layouts, then polish.

---

## Current State Analysis

### What exists
- **AppShell**: Fixed two-column flex layout (`w-64` sidebar + `flex-1` main). No responsive classes.
- **HouseholdSwitcher**: Compact `Select` dropdown in sidebar. Hidden when only 1 household.
- **DataTable**: Generic `@tanstack/react-table` wrapper used by Tasks, Reminders, Households pages. Column definitions defined inline in each page file.
- **Dashboard**: `md:grid-cols-3` card grid — the only existing responsive pattern.
- **Login/Onboarding**: Centered `max-w-md` cards — mostly mobile-friendly but untested at 320px.
- **Dialogs**: Already have responsive sizing (`sm:max-w-sm`, mobile padding).
- **Skeleton loading**: Present on all pages — preserved as-is.
- **Theme**: Light/dark mode via CSS custom properties and `.dark` class.
- **Button sizes**: `xs` (h-6/24px), `sm` (h-7/28px), `default` (h-8/32px), `lg` (h-9/36px). All below 44px touch target minimum.
- **Viewport meta**: Correctly set (`width=device-width, initial-scale=1.0`).

### What's missing
- No hamburger menu, mobile header, or navigation drawer.
- No card-based alternative to DataTable for small screens.
- No pull-to-refresh mechanism.
- No viewport detection hook.
- No mobile-specific household selector.
- Button/input tap targets below recommended 44px minimum.
- Page padding (`p-6`) does not adapt to smaller screens.

---

## Proposed Future State

```
Mobile (<768px)                        Desktop (≥768px)
┌──────────────────┐                   ┌──────┬────────────┐
│ [☰] Home Hub     │                   │ Side │  Content   │
├──────────────────┤                   │  bar │  (unchanged│
│ Pull-to-refresh  │                   │ 256px│  from      │
│ ┌──────────────┐ │                   │      │  today)    │
│ │ Card view    │ │                   │      │            │
│ │ with actions │ │                   └──────┴────────────┘
│ └──────────────┘ │
│ ┌──────────────┐ │
│ │ Card view    │ │
│ └──────────────┘ │
└──────────────────┘
```

- **Navigation**: Hamburger icon → slide-in drawer with all sidebar content.
- **Data views**: Card lists replace DataTables below `md`.
- **Household switcher**: Full-screen overlay on mobile.
- **Pull-to-refresh**: Touch-triggered refetch on Dashboard, Tasks, Reminders.
- **Touch targets**: All interactive elements ≥44px touch area on mobile.
- **Spacing**: Responsive padding (`p-4` mobile, `p-6` desktop).

---

## Implementation Phases

### Phase 1: Foundation (S)

Establish the responsive infrastructure that all subsequent phases depend on.

**1.1 Create `use-mobile` hook**
- File: `frontend/src/lib/hooks/use-mobile.ts`
- Uses `window.matchMedia("(max-width: 767px)")` with event listener for reactive updates.
- Returns `isMobile: boolean`.
- Effort: S
- Dependencies: None

**1.2 Create `use-pull-to-refresh` hook**
- File: `frontend/src/lib/hooks/use-pull-to-refresh.ts`
- Manages touch start/move/end events, pull distance tracking, threshold detection.
- Calls a provided `onRefresh` callback (async, returns when done).
- Returns `{ pullDistance, isRefreshing, handlers }` for binding to a container.
- Effort: M
- Dependencies: None

**1.3 Create `PullToRefresh` wrapper component**
- File: `frontend/src/components/common/pull-to-refresh.tsx`
- Wraps page content. Shows spinner indicator when pulling/refreshing.
- Only active when `isMobile` is true (from `use-mobile` hook).
- Props: `onRefresh: () => Promise<void>`, `children`.
- Effort: M
- Dependencies: 1.1, 1.2

### Phase 2: Navigation (M)

Replace the desktop sidebar with a responsive navigation system.

**2.1 Create `MobileHeader` component**
- File: `frontend/src/components/features/navigation/mobile-header.tsx`
- Renders: hamburger button (left), "Home Hub" branding (center/left), page title or user avatar (right).
- Visible only below `md` (`md:hidden`).
- Hamburger button calls `onMenuOpen` prop.
- Effort: S
- Dependencies: None

**2.2 Create `MobileDrawer` component**
- File: `frontend/src/components/features/navigation/mobile-drawer.tsx`
- Slide-in panel from left with overlay backdrop.
- Contains: HouseholdSwitcher (mobile variant), nav items, theme toggle, logout, user info.
- Controlled via `open` / `onClose` props.
- 250ms ease CSS transition.
- Focus trap when open (use `@base-ui/react` Dialog primitive or manual implementation).
- ARIA: `role="dialog"`, `aria-modal="true"`, `aria-label="Navigation"`.
- Clicking overlay or nav item calls `onClose`.
- Effort: M
- Dependencies: 2.3 (for mobile household switcher)

**2.3 Create `MobileHouseholdSelector` component**
- File: `frontend/src/components/features/households/mobile-household-selector.tsx`
- Full-screen overlay with header ("Select Household" + back/close button).
- Lists all households as large tappable rows (min 44px height).
- Checkmark on active household.
- Tapping a household: calls `setActiveHousehold`, closes overlay, closes drawer.
- Reuses existing `useHouseholds` and `useTenant` hooks.
- Preserves `households.length <= 1` → null behavior.
- Effort: M
- Dependencies: None (uses existing hooks)

**2.4 Refactor `AppShell` for responsive layout**
- File: `frontend/src/components/features/navigation/app-shell.tsx`
- Desktop (≥md): Current layout unchanged — fixed sidebar + main content.
- Mobile (<md): MobileHeader at top + main content below. No sidebar rendered.
- State: `drawerOpen` boolean, toggled by hamburger button.
- Use `hidden md:flex` on sidebar, `md:hidden` on MobileHeader.
- Effort: M
- Dependencies: 2.1, 2.2

### Phase 3: Mobile Data Views (L)

Create card-based alternatives to DataTable for each entity type.

**3.1 Create `TaskCard` component**
- File: `frontend/src/components/features/tasks/task-card.tsx`
- Renders: title (with strike-through if completed), status badge, due date, three-dot action menu.
- Action menu options: Complete/Reopen, Delete.
- Props: `task: Task`, `onToggleComplete`, `onDelete`.
- Min height: 44px touch targets on action buttons.
- Effort: S
- Dependencies: None

**3.2 Create `ReminderCard` component**
- File: `frontend/src/components/features/reminders/reminder-card.tsx`
- Renders: title, scheduled time, status badge, three-dot action menu.
- Action menu options: Snooze (if active), Dismiss (if active), Delete.
- Props: `reminder: Reminder`, `onSnooze`, `onDismiss`, `onDelete`.
- Effort: S
- Dependencies: None

**3.3 Create `HouseholdCard` component**
- File: `frontend/src/components/features/households/household-card.tsx`
- Renders: house icon, name, timezone/units, "Active" badge if current.
- Props: `household: Household`, `isActive: boolean`.
- Effort: S
- Dependencies: None

**3.4 Create `CardActionMenu` shared component**
- File: `frontend/src/components/common/card-action-menu.tsx`
- Three-dot icon button that opens a dropdown menu.
- Each menu item: icon + label, calls provided callback.
- Props: `actions: Array<{ icon, label, onClick, variant? }>`.
- Uses existing Button + a simple positioned dropdown (or repurpose existing popover patterns).
- Effort: S
- Dependencies: None

### Phase 4: Responsive Pages (L)

Integrate mobile views into each page, add pull-to-refresh, responsive spacing.

**4.1 Responsive TasksPage**
- Conditionally render TaskCard list (mobile) vs DataTable (desktop) using `use-mobile` hook.
- Extract `toggleComplete` and `handleDelete` handlers to be shared between table columns and card props.
- Wrap content in PullToRefresh, wired to `useTasks` refetch.
- Responsive padding: `p-4 md:p-6`.
- "New Task" button remains visible in page header on both viewports.
- Effort: M
- Dependencies: 1.1, 1.3, 3.1, 3.4

**4.2 Responsive RemindersPage**
- Same pattern as TasksPage: ReminderCard list on mobile, DataTable on desktop.
- Extract `handleSnooze`, `handleDismiss`, `handleDelete` handlers.
- Wrap in PullToRefresh, wired to `useReminders` refetch.
- Responsive padding.
- Effort: M
- Dependencies: 1.1, 1.3, 3.2, 3.4

**4.3 Responsive HouseholdsPage**
- HouseholdCard list on mobile, DataTable on desktop.
- No pull-to-refresh (not a data page per PRD).
- Responsive padding.
- Effort: S
- Dependencies: 1.1, 3.3

**4.4 Responsive DashboardPage**
- Already has `md:grid-cols-3` — verify single-column stacking works.
- Wrap in PullToRefresh, wired to task + reminder summary refetch.
- Responsive padding: `p-4 md:p-6`.
- Responsive heading size: `text-xl md:text-2xl`.
- Effort: S
- Dependencies: 1.1, 1.3

**4.5 Responsive SettingsPage**
- Responsive padding.
- Verify cards go full-width on mobile (they should — no `max-w` constraint).
- Ensure spacing between Profile/Appearance cards works on mobile.
- Effort: S
- Dependencies: None

**4.6 Verify Login & Onboarding pages at 320px**
- Add horizontal margin/padding so `max-w-md` card doesn't touch screen edges.
- Likely just needs `mx-4` or `px-4` on the outer container.
- Verify form inputs and buttons are touch-friendly.
- Effort: S
- Dependencies: None

### Phase 5: Polish & Testing (M)

**5.1 Touch target audit**
- Review all Button usages in mobile-visible contexts.
- The `xs` size (24px) and `sm` size (28px) are below 44px. Ensure mobile-visible buttons use `default` or larger, OR add minimum touch area via padding/hit-area expansion.
- Audit nav links in drawer for adequate touch targets (current `py-2` = ~32px line height).
- Effort: S
- Dependencies: Phase 2, Phase 4

**5.2 Typography & spacing pass**
- Add responsive heading sizes across all pages: `text-xl md:text-2xl`.
- Verify body text is ≥16px (Tailwind `text-sm` is 14px — check if base size needs adjustment on mobile or if it's fine at the current Geist font sizing).
- Ensure Input font-size is 16px on mobile to prevent iOS auto-zoom.
- Effort: S
- Dependencies: Phase 4

**5.3 Card-based empty states**
- Verify that card list views have proper empty states ("No tasks yet" with create button).
- Should mirror existing DataTable `emptyMessage` behavior.
- Effort: S
- Dependencies: Phase 4

**5.4 New tests: MobileDrawer**
- Test drawer opens on hamburger click.
- Test nav item click navigates and closes drawer.
- Test overlay click closes drawer.
- Test focus trap behavior.
- Mock `window.matchMedia` for viewport simulation.
- Effort: M
- Dependencies: 2.2, 2.4

**5.5 New tests: Mobile card views**
- Test TaskCard renders title, status, due date, action menu.
- Test ReminderCard renders title, time, status, action menu.
- Test HouseholdCard renders name, active badge.
- Test CardActionMenu opens and fires callbacks.
- Test responsive pages render cards when `isMobile` is true.
- Effort: M
- Dependencies: Phase 3, Phase 4

**5.6 Existing test verification**
- Run full test suite, ensure all existing tests pass without modification.
- If any fail due to conditional rendering, fix by ensuring tests run in "desktop" viewport by default.
- Effort: S
- Dependencies: 5.4, 5.5

**5.7 Cross-viewport manual verification**
- Test at: 320px, 375px, 414px, 768px, 1024px.
- Verify: no horizontal overflow, no overlapping elements, all pages accessible, all actions functional.
- Effort: S
- Dependencies: All prior phases

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Pull-to-refresh touch handling conflicts with native browser scroll | Medium | Medium | Test on real devices (iOS Safari, Chrome Android); use `touch-action` CSS property; consider a proven library if custom implementation is unreliable |
| Existing tests break due to conditional mobile/desktop rendering | Medium | Low | Default test viewport to desktop (≥768px); mock `matchMedia` to return `false` for mobile query |
| Action menu dropdown positioning issues on small screens | Low | Medium | Use viewport-aware positioning; test at 320px width |
| Performance degradation from dual component trees (card + table) | Low | Low | Only one is rendered at a time (conditional, not CSS hidden); no extra DOM nodes |
| iOS auto-zoom on input focus (font-size < 16px) | Medium | Low | Ensure all form inputs use ≥16px font size on mobile; Tailwind `text-base` = 16px |

---

## Success Metrics

1. All PRD acceptance criteria pass (see `prd.md` Section 10).
2. Existing test suite passes with zero modifications to test assertions.
3. No new JS dependencies added (or at most one small pull-to-refresh library).
4. Lighthouse mobile score ≥90 for performance on all pages.
5. No horizontal scrollbar at any viewport from 320px to 1440px.

---

## Required Resources and Dependencies

- **Team**: 1 frontend developer
- **Tools**: Browser dev tools with device emulation; real iOS/Android device for pull-to-refresh testing recommended
- **Packages**: All existing — React 19, Tailwind v4, @tanstack/react-table, @tanstack/react-query, lucide-react, @base-ui/react
- **Backend**: No changes required

---

## Timeline Estimates

| Phase | Effort | Description |
|-------|--------|-------------|
| Phase 1: Foundation | S | Hooks + PullToRefresh component |
| Phase 2: Navigation | M | Mobile header, drawer, household selector, AppShell refactor |
| Phase 3: Mobile Data Views | M | TaskCard, ReminderCard, HouseholdCard, CardActionMenu |
| Phase 4: Responsive Pages | L | Integrate mobile views into all 7 pages |
| Phase 5: Polish & Testing | M | Touch targets, typography, tests, verification |
