import { useEffect, useRef } from "react";
import { NavLink } from "react-router-dom";
import { Home, CheckSquare, Bell, Settings, LogOut, Moon, Sun, X } from "lucide-react";
import { createPortal } from "react-dom";
import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { MobileHouseholdSelector } from "@/components/features/households/mobile-household-selector";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/app", icon: Home, label: "Dashboard", end: true },
  { to: "/app/tasks", icon: CheckSquare, label: "Tasks", end: false },
  { to: "/app/reminders", icon: Bell, label: "Reminders", end: false },
  { to: "/app/households", icon: Home, label: "Households", end: false },
  { to: "/app/settings", icon: Settings, label: "Settings", end: false },
];

interface MobileDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function MobileDrawer({ open, onClose }: MobileDrawerProps) {
  const { user } = useAuth();
  const { theme, toggleTheme } = useThemeToggle();
  const logout = useLogout();
  const drawerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleKeyDown);
    // Prevent body scroll when drawer is open
    document.body.style.overflow = "hidden";
    // Focus the drawer
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

        {/* Nav items */}
        <nav className="flex-1 space-y-1 p-3">
          {navItems.map(({ to, icon: Icon, label, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
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
              <Icon className="h-5 w-5" />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* Footer actions */}
        <div className="border-t p-3 space-y-1">
          <Button
            variant="ghost"
            className="w-full justify-start gap-3 py-3"
            onClick={toggleTheme}
          >
            {theme === "light" ? <Moon className="h-5 w-5" /> : <Sun className="h-5 w-5" />}
            {theme === "light" ? "Dark Mode" : "Light Mode"}
          </Button>
          <Button
            variant="ghost"
            className="w-full justify-start gap-3 py-3 text-destructive"
            onClick={() => { logout.mutate(); onClose(); }}
          >
            <LogOut className="h-5 w-5" />
            Sign Out
          </Button>
        </div>

        {user && (
          <div className="border-t p-4">
            <p className="truncate text-sm font-medium">{user.attributes.displayName}</p>
            <p className="truncate text-xs text-muted-foreground">{user.attributes.email}</p>
          </div>
        )}
      </div>
    </>,
    document.body,
  );
}
