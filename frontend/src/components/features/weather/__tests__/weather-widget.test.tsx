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
          fetchedAt: new Date().toISOString(),
        },
      },
    },
  }),
  useWeatherForecast: () => ({ data: { data: [] } }),
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
});
