import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api/client", () => ({
  api: {
    setTenant: vi.fn(),
    upload: vi.fn().mockResolvedValue({ data: { id: "img-1", type: "recipe-images", attributes: {} } }),
    delete: vi.fn().mockResolvedValue(undefined),
    get: vi.fn().mockResolvedValue({ data: [], meta: { total: 0, page: 1, pageSize: 20 } }),
  },
}));

const tenant = { id: "tenant-1", type: "tenants" as const, attributes: { name: "T", createdAt: "", updatedAt: "" } };

describe("recipeService.listRecipes", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("sends sort=-usageCount when sort is descending", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, { sort: "-usageCount" });
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).toContain("sort=-usageCount");
  });

  it("sends sort=usageCount when sort is ascending", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, { sort: "usageCount" });
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).toContain("sort=usageCount");
  });

  it("omits sort when not provided", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, {});
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).not.toContain("sort=");
  });
});

describe("recipeService image methods", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uploadRecipeImage PUTs multipart FormData with the file part", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    const file = new File(["x"], "a.jpg", { type: "image/jpeg" });
    await recipeService.uploadRecipeImage(tenant, "r1", file);
    const [path, form, options] = (api.upload as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(path).toBe("/recipes/r1/image");
    expect(form).toBeInstanceOf(FormData);
    expect((form as FormData).get("file")).toBe(file);
    expect(options).toEqual({ method: "PUT" });
  });

  it("deleteRecipeImage DELETEs the image path", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.deleteRecipeImage(tenant, "r1");
    expect(api.delete).toHaveBeenCalledWith("/recipes/r1/image");
  });
});
