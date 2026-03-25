import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { reminderKeys } from "../use-reminders";
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

vi.mock("@/services/api/productivity", () => ({
  productivityService: {
    listReminders: vi.fn(),
    getReminderSummary: vi.fn(),
    createReminder: vi.fn(),
    updateReminder: vi.fn(),
    deleteReminder: vi.fn(),
    snoozeReminder: vi.fn(),
    dismissReminder: vi.fn(),
  },
}));

const t = (id: string): Tenant => ({ id, type: "tenants", attributes: { name: "", createdAt: "", updatedAt: "" } });
const h = (id: string): Household => ({ id, type: "households", attributes: { name: "", timezone: "", units: "imperial", latitude: null, longitude: null, locationName: null, createdAt: "", updatedAt: "" } });

describe("reminderKeys", () => {
  it("generates all key with tenant and household id", () => {
    expect(reminderKeys.all(t("t-1"), h("hh-1"))).toEqual(["reminders", "t-1", "hh-1"]);
  });

  it("generates all key with no-tenant and no-household fallbacks", () => {
    expect(reminderKeys.all(null, null)).toEqual(["reminders", "no-tenant", "no-household"]);
  });

  it("generates lists key", () => {
    expect(reminderKeys.lists(t("t-1"), h("hh-1"))).toEqual(["reminders", "t-1", "hh-1", "list"]);
  });

  it("generates details key", () => {
    expect(reminderKeys.details(t("t-1"), h("hh-1"))).toEqual(["reminders", "t-1", "hh-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(reminderKeys.detail(t("t-1"), h("hh-1"), "rem-42")).toEqual(["reminders", "t-1", "hh-1", "detail", "rem-42"]);
  });

  it("generates summary key", () => {
    expect(reminderKeys.summary(t("t-1"), h("hh-1"))).toEqual(["reminders", "t-1", "hh-1", "summary"]);
  });
});

describe("useReminders hook", () => {
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

  it("fetches reminders when tenant and household are available", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listReminders as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "1", type: "reminders", attributes: { title: "Test Reminder", active: true, scheduledFor: "2026-03-25T09:00:00Z" } }],
    });

    const { useReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useReminders(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]!.attributes.title).toBe("Test Reminder");
  });

  it("fetches reminder summary", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.getReminderSummary as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "s-1", type: "reminder-summaries", attributes: { dueNowCount: 3, upcomingCount: 5, snoozedCount: 1 } },
    });

    const { useReminderSummary } = await import("../use-reminders");
    const { result } = renderHook(() => useReminderSummary(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.dueNowCount).toBe(3);
  });

  it("creates a reminder", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.createReminder as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "new-1", type: "reminders", attributes: { title: "New Reminder", active: true } },
    });

    const { useCreateReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useCreateReminder(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ title: "New Reminder", scheduledFor: "2026-03-25T09:00:00Z" });
    expect(productivityService.createReminder).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      { title: "New Reminder", scheduledFor: "2026-03-25T09:00:00Z" },
    );
  });

  it("deletes a reminder", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.deleteReminder as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useDeleteReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useDeleteReminder(), { wrapper: createWrapper() });

    await result.current.mutateAsync("rem-1");
    expect(productivityService.deleteReminder).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "rem-1",
    );
  });

  it("snoozes a reminder", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.snoozeReminder as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useSnoozeReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useSnoozeReminder(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ id: "rem-1", minutes: 10 });
    expect(productivityService.snoozeReminder).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "rem-1",
      10,
    );
  });

  it("dismisses a reminder", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.dismissReminder as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useDismissReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useDismissReminder(), { wrapper: createWrapper() });

    await result.current.mutateAsync("rem-1");
    expect(productivityService.dismissReminder).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "rem-1",
    );
  });

  it("reports error state when fetch fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listReminders as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("network error"));

    const { useReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useReminders(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("reports loading state initially", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listReminders as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

    const { useReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useReminders(), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
  });

  it("rejects when createReminder mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.createReminder as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("create failed"));

    const { useCreateReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useCreateReminder(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync({ title: "Fail", scheduledFor: "2026-03-25T09:00:00Z" })).rejects.toThrow("create failed");
  });

  it("rejects when deleteReminder mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.deleteReminder as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("delete failed"));

    const { useDeleteReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useDeleteReminder(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync("rem-1")).rejects.toThrow("delete failed");
  });

  it("rejects when snoozeReminder mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.snoozeReminder as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("snooze failed"));

    const { useSnoozeReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useSnoozeReminder(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync({ id: "rem-1", minutes: 10 })).rejects.toThrow("snooze failed");
  });

  it("rejects when dismissReminder mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.dismissReminder as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("dismiss failed"));

    const { useDismissReminder } = await import("../use-reminders");
    const { result } = renderHook(() => useDismissReminder(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync("rem-1")).rejects.toThrow("dismiss failed");
  });
});

describe("useInvalidateReminders", () => {
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

  it("invalidateAll calls invalidateQueries with reminders key prefix", async () => {
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { useInvalidateReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useInvalidateReminders(), { wrapper: createWrapper() });

    await result.current.invalidateAll();
    expect(spy).toHaveBeenCalledWith({ queryKey: ["reminders", "tenant-1", "household-1"] });
  });

  it("invalidateLists calls invalidateQueries with lists key", async () => {
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { useInvalidateReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useInvalidateReminders(), { wrapper: createWrapper() });

    await result.current.invalidateLists();
    expect(spy).toHaveBeenCalledWith({ queryKey: ["reminders", "tenant-1", "household-1", "list"] });
  });

  it("invalidateSummary calls invalidateQueries with summary key", async () => {
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { useInvalidateReminders } = await import("../use-reminders");
    const { result } = renderHook(() => useInvalidateReminders(), { wrapper: createWrapper() });

    await result.current.invalidateSummary();
    expect(spy).toHaveBeenCalledWith({ queryKey: ["reminders", "tenant-1", "household-1", "summary"] });
  });
});

describe("usePrefetchReminders", () => {
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

  it("prefetch calls prefetchQuery with reminders lists key", async () => {
    const spy = vi.spyOn(queryClient, "prefetchQuery").mockResolvedValue(undefined);
    const { usePrefetchReminders } = await import("../use-reminders");
    const { result } = renderHook(() => usePrefetchReminders(), { wrapper: createWrapper() });

    result.current.prefetch();
    expect(spy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["reminders", "tenant-1", "household-1", "list"] }),
    );
  });

  it("prefetchSummary calls prefetchQuery with summary key", async () => {
    const spy = vi.spyOn(queryClient, "prefetchQuery").mockResolvedValue(undefined);
    const { usePrefetchReminders } = await import("../use-reminders");
    const { result } = renderHook(() => usePrefetchReminders(), { wrapper: createWrapper() });

    result.current.prefetchSummary();
    expect(spy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ["reminders", "tenant-1", "household-1", "summary"] }),
    );
  });
});
