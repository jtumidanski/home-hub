import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

const originalFetch = globalThis.fetch;
const originalLocation = window.location;

describe("401 Interceptor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    globalThis.fetch = vi.fn();
    // Mock window.location for redirect assertions
    Object.defineProperty(window, "location", {
      writable: true,
      value: { ...originalLocation, href: "http://localhost/app" },
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    Object.defineProperty(window, "location", {
      writable: true,
      value: originalLocation,
    });
  });

  async function getFreshApi() {
    // Clear module cache to get a fresh ApiClient instance
    vi.resetModules();
    const mod = await import("../client");
    return mod.api;
  }

  function mockResponse(status: number, body?: unknown): Response {
    return {
      ok: status >= 200 && status < 300,
      status,
      json: () => Promise.resolve(body ?? {}),
      text: () => Promise.resolve(JSON.stringify(body ?? {})),
      headers: new Headers({ "content-length": "1" }),
    } as unknown as Response;
  }

  it("refreshes token on 401 and retries the original request", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // First call: GET /test → 401
    // Second call: POST /auth/token/refresh → 200
    // Third call: GET /test retry → 200
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(200))
      .mockResolvedValueOnce(mockResponse(200, { data: { id: "1" } }));

    const result = await api.get("/test", { skipDeduplication: true });

    expect(result).toEqual({ data: { id: "1" } });
    expect(mockFetch).toHaveBeenCalledTimes(3);

    // Verify refresh was called
    const refreshCall = mockFetch.mock.calls[1]!;
    expect(refreshCall[0]).toBe("/api/v1/auth/token/refresh");
    expect(refreshCall[1].method).toBe("POST");
  });

  it("redirects to /login when refresh fails", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // GET /test → 401, then refresh → 401
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(401));

    await expect(
      api.get("/test", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth", status: 401 });

    expect(window.location.href).toBe("/login");
  });

  it("calls onAuthFailure callback when refresh fails", async () => {
    const api = await getFreshApi();
    const onAuthFailure = vi.fn();
    api.onAuthFailure = onAuthFailure;
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(401));

    await expect(
      api.get("/test", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth" });

    expect(onAuthFailure).toHaveBeenCalledTimes(1);
  });

  it("deduplicates concurrent refresh attempts", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // Two GET requests both get 401
    // One refresh attempt → 200
    // Both retry → 200
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))  // GET /a → 401
      .mockResolvedValueOnce(mockResponse(401))  // GET /b → 401
      .mockResolvedValueOnce(mockResponse(200))  // POST /auth/token/refresh → 200
      .mockResolvedValueOnce(mockResponse(200, { data: "a" }))  // GET /a retry
      .mockResolvedValueOnce(mockResponse(200, { data: "b" })); // GET /b retry

    const [resultA, resultB] = await Promise.all([
      api.get("/a", { skipDeduplication: true }),
      api.get("/b", { skipDeduplication: true }),
    ]);

    expect(resultA).toEqual({ data: "a" });
    expect(resultB).toEqual({ data: "b" });

    // Verify only one refresh was attempted
    const refreshCalls = mockFetch.mock.calls.filter(
      (call) => call[0] === "/api/v1/auth/token/refresh",
    );
    expect(refreshCalls).toHaveLength(1);
  });

  it("does not intercept 401 on refresh endpoint itself", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // POST /auth/token/refresh → 401 (should not trigger another refresh)
    mockFetch.mockResolvedValueOnce(mockResponse(401, {
      errors: [{ detail: "Refresh token expired" }],
    }));

    await expect(
      api.post("/auth/token/refresh"),
    ).rejects.toMatchObject({ type: "auth" });

    // Only one fetch call — no recursive refresh attempt
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });

  it("silently rejects after isRedirecting is set", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // First request triggers redirect
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(401)); // refresh fails

    await expect(
      api.get("/first", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth" });

    // Second request after redirect is in progress
    mockFetch.mockResolvedValueOnce(mockResponse(401));

    await expect(
      api.get("/second", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth" });

    // No additional refresh attempts after redirect
    const refreshCalls = mockFetch.mock.calls.filter(
      (call) => call[0] === "/api/v1/auth/token/refresh",
    );
    expect(refreshCalls).toHaveLength(1);
  });

  it("resetAuthState clears redirect flag", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    // Trigger redirect
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(401));

    await expect(
      api.get("/test", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth" });

    // Reset state
    api.resetAuthState();

    // Now a 401 should trigger a fresh refresh attempt
    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(200))
      .mockResolvedValueOnce(mockResponse(200, { data: "ok" }));

    const result = await api.get("/test", { skipDeduplication: true });
    expect(result).toEqual({ data: "ok" });
  });

  it("handles 401 on POST requests", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(200)) // refresh
      .mockResolvedValueOnce({
        ...mockResponse(200),
        status: 200,
        text: () => Promise.resolve(JSON.stringify({ data: { id: "new" } })),
        headers: new Headers({ "content-length": "42" }),
      } as unknown as Response);

    const result = await api.post("/tasks", { data: { type: "tasks" } });
    expect(result).toEqual({ data: { id: "new" } });
  });

  it("handles 401 on PATCH requests", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(200)) // refresh
      .mockResolvedValueOnce(mockResponse(200, { data: { id: "1", attributes: { status: "done" } } }));

    const result = await api.patch("/tasks/1", { data: { attributes: { status: "done" } } });
    expect(result).toEqual({ data: { id: "1", attributes: { status: "done" } } });
  });

  it("handles 401 on DELETE requests", async () => {
    const api = await getFreshApi();
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockResolvedValueOnce(mockResponse(200)) // refresh
      .mockResolvedValueOnce(mockResponse(200));

    await expect(api.delete("/tasks/1")).resolves.toBeUndefined();
  });

  it("handles network error during refresh", async () => {
    const api = await getFreshApi();
    const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    const mockFetch = globalThis.fetch as ReturnType<typeof vi.fn>;

    mockFetch
      .mockResolvedValueOnce(mockResponse(401))
      .mockRejectedValueOnce(new Error("network error")); // refresh fails

    await expect(
      api.get("/test", { skipDeduplication: true }),
    ).rejects.toMatchObject({ type: "auth" });

    expect(consoleSpy).toHaveBeenCalledWith("Token refresh failed due to network error");
    expect(window.location.href).toBe("/login");
    consoleSpy.mockRestore();
  });
});
