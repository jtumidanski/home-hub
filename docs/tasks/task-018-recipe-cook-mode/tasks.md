# Task 018 — Recipe Cook Mode: Task Checklist

Last Updated: 2026-03-31

---

## Phase 1: Extract Shared Segment Renderer

- [ ] 1.1 Create `StepSegment` component in `step-segment.tsx` with size variant support
- [ ] 1.2 Refactor `recipe-steps.tsx` to use `StepSegment` — verify no visual regression

## Phase 2: Cook Mode Core — All Steps View

- [ ] 2.1 Create `CookMode` overlay component with full-viewport layout and auto-scaled all-steps view
- [ ] 2.2 Add close functionality (X button + Escape key)
- [ ] 2.3 Add "Cook Mode" button to `RecipeDetailPage` alongside Edit/Delete

## Phase 3: Single Step View with Navigation

- [ ] 3.1 Add view mode toggle (All Steps / Single Step) in overlay header
- [ ] 3.2 Build single-step view with centered step, section header, and step counter
- [ ] 3.3 Add prev/next button navigation with large touch targets
- [ ] 3.4 Add keyboard navigation (Left/Right arrow keys)
- [ ] 3.5 Add swipe gesture navigation for touch devices

## Phase 4: Screen Wake Lock

- [ ] 4.1 Create `useWakeLock` hook with acquire/release/re-acquire lifecycle
- [ ] 4.2 Add wake lock indicator icon to overlay header

## Phase 5: Polish and Responsiveness

- [ ] 5.1 Tune responsive text scaling with `clamp()` across mobile/tablet/desktop
- [ ] 5.2 Verify light/dark theme support and segment color contrast
- [ ] 5.3 Run build and lint — fix any errors

## PRD Acceptance Criteria

- [ ] "Cook Mode" button visible on recipe detail page
- [ ] Clicking button opens full-screen overlay with all steps displayed
- [ ] Text auto-scales to fill viewport width using viewport-relative units
- [ ] Cooklang formatting preserved: ingredients (orange, with qty/unit), cookware (blue), timers (green), references (purple)
- [ ] Section headers display correctly for multi-section recipes
- [ ] Toggle switches between all-steps and single-step view
- [ ] Single-step view shows one step at maximum size with prev/next navigation
- [ ] Arrow key navigation works in single-step mode
- [ ] Swipe left/right navigates between steps in single-step mode on touch devices
- [ ] Step counter shows current position (e.g., "3 / 12") in single-step mode
- [ ] Screen Wake Lock is acquired on open, released on close (in supported browsers)
- [ ] Wake lock indicator icon visible when wake lock is active, hidden when unavailable
- [ ] X button and Escape key both close cook mode
- [ ] Works correctly in both light and dark themes
- [ ] Responsive on mobile, tablet, and desktop viewports
