import { useEffect, useState } from "react";
import { Check } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useTrackerToday, usePutEntry, useSkipEntry } from "@/lib/hooks/api/use-trackers";
import { cn } from "@/lib/utils";
import type { TrackerEntry, RangeConfig } from "@/types/models/tracker";

const colorDot: Record<string, string> = {
  red: "bg-red-500", orange: "bg-orange-500", amber: "bg-amber-500", yellow: "bg-yellow-500",
  lime: "bg-lime-500", green: "bg-green-500", emerald: "bg-emerald-500", teal: "bg-teal-500",
  cyan: "bg-cyan-500", blue: "bg-blue-500", indigo: "bg-indigo-500", violet: "bg-violet-500",
  purple: "bg-purple-500", fuchsia: "bg-fuchsia-500", pink: "bg-pink-500", rose: "bg-rose-500",
};

const colorBorderLeft: Record<string, string> = {
  red: "border-l-red-500", orange: "border-l-orange-500", amber: "border-l-amber-500", yellow: "border-l-yellow-500",
  lime: "border-l-lime-500", green: "border-l-green-500", emerald: "border-l-emerald-500", teal: "border-l-teal-500",
  cyan: "border-l-cyan-500", blue: "border-l-blue-500", indigo: "border-l-indigo-500", violet: "border-l-violet-500",
  purple: "border-l-purple-500", fuchsia: "border-l-fuchsia-500", pink: "border-l-pink-500", rose: "border-l-rose-500",
};

interface TodayItem {
  id: string;
  name: string;
  scale_type: string;
  scale_config: RangeConfig | null;
  color: string;
  sort_order: number;
}

export function TodayView() {
  const { data, isLoading } = useTrackerToday();
  const putEntry = usePutEntry();
  const skipEntry = useSkipEntry();
  const [notes, setNotes] = useState<Record<string, string>>({});

  if (isLoading) {
    return <div className="space-y-3">{[1, 2, 3].map((i) => <Skeleton key={i} className="h-24 w-full" />)}</div>;
  }

  const items: TodayItem[] = data?.data?.relationships?.items?.data ?? [];
  const entries: TrackerEntry[] = data?.data?.relationships?.entries?.data ?? [];
  const todayDate = data?.data?.attributes?.date ?? "";

  const entryMap = new Map<string, TrackerEntry>();
  entries.forEach((e) => {
    const itemId = e.attributes?.tracking_item_id ?? (e as unknown as { tracking_item_id: string }).tracking_item_id;
    entryMap.set(itemId, e);
  });

  const loggedCount = entries.filter((e) => !(e.attributes ?? e as unknown as TrackerEntry["attributes"])?.skipped).length;
  const skippedCount = entries.filter((e) => (e.attributes ?? e as unknown as TrackerEntry["attributes"])?.skipped).length;
  const filledAndSkipped = loggedCount + skippedCount;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Today — {new Date(todayDate + "T12:00:00").toLocaleDateString("en-US", { weekday: "long", month: "short", day: "numeric" })}</h2>
      </div>

      {items.length > 0 && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <div className="flex-1 bg-muted rounded-full h-2 overflow-hidden">
            <div className="bg-primary h-full transition-all" style={{ width: `${items.length > 0 ? (filledAndSkipped / items.length) * 100 : 0}%` }} />
          </div>
          <span>{filledAndSkipped}/{items.length} entries</span>
        </div>
      )}

      {items.length === 0 && (
        <p className="text-muted-foreground text-sm">No items scheduled for today.</p>
      )}

      {items.map((item) => {
        const entry = entryMap.get(item.id);
        const entryAttrs = entry?.attributes ?? (entry as unknown as TrackerEntry["attributes"]);
        const hasValue = entryAttrs && !entryAttrs.skipped && entryAttrs.value;
        const isSkipped = entryAttrs?.skipped;

        return (
          <Card key={item.id} className={cn(
            !hasValue && !isSkipped && "border-l-[3px]",
            !hasValue && !isSkipped && colorBorderLeft[item.color],
            isSkipped && "opacity-60",
          )}>
            <CardContent className="py-3 px-4 space-y-2">
              <div className="flex items-center gap-2">
                <span className={cn("w-3 h-3 rounded-full", colorDot[item.color])} />
                <span className="font-medium">{item.name}</span>
                {hasValue && <Badge className="ml-auto bg-green-100 text-green-700 dark:bg-green-950 dark:text-green-400 border-transparent"><Check className="h-3 w-3" />logged</Badge>}
                {isSkipped && <Badge variant="secondary" className="ml-auto">skipped</Badge>}
              </div>

              {item.scale_type === "sentiment" && (
                <SentimentInput itemId={item.id} date={todayDate} currentRating={(entryAttrs?.value as { rating?: string })?.rating} putEntry={putEntry} />
              )}
              {item.scale_type === "numeric" && (
                <NumericInput itemId={item.id} date={todayDate} currentCount={(entryAttrs?.value as { count?: number })?.count} putEntry={putEntry} />
              )}
              {item.scale_type === "range" && (
                <RangeInput itemId={item.id} date={todayDate} config={item.scale_config} currentValue={(entryAttrs?.value as { value?: number })?.value} putEntry={putEntry} />
              )}

              <div className="flex items-center gap-2">
                <Input
                  placeholder="Add a note..."
                  className="text-xs h-7"
                  value={notes[item.id] ?? entryAttrs?.note ?? ""}
                  onChange={(e) => setNotes((prev) => ({ ...prev, [item.id]: e.target.value }))}
                  onBlur={() => {
                    const note = notes[item.id];
                    if (note !== undefined && entryAttrs?.value) {
                      putEntry.mutate({ itemId: item.id, date: todayDate, value: entryAttrs.value, note: note || null });
                    }
                  }}
                />
                <Button variant="ghost" size="sm" className="text-xs"
                  onClick={() => skipEntry.mutate({ itemId: item.id, date: todayDate })}
                >Skip</Button>
              </div>
            </CardContent>
          </Card>
        );
      })}

      <p className="text-sm text-muted-foreground text-center">{loggedCount}/{items.length} logged today</p>

    </div>
  );
}

function SentimentInput({ itemId, date, currentRating, putEntry }: { itemId: string; date: string; currentRating?: string | undefined; putEntry: ReturnType<typeof usePutEntry> }) {
  const ratings = [
    { value: "positive", emoji: "😊" },
    { value: "neutral", emoji: "😐" },
    { value: "negative", emoji: "😞" },
  ];
  return (
    <div className="flex gap-2">
      {ratings.map((r) => (
        <Button key={r.value} variant={currentRating === r.value ? "default" : "outline"} size="sm"
          onClick={() => putEntry.mutate({ itemId, date, value: { rating: r.value } })}
        >{r.emoji}</Button>
      ))}
    </div>
  );
}

function NumericInput({ itemId, date, currentCount, putEntry }: { itemId: string; date: string; currentCount?: number | undefined; putEntry: ReturnType<typeof usePutEntry> }) {
  const isSet = currentCount !== undefined;
  const count = currentCount ?? 0;
  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={() => putEntry.mutate({ itemId, date, value: { count: Math.max(0, count - 1) } })}>-</Button>
      <span className={cn("w-8 text-center font-mono", !isSet && "text-muted-foreground")}>{count}</span>
      <Button variant="outline" size="sm" onClick={() => putEntry.mutate({ itemId, date, value: { count: count + 1 } })}>+</Button>
    </div>
  );
}

function RangeInput({ itemId, date, config, currentValue, putEntry }: { itemId: string; date: string; config: RangeConfig | null; currentValue?: number | undefined; putEntry: ReturnType<typeof usePutEntry> }) {
  const min = config?.min ?? 0;
  const max = config?.max ?? 100;
  const isSet = currentValue !== undefined;
  const [local, setLocal] = useState<number>(currentValue ?? Math.round((min + max) / 2));
  useEffect(() => {
    if (currentValue !== undefined) setLocal(currentValue);
  }, [currentValue]);
  const commit = (n: number) => putEntry.mutate({ itemId, date, value: { value: n } });
  return (
    <div className="flex items-center gap-2">
      <input type="range" min={min} max={max} value={local} className="flex-1"
        onChange={(e) => setLocal(parseInt(e.target.value))}
        onMouseUp={(e) => commit(parseInt((e.target as HTMLInputElement).value))}
        onTouchEnd={(e) => commit(parseInt((e.target as HTMLInputElement).value))}
        onKeyUp={(e) => commit(parseInt((e.target as HTMLInputElement).value))}
      />
      <span className={cn("w-10 text-center font-mono text-sm", !isSet && "text-muted-foreground")}>{isSet ? local : "Not set"}</span>
    </div>
  );
}
