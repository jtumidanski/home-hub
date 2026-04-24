import { WeatherWidget } from "@/components/features/weather/weather-widget";

export interface WeatherAdapterConfig {
  units: "imperial" | "metric";
  location: { lat: number; lon: number; label: string } | null;
}

/**
 * Registry adapter around WeatherWidget. Maps the dashboard widget config
 * shape to the component's prop shape.
 *
 * Note: `units` and `location` are accepted on WeatherWidget as optional
 * props, but not yet wired through the underlying `useCurrentWeather` hook
 * (which currently reads household settings). That wiring is a follow-up —
 * the registry shape is what this task establishes.
 */
export function WeatherWidgetAdapter({ config }: { config: WeatherAdapterConfig }) {
  return (
    <WeatherWidget
      units={config.units}
      locationOverride={config.location ?? undefined}
    />
  );
}
