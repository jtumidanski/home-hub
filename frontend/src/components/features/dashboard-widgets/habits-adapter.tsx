import { HabitsWidget } from "@/components/features/trackers/habits-widget";

export interface HabitsAdapterConfig {
  title?: string | undefined;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function HabitsAdapter({ config: _config }: { config: HabitsAdapterConfig }) {
  return <HabitsWidget />;
}
