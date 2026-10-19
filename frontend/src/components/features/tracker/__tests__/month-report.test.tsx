import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MonthReport } from "../month-report";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { id: "h1", attributes: { timezone: "America/New_York" } } }),
}));

vi.mock("@/lib/hooks/api/use-trackers", () => ({
  useMonthReport: () => ({
    isLoading: false,
    error: null,
    data: {
      data: {
        type: "tracker-reports",
        attributes: {
          month: "2026-05",
          summary: { total_items: 1, completion_rate: 0.5, skip_rate: 0, total_expected: 4, total_filled: 2, total_skipped: 0 },
          items: [
            {
              tracking_item_id: "i1",
              name: "Mood",
              scale_type: "sentiment",
              stats: {
                expected_days: 4, filled_days: 2, skipped_days: 0,
                positive: 2, neutral: 1, negative: 1, positive_ratio: 0.5, daily_values: [],
              },
            },
          ],
        },
      },
    },
  }),
}));

describe("MonthReport sentiment bar", () => {
  it("exposes the breakdown as an image with an accessible label", () => {
    render(<MonthReport month="2026-05" />);
    expect(screen.getByRole("img", { name: "2 positive, 1 neutral, 1 negative" })).toBeInTheDocument();
  });
});
