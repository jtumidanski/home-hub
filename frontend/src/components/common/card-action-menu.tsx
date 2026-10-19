import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface CardAction {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  variant?: "default" | "destructive";
}

interface CardActionMenuProps {
  actions: CardAction[];
}

export function CardActionMenu({ actions }: CardActionMenuProps) {
  return (
    <MenuPrimitive.Root>
      <MenuPrimitive.Trigger
        render={<Button variant="ghost" size="icon" aria-label="Actions" />}
        onClick={(e) => e.stopPropagation()}
      >
        <MoreVertical className="h-4 w-4" />
      </MenuPrimitive.Trigger>
      <MenuPrimitive.Portal>
        <MenuPrimitive.Positioner side="bottom" align="end" sideOffset={4} className="z-50">
          <MenuPrimitive.Popup
            className={cn(
              "min-w-40 rounded-lg bg-popover p-1 text-popover-foreground shadow-overlay ring-1 ring-foreground/10",
              "origin-(--transform-origin) transition-[transform,scale,opacity] duration-(--duration-fast) ease-out",
              "data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95",
              "data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
            )}
          >
            {actions.map((action) => (
              <MenuPrimitive.Item
                key={action.label}
                onClick={(e) => {
                  e.stopPropagation();
                  action.onClick();
                }}
                className={cn(
                  "flex w-full cursor-pointer items-center gap-2 rounded-md px-3 py-2.5 text-sm outline-none select-none coarse:min-h-11 data-highlighted:bg-accent data-highlighted:text-accent-foreground focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring",
                  action.variant === "destructive" && "text-destructive data-highlighted:text-destructive",
                )}
              >
                {action.icon}
                {action.label}
              </MenuPrimitive.Item>
            ))}
          </MenuPrimitive.Popup>
        </MenuPrimitive.Positioner>
      </MenuPrimitive.Portal>
    </MenuPrimitive.Root>
  );
}
