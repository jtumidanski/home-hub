"use client";

import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  HelpCircle,
  RefreshCw,
  Activity,
} from "lucide-react";
import {
  getAllServiceHealth,
  aggregateHealthStatus,
  getStatusMessage,
  formatUptime,
  formatRelativeTime,
  type ServiceHealth,
  type HealthStatus,
} from "@/lib/api/health";

interface SystemStatusCardProps {
  /**
   * Auto-refresh interval in milliseconds (default: 30000ms = 30 seconds)
   */
  refreshInterval?: number;

  /**
   * Enable auto-refresh (default: true)
   */
  autoRefresh?: boolean;
}

/**
 * System Status Card Component
 *
 * Displays the aggregate health status of all microservices.
 * Hover over the status badge to see individual service details.
 */
export function SystemStatusCard({
  refreshInterval = 30000,
  autoRefresh = true,
}: SystemStatusCardProps) {
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  // Fetch health data
  const fetchHealthData = async () => {
    try {
      setRefreshing(true);
      setError(null);

      const healthData = await getAllServiceHealth();
      setServices(healthData);
      setLastUpdated(new Date());
    } catch (err) {
      console.error("Failed to fetch health data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch health data");
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // Initial fetch
  useEffect(() => {
    fetchHealthData();
  }, []);

  // Auto-refresh effect
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(fetchHealthData, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  // Manual refresh handler
  const handleRefresh = () => {
    fetchHealthData();
  };

  // Calculate aggregate status
  const aggregateStatus = aggregateHealthStatus(services);

  // Get status icon component
  const getStatusIcon = (status: HealthStatus, size: "sm" | "lg" = "sm") => {
    const className = size === "lg" ? "h-5 w-5" : "h-4 w-4";

    switch (status) {
      case "healthy":
        return <CheckCircle2 className={`${className} text-green-600 dark:text-green-400`} />;
      case "degraded":
        return <AlertTriangle className={`${className} text-yellow-600 dark:text-yellow-400`} />;
      case "unhealthy":
        return <XCircle className={`${className} text-red-600 dark:text-red-400`} />;
      case "unknown":
        return <HelpCircle className={`${className} text-neutral-600 dark:text-neutral-400`} />;
    }
  };

  // Get status color for text
  const getStatusColor = (status: HealthStatus) => {
    switch (status) {
      case "healthy":
        return "text-green-600 dark:text-green-400";
      case "degraded":
        return "text-yellow-600 dark:text-yellow-400";
      case "unhealthy":
        return "text-red-600 dark:text-red-400";
      case "unknown":
        return "text-neutral-600 dark:text-neutral-400";
    }
  };

  // Get status text
  const getStatusText = (status: HealthStatus) => {
    switch (status) {
      case "healthy":
        return "Healthy";
      case "degraded":
        return "Degraded";
      case "unhealthy":
        return "Unhealthy";
      case "unknown":
        return "Unknown";
    }
  };

  // Loading state
  if (loading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">System Status</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-neutral-400">...</div>
          <p className="text-xs text-muted-foreground">Loading status</p>
        </CardContent>
      </Card>
    );
  }

  // Error state
  if (error && services.length === 0) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">System Status</CardTitle>
          <Button
            variant="ghost"
            size="icon"
            className="h-4 w-4"
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw
              className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
            />
          </Button>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-600">Error</div>
          <p className="text-xs text-muted-foreground">{error}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">System Status</CardTitle>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-4 w-4"
                onClick={handleRefresh}
                disabled={refreshing}
              >
                <RefreshCw
                  className={`h-3 w-3 ${refreshing ? "animate-spin" : ""}`}
                />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">
                {lastUpdated
                  ? `Updated ${formatRelativeTime(lastUpdated)}`
                  : "Click to refresh"}
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </CardHeader>
      <CardContent>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="cursor-help">
                <div className={`text-2xl font-bold ${getStatusColor(aggregateStatus)}`}>
                  {getStatusText(aggregateStatus)}
                </div>
                <p className="text-xs text-muted-foreground">
                  {getStatusMessage(aggregateStatus)}
                </p>
              </div>
            </TooltipTrigger>
            <TooltipContent side="bottom" className="w-80 bg-white dark:bg-neutral-950 border-neutral-200 dark:border-neutral-800">
              <div className="space-y-3 p-2">
                <div className="flex items-center justify-between border-b border-neutral-200 dark:border-neutral-800 pb-2">
                  <span className="text-sm font-semibold text-neutral-900 dark:text-neutral-100">Service Details</span>
                  <span className="text-xs text-neutral-600 dark:text-neutral-400">
                    {services.filter((s) => s.available).length}/{services.length} available
                  </span>
                </div>
                <div className="space-y-2">
                  {services.map((service) => (
                    <div
                      key={service.service}
                      className="flex items-start justify-between gap-3 rounded-md bg-neutral-100 dark:bg-neutral-900 p-2 border border-neutral-200 dark:border-neutral-800"
                    >
                      <div className="flex items-start gap-2 flex-1 min-w-0">
                        <div className="mt-0.5">
                          {service.available && service.health
                            ? getStatusIcon(service.health.status)
                            : getStatusIcon("unknown")}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate text-neutral-900 dark:text-neutral-100">
                            {service.service}
                          </p>
                          {service.available && service.health ? (
                            <p className="text-xs text-neutral-600 dark:text-neutral-400">
                              v{service.health.version} • {formatUptime(service.health.uptime_seconds)}
                            </p>
                          ) : (
                            <p className="text-xs text-red-600 dark:text-red-400">
                              {service.error || "Unavailable"}
                            </p>
                          )}
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-1 shrink-0">
                        {service.available && service.health && (
                          <>
                            <Badge
                              variant="outline"
                              className="text-xs h-5"
                            >
                              {service.health.status}
                            </Badge>
                            {service.responseTime && (
                              <span className="text-xs text-neutral-600 dark:text-neutral-400">
                                {Math.round(service.responseTime)}ms
                              </span>
                            )}
                          </>
                        )}
                        {!service.available && (
                          <Badge variant="outline" className="text-xs h-5 bg-neutral-100 dark:bg-neutral-800">
                            offline
                          </Badge>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </CardContent>
    </Card>
  );
}
