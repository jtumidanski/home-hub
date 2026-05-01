import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useDashboards, useSeedDashboard } from "@/lib/hooks/api/use-dashboards";
import {
  useHouseholdPreferences,
  useMarkKioskSeeded,
} from "@/lib/hooks/api/use-household-preferences";
import { useTenant } from "@/context/tenant-context";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { getErrorMessage } from "@/lib/api/errors";

/**
 * Resolves `/app/dashboard` (and `/app/dashboards`) to a concrete dashboard.
 *
 * On first visit, seeds two household dashboards in parallel:
 *   - "Home"   — the standard composition (seedLayout).
 *   - "Kiosk"  — the kiosk composition (kioskSeedLayout), gated by
 *                household_preferences.kioskDashboardSeeded so users who
 *                later delete the Kiosk dashboard don't get it re-seeded.
 *
 * Resolution order after seeding:
 *   1. household_preferences.defaultDashboardId, if it still exists.
 *   2. first household-scoped dashboard.
 *   3. first user-scoped dashboard.
 */
export function DashboardRedirect() {
  const navigate = useNavigate();
  const { tenant, household } = useTenant();
  const prefsQuery = useHouseholdPreferences();
  const dashboardsQuery = useDashboards();
  const homeSeed = useSeedDashboard();
  const kioskSeed = useSeedDashboard();
  const markKioskSeeded = useMarkKioskSeeded();
  const seededOnce = useRef(false);

  const { data: prefs, isError: prefsErr, error: prefsError } = prefsQuery;
  const { data: dashboards, refetch, isError: listErr, error: listError } = dashboardsQuery;

  useEffect(() => {
    if (!prefs || !dashboards) return;
    if (seededOnce.current) return;
    seededOnce.current = true;

    const list = dashboards.data;
    const prefRow = prefs.data[0];
    const kioskFlag = prefRow?.attributes.kioskDashboardSeeded ?? false;
    const prefId = prefRow?.id ?? null;

    const homeNeeded = !list.some((d) => d.attributes.scope === "household");
    const hasKioskRow = list.some(
      (d) => d.attributes.scope === "household" && d.attributes.name === "Kiosk",
    );
    const kioskNeeded = !kioskFlag && !hasKioskRow;

    const seedHome = homeNeeded
      ? homeSeed.mutateAsync({ name: "Home", layout: seedLayout(), key: "home" })
      : Promise.resolve(null);
    const seedKiosk = kioskNeeded
      ? kioskSeed.mutateAsync({ name: "Kiosk", layout: kioskSeedLayout(), key: "kiosk" })
      : Promise.resolve(null);

    Promise.allSettled([seedHome, seedKiosk]).then(async () => {
      const refreshed =
        homeNeeded || kioskNeeded ? (await refetch()).data?.data ?? list : list;
      const kioskNow = refreshed.some(
        (d) => d.attributes.scope === "household" && d.attributes.name === "Kiosk",
      );
      if (kioskNow && !kioskFlag && prefId) {
        markKioskSeeded.mutate(prefId);
      }
      const pref = prefRow?.attributes.defaultDashboardId ?? null;
      if (pref && refreshed.some((d) => d.id === pref)) {
        navigate(`/app/dashboards/${pref}`, { replace: true });
        return;
      }
      const householdDash = refreshed.find((d) => d.attributes.scope === "household");
      if (householdDash) {
        navigate(`/app/dashboards/${householdDash.id}`, { replace: true });
        return;
      }
      const userDash = refreshed.find((d) => d.attributes.scope === "user");
      if (userDash) {
        navigate(`/app/dashboards/${userDash.id}`, { replace: true });
        return;
      }
    });
  }, [prefs, dashboards, navigate, homeSeed, kioskSeed, markKioskSeeded, refetch]);

  if (!tenant?.id || !household?.id) {
    return (
      <DashboardMessage
        title="No household selected"
        body="Pick a household to view its dashboard."
      />
    );
  }
  if (prefsErr || listErr || homeSeed.isError || kioskSeed.isError) {
    const msg =
      getErrorMessage(prefsError, "") ||
      getErrorMessage(listError, "") ||
      getErrorMessage(homeSeed.error, "") ||
      getErrorMessage(kioskSeed.error, "") ||
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
