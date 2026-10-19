# Task 057 — Impeccable Onboarding, Design Audit & Redesign — Design

Status: **Approved by user 2026-10-19 (PRD FR-5).** See "Approval" at the bottom.

PRD: `docs/tasks/task-057-impeccable-redesign/prd.md`

---

## 1. Intent

Home Hub is a household's shared home tool. It's used at the counter, in the
grocery aisle, and on a wall kiosk, and right now it looks like an untouched
shadcn starter: every token in `frontend/src/index.css` is zero-chroma grey,
type is stock Geist, and about 30 pages each made their own visual calls. The
point of this task is to make the app feel like *ours*, warm, calm and
domestic, while making it measurably accessible (WCAG 2.2 AA in both themes).
It also leaves a written system (`frontend/DESIGN.md`) so future features
don't drift again.

Success looks like this: someone opens any page on any surface in either theme
and it reads as the same product. Kiosk widgets are legible from across the
kitchen. Dark mode looks deliberately designed. A developer adding a page
reaches for tokens and `components/common/*`, not Tailwind palette classes.

### Anti-goals (failure even if it "works")

- **Cottagecore pastiche.** No gingham, no hand-drawn doodles, no script
  fonts, no beige-on-beige low contrast. Warm has to stay legible.
- **Split-the-difference polish.** The grey shadcn look is treated as an
  anti-reference, not something to tint slightly orange.
- **Inverted dark mode.** Dark is a separate designed palette (warm charcoal,
  a "hearth at night" feel), not `invert()` or swapped lightness.
- **Layout churn on the kiosk.** Saved `react-grid-layout` positions and sizes
  render with the same structure. Only what's inside a widget changes.
- **Behaviour drift.** No interaction changes unless an audit finding needs
  one, and each such change is named in its PR.

## 2. Product context (Impeccable onboarding seed)

The PRD defers the product-context interview to this phase. These answers
are **inferred from the PRD and the codebase. They are my calls, not
confirmed answers.** Correct them here before approving. Impeccable `init`
still runs its own one-round confirmation when it writes
`frontend/PRODUCT.md` (Stage 0). This section is the brief it confirms
against.

| Field | Proposed answer |
|---|---|
| Platform | `web` (responsive; installed-PWA-style use on phones; wall kiosk in a browser) |
| Audience | Members of one or a few households: adults managing meals, shopping, calendars, chores, habits, workouts, packages. Mixed technical comfort. Also anyone glancing at the kiosk while walking past. |
| Primary contexts | (1) Phone, one-handed, in a shop or kitchen (shopping mode, today's tasks). (2) Desktop planning sessions (meal plan, week workouts, calendar). (3) Wall kiosk, read at 2–3 m, rarely touched. |
| Tone | Warm, calm, competent, quietly friendly. Like a well-kept kitchen notebook. Not cute, not corporate. |
| Brand personality | Domestic, dependable, unhurried. |
| References (loved) | Paper planners and recipe cards; Things 3's calm hierarchy; Apple Home/Weather at-a-glance legibility; warm editorial food sites (serif headings, generous whitespace). *(my guess, so replace with your own)* |
| Anti-references | Stock shadcn grey (the current app); dense enterprise admin dashboards; neon "smart home" dark UIs; pastel cottagecore. |
| Content truth | Real household data only. No placeholder marketing copy, no fabricated testimonials or stats. |

## 3. Visual direction — "The Kitchen Table"

Creative north star: **The Kitchen Table.** It's the place the household
gathers to plan the week. Paper-warm surfaces, ink-dark text, clay and sage
accents, soft but not bubbly shapes, and quiet motion.

### 3.1 Colour (OKLCH)

Hue families: **clay/terracotta (≈45°)** is the primary action colour.
**Sage (≈145°)** is the secondary accent and success colour. **Ochre (≈80°)**
is warning and highlight. **Warm ink (≈55°, low chroma)** is used for text
and neutrals. Every neutral carries 0.008–0.025 chroma toward warm. None of
them is pure grey.

Contrast numbers below are estimates from OKLCH L. The Stage 1 token test
(§5.3) computes the real WCAG ratio through sRGB, and any pair that falls
short gets its L nudged, not its role changed.

**Light — "paper"**

| Token | Value | Notes |
|---|---|---|
| `--background` | `oklch(0.985 0.008 85)` | warm paper |
| `--foreground` | `oklch(0.24 0.02 55)` | warm ink, ~14:1 |
| `--card` / `--popover` | `oklch(0.995 0.004 85)` | lifted paper |
| `--card-foreground` / `--popover-foreground` | `= --foreground` | |
| `--primary` | `oklch(0.53 0.13 45)` | clay; ≥4.5:1 vs paper and vs its fg |
| `--primary-foreground` | `oklch(0.985 0.01 85)` | |
| `--secondary` | `oklch(0.93 0.025 75)` | oat |
| `--secondary-foreground` | `oklch(0.28 0.025 55)` | |
| `--muted` | `oklch(0.955 0.012 80)` | |
| `--muted-foreground` | `oklch(0.47 0.025 60)` | ≥4.5:1 on background *and* muted |
| `--accent` | `oklch(0.93 0.035 145)` | sage wash (hover/selected) |
| `--accent-foreground` | `oklch(0.32 0.06 145)` | |
| `--destructive` | `oklch(0.52 0.19 25)` | true red, hue-separated from clay |
| `--destructive-foreground` *(new)* | `oklch(0.985 0.01 85)` | |
| `--success` / `-foreground` *(new)* | `oklch(0.48 0.1 150)` / paper | |
| `--warning` / `-foreground` *(new)* | `oklch(0.80 0.13 80)` / `oklch(0.26 0.04 60)` | ochre fill, ink text |
| `--info` / `-foreground` *(new)* | `oklch(0.50 0.08 235)` / paper | |
| `--border` | `oklch(0.89 0.015 75)` | decorative dividers only |
| `--input` | `oklch(0.62 0.02 65)` | control boundary, ≥3:1 (SC 1.4.11) |
| `--ring` | `oklch(0.53 0.13 45)` | focus, ≥3:1 |
| `--chart-1..5` | clay `0.58 0.13 45`, sage `0.55 0.09 145`, ochre `0.68 0.12 80`, slate-blue `0.52 0.08 240`, plum `0.48 0.09 340` | each ≥3:1 vs card |
| `--sidebar` | `oklch(0.965 0.014 80)` | slightly deeper paper than the canvas |
| `--sidebar-*` | follow the main roles | `sidebar-primary` = clay |

**Dark — "hearth"**

| Token | Value | Notes |
|---|---|---|
| `--background` | `oklch(0.19 0.012 55)` | warm charcoal |
| `--foreground` | `oklch(0.94 0.012 80)` | warm off-white, ~15:1 |
| `--card` | `oklch(0.23 0.014 55)` | surfaces step *up* in L; no shadows |
| `--popover` | `oklch(0.25 0.015 55)` | |
| `--primary` | `oklch(0.72 0.12 50)` | lit clay |
| `--primary-foreground` | `oklch(0.20 0.02 50)` | |
| `--secondary` | `oklch(0.29 0.018 60)` | |
| `--secondary-foreground` | `= --foreground` | |
| `--muted` | `oklch(0.27 0.014 55)` | |
| `--muted-foreground` | `oklch(0.74 0.02 75)` | ≥4.5:1 on background, card, and muted |
| `--accent` | `oklch(0.32 0.035 145)` | |
| `--accent-foreground` | `oklch(0.90 0.04 145)` | |
| `--destructive` | `oklch(0.71 0.16 25)` | raised from 0.68: the danger badge read 4.46:1 in the rendered mockup |
| `--success` / `--warning` / `--info` | `0.72 0.11 150` / `0.80 0.12 80` / `0.72 0.08 235`, ink foregrounds | |
| `--border` | `oklch(0.32 0.015 55)` | |
| `--input` | `oklch(0.52 0.02 60)` | ≥3:1 |
| `--ring` | `oklch(0.72 0.12 50)` | |
| `--chart-1..5` | same hues, L raised to 0.70–0.78 | each ≥3:1 vs card |
| `--sidebar` | `oklch(0.17 0.012 55)` | sidebar sits *below* the canvas in dark |

**Status never relies on colour alone (FR-20).** Destructive, warning,
success and info always come with an icon or text label. The `Badge`
variants enforce this by taking an optional leading icon, and the audit
checks that callers use it.

### 3.2 Data-driven and categorical colour

There are three places where colour comes from data or from a persisted enum
rather than from design:

1. **Tracker colours** (`TrackerColor` enum in `src/types/models/tracker.ts`,
   persisted by the backend). The names stay unchanged, so the data model is
   unchanged. Their visual values move from Tailwind palette classes
   (`bg-red-500`, `bg-red-100 dark:bg-red-950`, ~60 uses concentrated in
   `components/features/tracker/*`) into a **swatch token ramp**:
   `--swatch-<name>` (solid) and `--swatch-<name>-soft` (cell wash), with
   both light and dark values. Each is tuned to the warm world: slightly
   reduced chroma, solids ≥3:1 vs card, and soft washes that keep ink text
   ≥4.5:1. The lookup lives in one module (`src/lib/constants/swatches.ts`)
   that maps names to utility classes (`bg-swatch-red`, `bg-swatch-red-soft`)
   registered in `@theme inline`.
2. **Calendar user colours** (`attrs.userColor`, a hex chosen per user and
   returned by the API). This is user data, so it's a **documented exception**
   in `DESIGN.md`. Text on it gets a computed foreground (ink or paper, whichever
   clears 4.5:1) from a small pure helper in `src/lib/color-contrast.ts`, so
   event chips stay readable whatever colour a user picked. The hard-coded
   literals `#9ca3af` (busy) and `#0d9488` (package overlay) become tokens
   `--calendar-busy` and `--calendar-package`, referenced as
   `var(--calendar-busy)` in `style` props.
3. **Carrier icons, recipe step segments, weather icons.** These are
   categorical. They map to chart or swatch tokens, not new literals.

### 3.3 Typography

| Role | Family | Use |
|---|---|---|
| Display / headings | **Fraunces Variable** (wght axis only, opsz fixed) | page titles, section headings, card titles ≥ 18 px, kiosk widget titles, empty-state headlines |
| UI / body | **Figtree Variable** | everything else: body, controls, tables, labels, numerals |

- Why this pair: Fraunces is a soft, slightly old-style serif, which gives
  the recipe-card and planner warmth. Figtree is a friendly geometric sans
  with open apertures that stays very legible at 13–15 px and at kiosk sizes.
  Geist (the current font) reads technical, so it's replaced and its
  dependency removed.
- Fraunces is **never used below 18 px or for UI controls**, because serif
  legibility drops quickly at small sizes.
- Numerals: `tabular-nums` on times, quantities, counts, and kiosk metrics.
  Stage 1 checks that Figtree's `tnum` feature exists in the fontsource
  build. If it's missing, the fallback is Geist for numerals only, and that
  call is recorded in DESIGN.md.
- Scale (rem, 16 px root): `xs .75 · sm .875 · base 1 · lg 1.125 · xl 1.375 ·
  2xl 1.75 · 3xl 2.25 · display 3`. Line-height 1.5 for body, 1.2 for headings.
  The heading tracking is −0.01em.
- Loading: `@fontsource-variable/fraunces` and
  `@fontsource-variable/figtree`, both latin subset only, imported in
  `index.css`, `font-display: swap`. Stage 1 hand-authors metric-matched
  fallback `@font-face` rules (`size-adjust`, `ascent-override`,
  `descent-override` over local Georgia / Arial) so swapping causes no layout
  shift. Budget: combined default payload ≤150 KB (NFR). If Fraunces
  variable exceeds its share (~90 KB), fall back to static
  `@fontsource/fraunces` 600 only.

### 3.4 Shape, elevation, density, motion, icons

- **Radius:** `--radius: 0.75rem` (up from 0.625). Controls get `rounded-md`
  (≈9.6 px), cards `rounded-xl` (≈15 px), and pills/badges `rounded-full`.
  Nothing is fully squared, and nothing is bubbly.
- **Elevation, light:** two warm-tinted shadows,
  `--shadow-card: 0 1px 2px oklch(0.3 0.03 50 / 0.06), 0 2px 8px oklch(0.3 0.03 50 / 0.05)`
  and a stronger `--shadow-overlay` for popovers/dialogs. **Dark:** no
  shadows. Elevation comes from a surface L step plus a hairline border.
- **Density:** comfortable. Desktop control height is 36 px (`h-9`). On
  `pointer: coarse` devices, primary buttons, list rows and nav items get a
  44 px minimum through a new Tailwind custom variant
  `@custom-variant coarse (@media (pointer: coarse))`, used in
  `button.tsx`, `sidebar.tsx`, and list-row primitives (FR-21). Every
  icon-only control is ≥24×24 everywhere, ≥44×44 on coarse.
- **Focus:** one global visible ring: 2 px `--ring` with a 2 px offset in
  `--background`, applied through `focus-visible` in the base layer and in
  each `components/ui/*` primitive. It replaces the current `outline-ring/50`
  (50% alpha fails 3:1).
- **Motion:** tokens `--duration-fast 120ms`, `--duration-base 200ms`,
  `--ease-out cubic-bezier(.2,.7,.2,1)`. Motion only signals state (open,
  close, insert, remove). There's no decorative motion. A global
  `@media (prefers-reduced-motion: reduce)` rule in the base layer collapses
  animation and transition durations to ~0 and disables
  `tw-animate-css` keyframes (FR-22).
- **Icons:** keep `lucide-react`. Stroke 1.75, sizes 16/20/24 from a shared
  convention, and they're always paired with text or an `aria-label`.
- **Brand mark & favicon:** these **change** to **"Window light"** (user's
  pick out of four options; see `comps/brand-mark-options.html`, option B).
  It's a clay tile with a warm-ink house silhouette and one ochre four-pane
  window: someone's home. The indigo tile and lit door in
  `components/common/brand-mark.tsx` and `public/favicon.svg` are replaced
  by the 32×32 geometry in `comps/brand-mark-options.html` (`#mB`).
  The window panes are expected to disappear at 16 px; the yellow window
  alone carries the mark there. The SVG keeps literal fills (an image asset
  is a documented exception), and the two files stay in sync.

### 3.5 Kiosk legibility

Kiosk widgets live in `react-grid-layout` cells whose size is set by saved
layout data. Viewport breakpoints don't fit, because the same widget is tiny
on a phone dashboard and huge on the wall. So **widget internals use CSS
container queries**. Each widget root becomes an `@container` (Tailwind 4
native, no dependency), and its type, spacing and icon size step up at
container widths (e.g. `@md:` / `@xl:`). On the kiosk the headline metric
renders at display size with Fraunces titles and `tabular-nums` values, and
the same widget in a small cell stays compact.

- Layout JSON, `lib/dashboard/*` seed-layout semantics, and grid
  props are untouched. `lib/dashboard` tests must pass **unmodified** (PRD
  acceptance).
- Dark theme is the expected wall default at night. Both themes are audited.

### 3.6 Comp — dashboard

**The rendered comp is `comps/dashboard.html`.** Open it directly in a
browser; the fonts are next to it. It uses the exact §3.1 tokens, the real
Fraunces/Figtree files, and the Window light mark, and shows desktop light,
mobile light and dark, kiosk dark, and the palette/type sheet. The user
reviewed it during this phase. This is the approved comp that
`impeccable-finish-reviewer` checks against (FR-17). The ASCII sketches
below pin hierarchy and placement only.

Desktop ≥1280 px, light:

```
┌──────────────┬───────────────────────────────────────────────────────────┐
│ [■] Home Hub │  Good morning, Sam                         ☾  (avatar)    │ ← Fraunces 2xl, muted date below
│  (clay tile) │  Monday 19 October                                         │
│              ├───────────────────────────────────────────────────────────┤
│ ▸ Dashboard  │ ┌ Today ─────────────────┐ ┌ Weather ──┐ ┌ Meals ──────┐  │
│   Calendar   │ │ 3 tasks · 2 reminders  │ │  14°      │ │ Tonight     │  │ ← card: paper lifted, shadow-card
│   Tasks      │ │ ○ Bins out       7:00  │ │  ☁ Light  │ │ Dal & rice  │  │   titles Fraunces lg, body Figtree
│   Meals      │ │ ○ Call vet       12:30 │ │  rain     │ │ ─────────── │  │
│   Shopping   │ │ ● Pay water  ✓ done    │ │  H17 L9   │ │ Tue  Tacos  │  │ ← status: icon + text, never colour only
│   Habits     │ └────────────────────────┘ └───────────┘ └─────────────┘  │
│   Workouts   │ ┌ This week ─────────────────────────────┐ ┌ Packages ─┐  │
│   ...        │ │ M  T  W  T  F  S  S   (sage = done)    │ │ 2 arriving│  │
│ ──────────── │ │ ▇  ▇  ▁  ·  ·  ·  ·                    │ │ ▸ UPS Tue │  │
│ ⚙ Settings  │ └────────────────────────────────────────┘ └───────────┘  │
└──────────────┴───────────────────────────────────────────────────────────┘
 sidebar: deeper paper; active item = clay text + sage wash pill
```

Mobile ≈390 px, dark:

```
┌──────────────────────────────┐
│ ≡   [■] Home Hub        (av) │  ← 44px targets
├──────────────────────────────┤
│ Good evening, Sam            │  ← Fraunces xl
│ Monday 19 October            │
│ ┌ Today ───────────────────┐ │  ← card L+0.04 over charcoal, hairline border
│ │ ○ Bins out          7:00 │ │  ← rows ≥44px
│ │ ○ Call vet         12:30 │ │
│ └──────────────────────────┘ │
│ ┌ Weather ─┐ ┌ Tonight ────┐ │
│ │ 14° ☁    │ │ Dal & rice  │ │
│ └──────────┘ └─────────────┘ │
│ ┌ Packages ────────────────┐ │
│ │ 2 arriving · UPS Tue     │ │
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

Kiosk (same dashboard, wall display, dark): same grid, and container queries
push the "Today" count and weather temperature to `display` size. Widget
titles go to Fraunces 2xl and list rows to `xl`.

## 4. Alternatives considered

| Option | Verdict |
|---|---|
| **A. Retint only.** Swap OKLCH values and keep Geist, radius and component styles. | Rejected. That's split-the-difference polish on the discarded look, and it fixes none of the audit-class problems (focus, targets, states). |
| **B. New visual world through tokens plus primitives, rolled out by area (chosen).** | Fits the PRD's staged PRs. Most of the change lands once in `index.css` + `components/ui/*`, so feature stages are mostly literal removal and state fixes. |
| **C. Per-area accent themes** (e.g. green for meals, blue for calendar). | Rejected. It fragments the identity the task exists to unify and multiplies contrast verification by area × theme. |
| **D. Token pipeline** (Style Dictionary or a JSON → CSS generator). | Rejected. It adds a build step and a dependency for one consumer. `index.css` stays the single source, and DESIGN.md + `.impeccable/design.json` document it. The contrast test (§5.3) gives the mechanical safety a pipeline would have. |
| **E. Replace shadcn/base-ui or Tailwind.** | Out of scope (PRD non-goal). |
| **Kiosk: separate "kiosk mode" flag/route** vs. **container queries** | A flag is a new feature and touches dashboard data. Container queries need no data or route change and also improve phone dashboards. Chose container queries. |
| **Serif headings vs. all-sans warm** (e.g. Nunito everywhere) | All-sans warmth from rounded fonts tends to read childish. A serif display face carries "domestic" with more maturity. Chose Fraunces + Figtree. *(Swap this if you'd rather go all-sans.)* |

## 5. Architecture

### 5.1 Where things live

| Concern | Location |
|---|---|
| Token values (light, dark), theme mapping, custom variants, base layer (focus, reduced motion, fonts) | `frontend/src/index.css` (single source of truth) |
| Primitive variants | `frontend/src/components/ui/*` (button, badge, card, dialog, input, select, textarea, popover, tooltip, sidebar, calendar, form, radio-group, skeleton, label, user-avatar) |
| Shared composed patterns | `frontend/src/components/common/*`: existing `error-card`, `error-page`, `dashboard-skeleton`, `list-filter-bar`, `brand-mark`, plus **new** `page-header.tsx` (Fraunces title, description, actions slot) and `empty-state.tsx` (icon, headline, body, action). These are added in Stage 1 because nearly every page needs them and their absence is the most likely audit-finding cluster. Stage 0's audit confirms or drops them before Stage 1 starts. |
| Swatch mapping | `frontend/src/lib/constants/swatches.ts` |
| Contrast helper (calendar user colours, tests) | `frontend/src/lib/color-contrast.ts`: pure OKLCH/hex → linear sRGB → WCAG ratio, ~40 lines, no dependency |
| Product truth | `frontend/PRODUCT.md` (Impeccable `init`) |
| Design system record | `frontend/DESIGN.md` + `frontend/.impeccable/design.json` sidecar (Impeccable `document`, scan mode, run on the final build per FR-3) |
| Impeccable config | `frontend/.impeccable/config.json`. The Impeccable project root is `frontend/`; `impeccable context` is run with cwd `frontend/`. |
| Audit findings | `docs/tasks/task-057-impeccable-redesign/design-audit.md` |

`theme-provider.tsx` (class on `<html>`, `localStorage` + account-service
preference) is **not modified** (FR-8). Dark values hang off the existing
`.dark` class and `@custom-variant dark`.

### 5.2 Token rules

- Every token has a light *and* dark value (NFR Theming). The test enforces
  this.
- New semantic roles are added to both `:root`/`.dark` and `@theme inline` so
  utilities exist (`bg-success`, `text-warning-foreground`, `bg-swatch-red-soft`,
  `shadow-card`, `font-heading`).
- `--font-heading` stops aliasing `--font-sans` and maps to Fraunces.
- Feature and page code may use only token utilities. Tailwind palette
  classes (`bg-red-500`, `text-yellow-600`, …) and colour literals are
  banned in `components/features/**` and `pages/**` once their stage lands
  (§5.3).

### 5.3 Mechanical guards (added in Stage 1, run by vitest so `tools/verify.sh` enforces them)

1. **Token contrast test:** `src/__tests__/design-tokens.test.ts` parses
   `index.css` `:root` and `.dark` blocks and asserts (a) every token appears
   in both, and (b) a declared pair table meets its threshold:
   fg/bg pairs ≥4.5, large-text/non-text pairs (input, ring, chart vs card,
   swatch solids) ≥3.0, swatch-soft vs foreground ≥4.5. This is the
   recorded method for FR-18 token-level contrast.
2. **No-literal test:** `src/__tests__/no-color-literals.test.ts` scans
   `components/features/**` and `pages/**` (excluding `__tests__`) for hex,
   `rgb(`, `hsl(`, `oklch(` literals and Tailwind palette colour classes. It
   uses an explicit allowlist that mirrors DESIGN.md exceptions. While
   rollout is in progress it also carries a **per-directory pending list**;
   each feature stage deletes its directories from that list. When the list
   is empty, the NFR is met. This makes "no literals" monotonic and
   reviewable per PR.

Both tests are pure node/vitest with no new dependencies.

### 5.4 Dependencies

| Change | Justification |
|---|---|
| **add** `@fontsource-variable/fraunces` | Display face (§3.3), self-hosted (NFR Security) |
| **add** `@fontsource-variable/figtree` | UI face (§3.3), self-hosted |
| **remove** `@fontsource-variable/geist` | Replaced (unless the `tnum` fallback in §3.3 triggers) |

No other runtime dependencies are added.

### 5.5 Accessibility verification method (FR-18–22)

| Criterion | Method |
|---|---|
| Contrast (tokens) | §5.3 test, both themes |
| Contrast (rendered, incl. user-coloured chips, disabled-readable, placeholders) | Impeccable audit plus browser spot-check on the running app (`scripts/local-up.sh`), sampled per stage; results tabled in `design-audit.md` |
| Focus visible / keyboard / focus trap and return | Keyboard walk of every route per stage. Dialog/popover focus behaviour comes from base-ui and is asserted in existing or new RTL tests where a primitive is wrapped. |
| Labels / accessible names | RTL `getByRole(..., { name })` assertions in updated tests, plus a grep pass for icon-only `Button size="icon"` without `aria-label` |
| Target size | `coarse:` variants in primitives plus a mobile-width keyboard/touch walk |
| Reduced motion | Base-layer media rule, checked with the emulated preference in the browser |

## 6. Rollout stages

Each stage is a PR that passes the code-review step
(`frontend-guidelines-reviewer` + `plan-adherence-reviewer`) and a flagless
`tools/verify.sh` exit 0. They merge in order. How stage branches relate to
the task branch (stacked vs. sequential) is a `plan.md` decision per
`docs/git-workflow.md`. My recommendation: sequential stage branches, each
cut from `main` after the previous stage merges, so every PR diff shows only
that stage.

| Stage | Scope | Exit |
|---|---|---|
| **0. Onboard & audit** (docs only) | Impeccable `init` → `frontend/PRODUCT.md` (confirmed against §2). Record baselines: production JS gzip size and current font payload from `main` at task start. Run the Impeccable `audit` over every `App.tsx` route + shell + `components/ui|common` + dialogs/toasts/states, on desktop/mobile/kiosk × light/dark, against the task-start codebase (FR-12). Write `design-audit.md`: ID, severity, surface, theme, location, fix, **owning stage** (assigned by surface → table below). The audit can be sharded by area across fresh agents, each given the artifact path plus its own scope. | `design-audit.md` committed. Every finding has an owning stage. |
| **1. Foundation** | Tokens (§3.1–3.4), fonts, base layer (focus, reduced motion, metric fallbacks), `components/ui/*` variants + `coarse:` targets, `components/common/*` incl. new `page-header` / `empty-state`, swatch module, contrast helper, both §5.3 tests (no-literal pending list = all feature dirs), dependency changes, brand mark + favicon. | Token test green. Primitive-level findings resolved. Font budget measured. |
| **2. Shell** | App shell, sidebar, mobile header/drawer, Login, Onboarding, Error404/error pages, toasts (sonner theming via tokens). | Shell findings resolved. |
| **3. Dashboards & kiosk** | Dashboard renderer/shell/management/designer, all `dashboard-widgets/*`, container-query internals. | `lib/dashboard` tests unmodified and green; kiosk audited in both themes. |
| **4. Food** | Meals, recipes (list/detail/form), ingredients, shopping lists/detail, wish list. | Dirs removed from pending list. |
| **5. Time** | Calendar (incl. user-colour chips and package overlay), tasks, reminders. | ″ |
| **6. Body & home** | Habits/tracker (swatch migration), workouts (all sub-routes), packages, weather. | ″ |
| **7. Household & finish** | Households, members, settings, data retention. Then `impeccable-finish-reviewer` against this design and fix every material issue (FR-17). Then `impeccable-documenter` → `frontend/DESIGN.md` + `.impeccable/design.json` from the shipped build (FR-3). Record final bundle/font numbers vs. baseline (≤+10% JS gzip, ≤150 KB fonts). | Pending list empty. Every `design-audit.md` row marked with its resolving commit. |

Stages 4–6 are independent of one another once Stage 2 lands. The plan may
parallelise their implementation, but they still merge one PR at a time.

### Interim states

Stage 1 changes the look of every page at once, since pages already consume
tokens, while pages not yet migrated still use palette classes. That's
acceptable: palette classes keep rendering, and nothing breaks visually, it's
just inconsistent. Each feature stage removes the inconsistency for its area.
No feature flag is used; the interim look is the reason stages merge in quick
succession.

## 7. Testing impact

- Tests that assert palette classes (e.g.
  `components/features/tracker/__tests__/calendar-grid.test.tsx`,
  calendar and package-overlay colour tests) get **updated** to assert swatch
  or token classes and vars, never deleted (NFR Testing).
- Prefer role/name queries over class assertions in anything touched, which
  lowers future churn.
- New tests: the two §5.3 guards, `color-contrast.ts` unit tests,
  `swatches.ts` mapping completeness (every `TrackerColor` has a swatch), and
  `page-header` / `empty-state` render tests.

## 8. Left to the builder (`/plan-task` and execution)

- Exact component-level class recipes inside each `components/ui/*` variant.
- Exact L nudges to make the §5.3 pairs pass, as long as roles, hues and the
  dark "surfaces step up" rule hold.
- Container-query breakpoints per widget.
- The metric-override numbers for fallback fonts.
- Task granularity within stages and the stage-branch mechanics.
- Whether a Stage 0 finding warrants a primitive beyond `page-header` /
  `empty-state`. Add it to Stage 1 and note it in `design-audit.md`.

## 9. Calls I made (check these while reading)

1. Product context in §2 is inferred, not interviewed.
2. North star "The Kitchen Table": clay primary, sage accent, ochre warning,
   warm-charcoal dark.
3. Fraunces + Figtree replaces Geist.
4. ~~Brand mark recoloured~~. Superseded: the user picked "Window light" (§3.4).
5. Kiosk legibility through container queries, not a kiosk mode.
6. New `page-header` and `empty-state` common components.
7. Two vitest guard tests as the mechanical enforcement of contrast and
   no-literals.
8. Calendar user-picked colours stay as data, with computed foregrounds, as
   a documented exception.
9. ~~Rendered comp deferred to Stage 1~~. Superseded: rendered during
   this phase (`comps/`).
10. Audit runs as Stage 0 inside execution, against the task-start codebase,
    rather than as an extra pre-plan phase. Stage grouping is fixed here, so
    findings slot into stages by surface.

## Decisions confirmed during design review

- **Rendered comp reviewed** (`comps/dashboard.html`).
- **Fraunces and the warm paper background are kept.** Impeccable's
  detector flags both as AI-default choices (`overused-font`,
  `cream-palette`). The user was told this and replied "i'm good with
  that". The Fraunces suppression is recorded in `.impeccable/config.json`.
  Stage 1 should carry the same decision for `cream-palette` if the detector
  raises it on `index.css`.
- **Brand mark: "Window light"** (option B of four), chosen by the user.
- **Dark `--destructive` raised to L 0.71.** The rendered mockup measured
  the danger badge at 4.46:1.
- The mockup also showed that mid-L chart colours can't carry text in either
  ink or paper. Charts and swatches never hold text labels; labels sit beside
  or below the colour.

Builder note: `impeccable hooks` resolved its config at the **repository
root** (`.impeccable/config.json`), not `frontend/`. Stage 0 must settle
whether `frontend/` or the root is the Impeccable project root (§5.1 assumes
`frontend/`), and move that file plus PRODUCT.md/DESIGN.md to match. Drop
the scratch-file `nested-cards` ignore; it points at an untracked mockup.

## Approval

- [x] Direction approved by user. Date: 2026-10-19  Notes / changes: approved
  as written, including the "Decisions confirmed during design review"
  section (Fraunces + paper background kept, Window light brand mark, dark
  destructive at L 0.71).
