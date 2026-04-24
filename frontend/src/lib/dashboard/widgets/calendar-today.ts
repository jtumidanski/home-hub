import { z } from "zod";
import { CalendarAdapter } from "@/components/features/dashboard-widgets/calendar-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  horizonDays: z.union([z.literal(1), z.literal(3), z.literal(7)]).default(1),
  includeAllDay: z.boolean().default(true),
});

type Cfg = z.infer<typeof schema>;

export const calendarTodayWidget: WidgetDefinition<Cfg> = {
  type: "calendar-today",
  displayName: "Calendar",
  description: "Upcoming calendar events",
  component: CalendarAdapter,
  configSchema: schema,
  defaultConfig: { horizonDays: 1, includeAllDay: true },
  defaultSize: { w: 6, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 6 },
  dataScope: "household",
};
