# User Avatars & Profile Personalization — Context

Last Updated: 2026-03-26

---

## Key Files

### Backend (auth-service)

| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/auth-service/internal/user/entity.go` | GORM entity with `AvatarURL` field | Add `ProviderAvatarURL` field; add data migration function |
| `services/auth-service/internal/user/model.go` | Immutable domain model with accessors | Add `providerAvatarURL` field and accessor |
| `services/auth-service/internal/user/builder.go` | Fluent builder for domain model | Add `SetProviderAvatarURL` method |
| `services/auth-service/internal/user/rest.go` | JSON:API RestModel with `avatarUrl` | Add `ProviderAvatarURL` field; update Transform to compute effective avatar |
| `services/auth-service/internal/user/resource.go` | Route registration (GET /users/me, GET /users) | Add PATCH /users/me handler |
| `services/auth-service/internal/user/processor.go` | Business logic (FindOrCreate, ByID, etc.) | Add Update method, UpdateProviderAvatar method, avatar format validation |
| `services/auth-service/internal/authflow/processor.go` | OIDC callback orchestration | Call UpdateProviderAvatar after FindOrCreate |
| `services/auth-service/cmd/main.go` | Service entrypoint, migration registration | No changes (AutoMigrate picks up entity changes) |

### Frontend

| File | Purpose | Changes Needed |
|------|---------|----------------|
| `frontend/src/types/models/user.ts` | User type definitions | Add `providerAvatarUrl` to `UserAttributes` |
| `frontend/src/services/api/auth.ts` | Auth API client (getMe, getUsersByIds) | Add `updateMe()` method |
| `frontend/src/lib/hooks/api/use-auth.ts` | React Query hooks (useMe) | Add `useUpdateMe()` mutation hook |
| `frontend/src/components/ui/user-avatar.tsx` | **New file** — Reusable avatar component | Create with three-tier fallback, sm/md/lg sizes |
| `frontend/src/pages/SettingsPage.tsx` | Settings page (Profile + Appearance) | Add Avatar section with DiceBear picker |
| `frontend/src/components/features/navigation/user-menu.tsx` | User menu dropdown trigger | Add UserAvatar at sm size |
| `frontend/src/pages/HouseholdMembersPage.tsx` | Household members list | Add UserAvatar at md size per member |
| `frontend/src/components/providers/auth-provider.tsx` | Auth context provider | May need to expose providerAvatarUrl |

### Shared

| File | Purpose | Relevance |
|------|---------|-----------|
| `shared/go/server/response.go` | JSON:API marshal helpers | No changes — existing MarshalResponse works |
| `shared/go/server/request.go` | JSON:API unmarshal helpers | Used for PATCH request body parsing |

## Key Decisions

### 1. DiceBear: CDN vs Bundled JS

**Decision: Use CDN URL scheme** (`https://api.dicebear.com/9.x/{style}/svg?seed={seed}`)

- No npm dependency to manage
- Smaller bundle size
- Initials fallback covers CDN outages
- Trade-off: requires external network access for avatars

### 2. Avatar storage format

**Decision: Store DiceBear descriptor string** (`dicebear:{style}:{seed}`) in `avatar_url`

- Frontend resolves to CDN URL for rendering
- Backend validates format on PATCH (prevents arbitrary URL injection)
- Clean separation: backend stores intent, frontend resolves display

### 3. Data migration strategy

**Decision: Idempotent startup function after AutoMigrate**

- GORM AutoMigrate adds the new column
- A separate function runs: copy `avatar_url` → `provider_avatar_url` WHERE `provider_avatar_url IS NULL AND avatar_url != ''`, then clear `avatar_url`
- Idempotent: safe to run multiple times (checks for NULL provider_avatar_url)

### 4. Effective avatar computation

**Decision: Compute in REST transform layer** (not in domain model)

- `rest.go` Transform function: if `model.AvatarURL() != ""` → use it; else use `model.ProviderAvatarURL()`
- Both raw values also available in response for frontend decision-making
- Domain model stays pure — just carries data

### 5. Initials color generation

**Decision: Deterministic hash of user ID** → HSL color

- Hash user UUID to get a hue value (0-360)
- Fixed saturation (60-70%) and lightness (45-55%) for WCAG AA contrast with white text
- Same user always gets same color regardless of platform

## Dependencies

### External
- **DiceBear API** (api.dicebear.com) — avatar image CDN, no API key needed
- **Google userinfo** — OIDC picture claim (already integrated)

### Internal
- **shared/go/server** — JSON:API marshaling (existing, no changes)
- **shared/go/auth** — JWT claim extraction for PATCH endpoint auth (existing)

### NPM (no new packages required if using CDN approach)
- If bundling: `@dicebear/core`, `@dicebear/adventurer`, `@dicebear/bottts`, `@dicebear/fun-emoji`

## Architecture Notes

- **User entity is NOT tenant-scoped** — avatar is global per user, not per household
- **auth-service owns user data** — no cross-service writes needed
- **Batch user lookup** (`GET /users?filter[ids]=...`) already returns `avatarUrl` — adding `providerAvatarUrl` is the only change needed for household members page to work
- **OIDC callback flow**: `resource.go handleCallback` → `authflow.Processor.HandleCallback` → `user.Processor.FindOrCreate` — the refresh call goes after FindOrCreate in the authflow processor
