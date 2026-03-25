import { useQuery } from "@tanstack/react-query";
import { weatherService } from "@/services/api/weather";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";

export const weatherKeys = {
  all: ["weather"] as const,
  current: (householdId: string | undefined) =>
    [...weatherKeys.all, "current", householdId ?? "none"] as const,
  forecast: (householdId: string | undefined) =>
    [...weatherKeys.all, "forecast", householdId ?? "none"] as const,
  geocoding: (query: string) =>
    [...weatherKeys.all, "geocoding", query] as const,
};

export function useCurrentWeather() {
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);

  return useQuery({
    queryKey: weatherKeys.current(household?.id),
    queryFn: () =>
      weatherService.getCurrent(
        household!.attributes.latitude!,
        household!.attributes.longitude!,
        household!.attributes.units,
        household!.attributes.timezone,
      ),
    enabled: !!locationSet,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

export function useWeatherForecast() {
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);

  return useQuery({
    queryKey: weatherKeys.forecast(household?.id),
    queryFn: () =>
      weatherService.getForecast(
        household!.attributes.latitude!,
        household!.attributes.longitude!,
        household!.attributes.units,
        household!.attributes.timezone,
      ),
    enabled: !!locationSet,
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
