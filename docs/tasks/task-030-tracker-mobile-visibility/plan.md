# Tracker Mobile Visibility — Implementation Plan

Last Updated: 2026-04-09

## Executive Summary

Improve visual state indicators on the tracker page's mobile views (Today and Calendar day view) so users can immediately distinguish unset, logged, and skipped items. This is a frontend-only change touching two component files. No backend, API, or data model changes required.

## Current State Analysis

### Today View (`today-view.tsx`)
- **Unset items**: Card uses `border-primary/30` (slightly lighter border) — nearly invisible on mobile
- **Logged items**: Small green "logged" text (`text-xs text-green-600`) in the card header — easy to miss
- **Skipped items**: No visual indicator at all — indistinguishable from unset
- **Progress**: Text-only counter at bottom ("X/Y logged today"), no progress bar

### Calendar Mobile Day View (`calendar-grid.tsx` → `MobileDayView`)
- **Unset items**: Plain `Card` with no differentiation — identical to logged cards
- **Logged items**: Small green "logged" text badge — same as Today view
- **Skipped items**: Small muted "skipped" text — exists but inconsistent with Today view
- **No progress bar** (exists only in the desktop/month-level header)

### Available UI Primitives
- `Badge` component exists at `frontend/src/components/ui/badge.tsx` with variants: default, secondary, destructive, outline, ghost, link
- `colorDot` map (today-view.tsx) and `colorBg` map (calendar-grid.tsx) provide per-color Tailwind classes
- Neither file has a `colorBorder` map — needs to be added for left-border accent

## Proposed Future State

### Visual State Matrix

| State | Left Border | Badge | Card Background |
|-------|------------|-------|-----------------|
| Unset | 3px in item color (`border-l-[3px]`) | None | Default |
| Logged | None | `Badge` with `Check` icon, green tint | Default |
| Skipped | None | `Badge` with "skipped", muted | `bg-muted/50` or `opacity-60` |

### Shared Color Map

A new `colorBorderLeft` map provides `border-l-{color}-500` classes for all 16 tracker colors. This will be used in both files.

### Progress Bar

Today view gets a progress bar matching the Calendar view's existing pattern — `bg-primary` fill bar with "{filled + skipped}/{total} entries" text.

## Implementation Phases

### Phase 1: Add Color Border Map and Shared State Logic

Add a `colorBorderLeft` map to both files (or extract to a shared constants file). This maps each of the 16 tracker colors to a `border-l-{color}-500` Tailwind class.

### Phase 2: Update Today View

1. Import `Badge` component and `Check` icon from lucide-react
2. Add `colorBorderLeft` map
3. Update card rendering to apply left border for unset items
4. Replace green "logged" text with `Badge` + `Check` icon
5. Add skipped state detection and "skipped" Badge
6. Add muted styling for skipped cards
7. Add progress bar below the header

### Phase 3: Update Calendar Mobile Day View

1. Import `Badge` component and `Check` icon
2. Add `colorBorderLeft` map (or reuse from shared location)
3. Apply the same card state treatments as Today view
4. Ensure consistency between both views

### Phase 4: Verify and Polish

1. Test all three states in light and dark mode
2. Verify desktop calendar grid is unaffected
3. Check all 16 color variants render correctly for left border

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tailwind purge misses dynamic border-l classes | Medium | High | Use full class names in the map, not string interpolation |
| Left border shifts card content slightly | Low | Low | Use `border-l-[3px]` consistently; cards already have padding |
| Badge overflows on narrow screens | Low | Medium | Use `text-xs` sizing, test on 320px viewport |

## Success Metrics

- Users can distinguish all three states at a glance on mobile
- Visual treatment is consistent between Today and Calendar day views
- No regressions on desktop calendar grid

## Required Resources and Dependencies

- Frontend components: `Badge` (exists), `Check` icon from lucide-react
- Tailwind classes: `border-l-{color}-500` for 16 colors (must be full strings for purge)
- No external dependencies

## Effort Estimate

- **Total effort**: Small (S)
- **Phase 1**: Trivial — add color map
- **Phase 2**: Small — update Today view card rendering + progress bar
- **Phase 3**: Small — mirror changes in MobileDayView
- **Phase 4**: Trivial — visual verification
