# User Avatars & Profile Personalization — Implementation Plan

Last Updated: 2026-03-26

---

## Executive Summary

Add avatar display across all user-facing surfaces (user menu, settings, household members) with a three-tier fallback system: user-selected DiceBear procedural avatar > OIDC provider image > initials in a colored circle. The backend needs a new `provider_avatar_url` column, a PATCH endpoint for user profile updates, and OIDC login refresh logic. The frontend needs a reusable `UserAvatar` component, a DiceBear avatar picker on Settings, and integration into the user menu and members list.

## Current State Analysis

### Backend (auth-service)
- **User entity** has `avatar_url` (TEXT) — currently stores the OIDC provider's picture URL
- **FindOrCreate** sets `avatar_url` on first creation only; subsequent logins do not update it
- **No PATCH endpoint** exists for `/api/v1/users/me` — only GET is registered
- **REST model** exposes `avatarUrl` in JSON:API responses but has no `providerAvatarUrl` field
- **No update logic** exists in the user processor

### Frontend
- **UserAttributes** type already has `avatarUrl: string` and `UserUpdateAttributes` includes it — but nothing renders it
- **User menu** (`user-menu.tsx`) shows only text (display name + email)
- **Settings page** has Profile and Appearance sections — no Avatar section
- **Household members page** shows name, email, role — no avatar
- **No `UserAvatar` component** exists
- **No `updateMe()` API method** exists in the auth service client
- **No DiceBear dependency** installed

## Proposed Future State

### Backend
- `users` table gains `provider_avatar_url` column; existing `avatar_url` values migrate to it
- `avatar_url` stores user-selected DiceBear descriptor (`dicebear:{style}:{seed}`) or empty
- OIDC callback updates `provider_avatar_url` on every login (not just creation)
- New PATCH `/api/v1/users/me` endpoint validates avatar format and persists changes
- GET responses include both `avatarUrl` (effective) and `providerAvatarUrl`

### Frontend
- Reusable `UserAvatar` component with sm/md/lg sizes and three-tier fallback
- Settings page gains an Avatar section with DiceBear grid picker, reset, and remove
- User menu shows avatar at sm size
- Household members page shows avatar at md size
- New `updateMe()` API method and `useUpdateMe()` mutation hook

---

## Implementation Phases

### Phase 1: Backend — Data Model & Migration (auth-service)

Add the `provider_avatar_url` column and migrate existing data so the two avatar concepts are separated.

**Tasks:**
1. Add `ProviderAvatarURL` field to user entity, model, builder, and rest model
2. Write a GORM `AfterAutoMigrate` hook (or startup function) to copy `avatar_url` → `provider_avatar_url` and clear `avatar_url` for existing rows
3. Update `Transform` / `TransformSlice` to compute effective `avatarUrl` (user-selected if set, else provider) and expose `providerAvatarUrl`
4. Update unit tests for model/entity/rest changes

**Acceptance:** Entity has both columns; REST responses include `providerAvatarUrl`; existing data migrated correctly.

### Phase 2: Backend — OIDC Avatar Refresh (auth-service)

Update the OIDC callback flow so the provider picture is refreshed on every login without overwriting user-selected avatars.

**Tasks:**
1. Add an `UpdateProviderAvatar(userID, url)` method to the user processor
2. In `authflow.Processor.HandleCallbackWithUserInfo`, after `FindOrCreate`, call `UpdateProviderAvatar` with the latest picture URL
3. Add tests verifying: (a) provider avatar updates on login, (b) user-selected avatar is not overwritten

**Acceptance:** After login, `provider_avatar_url` matches the OIDC provider's current picture claim. A user-selected `avatar_url` remains unchanged.

### Phase 3: Backend — PATCH /api/v1/users/me Endpoint (auth-service)

Add the ability for users to update their avatar via a PATCH endpoint.

**Tasks:**
1. Add an `Update(userID, avatarURL)` method to the user processor that validates the format (`dicebear:{style}:{seed}` or empty string)
2. Create the PATCH handler in `resource.go` — parse JSON:API request body, validate, call processor, return updated user
3. Register the route in `InitializeRoutes`
4. Add handler tests for: valid DiceBear string, empty string (clear), invalid format (reject), unauthorized

**Acceptance:** PATCH `/api/v1/users/me` with `avatarUrl: "dicebear:adventurer:seed123"` persists and returns in response. Invalid formats return 400. Empty string clears selection.

### Phase 4: Frontend — UserAvatar Component

Build the reusable avatar component with three-tier fallback.

**Tasks:**
1. Install `@dicebear/core` and desired style packages (`@dicebear/adventurer`, `@dicebear/bottts`, `@dicebear/fun-emoji`) — OR use the DiceBear CDN URL scheme (`https://api.dicebear.com/9.x/{style}/svg?seed={seed}`)
2. Create `UserAvatar` component (`frontend/src/components/ui/user-avatar.tsx`) with props: `user` (or `avatarUrl`, `providerAvatarUrl`, `displayName`, `userId`), `size` (sm/md/lg)
3. Implement three-tier rendering:
   - If `avatarUrl` starts with `dicebear:` → resolve to DiceBear SVG URL
   - Else if `avatarUrl` is a URL → render `<img>` with `onError` fallback
   - Else → render initials circle with deterministic color from user ID
4. Handle image load errors gracefully (fall back to initials)
5. Ensure accessible `alt` text

**Acceptance:** Component renders correctly for all three tiers; handles load failures; supports sm/md/lg sizes.

### Phase 5: Frontend — API & Types

Wire up the frontend API layer for the new endpoint and updated response shape.

**Tasks:**
1. Add `providerAvatarUrl` to `UserAttributes` type
2. Add `updateMe(attributes)` method to `authService` calling PATCH `/api/v1/users/me`
3. Add `useUpdateMe()` mutation hook with cache invalidation for `authKeys.me()` and `userKeys`
4. Update `useAuth` context to expose `providerAvatarUrl` if needed

**Acceptance:** Types compile; `updateMe()` calls PATCH correctly; cache invalidates after mutation.

### Phase 6: Frontend — Settings Page Avatar Section

Add the avatar picker UI to the Settings page.

**Tasks:**
1. Add an "Avatar" card between Profile and Appearance in `SettingsPage.tsx`
2. Show current avatar at lg size using `UserAvatar`
3. Build a DiceBear grid: 2-3 style tabs, 12+ options per style with deterministic seeds
4. Add a "Shuffle" button to regenerate seeds
5. On selection, call `useUpdateMe()` with `avatarUrl: "dicebear:{style}:{seed}"`
6. Add "Reset to provider image" button (visible when user has provider avatar and has overridden it)
7. Add "Remove avatar" button to clear back to initials
8. Show loading/success/error states via Sonner toast

**Acceptance:** User can browse, select, shuffle DiceBear avatars; reset to provider image; remove avatar entirely. Changes persist across page reloads.

### Phase 7: Frontend — Integration Points

Display avatars in the user menu and household members page.

**Tasks:**
1. Update `user-menu.tsx` to show `UserAvatar` at sm (32px) size alongside/before the display name
2. Update `HouseholdMembersPage.tsx` to show `UserAvatar` at md (40px) size next to each member's name
3. Verify avatars load for batch-fetched users (the `useUsersByIds` hook)

**Acceptance:** Avatars visible in user menu dropdown trigger and next to each household member.

### Phase 8: Testing & Verification

**Tasks:**
1. Backend: unit tests for model, processor, handler (all three phases)
2. Frontend: component tests for `UserAvatar` (all three tiers, error handling, sizes)
3. Integration: verify OIDC login flow updates `provider_avatar_url`
4. Docker build verification for auth-service
5. Manual E2E walkthrough of full avatar lifecycle

**Acceptance:** All tests pass; Docker builds succeed; full lifecycle works end-to-end.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| GORM AutoMigrate doesn't handle data migration (copy + clear) | Medium | High | Write a separate startup migration function that runs after AutoMigrate; make it idempotent |
| DiceBear CDN availability | Low | Medium | Use CDN URL scheme (no bundled JS); initials fallback covers outages |
| External avatar images (Google) blocked by CSP | Low | Medium | Ensure CSP allows `img-src` from `*.googleusercontent.com` |
| Large DiceBear bundle size if using JS packages | Medium | Low | Prefer CDN URL approach over bundling; only bundle if latency is a concern |
| Race condition: OIDC refresh vs user avatar update | Low | Low | Provider avatar and user avatar are separate columns; no conflict |

## Success Metrics

- All 11 acceptance criteria from the PRD are met
- No regressions in existing auth flow
- Avatar component renders in <100ms for all tiers
- Settings page avatar picker is responsive and accessible (WCAG AA)

## Required Resources & Dependencies

- **DiceBear**: Either CDN (`api.dicebear.com`) or npm packages (`@dicebear/core`, style packages)
- **No new infrastructure**: No file storage, no image processing backend
- **No new services**: All changes in auth-service and frontend

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Data Model & Migration | M | None |
| Phase 2: OIDC Avatar Refresh | S | Phase 1 |
| Phase 3: PATCH Endpoint | M | Phase 1 |
| Phase 4: UserAvatar Component | M | None (parallel with backend) |
| Phase 5: API & Types | S | Phase 3, Phase 4 |
| Phase 6: Settings Avatar Section | L | Phase 4, Phase 5 |
| Phase 7: Integration Points | S | Phase 4, Phase 5 |
| Phase 8: Testing & Verification | M | All phases |

Phases 1-3 (backend) and Phase 4 (frontend component) can run in parallel. Phases 5-7 require both backend and component to be complete.
