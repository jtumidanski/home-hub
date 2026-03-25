import { api } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";
import type { WeatherCurrent, WeatherDaily, GeocodingResult } from "@/types/models/weather";

class WeatherService {
  getCurrent(lat: number, lon: number, units: string, timezone: string) {
    const params = new URLSearchParams({
      latitude: String(lat),
      longitude: String(lon),
      units,
      timezone,
    });
    return api.get<ApiResponse<WeatherCurrent>>(`/weather/current?${params}`);
  }

  getForecast(lat: number, lon: number, units: string, timezone: string) {
    const params = new URLSearchParams({
      latitude: String(lat),
      longitude: String(lon),
      units,
      timezone,
    });
    return api.get<ApiListResponse<WeatherDaily>>(`/weather/forecast?${params}`);
  }

  searchPlaces(query: string) {
    return api.get<ApiListResponse<GeocodingResult>>(`/weather/geocoding?q=${encodeURIComponent(query)}`);
  }
}

export const weatherService = new WeatherService();
