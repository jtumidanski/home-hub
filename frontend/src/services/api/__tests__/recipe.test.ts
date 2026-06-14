import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api/client", () => ({
  api: {
    setTenant: vi.fn(),
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
