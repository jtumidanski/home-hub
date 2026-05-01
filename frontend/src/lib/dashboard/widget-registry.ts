import type { ComponentType } from "react";
import type { z } from "zod";
import type { WidgetType } from "@/lib/dashboard/widget-types";
import { weatherWidget } from "@/lib/dashboard/widgets/weather";
import { tasksSummaryWidget } from "@/lib/dashboard/widgets/tasks-summary";
import { remindersSummaryWidget } from "@/lib/dashboard/widgets/reminders-summary";
import { overdueSummaryWidget } from "@/lib/dashboard/widgets/overdue-summary";
import { mealPlanTodayWidget } from "@/lib/dashboard/widgets/meal-plan-today";
import { calendarTodayWidget } from "@/lib/dashboard/widgets/calendar-today";
import { packagesSummaryWidget } from "@/lib/dashboard/widgets/packages-summary";
import { habitsTodayWidget } from "@/lib/dashboard/widgets/habits-today";
import { workoutTodayWidget } from "@/lib/dashboard/widgets/workout-today";
import { tasksTodayWidget } from "@/lib/dashboard/widgets/tasks-today";
import { remindersTodayWidget } from "@/lib/dashboard/widgets/reminders-today";
import { weatherTomorrowWidget } from "@/lib/dashboard/widgets/weather-tomorrow";
import { calendarTomorrowWidget } from "@/lib/dashboard/widgets/calendar-tomorrow";
import { tasksTomorrowWidget } from "@/lib/dashboard/widgets/tasks-tomorrow";

export type WidgetDefinition<TConfig> = {
  type: WidgetType;
  displayName: string;
  description: string;
  component: ComponentType<{ config: TConfig }>;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize: { w: number; h: number };
  dataScope: "household" | "user";
};

export type AnyWidgetDefinition = WidgetDefinition<unknown>;

// Each widget has its own TConfig — the `as unknown as` cast is intentional
// so the array can hold heterogeneous definitions under the common
// AnyWidgetDefinition face.
export const widgetRegistry: readonly AnyWidgetDefinition[] = [
  weatherWidget,
  tasksSummaryWidget,
  remindersSummaryWidget,
  overdueSummaryWidget,
  mealPlanTodayWidget,
  calendarTodayWidget,
  packagesSummaryWidget,
  habitsTodayWidget,
  workoutTodayWidget,
  tasksTodayWidget,
  remindersTodayWidget,
  weatherTomorrowWidget,
  calendarTomorrowWidget,
  tasksTomorrowWidget,
] as unknown as readonly AnyWidgetDefinition[];

export function findWidget(type: string): AnyWidgetDefinition | undefined {
  return widgetRegistry.find((w) => w.type === type);
}
