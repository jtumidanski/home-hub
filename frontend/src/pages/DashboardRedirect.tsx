import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDashboards, useSeedDashboard } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { useTenant } from "@/context/tenant-context";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { getErrorMessage } from "@/lib/api/errors";
import type { Dashboard } from "@/types/models/dashboard";

/**
 * Resolves `/app/dashboard` (and `/app/dashboards`) to a concrete dashboard.
 * Resolution order:
 *   1. household_preferences.defaultDashboardId, if it still exists.
 *   2. first household-scoped dashboard.
 *   3. first user-scoped dashboard.
 *   4. seed a first-run dashboard.
 */
export function DashboardRedirect() {
  const navigate = useNavigate();
  const { tenant, household } = useTenant();
  const prefsQuery = useHouseholdPreferences();
  const dashboardsQuery = useDashboards();
  const seed = useSeedDashboard();

  const { data: prefs, isError: prefsErr, error: prefsError } = prefsQuery;
  const { data: dashboards, refetch, isError: listErr, error: listError } = dashboardsQuery;

  useEffect(() => {
    if (!prefs || !dashboards) return;
    const list = dashboards.data;
    const prefRow = prefs.data[0];
    const pref = prefRow?.attributes.defaultDashboardId ?? null;

    if (pref && list.some((d) => d.id === pref)) {
      navigate(`/app/dashboards/${pref}`, { replace: true });
      return;
    }
    const householdDash = list.find((d) => d.attributes.scope === "household");
    if (householdDash) {
      navigate(`/app/dashboards/${householdDash.id}`, { replace: true });
      return;
    }
    const user = list.find((d) => d.attributes.scope === "user");
    if (user) {
      navigate(`/app/dashboards/${user.id}`, { replace: true });
      return;
    }
    if (seed.isPending || seed.isSuccess) return;
    seed.mutate(
      { name: "Home", layout: seedLayout() },
      {
        onSuccess: async (res) => {
          if (!Array.isArray(res.data)) {
            const single = res.data as Dashboard;
            navigate(`/app/dashboards/${single.id}`, { replace: true });
            return;
          }
          const first = res.data[0] ?? (await refetch()).data?.data?.[0];
          if (first) navigate(`/app/dashboards/${first.id}`, { replace: true });
        },
      },
    );
  }, [prefs, dashboards, navigate, seed, refetch]);

  if (!tenant?.id || !household?.id) {
    return <DashboardMessage title="No household selected" body="Pick a household to view its dashboard." />;
  }
  if (prefsErr || listErr || seed.isError) {
    const msg =
      getErrorMessage(prefsError, "") ||
      getErrorMessage(listError, "") ||
      getErrorMessage(seed.error, "") ||
      "The dashboard service is unavailable.";
    return <DashboardMessage title="Couldn't load dashboards" body={msg} />;
  }
  return <DashboardSkeleton />;
}

function DashboardMessage({ title, body }: { title: string; body: string }) {
  return (
    <div className="p-8 max-w-xl mx-auto">
      <h1 className="text-xl font-semibold mb-2">{title}</h1>
      <p className="text-muted-foreground">{body}</p>
    </div>
  );
}
