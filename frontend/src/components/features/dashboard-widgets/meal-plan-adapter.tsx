import { MealPlanWidget } from "@/components/features/meals/meal-plan-widget";

export interface MealPlanAdapterConfig {
  horizonDays: 1 | 3 | 7;
}

/**
 * Registry adapter around MealPlanWidget. `horizonDays` is accepted on
 * the registry config and forwarded to the widget — currently the widget
 * still renders today-only; wiring horizonDays through the meal plan
 * filter is a later task.
 */
export function MealPlanAdapter({ config }: { config: MealPlanAdapterConfig }) {
  return <MealPlanWidget horizonDays={config.horizonDays} />;
}
