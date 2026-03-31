# Task 018 — Recipe Cook Mode: Implementation Plan

Last Updated: 2026-03-31

---

## Executive Summary

Add a frontend-only "Cook Mode" feature to the recipe detail page. When activated, it opens a full-screen overlay displaying recipe steps with auto-scaled, large text optimized for reading from across a kitchen. The feature includes two view modes (all-steps scrollable and single-step focused), Screen Wake Lock API integration, Cooklang-formatted segments, and touch/keyboard navigation. No backend changes required.

---

## Current State Analysis

**Recipe Detail Page** (`frontend/src/pages/RecipeDetailPage.tsx`):
- Fetches recipe data via `useRecipe(id)` hook
- Renders header with Edit and Delete action buttons
- Displays metadata, tags, ingredients, and steps in a two-column layout
- Steps rendered by `RecipeSteps` component

**Step Rendering** (`frontend/src/components/features/recipes/recipe-steps.tsx`):
- Iterates `Step[]` array with section header tracking
- Segment switch renders: ingredient (orange), cookware (blue), timer (green), reference (purple), text (plain)
- Segment rendering logic is inline within the component — will need extraction for reuse

**Data Model** (`frontend/src/types/models/recipe.ts`):
- `Step` has `number`, optional `section`, and `segments: Segment[]`
- `Segment` has `type` (text|ingredient|cookware|timer|reference) with contextual fields
- All data already available from `RecipeDetailAttributes.steps` — no API calls needed

---

## Proposed Future State

A new `CookMode` component renders as a full-screen overlay triggered from the recipe detail page. It reuses a shared segment renderer extracted from `recipe-steps.tsx`. The overlay has two view modes toggled by a control in the header, wake lock integration, and keyboard/swipe navigation.

**Component Structure:**
```
RecipeDetailPage
├── "Cook Mode" Button (new, alongside Edit/Delete)
└── CookMode (new overlay component)
    ├── CookModeHeader (title, view toggle, wake lock indicator, close button)
    ├── CookModeAllSteps (scrollable large-text view)
    └── CookModeSingleStep (focused view with navigation)

Shared:
└── StepSegment (extracted from recipe-steps.tsx, reused by both)
```

---

## Implementation Phases

### Phase 1: Extract Shared Segment Renderer

Extract the Cooklang segment rendering logic from `recipe-steps.tsx` into a reusable component so both the existing step display and cook mode can share it.

### Phase 2: Build Cook Mode Core

Create the full-screen overlay with the all-steps view, basic header with close functionality, and Escape key handling.

### Phase 3: Single Step View with Navigation

Add the single-step focused view with prev/next buttons, keyboard arrow keys, swipe gesture support, and step counter.

### Phase 4: Screen Wake Lock

Integrate the Screen Wake Lock API with acquire/release lifecycle, visibility change re-acquisition, and status indicator.

### Phase 5: Polish and Responsiveness

Ensure responsive behavior across mobile/tablet/desktop, light/dark theme support, and auto-scaling text.

---

## Detailed Tasks

### Phase 1: Extract Shared Segment Renderer

**1.1** Create `StepSegment` component
- File: `frontend/src/components/features/recipes/step-segment.tsx`
- Extract the segment `switch` block from `recipe-steps.tsx` into a standalone `StepSegment` component
- Props: `segment: Segment`, optional `className` for text size overrides
- Accept an optional size variant (`"default"` | `"large"`) controlling font sizing
- Acceptance: existing recipe detail page renders identically after refactor
- Effort: S

**1.2** Refactor `recipe-steps.tsx` to use `StepSegment`
- Replace inline segment switch with `<StepSegment>` calls
- Verify no visual regression
- Acceptance: `RecipeSteps` renders identically
- Effort: S

### Phase 2: Cook Mode Core — All Steps View

**2.1** Create `CookMode` overlay component
- File: `frontend/src/components/features/recipes/cook-mode.tsx`
- Full-viewport overlay (`fixed inset-0 z-50`) with background respecting theme (`bg-background`)
- Props: `steps: Step[]`, `title: string`, `open: boolean`, `onClose: () => void`
- Render all steps in scrollable column using `StepSegment` with large size variant
- Section headers displayed between steps when section changes
- Step numbers shown as large badges
- Auto-scale text using `clamp()` with `vw` units via Tailwind arbitrary values
- Acceptance: overlay covers viewport, steps render with large text, sections display correctly
- Effort: M
- Depends on: 1.1

**2.2** Add close functionality
- X button in top-right corner using lucide `X` icon
- Escape key listener (`useEffect` with `keydown` handler)
- Calls `onClose` callback
- Acceptance: X button and Escape both close overlay
- Effort: S

**2.3** Add "Cook Mode" button to RecipeDetailPage
- File: `frontend/src/pages/RecipeDetailPage.tsx`
- Add button alongside Edit/Delete with `ChefHat` or `Utensils` lucide icon
- Manage `open` state with `useState`
- Render `<CookMode>` conditionally when open
- Acceptance: button visible, clicking opens overlay, closing returns to detail view
- Effort: S
- Depends on: 2.1

### Phase 3: Single Step View with Navigation

**3.1** Add view mode toggle
- Toggle control in overlay header: "All Steps" / "Single Step"
- Use shadcn `Button` group or simple toggle buttons
- State: `viewMode: "all" | "single"` with `useState`
- Acceptance: toggle switches between views
- Effort: S
- Depends on: 2.1

**3.2** Build single-step view
- Display one step at a time, centered in viewport
- Even larger text than all-steps view (more viewport space available)
- Show section header above step when applicable
- Step counter: "3 / 12" display
- State: `currentStep: number` index
- Acceptance: single step fills viewport with maximum-size text, counter accurate
- Effort: M
- Depends on: 3.1

**3.3** Add prev/next button navigation
- Large touch-target buttons (at least 48x48px) at bottom or sides of viewport
- Prev disabled on first step, Next disabled on last step
- Use lucide `ChevronLeft`/`ChevronRight` icons
- Acceptance: buttons navigate between steps, correct disable states
- Effort: S
- Depends on: 3.2

**3.4** Add keyboard navigation
- Left/Right arrow key handlers in single-step mode only
- `useEffect` with `keydown` listener, clean up on unmount/mode change
- Acceptance: arrow keys navigate steps in single-step mode, no effect in all-steps mode
- Effort: S
- Depends on: 3.2

**3.5** Add swipe gesture navigation
- Track `touchstart`/`touchend` events for horizontal swipe detection
- Swipe left = next step, swipe right = previous step
- Minimum swipe threshold (e.g., 50px) to prevent accidental triggers
- Acceptance: swipe gestures navigate steps on touch devices
- Effort: S
- Depends on: 3.2

### Phase 4: Screen Wake Lock

**4.1** Create `useWakeLock` hook
- File: `frontend/src/lib/hooks/use-wake-lock.ts`
- Request `navigator.wakeLock.request("screen")` when activated
- Release when deactivated or component unmounts
- Re-acquire on `visibilitychange` event (tab refocus)
- Gracefully handle unsupported browsers (`navigator.wakeLock === undefined`)
- Return `{ isSupported: boolean, isActive: boolean }`
- Acceptance: wake lock acquired on supported browsers, released on close, re-acquired on tab refocus
- Effort: M

**4.2** Add wake lock indicator to overlay header
- Show small lock/eye icon when wake lock is active
- Hide indicator when wake lock is unsupported or failed
- Use `useWakeLock` hook state
- Acceptance: indicator visible when active, hidden when unavailable
- Effort: S
- Depends on: 4.1

### Phase 5: Polish and Responsiveness

**5.1** Responsive text scaling
- Use CSS `clamp()` for text sizes: minimum readable size, preferred `vw`-based size, maximum cap
- All-steps view: `clamp(1.25rem, 3.5vw, 3rem)` (adjust based on testing)
- Single-step view: `clamp(1.5rem, 5vw, 5rem)` (adjust based on testing)
- Segment formatting (quantity parentheses, units) scales proportionally
- Test on mobile (375px), tablet (768px), desktop (1440px+) viewports
- Acceptance: text is large and readable at all breakpoints
- Effort: M

**5.2** Light/dark theme support
- Background: `bg-background` (already theme-aware)
- Text: `text-foreground` (already theme-aware)
- Segment colors already use dark: variants in `StepSegment`
- Verify contrast ratios at large sizes
- Acceptance: cook mode looks correct in both themes
- Effort: S

**5.3** Build and lint verification
- Run `npm run build` and `npm run lint` in frontend
- Fix any TypeScript or lint errors
- Acceptance: clean build with no errors or warnings
- Effort: S

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Wake Lock API not supported in Firefox | Known | Low | Graceful degradation — cook mode works without it |
| Text scaling too large/small on specific devices | Medium | Medium | Use `clamp()` with tested min/max bounds; QA on real devices |
| Swipe gestures conflict with browser back/forward | Low | Medium | Only detect horizontal swipes above threshold; use `touch-action` CSS |
| Segment rendering divergence between normal and cook mode | Low | High | Shared `StepSegment` component eliminates this risk |

---

## Success Metrics

- All 15 acceptance criteria from the PRD pass
- No visual regression on existing recipe detail page
- Clean frontend build with no new warnings
- Wake lock correctly acquired/released (verified in Chrome DevTools)

---

## Required Resources and Dependencies

**Dependencies:**
- No new npm packages required
- Lucide icons (already installed) for UI controls
- shadcn/ui components (already installed) for buttons and toggle
- Screen Wake Lock API (browser-native, no polyfill)

**Key Files:**
- `frontend/src/pages/RecipeDetailPage.tsx` — add button and render overlay
- `frontend/src/components/features/recipes/recipe-steps.tsx` — extract segment renderer
- `frontend/src/components/features/recipes/step-segment.tsx` — new shared component
- `frontend/src/components/features/recipes/cook-mode.tsx` — new overlay component
- `frontend/src/lib/hooks/use-wake-lock.ts` — new hook
- `frontend/src/types/models/recipe.ts` — existing types, no changes

---

## Timeline Estimate

| Phase | Effort | Cumulative |
|-------|--------|------------|
| Phase 1: Extract segment renderer | S | S |
| Phase 2: Cook mode core | M | M |
| Phase 3: Single step + navigation | M | L |
| Phase 4: Wake lock | M | L |
| Phase 5: Polish | M | XL |
| **Total** | **XL** | |
