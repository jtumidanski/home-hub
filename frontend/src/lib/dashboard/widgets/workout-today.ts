import { z } from "zod";
import { WorkoutAdapter } from "@/components/features/dashboard-widgets/workout-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const workoutTodayWidget: WidgetDefinition<Cfg> = {
  type: "workout-today",
  displayName: "Workout",
  description: "Today's workout plan",
  component: WorkoutAdapter,
  configSchema: schema,
  defaultConfig: {},
  defaultSize: { w: 6, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 4 },
  dataScope: "household",
};
