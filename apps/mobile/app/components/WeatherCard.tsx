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
      <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-1/3"></div>
          <div className="flex items-center justify-between">
            <div className="h-24 w-24 bg-muted rounded-full"></div>
            <div className="h-16 bg-muted rounded w-1/3"></div>
          </div>
          <div className="h-4 bg-muted rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-card border border-destructive/30 rounded-lg p-6 shadow-sm">
        <p className="text-sm text-destructive">Failed to load weather</p>
        <p className="text-xs text-muted-foreground mt-1">{error}</p>
      </div>
    );
  }

  if (!weather || !weather.current) {
    return (
      <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
        <p className="text-sm text-muted-foreground">No weather data available</p>
      </div>
    );
  }

  const { current, daily } = weather;
  const todayForecast = daily.length > 0 ? daily[0] : null;

  return (
    <div className="bg-card text-card-foreground border border-border rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <h3 className="text-sm font-medium text-muted-foreground mb-4">Current Weather</h3>

      <div className="flex items-center justify-between mb-4">
        <Sun className="w-16 h-16 text-yellow-500" />
        <div className="text-right">
          <div className="text-5xl font-bold">
            {formatTemperature(current.temperature_c)}
          </div>
          {todayForecast && (
            <div className="text-sm text-muted-foreground">
              H: {formatTemperature(todayForecast.tmax_c)} L: {formatTemperature(todayForecast.tmin_c)}
            </div>
          )}
        </div>
      </div>

      <div className="space-y-2">
        <p className="text-xs text-muted-foreground">
          Updated {formatRelativeTime(current.observed_at)}
        </p>
        {current.stale && (
          <p className="text-xs text-orange-600 dark:text-orange-500">
            Weather data may be outdated
          </p>
        )}
      </div>
    </div>
  );
}
