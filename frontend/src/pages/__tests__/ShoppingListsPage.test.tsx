import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { LocationProbe } from "@/test/location-probe";

const mockDeleteMutate = vi.fn();

const LIST = {
  id: "list-1",
  attributes: { name: "Groceries", item_count: 3, checked_count: 1, archived_at: null },
};

const ARCHIVED = {
  id: "list-2",
  attributes: { name: "Old List", item_count: 2, checked_count: 2, archived_at: "2026-01-01T00:00:00Z" },
};

vi.mock("@/lib/hooks/api/use-shopping", () => ({
  useShoppingLists: (status: string) => ({
    data: { data: status === "archived" ? [ARCHIVED] : [LIST] },
    isLoading: false,
  }),
  useCreateShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteShoppingList: () => ({ mutate: mockDeleteMutate, isPending: false }),
  useUpdateShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useUnarchiveShoppingList: () => ({ mutate: vi.fn() }),
}));

import { ShoppingListsPage } from "../ShoppingListsPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <ShoppingListsPage tab="active" />
    </MemoryRouter>,
  );
}

function renderRoutes(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/app/shopping/grocery" element={<ShoppingListsPage tab="active" />} />
        <Route path="/app/shopping/grocery/archived" element={<ShoppingListsPage tab="archived" />} />
      </Routes>
      <LocationProbe />
    </MemoryRouter>,
  );
}

describe("ShoppingListsPage — tabs", () => {
  it("marks Active pressed at the base route", () => {
    renderRoutes("/app/shopping/grocery");
    expect(screen.getByRole("button", { name: /^Active \(/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: /^History \(/ })).toHaveAttribute("aria-pressed", "false");
  });

  it("pushes the archived URL on History click and returns on back", async () => {
    renderRoutes("/app/shopping/grocery");
    await userEvent.click(screen.getByRole("button", { name: /^History \(/ }));
    expect(screen.getByTestId("location").textContent).toBe("/app/shopping/grocery/archived");
    expect(screen.getByTestId("nav-type").textContent).toBe("PUSH");
    expect(screen.getByRole("button", { name: /^History \(/ })).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(screen.getByText("probe-back"));
    expect(screen.getByTestId("location").textContent).toBe("/app/shopping/grocery");
  });

  it("renders the archived list when starting at /archived", () => {
    renderRoutes("/app/shopping/grocery/archived");
    expect(screen.getByRole("link", { name: "Old List" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Groceries" })).not.toBeInTheDocument();
  });
});

describe("ShoppingListsPage — delete confirmation", () => {
  beforeEach(() => vi.clearAllMocks());

  it("links each list by name for keyboard navigation", () => {
    renderPage();
    expect(screen.getByRole("link", { name: "Groceries" })).toHaveAttribute(
      "href",
      "/app/shopping/grocery/list-1",
    );
  });

  it("asks for confirmation before deleting", async () => {
    renderPage();
    await userEvent.click(screen.getByRole("button", { name: "Delete Groceries" }));
    expect(mockDeleteMutate).not.toHaveBeenCalled();
    expect(screen.getByText("Delete shopping list?")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(mockDeleteMutate).toHaveBeenCalledWith("list-1", expect.any(Object));
  });

  it("does not delete when the confirmation is cancelled", async () => {
    renderPage();
    await userEvent.click(screen.getByRole("button", { name: "Delete Groceries" }));
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(mockDeleteMutate).not.toHaveBeenCalled();
  });

  it("keeps the dialog open with Delete available when the delete fails", async () => {
    renderPage();
    await userEvent.click(screen.getByRole("button", { name: "Delete Groceries" }));
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    // onSuccess is never invoked on failure (the hook toasts the error)
    expect(screen.getByText("Delete shopping list?")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete" })).toBeEnabled();
  });
});
