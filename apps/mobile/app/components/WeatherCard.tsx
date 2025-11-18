import { Sun } from 'lucide-react';
import type { WeatherData } from '@/lib/api/weather';

interface WeatherCardProps {
  weather: WeatherData | null;
  temperatureUnit?: 'fahrenheit' | 'celsius';
  loading: boolean;
  error: string | null;
}

export function WeatherCard({
  weather,
  temperatureUnit = 'fahrenheit',
  loading,
  error,
}: WeatherCardProps) {
  const formatTemperature = (tempC: number) => {
    if (temperatureUnit === 'celsius') {
      return `${Math.round(tempC)}°C`;
    }
    // Convert Celsius to Fahrenheit
    const tempF = (tempC * 9) / 5 + 32;
    return `${Math.round(tempF)}°F`;
  };

  const formatRelativeTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 60) return 'just now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  if (loading) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="flex items-center justify-between">
            <div className="h-24 w-24 bg-gray-200 rounded-full"></div>
            <div className="h-16 bg-gray-200 rounded w-1/3"></div>
          </div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white border border-red-300 rounded-lg p-6 shadow-sm">
        <p className="text-sm text-red-600">Failed to load weather</p>
        <p className="text-xs text-gray-600 mt-1">{error}</p>
      </div>
    );
  }

  if (!weather || !weather.current) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
        <p className="text-sm text-gray-600">No weather data available</p>
      </div>
    );
  }

  const { current, daily } = weather;
  const todayForecast = daily.length > 0 ? daily[0] : null;

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <h3 className="text-sm font-medium text-gray-600 mb-4">Current Weather</h3>

      <div className="flex items-center justify-between mb-4">
        <Sun className="w-16 h-16 text-yellow-500" />
        <div className="text-right">
          <div className="text-5xl font-bold">
            {formatTemperature(current.temperature_c)}
          </div>
          {todayForecast && (
            <div className="text-sm text-gray-600">
              H: {formatTemperature(todayForecast.tmax_c)} L: {formatTemperature(todayForecast.tmin_c)}
            </div>
          )}
        </div>
      </div>

      <div className="space-y-2">
        <p className="text-xs text-gray-600">
          Updated {formatRelativeTime(current.observed_at)}
        </p>
        {current.stale && (
          <p className="text-xs text-orange-600">
            Weather data may be outdated
          </p>
        )}
      </div>
    </div>
  );
}
