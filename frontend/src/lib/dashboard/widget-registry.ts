import type { ComponentType } from "react";
import type { z } from "zod";
import type { WidgetType } from "@/lib/dashboard/widget-types";

export type WidgetDefinition<TConfig> = {
  type: WidgetType;
  displayName: string;
  description: string;
  component: ComponentType<{ config: TConfig }>;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize: { w: number; h: number };
  dataScope: "household" | "user";
};

export type AnyWidgetDefinition = WidgetDefinition<unknown>;
