import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { contextKeys } from "../use-context";

vi.mock("@/services/api/account", () => ({
  accountService: {
    getContext: vi.fn(),
  },
}));

describe("contextKeys", () => {
  it("generates all key", () => {
    expect(contextKeys.all).toEqual(["context"]);
  });

  it("generates current key", () => {
    expect(contextKeys.current).toEqual(["context", "current"]);
  });
});

describe("useAppContext hook", () => {
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

  it("fetches app context when enabled", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.getContext as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: {
        id: "current",
        type: "contexts",
        attributes: { resolvedTheme: "dark", resolvedRole: "owner", canCreateHousehold: true },
        relationships: {
          tenant: { data: { type: "tenants", id: "t-1" } },
          activeHousehold: { data: { type: "households", id: "hh-1" } },
        },
      },
    });

    const { useAppContext } = await import("../use-context");
    const { result } = renderHook(() => useAppContext(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.resolvedRole).toBe("owner");
  });

  it("does not fetch when disabled", async () => {
    const { accountService } = await import("@/services/api/account");
    const { useAppContext } = await import("../use-context");
    const { result } = renderHook(() => useAppContext(false), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe("idle");
    expect(accountService.getContext).not.toHaveBeenCalled();
  });

  it("does not retry on failure", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.getContext as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("unauthorized"));

    const { useAppContext } = await import("../use-context");
    const { result } = renderHook(() => useAppContext(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(accountService.getContext).toHaveBeenCalledTimes(1);
  });
});
