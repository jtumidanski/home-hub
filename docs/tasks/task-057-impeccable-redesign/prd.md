# Impeccable Onboarding, Design Audit & Redesign — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-10-19
---

## 1. Overview

The Home Hub frontend (React 19 + TypeScript, Vite, Tailwind 4, shadcn on
`@base-ui/react`) still ships the stock shadcn neutral theme. Every colour
token in `frontend/src/index.css` is a zero-chroma grey, the type is the
default Geist, and there is no written design system. ~30 pages across
meals/recipes, shopping, calendar, tasks/reminders, habits, workouts,
packages, weather, dashboards/kiosk, households/settings, and
login/onboarding have been built feature-by-feature with no shared visual
direction, so hierarchy, spacing, states, and accessibility vary page to page.

This task uses the Impeccable skill (`/impeccable:impeccable`) to (1) onboard
the product — interview product context and capture the current design system
into `frontend/DESIGN.md` and its sidecar — (2) establish a new, approved
visual direction seeded toward a **warm & domestic** feel, (3) run a
whole-app design audit across desktop, mobile, kiosk, and both themes, and
(4) remediate **every** finding while rolling the new direction out across
the app.

The work is one task delivered as **staged PRs** (foundation → shell →
feature areas), each passing its own code review and flagless
`tools/verify.sh` run. It is frontend-only: no backend, API, or data-model
changes.

## 2. Goals

Primary goals:
- Produce `frontend/DESIGN.md` (+ Impeccable sidecar/tokens) that documents
  the product context and the design system as actually shipped at the end
  of the task.
- Establish a new visual direction (colour, type, shape, density, motion)
  with a warm & domestic character, approved by the user in `design.md`
  before any rollout is planned.
- Produce a whole-app design audit report with severity-ranked findings.
- Fix every audit finding (Critical, High, Medium, Low).
- Meet WCAG 2.2 AA across all audited surfaces in both light and dark themes.
- Apply the new direction consistently across desktop, mobile, kiosk
  dashboard, light and dark modes.

Non-goals:
- Backend, API, JSON:API contract, or data-model changes.
- New features, new routes, or changes to information architecture beyond
  what an audit finding requires (e.g. a nav label or grouping fix is fine;
  a new page is not).
- Replacing shadcn / `@base-ui/react` or Tailwind 4.
- Changing the persisted dashboard layout format or migrating saved layouts.
- Native mobile apps.

## 3. User Stories

- As a household member, I want the app to feel warm and cohesive rather
  than generic, so that it feels like a home tool I enjoy opening.
- As a household member on my phone, I want touch targets, spacing, and
  mobile navigation to be comfortable, so that I can use it one-handed (e.g.
  in shopping mode).
- As a household viewing the wall kiosk, I want dashboard widgets to be
  legible at a distance in both themes, so that the at-a-glance view works.
- As a keyboard or screen-reader user, I want visible focus, correct labels,
  and sufficient contrast, so that I can use every page.
- As a user who prefers dark mode, I want dark mode to be a designed theme,
  not an inversion, so that it is as readable and pleasant as light mode.
- As a developer, I want a written `DESIGN.md` and tokens, so that new
  features follow the system without guessing.

## 4. Functional Requirements

### 4.1 Onboarding (Impeccable)
- FR-1: Run the Impeccable onboarding/capture flow against `frontend/`.
  Product context answers come from the user during the design phase.
- FR-2: Output `frontend/DESIGN.md` plus its sidecar/token file in the
  location Impeccable expects, configured to point at `frontend/`.
- FR-3: `DESIGN.md` is finalized at the end of the task from the *shipped*
  artifact (via `impeccable:impeccable-documenter`), not from intentions.

### 4.2 Visual direction
- FR-4: During `/design-task`, Impeccable proposes a direction (palette in
  OKLCH for light + dark, type family/scale, radius, elevation, density,
  motion, iconography usage) seeded toward "warm & domestic", with a comp or
  mock of at least one representative screen (dashboard) on desktop and
  mobile.
- FR-5: The user approves the direction in `design.md` before
  `/plan-task` runs. Rollout is not planned against an unapproved direction.
- FR-6: The direction is implemented through tokens first: CSS custom
  properties in `frontend/src/index.css` (`:root`, `.dark`, `@theme inline`)
  and `components/ui/*` variants. Pages consume tokens; hard-coded colours,
  radii, and font sizes in feature code are replaced with tokens/utilities.
- FR-7: Font family may change; any new font is self-hosted via an npm
  `@fontsource*` package (no external CDN requests). New runtime
  dependencies are permitted when the design justifies them; each is
  named and justified in `design.md`.
- FR-8: The existing light/dark mechanism (`theme-provider.tsx`, class on
  `<html>`, preference persisted via account-service) is retained; only
  token values and styling change.

### 4.3 Audit
- FR-9: Run the Impeccable audit across every route in `frontend/src/App.tsx`
  plus the shared shell (sidebar, mobile header/drawer), `components/ui/*`,
  `components/common/*`, dialogs/popovers/forms, toasts, and empty /
  loading / error states.
- FR-10: Each page is audited on desktop (≥1280px), mobile (≈390px), and —
  for dashboards — kiosk rendering, in both light and dark themes.
- FR-11: Findings are written to
  `docs/tasks/task-057-impeccable-redesign/design-audit.md` (distinct from
  the code-review `audit.md`), each with: ID, severity (Critical/High/
  Medium/Low), surface(s), theme(s), `file:line` or component, description,
  and proposed fix.
- FR-12: The audit is run against the codebase *as of the task start*, and
  its findings are mapped to the rollout stage that resolves them.

### 4.4 Remediation & rollout
- FR-13: Every finding in `design-audit.md` is resolved, and its row is
  marked with the commit that resolves it. No finding is deferred.
- FR-14: Rollout is staged; each stage is a separate PR off the task branch
  line, merged in order:
  1. Foundation: tokens, typography, `components/ui/*`, `components/common/*`.
  2. Shell: app shell, sidebar, mobile header/drawer, login, onboarding,
     error pages.
  3+. Feature areas (grouping finalized in `plan.md`), e.g. dashboards/kiosk;
     meals/recipes/ingredients/shopping; calendar/tasks/reminders;
     habits/workouts; packages/weather/wish list; households/settings/data
     retention.
- FR-15: Kiosk/dashboard: widgets restyled for legibility at distance;
  saved `react-grid-layout` layouts (positions, sizes, widget config) render
  with the same structure — no layout data migration, no change to
  `lib/dashboard` seed-layout semantics.
- FR-16: Behaviour is unchanged except where a finding requires a UX fix;
  each such behaviour change is called out in the stage's PR description.
- FR-17: After rollout, run `impeccable:impeccable-finish-reviewer` against
  the approved direction and fix every material issue it returns.

### 4.5 Accessibility (WCAG 2.2 AA)
- FR-18: Text contrast ≥ 4.5:1 (≥ 3:1 for large text and non-text UI /
  focus indicators) in both themes, including muted text, placeholders,
  disabled-but-readable states, chart colours, and badges.
- FR-19: Every interactive element has a visible focus indicator and is
  keyboard operable; dialogs/popovers trap and return focus.
- FR-20: Form controls have programmatic labels; icon-only buttons have
  accessible names; status is not conveyed by colour alone.
- FR-21: Pointer targets ≥ 24×24 CSS px (SC 2.5.8); mobile primary actions
  ≥ 44×44.
- FR-22: Motion respects `prefers-reduced-motion`.

## 5. API Surface

None. No endpoints are added or changed. The JSON:API contracts consumed by
the frontend are unchanged.

## 6. Data Model

None. No entities, fields, or migrations. The persisted theme preference
(`"light" | "dark"`) and dashboard layout JSON are unchanged.

## 7. Service Impact

- **frontend** — all changes: `src/index.css`, `src/components/ui/*`,
  `src/components/common/*`, `src/components/features/**`, `src/pages/**`,
  `public/` (favicon/brand mark if the direction changes them),
  `package.json` (font/other deps if justified), tests under `__tests__/`
  whose assertions depend on markup/classes that change.
- **Go services** — none. The `tools/verify.sh` docker leg rebuilds only
  the frontend image.
- **Docs** — new `frontend/DESIGN.md` (+ sidecar); task artifacts under
  `docs/tasks/task-057-impeccable-redesign/`.

## 8. Non-Functional Requirements

- **Performance:** production JS bundle (gzip) grows by no more than 10%
  versus `main` at task start; total self-hosted font payload for the
  default weights ≤ 150 KB. No layout shift from font loading (use
  `font-display` and matched fallbacks).
- **Security:** no third-party runtime requests (fonts, icons, images) —
  everything is bundled or self-hosted.
- **Multi-tenancy:** no change; tenant context handling in hooks/services is
  untouched.
- **Consistency:** no raw hex/rgb/oklch colour literals in
  `components/features/**` or `pages/**` after rollout (tokens only);
  exceptions listed in `DESIGN.md`.
- **Testing:** existing vitest suite passes; tests are updated, not deleted,
  when markup changes. Frontend guideline conventions (FE-* checklist) still
  apply.
- **Theming:** light and dark are both first-class; every token has a value
  in both.

## 9. Open Questions

- Product-context answers for onboarding (audience, tone, references,
  anti-references) are gathered interactively in `/design-task`.
- Exact feature-area grouping for stages 3+ and the number of PRs —
  settled in `plan.md` once the audit sizes the work.
- Whether the brand mark / favicon changes with the new direction — decided
  with the direction in `design.md`.
- How the staged PRs relate to the single task branch (stacked branches
  `task-057-impeccable-redesign-stage-N` vs. sequential merges of the task
  branch) — settled in `plan.md` per `docs/git-workflow.md`.

## 10. Acceptance Criteria

- [ ] `frontend/DESIGN.md` and its sidecar exist, describe product context
      and the shipped design system, and were produced by the Impeccable
      documenter from the final build.
- [ ] `design.md` records the approved visual direction (light + dark
      palettes, type, shape, motion, deps), with the user's approval noted.
- [ ] `design-audit.md` lists every finding with severity, surface, theme,
      location, and fix; every row is marked resolved with a commit ref.
- [ ] All routes in `App.tsx` and the shared shell reflect the new direction
      on desktop, mobile, and kiosk, in light and dark modes.
- [ ] WCAG 2.2 AA: contrast, focus visibility, keyboard operation, labels,
      target size, and reduced motion verified per FR-18–FR-22 in both
      themes; method and results recorded in `design-audit.md`.
- [ ] Saved dashboard layouts render with unchanged structure; existing
      `lib/dashboard` tests pass unmodified.
- [ ] No colour literals in `components/features/**` or `pages/**` outside
      documented exceptions.
- [ ] Bundle and font budgets in §8 met; numbers recorded in the final PR.
- [ ] `impeccable:impeccable-finish-reviewer` returns no unresolved material
      fixes.
- [ ] Each staged PR passed the code-review step (frontend-guidelines and
      plan-adherence reviewers) and a flagless `tools/verify.sh` exit 0.
