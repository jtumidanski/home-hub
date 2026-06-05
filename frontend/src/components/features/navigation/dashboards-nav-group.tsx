import { useMemo } from "react";
import { NavLink } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight, LayoutDashboard } from "lucide-react";
import { cn } from "@/lib/utils";
import { useDashboards } from "@/lib/hooks/api/use-dashboards";
import { sortDashboards } from "@/lib/dashboard/ordering";
import type { Dashboard } from "@/types/models/dashboard";

interface DashboardsNavGroupProps {
  isOpen: boolean;
  onToggle: () => void;
  onItemClick?: () => void;
  iconSize?: string;
  itemPadding?: string;
}

export function DashboardsNavGroup({
  isOpen,
  onToggle,
  onItemClick,
  iconSize = "h-4 w-4",
  itemPadding = "py-2",
}: DashboardsNavGroupProps) {
  const { data } = useDashboards();

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

  const renderLink = (dashboard: Dashboard) => (
    <NavLink
      key={dashboard.id}
      to={`/app/dashboards/${dashboard.id}`}
      onClick={onItemClick}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
          itemPadding,
          isActive
            ? "bg-sidebar-accent text-sidebar-accent-foreground"
            : "text-sidebar-foreground hover:bg-sidebar-accent/50",
        )
      }
    >
      <LayoutDashboard className={iconSize} />
      <span className="flex-1 truncate">{dashboard.attributes.name}</span>
    </NavLink>
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
          {householdList.length > 0 && (
            <div className="space-y-0.5">{householdList.map(renderLink)}</div>
          )}
          {userList.length > 0 && (
            <div className="space-y-0.5">
              <p className="px-3 pt-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/80">
                My Dashboards
              </p>
              {userList.map(renderLink)}
            </div>
          )}
        </div>
      </CollapsiblePrimitive.Panel>
    </CollapsiblePrimitive.Root>
  );
}
