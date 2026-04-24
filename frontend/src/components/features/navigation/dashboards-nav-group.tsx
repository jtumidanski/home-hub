import { useMemo } from "react";
import { NavLink } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight, LayoutDashboard, Plus } from "lucide-react";
import { cn } from "@/lib/utils";
import { useDashboards } from "@/lib/hooks/api/use-dashboards";
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

interface DashboardRowProps {
  dashboard: Dashboard;
}

function DashboardRow({ dashboard }: DashboardRowProps) {
  return (
    <div className="group/row flex items-center gap-0.5">
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
    </div>
  );
}

interface DashboardsNavGroupProps {
  isOpen: boolean;
  onToggle: () => void;
  onNewDashboard?: () => void;
}

export function DashboardsNavGroup({
  isOpen,
  onToggle,
  onNewDashboard,
}: DashboardsNavGroupProps) {
  const { data } = useDashboards();

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

  const renderList = (list: Dashboard[]) => (
    <div className="space-y-0.5">
      {list.map((dashboard) => (
        <DashboardRow key={dashboard.id} dashboard={dashboard} />
      ))}
    </div>
  );

  return (
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
          {householdList.length > 0 && renderList(householdList)}
          {userList.length > 0 && (
            <div className="space-y-0.5">
              <p className="px-3 pt-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/80">
                My Dashboards
              </p>
              {renderList(userList)}
            </div>
          )}
          <button
            type="button"
            onClick={onNewDashboard}
            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
          >
            <Plus className="h-4 w-4" />
            <span className="flex-1 text-left">New Dashboard</span>
          </button>
        </div>
      </CollapsiblePrimitive.Panel>
    </CollapsiblePrimitive.Root>
  );
}
