# Dashboard Designer — UX Flow

## Navigation

Sidebar renders a **Dashboards** group. Each dashboard visible to the caller is its own child entry with:
- Icon (shared dashboard icon for household, person icon for user-scoped)
- Name
- Drag handle (shown on hover on desktop)
- Kebab menu: Rename, Set as my default, Delete, (Promote / Copy to mine depending on scope)

Below the list: a **+ New dashboard** row.

Clicking a dashboard navigates to `/app/dashboards/{id}`. The currently active entry is highlighted.

Legacy `/app/dashboard` redirects via the resolution order in `data-model.md` §5.

## View mode

Header:
```
[Dashboard name]                                  [Edit] [⋮]
```

- `[Edit]` visible to anyone who can edit (disabled with tooltip on mobile).
- Kebab exposes: Rename, Set as my default, Delete, Promote / Copy to mine.
- Below the header is the grid-rendered dashboard (same widget components used today, selected from registry by `type`).
- Pull-to-refresh at the top invalidates all widget queries.

On viewports below the `md` breakpoint (roughly < 768px): widgets stack in single-column, row-major order. Widget heights adjust to content; horizontal size is always full width. The grid does not show gridlines.

## New dashboard modal

Fields:
- **Name** (required, 1–80 chars)
- **Scope** (radio): Household / Just me
- (Optional) **Start from**: Blank / Copy of an existing dashboard (dropdown of visible dashboards)

On create: navigate into edit mode on the new dashboard.

## Edit mode (designer)

Entry: `[Edit]` button on view mode, OR auto-entered after creating a new dashboard.

Layout on screen:
```
┌──────────────────────────────────────────────────────────┬──────────┐
│  [Save]  [Discard]  [Dashboard name ▾]    edit mode       │ Widgets  │
├──────────────────────────────────────────────────────────┤  palette  │
│                                                           │           │
│  [ weather banner, grid cell (0,0,12,3) ]                 │  weather  │
│                                                           │  tasks    │
│  [ tasks ] [ reminders ] [ overdue ]                      │  remind.  │
│                                                           │  overdue  │
│  [ meals ] [ habits ] [ packages ]                        │  meals    │
│                                                           │  ...      │
└──────────────────────────────────────────────────────────┴──────────┘
```

Widget in edit mode:
```
┌───────────────────────────┐
│  ⋮ Drag   [gear] [trash]  │   <- edit chrome revealed in edit mode only
├───────────────────────────┤
│                           │
│    (live widget preview)  │
│                           │
└───────────────────────────┘
     ↘ resize handle
```

Interactions:
- **Drag a widget**: picks it up, shows drop shadow, snaps on release. Collision pushes existing widgets downward.
- **Drag from palette**: shows live preview; drops at cursor cell with `defaultSize` and `defaultConfig`.
- **Resize**: corner and edge handles. Enforces `minSize` / `maxSize`. Snaps to grid cells.
- **Gear**: opens right-side config panel for that instance.
- **Trash**: removes the instance (no confirm; save is explicit).
- **Esc**: closes the palette or config panel.

Header controls in edit mode:
- `[Save]` — PATCH to server with the new layout and (if changed) name. Exit to view mode on success.
- `[Discard]` — warn if dirty, then reset to server copy.
- `[Rename]` — inline edit of name in the header.
- `[Scope toggle]` — dropdown; changing scope may create a copy (see PRD §4.5).
- `[Delete dashboard]` — confirmation dialog. On success, navigate to user's default (or fallback).

Dirty-state tracking:
- Any add/remove/move/resize/config change sets a dirty flag.
- Attempting to navigate away (sidebar click, back button) prompts the user to save or discard.

## Config panel

Opens as a side drawer when a widget's gear icon is clicked. Contents:
- Widget name + short description (read-only)
- Form rendered from the widget's Zod schema:
    - Enum → radio group or select
    - Boolean → toggle
    - String → text input (labeled, with char counter)
    - Number → numeric input with bounds
    - Nested object (e.g. `location`) → grouped fieldset
- `[Apply]` — merges into the in-memory layout.
- `[Cancel]` — discards changes for this widget only.
- `[Reset to defaults]` — loads the widget's `defaultConfig`.

Validation errors are inline and block `[Apply]`.

## Seeding

When the frontend loads `/app` for the first time against a brand-new household:

1. `GET /api/v1/dashboards` → `[]`
2. Frontend submits `POST /api/v1/dashboards/seed` with the seed body from `data-model.md` §6
3. Server returns 201 with the new "Home" dashboard
4. Frontend navigates to its id; view mode renders immediately

If a later visit finds the seed already exists, seeding is idempotent — the POST returns 200 and the existing list, and the frontend simply renders the first household dashboard.

## Deletion

- Deleting any dashboard: confirmation modal "Delete 'Weekend'? This can't be undone."
- If the deleted dashboard was the caller's default, clear `default_dashboard_id`.
- If the deleted dashboard was the user's currently open dashboard, navigate to the fallback resolution order.
- Deleting the last household dashboard is allowed — the next `/app/dashboard` visit re-seeds.

## Copy to mine / Promote

- Kebab menu on any household dashboard includes **Copy to mine** for every member. Clicking it creates a user-scoped copy and navigates to the copy in view mode.
- Kebab menu on a user-scoped dashboard (visible only to its owner) includes **Promote to household** for the owner. Confirmation modal warns other members will be able to see and edit it.
