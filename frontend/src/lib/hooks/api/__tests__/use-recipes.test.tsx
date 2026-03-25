import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { recipeKeys } from "../use-recipes";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

const mockTenant: Tenant = { id: "tenant-1", type: "tenants", attributes: { name: "Test", createdAt: "", updatedAt: "" } };
const mockHousehold: Household = { id: "household-1", type: "households", attributes: { name: "Home", timezone: "UTC", units: "imperial", latitude: null, longitude: null, locationName: null, createdAt: "", updatedAt: "" } };

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenant: mockTenant,
    household: mockHousehold,
    setActiveHousehold: vi.fn(),
  }),
}));

vi.mock("@/services/api/recipe", () => ({
  recipeService: {
    listRecipes: vi.fn(),
    getRecipe: vi.fn(),
    createRecipe: vi.fn(),
    updateRecipe: vi.fn(),
    deleteRecipe: vi.fn(),
    restoreRecipe: vi.fn(),
    listTags: vi.fn(),
    parseSource: vi.fn(),
  },
}));

const t = (id: string): Tenant => ({ id, type: "tenants", attributes: { name: "", createdAt: "", updatedAt: "" } });
const h = (id: string): Household => ({ id, type: "households", attributes: { name: "", timezone: "", units: "imperial", latitude: null, longitude: null, locationName: null, createdAt: "", updatedAt: "" } });

describe("recipeKeys", () => {
  it("generates all key with tenant and household id", () => {
    expect(recipeKeys.all(t("t-1"), h("hh-1"))).toEqual(["recipes", "t-1", "hh-1"]);
  });

  it("generates all key with no-tenant and no-household fallbacks", () => {
    expect(recipeKeys.all(null, null)).toEqual(["recipes", "no-tenant", "no-household"]);
  });

  it("generates lists key", () => {
    expect(recipeKeys.lists(t("t-1"), h("hh-1"))).toEqual(["recipes", "t-1", "hh-1", "list"]);
  });

  it("generates details key", () => {
    expect(recipeKeys.details(t("t-1"), h("hh-1"))).toEqual(["recipes", "t-1", "hh-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(recipeKeys.detail(t("t-1"), h("hh-1"), "recipe-42")).toEqual(["recipes", "t-1", "hh-1", "detail", "recipe-42"]);
  });

  it("generates tags key", () => {
    expect(recipeKeys.tags(t("t-1"), h("hh-1"))).toEqual(["recipes", "t-1", "hh-1", "tags"]);
  });

  it("returns readonly tuple arrays", () => {
    const key = recipeKeys.all(t("t-1"), h("hh-1"));
    expect(Array.isArray(key)).toBe(true);
    expect(key).toHaveLength(3);
  });

  it("different tenants produce different keys", () => {
    expect(recipeKeys.lists(t("tenant-1"), h("household-1"))).not.toEqual(
      recipeKeys.lists(t("tenant-2"), h("household-1"))
    );
  });
});

describe("useRecipes hook", () => {
  let queryClient: QueryClient;

  function createWrapper() {
    return ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  }

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
  });

  it("fetches recipes when tenant and household are available", async () => {
    const { recipeService } = await import("@/services/api/recipe");
    (recipeService.listRecipes as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "1", type: "recipes", attributes: { title: "Carbonara", tags: ["italian"] } }],
      meta: { total: 1, page: 1, pageSize: 20 },
    });

    const { useRecipes } = await import("../use-recipes");
    const { result } = renderHook(() => useRecipes(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]!.attributes.title).toBe("Carbonara");
  });

  it("fetches a single recipe", async () => {
    const { recipeService } = await import("@/services/api/recipe");
    (recipeService.getRecipe as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "1", type: "recipes", attributes: { title: "Carbonara", source: "Add @eggs{3}.", ingredients: [], steps: [], tags: [] } },
    });

    const { useRecipe } = await import("../use-recipes");
    const { result } = renderHook(() => useRecipe("1"), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.title).toBe("Carbonara");
  });

  it("creates a recipe and invalidates queries", async () => {
    const { recipeService } = await import("@/services/api/recipe");
    (recipeService.createRecipe as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "new-1", type: "recipes", attributes: { title: "New Recipe", source: "Stir.", ingredients: [], steps: [], tags: [] } },
    });

    const { useCreateRecipe } = await import("../use-recipes");
    const { result } = renderHook(() => useCreateRecipe(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ title: "New Recipe", source: "Stir." });
    expect(recipeService.createRecipe).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      { title: "New Recipe", source: "Stir." },
    );
  });

  it("deletes a recipe with optimistic update", async () => {
    const { recipeService } = await import("@/services/api/recipe");
    (recipeService.deleteRecipe as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useDeleteRecipe } = await import("../use-recipes");
    const { result } = renderHook(() => useDeleteRecipe(), { wrapper: createWrapper() });

    await result.current.mutateAsync("recipe-1");
    expect(recipeService.deleteRecipe).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "recipe-1",
    );
  });

  it("fetches recipe tags", async () => {
    const { recipeService } = await import("@/services/api/recipe");
    (recipeService.listTags as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [
        { id: "italian", type: "recipe-tags", attributes: { tag: "italian", count: 5 } },
        { id: "quick", type: "recipe-tags", attributes: { tag: "quick", count: 3 } },
      ],
    });

    const { useRecipeTags } = await import("../use-recipes");
    const { result } = renderHook(() => useRecipeTags(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(2);
  });
});
