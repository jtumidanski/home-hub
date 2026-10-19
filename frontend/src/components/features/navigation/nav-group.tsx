import { NavLink, useLocation } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { isNavItemActive, type NavGroup as NavGroupConfig } from "./nav-config";

interface NavGroupProps {
  group: NavGroupConfig;
  isOpen: boolean;
  onToggle: () => void;
  onItemClick?: () => void;
  size?: "default" | "lg";
  badges?: Record<string, number>;
}

/** The one group-label recipe, shared with DashboardsNavGroup. */
export const groupLabelClasses =
  "flex h-7 w-full shrink-0 cursor-pointer items-center gap-2 rounded-md px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground transition-colors hover:text-sidebar-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring motion-reduce:transition-none coarse:min-h-11";

export function NavGroup({
  group,
  isOpen,
  onToggle,
  onItemClick,
  size = "default",
  badges = {},
}: NavGroupProps) {
  const location = useLocation();
  const hasActiveRoute = group.items.some((item) =>
    isNavItemActive(location.pathname, item.to, item.end),
  );

  const open = hasActiveRoute || isOpen;

  return (
    <SidebarGroup>
      <CollapsiblePrimitive.Root open={open} onOpenChange={() => onToggle()}>
        <CollapsiblePrimitive.Trigger className={groupLabelClasses}>
          <ChevronRight
            className={cn("size-3 transition-transform duration-200 motion-reduce:transition-none", open && "rotate-90")}
          />
          {group.label}
        </CollapsiblePrimitive.Trigger>
        <CollapsiblePrimitive.Panel className="overflow-hidden">
          <SidebarGroupContent className="pl-2 pt-0.5">
            <SidebarMenu>
              {group.items.map(({ to, icon: Icon, label, end, badgeKey }) => (
                <SidebarMenuItem key={to}>
                  <SidebarMenuButton
                    asChild
                    size={size}
                    isActive={isNavItemActive(location.pathname, to, end)}
                  >
                    <NavLink
                      to={to}
                      end={end ?? false}
                      onClick={onItemClick}
                      className="aria-[current=page]:text-sidebar-primary"
                    >
                      <Icon />
                      <span className="flex-1">{label}</span>
                      {badgeKey && badges[badgeKey] != null && badges[badgeKey]! > 0 && (
                        <SidebarMenuBadge>{badges[badgeKey]}</SidebarMenuBadge>
                      )}
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </CollapsiblePrimitive.Panel>
      </CollapsiblePrimitive.Root>
    </SidebarGroup>
  );
}
