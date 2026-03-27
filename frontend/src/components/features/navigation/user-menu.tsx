import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { Moon, Sun, LogOut, ChevronDown } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";

interface UserMenuProps {
  onAction?: () => void;
  iconSize?: string;
}

export function UserMenu({ onAction, iconSize = "h-4 w-4" }: UserMenuProps) {
  const { user } = useAuth();
  const { theme, toggleTheme } = useThemeToggle();
  const logout = useLogout();

  if (!user) return null;

  return (
    <MenuPrimitive.Root>
      <MenuPrimitive.Trigger
        className={cn(
          "flex w-full cursor-pointer items-center justify-between rounded-md p-3 text-left transition-colors hover:bg-sidebar-accent/50 outline-none focus-visible:ring-2 focus-visible:ring-ring",
        )}
      >
        <UserAvatar
          avatarUrl={user.attributes.avatarUrl}
          providerAvatarUrl={user.attributes.providerAvatarUrl}
          displayName={user.attributes.displayName}
          userId={user.id}
          size="sm"
        />
        <div className="ml-2 min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{user.attributes.displayName}</p>
          <p className="truncate text-xs text-muted-foreground">{user.attributes.email}</p>
        </div>
        <ChevronDown className="ml-2 h-4 w-4 shrink-0 text-muted-foreground" />
      </MenuPrimitive.Trigger>
      <MenuPrimitive.Portal>
        <MenuPrimitive.Positioner side="top" sideOffset={8} align="start" className="z-50">
          <MenuPrimitive.Popup
            className={cn(
              "min-w-48 rounded-lg bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/10",
              "origin-(--transform-origin) transition-[transform,scale,opacity] duration-100",
              "data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95",
              "data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
            )}
          >
            <MenuPrimitive.Item
              className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
              onClick={() => {
                toggleTheme();
                onAction?.();
              }}
            >
              {theme === "light" ? (
                <Moon className={iconSize} />
              ) : (
                <Sun className={iconSize} />
              )}
              {theme === "light" ? "Dark Mode" : "Light Mode"}
            </MenuPrimitive.Item>
            <MenuPrimitive.Item
              className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm text-destructive outline-none select-none hover:bg-accent focus:bg-accent"
              onClick={() => {
                logout.mutate();
                onAction?.();
              }}
            >
              <LogOut className={iconSize} />
              Sign Out
            </MenuPrimitive.Item>
          </MenuPrimitive.Popup>
        </MenuPrimitive.Positioner>
      </MenuPrimitive.Portal>
    </MenuPrimitive.Root>
  );
}
