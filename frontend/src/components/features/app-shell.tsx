import { Outlet, NavLink } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Home, CheckSquare, Bell, Settings, LogOut, Moon, Sun } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useTheme } from "@/components/providers/theme-provider";
import { authService } from "@/services/api/auth";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { getErrorMessage } from "@/lib/api/errors";
import { HouseholdSwitcher } from "@/components/features/household-switcher";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/app", icon: Home, label: "Dashboard", end: true },
  { to: "/app/tasks", icon: CheckSquare, label: "Tasks", end: false },
  { to: "/app/reminders", icon: Bell, label: "Reminders", end: false },
  { to: "/app/households", icon: Home, label: "Households", end: false },
  { to: "/app/settings", icon: Settings, label: "Settings", end: false },
];

export function AppShell() {
  const { user, appContext } = useAuth();
  const { tenantId } = useTenant();
  const { theme, setTheme } = useTheme();
  const queryClient = useQueryClient();

  const handleLogout = async () => {
    try {
      await authService.logout();
    } finally {
      window.location.href = "/login";
    }
  };

  const handleThemeToggle = async () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    if (tenantId && appContext?.relationships?.preference?.data?.id) {
      try {
        await accountService.updatePreferenceTheme(
          tenantId,
          appContext.relationships.preference.data.id,
          newTheme
        );
        await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      } catch (error) {
        toast.error(getErrorMessage(error, "Failed to save theme preference"));
      }
    }
  };

  return (
    <div className="flex min-h-screen">
      <aside className="flex w-64 flex-col border-r bg-sidebar text-sidebar-foreground">
        <div className="flex h-14 items-center border-b px-4">
          <h1 className="text-lg font-semibold">Home Hub</h1>
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
            onClick={handleThemeToggle}
          >
            {theme === "light" ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
            {theme === "light" ? "Dark Mode" : "Light Mode"}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-3 text-destructive"
            onClick={handleLogout}
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

      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
