import { Clock, BellOff, Trash2 } from "lucide-react";
import { type Reminder, isReminderDismissed, isReminderSnoozed } from "@/types/models/reminder";
import { Card, CardHeader, CardTitle, CardAction, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CardActionMenu } from "@/components/common/card-action-menu";

interface ReminderCardProps {
  reminder: Reminder;
  ownerName?: string;
  onSnooze: (id: string) => void;
  onDismiss: (id: string) => void;
  onDelete: (id: string) => void;
}

export function ReminderCard({ reminder, ownerName, onSnooze, onDismiss, onDelete }: ReminderCardProps) {
  const { id, attributes } = reminder;

  const statusLabel = attributes.active
    ? "active"
    : isReminderDismissed(reminder)
      ? "dismissed"
      : isReminderSnoozed(reminder)
        ? "snoozed"
        : "inactive";

  const actions = [
    ...(attributes.active
      ? [
          { icon: <Clock className="h-4 w-4" />, label: "Snooze", onClick: () => onSnooze(id) },
          { icon: <BellOff className="h-4 w-4" />, label: "Dismiss", onClick: () => onDismiss(id) },
        ]
      : []),
    {
      icon: <Trash2 className="h-4 w-4" />,
      label: "Delete",
      onClick: () => onDelete(id),
      variant: "destructive" as const,
    },
  ];

  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle>{attributes.title}</CardTitle>
        <CardAction>
          <CardActionMenu actions={actions} />
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={attributes.active ? "default" : "secondary"}>
            {statusLabel}
          </Badge>
          <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
            <Clock className="h-3 w-3" />
            {new Date(attributes.scheduledFor).toLocaleString()}
          </span>
          {ownerName && (
            <span className="text-xs text-muted-foreground">{ownerName}</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
