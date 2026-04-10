import { Link } from "react-router-dom";
import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { UtensilsCrossed, ChevronRight } from "lucide-react";
import type { Slot } from "@/types/models/meal-plan";

const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

function getMonday(): string {
  const today = new Date();
  const day = today.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  const monday = new Date(today);
  monday.setDate(today.getDate() + diff);
  const year = monday.getFullYear();
  const month = String(monday.getMonth() + 1).padStart(2, "0");
  const d = String(monday.getDate()).padStart(2, "0");
  return `${year}-${month}-${d}`;
}

function getTodayStr(): string {
  const today = new Date();
  const year = today.getFullYear();
  const month = String(today.getMonth() + 1).padStart(2, "0");
  const day = String(today.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function MealPlanWidget() {
  const weekStart = getMonday();
  const todayStr = getTodayStr();

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
    return (
      <Card className="border-destructive">
        <CardContent className="py-3">
          <p className="text-sm text-destructive">Failed to load meal plan</p>
        </CardContent>
      </Card>
    );
  }

  const todayItems = planDetail?.data?.attributes?.items
    ?.filter((item) => item.day === todayStr)
    ?.sort((a, b) => {
      const slotOrder: Slot[] = ["breakfast", "lunch", "dinner", "snack", "side"];
      return slotOrder.indexOf(a.slot) - slotOrder.indexOf(b.slot);
    }) ?? [];

  return (
    <Link to="/app/meals" className="transition-opacity hover:opacity-80">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Meals</CardTitle>
          <CardAction>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {todayItems.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <UtensilsCrossed className="h-5 w-5" />
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
                    className="text-sm text-primary hover:underline truncate"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {item.recipe_title}
                  </Link>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
