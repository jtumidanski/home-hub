import { useQuery } from "@tanstack/react-query";
import { weatherService } from "@/services/api/weather";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";

export const weatherKeys = {
  all: ["weather"] as const,
  current: (householdId: string | undefined, locationId?: string) =>
    [
      ...weatherKeys.all,
      "current",
      householdId ?? "none",
      locationId ?? "primary",
    ] as const,
  forecast: (householdId: string | undefined, locationId?: string) =>
    [
      ...weatherKeys.all,
      "forecast",
      householdId ?? "none",
      locationId ?? "primary",
    ] as const,
  geocoding: (query: string) =>
    [...weatherKeys.all, "geocoding", query] as const,
};

export function useCurrentWeather(locationId?: string) {
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);
  const enabled = !!locationId || !!locationSet;

  return useQuery({
    queryKey: weatherKeys.current(household?.id, locationId),
    queryFn: () =>
      weatherService.getCurrent(
        household?.attributes.latitude ?? 0,
        household?.attributes.longitude ?? 0,
        household?.attributes.units ?? "metric",
        household?.attributes.timezone ?? "UTC",
        locationId,
      ),
    enabled,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

export function useWeatherForecast(locationId?: string) {
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);
  const enabled = !!locationId || !!locationSet;

  return useQuery({
    queryKey: weatherKeys.forecast(household?.id, locationId),
    queryFn: () =>
      weatherService.getForecast(
        household?.attributes.latitude ?? 0,
        household?.attributes.longitude ?? 0,
        household?.attributes.units ?? "metric",
        household?.attributes.timezone ?? "UTC",
        locationId,
      ),
    enabled,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

export function useGeocodingSearch(query: string) {
  return useQuery({
    queryKey: weatherKeys.geocoding(query),
    queryFn: () => weatherService.searchPlaces(query),
    enabled: query.length >= 2,
    staleTime: 10 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}
