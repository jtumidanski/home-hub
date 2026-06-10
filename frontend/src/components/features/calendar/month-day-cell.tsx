import type { CalendarEvent } from "@/types/models/calendar";
import { cn } from "@/lib/utils";
import { MonthEventChip } from "./month-event-chip";

interface DayBucket {
  allDay: CalendarEvent[];
  timed: CalendarEvent[];
}

interface MonthDayCellProps {
  day: Date;
  bucket: DayBucket;
  inMonth: boolean;
  isToday: boolean;
  isDesktop: boolean;
  timezone?: string | undefined;
  onDayClick: (day: Date) => void;
}

const MAX_DOTS = 4;

function eventColor(event: CalendarEvent): string {
  const { attributes: attrs } = event;
  return attrs.title === "Busy" && !attrs.isOwner ? "#9ca3af" : attrs.userColor;
}

function MonthDayDots({ events }: { events: CalendarEvent[] }) {
  const overflow = events.length > MAX_DOTS;
  const shown = overflow ? events.slice(0, MAX_DOTS - 1) : events;
  return (
    <div className="flex flex-wrap gap-0.5 mt-1" aria-hidden="true">
      {shown.map((evt) => (
        <span
          key={evt.id}
          className="w-1.5 h-1.5 rounded-full"
          style={{ backgroundColor: eventColor(evt) }}
        />
      ))}
      {overflow && (
        <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground" data-testid="dot-overflow" />
      )}
    </div>
  );
}

/**
 * One month-grid day cell. The cell is the single keyboard-focusable,
 * activatable control (design 2.5); chips/dots inside are presentational.
 */
export function MonthDayCell({
  day,
  bucket,
  inMonth,
  isToday,
  isDesktop,
  timezone,
  onDayClick,
}: MonthDayCellProps) {
  const ordered = [...bucket.allDay, ...bucket.timed];
  const dayLabel = day.toLocaleDateString("en-US", { month: "long", day: "numeric" });
  const label = `${dayLabel}, ${ordered.length} event${ordered.length === 1 ? "" : "s"}`;

  const cellTone = isToday ? "bg-primary/10" : !inMonth ? "bg-muted/30" : "";
  const numberTone = isToday ? "text-primary" : !inMonth ? "text-muted-foreground" : "";

  return (
    <button
      type="button"
      aria-label={label}
      onClick={() => onDayClick(day)}
      className={cn(
        "flex flex-col min-h-0 overflow-hidden border-r border-b last:border-r-0 text-left p-1 cursor-pointer hover:bg-accent/40",
        cellTone,
      )}
    >
      <span className={cn("text-xs font-medium px-0.5", numberTone)}>{day.getDate()}</span>
      {isDesktop ? (
        <div className="flex-1 overflow-y-auto flex flex-col gap-0.5 mt-0.5">
          {ordered.map((evt) => (
            <MonthEventChip key={evt.id} event={evt} timezone={timezone} />
          ))}
        </div>
      ) : (
        <MonthDayDots events={ordered} />
      )}
    </button>
  );
}
