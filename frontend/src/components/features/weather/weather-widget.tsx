import { Link, useNavigate } from "react-router-dom";
import { useCurrentWeather, useWeatherForecast } from "@/lib/hooks/api/use-weather";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";
import { WeatherIcon } from "@/components/common/weather-icon";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";
import { AlertTriangle, MapPin } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

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
      <span className="text-xs @xl:text-base text-muted-foreground">{label}</span>
      <span title={summary}>
        <WeatherIcon icon={icon} className="size-6 @xl:size-10 shrink-0 text-foreground" />
      </span>
      <div className="text-xs @xl:text-xl tabular-nums">
        <span className="font-medium">{Math.round(high)}°</span>
        <span className="text-muted-foreground ml-0.5">{Math.round(low)}°</span>
      </div>
    </div>
  );
}

export interface WeatherWidgetProps {
  /**
   * Units for temperature display. Currently informational — the underlying
   * weather API hook is driven by household settings. Config plumbing will
   * be wired in a later task; accepting the prop now preserves the widget
   * registry shape without breaking existing callers.
   */
  units?: "imperial" | "metric";
  /**
   * Per-widget location override. Same caveat as `units`: accepted on the
   * prop contract but not yet wired through `useCurrentWeather`; the widget
   * still reads from the household.
   */
  locationOverride?: { lat: number; lon: number; label: string };
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function WeatherWidget(_props: WeatherWidgetProps = {}) {
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
      <Card className="@container h-full">
        <CardContent className="flex flex-col items-center justify-center py-6 text-center">
          <MapPin aria-hidden="true" className="size-8 @xl:size-12 text-muted-foreground mb-2" />
          <p className="text-sm @xl:text-xl text-muted-foreground mb-3">
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
      <Card className="@container h-full">
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
    return <WidgetError message="Weather data unavailable" />;
  }

  return (
    <Card className="@container relative h-full cursor-pointer hover:bg-accent/50 transition-colors">
      <CardHeader className="pb-2">
        <CardTitle className="font-heading text-lg @xl:text-2xl">
          <Link
            to="/app/weather"
            className="rounded-sm outline-none after:absolute after:inset-0 after:content-[''] focus-visible:after:outline-2 focus-visible:after:outline-offset-2 focus-visible:after:outline-ring"
          >
            Weather
          </Link>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-3">
          <span title={weather.summary}>
            <WeatherIcon icon={weather.icon} className="size-10 @xl:size-14 shrink-0 text-foreground" />
          </span>
          <span className="tabular-nums text-display font-bold leading-none">
            {Math.round(weather.temperature)}{weather.temperatureUnit}
          </span>
        </div>
        <p className="mt-2 text-sm @xl:text-lg text-muted-foreground">{weather.summary}</p>
        <p className="mt-1 text-sm @xl:text-lg tabular-nums">
          <span className="font-medium">H {Math.round(weather.highTemperature)}°</span>
          <span className="ml-3 text-muted-foreground">L {Math.round(weather.lowTemperature)}°</span>
        </p>
        {forecastData?.data && forecastData.data.length >= 4 && (
          <div className="mt-3 grid grid-cols-3 gap-2 border-t border-border pt-3">
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
        <div className="flex items-center gap-1 mt-3 pt-3 border-t text-xs @xl:text-base text-muted-foreground">
          {household?.attributes.locationName && (
            <span>{household.attributes.locationName}</span>
          )}
          <span className={cn("ml-auto flex items-center gap-1 tabular-nums", isStale && "text-warning-text")}>
            {isStale && <AlertTriangle aria-hidden="true" className="size-4 @xl:size-6" />}
            {isStale && <span>Stale</span>}
            {formatTime(weather.fetchedAt)}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
