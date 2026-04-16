# Dashboard Widget Skip/Partial Status — Task Checklist

Last Updated: 2026-04-16

---

## Phase 1: Habits Widget

- [ ] **1.1** Build `skippedItemIds` set from entries where `skipped === true`
- [ ] **1.2** Import `SkipForward` icon and render skipped habits with muted strikethrough + aria-label

## Phase 2: Workout Widget

- [ ] **2.1** Import `SkipForward` and `CircleDot` icons; expand rendering to handle all four performance statuses (done, skipped, partial, pending) with distinct icon/color/text + aria-labels

## Phase 3: Verification

- [ ] **3.1** Start dev server and verify all states render correctly on the dashboard
- [ ] **3.2** Confirm no regressions: loading, error, empty states, links, pull-to-refresh
