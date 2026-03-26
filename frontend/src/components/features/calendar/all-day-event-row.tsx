import { useState } from "react";
import type { CalendarEvent } from "@/types/models/calendar";
import { EventPopover } from "./event-popover";

interface AllDayEventRowProps {
  events: CalendarEvent[];
}

export function AllDayEventRow({ events }: AllDayEventRowProps) {
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null);

  if (events.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-1 px-1 py-1 min-h-[28px]">
      {events.map((evt) => {
        const isBusy = evt.attributes.title === "Busy" && !evt.attributes.isOwner;
        return (
          <button
            key={evt.id}
            type="button"
            className="rounded px-2 py-0.5 text-xs text-white truncate max-w-full cursor-pointer hover:brightness-110 transition-all"
            style={{ backgroundColor: isBusy ? "#9ca3af" : evt.attributes.userColor }}
            onClick={() => setSelectedEvent(evt)}
          >
            {evt.attributes.title}
          </button>
        );
      })}
      {selectedEvent && (
        <EventPopover event={selectedEvent} onClose={() => setSelectedEvent(null)} />
      )}
    </div>
  );
}
