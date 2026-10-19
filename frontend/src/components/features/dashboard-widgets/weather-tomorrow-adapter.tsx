// Read-only widget — no mutations. See PRD §4.3.
import { useWeatherForecast } from "@/lib/hooks/api/use-weather";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "./widget-error";
import { Cloud } from "lucide-react";

export interface WeatherTomorrowConfig {
  units: "imperial" | "metric" | null;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function WeatherTomorrowAdapter({ config: _config }: { config: WeatherTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { data, isLoading, isError } = useWeatherForecast();

  if (isLoading) {
    return (
      <Card className="@container h-full">
        <CardHeader><Skeleton className="h-4 w-24" data-slot="skeleton" /></CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-1" data-slot="skeleton" />
          <Skeleton className="h-3 w-16" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <WidgetError message="Failed to load forecast" />
    );
  }

  const entry = (data?.data ?? []).find((d) => d.attributes.date === tomorrow);
  if (!entry) {
    return (
      <Card className="@container h-full">
        <CardHeader><CardTitle className="font-heading text-lg @xl:text-2xl">Tomorrow</CardTitle></CardHeader>
        <CardContent>
          <p className="text-sm @xl:text-xl text-muted-foreground">Tomorrow's forecast not available</p>
        </CardContent>
      </Card>
    );
  }

  const unit = entry.attributes.temperatureUnit;
  return (
    <Card className="@container h-full">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="font-heading text-lg @xl:text-2xl">Tomorrow</CardTitle>
        <Cloud className="size-4 @xl:size-6 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-2">
          <span className="tabular-nums text-3xl @md:text-display font-bold">{entry.attributes.highTemperature}{unit}</span>
          <span className="text-sm @xl:text-xl text-muted-foreground">/ {entry.attributes.lowTemperature}{unit}</span>
        </div>
        <p className="text-sm @xl:text-base text-muted-foreground">{entry.attributes.summary}</p>
      </CardContent>
    </Card>
  );
}
