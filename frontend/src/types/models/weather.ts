export interface WeatherCurrentAttributes {
  temperature: number;
  temperatureUnit: string;
  summary: string;
  icon: string;
  weatherCode: number;
  highTemperature: number;
  lowTemperature: number;
  fetchedAt: string;
}

export interface WeatherCurrent {
  id: string;
  type: "weather-current";
  attributes: WeatherCurrentAttributes;
}

export interface WeatherDailyAttributes {
  date: string;
  highTemperature: number;
  lowTemperature: number;
  temperatureUnit: string;
  summary: string;
  icon: string;
  weatherCode: number;
}

export interface WeatherDaily {
  id: string;
  type: "weather-daily";
  attributes: WeatherDailyAttributes;
}

export interface GeocodingResultAttributes {
  name: string;
  country: string;
  admin1: string;
  latitude: number;
  longitude: number;
}

export interface GeocodingResult {
  id: string;
  type: "geocoding-results";
  attributes: GeocodingResultAttributes;
}
