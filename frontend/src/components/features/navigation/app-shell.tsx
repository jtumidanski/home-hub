import { useMemo } from "react";
import { Outlet, Link } from "react-router-dom";
import { useAuth } from "@/components/providers/auth-provider";
import { BrandMark } from "@/components/common/brand-mark";
import { HouseholdSwitcher } from "@/components/features/households/household-switcher";
import { MobileHeader } from "@/components/features/navigation/mobile-header";
import { MobileDrawer } from "@/components/features/navigation/mobile-drawer";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { DashboardsNavGroup } from "@/components/features/navigation/dashboards-nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups } from "@/components/features/navigation/nav-config";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { usePackageSummary } from "@/lib/hooks/api/use-packages";

function AppShellContent() {
  const { toggleGroup, isGroupOpen } = useNavGroupState();
  const { appContext } = useAuth();
  const { data: packageSummary } = usePackageSummary();

  const navBadges = useMemo(() => ({
    pendingInvitationCount: appContext?.attributes.pendingInvitationCount ?? 0,
    inTransitCount: packageSummary?.data?.attributes?.inTransitCount ?? 0,
  }), [appContext?.attributes.pendingInvitationCount, packageSummary]);

  return (
    <>
      {/* Desktop sidebar */}
      <Sidebar>
        <SidebarHeader className="gap-3 border-b p-2">
          <Link
            to="/app"
            className="flex items-center gap-2.5 rounded-lg bg-sidebar-accent/50 px-2 py-2 transition-colors hover:bg-sidebar-accent"
          >
            <BrandMark className="size-8 shrink-0" />
            <span className="text-base font-semibold tracking-tight">Home Hub</span>
          </Link>
          <HouseholdSwitcher />
        </SidebarHeader>

        <SidebarContent>
          <DashboardsNavGroup
            isOpen={isGroupOpen("dashboards", true)}
            onToggle={() => toggleGroup("dashboards")}
          />
          {navGroups.map((group) => (
            <NavGroup
              key={group.key}
              group={group}
              isOpen={isGroupOpen(group.key, false)}
              onToggle={() => toggleGroup(group.key)}
              badges={navBadges}
            />
          ))}
        </SidebarContent>

        <SidebarFooter>
          <UserMenu />
        </SidebarFooter>
      </Sidebar>

      {/* Mobile header + drawer */}
      <MobileHeader />
      <MobileDrawer />

      <SidebarInset>
        <Outlet />
      </SidebarInset>
    </>
  );
}

export function AppShell() {
  return (
    <SidebarProvider>
      <AppShellContent />
    </SidebarProvider>
  );
}
