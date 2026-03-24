import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { taskKeys } from "../use-tasks";

// Mock useTenant to avoid context dependency in key factory tests
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenantId: "tenant-1",
    householdId: "household-1",
    tenant: null,
    household: null,
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

describe("taskKeys", () => {
  it("generates all key with household id", () => {
    expect(taskKeys.all("hh-1")).toEqual(["tasks", "hh-1"]);
  });

  it("generates all key with no-household fallback", () => {
    expect(taskKeys.all(null)).toEqual(["tasks", "no-household"]);
  });

  it("generates lists key", () => {
    expect(taskKeys.lists("hh-1")).toEqual(["tasks", "hh-1", "list"]);
  });

  it("generates list key (same as lists)", () => {
    expect(taskKeys.list("hh-1")).toEqual(["tasks", "hh-1", "list"]);
  });

  it("generates details key", () => {
    expect(taskKeys.details("hh-1")).toEqual(["tasks", "hh-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(taskKeys.detail("hh-1", "task-42")).toEqual(["tasks", "hh-1", "detail", "task-42"]);
  });

  it("generates summary key", () => {
    expect(taskKeys.summary("hh-1")).toEqual(["tasks", "hh-1", "summary"]);
  });

  it("returns readonly tuple arrays", () => {
    const key = taskKeys.all("hh-1");
    expect(Array.isArray(key)).toBe(true);
    expect(key).toHaveLength(2);
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
    expect(result.current.data?.data[0].attributes.title).toBe("Test");
  });
});
