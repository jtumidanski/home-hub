import { z } from "zod";
import { CalendarTomorrowAdapter } from "@/components/features/dashboard-widgets/calendar-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  includeAllDay: z.boolean().default(true),
  limit: z.number().int().min(1).max(10).default(5),
});

type Cfg = z.infer<typeof schema>;

export const calendarTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "calendar-tomorrow",
  displayName: "Tomorrow's Calendar",
  description: "Tomorrow's events",
  component: CalendarTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { includeAllDay: true, limit: 5 },
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 6 },
  dataScope: "household",
};
