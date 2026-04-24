import { z } from "zod";
import { RemindersSummaryWidget } from "@/components/features/dashboard-widgets/reminders-summary";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  filter: z.enum(["active", "snoozed", "upcoming"]),
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const remindersSummaryWidget: WidgetDefinition<Cfg> = {
  type: "reminders-summary",
  displayName: "Reminders Summary",
  description: "Count of reminders filtered by state",
  component: RemindersSummaryWidget,
  configSchema: schema,
  defaultConfig: { filter: "active" },
  defaultSize: { w: 4, h: 2 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 2 },
  dataScope: "household",
};
