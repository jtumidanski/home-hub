import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useLocalDate } from "../use-local-date";

describe("useLocalDate", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns today's date in the supplied timezone", () => {
    // 2026-04-08 22:00 in America/New_York (UTC-4)
    vi.setSystemTime(new Date("2026-04-09T02:00:00Z"));
    const { result } = renderHook(() => useLocalDate("America/New_York"));
    expect(result.current).toBe("2026-04-08");
  });

  it("re-renders when the local date crosses midnight", () => {
    // Start at 2026-04-08 23:59:30 EDT
    vi.setSystemTime(new Date("2026-04-09T03:59:30Z"));
    const { result } = renderHook(() => useLocalDate("America/New_York"));
    expect(result.current).toBe("2026-04-08");

    // Advance past local midnight (~30s + 60s poll => crosses midnight)
    act(() => {
      vi.setSystemTime(new Date("2026-04-09T04:00:30Z"));
      vi.advanceTimersByTime(60_000);
    });

    expect(result.current).toBe("2026-04-09");
  });

  it("does not re-render when the local date is unchanged", () => {
    vi.setSystemTime(new Date("2026-04-09T15:00:00Z"));
    let renders = 0;
    const { result } = renderHook(() => {
      renders++;
      return useLocalDate("America/New_York");
    });
    const initialRenders = renders;
    expect(result.current).toBe("2026-04-09");

    act(() => {
      vi.advanceTimersByTime(60_000);
    });

    expect(result.current).toBe("2026-04-09");
    // The mount effect bumps renders by 1, then a poll-tick with no change
    // should not cause an additional re-render.
    expect(renders).toBe(initialRenders);
  });
});
