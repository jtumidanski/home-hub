import { WorkoutWidget } from "@/components/features/workouts/workout-widget";

export interface WorkoutAdapterConfig {
  title?: string;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function WorkoutAdapter({ config: _config }: { config: WorkoutAdapterConfig }) {
  return <WorkoutWidget />;
}
