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

export interface WeatherData {
  current?: CurrentWeather;
  daily: DailyForecast[];
  units: string;
  stale: boolean;
  meta: WeatherMeta;
}

interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

interface JsonApiResponse<T> {
  data: JsonApiResource<T>;
}

/**
 * Fetch weather data for the specified household
 */
export async function getCurrentWeather(householdId: string, signal?: AbortSignal): Promise<WeatherData> {
  const response = await fetch(`/api/weather/combined?householdId=${encodeURIComponent(householdId)}`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`Weather API error: ${response.status} ${response.statusText}`);
  }

  const data = await response.json();

  // Handle JSON:API format
  if (data.data && data.data.attributes) {
    return data.data.attributes as WeatherData;
  }

  return data as WeatherData;
}
