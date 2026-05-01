// Read-only widget — no mutations. See PRD §4.2.
import { Link } from "react-router-dom";
import { useReminders } from "@/lib/hooks/api/use-reminders";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Bell, ChevronRight } from "lucide-react";
import type { Reminder } from "@/types/models/reminder";

export interface RemindersTodayConfig {
  title?: string | undefined;
  limit: number;
}

function formatRelative(iso: string): string {
  const ms = new Date(iso).getTime() - Date.now();
  if (Math.abs(ms) < 60_000) return "Now";
  const mins = Math.round(ms / 60_000);
  if (Math.abs(mins) < 60) return ms > 0 ? `in ${mins} min` : `${-mins} min ago`;
  const hours = Math.round(mins / 60);
  if (Math.abs(hours) < 24) return ms > 0 ? `in ${hours}h` : `${-hours}h ago`;
  const days = Math.round(hours / 24);
  return ms > 0 ? `in ${days}d` : `${-days}d ago`;
}

export function RemindersTodayAdapter({ config }: { config: RemindersTodayConfig }) {
  const { data, isLoading, isError } = useReminders();
  const title = config.title?.trim() || "Active Reminders";

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-32" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load reminders</p></CardContent>
      </Card>
    );
  }

  const reminders = ((data?.data ?? []) as Reminder[])
    .filter((r) => r.attributes.active)
    .sort((a, b) => a.attributes.scheduledFor.localeCompare(b.attributes.scheduledFor))
    .slice(0, config.limit);

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/reminders" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/reminders"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {reminders.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Bell className="h-5 w-5" />
            <p className="text-sm">No active reminders</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {reminders.map((r) => (
              <li key={r.id} className="flex items-baseline justify-between gap-2 text-sm">
                <span className="truncate">{r.attributes.title}</span>
                <span className="text-xs text-muted-foreground shrink-0">{formatRelative(r.attributes.scheduledFor)}</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
