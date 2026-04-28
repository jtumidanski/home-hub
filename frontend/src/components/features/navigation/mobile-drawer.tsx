import { useEffect, useRef } from "react";
import { NavLink } from "react-router-dom";
import { Settings, X } from "lucide-react";
import { createPortal } from "react-dom";
import { MobileHouseholdSelector } from "@/components/features/households/mobile-household-selector";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups, settingsNavItem } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface MobileDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function MobileDrawer({ open, onClose }: MobileDrawerProps) {
  const drawerRef = useRef<HTMLDivElement>(null);
  const { toggleGroup, isGroupOpen } = useNavGroupState();

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";
    drawerRef.current?.focus();
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
    };
  }, [open, onClose]);

  return createPortal(
    <>
      {/* Overlay */}
      <div
        className={cn(
          "fixed inset-0 z-50 bg-black/40 transition-opacity duration-250",
          open ? "opacity-100" : "pointer-events-none opacity-0",
        )}
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer panel */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Navigation"
        tabIndex={-1}
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex w-72 flex-col bg-sidebar text-sidebar-foreground shadow-lg transition-transform duration-250 ease-in-out outline-none",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        {/* Header */}
        <div className="flex h-14 items-center justify-between border-b px-4">
          <span className="text-lg font-semibold">Home Hub</span>
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close navigation menu">
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Household switcher */}
        <div className="border-b p-3">
          <MobileHouseholdSelector onNavigate={onClose} />
        </div>

        {/* Nav groups */}
        <nav className="flex-1 space-y-3 overflow-y-auto p-3">
          <DashboardsNavGroup
            isOpen={isGroupOpen("dashboards", true)}
            onToggle={() => toggleGroup("dashboards")}
            onItemClick={onClose}
            iconSize="h-5 w-5"
            itemPadding="py-3"
          />
          {navGroups.map((group) => (
            <NavGroup
              key={group.key}
              group={group}
              isOpen={isGroupOpen(group.key, false)}
              onToggle={() => toggleGroup(group.key)}
              onItemClick={onClose}
              iconSize="h-5 w-5"
              itemPadding="py-3"
            />
          ))}
        </nav>

        {/* Footer */}
        <div className="border-t p-3">
          <NavLink
            to={settingsNavItem.to}
            onClick={onClose}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent/50",
              )
            }
          >
            <Settings className="h-5 w-5" />
            {settingsNavItem.label}
          </NavLink>
        </div>

        <div className="border-t">
          <UserMenu onAction={onClose} iconSize="h-5 w-5" />
        </div>
      </div>
    </>,
    document.body,
  );
}
