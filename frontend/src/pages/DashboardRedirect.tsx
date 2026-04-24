import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDashboards, useSeedDashboard } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
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
  const { data: prefs } = useHouseholdPreferences();
  const { data: dashboards, refetch } = useDashboards();
  const seed = useSeedDashboard();

  useEffect(() => {
    if (!prefs || !dashboards) return;
    const list = dashboards.data;
    // getPreferences returns a single-row list; normalise to the first entry.
    const prefRow = prefs.data[0];
    const pref = prefRow?.attributes.defaultDashboardId ?? null;

    if (pref && list.some((d) => d.id === pref)) {
      navigate(`/app/dashboards/${pref}`, { replace: true });
      return;
    }
    const household = list.find((d) => d.attributes.scope === "household");
    if (household) {
      navigate(`/app/dashboards/${household.id}`, { replace: true });
      return;
    }
    const user = list.find((d) => d.attributes.scope === "user");
    if (user) {
      navigate(`/app/dashboards/${user.id}`, { replace: true });
      return;
    }
    // Seed — only fire once even if the effect retriggers (mutations have
    // their own in-flight guard, but we also avoid scheduling a second call
    // while the first is still pending).
    if (seed.isPending || seed.isSuccess) return;
    seed.mutate(
      { name: "Home", layout: seedLayout() },
      {
        onSuccess: async (res) => {
          // 201: single-resource body; 200 (idempotent no-op): list body.
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

  return <DashboardSkeleton />;
}
