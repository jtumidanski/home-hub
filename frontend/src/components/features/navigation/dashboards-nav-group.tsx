import { useMemo } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight, LayoutDashboard } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useDashboards } from "@/lib/hooks/api/use-dashboards";
import { sortDashboards } from "@/lib/dashboard/ordering";
import type { Dashboard } from "@/types/models/dashboard";

interface DashboardsNavGroupProps {
  isOpen: boolean;
  onToggle: () => void;
  onItemClick?: () => void;
  size?: "default" | "lg";
}

const groupLabelClasses =
  "flex h-7 w-full shrink-0 cursor-pointer items-center gap-2 rounded-md px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground transition-colors hover:text-sidebar-foreground";

export function DashboardsNavGroup({
  isOpen,
  onToggle,
  onItemClick,
  size = "default",
}: DashboardsNavGroupProps) {
  const { data } = useDashboards();
  const location = useLocation();

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
    <SidebarMenuItem key={dashboard.id}>
      <SidebarMenuButton
        asChild
        size={size}
        isActive={location.pathname === `/app/dashboards/${dashboard.id}`}
      >
        <NavLink to={`/app/dashboards/${dashboard.id}`} onClick={onItemClick}>
          <LayoutDashboard />
          <span className="flex-1 truncate">{dashboard.attributes.name}</span>
        </NavLink>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );

  return (
    <SidebarGroup>
      <CollapsiblePrimitive.Root open={isOpen} onOpenChange={() => onToggle()}>
        <CollapsiblePrimitive.Trigger className={groupLabelClasses}>
          <ChevronRight
            className={cn("h-3 w-3 transition-transform duration-200", isOpen && "rotate-90")}
          />
          Dashboards
        </CollapsiblePrimitive.Trigger>
        <CollapsiblePrimitive.Panel className="overflow-hidden transition-all duration-200 data-[state=closed]:animate-collapse data-[state=open]:animate-expand">
          <SidebarGroupContent className="space-y-2 pl-2 pt-0.5">
            {householdList.length > 0 && (
              <SidebarMenu>{householdList.map(renderLink)}</SidebarMenu>
            )}
            {userList.length > 0 && (
              <div className="space-y-0.5">
                <p className="px-3 pt-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/80">
                  My Dashboards
                </p>
                <SidebarMenu>{userList.map(renderLink)}</SidebarMenu>
              </div>
            )}
          </SidebarGroupContent>
        </CollapsiblePrimitive.Panel>
      </CollapsiblePrimitive.Root>
    </SidebarGroup>
  );
}
