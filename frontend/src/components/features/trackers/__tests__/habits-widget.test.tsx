import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const mockToday = vi.fn();
vi.mock("@/lib/hooks/api/use-trackers", () => ({
  useTrackerToday: () => mockToday(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "America/New_York" } } }),
}));
vi.mock("@/lib/hooks/use-local-date", () => ({
  useLocalDate: () => "2026-05-15",
}));

import { HabitsWidget } from "../habits-widget";

function renderWidget() {
  return render(
    <MemoryRouter>
      <HabitsWidget />
    </MemoryRouter>,
  );
}

beforeEach(() => mockToday.mockReset());

describe("HabitsWidget", () => {
  it("shows the shared widget error with an alert role on failure", () => {
    mockToday.mockReturnValue({ data: undefined, isLoading: false, isError: true });
    renderWidget();
    expect(screen.getByRole("alert")).toHaveTextContent("Failed to load habits");
  });

  it("links to the habits page and hovers via the card background, not opacity", () => {
    mockToday.mockReturnValue({
      data: { data: { relationships: { items: { data: [{ id: "i1", name: "Run" }] }, entries: { data: [] } } } },
      isLoading: false,
      isError: false,
    });
    renderWidget();
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/app/habits");
    expect(link.className).not.toContain("hover:opacity");
    expect(screen.getByText("Run")).toBeInTheDocument();
  });
});
