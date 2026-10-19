import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { ReauthorizeBanner } from "../reauthorize-banner";

vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useReauthorizeCalendar: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("ReauthorizeBanner", () => {
  it("renders a warning status banner with an upgrade action", () => {
    render(<ReauthorizeBanner />);
    const banner = screen.getByRole("status");
    expect(banner).toHaveAttribute("data-variant", "warning");
    expect(banner).toHaveTextContent(/upgrade your calendar connection/i);
    expect(screen.getByRole("button", { name: "Upgrade Access" })).toBeInTheDocument();
  });
});
