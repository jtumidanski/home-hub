/**
 * Weather Service API
 *
 * Client for interacting with the svc-weather service via the gateway.
 * Handles weather data retrieval and admin operations (cache management).
 */

import { get, post, del } from "./client";

/**
 * Current weather attributes
 */
export interface CurrentWeather {
  temperature_c: number;
  observed_at: string;
  stale: boolean;
  age_seconds: number;
}

/**
 * Daily forecast attributes
 */
export interface DailyForecast {
  date: string;
  tmax_c: number;
  tmin_c: number;
}

/**
 * Weather metadata
 */
export interface WeatherMeta {
  source: string;
  timezone: string;
  geokey: string;
  refreshed_at: string;
}

/**
 * Complete weather response
 */
export interface WeatherResponse {
  current?: CurrentWeather;
  daily: DailyForecast[];
  units: string;
  stale: boolean;
  meta: WeatherMeta;
}

/**
 * JSON:API response wrapper for weather data
 */
export interface JsonApiWeatherResponse {
  data: {
    type: string;
    id: string;
    attributes: WeatherResponse;
  };
}

/**
 * Get weather data for a specific household
 *
 * @param householdId - The UUID of the household
 * @returns Weather data including current conditions and forecast
 * @throws {ApiError} If the request fails
 */
export async function getWeather(householdId: string): Promise<WeatherResponse> {
  const response = await get<JsonApiWeatherResponse>(
    `/weather/combined?householdId=${encodeURIComponent(householdId)}`
  );
  return response.data.attributes;
}

/**
 * Trigger a refresh of cached weather data for a specific household
 *
 * This forces the weather service to fetch fresh data from the upstream provider.
 *
 * @param householdId - The UUID of the household
 * @returns void
 * @throws {ApiError} If the request fails
 */
export async function refreshWeatherCache(householdId: string): Promise<void> {
  await post<void>(
    `/admin/weather/refresh?householdId=${encodeURIComponent(householdId)}`
  );
}

/**
 * Purge cached weather data for a specific household
 *
 * This removes all cached weather data for the household. The next request
 * will trigger a fresh fetch from the upstream provider.
 *
 * @param householdId - The UUID of the household
 * @returns void
 * @throws {ApiError} If the request fails
 */
export async function purgeWeatherCache(householdId: string): Promise<void> {
  await del<void>(
    `/admin/weather/cache?householdId=${encodeURIComponent(householdId)}`
  );
}

/**
 * Purge all cached weather data across all households
 *
 * WARNING: This is a destructive operation that clears all weather caches
 * for all households in the system. Use with caution.
 *
 * @returns void
 * @throws {ApiError} If the request fails
 */
export async function purgeAllWeatherCaches(): Promise<void> {
  await del<void>("/admin/weather/cache/all");
}

/**
 * Convert Celsius to Fahrenheit
 *
 * @param celsius - Temperature in Celsius
 * @returns Temperature in Fahrenheit
 */
export function celsiusToFahrenheit(celsius: number): number {
  return (celsius * 9 / 5) + 32;
}

/**
 * Format temperature for display
 *
 * @param celsius - Temperature in Celsius
 * @param unit - Unit to display (celsius or fahrenheit)
 * @returns Formatted temperature string (e.g., "72°F" or "22°C")
 */
export function formatTemperature(
  celsius: number,
  unit: "celsius" | "fahrenheit" = "celsius"
): string {
  if (unit === "fahrenheit") {
    return `${Math.round(celsiusToFahrenheit(celsius))}°F`;
  }
  return `${Math.round(celsius)}°C`;
}

/**
 * Format relative time (e.g., "2 minutes ago")
 *
 * @param timestamp - ISO 8601 timestamp string
 * @returns Human-readable relative time string
 */
export function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}

/**
 * Format date for forecast display (e.g., "Mon 11/14")
 *
 * @param dateString - ISO 8601 date string
 * @returns Formatted date string with day name and date
 */
export function formatForecastDate(dateString: string): string {
  const date = new Date(dateString);
  const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
  const dayName = days[date.getDay()];
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${dayName} ${month}/${day}`;
}
