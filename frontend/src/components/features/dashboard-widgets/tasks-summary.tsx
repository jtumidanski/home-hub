import { Link } from "react-router-dom";
import { useTaskSummary } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckSquare, AlertTriangle } from "lucide-react";

export interface TasksSummaryConfig {
  status: "pending" | "overdue" | "completed";
  title?: string;
}

const DEFAULT_TITLES: Record<TasksSummaryConfig["status"], string> = {
  pending: "Pending Tasks",
  overdue: "Overdue",
  completed: "Completed",
};

const LINK_BY_STATUS: Record<TasksSummaryConfig["status"], string> = {
  pending: "/app/tasks?status=pending",
  overdue: "/app/tasks?status=overdue",
  completed: "/app/tasks?status=completed",
};

export function TasksSummaryWidget({ config }: { config: TasksSummaryConfig }) {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading } = useTaskSummary(today);

  const title = config.title && config.title.trim().length > 0 ? config.title : DEFAULT_TITLES[config.status];
  const Icon = config.status === "overdue" ? AlertTriangle : CheckSquare;

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-24" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-12 mb-1" />
          <Skeleton className="h-3 w-32" />
        </CardContent>
      </Card>
    );
  }

  const summary = data?.data?.attributes;
  const count = pickCount(summary, config.status);
  const sub = pickSub(summary, config.status);

  return (
    <Link to={LINK_BY_STATUS[config.status]} className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          <Icon className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{count ?? "-"}</div>
          {sub && <p className="text-xs text-muted-foreground">{sub}</p>}
        </CardContent>
      </Card>
    </Link>
  );
}

function pickCount(
  summary: { pendingCount?: number; overdueCount?: number; completedTodayCount?: number } | undefined,
  status: TasksSummaryConfig["status"],
): number | undefined {
  if (!summary) return undefined;
  if (status === "pending") return summary.pendingCount;
  if (status === "overdue") return summary.overdueCount;
  return summary.completedTodayCount;
}

function pickSub(
  summary: { pendingCount?: number; overdueCount?: number; completedTodayCount?: number } | undefined,
  status: TasksSummaryConfig["status"],
): string | null {
  if (!summary) return null;
  if (status === "pending") return `${summary.completedTodayCount ?? 0} completed today`;
  if (status === "overdue") return null;
  return null;
}
