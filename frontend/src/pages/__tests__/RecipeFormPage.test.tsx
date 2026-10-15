import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const navigate = vi.fn();
let params: { id?: string } = {};
vi.mock("react-router-dom", () => ({
  useParams: () => params,
  useNavigate: () => navigate,
}));

const toastError = vi.fn();
const toastSuccess = vi.fn();
vi.mock("sonner", () => ({ toast: { error: (...a: unknown[]) => toastError(...a), success: (...a: unknown[]) => toastSuccess(...a) } }));

const createMutateAsync = vi.fn();
const updateMutateAsync = vi.fn();
const uploadMutateAsync = vi.fn();
const deleteImageMutateAsync = vi.fn();
let uploadPending = false;
let existing: unknown = undefined;
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipe: () => ({ data: existing }),
  useCreateRecipe: () => ({ mutateAsync: createMutateAsync }),
  useUpdateRecipe: () => ({ mutateAsync: updateMutateAsync }),
  useUploadRecipeImage: () => ({ mutateAsync: uploadMutateAsync, isPending: uploadPending }),
  useDeleteRecipeImage: () => ({ mutateAsync: deleteImageMutateAsync, isPending: false }),
}));
vi.mock("@/lib/hooks/use-recipe-image-src", () => ({
  useRecipeImageSrc: (url: string | null) => ({ src: url ? `blob:${url}` : null, isError: false }),
}));
vi.mock("@/lib/hooks/use-cooklang-preview", () => ({
  useCooklangPreview: () => ({
    metadata: { title: "Soup" },
    ingredients: [],
    steps: [],
    errors: [],
    notes: [],
    normalization: [],
    isLoading: false,
  }),
}));
vi.mock("@/lib/hooks/use-mobile", () => ({ useMobile: () => false }));
vi.mock("@/components/features/recipes/cooklang-preview", () => ({ CooklangPreview: () => null }));
vi.mock("@/components/features/recipes/cooklang-help", () => ({ CooklangHelp: () => null }));
vi.mock("@/components/features/recipes/planner-config-form", () => ({ PlannerConfigForm: () => null }));
vi.mock("@/components/features/recipes/recipe-ingredients", () => ({ RecipeIngredients: () => null }));

import { RecipeFormPage } from "../RecipeFormPage";

const photo = () => new File(["x"], "a.jpg", { type: "image/jpeg" });
const existingImage = { thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 10, height: 10 };

function existingRecipe(image: unknown) {
  return {
    data: {
      id: "r1",
      type: "recipes",
      attributes: { title: "Soup", source: "---\ntitle: Soup\n---\nBoil.", tags: [], ingredients: [], image },
    },
  };
}

async function typeSource(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText("Recipe (Cooklang)"), "Boil water.");
}

describe("RecipeFormPage image save sequencing", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    params = {};
    existing = undefined;
    uploadPending = false;
    URL.createObjectURL = vi.fn(() => "blob:preview");
    URL.revokeObjectURL = vi.fn();
  });

  it("create + upload, then navigates", async () => {
    const user = userEvent.setup();
    createMutateAsync.mockResolvedValue({ data: { id: "r1" } });
    uploadMutateAsync.mockResolvedValue({});
    render(<RecipeFormPage />);
    await typeSource(user);
    const file = photo();
    await user.upload(screen.getByLabelText("Choose photo"), file);
    await user.click(screen.getByRole("button", { name: /create recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(uploadMutateAsync).toHaveBeenCalledWith({ id: "r1", file });
  });

  it("edit + remove deletes the image after the PATCH", async () => {
    const user = userEvent.setup();
    params = { id: "r1" };
    existing = existingRecipe(existingImage);
    updateMutateAsync.mockResolvedValue({});
    deleteImageMutateAsync.mockResolvedValue(undefined);
    render(<RecipeFormPage />);
    await user.click(screen.getByRole("button", { name: /^remove$/i }));
    await user.click(screen.getByRole("button", { name: /update recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(updateMutateAsync).toHaveBeenCalled();
    expect(deleteImageMutateAsync).toHaveBeenCalledWith("r1");
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("remove then undo makes no image request", async () => {
    const user = userEvent.setup();
    params = { id: "r1" };
    existing = existingRecipe(existingImage);
    updateMutateAsync.mockResolvedValue({});
    render(<RecipeFormPage />);
    await user.click(screen.getByRole("button", { name: /^remove$/i }));
    await user.click(screen.getByRole("button", { name: /undo/i }));
    await user.click(screen.getByRole("button", { name: /update recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(deleteImageMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("cancel makes no request", async () => {
    const user = userEvent.setup();
    render(<RecipeFormPage />);
    await user.upload(screen.getByLabelText("Choose photo"), photo());
    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(navigate).toHaveBeenCalledWith("/app/recipes");
    expect(createMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("partial failure keeps the user on the form and retries only the image", async () => {
    const user = userEvent.setup();
    createMutateAsync.mockResolvedValue({ data: { id: "r1" } });
    uploadMutateAsync.mockRejectedValueOnce(new Error("503")).mockResolvedValueOnce({});
    render(<RecipeFormPage />);
    await typeSource(user);
    const file = photo();
    await user.upload(screen.getByLabelText("Choose photo"), file);
    await user.click(screen.getByRole("button", { name: /create recipe/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Recipe saved, but the photo failed to upload — try again."),
    );
    expect(navigate).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: /update recipe/i }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(createMutateAsync).toHaveBeenCalledTimes(1);
    expect(updateMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).toHaveBeenCalledTimes(2);
    expect(uploadMutateAsync).toHaveBeenLastCalledWith({ id: "r1", file });
  });

  it("shows the uploading label while the image request is pending", () => {
    uploadPending = true;
    render(<RecipeFormPage />);
    const button = screen.getByRole("button", { name: "Uploading photo…" });
    expect(button).toBeDisabled();
  });
});
