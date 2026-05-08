import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

import type { MonthSummaryResponse, TrackerEntry } from "@/types/models/tracker";
import { CalendarGrid } from "../calendar-grid";

// --- Module-boundary mocks --------------------------------------------------

vi.mock("@/lib/hooks/api/use-trackers", () => ({
  useMonthSummary: vi.fn(),
  usePutEntry: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteEntry: () => ({ mutate: vi.fn(), isPending: false }),
  useSkipEntry: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    household: { id: "h1", attributes: { timezone: "America/New_York" } },
  }),
}));

vi.mock("@/lib/date-utils", async () => {
  const actual = await vi.importActual<typeof import("@/lib/date-utils")>(
    "@/lib/date-utils",
  );
  return {
    ...actual,
    getLocalTodayStr: () => "2026-05-15",
    getLocalMonth: () => "2026-05",
  };
});

import { useMonthSummary } from "@/lib/hooks/api/use-trackers";

const mockedUseMonthSummary = vi.mocked(useMonthSummary);

// --- Fixtures ---------------------------------------------------------------

function makeEntry(
  itemId: string,
  date: string,
  overrides: Partial<TrackerEntry["attributes"]> = {},
): TrackerEntry {
  return {
    id: `${itemId}-${date}`,
    type: "tracker_entries",
    attributes: {
      tracking_item_id: itemId,
      date,
      value: { rating: "positive" },
      skipped: false,
      note: null,
      ...overrides,
    },
  } as unknown as TrackerEntry;
}

function makeSummary(entries: TrackerEntry[]): MonthSummaryResponse {
  return {
    data: {
      id: "2026-05",
      type: "tracker_month_summary",
      attributes: {
        month: "2026-05",
        completion: { expected: 30, filled: 0, skipped: 0 },
        complete: false,
      },
      relationships: {
        items: {
          data: [
            {
              id: "item-1",
              name: "Run",
              color: "blue",
              scale_type: "sentiment",
              scale_config: null,
              sort_order: 0,
              active_from: "2026-01-01",
              active_until: null,
              schedule_snapshots: [
                {
                  effective_date: "2026-01-01",
                  schedule: [0, 1, 2, 3, 4, 5, 6],
                },
              ],
            },
          ],
        },
        entries: { data: entries },
      },
    },
  } as unknown as MonthSummaryResponse;
}

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true });
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

// --- Smoke test -------------------------------------------------------------

describe("CalendarGrid (desktop)", () => {
  it("renders the desktop calendar table for the given month", () => {
    mockedUseMonthSummary.mockReturnValue({
      data: makeSummary([]),
      isLoading: false,
    } as ReturnType<typeof useMonthSummary>);

    render(
      <CalendarGrid month="2026-05" onMonthChange={() => {}} onViewReport={() => {}} />,
    );

    expect(screen.getByText("May 2026")).toBeInTheDocument();
    // Both desktop table and mobile day view render the item name; the desktop
    // table is the surface under test, so target it explicitly.
    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.getAllByText("Run").length).toBeGreaterThan(0);
  });

  describe("note indicator", () => {
    it("renders the solid border and flange when the cell entry has a non-empty note", () => {
      mockedUseMonthSummary.mockReturnValue({
        data: makeSummary([makeEntry("item-1", "2026-05-10", { note: "ran 3 miles" })]),
        isLoading: false,
      } as ReturnType<typeof useMonthSummary>);

      render(
        <CalendarGrid month="2026-05" onMonthChange={() => {}} onViewReport={() => {}} />,
      );

      const trigger = screen
        .getAllByRole("button")
        .find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 10"));
      expect(trigger).toBeDefined();
      expect(trigger!.className).toContain("border-blue-500");
      const flange = trigger!.querySelector('[aria-hidden="true"]');
      expect(flange).not.toBeNull();
      expect(flange?.className).toContain("bg-blue-500");
      expect(flange?.getAttribute("class")).toMatch(/\[clip-path:polygon/);
    });

    it("shows the note text in a tooltip after hover, with the 200ms provider delay", async () => {
      mockedUseMonthSummary.mockReturnValue({
        data: makeSummary([makeEntry("item-1", "2026-05-10", { note: "line one\nline two" })]),
        isLoading: false,
      } as ReturnType<typeof useMonthSummary>);

      render(
        <CalendarGrid month="2026-05" onMonthChange={() => {}} onViewReport={() => {}} />,
      );

      const trigger = screen
        .getAllByRole("button")
        .find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 10"))!;

      // Base UI tooltip's pointer-enter open path does not flush under jsdom;
      // focus opens it on the same code path with the same content. The 200 ms
      // delay is exercised manually below to match the configured provider delay.
      fireEvent.focus(trigger);
      await act(async () => {
        vi.advanceTimersByTime(250);
      });

      const tooltip = await screen.findByText(/line one/);
      expect(tooltip).toHaveTextContent("line two");
      // Multiline preserved: the popup uses whitespace-pre-wrap.
      expect(tooltip.className).toContain("whitespace-pre-wrap");
    });
  });

  describe("cursor affordance", () => {
    it("applies cursor-pointer to past and today triggers and not to future placeholders", () => {
      mockedUseMonthSummary.mockReturnValue({
        data: makeSummary([
          makeEntry("item-1", "2026-05-10", { note: null }), // past, no note
          makeEntry("item-1", "2026-05-15", { note: null }), // today, no note
        ]),
        isLoading: false,
      } as ReturnType<typeof useMonthSummary>);

      render(
        <CalendarGrid month="2026-05" onMonthChange={() => {}} onViewReport={() => {}} />,
      );

      const buttons = screen.getAllByRole("button");
      const past = buttons.find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 10"))!;
      const today = buttons.find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 15"))!;
      expect(past.className).toContain("cursor-pointer");
      expect(today.className).toContain("cursor-pointer");

      // Future cells are <span>s, never <button>s.
      const futureTrigger = buttons.find((b) =>
        b.getAttribute("aria-label")?.startsWith("Run, May 20"),
      );
      expect(futureTrigger).toBeUndefined();
    });
  });

  describe("aria-label", () => {
    it("describes the cell with item name, formatted date, score, and note (when present)", () => {
      mockedUseMonthSummary.mockReturnValue({
        data: makeSummary([
          makeEntry("item-1", "2026-05-10", {
            value: { rating: "negative" },
            note: "tough day",
          }),
          makeEntry("item-1", "2026-05-11", {
            value: { rating: "positive" },
            note: null,
          }),
        ]),
        isLoading: false,
      } as ReturnType<typeof useMonthSummary>);

      render(
        <CalendarGrid month="2026-05" onMonthChange={() => {}} onViewReport={() => {}} />,
      );

      const buttons = screen.getAllByRole("button");
      const may10 = buttons.find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 10"))!;
      const may11 = buttons.find((b) => b.getAttribute("aria-label")?.startsWith("Run, May 11"))!;

      expect(may10.getAttribute("aria-label")).toBe(
        "Run, May 10. Sentiment negative. Note: tough day",
      );
      expect(may11.getAttribute("aria-label")).toBe(
        "Run, May 11. Sentiment positive.",
      );
      // No emoji in the aria-label, even though the visible glyph is one.
      expect(may10.getAttribute("aria-label")).not.toMatch(/😊|😞|😐/);
    });
  });
});
