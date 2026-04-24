import { z } from "zod";
import { WeatherWidgetAdapter } from "@/components/features/dashboard-widgets/weather-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  units: z.enum(["imperial", "metric"]).default("imperial"),
  location: z
    .object({
      lat: z.number(),
      lon: z.number(),
      label: z.string().max(200),
    })
    .nullable()
    .default(null),
});

type Cfg = z.infer<typeof schema>;

export const weatherWidget: WidgetDefinition<Cfg> = {
  type: "weather",
  displayName: "Weather",
  description: "Current conditions and forecast",
  component: WeatherWidgetAdapter,
  configSchema: schema,
  defaultConfig: { units: "imperial", location: null },
  defaultSize: { w: 12, h: 3 },
  minSize: { w: 6, h: 2 },
  maxSize: { w: 12, h: 4 },
  dataScope: "household",
};
