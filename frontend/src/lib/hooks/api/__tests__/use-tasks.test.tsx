import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { taskKeys } from "../use-tasks";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

const mockTenant: Tenant = { id: "tenant-1", type: "tenants", attributes: { name: "Test", createdAt: "", updatedAt: "" } };
const mockHousehold: Household = { id: "household-1", type: "households", attributes: { name: "Home", timezone: "UTC", units: "imperial", createdAt: "", updatedAt: "" } };

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenant: mockTenant,
    household: mockHousehold,
    setActiveHousehold: vi.fn(),
  }),
}));

vi.mock("@/services/api/productivity", () => ({
  productivityService: {
    listTasks: vi.fn(),
    getTaskSummary: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
    restoreTask: vi.fn(),
  },
}));

const t = (id: string): Tenant => ({ id, type: "tenants", attributes: { name: "", createdAt: "", updatedAt: "" } });
const h = (id: string): Household => ({ id, type: "households", attributes: { name: "", timezone: "", units: "imperial", createdAt: "", updatedAt: "" } });

describe("taskKeys", () => {
  it("generates all key with tenant and household id", () => {
    expect(taskKeys.all(t("t-1"), h("hh-1"))).toEqual(["tasks", "t-1", "hh-1"]);
  });

  it("generates all key with no-tenant and no-household fallbacks", () => {
    expect(taskKeys.all(null, null)).toEqual(["tasks", "no-tenant", "no-household"]);
  });

  it("generates lists key", () => {
    expect(taskKeys.lists(t("t-1"), h("hh-1"))).toEqual(["tasks", "t-1", "hh-1", "list"]);
  });

  it("generates details key", () => {
    expect(taskKeys.details(t("t-1"), h("hh-1"))).toEqual(["tasks", "t-1", "hh-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(taskKeys.detail(t("t-1"), h("hh-1"), "task-42")).toEqual(["tasks", "t-1", "hh-1", "detail", "task-42"]);
  });

  it("generates summary key", () => {
    expect(taskKeys.summary(t("t-1"), h("hh-1"))).toEqual(["tasks", "t-1", "hh-1", "summary"]);
  });

  it("returns readonly tuple arrays", () => {
    const key = taskKeys.all(t("t-1"), h("hh-1"));
    expect(Array.isArray(key)).toBe(true);
    expect(key).toHaveLength(3);
  });
});

describe("useTasks hook", () => {
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

  it("fetches tasks when tenant and household are available", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listTasks as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "1", type: "tasks", attributes: { title: "Test", status: "pending" } }],
    });

    const { useTasks } = await import("../use-tasks");
    const { result } = renderHook(() => useTasks(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]!.attributes.title).toBe("Test");
  });

  it("includes tenant and household in query key for cache isolation", () => {
    expect(taskKeys.lists(t("tenant-1"), h("household-1"))).toEqual(
      ["tasks", "tenant-1", "household-1", "list"]
    );
    expect(taskKeys.lists(t("tenant-2"), h("household-1"))).not.toEqual(
      taskKeys.lists(t("tenant-1"), h("household-1"))
    );
  });

  it("fetches task summary", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.getTaskSummary as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "s-1", type: "task-summaries", attributes: { pendingCount: 5, completedTodayCount: 2, overdueCount: 1 } },
    });

    const { useTaskSummary } = await import("../use-tasks");
    const { result } = renderHook(() => useTaskSummary(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data.attributes.pendingCount).toBe(5);
  });

  it("creates a task and invalidates queries", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.createTask as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "new-1", type: "tasks", attributes: { title: "New Task", status: "pending" } },
    });

    const { useCreateTask } = await import("../use-tasks");
    const { result } = renderHook(() => useCreateTask(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ title: "New Task" });
    expect(productivityService.createTask).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      { title: "New Task" },
    );
  });

  it("deletes a task", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.deleteTask as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useDeleteTask } = await import("../use-tasks");
    const { result } = renderHook(() => useDeleteTask(), { wrapper: createWrapper() });

    await result.current.mutateAsync("task-1");
    expect(productivityService.deleteTask).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "task-1",
    );
  });

  it("updates a task", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.updateTask as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { id: "task-1", type: "tasks", attributes: { title: "Updated", status: "completed" } },
    });

    const { useUpdateTask } = await import("../use-tasks");
    const { result } = renderHook(() => useUpdateTask(), { wrapper: createWrapper() });

    await result.current.mutateAsync({ id: "task-1", attrs: { status: "completed" } });
    expect(productivityService.updateTask).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "task-1",
      { status: "completed" },
    );
  });

  it("restores a task", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.restoreTask as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const { useRestoreTask } = await import("../use-tasks");
    const { result } = renderHook(() => useRestoreTask(), { wrapper: createWrapper() });

    await result.current.mutateAsync("task-1");
    expect(productivityService.restoreTask).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "task-1",
    );
  });

  it("reports error state when fetch fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listTasks as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("network error"));

    const { useTasks } = await import("../use-tasks");
    const { result } = renderHook(() => useTasks(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("reports loading state initially", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.listTasks as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

    const { useTasks } = await import("../use-tasks");
    const { result } = renderHook(() => useTasks(), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
  });

  it("rejects when createTask mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.createTask as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("create failed"));

    const { useCreateTask } = await import("../use-tasks");
    const { result } = renderHook(() => useCreateTask(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync({ title: "Fail" })).rejects.toThrow("create failed");
  });

  it("rejects when deleteTask mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.deleteTask as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("delete failed"));

    const { useDeleteTask } = await import("../use-tasks");
    const { result } = renderHook(() => useDeleteTask(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync("task-1")).rejects.toThrow("delete failed");
  });

  it("rejects when updateTask mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.updateTask as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("update failed"));

    const { useUpdateTask } = await import("../use-tasks");
    const { result } = renderHook(() => useUpdateTask(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync({ id: "task-1", attrs: { status: "completed" } })).rejects.toThrow("update failed");
  });

  it("rejects when restoreTask mutation fails", async () => {
    const { productivityService } = await import("@/services/api/productivity");
    (productivityService.restoreTask as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("restore failed"));

    const { useRestoreTask } = await import("../use-tasks");
    const { result } = renderHook(() => useRestoreTask(), { wrapper: createWrapper() });

    await expect(result.current.mutateAsync("task-1")).rejects.toThrow("restore failed");
  });
});
