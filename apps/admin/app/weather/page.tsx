"use client";

import { useEffect, useState, useCallback } from "react";
import { listHouseholds, Household } from "@/lib/api/households";
import {
  getWeather,
  WeatherResponse,
  refreshWeatherCache,
  purgeWeatherCache,
  purgeAllWeatherCaches,
} from "@/lib/api/weather";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, RefreshCw, Trash2, Cloud } from "lucide-react";
import { toast } from "sonner";
import { formatTemperature, formatRelativeTime, formatForecastDate } from "@/lib/api/weather";

export default function WeatherPage() {
  const [households, setHouseholds] = useState<Household[]>([]);
  const [selectedHouseholdId, setSelectedHouseholdId] = useState<string>("");
  const [weather, setWeather] = useState<WeatherResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [householdsLoading, setHouseholdsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [purging, setPurging] = useState(false);
  const [purgingAll, setPurgingAll] = useState(false);
  const [showPurgeAllDialog, setShowPurgeAllDialog] = useState(false);
  const [showPurgeDialog, setShowPurgeDialog] = useState(false);

  // Define functions before useEffect hooks
  const fetchHouseholds = async () => {
    try {
      setHouseholdsLoading(true);
      const data = await listHouseholds();
      setHouseholds(data);
    } catch (err) {
      console.error("Failed to fetch households:", err);
      toast.error("Failed to load households");
    } finally {
      setHouseholdsLoading(false);
    }
  };

  const fetchWeather = useCallback(async () => {
    if (!selectedHouseholdId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await getWeather(selectedHouseholdId);
      setWeather(data);
    } catch (err) {
      console.error("Failed to fetch weather:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch weather data");
      setWeather(null);
    } finally {
      setLoading(false);
    }
  }, [selectedHouseholdId]);

  // Fetch households on mount
  useEffect(() => {
    fetchHouseholds();
  }, []);

  // Fetch weather when household is selected
  useEffect(() => {
    if (selectedHouseholdId) {
      fetchWeather();
    }
  }, [selectedHouseholdId, fetchWeather]);

  const handleRefresh = async () => {
    if (!selectedHouseholdId) return;

    try {
      setRefreshing(true);
      await refreshWeatherCache(selectedHouseholdId);
      toast.success("Weather cache refreshed");
      // Refetch weather data after refresh
      await fetchWeather();
    } catch (err) {
      console.error("Failed to refresh cache:", err);
      toast.error("Failed to refresh weather cache");
    } finally {
      setRefreshing(false);
    }
  };

  const handlePurge = async () => {
    if (!selectedHouseholdId) return;

    try {
      setPurging(true);
      await purgeWeatherCache(selectedHouseholdId);
      toast.success("Weather cache purged");
      setWeather(null);
      setShowPurgeDialog(false);
    } catch (err) {
      console.error("Failed to purge cache:", err);
      toast.error("Failed to purge weather cache");
    } finally {
      setPurging(false);
    }
  };

  const handlePurgeAll = async () => {
    try {
      setPurgingAll(true);
      await purgeAllWeatherCaches();
      toast.success("All weather caches purged");
      setWeather(null);
      setShowPurgeAllDialog(false);
    } catch (err) {
      console.error("Failed to purge all caches:", err);
      toast.error("Failed to purge all weather caches");
    } finally {
      setPurgingAll(false);
    }
  };

  const selectedHousehold = households.find((h) => h.id === selectedHouseholdId);
  const temperatureUnit = selectedHousehold?.temperatureUnit || "celsius";

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Weather</h1>
        <p className="text-neutral-600 dark:text-neutral-400 mt-2">
          View weather data and manage weather cache
        </p>
      </div>

      {/* Global Admin Actions - Always available */}
      <Card>
        <CardHeader>
          <CardTitle>Global Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button
              onClick={() => setShowPurgeAllDialog(true)}
              disabled={purgingAll}
              variant="destructive"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              {purgingAll ? "Purging All..." : "Purge All Caches"}
            </Button>
          </div>
          <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-3">
            Purge all caches removes all cached weather data across all households.
            This action does not require a household to be selected.
          </p>
        </CardContent>
      </Card>

      {/* Household Selector */}
      <Card>
        <CardHeader>
          <CardTitle>Select Household</CardTitle>
        </CardHeader>
        <CardContent>
          <Select
            value={selectedHouseholdId}
            onValueChange={setSelectedHouseholdId}
            disabled={householdsLoading}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select a household to view weather" />
            </SelectTrigger>
            <SelectContent>
              {households.map((household) => (
                <SelectItem key={household.id} value={household.id}>
                  {household.name}
                  {household.timezone && (
                    <span className="text-neutral-500 ml-2">
                      ({household.timezone})
                    </span>
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      {/* Household Actions */}
      {selectedHouseholdId && (
        <Card>
          <CardHeader>
            <CardTitle>Household Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              <Button
                onClick={handleRefresh}
                disabled={refreshing}
                variant="default"
              >
                <RefreshCw
                  className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
                />
                {refreshing ? "Refreshing..." : "Refresh Cache"}
              </Button>

              <Button
                onClick={() => setShowPurgeDialog(true)}
                disabled={purging}
                variant="destructive"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                {purging ? "Purging..." : "Purge Cache"}
              </Button>
            </div>
            <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-3">
              Refresh cache forces a new fetch from the weather provider. Purge
              cache removes cached data for this household.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Weather Data Display */}
      {loading && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-8 w-8 animate-spin text-neutral-400" />
              <span className="ml-3 text-neutral-600 dark:text-neutral-400">
                Loading weather data...
              </span>
            </div>
          </CardContent>
        </Card>
      )}

      {!loading && !error && weather && (
        <>
          {/* Current Weather */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Cloud className="h-5 w-5" />
                Current Weather
              </CardTitle>
            </CardHeader>
            <CardContent>
              {weather.current ? (
                <div className="space-y-4">
                  <div className="flex items-baseline gap-2">
                    <span className="text-5xl font-bold">
                      {formatTemperature(weather.current.temperature_c, temperatureUnit)}
                    </span>
                    {weather.current.stale && (
                      <span className="text-sm text-amber-600 dark:text-amber-400">
                        (Stale)
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-neutral-600 dark:text-neutral-400 space-y-1">
                    <p>
                      Observed: {formatRelativeTime(weather.current.observed_at)}
                    </p>
                    <p>
                      Age: {Math.floor(weather.current.age_seconds / 60)} minutes
                    </p>
                  </div>
                  <div className="text-xs text-neutral-500 pt-2 border-t">
                    <p>Source: {weather.meta.source}</p>
                    <p>Location: {weather.meta.geokey}</p>
                    <p>Timezone: {weather.meta.timezone}</p>
                    <p>Refreshed: {formatRelativeTime(weather.meta.refreshed_at)}</p>
                  </div>
                </div>
              ) : (
                <p className="text-neutral-600 dark:text-neutral-400">
                  No current weather data available
                </p>
              )}
            </CardContent>
          </Card>

          {/* Forecast */}
          {weather.daily && weather.daily.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>7-Day Forecast</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {weather.daily.slice(0, 7).map((day) => (
                    <div
                      key={day.date}
                      className="p-4 border rounded-lg bg-neutral-50 dark:bg-neutral-800"
                    >
                      <p className="font-medium text-sm mb-2">
                        {formatForecastDate(day.date)}
                      </p>
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-neutral-600 dark:text-neutral-400">
                          High:
                        </span>
                        <span className="font-semibold">
                          {formatTemperature(day.tmax_c, temperatureUnit)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-neutral-600 dark:text-neutral-400">
                          Low:
                        </span>
                        <span className="font-semibold">
                          {formatTemperature(day.tmin_c, temperatureUnit)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

        </>
      )}

      {/* Empty State */}
      {!loading && !error && !weather && selectedHouseholdId && (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center py-8">
              <Cloud className="h-12 w-12 mx-auto text-neutral-400 mb-3" />
              <p className="text-neutral-600 dark:text-neutral-400">
                No weather data available for this household
              </p>
              <Button onClick={handleRefresh} className="mt-4" disabled={refreshing}>
                <RefreshCw
                  className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
                />
                Fetch Weather
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {!selectedHouseholdId && !householdsLoading && (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center py-8">
              <Cloud className="h-12 w-12 mx-auto text-neutral-400 mb-3" />
              <p className="text-neutral-600 dark:text-neutral-400">
                Select a household to view weather data
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Purge Confirmation Dialog */}
      <Dialog open={showPurgeDialog} onOpenChange={setShowPurgeDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Purge Weather Cache?</DialogTitle>
            <DialogDescription>
              This will remove all cached weather data for{" "}
              <strong>{selectedHousehold?.name}</strong>. The next weather request
              will fetch fresh data from the provider.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowPurgeDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handlePurge} disabled={purging} variant="destructive">
              {purging ? "Purging..." : "Purge Cache"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Purge All Confirmation Dialog */}
      <Dialog open={showPurgeAllDialog} onOpenChange={setShowPurgeAllDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Purge All Weather Caches?</DialogTitle>
            <DialogDescription>
              This will remove all cached weather data for <strong>ALL households</strong> in the system. This is a destructive operation. Are you sure?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowPurgeAllDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handlePurgeAll}
              disabled={purgingAll}
              variant="destructive"
            >
              {purgingAll ? "Purging All..." : "Purge All Caches"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
