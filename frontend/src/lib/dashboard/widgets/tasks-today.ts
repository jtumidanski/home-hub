import { z } from "zod";
import { TasksTodayAdapter } from "@/components/features/dashboard-widgets/tasks-today-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  includeCompleted: z.boolean().default(true),
});

type Cfg = z.infer<typeof schema>;

export const tasksTodayWidget: WidgetDefinition<Cfg> = {
  type: "tasks-today",
  displayName: "Today's Tasks",
  description: "Overdue plus today's incomplete tasks",
  component: TasksTodayAdapter,
  configSchema: schema,
  defaultConfig: { includeCompleted: true },
  defaultSize: { w: 4, h: 4 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 8 },
  dataScope: "household",
};
