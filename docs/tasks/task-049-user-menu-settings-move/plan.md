# User Menu Settings Move Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move "Settings" from the desktop sidebar / mobile drawer footers into the user-profile pop-over (alongside Sign Out), and drop the Light/Dark theme toggle from that pop-over.

**Architecture:** Three navigation files in `frontend/src/components/features/navigation/` change. `user-menu.tsx` gains a Settings menu item and loses the theme-toggle item. `app-shell.tsx` and `mobile-drawer.tsx` each lose their bordered Settings footer block. `nav-config.ts` is unchanged structurally — `settingsNavItem` (path `/app/settings`, label `Settings`, lucide `Settings` icon) keeps its shape; only its consumer relocates from the two layout files to `user-menu.tsx`. No backend, schema, or contract changes.

**Tech Stack:** React 19, TypeScript, Vite, react-router-dom (`useNavigate`), `@base-ui/react/menu`, lucide-react, Tailwind, vitest + @testing-library/react.

---

## Working Directory

All commands run from `frontend/` inside the task worktree:
`/home/tumidanski/source/home-hub/.worktrees/task-049-user-menu-settings-move/frontend`.

All file paths in this plan are relative to the worktree root unless explicitly noted.

---

## Task 1: Add `user-menu.test.tsx` covering the new pop-over contract (failing first)

We start with the new test file because `UserMenu` is the file with the most behavior change, and the design (§6.2) makes this the file that owns the pop-over contract going forward. Writing the test first locks the desired behavior before we touch `user-menu.tsx`.

**Files:**
- Create: `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx`

- [ ] **Step 1.1: Create the failing test file**

Create `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx` with this exact content:

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockUseAuth = vi.fn();
const mockLogoutMutate = vi.fn();
const mockNavigate = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/api/use-auth", () => ({
  useLogout: () => ({ mutate: mockLogoutMutate }),
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

import { UserMenu } from "../user-menu";

function renderUserMenu(props: { onAction?: () => void } = {}) {
  return render(
    <MemoryRouter>
      <UserMenu {...props} />
    </MemoryRouter>,
  );
}

describe("UserMenu", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: {
        id: "test-user-id",
        attributes: {
          displayName: "Test User",
          email: "test@example.com",
          avatarUrl: "",
          providerAvatarUrl: "",
        },
      },
    });
  });

  it("renders nothing when there is no user", () => {
    mockUseAuth.mockReturnValue({ user: null });
    const { container } = renderUserMenu();
    expect(container).toBeEmptyDOMElement();
  });

  it("renders only Settings and Sign Out items in the pop-over, in that order", async () => {
    const user = userEvent.setup();
    renderUserMenu();
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const items = await screen.findAllByRole("menuitem");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent(/settings/i);
    expect(items[1]).toHaveTextContent(/sign out/i);
  });

  it("does not render any theme toggle item", async () => {
    const user = userEvent.setup();
    renderUserMenu();
    await user.click(screen.getByRole("button", { name: /test user/i }));
    await screen.findByRole("menuitem", { name: /sign out/i });
    expect(screen.queryByRole("menuitem", { name: /dark mode/i })).toBeNull();
    expect(screen.queryByRole("menuitem", { name: /light mode/i })).toBeNull();
  });

  it("Settings item navigates to /app/settings and calls onAction", async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    renderUserMenu({ onAction });
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const settingsItem = await screen.findByRole("menuitem", { name: /settings/i });
    await user.click(settingsItem);
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    expect(mockNavigate).toHaveBeenCalledWith("/app/settings");
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it("Sign Out item triggers logout and calls onAction", async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    renderUserMenu({ onAction });
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const signOutItem = await screen.findByRole("menuitem", { name: /sign out/i });
    await user.click(signOutItem);
    expect(mockLogoutMutate).toHaveBeenCalledTimes(1);
    expect(onAction).toHaveBeenCalledTimes(1);
  });
});
```

- [ ] **Step 1.2: Run the new test to verify it fails**

Run from `frontend/`:

```bash
npm test -- src/components/features/navigation/__tests__/user-menu.test.tsx
```

Expected: the test file runs but several assertions FAIL. In particular:
- `renders only Settings and Sign Out items` — fails because today's pop-over renders the theme toggle as the first item, not Settings; `items[0]` text will be `Dark Mode` (or `Light Mode`), not `Settings`.
- `Settings item navigates to /app/settings` — fails with "Unable to find a menuitem with name /settings/i" because no such item exists yet.
- `does not render any theme toggle item` — fails because Dark Mode is currently rendered.

The `Sign Out` and null-user tests should pass even pre-implementation; that is fine.

If the file fails to compile (e.g. `UserMenu` is not exported, the test rendering crashes for some unrelated reason), stop and investigate before continuing — do not proceed to Task 2 with a non-compiling test file.

- [ ] **Step 1.3: Commit the failing test**

```bash
git add frontend/src/components/features/navigation/__tests__/user-menu.test.tsx
git commit -m "test(user-menu): add pop-over contract tests (failing)"
```

---

## Task 2: Update `user-menu.tsx` to make the new tests pass

Replace the theme-toggle item with a Settings item that uses `useNavigate`, drop unused theme imports, and add the new ones. The Sign Out item stays exactly as-is.

**Files:**
- Modify: `frontend/src/components/features/navigation/user-menu.tsx` (full replacement of lines 1–80)

- [ ] **Step 2.1: Rewrite `user-menu.tsx`**

Replace the entire content of `frontend/src/components/features/navigation/user-menu.tsx` with:

```tsx
import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { LogOut, ChevronDown, Settings } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/components/providers/auth-provider";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { settingsNavItem } from "@/components/features/navigation/nav-config";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";

interface UserMenuProps {
  onAction?: () => void;
  iconSize?: string;
}

export function UserMenu({ onAction, iconSize = "h-4 w-4" }: UserMenuProps) {
  const { user } = useAuth();
  const logout = useLogout();
  const navigate = useNavigate();

  if (!user) return null;

  return (
    <MenuPrimitive.Root>
      <MenuPrimitive.Trigger
        className={cn(
          "flex w-full cursor-pointer items-center justify-between rounded-md p-3 text-left transition-colors hover:bg-sidebar-accent/50 outline-none focus-visible:ring-2 focus-visible:ring-ring",
        )}
      >
        <UserAvatar
          avatarUrl={user.attributes.avatarUrl}
          providerAvatarUrl={user.attributes.providerAvatarUrl}
          displayName={user.attributes.displayName}
          userId={user.id}
          size="sm"
        />
        <div className="ml-2 min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{user.attributes.displayName}</p>
          <p className="truncate text-xs text-muted-foreground">{user.attributes.email}</p>
        </div>
        <ChevronDown className="ml-2 h-4 w-4 shrink-0 text-muted-foreground" />
      </MenuPrimitive.Trigger>
      <MenuPrimitive.Portal>
        <MenuPrimitive.Positioner side="top" sideOffset={8} align="start" className="z-50">
          <MenuPrimitive.Popup
            className={cn(
              "min-w-48 rounded-lg bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/10",
              "origin-(--transform-origin) transition-[transform,scale,opacity] duration-100",
              "data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95",
              "data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
            )}
          >
            <MenuPrimitive.Item
              className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
              onClick={() => {
                navigate(settingsNavItem.to);
                onAction?.();
              }}
            >
              <Settings className={iconSize} />
              {settingsNavItem.label}
            </MenuPrimitive.Item>
            <MenuPrimitive.Item
              className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm text-destructive outline-none select-none hover:bg-accent focus:bg-accent"
              onClick={() => {
                logout.mutate();
                onAction?.();
              }}
            >
              <LogOut className={iconSize} />
              Sign Out
            </MenuPrimitive.Item>
          </MenuPrimitive.Popup>
        </MenuPrimitive.Positioner>
      </MenuPrimitive.Portal>
    </MenuPrimitive.Root>
  );
}
```

Notable diffs vs. previous version:
- Removed imports: `Moon`, `Sun` (lucide-react); `useThemeToggle` (`@/lib/hooks/use-theme-toggle`).
- Added imports: `Settings` (lucide-react); `useNavigate` (`react-router-dom`); `settingsNavItem` (`@/components/features/navigation/nav-config`).
- Removed: `const { theme, toggleTheme } = useThemeToggle();`.
- Added: `const navigate = useNavigate();`.
- Replaced the first `MenuPrimitive.Item` (the theme-toggle item) with the Settings item. Same outer class string (neutral styling), `onClick` does `navigate(settingsNavItem.to); onAction?.()` in that order.
- Sign Out item is byte-for-byte unchanged.

- [ ] **Step 2.2: Run the new test file and confirm it passes**

```bash
npm test -- src/components/features/navigation/__tests__/user-menu.test.tsx
```

Expected: all 5 tests pass.

If any fail, do NOT loosen the test. Read the failure carefully:
- "menu primitive items not rendering" usually means `MemoryRouter` is missing or `useNavigate` is being called outside it — verify the test mocks match the production import path (`react-router-dom`).
- Order failures (`items[0]` was Sign Out instead of Settings) mean the JSX inside the popup is in the wrong order — Settings must be the first `MenuPrimitive.Item`.

- [ ] **Step 2.3: Commit**

```bash
git add frontend/src/components/features/navigation/user-menu.tsx
git commit -m "feat(user-menu): replace theme toggle with Settings navigation item"
```

---

## Task 3: Update `app-shell.test.tsx` to drop the obsolete assertions (failing first)

The existing sidebar test asserts the very things this task removes. Update the test before changing the component so we can watch the test fail-then-pass cycle.

**Files:**
- Modify: `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx`

- [ ] **Step 3.1: Edit `app-shell.test.tsx`**

Apply these three changes to `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx`:

**Change A — remove the `useThemeToggle` mock block (lines 14–16 in the current file).** Delete:

```tsx
vi.mock("@/lib/hooks/use-theme-toggle", () => ({
  useThemeToggle: () => ({ theme: "light", toggleTheme: mockToggleTheme }),
}));
```

Also delete the now-unused `const mockToggleTheme = vi.fn();` declaration at the top of the file.

**Change B — remove the `Settings` assertion from `"renders sidebar with navigation links"`.** Delete this line from inside that test:

```tsx
expect(screen.getByText("Settings")).toBeInTheDocument();
```

Leave the other assertions (`Home Hub`, `Dashboards`, `Tasks`, `Reminders`, `Households`) intact.

**Change C — delete the entire `"calls toggleTheme when theme button is clicked in user menu"` test.** Remove the whole `it("calls toggleTheme when theme button is clicked in user menu", ...)` block. The `"calls logout when sign out is clicked in user menu"` test directly below it stays unchanged.

After the edit, the file's mock list should be: `mockUseAuth`, `mockLogoutMutate` (no `mockToggleTheme`). The `useThemeToggle` mock block should be gone.

- [ ] **Step 3.2: Run the test file and verify it fails on the right thing**

Run from `frontend/`:

```bash
npm test -- src/components/features/navigation/__tests__/app-shell.test.tsx
```

Expected: `"renders sidebar with navigation links"` still passes (we just removed an assertion), but the suite will FAIL or warn on `"calls logout when sign out is clicked in user menu"` only if our test mocks for `useThemeToggle` were load-bearing — they shouldn't be. The most likely outcome is the file passes entirely. If it does pass, that's fine; the goal of step 3.1 was to clean up assertions that would be wrong after Task 4. Continue.

If the suite fails for any reason other than "Settings text still present" (which is impossible at this point — we haven't yet removed the sidebar Settings link), stop and investigate.

> Why we run the file even though it may pass: this is an explicit checkpoint. After Task 4, this same `npm test` invocation must continue to pass. Establishing the baseline now makes the post-Task-4 run unambiguous.

- [ ] **Step 3.3: Commit**

```bash
git add frontend/src/components/features/navigation/__tests__/app-shell.test.tsx
git commit -m "test(app-shell): drop assertions for removed sidebar Settings link and theme toggle"
```

---

## Task 4: Remove the Settings footer block from `app-shell.tsx`

**Files:**
- Modify: `frontend/src/components/features/navigation/app-shell.tsx`

- [ ] **Step 4.1: Delete the bordered Settings `NavLink` block**

Remove the following block (currently lines 55–70 in the file) entirely:

```tsx
        <div className="border-t p-2">
          <NavLink
            to={settingsNavItem.to}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent/50",
              )
            }
          >
            <Settings className="h-4 w-4" />
            {settingsNavItem.label}
          </NavLink>
        </div>
```

The `<div className="border-t"><UserMenu /></div>` block immediately after it MUST stay.

- [ ] **Step 4.2: Update imports**

In the same file:

- Remove `NavLink` from the `react-router-dom` import (`Outlet, NavLink, Link` → `Outlet, Link`). Confirm no other use of `NavLink` remains in the file by searching for it inside `app-shell.tsx`; if any other use exists, keep `NavLink`. As of the current file there is none.
- Remove `Settings` from the `lucide-react` import line entirely (`import { Settings } from "lucide-react";` is the only thing on its line — delete the whole line).
- Remove `settingsNavItem` from the `nav-config` import: `import { navGroups, settingsNavItem } from "..."` → `import { navGroups } from "..."`.

After the edit, the top of `app-shell.tsx` should look like:

```tsx
import { useState, useMemo } from "react";
import { Outlet, Link } from "react-router-dom";
import { useAuth } from "@/components/providers/auth-provider";
import { HouseholdSwitcher } from "@/components/features/households/household-switcher";
import { MobileHeader } from "@/components/features/navigation/mobile-header";
import { MobileDrawer } from "@/components/features/navigation/mobile-drawer";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { usePackageSummary } from "@/lib/hooks/api/use-packages";
import { cn } from "@/lib/utils";
```

`cn` MUST stay imported — it is still used in the `<div className="flex h-screen ...">` outer wrapper. (Sanity check: `grep -n "cn(" app-shell.tsx` should still return at least one hit. If it returns zero hits, `cn` can be dropped from imports too; otherwise leave it.)

- [ ] **Step 4.3: Run the navigation tests**

```bash
npm test -- src/components/features/navigation/__tests__/
```

Expected: all three navigation test files pass — `app-shell.test.tsx`, `user-menu.test.tsx`, `dashboards-nav-group.test.tsx`, `protected-route.test.tsx`.

If `app-shell.test.tsx`'s `"renders sidebar with navigation links"` fails on `getByText("Settings")` no longer being present, that means Step 3.1's Change B did not actually remove the assertion — go back to Task 3 and fix.

- [ ] **Step 4.4: TypeScript compile check**

Run from `frontend/`:

```bash
npm run build
```

Expected: build succeeds with no errors. The build step catches unused-import errors (TypeScript's `noUnusedLocals`/`noUnusedParameters` if configured, or the bundler's tree-shake warnings). If the build complains about an unused import in `app-shell.tsx`, fix the offending import line directly.

> Note: If `npm run build` is slow because it builds the entire bundle, an alternative quick check is `npx tsc -b --noEmit` (or equivalent — match what `package.json`'s `build` script does). Either is acceptable; the requirement is "TypeScript is happy."

- [ ] **Step 4.5: Commit**

```bash
git add frontend/src/components/features/navigation/app-shell.tsx
git commit -m "refactor(app-shell): remove Settings footer link from desktop sidebar"
```

---

## Task 5: Remove the Settings footer block from `mobile-drawer.tsx`

Same shape as Task 4, applied to the mobile drawer.

**Files:**
- Modify: `frontend/src/components/features/navigation/mobile-drawer.tsx`

- [ ] **Step 5.1: Delete the bordered Settings `NavLink` block**

Remove the following block (currently lines 96–113 in the file):

```tsx
        {/* Footer */}
        <div className="border-t p-3">
          <NavLink
            to={settingsNavItem.to}
            onClick={onClose}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent/50",
              )
            }
          >
            <Settings className="h-5 w-5" />
            {settingsNavItem.label}
          </NavLink>
        </div>
```

The `<div className="border-t"><UserMenu onAction={onClose} iconSize="h-5 w-5" /></div>` block immediately after it MUST stay. That is the surface the tests in Task 1 already cover via the `onAction` prop.

- [ ] **Step 5.2: Update imports**

In the same file:

- Remove `NavLink` from the `react-router-dom` import line entirely (it was the only named import: `import { NavLink } from "react-router-dom";` → delete the whole line).
- Remove `Settings` from the `lucide-react` import: `import { Settings, X } from "lucide-react";` → `import { X } from "lucide-react";`.
- Remove the entire `import { navGroups, settingsNavItem } from "..."` statement and replace with `import { navGroups } from "@/components/features/navigation/nav-config";`.

After the edit, the top of `mobile-drawer.tsx` should look like:

```tsx
import { useEffect, useRef } from "react";
import { X } from "lucide-react";
import { createPortal } from "react-dom";
import { MobileHouseholdSelector } from "@/components/features/households/mobile-household-selector";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
```

`cn` MUST stay imported — it is still used in the overlay class string and the drawer panel class string. (Sanity check: `grep -n "cn(" mobile-drawer.tsx` should return at least two hits.)

- [ ] **Step 5.3: Run the full frontend test suite**

```bash
npm test
```

Expected: every test file passes. We run the full suite here because this is the last code change of the plan; we want a clean baseline before final verification.

If any test outside `__tests__/navigation/` fails, that is a real regression — investigate before continuing. Do not skip or comment out unrelated tests.

- [ ] **Step 5.4: Commit**

```bash
git add frontend/src/components/features/navigation/mobile-drawer.tsx
git commit -m "refactor(mobile-drawer): remove Settings footer link"
```

---

## Task 6: Final verification

A single explicit checkpoint that exercises both build and tests in the configuration the user will see.

**Files:** none modified.

- [ ] **Step 6.1: Build**

From `frontend/`:

```bash
npm run build
```

Expected: build succeeds, no warnings about unused imports in `user-menu.tsx`, `app-shell.tsx`, or `mobile-drawer.tsx`. If TypeScript reports an unused-locals error in any of those files, fix the offending import; do not silence the warning.

- [ ] **Step 6.2: Test**

From `frontend/`:

```bash
npm test
```

Expected: every test passes.

This step is non-negotiable per repo policy (`CLAUDE.md` "Build & Verification") — running `npm run build` alone is not enough.

- [ ] **Step 6.3: Smoke check the diff**

From the worktree root:

```bash
git diff --stat main...HEAD
```

Expected file list (file count: 5):

- `docs/tasks/task-049-user-menu-settings-move/context.md` (added by `/plan-task`)
- `docs/tasks/task-049-user-menu-settings-move/plan.md` (added by `/plan-task`)
- `frontend/src/components/features/navigation/__tests__/app-shell.test.tsx`
- `frontend/src/components/features/navigation/__tests__/user-menu.test.tsx` (new)
- `frontend/src/components/features/navigation/app-shell.tsx`
- `frontend/src/components/features/navigation/mobile-drawer.tsx`
- `frontend/src/components/features/navigation/user-menu.tsx`

`nav-config.ts` MUST NOT appear in the diff. If it does, you accidentally edited it — revert it; the design phase explicitly committed to leaving `settingsNavItem` structurally unchanged.

`use-theme-toggle.ts` (and anything under `src/pages/settings/`) MUST NOT appear. The Settings page and its theme toggle remain untouched.

- [ ] **Step 6.4: Manual smoke test (optional but recommended)**

If the user has time, run the dev server and hand-verify the four flows:

1. Desktop sidebar: footer no longer shows Settings link; only the user-profile control with the chevron.
2. Click user-profile → pop-over shows two items in order: Settings, Sign Out. No Light/Dark toggle anywhere in the pop-over.
3. Click Settings → URL becomes `/app/settings`, page renders, no full reload (network tab does not show a document fetch).
4. Mobile drawer (resize browser to <768px): drawer footer has only the user-profile control. Open pop-over → Settings is the first item. Click Settings → drawer closes AND URL becomes `/app/settings`.

If any flow fails, that is a real bug — diagnose and fix; do not declare the task complete.

---

## Self-Review (already performed by the plan author)

- **Spec coverage:** Every PRD acceptance criterion (`prd.md` §10) and every design coverage row (`design.md` §8) is addressed by at least one task: pop-over content (Task 1, 2), navigation primitive (Task 2), drawer-close-after-activation (Task 1, 2), Sign Out preserved (Task 1, 2), theme toggle removed from pop-over but still present on Settings page (Task 2 plus untouched out-of-scope file), sidebar/drawer footer removal (Tasks 4, 5), keyboard accessibility (covered by base-ui defaults; tests in Task 1 exercise activation via the `menuitem` role), build/tests green (Task 6), no backend changes (entire plan is FE-only), test updates (Tasks 1, 3).
- **Placeholder scan:** No "TBD", "implement later", "appropriate error handling" markers; every code step contains the actual code.
- **Type consistency:** Across tasks, the prop name is `onAction`, the icon prop is `iconSize`, the route constant is `settingsNavItem.to` (= `/app/settings`), the navigation hook is `useNavigate`, the mock variable is `mockNavigate`. These match `prd.md` §4.1, `design.md` §3.1, and the existing source files verified during planning.
