---
name: Home Hub
description: A warm, calm household organiser that reads like a well-kept kitchen notebook, on a phone, a desktop or a wall kiosk.
colors:
  background: "oklch(0.985 0.008 85)"
  foreground: "oklch(0.24 0.02 55)"
  card: "oklch(0.995 0.004 85)"
  primary: "oklch(0.53 0.13 45)"
  primary-foreground: "oklch(0.985 0.01 85)"
  secondary: "oklch(0.93 0.025 75)"
  secondary-foreground: "oklch(0.28 0.025 55)"
  muted: "oklch(0.955 0.012 80)"
  muted-foreground: "oklch(0.47 0.025 60)"
  accent: "oklch(0.93 0.035 145)"
  accent-foreground: "oklch(0.32 0.06 145)"
  destructive: "oklch(0.52 0.19 25)"
  success: "oklch(0.48 0.1 150)"
  warning: "oklch(0.80 0.13 80)"
  warning-text: "oklch(0.50 0.11 70)"
  info: "oklch(0.50 0.08 235)"
  border: "oklch(0.89 0.015 75)"
  input: "oklch(0.62 0.02 65)"
  sidebar: "oklch(0.965 0.014 80)"
typography:
  display:
    fontFamily: "Fraunces Variable, Fraunces Fallback, Georgia, serif"
    fontSize: "3rem"
    fontWeight: 700
    lineHeight: 1.2
  headline:
    fontFamily: "Fraunces Variable, Fraunces Fallback, Georgia, serif"
    fontSize: "1.75rem"
    lineHeight: 1.2
    letterSpacing: "-0.01em"
  title:
    fontFamily: "Fraunces Variable, Fraunces Fallback, Georgia, serif"
    fontSize: "1.125rem"
    fontWeight: 500
    lineHeight: 1.5
    letterSpacing: "-0.01em"
  body:
    fontFamily: "Figtree Variable, Figtree Fallback, system-ui, sans-serif"
    fontSize: "0.875rem"
    lineHeight: 1.5
  label:
    fontFamily: "Figtree Variable, Figtree Fallback, system-ui, sans-serif"
    fontSize: "0.75rem"
    lineHeight: 1.5
rounded:
  sm: "calc(0.75rem * 0.6)"
  md: "calc(0.75rem * 0.8)"
  lg: "0.75rem"
  xl: "calc(0.75rem * 1.4)"
spacing:
  touch-min: "44px"
  control: "36px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
    height: "36px"
    padding: "0 10px"
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-foreground}"
    rounded: "{rounded.md}"
    height: "36px"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.xl}"
    padding: "16px"
  card-hover:
    backgroundColor: "{colors.accent}"
---

# Design System: Home Hub

## Overview

**Creative North Star: "The Kitchen Table"**

Warm paper, clay and herb green, set in a soft old-style serif for headings and a friendly geometric sans for everything that is read or tapped. The system is comfortable rather than dense, calm rather than busy, and both themes are first-class: light is cream paper, dark is warm brown-black, never blue-black or neon. Anti-references (PRODUCT.md): dense enterprise admin dashboards, neon smart-home dark UIs, pastel cottagecore.

One component set serves three contexts: a phone in one hand (coarse-pointer targets), a desktop planning session, and a wall kiosk read from 2-3 m (container-query type steps inside widgets). Content is the household's own data; nothing decorative fills space.

**Key Characteristics:**
- OKLCH tokens in `src/index.css`, one set for `:root` and one for `.dark`; components use token utilities only.
- Hue family: clay (45), herb green (145), ochre (80), warm neutrals (55-85).
- Two typefaces, self-hosted, no runtime third-party requests.
- Status is never colour alone: icon plus text.
- Light mode lifts surfaces with soft warm shadows; dark mode uses no shadows.

## Colors

Warm neutrals carry the surface, clay is the single action colour, herb green is the soft "selected / hover" wash. Values below are the shipped tokens (source of truth: `src/index.css`); the frontmatter lists the light-theme anchors.

### Token tables (as shipped)

| Token | Light | Dark |
|---|---|---|
| `--background` | `oklch(0.985 0.008 85)` | `oklch(0.19 0.012 55)` |
| `--foreground` | `oklch(0.24 0.02 55)` | `oklch(0.94 0.012 80)` |
| `--card` | `oklch(0.995 0.004 85)` | `oklch(0.23 0.014 55)` |
| `--popover` | `oklch(0.995 0.004 85)` | `oklch(0.25 0.015 55)` |
| `--primary` | `oklch(0.53 0.13 45)` | `oklch(0.72 0.12 50)` |
| `--primary-foreground` | `oklch(0.985 0.01 85)` | `oklch(0.20 0.02 50)` |
| `--secondary` | `oklch(0.93 0.025 75)` | `oklch(0.29 0.018 60)` |
| `--secondary-foreground` | `oklch(0.28 0.025 55)` | `var(--foreground)` |
| `--muted` | `oklch(0.955 0.012 80)` | `oklch(0.27 0.014 55)` |
| `--muted-foreground` | `oklch(0.47 0.025 60)` | `oklch(0.74 0.02 75)` |
| `--accent` | `oklch(0.93 0.035 145)` | `oklch(0.32 0.035 145)` |
| `--accent-foreground` | `oklch(0.32 0.06 145)` | `oklch(0.90 0.04 145)` |
| `--destructive` | `oklch(0.52 0.19 25)` | `oklch(0.71 0.16 25)` |
| `--destructive-foreground` | `oklch(0.985 0.01 85)` | `oklch(0.20 0.02 50)` |
| `--success` | `oklch(0.48 0.1 150)` | `oklch(0.72 0.11 150)` |
| `--success-foreground` | `oklch(0.985 0.008 85)` | `oklch(0.20 0.02 50)` |
| `--warning` | `oklch(0.80 0.13 80)` | `oklch(0.80 0.12 80)` |
| `--warning-foreground` | `oklch(0.26 0.04 60)` | `oklch(0.20 0.02 50)` |
| `--warning-text` | `oklch(0.50 0.11 70)` | `var(--warning)` |
| `--info` | `oklch(0.50 0.08 235)` | `oklch(0.72 0.08 235)` |
| `--info-foreground` | `oklch(0.985 0.008 85)` | `oklch(0.20 0.02 50)` |
| `--border` | `oklch(0.89 0.015 75)` | `oklch(0.32 0.015 55)` |
| `--input` (control border) | `oklch(0.62 0.02 65)` | `oklch(0.52 0.02 60)` |
| `--ring` | `oklch(0.53 0.13 45)` | `oklch(0.72 0.12 50)` |
| `--overlay` | `oklch(0.24 0.02 55 / 0.35)` | `oklch(0.10 0.01 55 / 0.6)` |
| `--sidebar` | `oklch(0.965 0.014 80)` | `oklch(0.17 0.012 55)` |
| `--chart-1` | `oklch(0.58 0.13 45)` | `oklch(0.74 0.13 45)` |
| `--chart-2` | `oklch(0.55 0.09 145)` | `oklch(0.72 0.09 145)` |
| `--chart-3` | `oklch(0.64 0.12 80)` | `oklch(0.78 0.12 80)` |
| `--chart-4` | `oklch(0.52 0.08 240)` | `oklch(0.70 0.08 240)` |
| `--chart-5` | `oklch(0.48 0.09 340)` | `oklch(0.70 0.09 340)` |
| `--on-color-ink` / `--on-color-paper` | `oklch(0.24 0.02 55)` / `oklch(0.985 0.008 85)` | not redefined (same in both themes) |
| `--calendar-busy` | `oklch(0.62 0.015 65)` | `oklch(0.66 0.015 65)` |
| `--calendar-busy-foreground` | `var(--on-color-ink)` | `var(--on-color-ink)` |
| `--calendar-package` | `oklch(0.54 0.09 185)` | `oklch(0.70 0.09 185)` |
| `--calendar-package-foreground` | `var(--on-color-paper)` | `var(--on-color-ink)` |

`--card-foreground`, `--popover-foreground` and every `--sidebar-*` role other than `--sidebar` itself alias the matching base token (`--sidebar-primary` = `--primary`, `--sidebar-accent` = `--accent`, `--sidebar-border` = `--border`, `--sidebar-ring` = `--ring`).

**Lightness nudged for contrast** (commit bf95a06, covered by the contrast test): `--chart-3` light 0.68 to 0.64 (was 2.88:1 on card), `--calendar-package` light 0.55 to 0.54 (paper text 4.44:1), `--calendar-busy` dark 0.55 to 0.66 (ink text 3.39:1).

### The Warning-Text Rule
`--warning` is an ochre fill (L 0.80) and fails as text on light surfaces. Warning-coloured text and icons use `--warning-text` (light `oklch(0.50 0.11 70)`; in dark it equals `--warning`). Use `--warning` for fills and borders only, with `--warning-foreground` on top. Example in the build: `StatusBanner` warning icon and the stale-weather timestamp use `text-warning-text`.

### The Single Action Colour Rule
Clay (`--primary`) is the only call-to-action colour and the focus ring. Herb green (`--accent`) is a quiet wash for hover, selection and active nav, never a button fill.

### Swatches (categorical colour, 16 hues)
Tracker/habit colours are tokens: `--swatch-<name>` (solid) and `--swatch-<name>-soft` (tint) for red, orange, amber, yellow, lime, green, emerald, teal, cyan, blue, indigo, violet, purple, fuchsia, pink, rose. Light solid is `oklch(0.62 C H)`, soft is `oklch(0.93 0.04 H)`; dark solid is `oklch(0.70 C H)`, soft is `oklch(0.30 0.04 H)`. Chroma and hue per colour are set in `index.css` and are identical across themes.

- **The Swatch Module Rule.** Components never spell a swatch class. They call `swatchClass(color, role)` from `src/lib/constants/swatches.ts`, roles `solid` (`bg-swatch-*`), `soft` (`bg-swatch-*-soft`) and `border` (`border-swatch-*`). The module spells every class as a whole literal so Tailwind's scanner sees it, and `TRACKER_COLORS` (16, typed against `TrackerColor`) is the ordered key list.
- **The Neutral Fallback Rule.** An unknown or empty colour string (including prototype keys) resolves to `bg-muted-foreground`, `bg-muted` or `border-muted-foreground` for solid, soft and border. A swatch never renders as `undefined`.
- Swatch hue alone does not carry meaning; pair it with a label or position.

### Calendar colours
- **User colours are data, not tokens.** Each connected calendar user has a `userColor` string from the API, applied inline (`style={{ backgroundColor: userColor }}`) on dots and chips in `event-popover`, `calendar-widget`, `user-legend`, `connection-status` and `calendar-selection-panel` (fallback `var(--calendar-busy)`). Do not mirror them into `index.css`.
- **Chip text is computed.** `eventChipColors` in `src/lib/calendar/event-colors.ts` picks `--on-color-ink` or `--on-color-paper` through `readableForeground` (`src/lib/color-contrast.ts`) so any user colour stays legible.
- **System events use token pairs:** non-owner "Busy" placeholders use `--calendar-busy` with `--calendar-busy-foreground`; package ETA events use `--calendar-package` with `--calendar-package-foreground` (sentinel `PACKAGE_EVENT_COLOR`).

### The Token-Only Rule
`src/__tests__/no-color-literals.test.ts` scans `components/features` and `pages` and fails on any hex-in-string, `rgb()/hsl()/oklch()` call, or Tailwind palette class (`bg-red-500`, `text-white`, `bg-black/30`, and so on). Only token utilities (`bg-primary`, `bg-swatch-red`, `border-l-swatch-blue`) pass. `components/ui` and `components/common` are not scanned but follow the same rule except for the documented exceptions below.

## Typography

**Display / headings:** Fraunces Variable (wght axis only; no opsz, SOFT or WONK), fallback `Fraunces Fallback` over local Georgia, then Georgia, serif. Utility `font-heading`.
**Body / UI:** Figtree Variable, fallback `Figtree Fallback` over local Arial, then system-ui, sans-serif. Set on `html` through `font-sans`.

**Character:** a soft serif gives the recipe-card warmth; a geometric sans with open apertures stays legible at 13-15 px and at kiosk sizes. Both fonts ship from `@fontsource-variable/*` (latin subset loaded by unicode-range, `font-display: swap`); the metric-matched fallbacks (`size-adjust`, ascent/descent overrides) prevent layout shift on swap.

### Hierarchy (scale from `@theme`)
- **Display** (3rem, line-height 1.2): kiosk headline metrics (`text-display`, bold, `tabular-nums`).
- **Headline** (`text-2xl` 1.75rem / `text-xl` 1.375rem, 1.2, tracking -0.01em): page titles (`PageHeader` h1 is `font-heading text-2xl`), kiosk widget titles.
- **Title** (`text-lg` 1.125rem, 1.5, weight 500, tracking -0.01em): `CardTitle`, section headings.
- **Body** (`text-base` 1rem and `text-sm` 0.875rem, 1.5): controls, lists, tables.
- **Label** (`text-xs` 0.75rem, 1.5).
- Scale: xs .75, sm .875, base 1, lg 1.125, xl 1.375, 2xl 1.75, 3xl 2.25, display 3 (rem). Line-height 1.5 up to `lg`, 1.2 from `xl`.
- Cook mode has its own fluid utilities (`text-cook-meta`, `-section`, `-step`, `-focus`, `-badge`, `size-cook-badge`), `clamp()`-based, for reading at arm's length.

### The 18px Serif Floor Rule
Fraunces is used only at 18 px (`text-lg`) or larger, and never for controls, labels, table cells or body copy. Below 18 px use Figtree.

### The Tabular Numerals Rule
Figtree's GSUB table includes `tnum` (and `pnum`); Fraunces has no `tnum` (only `liga`, `rvrn`). So no numeric fallback font exists (Geist was removed) and no `--font-numeric` token is defined. The build sets no `font-feature-settings` or `font-variant-numeric` globally. Instead the Tailwind `tabular-nums` utility is applied per element on times, counts, temperatures and kiosk metrics (task and reminder cards, weather, calendar widget times, summary widgets). Numerals therefore render in Figtree, never Fraunces. New numeric displays add `tabular-nums`.

## Layout

Comfortable density: 36 px desktop control height (`h-9`; inputs `h-8`, small `h-7`). A fixed sidebar plus content area; `html`, `body` and `#root` are `height: 100%` with `overflow: hidden`, so scrolling happens inside regions. Spacing follows the Tailwind 4px step scale.

### Coarse-pointer target rule
`@custom-variant coarse (@media (pointer: coarse))`. Every interactive control keeps at least a 24 px target on fine pointers (smallest icon button `icon-xs` is `size-6`; checkbox `size-6`; radio rows `min-h-6`; dropdown rows `min-h-9`) and grows on touch: buttons (all text sizes) `coarse:min-h-11`, icon buttons `coarse:size-11`, inputs/textarea/select `coarse:min-h-11`, sidebar items and group labels `coarse:min-h-11`, dropdown items `coarse:min-h-11`, radio rows `coarse:min-h-11`, checkbox gains a `coarse:before:-inset-2.5` hit area (44 px) without growing visually, list-row buttons such as `recipe-selector` use `coarse:min-h-11`. 44 px = `min-h-11` is the touch floor. Text inputs use `text-base` on coarse (`md:text-sm` otherwise) to avoid iOS zoom.

### Widget contract (container queries)
Kiosk widget cells are sized by saved layout data, not by viewport, so widget internals respond to their own container, not the viewport.
- The widget root `Card` carries `@container h-full`.
- Title: `CardTitle className="font-heading text-lg @xl:text-2xl"`.
- List rows and body text step `text-sm @xl:text-xl` (secondary text `text-xs @xl:text-base`; icons `size-5 @xl:size-6`; fixed time columns `w-20 @xl:w-28`). Headline metrics step `text-3xl @md:text-display` with `tabular-nums`.
- Loading state is a `Skeleton` card of the same shape; errors use `WidgetError` (`dashboard-widgets/widget-error.tsx`): an `@container h-full border-destructive` card with `role="alert"`, an `AlertTriangle` icon (`size-5 @xl:size-6`, `aria-hidden`) and `text-sm @xl:text-xl text-foreground` message, so status is icon plus text.
- Whole-card links: when a widget navigates, a `Link` wraps the Card as `group/<name> block h-full rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring`, and the Card carries `motion-safe:transition-colors group-hover/<name>:bg-accent`. A trailing `ChevronRight` (`size-4`, muted, `aria-hidden`) in `CardAction` signals navigation on calendar-style widgets.

## Elevation & Depth

Hybrid, theme-dependent. **Light:** two warm-tinted shadows. `--elevation-card: 0 1px 2px oklch(0.3 0.03 50 / 0.06), 0 2px 8px oklch(0.3 0.03 50 / 0.05)` (utility `shadow-card`) and `--elevation-overlay: 0 4px 12px oklch(0.3 0.03 50 / 0.10), 0 12px 32px oklch(0.3 0.03 50 / 0.12)` (`shadow-overlay`, popovers and dialogs). **Dark:** both elevation tokens are `none`; depth comes from the surface lightness step (background 0.19, card 0.23, popover 0.25) plus the card's `ring-1 ring-foreground/10` hairline. `--overlay` scrims dialogs.

- **The Tonal Dark Rule.** Never add a box-shadow in dark mode; use the surface step and hairline.

## Shapes

`--radius: 0.75rem`. Derived scale: sm 0.45rem, md 0.6rem (about 9.6 px, controls and buttons), lg 0.75rem (status banners), xl 1.05rem (cards, 15 px), 2xl-4xl above that. Badges and pills are `rounded-full`. Nothing is squared and nothing is bubbly. Cards have no border, only the `ring-1` hairline and `shadow-card`.

## Components

Components live in `src/components/ui` (shadcn on Base UI primitives, styled with cva) and `src/components/common`.

### Buttons
- **Shape:** `rounded-md`; default height 36 px, `xs` 24, `sm` 28, `lg` 40; icon sizes `size-8`/`size-6`/`size-7`/`size-9`, all `coarse:size-11`.
- **Variants:** default (clay fill, paper text, hover 80% opacity), outline, secondary, ghost, destructive, link.
- **Focus:** `focus-visible:outline-2 outline-offset-2 outline-ring` (destructive uses `outline-destructive`). Press nudges down 1 px. Transition uses `--duration-fast` and `ease-out`. Disabled is 50% opacity, no pointer events.

### Cards / Containers
- 15 px corner, `bg-card`, `py-4` and `gap-4` (`size="sm"` is `py-3`, `gap-3`), `shadow-card` plus hairline ring. `CardTitle` is `font-heading text-lg tracking-[-0.01em] font-medium`.

### Inputs / Fields
- Border uses `--input` (a 3:1 control border, darker than `--border`), transparent background (`dark:bg-input/30`), `rounded-md`, focus uses the same 2 px outline ring, invalid shows a `--destructive` border plus a 3 px 20% ring (40% in dark).

### Status banner and badges
- `StatusBanner` (`ui/status-banner.tsx`): `rounded-lg` border in the status colour, 10% tint fill, icon plus text; destructive is `role="alert"`, others `role="status"`. Icons: destructive/success/info use their colour, warning uses `text-warning-text`.
- `Badge` variants: default, success, warning, info, link, and others; filled with the token and its `-foreground`.

### Navigation
- Sidebar items: `text-sm font-medium`, active is `bg-sidebar-accent` with `text-sidebar-primary`, hover `bg-sidebar-accent/50`, `coarse:min-h-11`.

### Brand mark
- `components/common/brand-mark.tsx` ("Window light": ink house on a clay tile with an ochre window) mirrors `public/favicon.svg`. Default `size-8`.

## Do's and Don'ts

### Do:
- **Do** use token utilities (`bg-primary`, `text-muted-foreground`, `bg-swatch-*` via `swatchClass`) for all colour.
- **Do** put focus on every interactive element as a 2 px `--ring` outline, offset 2 px; the base layer provides it:
  `:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; }`
- **Do** honour reduced motion. The base layer ends with a global `@media (prefers-reduced-motion: reduce)` rule that sets `animation-duration` and `transition-duration` to 0.01ms, `animation-iteration-count` to 1 and `scroll-behavior: auto`, all `!important`, on `*, *::before, *::after`. New motion needs no extra code; hover colour fades on widget cards additionally use `motion-safe:transition-colors`. Motion tokens: `--duration-fast` 120ms, `--duration-base` 200ms, `--ease-out` `cubic-bezier(.2,.7,.2,1)`. Motion only signals state.
- **Do** convey status with an icon plus text, with colour as reinforcement only.
- **Do** keep touch targets at least 24 px and add `coarse:min-h-11` / `coarse:size-11` to any new control.
- **Do** use Fraunces only at `text-lg` or larger and add `tabular-nums` to numerals.
- **Do** build widgets as `@container` cards with `@xl:` steps and `WidgetError` for failures.

### Don't:
- **Don't** write palette classes, hex, `rgb()`, `hsl()` or `oklch()` literals in `components/features` or `pages`; the vitest guard fails the build.
- **Don't** use `--warning` as text colour; use `--warning-text`.
- **Don't** put calendar user colours into tokens, or hard-code ink/paper on them; compute with `eventChipColors`.
- **Don't** add shadows in dark mode.
- **Don't** use viewport breakpoints for widget internals; use container queries.
- **Don't** invent placeholder copy, stats or testimonials; only household data fills the UI (PRODUCT.md).

## Exceptions

- **`ALLOWED` in `src/__tests__/no-color-literals.test.ts`:** currently empty (`[]`). There are no permitted literal exceptions in `components/features` or `pages`, `PENDING` is also empty (migration complete), and a test fails if an `ALLOWED` entry goes stale. Any future entry must name the exact file, exact match and a reason, and be added here.
- **Brand mark and favicon image asset:** `src/components/common/brand-mark.tsx` carries literal `oklch()` fills (clay `oklch(0.53 0.13 45)`, ink `oklch(0.24 0.02 55)`, ochre `oklch(0.80 0.13 80)`), mirrored in `public/favicon.svg`. An image asset is a documented exception; it lives in `components/common`, outside the scanned roots, and the two files stay in sync by hand.
- **Calendar user colours:** runtime data from the API applied through inline `style`, not tokens (see Colors).

## Drift noted, not recorded as rules

- Whole-card widget links use `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring` (a ring with no offset) rather than the base 2 px offset outline. Recorded as shipped in the widget contract; new surfaces should prefer the base outline.

## Sidecar note

`.impeccable/design.json` includes `tonalRamp` arrays per colour. These are derived, illustrative OKLCH ramps synthesized around each canonical value for preview purposes. They are not shipped tokens; only the values in the token tables above exist in `src/index.css`.
