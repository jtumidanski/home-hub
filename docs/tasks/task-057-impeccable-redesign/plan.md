# Task 057 — Impeccable Onboarding, Design Audit & Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the stock grey shadcn look with the approved "Kitchen Table" direction across the whole frontend, fix every finding of a whole-app Impeccable design audit, meet WCAG 2.2 AA in both themes, and finish with a `frontend/DESIGN.md` generated from the shipped build.

**Architecture:** Tokens first. `frontend/src/index.css` is the single source of colour, type, shape, elevation and motion. `components/ui/*` and `components/common/*` consume the tokens, and feature/page code consumes only token utilities. Two vitest guards (token contrast, no colour literals) make the rules mechanical so `tools/verify.sh` enforces them. Rollout is seven staged PRs (Stage 0+1 together, then 2–7), each with its own review and flagless gate.

**Tech Stack:** React 19, TypeScript 6, Vite 8, Tailwind 4 (`@theme inline`, `@custom-variant`, container queries), shadcn on `@base-ui/react`, vitest 4 + Testing Library (jsdom), `@fontsource-variable/*`, Impeccable plugin (`/impeccable:impeccable`).

**Spec:** `docs/tasks/task-057-impeccable-redesign/design.md` (approved 2026-10-19). PRD: `docs/tasks/task-057-impeccable-redesign/prd.md`. Approved comp: `docs/tasks/task-057-impeccable-redesign/comps/dashboard.html`. Read `context.md` in the same folder first.

## Global Constraints

- Frontend only. No change under `services/`, `shared/`, or to any JSON:API contract or persisted data shape (PRD §5, §6).
- `frontend/src/components/providers/theme-provider.tsx` is not modified (FR-8). Dark styling hangs off the existing `.dark` class.
- `frontend/src/lib/dashboard/**` (including its `__tests__`) is not modified at all (PRD acceptance: tests pass unmodified).
- No third-party runtime requests. Fonts come only from `@fontsource-variable/fraunces` and `@fontsource-variable/figtree`; `@fontsource-variable/geist` is removed unless the `tnum` fallback (Task 6) triggers.
- No other new runtime dependencies. No new dev dependencies.
- Budgets: production JS gzip ≤ baseline × 1.10; default font payload ≤ 150 KB (NFR).
- Fraunces never below 18 px (`text-lg`) and never on controls. Figtree for everything else.
- Every colour token has a light (`:root`) and dark (`.dark`) value. Theme-invariant tokens are only those in the `THEME_INVARIANT` list of Task 5.
- After a stage lands, its directories contain no hex/`rgb(`/`hsl(`/`oklch(` literals and no Tailwind palette colour classes (`bg-red-500`, `text-white`, `bg-black/30`, …), except entries in the `ALLOWED` list of Task 9, each mirrored in `DESIGN.md`.
- Tailwind class names are always written as whole literal strings, never assembled (`bg-swatch-${c}` is forbidden); Tailwind's scanner only sees literals.
- Status is never colour-only: success/warning/destructive/info always carry an icon or a text label.
- Pointer targets ≥ 24×24 CSS px everywhere; primary actions, list rows, nav items and icon-only controls ≥ 44×44 under `coarse:`.
- Motion only signals state; all motion collapses under `prefers-reduced-motion: reduce`.
- Tests are updated, never deleted, when markup changes. Prefer role/name queries in anything you touch.
- No absolute home paths in anything under `docs/`.
- Commits: conventional style (`feat(frontend): …`, `test(frontend): …`, `docs(task-057): …`), ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Implementer verification is module-local: from `frontend/`, `npx vitest run <paths>`, `npx tsc -b`, `npx eslint <paths>`. The flagless `tools/verify.sh` is run by `task-verifier` at stage close-out only.

## Review Focus

1. **A user picked an extreme or malformed calendar colour** (`#ffff00`, `#000000`, `""`, `"teal"`): event chips must stay readable (≥ 4.5:1) and never throw. Pinned in Task 22 (`event-colors.test.ts`).
2. **A tracker arrives with a colour outside the `TrackerColor` union** (older or newer backend): the UI must render a neutral swatch, never `undefined` in a class list. Pinned in Task 8 (`swatchClass` fallback test).
3. **A saved layout puts a widget in a 1×1 cell on the kiosk or a full-width cell on a phone**: content must scale with the cell, not the viewport. Pinned in Task 17 (every widget root carries `@container`; titles use `font-heading`).
4. **Long names in the new serif headings at 390 px** (a long recipe or household name): titles wrap instead of overflowing or pushing actions off-screen. Pinned in Task 12 (`PageHeader` title has `min-w-0 break-words`; actions wrap).
5. **The OS asks for reduced motion**: dialogs, drawers and toasts must not animate. Pinned in Task 7 (token test asserts the `prefers-reduced-motion` rule and its declarations exist in `index.css`).

---

## Branch and PR mechanics (read before any task)

- **PR 1 = Stage 0 + Stage 1**, on the existing task branch `task-057-impeccable-redesign`. Stage 0 is docs only, so it ships with the foundation it sizes.
- **Stages 2–7** each get a fresh branch cut from `main` *after* the previous stage's PR has merged, in the same worktree (`.worktrees/task-057-impeccable-redesign/`):
  ```bash
  git fetch origin
  git merge-base --is-ancestor <previous-stage-branch> origin/main   # must exit 0
  git switch -c task-057-impeccable-redesign-stage-<N> origin/main
  ```
  The first task of each stage runs this. If the ancestor check fails, STOP: the previous PR has not merged.
- Stages run strictly in order 2 → 7. Stages 4–6 are independent in code, but they share the `PENDING` list in `no-color-literals.test.ts` and `design-audit.md`, so parallel branches would conflict for no real gain.
- PRs are opened with `tea` (Forgejo), never `gh`. Merging is the user's call: the controller opens the PR and stops.
- Each stage's PR is a durable boundary. After opening it, the controller hands off: the user merges, runs `/clear`, then `/execute-task task-057`, which resumes at the first unchecked task.

## Stage close-out procedure (controller, after the stage's last task)

Run in this order for every stage. Do not skip steps for "small" stages.

1. **Rendered check.** Bring the app up with `scripts/local-up.sh` (`run_in_background: true`). If a browser tool is available in the session, check each of the stage's routes at 1280 px and 390 px (plus kiosk for Stage 3) in light and dark: keyboard walk (visible focus, dialogs trap and return focus), reduced-motion emulation, and contrast spot-checks of user-coloured chips, placeholders and disabled-but-readable text. Append one row per route to the "Accessibility verification log" in `design-audit.md`. If no browser tool is available, write the row as `not rendered — browser unavailable` and list those routes in the PR body as a manual check for the user. Never record a rendered check that did not happen.
2. **Gate.** Dispatch `task-verifier` with the flagless invocation `tools/verify.sh`. It must exit 0. Fix and re-run until it does.
3. **Review.** Invoke `superpowers:requesting-code-review` to dispatch `frontend-guidelines-reviewer` and `plan-adherence-reviewer` in parallel. Scope is this stage's tasks and this branch's diff against `origin/main`. Findings go to `docs/tasks/task-057-impeccable-redesign/audit.md`. Fix every blocking finding, then re-run step 2.
4. **Mark findings.** In a separate commit (`docs(task-057): mark stage <N> findings resolved`), fill the `Resolved` column of every `design-audit.md` row owned by this stage with the short SHA of the commit that fixed it. A commit cannot contain its own SHA, which is why this is a separate commit.
5. **PR.** Push, write the body to a temp file outside the repo, and run
   `tea pr create --base main --head <branch> --title "task-057 stage <N>: <stage name>" --description-file <file>`.
   The body lists: the stage scope, the finding IDs resolved, every behaviour change (FR-16; "none" if none), the verify result, any manual rendered checks owed, and for Stage 7 the final bundle and font numbers next to the baseline. End it with `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.
6. **Stop and hand off** as described above.

---

## Stage 0 — Onboard & audit (docs only; ships in PR 1)

### Task 1: Impeccable project root and PRODUCT.md

**Files:**
- Create: `frontend/PRODUCT.md` **or** `PRODUCT.md` (repo root), depending on the decision below
- Move/Modify: `.impeccable/config.json` (currently at repo root)
- Create: `docs/tasks/task-057-impeccable-redesign/design-audit.md` (skeleton only: the "Method" section)

**Interfaces:**
- Produces: the **Impeccable root** (`frontend/` or repo root), recorded as the first line of the "Method" section of `design-audit.md`: `Impeccable root: <path>`. Tasks 3 and 32 read it from there.

- [x] **Step 1: Decide the root.** Move `.impeccable/config.json` to `frontend/.impeccable/config.json` (keep `.impeccable/hook.cache.json` out of git; check `.gitignore`). From `frontend/`, run the Impeccable `context` command and confirm it resolves the config; then trigger the Impeccable hook (edit-and-revert any `.tsx` under `frontend/src`) and confirm it reads the moved config. If both resolve, the root is `frontend/`. If either does not, move the config back and the root is the repo root.
- [x] **Step 2: Clean the config.** Remove the `nested-cards` ignore entry that points at `.superpowers/brainstorm/...` (untracked scratch file). Keep the Fraunces `overused-font` entry and the `comps/dashboard.html` `nested-cards` entry.
- [x] **Step 3: Run `/impeccable:impeccable init`** with the Impeccable root as project root. Give it design.md §2 as the product-context brief. When it runs its confirmation round, ask the user to confirm or correct those answers. Do not answer for them. Save `PRODUCT.md` at the Impeccable root.
- [x] **Step 4: Write the `design-audit.md` skeleton** with these headings, in order: `# Task 057 — Design Audit`, `## Method` (first line `Impeccable root: <path>`), `## Baselines`, `## Findings`, `## Accessibility verification log`. Findings table header (used by every later task):
  `| ID | Sev | Surface | Theme | Location | Description | Fix | Stage | Resolved |`
- [x] **Step 5: Commit**
  ```bash
  git add -A .impeccable frontend/.impeccable frontend/PRODUCT.md PRODUCT.md docs/tasks/task-057-impeccable-redesign/design-audit.md
  git commit -m "docs(task-057): impeccable root, PRODUCT.md, audit skeleton"
  ```
  (`git add` will warn on paths that do not exist; that is expected for the root you did not pick.)

### Task 2: Baselines

**Files:**
- Modify: `docs/tasks/task-057-impeccable-redesign/design-audit.md` (`## Baselines`)

- [x] **Step 1: Confirm the frontend is still at task start.** `git diff --quiet 9e00203 HEAD -- frontend` must exit 0 (9e00203 is `main` at task start).
- [x] **Step 2: Measure.** From `frontend/`: `npm ci && npm run build`, then
  ```bash
  for f in dist/assets/*.js; do gzip -9 -c "$f" | wc -c; done | awk '{s+=$1} END {print s}'
  ls -l dist/assets/*.woff2
  ```
- [x] **Step 3: Record** in `## Baselines`: ref `9e00203`, JS gzip total in bytes, the Geist woff2 files a latin-only page loads and their byte total, and the exact commands above. Later stages compare against these numbers.
- [x] **Step 4: Commit** `docs(task-057): record bundle and font baselines`.

### Task 3: Whole-app design audit

**Files:**
- Create (temporary, deleted in Step 4): `docs/tasks/task-057-impeccable-redesign/audit-shards/shard-{a..f}.md`
- Modify: `docs/tasks/task-057-impeccable-redesign/design-audit.md` (`## Method`, `## Findings`)

**Interfaces:**
- Produces: finding IDs `DA-001`…`DA-NNN`, each with a `Stage` value in `1..7`. Every later "Resolve Stage N findings" task consumes the rows whose `Stage` is N.

The audit runs against the task-start codebase (FR-12). It is the controller's job to shard it: dispatch six fresh `general-purpose` agents in one message (brief them; do not fork), each with the paths of `prd.md`, `design.md`, `PRODUCT.md`, the Impeccable root, its scope, and its output file.

| Shard | Scope | Default owning stage |
|---|---|---|
| a | `components/ui/*`, `components/common/*`, dialogs/popovers/forms, toasts, empty/loading/error states, `features/navigation/*`, `LoginPage`, `OnboardingPage`, `Error404Page` | 1 (primitive) or 2 (shell) |
| b | `DashboardManagementPage`, `DashboardShell`, `DashboardRenderer`, `DashboardDesigner`, `pages/dashboard-designer/*`, `features/dashboards/*`, `features/dashboard-widgets/*`; kiosk rendering | 3 |
| c | Meals, Recipes (list/detail/form), Ingredients (list/detail), Shopping lists/detail, Wish list; `features/meals|recipes|ingredients` | 4 |
| d | Calendar, Tasks, Reminders; `features/calendar|tasks|reminders`; package overlay on the calendar | 5 |
| e | Habits/Tracker, all Workouts routes, Packages, Weather; `features/tracker|trackers|workout|workouts|packages|weather` | 6 |
| f | Households, Household members, Settings, Data retention; `features/households` | 7 |

- [x] **Step 1: Dispatch the six shards.** Each agent runs `/impeccable:impeccable audit` over its scope, judging every page at desktop ≥1280 px, mobile ≈390 px, and (shard b) kiosk, in light and dark, against design.md §1 anti-goals and FR-18–FR-22. If it can render the app, it does; otherwise it audits from source and says so. It writes rows in the findings-table format with `ID` left as `?`, and puts any finding that is really a primitive problem (button, badge, input, focus ring, token) in Stage 1 regardless of shard. Shard a also states whether `page-header` and `empty-state` are confirmed as needed (design.md §5.1) and whether any further primitive is warranted (design.md §8).
- [x] **Step 2: Merge.** Concatenate shard rows into `## Findings`, de-duplicate (same location and same problem → one row, all surfaces listed), number `DA-001…` in Stage then Severity order, and check every row has a Stage. In `## Method`, record how each shard audited (rendered or source) and the page-header/empty-state decision. If a primitive beyond those two is warranted, add it as a Stage 1 row; Task 12 picks it up.
- [x] **Step 3: Check coverage.** Every route in `frontend/src/App.tsx` appears in at least one row's Surface or in a `Method` line saying it was audited with no findings.
- [x] **Step 4: Delete `audit-shards/` and commit** `docs(task-057): whole-app design audit`.

---

## Stage 1 — Foundation (ships in PR 1)

### Task 4: Colour contrast helper

**Files:**
- Create: `frontend/src/lib/color-contrast.ts`
- Test: `frontend/src/lib/__tests__/color-contrast.test.ts`

**Interfaces:**
- Produces:
  ```ts
  export interface Rgb { r: number; g: number; b: number } // gamma-encoded sRGB, each clamped to 0..1
  export function parseColor(input: string): Rgb | null    // "#rgb" | "#rrggbb" | "oklch(L C H)" | "oklch(L C H / a)"; L may be 0..1 or "NN%"; alpha ignored; anything else → null
  export function relativeLuminance(c: Rgb): number        // WCAG 2.x definition
  export function contrastRatio(a: string, b: string): number // throws Error(`unparseable colour: ${x}`) if either is null
  export const INK = "oklch(0.24 0.02 55)"
  export const PAPER = "oklch(0.985 0.008 85)"
  export function readableForeground(bg: string): "ink" | "paper" // higher contrast of INK/PAPER against bg; unparseable bg → "ink"
  ```

- [x] **Step 1: Write the failing tests**
  ```ts
  describe("parseColor", () => {
    it("parses short and long hex", () => {
      expect(parseColor("#fff")).toEqual({ r: 1, g: 1, b: 1 });
      expect(parseColor("#000000")).toEqual({ r: 0, g: 0, b: 0 });
    });
    it("parses oklch white and black", () => {
      const w = parseColor("oklch(1 0 0)")!; expect(w.r).toBeCloseTo(1, 3); expect(w.g).toBeCloseTo(1, 3); expect(w.b).toBeCloseTo(1, 3);
      expect(parseColor("oklch(0% 0 0)")).toEqual({ r: 0, g: 0, b: 0 });
    });
    it("ignores alpha", () => { expect(parseColor("oklch(1 0 0 / 10%)")).not.toBeNull(); });
    it("returns null for garbage", () => { for (const s of ["", "teal", "#ggg", "rgb(1,2,3)"]) expect(parseColor(s)).toBeNull(); });
  });
  describe("contrastRatio", () => {
    it("is 21 for black on white", () => { expect(contrastRatio("#000", "#fff")).toBeCloseTo(21, 1); });
    it("is symmetric", () => { expect(contrastRatio("#777", "#fff")).toBeCloseTo(contrastRatio("#fff", "#777"), 5); });
    it("matches a known pair", () => { expect(contrastRatio("#767676", "#ffffff")).toBeCloseTo(4.54, 1); });
    it("throws on unparseable input", () => { expect(() => contrastRatio("nope", "#fff")).toThrow("unparseable colour: nope"); });
  });
  describe("readableForeground", () => {
    it("picks ink on light and paper on dark", () => {
      expect(readableForeground("#ffff00")).toBe("ink");
      expect(readableForeground("#000000")).toBe("paper");
    });
    it("falls back to ink when unparseable", () => { expect(readableForeground("")).toBe("ink"); });
  });
  ```
- [x] **Step 2: Run** `npx vitest run src/lib/__tests__/color-contrast.test.ts` → FAIL (module not found).
- [x] **Step 3: Implement.** OKLCH → sRGB follows the OKLab reference (Ottosson): `a = C·cos(h°)`, `b = C·sin(h°)`;
  `l_ = L + 0.3963377774a + 0.2158037573b`, `m_ = L − 0.1055613458a − 0.0638541728b`, `s_ = L − 0.0894841775a − 1.2914855480b`; cube each;
  `R = 4.0767416621l − 3.3077115913m + 0.2309699292s`, `G = −1.2684380046l + 2.6097574011m − 0.3413193965s`, `B = −0.0041960863l − 0.7034186147m + 1.7076147010s`;
  clamp linear to 0..1, then gamma-encode (`x ≤ 0.0031308 ? 12.92x : 1.055x^(1/2.4) − 0.055`). Luminance linearises back with the WCAG formula. No dependency.
- [x] **Step 4: Run** the test → PASS. `npx tsc -b` clean.
- [x] **Step 5: Commit** `feat(frontend): add WCAG colour contrast helper`.

### Task 5: Colour tokens and the token contrast test

**Files:**
- Modify: `frontend/src/index.css` (`:root`, `.dark`, `@theme inline`)
- Create: `frontend/src/__tests__/design-tokens.test.ts`

**Interfaces:**
- Consumes: `parseColor`, `contrastRatio`, `INK`, `PAPER` from Task 4.
- Produces (utilities later tasks rely on): `bg|text|border-{success,warning,info,destructive}` and their `-foreground`; `text-warning-text`; `bg-overlay`; `bg-swatch-<name>`, `bg-swatch-<name>-soft`, `border-l-swatch-<name>` for the 16 `TrackerColor` names; CSS vars `--calendar-busy`, `--calendar-busy-foreground`, `--calendar-package`, `--calendar-package-foreground`, `--on-color-ink`, `--on-color-paper`; `shadow-card`, `shadow-overlay`.
- Produces (test helpers exported from the test file's sibling `frontend/src/__tests__/css-tokens.ts`, reused by Tasks 6–7):
  ```ts
  export function readIndexCss(): string
  export function tokenBlock(css: string, selector: ":root" | ".dark"): Map<string, string> // "--name" → raw value
  export function resolveToken(css: string, theme: "light" | "dark", name: string): string  // follows var(--x) chains; dark falls back to :root
  ```

Token values: use design.md §3.1 verbatim where given. Where the design is silent, the values below are the brief; nudge L only (design.md §8) until the test passes.
- `--card-foreground`, `--popover-foreground`: `var(--foreground)`. `--sidebar-*` (both themes): `sidebar-foreground`=foreground, `sidebar-primary`=primary, `sidebar-primary-foreground`=primary-foreground, `sidebar-accent`=accent, `sidebar-accent-foreground`=accent-foreground, `sidebar-border`=border, `sidebar-ring`=ring, all as `var(...)`.
- Dark `--destructive-foreground`, `--success-foreground`, `--warning-foreground`, `--info-foreground`: dark ink, `oklch(0.20 0.02 50)`.
- `--warning-text` (new; ochre that is legible as text/icon on the canvas): light `oklch(0.50 0.11 70)`, dark `var(--warning)`. Warning-coloured text and icons use `text-warning-text`, never `text-warning`.
- `--overlay`: light `oklch(0.24 0.02 55 / 0.35)`, dark `oklch(0.10 0.01 55 / 0.6)`.
- `--on-color-ink`: `INK`, `--on-color-paper`: `PAPER` (theme-invariant; the test asserts they equal the constants).
- `--calendar-busy`: warm grey, light `oklch(0.62 0.015 65)`, dark `oklch(0.55 0.015 65)`; `--calendar-package`: teal ≈ hue 185, light `oklch(0.55 0.09 185)`, dark `oklch(0.70 0.09 185)`; each `-foreground` is whichever of ink/paper passes.
- Swatches: for each of the 16 names, hue = the Tailwind hue of that name, chroma ≈ 0.8 × Tailwind-500 chroma. Solid: light L ≈ 0.62, dark L ≈ 0.70. Soft: light L ≈ 0.93 C ≈ 0.04; dark L ≈ 0.30 C ≈ 0.04.
- Chart dark: same hues as light, L 0.70–0.78.
- `@theme inline` maps every new role as `--color-<role>: var(--<role>)`, swatches included.

- [x] **Step 1: Write the failing test** (`design-tokens.test.ts`), with these exact tables:
  ```ts
  const THEME_INVARIANT = ["--radius", "--duration-fast", "--duration-base", "--on-color-ink", "--on-color-paper"];
  const TEXT_PAIRS: [fg: string, bg: string][] = [ // ≥ 4.5 in both themes
    ["--foreground","--background"], ["--card-foreground","--card"], ["--popover-foreground","--popover"],
    ["--primary-foreground","--primary"], ["--primary","--background"], ["--secondary-foreground","--secondary"],
    ["--muted-foreground","--background"], ["--muted-foreground","--muted"], ["--muted-foreground","--card"],
    ["--accent-foreground","--accent"], ["--destructive-foreground","--destructive"], ["--destructive","--background"],
    ["--success-foreground","--success"], ["--success","--background"], ["--warning-foreground","--warning"],
    ["--warning-text","--background"], ["--warning-text","--card"], ["--info-foreground","--info"], ["--info","--background"],
    ["--sidebar-foreground","--sidebar"], ["--sidebar-primary-foreground","--sidebar-primary"], ["--sidebar-accent-foreground","--sidebar-accent"],
    ["--calendar-busy-foreground","--calendar-busy"], ["--calendar-package-foreground","--calendar-package"],
    // plus ["--foreground", `--swatch-${name}-soft`] for all 16 TrackerColor names
  ];
  const NON_TEXT_PAIRS: [fg: string, bg: string][] = [ // ≥ 3.0 in both themes
    ["--input","--background"], ["--ring","--background"], ["--sidebar-ring","--sidebar"],
    ["--chart-1","--card"], ["--chart-2","--card"], ["--chart-3","--card"], ["--chart-4","--card"], ["--chart-5","--card"],
    ["--calendar-busy","--card"], ["--calendar-package","--card"],
    // plus [`--swatch-${name}`, "--card"] for all 16 names
  ];
  ```
  Tests: (a) `tokenBlock(":root")` keys minus `THEME_INVARIANT` equal `tokenBlock(".dark")` keys; (b) every pair in both tables meets its threshold in both themes, with the failure message naming theme, pair and measured ratio; (c) `resolveToken(light, "--on-color-ink") === INK` and the same for paper; (d) every `--color-*` in `@theme inline` points at a token that exists in `:root`. Build the 16 names from a `TRACKER_COLORS` array literal in the test (Task 8 will make the test import it from `swatches.ts` instead).
- [x] **Step 2: Run** `npx vitest run src/__tests__/design-tokens.test.ts` → FAIL (grey tokens fail most pairs; new roles missing).
- [x] **Step 3: Implement** the tokens in `index.css` per design.md §3.1 and the list above.
- [x] **Step 4: Run** the test → PASS. Record any L nudges you made (token, old L, new L) in the commit body; they go into DESIGN.md at Task 32.
- [x] **Step 5: Commit** `feat(frontend): Kitchen Table colour tokens with contrast test`.

### Task 6: Typography and fonts

**Files:**
- Modify: `frontend/package.json`, `frontend/package-lock.json`, `frontend/src/index.css`
- Test: `frontend/src/__tests__/design-tokens.test.ts` (add a `typography` describe)

**Interfaces:**
- Produces: `font-heading` (Fraunces stack), `font-sans` (Figtree stack), text scale `text-xs … text-3xl`, `text-display`.

- [x] **Step 1: Write the failing tests:** `index.css` contains no `geist` (case-insensitive) import; contains `@fontsource-variable/fraunces` and `@fontsource-variable/figtree` imports; `--font-heading` starts with `'Fraunces Variable', 'Fraunces Fallback'` and `--font-sans` starts with `'Figtree Variable', 'Figtree Fallback'`; `@font-face` blocks exist for `Fraunces Fallback` (src `local("Georgia")`) and `Figtree Fallback` (src `local("Arial")`), each with `size-adjust`, `ascent-override`, `descent-override`; `@theme` sets `--text-xl: 1.375rem`, `--text-2xl: 1.75rem`, `--text-3xl: 2.25rem`, `--text-display: 3rem`.
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Swap dependencies:** `npm install @fontsource-variable/fraunces @fontsource-variable/figtree && npm uninstall @fontsource-variable/geist`. If a package ships a latin-only CSS entry (look under `node_modules/@fontsource-variable/<name>/` for `*latin*.css`), import that; otherwise import the package root and rely on `unicode-range`. For Fraunces import the `wght`-only entry (no `opsz` axis).
- [x] **Step 4: Check `tnum`.** With any local tool (e.g. `python3 -c "from fontTools.ttLib import TTFont; …"` if fontTools exists; it is not a project dependency), list Figtree's GSUB features. If `tnum` is present, continue. If absent, keep `@fontsource-variable/geist`, add `--font-numeric: 'Geist Variable', …` with a `font-numeric` utility, and note it for DESIGN.md. Write the answer in the commit body.
- [x] **Step 5: Implement** in `index.css`: font imports; fallback `@font-face` rules with metric overrides derived from each woff2's `hhea`/`OS/2` and `unitsPerEm` (put the numbers and how you derived them in a CSS comment); `--font-heading`, `--font-sans` stacks; `font-display: swap` (fontsource default; confirm); text scale in `@theme` with `--text-*--line-height` 1.5 for xs–lg and 1.2 for xl–display; base layer `h1,h2,h3 { letter-spacing: -0.01em }` is **not** added globally: apply `tracking-[-0.01em]` where `font-heading` is used (Tasks 11–12).
- [x] **Step 6: Run** the test → PASS. `npm run build`, then sum the woff2 files a latin page loads; must be ≤ 150 KB. If Fraunces variable pushes it over, switch Fraunces to static `@fontsource/fraunces` weight 600 and re-measure. Record the number in `design-audit.md` `## Baselines` as "Stage 1 font payload".
- [x] **Step 7: Commit** `feat(frontend): Fraunces + Figtree self-hosted type with metric fallbacks`.

### Task 7: Base layer — focus, reduced motion, coarse pointer, shape, elevation, motion tokens

**Files:**
- Modify: `frontend/src/index.css`
- Test: `frontend/src/__tests__/design-tokens.test.ts` (add a `base layer` describe)

**Interfaces:**
- Produces: the `coarse:` variant; `--radius: 0.75rem`; `shadow-card`, `shadow-overlay` (via `--elevation-card`, `--elevation-overlay`; dark values `none`); `--duration-fast: 120ms`, `--duration-base: 200ms`, `--ease-out: cubic-bezier(.2,.7,.2,1)`; global focus ring.

- [x] **Step 1: Write the failing tests:** `index.css` contains `@custom-variant coarse (@media (pointer: coarse));`; does not contain `outline-ring/50`; contains a `:focus-visible` rule with `outline: 2px solid var(--ring)` and `outline-offset: 2px`; contains a `@media (prefers-reduced-motion: reduce)` block whose body includes `animation-duration: 0.01ms !important`, `transition-duration: 0.01ms !important` and `animation-iteration-count: 1 !important`; `:root --radius` is `0.75rem`; `--elevation-card` in `.dark` is `none`; `@theme inline` maps `--shadow-card: var(--elevation-card)` and `--shadow-overlay: var(--elevation-overlay)` (raw tokens must not share the `--shadow-*` name, or the mapping is circular).
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Implement.** Light `--elevation-card` is the exact value in design.md §3.4; `--elevation-overlay` is `0 4px 12px oklch(0.3 0.03 50 / 0.10), 0 12px 32px oklch(0.3 0.03 50 / 0.12)`. Reduced-motion block:
  ```css
  @media (prefers-reduced-motion: reduce) {
    *, *::before, *::after {
      animation-duration: 0.01ms !important;
      animation-iteration-count: 1 !important;
      transition-duration: 0.01ms !important;
      scroll-behavior: auto !important;
    }
  }
  ```
  Base `*` rule keeps `border-border` and drops `outline-ring/50`.
- [x] **Step 4: Run** → PASS.
- [x] **Step 5: Commit** `feat(frontend): focus ring, reduced motion, coarse variant, shape and elevation tokens`.

### Task 8: Swatch module

**Files:**
- Create: `frontend/src/lib/constants/swatches.ts`
- Test: `frontend/src/lib/__tests__/swatches.test.ts`
- Modify: `frontend/src/__tests__/design-tokens.test.ts` (import `TRACKER_COLORS` from the module instead of its local array)

**Interfaces:**
- Consumes: `TrackerColor` from `@/types/models/tracker`; swatch utilities from Task 5.
- Produces:
  ```ts
  export const TRACKER_COLORS: readonly TrackerColor[]          // the 16 names, in the union's order
  export type SwatchRole = "solid" | "soft" | "borderLeft"
  export const SWATCH_CLASSES: Record<SwatchRole, Record<TrackerColor, string>> // literal strings: "bg-swatch-red", "bg-swatch-red-soft", "border-l-swatch-red"
  export function swatchClass(color: string, role: SwatchRole): string           // unknown colour → neutral fallback: solid "bg-muted-foreground", soft "bg-muted", borderLeft "border-l-muted-foreground"
  ```
  Task 25 may add roles (e.g. `"text"`) if the tracker code needs them; each new role is a full 16-entry literal map, and the completeness test covers it automatically.

- [x] **Step 1: Write the failing tests:** `TRACKER_COLORS` has 16 unique entries; for every role and every colour, `SWATCH_CLASSES[role][c]` is a non-empty string containing `swatch-${c}`; `swatchClass("red","soft") === "bg-swatch-red-soft"`; **Review Focus 2:** `swatchClass("chartreuse","solid") === "bg-muted-foreground"` and `swatchClass("", "soft") === "bg-muted"`, and neither contains `undefined`.
- [x] **Step 2: Run** → FAIL. **Step 3: Implement.** **Step 4: Run** both this test and `design-tokens.test.ts` → PASS.
- [x] **Step 5: Commit** `feat(frontend): tracker swatch token mapping`.

### Task 9: No-colour-literal guard

**Files:**
- Create: `frontend/src/__tests__/no-color-literals.test.ts`

**Interfaces:**
- Produces, inside the test file (later stages edit these two constants only):
  ```ts
  const PENDING: string[]  // path prefixes relative to frontend/src; files under them are skipped
  const ALLOWED: { file: string; match: string; reason: string }[] // exact file + exact matched text; reason mirrors DESIGN.md
  ```

Rules: scan `.ts`/`.tsx` under `src/components/features/**` and `src/pages/**`, skipping any path containing `/__tests__/`. Flag hex `#[0-9a-fA-F]{3,8}\b` (only inside string or template literals, so `#` in comments is not flagged), `rgba?\(`, `hsla?\(`, `oklch\(`, and palette classes: optional variant prefixes (`dark:`, `hover:`, …), then `bg|text|border(-[trblxy])?|ring|ring-offset|fill|stroke|from|via|to|outline|decoration|divide|shadow|accent|caret|placeholder`, then `-` and one of `red orange amber yellow lime green emerald teal cyan sky blue indigo violet purple fuchsia pink rose slate gray zinc neutral stone` with `-\d{2,3}`, or `white|black`, optionally `/\d+`.

Initial `PENDING` (one entry per line, exactly this set):
`components/features/<dir>/` for each of calendar, dashboard-widgets, dashboards, households, ingredients, meals, navigation, packages, recipes, reminders, tasks, tracker, trackers, weather, workout, workouts;
`pages/dashboard-designer/`; and `pages/<File>.tsx` for every top-level page file (CalendarPage, DashboardDesigner, DashboardManagementPage, DashboardRedirect, DashboardRenderer, DashboardShell, DataRetentionPage, HouseholdMembersPage, HouseholdsPage, IngredientDetailPage, IngredientsPage, LoginPage, MealsPage, OnboardingPage, PackagesPage, RecipeDetailPage, RecipeFormPage, RecipesPage, RemindersPage, SettingsPage, ShoppingListDetailPage, ShoppingListsPage, TasksPage, TrackerPage, WeatherPage, WishListPage, WorkoutExercisesPage, WorkoutReviewPage, WorkoutTaxonomyPage, WorkoutTodayPage, WorkoutWeekPage).
Initial `ALLOWED`: empty.

- [x] **Step 1: Write the test** with four cases: (a) no unallowed literal in any non-pending file (failure message lists `file:line: match`); (b) every `PENDING` entry exists on disk; (c) every `ALLOWED` entry still matches its file (no stale entries); (d) a self-test that the matcher flags `"bg-red-500"`, `"dark:text-white"`, `"bg-black/30"`, `"#9ca3af"` and does not flag `"bg-primary"`, `"text-muted-foreground"`, `"bg-swatch-red"`, `"border-l-swatch-blue"`.
- [x] **Step 2: Run** `npx vitest run src/__tests__/no-color-literals.test.ts`. Expected: PASS, because everything with literals is pending. Then temporarily delete the `pages/IngredientsPage.tsx` entry and confirm it FAILS naming `text-yellow-600`; restore it.
- [x] **Step 3: Commit** `test(frontend): no-colour-literal guard with pending list`.

### Task 10: Control primitives

**Files:**
- Modify: `frontend/src/components/ui/{button,badge,input,select,textarea,radio-group,label,form}.tsx`
- Test: `frontend/src/components/ui/__tests__/button.test.tsx`, `badge.test.tsx` (create), plus any existing ui test whose class assertions change

**Interfaces:**
- Produces: `Badge` variants `default | secondary | destructive | success | warning | info | outline | ghost | link` and prop `icon?: LucideIcon` (rendered first, `aria-hidden`, `data-icon="inline-start"`). Button sizes unchanged in name.

- [x] **Step 1: Write the failing tests.**
  `button.test.tsx`: default button's class contains `focus-visible:outline-ring` and `coarse:min-h-11`, and contains neither `ring-ring/50` nor `outline-none`; `size="icon"` contains `coarse:size-11`; `size="icon-xs"` contains `size-6`.
  `badge.test.tsx`: `<Badge variant="success" icon={Check}>Done</Badge>` renders an `svg` with `aria-hidden="true"` before the text and has class `bg-success`; `variant="warning"` has `bg-warning` and `text-warning-foreground`; `variant="destructive"` has `bg-destructive` and `text-destructive-foreground`.
  A source-scan case in `button.test.tsx`: no file in `src/components/ui/` contains `ring-ring/50`.
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Implement.** Replace `focus-visible:ring-3 focus-visible:ring-ring/50` (and `outline-none`) with `focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring` in every control. Controls `rounded-md`; inputs/selects/textareas use `border-input`, placeholder `text-muted-foreground`, disabled keeps text readable (`disabled:opacity-60` on fill, not on text alone). Button `default` size `h-9`; add `coarse:min-h-11` to text sizes and `coarse:size-11` to icon sizes. Status badges are solid fills with their `-foreground` text, `rounded-full`. Transitions use `duration-(--duration-fast) ease-out`.
- [x] **Step 4: Run** `npx vitest run src/components/ui` → PASS; `npx tsc -b`; `npx eslint src/components/ui`.
- [x] **Step 5: Commit** `feat(frontend): restyle control primitives for Kitchen Table`.

### Task 11: Surface primitives

**Files:**
- Modify: `frontend/src/components/ui/{card,dialog,popover,tooltip,sidebar,skeleton,calendar,user-avatar}.tsx`
- Test: `frontend/src/components/ui/__tests__/{sidebar,user-avatar}.test.tsx` (existing), `card.test.tsx` and `dialog.test.tsx` (create)

**Interfaces:**
- Consumes: `readableForeground` (Task 4), `bg-overlay`, `shadow-card`, `shadow-overlay`, `coarse:`.

- [x] **Step 1: Write the failing tests:** `CardTitle` has `font-heading` and the card root has `rounded-xl` and `shadow-card`; the dialog backdrop has `bg-overlay` and not `bg-black/30`; the dialog popup has `shadow-overlay`; `SidebarMenuButton` (or the existing nav item primitive) has `coarse:min-h-11`; `UserAvatar` with a light user colour renders `color: var(--on-color-ink)` and with a dark colour `var(--on-color-paper)` (no `text-white`).
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Implement.** Cards: `bg-card rounded-xl shadow-card`, dark gets a hairline `border` (shadow is `none` there). Titles: `font-heading text-lg tracking-[-0.01em]` (18 px floor). Popover/tooltip/dialog: `shadow-overlay`, `bg-popover`, open/close animation durations from motion tokens. Sidebar: `bg-sidebar`; active item = `text-sidebar-primary bg-sidebar-accent rounded-md` (clay text on sage wash, design.md §3.6). Skeleton `bg-muted`. Calendar (day picker): selected = `bg-primary text-primary-foreground`, today = ring, focus per Task 10.
- [x] **Step 4: Run** `npx vitest run src/components/ui` → PASS; tsc; eslint.
- [x] **Step 5: Commit** `feat(frontend): restyle surface primitives`.

### Task 12: Common components, including PageHeader and EmptyState

**Files:**
- Create: `frontend/src/components/common/page-header.tsx`, `frontend/src/components/common/empty-state.tsx`
- Modify: `frontend/src/components/common/{error-card,error-page,dashboard-skeleton,list-filter-bar,card-action-menu,data-table,owner-select,pull-to-refresh,weather-icon}.tsx`
- Test: `frontend/src/components/common/__tests__/{page-header,empty-state}.test.tsx` (create); existing common tests updated as needed

Skip creating either new component if Task 3's `## Method` dropped it; add any extra primitive Task 3 added as a Stage 1 row, with its own render test.

**Interfaces:**
- Produces:
  ```ts
  export interface PageHeaderProps { title: React.ReactNode; description?: React.ReactNode; actions?: React.ReactNode; className?: string }
  export function PageHeader(props: PageHeaderProps): React.JSX.Element   // <header>, title in <h1>
  export interface EmptyStateProps { icon?: LucideIcon; title: string; description?: React.ReactNode; action?: React.ReactNode; className?: string }
  export function EmptyState(props: EmptyStateProps): React.JSX.Element   // title in <h2>, icon aria-hidden
  ```

- [x] **Step 1: Write the failing tests.** PageHeader: `getByRole("heading", { level: 1, name: "Recipes" })` has `font-heading`; **Review Focus 4:** the h1 has `min-w-0` and `break-words`, and the actions container has `flex-wrap`; description renders in `text-muted-foreground`; actions render. EmptyState: `getByRole("heading", { level: 2, name: "No recipes yet" })` has `font-heading` and `text-lg`; icon `svg` is `aria-hidden`; action renders.
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Implement** both; PageHeader title `font-heading text-2xl tracking-[-0.01em]`. Restyle the existing common components onto tokens (no palette classes; weather icons map to chart tokens), and resolve every Stage 1 `design-audit.md` row whose Location is in `components/common/`.
- [x] **Step 4: Run** `npx vitest run src/components/common` → PASS; tsc; eslint.
- [x] **Step 5: Commit** `feat(frontend): PageHeader, EmptyState, restyled common components`.

### Task 13: Brand mark and favicon

**Files:**
- Modify: `frontend/src/components/common/brand-mark.tsx`, `frontend/public/favicon.svg`
- Test: `frontend/src/components/common/__tests__/brand-mark.test.tsx` (create)

- [x] **Step 1: Write the failing test:** read `public/favicon.svg` with `fs`, render `BrandMark`, and assert the set of `fill`/`stop-color` values in each is identical (the two files stay in sync, design.md §3.4) and that neither contains `#6366f1`.
- [x] **Step 2: Run** → FAIL.
- [x] **Step 3: Implement** the "Window light" geometry from `docs/tasks/task-057-impeccable-redesign/comps/brand-mark-options.html` (element `#mB`, 32×32) in both files. Keep `role="img"` and the accessible name. Literal fills are the documented image-asset exception (`components/common` is outside the literal guard's scope).
- [x] **Step 4: Run** → PASS. **Step 5: Commit** `feat(frontend): Window light brand mark and favicon`.

### Task 14: Resolve remaining Stage 1 findings

**Files:** as named by the `design-audit.md` rows with `Stage = 1` not yet fixed by Tasks 5–13.

- [x] **Step 1:** List open Stage 1 rows. For each, write or update a test where the fix is testable in jsdom (role/name, class, attribute), then fix. Visual-only fixes need no test but must name the row ID in the commit message.
- [x] **Step 2:** Run the Impeccable detector on `frontend/src/index.css` and `components/ui|common`. If it raises `cream-palette`, add an ignore entry with reason `user confirmed warm paper background in task-057 design review` (design.md "Decisions confirmed"). Fix anything else it raises or add it as a row.
- [x] **Step 3:** `npx vitest run && npx tsc -b && npx eslint .` from `frontend/` → all pass.
- [x] **Step 4: Commit** each fix as you go: `fix(frontend): <summary> (DA-NNN)`.

Then run the **Stage close-out procedure** (PR 1 title: `task-057 stage 0+1: audit and foundation`).

---

## Stage 2 — Shell

### Task 15: App shell and navigation

**Files:**
- Modify: `frontend/src/components/features/navigation/{app-shell,nav-group,dashboards-nav-group,mobile-header,mobile-drawer,user-menu}.tsx`
- Modify: `frontend/src/__tests__/no-color-literals.test.ts` (remove `components/features/navigation/` from `PENDING`)
- Test: `frontend/src/components/features/navigation/__tests__/*` (update)

- [x] **Step 1:** Cut the stage branch (see "Branch and PR mechanics").
- [x] **Step 2: Remove the PENDING entry; run** `npx vitest run src/__tests__/no-color-literals.test.ts` → FAIL listing navigation literals.
- [x] **Step 3: Add failing tests:** mobile header menu button has an accessible name and `coarse:size-11`; the active nav item has `aria-current="page"`; the brand mark renders in the sidebar header and mobile header.
- [x] **Step 4: Implement** per design.md §3.6 sketch (sidebar deeper paper, active = clay text on sage wash, Fraunces only for the "Home Hub" wordmark at ≥ 18 px) and fix every Stage 2 row located in navigation.
- [x] **Step 5: Run** guard + navigation tests → PASS; tsc; eslint. **Step 6: Commit** `feat(frontend): restyle app shell and navigation`.

### Task 16: Login, onboarding, error pages, toasts

**Files:**
- Modify: `frontend/src/pages/{LoginPage,OnboardingPage}.tsx`, `frontend/src/components/common/error-page.tsx` (if Task 12 left shell-specific work), `frontend/src/App.tsx` (`<Toaster>` props only)
- Modify: `no-color-literals.test.ts` (remove `pages/LoginPage.tsx`, `pages/OnboardingPage.tsx`)
- Test: `frontend/src/pages/__tests__/*` for these pages (update/create)

- [x] **Step 1: Remove the PENDING entries; run the guard** → FAIL (or PASS if they had none; still remove them).
- [x] **Step 2: Add failing tests:** each page has exactly one `h1` with `font-heading`; every form control on Login/Onboarding is found by `getByLabelText`.
- [x] **Step 3: Implement.** Toaster: drop `richColors` and pass `toastOptions={{ classNames: { toast: "bg-popover text-popover-foreground border-border shadow-overlay", success: "…", error: "…", warning: "…", info: "…" } }}` using the semantic token pairs; keep `closeButton`. This changes toast colours only; list it in the PR body as a visual change, not a behaviour change. Fix every remaining Stage 2 row.
- [x] **Step 4: Run** `npx vitest run && npx tsc -b && npx eslint .` → PASS. **Step 5: Commit** `feat(frontend): restyle login, onboarding, errors, toasts`.

Then run the **Stage close-out procedure** (title: `task-057 stage 2: shell`).

---

## Stage 3 — Dashboards & kiosk

### Task 17: Widget internals with container queries

**Files:**
- Modify: every `.tsx` in `frontend/src/components/features/dashboard-widgets/` (adapters, summaries, `lossy-config-badge`, `unknown-widget-placeholder`)
- Modify: `no-color-literals.test.ts` (remove `components/features/dashboard-widgets/`)
- Test: `frontend/src/components/features/dashboard-widgets/__tests__/widget-container.test.tsx` (create)

- [x] **Step 1:** Cut the stage branch.
- [x] **Step 2: Write the failing test (Review Focus 3):** for each adapter exported from the widget registry (iterate `widget-registry.ts` so new widgets are covered automatically), render it with its data hooks mocked the way the existing `dashboard-widgets/__tests__` do (fixtures from `lib/dashboard/fixtures` may be imported read-only), and assert the root element has class `@container` and the widget title (if any) has `font-heading`. Remove the PENDING entry and run guard + this test → FAIL.
- [x] **Step 3: Implement.** Widget root `@container`; title `font-heading text-lg @xl:text-2xl`; headline metric (Today count, temperature) `tabular-nums text-3xl @md:text-display`; list rows `text-sm @xl:text-xl`; icons `size-4 @xl:size-6`. Breakpoints per widget are the implementer's call (design.md §8). Status rows carry icon + text. Leave calendar-widget `userColor` dots as they are (data, not a literal); Task 22 moves them to `eventChipColors`.
- [x] **Step 4: Run** the new test, the guard, and `npx vitest run src/lib/dashboard` → PASS. `git diff --stat origin/main -- frontend/src/lib/dashboard` must print nothing.
- [x] **Step 5: Commit** `feat(frontend): container-query widget internals for kiosk legibility`.

### Task 18: Dashboard pages and designer

**Files:**
- Modify: `frontend/src/pages/{DashboardManagementPage,DashboardShell,DashboardRenderer,DashboardDesigner,DashboardRedirect}.tsx`, `frontend/src/pages/dashboard-designer/{config-panel,palette-drawer,widget-chrome,designer-grid}.tsx`, `frontend/src/components/features/dashboards/*`
- Modify: `no-color-literals.test.ts` (remove those five page files, `pages/dashboard-designer/`, `components/features/dashboards/`)
- Test: existing `pages/__tests__` and `pages/dashboard-designer/__tests__` (update)

- [x] **Step 1: Remove PENDING entries; run guard** → FAIL naming `bg-black/30` in `config-panel.tsx` and `palette-drawer.tsx`.
- [x] **Step 2: Implement:** backdrops `bg-overlay`; management page uses `PageHeader` and `EmptyState`; greeting header per design.md §3.6 (Fraunces `text-2xl`, muted date). Grid props, layout JSON handling and `state.ts` logic unchanged.
- [x] **Step 3: Run** `npx vitest run src/pages src/lib/dashboard src/__tests__` → PASS; tsc; eslint. **Step 4: Commit** `feat(frontend): restyle dashboard pages and designer`.

### Task 19: Resolve remaining Stage 3 findings

Same steps as Task 14, for `Stage = 3` rows, then confirm again that `git diff --stat origin/main -- frontend/src/lib/dashboard` is empty. Then run the **Stage close-out procedure** (title: `task-057 stage 3: dashboards and kiosk`; the rendered check includes kiosk in both themes).

---

## Stage 4 — Food

### Task 20: Recipes, meals, ingredients, shopping, wish list

**Files:**
- Modify: `frontend/src/components/features/{recipes,meals,ingredients}/*`, `frontend/src/pages/{RecipesPage,RecipeDetailPage,RecipeFormPage,MealsPage,IngredientsPage,IngredientDetailPage,ShoppingListsPage,ShoppingListDetailPage,WishListPage}.tsx`
- Modify: `no-color-literals.test.ts` (remove those three feature dirs and nine page files)
- Test: existing tests in those areas (update); `frontend/src/components/features/recipes/__tests__/recipe-status.test.tsx` (create)

Known literals and their mapping (decided here so every implementer maps the same way):

| Current | Becomes |
|---|---|
| `text-green-500` resolved/check icons (`recipe-card`, `RecipeDetailPage`, `RecipeFormPage`, `recipe-ingredients`) | `text-success` + existing `title`/sr-only label kept |
| `text-yellow-500/600` unresolved/alert icons and text | `text-warning-text` |
| `IngredientsPage` "uncategorized" `text-yellow-600 border-yellow-300` outline badge | `<Badge variant="warning" icon={AlertTriangle}>uncategorized</Badge>` |
| `step-segment` ingredient `text-orange-600 dark:text-orange-400` | `text-primary` |
| `step-segment` cookware `text-blue-600 dark:text-blue-400` | `text-info` |
| `step-segment` timer `text-green-600 dark:text-green-400` | `text-success` |

- [x] **Step 1:** Cut the stage branch. **Step 2:** Remove the PENDING entries; run the guard → FAIL listing the literals above.
- [x] **Step 3: Write the failing test** `recipe-status.test.tsx`: the resolved icon in `RecipeCard` has `text-success` and an accessible name or adjacent text; the unresolved icon has `text-warning-text`.
- [x] **Step 4: Implement** the mapping; pages adopt `PageHeader`/`EmptyState` where they hand-roll them; fix every Stage 4 row in these files. Shopping mode rows get `coarse:min-h-11`.
- [x] **Step 5: Run** guard + area tests → PASS; tsc; eslint. **Step 6: Commit** in two commits if large (recipes+meals; ingredients+shopping+wish list), each `feat(frontend): restyle <area>`.

### Task 21: Resolve remaining Stage 4 findings

Same steps as Task 14, for `Stage = 4`. Then the **Stage close-out procedure** (title: `task-057 stage 4: food`).

---

## Stage 5 — Time

### Task 22: Calendar event colours

**Files:**
- Create: `frontend/src/lib/calendar/event-colors.ts`
- Test: `frontend/src/lib/calendar/__tests__/event-colors.test.ts`
- Modify: `frontend/src/components/features/calendar/{month-day-cell,event-block,all-day-event-row,month-event-chip,event-popover,user-legend,calendar-widget,connection-status}.tsx`, `frontend/src/components/features/packages/package-calendar-overlay.tsx`, `frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx`
- Modify: `frontend/src/components/features/packages/__tests__/package-calendar-overlay.test.ts`

**Interfaces:**
- Consumes: `readableForeground` (Task 4); `--calendar-*`, `--on-color-*` tokens (Task 5).
- Produces:
  ```ts
  export const PACKAGE_EVENT_COLOR = "var(--calendar-package)"
  export interface ChipColors { backgroundColor: string; color: string }
  export function eventChipColors(attrs: { title: string; isOwner: boolean; userColor: string }): ChipColors
  // busy (title "Busy" && !isOwner) → { var(--calendar-busy), var(--calendar-busy-foreground) }
  // userColor === PACKAGE_EVENT_COLOR → { var(--calendar-package), var(--calendar-package-foreground) }
  // otherwise → { userColor, readableForeground(userColor) === "ink" ? "var(--on-color-ink)" : "var(--on-color-paper)" }
  ```
  Check the exact attribute names against `src/types/models/calendar.ts` and the current busy condition in `month-day-cell.tsx:24` / `event-block.tsx:31`; if the busy rule differs between files today, keep `month-day-cell`'s and note the unification in the PR body as a behaviour change.

- [ ] **Step 1:** Cut the stage branch.
- [ ] **Step 2: Write the failing tests (Review Focus 1):** busy non-owner → busy vars; busy owner → user colour; package sentinel → package vars; `#ffff00` → `color: "var(--on-color-ink)"`; `#000000` → `"var(--on-color-paper)"`; `""` and `"teal"` → returns `backgroundColor` unchanged and `color: "var(--on-color-ink)"`, no throw. Update `package-calendar-overlay.test.ts` to expect `userColor: PACKAGE_EVENT_COLOR` instead of `#0d9488`.
- [ ] **Step 3: Run** → FAIL. **Step 4: Implement** the helper; replace every `#9ca3af` ternary and raw `backgroundColor: userColor` with `style={eventChipColors(attrs)}` (dots with no text may use only `backgroundColor`); overlay sets `userColor: PACKAGE_EVENT_COLOR`.
- [ ] **Step 5: Run** `npx vitest run src/lib/calendar src/components/features/calendar src/components/features/packages` → PASS. **Step 6: Commit** `feat(frontend): token- and contrast-aware calendar event colours`.

### Task 23: Calendar, tasks, reminders

**Files:**
- Modify: remaining `frontend/src/components/features/{calendar,tasks,reminders}/*`, `frontend/src/pages/{CalendarPage,TasksPage,RemindersPage}.tsx`
- Modify: `no-color-literals.test.ts` (remove those three feature dirs and three page files)

- [ ] **Step 1: Remove PENDING entries; run guard** → FAIL listing remaining calendar literals (e.g. `calendar-grid.tsx`, `connection-status.tsx`).
- [ ] **Step 2: Implement:** map palette classes to semantic tokens (status) or `--chart-*` (categorical); `PageHeader`/`EmptyState` where hand-rolled; task/reminder rows `coarse:min-h-11`, times `tabular-nums`; fix Stage 5 rows in these files. Update any test asserting old classes.
- [ ] **Step 3: Run** guard + area tests → PASS; tsc; eslint. **Step 4: Commit** `feat(frontend): restyle calendar, tasks, reminders`.

### Task 24: Resolve remaining Stage 5 findings

Same steps as Task 14, for `Stage = 5`. Then the **Stage close-out procedure** (title: `task-057 stage 5: time`).

---

## Stage 6 — Body & home

### Task 25: Tracker swatch migration

**Files:**
- Modify: `frontend/src/components/features/tracker/{calendar-grid,create-tracker-dialog,edit-tracker-dialog,today-view,tracker-setup,month-report}.tsx`, `frontend/src/components/features/trackers/*`, `frontend/src/pages/TrackerPage.tsx`
- Modify: `frontend/src/lib/constants/swatches.ts` only to add roles (see Task 8 Interfaces)
- Modify: `frontend/src/components/features/tracker/__tests__/calendar-grid.test.tsx` (line ~142 expects `bg-blue-500` → expect `bg-swatch-blue`)
- Modify: `no-color-literals.test.ts` (remove `components/features/tracker/`, `components/features/trackers/`, `pages/TrackerPage.tsx`)

- [ ] **Step 1:** Cut the stage branch. **Step 2:** Remove the PENDING entries; run the guard → FAIL (≈140 matches in `tracker/`).
- [ ] **Step 3:** Update `calendar-grid.test.tsx` to expect `bg-swatch-blue`; run → FAIL.
- [ ] **Step 4: Implement.** Delete the four duplicated `Record<TrackerColor, string>` maps (create/edit dialogs, today-view, tracker-setup) and use `SWATCH_CLASSES` / `swatchClass`. Sentiment colours in `month-report` (positive/neutral/negative bars) map to `bg-success` / `bg-muted-foreground` / `bg-destructive` with a text legend. Colour pickers keep the colour name as the accessible label.
- [ ] **Step 5: Run** guard + `npx vitest run src/components/features/tracker src/lib` → PASS. **Step 6: Commit** `feat(frontend): migrate tracker colours to swatch tokens`.

### Task 26: Workouts

**Files:**
- Modify: `frontend/src/components/features/{workout,workouts}/*`, `frontend/src/pages/{WorkoutTodayPage,WorkoutWeekPage,WorkoutExercisesPage,WorkoutTaxonomyPage,WorkoutReviewPage}.tsx`
- Modify: `no-color-literals.test.ts` (remove those two dirs and five page files)

Mapping: `bg-emerald-500/10 text-emerald-600` → `bg-success/10 text-success`; `bg-amber-500/10 text-amber-600` → `bg-warning/15 text-warning-text`; `WorkoutReviewPage` "Target met" `text-emerald-600` → `text-success` (keep its `aria-label`).

- [ ] **Step 1: Remove PENDING entries; run guard** → FAIL. **Step 2: Implement** the mapping, `PageHeader`, `tabular-nums` on sets/reps/weights, Stage 6 rows in these files. **Step 3: Run** guard + area tests → PASS; tsc; eslint. **Step 4: Commit** `feat(frontend): restyle workouts`.

### Task 27: Packages and weather

**Files:**
- Modify: `frontend/src/components/features/{packages,weather}/*`, `frontend/src/pages/{PackagesPage,WeatherPage}.tsx`
- Modify: `no-color-literals.test.ts` (remove those two dirs and two page files)

- [ ] **Step 1: Remove PENDING entries; run guard** → FAIL (packages ≈7 matches). **Step 2: Implement:** carrier icons map to `text-chart-1..5` (categorical, never carrying text); status (delivered/delayed/exception) uses `Badge` status variants with icons; weather temps `tabular-nums`. **Step 3: Run** guard + area tests → PASS. **Step 4: Commit** `feat(frontend): restyle packages and weather`.

### Task 28: Resolve remaining Stage 6 findings

Same steps as Task 14, for `Stage = 6`. Then the **Stage close-out procedure** (title: `task-057 stage 6: body and home`).

---

## Stage 7 — Household & finish

### Task 29: Households, members, settings, data retention

**Files:**
- Modify: `frontend/src/components/features/households/*`, `frontend/src/pages/{HouseholdsPage,HouseholdMembersPage,SettingsPage,DataRetentionPage}.tsx`, `frontend/src/pages/DashboardRedirect.tsx` if still pending
- Modify: `no-color-literals.test.ts`: remove the last entries, then add case (e) `expect(PENDING).toEqual([])`

- [ ] **Step 1:** Cut the stage branch. **Step 2:** Remove the entries, add case (e); run the guard → FAIL if any literal remains, PASS otherwise.
- [ ] **Step 3: Implement** token mapping, `PageHeader`/`EmptyState`, Stage 7 rows in these files. Settings' theme toggle keeps calling the unchanged theme provider.
- [ ] **Step 4: Run** `npx vitest run && npx tsc -b && npx eslint .` → PASS. **Step 5: Commit** `feat(frontend): restyle households and settings; literal guard complete`.

### Task 30: Resolve remaining Stage 7 findings and sweep

- [ ] **Step 1:** Same as Task 14 for `Stage = 7`.
- [ ] **Step 2: Sweep** for icon-only buttons without names: `grep -rn 'size="icon' frontend/src --include=*.tsx`; for each hit, confirm `aria-label`, `aria-labelledby`, or sr-only text. Fix misses with an RTL `getByRole("button", { name })` assertion in that component's test.
- [ ] **Step 3:** Every `design-audit.md` row now has a fix commit (the `Resolved` SHA is filled at close-out). List any row still open; there must be none.
- [ ] **Step 4: Commit** fixes as `fix(frontend): <summary> (DA-NNN)`.

### Task 31: Finish review

- [ ] **Step 1:** Dispatch `impeccable:impeccable-finish-reviewer` with: `design.md` (direction contract), `comps/dashboard.html` (approved comp), and the running app if a browser is available (otherwise the source tree; say which in the brief).
- [ ] **Step 2:** Add each material fix it returns to `design-audit.md` as a new row (`Stage 7`, ID continuing the sequence, Description prefixed `finish-review:`).
- [ ] **Step 3:** Fix each, with a test where jsdom can observe it. Re-dispatch the finish reviewer once; repeat until it returns no material fixes (FR-17, PRD acceptance).
- [ ] **Step 4: Commit** each fix `fix(frontend): <summary> (DA-NNN)`.

### Task 32: DESIGN.md from the shipped build

**Files:**
- Create: `DESIGN.md` and `.impeccable/design.json` at the Impeccable root recorded in `design-audit.md` (`frontend/` per design.md §5.1 unless Task 1 decided otherwise)

- [ ] **Step 1:** Dispatch `impeccable:impeccable-documenter` (scan mode) against the final build, giving it `PRODUCT.md`, `design.md`, and the paths of `index.css`, `components/ui`, `components/common`.
- [ ] **Step 2: Check the output contains**, and add if missing: the token tables as shipped (including Task 5's L nudges and the `--warning-text` role), the Fraunces ≥ 18 px rule, the `tnum` outcome from Task 6, the swatch and calendar-colour rules, the `coarse:` target rule, the focus and reduced-motion rules, and an "Exceptions" section listing every `ALLOWED` entry from `no-color-literals.test.ts` plus the brand-mark/favicon image asset and calendar user colours as data.
- [ ] **Step 3:** Confirm every `ALLOWED` entry's `reason` appears in DESIGN.md's Exceptions section (manual cross-check; list the pairs in the commit body).
- [ ] **Step 4: Commit** `docs(frontend): DESIGN.md and design.json from the shipped build`.

### Task 33: Final budgets

- [ ] **Step 1:** From `frontend/`, run the Task 2 measurement commands on the final branch.
- [ ] **Step 2:** JS gzip must be ≤ 1.10 × the Task 2 baseline; latin default font payload ≤ 150 KB. If either fails, STOP and report to the controller with the numbers; do not silently trim features.
- [ ] **Step 3:** Record both numbers next to the baseline in `design-audit.md` `## Baselines` and commit `docs(task-057): final bundle and font budgets`.

Then run the **Stage close-out procedure** (title: `task-057 stage 7: household and finish`; the PR body carries the budget numbers).
