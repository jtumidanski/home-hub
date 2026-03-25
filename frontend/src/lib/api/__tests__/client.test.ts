import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ApiRequestError } from "../client";

// We test ApiRequestError directly and the ApiClient via fetch mocking
const originalFetch = globalThis.fetch;

describe("ApiRequestError", () => {
  it("creates error with message and status", () => {
    const error = new ApiRequestError("Not Found", 404);
    expect(error.message).toBe("Not Found");
    expect(error.status).toBe(404);
    expect(error.name).toBe("ApiRequestError");
  });

  it("is an instance of Error", () => {
    const error = new ApiRequestError("Server Error", 500);
    expect(error).toBeInstanceOf(Error);
    expect(error).toBeInstanceOf(ApiRequestError);
  });
});

describe("ApiClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  async function getApi() {
    // Re-import to get fresh instance
    const mod = await import("../client");
    return mod.api;
  }

  it("sets tenant header on requests", async () => {
    const api = await getApi();
    api.setTenant({ id: "tenant-123" });

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });

    await api.get("/test");

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].headers["X-Tenant-ID"]).toBe("tenant-123");
  });

  it("clears tenant header after clearTenant", async () => {
    const api = await getApi();
    api.setTenant({ id: "tenant-123" });
    api.clearTenant();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });

    await api.get("/test");

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].headers["X-Tenant-ID"]).toBeUndefined();
  });

  it("sets JSON:API accept header", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });

    await api.get("/test");

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].headers.Accept).toBe("application/vnd.api+json");
  });

  it("throws ApiRequestError on non-ok response", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ errors: [{ detail: "Resource not found" }] }),
    });

    await expect(api.get("/missing")).rejects.toMatchObject({
      message: "Resource not found",
      status: 404,
      type: "not-found",
    });
    await expect(api.get("/missing", { skipDeduplication: true })).rejects.toMatchObject({
      message: "Resource not found",
      status: 404,
    });
  });

  it("throws ApiRequestError with status text when JSON parse fails", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 502,
      json: () => Promise.reject(new Error("not json")),
    });

    await expect(api.get("/bad-gateway", { skipDeduplication: true, maxRetries: 0 })).rejects.toMatchObject({
      message: "Request failed with status 502",
      status: 502,
    });
  });

  it("sends POST with JSON:API content type", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers({ "content-length": "42" }),
      text: () => Promise.resolve(JSON.stringify({ data: { id: "1" } })),
    });

    await api.post("/tasks", { data: { type: "tasks", attributes: { title: "Test" } } });

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].method).toBe("POST");
    expect(fetchCall[1].headers["Content-Type"]).toBe("application/vnd.api+json");
    expect(JSON.parse(fetchCall[1].body)).toEqual({
      data: { type: "tasks", attributes: { title: "Test" } },
    });
  });

  it("sends PATCH with JSON:API content type", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: { id: "1" } }),
    });

    await api.patch("/tasks/1", { data: { type: "tasks", id: "1", attributes: { status: "completed" } } });

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].method).toBe("PATCH");
  });

  it("sends DELETE without body", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
    });

    await api.delete("/tasks/1");

    const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(fetchCall[1].method).toBe("DELETE");
  });

  it("deduplicates concurrent GET requests", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });

    const [r1, r2] = await Promise.all([api.get("/tasks"), api.get("/tasks")]);

    expect(globalThis.fetch).toHaveBeenCalledTimes(1);
    expect(r1).toEqual(r2);
  });

  it("skips deduplication when requested", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: [] }),
    });

    await Promise.all([
      api.get("/tasks", { skipDeduplication: true }),
      api.get("/tasks", { skipDeduplication: true }),
    ]);

    expect(globalThis.fetch).toHaveBeenCalledTimes(2);
  });

  it("retries on 5xx errors for GET requests", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({ ok: false, status: 503 })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [] }),
      });

    const result = await api.get("/flaky", { skipDeduplication: true, retryDelay: 1, maxRetries: 1 });
    expect(result).toEqual({ data: [] });
    expect(globalThis.fetch).toHaveBeenCalledTimes(2);
  });

  it("does not retry POST requests", async () => {
    const api = await getApi();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ errors: [{ detail: "Server error" }] }),
    });

    await expect(api.post("/tasks", {})).rejects.toThrow("Server error");
    expect(globalThis.fetch).toHaveBeenCalledTimes(1);
  });
});
