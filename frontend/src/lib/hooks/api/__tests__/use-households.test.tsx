import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { householdKeys } from "../use-households";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenantId: "tenant-1",
    householdId: "household-1",
    tenant: null,
    household: null,
    setActiveHousehold: vi.fn(),
  }),
}));

vi.mock("@/services/api/account", () => ({
  accountService: {
    listHouseholds: vi.fn(),
    createHousehold: vi.fn(),
  },
}));

describe("householdKeys", () => {
  it("generates all key with tenant id", () => {
    expect(householdKeys.all("t-1")).toEqual(["households", "t-1"]);
  });

  it("generates all key with no-tenant fallback", () => {
    expect(householdKeys.all(null)).toEqual(["households", "no-tenant"]);
  });

  it("generates lists key", () => {
    expect(householdKeys.lists("t-1")).toEqual(["households", "t-1", "list"]);
  });

  it("generates details key", () => {
    expect(householdKeys.details("t-1")).toEqual(["households", "t-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(householdKeys.detail("t-1", "hh-42")).toEqual(["households", "t-1", "detail", "hh-42"]);
  });
});

describe("useHouseholds hook", () => {
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

  it("fetches households when tenant is available", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.listHouseholds as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "hh-1", type: "households", attributes: { name: "Main Home", timezone: "America/Chicago", units: "imperial" } }],
    });

    const { useHouseholds } = await import("../use-households");
    const { result } = renderHook(() => useHouseholds(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]!.attributes.name).toBe("Main Home");
  });

  it("does not fetch when explicitly disabled", async () => {
    const { accountService } = await import("@/services/api/account");
    const { useHouseholds } = await import("../use-households");
    const { result } = renderHook(() => useHouseholds(false), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe("idle");
    expect(accountService.listHouseholds).not.toHaveBeenCalled();
  });

  it("creates a household", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.createHousehold as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "hh-new", type: "households", attributes: { name: "Beach House", timezone: "America/New_York", units: "metric" } },
    });

    const { useCreateHousehold } = await import("../use-households");
    const { result } = renderHook(() => useCreateHousehold(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ name: "Beach House", timezone: "America/New_York", units: "metric" });
    expect(accountService.createHousehold).toHaveBeenCalledWith("tenant-1", "Beach House", "America/New_York", "metric");
  });

  it("reports error state when fetch fails", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.listHouseholds as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("network error"));

    const { useHouseholds } = await import("../use-households");
    const { result } = renderHook(() => useHouseholds(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("reports loading state initially", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.listHouseholds as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

    const { useHouseholds } = await import("../use-households");
    const { result } = renderHook(() => useHouseholds(), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
  });

  it("rejects when createHousehold mutation fails", async () => {
    const { accountService } = await import("@/services/api/account");
    (accountService.createHousehold as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("create failed"));

    const { useCreateHousehold } = await import("../use-households");
    const { result } = renderHook(() => useCreateHousehold(), { wrapper: createWrapper() });

    await expect(
      result.current.mutateAsync({ name: "Fail", timezone: "UTC", units: "metric" })
    ).rejects.toThrow("create failed");
  });
});
