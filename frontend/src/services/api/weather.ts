import { api } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";
import type { WeatherCurrent, WeatherDaily, GeocodingResult } from "@/types/models/weather";

class WeatherService {
  getCurrent(
    lat: number,
    lon: number,
    units: string,
    timezone: string,
    locationId?: string,
  ) {
    const params = new URLSearchParams();
    if (locationId) {
      params.set("locationId", locationId);
    } else {
      params.set("latitude", String(lat));
      params.set("longitude", String(lon));
    }
    params.set("units", units);
    params.set("timezone", timezone);
    return api.get<ApiResponse<WeatherCurrent>>(`/weather/current?${params}`);
  }

  getForecast(
    lat: number,
    lon: number,
    units: string,
    timezone: string,
    locationId?: string,
  ) {
    const params = new URLSearchParams();
    if (locationId) {
      params.set("locationId", locationId);
    } else {
      params.set("latitude", String(lat));
      params.set("longitude", String(lon));
    }
    params.set("units", units);
    params.set("timezone", timezone);
    return api.get<ApiListResponse<WeatherDaily>>(`/weather/forecast?${params}`);
  }

  searchPlaces(query: string) {
    return api.get<ApiListResponse<GeocodingResult>>(`/weather/geocoding?q=${encodeURIComponent(query)}`);
  }
}

export const weatherService = new WeatherService();
