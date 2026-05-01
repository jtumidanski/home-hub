import { z } from "zod";
import { RemindersTodayAdapter } from "@/components/features/dashboard-widgets/reminders-today-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const remindersTodayWidget: WidgetDefinition<Cfg> = {
  type: "reminders-today",
  displayName: "Active Reminders",
  description: "List of currently active reminders",
  component: RemindersTodayAdapter,
  configSchema: schema,
  defaultConfig: { limit: 5 },
  defaultSize: { w: 3, h: 4 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 8 },
  dataScope: "household",
};
