import { z } from "zod";
import { TasksSummaryWidget } from "@/components/features/dashboard-widgets/tasks-summary";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  status: z.enum(["pending", "overdue", "completed"]),
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const tasksSummaryWidget: WidgetDefinition<Cfg> = {
  type: "tasks-summary",
  displayName: "Tasks Summary",
  description: "Count of pending, overdue, or completed tasks",
  component: TasksSummaryWidget,
  configSchema: schema,
  defaultConfig: { status: "pending" },
  defaultSize: { w: 4, h: 2 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 2 },
  dataScope: "household",
};
