# Tracker Mobile Visibility — UX Flow

## Today View Card States

### Unset (Needs Attention)
```
┌─────────────────────────────────┐
│▌ [color dot] Item Name          │
│▌                                │
│▌ [input controls]               │
│▌ [note input] [Skip]            │
└─────────────────────────────────┘
 ^
 3px left border in item's color
```

### Logged
```
┌─────────────────────────────────┐
│  [color dot] Item Name  [✓ logged] │
│                                 │
│  [input controls - value shown] │
│  [note input] [Skip]            │
└─────────────────────────────────┘
                          ^
                    green Badge with check icon
```

### Skipped
```
┌─────────────────────────────────┐
│  [color dot] Item Name  [skipped]  │
│                                 │
│  [input controls]               │
│  [note input] [Skip]            │
└─────────────────────────────────┘
 ^                        ^
 no left border     muted gray Badge
 muted card background
```

## Today View Progress Bar

```
┌─────────────────────────────────┐
│ Today — Wednesday, Apr 9        │
├─────────────────────────────────┤
│ [████████░░░░░░░░] 4/7 entries  │
├─────────────────────────────────┤
│ [item cards...]                 │
└─────────────────────────────────┘
```

## Calendar Mobile Day View

Same card treatments as Today view — colored left border for unset, Badge for logged/skipped, muted style for skipped.

## Visual Hierarchy (Most to Least Prominent)

1. **Unset** — Colored left border draws the eye, signaling "action needed"
2. **Logged** — Clean card with green badge, feels complete
3. **Skipped** — Muted background and gray badge, visually recedes
