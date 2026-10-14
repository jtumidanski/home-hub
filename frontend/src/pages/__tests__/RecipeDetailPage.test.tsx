import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { RecipeDetail } from "@/types/models/recipe";

const useRecipeMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipe: () => useRecipeMock(),
  useDeleteRecipe: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@/lib/hooks/api/use-ingredient-normalization", () => ({
  useRenormalize: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock("react-router-dom", () => ({
  useParams: () => ({ id: "r1" }),
  useNavigate: () => vi.fn(),
}));
vi.mock("@/components/features/recipes/recipe-ingredients", () => ({ RecipeIngredients: () => null }));
vi.mock("@/components/features/recipes/recipe-steps", () => ({ RecipeSteps: () => null }));
vi.mock("@/components/features/recipes/cook-mode", () => ({ CookMode: () => null }));

import { RecipeDetailPage } from "../RecipeDetailPage";

function detail(image: RecipeDetail["attributes"]["image"]): { data: RecipeDetail } {
  return {
    data: {
      id: "r1",
      type: "recipes",
      attributes: {
        title: "Soup",
        tags: [],
        plannerReady: false,
        resolvedIngredients: 0,
        totalIngredients: 0,
        createdAt: "",
        updatedAt: "",
        image,
        source: "",
        ingredients: [],
        steps: [],
        plannerIssues: [],
      },
    },
  };
}

describe("RecipeDetailPage hero", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders no hero when the recipe has no image", () => {
    useRecipeMock.mockReturnValue({ data: detail(null), isLoading: false });
    render(<RecipeDetailPage />);
    expect(screen.queryByRole("img", { name: "Soup" })).not.toBeInTheDocument();
    expect(screen.queryByTestId("recipe-image-placeholder")).not.toBeInTheDocument();
  });

  it("renders the hero and opens an uncropped lightbox on click", async () => {
    const user = userEvent.setup();
    useRecipeMock.mockReturnValue({
      data: detail({ thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 1600, height: 900 }),
      isLoading: false,
    });
    render(<RecipeDetailPage />);
    const hero = screen.getByRole("img", { name: "Soup" });
    expect(hero).toHaveAttribute("src", "/f.jpg");

    await user.click(screen.getByRole("button", { name: /view photo full size/i }));
    // The modal aria-hides the hero, so scope to the dialog.
    const dialog = await screen.findByRole("dialog");
    const lightboxImg = within(dialog).getByRole("img", { name: "Soup" });
    expect(lightboxImg).toHaveAttribute("src", "/f.jpg");
    expect(lightboxImg).toHaveClass("object-contain");
  });
});
