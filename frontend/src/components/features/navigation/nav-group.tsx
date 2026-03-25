import { NavLink, useLocation } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import type { NavGroup as NavGroupConfig } from "./nav-config";

interface NavGroupProps {
  group: NavGroupConfig;
  isOpen: boolean;
  onToggle: () => void;
  onItemClick?: () => void;
  iconSize?: string;
  itemPadding?: string;
}

export function NavGroup({
  group,
  isOpen,
  onToggle,
  onItemClick,
  iconSize = "h-4 w-4",
  itemPadding = "py-2",
}: NavGroupProps) {
  const location = useLocation();
  const hasActiveRoute = group.items.some((item) =>
    item.end ? location.pathname === item.to : location.pathname.startsWith(item.to),
  );

  const open = hasActiveRoute || isOpen;

  return (
    <CollapsiblePrimitive.Root open={open} onOpenChange={() => onToggle()}>
      <CollapsiblePrimitive.Trigger
        className={cn(
          "flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground transition-colors hover:text-sidebar-foreground",
        )}
      >
        <ChevronRight
          className={cn(
            "h-3 w-3 transition-transform duration-200",
            open && "rotate-90",
          )}
        />
        {group.label}
      </CollapsiblePrimitive.Trigger>
      <CollapsiblePrimitive.Panel className="overflow-hidden transition-all duration-200 data-[state=closed]:animate-collapse data-[state=open]:animate-expand">
        <div className="space-y-0.5 pl-2">
          {group.items.map(({ to, icon: Icon, label, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end ?? false}
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
              <Icon className={iconSize} />
              {label}
            </NavLink>
          ))}
        </div>
      </CollapsiblePrimitive.Panel>
    </CollapsiblePrimitive.Root>
  );
}
