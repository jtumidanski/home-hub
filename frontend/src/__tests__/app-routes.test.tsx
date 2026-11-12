import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
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
  return function Stub() {
    const [sp] = useSearchParams();
    return <div data-testid="page">{`${name}:${sp.get("view") ?? sp.get("tab") ?? ""}`}</div>;
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
  });

  it("redirects /app/tracker to /app/habits preserving the query", () => {
    renderAt("/app/tracker?x=1");
    expect(screen.getByTestId("location").textContent).toBe("/app/habits?x=1");
  });

  it("redirects /app/shopping to /app/shopping/grocery", () => {
    renderAt("/app/shopping");
    expect(screen.getByTestId("location").textContent).toBe("/app/shopping/grocery");
  });

  it("routes an ingredient UUID to IngredientDetailPage", () => {
    renderAt("/app/ingredients/0b6c1f2e-3a4d-4e5f-8a9b-0c1d2e3f4a5b");
    expect(screen.getByTestId("page").textContent).toBe("IngredientDetailPage:");
  });

  it("routes /app/shopping/grocery/list-1 to ShoppingListDetailPage", () => {
    renderAt("/app/shopping/grocery/list-1");
    expect(screen.getByTestId("page").textContent).toBe("ShoppingListDetailPage:");
  });
});
