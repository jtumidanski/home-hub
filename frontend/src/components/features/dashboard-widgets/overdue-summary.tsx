import { TasksSummaryWidget } from "@/components/features/dashboard-widgets/tasks-summary";

export interface OverdueSummaryConfig {
  title?: string | undefined;
}

/**
 * Thin wrapper around TasksSummaryWidget with status pinned to "overdue".
 * Kept as its own widget type so users can add a dedicated overdue tile
 * without also exposing the status picker on the tasks-summary widget.
 */
export function OverdueSummaryWidget({ config }: { config: OverdueSummaryConfig }) {
  return (
    <TasksSummaryWidget
      config={{
        status: "overdue",
        ...(config.title !== undefined ? { title: config.title } : {}),
      }}
    />
  );
}
