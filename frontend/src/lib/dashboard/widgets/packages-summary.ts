import { z } from "zod";
import { PackagesSummaryAdapter } from "@/components/features/dashboard-widgets/packages-summary-adapter";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  title: z.string().max(80).optional(),
});

type Cfg = z.infer<typeof schema>;

export const packagesSummaryWidget: WidgetDefinition<Cfg> = {
  type: "packages-summary",
  displayName: "Packages",
  description: "Summary of tracked package activity",
  component: PackagesSummaryAdapter,
  configSchema: schema,
  defaultConfig: {},
  defaultSize: { w: 4, h: 3 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 12, h: 4 },
  dataScope: "household",
};
