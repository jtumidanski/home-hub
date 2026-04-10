# Tracker Emoji Ratings — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09
---

## 1. Overview

The tracker page's sentiment scale currently uses plain text characters (`+`, `~`, `-`) and labels ("Good", "OK", "Bad") to represent positive, neutral, and negative ratings. This makes the UI feel utilitarian and requires users to learn the character mapping.

Replacing these with face emoji (😊, 😐, 😞) makes the sentiment immediately intuitive and visually consistent across the today view, calendar grid, and month report. This is a frontend-only display change — the underlying data model is unaffected.

## 2. Goals

Primary goals:
- Replace all text/character-based sentiment indicators with face emoji across the tracker page
- Remove redundant text labels from sentiment buttons in the today view
- Maintain readability at all display sizes (calendar cells, buttons, report stats)

Non-goals:
- Changing the backend data model or API responses
- Modifying numeric or range scale types
- Adding user-configurable emoji choices

## 3. User Stories

- As a user, I want to see emoji faces on the sentiment rating buttons so I can instantly understand each option without reading text labels
- As a user, I want to see emoji in the calendar grid cells so I can quickly scan my month at a glance
- As a user, I want to see emoji in the month report stats so the summary is visually consistent with the rest of the tracker

## 4. Functional Requirements

### 4.1 Emoji Mapping

| Rating     | Current Display | New Display |
|------------|----------------|-------------|
| `positive` | `+` / "Good"  | 😊          |
| `neutral`  | `~` / "OK"    | 😐          |
| `negative` | `-` / "Bad"   | 😞          |

### 4.2 Today View — Sentiment Buttons

- Each button displays only the emoji (no text label)
- The button's selected/unselected visual state remains unchanged
- Button sizing should accommodate the emoji without layout shifts

### 4.3 Calendar Grid — Day Cells

- Replace the `+`/`~`/`-` character with the corresponding emoji
- Emoji should be sized appropriately for the calendar cell (not oversized)

### 4.4 Month Report — Sentiment Stats

- Replace the `+`/`~`/`-` prefix in stat counts with the corresponding emoji (e.g., `😊 5`, `😐 2`, `😞 1`)
- Color-coded bars remain unchanged

## 5. API Surface

No API changes. This is a frontend-only modification.

## 6. Data Model

No data model changes. The stored rating values (`"positive"`, `"neutral"`, `"negative"`) are unchanged.

## 7. Service Impact

| Service   | Impact |
|-----------|--------|
| Frontend  | Modify three component files to use emoji instead of text characters |
| tracker-service | None |

### Frontend Files Affected

- `frontend/src/components/features/tracker/today-view.tsx` — Update `ratings` array: replace `emoji` and `label` fields with face emoji, remove text labels from button rendering
- `frontend/src/components/features/tracker/calendar-grid.tsx` — Update sentiment display mapping to use emoji instead of `+`/`~`/`-`
- `frontend/src/components/features/tracker/month-report.tsx` — Update stat prefix characters to emoji

## 8. Non-Functional Requirements

- Emoji must render correctly across modern browsers (Chrome, Firefox, Safari, Edge)
- Calendar grid cells must not overflow or misalign due to emoji sizing differences
- No perceptible layout shift when switching between emoji and other scale type displays

## 9. Open Questions

None — all decisions resolved during scoping.

## 10. Acceptance Criteria

- [ ] Today view sentiment buttons show 😊, 😐, 😞 with no text labels
- [ ] Calendar grid cells show the corresponding face emoji for each logged sentiment entry
- [ ] Month report stats use face emoji as prefixes instead of `+`/`~`/`-`
- [ ] No backend or data model changes
- [ ] Emoji render correctly and are legible at the displayed sizes
- [ ] Existing button selection behavior and color-coded bars are unaffected
