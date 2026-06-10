import type { CalendarEvent } from "@/types/models/calendar";
import { formatChipTime } from "./calendar-utils";

interface MonthEventChipProps {
  event: CalendarEvent;
  timezone?: string | undefined;
}

/**
 * Presentational chip for one event in a month-grid day cell (desktop/tablet).
 * Not interactive — the enclosing day cell owns the click/focus (see design 2.5).
 */
export function MonthEventChip({ event, timezone }: MonthEventChipProps) {
  const { attributes: attrs } = event;
  const isBusy = attrs.title === "Busy" && !attrs.isOwner;
  const timeLabel = attrs.allDay ? "" : formatChipTime(attrs.startTime, timezone);

  return (
    <div
      className="rounded px-1 py-0.5 text-[10px] leading-tight truncate text-white"
      style={{ backgroundColor: isBusy ? "#9ca3af" : attrs.userColor }}
      title={attrs.title}
    >
      {timeLabel && <span className="font-semibold mr-1">{timeLabel}</span>}
      <span>{attrs.title}</span>
    </div>
  );
}
