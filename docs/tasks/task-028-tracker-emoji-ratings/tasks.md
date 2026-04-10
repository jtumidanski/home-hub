# Tracker Emoji Ratings — Tasks

Last Updated: 2026-04-09

---

## Implementation

- [ ] **Update today-view.tsx** — Change `ratings` array emoji values to 😊/😐/😞 and remove text labels from button render
  - File: `frontend/src/components/features/tracker/today-view.tsx`
  - Lines: 110-114 (array), 120 (render)
  - Acceptance: Buttons show only face emoji, no "Good"/"OK"/"Bad" text

- [ ] **Update calendar-grid.tsx CellContent** — Replace `+`/`-`/`~` ternary with emoji in cell display
  - File: `frontend/src/components/features/tracker/calendar-grid.tsx`
  - Line: 311
  - Acceptance: Calendar day cells show 😊/😐/😞

- [ ] **Update calendar-grid.tsx CellEditor** — Replace `+`/`-`/`~` in popover editor buttons with emoji
  - File: `frontend/src/components/features/tracker/calendar-grid.tsx`
  - Line: 381
  - Acceptance: Editor buttons show 😊/😐/😞

- [ ] **Update month-report.tsx** — Replace `+`/`~`/`-` stat prefixes with emoji
  - File: `frontend/src/components/features/tracker/month-report.tsx`
  - Lines: 95-97
  - Acceptance: Stats show `😊 N`, `😐 N`, `😞 N`

## Verification

- [ ] **Build check** — `npm run build` passes without errors
- [ ] **Visual check** — Emoji render correctly and fit within calendar cells without overflow
