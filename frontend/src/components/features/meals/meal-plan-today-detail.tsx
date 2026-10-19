import { Link } from "react-router-dom";
import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";
import { useTenant } from "@/context/tenant-context";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";
import { UtensilsCrossed, ChevronRight } from "lucide-react";
import { getLocalWeekStart, getLocalTodayStr, getLocalDateStrOffset, formatShortWeekday } from "@/lib/date-utils";
import type { Slot } from "@/types/models/meal-plan";

const SLOT_ORDER: Slot[] = ["breakfast", "lunch", "dinner", "snack", "side"];
const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

export interface MealPlanTodayDetailProps {
  horizonDays: 1 | 3 | 7;
}

export function MealPlanTodayDetail({ horizonDays }: MealPlanTodayDetailProps) {
  const { household } = useTenant();
  const tz = household?.attributes.timezone;
  const monday = getLocalWeekStart(tz);
  const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, "0")}-${String(monday.getDate()).padStart(2, "0")}`;
  const todayStr = getLocalTodayStr(tz);

  const { data: plansData, isLoading: plansLoading, isError: plansError } = usePlans({ starts_on: weekStart });
  const planId = plansData?.data?.[0]?.id ?? null;
  const { data: planDetail, isLoading: detailLoading, isError: detailError } = usePlan(planId);

  if (plansLoading || (planId !== null && detailLoading)) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-28" data-slot="skeleton" />
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (plansError || detailError) {
    return <WidgetError message="Failed to load meal plan" />;
  }

  const items = planDetail?.data?.attributes?.items ?? [];
  const today = items
    .filter((i) => i.day === todayStr)
    .sort((a, b) => SLOT_ORDER.indexOf(a.slot) - SLOT_ORDER.indexOf(b.slot));

  const followUps: Array<{ day: string; item: (typeof items)[number] | null }> = [];
  for (let n = 1; n < horizonDays; n++) {
    const day = getLocalDateStrOffset(tz, n);
    const dinner = items.find((i) => i.day === day && i.slot === "dinner") ?? null;
    followUps.push({ day, item: dinner });
  }

  return (
    <Link to="/app/meals" className="@container group/meals block h-full rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="@container h-full motion-safe:transition-colors group-hover/meals:bg-accent">
        <CardHeader>
          <CardTitle className="font-heading text-lg @xl:text-2xl">Meals</CardTitle>
          <CardAction>
            <ChevronRight className="size-4 @xl:size-6 text-muted-foreground" aria-hidden="true" />
          </CardAction>
        </CardHeader>
        <CardContent className="space-y-3">
          <section>
            <h4 className="text-xs font-medium text-muted-foreground">Today</h4>
            {today.length === 0 ? (
              <div className="flex items-center gap-2 text-muted-foreground mt-1">
                <UtensilsCrossed className="size-4 @xl:size-6 shrink-0" aria-hidden="true" />
                <p className="text-sm @xl:text-xl">No meals planned</p>
              </div>
            ) : (
              <ul className="mt-1 space-y-1">
                {today.map((item) => (
                  <li key={item.id} className="flex items-baseline gap-2">
                    <span className="text-xs font-medium text-muted-foreground w-16 shrink-0">
                      {SLOT_LABELS[item.slot]}
                    </span>
                    <span className="text-sm @xl:text-xl truncate">{item.recipe_title}</span>
                  </li>
                ))}
              </ul>
            )}
          </section>
          {followUps.length > 0 && (
            <section>
              <h4 className="text-xs font-medium text-muted-foreground">
                Next {followUps.length} {followUps.length === 1 ? "day" : "days"}
              </h4>
              <ul className="mt-1 space-y-1">
                {followUps.map((f) => (
                  <li key={f.day} className="flex items-baseline gap-2">
                    <span className="text-xs font-medium text-muted-foreground w-16 shrink-0">
                      {formatShortWeekday(f.day)}
                    </span>
                    <span className="text-sm @xl:text-xl truncate">{f.item?.recipe_title ?? "—"}</span>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
