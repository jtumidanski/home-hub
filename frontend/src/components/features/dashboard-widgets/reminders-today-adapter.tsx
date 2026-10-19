// Read-only widget — no mutations. See PRD §4.2.
import { Link } from "react-router-dom";
import { useReminders } from "@/lib/hooks/api/use-reminders";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "./widget-error";
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
      <Card className="@container h-full">
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
      <WidgetError message="Failed to load reminders" />
    );
  }

  const reminders = ((data?.data ?? []) as Reminder[])
    .filter((r) => r.attributes.active)
    .sort((a, b) => a.attributes.scheduledFor.localeCompare(b.attributes.scheduledFor))
    .slice(0, config.limit);

  return (
    <Card className="@container h-full">
      <CardHeader>
        <CardTitle className="font-heading text-lg @xl:text-2xl">
          <Link to="/app/reminders" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/reminders" aria-label="Open reminders" className="inline-flex size-6 items-center justify-center rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring coarse:size-11"><ChevronRight aria-hidden="true" className="size-4 @xl:size-6 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {reminders.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Bell className="size-5 @xl:size-6" />
            <p className="text-sm @xl:text-xl">No active reminders</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {reminders.map((r) => (
              <li key={r.id} className="flex items-baseline justify-between gap-2 text-sm @xl:text-xl">
                <span className="truncate">{r.attributes.title}</span>
                <span className="text-sm @xl:text-base text-muted-foreground shrink-0">{formatRelative(r.attributes.scheduledFor)}</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
