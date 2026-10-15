import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";

const download = vi.fn();
vi.mock("@/services/api/recipe", () => ({
  recipeService: { downloadImage: (...a: unknown[]) => download(...a) },
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenant: { id: "tenant-1" },
    household: { id: "household-1" },
  }),
}));

import { useRecipeImageSrc } from "../use-recipe-image-src";

const url = "/api/v1/recipes/r1/image?size=thumb&v=i1";

function setup(initial: string | null, enabled = true) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  const hook = renderHook(({ u }) => useRecipeImageSrc(u, { enabled }), { wrapper, initialProps: { u: initial } });
  return { client, ...hook };
}

describe("useRecipeImageSrc", () => {
  let n = 0;
  beforeEach(() => {
    n = 0;
    download.mockReset();
    URL.createObjectURL = vi.fn(() => `blob:${++n}`);
    URL.revokeObjectURL = vi.fn();
  });

  it("fetches through the api client (tenant headers) and returns an object URL", async () => {
    download.mockResolvedValue(new Blob(["x"]));
    const { result, client } = setup(url);
    await waitFor(() => expect(result.current.src).toBe("blob:1"));
    expect(download).toHaveBeenCalledWith(url);
    expect(client.getQueryCache().find({ queryKey: ["recipe-images", "tenant-1", "household-1", url] })).toBeDefined();
  });

  it("revokes the object URL on unmount", async () => {
    download.mockResolvedValue(new Blob(["x"]));
    const { result, unmount } = setup(url);
    await waitFor(() => expect(result.current.src).toBe("blob:1"));
    unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });

  it("does not fetch while disabled", () => {
    const { result } = setup(url, false);
    expect(download).not.toHaveBeenCalled();
    expect(result.current.src).toBeNull();
  });

  it("is not refetched by recipe invalidation", async () => {
    download.mockResolvedValue(new Blob(["x"]));
    const { result, client } = setup(url);
    await waitFor(() => expect(result.current.src).toBe("blob:1"));
    await client.invalidateQueries({ queryKey: ["recipes", "tenant-1", "household-1"] });
    expect(download).toHaveBeenCalledTimes(1);
  });

  it("does not fetch without a url", () => {
    const { result } = setup(null);
    expect(download).not.toHaveBeenCalled();
    expect(result.current.src).toBeNull();
  });

  it("reports an error when the fetch fails", async () => {
    download.mockRejectedValue(new Error("404"));
    const { result } = setup(url);
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.src).toBeNull();
  });
});
