import { useEffect, useRef } from "react";
import { X, Clock, MapPin, User, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { CalendarEvent } from "@/types/models/calendar";

interface EventPopoverProps {
  event: CalendarEvent;
  onClose: () => void;
  hasWriteAccess?: boolean | undefined;
  onEdit?: ((event: CalendarEvent) => void) | undefined;
  onDelete?: ((event: CalendarEvent) => void) | undefined;
}

export function EventPopover({ event, onClose, hasWriteAccess, onEdit, onDelete }: EventPopoverProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { attributes: attrs } = event;

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [onClose]);

  const startTime = new Date(attrs.startTime);
  const endTime = new Date(attrs.endTime);

  const timeDisplay = attrs.allDay
    ? "All day"
    : `${startTime.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })} – ${endTime.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}`;

  const canModify = attrs.isOwner && hasWriteAccess;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20" onClick={onClose}>
      <div
        ref={ref}
        className="bg-background border rounded-lg shadow-lg p-4 w-80 max-w-[90vw]"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full flex-shrink-0" style={{ backgroundColor: attrs.userColor }} />
            <h3 className="font-semibold text-sm">{attrs.title}</h3>
          </div>
          <button type="button" onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-2 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <Clock className="h-3.5 w-3.5" />
            <span>{timeDisplay}</span>
          </div>

          {attrs.location && (
            <div className="flex items-center gap-2">
              <MapPin className="h-3.5 w-3.5" />
              <span>{attrs.location}</span>
            </div>
          )}

          <div className="flex items-center gap-2">
            <User className="h-3.5 w-3.5" />
            <span>{attrs.userDisplayName}</span>
          </div>

          {attrs.description && (
            <p className="pt-2 border-t text-xs whitespace-pre-wrap">{attrs.description}</p>
          )}
        </div>

        {canModify && (
          <div className="flex gap-2 mt-3 pt-3 border-t">
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              onClick={() => onEdit?.(event)}
            >
              <Pencil className="h-3.5 w-3.5 mr-1.5" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="flex-1 text-destructive hover:text-destructive"
              onClick={() => onDelete?.(event)}
            >
              <Trash2 className="h-3.5 w-3.5 mr-1.5" />
              Delete
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
