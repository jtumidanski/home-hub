import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockCheckMutate = vi.fn();
const mockUncheckMutate = vi.fn();
const mockUseShoppingList = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useParams: () => ({ id: "list-1" }), useNavigate: () => vi.fn() };
});

vi.mock("@/lib/hooks/api/use-shopping", () => ({
  useShoppingList: () => mockUseShoppingList(),
  useAddShoppingItem: () => ({ mutate: vi.fn() }),
  useRemoveShoppingItem: () => ({ mutate: vi.fn() }),
  useCheckShoppingItem: () => ({ mutate: mockCheckMutate }),
  useUncheckAllItems: () => ({ mutate: mockUncheckMutate }),
  useArchiveShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useUnarchiveShoppingList: () => ({ mutate: vi.fn() }),
  useDeleteShoppingList: () => ({ mutate: vi.fn() }),
  useImportMealPlan: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("@/lib/hooks/api/use-meals", () => ({
  usePlans: () => ({ data: { data: [] } }),
}));

import { ShoppingListDetailPage, progressPercent } from "../ShoppingListDetailPage";

const ITEMS = [
  {
    id: "i1",
    name: "Milk",
    quantity: null,
    category_name: "Dairy",
    category_sort_order: 1,
    checked: false,
    position: 0,
  },
  {
    id: "i2",
    name: "Eggs",
    quantity: "12",
    category_name: "Dairy",
    category_sort_order: 1,
    checked: true,
    position: 1,
  },
];

function mockList(status: "active" | "archived" = "active") {
  return {
    data: {
      data: {
        id: "list-1",
        attributes: {
          name: "Groceries",
          status,
          archived_at: status === "archived" ? "2026-07-01T00:00:00Z" : null,
          items: ITEMS,
        },
      },
    },
    isLoading: false,
  };
}

function renderPage() {
  return render(
    <MemoryRouter>
      <ShoppingListDetailPage />
    </MemoryRouter>,
  );
}

describe("ShoppingListDetailPage — row behavior", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseShoppingList.mockReturnValue(mockList("active"));
  });

  it("toggles an item's checked state when a shopping-mode row is tapped", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    await userEvent.click(screen.getByText("Milk"));
    expect(mockCheckMutate).toHaveBeenCalledWith({ itemId: "i1", checked: true });
  });

  it("does not toggle items in the archived read-only view", async () => {
    mockUseShoppingList.mockReturnValue(mockList("archived"));
    renderPage();
    await userEvent.click(screen.getByText("Milk"));
    expect(mockCheckMutate).not.toHaveBeenCalled();
  });
});

describe("ShoppingListDetailPage — sticky mobile progress header", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseShoppingList.mockReturnValue(mockList("active"));
  });

  it("does not render the mobile progress header outside shopping mode", () => {
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-progress")).not.toBeInTheDocument();
  });

  it("renders the sticky progress header with count and Back to Edit in shopping mode", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    const header = screen.getByTestId("mobile-shopping-progress");
    expect(within(header).getByText("1 of 2 items")).toBeInTheDocument();
    expect(within(header).getByText("Back to Edit")).toBeInTheDocument();
  });

  it("does not render the mobile progress header in the archived view", () => {
    mockUseShoppingList.mockReturnValue(mockList("archived"));
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-progress")).not.toBeInTheDocument();
  });
});

describe("progressPercent", () => {
  it("returns 0 when total is 0 (no divide-by-zero)", () => {
    expect(progressPercent(0, 0)).toBe(0);
  });
  it("returns 0 when nothing is checked", () => {
    expect(progressPercent(0, 4)).toBe(0);
  });
  it("returns 50 at the halfway point", () => {
    expect(progressPercent(2, 4)).toBe(50);
  });
  it("returns 100 when everything is checked", () => {
    expect(progressPercent(4, 4)).toBe(100);
  });
  it("clamps to 100 when checked exceeds total", () => {
    expect(progressPercent(5, 4)).toBe(100);
  });
});
