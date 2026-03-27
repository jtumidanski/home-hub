# User Avatars & Profile Personalization — Task Checklist

Last Updated: 2026-03-26

---

## Phase 1: Backend — Data Model & Migration

- [ ] 1.1 Add `ProviderAvatarURL` field to `user/entity.go` (TEXT, nullable)
- [ ] 1.2 Add `providerAvatarURL` field and accessor to `user/model.go`
- [ ] 1.3 Add `SetProviderAvatarURL` to `user/builder.go`
- [ ] 1.4 Add `ProviderAvatarURL` to `user/rest.go` RestModel with JSON tag `providerAvatarUrl`
- [ ] 1.5 Update `Transform` to compute effective `avatarUrl` (user-selected > provider) and populate both fields
- [ ] 1.6 Write idempotent data migration function: copy `avatar_url` → `provider_avatar_url`, clear `avatar_url`
- [ ] 1.7 Register migration function to run after AutoMigrate in `cmd/main.go`
- [ ] 1.8 Update `Make` / `modelFromEntity` to map the new entity field
- [ ] 1.9 Add/update unit tests for entity, model, builder, rest transform

## Phase 2: Backend — OIDC Avatar Refresh

- [ ] 2.1 Add `UpdateProviderAvatar(userID uuid.UUID, url string) error` to `user/processor.go`
- [ ] 2.2 Add database update function in `user/processor.go` (update `provider_avatar_url` column only)
- [ ] 2.3 Call `UpdateProviderAvatar` in `authflow/processor.go` after `FindOrCreate`
- [ ] 2.4 Add tests: provider avatar updates on login; user-selected avatar is not overwritten

## Phase 3: Backend — PATCH /api/v1/users/me

- [ ] 3.1 Add avatar format validation function: accepts `dicebear:{style}:{seed}` or empty string; rejects all else
- [ ] 3.2 Add `UpdateAvatar(userID uuid.UUID, avatarURL string) (Model, error)` to `user/processor.go`
- [ ] 3.3 Create PATCH handler in `user/resource.go` — parse JSON:API body, validate, call processor, return user
- [ ] 3.4 Register `PATCH /users/me` route in `InitializeRoutes`
- [ ] 3.5 Add handler tests: valid DiceBear string, empty string, invalid format (400), unauthorized (401)

## Phase 4: Frontend — UserAvatar Component

- [ ] 4.1 Create `frontend/src/components/ui/user-avatar.tsx`
- [ ] 4.2 Implement DiceBear URL resolution (`dicebear:{style}:{seed}` → CDN SVG URL)
- [ ] 4.3 Implement provider image rendering with `<img>` and `onError` fallback
- [ ] 4.4 Implement initials fallback with deterministic color from user ID
- [ ] 4.5 Support sizes: sm (32px), md (40px), lg (80px)
- [ ] 4.6 Add accessible `alt` text (user display name)
- [ ] 4.7 Add component tests for all three tiers, error handling, sizes

## Phase 5: Frontend — API & Types

- [ ] 5.1 Add `providerAvatarUrl` to `UserAttributes` in `frontend/src/types/models/user.ts`
- [ ] 5.2 Add `updateMe(attributes)` to `frontend/src/services/api/auth.ts`
- [ ] 5.3 Add `useUpdateMe()` mutation hook in `frontend/src/lib/hooks/api/use-auth.ts`
- [ ] 5.4 Invalidate `authKeys.me()` and `userKeys.all` on successful mutation

## Phase 6: Frontend — Settings Page Avatar Section

- [ ] 6.1 Add Avatar card section to `SettingsPage.tsx` between Profile and Appearance
- [ ] 6.2 Display current avatar at lg size using `UserAvatar`
- [ ] 6.3 Build DiceBear style tabs (adventurer, bottts, fun-emoji)
- [ ] 6.4 Render grid of avatar options (12+ per style) with deterministic seeds
- [ ] 6.5 Add "Shuffle" button to regenerate seeds
- [ ] 6.6 Wire selection to `useUpdateMe()` with `avatarUrl: "dicebear:{style}:{seed}"`
- [ ] 6.7 Add "Reset to provider image" button (conditional on `providerAvatarUrl` being set and overridden)
- [ ] 6.8 Add "Remove avatar" button (calls updateMe with empty avatarUrl)
- [ ] 6.9 Show loading/success/error feedback via toast

## Phase 7: Frontend — Integration Points

- [ ] 7.1 Update `user-menu.tsx` to show `UserAvatar` at sm size in dropdown trigger
- [ ] 7.2 Update `HouseholdMembersPage.tsx` to show `UserAvatar` at md size per member row
- [ ] 7.3 Verify avatars render for batch-fetched users from `useUsersByIds`

## Phase 8: Testing & Verification

- [ ] 8.1 Run all backend unit tests — confirm pass
- [ ] 8.2 Run all frontend tests — confirm pass
- [ ] 8.3 Docker build auth-service — confirm success
- [ ] 8.4 Docker build frontend — confirm success
- [ ] 8.5 Manual E2E: login → see avatar in menu → go to Settings → pick DiceBear avatar → verify persistence → verify in members page → reset to provider → remove avatar → verify initials
- [ ] 8.6 Update auth-service docs (domain.md, rest.md, storage.md) for new field and endpoint
