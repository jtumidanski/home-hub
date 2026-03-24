import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { authKeys } from "../use-auth";

vi.mock("@/services/api/auth", () => ({
  authService: {
    getMe: vi.fn(),
    getProviders: vi.fn(),
  },
}));

describe("authKeys", () => {
  it("generates all key", () => {
    expect(authKeys.all).toEqual(["auth"]);
  });

  it("generates me key", () => {
    expect(authKeys.me()).toEqual(["auth", "me"]);
  });

  it("generates providers key", () => {
    expect(authKeys.providers()).toEqual(["auth", "providers"]);
  });
});

describe("useMe hook", () => {
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

  it("fetches current user", async () => {
    const { authService } = await import("@/services/api/auth");
    (authService.getMe as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "u-1", type: "users", attributes: { displayName: "John", email: "john@test.com" } },
    });

    const { useMe } = await import("../use-auth");
    const { result } = renderHook(() => useMe(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.displayName).toBe("John");
  });

  it("does not retry on failure", async () => {
    const { authService } = await import("@/services/api/auth");
    (authService.getMe as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("unauthorized"));

    const { useMe } = await import("../use-auth");
    const { result } = renderHook(() => useMe(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(authService.getMe).toHaveBeenCalledTimes(1);
  });
});

describe("useProviders hook", () => {
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

  it("fetches auth providers", async () => {
    const { authService } = await import("@/services/api/auth");
    (authService.getProviders as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "google", type: "auth-providers", attributes: { displayName: "Google" } }],
    });

    const { useProviders } = await import("../use-auth");
    const { result } = renderHook(() => useProviders(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]!.attributes.displayName).toBe("Google");
  });

  it("reports error state when providers fetch fails", async () => {
    const { authService } = await import("@/services/api/auth");
    (authService.getProviders as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("providers error"));

    const { useProviders } = await import("../use-auth");
    const { result } = renderHook(() => useProviders(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("reports loading state initially", async () => {
    const { authService } = await import("@/services/api/auth");
    (authService.getProviders as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

    const { useProviders } = await import("../use-auth");
    const { result } = renderHook(() => useProviders(), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
  });
});
