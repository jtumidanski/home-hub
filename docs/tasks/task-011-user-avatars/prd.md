# User Avatars & Profile Personalization — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-26
---

## 1. Overview

Users currently see only their name and email in the navigation menu and throughout the app. The system already captures the OIDC provider's `picture` claim and stores it as `avatar_url` on the user entity, but it is never displayed.

This feature adds avatar display everywhere users appear (user menu, settings page, household members list) and introduces a DiceBear-style procedural avatar picker as an alternative. Users who don't have a provider image — or who prefer something else — can choose from a generated set. A user-selected avatar always takes precedence over the OIDC-provided one.

Additionally, the OIDC avatar will be refreshed on each login so provider-side changes are reflected automatically.

## 2. Goals

Primary goals:
- Display user avatars across all surfaces where users appear
- Allow users to pick a procedural avatar from the Settings page
- Refresh the OIDC avatar on every login
- Show initials in a colored circle as the fallback when no avatar exists

Non-goals:
- Custom image upload (requires file storage infrastructure)
- Avatar cropping or editing
- Per-household avatars (one avatar per user globally)
- Animated avatars

## 3. User Stories

- As a logged-in user, I want to see my avatar in the navigation menu so the UI feels personalized.
- As a user with a Google account, I want my Google profile picture shown automatically without any setup.
- As a user without a provider picture (or who wants something different), I want to pick a procedural avatar from a set of generated options on the Settings page.
- As a user, I want my selected avatar to override my OIDC provider picture.
- As a user, I want the option to reset back to my provider picture if I change my mind.
- As a household member, I want to see other members' avatars on the household members page so I can visually identify people.

## 4. Functional Requirements

### 4.1 Avatar Display Component

- Create a reusable `UserAvatar` component that renders in three tiers:
  1. **User-selected avatar** (if set) — a DiceBear procedural avatar identified by style + seed
  2. **OIDC provider image** (if available) — rendered from `avatarUrl`
  3. **Initials fallback** — first letter of given name + first letter of family name (or first two letters of display name), rendered in a colored circle with deterministic color derived from user ID
- Support multiple sizes: `sm` (32px, user menu), `md` (40px, member lists), `lg` (80px, settings profile)

### 4.2 Avatar Selection (Settings Page)

- Add an "Avatar" section to the existing Settings page between Profile and Appearance
- Display the user's current avatar at `lg` size
- Show a grid of procedural avatar options generated via DiceBear HTTP API or a bundled JS library
  - Use 2-3 DiceBear styles (e.g., `adventurer`, `bottts`, `fun-emoji`)
  - Generate a set of options per style using deterministic seeds
  - Allow the user to regenerate/shuffle for new seeds
- Selecting an avatar calls PATCH to update the user's avatar
- Include a "Reset to provider image" button (visible only when user has an OIDC picture and has overridden it)
- Include a "Remove avatar" option to clear back to initials-only

### 4.3 OIDC Avatar Refresh

- On each OIDC login callback, update the stored `avatar_url` from the provider's `picture` claim (currently only set on first creation via `FindOrCreate`)
- This refresh should NOT overwrite a user-selected procedural avatar — it updates the stored provider URL so it's current if the user resets to provider image

### 4.4 Avatar in User Menu

- Update the user menu dropdown trigger to show the `UserAvatar` component at `sm` size alongside or replacing the current text-only display

### 4.5 Avatar on Household Members Page

- Show `UserAvatar` at `md` size next to each member's name and email in the members list

## 5. API Surface

### 5.1 Auth-Service: Update User Profile

**PATCH `/api/v1/users/me`**

Request (JSON:API):
```json
{
  "data": {
    "type": "users",
    "attributes": {
      "avatarUrl": "dicebear:adventurer:seed123"
    }
  }
}
```

- `avatarUrl` accepts either:
  - A DiceBear descriptor string: `dicebear:{style}:{seed}` — the frontend resolves this to an image URL
  - `""` (empty string) — clears the user-selected avatar, falling back to OIDC image or initials
- Returns the full user resource (same shape as GET `/api/v1/users/me`)

### 5.2 Auth-Service: User Resource (existing, updated)

**GET `/api/v1/users/me`** — no changes to shape, but clarify semantics:
- `avatarUrl` returns the **effective** avatar: the user-selected value if set, otherwise the OIDC provider URL

Add a new field to distinguish:
- `providerAvatarUrl` — always contains the OIDC provider's picture URL (or empty), so the frontend knows whether "Reset to provider image" is available

### 5.3 Auth-Service: Batch User Lookup (existing, updated)

**GET `/api/v1/users?filter[ids]=...`** — same field additions as above (`providerAvatarUrl`)

## 6. Data Model

### Auth-Service: User Entity Changes

Add one column to the `users` table:

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `provider_avatar_url` | TEXT | YES | NULL | Raw URL from OIDC provider's `picture` claim |

Rename semantics of existing `avatar_url`:
- `avatar_url` — the user's chosen avatar (DiceBear descriptor string or empty). This is what the user explicitly sets.
- `provider_avatar_url` — the OIDC provider's picture URL, refreshed on each login.

Migration:
1. Add `provider_avatar_url` column
2. Copy current `avatar_url` values to `provider_avatar_url` (since all existing values came from OIDC)
3. Set `avatar_url` to empty string for all rows (no user has explicitly selected an avatar yet)

### Domain Model Changes

Update the `User` model and entity to carry both fields. The REST layer computes `avatarUrl` (effective) and `providerAvatarUrl` for the response.

## 7. Service Impact

### auth-service
- **Migration**: Add `provider_avatar_url` column, migrate existing data
- **OIDC callback**: Update `provider_avatar_url` on every login (not just creation)
- **User processor**: Add update logic for PATCH on user profile
- **REST handler**: Register PATCH `/api/v1/users/me`, add `providerAvatarUrl` to response
- **User model/entity**: Add `ProviderAvatarURL` field

### frontend
- **New component**: `UserAvatar` — reusable avatar with three-tier fallback
- **Settings page**: Add avatar selection section with DiceBear grid
- **User menu**: Integrate `UserAvatar` into dropdown trigger
- **Household members page**: Add `UserAvatar` to member rows
- **API service**: Add `updateMe()` method for PATCH `/api/v1/users/me`
- **Types**: Add `providerAvatarUrl` to `UserAttributes`

### account-service
- No changes required

## 8. Non-Functional Requirements

- **Performance**: DiceBear avatars should be resolved client-side (either via their CDN URL scheme or a bundled JS package) to avoid backend image generation. Provider avatar URLs are external (e.g., Google) — render with `<img>` and handle load failures gracefully (fall back to initials).
- **Security**: Validate that `avatarUrl` on PATCH only accepts the `dicebear:{style}:{seed}` format or empty string — do not allow arbitrary URLs to be set by the user (prevents stored XSS or phishing via avatar URLs).
- **Accessibility**: Avatar images must have appropriate `alt` text (user's display name). Initials fallback must have sufficient color contrast (WCAG AA).
- **Multi-tenancy**: Avatar is a property of the user, not tenant-scoped. The user entity in auth-service is not tenant-scoped (it's global). No tenant considerations needed.

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

- [ ] User menu displays avatar (provider image, DiceBear, or initials fallback) at small size
- [ ] Settings page shows current avatar and allows selecting a DiceBear procedural avatar from a grid
- [ ] Settings page allows resetting to provider image (when available)
- [ ] Settings page allows clearing avatar entirely (back to initials)
- [ ] Household members page shows avatars next to each member
- [ ] OIDC provider avatar is refreshed on every login
- [ ] User-selected avatar takes precedence over OIDC provider image
- [ ] Initials fallback renders with deterministic colored circle when no avatar exists
- [ ] PATCH `/api/v1/users/me` validates avatar format (only `dicebear:{style}:{seed}` or empty)
- [ ] Existing users' OIDC avatars are migrated to `provider_avatar_url`
- [ ] Avatar component handles image load failures gracefully (falls back to initials)
