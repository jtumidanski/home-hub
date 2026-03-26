import { useState, useMemo } from "react";
import { Outlet, NavLink, Link } from "react-router-dom";
import { Settings } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { HouseholdSwitcher } from "@/components/features/households/household-switcher";
import { MobileHeader } from "@/components/features/navigation/mobile-header";
import { MobileDrawer } from "@/components/features/navigation/mobile-drawer";
import { NavGroup } from "@/components/features/navigation/nav-group";
import { UserMenu } from "@/components/features/navigation/user-menu";
import { navGroups, settingsNavItem } from "@/components/features/navigation/nav-config";
import { useNavGroupState } from "@/lib/hooks/use-nav-group-state";
import { cn } from "@/lib/utils";

export function AppShell() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { toggleGroup, isGroupOpen } = useNavGroupState();
  const { appContext } = useAuth();

  const navBadges = useMemo(() => ({
    pendingInvitationCount: appContext?.attributes.pendingInvitationCount ?? 0,
  }), [appContext?.attributes.pendingInvitationCount]);

  return (
    <div className="flex h-screen flex-col md:flex-row overflow-hidden">
      {/* Desktop sidebar */}
      <aside className="hidden md:flex w-64 flex-col border-r bg-sidebar text-sidebar-foreground">
        <div className="flex h-14 items-center border-b px-4">
          <Link to="/app" className="text-lg font-semibold hover:opacity-80 transition-opacity">Home Hub</Link>
        </div>

        <div className="border-b p-2">
          <HouseholdSwitcher />
        </div>

        <nav className="flex-1 space-y-3 overflow-y-auto p-2">
          {navGroups.map((group) => (
            <NavGroup
              key={group.key}
              group={group}
              isOpen={isGroupOpen(group.key, false)}
              onToggle={() => toggleGroup(group.key)}
              badges={navBadges}
            />
          ))}
        </nav>

        <div className="border-t p-2">
          <NavLink
            to={settingsNavItem.to}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent/50",
              )
            }
          >
            <Settings className="h-4 w-4" />
            {settingsNavItem.label}
          </NavLink>
        </div>

        <div className="border-t">
          <UserMenu />
        </div>
      </aside>

      {/* Mobile header */}
      <MobileHeader onMenuOpen={() => setDrawerOpen(true)} />
      <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />

      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
