# Household Calendar — UX Flow

## 1. Calendar Page (Default View)

**Route:** `/app/calendar`

**Layout:**

```
┌──────────────────────────────────────────────────────────────┐
│  ← Prev Week    March 23 – March 29, 2026    Next Week →    │
│                                          [Connect Calendar]  │
├──────────────────────────────────────────────────────────────┤
│           │ Mon 23 │ Tue 24 │ Wed 25 │ Thu 26*│ Fri 27 │ Sat 28 │ Sun 29 │
│  All Day  │ Jane:  │        │        │        │ John:  │        │        │
│           │ Vacation│       │        │        │ Day Off│        │        │
├───────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┤
│  6:00 AM  │        │        │        │        │        │        │        │
│  7:00 AM  │        │ ┌────┐ │        │        │        │        │        │
│  8:00 AM  │        │ │Jane│ │        │        │        │        │        │
│  9:00 AM  │ ┌────┐ │ │Gym │ │ ┌────┐ │ ┌────┐ │        │        │        │
│ 10:00 AM  │ │Jane│ │ └────┘ │ │John│ │ │Jane│ │        │        │        │
│ 11:00 AM  │ │Mtg │ │        │ │Busy│ │ │Mtg │ │        │        │        │
│ 12:00 PM  │ └────┘ │        │ └────┘ │ └────┘ │        │        │        │
│   ...     │        │        │        │        │        │        │        │
└───────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┘

Legend: ■ Jane Doe  ■ John Doe
        * = Today
```

**Behavior:**

- 7 columns for days, rows for hours (default 6 AM – 11 PM visible, scrollable)
- All times displayed in the household's timezone (from account-service household model)
- Today's column has a subtle highlight and current time indicator line
- Events are colored by user with a legend below or beside the grid
- Overlapping events within the same time slot render side-by-side, splitting the column width proportionally
- All-day events appear in a collapsible section above the hourly grid
- Private events from other users show as "[User]: Busy" with muted styling
- Clicking an event shows a popover with full details (if visible) or just "Busy"
- Custom-built grid component (no third-party calendar library)

## 2. Empty State (No Connections)

When no household members have connected a calendar:

```
┌──────────────────────────────────────────────┐
│                                              │
│          📅 No calendars connected           │
│                                              │
│   Connect your Google Calendar to see your   │
│   household's events in one place.           │
│                                              │
│        [Connect Google Calendar]             │
│                                              │
└──────────────────────────────────────────────┘
```

## 3. Connect Google Calendar Flow

**Trigger:** "Connect Calendar" button on calendar page

**Steps:**

1. User clicks "Connect Google Calendar"
2. Frontend calls `POST /api/v1/calendar/connections/google/authorize` with the callback redirect URI
3. Frontend receives the Google OAuth authorize URL
4. Frontend navigates to Google's consent screen (full page redirect)
5. User grants calendar read access on Google's consent screen
6. Google redirects to `/api/v1/calendar/connections/google/callback`
7. Calendar-service validates state, exchanges code for tokens, creates connection
8. Calendar-service redirects to `/app/calendar?connected=true`
9. Frontend shows success toast: "Google Calendar connected. Syncing events now..."
10. Calendar-service triggers immediate first sync (async goroutine) — events typically appear within seconds, not waiting for the 15-minute background tick

**Post-connection:**

After the OAuth flow completes and the user lands back on `/app/calendar?connected=true`:
- Show success toast: "Google Calendar connected. Events will appear after first sync."
- Automatically navigate to a calendar selection panel showing all the user's Google Calendars with toggles (all enabled by default)
- User can disable calendars they don't want on the household view (e.g., "Holidays in United States")

**Error handling:**

- If user denies consent: redirect to `/app/calendar?error=access_denied`, show toast: "Calendar access was not granted."
- If OAuth state is invalid: redirect to `/app/calendar?error=auth_failed`, show toast: "Connection failed. Please try again."

## 4. Calendar Selection

**Trigger:** After initial connection, or via a "Manage Calendars" button on the calendar page

**Layout:**

```
┌──────────────────────────────────────┐
│  My Google Calendars                 │
│                                      │
│  ☑ Personal (primary)                │
│  ☑ Work                              │
│  ☐ Holidays in United States         │
│  ☑ Family                            │
│                                      │
│                            [Done]    │
└──────────────────────────────────────┘
```

Toggling a calendar off hides its events from the household view immediately (no re-sync needed — events remain in DB but are filtered out).

## 5. Disconnect Flow

**Trigger:** Settings or calendar page connection management

**Steps:**

1. User clicks "Disconnect" next to their Google Calendar connection
2. Confirmation dialog: "Disconnect Google Calendar? Your events will be removed from the household calendar."
3. On confirm: `DELETE /api/v1/calendar/connections/{id}`
4. Events removed from household calendar immediately
5. Success toast: "Google Calendar disconnected."

## 6. Connection Status Indicators

Visible on the calendar page (small status area):

| Status | Display |
|--------|---------|
| `connected` | "Connected · Last synced 5 min ago" |
| `syncing` | "Syncing..." with spinner |
| `error` | "Sync error · Last synced 2 hours ago" with retry option |
| `disconnected` | "Disconnected — Google access was revoked" with reconnect option |

## 7. Event Popover

Clicking an event shows a popover:

**Visible event (own or public):**

```
┌─────────────────────────┐
│ Team Standup             │
│ 9:00 AM – 9:30 AM       │
│ 📍 Zoom                 │
│ Jane Doe                 │
│                          │
│ Daily sync meeting       │
└─────────────────────────┘
```

**Private event (other user):**

```
┌─────────────────────────┐
│ Busy                     │
│ 12:00 PM – 1:00 PM      │
│ John Doe                 │
└─────────────────────────┘
```

## 8. Navigation

- **Sidebar:** New "Calendar" entry in the navigation sidebar
- **Week navigation:** Left/right arrows or buttons to move between weeks
- **Today button:** Jumps back to the current week with today visible
