'use client';

import React from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import {
  WeatherResponse,
  formatTemperature,
  formatRelativeTime,
  formatForecastDate,
} from '@/lib/api/weather';
import { Cloud, CloudRain, Sun, Wind } from 'lucide-react';

interface WeatherCardProps {
  weather?: WeatherResponse | null;
  loading?: boolean;
  error?: string | null;
}

export function WeatherCard({ weather, loading, error }: WeatherCardProps) {
  if (error) {
    return (
      <Card title="Weather">
        <div className="text-center py-8">
          <Cloud className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-2" />
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Will retry automatically
          </p>
        </div>
      </Card>
    );
  }

  if (loading || !weather) {
    return <Card title="Weather" loading={true} />;
  }

  return (
    <Card title="Weather">
      <div className="space-y-6">
        {/* Current Weather */}
        {weather.current && (
          <CardSection>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-5xl font-bold text-gray-900 dark:text-white">
                  {formatTemperature(weather.current.temperature_c, weather.units)}
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {formatRelativeTime(weather.current.observed_at)}
                  {weather.current.stale && (
                    <span className="ml-2 text-yellow-600 dark:text-yellow-400">
                      (stale)
                    </span>
                  )}
                </div>
              </div>
              <Sun className="h-16 w-16 text-yellow-500" />
            </div>
          </CardSection>
        )}

        {/* 7-Day Forecast */}
        {weather.daily && weather.daily.length > 0 && (
          <CardSection title="7-Day Forecast">
            <div className="space-y-2">
              {weather.daily.map((day, index) => (
                <div
                  key={day.date}
                  className="flex items-center justify-between py-2 border-b border-gray-200 dark:border-gray-700 last:border-0"
                >
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300 w-24">
                    {index === 0 ? 'Today' : formatForecastDate(day.date)}
                  </span>
                  <div className="flex items-center gap-4 text-sm">
                    <span className="text-gray-900 dark:text-white font-medium">
                      {formatTemperature(day.tmax_c, weather.units)}
                    </span>
                    <span className="text-gray-500 dark:text-gray-400">
                      {formatTemperature(day.tmin_c, weather.units)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardSection>
        )}

        {/* Metadata */}
        <div className="text-xs text-gray-500 dark:text-gray-400 pt-2 border-t border-gray-200 dark:border-gray-700">
          <div className="flex justify-between">
            <span>Source: {weather.meta.source}</span>
            <span>{formatRelativeTime(weather.meta.refreshed_at)}</span>
          </div>
          {weather.stale && (
            <div className="text-yellow-600 dark:text-yellow-400 mt-1">
              Data may be outdated
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}
