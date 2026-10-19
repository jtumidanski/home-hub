import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { LocationSearch } from "../location-search";

const results = [
  {
    id: "g1",
    type: "geocoding-results",
    attributes: { name: "Portland", admin1: "Oregon", country: "United States", latitude: 45.5, longitude: -122.6 },
  },
];

vi.mock("@/lib/hooks/api/use-weather", () => ({
  useGeocodingSearch: (q: string) => ({ data: q.length >= 2 ? { data: results } : undefined }),
}));

describe("LocationSearch", () => {
  it("exposes a named combobox that is collapsed with no aria-controls while closed", () => {
    render(<LocationSearch value={null} onSelect={vi.fn()} onClear={vi.fn()} />);

    const combobox = screen.getByRole("combobox", { name: "Search for a city" });
    expect(combobox).toHaveAttribute("aria-expanded", "false");
    expect(combobox).not.toHaveAttribute("aria-controls");
  });

  it("sets aria-expanded and aria-controls to the rendered listbox when results show", async () => {
    const user = userEvent.setup();
    render(<LocationSearch value={null} onSelect={vi.fn()} onClear={vi.fn()} />);

    const combobox = screen.getByRole("combobox", { name: "Search for a city" });
    await user.type(combobox, "Por");

    const listbox = await screen.findByRole("listbox", { name: "Location results" });
    expect(combobox).toHaveAttribute("aria-expanded", "true");
    expect(combobox).toHaveAttribute("aria-controls", listbox.id);
  });

  it("offers a Clear location button when a value is set", async () => {
    const user = userEvent.setup();
    const onClear = vi.fn();
    render(<LocationSearch value="Portland, Oregon" onSelect={vi.fn()} onClear={onClear} />);

    await user.click(screen.getByRole("button", { name: "Clear location" }));
    expect(onClear).toHaveBeenCalled();
  });
});
