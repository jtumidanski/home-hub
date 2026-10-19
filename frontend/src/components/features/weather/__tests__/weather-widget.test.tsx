import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    household: { attributes: { latitude: 1, longitude: 2, locationName: "Home" } },
  }),
}));
vi.mock("@/lib/hooks/api/use-weather", () => ({
  useCurrentWeather: () => ({
    isLoading: false,
    isError: false,
    data: {
      data: {
        attributes: {
          temperature: 70,
          temperatureUnit: "F",
          summary: "Clear",
          icon: "clear",
          highTemperature: 75,
          lowTemperature: 55,
          fetchedAt: new Date().toISOString(),
        },
      },
    },
  }),
  useWeatherForecast: () => ({
    data: {
      data: ["a", "b", "c", "d"].map((id, i) => ({
        id,
        attributes: {
          date: `2026-03-0${i + 1}`,
          icon: "clear",
          summary: "Clear",
          highTemperature: 60 + i,
          lowTemperature: 40 + i,
        },
      })),
    },
  }),
}));

import { WeatherWidget } from "../weather-widget";

describe("WeatherWidget", () => {
  it("exposes a keyboard-reachable link to the weather page", () => {
    render(
      <MemoryRouter>
        <WeatherWidget />
      </MemoryRouter>,
    );
    expect(screen.getByRole("link", { name: "Weather" })).toHaveAttribute("href", "/app/weather");
  });

  it("shows the current temperature once, as its own focal row, with H/L and three forecast days", () => {
    render(
      <MemoryRouter>
        <WeatherWidget />
      </MemoryRouter>,
    );
    const temp = screen.getByText("70F");
    expect(temp).toHaveClass("tabular-nums", "text-display");
    expect(screen.getByText("H 75°")).toBeInTheDocument();
    expect(screen.getByText("L 55°")).toBeInTheDocument();
    expect(screen.getByText("61°")).toBeInTheDocument();
  });
});
