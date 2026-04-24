import { z } from "zod";
import { GRID_COLUMNS, LAYOUT_SCHEMA_VERSION, MAX_WIDGETS } from "@/lib/dashboard/widget-types";

export const widgetInstanceSchema = z
  .object({
    id: z.string().uuid(),
    type: z.string(),
    x: z.number().int().nonnegative(),
    y: z.number().int().nonnegative(),
    w: z.number().int().min(1),
    h: z.number().int().min(1),
    config: z.record(z.string(), z.unknown()).default({}),
  })
  .refine((w) => w.x + w.w <= GRID_COLUMNS, {
    message: "x + w must be <= 12",
    path: ["w"],
  });

export type WidgetInstance = z.infer<typeof widgetInstanceSchema>;

export const layoutSchema = z.object({
  version: z.literal(LAYOUT_SCHEMA_VERSION),
  widgets: z.array(widgetInstanceSchema).max(MAX_WIDGETS),
});

export type Layout = z.infer<typeof layoutSchema>;
