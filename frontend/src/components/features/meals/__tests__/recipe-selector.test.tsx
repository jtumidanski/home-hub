import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const useRecipesMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipes: (params: unknown) => useRecipesMock(params),
}));

import { RecipeSelector } from "../recipe-selector";

describe("RecipeSelector sort control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useRecipesMock.mockReturnValue({ data: { data: [] }, isLoading: false });
  });

  it("defaults to no sort param", () => {
    render(<RecipeSelector onSelectRecipe={vi.fn()} />);
    expect(useRecipesMock).toHaveBeenCalledWith(expect.objectContaining({ sort: undefined }));
  });

  it("passes sort=-usageCount after picking 'Most cooked'", async () => {
    const user = userEvent.setup();
    render(<RecipeSelector onSelectRecipe={vi.fn()} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /most cooked/i }));

    expect(useRecipesMock).toHaveBeenLastCalledWith(expect.objectContaining({ sort: "-usageCount" }));
  });
});
