import { z } from "zod";
import { OverdueSummaryWidget } from "@/components/features/dashboard-widgets/overdue-summary";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const overdueSummaryWidget: WidgetDefinition<Cfg> = {
  type: "overdue-summary",
  displayName: "Overdue Tasks",
  description: "Count of overdue tasks",
  component: OverdueSummaryWidget,
  configSchema: schema,
  defaultConfig: {},
  defaultSize: { w: 4, h: 2 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 2 },
  dataScope: "household",
};
