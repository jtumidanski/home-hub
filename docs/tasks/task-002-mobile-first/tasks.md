# Mobile-First Responsive UI — Task Checklist

Last Updated: 2026-03-25

---

## Phase 1: Foundation

- [x] **1.1** Create `use-mobile` hook (`lib/hooks/use-mobile.ts`) — `matchMedia` listener returning `isMobile` boolean at 767px threshold
- [x] **1.2** Create `use-pull-to-refresh` hook (`lib/hooks/use-pull-to-refresh.ts`) — touch event tracking, pull distance, threshold detection, async onRefresh callback
- [x] **1.3** Create `PullToRefresh` wrapper component (`components/common/pull-to-refresh.tsx`) — spinner indicator, only active on mobile, accepts `onRefresh` + `children`

## Phase 2: Navigation

- [x] **2.1** Create `MobileHeader` component (`components/features/navigation/mobile-header.tsx`) — hamburger button, branding, `md:hidden`
- [x] **2.2** Create `MobileDrawer` component (`components/features/navigation/mobile-drawer.tsx`) — slide-in panel, overlay, nav items, theme toggle, logout, user info, focus trap, ARIA
- [x] **2.3** Create `MobileHouseholdSelector` component (`components/features/households/mobile-household-selector.tsx`) — full-screen overlay, tappable household list, checkmark on active, back button
- [x] **2.4** Refactor `AppShell` — `hidden md:flex` sidebar, `md:hidden` MobileHeader, `drawerOpen` state, wire hamburger to drawer

## Phase 3: Mobile Data Views

- [x] **3.1** Create `TaskCard` component (`components/features/tasks/task-card.tsx`) — title, status badge, due date, three-dot action menu with Complete/Delete
- [x] **3.2** Create `ReminderCard` component (`components/features/reminders/reminder-card.tsx`) — title, scheduled time, status badge, action menu with Snooze/Dismiss/Delete
- [x] **3.3** Create `HouseholdCard` component (`components/features/households/household-card.tsx`) — house icon, name, timezone/units, active badge
- [x] **3.4** Create `CardActionMenu` component (`components/common/card-action-menu.tsx`) — three-dot button + dropdown menu, reusable across entity cards

## Phase 4: Responsive Pages

- [x] **4.1** TasksPage — conditional card/table rendering, extract shared handlers, PullToRefresh, responsive padding `p-4 md:p-6`
- [x] **4.2** RemindersPage — conditional card/table rendering, extract shared handlers, PullToRefresh, responsive padding
- [x] **4.3** HouseholdsPage — conditional card/table rendering, responsive padding
- [x] **4.4** DashboardPage — verify grid stacking, PullToRefresh on summary queries, responsive padding + heading
- [x] **4.5** SettingsPage — responsive padding, verify full-width cards on mobile
- [x] **4.6** LoginPage + OnboardingPage — add `px-4` horizontal padding, verify at 320px width, confirm touch-friendly inputs

## Phase 5: Polish & Testing

- [x] **5.1** Touch target audit — mobile drawer nav items use `py-3` + `h-5 w-5` icons; action menu items use `py-2.5`; buttons in mobile views use adequate sizing
- [x] **5.2** Typography & spacing pass — responsive headings `text-xl md:text-2xl` on all pages, responsive padding `p-4 md:p-6`
- [x] **5.3** Card-based empty states — empty state messages with create button in card list view for Tasks and Reminders
- [x] **5.4** Global `matchMedia` mock in `test/setup.ts` — defaults to desktop viewport
- [x] **5.5** Tests: CardActionMenu (4 tests) and TaskCard (6 tests) — open/close, action callbacks, rendering
- [x] **5.6** Existing test verification — all 256 original tests pass, 10 new tests added (266 total)
- [ ] **5.7** Cross-viewport manual verification — 320px, 375px, 414px, 768px, 1024px; no overflow, all pages functional

---

## Task Dependencies

```
1.1 ──┬──→ 1.3 ──→ 4.1, 4.2, 4.4
      │
      └──→ 4.1, 4.2, 4.3
1.2 ──→ 1.3

2.1 ──→ 2.4
2.2 ──→ 2.4
2.3 ──→ 2.2

3.1 ──→ 4.1
3.2 ──→ 4.2
3.3 ──→ 4.3
3.4 ──→ 3.1, 3.2

Phase 4 ──→ 5.1, 5.2, 5.3
Phase 2 + 4 ──→ 5.4, 5.5
5.4 + 5.5 ──→ 5.6
All ──→ 5.7
```

## Parallelization Opportunities

- **Phase 1 + Phase 3** can run in parallel (hooks and card components are independent)
- **Tasks 3.1, 3.2, 3.3** can all be built in parallel (independent entity cards)
- **Tasks 2.1, 2.3** can be built in parallel (header and household selector are independent)
- **Tasks 4.5, 4.6** have no dependencies on Phase 1-3 and can run anytime
- **Tasks 5.1, 5.2, 5.3** can run in parallel after Phase 4
