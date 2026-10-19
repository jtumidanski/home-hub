import { useMemo, useState } from "react";
import { CalendarOff, Check, ChevronLeft, ChevronRight, ListChecks } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import { EmptyState } from "@/components/common/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useMonthSummary, usePutEntry, useDeleteEntry, useSkipEntry } from "@/lib/hooks/api/use-trackers";
import { useTenant } from "@/context/tenant-context";
import { cn } from "@/lib/utils";
import { swatchClass } from "@/lib/constants/swatches";
import { getLocalTodayStr, getLocalMonth } from "@/lib/date-utils";
import type { MonthItemInfo, TrackerEntry, SentimentValue, NumericValue, RangeValue } from "@/types/models/tracker";

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
  onOpenSetup?: () => void;
}

export function CalendarGrid({ month, onMonthChange, onViewReport, onOpenSetup }: Props) {
  const { household } = useTenant();
  const timezone = household?.attributes.timezone;
  const today = getLocalTodayStr(timezone);
  const { data, isLoading } = useMonthSummary(month, today);
  const putEntry = usePutEntry();
  const deleteEntry = useDeleteEntry();
  const skipEntry = useSkipEntry();

  const summary = data?.data?.attributes;
  const items: MonthItemInfo[] = data?.data?.relationships?.items?.data ?? [];
  const entryMap = useMemo(() => {
    const entries: TrackerEntry[] = data?.data?.relationships?.entries?.data ?? [];
    const m = new Map<string, TrackerEntry>();
    entries.forEach((e) => {
      const attrs = e.attributes ?? e as unknown as TrackerEntry["attributes"];
      const key = `${attrs.tracking_item_id}:${attrs.date}`;
      m.set(key, e);
    });
    return m;
  }, [data?.data?.relationships?.entries?.data]);

  const daysInMonth = getDaysInMonth(month);
  const parts = month.split("-");
  const y = parts[0] ?? "2026";
  const m = parts[1] ?? "01";
  const currentMonth = getLocalMonth(timezone);

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
        <Button variant="ghost" size="icon" onClick={prevMonth} aria-label="Previous month"><ChevronLeft className="h-4 w-4" aria-hidden="true" /></Button>
        <h2 className="text-lg font-heading tracking-[-0.01em]">{formatMonth(month)}</h2>
        <Button variant="ghost" size="icon" onClick={nextMonth} disabled={month >= currentMonth} aria-label="Next month"><ChevronRight className="h-4 w-4" aria-hidden="true" /></Button>
      </div>

      {summary && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <div className="flex-1 bg-muted rounded-full h-2 overflow-hidden">
            <div className="bg-primary h-full motion-safe:transition-[width]" style={{ width: `${summary.completion.expected > 0 ? ((summary.completion.filled + summary.completion.skipped) / summary.completion.expected) * 100 : 0}%` }} />
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
          timezone={timezone}
          today={today}
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
                  <span className={cn("inline-block w-2 h-2 rounded-full mr-1", swatchClass(item.color, "solid"))} />
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
                        today={today}
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
        <EmptyState
          icon={ListChecks}
          title="No habits yet"
          description="Add habits in Setup to start tracking."
          action={onOpenSetup ? <Button variant="outline" onClick={onOpenSetup}>Open Setup</Button> : undefined}
        />
      )}
    </div>
  );
}

function MobileDayView({ month, items, entryMap, putEntry, deleteEntry, skipEntry, timezone, today }: {
  month: string;
  items: MonthItemInfo[];
  entryMap: Map<string, TrackerEntry>;
  putEntry: ReturnType<typeof usePutEntry>;
  deleteEntry: ReturnType<typeof useDeleteEntry>;
  skipEntry: ReturnType<typeof useSkipEntry>;
  timezone: string | undefined;
  today: string;
}) {
  void timezone;
  const daysInMonth = getDaysInMonth(month);
  const [y, m] = splitMonth(month);
  const monthKey = `${y}-${m}`;
  const isCurrentMonth = monthKey === today.slice(0, 7);
  const initialDay = isCurrentMonth ? parseInt(today.slice(8, 10)) : 1;

  const [selectedDay, setSelectedDay] = useState<number>(initialDay);
  const [renderedMonth, setRenderedMonth] = useState(month);
  if (month !== renderedMonth) {
    setRenderedMonth(month);
    setSelectedDay(isCurrentMonth ? parseInt(today.slice(8, 10)) : 1);
  }

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
        <Button variant="ghost" size="icon" aria-label="Previous day" onClick={() => setSelectedDay((d) => Math.max(1, d - 1))} disabled={selectedDay <= 1}>
          <ChevronLeft className="h-4 w-4" aria-hidden="true" />
        </Button>
        <p className="text-sm font-medium flex-1 text-center">{dateLabel}</p>
        <Button variant="ghost" size="icon" aria-label="Next day" onClick={() => setSelectedDay((d) => Math.min(daysInMonth, d + 1))} disabled={selectedDay >= daysInMonth}>
          <ChevronRight className="h-4 w-4" aria-hidden="true" />
        </Button>
      </div>

      {visibleItems.length === 0 && (
        <EmptyState className="py-6" headingLevel="h3" icon={CalendarOff} title="Nothing scheduled" description="No habits are scheduled for this day." />
      )}

      {visibleItems.map((item) => {
        const scheduled = isScheduledDay(selectedDay, month, item);
        const entry = entryMap.get(`${item.id}:${dateStr}`);
        const entryAttrs = entry?.attributes ?? (entry as unknown as TrackerEntry["attributes"]);
        const hasValue = entryAttrs && !entryAttrs.skipped && entryAttrs.value;
        const isSkipped = entryAttrs?.skipped;

        return (
          <Card key={item.id} className={cn(
            isSkipped && "opacity-60",
          )}>
            <CardContent className="py-3 px-4 space-y-2">
              <div className="flex items-center gap-2">
                <span className={cn("w-3 h-3 rounded-full", swatchClass(item.color, "solid"))} />
                <span className="font-medium text-sm">{item.name}</span>
                {hasValue && <Badge className="ml-auto bg-success/15 text-success border-transparent"><Check className="h-3 w-3" />logged</Badge>}
                {isSkipped && <Badge variant="secondary" className="ml-auto">skipped</Badge>}
                {!hasValue && !isSkipped && <Badge variant="outline" className="ml-auto">pending</Badge>}
              </div>
              {isFuture ? (
                <p className="text-xs text-muted-foreground">{scheduled ? "Scheduled — log on the day." : ""}</p>
              ) : (
                <CellEditor
                  itemId={item.id}
                  date={dateStr}
                  today={today}
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

function CellContent({ itemId, date, today, scaleType, scaleConfig, scheduled, entry, isFuture, color, itemName, putEntry, deleteEntry, skipEntry }: {
  itemId: string; date: string; today: string; scaleType: string; scaleConfig: { min: number; max: number } | null;
  scheduled: boolean; entry: TrackerEntry["attributes"] | undefined; isFuture: boolean; color: string; itemName: string;
  putEntry: ReturnType<typeof usePutEntry>; deleteEntry: ReturnType<typeof useDeleteEntry>; skipEntry: ReturnType<typeof useSkipEntry>;
}) {
  const [open, setOpen] = useState(false);

  if (!scheduled && !entry) {
    return <span className="text-muted-foreground">·</span>;
  }

  if (isFuture) {
    return <span className={cn("inline-flex size-7 coarse:size-11 items-center justify-center rounded text-muted-foreground", scheduled ? "border border-dashed border-muted-foreground/60" : "")}>
      {scheduled ? "" : "·"}
    </span>;
  }

  const hasValue = entry && !entry.skipped && entry.value;
  const isSkipped = entry?.skipped;
  const hasNote = !!entry?.note;

  let display = "";
  if (isSkipped) {
    display = "—";
  } else if (hasValue) {
    const val = entry.value;
    if (scaleType === "sentiment") {
      const r = (val as SentimentValue)?.rating;
      display = r === "positive" ? "😊" : r === "negative" ? "😞" : "😐";
    } else if (scaleType === "numeric") {
      display = String((val as NumericValue)?.count ?? 0);
    } else if (scaleType === "range") {
      display = String((val as RangeValue)?.value ?? 0);
    }
  }

  const dateLabel = new Date(date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" });
  const scoreLabel = isSkipped
    ? "Skipped"
    : hasValue
      ? scaleType === "sentiment"
        ? `Sentiment ${(entry.value as SentimentValue).rating}`
        : `Score ${display}`
      : "No entry";
  const ariaLabel = `${itemName}, ${dateLabel}. ${scoreLabel}.${hasNote ? ` Note: ${entry!.note}` : ""}`;

  const triggerClassName = cn(
    "relative inline-flex size-7 coarse:size-11 items-center justify-center rounded text-xs motion-safe:transition-colors cursor-pointer",
    hasValue && swatchClass(color, "soft"),
    isSkipped && "bg-muted text-muted-foreground line-through",
    !hasValue && !isSkipped && scheduled && !hasNote && "border border-dashed border-primary/40 hover:bg-primary/10",
    !hasValue && !isSkipped && !scheduled && !hasNote && "text-muted-foreground",
    hasNote && "border",
    hasNote && swatchClass(color, "border"),
  );

  const triggerNode = (
    <PopoverTrigger className={triggerClassName} aria-label={ariaLabel}>
      {display || (scheduled ? "" : "·")}
      {hasNote && (
        <span
          aria-hidden="true"
          className={cn(
            "absolute top-0 right-0 h-1.5 w-1.5",
            "[clip-path:polygon(100%_0,100%_100%,0_0)]",
            swatchClass(color, "solid"),
          )}
        />
      )}
    </PopoverTrigger>
  );

  return (
    <Popover open={open} onOpenChange={setOpen}>
      {hasNote ? (
        <Tooltip>
          <TooltipTrigger render={triggerNode} />
          <TooltipContent>{entry!.note}</TooltipContent>
        </Tooltip>
      ) : (
        triggerNode
      )}
      <PopoverContent className="w-48 p-2" align="center">
        <p className="text-xs font-medium mb-2">{itemName} — {new Date(date + "T12:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })}</p>
        <CellEditor
          itemId={itemId} date={date} today={today} scaleType={scaleType} scaleConfig={scaleConfig}
          scheduled={scheduled} entry={entry} putEntry={putEntry} deleteEntry={deleteEntry}
          skipEntry={skipEntry} onDone={() => setOpen(false)}
        />
      </PopoverContent>
    </Popover>
  );
}

function RangeEditor({ itemId, date, min, max, initial, onCommit }: { itemId: string; date: string; min: number; max: number; initial: number | undefined; onCommit: (n: number) => void }) {
  const [local, setLocal] = useState<number>(initial ?? Math.round((min + max) / 2));
  const [touched, setTouched] = useState(initial !== undefined);
  const resetKey = `${itemId}:${date}:${initial}:${min}:${max}`;
  const [prevResetKey, setPrevResetKey] = useState(resetKey);
  if (resetKey !== prevResetKey) {
    setPrevResetKey(resetKey);
    if (initial !== undefined) {
      setLocal(initial);
      setTouched(true);
    } else {
      setLocal(Math.round((min + max) / 2));
      setTouched(false);
    }
  }
  const handleCommit = (n: number) => { setTouched(true); onCommit(n); };
  return (
    <div className="space-y-1">
      <input type="range" min={min} max={max} value={local} className="w-full"
        onChange={(e) => { setLocal(parseInt(e.target.value)); setTouched(true); }}
        onMouseUp={(e) => handleCommit(parseInt((e.target as HTMLInputElement).value))}
        onTouchEnd={(e) => handleCommit(parseInt((e.target as HTMLInputElement).value))}
        onKeyUp={(e) => handleCommit(parseInt((e.target as HTMLInputElement).value))}
      />
      <p className={cn("text-center text-xs font-mono", !touched && "text-muted-foreground")}>{touched ? local : "Not set"}</p>
    </div>
  );
}

function CellEditor({ itemId, date, today, scaleType, scaleConfig, scheduled, entry, putEntry, deleteEntry, skipEntry, onDone }: {
  itemId: string; date: string; today: string; scaleType: string; scaleConfig: { min: number; max: number } | null;
  scheduled: boolean; entry: TrackerEntry["attributes"] | undefined;
  putEntry: ReturnType<typeof usePutEntry>; deleteEntry: ReturnType<typeof useDeleteEntry>;
  skipEntry: ReturnType<typeof useSkipEntry>; onDone: () => void;
}) {
  const [note, setNote] = useState(entry?.note ?? "");

  const save = (value: unknown) => {
    putEntry.mutate({ itemId, date, today, value, note: note || null }, { onSuccess: onDone });
  };

  return (
    <div className="space-y-2">
      {scaleType === "sentiment" && (
        <div className="flex gap-1">
          {(["positive", "neutral", "negative"] as const).map((r) => (
            <Button key={r} variant={(entry?.value as SentimentValue)?.rating === r ? "default" : "outline"} size="sm" className="text-xs flex-1"
              onClick={() => save({ rating: r })}
            >{r === "positive" ? "😊" : r === "negative" ? "😞" : "😐"}</Button>
          ))}
        </div>
      )}
      {scaleType === "numeric" && (() => {
        const hasEntry = entry?.value != null;
        const count = hasEntry ? ((entry.value as NumericValue)?.count ?? 0) : null;
        return (
          <div className="flex items-center gap-1">
            <Button variant="outline" size="sm" onClick={() => save({ count: Math.max(0, (count ?? 0) - 1) })}>-</Button>
            <span className={cn("flex-1 text-center font-mono", count === null && "text-muted-foreground")}>{count ?? "–"}</span>
            <Button variant="outline" size="sm" onClick={() => save({ count: (count ?? 0) + 1 })}>+</Button>
          </div>
        );
      })()}
      {scaleType === "range" && (
        <RangeEditor
          itemId={itemId}
          date={date}
          min={scaleConfig?.min ?? 0}
          max={scaleConfig?.max ?? 100}
          initial={(entry?.value as RangeValue)?.value}
          onCommit={(n) => save({ value: n })}
        />
      )}
      <Input placeholder="Note..." className="text-xs h-7" value={note} onChange={(e) => setNote(e.target.value)}
        onBlur={() => { if (entry?.value) putEntry.mutate({ itemId, date, today, value: entry.value, note: note || null }); }} />
      <div className="flex gap-1">
        {scheduled && <Button variant="ghost" size="sm" className="text-xs flex-1" onClick={() => { skipEntry.mutate({ itemId, date, today }); onDone(); }}>Skip</Button>}
        <Button variant="ghost" size="sm" className="text-xs flex-1" onClick={() => { deleteEntry.mutate({ itemId, date }); onDone(); }}>Clear</Button>
      </div>
    </div>
  );
}
