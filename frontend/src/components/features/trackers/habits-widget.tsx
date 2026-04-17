import { Link } from "react-router-dom";
import { useTrackerToday } from "@/lib/hooks/api/use-trackers";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ListChecks, Check, Circle, SkipForward, ChevronRight } from "lucide-react";

export function HabitsWidget() {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading, isError } = useTrackerToday(today);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-24" />
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="border-destructive">
        <CardContent className="py-3">
          <p className="text-sm text-destructive">Failed to load habits</p>
        </CardContent>
      </Card>
    );
  }

  const items = data?.data?.relationships?.items?.data ?? [];
  const entries = data?.data?.relationships?.entries?.data ?? [];
  const completedItemIds = new Set(
    entries
      .filter((e) => {
        const attrs = e.attributes ?? (e as unknown as Record<string, unknown>);
        return !attrs.skipped && attrs.value !== null && attrs.value !== undefined;
      })
      .map((e) => {
        const attrs = e.attributes ?? (e as unknown as Record<string, unknown>);
        return attrs.tracking_item_id as string;
      })
  );
  const skippedItemIds = new Set(
    entries
      .filter((e) => {
        const attrs = e.attributes ?? (e as unknown as Record<string, unknown>);
        return attrs.skipped === true;
      })
      .map((e) => {
        const attrs = e.attributes ?? (e as unknown as Record<string, unknown>);
        return attrs.tracking_item_id as string;
      })
  );

  return (
    <Link to="/app/habits" className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader>
          <CardTitle className="text-sm font-medium">Habits</CardTitle>
          <CardAction>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <ListChecks className="h-5 w-5" />
              <p className="text-sm">No habits scheduled for today</p>
            </div>
          ) : (
            <div className="space-y-2">
              {items.map((item) => {
                const done = completedItemIds.has(item.id);
                const skipped = skippedItemIds.has(item.id);
                return (
                  <div key={item.id} className="flex items-center gap-2">
                    {done ? (
                      <Check className="h-4 w-4 text-green-600 shrink-0" aria-label="Completed" />
                    ) : skipped ? (
                      <SkipForward className="h-4 w-4 text-muted-foreground shrink-0" aria-label="Skipped" />
                    ) : (
                      <Circle className="h-4 w-4 text-muted-foreground shrink-0" aria-label="Pending" />
                    )}
                    <span className={`text-sm truncate ${done || skipped ? "text-muted-foreground line-through" : ""}`}>
                      {item.name}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
