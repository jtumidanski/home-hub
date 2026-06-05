import { Link } from "react-router-dom";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, LayoutDashboard, Pencil } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { DashboardKebabMenu } from "@/components/features/dashboards/dashboard-kebab-menu";
import type { Dashboard } from "@/types/models/dashboard";

interface DashboardRowProps {
  dashboard: Dashboard;
  defaultDashboardId: string | null;
}

export function DashboardRow({ dashboard, defaultDashboardId }: DashboardRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: dashboard.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const isDefault = defaultDashboardId === dashboard.id;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex items-center gap-2 rounded-md border bg-card p-3"
    >
      <button
        type="button"
        aria-label={`Drag ${dashboard.attributes.name} to reorder`}
        className="flex h-9 w-9 shrink-0 cursor-grab touch-none items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="h-4 w-4" />
      </button>

      <LayoutDashboard className="h-4 w-4 shrink-0 text-muted-foreground" />

      <Link
        to={`/app/dashboards/${dashboard.id}`}
        className="min-w-0 flex-1 truncate text-sm font-medium hover:underline"
      >
        {dashboard.attributes.name}
      </Link>

      {isDefault && (
        <Badge variant="secondary" className="shrink-0">
          Default
        </Badge>
      )}

      <Link
        to={`/app/dashboards/${dashboard.id}/edit`}
        aria-label={`Edit ${dashboard.attributes.name}`}
        className={cn(
          "flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-muted-foreground",
          "hover:bg-accent hover:text-accent-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring",
        )}
      >
        <Pencil className="h-4 w-4" />
      </Link>

      <DashboardKebabMenu dashboard={dashboard} isDefault={isDefault} />
    </div>
  );
}
