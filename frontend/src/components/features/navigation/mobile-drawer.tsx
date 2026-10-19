import { X } from "lucide-react";
import { BrandMark } from "@/components/common/brand-mark";
import { MobileHouseholdSelector } from "@/components/features/households/mobile-household-selector";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
} from "@/components/ui/dialog";
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { useSidebar } from "@/components/ui/sidebar";

export function MobileDrawer() {
  const { openMobile, setOpenMobile } = useSidebar();
  const { toggleGroup, isGroupOpen } = useNavGroupState();
  const onClose = () => setOpenMobile(false);

  // base-ui Dialog unmounts the panel when closed, traps focus while open,
  // marks the page behind inert, and returns focus to the opener on close.
  return (
    <Dialog open={openMobile} onOpenChange={setOpenMobile}>
      <DialogPortal>
        <DialogOverlay />
        <DialogPrimitive.Popup
          className="fixed inset-y-0 left-0 z-50 flex w-72 flex-col bg-sidebar text-sidebar-foreground shadow-overlay outline-none duration-(--duration-base) data-open:animate-in data-open:slide-in-from-left data-closed:animate-out data-closed:slide-out-to-left"
        >
          <DialogTitle className="sr-only">Navigation</DialogTitle>
          {/* Header */}
          <div className="flex h-14 items-center justify-between border-b px-3">
            <div className="flex items-center gap-2">
              <BrandMark className="size-7 shrink-0" />
              <span className="font-heading text-lg leading-none tracking-[-0.01em]">Home Hub</span>
            </div>
            <DialogClose
              render={<Button variant="ghost" size="icon" aria-label="Close navigation menu" />}
            >
              <X className="size-5" />
            </DialogClose>
          </div>

          {/* Household switcher */}
          <div className="border-b p-3">
            <MobileHouseholdSelector onNavigate={onClose} />
          </div>

          {/* Nav groups */}
          <nav aria-label="Main" className="flex-1 space-y-3 overflow-y-auto p-2">
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
            <UserMenu onAction={onClose} size="lg" />
          </div>
        </DialogPrimitive.Popup>
      </DialogPortal>
    </Dialog>
  );
}
