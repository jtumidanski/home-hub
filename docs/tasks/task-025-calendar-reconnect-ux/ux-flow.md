# UX Flow — Calendar Reconnect

## States

The calendar connection row in the household calendar settings panel can be in one of four visible states.

### Healthy (`connected` or `syncing`)

```
● [color]  Jane Smith  ·  Last synced 2:14 PM        [↻ Sync] [⛓ Disconnect]
```

No badge, no error message. Sync button enabled. Behavior unchanged from today.

### Transient error (`error`)

```
● [color]  [⚠ Sync issues] Jane Smith                [↻ Sync] [⛓ Disconnect]
            Couldn't reach Google. Retrying automatically.
            Tried 4 min ago · Last success 6 hours ago        [Reconnect anyway]
```

- Amber badge with "Sync issues" label.
- User-facing message from §4.6 of the PRD on a second line.
- Both timestamps shown when they differ. If `lastSyncAttemptAt == lastSyncAt`, show only one.
- "Reconnect anyway" is a tertiary button — the system is still retrying, but the user can force the OAuth dance if they're impatient.
- Manual sync button remains enabled.

### Hard disconnect (`disconnected`)

```
● [color]  [✕ Disconnected] Jane Smith               [⛓ Disconnect]
            Access was revoked from your Google account.
            Reconnect to resume syncing.
            Tried 12 min ago · Last success yesterday
            [Reconnect]
```

- Red badge.
- User-facing message specific to the `error_code`.
- Prominent "Reconnect" button (primary style, full prominence within the row).
- Manual sync button hidden — sync cannot succeed without new tokens.

### Disconnected with unknown reason (legacy rows or pre-classification)

```
● [color]  [✕ Disconnected] Jane Smith               [⛓ Disconnect]
            This calendar is disconnected.
            [Reconnect]
```

For existing connections that were marked `disconnected` before this change shipped (no `error_code`), fall back to a generic message. They become first-class once the user reconnects, since the new code paths populate the fields going forward.

## Reconnect interaction

1. User clicks **Reconnect** on an `error` or `disconnected` row.
2. Frontend calls `useReauthorizeCalendar`, which `POST`s to `/calendar/connections/google/authorize` with `reauthorize: true` and the current page as redirect target.
3. The response includes the Google OAuth URL; the frontend redirects the browser to it (existing behavior).
4. User completes the Google consent screen.
5. Google redirects back to `/calendar/connections/google/callback`.
6. Backend callback handler:
   - Validates state, exchanges code for tokens (existing).
   - Encrypts and stores tokens (existing).
   - Resets `status` → `connected`.
   - Clears `error_code`, `error_message`, `last_error_at`, `consecutive_failures`.
   - Triggers an immediate sync (existing).
7. Browser redirects to `/app/calendar?connected=true`.
8. The connection row re-renders in the healthy state.

If the OAuth dance fails (user cancels, token exchange errors), the existing `?error=...` redirect surfaces a top-of-page toast as it does today. The connection row remains in its previous state.

## Edge cases

- **Transient → success**: counter resets, error fields clear, status returns to `connected`. UI re-renders without the error block on the next poll.
- **Transient → hard**: when classification flips mid-stream (e.g., several network errors followed by an `invalid_grant`), the hard error wins immediately and the row jumps from amber to red with an updated message.
- **User clicks Reconnect on a transient error**: works fine. The reauthorize endpoint doesn't care that the connection is technically still alive; it issues a fresh consent and replaces the tokens.
- **Multiple connections on one household**: each row is independent. One user's broken Google account does not visually affect another household member's healthy calendar.
