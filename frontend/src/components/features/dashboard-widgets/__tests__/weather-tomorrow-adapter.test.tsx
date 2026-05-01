import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeatherTomorrowAdapter } from "@/components/features/dashboard-widgets/weather-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-weather", () => ({ useWeatherForecast: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useWeatherForecast } from "@/lib/hooks/api/use-weather";

const daily = (date: string, hi: number, lo: number, unit = "F") => ({
  id: date, type: "weather-daily",
  attributes: { date, highTemperature: hi, lowTemperature: lo, temperatureUnit: unit, summary: "Sunny", icon: "sun", weatherCode: 0, hourlyForecast: [] },
});

describe("WeatherTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

  it("renders tomorrow's high/low", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useWeatherForecast as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        daily("2026-05-01", 70, 50),
        daily("2026-05-02", 75, 55),
      ] },
      isLoading: false, isError: false,
    });
    render(<WeatherTomorrowAdapter config={{ units: null }} />);
    expect(screen.getByText("75°F")).toBeInTheDocument();
    expect(screen.getByText("/ 55°F")).toBeInTheDocument();
  });

  it("shows fallback when tomorrow is missing from forecast", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useWeatherForecast as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [daily("2026-05-01", 70, 50)] },
      isLoading: false, isError: false,
    });
    render(<WeatherTomorrowAdapter config={{ units: null }} />);
    expect(screen.getByText(/Tomorrow's forecast not available/i)).toBeInTheDocument();
  });
});
