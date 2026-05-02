import { z } from "zod";
import { WeatherTomorrowAdapter } from "@/components/features/dashboard-widgets/weather-tomorrow-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  units: z.enum(["imperial", "metric"]).nullable().default(null),
});

type Cfg = z.infer<typeof schema>;

export const weatherTomorrowWidget: WidgetDefinition<Cfg> = {
  type: "weather-tomorrow",
  displayName: "Tomorrow's Weather",
  description: "Tomorrow's high and low",
  component: WeatherTomorrowAdapter,
  configSchema: schema,
  defaultConfig: { units: null },
  defaultSize: { w: 3, h: 2 },
  minSize: { w: 2, h: 2 },
  maxSize: { w: 6, h: 3 },
  dataScope: "household",
};
