import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";

describe("useLocalDateOffset", () => {
  afterEach(() => vi.useRealTimers());

  it("returns tomorrow's date in the given timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    const { result } = renderHook(() => useLocalDateOffset("UTC", 1));
    expect(result.current).toBe("2026-05-02");
  });
});
