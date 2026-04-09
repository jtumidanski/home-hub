import { useEffect, useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useMonthSummary, usePutEntry, useDeleteEntry, useSkipEntry } from "@/lib/hooks/api/use-trackers";
import { cn } from "@/lib/utils";
import type { MonthItemInfo, TrackerEntry, SentimentValue, NumericValue, RangeValue } from "@/types/models/tracker";

const colorBg: Record<string, string> = {
  red: "bg-red-100 dark:bg-red-950", orange: "bg-orange-100 dark:bg-orange-950",
  amber: "bg-amber-100 dark:bg-amber-950", yellow: "bg-yellow-100 dark:bg-yellow-950",
  lime: "bg-lime-100 dark:bg-lime-950", green: "bg-green-100 dark:bg-green-950",
  emerald: "bg-emerald-100 dark:bg-emerald-950", teal: "bg-teal-100 dark:bg-teal-950",
  cyan: "bg-cyan-100 dark:bg-cyan-950", blue: "bg-blue-100 dark:bg-blue-950",
  indigo: "bg-indigo-100 dark:bg-indigo-950", violet: "bg-violet-100 dark:bg-violet-950",
  purple: "bg-purple-100 dark:bg-purple-950", fuchsia: "bg-fuchsia-100 dark:bg-fuchsia-950",
  pink: "bg-pink-100 dark:bg-pink-950", rose: "bg-rose-100 dark:bg-rose-950",
};

function splitMonth(month: string): [string, string] {
  const parts = month.split("-");
  return [parts[0] ?? "2026", parts[1] ?? "01"];
}

function formatMonth(month: string) {
  const [y, m] = splitMonth(month);
  const date = new Date(parseInt(y), parseInt(m) - 1);
  return date.toLocaleDateString("en-US", { month: "long", year: "numeric" });
}

function getDaysInMonth(month: string): number {
  const [y, m] = splitMonth(month);
  return new Date(parseInt(y), parseInt(m), 0).getDate();
}

function isScheduledDay(day: number, month: string, item: MonthItemInfo): boolean {
  const [y, m] = splitMonth(month);
  const date = new Date(parseInt(y), parseInt(m) - 1, day);
  const dateStr = `${y}-${m}-${String(day).padStart(2, "0")}`;

  if (dateStr < item.active_from) return false;
  if (item.active_until && dateStr > item.active_until) return false;

  let applicable = item.schedule_snapshots?.[0];
  for (const snap of item.schedule_snapshots ?? []) {
    if (snap.effective_date <= dateStr) applicable = snap;
  }
  if (!applicable) return false;
  if (applicable.schedule.length === 0) return true;
  return applicable.schedule.includes(date.getDay());
}

interface Props {
  month: string;
  onMonthChange: (month: string) => void;
  onViewReport: () => void;
}

export function CalendarGrid({ month, onMonthChange, onViewReport }: Props) {
  const { data, isLoading } = useMonthSummary(month);
  const putEntry = usePutEntry();
  const deleteEntry = useDeleteEntry();
  const skipEntry = useSkipEntry();

  const summary = data?.data?.attributes;
  const items: MonthItemInfo[] = data?.data?.relationships?.items?.data ?? [];
  const entries: TrackerEntry[] = data?.data?.relationships?.entries?.data ?? [];

  const entryMap = useMemo(() => {
    const m = new Map<string, TrackerEntry>();
    entries.forEach((e) => {
      const attrs = e.attributes ?? e as unknown as TrackerEntry["attributes"];
      const key = `${attrs.tracking_item_id}:${attrs.date}`;
      m.set(key, e);
    });
    return m;
  }, [entries]);

  const daysInMonth = getDaysInMonth(month);
  const today = new Date().toISOString().slice(0, 10);
  const parts = month.split("-");
  const y = parts[0] ?? "2026";
  const m = parts[1] ?? "01";
  const currentMonth = `${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, "0")}`;

  const prevMonth = () => {
    const d = new Date(parseInt(y), parseInt(m) - 2);
    onMonthChange(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`);
  };
  const nextMonth = () => {
    const d = new Date(parseInt(y), parseInt(m));
    const next = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    if (next <= currentMonth) onMonthChange(next);
  };

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  const sortedItems = [...items].sort((a, b) => a.sort_order - b.sort_order);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Button variant="ghost" size="icon" onClick={prevMonth}><ChevronLeft className="h-4 w-4" /></Button>
        <h2 className="text-lg font-semibold">{formatMonth(month)}</h2>
        <Button variant="ghost" size="icon" onClick={nextMonth} disabled={month >= currentMonth}><ChevronRight className="h-4 w-4" /></Button>
      </div>

      {summary && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <div className="flex-1 bg-muted rounded-full h-2 overflow-hidden">
            <div className="bg-primary h-full transition-all" style={{ width: `${summary.completion.expected > 0 ? ((summary.completion.filled + summary.completion.skipped) / summary.completion.expected) * 100 : 0}%` }} />
          </div>
          <span>{summary.completion.filled + summary.completion.skipped}/{summary.completion.expected} entries</span>
          {summary.complete && <Button variant="link" size="sm" className="text-xs" onClick={onViewReport}>View Report</Button>}
        </div>
      )}

      <div className="md:hidden">
        <MobileDayView
          month={month}
          items={sortedItems}
          entryMap={entryMap}
          putEntry={putEntry}
          deleteEntry={deleteEntry}
          skipEntry={skipEntry}
        />
      </div>

      <div className="hidden md:block overflow-x-auto">
        <table className="w-full text-xs border-collapse">
          <thead>
            <tr>
              <th className="text-left p-1 min-w-[100px] sticky left-0 bg-background z-10">Item</th>
              {Array.from({ length: daysInMonth }, (_, i) => i + 1).map((day) => {
                const dateStr = `${y}-${m}-${String(day).padStart(2, "0")}`;
                return (
                  <th key={day} className={cn("p-1 text-center min-w-[32px]", dateStr === today && "bg-primary/10 rounded")}>
                    {day}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {sortedItems.map((item) => (
              <tr key={item.id}>
                <td className="p-1 font-medium truncate max-w-[120px] sticky left-0 bg-background z-10">
                  <span className={cn("inline-block w-2 h-2 rounded-full mr-1", colorBg[item.color]?.replace("100", "500").replace("950", "500"))} />
                  {item.name}
                </td>
                {Array.from({ length: daysInMonth }, (_, i) => i + 1).map((day) => {
                  const dateStr = `${y}-${m}-${String(day).padStart(2, "0")}`;
                  const scheduled = isScheduledDay(day, month, item);
                  const entry = entryMap.get(`${item.id}:${dateStr}`);
                  const entryAttrs = entry?.attributes ?? entry as unknown as TrackerEntry["attributes"];
                  const isFuture = dateStr > today;

                  return (
                    <td key={day} className={cn("p-0.5 text-center", dateStr === today && "bg-primary/5")}>
                      <CellContent
                        itemId={item.id}
                        date={dateStr}
                        scaleType={item.scale_type}
                        scaleConfig={item.scale_config}
                        scheduled={scheduled}
                        entry={entryAttrs}
                        isFuture={isFuture}
                        color={item.color}
                        itemName={item.name}
                        putEntry={putEntry}
                        deleteEntry={deleteEntry}
                        skipEntry={skipEntry}
                      />
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {sortedItems.length === 0 && (
        <p className="text-muted-foreground text-sm text-center py-8">No tracking items. Add items in Settings to get started.</p>
      )}
    </div>
  );
}

function MobileDayView({ month, items, entryMap, putEntry, deleteEntry, skipEntry }: {
  month: string;
  items: MonthItemInfo[];
  entryMap: Map<string, TrackerEntry>;
  putEntry: ReturnType<typeof usePutEntry>;
  deleteEntry: ReturnType<typeof useDeleteEntry>;
  skipEntry: ReturnType<typeof useSkipEntry>;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const daysInMonth = getDaysInMonth(month);
  const [y, m] = splitMonth(month);
  const monthKey = `${y}-${m}`;
  const isCurrentMonth = monthKey === today.slice(0, 7);
  const initialDay = isCurrentMonth ? parseInt(today.slice(8, 10)) : 1;

  const [selectedDay, setSelectedDay] = useState<number>(initialDay);

  useEffect(() => {
    setSelectedDay(isCurrentMonth ? parseInt(today.slice(8, 10)) : 1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [month]);

  const dayStr = String(selectedDay).padStart(2, "0");
  const dateStr = `${y}-${m}-${dayStr}`;
  const isFuture = dateStr > today;
  const dateLabel = new Date(dateStr + "T12:00:00").toLocaleDateString("en-US", {
    weekday: "long", month: "short", day: "numeric",
  });

  const visibleItems = items.filter((item) => {
    const scheduled = isScheduledDay(selectedDay, month, item);
    const entry = entryMap.get(`${item.id}:${dateStr}`);
    return scheduled || entry;
  });

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-2">
        <Button variant="ghost" size="icon" onClick={() => setSelectedDay((d) => Math.max(1, d - 1))} disabled={selectedDay <= 1}>
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <p className="text-sm font-medium flex-1 text-center">{dateLabel}</p>
        <Button variant="ghost" size="icon" onClick={() => setSelectedDay((d) => Math.min(daysInMonth, d + 1))} disabled={selectedDay >= daysInMonth}>
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>

      {visibleItems.length === 0 && (
        <p className="text-muted-foreground text-sm text-center py-4">Nothing scheduled for this day.</p>
      )}

      {visibleItems.map((item) => {
        const scheduled = isScheduledDay(selectedDay, month, item);
        const entry = entryMap.get(`${item.id}:${dateStr}`);
        const entryAttrs = entry?.attributes ?? (entry as unknown as TrackerEntry["attributes"]);
        const hasValue = entryAttrs && !entryAttrs.skipped && entryAttrs.value;
        const isSkipped = entryAttrs?.skipped;

        return (
          <Card key={item.id}>
            <CardContent className="py-3 px-4 space-y-2">
              <div className="flex items-center gap-2">
                <span className={cn("w-3 h-3 rounded-full", colorBg[item.color]?.replace("100", "500").replace("950", "500"))} />
                <span className="font-medium text-sm">{item.name}</span>
                {hasValue && <span className="text-xs text-green-600 ml-auto">logged</span>}
                {isSkipped && <span className="text-xs text-muted-foreground ml-auto">skipped</span>}
              </div>
              {isFuture ? (
                <p className="text-xs text-muted-foreground">{scheduled ? "Scheduled — log on the day." : ""}</p>
              ) : (
                <CellEditor
                  itemId={item.id}
                  date={dateStr}
                  scaleType={item.scale_type}
                  scaleConfig={item.scale_config}
                  scheduled={scheduled}
                  entry={entryAttrs}
                  putEntry={putEntry}
                  deleteEntry={deleteEntry}
                  skipEntry={skipEntry}
                  onDone={() => {}}
                />
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

function CellContent({ itemId, date, scaleType, scaleConfig, scheduled, entry, isFuture, color, itemName, putEntry, deleteEntry, skipEntry }: {
  itemId: string; date: string; scaleType: string; scaleConfig: { min: number; max: number } | null;
  scheduled: boolean; entry: TrackerEntry["attributes"] | undefined; isFuture: boolean; color: string; itemName: string;
  putEntry: ReturnType<typeof usePutEntry>; deleteEntry: ReturnType<typeof useDeleteEntry>; skipEntry: ReturnType<typeof useSkipEntry>;
}) {
  const [open, setOpen] = useState(false);

  if (!scheduled && !entry) {
    return <span className="text-muted-foreground/30">·</span>;
  }

  if (isFuture) {
    return <span className={cn("inline-block w-6 h-6 rounded text-center leading-6", scheduled ? "border border-dashed border-muted-foreground/30" : "")}>
      {scheduled ? "" : "·"}
    </span>;
  }

  const hasValue = entry && !entry.skipped && entry.value;
  const isSkipped = entry?.skipped;

  let display = "";
  if (isSkipped) {
    display = "—";
  } else if (hasValue) {
    const val = entry.value;
    if (scaleType === "sentiment") {
      const r = (val as SentimentValue)?.rating;
      display = r === "positive" ? "+" : r === "negative" ? "-" : "~";
    } else if (scaleType === "numeric") {
      display = String((val as NumericValue)?.count ?? 0);
    } else if (scaleType === "range") {
      display = String((val as RangeValue)?.value ?? 0);
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        className={cn(
          "w-6 h-6 rounded text-center leading-6 text-[10px] transition-colors",
          hasValue && colorBg[color],
          isSkipped && "bg-muted text-muted-foreground line-through",
          !hasValue && !isSkipped && scheduled && "border border-dashed border-primary/40 hover:bg-primary/10",
          !hasValue && !isSkipped && !scheduled && "text-muted-foreground/30"
        )}
      >
          {display || (scheduled ? "" : "·")}
      </PopoverTrigger>
      <PopoverContent className="w-48 p-2" align="center">
        <p className="text-xs font-medium mb-2">{itemName} — {new Date(date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })}</p>
        <CellEditor
          itemId={itemId} date={date} scaleType={scaleType} scaleConfig={scaleConfig}
          scheduled={scheduled} entry={entry} putEntry={putEntry} deleteEntry={deleteEntry}
          skipEntry={skipEntry} onDone={() => setOpen(false)}
        />
      </PopoverContent>
    </Popover>
  );
}

function RangeEditor({ min, max, initial, onCommit }: { min: number; max: number; initial: number | undefined; onCommit: (n: number) => void }) {
  const [local, setLocal] = useState<number>(initial ?? Math.round((min + max) / 2));
  useEffect(() => {
    if (initial !== undefined) setLocal(initial);
  }, [initial]);
  return (
    <div className="space-y-1">
      <input type="range" min={min} max={max} value={local} className="w-full"
        onChange={(e) => setLocal(parseInt(e.target.value))}
        onMouseUp={(e) => onCommit(parseInt((e.target as HTMLInputElement).value))}
        onTouchEnd={(e) => onCommit(parseInt((e.target as HTMLInputElement).value))}
        onKeyUp={(e) => onCommit(parseInt((e.target as HTMLInputElement).value))}
      />
      <p className="text-center text-xs font-mono">{local}</p>
    </div>
  );
}

function CellEditor({ itemId, date, scaleType, scaleConfig, scheduled, entry, putEntry, deleteEntry, skipEntry, onDone }: {
  itemId: string; date: string; scaleType: string; scaleConfig: { min: number; max: number } | null;
  scheduled: boolean; entry: TrackerEntry["attributes"] | undefined;
  putEntry: ReturnType<typeof usePutEntry>; deleteEntry: ReturnType<typeof useDeleteEntry>;
  skipEntry: ReturnType<typeof useSkipEntry>; onDone: () => void;
}) {
  const [note, setNote] = useState(entry?.note ?? "");

  const save = (value: unknown) => {
    putEntry.mutate({ itemId, date, value, note: note || null }, { onSuccess: onDone });
  };

  return (
    <div className="space-y-2">
      {scaleType === "sentiment" && (
        <div className="flex gap-1">
          {(["positive", "neutral", "negative"] as const).map((r) => (
            <Button key={r} variant={(entry?.value as SentimentValue)?.rating === r ? "default" : "outline"} size="sm" className="text-xs flex-1"
              onClick={() => save({ rating: r })}
            >{r === "positive" ? "+" : r === "negative" ? "-" : "~"}</Button>
          ))}
        </div>
      )}
      {scaleType === "numeric" && (
        <div className="flex items-center gap-1">
          <Button variant="outline" size="sm" onClick={() => save({ count: Math.max(0, ((entry?.value as NumericValue)?.count ?? 0) - 1) })}>-</Button>
          <span className="flex-1 text-center font-mono">{(entry?.value as NumericValue)?.count ?? 0}</span>
          <Button variant="outline" size="sm" onClick={() => save({ count: ((entry?.value as NumericValue)?.count ?? 0) + 1 })}>+</Button>
        </div>
      )}
      {scaleType === "range" && (
        <RangeEditor
          min={scaleConfig?.min ?? 0}
          max={scaleConfig?.max ?? 100}
          initial={(entry?.value as RangeValue)?.value}
          onCommit={(n) => save({ value: n })}
        />
      )}
      <Input placeholder="Note..." className="text-xs h-7" value={note} onChange={(e) => setNote(e.target.value)}
        onBlur={() => { if (entry?.value) putEntry.mutate({ itemId, date, value: entry.value, note: note || null }); }} />
      <div className="flex gap-1">
        {scheduled && <Button variant="ghost" size="sm" className="text-xs flex-1" onClick={() => { skipEntry.mutate({ itemId, date }); onDone(); }}>Skip</Button>}
        <Button variant="ghost" size="sm" className="text-xs flex-1" onClick={() => { deleteEntry.mutate({ itemId, date }); onDone(); }}>Clear</Button>
      </div>
    </div>
  );
}
