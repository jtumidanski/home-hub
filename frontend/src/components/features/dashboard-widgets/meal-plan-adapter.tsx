import { MealPlanWidget } from "@/components/features/meals/meal-plan-widget";
import { MealPlanTodayDetail } from "@/components/features/meals/meal-plan-today-detail";

export interface MealPlanAdapterConfig {
  horizonDays: 1 | 3 | 7;
  view?: "list" | "today-detail";
}

/**
 * Registry adapter around the meal plan widgets. Branches on `view`:
 * - "list" (default): renders the original today-list MealPlanWidget.
 * - "today-detail": renders MealPlanTodayDetail with full B/L/D for today
 *   plus dinners for the next horizonDays-1 days.
 */
export function MealPlanAdapter({ config }: { config: MealPlanAdapterConfig }) {
  if (config.view === "today-detail") {
    return <MealPlanTodayDetail horizonDays={config.horizonDays} />;
  }
  return <MealPlanWidget horizonDays={config.horizonDays} />;
}
