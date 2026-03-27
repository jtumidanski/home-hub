# Calendar Event CRUD — UX Flow

## Re-Authorization Flow

```
Calendar Page (existing connection without write access)
  │
  ├─ Banner: "Upgrade your calendar connection to add and edit events"
  │   └─ [Upgrade Access] button
  │       └─ POST /authorize with reauthorize=true
  │           └─ Redirect to Google consent screen (full calendar scope)
  │               └─ On success: redirect back, connection now has writeAccess=true
  │               └─ On decline: redirect back, read-only preserved, banner remains
  │
  └─ All existing read-only features continue working unchanged
```

## Event Creation Flow

### Via "Add Event" Button

```
Calendar Page
  │
  ├─ [+ Add Event] button (top bar, near navigation)
  │   └─ Opens EventFormDialog (empty form)
  │       ├─ Title (required, text input)
  │       ├─ All-day toggle
  │       │   ├─ Off: Date + time pickers for start and end
  │       │   └─ On: Date-only pickers for start and end
  │       ├─ Recurrence selector
  │       │   └─ Options: None, Daily, Weekly, Weekdays (Mon-Fri), Monthly, Yearly
  │       ├─ Location (optional, text input)
  │       ├─ Description (optional, textarea)
  │       ├─ Calendar picker (dropdown)
  │       │   └─ Lists user's visible calendar sources
  │       │   └─ Default: primary calendar
  │       │   └─ Only shown if user has write access on a connection
  │       ├─ [Cancel] — closes dialog
  │       └─ [Create Event] — submits
  │           ├─ Loading spinner on button
  │           ├─ POST to create endpoint
  │           ├─ Post-mutation sync completes
  │           ├─ Dialog closes
  │           ├─ Calendar view refreshes (query invalidation)
  │           └─ Toast: "Event created"
  │
  └─ Button disabled/hidden if no connections have write access
```

### Via Click on Calendar Grid

```
Calendar Grid (empty time slot)
  │
  └─ Click on empty area
      └─ Opens EventFormDialog pre-filled:
          ├─ Start: clicked time slot (rounded to nearest 15 min)
          ├─ End: start + 1 hour
          └─ All other fields empty/default
```

## Event Edit Flow

```
Calendar Grid
  │
  └─ Click on owned event
      └─ EventPopover (existing, extended)
          ├─ Event details (title, time, location, description)
          ├─ [Edit] button (only for owner, only if connection has write access)
          │   └─ Opens EventFormDialog pre-filled with event data
          │       ├─ If recurring event:
          │       │   └─ Prompt: "Edit this event only" / "Edit all events in series"
          │       │       ├─ "This event" → scope=single, form shows instance data
          │       │       └─ "All events" → scope=all, form shows series data
          │       │           (recurrence rule is not editable — only content fields)
          │       ├─ Edit fields, then [Save Changes]
          │       │   ├─ Loading spinner
          │       │   ├─ PATCH to update endpoint
          │       │   ├─ Post-mutation sync
          │       │   ├─ Dialog closes, popover closes
          │       │   ├─ Calendar refreshes
          │       │   └─ Toast: "Event updated"
          │       └─ [Cancel] — closes dialog, returns to popover
          │
          └─ [Delete] button (only for owner, only if connection has write access)
              └─ If recurring event:
              │   └─ Prompt: "Delete this event only" / "Delete all events in series"
              └─ Confirmation dialog: "Are you sure you want to delete [title]?"
                  ├─ [Cancel] — closes dialog
                  └─ [Delete] — submits
                      ├─ DELETE to endpoint
                      ├─ Post-mutation sync
                      ├─ Popover closes
                      ├─ Calendar refreshes
                      └─ Toast: "Event deleted"
```

## Event Form Dialog — Field States

| Field | Create (timed) | Create (all-day) | Edit (timed) | Edit (all-day) |
|-------|---------------|------------------|--------------|----------------|
| Title | Empty, focused | Empty, focused | Pre-filled | Pre-filled |
| All-day | Off | On | Off | On |
| Start date | Today | Today | Event date | Event date |
| Start time | Next hour | Hidden | Event time | Hidden |
| End date | Today | Today + 1 | Event date | Event date |
| End time | Start + 1hr | Hidden | Event time | Hidden |
| Recurrence | None | None | Current rule | Current rule |
| Location | Empty | Empty | Pre-filled | Pre-filled |
| Description | Empty | Empty | Pre-filled | Pre-filled |
| Calendar | Primary | Primary | Read-only (shows source name) | Read-only |

## Error States

| Scenario | User Experience |
|----------|----------------|
| No connections | "Connect your Google Calendar to add events" with connect button |
| No write access | "Upgrade your calendar connection to add events" with upgrade button |
| Google rejects write (read-only calendar) | Toast error: "This calendar doesn't allow new events. Try a different calendar." |
| Network error during create/edit/delete | Toast error: "Failed to [create/update/delete] event. Please try again." |
| Post-mutation sync fails | Event still created on Google; toast warning: "Event saved but may take a moment to appear" |
| Token expired during write | Auto-refresh token and retry (existing behavior); if refresh fails, prompt re-authorization |
