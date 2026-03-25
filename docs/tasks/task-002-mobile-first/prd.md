# Mobile-First Responsive UI — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25

---

## 1. Overview

Home Hub's frontend is currently desktop-only. The layout uses a fixed 256px sidebar that does not collapse, making the application unusable on phones and difficult on tablets. While a few Tailwind responsive classes exist (dashboard grid, dialog sizing), there is no systematic mobile support.

This feature reworks the frontend to treat mobile as a first-class citizen. Every page and interaction that works on desktop must work equally well on a phone. The approach is to retrofit responsive behavior into the existing component set using Tailwind's default breakpoints — no new frameworks, no native app, no backend changes.

The result is a responsive single-page app where mobile users get a hamburger-driven navigation, card-based data views, touch-friendly controls, and pull-to-refresh — while desktop users retain the current sidebar layout with no regressions.

## 2. Goals

Primary goals:
- Full feature parity between mobile and desktop viewports
- Hamburger menu navigation on small screens replacing the fixed sidebar
- Card-based mobile layouts for data-heavy pages (tasks, reminders, households)
- Touch-friendly tap targets and spacing across all interactive elements
- Pull-to-refresh support on mobile for data pages
- Preserve existing skeleton loading states during the transition

Non-goals:
- Progressive Web App (PWA) / offline support / service worker — future task
- Native mobile app (React Native, Capacitor, etc.)
- Swipe gestures or advanced touch interactions
- Mobile-only features not available on desktop
- Push notifications
- Custom breakpoints — use Tailwind defaults (sm:640, md:768, lg:1024, xl:1280)

## 3. User Stories

- As a mobile user, I want to navigate the app using a hamburger menu so that I can access all pages without a persistent sidebar consuming screen space.
- As a mobile user, I want to view my tasks and reminders as cards instead of data tables so that I can read and interact with them on a small screen.
- As a mobile user, I want to pull down on data pages to refresh content so that I can get updated information with a familiar gesture.
- As a mobile user, I want buttons, inputs, and other controls to be large enough to tap accurately so that I don't mis-tap.
- As a tablet user, I want the layout to adapt fluidly between mobile and desktop patterns so that I get the best experience for my screen size.
- As a desktop user, I want the existing sidebar layout and data table views to remain unchanged so that this work does not regress my experience.

## 4. Functional Requirements

### 4.1 Responsive Navigation

- Below the `md` breakpoint (768px), hide the fixed sidebar entirely.
- Show a top header bar containing: hamburger icon (left), "Home Hub" branding (center or left), and user avatar or action (right).
- Tapping the hamburger icon opens a slide-in drawer from the left containing all current sidebar content: household switcher, nav items, theme toggle, logout, user info.
- Tapping outside the drawer or tapping a nav item closes the drawer.
- The drawer must include an overlay/backdrop to indicate modal state.
- At `md` and above, the current fixed sidebar layout is preserved unchanged.

### 4.2 Responsive Page Layouts

**Dashboard:**
- Cards stack in a single column below `md`, current 3-column grid at `md` and above (already partially implemented via `md:grid-cols-3`).
- Summary sections should use full-width cards on mobile.

**Tasks Page:**
- Below `md`, replace the DataTable with a card-based list. Each card shows: task title, status badge, due date, and an actions menu.
- Above `md`, retain the existing DataTable.
- "Add Task" button should be accessible from a floating action button (FAB) or a prominent button in the mobile header area.

**Reminders Page:**
- Same card-based pattern as tasks on mobile: title, next trigger time, status, actions.
- Above `md`, retain the existing DataTable.

**Households Page:**
- Cards should stack vertically on mobile with full-width layout.
- Member lists within a household should be readable on small screens.

**Settings Page:**
- Form fields should go full-width on mobile.
- Adequate spacing between form groups for touch.

**Onboarding / Login:**
- Already use `max-w-md` centered cards — verify they work at 320px viewport width with appropriate padding so content does not touch screen edges.

### 4.3 Touch-Friendly Sizing

- All tap targets (buttons, links, interactive elements) must be at least 44x44px per Apple HIG / WCAG guidelines.
- Form inputs must have adequate height and padding for finger input.
- Spacing between interactive elements must prevent accidental adjacent taps.
- The existing Button component sizes (xs, sm, default, lg) should be audited — `xs` may need a minimum mobile size or be excluded from mobile use.

### 4.4 Mobile Household Switcher

- Below `md`, replace the compact Select dropdown with a full-screen selector overlay.
- The full-screen selector displays all available households as a tappable list with clear labels and large touch targets.
- The currently active household is visually indicated (checkmark or highlight).
- Tapping a household selects it, closes the overlay, and triggers the context switch.
- A back/close button or tap-outside dismisses the overlay without changing the selection.
- The selector is accessed from the same position in the navigation drawer where the desktop Select appears.

### 4.5 Pull-to-Refresh

- Data pages (Dashboard, Tasks, Reminders) support pull-to-refresh gesture on mobile.
- Pulling triggers a refetch of the page's primary query via TanStack React Query's `refetch()`.
- Visual indicator (spinner or pull indicator) during the gesture.
- Only active on touch devices / below `md` breakpoint.

### 4.6 Responsive Data Table / Card Toggle

- Create a responsive wrapper or pattern that renders a DataTable on `md`+ and a card list on smaller viewports.
- Card list items must support the same actions available in the DataTable (edit, delete, complete, snooze, dismiss, restore — as applicable per entity).
- Card actions should use a compact menu (e.g., three-dot icon opening a dropdown) to avoid clutter.

### 4.7 Typography & Spacing

- Base font sizes should be readable on mobile without zooming (minimum 16px for body text to prevent iOS auto-zoom on input focus).
- Headings should scale down on mobile using responsive Tailwind classes.
- Page padding should reduce on mobile (e.g., `p-4` on mobile vs `p-6` on desktop).

## 5. API Surface

No API changes required. This feature is entirely frontend.

## 6. Data Model

No data model changes required.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| frontend | All layout, navigation, and data display components modified for responsive behavior |
| auth-service | None |
| account-service | None |
| productivity-service | None |

## 8. Non-Functional Requirements

**Performance:**
- No additional JS bundle size beyond what's needed for the hamburger drawer and pull-to-refresh.
- Card-based mobile views should lazy-render or virtualize if lists exceed ~50 items.
- Responsive behavior via CSS/Tailwind where possible (avoid JS-based viewport detection when CSS media queries suffice).

**Accessibility:**
- Hamburger menu must be keyboard accessible and use proper ARIA attributes (aria-expanded, aria-label, role).
- Drawer focus trap when open.
- Card-based views must be navigable via keyboard/screen reader with the same functionality as DataTable views.

**Testing:**
- Existing component tests must pass without modification (desktop behavior preserved).
- New tests for mobile navigation (drawer open/close, nav item routing).
- New tests for responsive data card rendering.
- Visual verification at 320px, 375px, 414px, 768px, and 1024px viewports.

**Browser Support:**
- Mobile Safari (iOS 15+)
- Chrome for Android
- Desktop Chrome, Firefox, Safari, Edge (no regressions)

## 9. Open Questions

- Should the pull-to-refresh library be a third-party package or a lightweight custom implementation?

## 10. Acceptance Criteria

- [ ] At viewports below 768px, the sidebar is hidden and a hamburger menu is shown in a top header bar.
- [ ] Tapping the hamburger opens a slide-in drawer with all navigation items, household switcher, theme toggle, logout, and user info.
- [ ] Tapping a nav item in the drawer navigates to the page and closes the drawer.
- [ ] Tapping outside the drawer closes it.
- [ ] Household switcher renders as a full-screen selector overlay below 768px with tappable list items and active-household indicator.
- [ ] Tasks page renders as cards below 768px with title, status, due date, and action menu.
- [ ] Reminders page renders as cards below 768px with title, trigger time, status, and action menu.
- [ ] Dashboard cards stack in a single column below 768px.
- [ ] All tap targets are at least 44x44px on mobile viewports.
- [ ] Pull-to-refresh works on Dashboard, Tasks, and Reminders pages on touch devices.
- [ ] Login and onboarding pages render correctly at 320px width with no horizontal overflow.
- [ ] Desktop layout (sidebar, DataTable, existing spacing) is unchanged at 1024px+ viewports.
- [ ] Existing tests pass without modification.
- [ ] New tests cover drawer behavior and card-based responsive views.
- [ ] No new JS dependencies added beyond pull-to-refresh (if third-party).
