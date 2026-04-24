import { z } from "zod";
import { MealPlanAdapter } from "@/components/features/dashboard-widgets/meal-plan-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  horizonDays: z.union([z.literal(1), z.literal(3), z.literal(7)]).default(1),
});

type Cfg = z.infer<typeof schema>;

export const mealPlanTodayWidget: WidgetDefinition<Cfg> = {
  type: "meal-plan-today",
  displayName: "Meal Plan",
  description: "Upcoming planned meals",
  component: MealPlanAdapter,
  configSchema: schema,
  defaultConfig: { horizonDays: 1 },
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 6 },
  dataScope: "household",
};
