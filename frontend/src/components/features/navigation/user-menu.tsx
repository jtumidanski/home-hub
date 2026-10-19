import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { LogOut, ChevronDown, Settings } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/components/providers/auth-provider";
import { useLogout } from "@/lib/hooks/api/use-auth";
import { settingsNavItem } from "@/components/features/navigation/nav-config";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";

interface UserMenuProps {
  onAction?: () => void;
  size?: "default" | "lg";
}

const menuItemClasses =
  "flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none coarse:min-h-11 data-highlighted:bg-accent data-highlighted:text-accent-foreground focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring";

export function UserMenu({ onAction, size = "default" }: UserMenuProps) {
  const { user } = useAuth();
  const logout = useLogout();
  const navigate = useNavigate();

  if (!user) return null;

  const iconSize = size === "lg" ? "size-5" : "size-4";

  return (
    <MenuPrimitive.Root>
      <MenuPrimitive.Trigger
        className={cn(
          "flex w-full cursor-pointer items-center justify-between rounded-md p-3 text-left transition-colors motion-reduce:transition-none coarse:min-h-11 hover:bg-sidebar-accent/50 outline-none focus-visible:ring-2 focus-visible:ring-ring",
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
        <ChevronDown className="ml-2 size-4 shrink-0 text-muted-foreground" />
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
              className={menuItemClasses}
              onClick={() => {
                navigate(settingsNavItem.to);
                onAction?.();
              }}
            >
              <Settings className={iconSize} />
              {settingsNavItem.label}
            </MenuPrimitive.Item>
            <MenuPrimitive.Item
              className={cn(menuItemClasses, "text-destructive data-highlighted:text-destructive")}
              onClick={() => {
                logout.mutate();
                onAction?.();
              }}
            >
              <LogOut className={iconSize} />
              Sign out
            </MenuPrimitive.Item>
          </MenuPrimitive.Popup>
        </MenuPrimitive.Positioner>
      </MenuPrimitive.Portal>
    </MenuPrimitive.Root>
  );
}
