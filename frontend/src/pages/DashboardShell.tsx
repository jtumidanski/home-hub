import { Link, Outlet, useParams } from "react-router-dom";
import { useDashboard } from "@/lib/hooks/api/use-dashboards";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { Error404Page } from "@/components/common/error-page";
import { Button } from "@/components/ui/button";
import { LAYOUT_SCHEMA_VERSION } from "@/lib/dashboard/widget-types";
import { useMobile } from "@/lib/hooks/use-mobile";

export function DashboardShell() {
  const { dashboardId } = useParams<{ dashboardId: string }>();
  const { data, isLoading, isError } = useDashboard(dashboardId ?? null);
  const isMobile = useMobile();

  if (isLoading) return <DashboardSkeleton />;
  if (isError || !data?.data) return <Error404Page />;

  const dashboard = data.data;

  if (dashboard.attributes.schemaVersion > LAYOUT_SCHEMA_VERSION) {
    return (
      <div
        role="alert"
        className="flex min-h-[50vh] items-center justify-center p-6 text-center"
      >
        <div className="max-w-md space-y-2">
          <h2 className="text-lg font-semibold">Dashboard needs a newer app version</h2>
          <p className="text-sm text-muted-foreground">
            This dashboard was saved with a newer layout schema than this app supports.
            Please refresh your browser to load the latest version.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between p-4 md:px-6 md:pt-6">
        <h1 className="text-xl md:text-2xl font-semibold">{dashboard.attributes.name}</h1>
        {isMobile ? (
          <Button
            variant="outline"
            size="sm"
            disabled
            title="Editing requires a larger screen"
            data-testid="dashboard-shell-edit-disabled"
          >
            Edit
          </Button>
        ) : (
          <Button
            render={<Link to="./edit" />}
            variant="outline"
            size="sm"
            data-testid="dashboard-shell-edit"
          >
            Edit
          </Button>
        )}
      </div>
      <Outlet context={{ dashboard }} />
    </div>
  );
}
