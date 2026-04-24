import { z } from "zod";
import { HabitsAdapter } from "@/components/features/dashboard-widgets/habits-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const habitsTodayWidget: WidgetDefinition<Cfg> = {
  type: "habits-today",
  displayName: "Habits",
  description: "Today's habit check-ins",
  component: HabitsAdapter,
  configSchema: schema,
  defaultConfig: {},
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 4 },
  dataScope: "household",
};
