import React from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { AppRoutes } from "../app-routes";
import { LocationProbe } from "@/test/location-probe";

// Shared by Tasks 4-8: counts AppShell mounts so tests can assert the shell is
// not remounted by in-page navigation.
export const shellMounts = { count: 0 };

vi.mock("@/components/features/navigation/protected-route", () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock("@/components/features/navigation/app-shell", async () => {
  const React = await import("react");
  const { Outlet } = await import("react-router-dom");
  return {
    AppShell: () => {
      React.useEffect(() => {
        shellMounts.count += 1;
      }, []);
      return <Outlet />;
    },
  };
});

// Page stubs render `<name>:<view|tab>` so later tasks can assert which view
// a URL selects.
async function stub(name: string) {
  const { useSearchParams } = await import("react-router-dom");
  return function Stub(props: { tab?: string; view?: string }) {
    const [sp] = useSearchParams();
    return (
      <div data-testid="page">
        {`${name}:${props.view ?? props.tab ?? sp.get("view") ?? sp.get("tab") ?? ""}`}
      </div>
    );
  };
}

vi.mock("@/pages/TrackerPage", async () => ({ TrackerPage: await stub("TrackerPage") }));
vi.mock("@/pages/CalendarPage", async () => ({ CalendarPage: await stub("CalendarPage") }));
vi.mock("@/pages/IngredientsPage", async () => ({ IngredientsPage: await stub("IngredientsPage") }));
vi.mock("@/pages/IngredientDetailPage", async () => ({ IngredientDetailPage: await stub("IngredientDetailPage") }));
vi.mock("@/pages/ShoppingListsPage", async () => ({ ShoppingListsPage: await stub("ShoppingListsPage") }));
vi.mock("@/pages/ShoppingListDetailPage", async () => ({ ShoppingListDetailPage: await stub("ShoppingListDetailPage") }));
vi.mock("@/pages/MealsPage", async () => ({ MealsPage: await stub("MealsPage") }));

export function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <AppRoutes />
      <LocationProbe />
    </MemoryRouter>,
  );
}

describe("AppRoutes", () => {
  beforeEach(() => {
    shellMounts.count = 0;
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 10, 12, 12));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("redirects /app/tracker to /app/habits/today preserving the query", () => {
    renderAt("/app/tracker?x=1");
    expect(screen.getByTestId("location").textContent).toBe("/app/habits/today?x=1");
  });

  it.each(["today", "calendar", "setup", "report"])("routes /app/habits/%s to that view", (view) => {
    renderAt(`/app/habits/${view}`);
    expect(screen.getByTestId("page").textContent).toBe(`TrackerPage:${view}`);
  });

  it.each(["/app/habits", "/app/habits/bogus", "/app/tracker"])("redirects %s to /app/habits/today", (url) => {
    renderAt(url);
    expect(screen.getByTestId("location").textContent).toBe("/app/habits/today");
  });

  it("preserves month when redirecting /app/habits", () => {
    renderAt("/app/habits?month=2026-01");
    expect(screen.getByTestId("location").textContent).toBe("/app/habits/today?month=2026-01");
  });

  it("does not remount the shell when switching habits views", async () => {
    const { useNavigate } = await import("react-router-dom");
    function PushButton() {
      const navigate = useNavigate();
      return <button onClick={() => navigate("/app/habits/calendar")}>push-calendar</button>;
    }
    render(
      <MemoryRouter initialEntries={["/app/habits/today"]}>
        <AppRoutes />
        <LocationProbe />
        <PushButton />
      </MemoryRouter>,
    );
    expect(shellMounts.count).toBe(1);
    fireEvent.click(screen.getByText("push-calendar"));
    expect(screen.getByTestId("location").textContent).toBe("/app/habits/calendar");
    expect(shellMounts.count).toBe(1);
  });

  it("redirects /app/shopping to /app/shopping/grocery", () => {
    renderAt("/app/shopping");
    expect(screen.getByTestId("location").textContent).toBe("/app/shopping/grocery");
  });

  it("routes an ingredient UUID to IngredientDetailPage", () => {
    renderAt("/app/ingredients/0b6c1f2e-3a4d-4e5f-8a9b-0c1d2e3f4a5b");
    expect(screen.getByTestId("page").textContent).toBe("IngredientDetailPage:");
  });

  it("routes /app/shopping/grocery to the active tab", () => {
    renderAt("/app/shopping/grocery");
    expect(screen.getByTestId("page").textContent).toBe("ShoppingListsPage:active");
  });

  it("routes /app/shopping/grocery/archived to the archived tab, not :id", () => {
    renderAt("/app/shopping/grocery/archived");
    expect(screen.getByTestId("page").textContent).toBe("ShoppingListsPage:archived");
  });

  it("routes /app/shopping/grocery/list-1 to ShoppingListDetailPage", () => {
    renderAt("/app/shopping/grocery/list-1");
    expect(screen.getByTestId("page").textContent).toBe("ShoppingListDetailPage:");
  });

  it("redirects /app/ingredients to the list view", () => {
    renderAt("/app/ingredients");
    expect(screen.getByTestId("location").textContent).toBe("/app/ingredients/list");
  });

  it("preserves the query when redirecting /app/ingredients", () => {
    renderAt("/app/ingredients?category=abc");
    expect(screen.getByTestId("location").textContent).toBe("/app/ingredients/list?category=abc");
  });

  it.each(["list", "categories", "bulk"])("routes /app/ingredients/%s to that view", (view) => {
    renderAt(`/app/ingredients/${view}`);
    expect(screen.getByTestId("page").textContent).toBe(`IngredientsPage:${view}`);
  });

  it("routes an unknown ingredients segment to the detail page", () => {
    renderAt("/app/ingredients/not-a-view");
    expect(screen.getByTestId("page").textContent).toBe("IngredientDetailPage:");
  });

  it.each(["week", "month"])("routes /app/calendar/%s to that view", (view) => {
    renderAt(`/app/calendar/${view}`);
    expect(screen.getByTestId("page").textContent).toBe(`CalendarPage:${view}`);
  });

  it.each([
    ["/app/calendar", "/app/calendar/week"],
    ["/app/calendar/bogus", "/app/calendar/week"],
    ["/app/calendar?connected=true", "/app/calendar/week?connected=true"],
    ["/app/calendar?error=auth_failed", "/app/calendar/week?error=auth_failed"],
  ])("redirects %s to %s", (from, to) => {
    renderAt(from);
    expect(screen.getByTestId("location").textContent).toBe(to);
  });
});
