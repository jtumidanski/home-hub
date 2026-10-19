import { useState } from "react";
import type { CalendarEvent } from "@/types/models/calendar";
import { EventPopover } from "./event-popover";
import { eventChipColors } from "@/lib/calendar/event-colors";

interface AllDayEventRowProps {
  events: CalendarEvent[];
}

export function AllDayEventRow({ events }: AllDayEventRowProps) {
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null);

  if (events.length === 0) return null;

  return (
    <div className="flex flex-col gap-1 px-1 py-1 min-h-[28px]">
      {events.map((evt) => {
        return (
          <button
            key={evt.id}
            type="button"
            className="block w-full min-h-6 coarse:min-h-11 rounded px-2 py-0.5 text-left text-xs truncate cursor-pointer hover:brightness-110 transition-[filter] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            style={eventChipColors(evt.attributes)}
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
