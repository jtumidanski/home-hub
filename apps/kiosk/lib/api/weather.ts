export interface CurrentWeather {
  temperature_c: number;
  observed_at: string;
  stale: boolean;
  age_seconds: number;
}

export interface DailyForecast {
  date: string;
  tmax_c: number;
  tmin_c: number;
}

export interface WeatherMeta {
  source: string;
  timezone: string;
  geokey: string;
  refreshed_at: string;
}

export interface WeatherResponse {
  current?: CurrentWeather;
  daily: DailyForecast[];
  units: string;
  stale: boolean;
  meta: WeatherMeta;
}

/**
 * Fetch weather data for the specified household
 */
export async function getWeather(householdId: string, signal?: AbortSignal): Promise<WeatherResponse> {
  const response = await fetch(`/api/weather/combined?householdId=${encodeURIComponent(householdId)}`, {
    credentials: 'include',
    signal,
  });

  if (!response.ok) {
    throw new Error(`Weather API error: ${response.status} ${response.statusText}`);
  }

  const data = await response.json();

  // Handle JSON:API format
  if (data.data && data.data.attributes) {
    return data.data.attributes;
  }

  return data;
}

/**
 * Format temperature for display
 */
export function formatTemperature(celsius: number, units: string = 'celsius'): string {
  if (units === 'celsius') {
    return `${Math.round(celsius)}°C`;
  }
  // Future: Add Fahrenheit conversion if needed
  return `${Math.round(celsius)}°C`;
}

/**
 * Format relative time (e.g., "2 minutes ago")
 */
export function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}

/**
 * Format date for display (e.g., "Mon 11/14")
 */
export function formatForecastDate(dateString: string): string {
  const date = new Date(dateString);
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const dayName = days[date.getDay()];
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${dayName} ${month}/${day}`;
}
