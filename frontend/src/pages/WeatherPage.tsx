import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useWeatherForecast, useCurrentWeather } from "@/lib/hooks/api/use-weather";
import { useLocationsOfInterest } from "@/lib/hooks/api/use-locations-of-interest";
import { useTenant } from "@/context/tenant-context";
import { hasLocation } from "@/types/models/household";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { WeatherIcon } from "@/components/common/weather-icon";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { MapPin, AlertTriangle, ChevronDown, Settings } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ManageLocationsDialog } from "@/components/features/weather/manage-locations-dialog";
import { cn } from "@/lib/utils";
import type { HourlyForecastEntry } from "@/types/models/weather";

const PRIMARY_VALUE = "__primary__";

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
    <div className="border-t mt-2 pt-2 max-h-64 overflow-y-auto md:max-h-none md:overflow-visible">
      {filtered.map((entry) => (
        <div key={entry.time} className="flex items-center gap-3 py-1.5 px-1 text-sm">
          <span className="w-16 shrink-0 text-muted-foreground text-xs">
            {formatHourlyTime(entry.time)}
          </span>
          <WeatherIcon icon={entry.icon} className="size-6 shrink-0 text-foreground" />
          <span className="w-10 shrink-0 font-medium tabular-nums">{Math.round(entry.temperature)}°</span>
          <span className="w-24 shrink-0 text-xs text-muted-foreground truncate">
            {entry.summary}
          </span>
          <span className="w-8 shrink-0 text-xs text-muted-foreground">
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

  // Always defaults to undefined (= primary) on mount; reload resets.
  const [selectedLocationId, setSelectedLocationId] = useState<string | undefined>(undefined);
  const [manageOpen, setManageOpen] = useState(false);
  const [expandedDay, setExpandedDay] = useState<string | null>(null);

  const { data: locationsData } = useLocationsOfInterest();
  const savedLocations = locationsData?.data ?? [];

  const { data: forecastData, isLoading: forecastLoading, isError: forecastError, refetch: refetchForecast } =
    useWeatherForecast(selectedLocationId);
  const { data: currentData, refetch: refetchCurrent } = useCurrentWeather(selectedLocationId);

  const handleRefresh = useCallback(async () => {
    await Promise.all([refetchForecast(), refetchCurrent()]);
  }, [refetchForecast, refetchCurrent]);

  const handleSelectChange = (value: string | null) => {
    setSelectedLocationId(!value || value === PRIMARY_VALUE ? undefined : value);
  };

  const headerLocationName = (() => {
    if (selectedLocationId) {
      const match = savedLocations.find((l) => l.id === selectedLocationId);
      if (match) return match.attributes.label ?? match.attributes.placeName;
    }
    return household?.attributes.locationName ?? null;
  })();

  const header = (
    <PageHeader
      title={`Weather${headerLocationName ? ` \u2014 ${headerLocationName}` : ""}`}
      actions={
        <Button variant="ghost" size="sm" onClick={() => setManageOpen(true)}>
          <Settings className="h-3.5 w-3.5 mr-1" />
          Manage Locations
        </Button>
      }
    />
  );

  const selector =
    savedLocations.length > 0 ? (
      <div className="space-y-1">
        <Label htmlFor="weather-location-select" className="text-xs text-muted-foreground">
          Location
        </Label>
        <Select
          value={selectedLocationId ?? PRIMARY_VALUE}
          onValueChange={handleSelectChange}
        >
          <SelectTrigger id="weather-location-select" className="w-full max-w-xs">
            <SelectValue>
              {selectedLocationId
                ? (savedLocations.find((l) => l.id === selectedLocationId)?.attributes.label ??
                   savedLocations.find((l) => l.id === selectedLocationId)?.attributes.placeName)
                : (household?.attributes.locationName ?? "Primary Location")}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={PRIMARY_VALUE}>
              {household?.attributes.locationName ?? "Primary Location"}
            </SelectItem>
            {savedLocations.map((loc) => (
              <SelectItem key={loc.id} value={loc.id}>
                {loc.attributes.label ?? loc.attributes.placeName}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    ) : null;

  // Empty state: primary unset.
  if (!locationSet && !selectedLocationId) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        {header}
        <Card>
          <CardContent>
            <EmptyState
              icon={MapPin}
              title="No location set"
              description="Set a primary location for this household to see its forecast."
              action={
                <Button variant="outline" onClick={() => navigate("/app/households")}>
                  Set Location
                </Button>
              }
            />
            {savedLocations.length > 0 && (
              <div className="w-full pb-4">
                <p className="text-sm text-muted-foreground mb-2 text-center">
                  Or pick one of your saved locations:
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {savedLocations.map((loc) => (
                    <Button
                      key={loc.id}
                      variant="secondary"
                      size="sm"
                      onClick={() => setSelectedLocationId(loc.id)}
                    >
                      {loc.attributes.label ?? loc.attributes.placeName}
                    </Button>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
        <ManageLocationsDialog open={manageOpen} onOpenChange={setManageOpen} />
      </div>
    );
  }

  if (forecastLoading) {
    return (
      <div>
        <WeatherSkeleton />
        <ManageLocationsDialog open={manageOpen} onOpenChange={setManageOpen} />
      </div>
    );
  }

  if (forecastError || !forecastData?.data) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        {header}
        {selector}
        <Card>
          <CardContent className="flex items-center gap-2 py-6 text-muted-foreground">
            <AlertTriangle aria-hidden="true" className="h-4 w-4" />
            <span>Failed to load weather forecast. Please try again later.</span>
          </CardContent>
        </Card>
        <ManageLocationsDialog open={manageOpen} onOpenChange={setManageOpen} />
      </div>
    );
  }

  const forecast = forecastData.data;
  const fetchedAt = currentData?.data?.attributes?.fetchedAt;

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="space-y-2">
          {header}
          {fetchedAt && (
            <p className="text-xs text-muted-foreground">
              Updated {formatTime(fetchedAt)}
            </p>
          )}
          {selector}
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
                    <WeatherIcon icon={day.attributes.icon} className="size-10 shrink-0 text-foreground" />
                    <p className="text-sm text-muted-foreground flex-1 min-w-0 truncate">
                      {day.attributes.summary}
                    </p>
                    <div className="text-sm text-right shrink-0 tabular-nums">
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
      <ManageLocationsDialog open={manageOpen} onOpenChange={setManageOpen} />
    </PullToRefresh>
  );
}
