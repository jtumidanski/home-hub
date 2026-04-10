# Tracker Mobile Visibility — Tasks

Last Updated: 2026-04-09

## Phase 1: Today View Updates

- [ ] 1.1 Add `colorBorderLeft` map with all 16 tracker colors (`border-l-{color}-500`) to `today-view.tsx`
- [ ] 1.2 Import `Badge` from `@/components/ui/badge` and `Check` from `lucide-react`
- [ ] 1.3 Update unset card styling: apply `border-l-[3px]` + item color class, remove current `border-primary/30`
- [ ] 1.4 Update logged card styling: replace green text with `Badge` containing `Check` icon + "logged" text
- [ ] 1.5 Add skipped state detection and render "skipped" `Badge` (muted/gray)
- [ ] 1.6 Add muted card styling for skipped items (`bg-muted/50` or `opacity-60`)
- [ ] 1.7 Add progress bar below header (match Calendar view pattern: `bg-primary` fill bar with count text)

## Phase 2: Calendar Mobile Day View Updates

- [ ] 2.1 Add `colorBorderLeft` map to `calendar-grid.tsx`
- [ ] 2.2 Import `Badge` and `Check` icon
- [ ] 2.3 Update unset card styling: apply left border accent matching Today view
- [ ] 2.4 Update logged card styling: replace green text with `Badge` + `Check` icon
- [ ] 2.5 Update skipped card styling: ensure Badge + muted background match Today view
- [ ] 2.6 Verify skipped text already renders as Badge (currently it's plain `span`, needs update)

## Phase 3: Verification

- [ ] 3.1 Verify all three states render correctly in light mode
- [ ] 3.2 Verify all three states render correctly in dark mode
- [ ] 3.3 Verify desktop calendar grid (`<table>` view) is completely unchanged
- [ ] 3.4 Test with multiple color variants to ensure left border renders for all 16 colors
- [ ] 3.5 Build frontend successfully (`npm run build`)
