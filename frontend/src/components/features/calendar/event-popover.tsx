import { Clock, MapPin, User, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import type { CalendarEvent } from "@/types/models/calendar";

interface EventPopoverProps {
  event: CalendarEvent;
  onClose: () => void;
  hasWriteAccess?: boolean | undefined;
  onEdit?: ((event: CalendarEvent) => void) | undefined;
  onDelete?: ((event: CalendarEvent) => void) | undefined;
}

export function EventPopover({ event, onClose, hasWriteAccess, onEdit, onDelete }: EventPopoverProps) {
  const { attributes: attrs } = event;

  const startTime = new Date(attrs.startTime);
  const endTime = new Date(attrs.endTime);

  const timeDisplay = attrs.allDay
    ? "All day"
    : `${startTime.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })} – ${endTime.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}`;

  const canModify = attrs.isOwner && hasWriteAccess;

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose(); }}>
      <DialogContent className="w-80 max-w-[90vw] gap-0">
        <div className="mb-3 flex items-center gap-2">
          <div
            className="size-3 shrink-0 rounded-full ring-1 ring-border"
            style={{ backgroundColor: attrs.userColor }}
            aria-hidden="true"
          />
          <DialogTitle className="text-sm font-semibold">{attrs.title}</DialogTitle>
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
      </DialogContent>
    </Dialog>
  );
}
