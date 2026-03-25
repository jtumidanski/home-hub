import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWeatherForecast, useCurrentWeather } from "@/lib/hooks/api/use-weather";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";
import { WeatherIcon } from "@/components/common/weather-icon";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { MapPin, AlertTriangle, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { HourlyForecastEntry } from "@/types/models/weather";

function formatTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function formatDayName(dateStr: string, index: number): string {
  if (index === 0) return "Today";
  const date = new Date(dateStr + "T12:00:00");
  return date.toLocaleDateString([], { weekday: "short", month: "short", day: "numeric" });
}

function formatHourlyTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function getFilteredHourlyEntries(entries: HourlyForecastEntry[], isToday: boolean): HourlyForecastEntry[] {
  if (!isToday) return entries;
  const currentHour = new Date().getHours();
  return entries.filter((entry) => {
    const entryHour = new Date(entry.time).getHours();
    return entryHour >= currentHour;
  });
}

function HourlyBreakdown({ entries, isToday }: { entries: HourlyForecastEntry[]; isToday: boolean }) {
  const filtered = getFilteredHourlyEntries(entries, isToday);
  if (filtered.length === 0) return null;

  return (
    <div className="border-t mt-2 pt-2 max-h-64 overflow-y-auto">
      {filtered.map((entry) => (
        <div key={entry.time} className="flex items-center gap-3 py-1.5 px-1 text-sm">
          <span className="w-16 shrink-0 text-muted-foreground text-xs">
            {formatHourlyTime(entry.time)}
          </span>
          <WeatherIcon icon={entry.icon} className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="w-10 shrink-0 font-medium">{Math.round(entry.temperature)}°</span>
          <span className="text-xs text-muted-foreground flex-1 min-w-0 truncate">
            {entry.summary}
          </span>
          <span className="text-xs text-muted-foreground shrink-0">
            {entry.precipitationProbability}%
          </span>
        </div>
      ))}
    </div>
  );
}

function WeatherSkeleton() {
  return (
    <div className="p-4 md:p-6 space-y-4">
      <Skeleton className="h-8 w-48" />
      {Array.from({ length: 7 }).map((_, i) => (
        <Skeleton key={i} className="h-16" />
      ))}
    </div>
  );
}

export function WeatherPage() {
  const navigate = useNavigate();
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);
  const { data: forecastData, isLoading: forecastLoading, isError: forecastError, refetch: refetchForecast } = useWeatherForecast();
  const { data: currentData, refetch: refetchCurrent } = useCurrentWeather();
  const [expandedDay, setExpandedDay] = useState<string | null>(null);

  const handleRefresh = useCallback(async () => {
    await Promise.all([refetchForecast(), refetchCurrent()]);
  }, [refetchForecast, refetchCurrent]);

  if (!locationSet) {
    return (
      <div className="p-4 md:p-6">
        <h1 className="text-xl md:text-2xl font-semibold mb-4">Weather</h1>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-8 text-center">
            <MapPin className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="text-muted-foreground mb-4">
              No location set for this household.
            </p>
            <Button variant="outline" onClick={() => navigate("/app/households")}>
              Set Location
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (forecastLoading) {
    return <WeatherSkeleton />;
  }

  if (forecastError || !forecastData?.data) {
    return (
      <div className="p-4 md:p-6">
        <h1 className="text-xl md:text-2xl font-semibold mb-4">Weather</h1>
        <Card>
          <CardContent className="flex items-center gap-2 py-6 text-muted-foreground">
            <AlertTriangle className="h-4 w-4" />
            <span>Failed to load weather forecast. Please try again later.</span>
          </CardContent>
        </Card>
      </div>
    );
  }

  const forecast = forecastData.data;
  const fetchedAt = currentData?.data?.attributes?.fetchedAt;

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div>
          <h1 className="text-xl md:text-2xl font-semibold">
            Weather {household?.attributes.locationName ? `\u2014 ${household.attributes.locationName}` : ""}
          </h1>
          {fetchedAt && (
            <p className="text-xs text-muted-foreground mt-1">
              Updated {formatTime(fetchedAt)}
            </p>
          )}
        </div>

        <div className="space-y-2">
          {forecast.map((day, index) => {
            const isExpanded = expandedDay === day.id;
            const hasHourly = day.attributes.hourlyForecast?.length > 0;
            return (
              <Card
                key={day.id}
                className={cn(
                  hasHourly && "cursor-pointer",
                  index === 0 && "border-primary bg-primary/5",
                )}
                onClick={() => {
                  if (hasHourly) {
                    setExpandedDay(isExpanded ? null : day.id);
                  }
                }}
              >
                <CardContent className="py-3 px-4">
                  <div className="flex items-center gap-4">
                    <div className="w-28 shrink-0">
                      <p className="text-sm font-medium">
                        {formatDayName(day.attributes.date, index)}
                      </p>
                    </div>
                    <WeatherIcon icon={day.attributes.icon} className="h-6 w-6 shrink-0 text-muted-foreground" />
                    <p className="text-sm text-muted-foreground flex-1 min-w-0 truncate">
                      {day.attributes.summary}
                    </p>
                    <div className="text-sm text-right shrink-0">
                      <span className="font-medium">
                        {Math.round(day.attributes.highTemperature)}°
                      </span>
                      <span className="text-muted-foreground ml-2">
                        {Math.round(day.attributes.lowTemperature)}°
                      </span>
                    </div>
                    {hasHourly && (
                      <ChevronDown className={cn(
                        "h-4 w-4 shrink-0 text-muted-foreground transition-transform",
                        isExpanded && "rotate-180",
                      )} />
                    )}
                  </div>
                  {isExpanded && hasHourly && (
                    <HourlyBreakdown
                      entries={day.attributes.hourlyForecast}
                      isToday={index === 0}
                    />
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </PullToRefresh>
  );
}
