import { useMemo, useState } from "react";
import { NavLink } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight, GripVertical, LayoutDashboard, Plus } from "lucide-react";
import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { cn } from "@/lib/utils";
import { useDashboards, useReorderDashboards } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { NewDashboardModal } from "@/components/features/dashboards/new-dashboard-modal";
import { DashboardKebabMenu } from "@/components/features/dashboards/dashboard-kebab-menu";
import type { Dashboard } from "@/types/models/dashboard";

/**
 * Sort dashboards by sortOrder ASC, then createdAt ASC as a stable tiebreaker.
 */
export function sortDashboards(list: Dashboard[]): Dashboard[] {
  return [...list].sort((a, b) => {
    if (a.attributes.sortOrder !== b.attributes.sortOrder) {
      return a.attributes.sortOrder - b.attributes.sortOrder;
    }
    return a.attributes.createdAt.localeCompare(b.attributes.createdAt);
  });
}

/**
 * Given a sorted list and active/over ids from a dnd-kit drag-end, returns
 * the reorder payload with 0-indexed sortOrder. Returns null when the drag
 * is a no-op.
 */
export function computeReorderEntries(
  sorted: Dashboard[],
  activeId: string,
  overId: string,
): Array<{ id: string; sortOrder: number }> | null {
  if (activeId === overId) return null;
  const fromIdx = sorted.findIndex((d) => d.id === activeId);
  const toIdx = sorted.findIndex((d) => d.id === overId);
  if (fromIdx < 0 || toIdx < 0) return null;
  const next = [...sorted];
  const [moved] = next.splice(fromIdx, 1);
  if (!moved) return null;
  next.splice(toIdx, 0, moved);
  return next.map((d, i) => ({ id: d.id, sortOrder: i }));
}

interface SortableDashboardRowProps {
  dashboard: Dashboard;
  defaultDashboardId: string | null;
}

function SortableDashboardRow({ dashboard, defaultDashboardId }: SortableDashboardRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: dashboard.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="group/row flex items-center gap-0.5"
    >
      <button
        type="button"
        aria-label={`Drag ${dashboard.attributes.name} to reorder`}
        className="flex h-7 w-5 shrink-0 cursor-grab items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity hover:text-sidebar-foreground focus-visible:opacity-100 group-hover/row:opacity-100 touch-none outline-none"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="h-3 w-3" />
      </button>
      <NavLink
        to={`/app/dashboards/${dashboard.id}`}
        className={({ isActive }) =>
          cn(
            "flex flex-1 items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
            isActive
              ? "bg-sidebar-accent text-sidebar-accent-foreground"
              : "text-sidebar-foreground hover:bg-sidebar-accent/50",
          )
        }
      >
        <LayoutDashboard className="h-4 w-4" />
        <span className="flex-1 truncate">{dashboard.attributes.name}</span>
      </NavLink>
      <DashboardKebabMenu
        dashboard={dashboard}
        isDefault={defaultDashboardId === dashboard.id}
      />
    </div>
  );
}

interface DashboardsNavGroupProps {
  isOpen: boolean;
  onToggle: () => void;
}

export function DashboardsNavGroup({ isOpen, onToggle }: DashboardsNavGroupProps) {
  const { data } = useDashboards();
  const { data: prefsData } = useHouseholdPreferences();
  const reorderMutation = useReorderDashboards();
  const [modalOpen, setModalOpen] = useState(false);

  const dashboards = data?.data ?? [];

  const { householdList, userList } = useMemo(() => {
    const household = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "household"),
    );
    const user = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "user"),
    );
    return { householdList: household, userList: user };
  }, [dashboards]);

  const defaultDashboardId = prefsData?.data?.[0]?.attributes.defaultDashboardId ?? null;

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  // One reorder call per scope — the backend rejects mixed-scope batches.
  const handleDragEnd = (scope: "household" | "user") => (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;
    const list = scope === "household" ? householdList : userList;
    const entries = computeReorderEntries(list, String(active.id), String(over.id));
    if (!entries) return;
    reorderMutation.mutate(entries);
  };

  const renderList = (list: Dashboard[], scope: "household" | "user") => (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd(scope)}
    >
      <SortableContext items={list.map((d) => d.id)} strategy={verticalListSortingStrategy}>
        <div className="space-y-0.5">
          {list.map((dashboard) => (
            <SortableDashboardRow
              key={dashboard.id}
              dashboard={dashboard}
              defaultDashboardId={defaultDashboardId}
            />
          ))}
        </div>
      </SortableContext>
    </DndContext>
  );

  return (
    <>
      <CollapsiblePrimitive.Root open={isOpen} onOpenChange={() => onToggle()}>
        <CollapsiblePrimitive.Trigger
          className={cn(
            "flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground transition-colors hover:text-sidebar-foreground",
          )}
        >
          <ChevronRight
            className={cn(
              "h-3 w-3 transition-transform duration-200",
              isOpen && "rotate-90",
            )}
          />
          Dashboards
        </CollapsiblePrimitive.Trigger>
        <CollapsiblePrimitive.Panel className="overflow-hidden transition-all duration-200 data-[state=closed]:animate-collapse data-[state=open]:animate-expand">
          <div className="space-y-2 pl-2">
            {householdList.length > 0 && renderList(householdList, "household")}
            {userList.length > 0 && (
              <div className="space-y-0.5">
                <p className="px-3 pt-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/80">
                  My Dashboards
                </p>
                {renderList(userList, "user")}
              </div>
            )}
            <button
              type="button"
              onClick={() => setModalOpen(true)}
              className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
            >
              <Plus className="h-4 w-4" />
              <span className="flex-1 text-left">New Dashboard</span>
            </button>
          </div>
        </CollapsiblePrimitive.Panel>
      </CollapsiblePrimitive.Root>
      <NewDashboardModal open={modalOpen} onOpenChange={setModalOpen} />
    </>
  );
}
