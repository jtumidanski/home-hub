import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

const useRecipesMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipes: (params: unknown) => useRecipesMock(params),
  useRecipeTags: () => ({ data: { data: [] } }),
  useDeleteRecipe: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("react-router-dom", () => ({ useNavigate: () => vi.fn() }));
vi.mock("@/lib/hooks/use-mobile", () => ({ useMobile: () => false }));
vi.mock("@/components/common/pull-to-refresh", () => ({
  PullToRefresh: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

import { RecipesPage } from "../RecipesPage";

describe("RecipesPage sort control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useRecipesMock.mockReturnValue({ data: { data: [] }, isLoading: false, refetch: vi.fn() });
  });

  it("passes sort=usageCount after picking 'Least cooked'", async () => {
    const user = userEvent.setup();
    render(<RecipesPage />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /least cooked/i }));

    expect(useRecipesMock).toHaveBeenLastCalledWith(expect.objectContaining({ sort: "usageCount" }));
  });
});
