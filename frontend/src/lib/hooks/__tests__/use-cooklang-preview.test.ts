import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useCooklangPreview } from "@/lib/hooks/use-cooklang-preview";
import type { Tenant } from "@/types/models/tenant";

const mockTenant: Tenant = { id: "tenant-1", type: "tenants", attributes: { name: "Test", createdAt: "", updatedAt: "" } };

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ tenant: mockTenant }),
}));

vi.mock("@/services/api/recipe", () => ({
  recipeService: {
    parseSource: vi.fn(() =>
      new Promise(() => {
        // never resolves; these tests only care about the loading state
        // immediately after mount, before any debounced fetch settles.
      }),
    ),
  },
}));

describe("useCooklangPreview", () => {
  it("starts loading immediately when mounted with a non-empty source", () => {
    const { result } = renderHook(() => useCooklangPreview("ingredient flour"));

    expect(result.current.isLoading).toBe(true);
  });

  it("does not start loading when mounted with an empty source", () => {
    const { result } = renderHook(() => useCooklangPreview(""));

    expect(result.current.isLoading).toBe(false);
  });

  it("does not start loading when mounted with a whitespace-only source", async () => {
    const { result } = renderHook(() => useCooklangPreview("   "));

    expect(result.current.isLoading).toBe(false);
    await waitFor(() => expect(result.current.isLoading).toBe(false));
  });
});
