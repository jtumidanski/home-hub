# Recipe Cook Mode — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-31
---

## 1. Overview

Cook Mode is a full-screen overlay for the recipe detail page that maximizes step readability for hands-free use in the kitchen. When activated, recipe instructions fill the viewport with the largest possible auto-scaled text, preserving Cooklang rich formatting — highlighted ingredients with quantities/units, cookware, timers, and recipe references.

The feature is frontend-only. No backend or API changes are required — it reuses the already-parsed step data available on the recipe detail page.

## 2. Goals

Primary goals:
- Allow users to read recipe steps from across the room while cooking
- Preserve Cooklang inline formatting (ingredients, quantities, timers, cookware) at the enlarged scale
- Prevent the screen from sleeping during active cooking via the Screen Wake Lock API
- Provide both an all-steps scrollable view and a single-step focused view

Non-goals:
- Voice control or hands-free navigation
- Interactive timer start/stop functionality
- Step completion tracking or checkboxes
- Displaying the ingredients sidebar — only inline step formatting
- Offline/PWA support
- Backend or API changes

## 3. User Stories

- As a home cook, I want to blow up my recipe instructions to fill the screen so that I can read them from across the kitchen
- As a home cook, I want to see ingredient names, quantities, and cook times highlighted inline so that I don't miss measurements while cooking
- As a home cook, I want my screen to stay awake in cook mode so that I don't have to touch my device with messy hands
- As a home cook, I want to toggle between viewing all steps at once and focusing on a single step so that I can choose what works best for the recipe

## 4. Functional Requirements

### 4.1 Activation

- A "Cook Mode" button appears in the recipe detail page header action buttons (alongside Edit and Delete)
- Clicking the button opens a full-screen overlay that covers the entire viewport
- The overlay uses a clean background (respecting light/dark theme)

### 4.2 Display — All Steps View (Default)

- All recipe steps are displayed in a single scrollable column
- Text auto-scales using viewport-relative units (`clamp`/`vw`) to be as large as possible while fitting the viewport width
- Section headers are displayed when steps span multiple Cooklang sections
- Step numbers are clearly visible
- Cooklang segment formatting is preserved with the same color conventions:
  - Ingredients (orange) with quantity and unit shown inline
  - Cookware (blue)
  - Timers (green)
  - Recipe references (purple, italic)

### 4.3 Display — Single Step View

- A toggle switches between "All Steps" and "Single Step" modes
- In single step mode, one step fills the viewport with even larger text
- Previous/Next navigation buttons are provided (large touch targets)
- Current step number and total count are displayed (e.g., "3 / 12")
- Section header is shown above the step when applicable
- Keyboard navigation: left/right arrow keys for prev/next
- Swipe gestures: swipe left for next step, swipe right for previous step (touch devices)

### 4.4 Screen Wake Lock

- When cook mode is active, request a Screen Wake Lock via the Wake Lock API
- Release the wake lock when cook mode is closed
- If the Wake Lock API is unavailable (unsupported browser), degrade gracefully — cook mode still works, just without wake lock
- Re-acquire wake lock on visibility change (tab refocus) per API best practices
- A small indicator icon (e.g., eye or lock) is shown in the overlay header to confirm the wake lock is active; hidden when wake lock is unavailable or fails

### 4.5 Exit

- An X button in the top-right corner closes cook mode
- Pressing the Escape key closes cook mode
- Closing cook mode releases the wake lock and returns to the normal recipe detail view

## 5. API Surface

No API changes. This feature uses the existing `RecipeDetailAttributes.steps` data already fetched by the recipe detail page.

## 6. Data Model

No data model changes.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| Frontend | New `CookMode` component, button added to `RecipeDetailPage`, shared step segment rendering |
| All backend services | No changes |

## 8. Non-Functional Requirements

- **Performance**: The overlay should open instantly — no additional API calls needed
- **Responsiveness**: Must work well on tablets and phones (primary kitchen devices) as well as desktop/laptop screens
- **Accessibility**: Large text inherently improves accessibility; ensure sufficient color contrast for highlighted segments at large sizes; support keyboard navigation (Escape to close, arrow keys in single-step mode)
- **Browser support**: Screen Wake Lock API is supported in Chromium browsers and Safari 16.4+; Firefox does not support it — degrade gracefully

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

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
