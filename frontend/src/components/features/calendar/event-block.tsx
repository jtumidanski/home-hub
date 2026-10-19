import { useState } from "react";
import type { CalendarEvent } from "@/types/models/calendar";
import { EventPopover } from "./event-popover";
import { eventChipColors } from "./calendar-utils";

interface EventBlockProps {
  event: CalendarEvent;
  top: number;
  height: number;
  left: string;
  width: string;
  hasWriteAccess?: boolean | undefined;
  onEdit?: ((event: CalendarEvent) => void) | undefined;
  onDelete?: ((event: CalendarEvent) => void) | undefined;
}

export function EventBlock({ event, top, height, left, width, hasWriteAccess, onEdit, onDelete }: EventBlockProps) {
  const [showPopover, setShowPopover] = useState(false);
  const { attributes: attrs } = event;

  return (
    <>
      <button
        type="button"
        className="absolute rounded px-1.5 py-0.5 text-xs leading-tight overflow-hidden cursor-pointer border border-foreground/10 hover:brightness-110 transition-all"
        style={{
          top: `${top}px`,
          height: `${Math.max(height, 18)}px`,
          left,
          width,
          ...eventChipColors(event),
        }}
        onClick={() => setShowPopover(true)}
      >
        <div className="font-medium truncate">{attrs.title}</div>
        {height > 30 && (
          <div className="truncate opacity-80">
            {new Date(attrs.startTime).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}
          </div>
        )}
      </button>
      {showPopover && (
        <EventPopover
          event={event}
          onClose={() => setShowPopover(false)}
          hasWriteAccess={hasWriteAccess}
          onEdit={(evt) => {
            setShowPopover(false);
            onEdit?.(evt);
          }}
          onDelete={(evt) => {
            setShowPopover(false);
            onDelete?.(evt);
          }}
        />
      )}
    </>
  );
}
