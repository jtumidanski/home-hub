import { z } from "zod";
import { TasksTomorrowAdapter } from "@/components/features/dashboard-widgets/tasks-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const tasksTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "tasks-tomorrow",
  displayName: "Tomorrow's Tasks",
  description: "Incomplete tasks due tomorrow",
  component: TasksTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { limit: 5 },
  defaultSize: { w: 3, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 6 },
  dataScope: "household",
};
