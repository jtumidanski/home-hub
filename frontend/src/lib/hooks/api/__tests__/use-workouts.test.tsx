import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import {
  useWorkoutNearestPopulatedWeek,
  workoutKeys,
} from "../use-workouts";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

const mockTenant: Tenant = { id: "tenant-1", type: "tenants", attributes: { name: "Test", createdAt: "", updatedAt: "" } };
const mockHousehold: Household = {
  id: "household-1",
  type: "households",
  attributes: { name: "Home", timezone: "UTC", units: "imperial", latitude: null, longitude: null, locationName: null, createdAt: "", updatedAt: "" },
};

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenant: mockTenant,
    household: mockHousehold,
    setActiveHousehold: vi.fn(),
  }),
}));

const getNearestPopulatedWeek = vi.fn();
vi.mock("@/services/api/workout", () => ({
  workoutService: {
    getNearestPopulatedWeek: (...args: unknown[]) => getNearestPopulatedWeek(...args),
  },
}));

function wrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
}

describe("workoutKeys.nearestPopulated", () => {
  it("includes reference and direction in the cache key", () => {
    expect(workoutKeys.nearestPopulated(mockTenant, mockHousehold, "2026-04-13", "prev")).toEqual([
      "workouts",
      "tenant-1",
      "household-1",
      "nearest",
      "2026-04-13",
      "prev",
    ]);
  });
});

describe("useWorkoutNearestPopulatedWeek", () => {
  beforeEach(() => {
    getNearestPopulatedWeek.mockReset();
  });

  it("returns the pointer document on success", async () => {
    getNearestPopulatedWeek.mockResolvedValue({
      data: { type: "workoutWeekPointer", id: "2026-04-06", attributes: { weekStartDate: "2026-04-06" } },
    });

    const { result } = renderHook(
      () => useWorkoutNearestPopulatedWeek("2026-04-13", "prev", true),
      { wrapper: wrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.weekStartDate).toBe("2026-04-06");
    expect(getNearestPopulatedWeek).toHaveBeenCalledWith(mockTenant, "2026-04-13", "prev");
  });

  it("normalizes errors (e.g. 404) to null", async () => {
    getNearestPopulatedWeek.mockRejectedValue(new Error("not found"));

    const { result } = renderHook(
      () => useWorkoutNearestPopulatedWeek("2026-04-13", "next", true),
      { wrapper: wrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
  });

  it("respects the enabled flag", () => {
    const { result } = renderHook(
      () => useWorkoutNearestPopulatedWeek("2026-04-13", "prev", false),
      { wrapper: wrapper() },
    );
    expect(result.current.fetchStatus).toBe("idle");
    expect(getNearestPopulatedWeek).not.toHaveBeenCalled();
  });
});
