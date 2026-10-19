import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeatherIcon } from "../weather-icon";

describe("WeatherIcon", () => {
  it("is decorative by default and uses a chart token colour", () => {
    const { container } = render(<WeatherIcon icon="sun" />);
    const svg = container.querySelector("svg");
    expect(svg).toHaveAttribute("aria-hidden", "true");
    expect(svg).toHaveClass("text-chart-3");
  });

  it("is exposed as a named image when aria-label is given", () => {
    render(<WeatherIcon icon="cloud-rain" aria-label="Rain" />);
    expect(screen.getByRole("img", { name: "Rain" })).toBeInTheDocument();
  });
});
