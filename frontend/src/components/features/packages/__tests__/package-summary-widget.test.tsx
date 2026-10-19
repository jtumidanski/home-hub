import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const mockSummary = vi.fn();
vi.mock("@/lib/hooks/api/use-packages", () => ({
  usePackageSummary: () => mockSummary(),
}));

import { PackageSummaryWidget } from "../package-summary-widget";

function renderWidget() {
  return render(
    <MemoryRouter>
      <PackageSummaryWidget />
    </MemoryRouter>,
  );
}

describe("PackageSummaryWidget", () => {
  it("shows the shared widget error with an alert role on failure", () => {
    mockSummary.mockReturnValue({ data: undefined, isLoading: false, isError: true });
    renderWidget();
    expect(screen.getByRole("alert")).toHaveTextContent("Failed to load package summary");
  });

  it("links to packages with a focus ring and card-background hover, not opacity", () => {
    mockSummary.mockReturnValue({
      data: { data: { attributes: { arrivingTodayCount: 1, inTransitCount: 0, exceptionCount: 0 } } },
      isLoading: false,
      isError: false,
    });
    renderWidget();
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/app/packages");
    expect(link.className).toContain("focus-visible:ring-2");
    expect(link.className).not.toContain("opacity");
  });
});
