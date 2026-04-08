import { Circle, CheckCircle2, Trash2, Calendar } from "lucide-react";
import { type Task } from "@/types/models/task";
import { Card, CardHeader, CardTitle, CardAction, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CardActionMenu } from "@/components/common/card-action-menu";
import { cn } from "@/lib/utils";

interface TaskCardProps {
  task: Task;
  ownerName?: string;
  onToggleComplete: (id: string, currentStatus: string) => void;
  onDelete: (id: string) => void;
}

export function TaskCard({ task, ownerName, onToggleComplete, onDelete }: TaskCardProps) {
  const { id, attributes } = task;
  const isCompleted = attributes.status === "completed";

  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle
          className={cn(isCompleted && "line-through text-muted-foreground")}
        >
          {attributes.title}
        </CardTitle>
        <CardAction>
          <CardActionMenu
            actions={[
              {
                icon: isCompleted ? (
                  <CheckCircle2 className="h-4 w-4" />
                ) : (
                  <Circle className="h-4 w-4" />
                ),
                label: isCompleted ? "Mark incomplete" : "Mark complete",
                onClick: () => onToggleComplete(id, attributes.status),
              },
              {
                icon: <Trash2 className="h-4 w-4" />,
                label: "Delete",
                onClick: () => onDelete(id),
                variant: "destructive",
              },
            ]}
          />
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={isCompleted ? "secondary" : "default"}>
            {attributes.status}
          </Badge>
          {attributes.dueOn && (
            <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
              <Calendar className="h-3 w-3" />
              {attributes.dueOn}
            </span>
          )}
          {ownerName && (
            <span className="text-xs text-muted-foreground">{ownerName}</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
