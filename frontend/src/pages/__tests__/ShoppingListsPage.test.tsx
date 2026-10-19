import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockDeleteMutate = vi.fn();

const LIST = {
  id: "list-1",
  attributes: { name: "Groceries", item_count: 3, checked_count: 1, archived_at: null },
};

vi.mock("@/lib/hooks/api/use-shopping", () => ({
  useShoppingLists: () => ({ data: { data: [LIST] }, isLoading: false }),
  useCreateShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteShoppingList: () => ({ mutate: mockDeleteMutate, isPending: false }),
  useUpdateShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useUnarchiveShoppingList: () => ({ mutate: vi.fn() }),
}));

import { ShoppingListsPage } from "../ShoppingListsPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <ShoppingListsPage />
    </MemoryRouter>,
  );
}

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
});
