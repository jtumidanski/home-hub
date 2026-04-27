import { Link } from "react-router-dom";
import { useReminderSummary } from "@/lib/hooks/api/use-reminders";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Bell } from "lucide-react";

export interface RemindersSummaryConfig {
  filter: "active" | "snoozed" | "upcoming";
  title?: string | undefined;
}

const DEFAULT_TITLES: Record<RemindersSummaryConfig["filter"], string> = {
  active: "Active Reminders",
  snoozed: "Snoozed Reminders",
  upcoming: "Upcoming Reminders",
};

const LINK_BY_FILTER: Record<RemindersSummaryConfig["filter"], string> = {
  active: "/app/reminders?status=active",
  snoozed: "/app/reminders?status=snoozed",
  upcoming: "/app/reminders?status=upcoming",
};

export function RemindersSummaryWidget({ config }: { config: RemindersSummaryConfig }) {
  const { data, isLoading } = useReminderSummary();

  const title = config.title && config.title.trim().length > 0 ? config.title : DEFAULT_TITLES[config.filter];

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
  const count = pickCount(summary, config.filter);
  const sub = pickSub(summary, config.filter);

  return (
    <Link to={LINK_BY_FILTER[config.filter]} className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          <Bell className="h-4 w-4 text-muted-foreground" />
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
  summary: { dueNowCount?: number; snoozedCount?: number; upcomingCount?: number } | undefined,
  filter: RemindersSummaryConfig["filter"],
): number | undefined {
  if (!summary) return undefined;
  if (filter === "active") return summary.dueNowCount;
  if (filter === "snoozed") return summary.snoozedCount;
  return summary.upcomingCount;
}

function pickSub(
  summary: { dueNowCount?: number; snoozedCount?: number; upcomingCount?: number } | undefined,
  filter: RemindersSummaryConfig["filter"],
): string | null {
  if (!summary) return null;
  if (filter === "active") return `${summary.upcomingCount ?? 0} upcoming`;
  return null;
}
