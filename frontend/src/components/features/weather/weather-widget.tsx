import { useNavigate } from "react-router-dom";
import { useCurrentWeather } from "@/lib/hooks/api/use-weather";
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

export function WeatherWidget() {
  const navigate = useNavigate();
  const { household } = useTenant();
  const locationSet = household && hasLocation(household);
  const { data, isLoading, isError } = useCurrentWeather();

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

  const isStale = Date.now() - new Date(weather.fetchedAt).getTime() > 60 * 60 * 1000;

  return (
    <Card
      className="cursor-pointer hover:bg-accent/50 transition-colors"
      onClick={() => navigate("/app/weather")}
    >
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">Weather</CardTitle>
        <WeatherIcon icon={weather.icon} className="h-5 w-5 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-1">
          <span className="text-2xl font-bold">
            {Math.round(weather.temperature)}{weather.temperatureUnit}
          </span>
          <span className="text-sm text-muted-foreground">{weather.summary}</span>
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          H: {Math.round(weather.highTemperature)}° L: {Math.round(weather.lowTemperature)}°
        </p>
        <div className="flex items-center gap-1 mt-2 text-xs text-muted-foreground">
          {household?.attributes.locationName && (
            <span>{household.attributes.locationName}</span>
          )}
          <span className="ml-auto flex items-center gap-1">
            {isStale && <AlertTriangle className="h-3 w-3" />}
            Updated {formatTime(weather.fetchedAt)}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
