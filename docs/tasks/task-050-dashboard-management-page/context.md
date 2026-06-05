# Dashboard Management Page — Context

Companion to `plan.md`. Captures the key files, decisions, dependencies, and
gotchas an executor needs. All paths are relative to the repo root
(worktree `.worktrees/task-050-dashboard-management-page`).

## Scope

Frontend-only **move + extract** refactor. No Go services, shared libraries,
Docker builds, migrations, API contracts, or data-model changes. Every hook the
page needs already exists. Working directory for all commands is `frontend/`.

## Key files

### Sources (read before editing)
- `frontend/src/components/features/navigation/dashboards-nav-group.tsx` — today
  holds everything: `sortDashboards` + `computeReorderEntries` (exported),
  `SortableDashboardRow` (grip + NavLink + kebab), the group, dnd wiring, the
  "New Dashboard" button + `NewDashboardModal`. Source of the extraction.
- `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx` — the
  full action set (rename dialog, set-default, promote|copy, delete dialog).
  Reused wholesale; only the trigger className changes (Task 2).
- `frontend/src/components/features/dashboards/new-dashboard-modal.tsx` — reused
  as-is. Exports `dashboardNameSchema` (imported by the kebab). The modal
  navigates to `/app/dashboards/:id/edit` on create.
- `frontend/src/pages/HouseholdsPage.tsx` — the layout pattern to mirror:
  `p-4 md:p-6 space-y-4`, header `flex items-center justify-between` with
  `h1 text-xl md:text-2xl font-semibold`, loading `Skeleton` with
  `role="status" aria-label="Loading"`, error via `ErrorCard`, centered empty
  block.
- `frontend/src/App.tsx` — routes. `/app/dashboards` currently →
  `DashboardsIndexRedirect`. Pages are imported eagerly except the lazy
  `DashboardDesigner`.
- `frontend/src/components/features/navigation/nav-config.ts` — shared
  `navGroups`; both desktop sidebar and mobile drawer render the `management`
  group, so one entry appears in both.
- `frontend/src/pages/DashboardsIndexRedirect.tsx` — a one-line re-export of
  `DashboardRedirect`. Only `App.tsx` imports it (grep-verified). Safe to delete.

### Hooks (signatures verified in source)
- `useDashboards()` → `{ data, isLoading, isError }`, `data.data: Dashboard[]`.
- `useReorderDashboards()` → `.mutate(entries: DashboardOrderEntry[])`. **One
  call per scope; the backend rejects mixed-scope batches.**
- `useCreateDashboard / useUpdateDashboard / useDeleteDashboard /
  usePromoteDashboard / useCopyDashboardToMine` — used transitively via the
  kebab and modal; the page must mock the whole `use-dashboards` module in tests.
- `useHouseholdPreferences()` / `useUpdateHouseholdPreferences()` — from
  `@/lib/hooks/api/use-household-preferences`. **Gotcha:** existing code reads
  the default as `prefsData?.data?.[0]?.attributes.defaultDashboardId ?? null`
  (treats `.data` as an array and indexes `[0]`). Match this exact access shape
  — both `dashboard-kebab-menu.tsx` and the old nav group do.

### Types
- `frontend/src/types/models/dashboard.ts` — `Dashboard`, `DashboardAttributes`
  (`name`, `scope: "household" | "user"`, `sortOrder`, `createdAt`, …),
  `DashboardOrderEntry { id; sortOrder }`, `HouseholdPreferencesAttributes`.

## Decisions locked from design.md

| # | Decision |
|---|----------|
| Q1 | Per-row actions = **always-visible kebab** (reuse `DashboardKebabMenu`, drop hover-gated opacity). |
| Q2 | Reorder = **drag-only** (dnd-kit, keyboard-operable). No move-up/down buttons. |
| Q3 | Page header = `h1` "Dashboards" + primary "New Dashboard" button, no description. Mirror `HouseholdsPage`. |
| Q4 | Nav icon = **`LayoutGrid`** (distinct from the per-dashboard `LayoutDashboard`). |
| §3.2 | Helpers move to **`lib/dashboard/ordering.ts`** (option A), tests move with them. Nav group + page import from there. |
| §3.5 | **Two `DndContext`s, one per scope** — makes cross-scope drags structurally impossible. |

## Dependencies / ordering rationale

Tasks are ordered so each commit type-checks and tests green:
1. **Move helpers** first (Task 1) — both the page and the stripped nav group
   depend on `lib/dashboard/ordering.ts`. The nav group keeps importing them so
   it stays functional until Task 6.
2. **Kebab restyle** (Task 2) — small, isolated; the existing kebab test only
   uses the accessible label, so no test change is needed.
3. **`DashboardRow`** (Task 3) — depends on the restyled kebab; consumed by the page.
4. **Page** (Task 4) — depends on the ordering module + `DashboardRow` + modal.
5. **Wire routing + nav + delete redirect** (Task 5) — depends on the page existing.
6. **Strip nav group** (Task 6) — last, after the page owns reorder/kebab/new,
   so removing them from the sidebar is safe. `tsc -b` confirms no orphaned
   imports of the removed exports.
7. **Full gate** (Task 7) — `tsc -b`, `npm run lint`, `npm run test`.

## Reuse map

| Piece | Action |
|-------|--------|
| `sortDashboards`, `computeReorderEntries` | Move to `lib/dashboard/ordering.ts` (tests move too). |
| `SortableDashboardRow` markup | Extract → `DashboardRow` (full-width card, open + edit links, always-visible kebab). |
| `DashboardKebabMenu` | Reuse; restyle trigger only. |
| `NewDashboardModal` | Reuse as-is. |
| dnd-kit sensors (`PointerSensor` + `KeyboardSensor` w/ `sortableKeyboardCoordinates`) | Reuse, two contexts (one per scope). |
| `dashboardNameSchema` | Reused via the kebab's rename dialog. |

## Verification commands (run from `frontend/`)

- Single test file: `npx vitest run <path>`
- Type-check: `npx tsc -b`
- Lint: `npm run lint`
- Full suite: `npm run test`

## Test conventions (from existing suites)

- Mock hook modules with `vi.mock("@/lib/hooks/api/...")` returning plain
  objects; capture call spies with module-level `vi.fn()` + a getter
  (`useDashboards: () => mockUseDashboards()`), reset in `beforeEach` with
  `vi.clearAllMocks()`. See `dashboards-nav-group.test.tsx` and
  `HouseholdsPage.test.tsx`.
- Wrap renders in `<MemoryRouter>`.
- The `dash(...)` factory (id + attribute overrides) is the standard fixture
  shape; reuse it in new tests.
- dnd-kit drag is **not** simulated in unit tests — assert rendering, links,
  sections, empty/loading states, and modal open instead. `useSortable` renders
  fine outside a `SortableContext` (no-op draggable), so `DashboardRow` can be
  unit-tested in isolation.

## Risks (from design §9)

- Helper move breaks imports → update nav group import + test path in the same
  task; `tsc -b` catches stragglers.
- Kebab restyle leaks to another consumer → grep confirms only the nav group
  (which drops it) and tests import the kebab; restyle is page-only in practice.
- Two `DndContext`s + drag styling on the card layout → reuse the proven sensor
  config; `dashboard-row` test + manual mobile-width check.

## Acceptance criteria (PRD §10) → tasks

All FR/AC map to tasks per the plan's Self-Review section. The headline checks:
`/app/dashboards` renders the page (not a redirect); `/app`, `/app/dashboard`,
post-login landing still resolve to the default dashboard (untouched); Manage
Dashboards appears in the Management group; per-scope reorder persists via one
call; full action set on each row; sidebar is nav-only; `DashboardsIndexRedirect`
removed; type-check + lint + affected suites pass.
