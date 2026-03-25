import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";

interface MobileHeaderProps {
  onMenuOpen: () => void;
}

export function MobileHeader({ onMenuOpen }: MobileHeaderProps) {
  return (
    <header className="flex h-14 items-center border-b bg-sidebar px-4 md:hidden">
      <Button
        variant="ghost"
        size="icon"
        onClick={onMenuOpen}
        aria-label="Open navigation menu"
      >
        <Menu className="h-5 w-5" />
      </Button>
      <span className="ml-3 text-lg font-semibold">Home Hub</span>
    </header>
  );
}
