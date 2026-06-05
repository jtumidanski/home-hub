import { useMemo, useState } from "react";
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
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { Plus } from "lucide-react";
import { useDashboards, useReorderDashboards } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { sortDashboards, computeReorderEntries } from "@/lib/dashboard/ordering";
import { DashboardRow } from "@/components/features/dashboards/dashboard-row";
import { NewDashboardModal } from "@/components/features/dashboards/new-dashboard-modal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import type { Dashboard } from "@/types/models/dashboard";

export function DashboardManagementPage() {
  const { data, isLoading, isError } = useDashboards();
  const { data: prefsData } = useHouseholdPreferences();
  const reorderMutation = useReorderDashboards();
  const [modalOpen, setModalOpen] = useState(false);

  const dashboards = useMemo(() => data?.data ?? [], [data]);

  const { householdList, userList } = useMemo(() => {
    const household = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "household"),
    );
    const user = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "user"),
    );
    return { householdList: household, userList: user };
  }, [dashboards]);

  const defaultDashboardId =
    prefsData?.data?.[0]?.attributes.defaultDashboardId ?? null;

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

  const renderSection = (
    title: string,
    list: Dashboard[],
    scope: "household" | "user",
  ) => (
    <section className="space-y-2">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
        {title}
      </h2>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd(scope)}
      >
        <SortableContext items={list.map((d) => d.id)} strategy={verticalListSortingStrategy}>
          <div className="space-y-2">
            {list.map((dashboard) => (
              <DashboardRow
                key={dashboard.id}
                dashboard={dashboard}
                defaultDashboardId={defaultDashboardId}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>
    </section>
  );

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load dashboards. Try refreshing the page." />
      </div>
    );
  }

  const isEmpty = householdList.length === 0 && userList.length === 0;

  return (
    <div className="p-4 md:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl md:text-2xl font-semibold">Dashboards</h1>
        <Button size="sm" onClick={() => setModalOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />New Dashboard
        </Button>
      </div>

      <NewDashboardModal open={modalOpen} onOpenChange={setModalOpen} />

      {isEmpty ? (
        <div className="flex flex-col items-center justify-center py-12 text-center space-y-4">
          <p className="text-muted-foreground">No dashboards yet.</p>
          <Button variant="outline" onClick={() => setModalOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />Create First Dashboard
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          {householdList.length > 0 &&
            renderSection("Household Dashboards", householdList, "household")}
          {userList.length > 0 && renderSection("My Dashboards", userList, "user")}
        </div>
      )}
    </div>
  );
}
