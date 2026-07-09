import { useEffect, useRef } from "react";
import { X } from "lucide-react";
import { createPortal } from "react-dom";
import { BrandMark } from "@/components/common/brand-mark";
import { MobileHouseholdSelector } from "@/components/features/households/mobile-household-selector";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";

export function MobileDrawer() {
  const { openMobile, setOpenMobile } = useSidebar();
  const drawerRef = useRef<HTMLDivElement>(null);
  const { toggleGroup, isGroupOpen } = useNavGroupState();
  const onClose = () => setOpenMobile(false);

  useEffect(() => {
    if (!openMobile) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpenMobile(false);
    };
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";
    drawerRef.current?.focus();
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
    };
  }, [openMobile, setOpenMobile]);

  return createPortal(
    <>
      {/* Overlay */}
      <div
        className={cn(
          "fixed inset-0 z-50 bg-black/40 transition-opacity duration-250",
          openMobile ? "opacity-100" : "pointer-events-none opacity-0",
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
          openMobile ? "translate-x-0" : "-translate-x-full",
        )}
      >
        {/* Header */}
        <div className="flex h-14 items-center justify-between border-b px-3">
          <div className="flex items-center gap-2">
            <BrandMark className="size-7 shrink-0" />
            <span className="text-lg font-semibold leading-none">Home Hub</span>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close navigation menu">
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Household switcher */}
        <div className="border-b p-3">
          <MobileHouseholdSelector onNavigate={onClose} />
        </div>

        {/* Nav groups */}
        <nav className="flex-1 space-y-3 overflow-y-auto p-2">
          <DashboardsNavGroup
            isOpen={isGroupOpen("dashboards", true)}
            onToggle={() => toggleGroup("dashboards")}
            onItemClick={onClose}
            size="lg"
          />
          {navGroups.map((group) => (
            <NavGroup
              key={group.key}
              group={group}
              isOpen={isGroupOpen(group.key, false)}
              onToggle={() => toggleGroup(group.key)}
              onItemClick={onClose}
              size="lg"
            />
          ))}
        </nav>

        <div className="border-t">
          <UserMenu onAction={onClose} iconSize="h-5 w-5" />
        </div>
      </div>
    </>,
    document.body,
  );
}
