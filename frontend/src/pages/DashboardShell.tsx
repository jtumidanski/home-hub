import { Link, Outlet, useParams } from "react-router-dom";
import { useDashboard } from "@/lib/hooks/api/use-dashboards";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { Error404Page } from "@/components/common/error-page";
import { AlertTriangle } from "lucide-react";
import { useMemo } from "react";
import { EmptyState } from "@/components/common/empty-state";
import { PageHeader } from "@/components/common/page-header";
import { useTenant } from "@/context/tenant-context";
import { Button } from "@/components/ui/button";
import { LAYOUT_SCHEMA_VERSION } from "@/lib/dashboard/widget-types";
import { useAuth } from "@/components/providers/auth-provider";
import { useMobile } from "@/lib/hooks/use-mobile";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { useTimeOfDay } from "@/lib/hooks/use-time-of-day";

export function DashboardShell() {
  const { dashboardId } = useParams<{ dashboardId: string }>();
  const { data, isLoading, isError } = useDashboard(dashboardId ?? null);
  const isMobile = useMobile();
  const { household } = useTenant();
  const timezone = household?.attributes.timezone || undefined;

  const { user } = useAuth();
  const givenName = user?.attributes.givenName?.trim();
  const firstName = givenName || (user?.attributes.displayName?.trim().split(/\s+/)[0] ?? "");
  const localDate = useLocalDate(timezone);
  const timeOfDay = useTimeOfDay(timezone);

  const greeting = useMemo(() => {
    const salutation = `Good ${timeOfDay}`;
    return firstName ? `${salutation}, ${firstName}` : salutation;
  }, [firstName, timeOfDay]);
  const dateLabel = useMemo(() => {
    const [y = 1970, m = 1, d = 1] = localDate.split("-").map(Number);
    return new Date(Date.UTC(y, m - 1, d, 12)).toLocaleDateString("en-US", {
      weekday: "long",
      day: "numeric",
      month: "long",
      timeZone: "UTC",
    });
  }, [localDate]);

  if (isLoading) return <DashboardSkeleton />;
  if (isError || !data?.data) return <Error404Page />;

  const dashboard = data.data;

  if (dashboard.attributes.schemaVersion > LAYOUT_SCHEMA_VERSION) {
    return (
      <div role="alert" className="p-4 md:p-6">
        <EmptyState
          icon={AlertTriangle}
          title="Dashboard needs a newer app version"
          description="This dashboard was saved with a newer layout schema than this app supports. Please refresh your browser to load the latest version."
        />
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <div className="p-4 md:p-6">
        <PageHeader
          title={greeting}
          description={
            <>
              {dateLabel}
              <span aria-hidden="true"> · </span>
              <span>{dashboard.attributes.name}</span>
            </>
          }
          actions={
            isMobile ? (
              <p
                className="text-sm text-muted-foreground"
                data-testid="dashboard-shell-edit-disabled"
              >
                Editing needs a larger screen
              </p>
            ) : (
              <Button
                render={<Link to="./edit" />}
                variant="outline"
                size="sm"
                data-testid="dashboard-shell-edit"
              >
                Edit
              </Button>
            )
          }
        />
      </div>
      <Outlet context={{ dashboard }} />
    </div>
  );
}
