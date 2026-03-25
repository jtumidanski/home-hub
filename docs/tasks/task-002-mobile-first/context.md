# Mobile-First Responsive UI — Context

Last Updated: 2026-03-25

---

## Key Files

### Navigation & Layout
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `frontend/src/components/features/navigation/app-shell.tsx` | Main layout shell — fixed 256px sidebar + main content | Conditionally render sidebar (md+) vs top bar + hamburger drawer (<md) |
| `frontend/src/components/features/households/household-switcher.tsx` | Compact Select dropdown for switching households | Add mobile variant: full-screen overlay selector |

### Pages
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `frontend/src/pages/DashboardPage.tsx` | Summary cards in `md:grid-cols-3` grid | Add pull-to-refresh, verify single-column stacking on mobile |
| `frontend/src/pages/TasksPage.tsx` | Task list via DataTable + column defs | Add card-based mobile view, pull-to-refresh, mobile add button |
| `frontend/src/pages/RemindersPage.tsx` | Reminder list via DataTable + column defs | Add card-based mobile view, pull-to-refresh, mobile add button |
| `frontend/src/pages/HouseholdsPage.tsx` | Household list via DataTable | Add card-based mobile view |
| `frontend/src/pages/SettingsPage.tsx` | Profile + appearance cards | Responsive spacing/sizing audit |
| `frontend/src/pages/LoginPage.tsx` | Centered `max-w-md` card | Verify at 320px, add horizontal padding |
| `frontend/src/pages/OnboardingPage.tsx` | Multi-step form in centered card | Verify at 320px, add horizontal padding |

### Shared Components
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `frontend/src/components/common/data-table.tsx` | Generic DataTable using @tanstack/react-table | No changes — mobile card views are separate components |
| `frontend/src/components/ui/button.tsx` | Button with size variants (xs/sm/default/lg/icon) | Audit touch target sizes; `xs` is h-6 (24px), below 44px minimum |
| `frontend/src/components/ui/input.tsx` | Text input | Verify min 16px font size to prevent iOS zoom |
| `frontend/src/components/ui/select.tsx` | Select dropdown | Used in household switcher; mobile gets full-screen alternative |
| `frontend/src/components/ui/dialog.tsx` | Modal dialog | Already has responsive `sm:max-w-sm` + mobile padding |
| `frontend/src/components/ui/skeleton.tsx` | Loading placeholder | No changes — already used across all pages |

### New Components to Create
| File | Purpose |
|------|---------|
| `frontend/src/components/features/navigation/mobile-header.tsx` | Top bar with hamburger icon, branding, visible below md |
| `frontend/src/components/features/navigation/mobile-drawer.tsx` | Slide-in navigation drawer with overlay |
| `frontend/src/components/features/households/mobile-household-selector.tsx` | Full-screen household selector overlay |
| `frontend/src/components/common/responsive-list.tsx` | Wrapper: renders DataTable on md+, card list on mobile |
| `frontend/src/components/common/pull-to-refresh.tsx` | Pull-to-refresh wrapper component |
| `frontend/src/components/features/tasks/task-card.tsx` | Mobile card view for a single task |
| `frontend/src/components/features/reminders/reminder-card.tsx` | Mobile card view for a single reminder |
| `frontend/src/components/features/households/household-card.tsx` | Mobile card view for a single household |
| `frontend/src/lib/hooks/use-mobile.ts` | Hook to detect mobile viewport (below md breakpoint) |

### Styles
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `frontend/src/index.css` | Global CSS with Tailwind theme | No changes expected — responsive via Tailwind classes |

### Tests
| File | Purpose |
|------|---------|
| `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx` | Existing AppShell tests — must continue passing |
| `frontend/src/pages/__tests__/TasksPage.test.tsx` | Existing TasksPage tests — must continue passing |
| `frontend/src/pages/__tests__/RemindersPage.test.tsx` | Existing RemindersPage tests |
| `frontend/src/pages/__tests__/DashboardPage.test.tsx` | Existing DashboardPage tests |
| `frontend/src/pages/__tests__/HouseholdsPage.test.tsx` | Existing HouseholdsPage tests |
| `frontend/src/pages/__tests__/SettingsPage.test.tsx` | Existing SettingsPage tests |
| `frontend/src/pages/__tests__/LoginPage.test.tsx` | Existing LoginPage tests |

---

## Key Decisions

1. **Breakpoint boundary: `md` (768px)** — Everything below gets mobile layout; at and above gets current desktop layout. Uses Tailwind default breakpoints.

2. **CSS-first responsive where possible** — Use Tailwind `md:` prefixed classes and `hidden md:flex` / `md:hidden` patterns over JS-based viewport detection. Exception: `use-mobile` hook needed for conditional component rendering (card vs DataTable).

3. **No DataTable modification** — The existing DataTable component stays as-is. Mobile card views are separate components rendered conditionally based on viewport. This avoids breaking the existing desktop experience.

4. **Drawer vs sheet** — shadcn/ui doesn't currently include a Sheet/Drawer component in this project. The mobile drawer will be built with the existing Dialog primitive patterns (overlay + positioned panel + focus trap), or a simple custom implementation with Tailwind transitions.

5. **Pull-to-refresh approach** — TBD: third-party library (e.g., `react-simple-pull-to-refresh`) vs lightweight custom using touch events. Custom is preferred to avoid dependency, but may use a small library if touch handling proves complex.

6. **Household switcher** — Full-screen overlay on mobile, replaces the Select dropdown. Accessed from the same position in the navigation drawer.

7. **Action menus on cards** — Use a three-dot icon button that opens an existing dropdown/popover pattern. Must support all actions available in the DataTable row (complete, delete, snooze, dismiss, etc.).

---

## Dependencies

### NPM Packages
- No new packages required for core responsive work
- Possible addition: pull-to-refresh library (decision pending)
- All existing packages: React 19, Tailwind v4, @tanstack/react-table, @tanstack/react-query, lucide-react, react-hook-form, zod, sonner

### Internal Dependencies
- Card components (Card, CardContent, CardHeader, CardTitle) — already exist in `components/ui/card.tsx`
- Badge component — already exists in `components/ui/badge.tsx`
- Button component — already exists, may need touch size audit
- Dialog primitives — available via `@base-ui/react`

### External Dependencies
- None — no backend changes required

---

## Risk Areas

1. **Touch target sizing** — The `xs` button size is 24px tall, well below the 44px minimum. Need to ensure it's not used in mobile-visible contexts, or add a mobile override.

2. **DataTable column definitions live in page files** — Task columns with `toggleComplete` and `handleDelete` callbacks are defined inline in TasksPage.tsx. The card components will need access to the same mutation hooks and handlers. Extract shared action handlers or pass them as props.

3. **HouseholdSwitcher returns null for single household** — The `if (households.length <= 1) return null` logic (line 26) needs to be preserved in the mobile variant.

4. **Existing tests use shallow rendering patterns** — Tests mock child components and check for text content. Mobile-conditional rendering may need viewport mocking in tests (e.g., `window.matchMedia` mock).

5. **Dialog component already responsive** — The existing Dialog has `sm:max-w-sm` and mobile padding. Create/edit dialogs should work on mobile without changes, but need verification.
