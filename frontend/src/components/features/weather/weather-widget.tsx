import { useNavigate } from "react-router-dom";
import { useCurrentWeather, useWeatherForecast } from "@/lib/hooks/api/use-weather";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";
import { WeatherIcon } from "@/components/common/weather-icon";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertTriangle, MapPin } from "lucide-react";
import { Button } from "@/components/ui/button";

function formatTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function formatDayName(dateStr: string): string {
  const date = new Date(dateStr + "T12:00:00");
  return date.toLocaleDateString([], { weekday: "short" });
}

function ForecastDayColumn({ label, icon, high, low, summary }: {
  label: string;
  icon: string;
  high: number;
  low: number;
  summary: string;
}) {
  return (
    <div className="flex flex-col items-center text-center gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span title={summary}>
        <WeatherIcon icon={icon} className="h-4 w-4 text-muted-foreground" />
      </span>
      <div className="text-xs">
        <span className="font-medium">{Math.round(high)}°</span>
        <span className="text-muted-foreground ml-0.5">{Math.round(low)}°</span>
      </div>
    </div>
  );
}

function CurrentDayColumn({ icon, summary, temp, tempUnit, high, low }: {
  icon: string;
  summary: string;
  temp: number;
  tempUnit: string;
  high: number;
  low: number;
}) {
  return (
    <div className="flex flex-col items-center text-center gap-0.5">
      <div className="flex items-center gap-1">
        <span title={summary}>
          <WeatherIcon icon={icon} className="h-4 w-4 text-muted-foreground" />
        </span>
        <span className="text-lg font-bold">{Math.round(temp)}{tempUnit}</span>
      </div>
      <div className="text-xs">
        <span className="font-medium">{Math.round(high)}°</span>
        <span className="text-muted-foreground ml-0.5">{Math.round(low)}°</span>
      </div>
    </div>
  );
}

export function WeatherWidget() {
  const navigate = useNavigate();
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);
  const { data, isLoading, isError } = useCurrentWeather();
  const { data: forecastData } = useWeatherForecast();
  const fetchedAt = data?.data?.attributes?.fetchedAt;
  // eslint-disable-next-line react-hooks/purity -- Date.now() is intentionally impure: we need a fresh stale check each render
  const isStale = fetchedAt ? Date.now() - new Date(fetchedAt).getTime() > 60 * 60 * 1000 : false;

  if (!locationSet) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-6 text-center">
          <MapPin className="h-8 w-8 text-muted-foreground mb-2" />
          <p className="text-sm text-muted-foreground mb-3">
            Set your household location to see weather.
          </p>
          <Button variant="outline" size="sm" onClick={() => navigate("/app/households")}>
            Go to Settings
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-24" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-1" />
          <Skeleton className="h-4 w-32" />
        </CardContent>
      </Card>
    );
  }

  const weather = data?.data?.attributes;

  if (isError || !weather) {
    return (
      <Card>
        <CardContent className="flex items-center gap-2 py-6 text-muted-foreground">
          <AlertTriangle className="h-4 w-4" />
          <span className="text-sm">Weather data unavailable</span>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card
      className="cursor-pointer hover:bg-accent/50 transition-colors"
      onClick={() => navigate("/app/weather")}
    >
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">Weather</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Mobile: centered current conditions + 3-col preview */}
        <div className="md:hidden">
          <div className="flex items-center justify-center gap-2">
            <span title={weather.summary}>
              <WeatherIcon icon={weather.icon} className="h-5 w-5 text-muted-foreground" />
            </span>
            <span className="text-2xl font-bold">
              {Math.round(weather.temperature)}{weather.temperatureUnit}
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1 text-center">
            H:{Math.round(weather.highTemperature)}° L:{Math.round(weather.lowTemperature)}°
          </p>
          {forecastData?.data && forecastData.data.length >= 4 && (
            <div className="grid grid-cols-3 gap-1 mt-3 pt-3 border-t">
              {forecastData.data.slice(1, 4).map((day) => (
                <ForecastDayColumn
                  key={day.id}
                  label={formatDayName(day.attributes.date)}
                  icon={day.attributes.icon}
                  high={day.attributes.highTemperature}
                  low={day.attributes.lowTemperature}
                  summary={day.attributes.summary}
                />
              ))}
            </div>
          )}
        </div>
        {/* Desktop: 4 equal columns with Today including current temp */}
        {forecastData?.data && forecastData.data.length >= 4 && (
          <div className="hidden md:grid md:grid-cols-4 md:gap-2">
            <CurrentDayColumn
              icon={weather.icon}
              summary={weather.summary}
              temp={weather.temperature}
              tempUnit={weather.temperatureUnit}
              high={weather.highTemperature}
              low={weather.lowTemperature}
            />
            {forecastData.data.slice(1, 4).map((day) => (
              <ForecastDayColumn
                key={day.id}
                label={formatDayName(day.attributes.date)}
                icon={day.attributes.icon}
                high={day.attributes.highTemperature}
                low={day.attributes.lowTemperature}
                summary={day.attributes.summary}
              />
            ))}
          </div>
        )}
        <div className="flex items-center gap-1 mt-3 pt-3 border-t text-xs text-muted-foreground">
          {household?.attributes.locationName && (
            <span>{household.attributes.locationName}</span>
          )}
          <span className="ml-auto flex items-center gap-1">
            {isStale && <AlertTriangle className="h-3 w-3" />}
            {formatTime(weather.fetchedAt)}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
