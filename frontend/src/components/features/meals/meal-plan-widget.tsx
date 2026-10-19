import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";
import { useTenant } from "@/context/tenant-context";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";
import { UtensilsCrossed, ChevronRight } from "lucide-react";
import type { Slot } from "@/types/models/meal-plan";

const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

export interface MealPlanWidgetProps {
  /**
   * Optional planning horizon for future dashboard configurability. Accepted
   * on the prop contract so the widget registry can pass config through,
   * but not yet wired through the existing today-only filter. Extending the
   * filter is a follow-up task.
   */
  horizonDays?: 1 | 3 | 7;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function MealPlanWidget(_props: MealPlanWidgetProps = {}) {
  const { household } = useTenant();
  const timezone = household?.attributes.timezone;
  const localDate = useLocalDate(timezone);
  const weekStart = useMemo(() => {
    const [y = 1970, m = 1, d = 1] = localDate.split("-").map(Number);
    const date = new Date(y, m - 1, d);
    date.setDate(date.getDate() + (date.getDay() === 0 ? -6 : 1 - date.getDay()));
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  }, [localDate]);
  const todayStr = localDate;

  const { data: plansData, isLoading: plansLoading, isError: plansError } = usePlans({
    starts_on: weekStart,
  });

  const planId = plansData?.data?.[0]?.id ?? null;

  const { data: planDetail, isLoading: detailLoading, isError: detailError } = usePlan(planId);

  const isLoading = plansLoading || (planId !== null && detailLoading);
  const isError = plansError || detailError;

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-28" />
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return <WidgetError message="Failed to load meal plan" />;
  }

  const todayItems = planDetail?.data?.attributes?.items
    ?.filter((item) => item.day === todayStr)
    ?.sort((a, b) => {
      const slotOrder: Slot[] = ["breakfast", "lunch", "dinner", "snack", "side"];
      return slotOrder.indexOf(a.slot) - slotOrder.indexOf(b.slot);
    }) ?? [];

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="font-heading text-lg @xl:text-2xl">Meals</CardTitle>
        <CardAction>
          <Link
            to="/app/meals"
            aria-label="Open meals"
            className="inline-flex size-6 items-center justify-center rounded-md text-muted-foreground hover:text-foreground coarse:size-11"
          >
            <ChevronRight className="size-4" aria-hidden="true" />
          </Link>
        </CardAction>
      </CardHeader>
      <CardContent>
        {todayItems.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <UtensilsCrossed className="size-5 @xl:size-6 shrink-0" aria-hidden="true" />
            <p className="text-sm">No meals planned for today</p>
          </div>
        ) : (
          <div className="space-y-2">
            {todayItems.map((item) => (
              <div key={item.id} className="flex items-baseline gap-2">
                <span className="text-xs font-medium text-muted-foreground w-16 shrink-0">
                  {SLOT_LABELS[item.slot]}
                </span>
                <Link
                  to={`/app/recipes/${item.recipe_id}`}
                  className="inline-flex min-h-6 items-center text-sm @xl:text-xl text-primary hover:underline truncate coarse:min-h-11"
                >
                  {item.recipe_title}
                </Link>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
