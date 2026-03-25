import { useState } from "react";
import { Outlet, NavLink, Link } from "react-router-dom";
import { Home, CheckSquare, Bell, CloudSun, UtensilsCrossed, Settings, LogOut, Moon, Sun } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { HouseholdSwitcher } from "@/components/features/households/household-switcher";
import { MobileHeader } from "@/components/features/navigation/mobile-header";
import { MobileDrawer } from "@/components/features/navigation/mobile-drawer";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/app", icon: Home, label: "Dashboard", end: true },
  { to: "/app/tasks", icon: CheckSquare, label: "Tasks", end: false },
  { to: "/app/reminders", icon: Bell, label: "Reminders", end: false },
  { to: "/app/recipes", icon: UtensilsCrossed, label: "Recipes", end: false },
  { to: "/app/weather", icon: CloudSun, label: "Weather", end: false },
  { to: "/app/households", icon: Home, label: "Households", end: false },
  { to: "/app/settings", icon: Settings, label: "Settings", end: false },
];

export function AppShell() {
  const { user } = useAuth();
  const { theme, toggleTheme } = useThemeToggle();
  const logout = useLogout();
  const [drawerOpen, setDrawerOpen] = useState(false);

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

        <nav className="flex-1 space-y-1 p-2">
          {navItems.map(({ to, icon: Icon, label, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground"
                    : "text-sidebar-foreground hover:bg-sidebar-accent/50"
                )
              }
            >
              <Icon className="h-4 w-4" />
              {label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t p-2 space-y-1">
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-3"
            onClick={toggleTheme}
          >
            {theme === "light" ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
            {theme === "light" ? "Dark Mode" : "Light Mode"}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-3 text-destructive"
            onClick={() => logout.mutate()}
          >
            <LogOut className="h-4 w-4" />
            Sign Out
          </Button>
        </div>

        {user && (
          <div className="border-t p-4">
            <p className="truncate text-sm font-medium">{user.attributes.displayName}</p>
            <p className="truncate text-xs text-muted-foreground">{user.attributes.email}</p>
          </div>
        )}
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
