import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TrackerPage } from "../TrackerPage";

vi.mock("@/components/features/tracker/today-view", () => ({ TodayView: () => <div>today-view</div> }));
vi.mock("@/components/features/tracker/tracker-setup", () => ({ TrackerSetup: () => <div>setup-view</div> }));
vi.mock("@/components/features/tracker/month-report", () => ({ MonthReport: () => <div>report-view</div> }));
vi.mock("@/components/features/tracker/calendar-grid", () => ({
  CalendarGrid: ({ onOpenSetup }: { onOpenSetup: () => void }) => (
    <button onClick={onOpenSetup}>open-setup-from-calendar</button>
  ),
}));

describe("TrackerPage", () => {
  it("renders the Habits page heading", () => {
    render(<TrackerPage />);
    expect(screen.getByRole("heading", { level: 1, name: "Habits" })).toBeInTheDocument();
  });

  it("marks the active view button with aria-pressed and switches on click", () => {
    render(<TrackerPage />);
    expect(screen.getByRole("button", { name: /Today/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: /Calendar/ })).toHaveAttribute("aria-pressed", "false");

    fireEvent.click(screen.getByRole("button", { name: /Calendar/ }));
    expect(screen.getByRole("button", { name: /Calendar/ })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: /Today/ })).toHaveAttribute("aria-pressed", "false");
  });

  it("lets the calendar jump to Setup", () => {
    render(<TrackerPage />);
    fireEvent.click(screen.getByRole("button", { name: /Calendar/ }));
    fireEvent.click(screen.getByText("open-setup-from-calendar"));
    expect(screen.getByText("setup-view")).toBeInTheDocument();
  });
});
