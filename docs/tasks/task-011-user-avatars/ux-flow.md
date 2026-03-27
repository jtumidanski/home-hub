# User Avatars — UX Flow

## Avatar Display Hierarchy

```
Has user-selected avatar (dicebear:*)?
  → Yes: Render DiceBear SVG via CDN URL
  → No: Has provider avatar URL?
    → Yes: Render <img> with provider URL
      → On load error: Fall through to initials
    → No: Render initials in colored circle
```

## Initials Fallback

- **Letter selection**: First letter of `givenName` + first letter of `familyName`. If either is missing, use first two letters of `displayName`. If `displayName` is also empty, use first letter of `email`.
- **Color**: Deterministic based on user ID hash — pick from a palette of 8-10 distinct, accessible colors.
- **Shape**: Circle with white text, centered.

## Avatar Sizes

| Size | Pixels | Usage |
|------|--------|-------|
| `sm` | 32x32 | User menu trigger |
| `md` | 40x40 | Household member rows |
| `lg` | 80x80 | Settings page profile section |

## Settings Page: Avatar Section

Layout between the existing Profile and Appearance sections:

```
┌─────────────────────────────────────────┐
│ Avatar                                  │
│ Choose how you appear in Home Hub       │
│                                         │
│  ┌──────┐                               │
│  │ (lg) │  Current avatar               │
│  │avatar│  [Reset to provider image]    │
│  └──────┘  [Remove avatar]              │
│                                         │
│  ── Pick a new avatar ──────────────    │
│                                         │
│  Style: [Adventurer ▾]                  │
│                                         │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐  │
│  │ 1 │ │ 2 │ │ 3 │ │ 4 │ │ 5 │ │ 6 │  │
│  └───┘ └───┘ └───┘ └───┘ └───┘ └───┘  │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐  │
│  │ 7 │ │ 8 │ │ 9 │ │10 │ │11 │ │12 │  │
│  └───┘ └───┘ └───┘ └───┘ └───┘ └───┘  │
│                                         │
│  [Shuffle]                              │
│                                         │
└─────────────────────────────────────────┘
```

### Interactions

1. **Style selector**: Dropdown with options: Adventurer, Bottts, Fun Emoji. Changing style regenerates the grid.
2. **Avatar grid**: 12 options per style, generated with deterministic seeds (e.g., `{userId}-{index}`). Clicking one immediately saves it (optimistic update with PATCH call).
3. **Shuffle button**: Generates 12 new random seeds and re-renders the grid.
4. **Reset to provider image**: Only visible when user has a `providerAvatarUrl` AND has a user-selected avatar. Calls PATCH with `avatarUrl: ""`.
5. **Remove avatar**: Clears both user-selected avatar (PATCH with `avatarUrl: ""`). Falls back to initials. Only visible when user has an active avatar of any kind.

### Loading & Error States

- Grid shows skeleton placeholders while DiceBear images load
- Failed PATCH shows toast error, reverts optimistic update
- Failed image loads in the grid show a broken-image placeholder

## User Menu: Avatar Integration

```
Before:                    After:
┌──────────────┐          ┌──────────────────┐
│ Jane Doe     │          │ (sm)  Jane Doe    │
│ jane@ex.com  │          │ avatar jane@ex.com│
│──────────────│          │──────────────────│
│ Dark mode  ○ │          │ Dark mode  ○     │
│ Sign out     │          │ Sign out         │
└──────────────┘          └──────────────────┘
```

## Household Members Page

```
Before:                          After:
┌────────────────────────┐      ┌──────────────────────────────┐
│ Jane Doe               │      │ (md)  Jane Doe               │
│ jane@example.com       │      │ avatar jane@example.com      │
│ Joined: Jan 15, 2026   │      │ Joined: Jan 15, 2026         │
│ Role: [Owner ▾]        │      │ Role: [Owner ▾]              │
└────────────────────────┘      └──────────────────────────────┘
```
